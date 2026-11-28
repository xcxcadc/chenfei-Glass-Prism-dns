#!/usr/bin/env bash

set +x
set -Eeuo pipefail

VERSION="2.7.0"
SCRIPT_REF="${PRISM_SCRIPT_REF:-v2.0.1}"
INSTALL_DIR="/usr/local/lib/prismdns"
INSTALL_PATH="$INSTALL_DIR/prism_transport.sh"
STATE_DIR="/var/lib/prism-transport"
ENV_FILE="/etc/prismdns/transport.env"
SSH_KEY_FILE="$STATE_DIR/ssh_key"
READY_FILE="$STATE_DIR/ready-proxies.json"
ACTIVE_FILE="$STATE_DIR/active-proxies.json"
EGRESS_BINARY="/usr/local/bin/prism-egress-proxy"
EGRESS_SERVICE="prism-egress-proxy.service"
EGRESS_POLICY="/etc/prismdns/egress-policy.json"
EGRESS_HTTP_PORT=19080
EGRESS_HTTPS_PORT=19443
CLIENT_EGRESS_SERVICE="prism-client-egress-proxy.service"
CLIENT_EGRESS_POLICY="/etc/prismdns/client-egress-policy.json"
CLIENT_EGRESS_HTTP_PORT=19081
CLIENT_EGRESS_HTTPS_PORT=19444
EGRESS_REPO="${PRISM_EGRESS_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"
NODE_UPDATE_UNIT_DIR='/etc/systemd/system'
NODE_UPDATE_RUNTIME_DIR='/run'
MEDIA_WORKER_PATH='/usr/local/bin/prism-media-check'
ROLE=""
MASTER=""
MASTER_OVERRIDE=""
CREDENTIAL=""
ACTION="install"
TRANSPORT_MODE="${PRISM_TRANSPORT_MODE:-direct}"

log() { printf '[Prism Transport] %s\n' "$*"; }
fail() { printf '[Prism Transport] ERROR: %s\n' "$*" >&2; exit 1; }

require_root() {
  [[ ${EUID:-$(id -u)} -eq 0 ]] || fail "root privileges are required"
}

parse_args() {
  while (($#)); do
    case "$1" in
      --proxy) ROLE="proxy"; shift ;;
      --client) ROLE="client"; shift ;;
      --master) MASTER="${2:-}"; MASTER_OVERRIDE="$MASTER"; shift 2 ;;
      --secret|--token) CREDENTIAL="${2:-}"; shift 2 ;;
      --mode) TRANSPORT_MODE="${2:-}"; shift 2 ;;
      --sync) ACTION="sync"; shift ;;
      --refresh-runtime) ACTION="refresh-runtime"; shift ;;
      --uninstall) ACTION="uninstall"; shift ;;
      --status) ACTION="status"; shift ;;
      --help|-h)
        echo "Usage: prism_transport.sh (--proxy --secret SECRET | --client --token TOKEN) --master URL [--mode direct|encrypted] [--sync|--refresh-runtime|--uninstall|--status]"
        exit 0
        ;;
      *) fail "unknown argument: $1" ;;
    esac
  done
}

load_environment() {
  if [[ -f "$ENV_FILE" ]]; then
    # shellcheck disable=SC1090
    source "$ENV_FILE"
  fi
  [[ -z "$MASTER_OVERRIDE" ]] || MASTER="$MASTER_OVERRIDE"
  MASTER="${MASTER%/}"
  [[ "$TRANSPORT_MODE" == "direct" || "$TRANSPORT_MODE" == "encrypted" ]] || fail "invalid transport mode: $TRANSPORT_MODE"
}

detect_package_manager() {
  if command -v apt-get >/dev/null 2>&1; then echo apt
  elif command -v dnf >/dev/null 2>&1; then echo dnf
  elif command -v yum >/dev/null 2>&1; then echo yum
  elif command -v apk >/dev/null 2>&1; then echo apk
  else echo unknown
  fi
}

ensure_dependencies() {
  local ready=true
  for command_name in curl jq ssh ssh-keygen ip nft flock setsid; do
    command -v "$command_name" >/dev/null 2>&1 || ready=false
  done
  if [[ "$ROLE" == "proxy" ]]; then
    command -v sshd >/dev/null 2>&1 || ready=false
  fi
  $ready && return 0
  [[ -z "${PRISM_NODE_UPDATE_EXPECTED_VERSION:-}" ]] || fail "managed refresh requires existing transport dependencies"

  local manager
  manager=$(detect_package_manager)
  log "installing encrypted TCP transport dependencies"
  case "$manager" in
    apt)
      apt-get update -qq
      if [[ "$ROLE" == "proxy" ]]; then
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl jq openssh-client openssh-server iproute2 nftables util-linux
      else
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl jq openssh-client iproute2 nftables util-linux
      fi
      ;;
    dnf|yum)
      "$manager" install -y curl jq openssh-clients iproute nftables util-linux
      [[ "$ROLE" == "proxy" ]] && "$manager" install -y openssh-server
      ;;
    apk)
      apk add --no-cache curl jq openssh-client iproute2 nftables util-linux
      [[ "$ROLE" == "proxy" ]] && apk add --no-cache openssh-server
      ;;
    *) fail "unsupported package manager; install curl, jq, OpenSSH, iproute2 and util-linux first" ;;
  esac
}

save_environment() {
  mkdir -p "$(dirname "$ENV_FILE")"
  umask 077
  {
    printf 'ROLE=%q\n' "$ROLE"
    printf 'MASTER=%q\n' "$MASTER"
    printf 'CREDENTIAL=%q\n' "$CREDENTIAL"
    printf 'TRANSPORT_MODE=%q\n' "$TRANSPORT_MODE"
  } >"$ENV_FILE"
  chmod 600 "$ENV_FILE"
}

ensure_client_keypair() {
  mkdir -p "$STATE_DIR"
  chmod 700 "$STATE_DIR"
  if [[ ! -s "$SSH_KEY_FILE" ]]; then
    ssh-keygen -q -t ed25519 -N '' -f "$SSH_KEY_FILE"
  fi
  chmod 600 "$SSH_KEY_FILE"
  chmod 644 "$SSH_KEY_FILE.pub"
}

detect_architecture() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

download_dist_or_github() {
  local filename="$1" github_url="$2" target="$3"
  if [[ "${MASTER:-}" =~ ^https?://[^[:space:]@]+$ ]]; then
    if curl -fsSL --connect-timeout 6 --max-time 90 "$MASTER/enhancer/dist/$filename" -o "$target" 2>/dev/null && [[ -s "$target" ]]; then
      return 0
    fi
    rm -f -- "$target"
  fi
  curl -fL --retry 3 --connect-timeout 15 --max-time 120 -o "$target" "$github_url"
}

install_proxy_egress() {
  [[ "$ROLE" == "proxy" ]] || return 0
  local source="${PRISM_EGRESS_BINARY_FILE:-}" temporary architecture
  if [[ -n "$source" ]]; then
    [[ -s "$source" ]] || fail "PRISM_EGRESS_BINARY_FILE does not exist: $source"
  else
    architecture=$(detect_architecture)
    temporary=$(mktemp)
    download_dist_or_github \
      "prism-egress-proxy_linux_$architecture" \
      "https://github.com/$EGRESS_REPO/releases/download/$SCRIPT_REF/prism-egress-proxy_linux_$architecture" \
      "$temporary"
    source="$temporary"
  fi
  install -m 755 "$source" "$EGRESS_BINARY"
  [[ -z "${temporary:-}" ]] || rm -f "$temporary"
  mkdir -p "$(dirname "$EGRESS_POLICY")" "$STATE_DIR"
  if [[ ! -s "$EGRESS_POLICY" ]]; then
    printf '{"services":[]}\n' >"$EGRESS_POLICY"
    chmod 644 "$EGRESS_POLICY"
  fi
  cat >/etc/systemd/system/$EGRESS_SERVICE <<EOF
[Unit]
Description=Prism adaptive IPv4/IPv6 SNI egress
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$EGRESS_BINARY --mode proxy --warp-socks 127.0.0.1:40000 --http-listen 0.0.0.0:$EGRESS_HTTP_PORT --tls-listen 0.0.0.0:$EGRESS_HTTPS_PORT --policy $EGRESS_POLICY
Restart=always
RestartSec=2
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadOnlyPaths=$EGRESS_POLICY
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable "$EGRESS_SERVICE" >/dev/null
  systemctl restart "$EGRESS_SERVICE"
  for _ in {1..30}; do
    if systemctl is-active --quiet "$EGRESS_SERVICE" &&
      ss -lnt 2>/dev/null | awk -v http=":$EGRESS_HTTP_PORT" -v https=":$EGRESS_HTTPS_PORT" \
        '$4 ~ http "$" {h=1} $4 ~ https "$" {s=1} END {exit !(h && s)}'; then
      printf '%s\n' "$VERSION" >"$STATE_DIR/egress-version"
      return 0
    fi
    sleep 0.2
  done
  fail "adaptive egress did not open its listeners"
}

install_client_egress() {
  [[ "$ROLE" == "client" ]] || return 0
  local source="${PRISM_EGRESS_BINARY_FILE:-}" temporary architecture
  if [[ -n "$source" ]]; then
    [[ -s "$source" ]] || fail "PRISM_EGRESS_BINARY_FILE does not exist: $source"
  else
    architecture=$(detect_architecture)
    temporary=$(mktemp)
    download_dist_or_github \
      "prism-egress-proxy_linux_$architecture" \
      "https://github.com/$EGRESS_REPO/releases/download/$SCRIPT_REF/prism-egress-proxy_linux_$architecture" \
      "$temporary"
    source="$temporary"
  fi
  install -m 755 "$source" "$EGRESS_BINARY"
  [[ -z "${temporary:-}" ]] || rm -f "$temporary"
  mkdir -p "$(dirname "$CLIENT_EGRESS_POLICY")" "$STATE_DIR"
  if [[ ! -s "$CLIENT_EGRESS_POLICY" ]]; then
    printf '{"services":[]}\n' >"$CLIENT_EGRESS_POLICY"
    chmod 644 "$CLIENT_EGRESS_POLICY"
  fi
  cat >/etc/systemd/system/$CLIENT_EGRESS_SERVICE <<EOF
[Unit]
Description=Prism transparent client SNI egress
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$EGRESS_BINARY --mode client --http-listen 127.0.0.1:$CLIENT_EGRESS_HTTP_PORT --tls-listen 127.0.0.1:$CLIENT_EGRESS_HTTPS_PORT --policy $CLIENT_EGRESS_POLICY
Restart=always
RestartSec=2s
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadOnlyPaths=$CLIENT_EGRESS_POLICY
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable "$CLIENT_EGRESS_SERVICE" >/dev/null
  systemctl restart "$CLIENT_EGRESS_SERVICE"
  for _ in {1..30}; do
    if systemctl is-active --quiet "$CLIENT_EGRESS_SERVICE" &&
      ss -lnt 2>/dev/null | awk -v http=":$CLIENT_EGRESS_HTTP_PORT" -v https=":$CLIENT_EGRESS_HTTPS_PORT" \
        '$4 ~ http "$" {h=1} $4 ~ https "$" {s=1} END {exit !(h && s)}'; then
      printf '%s\n' "$VERSION" >"$STATE_DIR/client-egress-version"
      return 0
    fi
    sleep 0.2
  done
  fail "transparent client egress did not open its loopback listeners"
}

client_egress_ready() {
  [[ "$ROLE" == "client" ]] || return 0
  [[ -x "$EGRESS_BINARY" ]] &&
    [[ -f "$STATE_DIR/client-egress-version" && "$(<"$STATE_DIR/client-egress-version")" == "$VERSION" ]] &&
    grep -Fq -- '--mode client' "/etc/systemd/system/$CLIENT_EGRESS_SERVICE" 2>/dev/null &&
    systemctl is-active --quiet "$CLIENT_EGRESS_SERVICE" 2>/dev/null &&
    ss -lnt 2>/dev/null | awk -v http=":$CLIENT_EGRESS_HTTP_PORT" -v https=":$CLIENT_EGRESS_HTTPS_PORT" \
      '$4 ~ http "$" {h=1} $4 ~ https "$" {s=1} END {exit !(h && s)}'
}

proxy_egress_ready() {
  [[ "$ROLE" == "proxy" ]] || return 0
  [[ -x "$EGRESS_BINARY" ]] &&
    [[ -f "$STATE_DIR/egress-version" && "$(<"$STATE_DIR/egress-version")" == "$VERSION" ]] &&
    grep -Fq -- '--mode proxy' "/etc/systemd/system/$EGRESS_SERVICE" 2>/dev/null &&
    systemctl is-active --quiet "$EGRESS_SERVICE" 2>/dev/null &&
    ss -lnt 2>/dev/null | awk -v http=":$EGRESS_HTTP_PORT" -v https=":$EGRESS_HTTPS_PORT" \
      '$4 ~ http "$" {h=1} $4 ~ https "$" {s=1} END {exit !(h && s)}'
}

agent_master_endpoint() {
  local pid argument previous=""
  pid=$(systemctl show prism-agent.service -p MainPID --value 2>/dev/null) || return 1
  [[ "$pid" =~ ^[1-9][0-9]*$ && -r "/proc/$pid/cmdline" ]] || return 1
  while IFS= read -r -d '' argument; do
    if [[ "$previous" == "--master" ]]; then
      printf '%s\n' "${argument%/}"
      return 0
    fi
    case "$argument" in
      --master=*) printf '%s\n' "${argument#--master=}"; return 0 ;;
    esac
    previous="$argument"
  done <"/proc/$pid/cmdline"
  return 1
}

transport_response_valid() {
  jq -e --arg role "$ROLE" '.role == $role and (.peers | type == "array")' >/dev/null
}

post_transport_json() {
  local master="$1" endpoint="$2" body="$3"
  curl -fsS --connect-timeout 8 --max-time 20 \
    -H 'Content-Type: application/json' \
    --data-binary @- "$master$endpoint" <<<"$body"
}

post_json() {
  local endpoint="$1" body="$2" response candidate
  if response=$(post_transport_json "$MASTER" "$endpoint" "$body") &&
    transport_response_valid <<<"$response"; then
    [[ -z "$MASTER_OVERRIDE" ]] || save_environment
    printf '%s\n' "$response"
    return 0
  fi
  candidate=$(agent_master_endpoint) || return 1
  candidate="${candidate%/}"
  [[ "$candidate" =~ ^https?://[^[:space:]@]+$ && "$candidate" != "$MASTER" ]] || return 1
  if response=$(post_transport_json "$candidate" "$endpoint" "$body") &&
    transport_response_valid <<<"$response"; then
    MASTER="$candidate"
    save_environment
    log "transport controller recovered from the active Agent configuration" >&2
    printf '%s\n' "$response"
    return 0
  fi
  return 1
}

proxy_warp_listening() {
  ss -lnt 2>/dev/null | awk '$4 ~ /(127\.0\.0\.1|\[::1\]):40000$/ {found=1} END {exit !found}'
}

proxy_warp_ready() {
  proxy_warp_listening || return 1
  if [[ -f /run/prism-warp-liveness.ok && "$(</run/prism-warp-liveness.ok)" == "0" ]]; then
    return 1
  fi
  return 0
}

proxy_warp_provisioning() {
  systemctl is-active --quiet prism-warp-bootstrap.service 2>/dev/null
}

check_and_heal_warp_liveness() {
  local now last_check=0 ts_file="/run/prism-warp-liveness.ts" ok_file="/run/prism-warp-liveness.ok"
  local conf="/etc/prismdns/warp/wireproxy.conf"
  now=$(date +%s)
  [[ -f "$ts_file" ]] && last_check=$(cat "$ts_file" 2>/dev/null || echo 0)
  if (( now - last_check < 30 )); then
    return 0
  fi
  printf '%s\n' "$now" >"$ts_file"
  if curl -4 -fsS --socks5-hostname 127.0.0.1:40000 --connect-timeout 3 --max-time 4 https://www.cloudflare.com/cdn-cgi/trace 2>/dev/null | grep -q 'warp=on'; then
    printf '1\n' >"$ok_file"
    return 0
  fi
  log "WARP tunnel liveness check failed; restarting prism-warp.service to self-heal..."
  systemctl restart prism-warp.service >/dev/null 2>&1 || true
  sleep 2
  if curl -4 -fsS --socks5-hostname 127.0.0.1:40000 --connect-timeout 3 --max-time 4 https://www.cloudflare.com/cdn-cgi/trace 2>/dev/null | grep -q 'warp=on'; then
    printf '1\n' >"$ok_file"
    systemctl kill -s HUP "$EGRESS_SERVICE" >/dev/null 2>&1 || true
    log "WARP tunnel self-healed successfully"
    return 0
  fi
  # Try alternate Cloudflare WARP endpoints if default 2408 is blocked/degraded
  if [[ -f "$conf" ]]; then
    local ep
    for ep in "162.159.192.1:500" "162.159.192.1:1701" "162.159.192.1:4500" "162.159.193.1:854" "162.159.192.1:2408"; do
      sed -i "s/^Endpoint = .*/Endpoint = ${ep}/" "$conf" 2>/dev/null || true
      systemctl restart prism-warp.service >/dev/null 2>&1 || true
      sleep 2
      if curl -4 -fsS --socks5-hostname 127.0.0.1:40000 --connect-timeout 3 --max-time 4 https://www.cloudflare.com/cdn-cgi/trace 2>/dev/null | grep -q 'warp=on'; then
        printf '1\n' >"$ok_file"
        systemctl kill -s HUP "$EGRESS_SERVICE" >/dev/null 2>&1 || true
        log "WARP tunnel self-healed using alternate endpoint ${ep}"
        return 0
      fi
    done
  fi
  printf '0\n' >"$ok_file"
  log "WARP tunnel remains unhealthy; egress proxy will automatically bypass to native"
}

reconcile_local_warp() {
  [[ "$ROLE" == "proxy" ]] || return 0
  local response="$1" desired
  desired=$(jq -r '.warp_desired // false' <<<"$response" 2>/dev/null || echo false)
  if [[ "$desired" == "true" ]]; then
    if proxy_warp_listening; then
      check_and_heal_warp_liveness
      return 0
    fi
    if [[ -f /etc/systemd/system/prism-warp.service && ( -s /etc/prismdns/warp/wireproxy.conf || -s /etc/prism-warp/wireproxy.conf ) ]]; then
      systemctl enable --now prism-warp.service >/dev/null 2>&1 || true
      for _ in {1..15}; do
        if proxy_warp_listening; then
          rm -f /run/prism-warp-liveness.ts /run/prism-warp-liveness.ok
          check_and_heal_warp_liveness
          systemctl kill -s HUP "$EGRESS_SERVICE" >/dev/null 2>&1 || true
          log "auto-started local WARP service (127.0.0.1:40000)"
          return 0
        fi
        sleep 0.25
      done
    fi
    if proxy_warp_provisioning; then
      return 0
    fi
    log "panel requested WARP routing; auto-provisioning local non-global WARP in background..."
    local master_url="$MASTER"
    systemd-run --unit=prism-warp-bootstrap --quiet --collect \
      --property=TimeoutStartSec=120 \
      /bin/bash -c "
        set -euo pipefail
        if ! command -v dns >/dev/null 2>&1 || ! grep -q 'cmd_warp' /usr/local/bin/dns 2>/dev/null; then
          curl -fsSL --connect-timeout 8 --max-time 30 '${master_url}/enhancer/dist/dns.sh' -o /tmp/prism-dns-auto.sh || \
          curl -fsSL --connect-timeout 10 --max-time 30 'https://raw.githubusercontent.com/${EGRESS_REPO}/${SCRIPT_REF}/dns.sh' -o /tmp/prism-dns-auto.sh
          install -m 0755 /tmp/prism-dns-auto.sh /usr/local/bin/dns
          rm -f /tmp/prism-dns-auto.sh
        fi
        /usr/local/bin/dns warp enable
        rm -f /run/prism-warp-liveness.ts /run/prism-warp-liveness.ok
        systemctl kill -s HUP '$EGRESS_SERVICE' >/dev/null 2>&1 || true
      " >/dev/null 2>&1 || true
  else
    rm -f /run/prism-warp-liveness.ts /run/prism-warp-liveness.ok
    if systemctl is-active --quiet prism-warp.service 2>/dev/null || systemctl is-enabled --quiet prism-warp.service 2>/dev/null; then
      systemctl disable --now prism-warp.service >/dev/null 2>&1 || true
      log "auto-stopped local WARP service (no services configured for WARP)"
    fi
  fi
}

proxy_registration() {
  local host_key body egress_ready=false warp_ready=false warp_provisioning=false
  ssh-keygen -A >/dev/null 2>&1 || true
  host_key=$(awk '{print $1" "$2}' /etc/ssh/ssh_host_ed25519_key.pub)
  if proxy_egress_ready; then
    egress_ready=true
  fi
  if proxy_warp_ready; then
    warp_ready=true
  fi
  if proxy_warp_provisioning; then
    warp_provisioning=true
  fi
  body=$(jq -nc \
    --arg secret "$CREDENTIAL" \
    --arg ssh_host_key "$host_key" \
    --argjson egress_ready "$egress_ready" \
    --argjson warp_ready "$warp_ready" \
    --argjson warp_provisioning "$warp_provisioning" \
    '{secret:$secret,ssh_host_key:$ssh_host_key,egress_ready:$egress_ready,warp_ready:$warp_ready,warp_provisioning:$warp_provisioning}')
  post_json "/enhancer/api/transport/proxy" "$body"
}

client_registration() {
  local ready_json="${1:-}" public_key body
  public_key=$(awk '{print $1" "$2}' "$SSH_KEY_FILE.pub")
  if [[ -n "$ready_json" ]]; then
    body=$(jq -nc \
      --arg token "$CREDENTIAL" \
      --arg ssh_public_key "$public_key" \
      --argjson ready "$ready_json" \
      '{token:$token,ssh_public_key:$ssh_public_key,ready_proxies:$ready}')
  else
    body=$(jq -nc \
      --arg token "$CREDENTIAL" \
      --arg ssh_public_key "$public_key" \
      '{token:$token,ssh_public_key:$ssh_public_key}')
  fi
  post_json "/enhancer/api/transport/client" "$body"
}

ensure_proxy_user() {
  if ! id prism-tunnel >/dev/null 2>&1; then
    useradd --system --create-home --home-dir /var/lib/prism-tunnel --shell /bin/bash prism-tunnel
  fi
  install -d -m 700 -o prism-tunnel -g prism-tunnel /var/lib/prism-tunnel
  install -d -m 700 -o prism-tunnel -g prism-tunnel /var/lib/prism-tunnel/.ssh
}

apply_authorized_keys() {
  local response="$1" temporary target="/var/lib/prism-tunnel/.ssh/authorized_keys"
  ensure_proxy_user
  temporary=$(mktemp)
  while IFS= read -r public_key; do
    [[ -n "$public_key" ]] || continue
    printf 'restrict,port-forwarding,permitopen="127.0.0.1:80",permitopen="127.0.0.1:443",permitopen="127.0.0.1:%s",permitopen="127.0.0.1:%s",command="/bin/sleep 2147483647" %s\n' \
      "$EGRESS_HTTP_PORT" "$EGRESS_HTTPS_PORT" "$public_key"
  done < <(jq -r '.peers[]?.ssh_public_key // empty' <<<"$response" | sort -u) >"$temporary"
  install -m 600 -o prism-tunnel -g prism-tunnel "$temporary" "$target"
  rm -f "$temporary"
  sshd -t
}

apply_egress_policy() {
  local response="$1" temporary
  systemctl is-active --quiet "$EGRESS_SERVICE" 2>/dev/null || return 0
  temporary=$(mktemp)
  jq '{services:(.egress_services // [])}' <<<"$response" >"$temporary"
  if [[ ! -f "$EGRESS_POLICY" ]] || ! cmp -s "$temporary" "$EGRESS_POLICY"; then
    install -m 644 "$temporary" "$EGRESS_POLICY"
    systemctl kill -s HUP "$EGRESS_SERVICE" >/dev/null 2>&1 || true
  fi
  rm -f "$temporary"
}

apply_client_egress_policy() {
  [[ "$ROLE" == "client" ]] || return 0
  local response="$1" temporary
  temporary=$(mktemp)
  jq '{services:(.egress_services // [])}' <<<"$response" >"$temporary"
  if [[ ! -f "$CLIENT_EGRESS_POLICY" ]] || ! cmp -s "$temporary" "$CLIENT_EGRESS_POLICY"; then
    install -m 644 "$temporary" "$CLIENT_EGRESS_POLICY"
    systemctl kill -s HUP "$CLIENT_EGRESS_SERVICE" >/dev/null 2>&1 || true
  fi
  rm -f "$temporary"
}

apply_proxy_authorization() {
  local response="$1" temporary use_egress=false
  local -a clients4=()
  if proxy_egress_ready; then
    use_egress=true
  fi
  while IFS= read -r address; do
    [[ -n "$address" ]] || continue
    if [[ "$address" != *:* ]]; then
      clients4+=("$address")
    fi
  done < <(jq -r '.authorized_ips[]?' <<<"$response" | sort -u)
  join_csv() { local IFS=,; printf '%s' "$*"; }
  temporary=$(mktemp)
  {
    if nft list table inet prism_authorization >/dev/null 2>&1; then
      echo 'delete table inet prism_authorization'
    fi
    echo 'table inet prism_authorization {'
    printf '  set clients4 { type ipv4_addr; flags timeout; timeout 30s;'
    ((${#clients4[@]} > 0)) && printf ' elements = { %s };' "$(join_csv "${clients4[@]}")"
    echo ' }'
    echo '  chain input {'
    echo '    type filter hook input priority -25; policy accept;'
    printf '    iifname "lo" tcp dport { 53, 80, 443, %s, %s } accept\n' "$EGRESS_HTTP_PORT" "$EGRESS_HTTPS_PORT"
    echo '    iifname "lo" udp dport 53 accept'
    printf '    ip saddr @clients4 tcp dport { 53, 80, 443, %s, %s } accept\n' "$EGRESS_HTTP_PORT" "$EGRESS_HTTPS_PORT"
    echo '    ip saddr @clients4 udp dport 53 accept'
    printf '    meta nfproto ipv4 tcp dport { 53, 80, 443, %s, %s } drop\n' "$EGRESS_HTTP_PORT" "$EGRESS_HTTPS_PORT"
    echo '    meta nfproto ipv4 udp dport 53 drop'
    printf '    meta nfproto ipv6 tcp dport { 53, %s, %s } drop\n' "$EGRESS_HTTP_PORT" "$EGRESS_HTTPS_PORT"
    echo '    meta nfproto ipv6 udp dport 53 drop'
    echo '  }'
    if $use_egress; then
      echo '  chain prerouting {'
      echo '    type nat hook prerouting priority -100; policy accept;'
      printf '    ip saddr @clients4 tcp dport 80 redirect to :%s\n' "$EGRESS_HTTP_PORT"
      printf '    ip saddr @clients4 tcp dport 443 redirect to :%s\n' "$EGRESS_HTTPS_PORT"
      echo '  }'
    fi
    echo '}'
  } >"$temporary"
  nft -c -f "$temporary"
  nft -f "$temporary"
  rm -f "$temporary"
}

collect_transport_cgroups() {
  local unit descriptor control_group path container_id container_desc container_pid cg_rel
  while IFS= read -r unit; do
    [[ -n "$unit" ]] || continue
    case "$unit" in
      prism*|docker*|containerd*|ssh*|nginx*|komari*|flux*) continue ;;
    esac
    descriptor=$(systemctl show "$unit" -p ExecStart --value 2>/dev/null || true)
    if ! printf '%s\n' "$unit $descriptor" | grep -Eqi 'xray|xrayr|v2ray|v2bx|v2node|sing-box|hysteria|tuic|trojan|shadowsocks|ss-server|soga|clash|mihomo|brook|naiveproxy'; then
      continue
    fi
    control_group=$(systemctl show "$unit" -p ControlGroup --value 2>/dev/null || true)
    [[ "$control_group" == /* ]] || continue
    path="/sys/fs/cgroup${control_group}"
    [[ -d "$path" ]] || continue
    stat -c '%i' "$path" 2>/dev/null || true
  done < <(systemctl list-units --type=service --state=running --no-legend --no-pager 2>/dev/null | awk '{print $1}' | sort -u)
  if command -v docker >/dev/null 2>&1; then
    while IFS='|' read -r container_id container_desc; do
      [[ -n "$container_id" ]] || continue
      printf '%s\n' "$container_desc" | grep -Eqi 'xray|xrayr|v2ray|v2bx|v2node|sing-box|hysteria|tuic|trojan|shadowsocks|ss-server|soga|clash|mihomo|brook|naiveproxy' || continue
      container_pid=$(docker inspect -f '{{.State.Pid}}' "$container_id" 2>/dev/null || true)
      [[ "$container_pid" =~ ^[1-9][0-9]*$ && -r "/proc/$container_pid/cgroup" ]] || continue
      cg_rel=$(awk -F: 'NR==1 {print $3}' "/proc/$container_pid/cgroup" 2>/dev/null || true)
      [[ "$cg_rel" == /* && -d "/sys/fs/cgroup${cg_rel}" ]] || continue
      stat -c '%i' "/sys/fs/cgroup${cg_rel}" 2>/dev/null || true
    done < <(docker ps --format '{{.ID}}|{{.Image}} {{.Names}} {{.Command}}' 2>/dev/null || true)
  fi
}

apply_client_egress_rules() {
  [[ "$ROLE" == "client" ]] || return 0
  local temporary cgroup_id service_count
  local -a cgroups=()
  service_count=$(jq -er '(.egress_services // []) | if type == "array" then length else error("invalid egress services") end' <<<"$1") ||
    fail "invalid client egress configuration"
  if [[ ( "${TRANSPORT_MODE:-direct}" == "direct" || "${TRANSPORT_MODE:-direct}" == "encrypted" ) && "$service_count" != 0 ]]; then
    mapfile -t cgroups < <(collect_transport_cgroups | sort -u)
  fi
  ((${#cgroups[@]} > 0)) || {
    if nft list table inet prism_client_egress >/dev/null 2>&1; then
      nft delete table inet prism_client_egress
    fi
    return 0
  }
  temporary=$(mktemp)
  {
    if nft list table inet prism_client_egress >/dev/null 2>&1; then
      echo 'delete table inet prism_client_egress'
    fi
    echo 'table inet prism_client_egress {'
    echo '  chain redirect_output {'
    echo '    type nat hook output priority -110; policy accept;'
    for cgroup_id in "${cgroups[@]}"; do
      printf '    meta cgroup %s tcp dport 80 redirect to :%s\n' "$cgroup_id" "$CLIENT_EGRESS_HTTP_PORT"
      printf '    meta cgroup %s tcp dport 443 redirect to :%s\n' "$cgroup_id" "$CLIENT_EGRESS_HTTPS_PORT"
    done
    echo '  }'
    echo '  chain reject_quic {'
    echo '    type filter hook output priority -200; policy accept;'
    for cgroup_id in "${cgroups[@]}"; do
      printf '    meta cgroup %s udp dport 443 reject with icmpx type port-unreachable\n' "$cgroup_id"
    done
    echo '  }'
    echo '}'
  } >"$temporary"
  if ! nft -f "$temporary"; then
    rm -f "$temporary"
    fail "failed to install transparent SNI egress rules"
  fi
  rm -f "$temporary"
}

sync_proxy() {
  local response
  response=$(proxy_registration)
  jq -e '.role == "proxy" and (.peers | type == "array") and (.authorized_ips | type == "array")' <<<"$response" >/dev/null
  apply_authorized_keys "$response"
  apply_proxy_authorization "$response"
  reconcile_local_warp "$response"
  apply_egress_policy "$response"
  log "proxy transport synchronized: $(jq '.authorized_ips | length' <<<"$response") authorized IPs, $(jq '.peers | length' <<<"$response") encrypted clients"
}

service_id_for_proxy() {
  printf '%s' "$1" | sha256sum | cut -c1-12
}

write_tunnel_service() {
  local peer="$1" proxy_id proxy_ip ssh_host ssh_port host_key remote_http remote_https service_id service_path known_hosts temporary port_seed local_http local_https
  proxy_id=$(jq -r '.proxy_id' <<<"$peer")
  proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
  ssh_host=$(jq -r '.ssh_host' <<<"$peer")
  ssh_port=$(jq -r '.ssh_port' <<<"$peer")
  host_key=$(jq -r '.ssh_host_key' <<<"$peer")
  remote_http=$(jq -r '.remote_http // 80' <<<"$peer")
  remote_https=$(jq -r '.remote_https // 443' <<<"$peer")
  service_id=$(service_id_for_proxy "$proxy_id")
  port_seed=$((16#${service_id:0:4}))
  local_http=$((20000 + (port_seed % 10000)))
  local_https=$((30000 + (port_seed % 10000)))
  service_path="/etc/systemd/system/prism-transport-ssh-$service_id.service"
  known_hosts="$STATE_DIR/known-hosts-$service_id"
  printf 'prism-%s %s\n' "$service_id" "$host_key" >"$known_hosts"
  chmod 600 "$known_hosts"
  temporary=$(mktemp)
  cat >"$temporary" <<EOF
[Unit]
Description=Prism encrypted SNI tunnel for $proxy_id
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/bin/ssh -N -T -l prism-tunnel -p $ssh_port -i $SSH_KEY_FILE -o IdentitiesOnly=yes -o BatchMode=yes -o ExitOnForwardFailure=yes -o ConnectTimeout=8 -o ConnectionAttempts=3 -o TCPKeepAlive=yes -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -o IPQoS=none -o LogLevel=ERROR -o StrictHostKeyChecking=yes -o UserKnownHostsFile=$known_hosts -o HostKeyAlias=prism-$service_id -L 127.0.0.1:$local_http:127.0.0.1:$remote_http -L 127.0.0.1:$local_https:127.0.0.1:$remote_https $ssh_host
Restart=always
RestartSec=2
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF
  if [[ ! -f "$service_path" ]] || ! cmp -s "$temporary" "$service_path"; then
    install -m 644 "$temporary" "$service_path"
    systemctl daemon-reload
    systemctl enable "prism-transport-ssh-$service_id.service" >/dev/null 2>&1
    systemctl restart "prism-transport-ssh-$service_id.service"
  else
    systemctl is-active --quiet "prism-transport-ssh-$service_id.service" ||
      systemctl start "prism-transport-ssh-$service_id.service"
  fi
  rm -f "$temporary"
  jq -nc --arg proxy_id "$proxy_id" --arg proxy_ip "$proxy_ip" --arg service_id "$service_id" \
    --argjson local_http "$local_http" --argjson local_https "$local_https" \
    --argjson remote_http "$remote_http" --argjson remote_https "$remote_https" \
    '{proxy_id:$proxy_id,proxy_ip:$proxy_ip,service_id:$service_id,local_http:$local_http,local_https:$local_https,remote_http:$remote_http,remote_https:$remote_https}'
}

remove_stale_tunnels() {
  local current="$1" old_proxy service_id proxy_ip
  [[ -f "$ACTIVE_FILE" ]] || return 0
  while IFS= read -r old_proxy; do
    [[ -n "$old_proxy" ]] || continue
    if jq -e --arg proxy_id "$old_proxy" '.[] | select(.proxy_id == $proxy_id)' <<<"$current" >/dev/null; then
      continue
    fi
    service_id=$(jq -r --arg proxy_id "$old_proxy" '.[] | select(.proxy_id == $proxy_id) | .service_id' "$ACTIVE_FILE")
    if [[ -n "$service_id" && "$service_id" != "null" ]]; then
      systemctl disable --now "prism-transport-ssh-$service_id.service" >/dev/null 2>&1 || true
      rm -f "/etc/systemd/system/prism-transport-ssh-$service_id.service" "$STATE_DIR/known-hosts-$service_id"
    fi
  done < <(jq -r '.[].proxy_id' "$ACTIVE_FILE")
  systemctl daemon-reload
}

apply_client_rules() {
  local current="$1" temporary peer proxy_ip local_http local_https
  temporary=$(mktemp)
  {
    if nft list table inet prism_transport >/dev/null 2>&1; then
      echo 'delete table inet prism_transport'
    fi
    echo 'table inet prism_transport {'
    echo '  counter tx {}'
    echo '  counter rx {}'
    echo '  chain account_output {'
    echo '    type filter hook output priority -200; policy accept;'
    if [[ "$TRANSPORT_MODE" == "encrypted" ]]; then
      while IFS= read -r peer; do
        [[ -n "$peer" ]] || continue
        local_http=$(jq -r '.local_http' <<<"$peer")
        local_https=$(jq -r '.local_https' <<<"$peer")
        printf '    tcp sport { %s, %s } counter name "rx" return\n' "$local_http" "$local_https"
      done < <(jq -c 'unique_by(.proxy_ip)[]?' <<<"$current")
    fi
    while IFS= read -r peer; do
      [[ -n "$peer" ]] || continue
      proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
      printf '    ip daddr %s tcp dport { 80, 443 } counter name "tx"\n' "$proxy_ip"
    done < <(jq -c 'unique_by(.proxy_ip)[]?' <<<"$current")
    echo '  }'
    if [[ "$TRANSPORT_MODE" == "direct" ]]; then
      echo '  chain account_input {'
      echo '    type filter hook input priority -200; policy accept;'
      while IFS= read -r peer; do
        [[ -n "$peer" ]] || continue
        proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
        printf '    ip saddr %s tcp sport { 80, 443 } counter name "rx"\n' "$proxy_ip"
      done < <(jq -c 'unique_by(.proxy_ip)[]?' <<<"$current")
      echo '  }'
    else
      echo '  chain redirect_output {'
      echo '    type nat hook output priority -100; policy accept;'
      while IFS= read -r peer; do
        [[ -n "$peer" ]] || continue
        proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
        local_http=$(jq -r '.local_http' <<<"$peer")
        local_https=$(jq -r '.local_https' <<<"$peer")
        printf '    ip daddr %s tcp dport 80 redirect to :%s\n' "$proxy_ip" "$local_http"
        printf '    ip daddr %s tcp dport 443 redirect to :%s\n' "$proxy_ip" "$local_https"
      done < <(jq -c 'unique_by(.proxy_ip)[]?' <<<"$current")
      echo '  }'
    fi
    echo '}'
  } >"$temporary"
  if ! nft -f "$temporary"; then
    rm -f "$temporary"
    fail "failed to replace client transport rules"
  fi
  rm -f "$temporary"
}

direct_peer_ready() {
  local peer="$1" proxy_ip
  proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
  tunnel_port_ready "$proxy_ip" 80 &&
    tunnel_port_ready "$proxy_ip" 443 &&
    tunnel_https_probe "$proxy_ip"
}

tunnel_port_ready() {
  local proxy_ip="$1" port="$2"
  timeout 4 bash -c \
    'exec 3<>"/dev/tcp/$1/$2"; exec 3>&-; exec 3<&-' \
    prism-transport-probe "$proxy_ip" "$port" >/dev/null 2>&1
}

tunnel_https_probe() {
  local proxy_ip="$1" domain path
  while IFS='|' read -r domain path; do
    if curl -4 -fsS -o /dev/null \
      --resolve "${domain}:443:${proxy_ip}" \
      --connect-timeout 3 --max-time 5 \
      "https://${domain}${path}"; then
      return 0
    fi
  done <<'EOF'
www.gstatic.com|/generate_204
www.cloudflare.com|/cdn-cgi/trace
EOF
  return 1
}

tunnel_functional_ready() {
  local peer="$1" proxy_ip service_id now last_probe probe_file restart_file last_restart
  proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
  service_id=$(jq -r '.service_id' <<<"$peer")
  probe_file="$STATE_DIR/health-probe-$service_id.timestamp"
  restart_file="$STATE_DIR/health-restart-$service_id.timestamp"
  now=$(date +%s)
  last_probe=$(cat "$probe_file" 2>/dev/null || echo 0)
  if ((now - last_probe < 60)); then
    return 0
  fi
  if tunnel_https_probe "$proxy_ip"; then
    printf '%s\n' "$now" >"$probe_file"
    return 0
  fi
  last_restart=$(cat "$restart_file" 2>/dev/null || echo 0)
  if ((now - last_restart >= 60)); then
    systemctl restart "prism-transport-ssh-$service_id.service"
    printf '%s\n' "$now" >"$restart_file"
    for _ in $(seq 1 10); do
      tunnel_ready "$peer" && break
      sleep 1
    done
    if tunnel_https_probe "$proxy_ip"; then
      printf '%s\n' "$now" >"$probe_file"
      return 0
    fi
  fi
  printf '%s\n' "$now" >"$probe_file"
  return 1
}

tunnel_ready() {
  local peer="$1" proxy_ip service_id local_http local_https
  proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
  service_id=$(jq -r '.service_id' <<<"$peer")
  local_http=$(jq -r '.local_http' <<<"$peer")
  local_https=$(jq -r '.local_https' <<<"$peer")
  systemctl is-active --quiet "prism-transport-ssh-$service_id.service" &&
    ss -lnt 2>/dev/null | awk -v port=":$local_http" '$4 ~ port "$" {found=1} END {exit !found}' &&
    ss -lnt 2>/dev/null | awk -v port=":$local_https" '$4 ~ port "$" {found=1} END {exit !found}' &&
    tunnel_port_ready "$proxy_ip" 80 &&
    tunnel_port_ready "$proxy_ip" 443
}

run_client_native_probes() {
  [[ "$ROLE" == "client" ]] || return 0
  local response="$1" count task task_id svc_id domain probe_url real_ip tmp_body meta code final_url time_sec latency_ms snippet err_msg payload
  count=$(jq '(.native_probes // []) | length' <<<"$response" 2>/dev/null || echo 0)
  [[ "$count" -gt 0 ]] || return 0
  while IFS= read -r task; do
    [[ -n "$task" ]] || continue
    task_id=$(jq -r '.task_id // ""' <<<"$task")
    svc_id=$(jq -r '.service_id // ""' <<<"$task")
    domain=$(jq -r '.domain // ""' <<<"$task")
    probe_url=$(jq -r '.probe_url // ""' <<<"$task")
    [[ -n "$svc_id" && -n "$domain" && -n "$probe_url" ]] || continue
    real_ip=$(curl -4 -fsS --connect-timeout 3 --max-time 5 -H 'accept: application/dns-json' "https://1.1.1.1/dns-query?name=${domain}&type=A" 2>/dev/null | jq -r '[.Answer[]? | select(.type==1) | .data][0] // empty' 2>/dev/null || true)
    if [[ -z "$real_ip" ]]; then
      real_ip=$(curl -4 -fsS --connect-timeout 3 --max-time 5 -H 'accept: application/dns-json' "https://8.8.8.8/resolve?name=${domain}&type=A" 2>/dev/null | jq -r '[.Answer[]? | select(.type==1) | .data][0] // empty' 2>/dev/null || true)
    fi
    tmp_body=$(mktemp)
    code=0; final_url="$probe_url"; latency_ms=0; err_msg=""
    if [[ -n "$real_ip" ]]; then
      meta=$(curl -4 -sS -L --connect-timeout 4 --max-time 7 --resolve "${domain}:443:${real_ip}" \
        -A 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36' \
        -H 'Accept-Language: en-US,en;q=0.9' \
        -o "$tmp_body" -w '%{http_code}|%{url_effective}|%{time_total}' "$probe_url" 2>/dev/null) || err_msg="原生直连超时或连接失败"
      if [[ -z "$err_msg" ]]; then
        IFS='|' read -r code final_url time_sec <<<"$meta"
        latency_ms=$(awk -v t="${time_sec:-0}" 'BEGIN{printf "%d", t*1000}')
      fi
    else
      err_msg="公网 DNS 解析失败"
    fi
    snippet=$(head -c 262144 "$tmp_body" 2>/dev/null | tr -d '\000' || true)
    rm -f "$tmp_body"
    payload=$(jq -nc \
      --arg token "$CREDENTIAL" \
      --arg task_id "$task_id" \
      --arg service_id "$svc_id" \
      --argjson http_code "${code:-0}" \
      --arg final_url "$final_url" \
      --arg body_snippet "$snippet" \
      --argjson latency_ms "${latency_ms:-0}" \
      --arg error "$err_msg" \
      '{token:$token,task_id:$task_id,service_id:$service_id,http_code:$http_code,final_url:$final_url,body_snippet:$body_snippet,latency_ms:$latency_ms,error:$error}')
    post_json "/enhancer/api/client-native-check/report" "$payload" >/dev/null 2>&1 || true
  done < <(jq -c '.native_probes[]?' <<<"$response")
}

sync_client() {
  local response peer proxy_id ready_json old_ready="[]" old_current="[]" current_json changed_ready=false changed_current=false
  local -a ready=() current=()
  response=$(client_registration)
  jq -e '.role == "client" and (.peers | type == "array")' <<<"$response" >/dev/null
  run_client_native_probes "$response"
  apply_client_egress_policy "$response"
  while IFS= read -r peer; do
    [[ -n "$peer" ]] || continue
    if [[ "$TRANSPORT_MODE" == "direct" ]]; then
      current+=("$(jq -c '{proxy_id,proxy_ip}' <<<"$peer")")
    else
      current+=("$(write_tunnel_service "$peer")")
    fi
  done < <(jq -c '.peers[]?' <<<"$response")
  if ((${#current[@]})); then
    current_json=$(printf '%s\n' "${current[@]}" | jq -sc 'sort_by(.proxy_id)')
  else
    current_json='[]'
  fi
  [[ -f "$ACTIVE_FILE" ]] && old_current=$(jq -c 'sort_by(.proxy_id)' "$ACTIVE_FILE" 2>/dev/null || echo '[]')
  [[ "$(jq -c 'sort_by(.proxy_id)' <<<"$current_json")" != "$old_current" ]] && changed_current=true
  if [[ "$TRANSPORT_MODE" == "direct" ]]; then
    remove_stale_tunnels '[]'
  else
    remove_stale_tunnels "$current_json"
  fi
  if [[ "${PRISM_FORCE_TRANSPORT_RULES:-0}" == "1" ]] || $changed_current ||
    ! nft list counter inet prism_transport tx >/dev/null 2>&1 ||
    ! nft list counter inet prism_transport rx >/dev/null 2>&1; then
    apply_client_rules "$current_json"
  fi
  apply_client_egress_rules "$response"
  printf '%s\n' "$current_json" >"$ACTIVE_FILE"
  chmod 600 "$ACTIVE_FILE"
  sleep 2
  while IFS= read -r peer; do
    [[ -n "$peer" ]] || continue
    proxy_id=$(jq -r '.proxy_id' <<<"$peer")
    if [[ "$TRANSPORT_MODE" == "direct" ]] && direct_peer_ready "$peer"; then
      ready+=("$proxy_id")
    elif [[ "$TRANSPORT_MODE" == "encrypted" ]] && tunnel_ready "$peer"; then
      if ! tunnel_functional_ready "$peer"; then
        log "encrypted peer $proxy_id accepted TCP but failed HTTPS path probe; tunnel restart was requested without changing service routing"
      fi
      ready+=("$proxy_id")
    fi
  done < <(jq -c '.[]?' <<<"$current_json")
  ready_json=$(printf '%s\n' "${ready[@]}" | jq -Rsc 'split("\n") | map(select(length > 0)) | unique | sort')
  [[ -f "$READY_FILE" ]] && old_ready=$(jq -c 'sort' "$READY_FILE" 2>/dev/null || echo '[]')
  [[ "$(jq -c 'sort' <<<"$ready_json")" != "$old_ready" ]] && changed_ready=true
  client_registration "$ready_json" >/dev/null
  printf '%s\n' "$ready_json" >"$READY_FILE"
  chmod 600 "$READY_FILE"
  if $changed_ready && systemctl is-active --quiet prism-agent 2>/dev/null; then
    systemctl restart prism-agent
  fi
  systemctl disable --now wg-quick@prismwg0 >/dev/null 2>&1 || true
  rm -f /etc/wireguard/prismwg0.conf
  log "client transport synchronized: ${#ready[@]}/$(jq '.peers | length' <<<"$response") $TRANSPORT_MODE peers ready"
}

sync_transport() {
  load_environment
  [[ "$ROLE" == "proxy" || "$ROLE" == "client" ]] || fail "transport role is not configured"
  [[ "$MASTER" =~ ^https?://[^[:space:]]+$ ]] || fail "invalid panel URL"
  [[ -n "$CREDENTIAL" ]] || fail "transport credential is missing"
  ensure_dependencies
  if [[ "$ROLE" == "proxy" ]]; then
    proxy_egress_ready || install_proxy_egress
    sync_proxy
  else
    ensure_client_keypair
    client_egress_ready || install_client_egress
    sync_client
  fi
}

install_timer() {
  cat >/etc/systemd/system/prism-transport.service <<EOF
[Unit]
Description=Synchronize Prism encrypted SNI transport
After=network-online.target prism-agent.service
Wants=network-online.target

[Service]
Type=oneshot
Nice=10
IOSchedulingClass=best-effort
IOSchedulingPriority=6
ExecStart=$INSTALL_PATH --sync
EOF
  if [[ "$ROLE" == "proxy" ]]; then
    cat >/etc/systemd/system/prism-transport.timer <<'EOF'
[Unit]
Description=Refresh Prism encrypted SNI transport

[Timer]
OnActiveSec=10s
OnUnitInactiveSec=15s
AccuracySec=2s
Persistent=true

[Install]
WantedBy=timers.target
EOF
  else
    cat >/etc/systemd/system/prism-transport.timer <<'EOF'
[Unit]
Description=Refresh Prism encrypted SNI transport

[Timer]
OnActiveSec=15s
OnUnitInactiveSec=45s
AccuracySec=5s
Persistent=true

[Install]
WantedBy=timers.target
EOF
  fi
  find /run -maxdepth 1 -type f -name 'prismdns-media-check.*' -mmin +1440 -delete 2>/dev/null || true
  systemctl daemon-reload
  systemctl enable prism-transport.timer >/dev/null
  systemctl restart prism-transport.timer
}

install_transport() {
  [[ "$ROLE" == "proxy" || "$ROLE" == "client" ]] || fail "--proxy or --client is required"
  MASTER="${MASTER%/}"
  [[ "$MASTER" =~ ^https?://[^[:space:]]+$ ]] || fail "--master is invalid"
  [[ -n "$CREDENTIAL" ]] || fail "--secret or --token is required"
  ensure_dependencies
  mkdir -p "$INSTALL_DIR" "$STATE_DIR"
  if [[ "$(readlink -f "$0")" != "$INSTALL_PATH" ]]; then
    install -m 755 "$0" "$INSTALL_PATH"
  else
    chmod 755 "$INSTALL_PATH"
  fi
  save_environment
  install_media_worker
  [[ "$ROLE" == "client" ]] && ensure_client_keypair
  [[ "$ROLE" == "proxy" ]] && install_proxy_egress
  install_timer
  sync_transport
  log "Prism SNI transport v$VERSION installed for role $ROLE (mode: $TRANSPORT_MODE)"
}

refresh_transport_runtime() {
  [[ -s "$ENV_FILE" ]] || fail "transport configuration is missing"
  load_environment
  [[ "$ROLE" == "client" || "$ROLE" == "proxy" ]] || fail "runtime refresh requires an existing client or proxy transport"
  [[ "$MASTER" =~ ^https?://[^[:space:]]+$ && -n "$CREDENTIAL" ]] || fail "invalid transport configuration"
  ensure_dependencies
  if [[ -n "$MASTER_OVERRIDE" ]]; then
    save_environment
  fi
  mkdir -p "$INSTALL_DIR"
  if [[ "$(readlink -f "$0")" != "$INSTALL_PATH" ]]; then
    install -m 755 "$0" "$INSTALL_PATH"
  fi
  install_media_worker
  install_timer
  log "$ROLE transport runtime refreshed without Agent reinstall or immediate transport sync"
}

install_media_worker() {
  local installer="${PRISM_MEDIA_INSTALLER_FILE:-}" temporary=""
  if [[ -z "$installer" ]]; then
    temporary=$(mktemp /run/prism-media-installer.XXXXXX)
    installer="$temporary"
    if ! download_dist_or_github \
      "prism_media.sh" \
      "https://raw.githubusercontent.com/$EGRESS_REPO/refs/tags/$SCRIPT_REF/prism_media.sh" \
      "$installer"; then
      rm -f "$temporary"
      fail "could not download the IPv4 media check worker"
    fi
  fi
  bash -n "$installer" || fail "invalid IPv4 media check installer"
  bash "$installer" --install
  [[ -z "$temporary" ]] || rm -f "$temporary"
}

install_node_update_worker() (
  local installer="${PRISM_NODE_UPDATE_INSTALLER_FILE:-}" temporary='' http_code
  trap '[[ -z "$temporary" ]] || rm -f -- "$temporary"' EXIT
  if [[ -z "$installer" ]]; then
    temporary=$(mktemp "$NODE_UPDATE_RUNTIME_DIR/prism-node-update-installer.XXXXXX") || return 1
    installer="$temporary"
    http_code=$(curl -q -fsS --proto '=https' --tlsv1.2 --connect-timeout 8 --max-time 30 \
      --max-filesize 1048576 --write-out '%{http_code}' --output "$installer" \
      "https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/$SCRIPT_REF/prism_node_update.sh") &&
      [[ "$http_code" == 200 ]] || { fail "could not download the node update worker"; return 1; }
  fi
  [[ -s "$installer" ]] && grep -Fxq "NODE_UPDATE_REPO='xcxcadc/chenfei-Glass-Prism-dns'" "$installer" &&
    bash -n "$installer" || { fail "invalid node update worker"; return 1; }
  bash "$installer" --install "$SCRIPT_REF"
)

install_optional_node_update_worker() {
  install_node_update_worker || log 'node update worker unavailable; transport remains installed, retry a runtime refresh to enable online management' >&2
}

managed_transport_refresh() (
  local expected="$PRISM_NODE_UPDATE_EXPECTED_VERSION" marker="$STATE_DIR/node-updates/version"
  local snapshot path signature index timer
  local -a files=("$INSTALL_PATH" "$MEDIA_WORKER_PATH" "$INSTALL_DIR/prism_node_update.sh"
    "$NODE_UPDATE_UNIT_DIR"/prism-{transport,media-check,node-update}.{service,timer})
  local -a timers=(prism-transport.timer prism-media-check.timer prism-node-update.timer prismdns-service-audit.timer)
  local -A enabled=() active=()
  [[ -z "$MASTER_OVERRIDE" && -f "$marker" && ! -L "$marker" && "$(<"$marker")" == "$expected" ]] ||
    fail "node version or configuration changed before refresh"
  umask 077
  snapshot=$(mktemp -d "$NODE_UPDATE_RUNTIME_DIR/prism-node-refresh.XXXXXX")
  trap 'rm -rf -- "$snapshot"' EXIT
  for index in "${!files[@]}"; do
    path="${files[index]}"
    [[ -e "$path" || -L "$path" ]] || continue
    case "${path##*/}" in
      prism_transport.sh) signature='^INSTALL_PATH="\$INSTALL_DIR/prism_transport\.sh"$' ;;
      prism-media-check) signature='^INSTALL_PATH="/usr/local/bin/prism-media-check"$' ;;
      prism_node_update.sh) signature="^NODE_UPDATE_REPO='xcxcadc/chenfei-Glass-Prism-dns'$" ;;
      prism-transport.*) signature='^Description=(Synchronize|Refresh) Prism ' ;;
      prism-media-check.*) signature='^Description=Prism DNS (official IPv4 media check worker|media check job poller)$' ;;
      prism-node-update.*) signature='^Description=Prism DNS managed node update$' ;;
    esac
    [[ -f "$path" && ! -L "$path" ]] && grep -Eq "$signature" "$path" || fail "unowned file blocks managed refresh"
    cp -p -- "$path" "$snapshot/$index"
  done
  for timer in "${timers[@]}"; do
    if systemctl is-enabled --quiet "$timer"; then enabled["$timer"]=1; fi
    if systemctl is-active --quiet "$timer"; then active["$timer"]=1; fi
  done
  trap 'exit 143' TERM INT
  trap 'result=$?; trap - EXIT; if ((result != 0)) && [[ "$(<"$marker")" != "$SCRIPT_REF" ]]; then
    restore_transport_runtime; fi; [[ -z "$snapshot" ]] || rm -rf -- "$snapshot"; exit "$result"' EXIT
  refresh_transport_runtime
  install_node_update_worker
)

restore_transport_runtime() {
  local index path timer restore_error=0
  for timer in "${timers[@]}"; do
    if [[ -z "${active[$timer]:-}" ]] && systemctl is-active --quiet "$timer"; then
      systemctl stop "$timer" || restore_error=1
    fi
    if [[ -z "${enabled[$timer]:-}" ]] && systemctl is-enabled --quiet "$timer"; then
      systemctl disable "$timer" >/dev/null || restore_error=1
    fi
  done
  for index in "${!files[@]}"; do
    path="${files[index]}"
    if [[ -f "$snapshot/$index" ]]; then
      cp -p -- "$snapshot/$index" "$path" || restore_error=1
    else
      rm -f -- "$path" || restore_error=1
    fi
  done
  systemctl daemon-reload || restore_error=1
  for timer in "${timers[@]}"; do
    if [[ -n "${enabled[$timer]:-}" ]]; then
      systemctl enable "$timer" >/dev/null || restore_error=1
    fi
    if [[ -n "${active[$timer]:-}" ]]; then
      systemctl restart "$timer" || restore_error=1
    fi
  done
  if ((restore_error)); then
    log "managed refresh rollback incomplete; backup retained at $snapshot" >&2
    snapshot=''
  fi
}

uninstall_node_update_worker() {
  local worker="$INSTALL_DIR/prism_node_update.sh" unit path
  local -a owned_units=()
  if [[ -e "$worker" || -L "$worker" ]]; then
    [[ -f "$worker" && ! -L "$worker" ]] &&
      grep -Fxq "NODE_UPDATE_REPO='xcxcadc/chenfei-Glass-Prism-dns'" "$worker" || fail 'unowned node update worker'
    bash "$worker" --uninstall
    return
  fi
  for unit in prism-node-update.timer prism-node-update.service; do
    path="$NODE_UPDATE_UNIT_DIR/$unit"
    [[ -f "$path" && ! -L "$path" ]] && grep -Fxq 'Description=Prism DNS managed node update' "$path" || continue
    if [[ "$unit" == *.timer ]]; then
      systemctl disable --now "$unit" >/dev/null
    else
      systemctl stop "$unit"
    fi
    owned_units+=("$path")
  done
  if ((${#owned_units[@]})); then rm -f -- "${owned_units[@]}"; fi
}

uninstall_transport() {
  local service_id proxy_ip
  load_environment
  (uninstall_node_update_worker) || log 'could not remove node update worker; no unowned files were removed' >&2
  if [[ -x /usr/local/bin/prism-media-check ]]; then
    /usr/local/bin/prism-media-check --uninstall
  fi
  if [[ "$ROLE" == "client" && -n "$MASTER" && -n "$CREDENTIAL" && -s "$SSH_KEY_FILE.pub" ]]; then
    client_registration '[]' >/dev/null 2>&1 || true
  fi
  if [[ -f "$ACTIVE_FILE" ]]; then
    while IFS=$'\t' read -r service_id proxy_ip; do
      systemctl disable --now "prism-transport-ssh-$service_id.service" >/dev/null 2>&1 || true
      rm -f "/etc/systemd/system/prism-transport-ssh-$service_id.service" "$STATE_DIR/known-hosts-$service_id"
    done < <(jq -r '.[] | [.service_id,.proxy_ip] | @tsv' "$ACTIVE_FILE")
  fi
  systemctl disable --now prism-transport.timer prism-transport.service >/dev/null 2>&1 || true
  if [[ "$ROLE" == "proxy" ]]; then
    systemctl disable --now "$EGRESS_SERVICE" >/dev/null 2>&1 || true
    rm -f "/etc/systemd/system/$EGRESS_SERVICE" "$EGRESS_BINARY" "$EGRESS_POLICY"
  else
    systemctl disable --now "$CLIENT_EGRESS_SERVICE" >/dev/null 2>&1 || true
    rm -f "/etc/systemd/system/$CLIENT_EGRESS_SERVICE" "$EGRESS_BINARY" "$CLIENT_EGRESS_POLICY"
  fi
  systemctl disable --now wg-quick@prismwg0 >/dev/null 2>&1 || true
  rm -f /etc/systemd/system/prism-transport.service /etc/systemd/system/prism-transport.timer
  rm -f /etc/wireguard/prismwg0.conf "$ENV_FILE"
  nft delete table inet prism_transport >/dev/null 2>&1 || true
  nft delete table inet prism_client_egress >/dev/null 2>&1 || true
  nft delete table inet prism_authorization >/dev/null 2>&1 || true
  systemctl daemon-reload
  log "Prism SNI transport removed"
}

show_status() {
  load_environment
  echo "role=${ROLE:-unconfigured}"
  echo "mode=${TRANSPORT_MODE:-direct}"
  echo "timer=$(systemctl is-active prism-transport.timer 2>/dev/null || true)"
  if [[ "${ROLE:-}" == "proxy" ]]; then
    echo "egress=$(systemctl is-active "$EGRESS_SERVICE" 2>/dev/null || true)"
  else
    echo "client_egress=$(systemctl is-active "$CLIENT_EGRESS_SERVICE" 2>/dev/null || true)"
  fi
  if [[ -f "$READY_FILE" ]]; then
    echo "ready_proxies=$(jq -c . "$READY_FILE" 2>/dev/null || echo '[]')"
  fi
  systemctl --no-pager --plain list-units 'prism-transport-ssh-*.service' 2>/dev/null || true
}

main() {
  require_root
  parse_args "$@"
  if ! command -v flock >/dev/null 2>&1; then
    case "$ACTION" in
      install) ensure_dependencies ;;
      refresh-runtime) load_environment; ensure_dependencies ;;
      *) fail "flock is missing; install util-linux and refresh the runtime" ;;
    esac
  fi
  exec 9>/run/prism-transport.lock
  flock -w 30 9 || fail "another transport synchronization is still running"
  case "$ACTION" in
    sync) sync_transport ;;
    refresh-runtime)
      if [[ -n "${PRISM_NODE_UPDATE_EXPECTED_VERSION:-}" ]]; then
        managed_transport_refresh
      else
        refresh_transport_runtime
        install_optional_node_update_worker
      fi
      ;;
    uninstall) uninstall_transport ;;
    status) show_status ;;
    install) install_transport; install_optional_node_update_worker ;;
  esac
}

main "$@"
