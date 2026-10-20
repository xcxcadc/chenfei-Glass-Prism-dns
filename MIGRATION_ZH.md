# Prism DNS 面板迁移与灾备

本文适用于使用 `enhanced_install.sh` 安装的面板。当前面板/脚本为 `1.5.29`，transport 为 `2.5.1`，默认兼容 Agent 为上游 `v1.2.1`，版本不可混写。默认路径如下：

| 数据 | 默认路径 | 内容 |
|---|---|---|
| Controller 数据库 | `/opt/prism/data.db` | 管理员账号、节点、密钥、规则和运行状态 |
| Controller 环境 | `/opt/prism/.env` | JWT 密钥与 Controller 环境配置 |
| Enhancer 数据目录（`dataDir`） | `/var/lib/prism-enhancer` | 自定义服务与分类、内置服务域名覆盖及永久删除标记、IP 配置、路由、流量、节点中文名称/地区、品牌设置、传输信息、服务图标缓存，以及 `media-checks.json` 的媒体任务进度和历史 |

迁移面板时必须同时迁移这三项。只复制 `data.db` 会丢失 IP 配置、流量、服务分类、中文名称、地区、站点标题及媒体任务进度和历史。下面的整目录备份已覆盖 `media-checks.json`，不要将其排除或在恢复时重置为空文件；自定义 `--data-dir` 时，须将备份、恢复和权限命令中的默认目录替换为实际 `dataDir`。

## 一、旧服务器导出

以下命令需要 root 权限。停止时间通常只有生成压缩包所需的几十秒。

```bash
sudo -i
set -e

BACKUP="/root/prism-panel-$(date +%Y%m%d-%H%M%S).tar.gz"
systemctl stop prism-enhancer prism-controller

FILES=(
  opt/prism/data.db
  opt/prism/.env
  var/lib/prism-enhancer
)
[ -f /opt/prism/data.db-wal ] && FILES+=(opt/prism/data.db-wal)
[ -f /opt/prism/data.db-shm ] && FILES+=(opt/prism/data.db-shm)

tar -C / -czpf "$BACKUP" "${FILES[@]}"
(cd "$(dirname "$BACKUP")" && sha256sum "$(basename "$BACKUP")" > "$(basename "$BACKUP").sha256")
tar -tzf "$BACKUP"

systemctl start prism-controller prism-enhancer
echo "备份文件: $BACKUP"
echo "校验文件: ${BACKUP}.sha256"
```

确认压缩包列表中至少存在：

```text
opt/prism/data.db
opt/prism/.env
var/lib/prism-enhancer/
```

使用过新版媒体任务后，还应确认归档包含 `var/lib/prism-enhancer/media-checks.json`；从尚无此文件的旧版迁移时，新面板会在实际 `dataDir` 中创建该存储。保留文件可在关窗、重开或面板重启后读取已保存的进度和历史；迁移不会让已经中断或失联的 worker 进程从断点续跑。

## 配置文件导出与脱敏迁移

如果只需要迁移 IP 配置而不迁移整套面板数据库，可在旧面板“IP 配置”页面点击“导出配置”，把下载的 `prismdns-ip-configs-v1.json` 通过加密通道传到新面板，再点击“导入配置”。该文件不包含账号密码、节点密钥、安装令牌、流量、健康状态或 `media-checks.json` 的任务进度/历史；新面板会按解锁机公网 IPv4/名称匹配节点，匹配不到时拒绝导入，不会静默切换线路。

导入新目标 IP 后，必须在导入结果中打开“客户端脚本”，在对应目标服务器执行一次新命令；导入不会自动远程登录或把旧令牌复制到新面板。同一个目标 IP 再次导入会更新服务路由并保留其累计流量。若需要完整保留令牌、节点 ID、账号和流量，仍应使用本教程前面的整套压缩备份恢复流程。

如怀疑安装命令泄露，可在旧面板 IP 行点击“轮换令牌”。轮换会立即撤销旧命令并将目标状态置为待重新安装；在目标机执行新命令后再确认状态恢复为 `READY`。

将 `.tar.gz` 和 `.sha256` 一起复制到新服务器。备份中包含节点密钥和专属令牌，必须使用 SSH/SCP 等加密通道传输，不能公开上传。

## 二、新服务器安装基础程序

建议新旧面板使用相同域名和公开端口。这样只需把域名的 A/AAAA 记录切换到新服务器，现有解锁机与被解锁机不需要逐台修改面板地址。

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/enhanced_install.sh | sudo bash
```

如果旧服务器使用自定义端口，新服务器必须保持一致：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.29/enhanced_install.sh \
  | sudo PRISM_PORT=8081 PRISM_CORE_PORT=18080 bash
```

基础安装完成后，不要在新面板中新增节点或服务，直接进入恢复步骤。

## 三、新服务器恢复

把备份和校验文件按原文件名上传到 `/root/`，然后执行：

```bash
sudo -i
set -e

BACKUP="$(ls -1t /root/prism-panel-*.tar.gz | head -n 1)"
cd /root
sha256sum -c "$(basename "${BACKUP}.sha256")"
tar -tzf "$BACKUP"

systemctl stop prism-enhancer prism-controller
cp -a /opt/prism/data.db "/root/data.db.before-restore.$(date +%s)"
cp -a /var/lib/prism-enhancer "/root/prism-enhancer.before-restore.$(date +%s)"

tar -C / -xzpf "$BACKUP"
chown root:root /opt/prism/data.db /opt/prism/.env
chmod 600 /opt/prism/data.db /opt/prism/.env
chown -R root:root /var/lib/prism-enhancer
chmod 750 /var/lib/prism-enhancer
find /var/lib/prism-enhancer -type f -exec chmod 600 {} \;

systemctl restart prism-controller
sleep 2
systemctl restart prism-enhancer
sleep 2
systemctl is-active --quiet prism-controller
systemctl is-active --quiet prism-enhancer
```

## 四、恢复后验证

先在新服务器本机检查：

```bash
curl -fsS http://127.0.0.1:8080/enhancer/api/health
systemctl --no-pager --full status prism-controller prism-enhancer
journalctl -u prism-controller -u prism-enhancer --since "-10 min" --no-pager
```

如果公开端口不是 `8080`，请替换命令中的端口。

然后登录 Web 面板，逐项确认：

1. 原管理员账号可以登录。
2. 节点数量、名称、地区、分组和在线状态正确。
3. 自定义服务、域名、分类及服务分类归属正确。
4. 每个目标 IP 的服务选择、解锁机映射和流量数据正确。
5. 左上角网页名称和浏览器标签标题正确。
6. 重新打开媒体检测窗口，确认 `media-checks.json` 中已有任务进度、结果和历史可读；旧任务失联或失败应显示明确原因，不得伪装成通过。
7. 节点完成本版脚本升级后，手动重测一个 Proxy/Client，确认完整 IPv4 菜单 `0` 检测逐批每秒刷新，官方 label 的 `Yes` / `No`、地区、WAF 原值保留；DNS/AAAA、TLS/SNI 和代表页面作为独立诊断，不覆盖媒体结论。
8. 等待 1 分钟，确认目标 IP 流量继续上报，并检查节点的 `prism-media-check.timer` 已启用。

健康接口中的 `status` 应为 `ok`，`authorization_mode` 应为 `panel_allowlist`，`authorization_sync_seconds` 应为 `5`，所有已安装目标机应逐步恢复为 `READY`。

增强层首次启动会迁移旧路由数据：移除持久化 Proxy IPv6，将共享域名服务归一到同一 Proxy，清理旧版自动故障切换记录并保留用户选择的路由、节点 ID、目标 IP、令牌、累计流量和采样基线。媒体任务随 `media-checks.json` 保留；配置代变化后拒收旧 `generation` 的迟到结果，避免旧线路覆盖当前结果。旧 full / 四阶段检测不再作为最终媒体结论，DNS 与路由诊断独立保留。

**仅更换 enhancer 二进制不会自动刷新远端节点。** 每台 Proxy 需重新执行本版 `v1.5.29` Agent 命令，每台 Client 需重新执行本版“IP 配置”专属命令；已有有效配置的 Client 也可执行本版 `prismdns.sh --refresh-runtime`，详见 [已有节点升级](ENHANCED_ZH.md#已有节点升级)。这会刷新 runtime、transport `2.5.1` 和 `/usr/local/bin/prism-media-check`；只运行旧 transport 的 `--sync` 不会升级脚本或补装 worker。专用 dnsmasq、动态路由和本地 53 重定向由面板数据重新生成，不需要从旧目标机复制。所有未来新 Proxy/Client 默认随 transport 安装 worker，并随 transport / Agent 卸载。

worker 的 `prism-media-check.timer` / `prism-media-check.service` 空闲时每 5 秒 claim 任务，独立于 transport/traffic。运行等价于 `bash check.unlock.media -M4 -R0 -Een` 的官方脚本：交互菜单 `0`、只测跨国平台、仅 IPv4。有输出即逐批上报，界面每秒刷新；完整运行，无 30 秒硬停，15 分钟仅作异常安全兜底并明确 `failed`，90 秒无心跳显示失联/失败。每 6 小时自动刷新，也支持手动重测和配置代更新，关窗重开可读任务。**官方脚本结果不保证实际付费播放、账号登录或所有应用功能通过。**

## 五、切换域名或面板地址

最佳方案是保留原域名，只修改 DNS 解析到新面板 IP。切换前可把 DNS TTL 降至 60 秒。

如果必须更换域名、协议或端口：

1. 在每台解锁机的“节点管理”中重新复制 Agent 安装命令并执行。
2. 在每台被解锁机的“IP 配置”中重新复制专属命令并执行。
3. 不要删除并重建原节点或 IP 配置，否则会生成新的节点密钥和专属令牌。
4. 确认新地址可访问后，再关闭旧面板。

重新执行本版安装命令会更新兼容 Agent、面板地址、runtime、transport 和 worker。稳定安装器验证并锁定上游 Agent `v1.2.1`，不是 `v1.3`；安装和卸载会自动解除旧文件锁，不会读取、修改或重启 XrayR、MTProxy、V2bX、Docker 等其他业务。

transport `2.5.1` 保留显式 `--master` 的优先级，避免旧 `/etc/prismdns/transport.env` 覆盖新地址。旧 transport/client master 地址失效时，更新后的 transport 可从运行中 Agent 的 `--master` 恢复新地址，校验响应后写回配置，worker 下一轮读取同一配置；这要求 Agent 已指向新面板、原密钥/令牌有效，旧脚本仍需先升级。必要时可在更新脚本后手动同步新地址：

```bash
sudo /usr/local/lib/prismdns/prism_transport.sh --sync --master <新面板地址>
```

Proxy 每 5 秒从面板刷新授权白名单；AUTH nftables 规则先通过校验，再在单次事务中原子替换，避免先删后建造成的授权空窗。IPv4 限制 Prism DNS 53 与 SNI 80/443；IPv6 只拒绝 DNS 53，不拦截 IPv6 80/443，也不修改其他业务规则。迁移后确认 transport 配置指向新地址、worker 心跳及逐批结果恢复，再下线旧面板。

## 六、失败回滚

如果新面板恢复失败，先保持旧面板继续运行和解析。新服务器可使用恢复前自动复制的文件回滚：

```bash
sudo -i
systemctl stop prism-enhancer prism-controller

cp -a /root/data.db.before-restore.TIMESTAMP /opt/prism/data.db
mv /var/lib/prism-enhancer "/var/lib/prism-enhancer.failed.$(date +%s)"
cp -a /root/prism-enhancer.before-restore.TIMESTAMP /var/lib/prism-enhancer

systemctl restart prism-controller prism-enhancer
```

先用 `ls -lt /root/data.db.before-restore.* /root/prism-enhancer.before-restore.*` 找到同一时间生成的一组，再把命令中的 `TIMESTAMP` 替换为准确时间戳。

## 七、日常自动备份建议

至少每天备份一次，并保留 7 至 30 天：

```bash
sudo -i
install -d -m 700 /var/backups/prism
systemctl stop prism-enhancer prism-controller
tar -C / -czpf "/var/backups/prism/prism-$(date +%Y%m%d-%H%M%S).tar.gz" \
  opt/prism/data.db opt/prism/.env var/lib/prism-enhancer
systemctl start prism-controller prism-enhancer
find /var/backups/prism -type f -name 'prism-*.tar.gz' -mtime +30 -delete
```

生产环境建议把备份同步到另一台服务器或对象存储，并定期在临时机器上执行一次恢复演练。
