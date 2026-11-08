# v1.5.35：prismctl 数字菜单与日志查看

- 直接运行 `prismctl` 进入数字菜单：1 安装解锁机、2 安装被解锁机、3 安装面板、4 升级本机、5/6/7 分别卸载解锁机/被解锁机/面板、8 状态、9 日志、10 诊断、11 切换传输模式、12 更新自身、0 退出。输入编号即执行，安装时按提示输入面板地址、密钥或令牌。
- 日志：菜单 9 列出本机实际存在的 Prism 服务日志（Agent、出口代理、本地 DNS、路由同步、传输同步、加密隧道、流量上报、媒体检测、控制器、增强层），可选最近 100 行、实时跟踪、只看警告和错误、自定义行数。也可用子命令 `prismctl logs route-sync -f`。
- 菜单中单个操作失败不会退出整个菜单。
- 子命令方式保留，便于脚本调用。

## 使用

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.35/prismctl.sh | sudo bash -s -- self-install
sudo prismctl
```
