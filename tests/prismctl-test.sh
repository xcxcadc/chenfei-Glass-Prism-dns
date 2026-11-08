#!/usr/bin/env bash
# prismctl 单元测试：不需要 root，不访问网络，不触碰真实系统。
set -Eeuo pipefail
cd "$(dirname "$0")/.."
export PRISMCTL_SOURCE_ONLY=1
# shellcheck disable=SC1091
source ./prismctl.sh
set +e

failures=0
check() { # 描述 期望 实际
  if [[ "$2" == "$3" ]]; then printf 'PASS %s\n' "$1"; else printf 'FAIL %s: 期望 [%s] 实际 [%s]\n' "$1" "$2" "$3"; failures=$((failures + 1)); fi
}
check_contains() {
  if [[ "$3" == *"$2"* ]]; then printf 'PASS %s\n' "$1"; else printf 'FAIL %s: 输出中没有 [%s]\n' "$1" "$2"; failures=$((failures + 1)); fi
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
ETC_DIR="$tmp/etc"; ENV_FILE="$ETC_DIR/transport.env"; mkdir -p "$ETC_DIR"

# 无配置时
check "无配置时传输模式默认 direct" direct "$(current_mode)"
check "无配置时 env_value 为空" "" "$(env_value ROLE)"

# 读取已保存配置
printf 'ROLE=proxy\nMASTER=https://panel.example.com\nCREDENTIAL=secret123\nTRANSPORT_MODE=encrypted\n' >"$ENV_FILE"
check "读取 ROLE" proxy "$(env_value ROLE)"
check "读取 MASTER" https://panel.example.com "$(env_value MASTER)"
check "读取 CREDENTIAL" secret123 "$(env_value CREDENTIAL)"
check "读取传输模式" encrypted "$(current_mode)"

# printf %q 引用过的值也必须能读回
printf 'ROLE=client\nMASTER=https://a.example.com/x\ y\nCREDENTIAL=t\nTRANSPORT_MODE=direct\n' >"$ENV_FILE"
check "含转义的值可读回" 'https://a.example.com/x y' "$(env_value MASTER)"

# 角色探测：stub 掉 systemctl，避免依赖真实系统
systemctl() { return 1; }
has_panel() { return 1; }
printf 'ROLE=client\n' >"$ENV_FILE"
check "探测被解锁机角色" client "$(detect_roles)"
has_panel() { return 0; }
check "面板机同时是被解锁机" "panel client" "$(detect_roles)"
rm -f "$ENV_FILE"; has_panel() { return 1; }
check "无组件时角色为空" "" "$(detect_roles)"

# Agent 启动参数解析
agent_exec_line() { printf '/usr/local/bin/prism-agent --master https://p.example.com --secret abc --ip 203.0.113.9 --smart'; }
check "解析 --ip" 203.0.113.9 "$(agent_arg ip)"
check "解析 --secret" abc "$(agent_arg secret)"
agent_has_flag smart && smart=yes || smart=no
check "识别 --smart" yes "$smart"
agent_exec_line() { printf '/usr/local/bin/prism-agent --master https://p.example.com --secret abc'; }
check "无 --ip 时为空" "" "$(agent_arg ip)"
agent_has_flag smart && smart=yes || smart=no
check "无 --smart 时识别为否" no "$smart"

# 发布标签校验
check "合法标签通过" v1.5.35 "$(resolve_ref v1.5.35)"
out=$( (resolve_ref 'v1;rm -rf /') 2>&1 ); rc=$?
check "非法标签被拒绝" 1 "$rc"
check_contains "非法标签有提示" "无效的发布标签" "$out"

# 传输模式校验
out=$( (valid_mode weird) 2>&1 ); rc=$?
check "非法模式被拒绝" 1 "$rc"
valid_mode encrypted; check "合法模式 encrypted" 0 "$?"

# 帮助与版本、参数错误
out=$(main help 2>&1); check_contains "帮助包含 install" "install" "$out"; check_contains "帮助包含 doctor" "doctor" "$out"
out=$(main version 2>&1); check_contains "版本输出" "prismctl" "$out"
out=$( (main --bogus-flag) 2>&1 ); rc=$?
check_contains "未知命令/选项有提示" "未知" "$out"

if [[ $failures -eq 0 ]]; then echo "prismctl tests passed"; else echo "$failures prismctl test(s) failed"; exit 1; fi
