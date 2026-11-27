#!/usr/bin/env bash
set +x
set -Eeuo pipefail

NODE_UPDATE_REPO='xcxcadc/chenfei-Glass-Prism-dns'
ENV_FILE='/etc/prismdns/transport.env'
STATE_DIR='/var/lib/prism-transport/node-updates'
INSTALL_PATH='/usr/local/lib/prismdns/prism_node_update.sh'
UNIT_DIR='/etc/systemd/system'
AGENT_PATH='/usr/local/bin/prism-agent'
ROLE=''
MASTER=''
CREDENTIAL=''

fail() { printf 'Prism node update: %s\n' "$1" >&2; return 1; }

require_root() {
  [[ "$EUID" == 0 ]] || fail 'root privileges required'
}

valid_version() {
  [[ "$1" =~ ^v(0|[1-9][0-9]{0,8})\.(0|[1-9][0-9]{0,8})\.(0|[1-9][0-9]{0,8})$ ]]
}

newer_version() {
  [[ "$1" != "$2" && "$(printf '%s\n' "$1" "$2" | LC_ALL=C sort -V | tail -n1)" == "$1" ]]
}

tls_request() {
  local output="$1"
  shift
  HTTP_CODE=$(curl -q --silent --globoff --proto '=https' --proto-redir '=https' --tlsv1.2 \
    --connect-timeout 8 --max-time 30 --max-filesize 1048576 --output "$output" \
    --write-out '%{http_code}' "$@" 2>/dev/null) && [[ "$HTTP_CODE" =~ ^2[0-9]{2}$ ]]
}

post_worker() {
  local payload="$1" output="$2" body
  body=$(printf '%s' "$CREDENTIAL" | jq -Rsc --arg role "$ROLE" --argjson payload "$payload" \
    '$payload + {role:$role, credential:.}') || return 1
  tls_request "$output" --header 'Content-Type: application/json' --data-binary @- \
    "$MASTER/enhancer/api/node-updates/worker" <<<"$body"
}

inventory() {
  local version="$1" transport_version agent_version='' ipv4
  transport_version=$(sed -n 's/^VERSION="\([0-9.]*\)"$/\1/p' "${INSTALL_PATH%/*}/prism_transport.sh" 2>/dev/null | head -n1) || transport_version=''
  if [[ -x "$AGENT_PATH" ]]; then
    agent_version=$(timeout -k 1s 2s "$AGENT_PATH" --version 2>/dev/null | sed -nE \
      's/^(prism-agent[[:space:]]+)?([Vv]ersion[[:space:]]+)?(v?[0-9]+\.[0-9]+(\.[0-9]+)?)$/\3/p' | head -n1) || agent_version=''
  fi
  ipv4=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for (field=1; field<NF; field++) if ($field == "src") {print $(field+1); exit}}') || ipv4=''
  jq -cn --arg version "$version" --arg transport_version "$transport_version" \
    --arg agent_version "$agent_version" --arg ipv4 "$ipv4" \
    '{action:"poll", $version, $transport_version, $agent_version, $ipv4}'
}

refresh_release() {
  local version="$1" current="$2" staging="$3" filename from_dist=false
  if tls_request "$staging/release.json" --header 'Accept: application/vnd.github+json' \
    "https://api.github.com/repos/$NODE_UPDATE_REPO/releases/tags/$version"; then
    jq -e --arg version "$version" \
      '.tag_name == $version and .immutable == true and .draft == false and .prerelease == false' \
      "$staging/release.json" >/dev/null 2>&1 || return 1
  elif tls_request "$staging/manifest.json" "$MASTER/enhancer/dist/manifest.json" &&
       jq -e --arg version "$version" '.version == $version and (.files | has("prism_transport.sh") and has("prism_media.sh") and has("prism_node_update.sh"))' \
       "$staging/manifest.json" >/dev/null 2>&1; then
    from_dist=true
  else
    return 1
  fi
  for filename in prism_transport.sh prism_media.sh prism_node_update.sh; do
    if $from_dist; then
      tls_request "$staging/$filename" "$MASTER/enhancer/dist/$filename" || return 1
    else
      tls_request "$staging/$filename" "https://raw.githubusercontent.com/$NODE_UPDATE_REPO/refs/tags/$version/$filename" || return 1
    fi
    [[ -s "$staging/$filename" ]] && bash -n "$staging/$filename" >/dev/null 2>&1 || return 1
  done
  env -i PATH="$PATH" PRISM_SCRIPT_REF="$version" PRISM_NODE_UPDATE_EXPECTED_VERSION="$current" \
    PRISM_NODE_UPDATE_INSTALLER_FILE="$staging/prism_node_update.sh" PRISM_MEDIA_INSTALLER_FILE="$staging/prism_media.sh" \
    bash "$staging/prism_transport.sh" --refresh-runtime </dev/null >/dev/null 2>&1 9>&- || return 1
  [[ "$(<"$STATE_DIR/version")" == "$version" ]]
}

write_attempt() {
  local status="$1" reported="${2:-false}"
  jq -cn --argjson job "$job" --arg identity "$identity" --arg status "$status" --argjson reported "$reported" \
    '$job + {$identity, $status, $reported}' >"$staging/attempt.json" &&
    mv -fT -- "$staging/attempt.json" "$STATE_DIR/attempt.json"
}

report_attempt() {
  local status="$1" payload HTTP_CODE=''
  payload=$(jq -c --arg status "$status" '. + {action:"report", $status}' <<<"$job") || return 1
  if post_worker "$payload" "$staging/reply.json" || [[ "$HTTP_CODE" == 409 ]]; then
    write_attempt "$status" true
  else
    fail 'completion report could not be delivered'
  fi
}

private_state() {
  [[ -d "$STATE_DIR" && ! -L "$STATE_DIR" && "$(stat -c '%u:%a' "$STATE_DIR")" == 0:700 ]]
}

run_worker() (
  require_root
  [[ -f "$ENV_FILE" && ! -L "$ENV_FILE" && "$(stat -c '%u:%a' "$ENV_FILE")" == 0:600 ]] || return 0
  source "$ENV_FILE" >/dev/null 2>&1
  [[ ( "$ROLE" == client || "$ROLE" == proxy ) && -n "$CREDENTIAL" ]] || return 0
  [[ "$MASTER" =~ ^https://([a-zA-Z0-9.-]+|\[[a-fA-F0-9:.]+\])(:[0-9]{1,5})?/?$ ]] || return 0
  MASTER="${MASTER%/}"
  private_state || { fail 'update state is not private'; return 1; }
  umask 077
  exec 9>"$STATE_DIR/worker.lock"
  flock -n 9 || return 0
  local current staging identity payload job target status=failed saved_job='' saved_status=''
  current=$(<"$STATE_DIR/version")
  valid_version "$current" || { fail 'installed version is invalid'; return 1; }
  staging=$(mktemp -d "$STATE_DIR/job.XXXXXX")
  trap 'rm -rf -- "$staging"' EXIT
  trap 'exit 143' TERM INT
  identity=$(printf '%s\0' "$MASTER" "$ROLE" "$CREDENTIAL" | sha256sum)
  if [[ -f "$STATE_DIR/attempt.json" ]] && jq -e --arg identity "$identity" '.identity == $identity' "$STATE_DIR/attempt.json" >/dev/null 2>&1; then
    saved_job=$(jq -c '{id, version}' "$STATE_DIR/attempt.json")
    saved_status=$(jq -r .status "$STATE_DIR/attempt.json")
    if ! jq -e '.reported == true' "$STATE_DIR/attempt.json" >/dev/null; then
      job="$saved_job"
      if [[ "$saved_status" == applying ]]; then
        saved_status=failed
        [[ "$(jq -r .version <<<"$job")" != "$current" ]] || saved_status=succeeded
        write_attempt "$saved_status"
      fi
      report_attempt "$saved_status"
      return
    fi
  fi
  payload=$(inventory "$current")
  post_worker "$payload" "$staging/reply.json" || { fail 'poll failed'; return 1; }
  job=$(jq -cs '
    if length == 1 and (.[0] | type == "object" and has("job")) then
      .[0].job | if . == null then . elif type == "object" and
        (.id | type == "string" and length > 0 and length <= 128) and
        (.version | type == "string" and test("\\Av(0|[1-9][0-9]{0,8})\\.(0|[1-9][0-9]{0,8})\\.(0|[1-9][0-9]{0,8})\\z"))
      then {id, version} else error("invalid job") end
    else error("invalid poll response") end' "$staging/reply.json" 2>/dev/null) || { fail 'invalid poll response'; return 1; }
  [[ "$job" != null ]] || return 0
  target=$(jq -r .version <<<"$job")
  [[ "$job" != "$saved_job" ]] || { report_attempt "$saved_status"; return; }
  if { [[ "$target" == v1.5.37 ]] || newer_version "$target" v1.5.37; } && newer_version "$target" "$current"; then
    write_attempt applying
    if refresh_release "$target" "$current" "$staging"; then
      status=succeeded
    fi
  fi
  write_attempt "$status"
  report_attempt "$status"
)

check_owned_files() {
  local path signature
  for path in "$INSTALL_PATH" "$UNIT_DIR/prism-node-update.service" "$UNIT_DIR/prism-node-update.timer"; do
    [[ -e "$path" || -L "$path" ]] || continue
    case "$path" in
      "$INSTALL_PATH") signature="NODE_UPDATE_REPO='$NODE_UPDATE_REPO'" ;;
      *) signature='Description=Prism DNS managed node update' ;;
    esac
    [[ -f "$path" && ! -L "$path" ]] && grep -Fxq -- "$signature" "$path" || {
      fail 'existing node update file is not owned by Prism'; return 1;
    }
  done
}

install_worker() (
  require_root
  local version="$1" staging temporary=''
  valid_version "$version" || { fail 'invalid release version'; return 1; }
  check_owned_files
  umask 077
  if [[ -e "$STATE_DIR" || -L "$STATE_DIR" ]]; then
    private_state || { fail 'update state is not private'; return 1; }
  else
    install -d -m 700 "$STATE_DIR"
  fi
  staging=$(mktemp -d "$STATE_DIR/install.XXXXXX")
  trap 'rm -f -- "$temporary"; rm -rf -- "$staging"' EXIT
  temporary=$(mktemp "$INSTALL_PATH.XXXXXX")
  install -m 700 "${BASH_SOURCE[0]}" "$temporary"
  mv -fT -- "$temporary" "$INSTALL_PATH"
  cat >"$UNIT_DIR/prism-node-update.service" <<UNIT
[Unit]
Description=Prism DNS managed node update
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=$INSTALL_PATH --run
TimeoutStartSec=5min
UNIT
  cat >"$UNIT_DIR/prism-node-update.timer" <<'UNIT'
[Unit]
Description=Prism DNS managed node update

[Timer]
OnActiveSec=20s
OnUnitInactiveSec=20s
AccuracySec=2s

[Install]
WantedBy=timers.target
UNIT
  chmod 644 "$UNIT_DIR/prism-node-update.service" "$UNIT_DIR/prism-node-update.timer"
  systemctl daemon-reload
  systemctl enable --now prism-node-update.timer >/dev/null
  printf '%s\n' "$version" >"$staging/version"
  mv -fT -- "$staging/version" "$STATE_DIR/version"
)

uninstall_worker() {
  require_root
  check_owned_files
  if [[ -e "$STATE_DIR" || -L "$STATE_DIR" ]]; then
    private_state || { fail 'update state is not private'; return 1; }
  fi
  if [[ -f "$UNIT_DIR/prism-node-update.timer" ]]; then
    systemctl disable --now prism-node-update.timer >/dev/null
  fi
  if [[ -f "$UNIT_DIR/prism-node-update.service" ]]; then
    systemctl stop prism-node-update.service
  fi
  rm -f -- "$INSTALL_PATH" "$UNIT_DIR/prism-node-update.service" "$UNIT_DIR/prism-node-update.timer" \
    "$STATE_DIR/version" "$STATE_DIR/attempt.json" "$STATE_DIR/worker.lock"
  rmdir -- "$STATE_DIR" 2>/dev/null || true
  systemctl daemon-reload
}

main() {
  case "${1:-}" in
    --install) [[ $# == 2 ]] && install_worker "$2" ;;
    --run) [[ $# == 1 ]] && run_worker ;;
    --uninstall) [[ $# == 1 ]] && uninstall_worker ;;
    *) fail 'usage: prism_node_update.sh --install VERSION | --run | --uninstall' ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
