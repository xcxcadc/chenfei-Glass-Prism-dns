# v1.6.1 更新日志

## 核心新功能：面板按服务可控非全局 WARP 分流

1. **任意服务按需走 WARP，绝不影响其他原生解锁**
   - 出口代理 `prism-egress-proxy` 新增零依赖 RFC 1928 SOCKS5 拨号器（默认对接本地 `127.0.0.1:40000`），支持 5 种精细出口模式：
     - `原生直连 (direct)`：默认模式，完全走解锁机原生 IPv4/IPv6，不受 WARP 任何影响。
     - `自动兜底 (auto)`：按 `原生 IPv4 → 原生 IPv6 → WARP IPv6 → WARP IPv4` 自动择优，仅当原生被区域封锁/送中时才自动切入 WARP。
     - `WARP 智能 (warp)`：仅对该服务的域名走本地 WARP，并自动检测 WARP IPv6 与 IPv4 中更干净的非送中出口（例如 AKI-JP 自动优先 WARP IPv6 获得 `JPN` 解锁，SG 自动获得 `SGP` 解锁）。
     - `WARP IPv6 (warp6)` / `WARP IPv4 (warp4)`：显式指定该服务走 WARP IPv6 或 WARP IPv4 出口。
   - 若解锁机尚未开启 WARP 或 WARP 临时不可用，自动优雅降级回原生直连，不会导致断流。

2. **面板端实时配置与全局同步**
   - 新增持久化策略存储 `warp-policies.json` 与认证接口 `GET/PUT /enhancer/api/warp-policies`。
   - **节点管理**：每个解锁机卡片实时显示 `☁ WARP 就绪 / 未启用` 状态与已启用 WARP 的服务数量，点击 **WARP 分流** 可在全部 162+ 内置与自定义服务中搜索并单独设置出口模式，支持「一键 Google/Gemini 走 WARP」与「一键重置直连」。
   - **IP 配置与服务路由**：在 IP 配置弹窗的每个已选服务旁新增行内 `WARP 出口` 下拉选择器，改完即保存，约 5 秒内自动热同步至对应解锁机（`prism-transport.timer` + `SIGHUP`），无需重启连接；服务路由表格同步显示 `☁ WARP` 标记。

3. **解锁机一键非全局 WARP 命令 (`dns warp`)**
   - `dns.sh` 新增 `dns warp enable | disable | status`：在解锁机上部署轻量用户态 `wireproxy` + `wgcf`（`prism-warp.service`，仅监听 `127.0.0.1:40000`），**零修改系统内核路由表、不接管全局 IPv4/IPv6 流量**，对机上任何其他代理/解锁服务零侵入。
   - 传输组件 `prism_transport.sh` 升级至 `2.6.1`，自动检测 `127.0.0.1:40000` 并向面板上报 `warp_ready` 状态。

4. **内置修复**
   - 包含 `quick-check` 探测强制 HTTP/1.1 ALPN 修复，消除 Google/Gemini 节点探测偶发 HTTP/2 协议错误。
