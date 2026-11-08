#!/usr/bin/env bash
# 验证 enhanced_install.sh 升级时沿用现有端口与目录，且显式环境变量优先。
set -Eeuo pipefail
cd "$(dirname "$0")/.."
export PRISM_INSTALL_SOURCE_ONLY=1
failures=0
check() { if [[ "$2" == "$3" ]]; then echo "PASS $1"; else echo "FAIL $1: 期望 [$2] 实际 [$3]"; failures=$((failures + 1)); fi; }

run_case() { # 描述 期望公开端口 期望核心端口 期望数据目录 [环境变量...]
  local label="$1" want_pub="$2" want_core="$3" want_data="$4"; shift 4
  local result
  result=$(
    env "$@" bash -c '
      source ./enhanced_install.sh
      systemctl() {
        [[ "$1" == cat ]] || return 1
        case "$2" in
          prism-enhancer.service) echo "ExecStart=/opt/prism/prism-enhancer --listen 0.0.0.0:58082 --upstream http://127.0.0.1:18081 --data-dir /srv/enh" ;;
          prism-controller.service) echo "ExecStart=/opt/prism/prism-controller --host 127.0.0.1 --port 18081" ;;
          *) return 1 ;;
        esac
      }
      detect_existing_config
      echo "$PUBLIC_PORT $CORE_PORT $DATA_DIR"
    '
  )
  check "$label" "$want_pub $want_core $want_data" "$result"
}

run_case "沿用现有端口与数据目录" 58082 18081 /srv/enh A=1
run_case "显式 PRISM_PORT 优先" 9000 18081 /srv/enh PRISM_PORT=9000
run_case "显式 PRISM_CORE_PORT 与数据目录优先" 58082 19000 /tmp/x PRISM_CORE_PORT=19000 PRISM_ENHANCER_DATA_DIR=/tmp/x

# 没有已安装单元时保持默认值
result=$(bash -c 'source ./enhanced_install.sh; systemctl() { return 1; }; detect_existing_config; echo "$PUBLIC_PORT $CORE_PORT"')
check "全新安装保持默认端口" "8080 18080" "$result"

if [[ $failures -eq 0 ]]; then echo "enhanced-install config tests passed"; else echo "$failures test(s) failed"; exit 1; fi
