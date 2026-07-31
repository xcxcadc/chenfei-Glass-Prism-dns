# Prism DNS 一键卸载

以下命令只清理 Prism DNS 组件，不会修改 XrayR、V2bX、Docker 或 MTProxy。

## 卸载面板并保留数据

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/uninstall.sh | bash -s -- panel --yes
```

保留的数据位于 `/opt/prism` 和 `/var/lib/prism-enhancer`，重新安装时可以继续使用。

## 彻底删除面板和数据

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/uninstall.sh | bash -s -- panel --purge-data --yes
```

此命令会永久删除面板数据库、账号、节点、服务配置和自定义服务数据。

## 卸载解锁机 Agent

在提供流媒体出口的解锁机上执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/uninstall.sh | bash -s -- proxy --yes
```

## 卸载被解锁机 Agent

在使用 Prism DNS 的被解锁机上执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/uninstall.sh | bash -s -- client --yes
```

脚本会先停止 Prism DNS，再恢复安装前备份的系统 DNS。找不到备份但系统仍指向 `127.0.0.1` 时，会使用 `1.1.1.1` 和 `8.8.8.8`，避免卸载后断网。

## 自动识别 Agent 类型

不确定服务器是解锁机还是被解锁机时执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/uninstall.sh | bash -s -- agent --yes
```

面板卸载不会远程删除其他服务器上的 Agent。每台解锁机和被解锁机需要分别执行对应命令。
