# 版本对应关系

本 Fork 的面板/脚本版本、SNI transport 版本与上游 Agent 版本独立编号，不能混写。下表以 `v1.5.40` 为例；安装时按所选发布标签获取配套组件，不需要把表中的版本写进安装命令。其他版本见仓库的发布说明。

| 组件 | `v1.5.40` 发布组合 |
| --- | --- |
| Panel Enhancer | `1.5.40`（发布标签 `v1.5.40`） |
| 安装、升级、卸载及客户端脚本 | `v1.5.40` 标签 |
| 综合控制脚本与快捷方式 | `dns.sh` → `/usr/local/bin/dns`；`prismctl.sh` 保留兼容跳转，兼容 `PRISMCTL_*` |
| SNI transport (`prism_transport.sh`) | `2.5.3` |
| 官方媒体脚本 worker (`prism_media.sh`) | 随发布标签安装为 `/usr/local/bin/prism-media-check` |
| 节点更新 worker (`prism_node_update.sh`) | 随 transport 安装，验证正式不可变发布；旧节点须先接入一次 |
| Controller | 上游稳定 Controller |
| Proxy/DNS 兼容 Agent | 上游 `v1.2.1`（默认校验并锁定） |

## 综合控制入口

`bootstrap.sh` 解析 GitHub 最新公开稳定发布并执行对应标签的 `dns.sh`。以下命令在 root Bash 中执行；非 root 用户先运行 `sudo -i`，已是 root 则跳过，再在该 shell 中展开进程替换，避免跨 sudo 使用 `/dev/fd` 路径。

```bash
sudo -i
bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/bootstrap.sh) self-install
dns
```

不带子命令运行 `bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/bootstrap.sh)` 也会打开数字菜单，并自动安装快捷方式。`self-install` 原子安装 `/usr/local/bin/dns`，默认拒绝覆盖无关的同名 `dns`；已确认属于 Prism 的旧 `prismctl` 入口会原子迁移为指向 `dns` 的兼容符号链接，无关的 `prismctl` 保持原样。升级顺序为面板 → Proxy → Client；新版 `dns upgrade` 默认取最新公开稳定版并沿用凭据与传输模式，`dns upgrade panel` 保留现有端口和数据目录，并保留失败自动回滚。需要手动回滚时，用 `--ref vX.Y.Z` 指定目标发布标签。

默认选最新版的行为从 `v1.5.37` 起生效；更早安装的 `dns` 需运行 `dns upgrade --ref latest`，或通过 `bootstrap.sh` 运行 `upgrade`。

## 安装对应 Agent

Proxy 可在“节点管理”中复制专属安装命令，或使用已安装的快捷方式；Client 优先执行“IP 配置”生成的专属命令，以同时安装 runtime、transport 与 worker。以下命令中的地址和密钥须替换为面板生成的值：

```bash
dns install proxy --master 'https://panel.example.com' --secret '替换为节点密钥'
```

默认上游 Agent 由所选 Prism 发布确定。表中示例校验并锁定上游 `v1.2.1`，避免上游 `v1.3` 被动熔断影响固定路由；重装和卸载会自动解锁。这不代表把面板或 transport 降到同号版本。

## 已有节点与未来新机器

- **仅更换 enhancer 二进制不会自动更新远端节点。** 在每台已有有效配置的 Proxy/Client 上更新快捷方式后，运行 `dns upgrade`；缺少有效配置时重新执行当前面板生成的专属命令，刷新 runtime、transport 和 worker。旧脚本的 `--sync` 不等于升级。
- transport 安装默认同时安装 `prism-media-check.timer` / `prism-media-check.service`，空闲时每 5 秒 claim 任务，独立于 transport/traffic；卸载时一并移除。以后新增的 Proxy/Client 默认继承，无需逐台另装 worker。
- worker 执行官方脚本 `-M4 -R0 -Een`，即交互菜单 `0` 的跨国平台、仅 IPv4；逐批每秒更新原始 label 与 `Yes` / `No`、地区、WAF，不被 DNS/TLS 诊断覆盖。15 分钟仅作异常安全兜底，90 秒无心跳显示失联/失败。
- 六小时自动刷新、手动重测和配置代更新使用任务机制；旧 `generation` 结果拒收。官方脚本结果不保证实际付费播放或账号登录成功。
- 具体步骤见 [增强说明](ENHANCED_ZH.md#已有节点升级)、[迁移教程](MIGRATION_ZH.md) 和 [卸载说明](UNINSTALL_ZH.md)。
