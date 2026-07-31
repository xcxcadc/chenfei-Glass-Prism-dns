# 1.4.4 版本对应关系

`enhancer-v1.4.4` 是面板增强层版本。上游没有发布 `agent-v1.4.4`；在该面板版本发布时，稳定 Agent 发布线为上游 `v1.3`。

因此固定组合为：

| 组件 | 固定版本 |
| --- | --- |
| Panel Enhancer | `enhancer-v1.4.4` |
| Controller | 上游稳定 Controller |
| Proxy/DNS Agent | 上游 `v1.3` |

## 安装对应 Agent

面板添加节点后，在解锁机或被解锁机执行页面生成的命令，并显式锁定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh | bash -s -- --version v1.3 --master <面板地址> --secret <节点密钥>
```

安装器会校验 v1.3 Linux Agent 的 SHA256，不会静默回退到最新版本。

## 回退 Agent

需要与更早的面板/Agent 组合回退时，可以指定上游 `v1.2.1`：

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh | bash -s -- --version v1.2.1 --master <面板地址> --secret <节点密钥>
```

`--version` 只接受明确的稳定版本标签。`--beta` 仍然单独使用 Beta 发布线，不会覆盖稳定版本锁定逻辑。
