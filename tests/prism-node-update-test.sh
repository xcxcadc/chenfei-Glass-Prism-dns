#!/usr/bin/env bash
set -euo pipefail

source ./prism_node_update.sh

test_dir=$(mktemp -d)
trap 'rm -rf -- "$test_dir"' EXIT
STATE_DIR="$test_dir/state"
ENV_FILE="$test_dir/transport.env"
INSTALL_PATH="$test_dir/lib/prism_node_update.sh"
UNIT_DIR="$test_dir/systemd"
AGENT_PATH="$test_dir/prism-agent"
mkdir -p "$STATE_DIR" "$UNIT_DIR" "${INSTALL_PATH%/*}"
fixture='fixture credential; not a shell command'
role=client
master=https://panel.example.invalid
poll_reply='{"job":null}'
poll_code=200
report_code=200
release_json='{"immutable":true,"draft":false,"prerelease":false}'
refresh_succeeds=true
requests="$test_dir/requests"
systemd_log="$test_dir/systemd.calls"
events="$test_dir/events"
printf 'VERSION="2.5.2"\n' >"${INSTALL_PATH%/*}/prism_transport.sh"
printf '%s\n' '#!/usr/bin/env bash' 'printf "v1.2.1\n"' >"$AGENT_PATH"
chmod +x "$AGENT_PATH"

require_root() { :; }
stat() { if [[ "$3" == "$ENV_FILE" ]]; then echo 0:600; else echo 0:700; fi; }
ip() { [[ "$*" == '-4 route get 1.1.1.1' ]]; printf '1.1.1.1 via 192.0.2.1 src 192.0.2.8\n'; }
systemctl() {
  printf '%s\n' "$*" >>"$systemd_log"
  case "$1" in is-active|is-enabled) return 1 ;; esac
}

configure() {
  printf 'ROLE=%q\nMASTER=%q\nCREDENTIAL=%q\nTRANSPORT_MODE=direct\n' "$role" "$master" "$fixture" >"$ENV_FILE"
}

curl() {
  local output='' data='' url='' had_tls=false had_redirect_tls=false
  [[ "$1" == -q && " $* " != *' --location '* && " $* " != *' -L '* && " $* " != *"$fixture"* ]] || return 1
  while (($#)); do
    case "$1" in
      --proto) [[ "$2" == '=https' ]] || return 1; had_tls=true; shift 2 ;;
      --proto-redir) [[ "$2" == '=https' ]] || return 1; had_redirect_tls=true; shift 2 ;;
      --output) output="$2"; shift 2 ;;
      --data-binary) [[ "$2" == @- ]] || return 1; data=$(cat); shift 2 ;;
      https://*) url="$1"; shift ;;
      *) shift ;;
    esac
  done
  [[ "$had_tls" == true && "$had_redirect_tls" == true && -n "$output" ]] || return 1
  case "$url" in
    https://panel.example.invalid/enhancer/api/node-updates/worker)
      jq -e --arg role "$role" --arg credential "$fixture" '.role == $role and .credential == $credential' <<<"$data" >/dev/null || return 1
      jq -c 'del(.credential)' <<<"$data" >>"$requests"
      if [[ "$(jq -r .action <<<"$data")" == poll ]]; then
        printf '%s' "$poll_reply" >"$output"
        printf '%s' "$poll_code"
      else
        printf '{}' >"$output"
        printf '%s' "$report_code"
      fi
      ;;
    "https://api.github.com/repos/$NODE_UPDATE_REPO/releases/tags/"*)
      printf 'metadata\n' >>"$events"
      jq -c --arg tag "${url##*/}" '. + {tag_name:$tag}' <<<"$release_json" >"$output"
      printf '200'
      ;;
    "https://raw.githubusercontent.com/$NODE_UPDATE_REPO/refs/tags/"*)
      [[ -z "$data" && "$url" =~ /v1\.5\.(38|39|40|41)/prism_(transport|media|node_update)\.sh$ ]] || return 1
      printf '#!/usr/bin/env bash\n:\n' >"$output"
      printf '200'
      ;;
    *) return 1 ;;
  esac
}

env() {
  [[ "$1" == -i && "$3" == PRISM_SCRIPT_REF=* && "$4" == PRISM_NODE_UPDATE_EXPECTED_VERSION=* &&
     "$5" == PRISM_NODE_UPDATE_INSTALLER_FILE=* && "$6" == PRISM_MEDIA_INSTALLER_FILE=* &&
     "$7" == bash && "${8##*/}" == prism_transport.sh && "$9" == --refresh-runtime ]] || return 1
  printf 'refresh\n' >>"$events"
  [[ "$refresh_succeeds" == true ]] || return 1
  printf '%s\n' "${3#*=}" >"$STATE_DIR/version"
}

expect_failure() {
  "$@" >"$test_dir/failure.log" 2>&1 &
  local pid=$!
  if wait "$pid"; then
    printf 'Expected failure: %s\n' "$1" >&2
    return 1
  fi
}

configure
install_worker v1.5.37
[[ $(<"$STATE_DIR/version") == v1.5.37 ]]
grep -Fxq 'enable --now prism-node-update.timer' "$systemd_log"
grep -Fxq 'OnUnitInactiveSec=20s' "$UNIT_DIR/prism-node-update.timer"

run_worker
jq -e '.action == "poll" and .role == "client" and .version == "v1.5.37" and .transport_version == "2.5.2" and .agent_version == "v1.2.1" and .ipv4 == "192.0.2.8"' "$requests" >/dev/null
role=proxy
configure
run_worker
tail -n1 "$requests" | jq -e '.role == "proxy"' >/dev/null

request_count=$(wc -l <"$requests")
master=http://panel.example.invalid
configure
run_worker
master=https://panel.example.invalid
configure
(
  exec 8>"$STATE_DIR/worker.lock"
  flock -n 8
  run_worker
)
[[ $(wc -l <"$requests") == "$request_count" ]]

poll_reply='{"job":{"id":"successful","version":"v1.5.38","url":"https://untrusted.invalid","args":["--uninstall"]}}'
poll_code=504
expect_failure run_worker
[[ ! -e "$STATE_DIR/attempt.json" && ! -e "$events" ]]
poll_code=200
report_code=503
expect_failure run_worker
[[ $(<"$STATE_DIR/version") == v1.5.38 ]]
jq -e '.status == "succeeded" and .reported == false' "$STATE_DIR/attempt.json" >/dev/null
report_code=200
run_worker
run_worker
[[ $(grep -c '^refresh$' "$events") == 1 ]]
tail -n1 "$requests" | jq -e '.action == "report" and .status == "succeeded" and .id == "successful" and (keys | length == 5)' >/dev/null

poll_reply='{"job":{"id":"older","version":"v1.5.37"}}'
run_worker
tail -n1 "$requests" | jq -e '.status == "failed"' >/dev/null
poll_reply='{"job":{"id":"equal","version":"v1.5.38"}}'
run_worker
tail -n1 "$requests" | jq -e '.status == "failed"' >/dev/null
poll_reply='{"job":{"id":"invalid","version":"v1.5.39\n"}}'
expect_failure run_worker
poll_reply='{"job":{"id":"immutable","version":"v1.5.39"}}'
release_json='{"immutable":false,"draft":false,"prerelease":false}'
run_worker
[[ $(grep -c '^refresh$' "$events") == 1 && $(<"$STATE_DIR/version") == v1.5.38 ]]
tail -n1 "$requests" | jq -e '.status == "failed"' >/dev/null

release_json='{"immutable":true,"draft":false,"prerelease":false}'
refresh_succeeds=false
report_code=503
poll_reply='{"job":{"id":"failure","version":"v1.5.40"}}'
expect_failure run_worker
[[ $(<"$STATE_DIR/version") == v1.5.38 ]]
jq -e '.status == "failed" and .reported == false' "$STATE_DIR/attempt.json" >/dev/null
fixture='fixture rotated credential'
configure
poll_reply='{"job":null}'
run_worker
tail -n1 "$requests" | jq -e '.action == "poll" and .version == "v1.5.38"' >/dev/null
[[ $(grep -c '^refresh$' "$events") == 2 ]]

identity=$(printf '%s\0' "$master" "$role" "$fixture" | sha256sum)
report_code=200
jq -cn --arg identity "$identity" '{id:"interrupted",version:"v1.5.41",$identity,status:"applying",reported:false}' >"$STATE_DIR/attempt.json"
run_worker
tail -n1 "$requests" | jq -e '.id == "interrupted" and .status == "failed"' >/dev/null
jq -cn --arg identity "$identity" '{id:"committed",version:"v1.5.38",$identity,status:"applying",reported:false}' >"$STATE_DIR/attempt.json"
run_worker
tail -n1 "$requests" | jq -e '.id == "committed" and .status == "succeeded"' >/dev/null
[[ $(grep -c '^refresh$' "$events") == 2 ]]
if grep -Fq "$fixture" "$test_dir/failure.log" "$STATE_DIR/attempt.json" "$requests"; then
  printf 'Credential leaked into worker output or state\n' >&2
  exit 1
fi

printf 'unrelated\n' >"$UNIT_DIR/prism-node-update.timer"
expect_failure uninstall_worker
[[ $(<"$UNIT_DIR/prism-node-update.timer") == unrelated ]]
rm -f -- "$UNIT_DIR/prism-node-update.timer"
uninstall_worker
[[ ! -e "$INSTALL_PATH" && ! -e "$STATE_DIR/version" && ! -e "$UNIT_DIR/prism-node-update.service" ]]

for function_name in install_node_update_worker install_optional_node_update_worker managed_transport_refresh restore_transport_runtime uninstall_node_update_worker; do
  source <(sed -n "/^$function_name()[[:space:]]*[({]$/,/^[})]$/p" prism_transport.sh)
done
log() { printf '%s\n' "$*"; }
NODE_UPDATE_RUNTIME_DIR="$test_dir"
NODE_UPDATE_UNIT_DIR="$UNIT_DIR"
INSTALL_DIR="${INSTALL_PATH%/*}"
INSTALL_PATH="$INSTALL_DIR/prism_transport.sh"
MEDIA_WORKER_PATH="$INSTALL_DIR/media-worker"
SCRIPT_REF=v1.5.37
MASTER_OVERRIDE=''
curl() { return 22; }
install_optional_node_update_worker >"$test_dir/optional.log" 2>&1
grep -Fq 'transport remains installed' "$test_dir/optional.log"
[[ ! -e "$INSTALL_DIR/prism_node_update.sh" ]]

printf 'Description=Prism DNS managed node update\n' >"$UNIT_DIR/prism-node-update.timer"
printf 'unrelated\n' >"$UNIT_DIR/prism-node-update.service"
uninstall_node_update_worker
[[ ! -e "$UNIT_DIR/prism-node-update.timer" && $(<"$UNIT_DIR/prism-node-update.service") == unrelated ]]
rm -f -- "$UNIT_DIR/prism-node-update.service"

STATE_DIR="$test_dir/transport"
mkdir -p "$STATE_DIR/node-updates"
PRISM_NODE_UPDATE_EXPECTED_VERSION=v1.5.37
SCRIPT_REF=v1.5.38
printf '%s\n' "$PRISM_NODE_UPDATE_EXPECTED_VERSION" >"$STATE_DIR/node-updates/version"
printf '%s\n' 'INSTALL_PATH="$INSTALL_DIR/prism_transport.sh"' >"$INSTALL_PATH"
cp -- "$INSTALL_PATH" "$test_dir/original"
config_before=$(<"$ENV_FILE")
refresh_transport_runtime() {
  printf 'new transport\n' >"$INSTALL_PATH"
  printf 'new media\n' >"$MEDIA_WORKER_PATH"
}
install_node_update_worker() { return 1; }
expect_failure managed_transport_refresh
[[ $(<"$test_dir/original") == "$(<"$INSTALL_PATH")" ]]
[[ ! -e "$MEDIA_WORKER_PATH" && $(<"$STATE_DIR/node-updates/version") == v1.5.37 && $(<"$ENV_FILE") == "$config_before" ]]
printf 'Prism node update worker and lifecycle tests passed.\n'
