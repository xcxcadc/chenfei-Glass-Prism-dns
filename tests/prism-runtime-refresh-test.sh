#!/usr/bin/env bash
set -euo pipefail

source <(sed -n '/^refresh_runtime() {$/,/^}$/p' prismdns.sh)
source <(sed -n '/^refresh_transport_runtime() {$/,/^}$/p' prism_transport.sh)

events=''
record() { events+="$1 "; }
require_root() { :; }
ensure_dependencies() { :; }
load_config() { MASTER=https://panel.example.invalid; TOKEN=0123456789abcdef0123456789abcdef; }
validate_master() { :; }
curl() { printf '{"transport_installer":"https://panel.example.invalid/prism_transport.sh"}\n'; }
jq() { printf 'https://panel.example.invalid/prism_transport.sh\n'; }
install_client_transport() {
  [[ "$1" == https://panel.example.invalid/prism_transport.sh && "$2" == refresh ]]
  record transport
}
install_traffic_reporter() {
  [[ "$1" == refresh ]]
  record reporter
}

temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
ENV_FILE="$temp_dir/transport.env"
PRISM_TRANSPORT_ENV_FILE="$ENV_FILE"
INSTALL_DIR="$temp_dir/lib"
INSTALL_PATH="$INSTALL_DIR/prism_transport.sh"
printf 'ROLE=client\nMASTER=https://old.example.invalid\nCREDENTIAL=fixture\nTRANSPORT_MODE=direct\n' >"$ENV_FILE"
refresh_runtime
[[ "$events" == 'transport reporter ' ]]

events=''
MASTER_OVERRIDE=https://panel.example.invalid
load_environment() {
  source "$ENV_FILE"
  MASTER="$MASTER_OVERRIDE"
}
save_environment() { record save; }
install() { record install; }
install_timer() { record timer; }
install_media_worker() { record worker; }
log() { :; }
fail() { printf '%s\n' "$*" >&2; exit 1; }

refresh_transport_runtime
[[ "$events" == 'save install timer worker ' ]]

events=''
ROLE=proxy
load_environment() { ROLE=proxy; MASTER=https://panel.example.invalid; CREDENTIAL=fixture; }
if (refresh_transport_runtime 2>/dev/null); then
  printf 'proxy transport was accepted as a client runtime\n' >&2
  exit 1
fi
[[ -z "$events" ]]

printf 'Prism runtime refresh tests passed.\n'
