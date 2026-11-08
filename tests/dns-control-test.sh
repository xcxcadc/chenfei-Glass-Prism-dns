#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname "$0")/.."
export PRISMCTL_SOURCE_ONLY=1
source ./dns.sh

test_directory=$(mktemp -d)
trap 'rm -rf -- "$test_directory"' EXIT
BIN_PATH="$test_directory/bin/dns"
WORK_DIR="$test_directory/downloads"
mkdir -p "$WORK_DIR"
cp ./prismctl.sh "$test_directory/legacy-version"
mkdir -p "$(dirname "$BIN_PATH")"
cp ./prismctl.sh "$(dirname "$BIN_PATH")/prismctl"
cmd_self_install
[[ -x "$BIN_PATH" ]]
[[ "$(env -u PRISMCTL_SOURCE_ONLY "$BIN_PATH" version)" == 'dns 1.5.36' ]]
[[ -L "$(dirname "$BIN_PATH")/prismctl" ]]
[[ "$(env -u PRISMCTL_SOURCE_ONLY "$(dirname "$BIN_PATH")/prismctl" version)" == 'dns 1.5.36' ]]
cmd_self_install
[[ "$(env -u PRISMCTL_SOURCE_ONLY "$BIN_PATH" version)" == 'dns 1.5.36' ]]

BIN_PATH="$test_directory/unrelated/dns"
mkdir -p "$(dirname "$BIN_PATH")"
printf 'unrelated command\n' >"$BIN_PATH"
if (cmd_self_install >/dev/null 2>&1); then
  printf 'Unrelated dns command was overwritten\n' >&2
  exit 1
fi
[[ "$(cat "$BIN_PATH")" == 'unrelated command' ]]

BIN_PATH="$test_directory/custom/dns"
mkdir -p "$(dirname "$BIN_PATH")"
printf 'unrelated legacy command\n' >"$(dirname "$BIN_PATH")/prismctl"
cmd_self_install
[[ "$(cat "$(dirname "$BIN_PATH")/prismctl")" == 'unrelated legacy command' ]]
baseline=$(sha256sum "$BIN_PATH")
SCRIPT_REF=v99.0.0
fetch_script() { printf '%s' "$test_directory/invalid.sh"; }
printf '#!/usr/bin/env bash\nprintf other\n' >"$test_directory/invalid.sh"
if (cmd_self_install >/dev/null 2>&1); then
  printf 'Invalid control script replaced the working command\n' >&2
  exit 1
fi
[[ "$(sha256sum "$BIN_PATH")" == "$baseline" ]]
SCRIPT_REF=v1.5.36

mkdir -p "$test_directory/mockbin"
cat >"$test_directory/mockbin/curl" <<'SCRIPT'
#!/usr/bin/env bash
set -euo pipefail
while (($#)); do
  case "$1" in
    -o) cp "$DNS_CONTROL_FIXTURE" "$2"; exit 0 ;;
    *) shift ;;
  esac
done
exit 1
SCRIPT
chmod +x "$test_directory/mockbin/curl"
export DNS_CONTROL_FIXTURE="$PWD/dns.sh"
[[ "$(env -u PRISMCTL_SOURCE_ONLY bash <(cat ./dns.sh) version)" == 'dns 1.5.36' ]]
[[ "$(env -u PRISMCTL_SOURCE_ONLY bash ./prismctl.sh version)" == 'dns 1.5.36' ]]
[[ "$(PATH="$test_directory/mockbin:$PATH" env -u PRISMCTL_SOURCE_ONLY bash <(cat ./prismctl.sh) version)" == 'dns 1.5.36' ]]

require_root() { :; }
BIN_PATH="$test_directory/menu/dns"
TTY_DEV="$test_directory/menu-input"
printf '0\n' >"$TTY_DEV"
has_panel() { return 1; }
unit_exists() { return 1; }
ENV_FILE="$test_directory/missing.env"
WORK_DIR="$test_directory/menu-work"
mkdir -p "$WORK_DIR"
menu_output=$(main menu)
[[ "$menu_output" == *'已退出'* ]]
[[ -x "$BIN_PATH" ]]
[[ "$(env -u PRISMCTL_SOURCE_ONLY "$BIN_PATH" version)" == 'dns 1.5.36' ]]

printf 'DNS control command migration tests passed.\n'
