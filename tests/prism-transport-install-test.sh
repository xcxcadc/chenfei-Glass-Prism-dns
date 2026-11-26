#!/usr/bin/env bash
set -euo pipefail

source <(sed -n '/^install_transport() {$/,/^}$/p' prism_transport.sh)

temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
source <(sed -n '/^main() {$/,/^}$/p' prism_transport.sh | sed "s#/run/prism-transport.lock#$temp_dir/transport.lock#")
events="$temp_dir/events"
ROLE=proxy
MASTER=https://panel.example.invalid
CREDENTIAL=fixture
INSTALL_DIR="$temp_dir/lib"
STATE_DIR="$temp_dir/state"
INSTALL_PATH="$temp_dir/installed.sh"
VERSION=fixture
TRANSPORT_MODE=direct
record() { printf '%s\n' "$1" >>"$events"; }
ensure_dependencies() { record dependencies; }
install() { record install; }
save_environment() { record environment; }
install_media_worker() { record worker; }
install_proxy_egress() { record egress; }
install_timer() { record timer; }
sync_transport() { record sync; return 23; }
log() { record log; }
fail() { printf '%s\n' "$*" >&2; exit 1; }

set +e
(set -e; install_transport)
exit_code=$?
set -e
[[ "$exit_code" == 23 ]]
[[ "$(paste -sd ' ' "$events")" == 'dependencies install environment worker egress timer sync' ]]

require_root() { :; }
parse_args() { ACTION=install; }
command() {
  if [[ "$*" == '-v flock' && ! -f "$temp_dir/flock-ready" ]]; then
    return 1
  fi
  builtin command "$@"
}
ensure_dependencies() { record dependencies; touch "$temp_dir/flock-ready"; }
flock() { [[ -f "$temp_dir/flock-ready" ]]; record lock; }
install_transport() { record install; }
install_optional_node_update_worker() { record node-update; }
: >"$events"
main
[[ "$(paste -sd ' ' "$events")" == 'dependencies lock install node-update' ]]

source <(sed -n '/^apply_client_egress_rules() {$/,/^}$/p; /^apply_client_rules() {$/,/^}$/p' prism_transport.sh)
ROLE=client
CLIENT_EGRESS_HTTP_PORT=19081
CLIENT_EGRESS_HTTPS_PORT=19444
collect_transport_cgroups() { printf '%s\n' 12345; }
nft() {
  case "$1" in
    list) [[ "$2" == table && "$3" == inet ]] ;;
    delete) record "delete $4" ;;
    -f)
      record 'transaction'
      cp "$2" "$temp_dir/transaction"
      [[ ! -f "$temp_dir/reject-update" ]]
      ;;
    *) return 2 ;;
  esac
}
: >"$events"
apply_client_egress_rules '{"egress_services":[{}]}'
[[ "$(cat "$events")" == transaction ]]
[[ "$(head -n1 "$temp_dir/transaction")" == 'delete table inet prism_client_egress' ]]
grep -Fq 'meta cgroup 12345 tcp dport 443 redirect to :19444' "$temp_dir/transaction"
: >"$events"
apply_client_egress_rules '{}'
[[ "$(cat "$events")" == 'delete prism_client_egress' ]]

: >"$events"
touch "$temp_dir/reject-update"
if (apply_client_rules '[]' 2>/dev/null); then
  printf '%s\n' 'Failed client transport transaction must not report success.' >&2
  exit 1
fi
[[ "$(cat "$events")" == transaction ]]
[[ "$(head -n1 "$temp_dir/transaction")" == 'delete table inet prism_transport' ]]
grep -Fxq 'table inet prism_transport {' "$temp_dir/transaction"

printf 'Prism transport installation order tests passed.\n'
