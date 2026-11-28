# 更新日志 (CHANGELOG)

## v2.0.0 — Apple Liquid Glass 2.0 前后端全面翻新
- **全新 Apple Liquid Glass 2.0 视觉系统**：半透明毛玻璃侧边栏、顶部悬浮指令栏与柔光/深空双主题渐变背景，支持全局 `Ctrl+K` 快捷命令面板（秒搜任意服务、节点或一键操作）。
- **后端 v2 实时能力升级**：新增 `GET /enhancer/api/v2/dashboard` 聚合仪表盘接口与 `GET /enhancer/api/v2/events` SSE 实时推送流。
- **节点管理重构为高密度横列表格（默认）+ 卡片双视图**：大幅压缩纵向滚动距离，支持按角色（全部/解锁机/被解锁机/待升级）与搜索筛选；自动检测旧版本节点，提供顶部高亮升级提醒条、行内版本差值标签与一键热更新。
- **域名库与服务管理全面可编辑**：告别旧版只读分类计数，点击任意分类直接在页内浏览、搜索、编辑服务域名/关键词/CIDR、调整分类或删除服务。
- **日志与告警闭环增强**：审计页新增按目标 IP、结果状态与关键词筛选；告警中心自动纳入旧版本待升级提醒。

## v1.6.6
- **新增主流社交与应用分流服务（分类「社交与主流应用」）**：内置 **X (Twitter)**、**Instagram (INS)**、**Facebook**、**Threads**、**WhatsApp**、**Telegram**、**Discord**、**Reddit**、**Pinterest**、**Line** 共 10 项主流服务，覆盖主站、API、认证与核心图片/视频 CDN 域名（Telegram / X 同步包含官方 IP-CIDR）。
- **新增社交全家桶一键勾选预设**：IP 配置弹窗支持一键勾选主流社交应用，全节点并发秒测与被解锁机原生直连探测均已适配这 10 项服务。

## v1.6.5
- **修复 WARP 隧道假死导致的原生回退超时**：`egressproxy` 的 WARP 拨号改为独立 4 秒短超时，失败后使用全新上下文立即回退原生出口，不再因共享上下文耗尽导致连接超时。
- **新增 WARP 快速熔断器**：连续 2 次 WARP 拨号失败后自动旁路 WARP 25 秒直走原生，隧道恢复后自动复位。
- **新增 WARP 真实数据面探活与自动自愈**：`prism_transport.sh` 每 30 秒经本地 `127.0.0.1:40000` 探测真实连通性，检测到 WireGuard 握手异常时自动重启或轮换 Cloudflare 备用端口自愈。

## v1.6.4
- **被解锁机默认原生直连检测**：新增 `POST/GET /enhancer/api/client-native-check` 及客户端上报接口，客户端通过 DoH 解析真实公网 IP 并使用 `curl --resolve` 绕过本机代理 DNS 实测原生直连能力。
- **智能推荐闭环**：选服务或全节点秒测时并发检测本机原生直连与全解锁机；原生可用时推荐一键切换直连，原生不可用时推荐可用解锁机或一键开启 WARP。

## v1.6.3
- **极简统一对接短链**：面板新增 `GET /i/<credential>`，一条短命令 `curl -fsSL https://面板域名/i/<密钥或Token> | bash` 自动识别解锁机 / 受管被解锁机 / 手动 DNS 节点。
- **GitHub 极简安装入口**：新增根目录 `i.sh`，无需进程替换，兼容各类 Shell 与 `sudo`。
- **多系统 53 端口与依赖兼容**：`agent_install.sh` 自动检测并释放 Ubuntu / Debian / Fedora 的 `systemd-resolved` 53 端口以及 `dnsmasq` / `unbound` / `bind9` 等冲突服务，支持 `apt`/`dnf`/`yum`/`apk`/`pacman`/`zypper`。

## v1.6.2
- **全自动远程 WARP 启停**：面板向解锁机下发 `warp_desired` 期望状态，勾选任意服务走 WARP 时机器端自动后台部署/启动非全局 WARP，全部恢复原生后同步自动关闭。
- **不可用服务闭环修复**：IP 配置、全节点秒测与服务路由表全面接入不可用告警、推荐切换可用解锁机与一键 WARP 解锁按钮。

## v1.6.1
- **按服务非全局 WARP 分流**：出口代理新增零依赖 RFC 1928 SOCKS5 拨号器（对接本地 `127.0.0.1:40000`），支持为任意内置/自定义服务独立设置 `direct` / `warp` / `warp6` / `warp4` / `auto` 出口模式，不改写宿主机默认路由。
- **新增 `dns warp` 命令**：支持 `dns warp enable | disable | status` 一键部署轻量用户态 `wireproxy` + `wgcf`。

## v1.6.0
- **Liquid Glass 界面焕新与效率升级**：新增全节点并发秒测（`/enhancer/api/quick-check`）、节点能力标签、批量一键升级、AI/流媒体一键配置预设、失效残留路由一键迁移修复。
- **双栈防送中出口引擎 & 零误杀客户端分流**：增强 Gemini / Claude / ChatGPT / YouTube / Play Store 区域探测，客户端未命中服务直通原始目标 IP，修复规则集通配符与泛域名匹配。
