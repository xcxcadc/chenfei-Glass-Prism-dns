#!/usr/bin/env bash
set -euo pipefail

UPSTREAM_REPO="${PRISM_UPSTREAM_REPO:-mslxi/Liquid-Glass-Prism-dns}"
ENHANCER_REPO="${PRISM_ENHANCER_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"
INSTALL_DIR="${PRISM_INSTALL_DIR:-/opt/prism}"
DATA_DIR="${PRISM_ENHANCER_DATA_DIR:-/var/lib/prism-enhancer}"
PUBLIC_PORT="${PRISM_PORT:-8080}"
CORE_PORT="${PRISM_CORE_PORT:-18080}"
CONTROLLER_SERVICE="prism-controller"
ENHANCER_SERVICE="prism-enhancer"

info() { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
ok() { printf '\033[1;32m[OK]\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

require_root() {
    [ "$(id -u)" -eq 0 ] || fail "请使用 root 权限运行"
    command -v curl >/dev/null 2>&1 || fail "缺少 curl"
    command -v systemctl >/dev/null 2>&1 || fail "当前系统不支持 systemd"
}

prepare_fresh_install() {
    [ "${PRISM_FRESH_INSTALL:-0}" = "1" ] || return 0
    [ "${PRISM_CONFIRM_FRESH:-}" = "YES" ] || fail "PRISM_FRESH_INSTALL requires PRISM_CONFIRM_FRESH=YES"
    case "$INSTALL_DIR" in
        /opt/prism|/opt/prism/*) ;;
        *) fail "全新安装仅允许清理 /opt/prism 下的 Prism 目录" ;;
    esac
    case "$DATA_DIR" in
        /var/lib/prism-enhancer|/var/lib/prism-enhancer/*) ;;
        *) fail "全新安装仅允许清理 /var/lib/prism-enhancer 下的 Prism 目录" ;;
    esac
    info "清理本机已有 Prism 数据，开始全新安装"
    systemctl stop "$ENHANCER_SERVICE" "$CONTROLLER_SERVICE" 2>/dev/null || true
    rm -f "${INSTALL_DIR}/data.db" "${INSTALL_DIR}/.env" "${INSTALL_DIR}/initial-password"
    rm -rf -- "$DATA_DIR"
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) fail "不支持的架构: $(uname -m)" ;;
    esac
}

validate_port() {
    local port="$1"
    [[ "$port" =~ ^[0-9]+$ ]] && [ "$port" -ge 1 ] && [ "$port" -le 65535 ] || fail "无效端口: $port"
}

download() {
    local url="$1" target="$2"
    info "下载 $(basename "$target")"
    curl -fL --retry 3 --connect-timeout 15 -o "${target}.tmp" "$url"
    chmod 0755 "${target}.tmp"
    mv "${target}.tmp" "$target"
}

init_controller() {
    if [ -f "${INSTALL_DIR}/data.db" ]; then
        return
    fi
    info "初始化 Controller 数据库"
    local output password
    output=$(cd "$INSTALL_DIR" && timeout 8 ./prism-controller --host 127.0.0.1 --port 0 2>&1 || true)
    password=$(printf '%s' "$output" | sed -n 's/.*password=\([A-Za-z0-9]*\).*/\1/p' | head -n 1)
    if [ -n "$password" ]; then
        printf '%s' "$password" > "${INSTALL_DIR}/initial-password"
        chmod 0600 "${INSTALL_DIR}/initial-password"
    fi
}

write_services() {
    cat > "/etc/systemd/system/${CONTROLLER_SERVICE}.service" <<EOF
[Unit]
Description=Liquid Glass Prism Controller
After=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/prism-controller --host 127.0.0.1 --port ${CORE_PORT}
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

    cat > "/etc/systemd/system/${ENHANCER_SERVICE}.service" <<EOF
[Unit]
Description=Prism DNS Chinese Enhancer
After=network-online.target ${CONTROLLER_SERVICE}.service
Requires=${CONTROLLER_SERVICE}.service

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/prism-enhancer --listen 0.0.0.0:${PUBLIC_PORT} --upstream http://127.0.0.1:${CORE_PORT} --data-dir ${DATA_DIR}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$CONTROLLER_SERVICE" "$ENHANCER_SERVICE" >/dev/null
}

main() {
    require_root
    prepare_fresh_install
    validate_port "$PUBLIC_PORT"
    validate_port "$CORE_PORT"
    [ "$PUBLIC_PORT" != "$CORE_PORT" ] || fail "公开端口和内部端口不能相同"

    local arch controller_url enhancer_url
    arch=$(detect_arch)
    controller_url="https://github.com/${UPSTREAM_REPO}/releases/latest/download/prism-controller-linux-${arch}"
    enhancer_url="https://github.com/${ENHANCER_REPO}/releases/latest/download/prism-enhancer_linux_${arch}"

    mkdir -p "$INSTALL_DIR" "$DATA_DIR"
    if [ -f "${INSTALL_DIR}/data.db" ]; then
        cp "${INSTALL_DIR}/data.db" "${INSTALL_DIR}/data.db.backup.$(date +%Y%m%d%H%M%S)"
    fi
    download "$controller_url" "${INSTALL_DIR}/prism-controller"
    download "$enhancer_url" "${INSTALL_DIR}/prism-enhancer"

    if [ ! -f "${INSTALL_DIR}/.env" ]; then
        printf 'JWT_SECRET=%s\nGIN_MODE=release\n' "$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')" > "${INSTALL_DIR}/.env"
        chmod 0600 "${INSTALL_DIR}/.env"
    fi

    init_controller
    systemctl stop "$ENHANCER_SERVICE" "$CONTROLLER_SERVICE" 2>/dev/null || true
    write_services
    systemctl restart "$CONTROLLER_SERVICE"
    sleep 2
    systemctl restart "$ENHANCER_SERVICE"
    sleep 2
    systemctl is-active --quiet "$CONTROLLER_SERVICE" || fail "Controller 启动失败，请运行 journalctl -u ${CONTROLLER_SERVICE}"
    systemctl is-active --quiet "$ENHANCER_SERVICE" || fail "中文增强层启动失败，请运行 journalctl -u ${ENHANCER_SERVICE}"

    local ip password
    ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    password=""
    [ -f "${INSTALL_DIR}/initial-password" ] && password=$(cat "${INSTALL_DIR}/initial-password")
    echo
    ok "Prism DNS 中文增强版已启动"
    echo "地址: http://${ip:-YOUR_SERVER_IP}:${PUBLIC_PORT}"
    echo "用户名: admin"
    [ -n "$password" ] && echo "初始密码: $password"
    echo "自定义服务数据: ${DATA_DIR}/custom-services.json"
}

main "$@"
