# v1.6.3 更新日志

## 极简一键对接短链 & 多系统 53 端口自动兼容

1. **极简统一对接命令（告别冗长参数与多套命令差异）**
   - 面板新增短链入口 `GET /i/<credential>`（同时兼容 `/enhancer/i/<credential>`）：
     ```bash
     curl -fsSL https://你的面板域名/i/<密钥或Token> | bash
     ```
   - 自动识别该凭证属于 **解锁机 (Proxy)**、**受管 IP 被解锁机 (Client)** 还是 **手动 DNS 节点**，自动提权（非 root 自动 `sudo`）、优先从面板自托管 `/enhancer/dist/` 高速拉取脚本（免 GitHub 访问受限），彻底统一节点管理与 IP 配置两套冗长命令。
   - GitHub 仓库根目录新增极简入口 `i.sh`，支持跨系统一键安装面板/控制菜单：
     ```bash
     curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/i.sh | bash
     ```

2. **多 Linux 发行版 53 端口与依赖全自动兼容**
   - 修复 Ubuntu 18/20/22/24、Debian 12+、Fedora 等系统默认 `systemd-resolved` 占用 `127.0.0.53:53` 导致 `prism-agent` 启动报 `port 53 busy` 的问题：`agent_install.sh` 自动写入 `/etc/systemd/resolved.conf.d/99-prism-dns.conf` 关闭 `DNSStubListener=no`，自动备份并替换 stub `/etc/resolv.conf`，并兼容清理 `dnsmasq`、`sniproxy`、`unbound`、`named`、`bind9`、`smartdns`。
   - `ensure_dependencies()` 全面兼容 `apt-get`、`dnf`、`yum`、`apk`（Alpine）、`pacman`、`zypper`。
   - `prism-agent` 成功监听 53 端口后自动将 `/etc/resolv.conf` 首位指向 `127.0.0.1`（保留 `1.1.1.1` / `8.8.8.8` 兜底）。
