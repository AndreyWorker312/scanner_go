#!/usr/bin/env bash
# setup-dns.sh — подготовка хоста для запуска CoreDNS на порту 53
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[✓]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
error() { echo -e "${RED}[✗]${NC} $*"; exit 1; }

# ── 1. Узнаём LAN IP ──────────────────────────────────────────────────────────
DETECTED_IP=$(hostname -I | awk '{print $1}')
warn "Обнаруженный IP хоста: ${DETECTED_IP}"

ENV_FILE="$(dirname "$0")/.env"
if grep -q "^HOST_IP=" "$ENV_FILE" 2>/dev/null; then
    CURRENT=$(grep "^HOST_IP=" "$ENV_FILE" | cut -d= -f2)
    if [[ "$CURRENT" != "$DETECTED_IP" ]]; then
        warn "В .env прописан HOST_IP=${CURRENT}, обновляю на ${DETECTED_IP}"
        sed -i "s/^HOST_IP=.*/HOST_IP=${DETECTED_IP}/" "$ENV_FILE"
    fi
else
    echo "HOST_IP=${DETECTED_IP}" >> "$ENV_FILE"
fi
info "HOST_IP=${DETECTED_IP} записан в .env"

# ── 2. Освобождаем порт 53 от systemd-resolved ────────────────────────────────
if systemctl is-active --quiet systemd-resolved; then
    warn "systemd-resolved занимает порт 53 — отключаю stub listener..."

    CONF_DIR="/etc/systemd/resolved.conf.d"
    sudo mkdir -p "$CONF_DIR"
    sudo tee "$CONF_DIR/no-stub.conf" > /dev/null <<EOF
[Resolve]
DNSStubListener=no
EOF
    sudo systemctl restart systemd-resolved
    info "systemd-resolved stub listener отключён"
else
    info "systemd-resolved не активен, порт 53 свободен"
fi

# ── 3. Проверяем порт 53 ──────────────────────────────────────────────────────
if sudo lsof -i :53 2>/dev/null | grep -q LISTEN; then
    warn "Порт 53 всё ещё занят:"
    sudo lsof -i :53 2>/dev/null
    error "Освободи порт 53 вручную и перезапусти скрипт"
else
    info "Порт 53 свободен"
fi

# ── 4. Итог ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}Готово!${NC} Запусти сервисы:"
echo "  docker compose up -d"
echo ""
echo "Затем на других машинах сети пропиши DNS-сервер: ${DETECTED_IP}"
echo "  └─ или через настройки DHCP роутера → DNS Server → ${DETECTED_IP}"
echo ""
echo "Проверка (с любой машины в сети):"
echo "  nslookup scanner-vniitf.ru ${DETECTED_IP}"
echo "  curl http://scanner-vniitf.ru"

