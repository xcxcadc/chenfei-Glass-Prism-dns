# Prism DNS 中文增强版

本 Fork 在原 Controller 前增加一个独立、可审计的增强层。原 Controller 和 Agent 仍使用上游发布的二进制，增强层负责简体中文界面、服务级路由、自定义域名、动态规则集、UnlockTests 结果展示、账户安全和解锁链路统计。

- 仓库：[xcxcadc/chenfei-Glass-Prism-dns](https://github.com/xcxcadc/chenfei-Glass-Prism-dns)
- 最新版本：[GitHub Releases](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases/latest)

## 主要能力

- 默认简体中文，可随时切换英文和深浅主题。
- 动态读取 `stream.smartdns.list`，按分类和服务拆分为独立规则集。
- 每个 DNS 节点可把每项服务手动绑定到任意 Proxy Agent，并可恢复 Smart/Fallback 自动选择。
- 前端新增、编辑和删除自定义服务，支持任意名称、分类和域名列表。
- 自定义服务自动生成兼容 Prism 的 `DOMAIN-SUFFIX` 规则集。
- 节点检测完成后列出解锁机自身的 UnlockTests；目标服务器再运行同源 UnlockTests，IP 配置页显示最终实测结论。
- 新增 IP 配置闭环：保存时自动创建 DNS Client、服务规则和逐服务解锁机覆盖。
- 提供 `prismdns.sh` 客户端工具，支持 Agent 安装、DNS 测试、系统 DNS 接管、备份和恢复。
- 客户端使用 nftables 专用计数器，每分钟只上报与所选 Proxy IP 之间的 RX/TX。
- 点击右上角用户名可验证旧账号并修改 Controller 管理员用户名和密码，修改后旧会话全部失效。
- 自定义数据保存在 `/var/lib/prism-enhancer/custom-services.json`。

## IP 配置使用流程

1. 先在“节点管理”中添加并上线至少一台 Proxy Agent 解锁机。
2. 打开“IP 配置”，填写需要使用 DNS 解锁的目标服务器公网 IPv4 或 IPv6。
3. 选择需要解锁的服务，并为每项服务指定任意在线 Proxy Agent。
4. 保存后，系统会自动创建专属 DNS Client 节点、规则和服务覆盖，并生成专属配置命令。
5. 在目标服务器以 `root` 身份执行该命令。脚本会安装 DNS Agent、确认 `prism-agent` 实际占用 `127.0.0.1:53`、备份原 DNS，并在确认后接管系统 DNS；发现旧 `dnsmasq`/`sniproxy` 占用端口时先备份再停用，不修改 XrayR/V2bX。

需要重新打开命令时，可在该 IP 行点击“客户端脚本”。通用交互入口如下：

```bash
wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/prismdns.sh | sudo bash
```

## 流量统计口径

`prismdns.sh` 会为每台目标服务器创建独立 nftables 统计链：本机 Prism Agent 的 UDP/TCP 53 请求与响应计入 DNS 用量，只有与当前所选 Proxy IPv4/IPv6 之间的 TCP 80/443 才计入 SNI 解锁用量。普通公网 80/443、系统更新和其他非解锁机流量不会计入。systemd timer 启用后 15 秒内首次上报，之后每分钟上报一次；首次上报或清零后的首次上报只建立新基线。自动 UnlockTests 审计结束后会重置本地计数器，避免把检测流量计入用户用量。

“清零流量”只重置面板累计值和上报基线，不会影响 nftables 规则、DNS 节点和服务配置。IP 配置及令牌保存在 `/var/lib/prism-enhancer/ip-configs.json`，文件权限为 `0600`；请勿公开页面生成的专属命令或配置令牌。

## 解锁检测与账户安全

“节点管理”中的“运行解锁检测”调用解锁机 Agent 的检测任务，完成后直接汇总可用与不可用项目，但这只代表解锁机自身。目标 IP 安装 `prismdns.sh` 后，会在路由变化及定期周期内运行与 `dns_unlock.sh` 相同的 `/usr/bin/ut -m 4 -f 20 -b=false -s=false`，把每个已选服务的原始结果上报到“IP 配置”。服务配置页优先展示目标机结论，并只对一个已知可访问的代表域名做 TLS 诊断；列表中的其他根域可能仅是路由后缀，不提供 HTTPS，不能据此判定平台不可用。

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
