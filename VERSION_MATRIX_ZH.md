# 版本对应关系

本项目的面板/脚本版本、SNI transport 版本与上游 Agent 版本独立编号。下表以当前最新稳定版 `v1.6.5` 为例：

| 组件 | `v1.6.5` 发布组合 |
| --- | --- |
| Panel Enhancer | `1.6.5`（发布标签 `v1.6.5`） |
| 安装、升级、卸载及客户端脚本 | `v1.6.5` 标签（`i.sh` / `bootstrap.sh` / `dns.sh` / `agent_install.sh` / `prismdns.sh`） |
| 综合控制脚本与快捷方式 | `dns.sh` → `/usr/local/bin/dns`（含 `dns warp`、`dns doctor`、`dns upgrade`） |
| SNI transport (`prism_transport.sh`) | `2.6.5`（含自动 WARP 启停、30 秒数据面自愈与客户端原生直连探测） |
| 出口代理 (`prism-egress-proxy`) | `v1.6.5`（含按服务 WARP 分流、4 秒短超时原生回退与 25 秒熔断器） |
| 官方媒体脚本 worker (`prism_media.sh`) | 随发布标签安装为 `/usr/local/bin/prism-media-check` |
| 节点更新 worker (`prism_node_update.sh`) | 随 transport 安装，支持验证不可变 Release 与面板自托管分发 |
| Controller / Proxy-DNS Agent | 上游稳定 Controller + 锁定兼容 Agent `v1.2.1` |
