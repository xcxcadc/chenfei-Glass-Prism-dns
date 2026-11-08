#!/usr/bin/env bash
# dns —— Prism DNS 综合控制脚本：安装 / 升级 / 卸载 / 状态 / 诊断 / 传输模式切换
# 覆盖三种角色：面板(panel)、解锁机(proxy)、被解锁机(client)，形成安装→升级→诊断→卸载的闭环。
#
# 设计原则
#   1. 所有子脚本都按发布标签（默认本脚本对应版本）下载，下载后先校验非空与语法，再执行。
#   2. 升级时自动读取本机已保存的面板地址、密钥、传输模式并原样带上，不需要重新输入令牌。
#   3. 只处理 Prism 自有的服务、nft 表和 iptables 链，不触碰 V2bX、XrayR、Docker、nginx 等业务。
set -Eeuo pipefail

PRISMCTL_VERSION="1.5.36"
SCRIPT_REPO="${PRISM_SCRIPT_REPO:-xcxcadc/chenfei-Glass-Prism-dns}"
SCRIPT_REF="${PRISM_SCRIPT_REF:-v${PRISMCTL_VERSION}}"
ETC_DIR="${PRISMCTL_ETC_DIR:-/etc/prismdns}"
ENV_FILE="$ETC_DIR/transport.env"
LIB_DIR="${PRISMCTL_LIB_DIR:-/usr/local/lib/prismdns}"
BIN_PATH="${DNS_BIN:-${PRISMCTL_BIN:-/usr/local/bin/dns}}"
WORK_DIR=""

info() { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
ok() { printf '\033[1;32m[ OK ]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[WARN]\033[0m %s\n' "$*" >&2; }
fail() { printf '\033[1;31m[FAIL]\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<EOF
dns ${PRISMCTL_VERSION} —— Prism DNS 综合控制脚本

用法：dns                直接运行进入数字菜单（1 安装解锁机 … 9 查看日志）
      dns <命令> [角色] [选项]   子命令方式，便于脚本调用

命令：
  install   <panel|proxy|client>   安装
  upgrade   [panel|proxy|client]   升级（省略角色时按本机已安装的角色自动判断）
  uninstall [panel|proxy|client]   卸载（省略角色时自动判断）
  mode      [direct|encrypted]     查看或切换传输模式
  status                           查看本机角色、版本、服务和传输状态
  logs [名称]                      查看日志（agent egress client-egress dns route-sync transport tunnel traffic media controller enhancer all）；-n 行数 -f 跟踪 --errors 只看警告错误
  doctor                           诊断常见故障（防火墙、出口 IP、SNI 过滤等）
  self-install                     把 dns 安装到 ${BIN_PATH}
  self-update                      更新 dns 本身
  version | help

常用选项：
  --master URL       面板地址（升级时省略，自动读取本机配置）
  --secret TOKEN     解锁机节点密钥（install proxy）
  --token TOKEN      被解锁机令牌（install client）
  --ip ADDR          解锁机的真实公网 IPv4（install proxy，可选）
  --smart            解锁机启用智能检测（install proxy，可选）
  --mode MODE        direct 或 encrypted；默认沿用本机已有模式，新装为 direct
  --ref REF          指定发布标签（如 v1.5.36），latest 表示 GitHub 最新发布
  --purge-data       卸载面板时一并删除面板数据
  --yes              跳过确认

示例：
  dns install proxy  --master https://panel.example.com --secret <节点密钥>
  dns install client --master https://panel.example.com --token <被解锁机令牌> --mode encrypted
  dns upgrade --ref latest
  dns mode encrypted
  dns doctor
  dns uninstall --yes

无需预先安装，可直接运行：
  bash <(curl -fsSL https://raw.githubusercontent.com/${SCRIPT_REPO}/refs/tags/${SCRIPT_REF}/dns.sh) <命令> ...
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
  [[ -n "$WORK_DIR" ]] || WORK_DIR=$(mktemp -d /tmp/dns.XXXXXX)
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
  ok "解锁机安装完成。建议运行: dns doctor"
}

install_client() {
  local script
  [[ -n "${OPT_MASTER:-}" && -n "${OPT_TOKEN:-}" ]] || fail "install client 需要 --master 和 --token。"
  info "安装被解锁机（版本 ${SCRIPT_REF}，传输模式 ${OPT_MODE:-direct}）..."
  script=$(fetch_script prismdns.sh)
  PRISM_SCRIPT_REF="$SCRIPT_REF" PRISM_TRANSPORT_MODE="${OPT_MODE:-direct}" \
    bash "$script" --master "$OPT_MASTER" --token "$OPT_TOKEN" --one-click --non-interactive
  ok "被解锁机安装完成。建议运行: dns doctor"
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
  ok "升级完成。运行 dns status 查看结果。"
}

upgrade_panel() {
  local script stamp item rc=0 healthy=false _
  info "升级面板到 ${SCRIPT_REF}（面板数据目录会保留）..."
  script=$(fetch_script enhanced_install.sh)
  # 升级前备份二进制、数据库和 systemd 单元；升级后若面板没有恢复健康，自动按这份备份回滚。
  stamp=$(date +%Y%m%d-%H%M%S)
  for item in /opt/prism/prism-controller /opt/prism/prism-enhancer /opt/prism/data.db               /etc/systemd/system/prism-controller.service /etc/systemd/system/prism-enhancer.service; do
    [[ -f "$item" ]] && cp -a "$item" "${item}.pre-dns-${stamp}"
  done
  info "已备份到 *.pre-dns-${stamp}"
  PRISM_SCRIPT_REF="$SCRIPT_REF" bash "$script" || rc=$?
  if [[ $rc -eq 0 ]]; then
    for _ in {1..20}; do
      if unit_active prism-controller.service && unit_active prism-enhancer.service; then healthy=true; break; fi
      sleep 1
    done
  fi
  if [[ "$healthy" != true ]]; then
    warn "升级后面板未恢复健康，正在按备份自动回滚..."
    for item in /opt/prism/prism-controller /opt/prism/prism-enhancer /opt/prism/data.db                 /etc/systemd/system/prism-controller.service /etc/systemd/system/prism-enhancer.service; do
      [[ -f "${item}.pre-dns-${stamp}" ]] && cp -a "${item}.pre-dns-${stamp}" "$item"
    done
    systemctl daemon-reload
    systemctl restart prism-controller.service prism-enhancer.service || true
    fail "面板升级失败，已回滚到升级前的状态（备份后缀 .pre-dns-${stamp}）。"
  fi
  ok "面板升级成功。"
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
  printf 'dns %s（脚本标签 %s）\n' "$PRISMCTL_VERSION" "$SCRIPT_REF"
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
      id prism-tunnel >/dev/null 2>&1 && d_ok "隧道用户 prism-tunnel 存在" || d_bad "缺少隧道用户 prism-tunnel，请执行 dns mode encrypted 重建"
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
          d_bad "到 ${peer}: 普通域名可达，但 googlevideo 无响应。该网络很可能在过滤明文 SNI（表现为 YouTube 页面能开、视频转圈）。请执行: dns mode encrypted（解锁机与被解锁机两端都要切换）"
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
  local script temporary legacy source_path
  if [[ -e "$BIN_PATH" || -L "$BIN_PATH" ]]; then
    control_script_owned "$BIN_PATH" || fail "${BIN_PATH} 已被其他程序占用，未覆盖。"
  fi
  source_path="${BASH_SOURCE[0]}"
  if [[ -f "$source_path" && "$SCRIPT_REF" == "v$PRISMCTL_VERSION" ]]; then
    script="$source_path"
  else
    script=$(fetch_script dns.sh)
  fi
  control_script_owned "$script" || fail "下载内容不是 Prism DNS 控制脚本，已拒绝安装。"
  bash -n "$script" || fail "控制脚本语法检查失败，已保留原命令。"
  install -d -m 0755 "$(dirname "$BIN_PATH")"
  temporary=$(mktemp "${BIN_PATH}.XXXXXX")
  if ! install -m 0755 "$script" "$temporary" || ! mv -f -- "$temporary" "$BIN_PATH"; then
    rm -f -- "$temporary"
    fail "安装 dns 失败，已保留原命令。"
  fi
  legacy="$(dirname "$BIN_PATH")/prismctl"
  if [[ "$legacy" != "$BIN_PATH" && ( -e "$legacy" || -L "$legacy" ) ]] && control_script_owned "$legacy"; then
    temporary=$(mktemp "${legacy}.XXXXXX")
    rm -f -- "$temporary"
    if ! ln -s -- "$(basename "$BIN_PATH")" "$temporary" || ! mv -f -- "$temporary" "$legacy"; then
      rm -f -- "$temporary"
      warn "dns 已安装，但旧 prismctl 入口迁移失败，请直接运行 dns。"
    fi
  fi
  ok "已安装到 ${BIN_PATH}，之后可直接运行 dns。"
}

control_script_owned() {
  [[ -f "$1" ]] && grep -Eq '^PRISMCTL_(COMPAT_)?VERSION="[0-9]+\.[0-9]+\.[0-9]+"$' "$1" &&
    grep -Fq 'Prism DNS' "$1"
}

# ---------------------------------------------------------------- 日志
# 名称|说明|systemd 单元（可含通配符）
log_catalog() {
  cat <<'LIST'
agent|Agent（prism-agent，解锁机与被解锁机都有）|prism-agent.service
egress|解锁机出口代理（prism-egress-proxy）|prism-egress-proxy.service
client-egress|被解锁机 SNI 出口（prism-client-egress-proxy）|prism-client-egress-proxy.service
dns|被解锁机本地 DNS（prismdns-local-dns）|prismdns-local-dns.service
route-sync|被解锁机路由同步（面板改动在这里落地）|prismdns-route-sync.service
transport|传输同步（隧道与出口策略）|prism-transport.service
tunnel|加密隧道（encrypted 模式）|prism-transport-ssh-*.service
traffic|流量上报|prismdns-traffic.service
media|媒体检测 Worker|prism-media-check.service
controller|面板控制器（prism-controller）|prism-controller.service
enhancer|面板中文增强层（prism-enhancer）|prism-enhancer.service
LIST
}

# 输出本机实际存在的日志项（名称|说明|单元）
available_logs() {
  local name label units unit found
  while IFS="|" read -r name label units; do
    found=false
    for unit in $units; do
      if [[ "$unit" == *"*"* ]]; then
        if systemctl list-units --all --plain --no-legend "$unit" 2>/dev/null | grep -q .; then found=true; fi
      elif unit_exists "$unit"; then
        found=true
      fi
    done
    if [[ "$found" == true ]]; then printf '%s|%s|%s\n' "$name" "$label" "$units"; fi
  done < <(log_catalog)
}

show_log() { # 单元列表 行数 是否跟踪 是否只看错误
  local units="$1" lines="$2" follow="$3" errors="$4" args=() unit
  command -v journalctl >/dev/null 2>&1 || fail "本机没有 journalctl。"
  for unit in $units; do args+=(-u "$unit"); done
  args+=(-n "$lines" --no-pager -o short-iso)
  if [[ "$errors" == true ]]; then args+=(-p warning); fi
  if [[ "$follow" == true ]]; then
    echo "（实时跟踪中，按 Ctrl-C 结束）"
    journalctl "${args[@]}" -f
  else
    journalctl "${args[@]}"
  fi
}

cmd_logs() {
  local target="${1:-}" name label units
  if [[ -z "$target" || "$target" == all ]]; then
    units="prism-agent.service prism-egress-proxy.service prism-client-egress-proxy.service prismdns-local-dns.service prismdns-route-sync.service prism-transport.service prism-controller.service prism-enhancer.service"
    show_log "$units" "${OPT_LINES:-100}" "${OPT_FOLLOW:-false}" "${OPT_ERRORS:-false}"
    return 0
  fi
  while IFS="|" read -r name label units; do
    if [[ "$name" == "$target" ]]; then
      show_log "$units" "${OPT_LINES:-100}" "${OPT_FOLLOW:-false}" "${OPT_ERRORS:-false}"
      return 0
    fi
  done < <(log_catalog)
  fail "未知日志名称: ${target}。可用: agent egress client-egress dns route-sync transport tunnel traffic media controller enhancer all"
}

# ---------------------------------------------------------------- 交互菜单
TTY_DEV="${PRISMCTL_TTY:-/dev/tty}"

ask() { # 变量名 提示 [默认值] [secret]
  local __var="$1" __prompt="$2" __default="${3:-}" __secret="${4:-}" __value=""
  if [[ -n "$__default" ]]; then __prompt="${__prompt} [默认 ${__default}]"; fi
  if [[ "$__secret" == secret ]]; then
    read -r -s -p "${__prompt}: " __value <"$TTY_DEV" || true
    echo
  else
    read -r -p "${__prompt}: " __value <"$TTY_DEV" || true
  fi
  printf -v "$__var" '%s' "${__value:-$__default}"
}

ask_mode() {
  local choice
  echo "  传输模式：1) direct 直连（默认）   2) encrypted 加密隧道（机房过滤 SNI、YouTube 视频转圈时选这个）"
  ask choice "请选择 1/2" 1
  case "$choice" in 2) OPT_MODE=encrypted ;; *) OPT_MODE=direct ;; esac
}

role_label() {
  case "$1" in
    panel) printf '面板' ;;
    proxy) printf '解锁机' ;;
    client) printf '被解锁机' ;;
    agent) printf 'Agent（角色未知）' ;;
    *) printf '%s' "$1" ;;
  esac
}

local_role_text() {
  local roles role out=""
  roles=$(detect_roles)
  if [[ -z "$roles" ]]; then printf '未安装任何 Prism 组件'; return 0; fi
  for role in $roles; do out+="$(role_label "$role") "; done
  printf '%s' "$out"
}

print_menu() {
  local roles mode
  roles=$(local_role_text)
  mode=$(env_value TRANSPORT_MODE)
  echo
  echo "================ Prism DNS 控制台 v${PRISMCTL_VERSION} ================"
  echo "  本机角色：${roles}${mode:+   传输模式：${mode}}"
  echo "  版本标签：${SCRIPT_REF}"
  echo "------------------------------------------------------------"
  echo "  【安装】"
  echo "    1) 安装 解锁机    （提供解锁出口的服务器）"
  echo "    2) 安装 被解锁机  （需要被解锁的服务器）"
  echo "    3) 安装 面板"
  echo "  【升级】"
  echo "    4) 升级本机       （自动判断：面板 / 解锁机 / 被解锁机）"
  echo "  【卸载】"
  echo "    5) 卸载 解锁机"
  echo "    6) 卸载 被解锁机"
  echo "    7) 卸载 面板"
  echo "  【运维】"
  echo "    8) 查看状态"
  echo "    9) 查看日志"
  echo "   10) 故障诊断"
  echo "   11) 切换传输模式   （direct / encrypted）"
  echo "   12) 更新 dns 自身"
  echo "    0) 退出"
  echo "------------------------------------------------------------"
}

menu_logs() {
  local entries=() names=() i=1 choice view n lines follow errors line name label units
  mapfile -t entries < <(available_logs)
  if [[ ${#entries[@]} -eq 0 ]]; then
    echo "本机没有找到 Prism 服务日志。"
    return 0
  fi
  echo "要查看哪个日志？"
  for line in "${entries[@]}"; do
    IFS="|" read -r name label units <<<"$line"
    names+=("$name")
    printf '  %2d) %s\n' "$i" "$label"
    i=$((i + 1))
  done
  printf '  %2d) 全部 Prism 服务（合并）\n' "$i"
  echo "   0) 返回"
  ask choice "请输入编号"
  if [[ -z "$choice" || "$choice" == 0 ]]; then return 0; fi
  if ! [[ "$choice" =~ ^[0-9]+$ ]]; then echo "无效输入。"; return 0; fi
  if [[ "$choice" -eq "$i" ]]; then
    n=all
  elif [[ "$choice" -ge 1 && "$choice" -lt "$i" ]]; then
    n="${names[$((choice - 1))]}"
  else
    echo "无效编号。"
    return 0
  fi
  echo "查看方式：1) 最近 100 行   2) 实时跟踪   3) 只看警告和错误   4) 自定义行数"
  ask view "请选择" 1
  lines=100 follow=false errors=false
  case "$view" in
    2) follow=true; lines=30 ;;
    3) errors=true; lines=200 ;;
    4) ask lines "显示最近多少行" 300; [[ "$lines" =~ ^[0-9]+$ ]] || lines=300 ;;
  esac
  OPT_LINES="$lines" OPT_FOLLOW="$follow" OPT_ERRORS="$errors"
  ( trap - INT; cmd_logs "$n" ) || true
}

# 在子进程里执行，避免动作中的 fail/exit 把整个菜单带走；Ctrl-C 只中断当前动作。
run_action() {
  ( WORK_DIR=""; work_dir >/dev/null; trap cleanup EXIT; trap - INT; "$@" ) || warn "该操作未成功完成（见上方信息）。"
}

menu_install_proxy() {
  local smart
  ask OPT_MASTER "面板地址（例如 https://panel.example.com）" "$(env_value MASTER)"
  ask OPT_SECRET "解锁机节点密钥（在面板里获取）" "" secret
  ask OPT_IP "解锁机真实公网 IPv4（回车=自动探测，NAT 或多 IP 请手填）"
  ask_mode
  ask smart "是否启用智能检测 --smart？(y/N)" N
  if [[ "$smart" == y || "$smart" == Y ]]; then OPT_SMART=true; else OPT_SMART=false; fi
  OPT_MASTER="${OPT_MASTER%/}"
  cmd_install proxy
}

menu_install_client() {
  ask OPT_MASTER "面板地址（例如 https://panel.example.com）" "$(env_value MASTER)"
  ask OPT_TOKEN "被解锁机令牌（在面板「IP 配置」里获取）" "" secret
  ask_mode
  OPT_MASTER="${OPT_MASTER%/}"
  cmd_install client
}

menu_upgrade() {
  local ref
  ask ref "升级到哪个版本？回车=${SCRIPT_REF}，输入 latest=GitHub 最新发布，或输入标签如 v1.5.36"
  if [[ -n "$ref" ]]; then SCRIPT_REF=$(resolve_ref "$ref"); fi
  echo "升级顺序提示：先面板，再解锁机，最后被解锁机。"
  cmd_upgrade ""
}

menu_uninstall() { # 角色
  local role="$1" roles go purge
  roles=" $(detect_roles) "
  if [[ "$roles" != *" $role "* ]]; then
    warn "本机没有检测到「$(role_label "$role")」组件（检测到：$(local_role_text)）。"
    ask go "仍要对「$(role_label "$role")」执行卸载吗？(y/N)" N
    [[ "$go" == y || "$go" == Y ]] || return 0
  fi
  if [[ "$role" == panel ]]; then
    ask purge "是否同时永久删除面板数据？(y/N，默认保留)" N
    if [[ "$purge" == y || "$purge" == Y ]]; then OPT_PURGE=true; else OPT_PURGE=false; fi
  fi
  cmd_uninstall "$role"
}

menu_mode() {
  local choice
  cmd_mode ""
  echo "  1) 切换到 direct 直连   2) 切换到 encrypted 加密隧道   0) 返回"
  ask choice "请选择" 0
  case "$choice" in
    1) cmd_mode direct ;;
    2) cmd_mode encrypted ;;
  esac
}

pause_menu() {
  local _
  read -r -p "按回车返回菜单..." _ <"$TTY_DEV" || true
}

run_menu() {
  [[ -r "$TTY_DEV" ]] || fail "没有可用的终端，无法显示菜单。请直接使用子命令，例如 dns status（dns help 查看全部）。"
  trap ':' INT
  local choice
  while true; do
    print_menu
    ask choice "请输入编号"
    OPT_MASTER="" OPT_SECRET="" OPT_TOKEN="" OPT_IP="" OPT_MODE="" OPT_SMART=false OPT_PURGE=false OPT_YES=false
    OPT_LINES=100 OPT_FOLLOW=false OPT_ERRORS=false
    case "$choice" in
      1) run_action menu_install_proxy ;;
      2) run_action menu_install_client ;;
      3) run_action cmd_install panel ;;
      4) run_action menu_upgrade ;;
      5) run_action menu_uninstall proxy ;;
      6) run_action menu_uninstall client ;;
      7) run_action menu_uninstall panel ;;
      8) run_action cmd_status ;;
      9) menu_logs ;;
      10) run_action cmd_doctor ;;
      11) run_action menu_mode ;;
      12) run_action cmd_self_install ;;
      0|q|Q|exit) echo "已退出。"; return 0 ;;
      "") continue ;;
      *) echo "无效输入：${choice}，请输入 0-12。"; continue ;;
    esac
    pause_menu
  done
}

# ---------------------------------------------------------------- 参数解析
main() {
  OPT_MASTER="" OPT_SECRET="" OPT_TOKEN="" OPT_IP="" OPT_MODE="" OPT_SMART=false OPT_PURGE=false OPT_YES=false
  OPT_LINES=100 OPT_FOLLOW=false OPT_ERRORS=false
  local command="${1:-menu}"
  [[ $# -eq 0 ]] || shift
  case "$command" in
    help|-h|--help) usage; return 0 ;;
    version|-v|--version) echo "dns ${PRISMCTL_VERSION}"; return 0 ;;
    menu|install|upgrade|update|uninstall|remove|mode|status|doctor|logs|log|self-install|self-update) ;;
    *) fail "未知命令: ${command}（运行 dns help 查看用法）" ;;
  esac

  local positional=() ref_arg=""
  while (($#)); do
    case "$1" in
      --master) [[ $# -ge 2 ]] || fail "--master 需要值"; OPT_MASTER="${2%/}"; shift 2 ;;
      --secret) [[ $# -ge 2 ]] || fail "--secret 需要值"; OPT_SECRET="$2"; shift 2 ;;
      --token) [[ $# -ge 2 ]] || fail "--token 需要值"; OPT_TOKEN="$2"; shift 2 ;;
      --ip) [[ $# -ge 2 ]] || fail "--ip 需要值"; OPT_IP="$2"; shift 2 ;;
      --mode) [[ $# -ge 2 ]] || fail "--mode 需要值"; OPT_MODE="$2"; shift 2 ;;
      --ref) [[ $# -ge 2 ]] || fail "--ref 需要值"; ref_arg="$2"; shift 2 ;;
      -n|--lines) [[ $# -ge 2 && "$2" =~ ^[0-9]+$ ]] || fail "--lines 需要数字"; OPT_LINES="$2"; shift 2 ;;
      -f|--follow) OPT_FOLLOW=true; shift ;;
      --errors) OPT_ERRORS=true; shift ;;
      --smart) OPT_SMART=true; shift ;;
      --purge-data) OPT_PURGE=true; shift ;;
      --yes|-y) OPT_YES=true; shift ;;
      -h|--help) usage; return 0 ;;
      -*) fail "未知选项: $1（运行 dns help 查看用法）" ;;
      *) positional+=("$1"); shift ;;
    esac
  done
  [[ -z "$OPT_MODE" ]] || valid_mode "$OPT_MODE"

  require_root
  [[ -z "$ref_arg" ]] || SCRIPT_REF="$ref_arg"
  SCRIPT_REF=$(resolve_ref "$SCRIPT_REF")
  work_dir >/dev/null

  case "$command" in
    menu) cmd_self_install; run_menu ;;
    install) cmd_install "${positional[0]:-}" ;;
    upgrade|update) cmd_upgrade "${positional[0]:-}" ;;
    uninstall|remove) cmd_uninstall "${positional[0]:-}" ;;
    mode) cmd_mode "${positional[0]:-${OPT_MODE:-}}" ;;
    status) cmd_status ;;
    doctor) cmd_doctor ;;
    logs|log) cmd_logs "${positional[0]:-}" ;;
    self-install|self-update) cmd_self_install ;;
  esac
}

[[ -n "${PRISMCTL_SOURCE_ONLY:-}" ]] || main "$@"
