#!/usr/bin/env bash
set -euo pipefail

for dependency in jq curl setsid flock; do
  command -v "$dependency" >/dev/null || { printf 'Missing test dependency: %s\n' "$dependency" >&2; exit 1; }
done

temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
export PRISM_MEDIA_ENV_FILE="$temp_dir/transport.env"
export PRISM_MEDIA_CACHE_DIR="$temp_dir/cache"
export PRISM_MEDIA_SCRIPT_URL="file://$temp_dir/official.sh"
export PRISM_MEDIA_FALLBACK_URL="file://$temp_dir/missing.sh"
export PRISM_MEDIA_ROUTE_SYNC_PATH="$temp_dir/sync-routes.sh"
export PRISM_MEDIA_HEARTBEAT_INTERVAL=1
export PRISM_MEDIA_POLL_INTERVAL=0.1
source prism_media.sh

printf '%s\n' '#!/usr/bin/env bash' \
  "printf 'Checking Results Under IPv4\\nChatGPT: Yes\\nNetflix: Yes (HK)\\nChecking Results Under IPv6\\nChatGPT: No\\n'" \
  >"$temp_dir/official.sh"
printf '%s\n' '#!/usr/bin/env bash' 'exit 1' >"$temp_dir/sync-routes.sh"
chmod +x "$temp_dir/sync-routes.sh"

printf '%s\n' 'Advertisement: No' 'Checking Results Under IPv4' 'ChatGPT: Yes' \
  'Netflix: Yes (HK)' 'Google Gemini: No' 'Checking Results Under IPv6' 'ChatGPT: No' \
  >"$temp_dir/output"
parse_output "$temp_dir/output"
[[ "$(jq -r '.ChatGPT' <<<"$RESULTS")" == Yes ]]
[[ "$(jq -r '.Netflix' <<<"$RESULTS")" == 'Yes (HK)' ]]
[[ "$(jq -r '."Google Gemini"' <<<"$RESULTS")" == No ]]
[[ "$(jq 'length' <<<"$RESULTS")" == 3 ]]
jq -nc --argjson results "$RESULTS" '{results:$results}' >/dev/null
if parse_output "$temp_dir/output"; then
  printf 'Unchanged output triggered another report\n' >&2
  exit 1
fi

printf '%s\n' '<html>not a script</html>' >"$temp_dir/html"
if valid_script "$temp_dir/html"; then
  printf 'HTML was accepted as the official script\n' >&2
  exit 1
fi
cache_script
[[ -s "$SCRIPT_PATH" ]] && valid_script "$SCRIPT_PATH"

post_worker() {
  [[ "$1" != report ]] || printf '%s\n' "$2" >>"$temp_dir/reports"
  return 0
}
JOB_DEADLINE=$(($(date +%s) + 15))
NEXT_HEARTBEAT=0
run_stage sleep 3
[[ $(wc -l <"$temp_dir/reports") -ge 2 ]]

post_worker() { return 2; }
JOB_DEADLINE=$(($(date +%s) + 15))
NEXT_HEARTBEAT=0
if run_stage sleep 60; then
  printf 'HTTP 409 did not stop the stage\n' >&2
  exit 1
else
  [[ $? == 2 ]]
fi

run_fixture() (
  printf 'ROLE=%s\nMASTER=https://panel.example.invalid\nCREDENTIAL=fixture\n' "$1" >"$ENV_FILE"
  RESULTS='{}'
  post_worker() {
    if [[ "$1" == claim ]]; then
      JOB_ID=fixture
    else
      SEQUENCE=$((SEQUENCE + 1))
      printf '%s %s\n' "$SEQUENCE" "$2" >>"$temp_dir/$1-status"
    fi
    return 0
  }
  if [[ "$1" == client ]]; then
    if run_worker; then
      printf 'Client ignored failing route synchronization\n' >&2
      exit 1
    fi
  else
    run_worker
  fi
)

run_fixture proxy
awk '$1 != NR {exit 1} END {if (NR < 2) exit 1}' "$temp_dir/report-status"
[[ "$(tail -n 1 "$temp_dir/report-status")" == *completed ]]
rm -f "$temp_dir/report-status"
run_fixture client
[[ "$(tail -n 1 "$temp_dir/report-status")" == *failed ]]

printf 'Prism media worker tests passed.\n'
