# Prism DNS 中文增强版

本 Fork 在原 Controller 前增加一个独立、可审计的增强层。原 Controller 和 Agent 仍使用上游发布的二进制，增强层负责简体中文界面、服务级路由、自定义域名与分类、动态规则集、目标机媒体脚本结果展示、账户安全和解锁链路统计。

- 仓库：[xcxcadc/chenfei-Glass-Prism-dns](https://github.com/xcxcadc/chenfei-Glass-Prism-dns)
- 最新版本：[GitHub Releases](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases)

## 主要能力

- 默认简体中文，可随时切换英文和深浅主题。
- 动态读取 `stream.smartdns.list`，按分类和服务拆分为独立规则集。
- 每个受管目标 IP 的每项服务必须显式绑定到任意 Proxy Agent；只有用户保存配置才会改变出口，不跟随 Smart/Fallback 或检测结果自动跳转。
- 稳定通道默认安装并锁定上游 Agent `v1.2.1`，避免 `v1.3` 被动熔断在 WAF、拒绝连接或批量实测后把固定解锁 VIP 回退为公网 DNS；重新安装或卸载会自动解除锁定。
- 已纳管服务始终使用 Proxy IPv4；egress 不会因 Gemini/AI 等服务自动切到 IPv6，已选服务的 AAAA 会被抑制，防止客户端绕过解锁机走不解锁的 IPv6。
- 共用父域、子域或泛域名的服务自动联动到同一 Proxy，最后一次手动选择优先，避免 DNS 规则互相覆盖。
- 目标机实测只更新报告，WAF、超时或第三方检测波动都不会自动切换 Proxy；每项服务始终使用用户最后一次保存的 IPv4 节点。
- Proxy 每 5 秒同步面板已纳管目标 IPv4 白名单，仅放行这些地址访问 IPv4 DNS 53 与 SNI 80/443，只拒绝 IPv6 DNS 53，不拦截 IPv6 80/443；同机 MTProxy 等 IPv6 业务可继续使用。
- 目标机如果 XrayR、V2bX、sing-box、Hysteria、TUIC 等代理进程自行把客户端 DNS 发往 `8.8.8.8:53` 或 `1.1.1.1:53`，路由守卫会按 systemd cgroup 将这些 DNS 请求转交给本机 Prism Agent；只增加 Prism 自有 nftables 表，不改动代理程序配置或 Docker。
- 不会读取、修改或重启 XrayR、V2bX、MTProxy、Docker 或其他业务配置；代理进程只通过 Prism 自有 DNS 守卫接入本机解析器。
- 前端新增、编辑和删除自定义服务，支持任意名称、分类和域名列表；普通域名使用 `example.com`，泛域名使用 `*.example.com`，保存、重新打开和迁移后均保留原格式。
- 所有内置和自定义服务均可在服务表格或配置弹窗中编辑域名、增删域名并恢复域名库默认值；“彻底删除服务”会同步删除规则、目标 IP 路由、审计结果和无用解锁机引用，内置服务会写入永久隐藏列表，不会在下一次域名库同步时复现。
- 服务规则编辑器同时支持域名后缀、域名关键词和 IPv4/IPv6 `IP-CIDR`；规则列表、Agent 同步和传输配置会保留原始规则类型。Grok 默认包含 xAI/Twitter 域名、`twitter` 关键词及六段上游 CIDR。
- 服务库与 IP 服务选择器共用统一搜索，可按名称、中文别名、分类、服务 ID 或域名进行精确和模糊匹配，并忽略大小写及常见分隔符。
- 服务库可新建自定义分类，并把任意内置或自定义服务移动到任意分类；移动只覆盖显示分类，不改变稳定服务 ID、域名规则、IP 路由或客户端配置，可一键恢复原分类。
- 自定义服务自动生成兼容 Prism 的 `DOMAIN-SUFFIX` 规则集；泛域名在面板中保留 `*.`，下发时转换为基础后缀并去重，避免把无效的 `DOMAIN-SUFFIX,*.example.com` 发送给 Agent。
- 每台 Proxy/Client 默认安装官方媒体脚本 worker，完整执行 IPv4 菜单 `0` 的跨国平台检测，按官方 label 原样展示结果；Agent 自检只作参考，DNS/TLS/SNI 诊断独立保留。
- 新增 IP 配置闭环：保存时自动创建 DNS Client、服务规则和逐服务解锁机覆盖。
- 已安装目标机修改服务后自动生效：每 10 秒配置哈希守卫在路由变化时原子更新专用 dnsmasq 规则并清理旧 DNS 状态，Agent 只在健康恢复失败时兜底重启。
- 受限网络目标机通过加密 TCP SNI 传输连接解锁机；客户端每 60 秒同步并执行真实 HTTPS 路径探测，SSH 自身仍按 15 秒保活、连续 3 次失联判定，发现业务链路卡死时只重建对应隧道。
- 媒体任务每秒逐批刷新，支持六小时自动刷新、手动重测、配置代更新和关窗后重读进度；路由诊断保留精确 A 映射、AAAA 抑制、TLS/SNI 和代表页面信息，不覆盖媒体结论。
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
5. 在目标服务器以 `root` 身份执行该命令。脚本会备份原 DNS，安装兼容 Agent、监听 `127.0.0.1:5353` 的专用 dnsmasq、runtime、transport 和媒体 worker，再接管系统 DNS；发现冲突的旧 `dnsmasq`/`sniproxy` 时先备份再停用，不修改 XrayR/V2bX。
6. 安装完成后，在面板增删服务或切换解锁机只需点击“保存配置”。目标机每 10 秒检查配置并应用新路由，优先更新专用 DNS，Agent 仅在健康恢复失败时重启；新配置代会触发独立媒体任务。

需要重新打开命令时，可在该 IP 行点击“客户端脚本”。通用交互入口如下：

```bash
wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/prismdns.sh | sudo bash
```

## IP 配置备份与授权令牌

“IP 配置”页面提供“导出配置”和“导入配置”。导出的 v1 JSON 只包含目标 IP、备注、智能模式及服务到解锁机的标识（节点 ID、名称和公网 IP），不会包含安装令牌、节点密钥、账号、流量累计值、健康状态或服务审计结果。导入前会校验版本、IP、备注长度和解锁机；节点 ID 变化时会按公网 IPv4 或节点名称重新匹配。导入新目标 IP 后，面板会创建新的 DNS 节点并在结果中生成新的“客户端脚本”，必须在目标机执行一次；同 IP 导入则更新已有配置，不会重置其累计流量或令牌。

每个 IP 行的“轮换令牌”会立即撤销旧令牌、清除旧客户端健康/审计状态并生成新的安装命令。旧目标机不会继续获得配置，需在该目标机重新执行新命令。令牌轮换是显式管理操作，不设置自动过期，避免长期运行的生产节点在无人值守时被突然断开；删除 IP 配置同时撤销其令牌。

## 流量统计口径

`prismdns.sh` 会为每台目标服务器创建独立 nftables 统计链：本机 Prism DNS 的 UDP/TCP 53 请求与响应计入 DNS 用量，只有目标机与当前所选 Proxy IPv4 之间的 TCP 80/443 才计入 SNI 解锁用量。普通公网 80/443、SSH 控制流量、系统更新和其他非解锁机流量不会计入。systemd timer 启用后 15 秒内首次上报，之后每分钟上报一次；DNS RX/TX 与每个 SNI 解锁方向的 RX/TX 分开持久化到 `/var/lib/prismdns/traffic-cumulative.json`，单个 nftables 表重建不会丢失另一组计数器的历史增量，面板按单调采样差值入账。统计状态按 Agent enrollment token 隔离，重新绑定目标 IP 不会串用旧数据。流量上报与媒体 worker 各自调度，不等待媒体脚本结束。

“清零流量”只重置面板累计值并保留目标机当前采样基线，后续只累加清零后的新流量，不会把旧计数重新加回；该操作不影响 nftables 规则、DNS 节点和服务配置。IP 配置及令牌保存在 `/var/lib/prism-enhancer/ip-configs.json`，文件权限为 `0600`；请勿公开页面生成的专属命令或配置令牌。

## 路由自动应用

本次面板与发布脚本版本为 `1.5.29`，固定脚本标签为 `v1.5.29`，SNI transport 为 `2.5.1`，兼容 Agent 默认使用上游 `v1.2.1`；三者独立编号，不能混写为 Agent `v1.3`。安装、升级和卸载命令均使用本版标签，详见 [版本矩阵](VERSION_MATRIX_ZH.md)。

`prismdns.sh 1.5.29` 在每台目标机安装 `/usr/local/lib/prismdns/sync-routes.sh`、`prismdns-route-sync.timer`、专用 dnsmasq 和 DNS 守卫。增强层持久化服务到用户所选 Proxy IPv4 的映射，生成 `127.0.0.1:5353` 与可选 `[::1]:5353` 的路由；守卫每 10 秒检查配置和 DNS 监听，并通过 Prism 自有 nftables cgroup 规则接管受管代理进程的外部 DNS 请求。已选服务只返回指定 IPv4 并抑制 AAAA，未选服务保持直连。配置变化和 30 分钟路由健康缓存刷新仍校验全部域名，每 300 秒抽测各服务代表域名；这些属于路由诊断，不参与媒体结果判定。

`prism_transport.sh 2.5.1` 默认以授权源 IP 直连所选 Proxy 的 TCP 80/443，受限网络可显式选择 `--mode encrypted`。Proxy 每 5 秒同步授权名单，AUTH nftables 规则先校验，再以单次事务原子替换，不先单独删除旧授权表；仅限制 Prism 的 IPv4 DNS/SNI 和 IPv6 DNS 53，不拦截同机 IPv6 80/443。已选出口的 UDP/443 被拒绝，促使应用回落到 TCP/TLS；受管代理发往 Google 网段的 UDP/443 快速拒绝只促使 Google 直连回落，不改变其 DNS 或出口。

路由健康心跳、transport 同步和每分钟流量上报独立运行。安装只强制验证 DNS 是否映射到所选 Proxy，不等待完整媒体脚本；重装遇到本地 DNS 停止但 `/etc/resolv.conf` 仍指向回环时，会临时恢复引导 DNS 后继续。运行时最新服务目录排除已停止运营的 Crackle、Salto、GYAO。任何诊断或媒体结果都不会自动修改用户保存的路由。

## 官方媒体脚本实时任务

每台 Proxy/Client 随 transport 默认安装 `/usr/local/bin/prism-media-check`，由 `prism-media-check.timer` / `prism-media-check.service` 在空闲时每 5 秒 claim 待处理任务，独立于 transport 和 traffic。以后新增的机器执行本版安装命令也默认继承，无需单独部署 worker，不写死现有 IP、域名或节点 ID。安装 worker 会停用旧 `prismdns-service-audit.timer` 并停止旧审计服务，旧 full / 四阶段检测不再作为最终媒体结论；独立 DNS/路由诊断继续保留。

worker 下载官方 `check.unlock.media` 脚本后检查 Shell shebang、排除 HTML 并执行 `bash -n` 校验，然后完整运行以下等价命令（`check.unlock.media` 为下载后的本地脚本文件）：

```bash
bash check.unlock.media -M4 -R0 -Een
```

这相当于交互菜单 `0`，只检测跨国平台、仅 IPv4，使用英文官方 label；不运行其他地区菜单或 IPv6 项目。结果按官方 label 解析，`Yes` / `No`、地区、WAF 等原始值保持不变。完整表展示脚本实际输出，服务配置按对应 label 展示；没有对应项目的服务标为“未覆盖”。DNS A 映射、AAAA 抑制、TLS/SNI 和代表页面只作独立路由诊断，无论成功或失败都不覆盖媒体结果。Agent 自检也只作参考。

worker 每秒检查新输出，有输出即逐批上报，界面每 1 秒刷新已收到的结果；不等待整轮脚本完成。脚本执行没有 30 秒硬截止，也不再沿用旧的短时整轮等待上限；15 分钟仅作异常安全兜底，超限、下载/校验失败、脚本异常退出或没有有效 IPv4 平台结果均明确标记 `failed` 并给出原因。90 秒无心跳会显示 worker 失联并把任务标为失败。任务 `completed` 只表示脚本正常完成，不表示每个平台均为 `Yes`。

面板将任务 ID、状态、逐批结果、进度和历史持久化到 `dataDir/media-checks.json`，默认 `/var/lib/prism-enhancer/media-checks.json`。关闭窗口不会取消后台任务，重新打开可继续读取；这不表示已失败的进程会自动从中断位置续跑。自动刷新周期为 6 小时，也支持手动重测和配置代更新后重测。路由、Proxy IPv4 或重测请求变化时更新 `generation`；旧代迟到上报会被拒收，防止旧线路覆盖当前结果。

这里展示的是**官方脚本检测结果，不是实际付费播放、账号登录或所有应用功能通过的保证**，也不承诺已测试登录态账户权益。

升级会清理超过 24 小时的旧 `/run/prismdns-media-check.*` 遗留文件，避免旧检测填满临时内存盘而阻止 systemd 加载配置。路由守卫会核验本地 DNS 的实际监听状态；即使规则内容未变化，缺失 IPv4/IPv6 回环监听也会重启专用解析器恢复。Agent 的旧 Controller 地址会在成功取得当前 bootstrap 后同步，以上修复对后续新增和升级机器同样生效。

## 已有节点升级

**仅替换 enhancer 二进制不会自动刷新远端节点。** 面板升级与节点脚本升级应分别完成：

1. 面板使用本版 `v1.5.29/enhanced_install.sh`；保留整个 Enhancer `dataDir`，包括 `media-checks.json` 的进度和历史。
2. 每台既有 Proxy 重新执行“节点管理”生成的本版 `agent_install.sh` 命令，更新兼容 Agent、transport `2.5.1` 和媒体 worker。
3. 每台既有 Client 重新执行“IP 配置”生成的本版专属命令；已有有效本机配置时，也可用本版脚本刷新 runtime、transport 与 worker：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/prismdns.sh | sudo bash -s -- --refresh-runtime
```

`prism_transport.sh --sync` 只同步已安装的传输配置，不会下载新版 runtime/transport 或补装 worker。刷新后在每台节点检查 `systemctl is-enabled prism-media-check.timer`、`systemctl status prism-media-check.timer`，并在面板手动重测，确认逐批结果和任务心跳可读；oneshot service 空闲时不常驻是正常现象。

面板地址迁移时，更新后的 transport 优先使用显式 `--master`；旧地址不可达时可从运行中 Agent 的 `--master` 恢复新地址，经响应校验后持久化到 transport 配置，worker 随后读取该配置。完整步骤见 [迁移教程](MIGRATION_ZH.md)。worker 跟随 transport / Agent 卸载，timer、service 和可执行文件一并移除，见 [卸载说明](UNINSTALL_ZH.md)。

## 发布、图标与全新安装

仓库发布内容已脱敏：只包含程序代码、安装脚本和前端资源，不包含运行时数据库、节点密钥、账号密码、IP 配置、流量记录或任何当前服务器数据。新服务器执行一键安装时会创建空白数据，并生成新的 JWT 密钥和初始密码。

面板浏览器标签、首页图标和“账户安全”窗口使用透明 PNG 图标。默认资源位于 `enhancer/web/favicon.png`；登录后进入“设置中心 → 站点设置”，即可上传带透明背景的 PNG（最大 2 MB），上传后立即替换并持久化到 `/var/lib/prism-enhancer/branding-icon.png`，也可以点击“恢复默认图标”。如需从源码替换默认图标，直接覆盖该文件后重新构建发布即可。运行中的旧浏览器缓存可用 `Ctrl+F5` 刷新。

如需在曾安装过 Prism 的主机上明确清空 Prism 数据并重新初始化，执行以下命令：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/enhanced_install.sh \\
  | sudo env PRISM_FRESH_INSTALL=1 PRISM_CONFIRM_FRESH=YES bash
```

该模式只清理 `/opt/prism/data.db`、`/opt/prism/.env`、初始密码文件和 `/var/lib/prism-enhancer/`，不会删除或修改 Docker、XrayR、V2bX、MTProxy 等其他业务。需要保留旧面板时，必须按 `MIGRATION_ZH.md` 显式备份和恢复；安装脚本不会自动带入旧数据。

## 账户安全

点击页面右上角当前用户名可打开“账户安全”。系统先通过 Controller 验证旧用户名和旧密码，再原子更新 SQLite 用户记录并清理旧会话；更新成功后必须使用新账户重新登录。

## 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/enhanced_install.sh | sudo bash
```

可通过环境变量调整端口：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/enhanced_install.sh \
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
bash tests/prismdns-audit-test.sh
```

## 数据来源与边界

细化域名目录运行时读取自 [1-stream/1stream-public-utils](https://github.com/1-stream/1stream-public-utils) 的 `stream.smartdns.list`。该仓库目前未声明许可证，本 Fork 不提交其名单副本，只在部署实例中按用户请求实时读取和转换。

上游 `mslxi/Liquid-Glass-Prism-dns` 同样未提供 Controller/Agent 源码及许可证。增强层不复制、不修改上游二进制，只通过已公开的 HTTP API 与其协作。上游二进制的使用和分发仍受原作者权利约束。

## 节点接入与解锁配置

“节点管理”直接新增的 DNS 客户端仅完成上游 Agent 入网。它在已上线但尚未选择解锁服务时，会显示为“未纳管”，而不会误报为等待健康上报。

点击该节点卡片的“接入 IP 服务”，选择解锁机和要启用的服务，保存后即会复用原 DNS 节点的 Agent 密钥生成专属 `prismdns.sh` 安装命令。完成安装后，该节点会自动上报 DNS 、路由、解锁链路流量和目标测试结果。

对这类被接管的既有节点，删除 IP 配置只会清理对应解锁路由和本面板配置，不会删除原手动创建的 DNS 节点。节点名称和分组支持简体中文、英文、数字、空格、点、下划线和连字符。

由于上游 Controller 只接受 ASCII 节点名称和分组，增强层会把中文显示名称持久化到 `/var/lib/prism-enhancer/node-labels.json`，并向 Controller 提交兼容名称。新增、编辑、删除及面板重启后都会自动保持映射，不改变节点 ID、Agent 密钥或路由关系。
