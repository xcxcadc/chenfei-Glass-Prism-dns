# chenfei Glass Prism DNS

This is the enhanced fork at [xcxcadc/chenfei-Glass-Prism-dns](https://github.com/xcxcadc/chenfei-Glass-Prism-dns). It adds a Simplified Chinese UI, custom service domains and categories, per-service proxy selection, IP configuration, unlock-link traffic accounting, account security, and client tool `1.4.0`. See [ENHANCED_ZH.md](ENHANCED_ZH.md) and the [latest release](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases/latest).

Prism-Gateway is a lightweight, non-intrusive DNS-based traffic routing management panel. It supports streaming unlock and smart AI services unlock detection. Features a beautiful Liquid Glass-inspired UI.

[中文](README.md) | English

## 🌐 Project Notes

The upstream demo at [prism.ciii.club](https://prism.ciii.club) only demonstrates the original Controller. Deploy this fork to use the enhanced UI and routing features.

## 💬 Join the Community

**Telegram Group**: [https://t.me/Prism_Gateway](https://t.me/Prism_Gateway)

## Features

### Core Features

- **Smart DNS Routing** - Route traffic through different Proxy Agents based on domain rules
- **External Ruleset Support** - Import external ruleset files for quick configuration of common services
- **Streaming Unlock Detection** - Auto-detect unlock status for Netflix, Disney+, HBO Max, and 20+ services
- **AI Services Unlock Detection** - Auto-detect availability of OpenAI, Claude, Gemini, Copilot and other AI services
- **Dual-Stack IPv4/IPv6** - Full support for both protocols
- **Real-time Monitoring** - SSE-based live node status updates
- **Modern UI** - Three-column route planning workspace with desktop/mobile layouts, light/dark themes, Chinese and English
- **Chinese Enhanced UI** - Simplified Chinese by default, with English and theme switching
- **Detailed Service Catalog** - Dynamically parses `stream.smartdns.list` into 178 active services; discontinued Crackle, Salto, and GYAO entries are excluded
- **Custom Services** - Add, edit, and delete arbitrary service names and domain lists in the UI; both `example.com` and `*.example.com` are preserved and compiled into deduplicated `DOMAIN-SUFFIX` rules
- **Unified Service Search** - The service library and IP picker support exact and fuzzy matching across names, localized aliases, categories, service IDs, and domains while ignoring case and common separators
- **Flexible Categories** - Create custom categories, move any built-in or custom service between them, and restore the original category without changing service IDs, domains, or deployed routes
- **Per-Service Routing** - Bind each service on each DNS client to any proxy agent, or restore Smart/Fallback selection
- **Two-Layer Unlock Audit** - Shows proxy-side Agent UnlockTests and reruns the same detector on each target IP, avoiding false confidence from proxy-only checks
- **Target Compatibility First** - Node cards, service badges, and test dialogs prioritize per-target audits; Agent self-checks are reference-only and inconsistent three-run results are marked `UNSTABLE`
- **IP Configuration Workflow** - Add a target IP, choose services and proxy agents, then create DNS nodes and overrides automatically
- **Automatic Route Application** - The enhancer restores persistent per-node routes on every Agent sync; a generic 10-second guard also verifies real DNS routes and safely restarts a stalled or stale Agent with rate limiting
- **Restricted-Network Transport** - Targets can use encrypted TCP SNI transport to avoid cross-border UDP 53 loss, poisoning, and latency while preserving per-service SG/VN selection
- **Multi-Domain Audit Fallback** - Target DNS/HTTPS audits try multiple representative domains so a throttled, retired, or non-Web suffix cannot create a false failure
- **Coexistence Watchdog** - Monitors and recovers only `prism-agent`; co-located MTProxy, XrayR, and V2bX services are left untouched
- **Service Brand Icons** - Service cards and the IP picker load matching site icons, with deterministic local fallbacks
- **Client Management Script** - Installs the DNS Agent, tests local DNS, and takes over, backs up, or restores system DNS
- **Traffic Accounting** - Per target IP, counts local Prism DNS UDP/TCP 53 plus TCP 80/443 exchanged with selected proxy IPs; whole-interface traffic is excluded
- **Account Security** - Click the username in the top bar to change the administrator username and password after verifying the current credentials

### Routing Modes

This project supports flexible node grouping and routing strategies:

#### Group (Node Grouping)

Combine multiple Proxy Agents into a group, each node can have a priority (higher value = higher priority). Two selection strategies are supported within a group:

| Strategy | Description |
|----------|-------------|
| **Smart** | Intelligent Selection - Automatically choose the node with best unlock status within the group |
| **Fallback** | Failover - Try nodes from highest to lowest priority, switch to next when current fails |

#### Priority Rules

Priority is sorted by value from high to low:

```
Priority 100 (Highest) → Priority 50 → Priority 10 → Priority 1 (Lowest)
```

#### How It Works

**Fallback Mode**: Try nodes from highest to lowest priority

```mermaid
flowchart LR
    A[Node A<br/>Priority 100] -->|A fails| B[Node B<br/>Priority 50]
    B -->|B fails| C[Node C<br/>Priority 10]
```

**Smart Mode**: Automatically select the node with best unlock status

```mermaid
flowchart LR
    subgraph Group[Node Group]
        A[Node A<br/>Unlocked ✓]
        B[Node B<br/>Locked ✗]
        C[Node C<br/>Unlocked ✓]
    end
    Group -->|Smart Select| D[Choose unlocked node]
```

**Example Scenarios**:
- **Fallback Mode**: Priority 100 > 50 > 10, prefer highest priority node, degrade sequentially on failure
- **Smart Mode**: Auto-detect unlock status of all nodes in group, select the one that is unlocked

### Unlock Detection

Automatically detect unlock status for the following services:

Node Management shows the proxy Agent's own UnlockTests output. IP Configs reruns the same detector used by the referenced `dns_unlock.sh` on the actual target server and stores the raw result per selected service. Service pages prefer this target-side verdict and run TLS against one known live representative hostname, instead of incorrectly treating routing-only suffix roots as failed websites. Target audits run after route changes and refresh periodically.

**Streaming Services**
- Netflix, Disney+, HBO Max, Amazon Prime Video
- Hulu, Paramount+, Peacock, Discovery+
- YouTube Premium, Spotify, Apple TV+
- BBC iPlayer, ITV, Channel 4, Channel 5
- And more...

**AI Services**
- OpenAI (ChatGPT)
- Anthropic (Claude)
- Google (Gemini)
- GitHub Copilot
- And more...

## Installation

### One-Click Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/enhanced_install.sh | sudo bash
```

Custom ports:

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/enhanced_install.sh \
  | sudo PRISM_PORT=8080 PRISM_CORE_PORT=18080 bash
```

The script installs or upgrades the upstream Controller and this fork's enhancer, backing up `data.db` before upgrades.

After installation:
- Web UI: `http://YOUR_IP:PORT`
- Username: `admin`
- Password: Displayed after installation
- Custom service data: `/var/lib/prism-enhancer/custom-services.json`

### Manual Installation

Controller and Agent binaries come from the [upstream releases](https://github.com/mslxi/Liquid-Glass-Prism-dns/releases). Enhancer binaries come from [this fork's releases](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases). The one-click installer is recommended.

```bash
# Download
wget https://github.com/mslxi/Liquid-Glass-Prism-dns/releases/latest/download/prism-controller-linux-amd64
chmod +x prism-controller-linux-amd64
mkdir -p /opt/prism
mv prism-controller-linux-amd64 /opt/prism/prism-controller

# Create environment file
echo "JWT_SECRET=$(openssl rand -hex 16)" > /opt/prism/.env

# Run
cd /opt/prism && ./prism-controller --host 0.0.0.0 --port 8080
```

## Agent Installation

Install Agent on node servers:

```bash
curl -sL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh | bash -s -- --master <Controller_URL> --secret <Node_Secret>
```

### IP Client Tool

First add the target IP in the Web UI's IP Configs view, select services and proxy agents, and save. The UI generates a dedicated command containing the panel URL and enrollment token. The generic interactive tool is also available:

```bash
wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/prismdns.sh | sudo bash
```

The dedicated command is only needed for the target's first installation. Later service additions, removals, and proxy switches apply automatically when saved in the panel. The generic tool installs/connects the DNS Client Agent, tests local DNS, applies permanent or temporary system DNS, and supports backup, restore, and status checks. The installer backs up and disables conflicting legacy `dnsmasq`/`sniproxy` services, but does not modify MTProxy, XrayR, or V2bX. Every new target receives a generic 10-second configuration hash guard that reads that machine's own panel URL and token. When routes change, it restarts only `prism-agent`, waits for port 53, and then commits the new hash so upstream hot-reload cannot leave stale DNS entries. Restricted-network targets use encrypted TCP SNI transport instead of unreliable cross-border UDP 53, and service audits fall back across multiple candidate domains. Initial installation requires every selected domain to resolve to a configured proxy. Third-party HTTPS probe failures remain warnings, so regional policy, rate limiting, or transient TLS failures cannot block system DNS activation. Dedicated nftables counters record full RX/TX for local Prism DNS UDP/TCP 53 plus TCP 80/443 exchanged with selected proxy IPs. The first report runs within 15 seconds and repeats every minute; automated UnlockTests traffic is removed from the counters after the audit.

The guard, traffic accounting, and automatic-apply logic are generated from each target's own configuration and contain no fixed host addresses, so future proxy and target machines inherit the same behavior by default.

IP configurations, node secrets, enrollment tokens, and traffic baselines are stored in `/var/lib/prism-enhancer/ip-configs.json` with mode `0600`. Do not publish a dedicated command or enrollment token.

### Agent Parameters

| Parameter | Description | Applicable Nodes |
|-----------|-------------|------------------|
| `--master` | Controller URL, e.g., `http://192.168.1.1:8080` | All nodes |
| `--secret` | Node secret generated when creating node in Controller | All nodes |
| `--smart` | Enable smart unlock detection mode | DNS Client only |
| `--beta` | Use beta version | All nodes |

## Architecture

```mermaid
flowchart TB
    Controller[Controller<br/>Web UI + API + Rule Engine + Unlock Detection]
    Controller -->|Push Rules / Status Report| DNS[DNS Client<br/>Edge Node<br/>Receive DNS]
    Controller -->|Push Rules / Status Report| Proxy1[Proxy Agent<br/>US Node<br/>Unlock Netflix]
    Controller -->|Push Rules / Status Report| Proxy2[Proxy Agent<br/>JP Node<br/>Unlock DMM]
```

| Component | Description |
|-----------|-------------|
| **Controller** | Central controller with Web UI, API, rule engine and unlock detection |
| **DNS Client** | Edge node that receives DNS queries, forwards to corresponding Proxy Agent based on rules |
| **Proxy Agent** | Exit node that forwards traffic to target servers, reports unlock status. Supports nested unlocking: if a VPS provider offers DNS unlock service, that VPS can serve as a Proxy Agent to provide proxy for other DNS Clients |

## Workflow

1. **Install Controller** - Install on central server
2. **Create Nodes** - Create DNS Client and Proxy Agent nodes in Web UI
3. **Install Agents** - Install Agent on each node server
4. **Configure Rules** - Create DNS rules, select routing mode and target nodes
5. **Start Using** - Point client DNS to DNS Client node

## Service Management

```bash
# Check status
sudo systemctl status prism-controller

# Restart service
sudo systemctl restart prism-controller

# View logs
journalctl -u prism-controller -f
```

## 📸 Screenshots

![Dashboard](screenshots/1-dashboard.webp)
![Nodes](screenshots/2-nodes.webp)
![Rules](screenshots/3-rules.webp)
![Unlock](screenshots/4-unlock.webp)
