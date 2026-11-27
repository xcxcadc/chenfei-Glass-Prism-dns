# v1.6.0：流光玻璃界面焕新、自适应双栈防送中与全节点秒级快测

- **全链路 Gemini / Google 防送中双栈自适应分流（`prism-egress-proxy` + `prism_transport.sh` v2.6.0）**：
  - 解锁机（`proxy`）在 `direct` 与 `encrypted` 两种传输模式下均统一由 `prism-egress-proxy` 承载 `80/443` 流量，彻底解决直连模式下外部流量绕过双栈分流器、命中 Google“IPv4 送中（CHN）”的问题。
  - 升级现代版 Gemini（`45631641` 新版地区标记）、YouTube CN、Google Play Store China 与 Claude 区域重定向特征识别，代表性域名轻量探测（1~2 个核心域名）取代旧版全域名风暴，并在配置热重载时平滑保留最优 IPv6 选路状态。
- **被解锁机零误杀 `SO_ORIGINAL_DST` 原目标直通（`--mode client`）**：
  - 透明 SNI 代理新增 Linux `SO_ORIGINAL_DST` / `IP6T_SO_ORIGINAL_DST` 原始目标提取：未纳管域名、非 TLS 协议与 ECH 流量直接按原目标 IP 原样透传，消除二次 DNS 解析与连接中断；同时支持域名关键词、IP-CIDR、Docker 代理容器 cgroup 以及 `encrypted` 加密隧道串联。
- **流光玻璃 UI 焕新与丝滑交互体验（`prism-enhancer` v1.6.0）**：
  - 全新浅色/深色毛玻璃（Liquid Glass）视觉体系、卡片悬浮微动效与解锁机实时解锁能力标签条（Gemini / ChatGPT / Claude / Netflix / Disney+ / YouTube + IPv6 防送中徽章）。
  - 新增 **⚡ 全节点秒测（`/enhancer/api/quick-check`）**：无需等待数分钟全量脚本，1~3 秒并发经所有解锁机 443 出口实测单服务真实可用性、地区与延迟，支持一键选用并下发。
  - IP 配置弹窗新增 **AI 全家桶 / 流媒体全家桶** 智能优选勾选、已选服务批量改节点，以及下拉框节点可用性/送中状态直观标注。
  - 自动检测指向已删除旧解锁机的残留孤儿路由，提供顶栏 **一键迁移修复 / 清除失效项**。
- **面板自托管极速分发与媒体检测精准修正（`prism_media.sh` / `/enhancer/dist/`）**：
  - 面板内置 `/enhancer/dist/` 自托管脚本与二进制分发及 **一键升级全部节点** 批量接口，国内/受限网络免 GitHub 亦可秒装秒升。
  - 修正官方检测脚本对新版 Google Gemini（`45631641\":true`）与 Claude（`unavailable-in-region`）的误判，并在解锁机自检时真实穿过本地自适应双栈出口，容忍单子项非零退出码保留有效检测结果。
