# chenfei Glass Prism DNS

基于 DNS 与智能 SNI 出口代理的流媒体与 AI 服务解锁管理面板（Liquid Glass 风格中文增强版）。轻量、非侵入式部署，不干扰机器上已有的 XrayR / V2bX / sing-box / Docker / Nginx 等业务服务。

---

## ⚡ 极简一键安装与节点对接

兼容 **Debian / Ubuntu / CentOS / Rocky / Alma / Fedora / Alpine / Arch**。

### 1. 一键安装面板 / 打开综合控制菜单
```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/i.sh | bash
```

### 2. 一键对接任意解锁机或被解锁机
在面板 **「节点管理 → 安装命令」** 或 **「IP 配置 → 客户端脚本」** 中直接复制极简短链执行即可：
```bash
curl -fsSL https://你的面板域名/i/<节点密钥或Token> | bash
```
- 自动识别该凭证属于 **解锁机 (Proxy)**、**受管 IP 被解锁机 (Client)** 或 **手动 DNS 节点**。
- 非 `root` 自动 `sudo` 提权；优先从面板自托管 `/enhancer/dist/` 拉取组件，免受 GitHub 访问限制。
- 自动处理 Ubuntu / Debian / Fedora 的 `systemd-resolved`（53 端口冲突）及 `dnsmasq` / `unbound` / `bind9` 等冲突服务，自动备份并恢复系统 DNS。

### 3. 常用运维命令（安装后可直接使用 `dns`）
```bash
dns                  # 打开交互式控制菜单
dns status           # 查看本机角色、组件状态与传输模式
dns upgrade          # 一键升级至最新正式版（保留现有数据与配置）
dns mode encrypted   # 切换加密 SNI 隧道模式（解决机房过滤明文 SNI）
dns warp status      # 查看本机非全局 WARP 分流状态
dns doctor           # 一键自检端口、防火墙与解锁链路
dns uninstall        # 卸载 Prism 组件并恢复系统原 DNS
```

---

## ✨ 核心特性（v1.6.x）

- **按服务可控非全局 WARP 分流（全自动启停）**
  - 面板端可为任意解锁机的**任意内置或自定义服务**单独切换出口模式：`原生直连` / `☁ WARP 智能优选` / `WARP IPv6` / `WARP IPv4` / `自动防送中兜底`，绝不影响同一解锁机上其他服务的原生解锁能力。
  - **零手动 SSH**：面板开启任意服务的 WARP 后，解锁机在 5 秒内自动部署/拉起纯本地用户态 `wireproxy`（`127.0.0.1:40000`，**不改写宿主机内核默认路由**）；全部恢复原生后自动关闭。
  - **高可用熔断与自愈**：WARP 拨号采用 4 秒独立短超时，异常时立即使用全新上下文回退原生出口并触发 25 秒熔断；每 30 秒真实数据面探活，握手卡死自动重启或切换 Cloudflare 备用端口自愈。
- **被解锁机默认原生直连检测 & 智能推荐闭环**
  - 在服务编排或 IP 配置中检测/选择服务时，自动并发执行「被解锁机原生直连探测（绕过本机代理 DNS）」与「全解锁机实时秒测」：
    - 本机原生直连即可解锁 → 突出推荐 **采用原生直连**（一键移出代理分流，降低延迟、节省解锁机流量）。
    - 本机原生不可用 → 立即推荐本次实测可用的最快解锁机（一键切换）或为任意解锁机 **一键开启 WARP 解锁**。
- **细化服务库 & 完整分流规则**
  - 支持域名后缀（`DOMAIN-SUFFIX`）、域名关键词（`DOMAIN-KEYWORD`）与 `IP-CIDR` 规则，内置 160+ 常见流媒体与 AI 服务，支持自定义服务、自定义分类、域名编辑与共享域名自动联动。
- **业务隔离与授权白名单**
  - 解锁机每 5 秒同步受管客户端 IPv4 白名单，仅放行授权 IP 访问解锁端口；客户端通过 cgroup 仅接管代理进程 DNS，绝不修改业务程序配置。
- **独立流量统计与多维度实测**
  - 独立统计每台被解锁机到各解锁机的解锁链路 RX/TX 流量；支持官方跨国流媒体脚本实测与全节点并发秒测。

---

## 📄 更多文档

- 版本更新记录：[CHANGELOG.md](CHANGELOG.md)
- 组件版本对应表：[VERSION_MATRIX_ZH.md](VERSION_MATRIX_ZH.md)
