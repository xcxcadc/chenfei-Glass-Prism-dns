# v1.5.33：综合控制脚本 prismctl 与面板改动更快生效

## 新增：综合控制脚本 prismctl

面板、解锁机、被解锁机的安装、升级、卸载、状态、诊断、传输模式切换统一到一个脚本，形成闭环：

- `install panel|proxy|client`、`upgrade`、`uninstall`、`mode`、`status`、`doctor`、`self-install`、`self-update`。
- 升级自动读取本机已保存的面板地址、节点密钥和传输模式并原样带上，不需要重新输入令牌；**升级不会把加密传输模式重置回直连**（之前直接重跑安装脚本会重置）。
- 子脚本按发布标签下载，下载后先校验非空与语法再执行；标签格式有校验，`--ref latest` 取 GitHub 最新发布。
- `doctor` 检查服务状态、解锁机 `--ip` 是否为本机真实地址、443 监听与 ufw 放行、隧道用户，以及被解锁机到解锁机的链路是否被过滤明文 SNI，并给出处理建议。
- 只处理 Prism 自有的服务、nft 表和 iptables 链，不触碰 V2bX、XrayR、Docker、nginx。

## 改进：面板里新增/删除服务后更快生效

- 被解锁机路由变化后重启 Agent 的限流从 60 秒降到 10 秒，连续两次面板操作不再被拖慢。
- 路由变化时主动触发一次传输同步，出口代理的域名策略与隧道立即跟上。
- 被解锁机传输同步周期从 60 秒缩短到 10 秒（解锁机本来就是 5 秒）。
- 路由同步本身一直是每 10 秒轮询。

## 测试

- 新增 `tests/prismctl-test.sh`（配置读取、角色探测、参数解析、标签与模式校验）。
- 全部 8 个 shell 测试在 Linux 上通过；Windows 下 `prism-media-test.sh`、`prismdns-audit-test.sh` 因缺 `setsid`、时序差异会失败，请在 Linux 上运行。

## 升级

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.33/prismctl.sh | sudo bash -s -- self-install
sudo prismctl upgrade
```

顺序：先面板，再解锁机，最后被解锁机。
