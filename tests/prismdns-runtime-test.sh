#!/usr/bin/env bash
set -euo pipefail

temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
mkdir -p "$temp_dir/etc/systemd/system" "$temp_dir/lib/systemd/system" "$temp_dir/etc/prismdns"
source <(sed -n '/^render_local_routes() {$/,/^}$/p' prismdns.sh | sed "s|/etc/prismdns/|$temp_dir/etc/prismdns/|g")
source <(sed -n '/^ensure_agent_mode() {$/,/^}$/p' prismdns.sh | sed "s|/etc/systemd/system/|$temp_dir/etc/systemd/system/|g;s|/lib/systemd/system/|$temp_dir/lib/systemd/system/|g")
systemctl() { printf '%s\n' "$*" >>"$temp_dir/systemctl.calls"; }
dnsmasq() { return 0; }
local_dns_ready() { [[ "$listener_ready" == true ]]; }
DNSMASQ_ROUTES="$temp_dir/etc/prismdns/dnsmasq-routes.conf"
DNSMASQ_CONFIG="$temp_dir/etc/prismdns/dnsmasq.conf"
LOCAL_DNS_SERVICE=prismdns-local-dns.service
HASH_FILE="$temp_dir/route.hash"
RESTART_FILE="$temp_dir/restart.time"
BOOTSTRAP='{"master":"https://panel.example.invalid","smart":true,"health_probes":[{"domain":"chatgpt.com","traffic_peers":["198.51.100.1"]}]}'
listener_ready=true
render_local_routes
[[ $(grep -c '^restart ' "$temp_dir/systemctl.calls") == 1 ]]
render_local_routes
[[ $(grep -c '^restart ' "$temp_dir/systemctl.calls") == 1 ]]
listener_ready=false
render_local_routes
[[ $(grep -c '^restart ' "$temp_dir/systemctl.calls") == 2 ]]

printf '[Service]\nExecStart=/usr/local/bin/prism-agent --master http://old.example.invalid:8081 --secret fixture --smart\n' >"$temp_dir/etc/systemd/system/prism-agent.service"
touch "$HASH_FILE" "$RESTART_FILE"
ensure_agent_mode
grep -q -- '--master https://panel.example.invalid --secret fixture --smart' "$temp_dir/etc/systemd/system/prism-agent.service"
[[ ! -f "$HASH_FILE" && ! -f "$RESTART_FILE" ]]
before=$(grep -c '^daemon-reload' "$temp_dir/systemctl.calls")
ensure_agent_mode
[[ $(grep -c '^daemon-reload' "$temp_dir/systemctl.calls") == "$before" ]]
BOOTSTRAP='{"master":"https://panel.example.invalid/#invalid","smart":true}'
ensure_agent_mode
grep -q -- '--master https://panel.example.invalid --secret fixture --smart' "$temp_dir/etc/systemd/system/prism-agent.service"
printf 'Prism DNS runtime recovery tests passed.\n'
