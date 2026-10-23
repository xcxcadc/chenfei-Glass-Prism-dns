#!/usr/bin/env bash
set +x
set -Eeuo pipefail

ENV_FILE="${PRISM_MEDIA_ENV_FILE:-/etc/prismdns/transport.env}"
CACHE_DIR="${PRISM_MEDIA_CACHE_DIR:-/var/lib/prism-media-check}"
SCRIPT_URL="${PRISM_MEDIA_SCRIPT_URL:-http://check.unlock.media}"
FALLBACK_URL="${PRISM_MEDIA_FALLBACK_URL:-https://raw.githubusercontent.com/lmc999/RegionRestrictionCheck/main/check.sh}"
CACHE_TTL="${PRISM_MEDIA_CACHE_TTL:-86400}"
RUN_LIMIT="${PRISM_MEDIA_RUN_LIMIT:-900}"
POLL_INTERVAL="${PRISM_MEDIA_POLL_INTERVAL:-1}"
HEARTBEAT_INTERVAL="${PRISM_MEDIA_HEARTBEAT_INTERVAL:-20}"
INSTALL_PATH="/usr/local/bin/prism-media-check"
UNIT_PATH="/etc/systemd/system/prism-media-check.service"
TIMER_PATH="/etc/systemd/system/prism-media-check.timer"
ROUTE_SYNC_PATH="${PRISM_MEDIA_ROUTE_SYNC_PATH:-/usr/local/lib/prismdns/sync-routes.sh}"
SCRIPT_PATH="$CACHE_DIR/official-check.sh"
SELF_PATH="${BASH_SOURCE[0]}"
RESULTS='{}'
SEQUENCE=0
SCRIPT_PID=''
STAGE_PID=''
RUN_OUTPUT=''
JOB_ID=''
JOB_GENERATION=''
JOB_DEADLINE=0
NEXT_HEARTBEAT=0
ROLE=''
MASTER=''
CREDENTIAL=''

declare -A MEDIA_LABELS=()
for label in 'Dazn' 'Disney+' 'Netflix' 'YouTube Premium' 'Amazon Prime Video' 'TVBAnywhere+' \
  'Spotify Registration' 'OneTrust Region' 'iQyi Oversea Region' 'Bing Region' 'Apple Region' \
  'YouTube CDN' 'Netflix Preferred CDN' 'ChatGPT' 'Google Gemini' 'Claude' \
  'Wikipedia Editability' 'Google Play Store' 'Google Search CAPTCHA Free' \
  'Steam Currency' 'Reddit' 'SD Gundam G Generation Eternal'; do
  MEDIA_LABELS["$label"]=1
done

fail() { printf 'Prism media worker: %s\n' "$1" >&2; return 1; }

ensure_dependencies() {
  local name missing=false
  for name in bash jq curl systemctl setsid flock; do
    command -v "$name" >/dev/null 2>&1 || missing=true
  done
  $missing || return 0
  if command -v apt-get >/dev/null 2>&1; then
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends bash jq curl ca-certificates util-linux systemd
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y bash jq curl ca-certificates util-linux systemd
  elif command -v yum >/dev/null 2>&1; then
    yum install -y bash jq curl ca-certificates util-linux systemd
  else
    fail 'install bash, jq, curl, systemd and util-linux before installing the worker'
  fi
}

load_environment() {
  [[ -r "$ENV_FILE" ]] || { fail 'transport configuration unavailable'; return 1; }
  ROLE=''; MASTER=''; CREDENTIAL=''
  source "$ENV_FILE"
  [[ "$ROLE" == proxy || "$ROLE" == client ]] || { fail 'invalid role'; return 1; }
  [[ -n "$CREDENTIAL" && "$MASTER" =~ ^https?://[^/@[:space:]]+(:[0-9]+)?/?$ ]] || {
    fail 'invalid controller configuration'; return 1;
  }
  MASTER="${MASTER%/}"
}

valid_script() {
  [[ -s "$1" ]] && head -n 1 "$1" | grep -Eq '^#!.*(ba)?sh([[:space:]]|$)' &&
    ! grep -Eiq '^[[:space:]]*(<!doctype[[:space:]]+html|<html([[:space:]>]|$))' "$1" &&
    bash -n "$1" >/dev/null 2>&1
}

cache_script() {
  local now stamp url temp
  mkdir -p "$CACHE_DIR"
  chmod 700 "$CACHE_DIR"
  now=$(date +%s)
  if [[ -f "$SCRIPT_PATH" ]] && valid_script "$SCRIPT_PATH"; then
    stamp=$(stat -c %Y "$SCRIPT_PATH")
    ((now - stamp < CACHE_TTL)) && return 0
  fi
  for url in "$SCRIPT_URL" "$FALLBACK_URL"; do
    temp=$(mktemp "$CACHE_DIR/check.XXXXXX")
    if curl -4 -fsSL --connect-timeout 8 --max-time 30 -o "$temp" "$url" && valid_script "$temp"; then
      chmod 700 "$temp"
      mv -f "$temp" "$SCRIPT_PATH"
      return 0
    fi
    rm -f -- "$temp"
  done
  fail 'official script download or validation failed'
}

parse_output() {
  local file="$1" line clean label value updated section='' changed=0
  local pattern='^[[:space:]]*([^:]+):[[:space:]]*(.+)$'
  while IFS= read -r line; do
    clean=$(printf '%s' "$line" | sed -E $'s/\033\\[[0-9;]*[[:alpha:]]//g')
    clean="${clean//$'\r'/}"
    if [[ "$clean" == *'Checking Results Under IPv4'* ]]; then
      section=4
      continue
    fi
    if [[ "$clean" == *'Checking Results Under IPv6'* ]]; then
      section=6
      continue
    fi
    [[ "$section" == 4 && "$clean" =~ $pattern ]] || continue
    label="${BASH_REMATCH[1]}"
    label="${label#"${label%%[![:space:]]*}"}"
    label="${label%"${label##*[![:space:]]}"}"
    [[ -v MEDIA_LABELS["$label"] ]] || continue
    value="${BASH_REMATCH[2]}"
    [[ ${#value} -le 1024 ]] || continue
    if [[ "$(jq -r --arg platform "$label" '.[$platform] // empty' <<<"$RESULTS")" != "$value" ]]; then
      updated=$(jq -c --arg platform "$label" --arg value "$value" '. + {($platform): $value}' <<<"$RESULTS") || return 2
      RESULTS="$updated"
      changed=1
    fi
  done < "$file"
  return "$((1 - changed))"
}

post_worker() {
  local action="$1" status="${2:-}" error="${3:-}" response http_code payload
  response=$(mktemp "$CACHE_DIR/response.XXXXXX")
  if [[ "$action" == report ]]; then
    SEQUENCE=$((SEQUENCE + 1))
  fi
  payload=$(PRISM_MEDIA_CREDENTIAL="$CREDENTIAL" jq -nc \
    --arg role "$ROLE" --arg action "$action" --arg job_id "$JOB_ID" \
    --argjson sequence "$SEQUENCE" --arg status "$status" \
    --argjson results "$RESULTS" --arg error "$error" \
    '{role:$role,credential:env.PRISM_MEDIA_CREDENTIAL,action:$action} +
      (if $action == "report" then
        {job_id:$job_id,sequence:$sequence,status:$status,results:$results} +
        (if $error == "" then {} else {error:$error} end)
      else {} end)')
  http_code=$(printf '%s' "$payload" | curl -sS --connect-timeout 5 --max-time 15 \
    -H 'Content-Type: application/json' --data-binary @- -o "$response" -w '%{http_code}' \
    "$MASTER/enhancer/api/media-checks/worker" 2>/dev/null) || http_code=000
  if [[ "$http_code" == 200 ]]; then
    if [[ "$action" == claim ]]; then
      JOB_ID=$(jq -r '.job.id // empty' "$response" 2>/dev/null) || JOB_ID=''
      JOB_GENERATION=$(jq -r '.job.generation // empty' "$response" 2>/dev/null) || JOB_GENERATION=''
    fi
    rm -f -- "$response"
    return 0
  fi
  rm -f -- "$response"
  [[ "$http_code" != 409 ]] || return 2
  return 1
}

process_group_alive() {
  kill -0 -- "-$1" 2>/dev/null
}

terminate_process_group() {
  local pid="$1" attempt
  process_group_alive "$pid" || { wait "$pid" 2>/dev/null || true; return 0; }
  kill -TERM -- "-$pid" 2>/dev/null || true
  for attempt in {1..5}; do
    process_group_alive "$pid" || break
    sleep 0.2
  done
  process_group_alive "$pid" && kill -KILL -- "-$pid" 2>/dev/null || true
  wait "$pid" 2>/dev/null || true
}

report_running_if_due() {
  local now report_code
  now=$(date +%s)
  ((now < NEXT_HEARTBEAT)) && return 0
  if post_worker report running; then
    NEXT_HEARTBEAT=$((now + HEARTBEAT_INTERVAL))
    return 0
  else
    report_code=$?
    ((report_code == 2)) && return 2
  fi
  NEXT_HEARTBEAT=$((now + 1))
  return 0
}

run_stage() {
  local exit_code=0 report_code now
  setsid "$@" </dev/null >/dev/null 2>&1 &
  STAGE_PID=$!
  while kill -0 "$STAGE_PID" 2>/dev/null; do
    now=$(date +%s)
    if ((now >= JOB_DEADLINE)); then
      terminate_process_group "$STAGE_PID"
      STAGE_PID=''
      return 124
    fi
    if report_running_if_due; then
      :
    else
      report_code=$?
      if ((report_code == 2)); then
        terminate_process_group "$STAGE_PID"
        STAGE_PID=''
        return 2
      fi
    fi
    sleep "$POLL_INTERVAL"
  done
  wait "$STAGE_PID" || exit_code=$?
  if process_group_alive "$STAGE_PID"; then
    terminate_process_group "$STAGE_PID"
  fi
  STAGE_PID=''
  return "$exit_code"
}

stop_script() {
  [[ -n "$SCRIPT_PID" ]] || return 0
  terminate_process_group "$SCRIPT_PID"
  SCRIPT_PID=''
}

cleanup_worker() {
  stop_script
  if [[ -n "$STAGE_PID" ]]; then
    terminate_process_group "$STAGE_PID"
    STAGE_PID=''
  fi
  [[ -z "$RUN_OUTPUT" ]] || rm -f -- "$RUN_OUTPUT"
}

run_worker() {
  local output now exit_code=0 report_code=0 timed_out=false
  umask 077
  [[ -r "$ENV_FILE" ]] || return 0
  load_environment || return 1
  mkdir -p "$CACHE_DIR"
  chmod 700 "$CACHE_DIR"
  exec 9>"$CACHE_DIR/worker.lock"
  flock -n 9 || return 0
  RESULTS='{}'; SEQUENCE=0; JOB_ID=''; JOB_GENERATION=''
  trap cleanup_worker EXIT
  trap 'exit 143' TERM
  trap 'exit 130' INT
  post_worker claim || { fail 'unable to claim media job'; return 1; }
  [[ -n "$JOB_ID" ]] || return 0
  JOB_DEADLINE=$(($(date +%s) + RUN_LIMIT))
  NEXT_HEARTBEAT=$(($(date +%s) + HEARTBEAT_INTERVAL))
  output=$(mktemp "$CACHE_DIR/output.XXXXXX")
  RUN_OUTPUT="$output"
  chmod 600 "$output"
  if ! post_worker report running; then
    fail 'unable to start media job'; return 1
  fi
  export PRISM_MEDIA_ENV_FILE="$ENV_FILE" PRISM_MEDIA_CACHE_DIR="$CACHE_DIR"
  export PRISM_MEDIA_SCRIPT_URL="$SCRIPT_URL" PRISM_MEDIA_FALLBACK_URL="$FALLBACK_URL"
  if run_stage bash "$SELF_PATH" --cache; then
    :
  else
    report_code=$?
    if ((report_code == 2)); then return 0; fi
    if ((report_code == 124)); then
      post_worker report failed 'Media job exceeded the 15 minute safety limit' || true
    else
      post_worker report failed 'Official script download or validation failed' || true
    fi
    return 1
  fi
  if [[ "$ROLE" == client && -x "$ROUTE_SYNC_PATH" ]]; then
    if run_stage env PRISM_SKIP_SERVICE_AUDIT=1 "$ROUTE_SYNC_PATH"; then
      :
    else
      report_code=$?
      if ((report_code == 2)); then return 0; fi
      if ((report_code == 124)); then
        post_worker report failed 'Media job exceeded the 15 minute safety limit' || true
      else
        post_worker report failed 'Current DNS routes could not be synchronized' || true
      fi
      return 1
    fi
  fi
  setsid bash "$SCRIPT_PATH" -M 4 -R 0 -E en 9>&- </dev/null >"$output" 2>&1 &
  SCRIPT_PID=$!
  while kill -0 "$SCRIPT_PID" 2>/dev/null; do
    now=$(date +%s)
    if ((now >= JOB_DEADLINE)); then
      timed_out=true
      stop_script
      break
    fi
    changed=false
    parse_output "$output" && changed=true || true
    if $changed || ((now >= NEXT_HEARTBEAT)); then
      if post_worker report running; then
        NEXT_HEARTBEAT=$((now + HEARTBEAT_INTERVAL))
      else
        report_code=$?
        if ((report_code == 2)); then
          stop_script
          return 0
        fi
        NEXT_HEARTBEAT=$((now + 1))
      fi
    fi
    sleep "$POLL_INTERVAL"
  done
  if ! $timed_out; then
    wait "$SCRIPT_PID" || exit_code=$?
    if process_group_alive "$SCRIPT_PID"; then
      terminate_process_group "$SCRIPT_PID"
    fi
    SCRIPT_PID=''
  fi
  parse_output "$output" || true
  if $timed_out; then
    post_worker report failed 'Official media script exceeded 15 minute safety limit' || true
    return 1
  fi
  if ((exit_code != 0)); then
    post_worker report failed "Official media script exited with status $exit_code" || true
    return 1
  fi
  if [[ "$RESULTS" == '{}' ]]; then
    post_worker report failed 'Official media script returned no IPv4 platform results' || true
    return 1
  fi
  if post_worker report completed; then
    return 0
  else
    report_code=$?
    ((report_code == 2)) && return 0
    return 1
  fi
}

install_worker() {
  [[ ${EUID:-$(id -u)} -eq 0 ]] || { fail 'root privileges required'; return 1; }
  ensure_dependencies
  install -d -m 700 "$CACHE_DIR"
  if [[ "$(readlink -f "${BASH_SOURCE[0]}")" != "$(readlink -f "$INSTALL_PATH")" ]]; then
    install -m 700 "${BASH_SOURCE[0]}" "$INSTALL_PATH.new"
    mv -f "$INSTALL_PATH.new" "$INSTALL_PATH"
  else
    chmod 700 "$INSTALL_PATH"
  fi
  cat > "$UNIT_PATH" <<'UNIT'
[Unit]
Description=Prism DNS official IPv4 media check worker
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/prism-media-check --run
TimeoutStartSec=16min
KillMode=control-group
MemoryMax=512M
TasksMax=128
NoNewPrivileges=yes
UNIT
  cat > "$TIMER_PATH" <<'TIMER'
[Unit]
Description=Prism DNS media check job poller

[Timer]
OnBootSec=15s
OnUnitInactiveSec=5s
AccuracySec=1s
Unit=prism-media-check.service

[Install]
WantedBy=timers.target
TIMER
  systemctl disable --now prismdns-service-audit.timer >/dev/null 2>&1 || true
  systemctl stop prismdns-service-audit.service >/dev/null 2>&1 || true
  find /run -maxdepth 1 -type f -name 'prismdns-media-check.*' -mmin +1440 -delete 2>/dev/null || true
  systemctl daemon-reload
  systemctl enable --now prism-media-check.timer >/dev/null
}

uninstall_worker() {
  [[ ${EUID:-$(id -u)} -eq 0 ]] || { fail 'root privileges required'; return 1; }
  systemctl disable --now prism-media-check.timer >/dev/null 2>&1 || true
  systemctl stop prism-media-check.service >/dev/null 2>&1 || true
  rm -f -- "$UNIT_PATH" "$TIMER_PATH" "$INSTALL_PATH"
  systemctl daemon-reload
}

main() {
  case "${1:-}" in
    --install) install_worker ;;
    --run) run_worker ;;
    --cache) cache_script ;;
    --uninstall) uninstall_worker ;;
    *) printf 'Usage: prism_media.sh --install|--run|--uninstall\n' >&2; return 2 ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
