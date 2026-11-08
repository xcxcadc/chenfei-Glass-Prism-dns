# v1.5.31：服务目录与解锁审计修复

- 修复上游名称带区域限定后缀（如 `Claude 2 /HK Only/`、`Google Gemini /HK Only/`、`Google AiStudio /HK Only/`）时，别名表、补充域名表、探测域名和审计标签全部匹配失败的问题。表现为 Gemini 只下发 7 个域名、Claude 与 AI Studio 审计显示“未覆盖”或失败。服务 ID 保持不变，已有路由不受影响。
- 域名校验拒绝单标签（如 `bbc`、`colab`、`twitter`）。已保存的偏好文件加载时静默剔除这类旧脏数据，不再因此导致服务无法启动。
- YouTube 补充内嵌播放器、播放器初始化、儿童版、教育版和短链域名，均已联网核对并在被解锁机验证可解析。
- 新增 Google Play 商店内置服务（`Google Play Store`），可被媒体审计的 `Google Play Store` 标签覆盖；域名刻意不与 YouTube 重叠，避免被强制绑定到同一节点。
- `prismdns.sh` 在停用系统自带 `dnsmasq.service` 之前先备份配置并写入 `was-active`/`was-enabled` 标记，卸载时可原样恢复。
- 不影响被解锁机上已有的 V2bX、Docker 等业务：Prism 仅监听 `127.0.0.1:53/5353`，使用自有 nft 表与 iptables 链。

## 升级

面板：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.31/enhanced_install.sh | sudo bash
```

已安装的节点无需重装，面板升级后路由会随同步自动生效。

## 部署排障经验（解锁机 Agent）

- 解锁机 `prism-agent` 的 `--ip` 参数必须是本机真实公网 IPv4。若写成不可达地址，面板会把它记入该节点的出口列表，被解锁机的 dnsmasq 会把服务解析到该死地址，表现为 YouTube、Play 商店全部超时。
- 解锁机的防火墙（如 ufw）需放行 TCP 443；Prism 的授权白名单只允许已授权的被解锁机访问，443 对其他来源仍被丢弃。
- 已知问题：个别解锁机上 `play.google.com` 经 Agent 的 443 透明转发会被丢弃，而同机其他 Play 域名正常，原因待上游 Agent 排查。

## 卸载补充

`uninstall.sh` 现在会一并清理 DNS 守卫创建的 nft 表 `prismdns_dns_guard` 与 iptables 链 `PRISMDNS_DNS_GUARD`、`PRISMDNS_DNS_GUARD6`、`PRISMDNS_GOOGLE_QUIC`，只处理 Prism 自有对象，不触碰 V2bX、Docker、nginx。
