#!/usr/bin/env bash

set -Eeuo pipefail

MODE=""
ASSUME_YES=false
PURGE_DATA=false
DNS_RESTORED_FROM=""

info() { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
ok() { printf '\033[1;32m[OK]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[WARN]\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: uninstall.sh MODE [--yes] [--purge-data]

MODE:
  panel   Remove Controller and Chinese enhancer from this server
  proxy   Remove the unlock-server Agent from this server
  client  Remove the unlocked-server Agent and restore system DNS
  agent   Detect proxy/client role, then remove the Agent

Options:
  --yes         Skip the confirmation prompt
  --purge-data  With panel mode, also permanently delete panel databases
EOF
}

require_root() {
  [[ ${EUID:-$(id -u)} -eq 0 ]] || fail "Please run as root."
  command -v systemctl >/dev/null 2>&1 || fail "systemd is required."
}

parse_args() {
  (($# > 0)) || { usage; exit 1; }
  MODE="$1"
  shift
  case "$MODE" in
    panel|proxy|client|agent) ;;
    --help|-h) usage; exit 0 ;;
    *) fail "Unknown uninstall mode: $MODE" ;;
  esac
  while (($#)); do
    case "$1" in
      --yes) ASSUME_YES=true ;;
      --purge-data) PURGE_DATA=true ;;
      --help|-h) usage; exit 0 ;;
      *) fail "Unknown option: $1" ;;
    esac
    shift
  done
  if [[ "$MODE" != "panel" && "$PURGE_DATA" == true ]]; then
    fail "--purge-data is only valid with panel mode."
  fi
}

confirm_uninstall() {
  $ASSUME_YES && return 0
  local message answer
  case "$MODE" in
    panel)
      if $PURGE_DATA; then
        message="Remove the panel and permanently delete all panel data?"
      else
        message="Remove the panel programs but keep panel data?"
      fi
      ;;
    proxy) message="Remove the Prism unlock-server Agent?" ;;
    client) message="Remove the Prism client Agent and restore system DNS?" ;;
    agent) message="Detect and remove the local Prism Agent?" ;;
  esac
  if [[ -r /dev/tty ]]; then
    read -r -p "$message [y/N]: " answer </dev/tty
  else
    read -r -p "$message [y/N]: " answer
  fi
  [[ "$answer" =~ ^[Yy]$ ]] || { info "Cancelled."; exit 0; }
}

stop_unit() {
  systemctl disable --now "$1" >/dev/null 2>&1 || true
}

remove_unit_file() {
  local unit="$1"
  stop_unit "$unit"
  rm -f "/etc/systemd/system/$unit" "/lib/systemd/system/$unit"
}

remove_transport_units() {
  local path unit
  for path in /etc/systemd/system/prism-transport-ssh-*.service; do
    [[ -e "$path" ]] || continue
    unit=$(basename "$path")
    stop_unit "$unit"
    rm -f "$path"
  done
  remove_unit_file prism-transport.timer
  remove_unit_file prism-transport.service
  remove_unit_file prism-egress-proxy.service
  stop_unit wg-quick@prismwg0.service
}

latest_dns_backup() {
  local candidate=""
  if [[ -d /var/lib/prismdns/backups ]]; then
    candidate=$(find /var/lib/prismdns/backups -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort -r | head -n 1 || true)
  fi
  printf '%s' "$candidate"
}

restore_system_dns() {
  local backup legacy_backup target temporary
  backup=$(latest_dns_backup)
  chattr -i /etc/resolv.conf >/dev/null 2>&1 || true
  rm -f /etc/NetworkManager/conf.d/prismdns.conf

  if [[ -n "$backup" ]]; then
    if [[ -s "$backup/resolv.conf.symlink" ]]; then
      target=$(cat "$backup/resolv.conf.symlink")
      rm -f /etc/resolv.conf
      ln -s "$target" /etc/resolv.conf
    elif [[ -f "$backup/resolv.conf" ]]; then
      rm -f /etc/resolv.conf
      cp -a "$backup/resolv.conf" /etc/resolv.conf
    fi
    [[ -f "$backup/resolved.conf" ]] && cp -a "$backup/resolved.conf" /etc/systemd/resolved.conf
    [[ -f "$backup/NetworkManager.conf" ]] && cp -a "$backup/NetworkManager.conf" /etc/NetworkManager/NetworkManager.conf
    [[ -f "$backup/resolved.enabled" ]] && systemctl enable systemd-resolved >/dev/null 2>&1 || true
    [[ -f "$backup/resolved.active" ]] && systemctl restart systemd-resolved >/dev/null 2>&1 || true
    DNS_RESTORED_FROM="$backup"
  else
    legacy_backup=$(find /etc -maxdepth 1 -type f -name 'resolv.conf.prism-backup-*' 2>/dev/null | sort -r | head -n 1 || true)
    if [[ -n "$legacy_backup" ]]; then
      rm -f /etc/resolv.conf
      cp -a "$legacy_backup" /etc/resolv.conf
      DNS_RESTORED_FROM="$legacy_backup"
    elif grep -Eq '^nameserver[[:space:]]+(127\.0\.0\.1|::1)([[:space:]]|$)' /etc/resolv.conf 2>/dev/null; then
      temporary=$(mktemp /etc/resolv.conf.prism-uninstall.XXXXXX)
      printf '# Restored by Prism DNS uninstaller\nnameserver 1.1.1.1\nnameserver 8.8.8.8\noptions timeout:2 attempts:2\n' >"$temporary"
      chmod 644 "$temporary"
      rm -f /etc/resolv.conf
      mv "$temporary" /etc/resolv.conf
      DNS_RESTORED_FROM="public fallback resolvers"
      warn "No DNS backup was found; restored public fallback resolvers."
    else
      DNS_RESTORED_FROM="unchanged"
    fi
  fi
  systemctl reload NetworkManager >/dev/null 2>&1 || true
}

restore_legacy_service() {
  local unit="$1" backup
  [[ -d /var/lib/prism-agent/conflict-backups ]] || return 0
  backup=$(find /var/lib/prism-agent/conflict-backups -mindepth 1 -maxdepth 1 -type d -name "*-$unit" 2>/dev/null | sort -r | head -n 1 || true)
  [[ -n "$backup" ]] || return 0
  case "$unit" in
    dnsmasq)
      [[ -f "$backup/dnsmasq.conf" ]] && cp -a "$backup/dnsmasq.conf" /etc/dnsmasq.conf
      [[ -d "$backup/dnsmasq.d" ]] && { rm -rf /etc/dnsmasq.d; cp -a "$backup/dnsmasq.d" /etc/dnsmasq.d; }
      ;;
    sniproxy)
      [[ -f "$backup/sniproxy.conf" ]] && cp -a "$backup/sniproxy.conf" /etc/sniproxy.conf
      [[ -d "$backup/sniproxy" ]] && { rm -rf /etc/sniproxy; cp -a "$backup/sniproxy" /etc/sniproxy; }
      ;;
  esac
  [[ -f "$backup/was-enabled" ]] && systemctl enable "$unit" >/dev/null 2>&1 || true
  [[ -f "$backup/was-active" ]] && systemctl restart "$unit" >/dev/null 2>&1 || true
}

remove_agent_runtime() {
  remove_transport_units
  remove_unit_file prism-agent-watchdog.timer
  remove_unit_file prism-agent-watchdog.service
  remove_unit_file prismdns-traffic.timer
  remove_unit_file prismdns-traffic.service
  remove_unit_file prismdns-route-sync.timer
  remove_unit_file prismdns-route-sync.service
  remove_unit_file prismdns-service-audit.service
  remove_unit_file prismdns-local-dns.service
  remove_unit_file prism-agent.service
  rm -rf /etc/systemd/system/prism-agent.service.d

  chattr -i /usr/local/bin/prism-agent >/dev/null 2>&1 || true
  rm -f /usr/local/bin/prism-agent /usr/local/bin/prism-egress-proxy
  rm -f /etc/cron.d/prismdns-traffic /etc/wireguard/prismwg0.conf
  ip link delete prismwg0 >/dev/null 2>&1 || true
  if command -v nft >/dev/null 2>&1; then
    nft delete table inet prismdns_traffic >/dev/null 2>&1 || true
    nft delete table inet prismdns_local_dns >/dev/null 2>&1 || true
    nft delete table inet prism_transport >/dev/null 2>&1 || true
    nft delete table inet prism_authorization >/dev/null 2>&1 || true
  fi
  systemctl daemon-reload
  systemctl reset-failed >/dev/null 2>&1 || true
}

purge_agent_files() {
  rm -rf /usr/local/lib/prismdns /etc/prismdns
  rm -rf /var/lib/prismdns /var/lib/prism-transport /var/lib/prism-agent
  if id prism-tunnel >/dev/null 2>&1; then
    userdel -r prism-tunnel >/dev/null 2>&1 || true
  else
    rm -rf /var/lib/prism-tunnel
  fi
}

detect_agent_role() {
  if [[ -f /var/lib/prismdns/system-dns.enabled ]] ||
     grep -Eq '^nameserver[[:space:]]+(127\.0\.0\.1|::1)([[:space:]]|$)' /etc/resolv.conf 2>/dev/null ||
     systemctl is-enabled prismdns-route-sync.timer >/dev/null 2>&1; then
    MODE="client"
  else
    MODE="proxy"
  fi
  info "Detected Agent role: $MODE"
}

uninstall_agent() {
  [[ "$MODE" == "agent" ]] && detect_agent_role
  remove_agent_runtime
  if [[ "$MODE" == "client" ]]; then
    restore_system_dns
  else
    restore_legacy_service dnsmasq
    restore_legacy_service sniproxy
  fi
  purge_agent_files
  if [[ "$MODE" == "client" ]]; then
    ok "Client Agent removed; system DNS restored from: $DNS_RESTORED_FROM"
  else
    ok "Unlock-server Agent removed."
  fi
}

uninstall_panel() {
  remove_unit_file prism-enhancer.service
  remove_unit_file prism-controller.service
  systemctl daemon-reload
  if $PURGE_DATA; then
    rm -rf /opt/prism /var/lib/prism-enhancer
    ok "Panel programs and all panel data removed."
  else
    rm -f /opt/prism/prism-controller /opt/prism/prism-enhancer
    ok "Panel programs removed; data retained in /opt/prism and /var/lib/prism-enhancer."
  fi
}

main() {
  parse_args "$@"
  require_root
  confirm_uninstall
  case "$MODE" in
    panel) uninstall_panel ;;
    proxy|client|agent) uninstall_agent ;;
  esac
  ok "Uninstallation completed. XrayR, V2bX, Docker and MTProxy were not modified."
}

main "$@"
