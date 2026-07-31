#!/usr/bin/env bash

set -Eeuo pipefail

VERSION="1.4.5"
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
  local expected detected secret smart installer transport_installer configured_master
  expected=$(jq -r '.expected_ip // empty' <<<"$response")
  detected=$(jq -r '.detected_ip // empty' <<<"$response")
  secret=$(jq -r '.secret // empty' <<<"$response")
  smart=$(jq -r '.smart // true' <<<"$response")
  installer=$(jq -r '.agent_installer // empty' <<<"$response")
  transport_installer=$(jq -r '.transport_installer // empty' <<<"$response")
  configured_master=$(jq -r '.master // empty' <<<"$response")
  [[ -n "$secret" && -n "$installer" && -n "$transport_installer" ]] || fail "面板返回的节点配置不完整。"
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
  install_client_transport "$transport_installer"
  install_traffic_reporter
  wait_for_local_dns
}

install_client_transport() {
  local installer_url="$1" installer="/tmp/prism_transport.sh"
  info "安装 Prism 加密解锁传输..."
  if [[ -n "${PRISM_TRANSPORT_INSTALLER_FILE:-}" ]]; then
    installer="$PRISM_TRANSPORT_INSTALLER_FILE"
  else
    curl -fsSL "$installer_url" -o "$installer"
  fi
  chmod +x "$installer"
  bash "$installer" --client --master "$MASTER" --token "$TOKEN"
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
LOCK_FILE="/run/prismdns-report.lock"
NFT_TABLE="prismdns_traffic"
COUNTER_VERSION="dns-sni-ports-v10-smart-dual-stack-rx-tx"
AUDIT_INTERVAL=21600
AUDIT_VERSION="selected-providers-browser-path-ipv4-v12"
HEALTH_CACHE_FILE="/var/lib/prismdns/route-health-report.json"
HEALTH_CACHE_INTERVAL=1800
[[ -f "$CONFIG_FILE" ]] || exit 0
if command -v flock >/dev/null 2>&1; then
  exec 8>"$LOCK_FILE"
  flock -n 8 || exit 0
else
  LOCK_DIR="${LOCK_FILE}.d"
  mkdir "$LOCK_DIR" 2>/dev/null || exit 0
  trap 'rmdir "$LOCK_DIR" 2>/dev/null || true' EXIT
fi
MASTER=$(sed -n 's/^master=//p' "$CONFIG_FILE" | head -1)
TOKEN=$(sed -n 's/^token=//p' "$CONFIG_FILE" | head -1)
[[ -n "$MASTER" && -n "$TOKEN" ]] || exit 0
if ! BOOTSTRAP=$(curl -fsSL --connect-timeout 8 --max-time 15 "$MASTER/enhancer/api/bootstrap/$TOKEN"); then
  exit 0
fi
mapfile -t PEERS < <(jq -r '.traffic_peers[]?' <<<"$BOOTSTRAP" | sort -u)
mapfile -t PEERS4 < <(printf '%s\n' "${PEERS[@]}" | awk 'index($0, ":") == 0 && length($0) > 0')
mapfile -t PEERS6 < <(printf '%s\n' "${PEERS[@]}" | awk 'index($0, ":") > 0')
((${#PEERS[@]} > 0)) || exit 0
PEER_HASH=$({ printf '%s\n' "$COUNTER_VERSION"; printf '%s\n' "${PEERS[@]}"; } | sha256sum | awk '{print $1}')
CURRENT_HASH=$(cat "$PEER_HASH_FILE" 2>/dev/null || true)
if [[ "$PEER_HASH" != "$CURRENT_HASH" ]] || ! nft list table inet "$NFT_TABLE" >/dev/null 2>&1; then
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
    printf '  chain output { type filter hook output priority -10; policy accept; ip daddr @peers4 udp dport 443 reject with icmp type port-unreachable; ip6 daddr @peers6 udp dport 443 reject with icmpv6 type port-unreachable; tcp sport 53 counter name tx; udp sport 53 counter name tx; ip daddr @peers4 tcp dport { 80, 443 } counter name tx; ip6 daddr @peers6 tcp dport { 80, 443 } counter name tx; }\n'
    printf '}\n'
  } > "$RULE_FILE"
  nft delete table inet "$NFT_TABLE" >/dev/null 2>&1 || true
  nft -f "$RULE_FILE"
  rm -f "$RULE_FILE"
  printf '%s\n' "$PEER_HASH" > "$PEER_HASH_FILE"
fi
RX=$(nft -j list counter inet "$NFT_TABLE" rx | jq '[.nftables[].counter? | .bytes] | add // 0')
TX=$(nft -j list counter inet "$NFT_TABLE" tx | jq '[.nftables[].counter? | .bytes] | add // 0')
if nft list counter inet prism_transport tx >/dev/null 2>&1; then
  TRANSPORT_TX=$(nft -j list counter inet prism_transport tx | jq '[.nftables[].counter? | .bytes] | add // 0')
  TX=$((TX + TRANSPORT_TX))
fi
if nft list counter inet prism_transport rx >/dev/null 2>&1; then
  TRANSPORT_RX=$(nft -j list counter inet prism_transport rx | jq '[.nftables[].counter? | .bytes] | add // 0')
  RX=$((RX + TRANSPORT_RX))
fi
DNS_READY=false
SYSTEM_DNS_READY=false
ROUTES_READY=false
HEALTHY_ROUTES=0
EXPECTED_ROUTES=0
HEALTH_MESSAGE=""
if systemctl is-active --quiet prism-agent 2>/dev/null &&
   ss -lntup 2>/dev/null | awk '$5 ~ /:53$/ && /prism-agent/ {found=1} END {exit !found}'; then
  DNS_READY=true
else
  HEALTH_MESSAGE="Prism Agent 未接管 53 端口"
fi
if awk '$1 == "nameserver" && ($2 == "127.0.0.1" || $2 == "::1") {found=1} END {exit !found}' /etc/resolv.conf 2>/dev/null; then
  SYSTEM_DNS_READY=true
else
  HEALTH_MESSAGE="${HEALTH_MESSAGE:+$HEALTH_MESSAGE；}系统 DNS 未使用 Prism"
fi
route_ready() {
  local domain="$1" answer peer
  local -a expected answers
  shift
  expected=("$@")
  ((${#expected[@]} > 0)) || return 1
  mapfile -t answers < <({
    dig @127.0.0.1 "$domain" A +short +time=2 +tries=1
    dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1
  } 2>/dev/null | sed '/^$/d' | sort -u)
  for answer in "${answers[@]}"; do
    for peer in "${expected[@]}"; do
      if [[ "$answer" == "$peer" ]]; then
        return 0
      fi
    done
  done
  return 1
}
ROUTE_HEALTH_KEY=$(jq -Sc '
  {mode:"agent-smart-dual-stack-v1",probes:([.health_probes[]? | {service_id,domain,probe_domains,route_domains,traffic_peers}] | sort_by(.service_id))}
' <<<"$BOOTSTRAP" | sha256sum | awk '{print $1}')
NOW=$(date +%s)
USE_HEALTH_CACHE=false
if [[ -s "$HEALTH_CACHE_FILE" ]]; then
  CACHE_KEY=$(jq -r '.key // empty' "$HEALTH_CACHE_FILE" 2>/dev/null || true)
  CACHE_TIME=$(jq -r '.checked_at // 0' "$HEALTH_CACHE_FILE" 2>/dev/null || echo 0)
  if [[ "$CACHE_KEY" == "$ROUTE_HEALTH_KEY" ]] && ((NOW - CACHE_TIME < HEALTH_CACHE_INTERVAL)); then
    HEALTHY_ROUTES=$(jq -r '.healthy // 0' "$HEALTH_CACHE_FILE")
    EXPECTED_ROUTES=$(jq -r '.expected // 0' "$HEALTH_CACHE_FILE")
    USE_HEALTH_CACHE=true
  fi
fi
if ! $USE_HEALTH_CACHE; then
  while IFS= read -r probe; do
    mapfile -t PROBE_PEERS < <(jq -r '.traffic_peers[]?' <<<"$probe" | sort -u)
    mapfile -t PROBE_DOMAINS < <(jq -r '[(.route_domains[]?, .probe_domains[]?, .domain // empty) | select(length > 0)] | unique[]' <<<"$probe")
    for domain in "${PROBE_DOMAINS[@]}"; do
      EXPECTED_ROUTES=$((EXPECTED_ROUTES + 1))
      if route_ready "$domain" "${PROBE_PEERS[@]}"; then
        HEALTHY_ROUTES=$((HEALTHY_ROUTES + 1))
      fi
    done
  done < <(jq -c '.health_probes[]?' <<<"$BOOTSTRAP")
  HEALTH_CACHE_TEMP="${HEALTH_CACHE_FILE}.tmp.$$"
  jq -nc --arg key "$ROUTE_HEALTH_KEY" --argjson checked_at "$NOW" \
    --argjson healthy "$HEALTHY_ROUTES" --argjson expected "$EXPECTED_ROUTES" \
    '{key:$key,checked_at:$checked_at,healthy:$healthy,expected:$expected}' >"$HEALTH_CACHE_TEMP"
  mv -f "$HEALTH_CACHE_TEMP" "$HEALTH_CACHE_FILE"
fi
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
reset_traffic_counters() {
  nft reset counter inet "$NFT_TABLE" rx >/dev/null 2>&1 || true
  nft reset counter inet "$NFT_TABLE" tx >/dev/null 2>&1 || true
  nft reset counter inet prism_transport tx >/dev/null 2>&1 || true
  nft reset counter inet prism_transport rx >/dev/null 2>&1 || true
}

run_service_audit() {
  local audit_hash="$1" now="$2" output="" attempt_output attempt_block plain="" results probe service_id service_name result provider_summary matched_peer answer peer mapping_output
  local providers_csv candidate failure test_provider probe_domain route_domain http_code resolve_peer
  local attempt attempt_count pass_count youtube_pass route_total route_pass https_total https_pass https_success
  local probe_attempt tls_attempt_pass page_ok
  local -a provider_results test_providers matched_results probe_domains route_domains probe_peers
  ut_supports_selected() {
    local help_output
    help_output=$(/usr/bin/ut -h 2>&1 || true)
    grep -q -- '-test string' <<<"$help_output"
  }
  providers_csv=$(jq -r '[.health_probes[]? | (.unlock_tests[]?, .unlock_test // empty) | select(length > 0)] | unique | join(",")' <<<"$BOOTSTRAP")
  if [[ -n "$providers_csv" ]]; then
    if [[ ! -x /usr/bin/ut ]] || ! ut_supports_selected; then
      timeout 180 bash -c 'curl -sL https://raw.githubusercontent.com/oneclickvirt/UnlockTests/main/ut_install.sh -sSf | bash' >/dev/null 2>&1 || true
    fi
    if [[ -x /usr/bin/ut ]]; then
      for attempt in 1 2 3; do
        if ut_supports_selected; then
          attempt_output=$(timeout 180 /usr/bin/ut -m 4 -test "$providers_csv" -b=false -s=false 2>&1 || true)
        else
          attempt_output=$(timeout 300 /usr/bin/ut -m 4 -f 20 -b=false -s=false 2>&1 || true)
        fi
        printf -v attempt_block '[attempt %d]\n%s\n' "$attempt" "$attempt_output"
        output+="$attempt_block"
        if ((attempt < 3)); then
          sleep 2
        fi
      done
      plain=$(sed $'s/\033\\[[0-9;]*[mK]//g' <<<"$output" | tr -d '\r')
    fi
  fi
  printf '%s\n' "$plain" > "$AUDIT_OUTPUT_FILE"
  results='{}'
  while IFS= read -r probe; do
    service_id=$(jq -r '.service_id // empty' <<<"$probe")
    service_name=$(jq -r '.name // empty' <<<"$probe")
    [[ -n "$service_id" ]] || continue
    result=""
    if [[ "$service_name" == "YouTube" ]]; then
      youtube_pass=0
      for attempt in 1 2 3; do
        mapping_output=$(curl -4 -fsSL --connect-timeout 6 --max-time 20 "https://redirector.googlevideo.com/report_mapping" || true)
        if curl -4 -fsSL --connect-timeout 6 --max-time 20 -o /dev/null "https://www.youtube.com/" &&
          grep -q '=> ' <<<"$mapping_output"; then
          youtube_pass=$((youtube_pass + 1))
        fi
      done
      if ((youtube_pass == 3)); then
        result="YES (Playback + CDN 3/3; Premium region is informational)"
      elif ((youtube_pass > 0)); then
        result="UNSTABLE (Playback + CDN ${youtube_pass}/3)"
      else
        result="FAIL (Playback CDN unavailable)"
      fi
    fi
    mapfile -t test_providers < <(jq -r '[(.unlock_tests[]?, .unlock_test // empty) | select(length > 0)] | unique[]' <<<"$probe")
    if [[ -z "$result" ]] && ((${#test_providers[@]} > 0)); then
      provider_results=()
      for test_provider in "${test_providers[@]}"; do
        mapfile -t matched_results < <(awk -v provider="$test_provider" '
          index($0, provider) == 1 && substr($0, length(provider) + 1, 2) ~ /^[[:space:]][[:space:]]$/ {
            value = substr($0, length(provider) + 1)
            sub(/^[[:space:]]+/, "", value)
            print value
          }
        ' <<<"$plain")
        provider_results+=("${matched_results[@]}")
      done
      attempt_count=${#provider_results[@]}
      pass_count=0
      failure=""
      for candidate in "${provider_results[@]}"; do
        if [[ "$candidate" =~ ^YES([[:space:]]|$) ]]; then
          pass_count=$((pass_count + 1))
        else
          failure="$candidate"
        fi
      done
      if ((attempt_count > 0 && pass_count == attempt_count)); then
        if ((${#test_providers[@]} > 1)); then
          result="YES (${#test_providers[@]} checks; ${pass_count}/${attempt_count} passes)"
        else
          result="${provider_results[attempt_count - 1]}"
        fi
      elif ((pass_count > 0)); then
        result="UNSTABLE (${pass_count}/${attempt_count} YES; ${failure:-intermittent failure})"
      elif ((attempt_count > 0)); then
        result="${provider_results[attempt_count - 1]}"
      fi
    fi
    provider_summary="$result"
    mapfile -t route_domains < <(jq -r '[(.route_domains[]?, .probe_domains[]?, .domain // empty) | select(length > 0)] | unique[]' <<<"$probe")
    mapfile -t probe_domains < <(jq -r '[(.probe_domains[]?, .domain // empty) | select(length > 0)] | unique[]' <<<"$probe")
    mapfile -t probe_peers < <(jq -r '.traffic_peers[]?' <<<"$probe" | sort -u)
    route_total=${#route_domains[@]}
    route_pass=0
    for route_domain in "${route_domains[@]}"; do
      if route_ready "$route_domain" "${probe_peers[@]}"; then
        route_pass=$((route_pass + 1))
      fi
    done
    https_total=${#probe_domains[@]}
    https_pass=0
    https_success=0
    for probe_domain in "${probe_domains[@]}"; do
      matched_peer=""
      while IFS= read -r answer; do
        for peer in "${probe_peers[@]}"; do
          if [[ "$answer" == "$peer" ]]; then
            matched_peer="$peer"
            break 2
          fi
        done
      done < <({
        dig @127.0.0.1 "$probe_domain" A +short +time=2 +tries=1
        dig @127.0.0.1 "$probe_domain" AAAA +short +time=2 +tries=1
      } 2>/dev/null | sed '/^$/d' | sort -u)
      [[ -n "$matched_peer" ]] || continue
      resolve_peer="$matched_peer"
      [[ "$resolve_peer" == *:* ]] && resolve_peer="[$resolve_peer]"
      tls_attempt_pass=0
      page_ok=false
      for probe_attempt in 1 2 3; do
        http_code=$(curl -sS -o /dev/null -w '%{http_code}' \
          -A 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36' \
          --resolve "$probe_domain:443:$resolve_peer" --connect-timeout 6 --max-time 25 "https://$probe_domain/" || true)
        if [[ "$http_code" =~ ^[1-5][0-9][0-9]$ ]]; then
          tls_attempt_pass=$((tls_attempt_pass + 1))
        fi
        if [[ "$http_code" =~ ^[23][0-9][0-9]$ ]]; then
          page_ok=true
        fi
        ((tls_attempt_pass >= 2)) && break
      done
      if ((tls_attempt_pass >= 2)); then
        https_pass=$((https_pass + 1))
      fi
      if $page_ok; then
        https_success=$((https_success + 1))
      fi
    done
    if ((route_total == 0 || route_pass != route_total)); then
      result="FAIL (DNS routes ${route_pass}/${route_total}; provider ${provider_summary:-not checked})"
    elif ((https_total == 0)); then
      result="FAIL (no TLS/SNI probe domains; DNS ${route_pass}/${route_total}; provider ${provider_summary:-not checked})"
    elif ((https_pass != https_total)); then
      result="FAIL (TLS/SNI ${https_pass}/${https_total}; page success ${https_success}; DNS ${route_pass}/${route_total}; provider ${provider_summary:-not checked})"
    elif ((https_success == 0)); then
      result="FAIL (DNS and TLS/SNI passed, but representative pages did not load; provider ${provider_summary:-not checked})"
    elif [[ "$provider_summary" =~ ^YES([[:space:]]|$) ]]; then
      result="PASS (DNS ${route_pass}/${route_total}; TLS/SNI ${https_pass}/${https_total}; provider verified; page success ${https_success}/${https_total})"
    else
      result="PASS (DNS ${route_pass}/${route_total}; TLS/SNI ${https_pass}/${https_total}; page success ${https_success}/${https_total}; provider reference ${provider_summary:-not configured})"
    fi
    results=$(jq -c --arg id "$service_id" --arg result "$result" '. + {($id):$result}' <<<"$results")
  done < <(jq -c '.health_probes[]?' <<<"$BOOTSTRAP")
  if [[ "$(jq 'length' <<<"$results")" -gt 0 ]]; then
    jq -nc --arg token "$TOKEN" --argjson results "$results" '{token:$token,scope:"unlock_services",results:$results}' |
      curl -fsSL --connect-timeout 8 --max-time 20 -H 'Content-Type: application/json' -d @- "$MASTER/enhancer/api/audit/report" >/dev/null || return 0
    printf '%s\n' "$audit_hash" > "$AUDIT_HASH_FILE"
    printf '%s\n' "$now" > "$AUDIT_TIME_FILE"
  fi
}

if [[ "${PRISM_SKIP_SERVICE_AUDIT:-0}" != "1" ]] &&
  { [[ "${PRISM_RUN_SERVICE_AUDIT:-0}" == "1" ]] || ! command -v systemctl >/dev/null 2>&1; } &&
  $DNS_READY && $SYSTEM_DNS_READY && $ROUTES_READY; then
  AUDIT_HASH=$({
    printf '%s\n' "$AUDIT_VERSION"
    jq -Sc '{
      service_audit_requested_at:(.service_audit_requested_at // ""),
      traffic_peers:((.traffic_peers // []) | sort),
      health_probes:([.health_probes[]? | {service_id,unlock_test,unlock_tests,domain,probe_domains,route_domains,traffic_peers}] | sort_by(.service_id))
    }' <<<"$BOOTSTRAP"
  } | sha256sum | awk '{print $1}')
  LAST_AUDIT_HASH=$(cat "$AUDIT_HASH_FILE" 2>/dev/null || true)
  LAST_AUDIT_TIME=$(cat "$AUDIT_TIME_FILE" 2>/dev/null || echo 0)
  NOW=$(date +%s)
  if [[ "$AUDIT_HASH" != "$LAST_AUDIT_HASH" ]] || ((NOW - LAST_AUDIT_TIME >= AUDIT_INTERVAL)); then
    run_service_audit "$AUDIT_HASH" "$NOW" || true
    reset_traffic_counters
  fi
fi
EOF
  chmod 700 /usr/local/lib/prismdns/report-traffic.sh
  cat > /usr/local/lib/prismdns/sync-routes.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
CONFIG_FILE="/var/lib/prismdns/client.conf"
HASH_FILE="/var/lib/prismdns/route-config.sha256"
RESTART_FILE="/var/lib/prismdns/route-restart.timestamp"
AUDIT_HASH_FILE="/var/lib/prismdns/service-audit.sha256"
AUDIT_VERSION="selected-providers-browser-path-ipv4-v12"
HEALTH_CHECK_FILE="/var/lib/prismdns/route-health.timestamp"
HEALTH_CHECK_INTERVAL=300
DNSMASQ_CONFIG="/etc/prismdns/dnsmasq.conf"
DNSMASQ_ROUTES="/etc/prismdns/dnsmasq-routes.conf"
LOCAL_DNS_SERVICE="prismdns-local-dns.service"
LOCAL_DNS_TABLE="prismdns_local_dns"
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
ensure_system_dns() {
  [[ -f /var/lib/prismdns/system-dns.enabled ]] || return 0
  if awk '$1 == "nameserver" && ($2 == "127.0.0.1" || $2 == "::1") {found=1} END {exit !found}' /etc/resolv.conf 2>/dev/null; then
    return 0
  fi
  local temporary
  temporary=$(mktemp /etc/resolv.conf.prismdns.XXXXXX)
  printf '# Managed by Prism DNS\nnameserver 127.0.0.1\noptions timeout:2 attempts:2\n' > "$temporary"
  chmod 644 "$temporary"
  chattr -i /etc/resolv.conf 2>/dev/null || true
  [[ -L /etc/resolv.conf ]] && rm -f /etc/resolv.conf
  mv -f "$temporary" /etc/resolv.conf
}
ensure_system_dns
ensure_local_dns_service() {
  command -v dnsmasq >/dev/null 2>&1 || return 1
  install -d -m 755 /etc/prismdns
  touch "$DNSMASQ_ROUTES"
  chmod 644 "$DNSMASQ_ROUTES"
  local config_tmp unit_tmp reload=false
  config_tmp=$(mktemp /etc/prismdns/dnsmasq.conf.XXXXXX)
  cat >"$config_tmp" <<'DNSMASQ_CONFIG_FILE'
domain-needed
bogus-priv
no-resolv
server=1.1.1.1
server=8.8.8.8
listen-address=127.0.0.1
port=5353
bind-interfaces
cache-size=0
conf-file=/etc/prismdns/dnsmasq-routes.conf
DNSMASQ_CONFIG_FILE
  chmod 644 "$config_tmp"
  if ! cmp -s "$config_tmp" "$DNSMASQ_CONFIG" 2>/dev/null; then
    mv -f "$config_tmp" "$DNSMASQ_CONFIG"
  else
    rm -f "$config_tmp"
  fi
  if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet dnsmasq.service 2>/dev/null ||
       systemctl is-enabled --quiet dnsmasq.service 2>/dev/null; then
      systemctl disable --now dnsmasq.service >/dev/null 2>&1 || true
    fi
    unit_tmp=$(mktemp /etc/systemd/system/prismdns-local-dns.service.XXXXXX)
    cat >"$unit_tmp" <<'DNSMASQ_UNIT'
[Unit]
Description=Prism deterministic local DNS routes
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStartPre=/usr/sbin/dnsmasq --test --conf-file=/etc/prismdns/dnsmasq.conf
ExecStart=/usr/sbin/dnsmasq --keep-in-foreground --conf-file=/etc/prismdns/dnsmasq.conf
Restart=always
RestartSec=2s

[Install]
WantedBy=multi-user.target
DNSMASQ_UNIT
    chmod 644 "$unit_tmp"
    if ! cmp -s "$unit_tmp" "/etc/systemd/system/$LOCAL_DNS_SERVICE" 2>/dev/null; then
      mv -f "$unit_tmp" "/etc/systemd/system/$LOCAL_DNS_SERVICE"
      reload=true
    else
      rm -f "$unit_tmp"
    fi
    $reload && systemctl daemon-reload
  fi
  if ! nft list table inet "$LOCAL_DNS_TABLE" >/dev/null 2>&1 ||
     ! nft list chain inet "$LOCAL_DNS_TABLE" output 2>/dev/null | grep -q 'redirect to :5353'; then
    nft delete table inet "$LOCAL_DNS_TABLE" >/dev/null 2>&1 || true
    nft -f - <<'LOCAL_DNS_NFT'
table inet prismdns_local_dns {
  chain output {
    type nat hook output priority -100; policy accept;
    ip daddr 127.0.0.1 udp dport 53 redirect to :5353
    ip daddr 127.0.0.1 tcp dport 53 redirect to :5353
    ip6 daddr ::1 udp dport 53 redirect to :5353
    ip6 daddr ::1 tcp dport 53 redirect to :5353
  }
}
LOCAL_DNS_NFT
  fi
}
render_local_routes() {
  local rules_tmp domain peer previous changed=false
  local -A domain_routes=()
  rules_tmp=$(mktemp /etc/prismdns/dnsmasq-routes.conf.XXXXXX)
  while IFS='|' read -r domain peer; do
    [[ -n "$domain" && -n "$peer" && "$peer" != *:* ]] || continue
    domain="${domain#.}"
    domain="${domain#\*.}"
    previous="${domain_routes[$domain]:-}"
    if [[ -n "$previous" && "$previous" != "$peer" ]]; then
      printf 'conflicting deterministic route for %s: %s != %s\n' "$domain" "$previous" "$peer" >&2
      rm -f "$rules_tmp"
      return 1
    fi
    domain_routes[$domain]="$peer"
  done < <(jq -r '
    .health_probes[]? |
    ([.traffic_peers[]? | select(contains(":") | not)] | first // "") as $peer |
    [.route_domains[]?, .probe_domains[]?, (.domain // empty)] | unique[] |
    select(length > 0) |
    "\(.)|\($peer)"
  ' <<<"$BOOTSTRAP" | sort -u)
  for domain in "${!domain_routes[@]}"; do
    printf 'local=/%s/\naddress=/%s/%s\n' \
      "$domain" "$domain" "${domain_routes[$domain]}"
  done | sort >"$rules_tmp"
  chmod 644 "$rules_tmp"
  if ! cmp -s "$rules_tmp" "$DNSMASQ_ROUTES" 2>/dev/null; then
    mv -f "$rules_tmp" "$DNSMASQ_ROUTES"
    changed=true
  else
    rm -f "$rules_tmp"
  fi
  if command -v systemctl >/dev/null 2>&1; then
    systemctl is-enabled --quiet "$LOCAL_DNS_SERVICE" 2>/dev/null ||
      systemctl enable "$LOCAL_DNS_SERVICE" >/dev/null 2>&1 || true
    if $changed || ! systemctl is-active --quiet "$LOCAL_DNS_SERVICE"; then
      dnsmasq --test --conf-file="$DNSMASQ_CONFIG" >/dev/null
      systemctl restart "$LOCAL_DNS_SERVICE"
    fi
  fi
}
remove_deterministic_dns() {
  systemctl disable --now "$LOCAL_DNS_SERVICE" >/dev/null 2>&1 || true
  rm -f "/etc/systemd/system/$LOCAL_DNS_SERVICE" "$DNSMASQ_CONFIG" "$DNSMASQ_ROUTES"
  nft delete table inet "$LOCAL_DNS_TABLE" >/dev/null 2>&1 || true
  systemctl daemon-reload >/dev/null 2>&1 || true
}
ensure_agent_mode() {
  local unit="" smart changed=false
  for candidate in /etc/systemd/system/prism-agent.service /lib/systemd/system/prism-agent.service; do
    if [[ -s "$candidate" ]]; then
      unit="$candidate"
      break
    fi
  done
  [[ -n "$unit" ]] || return 0
  smart=$(jq -r '.smart // true' <<<"$BOOTSTRAP")
  if [[ "$smart" == "true" ]] && ! grep -Eq '^ExecStart=.*(^|[[:space:]])--smart([[:space:]]|$)' "$unit"; then
    sed -i -E '/^ExecStart=/ s/$/ --smart/' "$unit"
    changed=true
  elif [[ "$smart" != "true" ]] && grep -Eq '^ExecStart=.*(^|[[:space:]])--smart([[:space:]]|$)' "$unit"; then
    sed -i -E '/^ExecStart=/ s/[[:space:]]+--smart([[:space:]]|$)/\1/g' "$unit"
    changed=true
  fi
  if $changed; then
    systemctl daemon-reload
    rm -f "$HASH_FILE" "$RESTART_FILE"
  fi
}
remove_deterministic_dns
ensure_agent_mode
CURRENT_HASH=$(jq -Sc '{
  mode:"agent-smart-dual-stack-v1",
  smart:(.smart // true),
  traffic_peers:((.traffic_peers // []) | sort),
  health_probes:([.health_probes[]? | {service_id,domain,probe_domains,route_domains,unlock_test,unlock_tests,traffic_peers}] | sort_by(.service_id))
}' <<<"$BOOTSTRAP" | sha256sum | awk '{print $1}')
LAST_HASH=$(cat "$HASH_FILE" 2>/dev/null || true)
NOW=$(date +%s)
AUDIT_HASH=$({
  printf '%s\n' "$AUDIT_VERSION"
  jq -Sc '{
    service_audit_requested_at:(.service_audit_requested_at // ""),
    traffic_peers:((.traffic_peers // []) | sort),
    health_probes:([.health_probes[]? | {service_id,unlock_test,unlock_tests,domain,probe_domains,route_domains,traffic_peers}] | sort_by(.service_id))
  }' <<<"$BOOTSTRAP"
} | sha256sum | awk '{print $1}')
LAST_AUDIT_HASH=$(cat "$AUDIT_HASH_FILE" 2>/dev/null || true)
trigger_service_audit() {
  [[ "$AUDIT_HASH" != "$LAST_AUDIT_HASH" ]] || return 0
  command -v systemctl >/dev/null 2>&1 || return 0
  systemctl is-active --quiet prismdns-service-audit.service && return 0
  systemctl start --no-block prismdns-service-audit.service >/dev/null 2>&1 || true
}
route_ready() {
  local domain="$1" answer peer
  local -a expected answers
  shift
  expected=("$@")
  ((${#expected[@]} > 0)) || return 1
  mapfile -t answers < <({
    dig @127.0.0.1 "$domain" A +short +time=2 +tries=1
    dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1
  } 2>/dev/null | sed '/^$/d' | sort -u)
  for answer in "${answers[@]}"; do
    for peer in "${expected[@]}"; do
      if [[ "$answer" == "$peer" ]]; then
        return 0
      fi
    done
  done
  return 1
}
basic_health_ok() {
  systemctl is-active --quiet prism-agent 2>/dev/null || return 1
  ss -lntup 2>/dev/null | awk '$5 ~ /:53$/ && /prism-agent/ {found=1} END {exit !found}'
}
route_health_ok() {
  basic_health_ok || return 1
  local probe domain
  local -a expected domains
  while IFS= read -r probe; do
    mapfile -t expected < <(jq -r '.traffic_peers[]?' <<<"$probe" | sort -u)
    mapfile -t domains < <(jq -r '[(.route_domains[]?, .probe_domains[]?, .domain // empty) | select(length > 0)] | unique[]' <<<"$probe")
    for domain in "${domains[@]}"; do
      route_ready "$domain" "${expected[@]}" || return 1
    done
  done < <(jq -c '.health_probes[]?' <<<"$BOOTSTRAP")
}
route_health_sample_ok() {
  basic_health_ok || return 1
  local probe domain
  local -a expected
  while IFS= read -r probe; do
    mapfile -t expected < <(jq -r '.traffic_peers[]?' <<<"$probe" | sort -u)
    domain=$(jq -r '[(.domain // empty), .probe_domains[]?, .route_domains[]?] | map(select(length > 0)) | unique | .[0] // empty' <<<"$probe")
    [[ -n "$domain" ]] || continue
    route_ready "$domain" "${expected[@]}" || return 1
  done < <(jq -c '.health_probes[]?' <<<"$BOOTSTRAP")
}
if [[ -n "$LAST_HASH" && "$CURRENT_HASH" == "$LAST_HASH" ]] && basic_health_ok; then
  LAST_HEALTH_CHECK=$(cat "$HEALTH_CHECK_FILE" 2>/dev/null || echo 0)
  if ((NOW - LAST_HEALTH_CHECK < HEALTH_CHECK_INTERVAL)) || route_health_sample_ok; then
    if ((NOW - LAST_HEALTH_CHECK >= HEALTH_CHECK_INTERVAL)); then
      printf '%s\n' "$NOW" > "$HEALTH_CHECK_FILE"
    fi
    trigger_service_audit
    exit 0
  fi
fi
LAST_RESTART=$(cat "$RESTART_FILE" 2>/dev/null || echo 0)
if ((NOW - LAST_RESTART < 60)); then
  exit 0
fi
printf '%s\n' "$NOW" > "$RESTART_FILE"
systemctl restart prism-agent
for _ in {1..30}; do
  if route_health_ok; then
    printf '%s\n' "$CURRENT_HASH" > "$HASH_FILE"
    printf '%s\n' "$NOW" > "$HEALTH_CHECK_FILE"
    trigger_service_audit
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
ExecStartPre=/bin/sleep 7
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
    cat > /etc/systemd/system/prismdns-service-audit.service <<'EOF'
[Unit]
Description=Run Prism DNS selected-service audit
After=network-online.target prism-agent.service

[Service]
Type=oneshot
Environment=PRISM_RUN_SERVICE_AUDIT=1
ExecStart=/usr/local/lib/prismdns/report-traffic.sh
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
  PRISM_SKIP_SERVICE_AUDIT=1 /usr/local/lib/prismdns/report-traffic.sh || warn "首次流量上报失败，定时任务会稍后重试。"
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
  local bootstrap probe domain matched_peer expected=0 mapped=0 healthy=0
  local -a peers domains answers
  bootstrap=$(curl -fsSL --connect-timeout 8 --max-time 15 "$MASTER/enhancer/api/bootstrap/$TOKEN")
  while IFS= read -r probe; do
    mapfile -t peers < <(jq -r '.traffic_peers[]?' <<<"$probe" | sort -u)
    mapfile -t domains < <(jq -r '[(.route_domains[]?, .probe_domains[]?, .domain // empty) | select(length > 0)] | unique[]' <<<"$probe")
    for domain in "${domains[@]}"; do
      expected=$((expected + 1))
      matched_peer=""
      probe_ok=false
      mapfile -t answers < <({
        dig @127.0.0.1 "$domain" A +short +time=2 +tries=1
        dig @127.0.0.1 "$domain" AAAA +short +time=2 +tries=1
      } 2>/dev/null | sed '/^$/d' | sort -u)
      for answer in "${answers[@]}"; do
        for peer in "${peers[@]}"; do
          if [[ "$answer" == "$peer" ]]; then
            matched_peer="$peer"
            break 2
          fi
        done
      done
      [[ -n "$matched_peer" ]] || continue
      mapped=$((mapped + 1))
      for probe_attempt in 1 2; do
        if curl -sS -o /dev/null --resolve "$domain:443:$matched_peer" --connect-timeout 5 --max-time 12 "https://$domain/"; then
          probe_ok=true
          break
        fi
        sleep 1
      done
      if $probe_ok; then
        healthy=$((healthy + 1))
      fi
    done
  done < <(jq -c '.health_probes[]?' <<<"$bootstrap")
  ((expected == mapped)) || fail "所选服务 DNS 路由验证失败：$mapped/$expected 个域名正确映射到已配置解锁机。"
  if ((healthy < expected)); then
    warn "所选服务 DNS 路由正确（$mapped/$expected）；HTTPS 探测通过 $healthy/$expected。第三方服务策略或瞬时网络异常不会阻断 DNS 接管。"
  else
    ok "所选服务 DNS 路由和 HTTPS 验证通过：$healthy/$expected。"
  fi
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
  touch "$STATE_DIR/system-dns.enabled"
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
  rm -f "$STATE_DIR/system-dns.enabled"
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

refresh_runtime() {
  require_root
  ensure_dependencies
  load_config
  validate_master
  install_traffic_reporter
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
      --refresh-runtime) ACTION="refresh-runtime"; shift ;;
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
    refresh-runtime) refresh_runtime; exit ;;
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
