#!/usr/bin/env bash

set -Eeuo pipefail

VERSION="1.3.3"
STATE_DIR="/var/lib/prismdns"
BACKUP_DIR="$STATE_DIR/backups"
CONFIG_FILE="$STATE_DIR/client.conf"
BOOTSTRAP_FILE="$STATE_DIR/bootstrap.json"
TEST_DOMAIN="www.google.com"
MASTER=""
TOKEN=""
NON_INTERACTIVE=false
ACTION="menu"

if [[ -t 1 ]]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[0;33m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
else
  RED=''; GREEN=''; YELLOW=''; CYAN=''; BOLD=''; NC=''
fi

info() { printf '%b\n' "${CYAN}[信息]${NC} $*"; }
ok() { printf '%b\n' "${GREEN}[成功]${NC} $*"; }
warn() { printf '%b\n' "${YELLOW}[注意]${NC} $*"; }
fail() { printf '%b\n' "${RED}[错误]${NC} $*" >&2; exit 1; }

prompt() {
  local message="$1" value=""
  if [[ -r /dev/tty ]]; then
    read -r -p "$message" value </dev/tty
  else
    read -r -p "$message" value
  fi
  printf '%s' "$value"
}

confirm() {
  $NON_INTERACTIVE && return 0
  local answer
  answer=$(prompt "$1 [y/N]: ")
  [[ "$answer" =~ ^[Yy]$ ]]
}

require_root() {
  [[ ${EUID:-$(id -u)} -eq 0 ]] || fail "此操作需要 root 权限，请使用 sudo 运行。"
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
  local missing=()
  command -v curl >/dev/null 2>&1 || missing+=(curl)
  command -v jq >/dev/null 2>&1 || missing+=(jq)
  command -v dig >/dev/null 2>&1 || missing+=(dig)
  command -v nft >/dev/null 2>&1 || missing+=(nftables)
  ((${#missing[@]} == 0)) && return 0
  require_root
  local manager
  manager=$(detect_package_manager)
  info "安装依赖: ${missing[*]}"
  case "$manager" in
    apt) apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl jq dnsutils nftables ;;
    dnf) dnf install -y curl jq bind-utils nftables ;;
    yum) yum install -y curl jq bind-utils nftables ;;
    apk) apk add --no-cache curl jq bind-tools nftables ;;
    *) fail "无法识别包管理器，请手动安装 curl、jq、dig 和 nftables。" ;;
  esac
}

load_config() {
  [[ -f "$CONFIG_FILE" ]] || return 0
  [[ -z "$MASTER" ]] && MASTER=$(sed -n 's/^master=//p' "$CONFIG_FILE" | head -1)
  [[ -z "$TOKEN" ]] && TOKEN=$(sed -n 's/^token=//p' "$CONFIG_FILE" | head -1)
  return 0
}

save_config() {
  mkdir -p "$STATE_DIR"
  umask 077
  printf 'master=%s\ntoken=%s\n' "$MASTER" "$TOKEN" > "$CONFIG_FILE"
  chmod 600 "$CONFIG_FILE"
}

validate_master() {
  MASTER="${MASTER%/}"
  [[ "$MASTER" =~ ^https?://[^[:space:]]+$ ]] || fail "Controller 地址无效: $MASTER"
  [[ "$TOKEN" =~ ^[a-fA-F0-9]{32,}$ ]] || fail "配置令牌无效。"
}

bootstrap() {
  ensure_dependencies
  load_config
  [[ -n "$MASTER" ]] || MASTER=$(prompt "请输入 Prism DNS 面板地址: ")
  [[ -n "$TOKEN" ]] || TOKEN=$(prompt "请输入 IP 配置令牌: ")
  validate_master
  local response
  response=$(curl -fsSL --connect-timeout 8 --max-time 20 "$MASTER/enhancer/api/bootstrap/$TOKEN") || fail "无法从面板获取配置，请检查地址和令牌。"
  local expected detected secret smart installer configured_master
  expected=$(jq -r '.expected_ip // empty' <<<"$response")
  detected=$(jq -r '.detected_ip // empty' <<<"$response")
  secret=$(jq -r '.secret // empty' <<<"$response")
  smart=$(jq -r '.smart // true' <<<"$response")
  installer=$(jq -r '.agent_installer // empty' <<<"$response")
  configured_master=$(jq -r '.master // empty' <<<"$response")
  [[ -n "$secret" && -n "$installer" ]] || fail "面板返回的节点配置不完整。"
  if [[ -n "$expected" && -n "$detected" && "$expected" != "$detected" ]]; then
    warn "控制台配置 IP 为 $expected，但面板检测到当前出口 IP 为 $detected。"
    confirm "仍要继续安装吗？" || return 1
  fi
  [[ -n "$configured_master" ]] && MASTER="${configured_master%/}"
  save_config
  umask 077
  printf '%s\n' "$response" > "$BOOTSTRAP_FILE"
  chmod 600 "$BOOTSTRAP_FILE"
  info "安装 Prism DNS Client Agent..."
  local args=(--master "$MASTER" --secret "$secret")
  [[ "$smart" == "true" ]] && args+=(--smart)
  if [[ -n "${PRISM_AGENT_INSTALLER_FILE:-}" ]]; then
    bash "$PRISM_AGENT_INSTALLER_FILE" "${args[@]}"
  else
    curl -fsSL "$installer" | bash -s -- "${args[@]}"
  fi
  install_traffic_reporter
  wait_for_local_dns
}

install_traffic_reporter() {
  require_root
  mkdir -p /usr/local/lib/prismdns
  cat > /usr/local/lib/prismdns/report-traffic.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
CONFIG_FILE="/var/lib/prismdns/client.conf"
PEER_HASH_FILE="/var/lib/prismdns/traffic-peers.sha256"
AUDIT_HASH_FILE="/var/lib/prismdns/service-audit.sha256"
AUDIT_TIME_FILE="/var/lib/prismdns/service-audit.timestamp"
AUDIT_OUTPUT_FILE="/var/lib/prismdns/service-audit.txt"
NFT_TABLE="prismdns_traffic"
COUNTER_VERSION="dns-sni-ports-v4"
AUDIT_INTERVAL=21600
[[ -f "$CONFIG_FILE" ]] || exit 0
MASTER=$(sed -n 's/^master=//p' "$CONFIG_FILE" | head -1)
TOKEN=$(sed -n 's/^token=//p' "$CONFIG_FILE" | head -1)
[[ -n "$MASTER" && -n "$TOKEN" ]] || exit 0
if ! BOOTSTRAP=$(curl -fsSL --connect-timeout 8 --max-time 15 "$MASTER/enhancer/api/bootstrap/$TOKEN"); then
  exit 0
fi
mapfile -t PEERS < <(jq -r '.traffic_peers[]?' <<<"$BOOTSTRAP" | sort -u)
mapfile -t PROBES < <(jq -r '.health_probes[]?.domain // empty' <<<"$BOOTSTRAP" | sort -u)
((${#PEERS[@]} > 0)) || exit 0
PEER_HASH=$({ printf '%s\n' "$COUNTER_VERSION"; printf '%s\n' "${PEERS[@]}"; } | sha256sum | awk '{print $1}')
CURRENT_HASH=$(cat "$PEER_HASH_FILE" 2>/dev/null || true)
if [[ "$PEER_HASH" != "$CURRENT_HASH" ]] || ! nft list table inet "$NFT_TABLE" >/dev/null 2>&1; then
  PEERS4=()
  PEERS6=()
  for peer in "${PEERS[@]}"; do
    if [[ "$peer" == *:* ]]; then PEERS6+=("$peer"); else PEERS4+=("$peer"); fi
  done
  join_csv() { local IFS=,; printf '%s' "$*"; }
  RULE_FILE=$(mktemp /var/lib/prismdns/traffic-rules.XXXXXX)
  {
    printf 'table inet %s {\n' "$NFT_TABLE"
    printf '  counter rx {}\n  counter tx {}\n'
    printf '  set peers4 { type ipv4_addr;'
    ((${#PEERS4[@]} > 0)) && printf ' elements = { %s };' "$(join_csv "${PEERS4[@]}")"
    printf ' }\n'
    printf '  set peers6 { type ipv6_addr;'
    ((${#PEERS6[@]} > 0)) && printf ' elements = { %s };' "$(join_csv "${PEERS6[@]}")"
    printf ' }\n'
    printf '  chain input { type filter hook input priority -10; policy accept; tcp dport 53 counter name rx; udp dport 53 counter name rx; ip saddr @peers4 tcp sport { 80, 443 } counter name rx; ip6 saddr @peers6 tcp sport { 80, 443 } counter name rx; }\n'
    printf '  chain output { type filter hook output priority -10; policy accept; tcp sport 53 counter name tx; udp sport 53 counter name tx; ip daddr @peers4 tcp dport { 80, 443 } counter name tx; ip6 daddr @peers6 tcp dport { 80, 443 } counter name tx; }\n'
    printf '}\n'
  } > "$RULE_FILE"
  nft delete table inet "$NFT_TABLE" >/dev/null 2>&1 || true
  nft -f "$RULE_FILE"
  rm -f "$RULE_FILE"
  printf '%s\n' "$PEER_HASH" > "$PEER_HASH_FILE"
fi
RX=$(nft -j list counter inet "$NFT_TABLE" rx | jq '[.nftables[].counter? | .bytes] | add // 0')
TX=$(nft -j list counter inet "$NFT_TABLE" tx | jq '[.nftables[].counter? | .bytes] | add // 0')
DNS_READY=false
SYSTEM_DNS_READY=false
ROUTES_READY=false
HEALTHY_ROUTES=0
EXPECTED_ROUTES=${#PROBES[@]}
HEALTH_MESSAGE=""
if systemctl is-active --quiet prism-agent 2>/dev/null && ss -lntup 2>/dev/null | awk '$5 ~ /:53$/ && /prism-agent/ {found=1} END {exit !found}'; then
  DNS_READY=true
else
  HEALTH_MESSAGE="Prism Agent 未接管 53 端口"
fi
if awk '$1 == "nameserver" && ($2 == "127.0.0.1" || $2 == "::1") {found=1} END {exit !found}' /etc/resolv.conf 2>/dev/null; then
  SYSTEM_DNS_READY=true
else
  HEALTH_MESSAGE="${HEALTH_MESSAGE:+$HEALTH_MESSAGE；}系统 DNS 未使用 Prism"
fi
for domain in "${PROBES[@]}"; do
  mapfile -t ANSWERS < <({ dig @127.0.0.1 "$domain" A +short +time=2 +tries=1; dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1; } 2>/dev/null | sed '/^$/d' | sort -u)
  matched=false
  for answer in "${ANSWERS[@]}"; do
    for peer in "${PEERS[@]}"; do
      if [[ "$answer" == "$peer" ]]; then
        matched=true
        break 2
      fi
    done
  done
  if $matched; then
    HEALTHY_ROUTES=$((HEALTHY_ROUTES + 1))
  fi
done
if ((EXPECTED_ROUTES == HEALTHY_ROUTES)) && $DNS_READY && $SYSTEM_DNS_READY; then
  ROUTES_READY=true
else
  HEALTH_MESSAGE="${HEALTH_MESSAGE:+$HEALTH_MESSAGE；}路由探针 ${HEALTHY_ROUTES}/${EXPECTED_ROUTES}"
fi
PAYLOAD=$(jq -nc \
  --arg token "$TOKEN" --argjson rx "$RX" --argjson tx "$TX" \
  --argjson dns_ready "$DNS_READY" --argjson system_dns_ready "$SYSTEM_DNS_READY" --argjson routes_ready "$ROUTES_READY" \
  --argjson healthy_routes "$HEALTHY_ROUTES" --argjson expected_routes "$EXPECTED_ROUTES" --arg health_message "$HEALTH_MESSAGE" \
  '{token:$token,scope:"unlock_peers",interface:"nftables-dns-sni",rx_bytes:$rx,tx_bytes:$tx,dns_ready:$dns_ready,system_dns_ready:$system_dns_ready,routes_ready:$routes_ready,healthy_routes:$healthy_routes,expected_routes:$expected_routes,health_message:$health_message}')
curl -fsSL --connect-timeout 8 --max-time 15 -H 'Content-Type: application/json' -d "$PAYLOAD" "$MASTER/enhancer/api/traffic/report" >/dev/null || exit 0

run_service_audit() {
  local audit_hash="$1" now="$2" output plain results probe service_id provider domain result matched_peer answer peer
  if [[ ! -x /usr/bin/ut ]]; then
    timeout 180 bash -c 'curl -sL https://raw.githubusercontent.com/oneclickvirt/UnlockTests/main/ut_install.sh -sSf | bash' >/dev/null 2>&1 || return 0
  fi
  output=$(timeout 300 /usr/bin/ut -m 4 -f 20 -b=false -s=false 2>&1 || true)
  [[ -n "$output" ]] || return 0
  plain=$(sed $'s/\033\\[[0-9;]*[mK]//g' <<<"$output" | tr -d '\r')
  printf '%s\n' "$plain" > "$AUDIT_OUTPUT_FILE"
  results='{}'
  while IFS= read -r probe; do
    service_id=$(jq -r '.service_id // empty' <<<"$probe")
    provider=$(jq -r '.unlock_test // empty' <<<"$probe")
    domain=$(jq -r '.domain // empty' <<<"$probe")
    [[ -n "$service_id" ]] || continue
    result=""
    if [[ -n "$provider" ]]; then
      result=$(awk -v provider="$provider" '
        index($0, provider) == 1 && substr($0, length(provider) + 1, 1) ~ /[[:space:]]/ {
          value = substr($0, length(provider) + 1)
          sub(/^[[:space:]]+/, "", value)
          print value
          exit
        }
      ' <<<"$plain")
    fi
    if [[ -z "$result" && -n "$domain" ]]; then
      matched_peer=""
      while IFS= read -r answer; do
        for peer in "${PEERS[@]}"; do
          if [[ "$answer" == "$peer" ]]; then
            matched_peer="$peer"
            break 2
          fi
        done
      done < <({ dig @127.0.0.1 "$domain" A +short +time=2 +tries=1; dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1; } 2>/dev/null | sed '/^$/d' | sort -u)
      if [[ -z "$matched_peer" ]]; then
        result="FAIL (DNS route mismatch)"
      elif curl -sS -o /dev/null --resolve "$domain:443:$matched_peer" --connect-timeout 5 --max-time 15 "https://$domain/"; then
        result="PASS (HTTPS reachable) [Via DNS]"
      else
        result="FAIL (HTTPS unavailable)"
      fi
    fi
    [[ -n "$result" ]] || result="N/A (No UnlockTests result)"
    results=$(jq -c --arg id "$service_id" --arg result "$result" '. + {($id):$result}' <<<"$results")
  done < <(jq -c '.health_probes[]?' <<<"$BOOTSTRAP")
  if [[ "$(jq 'length' <<<"$results")" -gt 0 ]]; then
    jq -nc --arg token "$TOKEN" --argjson results "$results" '{token:$token,scope:"unlock_services",results:$results}' |
      curl -fsSL --connect-timeout 8 --max-time 20 -H 'Content-Type: application/json' -d @- "$MASTER/enhancer/api/audit/report" >/dev/null || return 0
    printf '%s\n' "$audit_hash" > "$AUDIT_HASH_FILE"
    printf '%s\n' "$now" > "$AUDIT_TIME_FILE"
  fi
  nft reset counter inet "$NFT_TABLE" rx >/dev/null 2>&1 || true
  nft reset counter inet "$NFT_TABLE" tx >/dev/null 2>&1 || true
}

AUDIT_HASH=$(jq -Sc '[.health_probes[]? | {service_id,unlock_test,domain}]' <<<"$BOOTSTRAP" | sha256sum | awk '{print $1}')
LAST_AUDIT_HASH=$(cat "$AUDIT_HASH_FILE" 2>/dev/null || true)
LAST_AUDIT_TIME=$(cat "$AUDIT_TIME_FILE" 2>/dev/null || echo 0)
NOW=$(date +%s)
if [[ "$AUDIT_HASH" != "$LAST_AUDIT_HASH" ]] || ((NOW - LAST_AUDIT_TIME >= AUDIT_INTERVAL)); then
  run_service_audit "$AUDIT_HASH" "$NOW"
fi
EOF
  chmod 700 /usr/local/lib/prismdns/report-traffic.sh
  cat > /usr/local/lib/prismdns/sync-routes.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
CONFIG_FILE="/var/lib/prismdns/client.conf"
HASH_FILE="/var/lib/prismdns/route-config.sha256"
RESTART_FILE="/var/lib/prismdns/route-restart.timestamp"
LOCK_FILE="/run/prismdns-route-sync.lock"
[[ -f "$CONFIG_FILE" ]] || exit 0
MASTER=$(sed -n 's/^master=//p' "$CONFIG_FILE" | head -1)
TOKEN=$(sed -n 's/^token=//p' "$CONFIG_FILE" | head -1)
[[ -n "$MASTER" && -n "$TOKEN" ]] || exit 0
if command -v flock >/dev/null 2>&1; then
  exec 9>"$LOCK_FILE"
  flock -n 9 || exit 0
else
  LOCK_DIR="${LOCK_FILE}.d"
  mkdir "$LOCK_DIR" 2>/dev/null || exit 0
  trap 'rmdir "$LOCK_DIR" 2>/dev/null || true' EXIT
fi
if ! BOOTSTRAP=$(curl -fsSL --connect-timeout 5 --max-time 10 "$MASTER/enhancer/api/bootstrap/$TOKEN"); then
  exit 0
fi
CURRENT_HASH=$(jq -Sc '{
  smart:(.smart // true),
  traffic_peers:((.traffic_peers // []) | sort),
  health_probes:([.health_probes[]? | {service_id,domain,unlock_test}] | sort_by(.service_id))
}' <<<"$BOOTSTRAP" | sha256sum | awk '{print $1}')
LAST_HASH=$(cat "$HASH_FILE" 2>/dev/null || true)
mapfile -t PEERS < <(jq -r '.traffic_peers[]?' <<<"$BOOTSTRAP" | sort -u)
mapfile -t PROBES < <(jq -r '.health_probes[]?.domain // empty' <<<"$BOOTSTRAP" | sort -u)
route_health_ok() {
  systemctl is-active --quiet prism-agent 2>/dev/null || return 1
  ss -lntup 2>/dev/null | awk '$5 ~ /:53$/ && /prism-agent/ {found=1} END {exit !found}' || return 1
  ((${#PROBES[@]} > 0)) || return 0
  local domain answer peer matched
  for domain in "${PROBES[@]}"; do
    matched=false
    while IFS= read -r answer; do
      for peer in "${PEERS[@]}"; do
        if [[ "$answer" == "$peer" ]]; then
          matched=true
          break 2
        fi
      done
    done < <({ dig @127.0.0.1 "$domain" A +short +time=2 +tries=1; dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1; } 2>/dev/null | sed '/^$/d' | sort -u)
    $matched || return 1
  done
}
if [[ -n "$LAST_HASH" && "$CURRENT_HASH" == "$LAST_HASH" ]] && route_health_ok; then
  exit 0
fi
NOW=$(date +%s)
LAST_RESTART=$(cat "$RESTART_FILE" 2>/dev/null || echo 0)
if ((NOW - LAST_RESTART < 60)); then
  exit 0
fi
printf '%s\n' "$NOW" > "$RESTART_FILE"
systemctl restart prism-agent
for _ in {1..30}; do
  if route_health_ok; then
    printf '%s\n' "$CURRENT_HASH" > "$HASH_FILE"
    exit 0
  fi
  sleep 1
done
exit 1
EOF
  chmod 700 /usr/local/lib/prismdns/sync-routes.sh
  if command -v systemctl >/dev/null 2>&1; then
    cat > /etc/systemd/system/prismdns-traffic.service <<'EOF'
[Unit]
Description=Prism DNS unlock-link traffic report
After=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/lib/prismdns/report-traffic.sh
EOF
    cat > /etc/systemd/system/prismdns-traffic.timer <<'EOF'
[Unit]
Description=Report Prism DNS unlock-link traffic every minute

[Timer]
OnActiveSec=15s
OnUnitActiveSec=1min
AccuracySec=5s
Persistent=true

[Install]
WantedBy=timers.target
EOF
    cat > /etc/systemd/system/prismdns-route-sync.service <<'EOF'
[Unit]
Description=Apply Prism DNS route changes and flush Agent DNS cache
After=network-online.target prism-agent.service

[Service]
Type=oneshot
ExecStart=/usr/local/lib/prismdns/sync-routes.sh
EOF
    cat > /etc/systemd/system/prismdns-route-sync.timer <<'EOF'
[Unit]
Description=Check Prism DNS route changes every 10 seconds

[Timer]
OnBootSec=10s
OnUnitActiveSec=10s
AccuracySec=1s
Persistent=true

[Install]
WantedBy=timers.target
EOF
    systemctl daemon-reload
    systemctl enable --now prismdns-traffic.timer >/dev/null
    systemctl enable --now prismdns-route-sync.timer >/dev/null
  elif [[ -d /etc/cron.d ]]; then
    printf '* * * * * root /usr/local/lib/prismdns/report-traffic.sh\n* * * * * root /usr/local/lib/prismdns/sync-routes.sh\n' > /etc/cron.d/prismdns-traffic
    chmod 644 /etc/cron.d/prismdns-traffic
  else
    warn "系统没有 systemd 或 cron，无法启用定时流量上报。"
  fi
  /usr/local/lib/prismdns/sync-routes.sh || warn "首次路由同步守卫初始化失败，定时任务会稍后重试。"
  /usr/local/lib/prismdns/report-traffic.sh || warn "首次流量上报失败，定时任务会稍后重试。"
  ok "每 IP 解锁链路流量统计已启用（本机 Prism DNS 的 UDP/TCP 53，加上目标服务器与已选解锁机之间 TCP 80/443；每分钟上报）。"
}

wait_for_local_dns() {
  local attempt
  info "等待 Prism Agent 接管本机 53 端口..."
  for attempt in {1..30}; do
    if systemctl is-active --quiet prism-agent 2>/dev/null &&
       ss -lntup 2>/dev/null | awk '$5 ~ /:53$/ && /prism-agent/ {found=1} END {exit !found}' &&
       dig @127.0.0.1 "$TEST_DOMAIN" +time=1 +tries=1 +short 2>/dev/null | grep -q .; then
      ok "Prism Agent 已接管本机 DNS。"
      verify_route_probes
      return 0
    fi
    sleep 1
  done
  journalctl -u prism-agent -n 20 --no-pager >&2 2>/dev/null || true
  fail "Prism Agent 未能接管 53 端口；不能把其他 DNS 服务的响应误判为安装成功。"
}

verify_route_probes() {
  local bootstrap domain answer peer matched_peer expected=0 healthy=0
  local -a peers
  bootstrap=$(curl -fsSL --connect-timeout 8 --max-time 15 "$MASTER/enhancer/api/bootstrap/$TOKEN")
  mapfile -t peers < <(jq -r '.traffic_peers[]?' <<<"$bootstrap" | sort -u)
  while IFS= read -r domain; do
    [[ -n "$domain" ]] || continue
    expected=$((expected + 1))
    matched=false
    matched_peer=""
    while IFS= read -r answer; do
      for peer in "${peers[@]}"; do
        if [[ "$answer" == "$peer" ]]; then
          matched=true
          matched_peer="$peer"
          break 2
        fi
      done
    done < <({ dig @127.0.0.1 "$domain" A +short +time=2 +tries=1; dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1; } 2>/dev/null | sed '/^$/d' | sort -u)
    probe_ok=false
    for probe_attempt in 1 2; do
      if $matched && curl -sS -o /dev/null --resolve "$domain:443:$matched_peer" --connect-timeout 5 --max-time 12 "https://$domain/"; then
        probe_ok=true
        break
      fi
      sleep 1
    done
    if $probe_ok; then
      healthy=$((healthy + 1))
    fi
  done < <(jq -r '.health_probes[]?.domain // empty' <<<"$bootstrap" | sort -u)
  ((expected == healthy)) || fail "所选服务端到端验证失败：$healthy/$expected 同时通过 DNS 映射与 HTTPS。"
  ok "所选服务端到端验证通过：$healthy/$expected。"
}

test_dns() {
  ensure_dependencies
  printf '\n%b\n' "${BOLD}Prism DNS 连通性测试${NC}"
  local started elapsed result transport="UDP"
  started=$(date +%s%3N 2>/dev/null || date +%s000)
  result=$(dig @127.0.0.1 "$TEST_DOMAIN" +time=2 +tries=1 +short 2>/dev/null || true)
  if [[ -z "$result" ]]; then
    transport="TCP"
    result=$(dig @127.0.0.1 "$TEST_DOMAIN" +tcp +time=2 +tries=1 +short 2>/dev/null || true)
  fi
  [[ -n "$result" ]] || fail "UDP/TCP DNS 查询均失败。"
  elapsed=$(( $(date +%s%3N 2>/dev/null || date +%s000) - started ))
  ok "$transport 查询成功，耗时 ${elapsed}ms"
  printf '  域名: %s\n  结果: %s\n' "$TEST_DOMAIN" "$(tr '\n' ' ' <<<"$result")"
}

backup_dns() {
  require_root
  mkdir -p "$BACKUP_DIR"
  local timestamp path
  timestamp=$(date +%Y%m%d-%H%M%S)
  path="$BACKUP_DIR/$timestamp"
  mkdir -p "$path"
  chmod 700 "$path"
  if [[ -L /etc/resolv.conf ]]; then
    readlink /etc/resolv.conf > "$path/resolv.conf.symlink"
    cp -aL /etc/resolv.conf "$path/resolv.conf" 2>/dev/null || true
  elif [[ -e /etc/resolv.conf ]]; then
    cp -a /etc/resolv.conf "$path/resolv.conf"
  fi
  [[ -f /etc/systemd/resolved.conf ]] && cp -a /etc/systemd/resolved.conf "$path/resolved.conf"
  [[ -f /etc/NetworkManager/NetworkManager.conf ]] && cp -a /etc/NetworkManager/NetworkManager.conf "$path/NetworkManager.conf"
  if command -v systemctl >/dev/null 2>&1; then
    systemctl is-active --quiet systemd-resolved 2>/dev/null && echo yes > "$path/resolved.active" || true
    systemctl is-enabled --quiet systemd-resolved 2>/dev/null && echo yes > "$path/resolved.enabled" || true
  fi
  ok "DNS 配置已备份到 $path"
  printf '%s' "$path"
}

write_resolv_conf() {
  local dns="$1" temporary
  temporary=$(mktemp /etc/resolv.conf.prismdns.XXXXXX)
  printf '# Managed by Prism DNS\nnameserver %s\noptions timeout:2 attempts:2\n' "$dns" > "$temporary"
  chmod 644 "$temporary"
  chattr -i /etc/resolv.conf 2>/dev/null || true
  [[ -L /etc/resolv.conf ]] && rm -f /etc/resolv.conf
  mv -f "$temporary" /etc/resolv.conf
}

apply_permanent() {
  require_root
  test_dns
  confirm "将系统 DNS 永久接管为 127.0.0.1，并自动备份原配置，继续吗？" || return 0
  backup_dns >/dev/null
  if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet systemd-resolved 2>/dev/null; then
    systemctl disable --now systemd-resolved 2>/dev/null || warn "无法停用 systemd-resolved，将继续写入 resolv.conf。"
  fi
  if [[ -d /etc/NetworkManager/conf.d ]]; then
    cat > /etc/NetworkManager/conf.d/prismdns.conf <<'EOF'
[main]
dns=none
EOF
    systemctl reload NetworkManager 2>/dev/null || true
  fi
  write_resolv_conf 127.0.0.1
  test_dns
  grep -Eq '^nameserver[[:space:]]+127\.0\.0\.1([[:space:]]|$)' /etc/resolv.conf || fail "系统 DNS 未成功切换到 127.0.0.1。"
  getent ahosts "$TEST_DOMAIN" >/dev/null || fail "系统解析器未能通过 Prism DNS 完成查询。"
  ok "系统 DNS 已永久设置为 Prism DNS。"
}

apply_temporary() {
  require_root
  test_dns
  backup_dns >/dev/null
  if command -v resolvectl >/dev/null 2>&1 && systemctl is-active --quiet systemd-resolved 2>/dev/null; then
    local iface
    iface=$(ip route show default 2>/dev/null | awk 'NR==1 {print $5}')
    [[ -n "$iface" ]] || fail "无法识别默认网络接口。"
    resolvectl dns "$iface" 127.0.0.1
    resolvectl domain "$iface" '~.'
  else
    write_resolv_conf 127.0.0.1
  fi
  test_dns
  ok "已临时使用 Prism DNS。"
}

restore_dns() {
  require_root
  [[ -d "$BACKUP_DIR" ]] || fail "没有可用备份。"
  local path
  path=$(find "$BACKUP_DIR" -mindepth 1 -maxdepth 1 -type d | sort -r | head -1)
  [[ -n "$path" ]] || fail "没有可用备份。"
  confirm "从 $(basename "$path") 恢复 DNS 配置吗？" || return 0
  rm -f /etc/NetworkManager/conf.d/prismdns.conf 2>/dev/null || true
  if [[ -f "$path/resolv.conf.symlink" ]]; then
    rm -f /etc/resolv.conf
    ln -s "$(cat "$path/resolv.conf.symlink")" /etc/resolv.conf
  elif [[ -f "$path/resolv.conf" ]]; then
    rm -f /etc/resolv.conf
    cp -a "$path/resolv.conf" /etc/resolv.conf
  fi
  [[ -f "$path/resolved.conf" ]] && cp -a "$path/resolved.conf" /etc/systemd/resolved.conf
  [[ -f "$path/NetworkManager.conf" ]] && cp -a "$path/NetworkManager.conf" /etc/NetworkManager/NetworkManager.conf
  if command -v systemctl >/dev/null 2>&1; then
    [[ -f "$path/resolved.enabled" ]] && systemctl enable systemd-resolved 2>/dev/null || true
    [[ -f "$path/resolved.active" ]] && systemctl restart systemd-resolved 2>/dev/null || true
    systemctl reload NetworkManager 2>/dev/null || true
  fi
  ok "DNS 配置已恢复。"
}

show_status() {
  load_config
  local backup_count=0
  if [[ -d "$BACKUP_DIR" ]]; then
    backup_count=$(find "$BACKUP_DIR" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l || true)
  fi
  printf '\n%b\n' "${BOLD}Prism DNS 系统状态${NC}"
  printf '  版本             : %s\n' "$VERSION"
  printf '  面板             : %s\n' "${MASTER:-未配置}"
  printf '  Agent 服务       : '
  if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet prism-agent 2>/dev/null; then echo "运行中"; else echo "未运行"; fi
  printf '  当前 DNS         : %s\n' "$(awk '/^nameserver/ {printf "%s ", $2}' /etc/resolv.conf 2>/dev/null || true)"
  printf '  本机 53 端口     : '
  if command -v ss >/dev/null 2>&1 && ss -lntu 2>/dev/null | grep -qE '[:.]53[[:space:]]'; then echo "监听中"; else echo "未监听"; fi
  printf '  备份数量         : %s\n' "$backup_count"
  printf '  解锁链路流量上报 : '
  if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet prismdns-traffic.timer 2>/dev/null; then echo "每分钟"; elif [[ -f /etc/cron.d/prismdns-traffic ]]; then echo "每分钟 (cron)"; else echo "未启用"; fi
  printf '  路由热更新守卫     : '
  if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet prismdns-route-sync.timer 2>/dev/null; then echo "每 10 秒"; elif [[ -f /etc/cron.d/prismdns-traffic ]]; then echo "每分钟 (cron)"; else echo "未启用"; fi
  echo ""
}

one_click() {
  require_root
  bootstrap
  test_dns
  apply_permanent
}

show_menu() {
  clear 2>/dev/null || true
  printf '%b\n' "${CYAN}${BOLD}PRISM DNS${NC} v$VERSION - DNS Client 管理工具"
  printf '%s\n' "================================================"
  show_status
  printf '%s\n' "推荐流程: 控制台添加 IP 并选择服务 -> 执行 4 一键安装并应用"
  printf '\n%s\n' "请选择操作:"
  printf '  %b\n' "${GREEN}1)${NC} 安装 / 连接 DNS Client Agent"
  printf '  %b\n' "${GREEN}2)${NC} DNS 连通性测试"
  printf '  %b\n' "${GREEN}3)${NC} 永久设为系统 DNS"
  printf '  %b\n' "${GREEN}4)${NC} 一键安装、测试并应用"
  printf '  %b\n' "${GREEN}5)${NC} 临时设为系统 DNS"
  printf '  %b\n' "${GREEN}6)${NC} 备份当前 DNS 配置"
  printf '  %b\n' "${GREEN}7)${NC} 恢复最近 DNS 备份"
  printf '  %b\n' "${GREEN}8)${NC} 查看系统状态"
  printf '  %b\n' "${GREEN}0)${NC} 退出"
}

parse_args() {
  while (($#)); do
    case "$1" in
      --master) MASTER="${2:-}"; shift 2 ;;
      --token) TOKEN="${2:-}"; shift 2 ;;
      --non-interactive) NON_INTERACTIVE=true; shift ;;
      --install) ACTION="install"; shift ;;
      --apply) ACTION="apply"; shift ;;
      --one-click) ACTION="one-click"; shift ;;
      --restore) ACTION="restore"; shift ;;
      --status) ACTION="status"; shift ;;
      --help|-h) echo "Usage: prismdns.sh [--master URL --token TOKEN] [--install|--apply|--one-click|--restore|--status]"; exit 0 ;;
      *) fail "未知参数: $1" ;;
    esac
  done
}

main() {
  parse_args "$@"
  mkdir -p "$STATE_DIR" 2>/dev/null || true
  case "$ACTION" in
    install) require_root; bootstrap; exit ;;
    apply) require_root; apply_permanent; exit ;;
    one-click) one_click; exit ;;
    restore) restore_dns; exit ;;
    status) show_status; exit ;;
  esac
  while true; do
    show_menu
    local choice
    choice=$(prompt "请选择 [0-8]: ")
    case "$choice" in
      1) require_root; bootstrap ;;
      2) test_dns ;;
      3) apply_permanent ;;
      4) one_click ;;
      5) apply_temporary ;;
      6) backup_dns ;;
      7) restore_dns ;;
      8) show_status ;;
      0) exit 0 ;;
      *) warn "无效选项。" ;;
    esac
    prompt "按 Enter 返回菜单..." >/dev/null
  done
}

main "$@"
