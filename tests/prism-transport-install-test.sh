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
: >"$events"
main
[[ "$(paste -sd ' ' "$events")" == 'dependencies lock install' ]]

printf 'Prism transport installation order tests passed.\n'
