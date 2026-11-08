#!/usr/bin/env bash
set -Eeuo pipefail

PRISMCTL_COMPAT_VERSION="1.5.36"
PRISMCTL_COMPAT_NAME="Prism DNS"
compat_source="${BASH_SOURCE[0]}"
compat_ref="${PRISM_SCRIPT_REF:-v${PRISMCTL_COMPAT_VERSION}}"
compat_repo="${PRISM_SCRIPT_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"

if [[ -f "$compat_source" && -r "$(dirname "$compat_source")/dns.sh" ]]; then
  source "$(dirname "$compat_source")/dns.sh" "$@"
else
  compat_directory=$(mktemp -d /tmp/prism-dns-compat.XXXXXX)
  trap 'rm -rf -- "$compat_directory"' EXIT
  curl -fsSL --connect-timeout 8 --max-time 60 \
    "https://raw.githubusercontent.com/$compat_repo/refs/tags/$compat_ref/dns.sh" \
    -o "$compat_directory/dns.sh"
  bash -n "$compat_directory/dns.sh"
  bash "$compat_directory/dns.sh" "$@"
fi
