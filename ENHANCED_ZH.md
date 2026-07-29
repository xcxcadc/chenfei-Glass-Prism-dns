# Prism DNS 中文增强版

本 Fork 在原 Controller 前增加一个独立、可审计的增强层。原 Controller 和 Agent 仍使用上游发布的二进制，增强层负责简体中文界面、服务级路由、自定义域名与分类、动态规则集、UnlockTests 结果展示、账户安全和解锁链路统计。

- 仓库：[xcxcadc/chenfei-Glass-Prism-dns](https://github.com/xcxcadc/chenfei-Glass-Prism-dns)
- 最新版本：[GitHub Releases](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases/latest)

## 主要能力

- 默认简体中文，可随时切换英文和深浅主题。
- 动态读取 `stream.smartdns.list`，按分类和服务拆分为独立规则集。
- 每个受管目标 IP 的每项服务必须显式绑定到任意 Proxy Agent；只有用户保存配置才会改变出口，不跟随 Smart/Fallback 或检测结果自动跳转。
- 稳定通道默认安装并锁定上游 Agent `v1.2.1`，避免 `v1.3` 被动熔断在 WAF、拒绝连接或批量实测后把固定解锁 VIP 回退为公网 DNS；重新安装或卸载会自动解除锁定。
- 已纳管服务始终使用 Proxy IPv4；已选服务的 AAAA 会被抑制，防止客户端绕过解锁机走不解锁的 IPv6。
- 共用父域、子域或泛域名的服务自动联动到同一 Proxy，最后一次手动选择优先，避免 DNS 规则互相覆盖。
- 目标机实测只更新报告，WAF、超时或第三方检测波动都不会自动切换 Proxy；每项服务始终使用用户最后一次保存的 IPv4 节点。
- Proxy 每 5 秒同步面板已纳管目标 IPv4 白名单，仅放行这些地址访问 IPv4 DNS 53 与 SNI 80/443，并拒绝 IPv6 访问这些 Prism 端口；其他 IPv6 端口不受影响，同机 MTProxy 可继续使用。
- 前端新增、编辑和删除自定义服务，支持任意名称、分类和域名列表；普通域名使用 `example.com`，泛域名使用 `*.example.com`，保存、重新打开和迁移后均保留原格式。
- 服务库与 IP 服务选择器共用统一搜索，可按名称、中文别名、分类、服务 ID 或域名进行精确和模糊匹配，并忽略大小写及常见分隔符。
- 服务库可新建自定义分类，并把任意内置或自定义服务移动到任意分类；移动只覆盖显示分类，不改变稳定服务 ID、域名规则、IP 路由或客户端配置，可一键恢复原分类。
- 自定义服务自动生成兼容 Prism 的 `DOMAIN-SUFFIX` 规则集；泛域名在面板中保留 `*.`，下发时转换为基础后缀并去重，避免把无效的 `DOMAIN-SUFFIX,*.example.com` 发送给 Agent。
- 节点检测完成后列出解锁机自身的 UnlockTests；目标服务器再运行同源 UnlockTests，IP 配置页显示最终实测结论。
- 新增 IP 配置闭环：保存时自动创建 DNS Client、服务规则和逐服务解锁机覆盖。
- 已安装目标机修改服务后自动生效：每 10 秒配置哈希守卫在路由变化时原子更新专用 dnsmasq 规则并清理旧 DNS 状态，Agent 只在健康恢复失败时兜底重启。
- 受限网络目标机通过加密 TCP SNI 传输连接解锁机；客户端每 60 秒同步并执行真实 HTTPS 路径探测，SSH 自身仍按 15 秒保活、连续 3 次失联判定，发现业务链路卡死时只重建对应隧道。
- 解锁审计逐一检查全部路由域名的精确 A 映射、AAAA 抑制、TLS/SNI 握手及代表页面/服务方结论，并显示 `DNS x/x`、`TLS/SNI y/y` 与页面成功数；不再只抽查一个首页域名。
- 守护任务只操作 Prism 的 `prism-agent`、专用 dnsmasq 与传输服务，不会重启或覆盖同机 MTProxy、XrayR、V2bX。
- 选择服务时原地更新勾选状态并保持滚动位置；图标未返回前立即显示本地首字占位，图标请求带版本指纹，增强层启动时优先预热已配置服务，并把真实品牌图标持久化到 `/var/lib/prism-enhancer/icon-cache`。
- 服务编排使用表格控制台，可在同一页完成服务检索、状态与节点筛选、分页、批量选择、逐项节点配置和真实目标机测试；侧栏的节点管理、IP 配置、域名同步、日志审计、告警中心和设置中心均为可操作页面。
- 提供 `prismdns.sh` 客户端工具，支持 Agent 安装、DNS 测试、系统 DNS 接管、备份和恢复。
- 客户端使用 nftables 专用计数器，每分钟只上报本机 Prism DNS UDP/TCP 53 与所选 Proxy IP TCP 80/443 的 RX/TX。
- 点击右上角用户名可验证旧账号并修改 Controller 管理员用户名和密码，修改后旧会话全部失效。
- 自定义服务保存在 `/var/lib/prism-enhancer/custom-services.json`，自定义分类和服务分类覆盖保存在 `/var/lib/prism-enhancer/catalog-preferences.json`。

## IP 配置使用流程

1. 先在“节点管理”中添加并上线至少一台 Proxy Agent 解锁机。
2. 打开“IP 配置”，填写需要使用 DNS 解锁的目标服务器公网 IP；实际连接解锁机时固定选择 Proxy IPv4。
3. 选择需要解锁的服务，并为每项服务指定任意在线 Proxy Agent。
4. 首次保存后，系统会自动创建专属 DNS Client 节点、规则和服务覆盖，并生成专属配置命令。
5. 在目标服务器以 `root` 身份执行该命令。脚本会安装 DNS Agent、确认 `prism-agent` 实际占用 `127.0.0.1:53`、备份原 DNS，并在确认后接管系统 DNS；发现旧 `dnsmasq`/`sniproxy` 占用端口时先备份再停用，不修改 XrayR/V2bX。
6. 安装完成后，在面板增删服务或切换解锁机只需点击“保存配置”。页面直接关闭并自动下发，不再重复显示安装命令；目标机通常在 10 秒内安全重启 Agent 并应用新路由。

需要重新打开命令时，可在该 IP 行点击“客户端脚本”。通用交互入口如下：

```bash
wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/prismdns.sh | sudo bash
```

## 流量统计口径

`prismdns.sh` 会为每台目标服务器创建独立 nftables 统计链：本机 Prism DNS 的 UDP/TCP 53 请求与响应计入 DNS 用量，只有与当前所选 Proxy IPv4 或其专属加密传输端点之间的 TCP 80/443 才计入 SNI 解锁用量。普通公网 80/443、系统更新和其他非解锁机流量不会计入。systemd timer 启用后 15 秒内首次上报，之后每分钟上报一次；计数器在正常上报之间保持单调累计，面板按上次采样差值入账，不会因某一分钟流量更大而少算。自动 UnlockTests 审计结束后会开启新的计数纪元，避免把检测流量计入用户用量。

“清零流量”只重置面板累计值并保留目标机当前采样基线，后续只累加清零后的新流量，不会把旧计数重新加回；该操作不影响 nftables 规则、DNS 节点和服务配置。IP 配置及令牌保存在 `/var/lib/prism-enhancer/ip-configs.json`，文件权限为 `0600`；请勿公开页面生成的专属命令或配置令牌。

## 路由自动应用

`prismdns.sh 1.4.10` 会在每台目标机安装 `/usr/local/lib/prismdns/sync-routes.sh`、`prismdns-route-sync.timer` 和专用 `prismdns-local-dns.service`。增强层把服务到用户所选 Proxy IPv4 的映射持久化，动态生成 `/etc/prismdns/dnsmasq-routes.conf`；专用 dnsmasq 仅监听 `127.0.0.1:5353`，nftables 把本机 53 请求重定向到该端口，精确返回所选 IPv4 并抑制对应 AAAA。它不依赖 Agent 的 Smart/Fallback 或熔断状态，Agent 保留用于面板同步、授权和报告。稳定安装器固定使用上游 Agent `v1.2.1` 并校验 SHA-256、锁定可执行文件；配置变化优先原子替换规则并重启专用 dnsmasq，只有健康恢复失败才重启 Agent。受限网络链路会通过 `prism_transport.sh 2.2.3` 建立加密 TCP SNI 传输；客户端每 60 秒同步并用双站点 HTTPS 探针识别“SSH/TCP 仍在线但业务转发卡死”的假健康，SSH 本身继续按 15 秒保活、连续 3 次失联判定；修复时只重建对应隧道，不改服务路由。守卫每 10 秒检查配置哈希和本地 DNS 监听，每 300 秒抽测每项服务一个代表域名；配置变化、30 分钟健康缓存刷新和服务审计仍会全量验证所有域名。发往已选解锁机的 UDP/443 会被明确拒绝，促使支持 QUIC 的应用回落到 SNIproxy 可处理的 TCP/TLS。每分钟定时器只负责流量上报，耗时的四阶段检测由独立 `prismdns-service-audit.service` 执行。守卫只操作 Prism 服务，不触碰 MTProxy、XrayR、V2bX；面板或网络瞬时不可达时任务会正常退出并等待下一轮。

目标机只检测该 IP 已选择的服务。每项服务必须依次通过：全部路由域名精确解析到用户所选 Proxy IPv4、AAAA 为空、TLS/SNI 连续三次至少两次成功、代表页面或对应 UnlockTests 服务方项目可用。Gemini 会覆盖 26 个网页与移动端依赖，而不是只检查首页。服务方明确返回 `NO`、`Banned`、WAF 或稳定性不足时直接显示失败；没有专属 UnlockTests 项目的自定义服务则按自己的 DNS、TLS/SNI 和代表页面判定，不会因检测器没有输出而误报。任何实测结果都不会修改路由或自动跳转节点，用户需要切换线路时必须手动选择并保存。长时间审计会被守卫识别并去重，不会每 10 秒重复排队。节点页只把解锁机 Agent 自检保留为参考，目标机四阶段结果才是最终结论。

该逻辑不写死任何现有 IP、域名或节点 ID。以后从面板新增的解锁机和被解锁机，只要执行页面首次生成的客户端命令，就会默认获得相同的 IPv4 优先、AAAA 抑制、共享域名联动、只读实测、5 秒授权、缓存清理、流量统计和健康上报能力。无 `flock` 的精简系统会自动使用原子目录锁降级。

`1.4.10` 回归使用运行时最新有效服务目录，并对每台目标机逐项验证 DNS、AAAA、TLS/SNI、代表页面和服务方结论。面板不会把解锁机自身自检结果伪装成目标机可用，不会因第三方检测波动自动切换节点，也不会把共享域名拆到互相覆盖的出口。Crackle、Salto、GYAO 已停止运营，不再进入有效服务目录。

## 解锁检测与账户安全

“节点管理”中的“运行解锁检测”调用解锁机 Agent 的检测任务，完成后直接汇总可用与不可用项目，但这只代表解锁机自身。目标 IP 安装 `prismdns.sh` 后，会在路由变化及定期周期内运行与 `dns_unlock.sh` 同源的 UnlockTests，再逐一核对全部路由域名、AAAA 抑制、TLS/SNI 握手和代表页面，把结果上报到“IP 配置”。服务配置页优先展示目标机四阶段结论；审计只写报告，不修改路由。

点击页面右上角当前用户名可打开“账户安全”。系统先通过 Controller 验证旧用户名和旧密码，再原子更新 SQLite 用户记录并清理旧会话；更新成功后必须使用新账户重新登录。

## 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/enhanced_install.sh | sudo bash
```

可通过环境变量调整端口：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/enhanced_install.sh \
  | sudo PRISM_PORT=8080 PRISM_CORE_PORT=18080 bash
```

## 本地开发

```bash
cd enhancer
go test ./...
go run . --listen 127.0.0.1:8080 --upstream http://127.0.0.1:18080 --data-dir ./data
```

提交前可执行：

```bash
cd enhancer && go test -race ./... && go vet ./...
cd .. && bash -n prismdns.sh enhanced_install.sh agent_install.sh
```

## 数据来源与边界

细化域名目录运行时读取自 [1-stream/1stream-public-utils](https://github.com/1-stream/1stream-public-utils) 的 `stream.smartdns.list`。该仓库目前未声明许可证，本 Fork 不提交其名单副本，只在部署实例中按用户请求实时读取和转换。

上游 `mslxi/Liquid-Glass-Prism-dns` 同样未提供 Controller/Agent 源码及许可证。增强层不复制、不修改上游二进制，只通过已公开的 HTTP API 与其协作。上游二进制的使用和分发仍受原作者权利约束。

## 节点接入与解锁配置

“节点管理”直接新增的 DNS 客户端仅完成上游 Agent 入网。它在已上线但尚未选择解锁服务时，会显示为“未纳管”，而不会误报为等待健康上报。

点击该节点卡片的“接入 IP 服务”，选择解锁机和要启用的服务，保存后即会复用原 DNS 节点的 Agent 密钥生成专属 `prismdns.sh` 安装命令。完成安装后，该节点会自动上报 DNS 、路由、解锁链路流量和目标测试结果。

对这类被接管的既有节点，删除 IP 配置只会清理对应解锁路由和本面板配置，不会删除原手动创建的 DNS 节点。节点名称和分组支持简体中文、英文、数字、空格、点、下划线和连字符。

由于上游 Controller 只接受 ASCII 节点名称和分组，增强层会把中文显示名称持久化到 `/var/lib/prism-enhancer/node-labels.json`，并向 Controller 提交兼容名称。新增、编辑、删除及面板重启后都会自动保持映射，不改变节点 ID、Agent 密钥或路由关系。
