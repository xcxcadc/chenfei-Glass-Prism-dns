# Prism DNS 面板迁移与灾备

本文适用于使用 `enhanced_install.sh` 安装的面板。默认路径如下：

| 数据 | 默认路径 | 内容 |
|---|---|---|
| Controller 数据库 | `/opt/prism/data.db` | 管理员账号、节点、密钥、规则和运行状态 |
| Controller 环境 | `/opt/prism/.env` | JWT 密钥与 Controller 环境配置 |
| Enhancer 数据目录 | `/var/lib/prism-enhancer` | 自定义服务与分类、IP 配置、路由、流量、节点中文名称/地区、品牌设置和传输信息 |

迁移面板时必须同时迁移这三项。只复制 `data.db` 会丢失 IP 配置、流量、服务分类、中文名称、地区和站点标题。

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

将 `.tar.gz` 和 `.sha256` 一起复制到新服务器。备份中包含节点密钥和专属令牌，必须使用 SSH/SCP 等加密通道传输，不能公开上传。

## 二、新服务器安装基础程序

建议新旧面板使用相同域名和公开端口。这样只需把域名的 A/AAAA 记录切换到新服务器，现有解锁机与被解锁机不需要逐台修改面板地址。

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/enhanced_install.sh | sudo bash
```

如果旧服务器使用自定义端口，新服务器必须保持一致：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/enhanced_install.sh \
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
6. 随机选择一个目标 IP，执行 Gemini、Claude、Disney+ 或 YouTube 的真实解锁测试。
7. 等待 1 分钟，确认目标 IP 流量继续上报。

健康接口中的 `status` 应为 `ok`，所有已安装目标机应逐步恢复为 `READY`。

## 五、切换域名或面板地址

最佳方案是保留原域名，只修改 DNS 解析到新面板 IP。切换前可把 DNS TTL 降至 60 秒。

如果必须更换域名、协议或端口：

1. 在每台解锁机的“节点管理”中重新复制 Agent 安装命令并执行。
2. 在每台被解锁机的“IP 配置”中重新复制专属命令并执行。
3. 不要删除并重建原节点或 IP 配置，否则会生成新的节点密钥和专属令牌。
4. 确认新地址可访问后，再关闭旧面板。

重新执行安装命令只更新 Prism Agent、面板地址和守卫配置，不会修改 MTProxy、XrayR 或 V2bX。

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
