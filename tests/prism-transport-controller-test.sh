#!/usr/bin/env bash
set -euo pipefail

source <(sed -n '/^transport_response_valid() {$/,/^}$/p' prism_transport.sh)
source <(sed -n '/^post_json() {$/,/^}$/p' prism_transport.sh)

ROLE=proxy
MASTER_OVERRIDE=""
MASTER=https://old.example.invalid
candidate_master=https://new.example.invalid
saved_master=""
mode=healthy

agent_master_endpoint() { printf '%s\n' "$candidate_master"; }
save_environment() { saved_master="$MASTER"; }
log() { :; }
post_transport_json() {
  if [[ "$mode" == healthy && "$1" == https://old.example.invalid ]]; then
    printf '{"role":"proxy","peers":[]}\n'
  elif [[ "$mode" == recovery && "$1" == https://new.example.invalid ]]; then
    printf '{"role":"proxy","peers":[],"authorized_ips":["192.0.2.1"]}\n'
  elif [[ "$mode" == invalid && "$1" == https://new.example.invalid ]]; then
    printf '{"role":"client","peers":[]}\n'
  else
    return 7
  fi
}

post_json /enhancer/api/transport/proxy '{}' >/dev/null
[[ "$MASTER" == https://old.example.invalid && -z "$saved_master" ]]

mode=recovery
post_json /enhancer/api/transport/proxy '{}' >/dev/null
[[ "$MASTER" == https://new.example.invalid && "$saved_master" == "$MASTER" ]]

MASTER=https://old.example.invalid
saved_master=""
mode=invalid
if post_json /enhancer/api/transport/proxy '{}' >/dev/null; then
  printf 'invalid controller registration was accepted\n' >&2
  exit 1
fi
[[ "$MASTER" == https://old.example.invalid && -z "$saved_master" ]]

mode=recovery
candidate_master=https://username:password@new.example.invalid
if post_json /enhancer/api/transport/proxy '{}' >/dev/null; then
  printf 'controller URL with credentials was accepted\n' >&2
  exit 1
fi
[[ -z "$saved_master" ]]

candidate_master=https://old.example.invalid
if post_json /enhancer/api/transport/proxy '{}' >/dev/null; then
  printf 'unreachable unchanged controller was accepted\n' >&2
  exit 1
fi
[[ -z "$saved_master" ]]

printf 'Prism transport controller recovery tests passed.\n'
