# v1.5.36：dns 快捷命令与 Bash 入口

## 综合控制入口

- 本地主命令改为 `dns`，规范脚本为 `dns.sh`；`prismctl.sh` 保留兼容跳转，`PRISMCTL_*` 旧环境变量继续可用。
- 继续从 GitHub 发布标签 URL 获取脚本，推荐使用 `bash <(curl -fsSL ...)`；直接运行菜单入口会自动安装 `dns` 快捷方式。
- `self-install` 将校验后的脚本原子安装到 `/usr/local/bin/dns`，默认拒绝覆盖无关的同名 `dns`。已确认属于 Prism 的旧 `prismctl` 入口会原子迁移为指向 `dns` 的兼容符号链接，无关的 `prismctl` 保持原样。
- 子脚本仍按标签下载，检查非空和 Bash 语法后执行；`--ref latest` 仍解析 GitHub 最新发布。

## 保留的菜单、日志与升级行为

- 数字菜单继续提供：1 安装解锁机、2 安装被解锁机、3 安装面板、4 升级本机、5/6/7 卸载解锁机/被解锁机/面板、8 状态、9 日志、10 诊断、11 切换传输模式、12 更新自身、0 退出。单项操作失败返回菜单，子命令调用方式保留。
- 菜单 9 只列出本机实际存在的 Prism 服务，支持最近 100 行、实时跟踪、警告错误、自定义行数；Ctrl-C 结束当前跟踪。
- 日志目标保留 `agent`、`egress`、`client-egress`、`dns`、`route-sync`、`transport`、`tunnel`、`traffic`、`media`、`controller`、`enhancer`、`all`，保留 `logs/log`、`-n/--lines`、`-f/--follow`、`--errors`。
- `dns upgrade` 继续自动探测本机角色；解锁机升级沿用面板地址、密钥、出口 IPv4、`--smart` 和传输模式，被解锁机刷新已有 runtime、transport 与媒体 worker，保留令牌、路由和累计流量。首次安装默认 `direct`，升级不会把 `encrypted` 重置为直连。
- `dns upgrade panel` 保留现有公开端口、控制器端口和 Enhancer 数据目录；只有显式设置 `PRISM_PORT`、`PRISM_CORE_PORT`、`PRISM_ENHANCER_DATA_DIR` 才覆盖。升级前备份二进制、数据库和 systemd 单元，升级后检查服务，失败自动回滚并返回失败。
- 卸载继续按角色执行，面板默认保留数据，`--purge-data` 显式删除数据，`--yes` 跳过确认；被解锁机恢复原系统 DNS，transport/Agent 卸载同时清理媒体 worker。只操作 Prism 自有服务与防火墙对象。
- 延续 [v1.5.31](CHANGELOG_v1.5.31.md) 的服务目录与审计修复、[v1.5.32](CHANGELOG_v1.5.32.md) 的加密传输排障、[v1.5.33](CHANGELOG_v1.5.33.md) 的配置复用与同步周期、[v1.5.34](CHANGELOG_v1.5.34.md) 的保端口和回滚、[v1.5.35](CHANGELOG_v1.5.35.md) 的数字菜单与日志行为。

## 使用与升级

以下命令在 **root Bash** 中运行。非 root 用户先执行 `sudo -i`，已是 root 则跳过；进程替换必须在该 root Bash 中展开，避免跨 sudo 传递 `/dev/fd` 路径。

```bash
sudo -i
bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.36/dns.sh) self-install
dns
```

直接打开菜单并自动安装快捷方式：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.36/dns.sh)
```

升级顺序为面板 → 解锁机 → 被解锁机。在每台对应主机安装本版快捷方式后执行：

```bash
dns upgrade --ref v1.5.36
dns status
dns logs route-sync -f
```

切换加密模式使用 `dns mode encrypted`，先切解锁机，再切被解锁机，两端模式须一致。路由守卫和 Client transport 同步周期为 10 秒，Proxy transport 为 5 秒。面板决定 Client bootstrap 返回的脚本地址，应先升级面板。

面板与发布脚本为 `1.5.36`；SNI transport 仍为 `2.5.2`，默认上游兼容 Agent 仍为 `v1.2.1`，版本独立编号。完整入口见 [README](README.md)、[中文增强说明](ENHANCED_ZH.md)、[迁移教程](MIGRATION_ZH.md)、[卸载说明](UNINSTALL_ZH.md) 和 [版本矩阵](VERSION_MATRIX_ZH.md)。
