#!/usr/bin/env bash
set -euo pipefail

source <(sed -n '/^validate_provider_summary() {$/,/^}$/p' prismdns.sh)
source <(sed -n '/^http_status_is_response() {$/,/^}$/p' prismdns.sh)
source <(sed -n '/^http_status_is_page_success() {$/,/^}$/p' prismdns.sh)
source <(sed -n '/^  download_media_check_script() {$/,/^  }$/p' prismdns.sh | sed 's/^  //')
source <(sed -n '/^  media_results_for_label() {$/,/^  }$/p' prismdns.sh | sed 's/^  //')
source <(sed -n '/^  media_result_is_positive() {$/,/^  }$/p' prismdns.sh | sed 's/^  //')
source <(sed -n '/^  media_summary_for_probe() {$/,/^  }$/p' prismdns.sh | sed 's/^  //')

curl() {
  local output="" url=""
  while (($#)); do
    case "$1" in
      -o) output="$2"; shift 2 ;;
      -*) shift ;;
      *) url="$1"; shift ;;
    esac
  done
  download_attempts=$((download_attempts + 1))
  if [[ "$MOCK_DOWNLOAD_MODE" == "html" || "$url" == "$MEDIA_CHECK_URL" ]]; then
    printf '<!doctype html><html></html>\n' >"$output"
  else
    printf '#!/bin/bash\nexit 0\n' >"$output"
  fi
}

media_script=$(mktemp)
MEDIA_CHECK_URL="https://check.example.invalid"
MEDIA_CHECK_FALLBACK_URL="https://fallback.example.invalid"
download_attempts=0
MOCK_DOWNLOAD_MODE="fallback"
if ! download_media_check_script || ((download_attempts != 2)) || ! grep -q '^#!/bin/bash$' "$media_script"; then
  printf 'valid fallback script was not accepted after rejecting HTML\n' >&2
  exit 1
fi

download_attempts=0
MOCK_DOWNLOAD_MODE="html"
if download_media_check_script || [[ -s "$media_script" ]]; then
  printf 'HTML responses were accepted as a shell script\n' >&2
  exit 1
fi
rm -f "$media_script"

assert_summary() {
  local expected="$1"
  shift
  local actual
  actual=$(validate_provider_summary "$@")
  if [[ "$actual" != "$expected" ]]; then
    printf 'expected %s, got %s\n' "$expected" "$actual" >&2
    exit 1
  fi
}

assert_summary 'PASS' 'PASS' false 1 1 1 1
assert_summary 'FAIL (target DNS route or required TLS/SNI validation failed)' 'PASS' false 1 0 1 1
assert_summary 'FAIL (target DNS route or required TLS/SNI validation failed)' 'PASS' false 1 1 0 1
assert_summary 'NOT TESTED (total route audit exceeded its 240s time budget)' 'PASS' true 1 1 1 1
assert_summary 'FAIL' 'FAIL' false 1 1 1 1

if ! http_status_is_response 403 || ! http_status_is_response 502 || http_status_is_response 000; then
  printf 'HTTP response classification regression\n' >&2
  exit 1
fi
if ! http_status_is_page_success 200 || ! http_status_is_page_success 302 || http_status_is_page_success 403; then
  printf 'HTTP page-success classification regression\n' >&2
  exit 1
fi

https_page_success=0
assert_summary 'PASS' 'PASS' false 1 1 1 1

assert_media_status() {
  local expected="$1" probe="$2" actual
  actual=$(media_summary_for_probe "$probe")
  if [[ "$actual" != "$expected"* ]]; then
    printf 'expected media status %s, got %s\n' "$expected" "$actual" >&2
    exit 1
  fi
}

media_error='script exceeded the 120s audit limit'
plain=$'ChatGPT: Yes\nNetflix: Yes (HK)\nDisney+: Yes (HK)\nDAZN: Yes (JO)\nGrok: Network Connection\nGemini: Ye'
assert_media_status 'PASS' '{"media_required":["ChatGPT"]}'
assert_media_status 'PASS' '{"media_required":["Netflix"]}'
assert_media_status 'PASS' '{"media_required":["Disney+"]}'
assert_media_status 'PASS' '{"media_required":["DAZN"]}'
assert_media_status 'FAIL' '{"media_required":["Grok"]}'
assert_media_status 'NOT TESTED' '{"media_required":["Gemini"]}'
assert_media_status 'NOT TESTED' '{"media_required":["ChatGPT","Gemini"]}'
assert_media_status 'PASS' '{"media_any":["Netflix","Gemini"]}'

media_error=''
assert_media_status 'NOT COVERED' '{"media_required":["Gemini"]}'
plain=$'ChatGPT: Network Connection\n'
assert_media_status 'FAIL' '{"media_required":["ChatGPT"]}'

printf 'Prism DNS audit classification tests passed.\n'
