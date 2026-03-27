"""
test_inject.py — скрипт для тестирования change_detector.

Вбрасывает пары фейковых сканов в MongoDB так, чтобы детектор
нашёл каждый тип события:

  CRITICAL  — новый порт 4444  (backdoor)
  HIGH      — новый порт 9090  (обычный)
  HIGH      — закрылся порт 22 (SSH — критический сервис)
  HIGH      — новое устройство в сети  (ARP)
  MEDIUM    — сменился сервис на порту 8080
  MEDIUM    — устройство пропало из сети (ARP)
  LOW       — закрылся порт 3000 (не критичный)

Запуск (локально, пока стек поднят):
    python3 test_inject.py

Или прямо внутри контейнера:
    docker compose exec change-detector python test_inject.py
"""

import os
import time
from datetime import datetime, timezone, timedelta
from pymongo import MongoClient

MONGO_URI = os.getenv("MONGODB_URI", "mongodb://localhost:27017")
DB_NAME   = os.getenv("MONGODB_DATABASE", "network_scanner")

client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=5000)
db     = client[DB_NAME]

# ────────────────────────────────────────────────────────────────────────────
# Вспомогательные функции
# ────────────────────────────────────────────────────────────────────────────

def nmap_record(ip, ports: list[dict], minutes_ago: int = 0) -> dict:
    """
    Собирает запись в формате NmapTcpUdpHistoryRecord.

    ports — список словарей:
        {"port": 22, "service": "ssh", "protocol": "tcp", "state": "open"}
    """
    all_ports    = [p["port"]     for p in ports]
    protocols    = [p.get("protocol", "tcp") for p in ports]
    states       = [p.get("state",    "open") for p in ports]
    service_name = [p.get("service",  "unknown") for p in ports]

    return {
        "task_id":      f"test-nmap-{ip}-{minutes_ago}",
        "ip":           ip,
        "scanner_type": "tcp",
        "ports":        "1-65535",
        "host":         ip,
        "status":       "done",
        "port_info": [
            {
                "status":       "done",
                "all_ports":    all_ports,
                "protocols":    protocols,
                "state":        states,
                "service_name": service_name,
            }
        ],
        "created_at": datetime.now(timezone.utc) - timedelta(minutes=minutes_ago),
    }


def arp_record(ip_range, devices: list[dict], minutes_ago: int = 0) -> dict:
    """
    Собирает запись в формате ARPHistoryRecord.
    """
    online = [d for d in devices if d.get("status", "online") == "online"]
    return {
        "task_id":        f"test-arp-{ip_range}-{minutes_ago}",
        "interface_name": "eth0",
        "ip_range":       ip_range,
        "status":         "done",
        "devices":        devices,
        "online_devices": online,
        "offline_devices": [],
        "total_count":    len(devices),
        "online_count":   len(online),
        "offline_count":  0,
        "created_at": datetime.now(timezone.utc) - timedelta(minutes=minutes_ago),
    }


def insert_pair(collection, old_doc, new_doc, label: str):
    """Вставляет старый скан (раньше), потом новый (сейчас)."""
    db[collection].insert_one(old_doc)
    db[collection].insert_one(new_doc)
    print(f"  ✅  [{label}] вставлено 2 записи в '{collection}'")


# ────────────────────────────────────────────────────────────────────────────
# Чистим старые тестовые данные + события от прошлых запусков
# ────────────────────────────────────────────────────────────────────────────

print("\n🗑  Удаляю старые тестовые записи...")
db.nmap_tcp_udp_history.delete_many({"task_id": {"$regex": "^test-"}})
db.arp_history.delete_many(          {"task_id": {"$regex": "^test-"}})
db.change_events.delete_many(        {"event_id": {"$regex": "^(NEW_PORT|PORT_CLOSED|VERSION_CHANGE|NEW_DEVICE|DEVICE_GONE):.*"}})

print("\n📥  Вбрасываю тестовые пары сканов...\n")

# ──────────────────────────────────────────────────
# NMAP — IP 192.168.1.100
# ──────────────────────────────────────────────────
BASE_PORTS = [
    {"port": 22,   "service": "ssh",   "protocol": "tcp", "state": "open"},
    {"port": 80,   "service": "http",  "protocol": "tcp", "state": "open"},
    {"port": 443,  "service": "https", "protocol": "tcp", "state": "open"},
    {"port": 3000, "service": "node",  "protocol": "tcp", "state": "open"},
    {"port": 8080, "service": "nginx", "protocol": "tcp", "state": "open"},
]

NEW_PORTS = [
    # порт 22 (SSH) — ПРОПАЛ   → HIGH
    # порт 3000 (node) — ПРОПАЛ → LOW
    # порт 8080 — сервис nginx→apache → MEDIUM
    {"port": 80,   "service": "http",        "protocol": "tcp", "state": "open"},
    {"port": 443,  "service": "https",       "protocol": "tcp", "state": "open"},
    {"port": 8080, "service": "apache",      "protocol": "tcp", "state": "open"},  # сменился сервис
    {"port": 4444, "service": "unknown",     "protocol": "tcp", "state": "open"},  # CRITICAL — backdoor
    {"port": 9090, "service": "prometheus",  "protocol": "tcp", "state": "open"},  # HIGH — новый порт
]

insert_pair(
    "nmap_tcp_udp_history",
    old_doc=nmap_record("192.168.1.100", BASE_PORTS, minutes_ago=10),
    new_doc=nmap_record("192.168.1.100", NEW_PORTS,  minutes_ago=0),
    label="CRITICAL+HIGH+HIGH+MEDIUM+LOW",
)

# ──────────────────────────────────────────────────
# ARP — подсеть 192.168.1.0/24
# ──────────────────────────────────────────────────
OLD_DEVICES = [
    {"ip": "192.168.1.1",   "mac": "aa:bb:cc:dd:ee:01", "vendor": "Cisco",  "status": "online"},
    {"ip": "192.168.1.50",  "mac": "aa:bb:cc:dd:ee:02", "vendor": "Dell",   "status": "online"},
    {"ip": "192.168.1.99",  "mac": "aa:bb:cc:dd:ee:03", "vendor": "HP",     "status": "online"},
]
NEW_DEVICES = [
    {"ip": "192.168.1.1",   "mac": "aa:bb:cc:dd:ee:01", "vendor": "Cisco",    "status": "online"},
    # 192.168.1.50 — ПРОПАЛО → MEDIUM
    {"ip": "192.168.1.99",  "mac": "aa:bb:cc:dd:ee:03", "vendor": "HP",       "status": "online"},
    {"ip": "192.168.1.200", "mac": "ff:ee:dd:cc:bb:aa", "vendor": "unknown",  "status": "online"},  # HIGH — новое устройство
]

insert_pair(
    "arp_history",
    old_doc=arp_record("192.168.1.0/24", OLD_DEVICES, minutes_ago=10),
    new_doc=arp_record("192.168.1.0/24", NEW_DEVICES, minutes_ago=0),
    label="HIGH+MEDIUM (ARP)",
)

# ────────────────────────────────────────────────────────────────────────────
# Итог
# ────────────────────────────────────────────────────────────────────────────
print("""
┌─────────────────────────────────────────────────────────────────────┐
│  Ожидаемые события после следующего цикла детектора (≤30 сек):      │
├──────────┬──────────────────┬──────────────────────────────────────┤
│ CRITICAL │ NEW_PORT         │ Порт 4444 открыт на 192.168.1.100    │
│ HIGH     │ NEW_PORT         │ Порт 9090 открыт на 192.168.1.100    │
│ HIGH     │ PORT_CLOSED      │ Порт 22 (SSH) закрыт на 192.168.1.100│
│ HIGH     │ NEW_DEVICE       │ 192.168.1.200 появился в сети         │
│ MEDIUM   │ VERSION_CHANGE   │ 8080: nginx → apache                  │
│ MEDIUM   │ DEVICE_GONE      │ 192.168.1.50 пропал из сети           │
│ LOW      │ PORT_CLOSED      │ Порт 3000 (node) закрыт               │
└──────────┴──────────────────┴──────────────────────────────────────┘
""")

print("⏳  Жду запуска цикла детектора (до 35 сек)...")
time.sleep(35)

print("\n📊  Проверяю результаты в MongoDB...\n")
events = list(db.change_events.find(
    {},
    {"_id": 0, "event_id": 0, "created_at": 0, "details": 0},
    sort=[("created_at", -1)],
    limit=20,
))

if not events:
    print("  ⚠️  Событий нет. Возможно детектор ещё не запущен или не подключился к MongoDB.")
    print("  Проверь: docker compose logs change-detector --tail=30")
else:
    print(f"  Найдено {len(events)} событий:\n")
    for e in events:
        icon = {"CRITICAL": "🔴", "HIGH": "🟠", "MEDIUM": "🟡", "LOW": "🟢"}.get(e.get("severity", ""), "⚪")
        print(f"  {icon} [{e.get('severity','?')}] {e.get('event_type','?')} — {e.get('title','?')}")

client.close()

