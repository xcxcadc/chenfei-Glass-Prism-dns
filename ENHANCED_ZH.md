# Prism DNS 中文增强版

本 Fork 在原 Controller 前增加一个独立、可审计的增强层。原 Controller 和 Agent 仍使用上游发布的二进制，增强层负责简体中文界面、服务级路由、自定义域名、动态规则集和连通性测试。

- 仓库：[xcxcadc/chenfei-Glass-Prism-dns](https://github.com/xcxcadc/chenfei-Glass-Prism-dns)
- 最新版本：[GitHub Releases](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases/latest)

## 主要能力

- 默认简体中文，可随时切换英文和深浅主题。
- 动态读取 `stream.smartdns.list`，按分类和服务拆分为独立规则集。
- 每个 DNS 节点可把每项服务手动绑定到任意 Proxy Agent，并可恢复 Smart/Fallback 自动选择。
- 前端新增、编辑和删除自定义服务，支持任意名称、分类和域名列表。
- 自定义服务自动生成兼容 Prism 的 `DOMAIN-SUFFIX` 规则集。
- 同时提供上游 Agent 解锁检测和通用 DNS/TLS 连通性测试。
- 新增 IP 配置闭环：保存时自动创建 DNS Client、服务规则和逐服务解锁机覆盖。
- 提供 `prismdns.sh` 客户端工具，支持 Agent 安装、DNS 测试、系统 DNS 接管、备份和恢复。
- 客户端每分钟上报默认网卡 RX/TX 增量，面板显示全体与每 IP 流量并支持清零。
- 自定义数据保存在 `/var/lib/prism-enhancer/custom-services.json`。

## IP 配置使用流程

1. 先在“节点管理”中添加并上线至少一台 Proxy Agent 解锁机。
2. 打开“IP 配置”，填写需要使用 DNS 解锁的目标服务器公网 IPv4 或 IPv6。
3. 选择需要解锁的服务，并为每项服务指定任意在线 Proxy Agent。
4. 保存后，系统会自动创建专属 DNS Client 节点、规则和服务覆盖，并生成专属配置命令。
5. 在目标服务器以 `root` 身份执行该命令。脚本会安装 DNS Agent、检查本机 `127.0.0.1:53`、备份原 DNS，并在确认后接管系统 DNS。

需要重新打开命令时，可在该 IP 行点击“客户端脚本”。通用交互入口如下：

```bash
wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/prismdns.sh | sudo bash
```

## 流量统计口径

流量数字来自目标服务器默认网卡 `/sys/class/net/<interface>/statistics/{rx,tx}_bytes` 的计数差值，每分钟上报一次。它代表该服务器默认网卡的全部 RX/TX 流量，不是仅由 DNS 解锁服务产生的流量。首次上报只建立基线，不计入历史流量；服务器或网卡计数器重置后会从新值继续累计。

“清零流量”只重置面板累计值，不会清除系统网卡计数器，也不会影响 DNS 节点和服务配置。IP 配置及令牌保存在 `/var/lib/prism-enhancer/ip-configs.json`，文件权限为 `0600`；请勿公开页面生成的专属命令或配置令牌。

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
