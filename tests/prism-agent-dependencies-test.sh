#!/usr/bin/env bash
set -euo pipefail

source <(sed -n '/^ensure_dependencies() {$/,/^}$/p' agent_install.sh)
source <(sed -n '/^main() {$/,/^}$/p' agent_install.sh)

step() { :; }
error() { printf '%s\n' "$*" >&2; exit 1; }

package_manager=apt-get
ss_ready=false
install_ss=true
dependency_calls=''

command() {
  [[ $# == 2 && "$1" == -v ]] || return 1
  case "$2" in
    ss) [[ "$ss_ready" == true ]] ;;
    apt-get|dnf|yum) [[ "$2" == "$package_manager" ]] ;;
    *) return 1 ;;
  esac
}

apt-get() {
  dependency_calls+="apt-get $*;"
  if [[ "$1" == install ]]; then
    [[ "${DEBIAN_FRONTEND:-}" == noninteractive ]] || exit 1
    [[ "$install_ss" != true ]] || ss_ready=true
  fi
  return 0
}

dnf() {
  dependency_calls+="dnf $*;"
  [[ "$install_ss" != true ]] || ss_ready=true
  return 0
}

yum() {
  dependency_calls+="yum $*;"
  [[ "$install_ss" != true ]] || ss_ready=true
  return 0
}

for package_manager in apt-get dnf yum; do
  ss_ready=false
  dependency_calls=''
  ensure_dependencies
  [[ "$ss_ready" == true ]]
  case "$package_manager" in
    apt-get) [[ "$dependency_calls" == 'apt-get update -qq;apt-get install -y -qq iproute2;' ]] ;;
    dnf|yum) [[ "$dependency_calls" == "$package_manager install -y iproute;" ]] ;;
  esac
done

dependency_calls=''
ensure_dependencies
[[ -z "$dependency_calls" ]]

ss_ready=false
package_manager=unsupported
if (ensure_dependencies >/dev/null 2>&1); then
  printf 'Missing ss was accepted without a supported package manager\n' >&2
  exit 1
fi

package_manager=apt-get
install_ss=false
if (ensure_dependencies >/dev/null 2>&1); then
  printf 'Dependency installation succeeded without providing ss\n' >&2
  exit 1
fi

install_ss=true
dependency_calls=''
download_started=false
service_started=false
show_banner() { :; }
check_root() { :; }
parse_args() { UNINSTALL_MODE=false; }
detect_system() { :; }
download_binary() {
  [[ "$ss_ready" == true ]] || exit 1
  download_started=true
}
configure_service() { :; }
prepare_legacy_conflicts() { :; }
start_service() {
  [[ "$ss_ready" == true ]] || exit 1
  service_started=true
}
install_proxy_watchdog() { :; }
install_proxy_transport() { :; }
show_result() { :; }

main
[[ "$download_started" == true && "$service_started" == true ]]
[[ "$dependency_calls" == 'apt-get update -qq;apt-get install -y -qq iproute2;' ]]

printf 'Prism Agent dependency tests passed.\n'
