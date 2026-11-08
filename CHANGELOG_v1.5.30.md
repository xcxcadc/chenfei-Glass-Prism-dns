# v1.5.30：新增节点检测恢复

- 修复前端生成的节点安装命令仍固定引用 `v1.5.28` 的问题。发布引用由面板统一渲染，避免面板、页面和脚本版本漂移。
- 修复新节点 Agent 已在线，但旧安装流程没有安装官方 IPv4 检测 Worker，导致“任务失败 · 已返回项目 0”。Agent 在线与媒体检测完成仍为独立状态。
- 首次安装先安装检测 Worker，再执行首次传输同步；同步暂时失败不再导致检测组件被跳过。补齐 `ss`、`flock`、`setsid` 前置依赖。
- 安装或恢复 Worker 后，自动重建此前因无人领取而超时的任务，不必等待六小时。运行过的脚本错误不会无限自动重试，历史失败和原始结果保留。
- 解锁机和被解锁机均支持非侵入式 runtime 刷新，不重装 Agent、不修改业务代理配置。
- 节点卡片明确区分“检测未启动”和服务结果；失败原因可查看，不能把零项任务失败当成全部服务不可用。

## 已有节点升级

面板升级后，所有新生成安装命令自动使用 `v1.5.30`。已经安装旧 runtime 的服务器可运行：

```bash
installer="$(mktemp)"
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/prism_transport.sh -o "$installer" \
  && sudo bash "$installer" --refresh-runtime
rm -f "$installer"
```

该命令复用 `/etc/prismdns/transport.env`，支持 Proxy 和已有 Client，不重置路由、密钥或累计流量。Client 完整刷新也可使用面板 IP 配置专属命令或 `prismdns.sh --refresh-runtime`。

## 检查与回滚

```bash
systemctl is-active prism-media-check.timer
journalctl -u prism-media-check.service -n 30 --no-pager
```

重新打开“解锁机自检 / 查看任务”，结果应随官方 `check.unlock.media` 菜单 `0` 的 IPv4 输出逐批出现，直至真正结束。`Yes`、`No`、地区和 WAF 原样保留；检测完成不等于所有平台均可用。

升级前请备份面板二进制、`/opt/prism/data.db`、`/var/lib/prism-enhancer`，以及节点的 `/etc/prismdns`、Prism runtime 和 systemd 单元。需要回滚时仅恢复对应 Prism 文件与服务，不操作 XrayR、V2bX、MTProxy 等业务；完整迁移与卸载见 `MIGRATION_ZH.md` 和 `UNINSTALL_ZH.md`。
