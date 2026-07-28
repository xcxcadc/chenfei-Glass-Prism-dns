#!/bin/bash

set -e

SCRIPT_REPO="${PRISM_SCRIPT_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"
REPO="${PRISM_AGENT_REPO:-mslxi/Liquid-Glass-Prism-dns}"
BINARY_NAME="prism-agent"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="prism-agent"
SCRIPT_URL="https://raw.githubusercontent.com/${SCRIPT_REPO}/main/agent_install.sh"
TRANSPORT_URL="https://raw.githubusercontent.com/${SCRIPT_REPO}/main/prism_transport.sh"
CUSTOM_IP=""
CONFLICT_BACKUP_DIR="/var/lib/prism-agent/conflict-backups"
WATCHDOG_PATH="/usr/local/lib/prismdns/prism-agent-watchdog.sh"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m'

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
step() { echo -e "${BLUE}[STEP]${NC} $1"; }

check_root() {
    if [ "$EUID" -ne 0 ]; then
        error "Please run as root (sudo)"
    fi
}

parse_args() {
    MASTER_ADDR=""
    SECRET_TOKEN=""
    UNINSTALL_MODE=false
    BETA_MODE=false
    SMART_MODE=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --master)
                if [ -n "$2" ] && [ "${2:0:2}" != "--" ]; then
                    MASTER_ADDR="$2"
                    shift 2
                else
                    error "--master requires a value"
                fi
                ;;
            --secret)
                if [ -n "$2" ] && [ "${2:0:2}" != "--" ]; then
                    SECRET_TOKEN="$2"
                    shift 2
                else
                    error "--secret requires a value"
                fi
                ;;
            --name)
                if [ -n "$2" ] && [ "${2:0:2}" != "--" ]; then
                    SERVICE_NAME="$2"
                    shift 2
                else
                    shift 1
                fi
                ;;
            --ip)
                if [ -n "$2" ] && [ "${2:0:2}" != "--" ]; then
                    CUSTOM_IP="$2"
                    shift 2
                else
                    shift 1
                fi
                ;;
            --uninstall)
                UNINSTALL_MODE=true
                shift
                ;;
            --beta)
                BETA_MODE=true
                shift
                ;;
            --smart)
                SMART_MODE=true
                shift
                ;;
            *)
                shift
                ;;
        esac
    done

    if [ "$UNINSTALL_MODE" = true ]; then
        return
    fi

    if [ -z "$MASTER_ADDR" ] || [ -z "$SECRET_TOKEN" ]; then
        echo -e "${YELLOW}Missing parameters!${NC}"
        echo -e "Usage: ... | bash -s -- --master URL --secret TOKEN [--beta] [--smart]"
        exit 1
    fi
}

uninstall_agent() {
    step "Uninstalling Prism Agent ($SERVICE_NAME)..."

    if [ -x "/usr/local/lib/prismdns/prism_transport.sh" ]; then
        /usr/local/lib/prismdns/prism_transport.sh --uninstall || true
    fi
    
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable --now prism-agent-watchdog.timer prism-agent-watchdog.service 2>/dev/null || true
    
    if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
        rm "/etc/systemd/system/${SERVICE_NAME}.service"
        systemctl daemon-reload
    fi
    
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        rm "$INSTALL_DIR/$BINARY_NAME"
    fi
    rm -f /etc/systemd/system/prism-agent-watchdog.service
    rm -f /etc/systemd/system/prism-agent-watchdog.timer
    rm -f "$WATCHDOG_PATH"
    systemctl daemon-reload
    
    info "Uninstallation completed."
    exit 0
}

detect_system() {
    ARCH=$(uname -m)
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')

    case "$ARCH" in
        x86_64) ARCH_SUFFIX="amd64" ;;
        aarch64|arm64) ARCH_SUFFIX="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    ASSET_NAME="${BINARY_NAME}_${OS}_${ARCH_SUFFIX}"
    info "Detected: ${OS} / ${ARCH_SUFFIX}"
}

download_binary() {
    step "Fetching version info..."

    API_URL="https://api.github.com/repos/$REPO/releases"
    
    if [ "$BETA_MODE" = true ]; then
        info "Mode: ${YELLOW}Beta Channel (Pre-release)${NC}"
    else
        info "Mode: ${GREEN}Stable Channel (Official)${NC}"
    fi
    
    RESP=$(curl -s --connect-timeout 10 "$API_URL")

    if [ "$BETA_MODE" = true ]; then
        # Beta: find the beta release with largest timestamp (format: beta-YYYYMMDDHHMMSS)
        DOWNLOAD_URL=$(echo "$RESP" | awk -v asset="$ASSET_NAME" '
            BEGIN { latest_ts = ""; latest_url = "" }
            /"tag_name":/ { 
                tag = $0
                gsub(/.*"tag_name": *"|".*/, "", tag)
                current_tag = tag
            }
            /"browser_download_url":/ && index($0, asset) {
                url = $0
                gsub(/.*"browser_download_url": *"|".*/, "", url)
                if (index(current_tag, "beta-") == 1) {
                    ts = current_tag
                    gsub(/^beta-/, "", ts)
                    if (ts > latest_ts) {
                        latest_ts = ts
                        latest_url = url
                    }
                }
            }
            END { print latest_url }
        ')
        VERSION=$(echo "$DOWNLOAD_URL" | grep -oE 'beta-[0-9]+' | head -1)
    else
        # Stable: find first non-prerelease with agent asset
        DOWNLOAD_URL=$(echo "$RESP" | grep -E '"tag_name"|"prerelease"|"browser_download_url".*prism-agent' | \
            awk -v asset="$ASSET_NAME" '
                /"tag_name":/ { tag=$0; gsub(/.*"tag_name": *"|".*/, "", tag) }
                /"prerelease":/ { prerelease=$0; gsub(/.*"prerelease": *|,.*/, "", prerelease) }
                /"browser_download_url":/ && index($0, asset) { 
                    url=$0; gsub(/.*"browser_download_url": *"|".*/, "", url)
                    if (prerelease == "false") { print url; exit }
                }
            ')
        VERSION=$(echo "$DOWNLOAD_URL" | grep -oE 'v[0-9]+\.[0-9]+[^/]*' | head -1)
    fi

    if [ -z "$DOWNLOAD_URL" ]; then
        warn "Smart search failed, trying fallback..."
        DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"
    fi

    if [ -n "$VERSION" ]; then
        info "Found agent version: ${CYAN}${VERSION}${NC}"
    fi

    info "Download URL: $DOWNLOAD_URL"
    curl -L -o "/tmp/$BINARY_NAME" "$DOWNLOAD_URL" --progress-bar

    if [ ! -f "/tmp/$BINARY_NAME" ] || [ ! -s "/tmp/$BINARY_NAME" ]; then
        error "Download failed. Please check network or GitHub access."
    fi

    chmod +x "/tmp/$BINARY_NAME"
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        info "Stopping old service..."
        systemctl stop "$SERVICE_NAME"
    fi

    mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
}

configure_service() {
    step "Configuring systemd service..."
    SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

    EXEC_ARGS="--master \"$MASTER_ADDR\" --secret \"$SECRET_TOKEN\""
    
    if [ "$SMART_MODE" = true ]; then
        EXEC_ARGS="$EXEC_ARGS --smart"
    fi

    if [ -n "$CUSTOM_IP" ]; then
        EXEC_ARGS="$EXEC_ARGS --ip \"$CUSTOM_IP\""
    fi

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Liquid Glass Prism Agent ($SERVICE_NAME)
After=network.target

[Service]
Type=simple
User=root
Restart=always
RestartSec=5s
ExecStart=$INSTALL_DIR/$BINARY_NAME $EXEC_ARGS
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME" >/dev/null 2>&1
}

backup_and_stop_conflicting_service() {
    local unit="$1"
    if ! systemctl is-active --quiet "$unit" && ! systemctl is-enabled --quiet "$unit" 2>/dev/null; then
        return
    fi

    local stamp backup
    stamp=$(date +%Y%m%d-%H%M%S)
    backup="$CONFLICT_BACKUP_DIR/$stamp-$unit"
    mkdir -p "$backup"
    systemctl is-active --quiet "$unit" && echo active > "$backup/was-active" || true
    systemctl is-enabled --quiet "$unit" 2>/dev/null && echo enabled > "$backup/was-enabled" || true
    systemctl cat "$unit" > "$backup/unit.txt" 2>/dev/null || true
    case "$unit" in
        dnsmasq)
            cp -a /etc/dnsmasq.conf /etc/dnsmasq.d "$backup/" 2>/dev/null || true
            ;;
        sniproxy)
            cp -a /etc/sniproxy.conf /etc/sniproxy "$backup/" 2>/dev/null || true
            ;;
    esac
    warn "Stopping legacy $unit because it conflicts with Prism Agent. Backup: $backup"
    systemctl disable --now "$unit" >/dev/null 2>&1 || true
}

prepare_legacy_conflicts() {
    step "Checking DNS/proxy port conflicts..."
    backup_and_stop_conflicting_service dnsmasq
    backup_and_stop_conflicting_service sniproxy
}

listener_owned_by_agent() {
    local port="$1"
    ss -lntup 2>/dev/null | awk -v port=":$port" '$5 ~ port"$" && /prism-agent/ {found=1} END {exit !found}'
}

start_service() {
    step "Starting service..."
    AGENT_STARTED_AT=$(date '+%Y-%m-%d %H:%M:%S')
    systemctl restart "$SERVICE_NAME"
    
    info "Waiting for initialization..."
    for _ in {1..30}; do
        if ! systemctl is-active --quiet "$SERVICE_NAME"; then
            error "Failed to start! Check logs: journalctl -u $SERVICE_NAME -n 20"
        fi
        LOGS=$(journalctl -u "$SERVICE_NAME" --since "$AGENT_STARTED_AT" --no-pager 2>/dev/null || true)
        if echo "$LOGS" | grep -q "DNS Server Started" && listener_owned_by_agent 53; then
            AGENT_MODE="dns"
            return 0
        fi
        if echo "$LOGS" | grep -q "Proxy Server Started" && listener_owned_by_agent 80 && listener_owned_by_agent 443; then
            AGENT_MODE="proxy"
            return 0
        fi
        if echo "$LOGS" | grep -Eq "port (53|80|443) busy"; then
            echo "$LOGS" | tail -20 >&2
            error "A required port is still occupied. Prism Agent did not enter the data path."
        fi
        sleep 1
    done
    journalctl -u "$SERVICE_NAME" --since "$AGENT_STARTED_AT" -n 30 --no-pager >&2 2>/dev/null || true
    error "Prism Agent is online but its DNS/proxy listener was not ready within 30 seconds."
}

install_proxy_transport() {
    if [ "${AGENT_MODE:-}" != "proxy" ]; then
        return
    fi
    step "Installing encrypted SNI transport..."
    local installer="/tmp/prism_transport.sh"
    if [ -n "${PRISM_TRANSPORT_INSTALLER_FILE:-}" ]; then
        installer="$PRISM_TRANSPORT_INSTALLER_FILE"
    else
        curl -fsSL "$TRANSPORT_URL" -o "$installer"
    fi
    chmod +x "$installer"
    bash "$installer" --proxy --master "$MASTER_ADDR" --secret "$SECRET_TOKEN"
}

install_proxy_watchdog() {
    if [ "${AGENT_MODE:-}" != "proxy" ]; then
        systemctl disable --now prism-agent-watchdog.timer prism-agent-watchdog.service >/dev/null 2>&1 || true
        rm -f /etc/systemd/system/prism-agent-watchdog.service
        rm -f /etc/systemd/system/prism-agent-watchdog.timer
        rm -f "$WATCHDOG_PATH"
        systemctl daemon-reload
        return
    fi

    step "Installing Prism Agent resource watchdog..."
    mkdir -p "$(dirname "$WATCHDOG_PATH")" /var/lib/prism-agent
    cat > "$WATCHDOG_PATH" <<'EOF'
#!/bin/bash

set -u

SERVICE="prism-agent"
STATE_FILE="/var/lib/prism-agent/watchdog-failures"
MAX_FDS=4096
MAX_CLOSE_WAIT=1024

restart_agent() {
    local reason="$1"
    logger -t prism-agent-watchdog "restarting prism-agent: $reason"
    printf '0\n' > "$STATE_FILE"
    systemctl restart "$SERVICE"
}

systemctl is-active --quiet "$SERVICE" || exit 0
pid=$(systemctl show -p MainPID --value "$SERVICE" 2>/dev/null)
[[ "$pid" =~ ^[1-9][0-9]*$ && -d "/proc/$pid" ]] || exit 0

fds=$(find "/proc/$pid/fd" -maxdepth 1 -type l 2>/dev/null | wc -l)
close_wait=$(ss -tanp state close-wait 2>/dev/null | grep -c "pid=$pid," || true)
if ((fds >= MAX_FDS || close_wait >= MAX_CLOSE_WAIT)); then
    restart_agent "resource leak detected (fds=$fds close_wait=$close_wait)"
    exit 0
fi

healthy=false
for domain in www.cloudflare.com www.apple.com; do
    if curl -4 -ksS -o /dev/null \
        --connect-timeout 4 --max-time 8 \
        --resolve "$domain:443:127.0.0.1" "https://$domain/"; then
        healthy=true
        break
    fi
done

if $healthy; then
    printf '0\n' > "$STATE_FILE"
    exit 0
fi

failures=0
[[ -f "$STATE_FILE" ]] && read -r failures < "$STATE_FILE"
[[ "$failures" =~ ^[0-9]+$ ]] || failures=0
failures=$((failures + 1))
printf '%s\n' "$failures" > "$STATE_FILE"
if ((failures >= 3)); then
    restart_agent "three consecutive local TLS probes failed"
fi
EOF
    chmod 755 "$WATCHDOG_PATH"

    cat > /etc/systemd/system/prism-agent-watchdog.service <<EOF
[Unit]
Description=Prism Agent resource and TLS watchdog
After=network-online.target $SERVICE_NAME.service

[Service]
Type=oneshot
ExecStart=$WATCHDOG_PATH
EOF

    cat > /etc/systemd/system/prism-agent-watchdog.timer <<'EOF'
[Unit]
Description=Check Prism Agent resource and TLS health

[Timer]
OnBootSec=2min
OnUnitActiveSec=1min
AccuracySec=10s
Persistent=true

[Install]
WantedBy=timers.target
EOF

    systemctl daemon-reload
    systemctl enable --now prism-agent-watchdog.timer >/dev/null
}

show_result() {
    LOGS=$(journalctl -u "$SERVICE_NAME" --since "${AGENT_STARTED_AT:-5 minutes ago}" --no-pager 2>/dev/null || true)
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════${NC}"
    echo -e "${GREEN}   Liquid Glass Prism Agent Installed!         ${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════${NC}"
    echo ""
    
    if [ "$BETA_MODE" = true ]; then
        echo -e "  Version: ${YELLOW}Beta (Pre-release)${NC}"
    fi
    
    if [ "$SMART_MODE" = true ]; then
        echo -e "  Feature: ${CYAN}Smart Mode Enabled${NC}"
    fi

    if [ "${AGENT_MODE:-}" = "dns" ]; then
        echo -e "  Mode:    ${CYAN}DNS Client${NC} (Set DNS to 127.0.0.1)"
    elif [ "${AGENT_MODE:-}" = "proxy" ]; then
        echo -e "  Mode:    ${CYAN}Proxy Agent${NC} (Open ports 80/443)"
    else
        error "Agent data listener verification failed."
    fi
    
    echo ""
    echo -e "  Uninstall: ${GREEN}curl -sL $SCRIPT_URL | bash -s -- --uninstall${NC}"
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════${NC}"
}

show_banner() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║       Liquid Glass Prism Agent Installer         ║${NC}"
    echo -e "${BLUE}║      github.com/xcxcadc/chenfei-Glass-Prism-dns  ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
    echo ""
}

main() {
    show_banner
    check_root
    parse_args "$@"
    
    if [ "$UNINSTALL_MODE" = true ]; then
        uninstall_agent
    fi

    detect_system
    download_binary
    configure_service
    prepare_legacy_conflicts
    start_service
    install_proxy_watchdog
    install_proxy_transport
    show_result
}

main "$@"
