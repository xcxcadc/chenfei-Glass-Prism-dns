const translations = {
  zh: {
    title: "Prism DNS 中文增强版", online: "控制器已连接", loginTitle: "登录 Prism Gateway", loginHint: "使用 Controller 管理员账号登录",
    username: "用户名", password: "密码", login: "登录", loggingIn: "正在登录...", services: "服务编排", nodes: "节点管理", ipConfigs: "IP 配置",
    search: "搜索服务或域名", allCategories: "全部分类", dnsClient: "DNS 节点", selectDNS: "请选择 DNS 节点", addService: "新增服务",
    addNode: "新增节点", refresh: "刷新", logout: "退出", totalServices: "服务数量", configured: "已配置", proxyNodes: "解锁机",
    customServices: "自定义服务", notConfigured: "未配置", auto: "自动选择", manual: "手动", open: "配置", domains: "域名",
    serviceConfig: "配置服务", currentRoute: "当前路由", targetServer: "目标解锁机", assign: "保存并切换", reset: "恢复自动",
    connectivity: "DNS 解锁检测", testing: "正在检测...", triggerUnlock: "运行解锁检测", customEdit: "编辑服务", customDelete: "删除服务",
    serviceName: "服务名称", category: "分类", domainList: "域名列表", save: "保存", cancel: "取消", deleteConfirm: "确认删除这个自定义服务？",
    nodeName: "节点名称", role: "节点类型", proxy: "解锁机", dns: "DNS 客户端", address: "连接 IP / DNS 地址", country: "地区",
    group: "分组", priority: "优先级", smartMode: "智能模式", standardMode: "标准模式", createNode: "创建节点", editNode: "编辑节点", deleteNode: "删除节点", installCommand: "Agent 安装命令", showInstallCommand: "安装命令",
    nextStep: "下一步", back: "上一步", copy: "复制命令", copied: "已复制", close: "关闭", nodeHint: "支持字母、数字和空格；点、短横线、下划线会自动转换为空格。", groupHint: "支持字母、数字、空格和逗号；其他分隔符会自动转换为空格。",
    nameRequired: "请输入节点名称", nameInvalid: "节点名称只能包含字母、数字和空格", nameTooLong: "节点名称不能超过 64 个字符", groupInvalid: "分组只能包含字母、数字、空格和逗号", groupTooLong: "分组不能超过 64 个字符", addressInvalid: "连接地址必须是有效的 IPv4 或 IPv6 地址", priorityInvalid: "优先级必须在 1 到 100 之间", deleteNodeConfirm: "确定删除这个节点吗？此操作不可撤销。",
    noDNS: "尚未创建 DNS 节点", noProxy: "尚未添加解锁机", noServices: "没有匹配的服务", sourceUpdated: "名单更新时间",
    supported: "DNS 检测可用", unavailable: "DNS 检测不可用", unknown: "未检测", active: "正在使用", manualOffline: "手动目标离线", ruleCreated: "服务规则已创建",
    saved: "保存成功", deleted: "删除成功", switched: "切换成功", resetDone: "已恢复自动选择", testDone: "测试完成", failed: "操作失败",
    catalogRefresh: "同步域名名单", sourceList: "细化域名库", theme: "切换主题", language: "English", originalStatus: "DNS 解锁检测",
    addIP: "添加 IP", editIP: "配置服务", targetIP: "目标 IP", note: "备注", defaultProxy: "默认解锁机", chooseServices: "选择服务", selectedServices: "已选服务", clientScript: "客户端脚本", runScriptHint: "在目标服务器以 root 身份执行，脚本会安装 DNS Agent、测试本机 DNS，并可接管或恢复系统 DNS。", noIPConfigs: "尚未添加 IP 配置", ipDeleteConfirm: "删除该 IP 配置及其 DNS 节点？", saveConfig: "保存配置", scriptCommand: "一键配置命令",
    totalTraffic: "全体解锁流量", clientTraffic: "解锁链路流量", clearTraffic: "清零流量", clearAllTraffic: "全部清零", trafficHint: "按目标 IP 独立统计本机 Prism DNS 的 UDP/TCP 53，以及到已选解锁机 TCP 80/443 的 RX/TX；不统计整机网卡流量。", trafficUpdated: "流量更新时间",
    clientState: "数据面状态", ready: "READY", degraded: "ERROR", pending: "PENDING", stale: "STALE", healthRoutes: "路由探针",
    actualAudit: "目标机实测", auditPending: "待实测", targetAvailable: "目标机可用", targetProblem: "目标机异常", targetCompatibility: "目标机兼容性",
    actualAvailable: "实测可用", actualProblem: "异常或波动", referenceOnly: "仅节点自检", agentReference: "节点自检（仅参考）", noTargetAudit: "尚无目标机实测",
    accountSettings: "账户安全", accountHint: "验证旧账号后修改管理员用户名和密码。", oldUsername: "旧用户名", oldPassword: "旧密码", newUsername: "新用户名", newPassword: "新密码", confirmPassword: "确认新密码", updateAccount: "更新账户", credentialsInvalid: "旧用户名或旧密码不正确", passwordMismatch: "两次输入的新密码不一致", accountUpdated: "账户已更新，请使用新账号重新登录",
    siteSettings: "站点设置", siteSettingsHint: "自定义左上角产品名称、说明文字和浏览器标签标题。", siteName: "网页名称", browserTitle: "页面标签名称", siteTagline: "网页说明", saveBranding: "保存站点设置", brandingUpdated: "站点名称已更新",
    unlockResults: "解锁检测结果", availableServices: "可用服务", unavailableServices: "不可用服务", unlockSourceHint: "节点自检仅供参考；已配置服务会在同一行列出每台目标 IP 的三次聚合实测。", checkAgain: "重新检测", waitingResults: "正在等待节点返回检测结果...",
  },
  en: {
    title: "Prism DNS Enhanced", online: "Controller connected", loginTitle: "Sign in to Prism Gateway", loginHint: "Use your Controller administrator account",
    username: "Username", password: "Password", login: "Sign in", loggingIn: "Signing in...", services: "Service Routing", nodes: "Nodes", ipConfigs: "IP Configs",
    search: "Search services or domains", allCategories: "All categories", dnsClient: "DNS client", selectDNS: "Select a DNS client", addService: "Add service",
    addNode: "Add node", refresh: "Refresh", logout: "Sign out", totalServices: "Services", configured: "Configured", proxyNodes: "Proxy nodes",
    customServices: "Custom services", notConfigured: "Not configured", auto: "Automatic", manual: "Manual", open: "Configure", domains: "Domains",
    serviceConfig: "Configure service", currentRoute: "Current route", targetServer: "Target proxy", assign: "Save and switch", reset: "Reset to auto",
    connectivity: "DNS unlock check", testing: "Testing...", triggerUnlock: "Run unlock check", customEdit: "Edit service", customDelete: "Delete service",
    serviceName: "Service name", category: "Category", domainList: "Domain list", save: "Save", cancel: "Cancel", deleteConfirm: "Delete this custom service?",
    nodeName: "Node name", role: "Node role", proxy: "Proxy agent", dns: "DNS client", address: "Connect IP / DNS address", country: "Region",
    group: "Group", priority: "Priority", smartMode: "Smart mode", standardMode: "Standard mode", createNode: "Create node", editNode: "Edit node", deleteNode: "Delete node", installCommand: "Agent install command", showInstallCommand: "Install command",
    nextStep: "Next step", back: "Back", copy: "Copy command", copied: "Copied", close: "Close", nodeHint: "Use letters, numbers, and spaces; dots, hyphens, and underscores become spaces.", groupHint: "Use letters, numbers, spaces, and commas; other separators become spaces.",
    nameRequired: "Node name is required", nameInvalid: "Node name may contain letters, numbers, and spaces only", nameTooLong: "Node name must be 64 characters or fewer", groupInvalid: "Group may contain letters, numbers, spaces, and commas only", groupTooLong: "Group must be 64 characters or fewer", addressInvalid: "Connect address must be a valid IPv4 or IPv6 address", priorityInvalid: "Priority must be between 1 and 100", deleteNodeConfirm: "Delete this node? This action cannot be undone.",
    noDNS: "No DNS client has been created", noProxy: "No proxy agent has been added", noServices: "No matching services", sourceUpdated: "Catalog updated",
    supported: "DNS check passed", unavailable: "DNS check failed", unknown: "Not checked", active: "Active", manualOffline: "Manual target offline", ruleCreated: "Service rule created",
    saved: "Saved", deleted: "Deleted", switched: "Switched", resetDone: "Automatic selection restored", testDone: "Test completed", failed: "Operation failed",
    catalogRefresh: "Sync domain catalog", sourceList: "Detailed domain catalog", theme: "Toggle theme", language: "简体中文", originalStatus: "DNS unlock check",
    addIP: "Add IP", editIP: "Configure services", targetIP: "Target IP", note: "Note", defaultProxy: "Default proxy", chooseServices: "Choose services", selectedServices: "Selected services", clientScript: "Client script", runScriptHint: "Run as root on the target server. The script installs the DNS Agent, tests local DNS, and can take over or restore system DNS.", noIPConfigs: "No IP configuration yet", ipDeleteConfirm: "Delete this IP configuration and its DNS node?", saveConfig: "Save configuration", scriptCommand: "One-click command",
    totalTraffic: "Total unlock traffic", clientTraffic: "Unlock link traffic", clearTraffic: "Clear traffic", clearAllTraffic: "Clear all", trafficHint: "Each target IP counts local Prism DNS UDP/TCP 53 plus TCP 80/443 to selected unlock proxies. Whole-interface traffic is excluded.", trafficUpdated: "Traffic updated",
    clientState: "Data-plane status", ready: "READY", degraded: "ERROR", pending: "PENDING", stale: "STALE", healthRoutes: "Route probes",
    actualAudit: "Target audit", auditPending: "Not audited", targetAvailable: "Target passed", targetProblem: "Target issue", targetCompatibility: "Target compatibility",
    actualAvailable: "Verified passed", actualProblem: "Failed or unstable", referenceOnly: "Agent only", agentReference: "Agent self-check (reference only)", noTargetAudit: "No target audit yet",
    accountSettings: "Account security", accountHint: "Verify the current credentials before changing the administrator username and password.", oldUsername: "Current username", oldPassword: "Current password", newUsername: "New username", newPassword: "New password", confirmPassword: "Confirm new password", updateAccount: "Update account", credentialsInvalid: "Current username or password is incorrect", passwordMismatch: "The new passwords do not match", accountUpdated: "Account updated. Sign in with the new credentials.",
    siteSettings: "Site settings", siteSettingsHint: "Customize the product name, supporting text, and browser tab title.", siteName: "Website name", browserTitle: "Browser tab title", siteTagline: "Website description", saveBranding: "Save site settings", brandingUpdated: "Site branding updated",
    unlockResults: "Unlock check results", availableServices: "Available services", unavailableServices: "Unavailable services", unlockSourceHint: "Agent self-checks are reference-only. Configured services list each target IP's aggregated three-run audit on the same row.", checkAgain: "Check again", waitingResults: "Waiting for the node to return results...",
  }
};

const commonNames = {
  "ChatGPT / OpenAI": "ChatGPT / OpenAI", "Disney+": "Disney+", "Prime Video": "Prime Video", "Apple TV+": "Apple TV+",
  "Bilibili": "哔哩哔哩", "YouTube": "YouTube", "Gemini": "Gemini", "Claude": "Claude", "Netflix": "Netflix",
  "TikTok": "TikTok", "Spotify": "Spotify", "Meta AI": "Meta AI", "Suno": "Suno"
};

const categoryNames = {
  "AI Platform": "AI 服务", "Canada Media": "加拿大媒体", "China Media": "中国大陆媒体", "Europe Media": "欧洲媒体",
  "Global Plaform": "全球服务", "Global Platform": "全球服务", "Hong Kong Media": "香港媒体", "Indian Media": "印度媒体",
  "Japan Media": "日本媒体", "Korean Media": "韩国媒体", "North America Media": "北美媒体", "Oceania Media": "大洋洲媒体",
  "Others": "其他服务", "South America Media": "南美媒体", "Southeast Asia Media": "东南亚媒体", "Sports Media": "体育服务", "Taiwan Media": "台湾媒体"
};

const state = {
  token: localStorage.getItem("prism_token") || "",
  user: JSON.parse(localStorage.getItem("prism_user") || "{}"),
  lang: localStorage.getItem("enhancer_lang") || "zh",
  theme: localStorage.getItem("prism_theme_v2") || "light",
  tab: "services", nodes: [], rules: [], catalog: [], catalogMeta: {}, ipConfigs: [], dnsNodeId: localStorage.getItem("enhancer_dns") || "",
  overrides: {}, activeSelections: {}, search: "", category: "", loading: false, modal: null, testResults: null, scrollPositions: {},
  branding: {site_name:"", browser_title:"", site_tagline:""}
};

function t(key) { return translations[state.lang][key] || key; }
function escapeHTML(value = "") { return String(value).replace(/[&<>'"]/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"}[char])); }
function siteName() { return String(state.branding.site_name || "Prism DNS"); }
function browserTitle() { return String(state.branding.browser_title || t("title")); }
function siteTagline() { return String(state.branding.site_tagline || (state.lang === "zh" ? "全局解锁编排" : "Global orchestration")); }

const scrollSelectors = [".container", ".service-grid", ".selected-service-list", ".route-node-list"];
function captureScrollPositions() {
  const container = document.querySelector(".container[data-view]");
  const view = container?.dataset.view;
  if (!view) return;
  const positions = {...(state.scrollPositions[view] || {})};
  scrollSelectors.forEach(selector => {
    const element = document.querySelector(selector);
    if (element) positions[selector] = {top:element.scrollTop, left:element.scrollLeft};
  });
  state.scrollPositions[view] = positions;
}
function restoreScrollPositions(view) {
  const positions = state.scrollPositions[view];
  if (!positions) return;
  scrollSelectors.forEach(selector => {
    const position = positions[selector];
    const element = document.querySelector(selector);
    if (element && position) {
      element.scrollTop = position.top;
      element.scrollLeft = position.left;
    }
  });
}
function nodeID(value) { return value == null ? "" : String(value); }
function selectedDNS() { return state.nodes.find(node => nodeID(node.id) === nodeID(state.dnsNodeId)); }
function proxyNodes() { return state.nodes.filter(node => node.role === "proxy"); }
function dnsNodes() { return state.nodes.filter(node => node.role === "dns"); }
function parseUnlock(node) {
  if (!node) return {};
  if (node.unlock_data && typeof node.unlock_data === "object") return node.unlock_data;
  try { return JSON.parse(node.unlock_json || "{}"); } catch { return {}; }
}
function unlockEntries(node) {
  return Object.entries(parseUnlock(node)).filter(([, value]) => String(value || "").trim()).sort(([left], [right]) => left.localeCompare(right));
}
function unlockPassed(value) { return /^YES\b/i.test(String(value || "").trim()); }
function serviceDetectorKey(service) {
  const value = `${service.name} ${service.domains.join(" ")}`.toLowerCase();
  const tests = [
    ["OpenAI", ["openai", "chatgpt"]], ["Gemini", ["gemini", "bard"]], ["Claude", ["claude", "anthropic"]],
    ["Copilot", ["copilot", "github.dev"]], ["Perplexity", ["perplexity"]], ["Meta AI", ["meta.ai", "llama"]], ["Suno", ["suno"]],
    ["Netflix", ["netflix"]], ["Disney+", ["disney", "bamgrid"]], ["YouTube", ["youtube", "googlevideo"]],
    ["HBO Max", ["hbomax", "hbo.com", "max.com"]], ["Prime Video", ["primevideo", "amazonvideo"]], ["Hulu", ["hulu"]],
    ["Apple TV+", ["tv.apple", "appletv"]], ["Paramount+", ["paramount"]], ["Peacock", ["peacock"]],
    ["Crunchyroll", ["crunchyroll"]], ["DAZN", ["dazn"]], ["Bilibili", ["bilibili"]], ["Spotify", ["spotify"]], ["TikTok", ["tiktok"]]
  ];
  const match = tests.find(([, keywords]) => keywords.some(keyword => value.includes(keyword)));
  return match ? match[0] : "";
}
function serviceRule(service) {
  return state.rules.find(rule => String(rule.value || "").includes(`/enhancer/rules/${service.id}.list`)) ||
    state.rules.find(rule => String(rule.name || "").toLowerCase() === `stream · ${service.name}`.toLowerCase());
}
function normalizeOverride(value) {
  if (value == null) return "";
  if (typeof value === "object") return nodeID(value.proxy_node_id ?? value.node_id ?? value.id);
  return nodeID(value);
}
function overrideFor(rule) {
  if (!rule) return "";
  if (Array.isArray(state.overrides)) {
    const item = state.overrides.find(entry => nodeID(entry.rule_id) === nodeID(rule.id));
    return normalizeOverride(item);
  }
  return normalizeOverride(state.overrides[nodeID(rule.id)] ?? state.overrides[rule.id]);
}
function activeFor(rule) {
  if (!rule) return null;
  return state.activeSelections[rule.value] || state.activeSelections[String(rule.value || "").toLowerCase()] || null;
}
function activeNodeFor(rule) {
  const override = overrideFor(rule);
  if (override) return state.nodes.find(node => nodeID(node.id) === override) || null;
  const active = activeFor(rule);
  if (active && typeof active === "object" && active.node_id) return state.nodes.find(node => nodeID(node.id) === nodeID(active.node_id)) || null;
  const ip = typeof active === "string" ? active : active?.ip;
  if (ip) return proxyNodes().find(node => String(node.address || "").split(",").map(item => item.trim()).includes(ip)) || null;
  if (rule?.target_type === "node") return state.nodes.find(node => nodeID(node.id) === nodeID(rule.target_val)) || null;
  return null;
}
function serviceStatus(service, node) {
  const targetConfig = configForNode(selectedDNS());
  const targetResult = targetConfig?.service_results?.[service.id];
  if (targetResult && nodeID(targetConfig?.routes?.[service.id]) === nodeID(node?.id)) {
    const audit = auditResultState(targetResult);
    return {kind:audit.kind, label:audit.kind === "good" ? t("targetAvailable") : t("targetProblem"), raw:targetResult};
  }
  const key = serviceDetectorKey(service);
  const value = key ? parseUnlock(node)[key] || "" : "";
  if (value) return {kind:"warn", label:t("referenceOnly"), raw:value};
  return {kind:"", label:t("unknown"), raw:""};
}
function serviceIconHTML(service) {
  const name = displayServiceName(service);
  return `<span class="service-icon"><img src="/enhancer/icons/${encodeURIComponent(service.id)}.png" alt="" title="${escapeHTML(name)}" loading="lazy" decoding="async"></span>`;
}
function formatDate(value) { if (!value) return "-"; try { return new Date(value).toLocaleString(state.lang === "zh" ? "zh-CN" : "en-US"); } catch { return value; } }
function formatBytes(value) { const bytes = Number(value) || 0; if (bytes < 1024) return `${bytes} B`; const units = ["KB", "MB", "GB", "TB"]; let size = bytes; let unit = -1; do { size /= 1024; unit++; } while (size >= 1024 && unit < units.length - 1); return `${size >= 100 ? size.toFixed(0) : size.toFixed(2)} ${units[unit]}`; }
function displayCategory(value) { return state.lang === "zh" ? categoryNames[value] || value : value; }
function displayServiceName(service) { return state.lang === "zh" ? commonNames[service.name] || service.name : service.name; }
function serviceMatches(service, query) { return !query || `${service.name} ${displayServiceName(service)} ${service.category} ${service.domains.join(" ")}`.toLowerCase().includes(query); }
function isOnline(node) { if (!node.last_heartbeat) return false; return Date.now() - new Date(node.last_heartbeat).getTime() < 90000; }
function configForNode(node) { return state.ipConfigs.find(config => nodeID(config.dns_node_id) === nodeID(node?.id)); }
function clientState(config, node) {
  if (!isOnline(node)) return {kind:"bad", label:"OFFLINE", detail:state.lang === "zh" ? "控制通道离线" : "Control channel offline"};
  if (!config) return {kind:"warn", label:state.lang === "zh" ? "未纳管" : "UNMANAGED", detail:state.lang === "zh" ? "Agent 已在线，尚未配置解锁服务" : "Agent connected; unlock services are not configured"};
  if (!config.health_updated_at) return {kind:"warn", label:t("pending"), detail:state.lang === "zh" ? "等待客户端健康上报" : "Waiting for client health report"};
  if (Date.now() - new Date(config.health_updated_at).getTime() > 180000) return {kind:"warn", label:t("stale"), detail:state.lang === "zh" ? "健康上报已超时" : "Health report is stale"};
  const routes = `${Number(config.healthy_routes || 0)}/${Number(config.expected_routes || 0)}`;
  if (config.dns_ready && config.system_dns_ready && config.routes_ready) return {kind:"good", label:t("ready"), detail:`${t("healthRoutes")} ${routes}`};
  return {kind:"bad", label:t("degraded"), detail:config.health_message || `${t("healthRoutes")} ${routes}`};
}
function auditResultState(result) {
  const value = String(result || "").trim();
  if (!value) return {kind:"", label:t("auditPending")};
  if (/^YES\b|^PASS\b/i.test(value)) return {kind:"good", label:value};
  if (/UNSTABLE|Restricted|Partial|Banned|WAF|Error|Failed|N\/A/i.test(value)) return {kind:"warn", label:value};
  return {kind:"bad", label:value};
}

function targetAuditObservations(proxyId, service) {
  return state.ipConfigs
    .filter(config => nodeID(config?.routes?.[service.id]) === nodeID(proxyId))
    .map(config => ({ip:config.ip, result:String(config?.service_results?.[service.id] || "").trim()}));
}

function targetCompatibility(proxyId, service) {
  const observations = targetAuditObservations(proxyId, service);
  const measured = observations.filter(item => item.result);
  const passed = measured.filter(item => auditResultState(item.result).kind === "good").length;
  let kind = "warn";
  if (measured.length && passed === measured.length && measured.length === observations.length) kind = "good";
  else if (measured.length && passed === 0) kind = "bad";
  const label = observations.length ? `${t("actualAudit")} ${passed}/${observations.length}` : t("referenceOnly");
  return {observations, measured, passed, kind, label};
}

function proxyCompatibilitySummary(proxyId) {
  const services = state.catalog.filter(service => targetAuditObservations(proxyId, service).length);
  const results = services.map(service => targetCompatibility(proxyId, service));
  const passed = results.filter(result => result.kind === "good").length;
  const problem = results.filter(result => result.measured.length && result.kind !== "good").length;
  const pending = results.filter(result => !result.measured.length).length;
  return {passed, problem, pending, total:services.length};
}

function preferredServiceTestDomain(service) {
  const preferred = {
    "Apple TV+": ["tv.apple.com"],
    "Bilibili": ["bilibili.com"],
    "ChatGPT / OpenAI": ["chatgpt.com", "openai.com"],
    "Claude": ["claude.ai", "anthropic.com"],
    "Crunchyroll": ["crunchyroll.com"],
    "DAZN": ["dazn.com"],
    "Disney+": ["disneyplus.com", "bamgrid.com"],
    "Gemini": ["gemini.google.com", "bard.google.com"],
    "Google AI Studio": ["aistudio.google.com"],
    "HBO / Max": ["max.com", "hbomax.com"],
    "Microsoft Copilot Image Creator": ["copilot.microsoft.com"],
    "Netflix": ["netflix.com"],
    "Paramount+": ["paramountplus.com"],
    "Spotify": ["spotify.com"],
    "Suno": ["suno.com", "suno.ai"],
    "TikTok": ["tiktok.com"],
    "YouTube": ["youtube.com"],
  };
  return (preferred[service.name] || []).find(candidate =>
    service.domains.some(domain => candidate === domain || candidate.endsWith(`.${domain}`))
  ) || service.domains[0];
}

async function api(path, options = {}) {
  const headers = {...(options.headers || {})};
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  if (options.body && !(options.body instanceof FormData)) headers["Content-Type"] = "application/json";
  const response = await fetch(path, {...options, headers});
  if (response.status === 401) {
    logout(false);
    throw new Error(state.lang === "zh" ? "登录已失效" : "Session expired");
  }
  if (response.status === 204) return null;
  const type = response.headers.get("content-type") || "";
  const data = type.includes("application/json") ? await response.json() : await response.text();
  if (!response.ok) throw new Error(data?.error || data || `${response.status}`);
  return data;
}

async function loadBranding() {
  try {
    const response = await fetch("/enhancer/api/branding", {cache:"no-store"});
    if (!response.ok) return;
    const branding = await response.json();
    state.branding = branding && typeof branding === "object" ? branding : state.branding;
  } catch {}
}

async function login(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const button = event.currentTarget.querySelector("button[type=submit]");
  const error = document.querySelector(".form-error");
  button.disabled = true; button.textContent = t("loggingIn"); error.textContent = "";
  try {
    const response = await fetch("/api/login", {method:"POST", headers:{"Content-Type":"application/json"}, body:JSON.stringify({username:form.get("username"), password:form.get("password")})});
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || t("failed"));
    state.token = data.token; state.user = data.user || {username:data.username || ""};
    localStorage.setItem("prism_token", state.token); localStorage.setItem("prism_user", JSON.stringify(state.user));
    await loadAll();
  } catch (loginError) { error.textContent = loginError.message; }
  finally { button.disabled = false; button.textContent = t("login"); }
}

function logout(callAPI = true) {
  if (callAPI && state.token) fetch("/api/logout", {method:"POST", headers:{Authorization:`Bearer ${state.token}`}}).catch(() => {});
  if (eventSource) { eventSource.close(); eventSource = null; }
  clearTimeout(reloadTimer);
  state.token = ""; state.user = {}; state.nodes = []; state.rules = []; state.ipConfigs = [];
  localStorage.removeItem("prism_token"); localStorage.removeItem("prism_user");
  render();
}

async function loadAll(silent = false) {
  if (!state.token) return render();
  state.loading = !silent;
  if (!silent) render();
  try {
    const [nodes, rules, catalog, ipConfigs] = await Promise.all([api("/enhancer/api/nodes"), api("/api/rules"), api("/enhancer/api/catalog"), api("/enhancer/api/ip-configs")]);
    state.nodes = Array.isArray(nodes) ? nodes : [];
    state.rules = Array.isArray(rules) ? rules : [];
    state.catalog = catalog.services || [];
    state.catalogMeta = catalog;
    state.ipConfigs = Array.isArray(ipConfigs) ? ipConfigs : [];
    if (!dnsNodes().some(node => nodeID(node.id) === nodeID(state.dnsNodeId))) state.dnsNodeId = dnsNodes()[0]?.id || "";
    localStorage.setItem("enhancer_dns", nodeID(state.dnsNodeId));
    await loadRoutingState();
    connectSSE();
  } catch (error) { toast(error.message, "error"); }
  finally {
    state.loading = false;
    const preserveDraft = silent && ["service-form", "node-form", "ip-form"].includes(state.modal?.type);
    if (!preserveDraft) render();
  }
}

async function loadRoutingState() {
  if (!state.dnsNodeId) { state.overrides = {}; state.activeSelections = {}; return; }
  const [overrides, selections] = await Promise.all([
    api(`/api/overrides?dns_node_id=${encodeURIComponent(state.dnsNodeId)}`).catch(() => ({})),
    api(`/api/active_selections?dns_node_id=${encodeURIComponent(state.dnsNodeId)}`).catch(() => ({}))
  ]);
  state.overrides = overrides || {}; state.activeSelections = selections || {};
}

let eventSource;
let reloadTimer;
function scheduleLiveReload() {
  clearTimeout(reloadTimer);
  reloadTimer = setTimeout(() => {
    const active = document.activeElement;
    if (state.modal || active?.matches?.("input, textarea, select, [contenteditable=true]")) { scheduleLiveReload(); return; }
    loadAll(true);
  }, 700);
}
function connectSSE() {
  if (eventSource || !state.token) return;
  eventSource = new EventSource(`/api/sse?token=${encodeURIComponent(state.token)}`);
  eventSource.onmessage = scheduleLiveReload;
  eventSource.onerror = () => { eventSource.close(); eventSource = null; setTimeout(connectSSE, 3500); };
}

function render() {
  captureScrollPositions();
  document.documentElement.dataset.theme = state.theme === "light" ? "light" : "dark";
  document.documentElement.lang = state.lang === "zh" ? "zh-CN" : "en";
  document.title = browserTitle();
  const app = document.getElementById("app");
  if (!state.token) { app.innerHTML = loginHTML(); bindLogin(); return; }
  app.innerHTML = shellHTML();
  bindShell();
  renderModal();
  restoreScrollPositions(state.tab);
}

function loginHTML() {
  const loginTitle = state.lang === "zh" ? `登录 ${siteName()}` : `Sign in to ${siteName()}`;
  return `<main class="login-wrap"><section class="login-box panel"><h1>${escapeHTML(loginTitle)}</h1><p>${escapeHTML(t("loginHint"))}</p>
    <form class="form-stack" id="login-form"><div class="field"><label>${t("username")}</label><input class="input" name="username" autocomplete="username" required></div>
    <div class="field"><label>${t("password")}</label><input class="input" type="password" name="password" autocomplete="current-password" required></div>
    <div class="form-error"></div><button class="btn primary" type="submit">${t("login")}</button></form></section></main>`;
}
function bindLogin() { document.getElementById("login-form").addEventListener("submit", login); }

function toolbarActionsHTML() {
  if (state.tab === "services") return `<button class="btn" id="refresh-catalog">${t("catalogRefresh")}</button><button class="btn primary" id="add-service">${t("addService")}</button>`;
  if (state.tab === "nodes") return `<button class="btn primary" id="add-node">${t("addNode")}</button>`;
  return `<button class="btn primary" id="add-ip">${t("addIP")}</button>`;
}

function activeContentHTML() {
  if (state.tab === "services") return servicesHTML();
  if (state.tab === "nodes") return nodesHTML();
  return ipConfigsHTML();
}

function shellHTML() {
  const pageTitle = state.tab === "services" ? t("services") : state.tab === "nodes" ? t("nodes") : t("ipConfigs");
  const pageHint = state.lang === "zh"
    ? (state.tab === "services" ? "编排解锁服务、验证真实可用性并选择最佳节点" : state.tab === "nodes" ? "管理解锁机与被解锁机的在线状态" : "按目标 IP 配置服务、流量与自动下发")
    : (state.tab === "services" ? "Route services, verify availability, and choose the best node" : state.tab === "nodes" ? "Manage unlock and client node health" : "Configure services, traffic, and delivery by target IP");
  return `<div class="shell"><aside class="sidebar"><div class="sidebar-brand"><strong>${escapeHTML(siteName())}</strong><span>${escapeHTML(siteTagline())}</span></div>
    <nav class="tabs" aria-label="${escapeHTML(browserTitle())}"><button class="tab ${state.tab === "services" ? "active" : ""}" data-tab="services">${t("services")}</button><button class="tab ${state.tab === "nodes" ? "active" : ""}" data-tab="nodes">${t("nodes")}</button><button class="tab ${state.tab === "ips" ? "active" : ""}" data-tab="ips">${t("ipConfigs")}</button></nav>
    <div class="sidebar-footer"><button class="user-button account-settings-trigger" id="account-settings" title="${t("accountSettings")}"><strong>${escapeHTML(state.user.username || "admin")}</strong><span>${state.lang === "zh" ? "管理员账户" : "Administrator"}</span></button>
    <button class="btn small site-settings-trigger">${t("siteSettings")}</button><div class="sidebar-tools"><button class="btn small theme-toggle-trigger" id="theme-toggle">${state.theme === "light" ? (state.lang === "zh" ? "深色" : "Dark") : (state.lang === "zh" ? "浅色" : "Light")}</button><button class="btn small lang-toggle-trigger" id="lang-toggle">${t("language")}</button><button class="btn small danger logout-trigger" id="logout">${t("logout")}</button></div></div></aside>
    <section class="workspace"><header class="topbar"><div class="page-heading"><h1>${escapeHTML(pageTitle)}</h1><p>${escapeHTML(pageHint)}</p></div><div class="toolbar-right"><span class="live-state"><span class="status-dot"></span>${escapeHTML(t("online"))}</span>${toolbarActionsHTML()}<button class="btn" id="refresh">${t("refresh")}</button><div class="mobile-controls"><button class="btn small account-settings-trigger">${escapeHTML(state.user.username || "admin")}</button><button class="btn small site-settings-trigger">${t("siteSettings")}</button><button class="btn small theme-toggle-trigger">${state.theme === "light" ? (state.lang === "zh" ? "深色" : "Dark") : (state.lang === "zh" ? "浅色" : "Light")}</button><button class="btn small lang-toggle-trigger">${t("language")}</button><button class="btn small danger logout-trigger">${t("logout")}</button></div></div></header>
    <main class="container" data-view="${escapeHTML(state.tab)}">${state.loading ? `<div class="loading panel"><div class="spinner"></div></div>` : activeContentHTML()}</main></section></div>`;
}

function servicesHTML() {
  const dns = dnsNodes(); const proxies = proxyNodes();
  const categories = [...new Set(state.catalog.map(service => service.category))].sort((a,b) => a.localeCompare(b));
  const filtered = state.catalog.filter(service => !state.category || service.category === state.category);
  const configuredServices = state.catalog.filter(service => {
    const rule = serviceRule(service);
    return rule && activeNodeFor(rule);
  });
  const configured = configuredServices.length;
  const targetNode = selectedDNS();
  const targetConfig = configForNode(targetNode);
  const targetHealth = targetNode ? clientState(targetConfig, targetNode) : {kind:"warn", label:t("pending"), detail:t("selectDNS")};
  const targetIP = targetConfig?.ip || targetNode?.public_ip || targetNode?.address || "-";
  return `${dns.length === 0 ? `<div class="panel empty"><strong>${t("noDNS")}</strong><button class="btn primary" id="empty-add-node">＋ ${t("addNode")}</button></div>` : ""}
    <section class="orchestration-summary panel"><div class="summary-target"><span>${t("targetIP")}</span><strong>${escapeHTML(targetIP)}</strong><select class="select" id="dns-select"><option value="">${t("selectDNS")}</option>${dns.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(state.dnsNodeId) ? "selected" : ""}>${escapeHTML(node.name)}</option>`).join("")}</select></div>
    <div class="summary-health"><span>${state.lang === "zh" ? "路线健康" : "Route health"}</span><strong class="${targetHealth.kind}">${escapeHTML(targetHealth.label)}</strong><small title="${escapeHTML(targetHealth.detail)}">${escapeHTML(targetHealth.detail)}</small></div>
    <div class="summary-metric"><span>${t("selectedServices")}</span><strong>${configured} <small>/ ${state.catalog.length}</small></strong></div>
    <div class="summary-metric"><span>${state.lang === "zh" ? "可用节点" : "Available nodes"}</span><strong>${proxies.filter(isOnline).length} <small>/ ${proxies.length}</small></strong></div>
    <div class="summary-metric"><span>${t("customServices")}</span><strong>${state.catalog.filter(item => item.custom).length}</strong></div></section>
    <section class="orchestration-grid"><div class="panel orchestration-column service-library"><header class="section-head"><div><h2>${state.lang === "zh" ? "服务库" : "Service library"}</h2><p>${state.catalog.length} ${t("totalServices")}</p></div><button class="btn small" id="clear-filter">${state.lang === "zh" ? "重置" : "Reset"}</button></header>
    <div class="library-filters"><input class="input" id="service-search" value="${escapeHTML(state.search)}" placeholder="${t("search")}"><select class="select" id="category-filter"><option value="">${t("allCategories")}</option>${categories.map(category => `<option value="${escapeHTML(category)}" ${state.category === category ? "selected" : ""}>${escapeHTML(displayCategory(category))}</option>`).join("")}</select></div>
    ${filtered.length ? `<div class="service-grid">${filtered.map(serviceCardHTML).join("")}</div><div class="empty" id="service-filter-empty" hidden><strong>${t("noServices")}</strong></div>` : `<div class="empty"><strong>${t("noServices")}</strong></div>`}</div>
    <div class="panel orchestration-column selected-services"><header class="section-head"><div><h2>${t("selectedServices")}</h2><p>${configured} ${state.lang === "zh" ? "项已路由" : "routed"}</p></div></header><div class="selected-service-list">${configuredServices.length ? configuredServices.map(selectedServiceRowHTML).join("") : `<div class="empty compact"><strong>${t("notConfigured")}</strong><span>${state.lang === "zh" ? "从左侧选择服务开始配置" : "Choose a service from the library"}</span></div>`}</div></div>
    <div class="panel orchestration-column route-nodes"><header class="section-head"><div><h2>${state.lang === "zh" ? "节点选择与线路表现" : "Nodes and route health"}</h2><p>${state.lang === "zh" ? "依据真实目标机检测结果选择" : "Based on real target audits"}</p></div></header><div class="route-node-list">${proxies.length ? proxies.map(node => routeNodeCardHTML(node, configuredServices)).join("") : `<div class="empty compact"><strong>${t("noProxy")}</strong></div>`}</div><div class="route-note"><strong>${state.lang === "zh" ? "线路测试说明" : "Route testing"}</strong><span>${state.lang === "zh" ? "Agent 按真实 DNS、HTTPS 与播放链路定期复测；保存后自动下发并刷新 DNS。" : "Agents retest real DNS, HTTPS, and playback routes. Saves are delivered automatically."}</span></div></div></section>`;
}

function serviceCardHTML(service) {
  const rule = serviceRule(service); const node = activeNodeFor(rule); const status = serviceStatus(service, node); const manual = !!overrideFor(rule);
  const displayName = displayServiceName(service);
  return `<article class="service-card ${rule ? "configured" : ""} ${service.custom ? "custom" : ""}" data-service-id="${escapeHTML(service.id)}"><div class="service-row-main">${serviceIconHTML(service)}<div><div class="service-name" title="${escapeHTML(displayName)}">${escapeHTML(displayName)}</div><div class="service-category">${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")}</div></div></div>
    <div class="service-row-state"><span class="badge ${status.kind}" title="${escapeHTML(status.raw || status.label)}">${escapeHTML(status.kind ? status.label : (rule ? t("configured") : t("notConfigured")))}</span><button class="btn small service-open" data-service-id="${escapeHTML(service.id)}">${t("open")}</button></div></article>`;
}

function selectedServiceRowHTML(service) {
  const rule = serviceRule(service); const node = activeNodeFor(rule); const status = serviceStatus(service, node);
  return `<article class="selected-service-row" data-service-id="${escapeHTML(service.id)}">${serviceIconHTML(service)}<div class="selected-service-main"><strong>${escapeHTML(displayServiceName(service))}</strong><span>${escapeHTML(node?.name || t("auto"))}</span></div><span class="badge ${status.kind}" title="${escapeHTML(status.raw || status.label)}">${escapeHTML(status.kind ? status.label : t("configured"))}</span><button class="btn small service-open" data-service-id="${escapeHTML(service.id)}">${state.lang === "zh" ? "查看" : "View"}</button></article>`;
}

function routeNodeCardHTML(node, configuredServices) {
  const online = isOnline(node);
  const compatibility = proxyCompatibilitySummary(node.id);
  const assigned = configuredServices.filter(service => nodeID(activeNodeFor(serviceRule(service))?.id) === nodeID(node.id));
  const detail = compatibility.total ? `${compatibility.passed}/${compatibility.total} ${t("actualAudit")}` : t("noTargetAudit");
  return `<article class="route-node-card ${online ? "online" : "offline"}"><header><div><strong>${escapeHTML(node.name)}</strong><span>${escapeHTML(node.country || node.public_ip || node.address || "-")}</span></div><span class="badge ${online ? "good" : "bad"}">${online ? "ONLINE" : "OFFLINE"}</span></header><div class="route-node-metrics"><div><span>${state.lang === "zh" ? "目标实测" : "Target audits"}</span><strong>${escapeHTML(detail)}</strong></div><div><span>${t("selectedServices")}</span><strong>${assigned.length}</strong></div></div><div class="node-service-icons">${assigned.slice(0, 8).map(serviceIconHTML).join("")}${assigned.length > 8 ? `<span class="badge">+${assigned.length - 8}</span>` : ""}</div></article>`;
}

function nodesHTML() {
  if (!state.nodes.length) return `<div class="panel empty"><strong>${state.lang === "zh" ? "还没有节点" : "No nodes"}</strong><button class="btn primary" id="empty-add-node">＋ ${t("addNode")}</button></div>`;
  return `<section class="node-grid">${state.nodes.map(node => {
    const online = isOnline(node); const compatibility = proxyCompatibilitySummary(node.id);
    const proxyDetail = compatibility.total ? `${compatibility.passed}/${compatibility.total}` : t("noTargetAudit");
    const status = node.role === "dns" ? clientState(configForNode(node), node) : {kind:online ? "good" : "bad", label:online ? "ONLINE" : "OFFLINE", detail:proxyDetail};
    return `<article class="panel node-card"><div class="node-title"><div><h3>${escapeHTML(node.name)}</h3><div class="brand-meta"><span class="status-dot" style="background:${status.kind === "good" ? "var(--good)" : status.kind === "warn" ? "var(--warn)" : "var(--bad)"}"></span>${node.role === "proxy" ? t("proxy") : t("dns")}</div></div><span class="badge ${status.kind}" title="${escapeHTML(status.detail)}">${escapeHTML(status.label)}</span></div>
      <div class="node-meta"><div><span>${t("address")}</span><strong>${escapeHTML(node.public_ip || node.address || "-")}</strong></div><div><span>${t("country")}</span><strong>${escapeHTML(node.country || "-")}</strong></div><div><span>${t("priority")}</span><strong>${node.priority || "-"}</strong></div><div><span>${node.role === "dns" ? t("clientState") : t("targetCompatibility")}</span><strong title="${escapeHTML(status.detail)}">${escapeHTML(status.detail)}</strong></div></div>
      <div class="node-actions"><button class="btn small node-install" data-node-id="${nodeID(node.id)}">${t("showInstallCommand")}</button>${node.role === "dns" && !configForNode(node) ? `<button class="btn small primary node-manage-ip" data-node-id="${nodeID(node.id)}">${state.lang === "zh" ? "接入 IP 服务" : "Configure IP services"}</button>` : ""}<button class="btn small node-test" data-node-id="${nodeID(node.id)}" ${node.role !== "proxy" ? "disabled" : ""}>${t("triggerUnlock")}</button><button class="btn small node-edit" data-node-id="${nodeID(node.id)}">${t("editNode")}</button><button class="btn small danger node-delete" data-node-id="${nodeID(node.id)}">${t("deleteNode")}</button></div></article>`;
  }).join("")}</section>`;
}

function nodeCheckHTML() {
  const node = state.nodes.find(item => nodeID(item.id) === nodeID(state.modal.nodeId));
  const entries = unlockEntries(node);
  const rowsData = entries.map(([name, value]) => {
    const candidates = state.catalog.filter(item => serviceDetectorKey(item) === name);
    const service = candidates.find(item => targetAuditObservations(node?.id, item).length) || candidates[0];
    const compatibility = service ? targetCompatibility(node?.id, service) : {observations:[], measured:[], passed:0, kind:"warn", label:t("referenceOnly")};
    return {name, value, compatibility};
  });
  const actual = rowsData.filter(item => item.compatibility.observations.length);
  const available = actual.filter(item => item.compatibility.kind === "good");
  const problems = actual.filter(item => item.compatibility.kind !== "good");
  const reference = rowsData.filter(item => !item.compatibility.observations.length);
  const rows = rowsData.map(({name, value, compatibility}) => {
    const targetLines = compatibility.observations.map(item => `<span>${escapeHTML(item.ip)} · ${escapeHTML(item.result || t("auditPending"))}</span>`).join("");
    return `<div class="unlock-result-row"><div><strong>${escapeHTML(name)}</strong><span>${t("agentReference")} · ${escapeHTML(value)}</span>${targetLines}</div><span class="badge ${compatibility.kind}">${escapeHTML(compatibility.label)}</span></div>`;
  }).join("");
  return `<div class="modal-backdrop"><section class="modal panel"><header class="modal-head"><div><h2>${t("unlockResults")} · ${escapeHTML(node?.name || "-")}</h2><p>${t("unlockSourceHint")}</p></div><button class="btn icon modal-close">×</button></header><div class="modal-body"><div class="unlock-summary"><div><span>${t("actualAvailable")}</span><strong>${available.length}</strong></div><div><span>${t("actualProblem")}</span><strong>${problems.length}</strong></div><div><span>${t("referenceOnly")}</span><strong>${reference.length}</strong></div></div>${state.modal.running ? `<div class="unlock-wait"><div class="spinner"></div><span>${t("waitingResults")}</span></div>` : ""}<div class="unlock-results">${rows || `<div class="empty"><strong>${t("unknown")}</strong></div>`}</div>${state.modal.error ? `<div class="form-error">${escapeHTML(state.modal.error)}</div>` : ""}</div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close">${t("close")}</button><button class="btn primary" id="node-check-retry" ${state.modal.running ? "disabled" : ""}>${t("checkAgain")}</button></div></footer></section></div>`;
}

function ipConfigNode(config) { return state.nodes.find(node => nodeID(node.id) === nodeID(config.dns_node_id)); }

function ipConfigsHTML() {
  if (!state.ipConfigs.length) return `<div class="panel empty"><strong>${t("noIPConfigs")}</strong><button class="btn primary" id="empty-add-ip">＋ ${t("addIP")}</button></div>`;
  const totalTraffic = state.ipConfigs.reduce((total, config) => total + Number(config.traffic_rx_bytes || 0) + Number(config.traffic_tx_bytes || 0), 0);
  return `<section class="traffic-summary panel"><div><span>${t("totalTraffic")}</span><strong>${formatBytes(totalTraffic)}</strong><small>${t("trafficHint")}</small></div><button class="btn danger" id="clear-all-traffic">${t("clearAllTraffic")}</button></section><section class="ip-list panel"><div class="ip-list-head"><span>${t("targetIP")}</span><span>${t("selectedServices")}</span><span>${t("clientTraffic")}</span><span>${t("dnsClient")}</span><span></span></div>${state.ipConfigs.map(config => {
    const node = ipConfigNode(config); const status = clientState(config, node); const count = Object.keys(config.routes || {}).length; const traffic = Number(config.traffic_rx_bytes || 0) + Number(config.traffic_tx_bytes || 0);
    const audited = Object.values(config.service_results || {}).map(auditResultState); const passed = audited.filter(result => result.kind === "good").length;
    return `<article class="ip-row" data-ip-id="${escapeHTML(config.id)}" role="button" tabindex="0" aria-label="${escapeHTML(`${t("editIP")} ${config.ip}`)}"><div class="ip-main"><strong>${escapeHTML(config.ip)}</strong><span>${escapeHTML(config.note || config.node_name || "-")}</span></div><div class="ip-selection"><span class="badge good">${count} / ${state.catalog.length}</span>${audited.length ? `<small>${t("actualAudit")} ${passed}/${audited.length}</small>` : ""}</div><div class="ip-traffic" title="RX ${formatBytes(config.traffic_rx_bytes)} · TX ${formatBytes(config.traffic_tx_bytes)}"><strong>${formatBytes(traffic)}</strong><span>${t("trafficUpdated")}: ${formatDate(config.traffic_updated_at)}</span></div><div class="ip-node-state" title="${escapeHTML(status.detail)}"><span class="status-dot" style="background:${status.kind === "good" ? "var(--good)" : status.kind === "warn" ? "var(--warn)" : "var(--bad)"}"></span><span>${escapeHTML(status.label)}</span></div><div class="ip-row-actions"><button class="btn small ip-script" data-ip-id="${escapeHTML(config.id)}">${t("clientScript")}</button><button class="btn small primary ip-edit" data-ip-id="${escapeHTML(config.id)}">${t("editIP")}</button><button class="btn small ip-clear-traffic" data-ip-id="${escapeHTML(config.id)}">${t("clearTraffic")}</button><button class="btn small danger ip-delete" data-ip-id="${escapeHTML(config.id)}">${t("deleteNode")}</button></div></article>`;
  }).join("")}</section>`;
}

function bindShell() {
  document.querySelectorAll("[data-tab]").forEach(button => button.addEventListener("click", () => { state.tab = button.dataset.tab; render(); }));
  document.querySelectorAll(".theme-toggle-trigger").forEach(button => button.addEventListener("click", () => { state.theme = state.theme === "light" ? "dark" : "light"; localStorage.setItem("prism_theme_v2", state.theme); render(); }));
  document.querySelectorAll(".lang-toggle-trigger").forEach(button => button.addEventListener("click", () => { state.lang = state.lang === "zh" ? "en" : "zh"; localStorage.setItem("enhancer_lang", state.lang); render(); }));
  document.querySelectorAll(".logout-trigger").forEach(button => button.addEventListener("click", () => logout(true)));
  document.querySelectorAll(".account-settings-trigger").forEach(button => button.addEventListener("click", openAccountSettings));
  document.querySelectorAll(".site-settings-trigger").forEach(button => button.addEventListener("click", openBrandingSettings));
  document.getElementById("refresh").onclick = () => loadAll();
  document.getElementById("add-service")?.addEventListener("click", () => openServiceForm());
  document.getElementById("add-node")?.addEventListener("click", () => openNodeForm());
  document.getElementById("add-ip")?.addEventListener("click", () => openIPForm());
  document.getElementById("empty-add-ip")?.addEventListener("click", () => openIPForm());
  document.getElementById("empty-add-node")?.addEventListener("click", () => openNodeForm());
  document.getElementById("refresh-catalog")?.addEventListener("click", refreshCatalog);
  document.getElementById("service-search")?.addEventListener("input", event => { state.search = event.target.value; filterServiceCards(); });
  document.getElementById("category-filter")?.addEventListener("change", event => { state.category = event.target.value; render(); });
  document.getElementById("dns-select")?.addEventListener("change", async event => { state.dnsNodeId = event.target.value; localStorage.setItem("enhancer_dns", state.dnsNodeId); await loadRoutingState(); render(); });
  document.getElementById("clear-filter")?.addEventListener("click", () => { state.search = ""; state.category = ""; render(); });
  document.querySelectorAll(".service-open").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openService(button.dataset.serviceId); }));
  document.querySelectorAll(".service-card").forEach(card => card.addEventListener("click", () => openService(card.dataset.serviceId)));
  document.querySelectorAll(".selected-service-row").forEach(row => row.addEventListener("click", () => openService(row.dataset.serviceId)));
  document.querySelectorAll(".node-test").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); triggerNodeCheck(button.dataset.nodeId); }));
  document.querySelectorAll(".node-install").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openInstallCommand(button.dataset.nodeId); }));
  document.querySelectorAll(".node-manage-ip").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openIPForm(null, 1, state.nodes.find(node => nodeID(node.id) === nodeID(button.dataset.nodeId))); }));
  document.querySelectorAll(".node-edit").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openNodeForm(state.nodes.find(node => nodeID(node.id) === nodeID(button.dataset.nodeId))); }));
  document.querySelectorAll(".node-delete").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openDeleteNode(button.dataset.nodeId); }));
  document.querySelectorAll(".ip-row").forEach(row => {
    const open = () => openIPForm(state.ipConfigs.find(config => config.id === row.dataset.ipId), 2);
    row.addEventListener("click", open);
    row.addEventListener("keydown", event => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); open(); } });
  });
  document.querySelectorAll(".ip-edit").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openIPForm(state.ipConfigs.find(config => config.id === button.dataset.ipId), 2); }));
  document.querySelectorAll(".ip-script").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openIPScript(button.dataset.ipId); }));
  document.querySelectorAll(".ip-delete").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openIPDelete(button.dataset.ipId); }));
  document.querySelectorAll(".ip-clear-traffic").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); clearTraffic(button.dataset.ipId); }));
  document.getElementById("clear-all-traffic")?.addEventListener("click", clearAllTraffic);
  filterServiceCards();
}

function filterServiceCards() {
  const query = state.search.trim().toLowerCase();
  let visible = 0;
  document.querySelectorAll(".service-card").forEach(card => {
    const service = state.catalog.find(item => item.id === card.dataset.serviceId);
    const matches = service && serviceMatches(service, query);
    card.hidden = !matches;
    if (matches) visible++;
  });
  const empty = document.getElementById("service-filter-empty");
  if (empty) empty.hidden = visible !== 0;
}

async function refreshCatalog() {
  try { await api("/enhancer/api/catalog?refresh=1"); toast(t("saved"), "good"); await loadAll(); } catch (error) { toast(error.message, "error"); }
}

function openService(id) {
  const service = state.catalog.find(item => item.id === id);
  if (!service) return;
  const current = activeNodeFor(serviceRule(service));
  const fallback = proxyNodes().find(isOnline) || proxyNodes()[0];
  state.modal = {type:"service", service, proxyId:nodeID(current?.id || fallback?.id)};
  state.testResults = null;
  renderModal();
}
function openServiceForm(service = null) { state.modal = {type:"service-form", service}; renderModal(); }
function openAccountSettings() { state.modal = {type:"account", error:"", busy:false}; renderModal(); }
function openBrandingSettings() { state.modal = {type:"branding", error:"", busy:false}; renderModal(); }
function randomSecret() { return Math.random().toString(36).slice(-8); }
function emptyNodeDraft() { return {id:"", name:"", role:"proxy", public_ip:"", country:"", group:"", priority:1, secret:randomSecret()}; }
function openNodeForm(node = null) {
  const draft = node ? {...node, public_ip:node.public_ip || "", country:node.country || "", group:node.group || "", priority:node.priority || 1} : emptyNodeDraft();
  state.modal = {type:"node-form", node:node ? {...node} : null, draft, step:node ? 1 : 1, smartMode:node?.smart ? "smart" : "standard", error:""};
  renderModal();
}
function openDeleteNode(id) {
  const node = state.nodes.find(item => nodeID(item.id) === nodeID(id));
  if (!node) return;
  state.modal = {type:"node-delete", node};
  renderModal();
}
function openInstallCommand(id) {
  const node = state.nodes.find(item => nodeID(item.id) === nodeID(id));
  if (!node) return;
  const smartMode = node.role === "dns" && (node.smart === true || String(node.smart).toLowerCase() === "true") ? "smart" : "standard";
  state.modal = {type:"install-command", node, command:installCommand(node, smartMode)};
  renderModal();
}
function openIPForm(config = null, step = config ? 2 : 1, existingNode = null) {
  const routes = {...(config?.routes || {})};
  const defaultProxy = Object.values(routes)[0] || nodeID(proxyNodes()[0]?.id);
  state.modal = {type:"ip-form", config, existingNode, step, draft:{ip:config?.ip || serverHost(existingNode || {}) || "", note:config?.note || existingNode?.name || "", smart:config?.smart !== false}, routes, defaultProxy, serviceSearch:"", error:""};
  renderModal();
}
function ipScriptCommand(config) {
  return `wget -qO- https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/prismdns.sh | sudo bash -s -- --master ${location.origin} --token ${config.enrollment_token} --one-click --non-interactive`;
}
function openIPScript(id) {
  const config = state.ipConfigs.find(item => item.id === id);
  if (!config) return;
  state.modal = {type:"ip-script", config, command:ipScriptCommand(config)};
  renderModal();
}
function openIPDelete(id) {
  const config = state.ipConfigs.find(item => item.id === id);
  if (!config) return;
  state.modal = {type:"ip-delete", config};
  renderModal();
}
function closeModal() { state.modal = null; state.testResults = null; renderModal(); }

function renderModal() {
  const root = document.getElementById("modal-root");
  if (!state.modal) { root.innerHTML = ""; return; }
  if (state.modal.type === "service") root.innerHTML = serviceModalHTML(state.modal.service);
  if (state.modal.type === "service-form") root.innerHTML = serviceFormHTML(state.modal.service);
  if (state.modal.type === "node-form") root.innerHTML = nodeFormHTML();
  if (state.modal.type === "install-command") root.innerHTML = installCommandHTML();
  if (state.modal.type === "node-delete") root.innerHTML = nodeDeleteHTML();
  if (state.modal.type === "node-check") root.innerHTML = nodeCheckHTML();
  if (state.modal.type === "ip-form") root.innerHTML = ipFormHTML();
  if (state.modal.type === "ip-script") root.innerHTML = ipScriptHTML();
  if (state.modal.type === "ip-delete") root.innerHTML = ipDeleteHTML();
  if (state.modal.type === "account") root.innerHTML = accountModalHTML();
  if (state.modal.type === "branding") root.innerHTML = brandingModalHTML();
  bindModal();
}

function brandingModalHTML() {
  const branding = state.branding || {};
  return `<div class="modal-backdrop"><form class="modal medium panel" id="branding-form"><header class="modal-head"><div><h2>${t("siteSettings")}</h2><p>${t("siteSettingsHint")}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><div class="field"><label>${t("siteName")}</label><input class="input" name="site_name" value="${escapeHTML(branding.site_name || siteName())}" required maxlength="48" autocomplete="off"></div><div class="field"><label>${t("browserTitle")}</label><input class="input" name="browser_title" value="${escapeHTML(branding.browser_title || browserTitle())}" required maxlength="96" autocomplete="off"></div><div class="field"><label>${t("siteTagline")}</label><input class="input" name="site_tagline" value="${escapeHTML(branding.site_tagline || siteTagline())}" maxlength="120" autocomplete="off"></div><div class="branding-preview"><span>${t("siteName")}</span><strong id="branding-preview-name">${escapeHTML(siteName())}</strong><small id="branding-preview-tagline">${escapeHTML(siteTagline())}</small></div><div class="form-error">${escapeHTML(state.modal.error || "")}</div></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button" ${state.modal.busy ? "disabled" : ""}>${t("cancel")}</button><button class="btn primary" type="submit" ${state.modal.busy ? "disabled" : ""}>${t("saveBranding")}</button></div></footer></form></div>`;
}

function accountModalHTML() {
  const username = state.user.username || "admin";
  return `<div class="modal-backdrop"><form class="modal medium panel" id="account-form"><header class="modal-head"><div><h2>${t("accountSettings")}</h2><p>${t("accountHint")}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><div class="form-columns"><div class="field"><label>${t("oldUsername")}</label><input class="input" name="old_username" value="${escapeHTML(username)}" required autocomplete="username" autocapitalize="off" spellcheck="false"></div><div class="field"><label>${t("oldPassword")}</label><input class="input" type="password" name="old_password" required autocomplete="current-password"></div></div><div class="field"><label>${t("newUsername")}</label><input class="input" name="new_username" value="${escapeHTML(username)}" required minlength="3" maxlength="20" pattern="[A-Za-z0-9_]+" autocomplete="off" autocapitalize="off" spellcheck="false"><span class="hint">${state.lang === "zh" ? "3-20 位字母、数字或下划线" : "3-20 letters, numbers, or underscores"}</span></div><div class="form-columns"><div class="field"><label>${t("newPassword")}</label><input class="input" type="password" name="new_password" required minlength="6" autocomplete="new-password"></div><div class="field"><label>${t("confirmPassword")}</label><input class="input" type="password" name="confirm_password" required minlength="6" autocomplete="new-password"></div></div><div class="form-error">${escapeHTML(state.modal.error || "")}</div></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button" ${state.modal.busy ? "disabled" : ""}>${t("cancel")}</button><button class="btn primary" type="submit" ${state.modal.busy ? "disabled" : ""}>${t("updateAccount")}</button></div></footer></form></div>`;
}

function serviceModalHTML(service) {
  const rule = serviceRule(service); const current = activeNodeFor(rule); const manual = overrideFor(rule); const proxies = proxyNodes();
  const status = serviceStatus(service, current);
  return `<div class="modal-backdrop"><section class="modal panel"><header class="modal-head"><div><h2>${t("serviceConfig")} · ${escapeHTML(service.name)}</h2><p>${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")} · ${status.label}</p></div><button class="btn icon modal-close">×</button></header>
    <div class="modal-body"><div class="field"><label>${t("targetServer")}</label><div class="server-list">${proxies.length ? proxies.map(node => {
      const selected = nodeID(state.modal.proxyId) === nodeID(node.id); const nodeStatus = serviceStatus(service, node);
      return `<label class="server-option ${selected ? "selected" : ""}"><input type="radio" name="proxy" value="${nodeID(node.id)}" ${selected ? "checked" : ""}><div class="server-name"><strong>${escapeHTML(node.name)}</strong><span>${escapeHTML(node.country || node.public_ip || node.address || "-")}</span></div><span class="badge ${nodeStatus.kind}">${nodeStatus.label}</span></label>`;
    }).join("") : `<div class="empty"><strong>${t("noProxy")}</strong></div>`}</div></div>
    <div class="inline" style="margin-top:15px"><span class="badge ${manual ? "warn" : ""}">${manual ? t("manual") : t("auto")}</span><span class="hint">${t("currentRoute")}: ${escapeHTML(current?.name || "-")}</span></div>
    ${state.testResults ? testResultsHTML(state.testResults) : ""}</div>
    <footer class="modal-foot"><div class="inline">${service.custom ? `<button class="btn small" id="edit-custom">${t("customEdit")}</button><button class="btn small danger" id="delete-custom">${t("customDelete")}</button>` : ""}</div><div class="modal-foot-right"><button class="btn" id="test-service" ${!state.modal.proxyId ? "disabled" : ""}>${t("connectivity")}</button><button class="btn" id="reset-service" ${!rule || !manual ? "disabled" : ""}>${t("reset")}</button><button class="btn primary" id="assign-service" ${!state.modal.proxyId || !state.dnsNodeId ? "disabled" : ""}>${t("assign")}</button></div></footer></section></div>`;
}

function testResultsHTML(results) {
  return `<div class="test-results">${results.map(result => {
    const detail = result.detail || result.error || (result.addresses || []).join(", ");
    const label = result.detail ? (result.success ? t("supported") : t("unavailable")) : (result.success ? `TLS ${result.tls_ms}ms` : "FAIL");
    return `<div class="test-row"><span title="${escapeHTML(detail)}">${escapeHTML(result.domain)} · ${escapeHTML(detail)}</span><strong style="color:${result.success ? "var(--good)" : "var(--bad)"}">${label}</strong></div>`;
  }).join("")}</div>`;
}

function serviceFormHTML(service) {
  return `<div class="modal-backdrop"><form class="modal medium panel" id="service-form"><header class="modal-head"><div><h2>${service ? t("customEdit") : t("addService")}</h2></div><button class="btn icon modal-close" type="button">×</button></header>
    <div class="modal-body form-stack"><div class="field"><label>${t("serviceName")}</label><input class="input" name="name" value="${escapeHTML(service?.name || "")}" required maxlength="80"></div>
    <div class="field"><label>${t("category")}</label><input class="input" name="category" value="${escapeHTML(service?.category || (state.lang === "zh" ? "自定义服务" : "Custom services"))}" maxlength="80"></div>
    <div class="field"><label>${t("domainList")}</label><textarea class="textarea" name="domains" placeholder="example.com&#10;cdn.example.com" required>${escapeHTML((service?.domains || []).join("\n"))}</textarea><span class="hint">${state.lang === "zh" ? "每行一个域名，也支持逗号或分号分隔。" : "One domain per line; commas and semicolons are also accepted."}</span></div></div>
    <footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit">${t("save")}</button></div></footer></form></div>`;
}

function nodeFormHTML() {
  const modal = state.modal; const draft = modal.draft || emptyNodeDraft(); const editing = !!modal.node;
  if (!editing && modal.step === 2) {
    return `<div class="modal-backdrop"><form class="modal medium panel" id="node-form"><header class="modal-head"><div><h2>${t("createNode")}</h2><p>${t("nextStep")}</p></div><button class="btn icon modal-close" type="button">×</button></header>
      <div class="modal-body form-stack"><div class="node-summary"><strong>${escapeHTML(draft.name)}</strong><span>${draft.role === "proxy" ? t("proxy") : t("dns")} · ${escapeHTML(draft.public_ip || "-")}</span><span>${escapeHTML(draft.group || "-")} · ${t("priority")} ${draft.priority}</span></div>
      ${draft.role === "dns" ? `<div class="field"><label>${t("smartMode")}</label><select class="select" name="smart_mode"><option value="standard" ${modal.smartMode !== "smart" ? "selected" : ""}>${t("standardMode")}</option><option value="smart" ${modal.smartMode === "smart" ? "selected" : ""}>${t("smartMode")}</option></select><span class="hint">DNS Agent</span></div>` : ""}
      <div class="form-error">${escapeHTML(modal.error || "")}</div></div>
      <footer class="modal-foot"><button class="btn" id="node-back" type="button">${t("back")}</button><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit">${t("createNode")}</button></div></footer></form></div>`;
  }
  return `<div class="modal-backdrop"><form class="modal medium panel" id="node-form"><header class="modal-head"><div><h2>${editing ? t("editNode") : t("addNode")}</h2><p>${editing ? t("nodes") : t("nextStep")}</p></div><button class="btn icon modal-close" type="button">×</button></header>
    <div class="modal-body form-stack"><div class="field"><label>${t("nodeName")}</label><input class="input" name="name" value="${escapeHTML(draft.name || "")}" placeholder="Tokyo Server 01" required maxlength="64" autocomplete="off"><span class="hint">${t("nodeHint")}</span></div>
    <div class="field"><label>${t("role")}</label><div class="segmented node-role-toggle"><button type="button" class="segment ${draft.role === "proxy" ? "active" : ""}" data-node-role="proxy">${t("proxy")}</button><button type="button" class="segment ${draft.role === "dns" ? "active" : ""}" data-node-role="dns">${t("dns")}</button></div><input type="hidden" name="role" value="${escapeHTML(draft.role || "proxy")}"></div>
    <div class="field"><label>${t("address")}</label><input class="input" name="public_ip" value="${escapeHTML(draft.public_ip || "")}" placeholder="203.0.113.10 或 2001:db8::10" autocomplete="off"></div>
    <div class="form-columns"><div class="field"><label>${t("country")}</label><input class="input" name="country" value="${escapeHTML(draft.country || "")}" placeholder="SG / JP / US"></div><div class="field"><label>${t("priority")}</label><input class="input" type="number" name="priority" min="1" max="100" value="${escapeHTML(draft.priority || 1)}"></div></div>
    <div class="field"><label>${t("group")}</label><input class="input" name="group" value="${escapeHTML(draft.group || "")}" maxlength="64"><span class="hint">${t("groupHint")}</span></div><div class="form-error">${escapeHTML(modal.error || "")}</div></div>
    <footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit">${editing ? t("save") : t("nextStep")}</button></div></footer></form></div>`;
}

function installCommandHTML() {
  const modal = state.modal;
  return `<div class="modal-backdrop"><section class="modal medium panel"><header class="modal-head"><div><h2>${t("installCommand")}</h2><p>${escapeHTML(modal.node?.name || "")}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body"><p class="hint">${state.lang === "zh" ? "请在目标节点服务器执行以下命令，Agent 连接后节点会自动变为在线。" : "Run this command on the target node. The node becomes online after the Agent connects."}</p><pre class="code" id="install-command">${escapeHTML(modal.command)}</pre></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn" id="copy-install-command" type="button">${t("copy")}</button><button class="btn primary modal-close" type="button">${t("close")}</button></div></footer></section></div>`;
}

function nodeDeleteHTML() {
  const node = state.modal.node;
  return `<div class="modal-backdrop"><section class="modal medium panel"><header class="modal-head"><div><h2>${t("deleteNode")}</h2><p>${escapeHTML(node.name)}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body"><p>${t("deleteNodeConfirm")}</p></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn danger" id="confirm-delete-node" type="button">${t("deleteNode")}</button></div></footer></section></div>`;
}

function ipFormHTML() {
  const modal = state.modal; const draft = modal.draft; const editing = !!modal.config; const proxies = proxyNodes();
  if (modal.step === 1) {
    return `<div class="modal-backdrop"><form class="modal medium panel" id="ip-form"><header class="modal-head"><div><h2>${editing ? t("editIP") : t("addIP")}</h2><p>${t("targetIP")}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><div class="field"><label>${t("targetIP")}</label><input class="input" name="ip" value="${escapeHTML(draft.ip)}" placeholder="203.0.113.10" ${editing ? "disabled" : ""} required></div><div class="field"><label>${t("note")}</label><input class="input" name="note" value="${escapeHTML(draft.note)}" maxlength="80"></div><div class="field"><label>${t("defaultProxy")}</label><select class="select" name="default_proxy" required><option value="">${t("noProxy")}</option>${proxies.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(modal.defaultProxy) ? "selected" : ""}>${escapeHTML(node.name)} · ${escapeHTML(node.country || node.public_ip || node.address || "-")}</option>`).join("")}</select></div><label class="toggle-row"><input type="checkbox" name="smart" ${draft.smart ? "checked" : ""}><span><strong>${t("smartMode")}</strong><small>DNS Client Agent</small></span></label><div class="form-error">${escapeHTML(modal.error || "")}</div></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit" ${!proxies.length ? "disabled" : ""}>${t("nextStep")}</button></div></footer></form></div>`;
  }
  const services = state.catalog;
  const selectedCount = Object.keys(modal.routes).length;
  return `<div class="modal-backdrop"><form class="modal ip-modal panel" id="ip-form"><header class="modal-head"><div><h2>${t("chooseServices")}</h2><p>${escapeHTML(draft.ip)} · ${t("selectedServices")} <span id="ip-selected-count">${selectedCount}</span></p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body ip-config-body"><input class="input" id="ip-service-search" value="${escapeHTML(modal.serviceSearch)}" placeholder="${t("search")}" autocomplete="off" autocapitalize="off" spellcheck="false"><div class="ip-service-picker">${services.map(service => {
    const selectedProxy = modal.routes[service.id] || ""; const checked = !!selectedProxy; const audit = auditResultState(modal.config?.service_results?.[service.id]);
    return `<article class="ip-service-option ${checked ? "selected" : ""}" data-service-id="${escapeHTML(service.id)}"><label><input type="checkbox" class="ip-service-check" data-service-id="${escapeHTML(service.id)}" ${checked ? "checked" : ""}>${serviceIconHTML(service)}<span class="ip-service-name"><strong>${escapeHTML(displayServiceName(service))}</strong><small>${escapeHTML(displayCategory(service.category))}</small></span></label><select class="select ip-service-proxy" data-service-id="${escapeHTML(service.id)}" ${checked ? "" : "disabled"}>${proxies.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(selectedProxy || modal.defaultProxy) ? "selected" : ""}>${escapeHTML(node.name)}</option>`).join("")}</select><div class="service-audit" ${checked ? "" : "hidden"}><span>${t("actualAudit")}</span><span class="badge ${audit.kind}" title="${escapeHTML(audit.label)}">${escapeHTML(audit.label)}</span></div></article>`;
  }).join("")}</div><div class="empty" id="ip-service-filter-empty" hidden><strong>${t("noServices")}</strong></div><div class="form-error">${escapeHTML(modal.error || "")}</div></div><footer class="modal-foot delivery-foot"><button class="btn" id="ip-back" type="button">${t("back")}</button><div class="delivery-impact"><strong>${state.lang === "zh" ? "保存后自动下发" : "Automatic delivery after save"}</strong><span>${state.lang === "zh" ? "Agent 将自动同步服务增删并安全刷新 DNS" : "The Agent applies additions and removals, then safely refreshes DNS"}</span></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" id="ip-save-config" type="submit" ${!selectedCount ? "disabled" : ""}>${state.lang === "zh" ? "保存并自动下发" : "Save and deliver"}</button></div></footer></form></div>`;
}

function ipScriptHTML() {
  const modal = state.modal;
  return `<div class="modal-backdrop"><section class="modal medium panel"><header class="modal-head"><div><h2>${t("scriptCommand")}</h2><p>${escapeHTML(modal.config.ip)}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body"><p class="hint">${t("runScriptHint")}</p><pre class="code" id="install-command">${escapeHTML(modal.command)}</pre></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn" id="copy-install-command" type="button">${t("copy")}</button><button class="btn primary modal-close" type="button">${t("close")}</button></div></footer></section></div>`;
}

function ipDeleteHTML() {
  const config = state.modal.config;
  const message = config.external_dns_node ? (state.lang === "zh" ? "删除此 IP 配置？手动创建的 DNS 节点将被保留。" : "Delete this IP configuration? The manually created DNS node will be kept.") : t("ipDeleteConfirm");
  return `<div class="modal-backdrop"><section class="modal medium panel"><header class="modal-head"><div><h2>${t("deleteNode")}</h2><p>${escapeHTML(config.ip)}</p></div><button class="btn icon modal-close" type="button">&times;</button></header><div class="modal-body"><p>${message}</p></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn danger" id="confirm-delete-ip" type="button">${t("deleteNode")}</button></div></footer></section></div>`;
}
function bindModal() {
  document.querySelectorAll(".modal-close").forEach(button => button.addEventListener("click", closeModal));
  document.querySelector(".modal-backdrop")?.addEventListener("click", event => { if (event.target.classList.contains("modal-backdrop")) closeModal(); });
  document.querySelectorAll("input[name=proxy]").forEach(input => input.addEventListener("change", () => { state.modal.proxyId = input.value; renderModal(); }));
  document.getElementById("assign-service")?.addEventListener("click", assignService);
  document.getElementById("reset-service")?.addEventListener("click", resetService);
  document.getElementById("test-service")?.addEventListener("click", testService);
  document.getElementById("edit-custom")?.addEventListener("click", () => openServiceForm(state.modal.service));
  document.getElementById("delete-custom")?.addEventListener("click", deleteCustomService);
  document.getElementById("service-form")?.addEventListener("submit", saveCustomService);
  document.querySelectorAll("[data-node-role]").forEach(button => button.addEventListener("click", () => {
    const form = document.getElementById("node-form"); form.querySelector('input[name="role"]').value = button.dataset.nodeRole;
    document.querySelectorAll("[data-node-role]").forEach(item => item.classList.toggle("active", item === button));
  }));
  document.getElementById("node-back")?.addEventListener("click", () => { state.modal.step = 1; state.modal.error = ""; renderModal(); });
  document.getElementById("node-form")?.addEventListener("submit", submitNodeForm);
  document.getElementById("copy-install-command")?.addEventListener("click", copyInstallCommand);
  document.getElementById("confirm-delete-node")?.addEventListener("click", deleteNode);
  document.getElementById("node-check-retry")?.addEventListener("click", () => triggerNodeCheck(state.modal.nodeId));
  document.getElementById("ip-form")?.addEventListener("submit", submitIPForm);
  document.getElementById("ip-back")?.addEventListener("click", () => { state.modal.step = 1; state.modal.error = ""; renderModal(); });
  document.getElementById("ip-service-search")?.addEventListener("input", event => { state.modal.serviceSearch = event.target.value; filterIPServiceOptions(); });
  document.querySelectorAll(".ip-service-check").forEach(input => input.addEventListener("change", () => {
    const option = input.closest(".ip-service-option");
    const select = option?.querySelector(".ip-service-proxy");
    const audit = option?.querySelector(".service-audit");
    if (input.checked) {
      state.modal.routes[input.dataset.serviceId] = select?.value || state.modal.defaultProxy;
    } else {
      delete state.modal.routes[input.dataset.serviceId];
    }
    option?.classList.toggle("selected", input.checked);
    if (select) select.disabled = !input.checked;
    if (audit) audit.hidden = !input.checked;
    const selectedCount = Object.keys(state.modal.routes).length;
    const count = document.getElementById("ip-selected-count");
    const save = document.getElementById("ip-save-config");
    if (count) count.textContent = String(selectedCount);
    if (save) save.disabled = selectedCount === 0;
  }));
  document.querySelectorAll(".ip-service-proxy").forEach(select => select.addEventListener("change", () => {
    if (select.closest(".ip-service-option")?.querySelector(".ip-service-check")?.checked) state.modal.routes[select.dataset.serviceId] = select.value;
  }));
  document.getElementById("confirm-delete-ip")?.addEventListener("click", deleteIPConfig);
  document.getElementById("account-form")?.addEventListener("submit", updateAccount);
  const brandingForm = document.getElementById("branding-form");
  brandingForm?.addEventListener("submit", updateBranding);
  brandingForm?.addEventListener("input", () => {
    const form = new FormData(brandingForm);
    document.getElementById("branding-preview-name").textContent = String(form.get("site_name") || "");
    document.getElementById("branding-preview-tagline").textContent = String(form.get("site_tagline") || "");
  });
  filterIPServiceOptions();
}

async function updateBranding(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const payload = {
    site_name:String(form.get("site_name") || "").trim(),
    browser_title:String(form.get("browser_title") || "").trim(),
    site_tagline:String(form.get("site_tagline") || "").trim()
  };
  state.modal.busy = true; state.modal.error = ""; renderModal();
  try {
    state.branding = await api("/enhancer/api/branding", {method:"PUT", body:JSON.stringify(payload)});
    state.modal = null;
    render();
    toast(t("brandingUpdated"), "good");
  } catch (error) {
    state.modal.busy = false;
    state.modal.error = error.message;
    renderModal();
  }
}

async function updateAccount(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const oldUsername = String(form.get("old_username") || "").trim();
  const oldPassword = String(form.get("old_password") || "");
  const newUsername = String(form.get("new_username") || "").trim();
  const newPassword = String(form.get("new_password") || "");
  if (newPassword !== String(form.get("confirm_password") || "")) { state.modal.error = t("passwordMismatch"); renderModal(); return; }
  if (!/^[A-Za-z0-9_]{3,20}$/.test(newUsername) || newPassword.length < 6) { state.modal.error = t("failed"); renderModal(); return; }
  state.modal.busy = true; state.modal.error = ""; renderModal();
  try {
    await api("/enhancer/api/account", {method:"POST", body:JSON.stringify({old_username:oldUsername, old_password:oldPassword, new_username:newUsername, new_password:newPassword})});
    localStorage.removeItem("prism_token"); localStorage.removeItem("prism_user");
    state.token = ""; state.user = {}; state.modal = null;
    toast(t("accountUpdated"), "good");
    setTimeout(() => location.reload(), 900);
  } catch (error) {
    state.modal.busy = false;
    state.modal.error = /unauthorized|invalid|incorrect/i.test(error.message) ? t("credentialsInvalid") : error.message;
    renderModal();
  }
}

function filterIPServiceOptions() {
  if (state.modal?.type !== "ip-form" || state.modal.step !== 2) return;
  const query = state.modal.serviceSearch.trim().toLowerCase();
  let visible = 0;
  document.querySelectorAll(".ip-service-option").forEach(option => {
    const service = state.catalog.find(item => item.id === option.dataset.serviceId);
    const matches = service && serviceMatches(service, query);
    option.hidden = !matches;
    if (matches) visible++;
  });
  const empty = document.getElementById("ip-service-filter-empty");
  if (empty) empty.hidden = visible !== 0;
}

async function ensureRule(service, proxyId) {
  let rule = serviceRule(service);
  if (rule) return rule;
  const payload = {name:`Stream · ${service.name}`, type:"RULE-SET", source_type:"all", source_val:"", target_type:"node", target_val:proxyId, value:`${location.origin}/enhancer/rules/${service.id}.list`, enabled:true};
  const response = await api("/api/rules", {method:"POST", body:JSON.stringify(payload)});
  rule = response?.id ? response : null;
  if (!rule) {
    const rules = await api("/api/rules"); state.rules = Array.isArray(rules) ? rules : state.rules; rule = serviceRule(service);
  } else state.rules.push(rule);
  if (!rule) throw new Error(t("failed"));
  toast(t("ruleCreated"), "good");
  return rule;
}

async function assignService() {
  const service = state.modal.service; const proxyId = state.modal.proxyId;
  try {
    const rule = await ensureRule(service, proxyId);
    await api(`/api/rules/${rule.id}/override`, {method:"POST", body:JSON.stringify({dns_node_id:state.dnsNodeId, proxy_node_id:proxyId})});
    toast(t("switched"), "good"); closeModal(); await loadAll(true);
  } catch (error) { toast(error.message, "error"); }
}

async function resetService() {
  const rule = serviceRule(state.modal.service); if (!rule) return;
  try { await api(`/api/rules/${rule.id}/override`, {method:"POST", body:JSON.stringify({dns_node_id:state.dnsNodeId, proxy_node_id:""})}); toast(t("resetDone"), "good"); closeModal(); await loadAll(true); }
  catch (error) { toast(error.message, "error"); }
}

function serverHost(node) {
  const value = String(node.public_ip || node.address || "").split(",")[0].trim();
  if (value.startsWith("[")) return value.slice(1, value.indexOf("]"));
  if ((value.match(/:/g) || []).length === 1 && value.includes(".")) return value.split(":")[0];
  return value;
}

async function testService() {
  const service = state.modal.service; const node = state.nodes.find(item => nodeID(item.id) === nodeID(state.modal.proxyId)); if (!node) return;
  const button = document.getElementById("test-service"); button.disabled = true; button.textContent = t("testing");
  try {
    const results = [];
    const targetConfig = state.ipConfigs.find(config => nodeID(config.dns_node_id) === nodeID(state.dnsNodeId));
    const targetResult = targetConfig?.service_results?.[service.id];
    const detector = serviceDetectorKey(service);
    if (targetResult) {
      results.push({domain:`${t("actualAudit")} · UnlockTests`, detail:targetResult, success:auditResultState(targetResult).kind === "good"});
    } else if (detector) {
      try {
        const latest = await refreshNodeUnlock(node.id, true);
        const value = parseUnlock(latest)[detector] || "No result";
        results.push({domain:`UnlockTests · ${detector}`, detail:value, success:unlockPassed(value)});
      } catch (error) {
        results.push({domain:`UnlockTests · ${detector}`, detail:error.message, success:false});
      }
    }
    const probeDomain = preferredServiceTestDomain(service);
    const data = await api("/enhancer/api/connectivity", {method:"POST", body:JSON.stringify({proxy_server:serverHost(node), domains:probeDomain ? [probeDomain] : [], timeout_ms:12000})});
    results.push(...(data.results || []));
    state.testResults = results;
    toast(t("testDone"), "good"); renderModal();
  } catch (error) { toast(error.message, "error"); button.disabled = false; button.textContent = t("connectivity"); }
}

async function saveCustomService(event) {
  event.preventDefault(); const form = new FormData(event.currentTarget); const existing = state.modal.service;
  const payload = {name:form.get("name"), category:form.get("category"), domains:String(form.get("domains")).split(/\r?\n|,|;/)};
  try {
    const path = existing ? `/enhancer/api/custom-services/${existing.id}` : "/enhancer/api/custom-services";
    await api(path, {method:existing ? "PUT" : "POST", body:JSON.stringify(payload)});
    const rule = existing ? serviceRule(existing) : null;
    if (rule) await api(`/api/rules/${rule.id}`, {method:"PUT", body:JSON.stringify({...rule, name:`Stream · ${payload.name}`})});
    toast(t("saved"), "good"); closeModal(); await loadAll(true);
  } catch (error) { toast(error.message, "error"); }
}

async function deleteCustomService() {
  const service = state.modal.service; if (!confirm(t("deleteConfirm"))) return;
  try {
    const rule = serviceRule(service);
    if (rule) await api(`/api/rules/${rule.id}`, {method:"DELETE"});
    await api(`/enhancer/api/custom-services/${service.id}`, {method:"DELETE"}); toast(t("deleted"), "good"); closeModal(); await loadAll(true);
  }
  catch (error) { toast(error.message, "error"); }
}

function submitIPForm(event) {
  event.preventDefault();
  if (state.modal.step === 2) { saveIPConfig(); return; }
  const form = new FormData(event.currentTarget);
  const ip = state.modal.config?.ip || String(form.get("ip") || "").trim();
  const defaultProxy = String(form.get("default_proxy") || "");
  if (!ip) { state.modal.error = t("addressInvalid"); renderModal(); return; }
  if (!defaultProxy) { state.modal.error = t("noProxy"); renderModal(); return; }
  state.modal.draft = {ip, note:String(form.get("note") || "").trim(), smart:form.get("smart") === "on", existing_dns_node_id:nodeID(state.modal.existingNode?.id)};
  state.modal.defaultProxy = defaultProxy;
  state.modal.step = 2;
  state.modal.error = "";
  renderModal();
}

async function saveIPConfig() {
  const modal = state.modal;
  if (!Object.keys(modal.routes).length) return;
  const payload = {...modal.draft, routes:modal.routes};
  try {
    const editing = !!modal.config;
    const path = editing ? `/enhancer/api/ip-configs/${encodeURIComponent(modal.config.id)}` : "/enhancer/api/ip-configs";
    const config = await api(path, {method:editing ? "PUT" : "POST", body:JSON.stringify(payload)});
    state.modal = editing ? null : {type:"ip-script", config, command:ipScriptCommand(config)};
    await loadAll(true);
    renderModal();
    toast(t("saved"), "good");
  } catch (error) { state.modal.error = error.message; renderModal(); }
}

async function deleteIPConfig() {
  const config = state.modal?.config;
  if (!config) return;
  try {
    await api(`/enhancer/api/ip-configs/${encodeURIComponent(config.id)}`, {method:"DELETE"});
    state.modal = null;
    await loadAll(true);
    toast(t("deleted"), "good");
  } catch (error) { toast(error.message, "error"); }
}

async function clearTraffic(id) {
  if (!confirm(state.lang === "zh" ? "清零该 IP 的已用流量？" : "Clear traffic for this IP?")) return;
  try {
    await api(`/enhancer/api/traffic/${encodeURIComponent(id)}`, {method:"DELETE"});
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) { toast(error.message, "error"); }
}

async function clearAllTraffic() {
  if (!confirm(state.lang === "zh" ? "清零全部 IP 的已用流量？" : "Clear traffic for every IP?")) return;
  try {
    await Promise.all(state.ipConfigs.map(config => api(`/enhancer/api/traffic/${encodeURIComponent(config.id)}`, {method:"DELETE"})));
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) { toast(error.message, "error"); }
}

function readNodeDraft(form) {
  const current = state.modal.draft || emptyNodeDraft();
  const normalizeLabel = (value, keepComma = false) => String(value || "").trim().replace(keepComma ? /[^\p{L}\p{N} ,._-]+/gu : /[^\p{L}\p{N} ._-]+/gu, " ").replace(/\s+/g, " ").trim();
  return {...current, name:normalizeLabel(form.get("name")), role:String(form.get("role") || "proxy"), public_ip:String(form.get("public_ip") || "").trim(), country:String(form.get("country") || "").trim(), group:normalizeLabel(form.get("group"), true), priority:Number(form.get("priority")) || 0};
}
function validAddress(value) {
  if (!value) return true;
  const candidate = value.replace(/^\[|\]$/g, "");
  const ipv4 = /^(\d{1,3}\.){3}\d{1,3}$/.test(candidate) && candidate.split(".").every(part => Number(part) <= 255);
  const ipv6 = candidate.includes(":") && /^[0-9a-fA-F:]+$/.test(candidate);
  return ipv4 || ipv6;
}

function validateNodeDraft(draft) {
  if (!draft.name) return t("nameRequired");
  if (draft.name.length > 64) return t("nameTooLong");
  if (!/^[\p{L}\p{N} ._-]+$/u.test(draft.name)) return t("nameInvalid");
  if (draft.group.length > 64) return t("groupTooLong");
  if (!/^[\p{L}\p{N} ,._-]*$/u.test(draft.group)) return t("groupInvalid");
  if (!validAddress(draft.public_ip)) return t("addressInvalid");
  if (draft.role === "proxy" && (!Number.isInteger(draft.priority) || draft.priority < 1 || draft.priority > 100)) return t("priorityInvalid");
  return "";
}

function submitNodeForm(event) {
  event.preventDefault();
  if (!state.modal.node && state.modal.step === 2) {
    state.modal.smartMode = event.currentTarget.querySelector('[name="smart_mode"]')?.value || "standard";
    createNode(state.modal.draft);
    return;
  }
  const draft = readNodeDraft(new FormData(event.currentTarget));
  const error = validateNodeDraft(draft);
  if (error) { state.modal.error = error; renderModal(); return; }
  state.modal.draft = draft; state.modal.error = "";
  if (state.modal.node) { saveNode(draft); return; }
  if (state.modal.step === 1) { state.modal.step = 2; renderModal(); return; }
}

async function saveNode(draft) {
  const node = state.modal.node;
  try {
    await api(`/enhancer/api/nodes/${encodeURIComponent(node.id)}`, {method:"PUT", body:JSON.stringify({...node, ...draft})});
    state.modal = null; toast(t("saved"), "good"); await loadAll(true);
  } catch (error) { state.modal.error = error.message; renderModal(); }
}

function installCommand(node, smartMode) {
  const smart = node.role === "dns" && smartMode === "smart" ? " --smart" : "";
  const ip = node.role === "proxy" && node.public_ip ? ` --ip "${node.public_ip.replace(/"/g, "")}"` : "";
  return `curl -fsSL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh | bash -s -- --master ${location.origin} --secret ${node.secret || "<secret>"}${smart}${ip}`;
}

async function createNode(draft) {
  const payload = {...draft, secret:draft.secret || randomSecret()};
  try {
    const node = await api("/enhancer/api/nodes", {method:"POST", body:JSON.stringify(payload)});
    const created = node && typeof node === "object" ? {...payload, ...node} : payload;
    state.modal = {type:"install-command", node:created, command:installCommand(created, state.modal.smartMode)};
    await loadAll(true);
    renderModal();
    toast(t("saved"), "good");
  } catch (error) { state.modal.error = error.message; state.modal.step = 1; renderModal(); }
}

async function deleteNode() {
  const node = state.modal?.node;
  if (!node) return;
  try {
    await api(`/enhancer/api/nodes/${encodeURIComponent(node.id)}`, {method:"DELETE"});
    if (nodeID(state.dnsNodeId) === nodeID(node.id)) { state.dnsNodeId = ""; localStorage.removeItem("enhancer_dns"); }
    state.modal = null; toast(t("deleted"), "good"); await loadAll(true);
  } catch (error) { toast(error.message, "error"); }
}

async function copyInstallCommand() {
  const command = state.modal?.command || "";
  try {
    if (navigator.clipboard?.writeText) await navigator.clipboard.writeText(command);
    else { const input = document.createElement("textarea"); input.value = command; document.body.appendChild(input); input.select(); document.execCommand("copy"); input.remove(); }
    toast(t("copied"), "good");
  } catch (error) { toast(error.message, "error"); }
}

function delay(milliseconds) { return new Promise(resolve => setTimeout(resolve, milliseconds)); }

async function refreshNodeUnlock(id, trigger = false) {
  const previous = JSON.stringify(parseUnlock(state.nodes.find(node => nodeID(node.id) === nodeID(id))));
  if (trigger) await api(`/api/nodes/${encodeURIComponent(id)}/check_unlock`, {method:"POST"});
  const started = Date.now();
  let latest = state.nodes.find(node => nodeID(node.id) === nodeID(id));
  while (Date.now() - started < 45000) {
    await delay(2000);
    const nodes = await api("/enhancer/api/nodes");
    state.nodes = Array.isArray(nodes) ? nodes : state.nodes;
    latest = state.nodes.find(node => nodeID(node.id) === nodeID(id));
    const current = JSON.stringify(parseUnlock(latest));
    if (current !== "{}" && (current !== previous || Date.now() - started >= 6000)) return latest;
  }
  if (!latest || !unlockEntries(latest).length) throw new Error(state.lang === "zh" ? "节点未在 45 秒内返回检测结果" : "The node did not return results within 45 seconds");
  return latest;
}

async function triggerNodeCheck(id) {
  state.modal = {type:"node-check", nodeId:id, running:true, error:""};
  renderModal();
  try {
    await refreshNodeUnlock(id, true);
    if (state.modal?.type === "node-check" && nodeID(state.modal.nodeId) === nodeID(id)) {
      state.modal.running = false;
      renderModal();
    }
  } catch (error) {
    if (state.modal?.type === "node-check" && nodeID(state.modal.nodeId) === nodeID(id)) {
      state.modal.running = false;
      state.modal.error = error.message;
      renderModal();
    } else toast(error.message, "error");
  }
}

function toast(message, type = "") {
  const root = document.getElementById("toast-root");
  root.replaceChildren();
  const item = document.createElement("div"); item.className = `toast ${type}`; item.textContent = message; root.appendChild(item); setTimeout(() => item.remove(), 3200);
}

async function initialize() {
  await loadBranding();
  render();
  if (state.token) loadAll();
}
initialize();
