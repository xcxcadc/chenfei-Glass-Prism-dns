# chenfei Glass Prism DNS

This is the enhanced fork at [xcxcadc/chenfei-Glass-Prism-dns](https://github.com/xcxcadc/chenfei-Glass-Prism-dns). It adds a Simplified Chinese UI, custom service domains and categories, per-service proxy selection, IP configuration, unlock-link traffic accounting, account security, panel and release scripts `1.5.30`, and SNI transport `2.5.2`. See [ENHANCED_ZH.md](ENHANCED_ZH.md) and the [latest release](https://github.com/xcxcadc/chenfei-Glass-Prism-dns/releases).

Every Proxy/Client receives an independent official media-script worker. It runs the complete IPv4 menu `0` check, streams results as they arrive, and rejects reports from an old route generation. Official script results do not guarantee paid playback or account login.

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
- **Streaming Unlock Detection** - Display all official IPv4 menu `0` platform results, including Netflix and Disney+; unsupported services are explicitly marked as not covered
- **AI Services Unlock Detection** - Display official ChatGPT, Claude, and Google Gemini results verbatim; these checks do not verify account sign-in or paid features
- **Dual-Stack Node Management** - The panel still manages IPv4/IPv6 nodes, while managed unlock routes always use proxy IPv4, never auto-switch egress to IPv6, and suppress AAAA for selected services
- **Real-time Monitoring** - SSE-based live node status updates
- **Modern Console** - Table-based orchestration and overview cards with seven functional pages for services, nodes, target IPs, catalog sync, audit logs, alerts, and settings; responsive desktop/mobile layouts with light/dark themes and Chinese/English
- **Chinese Enhanced UI** - Simplified Chinese by default, with English and theme switching
- **Detailed Service Catalog** - Dynamically parses the latest active services from `stream.smartdns.list`; discontinued Crackle, Salto, and GYAO entries are excluded
- **Service Domains and Deletion** - Edit domains for any built-in or custom service, add or remove entries, restore catalog defaults, and permanently delete services from the table or configuration dialog. Deletion removes rules, target-IP routes, audit results, and persistently hides built-in services from future catalog syncs
- **Typed Split Rules** - The service editor supports domain suffixes, domain keywords, and IPv4/IPv6 `IP-CIDR` networks. Rule lists and Agent sync preserve `DOMAIN-SUFFIX`, `DOMAIN-KEYWORD`, and `IP-CIDR` types; built-in Grok includes the xAI/Twitter domains, the `twitter` keyword, and the six upstream CIDR ranges
- **Unified Service Search** - The service library and IP picker support exact and fuzzy matching across names, localized aliases, categories, service IDs, and domains while ignoring case and common separators
- **Flexible Categories** - Create custom categories, move any built-in or custom service between them, and restore the original category without changing service IDs, domains, or deployed routes
- **Explicit Per-Service Routing** - Bind every service on a managed target IP to a proxy explicitly; routes change only after a save and never follow Smart/Fallback or audit results automatically
- **Stable Agent Data Plane** - The installer pins upstream Agent `v1.2.1`, preventing the `v1.3` passive circuit breaker from replacing a fixed unlock VIP with public DNS after WAF responses, refused connections, or batch audits
- **Overlapping-Domain Linking** - Services sharing parent, child, or wildcard domains automatically use one proxy; the latest selection wins
- **Audits Never Change Routes** - Target audits update reports only; WAF responses, timeouts, and third-party detector instability never switch the proxy selected by the user
- **Five-Second Allowlist** - Proxies refresh managed target IPv4 addresses every five seconds, restrict IPv4 DNS 53 plus SNI 80/443, reject IPv6 DNS 53 only, and do not intercept IPv6 80/443 so co-hosted MTProxy or other IPv6 services keep working
- **Official Media Script Results** - Each Proxy/Client runs the IPv4 international-platform check and preserves official labels and raw `Yes`, `No`, region, and WAF values
- **Separate Route Diagnostics** - DNS/TLS/SNI and page probes remain available independently and never override media results; Agent self-checks are reference-only
- **IP Configuration Workflow** - Add a target IP, choose services and proxy agents, then create DNS nodes and overrides automatically
- **Automatic Route Application** - The enhancer persists per-node routes and generates dedicated dnsmasq rules; a 10-second guard atomically applies saved changes and checks the DNS listener, while a 300-second sample checks one representative domain per selected service
- **Restricted-Network Transport** - Targets synchronize and probe encrypted TCP SNI transport every 60 seconds; SSH still uses 15-second keepalives with three missed checks before failure, and a false-positive TCP path only rebuilds that tunnel
- **Persistent Live Checks** - Results refresh every second in batches, with six-hour automatic refresh, manual retests, configuration-generation updates, and progress available after reopening the window
- **Coexistence Watchdog** - Monitors and recovers only `prism-agent`; co-located MTProxy, XrayR, and V2bX services are left untouched
- **Service Brand Icons** - Service cards and the IP picker show an immediate local placeholder, replace it with the matching brand icon, and use versioned asset URLs so corrected icons appear immediately instead of reusing stale browser cache
- **Client Management Script** - Installs the DNS Agent, tests local DNS, and takes over, backs up, or restores system DNS
- **Traffic Accounting** - Per target IP, counts local Prism DNS UDP/TCP 53 plus TCP 80/443 exchanged with selected proxy IPs; whole-interface traffic is excluded

- **XrayR System DNS** - In system-DNS mode, if XrayR has custom DNS enabled, the installer backs up `/etc/XrayR/config.yml`, switches XrayR to `/etc/resolv.conf`, and restarts XrayR once; V2bX, MTProxy, Docker, and other service configurations are untouched
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

Official script menu `0` currently covers these platforms and region checks:

Node Management and IP Configs use `/usr/local/bin/prism-media-check` on the corresponding Proxy/Client. Its independent `prism-media-check.timer` / `prism-media-check.service` attempts to claim pending jobs every five seconds while idle, separately from transport synchronization and minute-based traffic reporting. After downloading and validating the official script, it runs the equivalent of `bash check.unlock.media -M4 -R0 -Een`, where `check.unlock.media` is the downloaded local script: interactive menu `0`, international platforms only, IPv4 only, with English official labels.

Results are parsed by official label and preserve raw `Yes` / `No`, region, WAF, and other values. Missing platform coverage is shown as not covered. DNS A mapping, AAAA suppression, TLS/SNI, and page probes remain separate route diagnostics and never replace media results. The old full/four-stage audit and Agent self-checks no longer provide final media conclusions. These are **official script results, not a guarantee of paid playback, account login, or all application features**.

The worker checks output every second, reports new results in batches, and the UI refreshes every second without waiting for completion. The full script runs without a 30-second hard stop. A 15-minute safety limit handles abnormal runs; exceeded limits, script failures, or no valid results produce an explicit `failed` state. No heartbeat for 90 seconds reports a lost worker and a failed job. A `completed` job means the script finished, not that every platform returned `Yes`.

Jobs, progress, results, and history persist in `dataDir/media-checks.json` (default `/var/lib/prism-enhancer/media-checks.json`). Closing the window does not cancel the worker; reopening reads the saved job. Checks refresh automatically every six hours, on manual retest, and after configuration-generation changes. Reports from an old `generation` are rejected so delayed results cannot overwrite the current route. Checks never change the selected proxy.

**Streaming Services**
- Netflix, Disney+, DAZN, Amazon Prime Video, TVBAnywhere+
- YouTube Premium, Spotify Registration, iQyi Oversea Region
- YouTube/Netflix CDN and Apple, Bing, Google Play, Steam region information

**AI Services**
- OpenAI (ChatGPT)
- Anthropic (Claude)
- Google (Gemini)

HBO Max, Hulu, Copilot, and other catalog services can have routing rules, but menu `0` does not cover them and cannot certify their availability.

## Installation

### One-Click Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/enhanced_install.sh | sudo bash
```

Custom ports:

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/enhanced_install.sh \
  | sudo PRISM_PORT=8080 PRISM_CORE_PORT=18080 bash
```

The script installs or upgrades the upstream Controller and this fork's enhancer, backing up `data.db` before upgrades.

Panel/scripts `1.5.30`, transport `2.5.2`, and the default upstream compatibility Agent `v1.2.1` have independent version numbers; Agent `v1.3` is not the default. See the [version matrix](VERSION_MATRIX_ZH.md).

After installation:
- Web UI: `http://YOUR_IP:PORT`
- Username: `admin`
- Password: Displayed after installation
- Custom service data: `/var/lib/prism-enhancer/custom-services.json`

### Fresh Installation and Data Sanitization

The repository contains only program code, installer scripts, and frontend assets. It does not contain runtime databases, node secrets, credentials, IP routes, traffic records, or data from the current server. A new-server install creates empty stores and generates a new JWT secret and initial password; it never copies this panel's nodes or service configuration. After login, the icon can be replaced with a transparent PNG from `Settings -> Site settings`; the runtime file is stored at `/var/lib/prism-enhancer/branding-icon.png`.

To explicitly perform a sanitized fresh install on a host that previously ran Prism, first confirm that its old panel data is no longer needed, then run:

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/enhanced_install.sh \\
  | sudo env PRISM_FRESH_INSTALL=1 PRISM_CONFIRM_FRESH=YES bash
```

This mode removes only Prism's `/opt/prism/data.db`, `/opt/prism/.env`, initial-password file, and `/var/lib/prism-enhancer/`. It does not remove or modify Docker, XrayR, V2bX, MTProxy, or other services. Restoring an existing panel is always explicit through [MIGRATION_ZH.md](MIGRATION_ZH.md); the installer never imports old data automatically.

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
curl -sL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/agent_install.sh | bash -s -- --master <Controller_URL> --secret <Node_Secret>
```

### IP Client Tool

First add the target IP in the Web UI's IP Configs view, select services and proxy agents, and save. The UI generates a dedicated command containing the panel URL and enrollment token. The generic interactive tool is also available:

```bash
wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/prismdns.sh | sudo bash
```

Use the dedicated command for first installation and node upgrades. Routine additions, removals, and proxy changes apply after a panel save. The installer backs up conflicting legacy `dnsmasq`/`sniproxy` services, installs dedicated dnsmasq on `127.0.0.1:5353`, and redirects local DNS requests to its generated IPv4 and AAAA-suppression rules. Stable installs verify and pin upstream Agent `v1.2.1`; reinstall and uninstall unlock it automatically. The guard checks configuration and DNS every 10 seconds, samples representative domains every 300 seconds, and checks all routed domains on changes and the 30-minute route-health refresh. Installation blocks only on incorrect DNS mapping, while media jobs run independently. Route updates restart the dedicated resolver first; Agent restarts are a recovery fallback.

Transport `2.5.2` defaults to authorized direct TCP SNI egress; restricted networks can explicitly select encrypted mode. UDP/443 to selected proxies is rejected to encourage TCP/TLS fallback. Proxy AUTH nftables rules are validated and replaced atomically in one transaction every five seconds, protecting Prism IPv4 DNS/SNI and IPv6 DNS 53 while leaving IPv6 80/443 available. Minute-based traffic reports do not wait for media checks, and clearing traffic preserves the sample baseline. Installing the media worker disables the old audit timer and stops its service.

IPv4 preference, AAAA suppression, overlapping-domain linking, independent route diagnostics, five-second authorization, traffic accounting, automatic apply, and the media worker are defaults for all future Proxy/Client installations, with no fixed host addresses. The worker is installed with transport and removed with transport/Agent, including its executable, timer, and service. All installation and uninstall entry points use `v1.5.30`; see [uninstall instructions](UNINSTALL_ZH.md).

### Upgrade Existing Nodes

**Replacing only the enhancer binary does not refresh remote nodes.** Rerun each Proxy's `v1.5.30` Agent installation command and each Client's current dedicated IP Configs command to refresh runtime, transport `2.5.2`, and the media worker. Clients with valid saved configuration can instead run the current runtime refresh script:

```bash
curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/v1.5.30/prismdns.sh | sudo bash -s -- --refresh-runtime
```

The installed `prism_transport.sh --sync` synchronizes configuration only; it does not download newer runtime/transport scripts or install a missing worker. Check `systemctl is-enabled prism-media-check.timer` on each node, then trigger a panel retest and verify live results. An idle oneshot service need not remain active.

For panel address changes, updated transport honors an explicit `--master`. If its old endpoint fails, it can recover the address from the running Agent's `--master`, validate the response, and persist the new endpoint for subsequent transport and worker requests. Preserve the entire enhancer `dataDir`, including `media-checks.json` progress/history, during migration; see [migration instructions](MIGRATION_ZH.md).

IP configurations, node secrets, enrollment tokens, and traffic baselines are stored in `/var/lib/prism-enhancer/ip-configs.json` with mode `0600`. Each target also persists DNS RX/TX and transport RX/TX independently in `/var/lib/prismdns/traffic-cumulative.json`; rebuilding one nftables table cannot discard the other component's history. The state is isolated by enrollment token. Do not publish a dedicated command or enrollment token.

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
