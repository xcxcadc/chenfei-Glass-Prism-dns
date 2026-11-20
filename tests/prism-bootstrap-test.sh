#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

curl() {
  local url="${!#}" output=''
  local base='https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases'
  while (($#)); do
    case "$1" in
      --output) output="$2"; shift 2 ;;
      *) shift ;;
    esac
  done
  case "$url" in
    "$base/latest")
      case "$BOOTSTRAP_CASE" in
        no-release) printf '%s' "$base" ;;
        prerelease) printf '%s' "$base/tag/v9.8.7-rc.1" ;;
        wrong-repo) printf '%s' 'https://github.com/other/other/releases/tag/v9.8.7' ;;
        offline) return 22 ;;
        *) printf '%s' "$base/tag/v9.8.7" ;;
      esac
      ;;
    'https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v9.8.7/dns.sh')
      case "$BOOTSTRAP_CASE" in
        bad-version) printf 'PRISMCTL_VERSION="1.0.0"\n' >"$output" ;;
        bad-syntax) printf 'PRISMCTL_VERSION="9.8.7"\nif\n' >"$output" ;;
        missing-script) return 22 ;;
        *) printf '%s\n' 'PRISMCTL_VERSION="9.8.7"' 'printf "%s\n" "$PRISM_SCRIPT_REF" "$@"' >"$output" ;;
      esac
      ;;
    *) return 2 ;;
  esac
}
export -f curl

result=$(BOOTSTRAP_CASE=success PRISM_SCRIPT_REF=v1.0.0 bash ./bootstrap.sh install client --master 'https://example.invalid')
[[ "$result" == $'v9.8.7\ninstall\nclient\n--master\nhttps://example.invalid' ]]
for test_case in no-release prerelease wrong-repo offline bad-version bad-syntax missing-script; do
  if BOOTSTRAP_CASE="$test_case" bash ./bootstrap.sh version >/dev/null 2>&1; then
    printf 'Expected bootstrap to refuse %s\n' "$test_case" >&2
    exit 1
  fi
done
printf 'Prism stable release bootstrap tests passed.\n'
