# Prism DNS 一键卸载

以下安装/卸载配套入口统一使用 `v1.5.29`，只清理 Prism DNS 组件，不会修改 XrayR、V2bX、Docker 或 MTProxy。节点侧的媒体 worker 跟随 transport 安装和卸载，适用于现有节点及未来默认安装的所有 Proxy/Client。

## 卸载面板并保留数据

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/uninstall.sh | bash -s -- panel --yes
```

保留的数据位于 `/opt/prism` 和 Enhancer 的 `dataDir`（默认 `/var/lib/prism-enhancer`），包括 `media-checks.json` 中的媒体任务进度和历史，重新安装时可以继续使用；自定义目录请按 [迁移教程](MIGRATION_ZH.md) 另行确认备份。

## 彻底删除面板和数据

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/uninstall.sh | bash -s -- panel --purge-data --yes
```

此命令会永久删除面板数据库、账号、节点、服务配置、自定义服务数据，以及默认数据目录中 `media-checks.json` 的媒体任务进度和历史；执行前请备份整个数据目录。

## 卸载解锁机 Agent

在提供流媒体出口的解锁机上执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/uninstall.sh | bash -s -- proxy --yes
```

## 卸载被解锁机 Agent

在使用 Prism DNS 的被解锁机上执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/uninstall.sh | bash -s -- client --yes
```

脚本会先停止 Prism DNS，再恢复安装前备份的系统 DNS。找不到备份但系统仍指向 `127.0.0.1` 时，会使用 `1.1.1.1` 和 `8.8.8.8`，避免卸载后断网。

## 自动识别 Agent 类型

不确定服务器是解锁机还是被解锁机时执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/uninstall.sh | bash -s -- agent --yes
```

上述 `proxy` / `client` / `agent` 卸载会一并清理本机 transport 与 `/usr/local/bin/prism-media-check`，停止并移除 `prism-media-check.timer` / `prism-media-check.service`，终止正在运行的本机检测，不留下持续 claim 的 worker；其他 Prism 自有服务、授权规则及计数器按节点类型清理。单独卸载 transport 也会卸载媒体 worker，无需另外运行媒体卸载命令。

面板卸载不会远程删除其他服务器上的 Agent、transport 或媒体 worker。每台解锁机和被解锁机需要分别执行对应命令；只删除面板或只更换 enhancer 二进制都不等于刷新或清理远端运行时。
