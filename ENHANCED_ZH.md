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
- 自定义数据保存在 `/var/lib/prism-enhancer/custom-services.json`。

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

## 数据来源与边界

细化域名目录运行时读取自 [1-stream/1stream-public-utils](https://github.com/1-stream/1stream-public-utils) 的 `stream.smartdns.list`。该仓库目前未声明许可证，本 Fork 不提交其名单副本，只在部署实例中按用户请求实时读取和转换。

上游 `mslxi/Liquid-Glass-Prism-dns` 同样未提供 Controller/Agent 源码及许可证。增强层不复制、不修改上游二进制，只通过已公开的 HTTP API 与其协作。上游二进制的使用和分发仍受原作者权利约束。
