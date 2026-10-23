# 1.5.30 版本对应关系

本 Fork 的面板/脚本版本、SNI transport 版本与上游 Agent 版本独立编号，不能混写。当前发布组合如下；兼容 Agent 默认固定为上游 `v1.2.1`，不是 `v1.3`，也不存在与本 Fork 同号的 `agent-v1.5.30` 安装要求。

| 组件 | 固定版本 |
| --- | --- |
| Panel Enhancer | `1.5.30`（发布标签 `v1.5.30`） |
| 安装、升级、卸载及客户端脚本 | `v1.5.30` 标签 |
| SNI transport (`prism_transport.sh`) | `2.5.2` |
| 官方媒体脚本 worker (`prism_media.sh`) | 随 `v1.5.30` 发布，安装为 `/usr/local/bin/prism-media-check` |
| Controller | 上游稳定 Controller |
| Proxy/DNS 兼容 Agent | 上游 `v1.2.1`（默认校验并锁定） |

## 安装对应 Agent

Proxy 在“节点管理”中复制本版命令；Client 优先执行“IP 配置”生成的专属命令，以同时安装 runtime、transport 与 worker。以下通用 Agent 示例显式固定兼容版本，不替代 Client 的专属配置流程：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/agent_install.sh | sudo bash -s -- --version v1.2.1 --master <面板地址> --secret <节点密钥>
```

稳定安装器省略 `--version` 时同样默认使用 `v1.2.1`，校验 SHA-256 并锁定二进制；重装和卸载会自动解锁。该兼容选择避免上游 `v1.3` 被动熔断影响固定路由，不代表把面板或 transport 降到同号版本。`--beta` 是单独的 Beta 发布线，不属于上述默认稳定组合。

## 已有节点与未来新机器

- **仅更换 enhancer 二进制不会自动更新远端节点。** 已有 Proxy 需重新执行 `v1.5.30` 的 Agent 安装命令，已有 Client 需重新执行本版专属命令或用本版 `prismdns.sh --refresh-runtime`，刷新 runtime、transport 和 worker；旧脚本的 `--sync` 不等于升级。
- transport 安装默认同时安装 `prism-media-check.timer` / `prism-media-check.service`，空闲时每 5 秒 claim 任务，独立于 transport/traffic；卸载时一并移除。以后新增的 Proxy/Client 默认继承，无需逐台另装 worker。
- worker 执行官方脚本 `-M4 -R0 -Een`，即交互菜单 `0` 的跨国平台、仅 IPv4；逐批每秒更新原始 label 与 `Yes` / `No`、地区、WAF，不被 DNS/TLS 诊断覆盖。15 分钟仅作异常安全兜底，90 秒无心跳显示失联/失败。
- 六小时自动刷新、手动重测和配置代更新使用任务机制；旧 `generation` 结果拒收。官方脚本结果不保证实际付费播放或账号登录成功。
- 安装与卸载入口统一使用 `v1.5.30`；具体步骤见 [增强说明](ENHANCED_ZH.md#已有节点升级)、[迁移教程](MIGRATION_ZH.md) 和 [卸载说明](UNINSTALL_ZH.md)。
