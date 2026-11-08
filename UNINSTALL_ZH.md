# Prism DNS 一键卸载

以下安装/卸载配套入口统一使用 `v1.5.36`，只清理 Prism DNS 组件，不会修改 XrayR、V2bX、Docker 或 MTProxy。节点侧的媒体 worker 跟随 transport 安装和卸载，适用于现有节点及未来默认安装的所有 Proxy/Client。

以下快捷命令均在 root Bash 中运行。非 root 用户先执行 `sudo -i`，已是 root 则跳过；进程替换在该 root Bash 中展开，避免跨 sudo 传递 `/dev/fd`。尚未安装本版快捷方式时，先执行：

```bash
sudo -i
bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.36/dns.sh) self-install
```

已安装后可运行 `dns`，选择菜单 5/6/7 分别卸载解锁机/被解锁机/面板。也可直接运行 `bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.36/dns.sh)` 打开菜单并自动安装快捷方式。`self-install` 原子写入 `/usr/local/bin/dns`，默认拒绝覆盖无关的同名 `dns`；已确认属于 Prism 的旧 `prismctl` 入口会原子迁移为指向 `dns` 的兼容符号链接，无关的 `prismctl` 保持原样。旧 `prismctl.sh` 与 `PRISMCTL_*` 环境变量继续兼容。

## 卸载面板并保留数据

```bash
dns uninstall panel --yes
```

保留的数据位于 `/opt/prism` 和 Enhancer 的 `dataDir`（默认 `/var/lib/prism-enhancer`），包括 `media-checks.json` 中的媒体任务进度和历史，重新安装时可以继续使用；自定义目录请按 [迁移教程](MIGRATION_ZH.md) 另行确认备份。

## 彻底删除面板和数据

```bash
dns uninstall panel --purge-data --yes
```

此命令会永久删除面板数据库、账号、节点、服务配置、自定义服务数据，以及默认数据目录中 `media-checks.json` 的媒体任务进度和历史；执行前请备份整个数据目录。

## 卸载解锁机 Agent

在提供流媒体出口的解锁机上执行：

```bash
dns uninstall proxy --yes
```

## 卸载被解锁机 Agent

在使用 Prism DNS 的被解锁机上执行：

```bash
dns uninstall client --yes
```

脚本会先停止 Prism DNS，再恢复安装前备份的系统 DNS。找不到备份但系统仍指向 `127.0.0.1` 时，会使用 `1.1.1.1` 和 `8.8.8.8`，避免卸载后断网。

## 自动识别 Agent 类型

不确定服务器是解锁机还是被解锁机时执行：

```bash
dns uninstall agent --yes
```

省略 `--yes` 时会要求确认；`--purge-data` 仅用于面板。尚未安装快捷方式时，也可在 root Bash 中直接调用规范脚本，例如：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.36/dns.sh) uninstall client --yes
```

上述 `proxy` / `client` / `agent` 卸载会一并清理本机 transport 与 `/usr/local/bin/prism-media-check`，停止并移除 `prism-media-check.timer` / `prism-media-check.service`，终止正在运行的本机检测，不留下持续 claim 的 worker；其他 Prism 自有服务、授权规则及计数器按节点类型清理。单独卸载 transport 也会卸载媒体 worker，无需另外运行媒体卸载命令。

面板卸载不会远程删除其他服务器上的 Agent、transport 或媒体 worker。每台解锁机和被解锁机需要分别执行对应命令；只删除面板或只更换 enhancer 二进制都不等于刷新或清理远端运行时。
