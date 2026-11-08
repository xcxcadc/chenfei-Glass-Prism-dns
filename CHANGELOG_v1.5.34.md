# v1.5.34：面板升级保留端口、失败自动回滚

## 修复

- **`enhanced_install.sh` 重装/升级时不再改掉已有端口与目录。** 之前重新生成 systemd 单元时，会把增强层监听端口写回默认的 8080；若面板由 nginx 反代到自定义端口（例如 58082）而 8080 已被占用，增强层会因 `address already in use` 反复重启，面板不可用。现在会沿用现有单元里的监听端口、控制器端口和数据目录；只有显式设置 `PRISM_PORT`、`PRISM_CORE_PORT`、`PRISM_ENHANCER_DATA_DIR` 时才覆盖。
- **`prismctl upgrade panel` 升级前备份二进制、数据库和 systemd 单元，升级后检查控制器与增强层是否恢复运行；没有恢复则自动回滚并报错。**
- 发布附带 `prism-enhancer`、`prism-egress-proxy` 的 linux amd64/arm64 二进制和 `SHA256SUMS.txt`（`enhanced_install.sh` 与 `prism_transport.sh` 从 `releases/latest/download` 下载它们，没有资产会 404）。

## 说明

- 被解锁机拿到的传输脚本地址由**面板**返回，所以必须先升级面板，再升级解锁机与被解锁机，新的同步周期才会落到被解锁机上。
- 新增 `tests/enhanced-install-config-test.sh`。

## 升级

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.34/prismctl.sh | sudo bash -s -- self-install
sudo prismctl upgrade
```

顺序：面板 → 解锁机 → 被解锁机。
