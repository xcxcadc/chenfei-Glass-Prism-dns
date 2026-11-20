# v1.5.37：Gemini 域名与节点在线维护

## Gemini

- 综合 Gemini 网页启动配置、SmartDNS 名单和用户提供的参考规则，补齐 `geminiweb-pa.clients.google.com`、`geminiweb-pa.clients6.google.com`、`push.clients6.google.com`、Antigravity 和 Cloud Code 的具体主机名。
- 保留 Google AI Studio 为独立服务，`aistudio.google.com`、`makersuite.google.com`、`generativelanguage.googleapis.com` 归入该服务。未加入整个 `clients6.google.com`、`googleusercontent.com`、`googleapis.com` 或不限定范围的关键词匹配。
- 内置域名更新不覆盖管理员手动编辑的名单。如果 Gemini 显示已自定义域名，在保存时将所需的新域名合并进去；不要为了获取新域名而丢掉自己的规则。
- Gemini 的地区提示是应用层结果。DNS 返回正确地址或 HTTPS 返回 200 不代表可用；应从目标用户的实际节点访问 Gemini 网页，确认出现会话输入框，并可发送消息获得响应。

## 节点管理

- 增加节点版本、传输版本、Agent 版本、节点报告的 IPv4 和滚动更新的心跳；配置 IP 与节点报告的本机出口路由源 IPv4 分别展示。NAT 下两者可能不同。
- 新节点随 Prism Transport 安装管理定时器；旧节点需先通过该角色的原有安装或升级入口接入一次。面板不能直接更新没有安装管理程序的节点。
- 节点通过 HTTPS 使用现有凭据上报。面板仅在对应正式不可变版本可用时接受更新请求。管理员逐台确认后，节点独立核验发布，从固定 GitHub 仓库下载；不会自动升级、降级，也不会接受面板提供的任意命令或下载地址。
- 在线更新覆盖 Transport、媒体检测和管理程序，不替换上游 Agent 二进制、节点凭据和路由配置，不操作业务服务。面板自身以及需要替换上游 Agent 的升级仍使用原有 `dns` 命令。
- 发布顺序：为仓库启用不可变发布，完整发布 `v1.5.37` 后再升级面板；接入一台测试节点确认心跳后再逐步推广。后续版本先完成发布，再升级面板并逐台更新节点。

## 安装与升级

- `dns upgrade panel --ref v1.5.37` 使用同一发布标签的增强层二进制，不再隐式下载 GitHub 的 `latest` 资产。只有版本已发布且资产齐全后才运行该命令。
- 升级前在服务器本地备份 Controller 数据库、环境文件、整个 Enhancer 数据目录、二进制和 systemd 单元。保留旧版本，确认公网面板和实际解锁链路正常后，再逐台升级其他节点。
