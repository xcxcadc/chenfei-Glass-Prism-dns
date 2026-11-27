#!/usr/bin/env bash
# Ultra-short Prism DNS entry script (multi-OS, no process substitution required)
# Usage: curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/i.sh | bash
set -euo pipefail
REPO="${PRISM_SCRIPT_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    exec sudo -E bash -c "curl -fsSL https://raw.githubusercontent.com/${REPO}/main/i.sh | bash"
  fi
  echo "[ERROR] Please run as root (or with sudo)." >&2
  exit 1
fi
TMP=$(mktemp /tmp/prism-bootstrap.XXXXXX.sh)
trap 'rm -f "$TMP"' EXIT
curl -fsSL --connect-timeout 10 --max-time 60 "https://raw.githubusercontent.com/${REPO}/main/bootstrap.sh" -o "$TMP"
if [ -r /dev/tty ]; then
  bash "$TMP" "$@" </dev/tty
else
  bash "$TMP" "$@"
fi
