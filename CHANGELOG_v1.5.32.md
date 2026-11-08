# v1.5.32：YouTube 视频无法播放的根因与修复

## 现象

被解锁机能打开 YouTube 页面，但视频一直转圈；Google Play 商店网页打不开。

## 根因

部分机房网络会对 TLS 握手里**明文的 SNI** 做过滤。实测同一台被解锁机上，只要 SNI 是 `googlevideo.com`、`yt3.ggpht.com` 或 `play.google.com`，连接就被丢弃，**与目的 IP 无关**——无论直连 Google 还是转发给解锁机；把 SNI 改成大小写混合又能通。因此：

- 直连模式（`direct`）下，被解锁机到解锁机的 443 也是明文 SNI，同样被丢弃，解锁机根本收不到请求；
- 只调整域名清单无法解决，必须让 SNI 不再以明文出现在这段链路上。

## 修复

- 在这类网络里，把被解锁机与解锁机都切换到**加密传输模式**。SNI 被封装进 SSH 隧道后，视频流、图片 CDN 与 Play 商店域名全部恢复可达。
- 视频流与图片 CDN（`googlevideo.com`、`ggpht.com`、`ytimg.com`、`gvt1.com`、`gvt2.com`）保持经解锁机转发，并在代码中注明它们依赖加密模式。
- 探测域名与图标等细节随 YouTube、Google Play Store 一并校正。

## 如何判断需要加密模式

在被解锁机上执行：

```bash
curl -4 -sk -m6 -o /dev/null -w '%{http_code}\n' --resolve www.example.com:443:<解锁机IP> https://www.example.com/
curl -4 -sk -m6 -o /dev/null -w '%{http_code}\n' --resolve redirector.googlevideo.com:443:<解锁机IP> https://redirector.googlevideo.com/
```

第一条有响应、第二条无响应（`000`），即该网络在过滤 SNI，请启用加密模式。

## 启用加密模式

只会新增解锁机上的系统用户 `prism-tunnel`（走现有 sshd）和被解锁机上仅针对“目的地为解锁机 IP 的 80/443”的本地转发，不影响 V2bX、Docker、nginx 等业务。先在解锁机、再在被解锁机执行：

```bash
# 解锁机（Proxy）
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.32/prism_transport.sh -o /tmp/prism_transport.sh \
  && sudo bash /tmp/prism_transport.sh --proxy --secret <节点密钥> --master <面板地址> --mode encrypted

# 被解锁机（Client）
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.32/prism_transport.sh -o /tmp/prism_transport.sh \
  && sudo bash /tmp/prism_transport.sh --client --token <被解锁机令牌> --master <面板地址> --mode encrypted
```

用 `/usr/local/lib/prismdns/prism_transport.sh --status` 确认 `mode=encrypted`。改回明文模式只需把 `--mode` 换成 `direct` 重新执行。

## 其他

- 解锁机 `prism-agent` 的 `--ip` 必须是本机真实公网 IPv4，写成不可达地址会让面板把它记为出口，被解锁机的解析全部指向死地址。
- 解锁机防火墙（如 ufw）需放行 TCP 443。
- 无法确认：服务端无法在没有浏览器的情况下取到视频格式（YouTube 对机房 IP 要求 PO token），整段视频播放请以浏览器实测为准。
