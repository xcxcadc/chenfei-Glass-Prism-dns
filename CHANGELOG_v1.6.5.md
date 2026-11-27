# v1.6.5 稳定性修复日志

## 修复 WARP 隧道假死导致回退原生超时的问题（彻底杜绝连接挂死）

1. **出口代理独立上下文回退 + WARP 快速熔断 (`egressproxy`)**
   - 修复 `dialViaWARP` 与原生回退 `dialHost` 共享同一个 `context.WithTimeout` 的严重缺陷：此前当 `wireproxy` 出现 WireGuard 握手卡死时，WARP 拨号耗尽全部 20 秒超时，导致后续回退原生直连因上下文已过期而同步失败（`lookup ...: i/o timeout`）。
   - 现在 `dialViaWARP` 使用独立的 **4 秒短超时**，失败后立即使用全新上下文回退原生出口（毫秒级完成），绝不阻塞业务访问。
   - 新增 **WARP 隧道快速熔断器**：连续 2 次 WARP 拨号失败后自动旁路 WARP 25 秒，所有连接直走原生出口，待隧道恢复后自动切回。

2. **WARP 真实数据面保活与自动自愈 (`prism_transport.sh`)**
   - 修复 `proxy_warp_ready()` 仅检查 `127.0.0.1:40000` 端口监听、不检测 WireGuard 真实通断的缺陷。
   - 新增每 30 秒一次的真实数据面探测（经 `127.0.0.1:40000` 访问 `cloudflare.com/cdn-cgi/trace` 验证 `warp=on`）：一旦检测到 WireGuard 握手卡死，自动重启 `prism-warp.service` 自愈，若默认 `2408` 端口异常则自动轮换备用 Cloudflare WARP 端口（`500`/`1701`/`4500`/`854`）并热通知 `prism-egress-proxy`。
   - 修复 `reconcile_local_warp()` 复用已有配置时路径检查（`/etc/prismdns/warp/wireproxy.conf`）。
