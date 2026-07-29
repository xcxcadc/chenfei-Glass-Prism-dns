#!/usr/bin/env bash

set -Eeuo pipefail

VERSION="2.2.1"
INSTALL_DIR="/usr/local/lib/prismdns"
INSTALL_PATH="$INSTALL_DIR/prism_transport.sh"
STATE_DIR="/var/lib/prism-transport"
ENV_FILE="/etc/prismdns/transport.env"
SSH_KEY_FILE="$STATE_DIR/ssh_key"
READY_FILE="$STATE_DIR/ready-proxies.json"
ACTIVE_FILE="$STATE_DIR/active-proxies.json"
ROLE=""
MASTER=""
CREDENTIAL=""
ACTION="install"

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
      --master) MASTER="${2:-}"; shift 2 ;;
      --secret|--token) CREDENTIAL="${2:-}"; shift 2 ;;
      --sync) ACTION="sync"; shift ;;
      --uninstall) ACTION="uninstall"; shift ;;
      --status) ACTION="status"; shift ;;
      --help|-h)
        echo "Usage: prism_transport.sh (--proxy --secret SECRET | --client --token TOKEN) --master URL [--sync|--uninstall|--status]"
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
  MASTER="${MASTER%/}"
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
  for command_name in curl jq ssh ssh-keygen ip nft; do
    command -v "$command_name" >/dev/null 2>&1 || ready=false
  done
  if [[ "$ROLE" == "proxy" ]]; then
    command -v sshd >/dev/null 2>&1 || ready=false
  fi
  $ready && return 0

  local manager
  manager=$(detect_package_manager)
  log "installing encrypted TCP transport dependencies"
  case "$manager" in
    apt)
      apt-get update -qq
      if [[ "$ROLE" == "proxy" ]]; then
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl jq openssh-client openssh-server iproute2 nftables
      else
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl jq openssh-client iproute2 nftables
      fi
      ;;
    dnf|yum)
      "$manager" install -y curl jq openssh-clients iproute nftables
      [[ "$ROLE" == "proxy" ]] && "$manager" install -y openssh-server
      ;;
    apk)
      apk add --no-cache curl jq openssh-client iproute2 nftables
      [[ "$ROLE" == "proxy" ]] && apk add --no-cache openssh-server
      ;;
    *) fail "unsupported package manager; install curl, jq, OpenSSH and iproute2 first" ;;
  esac
}

save_environment() {
  mkdir -p "$(dirname "$ENV_FILE")"
  umask 077
  {
    printf 'ROLE=%q\n' "$ROLE"
    printf 'MASTER=%q\n' "$MASTER"
    printf 'CREDENTIAL=%q\n' "$CREDENTIAL"
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

post_json() {
  local endpoint="$1" body="$2"
  curl -fsSL --connect-timeout 8 --max-time 20 \
    -H 'Content-Type: application/json' \
    --data-binary "$body" \
    "$MASTER$endpoint"
}

proxy_registration() {
  local host_key body
  ssh-keygen -A >/dev/null 2>&1 || true
  host_key=$(awk '{print $1" "$2}' /etc/ssh/ssh_host_ed25519_key.pub)
  body=$(jq -nc \
    --arg secret "$CREDENTIAL" \
    --arg ssh_host_key "$host_key" \
    '{secret:$secret,ssh_host_key:$ssh_host_key}')
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
    printf 'restrict,port-forwarding,permitopen="127.0.0.1:80",permitopen="127.0.0.1:443",command="/bin/sleep 2147483647" %s\n' "$public_key"
  done < <(jq -r '.peers[]?.ssh_public_key // empty' <<<"$response" | sort -u) >"$temporary"
  install -m 600 -o prism-tunnel -g prism-tunnel "$temporary" "$target"
  rm -f "$temporary"
  sshd -t
}

apply_proxy_authorization() {
  local response="$1" temporary
  local -a clients4=()
  while IFS= read -r address; do
    [[ -n "$address" ]] || continue
    if [[ "$address" != *:* ]]; then
      clients4+=("$address")
    fi
  done < <(jq -r '.authorized_ips[]?' <<<"$response" | sort -u)
  join_csv() { local IFS=,; printf '%s' "$*"; }
  temporary=$(mktemp)
  {
    echo 'table inet prism_authorization {'
    printf '  set clients4 { type ipv4_addr; flags timeout; timeout 30s;'
    ((${#clients4[@]} > 0)) && printf ' elements = { %s };' "$(join_csv "${clients4[@]}")"
    echo ' }'
    echo '  chain input {'
    echo '    type filter hook input priority -25; policy accept;'
    echo '    iifname "lo" tcp dport { 53, 80, 443 } accept'
    echo '    iifname "lo" udp dport 53 accept'
    echo '    ip saddr @clients4 tcp dport { 53, 80, 443 } accept'
    echo '    ip saddr @clients4 udp dport 53 accept'
    echo '    meta nfproto ipv4 tcp dport { 53, 80, 443 } drop'
    echo '    meta nfproto ipv4 udp dport 53 drop'
    echo '    meta nfproto ipv6 tcp dport { 53, 80, 443 } drop'
    echo '    meta nfproto ipv6 udp dport 53 drop'
    echo '  }'
    echo '}'
  } >"$temporary"
  nft delete table inet prism_authorization >/dev/null 2>&1 || true
  nft -f "$temporary"
  rm -f "$temporary"
}

sync_proxy() {
  local response
  response=$(proxy_registration)
  jq -e '.role == "proxy" and (.peers | type == "array") and (.authorized_ips | type == "array")' <<<"$response" >/dev/null
  apply_authorized_keys "$response"
  apply_proxy_authorization "$response"
  log "proxy transport synchronized: $(jq '.authorized_ips | length' <<<"$response") authorized IPs, $(jq '.peers | length' <<<"$response") encrypted clients"
}

service_id_for_proxy() {
  printf '%s' "$1" | sha256sum | cut -c1-12
}

write_tunnel_service() {
  local peer="$1" proxy_id proxy_ip ssh_host ssh_port host_key service_id service_path known_hosts temporary port_seed local_http local_https
  proxy_id=$(jq -r '.proxy_id' <<<"$peer")
  proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
  ssh_host=$(jq -r '.ssh_host' <<<"$peer")
  ssh_port=$(jq -r '.ssh_port' <<<"$peer")
  host_key=$(jq -r '.ssh_host_key' <<<"$peer")
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
ExecStart=/usr/bin/ssh -N -T -l prism-tunnel -p $ssh_port -i $SSH_KEY_FILE -o IdentitiesOnly=yes -o BatchMode=yes -o ExitOnForwardFailure=yes -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -o StrictHostKeyChecking=yes -o UserKnownHostsFile=$known_hosts -o HostKeyAlias=prism-$service_id -L 127.0.0.1:$local_http:127.0.0.1:80 -L 127.0.0.1:$local_https:127.0.0.1:443 $ssh_host
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
    '{proxy_id:$proxy_id,proxy_ip:$proxy_ip,service_id:$service_id,local_http:$local_http,local_https:$local_https}'
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
    systemctl disable --now "prism-transport-ssh-$service_id.service" >/dev/null 2>&1 || true
    rm -f "/etc/systemd/system/prism-transport-ssh-$service_id.service" "$STATE_DIR/known-hosts-$service_id"
  done < <(jq -r '.[].proxy_id' "$ACTIVE_FILE")
}

apply_redirect_rules() {
  local current="$1" temporary peer proxy_ip local_http local_https
  temporary=$(mktemp)
  {
    echo 'table inet prism_transport {'
    echo '  counter tx {}'
    echo '  counter rx {}'
    echo '  chain account_output {'
    echo '    type filter hook output priority -200; policy accept;'
    while IFS= read -r peer; do
      [[ -n "$peer" ]] || continue
      local_http=$(jq -r '.local_http' <<<"$peer")
      local_https=$(jq -r '.local_https' <<<"$peer")
      printf '    tcp sport { %s, %s } counter name "rx" return\n' "$local_http" "$local_https"
    done < <(jq -c 'unique_by(.proxy_ip)[]?' <<<"$current")
    while IFS= read -r peer; do
      [[ -n "$peer" ]] || continue
      proxy_ip=$(jq -r '.proxy_ip' <<<"$peer")
      printf '    ip daddr %s counter name "tx"\n' "$proxy_ip"
    done < <(jq -c 'unique_by(.proxy_ip)[]?' <<<"$current")
    echo '  }'
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
    echo '}'
  } >"$temporary"
  nft delete table inet prism_transport >/dev/null 2>&1 || true
  nft -f "$temporary"
  rm -f "$temporary"
}

tunnel_port_ready() {
  local proxy_ip="$1" port="$2"
  timeout 4 bash -c \
    'exec 3<>"/dev/tcp/$1/$2"; exec 3>&-; exec 3<&-' \
    prism-transport-probe "$proxy_ip" "$port" >/dev/null 2>&1
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

sync_client() {
  local response peer proxy_id ready_json old_ready="[]" old_current="[]" current_json changed_ready=false changed_current=false
  local -a ready=() current=()
  response=$(client_registration)
  jq -e '.role == "client" and (.peers | type == "array")' <<<"$response" >/dev/null
  while IFS= read -r peer; do
    [[ -n "$peer" ]] || continue
    current+=("$(write_tunnel_service "$peer")")
  done < <(jq -c '.peers[]?' <<<"$response")
  if ((${#current[@]})); then
    current_json=$(printf '%s\n' "${current[@]}" | jq -sc 'sort_by(.proxy_id)')
  else
    current_json='[]'
  fi
  [[ -f "$ACTIVE_FILE" ]] && old_current=$(jq -c 'sort_by(.proxy_id)' "$ACTIVE_FILE" 2>/dev/null || echo '[]')
  [[ "$(jq -c 'sort_by(.proxy_id)' <<<"$current_json")" != "$old_current" ]] && changed_current=true
  remove_stale_tunnels "$current_json"
  if $changed_current ||
    ! nft list counter inet prism_transport tx >/dev/null 2>&1 ||
    ! nft list counter inet prism_transport rx >/dev/null 2>&1; then
    apply_redirect_rules "$current_json"
  fi
  printf '%s\n' "$current_json" >"$ACTIVE_FILE"
  chmod 600 "$ACTIVE_FILE"
  sleep 2
  while IFS= read -r peer; do
    [[ -n "$peer" ]] || continue
    proxy_id=$(jq -r '.proxy_id' <<<"$peer")
    if tunnel_ready "$peer"; then
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
  log "client transport synchronized: ${#ready[@]}/$(jq '.peers | length' <<<"$response") encrypted peers ready"
}

sync_transport() {
  load_environment
  [[ "$ROLE" == "proxy" || "$ROLE" == "client" ]] || fail "transport role is not configured"
  [[ "$MASTER" =~ ^https?://[^[:space:]]+$ ]] || fail "invalid panel URL"
  [[ -n "$CREDENTIAL" ]] || fail "transport credential is missing"
  ensure_dependencies
  if [[ "$ROLE" == "proxy" ]]; then
    sync_proxy
  else
    ensure_client_keypair
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
ExecStart=$INSTALL_PATH --sync
EOF
  cat >/etc/systemd/system/prism-transport.timer <<'EOF'
[Unit]
Description=Refresh Prism encrypted SNI transport

[Timer]
OnBootSec=5s
OnUnitActiveSec=5s
AccuracySec=1s
Persistent=true

[Install]
WantedBy=timers.target
EOF
  systemctl daemon-reload
  systemctl enable --now prism-transport.timer >/dev/null
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
  [[ "$ROLE" == "client" ]] && ensure_client_keypair
  install_timer
  sync_transport
  log "encrypted TCP transport v$VERSION installed for role $ROLE"
}

uninstall_transport() {
  local service_id proxy_ip
  load_environment
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
  systemctl disable --now wg-quick@prismwg0 >/dev/null 2>&1 || true
  rm -f /etc/systemd/system/prism-transport.service /etc/systemd/system/prism-transport.timer
  rm -f /etc/wireguard/prismwg0.conf "$ENV_FILE"
  nft delete table inet prism_transport >/dev/null 2>&1 || true
  nft delete table inet prism_authorization >/dev/null 2>&1 || true
  systemctl daemon-reload
  log "encrypted transport removed"
}

show_status() {
  load_environment
  echo "role=${ROLE:-unconfigured}"
  echo "timer=$(systemctl is-active prism-transport.timer 2>/dev/null || true)"
  if [[ -f "$READY_FILE" ]]; then
    echo "ready_proxies=$(jq -c . "$READY_FILE" 2>/dev/null || echo '[]')"
  fi
  systemctl --no-pager --plain list-units 'prism-transport-ssh-*.service' 2>/dev/null || true
}

main() {
  require_root
  parse_args "$@"
  exec 9>/run/prism-transport.lock
  flock -w 30 9 || fail "another transport synchronization is still running"
  case "$ACTION" in
    sync) sync_transport ;;
    uninstall) uninstall_transport ;;
    status) show_status ;;
    install) install_transport ;;
  esac
}

main "$@"
