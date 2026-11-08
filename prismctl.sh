#!/usr/bin/env bash
# prismctl —— Prism DNS 综合控制脚本：安装 / 升级 / 卸载 / 状态 / 诊断 / 传输模式切换
# 覆盖三种角色：面板(panel)、解锁机(proxy)、被解锁机(client)，形成安装→升级→诊断→卸载的闭环。
#
# 设计原则
#   1. 所有子脚本都按发布标签（默认本脚本对应版本）下载，下载后先校验非空与语法，再执行。
#   2. 升级时自动读取本机已保存的面板地址、密钥、传输模式并原样带上，不需要重新输入令牌。
#   3. 只处理 Prism 自有的服务、nft 表和 iptables 链，不触碰 V2bX、XrayR、Docker、nginx 等业务。
set -Eeuo pipefail

PRISMCTL_VERSION="1.5.33"
SCRIPT_REPO="${PRISM_SCRIPT_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"
SCRIPT_REF="${PRISM_SCRIPT_REF:-v${PRISMCTL_VERSION}}"
ETC_DIR="${PRISMCTL_ETC_DIR:-/etc/prismdns}"
ENV_FILE="$ETC_DIR/transport.env"
LIB_DIR="${PRISMCTL_LIB_DIR:-/usr/local/lib/prismdns}"
BIN_PATH="${PRISMCTL_BIN:-/usr/local/bin/prismctl}"
WORK_DIR=""

info() { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
ok() { printf '\033[1;32m[ OK ]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[WARN]\033[0m %s\n' "$*" >&2; }
fail() { printf '\033[1;31m[FAIL]\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<EOF
prismctl ${PRISMCTL_VERSION} —— Prism DNS 综合控制脚本

用法：prismctl <命令> [角色] [选项]

命令：
  install   <panel|proxy|client>   安装
  upgrade   [panel|proxy|client]   升级（省略角色时按本机已安装的角色自动判断）
  uninstall [panel|proxy|client]   卸载（省略角色时自动判断）
  mode      [direct|encrypted]     查看或切换传输模式
  status                           查看本机角色、版本、服务和传输状态
  doctor                           诊断常见故障（防火墙、出口 IP、SNI 过滤等）
  self-install                     把 prismctl 安装到 ${BIN_PATH}
  self-update                      更新 prismctl 本身
  version | help

常用选项：
  --master URL       面板地址（升级时省略，自动读取本机配置）
  --secret TOKEN     解锁机节点密钥（install proxy）
  --token TOKEN      被解锁机令牌（install client）
  --ip ADDR          解锁机的真实公网 IPv4（install proxy，可选）
  --smart            解锁机启用智能检测（install proxy，可选）
  --mode MODE        direct 或 encrypted；默认沿用本机已有模式，新装为 direct
  --ref REF          指定发布标签（如 v1.5.33），latest 表示 GitHub 最新发布
  --purge-data       卸载面板时一并删除面板数据
  --yes              跳过确认

示例：
  prismctl install proxy  --master https://panel.example.com --secret <节点密钥>
  prismctl install client --master https://panel.example.com --token <被解锁机令牌> --mode encrypted
  prismctl upgrade --ref latest
  prismctl mode encrypted
  prismctl doctor
  prismctl uninstall --yes

无需预先安装，可直接运行：
  curl -fsSL https://raw.githubusercontent.com/${SCRIPT_REPO}/refs/tags/${SCRIPT_REF}/prismctl.sh | sudo bash -s -- <命令> ...
EOF
}

cleanup() { [[ -z "$WORK_DIR" ]] || rm -rf -- "$WORK_DIR"; }
trap cleanup EXIT

require_root() {
  [[ ${EUID:-$(id -u)} -eq 0 ]] || fail "需要 root 权限，请使用 sudo 运行。"
  command -v curl >/dev/null 2>&1 || fail "缺少 curl，请先安装。"
  command -v systemctl >/dev/null 2>&1 || fail "当前系统不支持 systemd。"
}

work_dir() {
  [[ -n "$WORK_DIR" ]] || WORK_DIR=$(mktemp -d /tmp/prismctl.XXXXXX)
  printf '%s' "$WORK_DIR"
}

resolve_ref() {
  local ref="$1" tag
  if [[ "$ref" == latest ]]; then
    tag=$(curl -fsSL --connect-timeout 8 --max-time 20 "https://api.github.com/repos/${SCRIPT_REPO}/releases/latest" \
      | sed -n 's/^[[:space:]]*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
    [[ -n "$tag" ]] || fail "无法获取 GitHub 最新发布版本。"
    ref="$tag"
  fi
  [[ "$ref" =~ ^[A-Za-z0-9._-]+$ ]] || fail "无效的发布标签: $ref"
  printf '%s' "$ref"
}

# 下载指定发布标签中的脚本到临时目录，校验后返回路径。
fetch_script() {
  local file="$1" target
  target="$(work_dir)/$file"
  curl -fsSL --connect-timeout 8 --max-time 60 \
    "https://raw.githubusercontent.com/${SCRIPT_REPO}/refs/tags/${SCRIPT_REF}/${file}" -o "$target" \
    || fail "下载 ${file} 失败（标签 ${SCRIPT_REF}）。请检查网络或标签是否存在。"
  [[ -s "$target" ]] || fail "${file} 内容为空。"
  bash -n "$target" || fail "${file} 语法检查失败，已拒绝执行。"
  printf '%s' "$target"
}

env_value() {
  local key="$1"
  [[ -f "$ENV_FILE" ]] || return 0
  ( set +u; . "$ENV_FILE" >/dev/null 2>&1; printf '%s' "${!key:-}" )
}

unit_exists() { systemctl cat "$1" >/dev/null 2>&1; }
unit_active() { systemctl is-active --quiet "$1" 2>/dev/null; }

has_panel() { unit_exists prism-controller.service || [[ -x /opt/prism/prism-controller ]]; }

# 输出本机已安装的角色（可多个，空格分隔）
detect_roles() {
  local roles=() role
  has_panel && roles+=(panel)
  role=$(env_value ROLE)
  case "$role" in proxy|client) roles+=("$role") ;; esac
  if [[ ${#roles[@]} -eq 0 ]] && unit_exists prism-agent.service; then roles+=(agent); fi
  printf '%s' "${roles[*]:-}"
}

agent_exec_line() { systemctl cat prism-agent.service 2>/dev/null | sed -n 's/^ExecStart=//p' | head -n 1 | tr -d '"'; }

agent_arg() {
  local name="$1" line
  line=$(agent_exec_line)
  [[ -n "$line" ]] || return 0
  printf '%s' "$line" | sed -nE "s/.*--${name}[= ]([^ ]+).*/\\1/p"
}

agent_has_flag() { agent_exec_line | grep -Eq -- "(^| )--$1( |$)"; }

current_mode() {
  local mode
  mode=$(env_value TRANSPORT_MODE)
  printf '%s' "${mode:-direct}"
}

confirm() {
  [[ "${OPT_YES:-false}" == true ]] && return 0
  local answer
  read -r -p "$1 [y/N] " answer </dev/tty || return 1
  [[ "$answer" == y || "$answer" == Y ]]
}

valid_mode() { [[ "$1" == direct || "$1" == encrypted ]] || fail "传输模式只能是 direct 或 encrypted，收到: $1"; }

# ---------------------------------------------------------------- 安装
cmd_install() {
  local role="${1:-}"
  case "$role" in
    panel) install_panel ;;
    proxy) install_proxy ;;
    client) install_client ;;
    *) fail "install 需要指定角色: panel | proxy | client" ;;
  esac
}

install_panel() {
  local script
  info "安装面板（控制器 + 中文增强层，版本 ${SCRIPT_REF}）..."
  script=$(fetch_script enhanced_install.sh)
  PRISM_SCRIPT_REF="$SCRIPT_REF" bash "$script"
  ok "面板安装完成。"
}

install_proxy() {
  local script args=()
  [[ -n "${OPT_MASTER:-}" && -n "${OPT_SECRET:-}" ]] || fail "install proxy 需要 --master 和 --secret。"
  info "安装解锁机（版本 ${SCRIPT_REF}，传输模式 ${OPT_MODE:-direct}）..."
  script=$(fetch_script agent_install.sh)
  args=(--master "$OPT_MASTER" --secret "$OPT_SECRET")
  [[ -z "${OPT_IP:-}" ]] || args+=(--ip "$OPT_IP")
  [[ "${OPT_SMART:-false}" != true ]] || args+=(--smart)
  PRISM_SCRIPT_REF="$SCRIPT_REF" PRISM_TRANSPORT_MODE="${OPT_MODE:-direct}" bash "$script" "${args[@]}"
  ok "解锁机安装完成。建议运行: prismctl doctor"
}

install_client() {
  local script
  [[ -n "${OPT_MASTER:-}" && -n "${OPT_TOKEN:-}" ]] || fail "install client 需要 --master 和 --token。"
  info "安装被解锁机（版本 ${SCRIPT_REF}，传输模式 ${OPT_MODE:-direct}）..."
  script=$(fetch_script prismdns.sh)
  PRISM_SCRIPT_REF="$SCRIPT_REF" PRISM_TRANSPORT_MODE="${OPT_MODE:-direct}" \
    bash "$script" --master "$OPT_MASTER" --token "$OPT_TOKEN" --one-click --non-interactive
  ok "被解锁机安装完成。建议运行: prismctl doctor"
}

# ---------------------------------------------------------------- 升级
cmd_upgrade() {
  local requested="${1:-}" roles role
  if [[ -n "$requested" ]]; then
    roles="$requested"
  else
    roles=$(detect_roles)
  fi
  [[ -n "$roles" ]] || fail "本机未检测到 Prism 组件，请先使用 install。"
  for role in $roles; do
    case "$role" in
      panel) upgrade_panel ;;
      proxy|agent) upgrade_proxy ;;
      client) upgrade_client ;;
      *) fail "未知角色: $role" ;;
    esac
  done
  ok "升级完成。运行 prismctl status 查看结果。"
}

upgrade_panel() {
  local script
  info "升级面板到 ${SCRIPT_REF}（面板数据目录会保留）..."
  script=$(fetch_script enhanced_install.sh)
  # 升级前备份二进制与数据库，失败时可原样回滚（数据库也会由安装脚本再备份一份）。
  local stamp binary
  stamp=$(date +%Y%m%d-%H%M%S)
  for binary in /opt/prism/prism-controller /opt/prism/prism-enhancer; do
    [[ -f "$binary" ]] && cp -a "$binary" "${binary}.pre-prismctl-${stamp}"
  done
  [[ -f /opt/prism/data.db ]] && cp -a /opt/prism/data.db "/opt/prism/data.db.pre-prismctl-${stamp}"
  info "已备份到 /opt/prism/*.pre-prismctl-${stamp}"
  PRISM_SCRIPT_REF="$SCRIPT_REF" bash "$script"
}

upgrade_proxy() {
  local script master secret ip mode args=()
  master="${OPT_MASTER:-$(env_value MASTER)}"
  secret="${OPT_SECRET:-$(env_value CREDENTIAL)}"
  mode="${OPT_MODE:-$(current_mode)}"
  ip=$(agent_arg ip)
  [[ -n "$master" && -n "$secret" ]] || fail "读取不到本机的面板地址或节点密钥，请显式传入 --master 与 --secret。"
  valid_mode "$mode"
  info "升级解锁机到 ${SCRIPT_REF}（沿用传输模式 ${mode}${ip:+，出口 IP ${ip}}）..."
  script=$(fetch_script agent_install.sh)
  args=(--master "$master" --secret "$secret")
  [[ -z "$ip" ]] || args+=(--ip "$ip")
  ! agent_has_flag smart || args+=(--smart)
  PRISM_SCRIPT_REF="$SCRIPT_REF" PRISM_TRANSPORT_MODE="$mode" bash "$script" "${args[@]}"
}

upgrade_client() {
  local script master mode
  master="${OPT_MASTER:-$(env_value MASTER)}"
  mode="${OPT_MODE:-$(current_mode)}"
  [[ -n "$master" ]] || fail "读取不到本机的面板地址，请显式传入 --master。"
  valid_mode "$mode"
  info "升级被解锁机到 ${SCRIPT_REF}（沿用传输模式 ${mode}）..."
  script=$(fetch_script prismdns.sh)
  PRISM_SCRIPT_REF="$SCRIPT_REF" PRISM_TRANSPORT_MODE="$mode" \
    bash "$script" --master "$master" --refresh-runtime --non-interactive
}

# ---------------------------------------------------------------- 卸载
cmd_uninstall() {
  local requested="${1:-}" roles role script args
  if [[ -n "$requested" ]]; then roles="$requested"; else roles=$(detect_roles); fi
  [[ -n "$roles" ]] || fail "本机未检测到 Prism 组件，无需卸载。"
  confirm "将卸载本机的 Prism 组件: ${roles}（不影响 V2bX、Docker、nginx 等业务）。继续吗？" || fail "已取消。"
  script=$(fetch_script uninstall.sh)
  for role in $roles; do
    args=("$role" --yes)
    if [[ "$role" == panel && "${OPT_PURGE:-false}" == true ]]; then args+=(--purge-data); fi
    info "卸载 ${role} ..."
    bash "$script" "${args[@]}"
  done
  ok "卸载完成。"
}

# ---------------------------------------------------------------- 传输模式
cmd_mode() {
  local target="${1:-}" role master credential flag script
  role=$(env_value ROLE)
  if [[ -z "$target" ]]; then
    printf '角色=%s 传输模式=%s\n' "${role:-未安装}" "$(current_mode)"
    return 0
  fi
  valid_mode "$target"
  case "$role" in
    proxy) flag=--secret ;;
    client) flag=--token ;;
    *) fail "本机不是解锁机或被解锁机，无法切换传输模式。" ;;
  esac
  master=$(env_value MASTER)
  credential=$(env_value CREDENTIAL)
  [[ -n "$master" && -n "$credential" ]] || fail "读取不到本机传输配置。"
  info "切换传输模式: $(current_mode) -> ${target}"
  script=$(fetch_script prism_transport.sh)
  PRISM_SCRIPT_REF="$SCRIPT_REF" bash "$script" "--${role}" "$flag" "$credential" --master "$master" --mode "$target"
  ok "已切换。请先切解锁机、再切被解锁机，两端模式必须一致。"
}

# ---------------------------------------------------------------- 状态
unit_line() {
  local unit="$1" label="$2"
  if unit_exists "$unit"; then
    printf '  %-34s %s\n' "$label" "$(systemctl is-active "$unit" 2>/dev/null || true)"
  fi
}

cmd_status() {
  local roles agent_version
  roles=$(detect_roles)
  printf 'prismctl %s（脚本标签 %s）\n' "$PRISMCTL_VERSION" "$SCRIPT_REF"
  printf '本机角色: %s\n' "${roles:-未检测到 Prism 组件}"
  if [[ -x /usr/local/bin/prism-agent ]]; then
    agent_version=$(/usr/local/bin/prism-agent -version 2>&1 | head -n 1 || true)
    printf 'Agent 版本: %s\n' "${agent_version:-未知}"
  fi
  [[ -z "$(env_value ROLE)" ]] || printf '传输配置: 角色=%s 模式=%s 面板=%s\n' "$(env_value ROLE)" "$(current_mode)" "$(env_value MASTER)"
  echo "Prism 服务:"
  unit_line prism-controller.service "控制器 prism-controller"
  unit_line prism-enhancer.service "增强层 prism-enhancer"
  unit_line prism-agent.service "Agent prism-agent"
  unit_line prism-egress-proxy.service "解锁机出口代理"
  unit_line prism-client-egress-proxy.service "被解锁机 SNI 出口"
  unit_line prismdns-local-dns.service "本地 DNS prismdns-local-dns"
  unit_line prism-transport.timer "传输同步 timer"
  echo "业务服务（非 Prism 管理，仅展示）:"
  local unit
  for unit in V2bX XrayR v2node docker nginx; do
    unit_line "${unit}.service" "$unit"
  done
  if [[ -x "$LIB_DIR/prism_transport.sh" ]]; then
    echo "传输状态:"
    "$LIB_DIR/prism_transport.sh" --status 2>/dev/null | sed 's/^/  /' || true
  fi
  if grep -Eq '^nameserver[[:space:]]+(127\.0\.0\.1|::1)' /etc/resolv.conf 2>/dev/null; then
    echo "系统 DNS: 已由 Prism 接管 (127.0.0.1)"
  else
    echo "系统 DNS: 未接管"
  fi
}

# ---------------------------------------------------------------- 诊断
DOCTOR_FAIL=0
d_ok() { printf '  \033[1;32m通过\033[0m  %s\n' "$*"; }
d_bad() { printf '  \033[1;31m异常\033[0m  %s\n' "$*"; DOCTOR_FAIL=$((DOCTOR_FAIL + 1)); }
d_note() { printf '  \033[1;33m提示\033[0m  %s\n' "$*"; }

sni_code() { # IP SNI
  curl -4 -sk --connect-timeout 5 -m 8 -o /dev/null -w '%{http_code}' --resolve "$2:443:$1" "https://$2/" 2>/dev/null || true
}

local_ipv4s() { ip -4 -o addr show scope global 2>/dev/null | awk '{sub(/\/.*/,"",$4); print $4}'; }

cmd_doctor() {
  local role mode unit configured peers peer plain video
  role=$(env_value ROLE)
  mode=$(current_mode)
  echo "角色: ${role:-$(detect_roles)}  传输模式: ${mode}"
  echo "Prism 服务:"
  for unit in prism-agent.service prismdns-local-dns.service prism-client-egress-proxy.service prism-egress-proxy.service; do
    if unit_exists "$unit"; then
      if unit_active "$unit"; then d_ok "$unit 运行中"; else d_bad "$unit 未运行（systemctl status $unit）"; fi
    fi
  done

  if [[ "$role" == proxy ]]; then
    echo "解锁机检查:"
    configured=$(agent_arg ip)
    if [[ -n "$configured" ]]; then
      if local_ipv4s | grep -qx -- "$configured"; then
        d_ok "Agent --ip ${configured} 是本机网卡上的地址"
      else
        d_bad "Agent --ip ${configured} 不在本机网卡上。若这是 NAT 或旧地址，面板会把它记为出口，被解锁机会解析到不可达地址。请改成真实公网 IPv4 后重装。"
      fi
    else
      d_note "Agent 未指定 --ip，使用自动探测的出口地址。"
    fi
    if command -v ss >/dev/null 2>&1; then
      if ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq '(:|\])443$'; then d_ok "TCP 443 已在监听"; else d_bad "TCP 443 没有监听"; fi
    fi
    if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -q '^Status: active'; then
      if ufw status 2>/dev/null | grep -Eq '(^|[[:space:]])443(/tcp)?[[:space:]]+ALLOW'; then
        d_ok "ufw 已放行 443"
      else
        d_bad "ufw 已启用但未放行 443。执行: ufw allow 443/tcp（Prism 授权白名单仍会挡住未授权来源）"
      fi
    fi
    if [[ "$mode" == encrypted ]]; then
      id prism-tunnel >/dev/null 2>&1 && d_ok "隧道用户 prism-tunnel 存在" || d_bad "缺少隧道用户 prism-tunnel，请执行 prismctl mode encrypted 重建"
    fi
  fi

  if [[ "$role" == client ]]; then
    echo "被解锁机检查:"
    if grep -Eq '^nameserver[[:space:]]+(127\.0\.0\.1|::1)' /etc/resolv.conf 2>/dev/null; then d_ok "系统 DNS 已接管"; else d_bad "系统 DNS 未指向 127.0.0.1"; fi
    peers=""
    if [[ -f "$ETC_DIR/dnsmasq-routes.conf" ]]; then
      peers=$(sed -n 's#^address=/[^/]*/\([0-9.]*\)$#\1#p' "$ETC_DIR/dnsmasq-routes.conf" | sort -u | grep -v '^127\.' | head -n 3 || true)
    fi
    if [[ -z "$peers" ]]; then
      d_note "当前没有下发任何路由，跳过出口连通性检查。"
    elif [[ "$mode" == encrypted ]]; then
      d_note "加密模式下，到解锁机的流量经本机隧道转发，明文 SNI 检查不适用；请直接实测业务。"
    else
      for peer in $peers; do
        plain=$(sni_code "$peer" www.example.com)
        video=$(sni_code "$peer" redirector.googlevideo.com)
        if [[ "$plain" == 000 || -z "$plain" ]]; then
          d_bad "到解锁机 ${peer}:443 不通。检查解锁机防火墙是否放行 443、Agent --ip 是否正确。"
        elif [[ "$video" == 000 || -z "$video" ]]; then
          d_bad "到 ${peer}: 普通域名可达，但 googlevideo 无响应。该网络很可能在过滤明文 SNI（表现为 YouTube 页面能开、视频转圈）。请执行: prismctl mode encrypted（解锁机与被解锁机两端都要切换）"
        else
          d_ok "到 ${peer}: 普通域名与视频域名均可达"
        fi
      done
    fi
  fi

  echo "---"
  if [[ "$DOCTOR_FAIL" -eq 0 ]]; then ok "未发现异常。"; else fail "发现 ${DOCTOR_FAIL} 项异常，见上方建议。"; fi
}

# ---------------------------------------------------------------- 自身安装与更新
cmd_self_install() {
  local script
  script=$(fetch_script prismctl.sh)
  install -m 0755 "$script" "$BIN_PATH"
  ok "已安装到 ${BIN_PATH}，之后可直接运行 prismctl。"
}

# ---------------------------------------------------------------- 参数解析
main() {
  local command="${1:-help}"
  [[ $# -eq 0 ]] || shift
  case "$command" in
    help|-h|--help) usage; return 0 ;;
    version|-v|--version) echo "prismctl ${PRISMCTL_VERSION}"; return 0 ;;
    install|upgrade|update|uninstall|remove|mode|status|doctor|self-install|self-update) ;;
    *) fail "未知命令: ${command}（运行 prismctl help 查看用法）" ;;
  esac

  local positional=()
  OPT_MASTER="" OPT_SECRET="" OPT_TOKEN="" OPT_IP="" OPT_MODE="" OPT_SMART=false OPT_PURGE=false OPT_YES=false
  local ref_arg=""
  while (($#)); do
    case "$1" in
      --master) [[ $# -ge 2 ]] || fail "--master 需要值"; OPT_MASTER="${2%/}"; shift 2 ;;
      --secret) [[ $# -ge 2 ]] || fail "--secret 需要值"; OPT_SECRET="$2"; shift 2 ;;
      --token) [[ $# -ge 2 ]] || fail "--token 需要值"; OPT_TOKEN="$2"; shift 2 ;;
      --ip) [[ $# -ge 2 ]] || fail "--ip 需要值"; OPT_IP="$2"; shift 2 ;;
      --mode) [[ $# -ge 2 ]] || fail "--mode 需要值"; OPT_MODE="$2"; shift 2 ;;
      --ref) [[ $# -ge 2 ]] || fail "--ref 需要值"; ref_arg="$2"; shift 2 ;;
      --smart) OPT_SMART=true; shift ;;
      --purge-data) OPT_PURGE=true; shift ;;
      --yes|-y) OPT_YES=true; shift ;;
      -h|--help) usage; return 0 ;;
      -*) fail "未知选项: $1（运行 prismctl help 查看用法）" ;;
      *) positional+=("$1"); shift ;;
    esac
  done
  [[ -z "$OPT_MODE" ]] || valid_mode "$OPT_MODE"

  require_root
  [[ -z "$ref_arg" ]] || SCRIPT_REF="$ref_arg"
  SCRIPT_REF=$(resolve_ref "$SCRIPT_REF")

  case "$command" in
    install) cmd_install "${positional[0]:-}" ;;
    upgrade|update) cmd_upgrade "${positional[0]:-}" ;;
    uninstall|remove) cmd_uninstall "${positional[0]:-}" ;;
    mode) cmd_mode "${positional[0]:-${OPT_MODE:-}}" ;;
    status) cmd_status ;;
    doctor) cmd_doctor ;;
    self-install|self-update) cmd_self_install ;;
    *) fail "未知命令: ${command}（运行 prismctl help 查看用法）" ;;
  esac
}

[[ -n "${PRISMCTL_SOURCE_ONLY:-}" ]] || main "$@"
