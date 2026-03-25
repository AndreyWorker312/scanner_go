"""
Core change detection logic.
Compares consecutive scan results from MongoDB and generates ChangeEvent records.
"""
import logging
from datetime import datetime, timezone

from pymongo import MongoClient, ASCENDING, DESCENDING

logger = logging.getLogger(__name__)

# Ports that are typical indicators of malware / backdoors
BACKDOOR_PORTS = {
    4444, 1337, 31337, 5555, 6666, 7777, 8888, 9999,
    12345, 54321, 1234, 6667, 6668, 4899, 2222, 3333,
    1111, 2323, 6969, 27374, 65000, 65535,
}

# Ports associated with important services (disappearance = HIGH alert)
CRITICAL_SERVICES = {
    22: "SSH", 80: "HTTP", 443: "HTTPS", 3306: "MySQL",
    5432: "PostgreSQL", 6379: "Redis", 27017: "MongoDB",
    8080: "HTTP-Alt", 8443: "HTTPS-Alt",
}


class ChangeDetector:
    def __init__(self, mongo_uri: str, db_name: str):
        self.client = MongoClient(mongo_uri, serverSelectionTimeoutMS=5000)
        self.db = self.client[db_name]

    def run_detection(self) -> list:
        """
        Run all detection routines, save new events to MongoDB and return them.

        Returns a list of newly saved events so the caller (main.py) can
        publish them to RabbitMQ for backend → WebSocket broadcast.
        """
        logger.info("Running change detection cycle...")
        saved: list = []

        for detect_fn in (self._detect_arp_changes, self._detect_nmap_changes):
            try:
                events = detect_fn()
                for event in events:
                    if self._save_event_if_new(event):
                        saved.append(event)
            except Exception as exc:
                logger.error("Detection error in %s: %s", detect_fn.__name__, exc, exc_info=True)

        logger.info("Detection cycle complete — %d new event(s) saved.", len(saved))
        return saved

    # ─────────────────────────────────────────────────────────────────────────
    # ARP detection
    # ─────────────────────────────────────────────────────────────────────────

    def _detect_arp_changes(self) -> list:
        events = []
        ip_ranges = self.db.arp_history.distinct("ip_range")

        for ip_range in ip_ranges:
            scans = list(
                self.db.arp_history.find(
                    {"ip_range": ip_range},
                    sort=[("created_at", DESCENDING)],
                    limit=2,
                )
            )
            if len(scans) < 2:
                continue

            new_scan, old_scan = scans[0], scans[1]
            scan_b_id = str(new_scan["_id"])

            new_devices = {d["ip"]: d for d in new_scan.get("online_devices", [])}
            old_devices = {d["ip"]: d for d in old_scan.get("online_devices", [])}

            # New devices
            for ip, device in new_devices.items():
                if ip not in old_devices:
                    mac = device.get("mac", "unknown")
                    vendor = device.get("vendor", "unknown")
                    events.append({
                        "event_id":    f"NEW_DEVICE:{ip}:{scan_b_id}",
                        "event_type":  "NEW_DEVICE",
                        "severity":    "HIGH",
                        "title":       f"Новое устройство {ip} появилось в сети",
                        "description": f"MAC: {mac}, Производитель: {vendor}, Сеть: {ip_range}",
                        "target":      ip,
                        "service":     "arp",
                        "action":      "Проверить MAC-адрес и авторизацию устройства",
                        "scanner":     "arp",
                        "details":     {"ip": ip, "mac": mac, "vendor": vendor, "ip_range": ip_range},
                    })

            # Disappeared devices
            for ip, device in old_devices.items():
                if ip not in new_devices:
                    mac = device.get("mac", "unknown")
                    svc_name = CRITICAL_SERVICES.get(0, "")
                    severity = "HIGH" if ip in [d["ip"] for d in new_scan.get("devices", [])] else "MEDIUM"
                    events.append({
                        "event_id":    f"DEVICE_GONE:{ip}:{scan_b_id}",
                        "event_type":  "DEVICE_GONE",
                        "severity":    "MEDIUM",
                        "title":       f"Устройство {ip} пропало из сети",
                        "description": f"MAC: {mac} больше не отвечает в сети {ip_range}",
                        "target":      ip,
                        "service":     "arp",
                        "action":      "Проверить доступность и состояние устройства",
                        "scanner":     "arp",
                        "details":     {"ip": ip, "mac": mac, "ip_range": ip_range},
                    })

        return events

    # ─────────────────────────────────────────────────────────────────────────
    # Nmap detection
    # ─────────────────────────────────────────────────────────────────────────

    def _detect_nmap_changes(self) -> list:
        events = []
        ips = self.db.nmap_tcp_udp_history.distinct("ip")

        for ip in ips:
            scans = list(
                self.db.nmap_tcp_udp_history.find(
                    {"ip": ip},
                    sort=[("created_at", DESCENDING)],
                    limit=2,
                )
            )
            if len(scans) < 2:
                continue

            new_scan, old_scan = scans[0], scans[1]
            scan_b_id = str(new_scan["_id"])

            new_ports = self._extract_open_ports(new_scan)
            old_ports = self._extract_open_ports(old_scan)

            # New ports opened
            for port, info in new_ports.items():
                if port not in old_ports:
                    service  = info.get("service", "unknown")
                    protocol = info.get("protocol", "tcp")
                    is_back  = port in BACKDOOR_PORTS
                    severity = "CRITICAL" if is_back else "HIGH"
                    action   = (
                        "⚠️ Возможный бэкдор! Немедленно проверить процессы!"
                        if is_back
                        else f"Проверить назначение порта {port}/{protocol}"
                    )
                    events.append({
                        "event_id":    f"NEW_PORT:{ip}:{port}:{scan_b_id}",
                        "event_type":  "NEW_PORT",
                        "severity":    severity,
                        "title":       f"Новый порт {port} открыт на {ip}",
                        "description": f"Сервис: {service}, Протокол: {protocol}",
                        "target":      ip,
                        "service":     service,
                        "action":      action,
                        "scanner":     "nmap",
                        "details":     {"ip": ip, "port": port, "service": service, "protocol": protocol},
                    })

            # Ports closed
            for port, info in old_ports.items():
                if port not in new_ports:
                    service  = info.get("service", "unknown")
                    crit_svc = CRITICAL_SERVICES.get(port)
                    severity = "HIGH" if crit_svc else "LOW"
                    action   = (
                        f"⚠️ Критический сервис {crit_svc} недоступен! Проверить сервер."
                        if crit_svc
                        else "Порт закрыт — проверить, было ли это запланировано"
                    )
                    events.append({
                        "event_id":    f"PORT_CLOSED:{ip}:{port}:{scan_b_id}",
                        "event_type":  "PORT_CLOSED",
                        "severity":    severity,
                        "title":       f"Порт {port} закрыт на {ip}",
                        "description": f"Сервис {service} ({port}/{info.get('protocol','tcp')}) больше не доступен",
                        "target":      ip,
                        "service":     service,
                        "action":      action,
                        "scanner":     "nmap",
                        "details":     {"ip": ip, "port": port, "service": service},
                    })

            # Service / version changes
            for port, new_info in new_ports.items():
                if port in old_ports:
                    old_svc = old_ports[port].get("service", "")
                    new_svc = new_info.get("service", "")
                    if old_svc and new_svc and old_svc != new_svc:
                        events.append({
                            "event_id":    f"VERSION_CHANGE:{ip}:{port}:{scan_b_id}",
                            "event_type":  "VERSION_CHANGE",
                            "severity":    "MEDIUM",
                            "title":       f"Сервис на порту {port} изменился на {ip}",
                            "description": f"Было: {old_svc} → Стало: {new_svc}",
                            "target":      ip,
                            "service":     new_svc,
                            "action":      "Проверить корректность обновления или несанкционированное изменение",
                            "scanner":     "nmap",
                            "details":     {"ip": ip, "port": port, "old_service": old_svc, "new_service": new_svc},
                        })

        return events

    # ─────────────────────────────────────────────────────────────────────────
    # Helpers
    # ─────────────────────────────────────────────────────────────────────────

    @staticmethod
    def _extract_open_ports(scan: dict) -> dict:
        """Return {port_number: {service, protocol, state}} for all open ports."""
        ports: dict = {}
        for port_info in scan.get("port_info", []):
            all_ports    = port_info.get("all_ports", [])
            service_names = port_info.get("service_name", [])
            protocols    = port_info.get("protocols", [])
            states       = port_info.get("state", [])

            for i, port in enumerate(all_ports):
                state    = states[i]       if i < len(states)        else "open"
                service  = service_names[i] if i < len(service_names) else "unknown"
                protocol = protocols[i]    if i < len(protocols)     else "tcp"
                if str(state).lower() == "open":
                    ports[int(port)] = {"service": service, "protocol": protocol}
        return ports

    def _save_event_if_new(self, event: dict) -> bool:
        """Insert event into change_events collection; skip if event_id already exists."""
        if self.db.change_events.find_one({"event_id": event["event_id"]}, {"_id": 1}):
            return False
        event["created_at"] = datetime.now(timezone.utc)
        try:
            self.db.change_events.insert_one(event)
            logger.info("[%s] %s — %s", event["severity"], event["event_type"], event["title"])
            return True
        except Exception as exc:
            # Duplicate key race-condition is fine — just skip
            logger.debug("Skipping duplicate event %s: %s", event["event_id"], exc)
            return False

    def ensure_indexes(self):
        self.db.change_events.create_index("event_id",   unique=True)
        self.db.change_events.create_index("created_at")
        self.db.change_events.create_index("severity")
        self.db.change_events.create_index("event_type")
        logger.info("MongoDB indexes ensured on change_events.")

    def close(self):
        self.client.close()

