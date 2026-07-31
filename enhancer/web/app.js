const translations = {
  zh: {
    title: "Prism DNS 中文增强版", online: "控制器已连接", loginTitle: "登录 Prism Gateway", loginHint: "使用 Controller 管理员账号登录",
    username: "用户名", password: "密码", login: "登录", loggingIn: "正在登录...", services: "服务编排", nodes: "节点管理", ipConfigs: "IP 配置",
    search: "搜索服务或域名", allCategories: "全部分类", dnsClient: "DNS 节点", selectDNS: "请选择 DNS 节点", addService: "新增服务",
    addNode: "新增节点", refresh: "刷新", logout: "退出", totalServices: "服务数量", configured: "已配置", proxyNodes: "解锁机",
    customServices: "自定义服务", notConfigured: "未配置", auto: "自动选择", manual: "手动", open: "配置", domains: "域名",
    serviceConfig: "配置服务", currentRoute: "当前路由", targetServer: "目标解锁机", assign: "保存并切换", reset: "恢复自动",
    connectivity: "DNS 解锁检测", testing: "正在检测...", triggerUnlock: "运行解锁检测", customEdit: "编辑服务", customDelete: "删除服务",
    serviceName: "服务名称", category: "分类", domainList: "域名列表", viewDomains: "查看/编辑域名", restoreDomains: "恢复默认域名", domainOverrideHint: "已保存自定义域名覆盖；恢复后使用域名库当前值。", save: "保存", cancel: "取消", deleteConfirm: "确认删除这个自定义服务？",
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
    actualAudit: "目标机实测", auditPending: "待实测", targetAvailable: "目标机可用", targetUncertain: "探测待确认", targetProblem: "目标机异常", targetCompatibility: "目标机兼容性",
    actualAvailable: "实测可用", actualProblem: "异常或波动", referenceOnly: "仅节点自检", agentReference: "节点自检（仅参考）", noTargetAudit: "尚无目标机实测",
    accountSettings: "账户安全", accountHint: "验证旧账号后修改管理员用户名和密码。", oldUsername: "旧用户名", oldPassword: "旧密码", newUsername: "新用户名", newPassword: "新密码", confirmPassword: "确认新密码", updateAccount: "更新账户", credentialsInvalid: "旧用户名或旧密码不正确", passwordMismatch: "两次输入的新密码不一致", accountUpdated: "账户已更新，请使用新账号重新登录",
    siteSettings: "站点设置", siteSettingsHint: "自定义左上角产品名称、说明文字和浏览器标签标题。", siteName: "网页名称", browserTitle: "页面标签名称", siteTagline: "网页说明", saveBranding: "保存站点设置", brandingUpdated: "站点名称已更新",
    manageCategories: "分类管理", categoryHint: "服务分类只影响整理和筛选，不会修改域名、路由或客户端配置。", newCategory: "新建分类", categoryName: "分类名称", categoryCreated: "分类已创建", categoryDeleted: "分类已删除", editCategory: "分类", serviceCategory: "调整服务分类", restoreCategory: "恢复原分类", originalCategory: "原始分类", useCategory: "使用", builtInCategory: "内置", customCategory: "自定义", categoryInUse: "该分类仍有服务，请先移动这些服务", categoryDeleteConfirm: "删除这个空分类？",
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
    serviceName: "Service name", category: "Category", domainList: "Domain list", viewDomains: "View/edit domains", restoreDomains: "Restore default domains", domainOverrideHint: "A custom domain override is active. Restore to use the current catalog values.", save: "Save", cancel: "Cancel", deleteConfirm: "Delete this custom service?",
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
    actualAudit: "Target audit", auditPending: "Not audited", targetAvailable: "Target passed", targetUncertain: "Probe inconclusive", targetProblem: "Target issue", targetCompatibility: "Target compatibility",
    actualAvailable: "Verified passed", actualProblem: "Failed or unstable", referenceOnly: "Agent only", agentReference: "Agent self-check (reference only)", noTargetAudit: "No target audit yet",
    accountSettings: "Account security", accountHint: "Verify the current credentials before changing the administrator username and password.", oldUsername: "Current username", oldPassword: "Current password", newUsername: "New username", newPassword: "New password", confirmPassword: "Confirm new password", updateAccount: "Update account", credentialsInvalid: "Current username or password is incorrect", passwordMismatch: "The new passwords do not match", accountUpdated: "Account updated. Sign in with the new credentials.",
    siteSettings: "Site settings", siteSettingsHint: "Customize the product name, supporting text, and browser tab title.", siteName: "Website name", browserTitle: "Browser tab title", siteTagline: "Website description", saveBranding: "Save site settings", brandingUpdated: "Site branding updated",
    manageCategories: "Manage categories", categoryHint: "Categories only organize and filter services. Domains, routes, and client configuration remain unchanged.", newCategory: "New category", categoryName: "Category name", categoryCreated: "Category created", categoryDeleted: "Category deleted", editCategory: "Category", serviceCategory: "Change service category", restoreCategory: "Restore original", originalCategory: "Original category", useCategory: "Use", builtInCategory: "Built-in", customCategory: "Custom", categoryInUse: "This category still contains services. Move them first.", categoryDeleteConfirm: "Delete this empty category?",
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
  tab: localStorage.getItem("enhancer_tab") || "services", nodes: [], rules: [], catalog: [], categories: [], customCategories: [], catalogMeta: {}, ipConfigs: [], dnsNodeId: localStorage.getItem("enhancer_dns") || "",
  overrides: {}, activeSelections: {}, search: "", category: "", statusFilter: "", nodeFilter: "", page: 1, pageSize: 20, selectedServiceIds: new Set(),
  loading: false, modal: null, testResults: null, scrollPositions: {}, backgroundPending: false, lastBackgroundSync: "",
  branding: {site_name:"", browser_title:"", site_tagline:""}
};

function t(key) { return translations[state.lang][key] || key; }
function escapeHTML(value = "") { return String(value).replace(/[&<>'"]/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"}[char])); }
function siteName() { return String(state.branding.site_name || "Prism DNS"); }
const iconAssetVersion = "1.5.7-final";

function browserTitle() { return String(state.branding.browser_title || t("title")); }
function siteTagline() { return String(state.branding.site_tagline || (state.lang === "zh" ? "全局解锁编排" : "Global orchestration")); }
function markLoadedServiceIcons(root = document) {
  root.querySelectorAll?.(".service-icon img").forEach(image => {
    image.parentElement?.classList.toggle("loaded", image.complete && image.naturalWidth > 0);
  });
}
document.addEventListener("load", event => {
  if (event.target instanceof HTMLImageElement && event.target.matches(".service-icon img")) {
    event.target.parentElement?.classList.add("loaded");
  }
}, true);
document.addEventListener("error", event => {
  if (event.target instanceof HTMLImageElement && event.target.matches(".service-icon img")) {
    event.target.parentElement?.classList.remove("loaded");
  }
}, true);
new MutationObserver(() => requestAnimationFrame(() => markLoadedServiceIcons()))
  .observe(document.documentElement, {childList:true, subtree:true});

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
  const ids = serviceIDs(service);
  return state.rules.find(rule => ids.some(id => String(rule.value || "").includes(`/enhancer/rules/${id}.list`))) ||
    state.rules.find(rule => String(rule.name || "").toLowerCase() === `stream · ${service.name}`.toLowerCase());
}
function serviceIDs(service) {
  return [...new Set([service?.id, ...(service?.aliases || [])].filter(Boolean))];
}
function serviceRouteEntry(config, service) {
  if (!config?.routes || !service) return null;
  for (const id of serviceIDs(service)) {
    if (config.routes[id]) return {id, value:config.routes[id]};
  }
  return null;
}
function serviceResult(config, service) {
  if (!config?.service_results || !service) return "";
  for (const id of serviceIDs(service)) {
    if (config.service_results[id]) return config.service_results[id];
  }
  return "";
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
function routedNodeForService(service) {
  const targetConfig = configForNode(selectedDNS());
  const proxyId = nodeID(serviceRouteEntry(targetConfig, service)?.value);
  if (proxyId) return state.nodes.find(node => nodeID(node.id) === proxyId) || null;
  return activeNodeFor(serviceRule(service));
}
function serviceStatus(service, node) {
  const targetConfig = configForNode(selectedDNS());
  const targetResult = serviceResult(targetConfig, service);
  if (targetResult && nodeID(serviceRouteEntry(targetConfig, service)?.value) === nodeID(node?.id)) {
    const audit = auditResultState(targetResult);
    return {kind:audit.kind, label:audit.kind === "good" ? t("targetAvailable") : audit.kind === "warn" ? t("targetUncertain") : t("targetProblem"), raw:targetResult};
  }
  const key = serviceDetectorKey(service);
  const value = key ? parseUnlock(node)[key] || "" : "";
  if (value) return {kind:"warn", label:t("referenceOnly"), raw:value};
  return {kind:"", label:t("unknown"), raw:""};
}
function serviceIconHTML(service) {
  const name = displayServiceName(service);
  const initial = Array.from(String(name || "P").trim())[0] || "P";
  const iconDomain = serviceRouteDomains(service)[0] || service.id;
  return `<span class="service-icon" data-initial="${escapeHTML(initial)}"><span class="service-icon-fallback" aria-hidden="true">${escapeHTML(initial)}</span><img src="/enhancer/icons/${encodeURIComponent(service.id)}.png?domain=${encodeURIComponent(iconDomain)}&v=${iconAssetVersion}" alt="${escapeHTML(name)}" title="${escapeHTML(name)}" loading="eager" decoding="async" fetchpriority="high"></span>`;
}
function formatDate(value) { if (!value) return "-"; try { return new Date(value).toLocaleString(state.lang === "zh" ? "zh-CN" : "en-US"); } catch { return value; } }
function formatRelativeDate(value) {
  if (!value) return state.lang === "zh" ? "待检测" : "Pending";
  const milliseconds = Date.now() - new Date(value).getTime();
  if (!Number.isFinite(milliseconds) || milliseconds < 0) return formatDate(value);
  const minutes = Math.floor(milliseconds / 60000);
  if (minutes < 1) return state.lang === "zh" ? "刚刚" : "Just now";
  if (minutes < 60) return state.lang === "zh" ? `${minutes} 分钟前` : `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return state.lang === "zh" ? `${hours} 小时前` : `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return state.lang === "zh" ? `${days} 天前` : `${days}d ago`;
}
function formatBytes(value) { const bytes = Number(value) || 0; if (bytes < 1024) return `${bytes} B`; const units = ["KB", "MB", "GB", "TB"]; let size = bytes; let unit = -1; do { size /= 1024; unit++; } while (size >= 1024 && unit < units.length - 1); return `${size >= 100 ? size.toFixed(0) : size.toFixed(2)} ${units[unit]}`; }
function displayCategory(value) { return state.lang === "zh" ? categoryNames[value] || value : value; }
function displayServiceName(service) { return state.lang === "zh" ? commonNames[service.name] || service.name : service.name; }
function serviceRouteDomains(service) {
  return [...new Set((service?.domains || []).flatMap(value => String(value || "").split(/[\s,]+/)).map(value => value.trim().toLowerCase().replace(/^\*\./, "")).filter(Boolean))];
}
function routeDomainsOverlap(left, right) {
  return left.some(leftDomain => right.some(rightDomain => leftDomain === rightDomain || leftDomain.endsWith(`.${rightDomain}`) || rightDomain.endsWith(`.${leftDomain}`)));
}
function linkedRouteServiceIds(serviceId, routes) {
  const selectedIds = new Set([...Object.keys(routes || {}).filter(id => routes[id]), serviceId]);
  const services = new Map(state.catalog
    .filter(service => serviceIDs(service).some(id => selectedIds.has(id)))
    .map(service => [service.id, service]));
  if (!services.has(serviceId)) return [serviceId];
  const linked = [serviceId];
  const seen = new Set(linked);
  for (let index = 0; index < linked.length; index++) {
    const currentDomains = serviceRouteDomains(services.get(linked[index]));
    services.forEach((service, candidateId) => {
      if (!seen.has(candidateId) && routeDomainsOverlap(currentDomains, serviceRouteDomains(service))) {
        seen.add(candidateId);
        linked.push(candidateId);
      }
    });
  }
  return linked.sort();
}
function applyLinkedRouteProxy(serviceId, proxyId, announce = false) {
  const linkedIds = linkedRouteServiceIds(serviceId, state.modal?.routes || {});
  linkedIds.forEach(linkedId => {
    if (!state.modal.routes[linkedId]) return;
    state.modal.routes[linkedId] = proxyId;
    const select = document.querySelector(`.ip-service-proxy[data-service-id="${CSS.escape(linkedId)}"]`);
    if (select) select.value = proxyId;
  });
  if (announce && linkedIds.length > 1) {
    toast(state.lang === "zh" ? `${linkedIds.length} 个共享域名服务已联动到同一解锁机` : `${linkedIds.length} overlapping-domain services now share one proxy`, "warn");
  }
}
function normalizeSearchText(value) {
  return String(value || "").normalize("NFKC").toLocaleLowerCase(state.lang === "zh" ? "zh-CN" : "en-US").replace(/\s+/g, " ").trim();
}
function compactSearchText(value) {
  return normalizeSearchText(value).replace(/[\s./_+&:·-]+/g, "");
}
function serviceSearchFields(service) {
  const domains = Array.isArray(service.domains) ? service.domains : [];
  return [
    service.name,
    displayServiceName(service),
    service.id,
    service.category,
    displayCategory(service.category),
    serviceDetectorKey(service),
    ...domains,
    ...domains.map(domain => String(domain).replace(/^www\./i, ""))
  ].filter(Boolean);
}
function searchTokenMatches(field, token) {
  if (/^[a-z0-9]{1,2}$/.test(token)) {
    return field.split(/[^\p{L}\p{N}]+/u).includes(token);
  }
  return field.includes(token);
}
function serviceMatches(service, rawQuery) {
  const query = normalizeSearchText(rawQuery);
  if (!query) return true;
  const fields = serviceSearchFields(service).map(normalizeSearchText);
  if (fields.some(field => field === query)) return true;
  const tokens = query.split(" ").filter(Boolean);
  if (tokens.every(token => fields.some(field => searchTokenMatches(field, token)))) return true;
  const compactQuery = compactSearchText(query);
  return compactQuery.length >= 2 && fields.some(field => compactSearchText(field).includes(compactQuery));
}
function catalogCategories() {
  return [...new Set([...(state.categories || []), ...state.catalog.map(service => service.category)].filter(Boolean))]
    .sort((left, right) => displayCategory(left).localeCompare(displayCategory(right), state.lang === "zh" ? "zh-CN" : "en"));
}
function isOnline(node) { if (!node?.last_heartbeat) return false; return Date.now() - new Date(node.last_heartbeat).getTime() < 90000; }
function normalizeIP(value) {
  let candidate = String(value || "").split(",")[0].trim();
  if (candidate.startsWith("[")) candidate = candidate.slice(1, candidate.indexOf("]"));
  if ((candidate.match(/:/g) || []).length === 1 && candidate.includes(".")) candidate = candidate.split(":")[0];
  return candidate.toLowerCase();
}
function configForNode(node) {
  if (!node) return null;
  const byNode = state.ipConfigs.find(config => nodeID(config.dns_node_id) === nodeID(node.id));
  if (byNode) return byNode;
  const address = normalizeIP(node.public_ip || node.address);
  return address ? state.ipConfigs.find(config => normalizeIP(config.ip) === address) || null : null;
}
function clientState(config, node) {
  if (!isOnline(node)) return {kind:"bad", label:"OFFLINE", detail:state.lang === "zh" ? "控制通道离线" : "Control channel offline"};
  if (!config) return {kind:"warn", label:state.lang === "zh" ? "未纳管" : "UNMANAGED", detail:state.lang === "zh" ? "Agent 已在线，尚未配置解锁服务" : "Agent connected; unlock services are not configured"};
  if (!config.health_updated_at) return {kind:"warn", label:t("pending"), detail:state.lang === "zh" ? "等待客户端健康上报" : "Waiting for client health report"};
  if (Date.now() - new Date(config.health_updated_at).getTime() > 180000) return {kind:"warn", label:t("stale"), detail:state.lang === "zh" ? "健康上报已超时" : "Health report is stale"};
  const routes = `${Number(config.healthy_routes || 0)}/${Number(config.expected_routes || 0)}`;
  if (config.dns_ready && config.system_dns_ready && config.routes_ready) return {kind:"good", label:t("ready"), detail:`${t("healthRoutes")} ${routes}`};
  return {kind:"bad", label:t("degraded"), detail:config.health_message || `${t("healthRoutes")} ${routes}`};
}

function viewFingerprint() {
  const nodes = state.nodes.map(node => ({
    id:nodeID(node.id), role:node.role, name:node.name, address:node.address, public_ip:node.public_ip,
    country:node.country, group:node.group, priority:node.priority, online:isOnline(node),
    unlock:state.tab === "nodes" || state.tab === "audit" || state.tab === "alerts" ? parseUnlock(node) : undefined
  }));
  const configs = state.ipConfigs.map(config => ({
    id:config.id, ip:config.ip, note:config.note, dns_node_id:nodeID(config.dns_node_id), routes:config.routes,
    dns_ready:config.dns_ready, system_dns_ready:config.system_dns_ready, routes_ready:config.routes_ready,
    healthy_routes:config.healthy_routes, expected_routes:config.expected_routes, health_message:config.health_message,
    service_results:config.service_results, service_audited_at:config.service_audited_at,
    traffic_rx_bytes:state.tab === "ips" ? config.traffic_rx_bytes : undefined,
    traffic_tx_bytes:state.tab === "ips" ? config.traffic_tx_bytes : undefined
  }));
  return JSON.stringify({
    tab:state.tab,
    nodes,
    configs,
    rules:state.rules,
    overrides:state.overrides,
    activeSelections:state.activeSelections,
    catalog:state.catalog.map(service => ({id:service.id, aliases:service.aliases, name:service.name, category:service.category, domains:service.domains, custom:service.custom})),
    categories:state.categories,
    branding:state.branding
  });
}

function updateBackgroundIndicator() {
  const indicator = document.getElementById("background-sync-state");
  if (!indicator) return;
  indicator.classList.toggle("pending", !!state.backgroundPending);
  const label = indicator.querySelector("span:last-child");
  if (label) {
    label.textContent = state.backgroundPending
      ? (state.lang === "zh" ? "有后台更新" : "Background update")
      : (state.lang === "zh" ? "数据已同步" : "Data synced");
  }
}
function auditResultState(result) {
  const value = String(result || "").trim();
  if (!value) return {kind:"", label:t("auditPending"), detail:""};
  if (/^YES\b|^PASS\b/i.test(value)) {
    return {kind:"good", label:state.lang === "zh" ? "完整应用链实测通过" : "Full application path verified", detail:value};
  }
  const ratio = value.match(/(\d+)\s*\/\s*(\d+)(?:\s*YES)?/i);
  if (/INCONCLUSIVE|DEGRADED|UNSTABLE|Banned|WAF/i.test(value) && ratio) {
    const passed = Number(ratio[1]); const total = Number(ratio[2]);
    const wafCount = /WAF|Banned/i.test(value) ? Math.max(1, total - passed) : 0;
    const suffix = wafCount
      ? (state.lang === "zh" ? `，${wafCount} 次被 WAF 拒绝` : `, ${wafCount} WAF rejection${wafCount > 1 ? "s" : ""}`)
      : "";
    return {
      kind:"bad",
      label:state.lang === "zh"
        ? `不可用（${total} 次仅 ${passed} 次成功${suffix}）`
        : `Unavailable (${passed} of ${total} passed${suffix})`,
      detail:value
    };
  }
  if (/INCONCLUSIVE|DEGRADED|UNSTABLE|Banned|WAF/i.test(value)) {
    return {kind:"bad", label:state.lang === "zh" ? "不可用（真实检测未通过）" : "Unavailable (real check failed)", detail:value};
  }
  if (/Restricted|Partial|Error|Failed|FAIL|N\/A|Unavailable|No result|^NO\b/i.test(value)) {
    return {kind:"bad", label:state.lang === "zh" ? "不可用" : "Unavailable", detail:value};
  }
  return {kind:"bad", label:state.lang === "zh" ? "不可用" : "Unavailable", detail:value};
}

function targetAuditObservations(proxyId, service) {
  return state.ipConfigs
    .filter(config => nodeID(serviceRouteEntry(config, service)?.value) === nodeID(proxyId))
    .map(config => ({ip:config.ip, result:String(serviceResult(config, service) || "").trim()}));
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
  const domains = service.domains.map(domain => domain.replace(/^\*\./, "").replace(/^\./, ""));
  return (preferred[service.name] || []).find(candidate =>
    domains.some(domain => candidate === domain || candidate.endsWith(`.${domain}`))
  ) || domains[0];
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
  const previousFingerprint = viewFingerprint();
  state.loading = !silent;
  if (!silent) render();
  try {
    const [nodes, rules, catalog, ipConfigs, categoryData] = await Promise.all([api("/enhancer/api/nodes"), api("/api/rules"), api("/enhancer/api/catalog"), api("/enhancer/api/ip-configs"), api("/enhancer/api/categories")]);
    state.nodes = Array.isArray(nodes) ? nodes : [];
    state.rules = Array.isArray(rules) ? rules : [];
    state.catalog = catalog.services || [];
    state.categories = Array.isArray(categoryData?.categories) ? categoryData.categories : (catalog.categories || []);
    state.customCategories = Array.isArray(categoryData?.custom_categories) ? categoryData.custom_categories : [];
    state.catalogMeta = catalog;
    state.ipConfigs = Array.isArray(ipConfigs) ? ipConfigs : [];
    if (!dnsNodes().some(node => nodeID(node.id) === nodeID(state.dnsNodeId))) state.dnsNodeId = dnsNodes()[0]?.id || "";
    localStorage.setItem("enhancer_dns", nodeID(state.dnsNodeId));
    await loadRoutingState();
    connectSSE();
  } catch (error) { toast(error.message, "error"); }
  finally {
    state.loading = false;
    const preserveDraft = silent && ["service-form", "service-category", "category-manager", "node-form", "ip-form"].includes(state.modal?.type);
    const changed = previousFingerprint !== viewFingerprint();
    state.lastBackgroundSync = new Date().toISOString();
    state.backgroundPending = preserveDraft && changed;
    if (!preserveDraft && (!silent || changed)) render();
    else updateBackgroundIndicator();
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
let searchRenderTimer;
let searchComposing = false;
let clockTimer;
function scheduleLiveReload() {
  if (reloadTimer) return;
  reloadTimer = setTimeout(() => {
    reloadTimer = null;
    const active = document.activeElement;
    if (state.modal || active?.matches?.("input, textarea, select, [contenteditable=true]")) {
      state.backgroundPending = true;
      updateBackgroundIndicator();
      scheduleLiveReload();
      return;
    }
    loadAll(true);
  }, 5000);
}
function connectSSE() {
  if (eventSource || !state.token) return;
  eventSource = new EventSource(`/api/sse?token=${encodeURIComponent(state.token)}`);
  eventSource.onmessage = scheduleLiveReload;
  eventSource.onerror = () => { eventSource.close(); eventSource = null; setTimeout(connectSSE, 3500); };
}

function updateClock() {
  const element = document.getElementById("topbar-clock");
  if (!element) return;
  const now = new Date();
  const date = now.toLocaleDateString(state.lang === "zh" ? "zh-CN" : "en-US", {year:"numeric", month:"2-digit", day:"2-digit", weekday:"short"});
  const time = now.toLocaleTimeString(state.lang === "zh" ? "zh-CN" : "en-US", {hour12:false});
  element.textContent = `${date}  ${time}`;
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
  updateBackgroundIndicator();
}

function loginHTML() {
  const loginTitle = state.lang === "zh" ? `登录 ${siteName()}` : `Sign in to ${siteName()}`;
  return `<main class="login-wrap"><section class="login-box panel"><h1>${escapeHTML(loginTitle)}</h1><p>${escapeHTML(t("loginHint"))}</p>
    <form class="form-stack" id="login-form"><div class="field"><label>${t("username")}</label><input class="input" name="username" autocomplete="username" required></div>
    <div class="field"><label>${t("password")}</label><input class="input" type="password" name="password" autocomplete="current-password" required></div>
    <div class="form-error"></div><button class="btn primary" type="submit">${t("login")}</button></form></section></main>`;
}
function bindLogin() { document.getElementById("login-form").addEventListener("submit", login); }

function activeContentHTML() {
  if (state.tab === "services") return servicesHTML();
  if (state.tab === "nodes") return nodesHTML();
  if (state.tab === "ips") return ipConfigsHTML();
  if (state.tab === "sync") return syncHTML();
  if (state.tab === "audit") return auditHTML();
  if (state.tab === "alerts") return alertsHTML();
  return settingsHTML();
}

function pageMeta() {
  const pages = {
    services: ["服务编排", "Service Routing", "配置真实可用的 DNS 解锁线路"],
    nodes: ["节点管理", "Nodes", "管理解锁机与被解锁机"],
    ips: ["IP 配置", "IP Configs", "按目标 IP 管理服务与流量"],
    sync: ["同步域名名单", "Domain Catalog", "更新和检查细化域名库"],
    audit: ["日志与审计", "Logs & Audits", "查看目标机真实复测结果"],
    alerts: ["告警中心", "Alerts", "集中处理当前异常与待确认项目"],
    settings: ["设置中心", "Settings", "管理站点、账户与服务分类"]
  };
  const current = pages[state.tab] || pages.services;
  return {title:state.lang === "zh" ? current[0] : current[1], hint:current[2]};
}

function navigationHTML() {
  const items = [
    ["services", "bi-grid-fill", state.lang === "zh" ? "服务编排" : "Service Routing"],
    ["nodes", "bi-diagram-3", state.lang === "zh" ? "节点管理" : "Nodes"],
    ["ips", "bi-globe2", state.lang === "zh" ? "IP 配置" : "IP Configs"],
    ["sync", "bi-arrow-repeat", state.lang === "zh" ? "同步域名名单" : "Domain Catalog"],
    ["audit", "bi-journal-text", state.lang === "zh" ? "日志与审计" : "Logs & Audits"],
    ["alerts", "bi-bell", state.lang === "zh" ? "告警中心" : "Alerts"],
    ["settings", "bi-gear", state.lang === "zh" ? "设置中心" : "Settings"]
  ];
  return items.map(([id, icon, label]) => `<button type="button" class="tab ${state.tab === id ? "active" : ""}" data-tab="${id}"><i class="bi ${icon}" aria-hidden="true"></i><span>${escapeHTML(label)}</span></button>`).join("");
}
function selectTab(tab) {
  if (!["services", "nodes", "ips", "sync", "audit", "alerts", "settings"].includes(tab)) return;
  if (state.tab === tab && document.querySelector(".container")?.dataset.view === tab) return;
  state.tab = tab;
  state.page = 1;
  localStorage.setItem("enhancer_tab", state.tab);
  render();
}

let navigationEventsBound = false;
function bindNavigation() {
  if (navigationEventsBound) return;
  const activate = event => {
    const target = event.target;
    const button = target instanceof Element ? target.closest(".tab[data-tab]") : null;
    if (!button) return;
    event.preventDefault();
    event.stopPropagation();
    selectTab(button.dataset.tab);
  };
  document.addEventListener("click", activate, true);
  document.addEventListener("pointerup", activate, true);
  document.addEventListener("keydown", event => {
    if (event.key !== "Enter" && event.key !== " ") return;
    const target = event.target;
    const button = target instanceof Element ? target.closest(".tab[data-tab]") : null;
    if (!button) return;
    event.preventDefault();
    event.stopPropagation();
    selectTab(button.dataset.tab);
  }, true);
  navigationEventsBound = true;
}

function shellHTML() {
  const meta = pageMeta();
  return `<div class="shell"><aside class="sidebar"><div class="sidebar-brand"><span class="brand-mark"><i class="bi bi-triangle-fill" aria-hidden="true"></i></span><div><strong>${escapeHTML(siteName())}</strong><span>${escapeHTML(siteTagline())}</span></div></div>
    <nav class="tabs" aria-label="${escapeHTML(browserTitle())}">${navigationHTML()}</nav>
    <div class="sidebar-footer"><button class="user-button account-settings-trigger" id="account-settings" title="${t("accountSettings")}"><span class="user-avatar">${escapeHTML(Array.from(state.user.username || "A")[0] || "A")}</span><span class="user-copy"><strong>${escapeHTML(state.user.username || "admin")}</strong><small>${state.lang === "zh" ? "超级管理员" : "Administrator"}</small></span><i class="bi bi-chevron-up" aria-hidden="true"></i></button>
    <div class="sidebar-tools"><button class="btn small theme-toggle-trigger" id="theme-toggle" title="${t("theme")}"><i class="bi ${state.theme === "light" ? "bi-moon" : "bi-sun"}" aria-hidden="true"></i></button><button class="btn small lang-toggle-trigger" id="lang-toggle">${t("language")}</button><button class="btn small danger logout-trigger" id="logout">${t("logout")}</button></div></div></aside>
    <section class="workspace"><header class="topbar"><div class="page-heading"><h1>${escapeHTML(meta.title)}</h1><p>${escapeHTML(meta.hint)}</p></div><div class="toolbar-right"><span class="topbar-clock" id="topbar-clock"></span><span class="topbar-divider"></span><span class="live-state authorization-state" title="${state.lang === "zh" ? "只有面板授权目标 IP 可使用 DNS 解锁" : "Only panel-authorized target IPs may use unlock DNS"}"><span class="status-dot"></span>${state.lang === "zh" ? "授权白名单" : "Allowlist"}</span><span class="live-state" id="background-sync-state"><span class="status-dot"></span><span>${state.lang === "zh" ? "数据已同步" : "Data synced"}</span></span><button class="btn icon" id="refresh" title="${t("refresh")}"><i class="bi bi-arrow-clockwise" aria-hidden="true"></i></button><button class="btn icon theme-toggle-trigger" title="${t("theme")}"><i class="bi ${state.theme === "light" ? "bi-moon" : "bi-sun"}" aria-hidden="true"></i></button><button class="btn lang-toggle-trigger"><i class="bi bi-translate" aria-hidden="true"></i>${state.lang === "zh" ? "中文" : "EN"}</button><div class="mobile-controls"><button class="btn small account-settings-trigger">${escapeHTML(state.user.username || "admin")}</button><button class="btn small danger logout-trigger">${t("logout")}</button></div></div></header>
    <main class="container" data-view="${escapeHTML(state.tab)}">${state.loading ? `<div class="loading panel"><div class="spinner"></div></div>` : activeContentHTML()}</main></section></div>`;
}

function servicesHTML() {
  const dns = dnsNodes(); const proxies = proxyNodes();
  const categories = catalogCategories();
  const targetNode = selectedDNS();
  const targetConfig = configForNode(targetNode);
  const totalTraffic = state.ipConfigs.reduce((total, config) => total + Number(config.traffic_rx_bytes || 0) + Number(config.traffic_tx_bytes || 0), 0);
  const configuredServices = state.catalog.filter(service => {
    return !!routedNodeForService(service);
  });
  const configured = configuredServices.length;
  const targetHealth = targetNode ? clientState(targetConfig, targetNode) : {kind:"warn", label:t("pending"), detail:t("selectDNS")};
  const targetIP = targetConfig?.ip || targetNode?.public_ip || targetNode?.address || "-";
  const filtered = state.catalog.filter(service => {
    if (state.category && service.category !== state.category) return false;
    if (!serviceMatches(service, state.search)) return false;
    const node = routedNodeForService(service);
    const status = serviceStatus(service, node);
    if (state.nodeFilter && nodeID(node?.id) !== nodeID(state.nodeFilter)) return false;
    if (state.statusFilter === "configured" && !node) return false;
    if (state.statusFilter === "available" && status.kind !== "good") return false;
    if (state.statusFilter === "issue" && !["warn", "bad"].includes(status.kind)) return false;
    if (state.statusFilter === "unconfigured" && node) return false;
    return true;
  });
  const pageCount = Math.max(1, Math.ceil(filtered.length / state.pageSize));
  state.page = Math.min(Math.max(1, state.page), pageCount);
  const pageStart = (state.page - 1) * state.pageSize;
  const pageServices = filtered.slice(pageStart, pageStart + state.pageSize);
  return `${dns.length === 0 ? `<div class="panel empty"><strong>${t("noDNS")}</strong><button class="btn primary" id="empty-add-node"><i class="bi bi-plus-circle"></i>${t("addNode")}</button></div>` : ""}
    <section class="dashboard-summary panel">
      <div class="dashboard-stat"><span class="stat-icon blue"><i class="bi bi-grid-fill"></i></span><div><span>${state.lang === "zh" ? "活跃服务" : "Active services"}</span><strong>${configured}</strong><small>${state.lang === "zh" ? `总计 ${state.catalog.length}` : `${state.catalog.length} total`}</small></div></div>
      <div class="dashboard-stat"><span class="stat-icon cyan"><i class="bi bi-server"></i></span><div><span>${state.lang === "zh" ? "DNS 节点" : "DNS nodes"}</span><strong>${dns.length}</strong><small class="good-copy"><i class="bi bi-circle-fill"></i>${dns.filter(isOnline).length}/${dns.length} ${state.lang === "zh" ? "在线" : "online"}</small></div></div>
      <div class="dashboard-stat"><span class="stat-icon violet"><i class="bi bi-shield-check"></i></span><div><span>${t("targetIP")}</span><strong class="summary-ip">${escapeHTML(targetIP)}</strong><small>${escapeHTML(targetNode?.country || targetNode?.name || "-")}</small></div></div>
      <div class="dashboard-stat"><span class="stat-icon green"><i class="bi bi-activity"></i></span><div><span>${t("totalTraffic")}</span><strong>${formatBytes(totalTraffic)}</strong><small class="${targetHealth.kind}-copy">${escapeHTML(targetHealth.label)}</small></div></div>
    </section>
    <section class="service-dashboard panel">
      <div class="service-toolbar">
        <label class="search-control"><i class="bi bi-search"></i><input id="service-search" value="${escapeHTML(state.search)}" placeholder="${t("search")}"></label>
        <select class="select compact-select" id="category-filter"><option value="">${t("allCategories")}</option>${categories.map(category => `<option value="${escapeHTML(category)}" ${state.category === category ? "selected" : ""}>${escapeHTML(displayCategory(category))}</option>`).join("")}</select>
        <select class="select compact-select" id="status-filter"><option value="">${state.lang === "zh" ? "全部状态" : "All statuses"}</option><option value="available" ${state.statusFilter === "available" ? "selected" : ""}>${state.lang === "zh" ? "实测可用" : "Verified"}</option><option value="issue" ${state.statusFilter === "issue" ? "selected" : ""}>${state.lang === "zh" ? "异常/待确认" : "Issues"}</option><option value="configured" ${state.statusFilter === "configured" ? "selected" : ""}>${t("configured")}</option><option value="unconfigured" ${state.statusFilter === "unconfigured" ? "selected" : ""}>${t("notConfigured")}</option></select>
        <select class="select compact-select" id="node-filter"><option value="">${state.lang === "zh" ? "全部节点" : "All nodes"}</option>${proxies.map(node => `<option value="${nodeID(node.id)}" ${nodeID(state.nodeFilter) === nodeID(node.id) ? "selected" : ""}>${escapeHTML(node.name)}</option>`).join("")}</select>
        <span class="toolbar-spacer"></span>
        <select class="select compact-select" id="batch-action"><option value="">${state.lang === "zh" ? "批量操作" : "Bulk actions"}</option><option value="select-page">${state.lang === "zh" ? "选择当前页" : "Select page"}</option><option value="clear">${state.lang === "zh" ? "取消选择" : "Clear selection"}</option><option value="configure">${state.lang === "zh" ? "配置所选服务" : "Configure selected"}</option></select>
        <button class="btn" id="clear-filter"><i class="bi bi-arrow-counterclockwise"></i>${state.lang === "zh" ? "重置" : "Reset"}</button>
        <button class="btn" id="manage-categories"><i class="bi bi-tags"></i>${t("manageCategories")}</button>
        <button class="btn primary" id="add-service"><i class="bi bi-plus-circle"></i>${t("addService")}</button>
      </div>
      <div class="target-strip"><div><span>${state.lang === "zh" ? "当前目标 IP" : "Current target"}</span><strong>${escapeHTML(targetIP)}</strong><span class="region-pill">${escapeHTML(targetNode?.country || "-")}</span><span>${escapeHTML(targetNode?.name || "-")}</span></div><label><span>${state.lang === "zh" ? "更换" : "Switch"}</span><select class="select" id="dns-select"><option value="">${t("selectDNS")}</option>${dns.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(state.dnsNodeId) ? "selected" : ""}>${escapeHTML(node.name)} · ${escapeHTML(node.public_ip || node.address || "")}</option>`).join("")}</select></label></div>
      <div class="service-table-wrap"><table class="service-table"><thead><tr><th><input type="checkbox" id="select-page-services" aria-label="${state.lang === "zh" ? "选择当前页" : "Select page"}"></th><th>${state.lang === "zh" ? "服务名称" : "Service"}</th><th>${state.lang === "zh" ? "状态" : "Status"}</th><th>${state.lang === "zh" ? "解析结果" : "Audit"}</th><th>${state.lang === "zh" ? "路由节点" : "Route node"}</th><th>${state.lang === "zh" ? "上次检测" : "Last test"}</th><th>${state.lang === "zh" ? "操作" : "Actions"}</th></tr></thead><tbody>${pageServices.map(serviceTableRowHTML).join("")}</tbody></table>${pageServices.length ? "" : `<div class="empty"><strong>${t("noServices")}</strong></div>`}</div>
      <footer class="table-footer"><div><strong>${filtered.length}</strong> ${state.lang === "zh" ? "项" : "items"}${state.selectedServiceIds.size ? `<span> · ${state.lang === "zh" ? "已选择" : "selected"} ${state.selectedServiceIds.size}</span>` : ""}</div><div class="pagination"><select class="select" id="page-size"><option value="10" ${state.pageSize === 10 ? "selected" : ""}>10 / ${state.lang === "zh" ? "页" : "page"}</option><option value="20" ${state.pageSize === 20 ? "selected" : ""}>20 / ${state.lang === "zh" ? "页" : "page"}</option><option value="50" ${state.pageSize === 50 ? "selected" : ""}>50 / ${state.lang === "zh" ? "页" : "page"}</option></select><button class="btn icon page-prev" ${state.page <= 1 ? "disabled" : ""}><i class="bi bi-chevron-left"></i></button><span>${state.page} / ${pageCount}</span><button class="btn icon page-next" ${state.page >= pageCount ? "disabled" : ""}><i class="bi bi-chevron-right"></i></button></div></footer>
    </section>`;
}

function serviceTableRowHTML(service) {
  const node = routedNodeForService(service);
  const status = serviceStatus(service, node);
  const targetConfig = configForNode(selectedDNS());
  const result = serviceResult(targetConfig, service);
  const configured = !!node;
  const checked = state.selectedServiceIds.has(service.id);
  const statusClass = !configured ? "neutral" : status.kind === "warn" ? "warn" : status.kind === "bad" ? "bad" : "";
  const statusLabel = configured ? (status.kind === "bad" ? (state.lang === "zh" ? "异常" : "Issue") : (state.lang === "zh" ? "已配置" : "Configured")) : t("notConfigured");
  const auditLabel = status.kind === "good" ? (state.lang === "zh" ? "成功" : "Passed") : status.kind === "warn" ? (state.lang === "zh" ? "待确认" : "Review") : status.kind === "bad" ? (state.lang === "zh" ? "失败" : "Failed") : t("auditPending");
  return `<tr class="${configured ? "configured" : ""}" data-service-id="${escapeHTML(service.id)}"><td><input class="service-select" type="checkbox" data-service-id="${escapeHTML(service.id)}" ${checked ? "checked" : ""}></td><td><button class="service-identity service-open" data-service-id="${escapeHTML(service.id)}">${serviceIconHTML(service)}<span><strong>${escapeHTML(displayServiceName(service))}</strong><small>${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")}</small></span></button></td><td><span class="table-status"><i class="status-dot ${statusClass}"></i>${escapeHTML(statusLabel)}</span></td><td><span class="result-badge ${status.kind || "neutral"}" title="${escapeHTML(result || status.raw || status.label)}">${escapeHTML(auditLabel)}</span></td><td>${node ? `<div class="route-node-cell"><span class="country-chip">${escapeHTML((node.country || "--").slice(0, 3).toUpperCase())}</span><span><strong>${escapeHTML(node.name)}</strong><small>${escapeHTML(node.country || node.public_ip || node.address || "-")}</small></span></div>` : `<span class="muted-cell">${t("notConfigured")}</span>`}</td><td><span class="last-audit" title="${escapeHTML(formatDate(targetConfig?.service_audited_at))}">${result ? escapeHTML(formatRelativeDate(targetConfig?.service_audited_at)) : t("auditPending")}</span></td><td><div class="table-actions">${configured ? `<button class="btn small icon service-audit-run" data-service-id="${escapeHTML(service.id)}" title="${state.lang === "zh" ? "立即实测" : "Test now"}"><i class="bi bi-play-circle"></i></button>` : ""}<button class="btn small service-open" data-service-id="${escapeHTML(service.id)}">${t("open")}</button><button class="btn small icon service-category-open" data-service-id="${escapeHTML(service.id)}" title="${t("editCategory")}"><i class="bi bi-three-dots"></i></button></div></td></tr>`;
}

function contentHeaderHTML(title, hint, actions = "") {
  return `<header class="content-header"><div><h2>${escapeHTML(title)}</h2><p>${escapeHTML(hint)}</p></div><div>${actions}</div></header>`;
}

function syncHTML() {
  const updatedAt = state.catalogMeta.updated_at || state.catalogMeta.updated || state.catalogMeta.source_updated_at;
  const error = state.catalogMeta.error || "";
  const categoryRows = catalogCategories().map(category => `<div class="compact-row"><span>${escapeHTML(displayCategory(category))}</span><strong>${state.catalog.filter(service => service.category === category).length}</strong></div>`).join("");
  return `${contentHeaderHTML(state.lang === "zh" ? "同步域名名单" : "Domain catalog", state.lang === "zh" ? "从细化域名库更新服务、泛域名与分类，现有目标路由不会被自动切换。" : "Refresh services, wildcards, and categories without changing selected target routes.", `<button class="btn primary" id="refresh-catalog"><i class="bi bi-arrow-repeat"></i>${t("catalogRefresh")}</button>`)}
    <section class="two-column-layout"><article class="panel info-card"><span class="info-card-icon"><i class="bi bi-database-check"></i></span><div><span>${state.lang === "zh" ? "当前服务库" : "Current catalog"}</span><strong>${state.catalog.length}</strong><small>${state.lang === "zh" ? `自定义服务 ${state.catalog.filter(service => service.custom).length} 项` : `${state.catalog.filter(service => service.custom).length} custom services`}</small></div></article><article class="panel info-card ${error ? "danger-card" : ""}"><span class="info-card-icon"><i class="bi ${error ? "bi-exclamation-triangle" : "bi-cloud-check"}"></i></span><div><span>${state.lang === "zh" ? "上次同步" : "Last sync"}</span><strong>${escapeHTML(formatRelativeDate(updatedAt))}</strong><small>${escapeHTML(error || formatDate(updatedAt))}</small></div></article></section>
    <section class="panel settings-card"><div class="settings-card-head"><div><h3>${state.lang === "zh" ? "服务分类概览" : "Category overview"}</h3><p>${state.lang === "zh" ? "同步只更新域名库；你的自定义分类、节点和目标路由保持不变。" : "Sync updates the catalog only. Custom categories, nodes, and target routes remain unchanged."}</p></div><button class="btn" id="manage-categories">${t("manageCategories")}</button></div><div class="compact-list">${categoryRows}</div></section>`;
}

function auditRows() {
  const services = new Map();
  state.catalog.forEach(service => serviceIDs(service).forEach(id => services.set(id, service)));
  return state.ipConfigs.flatMap(config => Object.entries(config.service_results || {}).map(([serviceID, raw]) => {
    const service = services.get(serviceID);
    const audit = auditResultState(raw);
    const proxy = state.nodes.find(node => nodeID(node.id) === nodeID(service ? serviceRouteEntry(config, service)?.value : config.routes?.[serviceID]));
    return {config, service, serviceID, raw, audit, proxy};
  })).sort((left, right) => String(right.config.service_audited_at || "").localeCompare(String(left.config.service_audited_at || "")));
}

function auditHTML() {
  const rows = auditRows();
  return `${contentHeaderHTML(state.lang === "zh" ? "日志与审计" : "Logs & Audits", state.lang === "zh" ? "这里仅展示目标机经过真实 DNS、HTTPS 与服务依赖路径后的结果；节点自检不等同于可用。" : "Only target-machine DNS, HTTPS, and dependency-path results are shown here. Node self-checks are not treated as proof.", `<button class="btn" id="refresh-audits"><i class="bi bi-arrow-clockwise"></i>${t("refresh")}</button>`)}
    <section class="panel audit-table"><div class="audit-table-head"><span>${state.lang === "zh" ? "目标 / 服务" : "Target / service"}</span><span>${state.lang === "zh" ? "路由节点" : "Route node"}</span><span>${state.lang === "zh" ? "真实结果" : "Real result"}</span><span>${state.lang === "zh" ? "检测时间" : "Checked"}</span></div>${rows.length ? rows.map(item => `<article class="audit-row"><div>${item.service ? serviceIconHTML(item.service) : ""}<span><strong>${escapeHTML(item.service ? displayServiceName(item.service) : item.serviceID)}</strong><small>${escapeHTML(item.config.ip)}</small></span></div><div><strong>${escapeHTML(item.proxy?.name || "-")}</strong><small>${escapeHTML(item.proxy?.country || "")}</small></div><div><span class="result-badge ${item.audit.kind}">${escapeHTML(item.audit.label)}</span><small title="${escapeHTML(item.raw)}">${escapeHTML(item.raw)}</small></div><div><strong>${escapeHTML(formatRelativeDate(item.config.service_audited_at))}</strong><small>${escapeHTML(formatDate(item.config.service_audited_at))}</small></div></article>`).join("") : `<div class="empty"><strong>${t("noTargetAudit")}</strong></div>`}</section>`;
}

function currentAlerts() {
  const alerts = [];
  state.nodes.forEach(node => {
    if (!isOnline(node)) alerts.push({kind:"bad", title:`${node.name} · OFFLINE`, detail:state.lang === "zh" ? "控制通道已离线" : "Control channel is offline", tab:"nodes"});
  });
  state.ipConfigs.forEach(config => {
    const node = ipConfigNode(config);
    const health = clientState(config, node);
    if (health.kind !== "good") alerts.push({kind:health.kind, title:`${config.ip} · ${health.label}`, detail:health.detail, tab:"ips"});
    Object.entries(config.service_results || {}).forEach(([serviceID, raw]) => {
      const audit = auditResultState(raw);
      if (audit.kind === "good") return;
      const service = state.catalog.find(item => serviceIDs(item).includes(serviceID));
      alerts.push({kind:audit.kind, title:`${config.ip} · ${service ? displayServiceName(service) : serviceID}`, detail:audit.label, tab:"audit"});
    });
  });
  return alerts;
}

function alertsHTML() {
  const alerts = currentAlerts();
  return `${contentHeaderHTML(state.lang === "zh" ? "告警中心" : "Alerts", state.lang === "zh" ? "异常不会自动切换解锁机，避免线路来回跳转；请在确认后手动调整。" : "Alerts never switch proxies automatically. Review and change routes manually when needed.")}
    <section class="alert-summary"><article class="panel info-card"><span class="info-card-icon"><i class="bi bi-shield-check"></i></span><div><span>${state.lang === "zh" ? "当前异常" : "Current alerts"}</span><strong>${alerts.length}</strong><small>${alerts.length ? (state.lang === "zh" ? "需要人工确认" : "Needs review") : (state.lang === "zh" ? "运行正常" : "All clear")}</small></div></article></section>
    <section class="panel alert-list">${alerts.length ? alerts.map((alert, index) => `<article class="alert-item ${alert.kind}"><span class="alert-symbol"><i class="bi ${alert.kind === "bad" ? "bi-x-circle" : "bi-exclamation-triangle"}"></i></span><div><strong>${escapeHTML(alert.title)}</strong><p>${escapeHTML(alert.detail)}</p></div><button class="btn small alert-open" data-alert-tab="${alert.tab}">${state.lang === "zh" ? "查看" : "View"}</button></article>`).join("") : `<div class="empty"><i class="bi bi-check-circle good-copy"></i><strong>${state.lang === "zh" ? "当前没有告警" : "No current alerts"}</strong></div>`}</section>`;
}

function settingsHTML() {
  return `${contentHeaderHTML(state.lang === "zh" ? "设置中心" : "Settings", state.lang === "zh" ? "管理站点显示、管理员账户、服务分类和外观。" : "Manage branding, account security, categories, and appearance.")}
    <section class="settings-grid">
      <article class="panel settings-card"><span class="settings-icon"><i class="bi bi-window"></i></span><div><h3>${t("siteSettings")}</h3><p>${t("siteSettingsHint")}</p><button class="btn site-settings-trigger">${state.lang === "zh" ? "编辑站点信息" : "Edit branding"}</button></div></article>
      <article class="panel settings-card"><span class="settings-icon"><i class="bi bi-person-lock"></i></span><div><h3>${t("accountSettings")}</h3><p>${t("accountHint")}</p><button class="btn account-settings-trigger">${state.lang === "zh" ? "修改账户密码" : "Update credentials"}</button></div></article>
      <article class="panel settings-card"><span class="settings-icon"><i class="bi bi-tags"></i></span><div><h3>${t("manageCategories")}</h3><p>${t("categoryHint")}</p><button class="btn" id="manage-categories">${state.lang === "zh" ? "管理服务分类" : "Manage categories"}</button></div></article>
      <article class="panel settings-card"><span class="settings-icon"><i class="bi bi-palette"></i></span><div><h3>${state.lang === "zh" ? "外观与语言" : "Appearance & language"}</h3><p>${state.lang === "zh" ? "切换深浅主题与中英文界面。" : "Switch theme and interface language."}</p><div class="inline-actions"><button class="btn theme-toggle-trigger">${state.theme === "light" ? (state.lang === "zh" ? "切换深色" : "Use dark") : (state.lang === "zh" ? "切换浅色" : "Use light")}</button><button class="btn lang-toggle-trigger">${t("language")}</button></div></div></article>
    </section>`;
}

function serviceCardHTML(service) {
  const rule = serviceRule(service); const node = activeNodeFor(rule); const status = serviceStatus(service, node); const manual = !!overrideFor(rule);
  const displayName = displayServiceName(service);
  return `<article class="service-card ${rule ? "configured" : ""} ${service.custom ? "custom" : ""}" data-service-id="${escapeHTML(service.id)}"><div class="service-row-main">${serviceIconHTML(service)}<div><div class="service-name" title="${escapeHTML(displayName)}">${escapeHTML(displayName)}</div><div class="service-category">${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")}</div></div></div>
    <div class="service-row-state"><span class="badge ${status.kind}" title="${escapeHTML(status.raw || status.label)}">${escapeHTML(status.kind ? status.label : (rule ? t("configured") : t("notConfigured")))}</span><button class="btn small service-category-open" data-service-id="${escapeHTML(service.id)}">${t("editCategory")}</button><button class="btn small service-open" data-service-id="${escapeHTML(service.id)}">${t("open")}</button></div></article>`;
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
  const header = contentHeaderHTML(state.lang === "zh" ? "节点管理" : "Nodes", state.lang === "zh" ? "统一管理解锁机和被解锁机；节点状态变化不会自动改写服务路由。" : "Manage proxy and target nodes. Status changes never rewrite service routes automatically.", `<button class="btn primary" id="add-node"><i class="bi bi-plus-circle"></i>${t("addNode")}</button>`);
  if (!state.nodes.length) return `${header}<div class="panel empty"><strong>${state.lang === "zh" ? "还没有节点" : "No nodes"}</strong><button class="btn primary" id="empty-add-node"><i class="bi bi-plus-circle"></i>${t("addNode")}</button></div>`;
  return `${header}<section class="node-grid">${state.nodes.map(node => {
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
    const targetLines = compatibility.observations.map(item => {
      const audit = auditResultState(item.result);
      return `<span title="${escapeHTML(audit.detail)}">${escapeHTML(item.ip)} · ${escapeHTML(audit.label)}</span>`;
    }).join("");
    return `<div class="unlock-result-row"><div><strong>${escapeHTML(name)}</strong><span>${t("agentReference")} · ${escapeHTML(value)}</span>${targetLines}</div><span class="badge ${compatibility.kind}">${escapeHTML(compatibility.label)}</span></div>`;
  }).join("");
  return `<div class="modal-backdrop"><section class="modal panel"><header class="modal-head"><div><h2>${t("unlockResults")} · ${escapeHTML(node?.name || "-")}</h2><p>${t("unlockSourceHint")}</p></div><button class="btn icon modal-close">×</button></header><div class="modal-body"><div class="unlock-summary"><div><span>${t("actualAvailable")}</span><strong>${available.length}</strong></div><div><span>${t("actualProblem")}</span><strong>${problems.length}</strong></div><div><span>${t("referenceOnly")}</span><strong>${reference.length}</strong></div></div>${state.modal.running ? `<div class="unlock-wait"><div class="spinner"></div><span>${t("waitingResults")}</span></div>` : ""}<div class="unlock-results">${rows || `<div class="empty"><strong>${t("unknown")}</strong></div>`}</div>${state.modal.error ? `<div class="form-error">${escapeHTML(state.modal.error)}</div>` : ""}</div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close">${t("close")}</button><button class="btn primary" id="node-check-retry" ${state.modal.running ? "disabled" : ""}>${t("checkAgain")}</button></div></footer></section></div>`;
}

function ipConfigNode(config) { return config ? state.nodes.find(node => nodeID(node.id) === nodeID(config.dns_node_id)) : null; }

function ipConfigsHTML() {
  const header = contentHeaderHTML(state.lang === "zh" ? "IP 配置" : "IP Configs", state.lang === "zh" ? "每个目标 IP 独立配置解锁服务、统计 DNS 53 与解锁 TCP 80/443 流量。" : "Configure each target independently and count only DNS 53 plus unlock TCP 80/443 traffic.", `<button class="btn primary" id="add-ip"><i class="bi bi-plus-circle"></i>${t("addIP")}</button>`);
  if (!state.ipConfigs.length) return `${header}<div class="panel empty"><strong>${t("noIPConfigs")}</strong><button class="btn primary" id="empty-add-ip"><i class="bi bi-plus-circle"></i>${t("addIP")}</button></div>`;
  const totalTraffic = state.ipConfigs.reduce((total, config) => total + Number(config.traffic_rx_bytes || 0) + Number(config.traffic_tx_bytes || 0), 0);
  return `${header}<section class="traffic-summary panel"><div><span>${t("totalTraffic")}</span><strong>${formatBytes(totalTraffic)}</strong><small>${t("trafficHint")}</small></div><button class="btn danger" id="clear-all-traffic">${t("clearAllTraffic")}</button></section><section class="ip-list panel"><div class="ip-list-head"><span>${t("targetIP")}</span><span>${t("selectedServices")}</span><span>${t("clientTraffic")}</span><span>${t("dnsClient")}</span><span></span></div>${state.ipConfigs.map(config => {
    const node = ipConfigNode(config); const status = clientState(config, node); const count = Object.keys(config.routes || {}).length; const traffic = Number(config.traffic_rx_bytes || 0) + Number(config.traffic_tx_bytes || 0);
    const audited = Object.values(config.service_results || {}).map(auditResultState); const passed = audited.filter(result => result.kind === "good").length;
    return `<article class="ip-row" data-ip-id="${escapeHTML(config.id)}" role="button" tabindex="0" aria-label="${escapeHTML(`${t("editIP")} ${config.ip}`)}"><div class="ip-main"><strong>${escapeHTML(config.ip)}</strong><span>${escapeHTML(config.note || config.node_name || "-")}</span></div><div class="ip-selection"><span class="badge good">${count} / ${state.catalog.length}</span>${audited.length ? `<small>${t("actualAudit")} ${passed}/${audited.length}</small>` : ""}</div><div class="ip-traffic" title="RX ${formatBytes(config.traffic_rx_bytes)} · TX ${formatBytes(config.traffic_tx_bytes)}"><strong>${formatBytes(traffic)}</strong><span>${t("trafficUpdated")}: ${formatDate(config.traffic_updated_at)}</span></div><div class="ip-node-state" title="${escapeHTML(status.detail)}"><span class="status-dot" style="background:${status.kind === "good" ? "var(--good)" : status.kind === "warn" ? "var(--warn)" : "var(--bad)"}"></span><span>${escapeHTML(status.label)}</span></div><div class="ip-row-actions"><button class="btn small ip-script" data-ip-id="${escapeHTML(config.id)}">${t("clientScript")}</button><button class="btn small primary ip-edit" data-ip-id="${escapeHTML(config.id)}">${t("editIP")}</button><button class="btn small ip-clear-traffic" data-ip-id="${escapeHTML(config.id)}">${t("clearTraffic")}</button><button class="btn small danger ip-delete" data-ip-id="${escapeHTML(config.id)}">${t("deleteNode")}</button></div></article>`;
  }).join("")}</section>`;
}

function bindShell() {
  updateClock();
  if (!clockTimer) clockTimer = setInterval(updateClock, 1000);
  document.querySelectorAll(".theme-toggle-trigger").forEach(button => button.addEventListener("click", () => { state.theme = state.theme === "light" ? "dark" : "light"; localStorage.setItem("prism_theme_v2", state.theme); render(); }));
  document.querySelectorAll(".lang-toggle-trigger").forEach(button => button.addEventListener("click", () => { state.lang = state.lang === "zh" ? "en" : "zh"; localStorage.setItem("enhancer_lang", state.lang); render(); }));
  document.querySelectorAll(".logout-trigger").forEach(button => button.addEventListener("click", () => logout(true)));
  document.querySelectorAll(".account-settings-trigger").forEach(button => button.addEventListener("click", openAccountSettings));
  document.querySelectorAll(".site-settings-trigger").forEach(button => button.addEventListener("click", openBrandingSettings));
  document.getElementById("refresh")?.addEventListener("click", () => loadAll());
  document.getElementById("refresh-audits")?.addEventListener("click", () => loadAll());
  document.getElementById("add-service")?.addEventListener("click", () => openServiceForm());
  document.getElementById("add-node")?.addEventListener("click", () => openNodeForm());
  document.getElementById("add-ip")?.addEventListener("click", () => openIPForm());
  document.getElementById("empty-add-ip")?.addEventListener("click", () => openIPForm());
  document.getElementById("empty-add-node")?.addEventListener("click", () => openNodeForm());
  document.getElementById("refresh-catalog")?.addEventListener("click", refreshCatalog);
  document.getElementById("manage-categories")?.addEventListener("click", () => openCategoryManager());
  const search = document.getElementById("service-search");
  search?.addEventListener("compositionstart", () => { searchComposing = true; });
  search?.addEventListener("compositionend", event => { searchComposing = false; scheduleServiceSearch(event.target); });
  search?.addEventListener("input", event => { if (!searchComposing) scheduleServiceSearch(event.target); });
  document.getElementById("category-filter")?.addEventListener("change", event => { state.category = event.target.value; state.page = 1; render(); });
  document.getElementById("status-filter")?.addEventListener("change", event => { state.statusFilter = event.target.value; state.page = 1; render(); });
  document.getElementById("node-filter")?.addEventListener("change", event => { state.nodeFilter = event.target.value; state.page = 1; render(); });
  document.getElementById("dns-select")?.addEventListener("change", async event => { state.dnsNodeId = event.target.value; localStorage.setItem("enhancer_dns", state.dnsNodeId); await loadRoutingState(); render(); });
  document.getElementById("clear-filter")?.addEventListener("click", () => { state.search = ""; state.category = ""; state.statusFilter = ""; state.nodeFilter = ""; state.page = 1; render(); });
  document.getElementById("page-size")?.addEventListener("change", event => { state.pageSize = Number(event.target.value) || 20; state.page = 1; render(); });
  document.querySelector(".page-prev")?.addEventListener("click", () => { state.page = Math.max(1, state.page - 1); render(); });
  document.querySelector(".page-next")?.addEventListener("click", () => { state.page += 1; render(); });
  document.getElementById("select-page-services")?.addEventListener("change", event => {
    document.querySelectorAll(".service-select").forEach(input => {
      input.checked = event.target.checked;
      if (input.checked) state.selectedServiceIds.add(input.dataset.serviceId);
      else state.selectedServiceIds.delete(input.dataset.serviceId);
    });
  });
  document.getElementById("batch-action")?.addEventListener("change", event => {
    const action = event.target.value;
    event.target.value = "";
    if (action === "select-page") {
      document.querySelectorAll(".service-select").forEach(input => {
        input.checked = true;
        state.selectedServiceIds.add(input.dataset.serviceId);
      });
    } else if (action === "clear") {
      state.selectedServiceIds.clear();
      document.querySelectorAll(".service-select").forEach(input => { input.checked = false; });
    } else if (action === "configure") {
      openBulkServiceConfig();
    }
  });
  document.querySelectorAll(".service-select").forEach(input => input.addEventListener("change", () => {
    if (input.checked) state.selectedServiceIds.add(input.dataset.serviceId);
    else state.selectedServiceIds.delete(input.dataset.serviceId);
  }));
  document.querySelectorAll(".alert-open").forEach(button => button.addEventListener("click", () => {
    state.tab = button.dataset.alertTab;
    localStorage.setItem("enhancer_tab", state.tab);
    render();
  }));
  document.querySelectorAll(".service-open").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openServiceRoute(button.dataset.serviceId); }));
  document.querySelectorAll(".service-category-open").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openServiceCategory(button.dataset.serviceId); }));
  document.querySelectorAll(".service-audit-run").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); runTableServiceAudit(button.dataset.serviceId); }));
  document.querySelectorAll(".service-card").forEach(card => card.addEventListener("click", () => openService(card.dataset.serviceId)));
  document.querySelectorAll(".selected-service-row").forEach(row => row.addEventListener("click", () => openService(row.dataset.serviceId)));
  const container = document.querySelector(".container");
  container?.addEventListener("click", event => {
    const button = event.target.closest?.("button");
    if (button && container.contains(button)) {
      const node = state.nodes.find(item => nodeID(item.id) === nodeID(button.dataset.nodeId));
      if (button.matches(".node-test")) { event.preventDefault(); triggerNodeCheck(button.dataset.nodeId); return; }
      if (button.matches(".node-install")) { event.preventDefault(); openInstallCommand(button.dataset.nodeId); return; }
      if (button.matches(".node-manage-ip")) { event.preventDefault(); const existing = configForNode(node); openIPForm(existing, existing ? 2 : 1, node); return; }
      if (button.matches(".node-edit")) { event.preventDefault(); openNodeForm(node); return; }
      if (button.matches(".node-delete")) { event.preventDefault(); openDeleteNode(button.dataset.nodeId); return; }
      if (button.matches(".ip-edit")) { event.preventDefault(); openIPForm(state.ipConfigs.find(config => nodeID(config.id) === nodeID(button.dataset.ipId)), 2); return; }
      if (button.matches(".ip-script")) { event.preventDefault(); openIPScript(button.dataset.ipId); return; }
      if (button.matches(".ip-delete")) { event.preventDefault(); openIPDelete(button.dataset.ipId); return; }
      if (button.matches(".ip-clear-traffic")) { event.preventDefault(); clearTraffic(button.dataset.ipId); return; }
      return;
    }
    const row = event.target.closest?.(".ip-row");
    if (row && container.contains(row)) {
      openIPForm(state.ipConfigs.find(config => nodeID(config.id) === nodeID(row.dataset.ipId)), 2);
    }
  });
  container?.addEventListener("keydown", event => {
    if (event.key !== "Enter" && event.key !== " ") return;
    const row = event.target.closest?.(".ip-row");
    if (!row || !container.contains(row)) return;
    event.preventDefault();
    openIPForm(state.ipConfigs.find(config => nodeID(config.id) === nodeID(row.dataset.ipId)), 2);
  });
  document.getElementById("clear-all-traffic")?.addEventListener("click", clearAllTraffic);
  markLoadedServiceIcons();
}

function scheduleServiceSearch(input) {
  state.search = input.value;
  state.page = 1;
  const caret = input.selectionStart;
  clearTimeout(searchRenderTimer);
  searchRenderTimer = setTimeout(() => {
    render();
    const next = document.getElementById("service-search");
    if (next) {
      next.focus({preventScroll:true});
      if (Number.isInteger(caret)) next.setSelectionRange(caret, caret);
    }
  }, 180);
}

function filterServiceCards() {
  const query = state.search;
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
  const current = routedNodeForService(service);
  const fallback = proxyNodes().find(isOnline) || proxyNodes()[0];
  state.modal = {type:"service", service, proxyId:nodeID(current?.id || fallback?.id)};
  state.testResults = null;
  renderModal();
}
function openServiceDomains(id) {
  const service = state.catalog.find(item => item.id === id);
  if (!service) return;
  if (service.custom) {
    openServiceForm(service);
    return;
  }
  state.modal = {type:"service-domains", service, error:"", busy:false};
  renderModal();
}
function openServiceRoute(id) {
  const service = state.catalog.find(item => item.id === id);
  if (!service) return;
  const config = configForNode(selectedDNS());
  if (!config) {
    openService(id);
    return;
  }
  openIPForm(config, 2);
  state.modal.serviceSearch = displayServiceName(service);
  renderModal();
}
function openBulkServiceConfig() {
  const config = configForNode(selectedDNS());
  if (!config) {
    toast(state.lang === "zh" ? "请先选择已纳管的目标 IP" : "Select a managed target first", "warn");
    return;
  }
  if (!state.selectedServiceIds.size) {
    toast(state.lang === "zh" ? "请先选择服务" : "Select services first", "warn");
    return;
  }
  openIPForm(config, 2);
  const fallback = nodeID(state.nodeFilter) || nodeID(Object.values(state.modal.routes)[0]) || nodeID(proxyNodes().find(isOnline)?.id || proxyNodes()[0]?.id);
  state.selectedServiceIds.forEach(serviceId => {
    if (!state.modal.routes[serviceId] && fallback) state.modal.routes[serviceId] = fallback;
  });
  state.modal.serviceSearch = "";
  renderModal();
}
async function runTableServiceAudit(id) {
  const service = state.catalog.find(item => item.id === id);
  const config = configForNode(selectedDNS());
  if (!service || !serviceRouteEntry(config, service)) {
    toast(state.lang === "zh" ? "请先为当前目标配置该服务" : "Configure this service for the current target first", "warn");
    return;
  }
  openIPForm(config, 2);
  state.modal.serviceSearch = displayServiceName(service);
  renderModal();
  await triggerIPServiceAudit(id);
}
function openServiceForm(service = null) { state.modal = {type:"service-form", service}; renderModal(); }
function openServiceCategory(id) {
  const service = state.catalog.find(item => item.id === id);
  if (!service) return;
  state.modal = {type:"service-category", service, error:"", busy:false};
  renderModal();
}
function openCategoryManager(serviceId = "") { state.modal = {type:"category-manager", serviceId, error:"", busy:false}; renderModal(); }
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
  if (state.modal.type === "service-domains") root.innerHTML = serviceDomainFormHTML(state.modal.service);
  if (state.modal.type === "service-form") root.innerHTML = serviceFormHTML(state.modal.service);
  if (state.modal.type === "service-category") root.innerHTML = serviceCategoryHTML();
  if (state.modal.type === "category-manager") root.innerHTML = categoryManagerHTML();
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
  const visibleDomains = service.domains.slice(0, 18);
  const hiddenDomainCount = Math.max(0, service.domains.length - visibleDomains.length);
  return `<div class="modal-backdrop"><section class="modal panel"><header class="modal-head"><div><h2>${t("serviceConfig")} · ${escapeHTML(service.name)}</h2><p>${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")} · ${status.label}</p></div><button class="btn icon modal-close" type="button">×</button></header>
     <div class="modal-body"><div class="field"><label>${t("targetServer")}</label><div class="server-list">${proxies.length ? proxies.map(node => {
       const selected = nodeID(state.modal.proxyId) === nodeID(node.id); const nodeStatus = serviceStatus(service, node);
       return `<label class="server-option ${selected ? "selected" : ""}"><input type="radio" name="proxy" value="${nodeID(node.id)}" ${selected ? "checked" : ""}><div class="server-name"><strong>${escapeHTML(node.name)}</strong><span>${escapeHTML(node.country || node.public_ip || node.address || "-")}</span></div><span class="badge ${nodeStatus.kind}">${nodeStatus.label}</span></label>`;
     }).join("") : `<div class="empty"><strong>${t("noProxy")}</strong></div>`}</div></div>
     <div class="service-domain-box"><div class="service-domain-head"><div><strong>${t("domainList")}</strong><span>${service.domain_override ? t("domainOverrideHint") : `${service.domains.length} ${t("domains")}`}</span></div><button class="btn small" id="edit-service-domains" type="button">${t("viewDomains")}</button></div><div class="domain-chip-list">${visibleDomains.map(domain => `<code>${escapeHTML(domain)}</code>`).join("")}${hiddenDomainCount ? `<span class="domain-chip-more">+${hiddenDomainCount}</span>` : ""}</div>${service.domain_override ? `<button class="btn small" id="restore-service-domains" type="button">${t("restoreDomains")}</button>` : ""}</div>
     <div class="inline" style="margin-top:15px"><span class="badge ${manual ? "warn" : ""}">${manual ? t("manual") : t("auto")}</span><span class="hint">${t("currentRoute")}: ${escapeHTML(current?.name || "-")}</span></div>
     ${state.testResults ? testResultsHTML(state.testResults) : ""}</div>
     <footer class="modal-foot"><div class="inline"><button class="btn small" id="edit-service-category" type="button">${t("serviceCategory")}</button>${service.custom ? `<button class="btn small" id="edit-custom" type="button">${t("customEdit")}</button><button class="btn small danger" id="delete-custom" type="button">${t("customDelete")}</button>` : ""}</div><div class="modal-foot-right"><button class="btn" id="test-service" type="button" ${!state.modal.proxyId ? "disabled" : ""}>${t("connectivity")}</button><button class="btn" id="reset-service" type="button" ${!rule || !manual ? "disabled" : ""}>${t("reset")}</button><button class="btn primary" id="assign-service" type="button" ${!state.modal.proxyId || !state.dnsNodeId ? "disabled" : ""}>${t("assign")}</button></div></footer></section></div>`;
}

function serviceDomainFormHTML(service) {
  return `<div class="modal-backdrop"><form class="modal medium panel" id="service-domains-form"><header class="modal-head"><div><h2>${t("viewDomains")}</h2><p>${escapeHTML(displayServiceName(service))}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><div class="field"><label>${t("domainList")}</label><textarea class="textarea" name="domains" required spellcheck="false">${escapeHTML((service.domains || []).join("\n"))}</textarea><span class="hint">${state.lang === "zh" ? "每行一个域名；泛域名请使用 *.example.com，也支持逗号和分号分隔。" : "One domain per line. Use *.example.com for wildcards; commas and semicolons are accepted."}</span></div><div class="form-error">${escapeHTML(state.modal.error || "")}</div></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit" ${state.modal.busy ? "disabled" : ""}>${t("save")}</button></div></footer></form></div>`;
}

function testResultsHTML(results) {
  return `<div class="test-results">${results.map(result => {
    const detail = result.detail || result.error || (result.addresses || []).join(", ");
    const label = result.detail ? (result.success ? t("supported") : t("unavailable")) : (result.success ? `TLS ${result.tls_ms}ms` : "FAIL");
    return `<div class="test-row"><span title="${escapeHTML(detail)}">${escapeHTML(result.domain)} · ${escapeHTML(detail)}</span><strong style="color:${result.success ? "var(--good)" : "var(--bad)"}">${label}</strong></div>`;
  }).join("")}</div>`;
}

function serviceFormHTML(service) {
  const categoryOptions = catalogCategories().map(category => `<option value="${escapeHTML(category)}">${escapeHTML(displayCategory(category))}</option>`).join("");
  return `<div class="modal-backdrop"><form class="modal medium panel" id="service-form"><header class="modal-head"><div><h2>${service ? t("customEdit") : t("addService")}</h2></div><button class="btn icon modal-close" type="button">×</button></header>
    <div class="modal-body form-stack"><div class="field"><label>${t("serviceName")}</label><input class="input" name="name" value="${escapeHTML(service?.name || "")}" required maxlength="80"></div>
    <div class="field"><label>${t("category")}</label><input class="input" name="category" list="service-category-options" value="${escapeHTML(service?.category || (state.lang === "zh" ? "自定义服务" : "Custom services"))}" maxlength="64"><datalist id="service-category-options">${categoryOptions}</datalist></div>
    <div class="field"><label>${t("domainList")}</label><textarea class="textarea" name="domains" placeholder="example.com&#10;*.example.com" required>${escapeHTML((service?.domains || []).join("\n"))}</textarea><span class="hint">${state.lang === "zh" ? "每行一个域名；泛域名请写成 *.example.com，也支持逗号或分号分隔。" : "One domain per line. Use *.example.com for wildcards; commas and semicolons are also accepted."}</span></div></div>
    <footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit">${t("save")}</button></div></footer></form></div>`;
}

function serviceCategoryHTML() {
  const service = state.modal.service;
  const categories = catalogCategories();
  return `<div class="modal-backdrop"><form class="modal medium panel" id="service-category-form"><header class="modal-head"><div><h2>${t("serviceCategory")}</h2><p>${escapeHTML(displayServiceName(service))}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><div class="category-service-summary">${serviceIconHTML(service)}<div><strong>${escapeHTML(displayServiceName(service))}</strong><span>${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")}</span></div></div><div class="field"><label>${t("category")}</label><select class="select" name="category">${categories.map(category => `<option value="${escapeHTML(category)}" ${category === service.category ? "selected" : ""}>${escapeHTML(displayCategory(category))}</option>`).join("")}</select></div>${service.original_category ? `<div class="hint">${t("originalCategory")}: ${escapeHTML(displayCategory(service.original_category))}</div>` : ""}<div class="category-note">${t("categoryHint")}</div><button class="btn" id="open-category-manager" type="button">＋ ${t("newCategory")}</button><div class="form-error">${escapeHTML(state.modal.error || "")}</div></div><footer class="modal-foot"><div>${service.original_category ? `<button class="btn" id="restore-service-category" type="button" ${state.modal.busy ? "disabled" : ""}>${t("restoreCategory")}</button>` : ""}${service.custom ? `<button class="btn danger" id="delete-custom-category-view" type="button" ${state.modal.busy ? "disabled" : ""}>${t("customDelete")}</button>` : ""}</div><div class="modal-foot-right"><button class="btn modal-close" type="button" ${state.modal.busy ? "disabled" : ""}>${t("cancel")}</button><button class="btn primary" type="submit" ${state.modal.busy || !categories.length ? "disabled" : ""}>${t("save")}</button></div></footer></form></div>`;
}

function categoryManagerHTML() {
  const categories = catalogCategories();
  const custom = new Set(state.customCategories || []);
  const serviceId = state.modal.serviceId || "";
  return `<div class="modal-backdrop"><section class="modal medium panel"><header class="modal-head"><div><h2>${t("manageCategories")}</h2><p>${t("categoryHint")}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><form class="category-create" id="category-create-form"><input class="input" name="name" placeholder="${t("categoryName")}" maxlength="64" required autocomplete="off"><button class="btn primary" type="submit" ${state.modal.busy ? "disabled" : ""}>＋ ${t("newCategory")}</button></form><div class="category-list">${categories.map(category => {
    const count = state.catalog.filter(service => service.category === category).length;
    const customCategory = custom.has(category);
    return `<div class="category-row"><button class="category-row-main category-filter-select" type="button" data-category="${escapeHTML(category)}"><strong>${escapeHTML(displayCategory(category))}</strong><span>${count} ${state.lang === "zh" ? "项服务" : "services"} · ${customCategory ? t("customCategory") : t("builtInCategory")}</span></button><div class="category-row-actions">${serviceId ? `<button class="btn small category-apply" type="button" data-category="${escapeHTML(category)}">${t("useCategory")}</button>` : ""}${customCategory ? `<button class="btn small danger category-delete" type="button" data-category="${escapeHTML(category)}" ${count ? "disabled" : ""} title="${count ? escapeHTML(t("categoryInUse")) : ""}">${state.lang === "zh" ? "删除" : "Delete"}</button>` : ""}</div></div>`;
  }).join("")}</div><div class="form-error">${escapeHTML(state.modal.error || "")}</div></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn primary modal-close" type="button">${t("close")}</button></div></footer></section></div>`;
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
    return `<div class="modal-backdrop"><form class="modal medium panel" id="ip-form"><header class="modal-head"><div><h2>${editing ? t("editIP") : t("addIP")}</h2><p>${t("targetIP")}</p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body form-stack"><div class="field"><label>${t("targetIP")}</label><input class="input" name="ip" value="${escapeHTML(draft.ip)}" placeholder="203.0.113.10" ${editing ? "disabled" : ""} required></div><div class="field"><label>${t("note")}</label><input class="input" name="note" value="${escapeHTML(draft.note)}" maxlength="80"></div><div class="field"><label>${t("defaultProxy")}</label><select class="select" name="default_proxy" required><option value="">${t("noProxy")}</option>${proxies.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(modal.defaultProxy) ? "selected" : ""}>${escapeHTML(node.name)} · ${escapeHTML(node.country || node.public_ip || node.address || "-")}</option>`).join("")}</select></div><label class="toggle-row"><input type="checkbox" name="smart" ${draft.smart ? "checked" : ""}><span><strong>${t("smartMode")}</strong><small>${state.lang === "zh" ? "保留双栈并由 Agent 选择可用出口" : "Preserve dual stack and let Agent select a working egress"}</small></span></label><div class="form-error">${escapeHTML(modal.error || "")}</div></div><footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit" ${!proxies.length ? "disabled" : ""}>${t("nextStep")}</button></div></footer></form></div>`;
  }
  const services = state.catalog;
  const selectedCount = Object.keys(modal.routes).length;
  return `<div class="modal-backdrop"><form class="modal ip-modal panel" id="ip-form"><header class="modal-head"><div><h2>${t("chooseServices")}</h2><p>${escapeHTML(draft.ip)} · ${t("selectedServices")} <span id="ip-selected-count">${selectedCount}</span></p></div><button class="btn icon modal-close" type="button">×</button></header><div class="modal-body ip-config-body"><input class="input" id="ip-service-search" value="${escapeHTML(modal.serviceSearch)}" placeholder="${t("search")}" autocomplete="off" autocapitalize="off" spellcheck="false"><div class="ip-service-picker">${services.map(service => {
     const selectedProxy = serviceRouteEntry({routes:modal.routes}, service)?.value || ""; const checked = !!selectedProxy; const audit = auditResultState(serviceResult(modal.config, service));
    const canAudit = checked && !!modal.config?.id;
    return `<article class="ip-service-option ${checked ? "selected" : ""}" data-service-id="${escapeHTML(service.id)}"><label><input type="checkbox" class="ip-service-check" data-service-id="${escapeHTML(service.id)}" ${checked ? "checked" : ""}>${serviceIconHTML(service)}<span class="ip-service-name"><strong>${escapeHTML(displayServiceName(service))}</strong><small>${escapeHTML(displayCategory(service.category))}</small></span></label><select class="select ip-service-proxy" data-service-id="${escapeHTML(service.id)}" ${checked ? "" : "disabled"}>${proxies.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(selectedProxy || modal.defaultProxy) ? "selected" : ""}>${escapeHTML(node.name)}</option>`).join("")}</select><div class="service-audit" ${checked ? "" : "hidden"}><div class="service-audit-copy"><span>${t("actualAudit")}</span><small>${state.lang === "zh" ? "只报告结果，不自动切换节点" : "Report only; routing never changes automatically"}</small></div><div class="service-audit-actions"><span class="badge ${audit.kind}" title="${escapeHTML(audit.detail)}">${escapeHTML(audit.label)}</span><button class="btn small ip-service-audit" type="button" data-service-id="${escapeHTML(service.id)}" ${canAudit ? "" : "disabled"} title="${canAudit ? "" : escapeHTML(state.lang === "zh" ? "请先保存并安装客户端" : "Save and install the client first")}">${state.lang === "zh" ? "立即实测" : "Test now"}</button></div></div></article>`;
  }).join("")}</div><div class="empty" id="ip-service-filter-empty" hidden><strong>${t("noServices")}</strong></div><div class="form-error">${escapeHTML(modal.error || "")}</div></div><footer class="modal-foot delivery-foot"><button class="btn" id="ip-back" type="button">${t("back")}</button><div class="delivery-impact"><strong>${state.lang === "zh" ? "保存后自动下发" : "Automatic delivery after save"}</strong><span>${state.lang === "zh" ? "Agent 自动同步配置；实测只更新结果，不会改动您选择的节点" : "The Agent syncs configuration; audits update reports without changing your selected node"}</span></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" id="ip-save-config" type="submit" ${!selectedCount ? "disabled" : ""}>${state.lang === "zh" ? "保存并自动下发" : "Save and deliver"}</button></div></footer></form></div>`;
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
  document.getElementById("edit-service-domains")?.addEventListener("click", () => openServiceDomains(state.modal.service.id));
  document.getElementById("restore-service-domains")?.addEventListener("click", restoreServiceDomains);
  document.getElementById("edit-custom")?.addEventListener("click", () => openServiceForm(state.modal.service));
  document.getElementById("edit-service-category")?.addEventListener("click", () => openServiceCategory(state.modal.service.id));
  document.getElementById("delete-custom")?.addEventListener("click", deleteCustomService);
  document.getElementById("delete-custom-category-view")?.addEventListener("click", deleteCustomService);
  document.getElementById("service-form")?.addEventListener("submit", saveCustomService);
  document.getElementById("service-domains-form")?.addEventListener("submit", saveServiceDomains);
  document.getElementById("service-category-form")?.addEventListener("submit", saveServiceCategory);
  document.getElementById("restore-service-category")?.addEventListener("click", restoreServiceCategory);
  document.getElementById("open-category-manager")?.addEventListener("click", () => openCategoryManager(state.modal.service.id));
  document.getElementById("category-create-form")?.addEventListener("submit", createCategory);
  document.querySelectorAll(".category-delete").forEach(button => button.addEventListener("click", () => deleteCategory(button.dataset.category)));
  document.querySelectorAll(".category-apply").forEach(button => button.addEventListener("click", () => applyManagedCategory(button.dataset.category)));
  document.querySelectorAll(".category-filter-select").forEach(button => button.addEventListener("click", () => {
    if (state.modal.serviceId) return;
    state.category = button.dataset.category;
    closeModal();
    render();
  }));
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
      const linkedProxy = linkedRouteServiceIds(input.dataset.serviceId, state.modal.routes)
        .map(serviceId => state.modal.routes[serviceId])
        .find(Boolean);
      state.modal.routes[input.dataset.serviceId] = linkedProxy || select?.value || state.modal.defaultProxy;
      if (select) select.value = state.modal.routes[input.dataset.serviceId];
      applyLinkedRouteProxy(input.dataset.serviceId, state.modal.routes[input.dataset.serviceId]);
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
    if (select.closest(".ip-service-option")?.querySelector(".ip-service-check")?.checked) {
      state.modal.routes[select.dataset.serviceId] = select.value;
      applyLinkedRouteProxy(select.dataset.serviceId, select.value, true);
    }
  }));
  document.querySelectorAll(".ip-service-audit").forEach(button => button.addEventListener("click", () => triggerIPServiceAudit(button.dataset.serviceId)));
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

async function saveServiceCategory(event) {
  event.preventDefault();
  const service = state.modal.service;
  const category = String(new FormData(event.currentTarget).get("category") || "");
  state.modal.busy = true; state.modal.error = ""; renderModal();
  try {
    await api(`/enhancer/api/service-categories/${encodeURIComponent(service.id)}`, {method:"PUT", body:JSON.stringify({category})});
    state.modal = null;
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) {
    state.modal.busy = false; state.modal.error = error.message; renderModal();
  }
}

async function saveServiceDomains(event) {
  event.preventDefault();
  const service = state.modal.service;
  const form = new FormData(event.currentTarget);
  const domains = String(form.get("domains") || "").split(/\r?\n|,|;/);
  state.modal.busy = true;
  state.modal.error = "";
  renderModal();
  try {
    await api(`/enhancer/api/service-domains/${encodeURIComponent(service.id)}`, {method:"PUT", body:JSON.stringify({domains})});
    state.modal = null;
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) {
    state.modal.busy = false;
    state.modal.error = error.message;
    renderModal();
  }
}

async function restoreServiceDomains() {
  const service = state.modal.service;
  state.modal.busy = true;
  state.modal.error = "";
  renderModal();
  try {
    await api(`/enhancer/api/service-domains/${encodeURIComponent(service.id)}`, {method:"DELETE"});
    state.modal = null;
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) {
    state.modal.busy = false;
    state.modal.error = error.message;
    renderModal();
  }
}

async function restoreServiceCategory() {
  const service = state.modal.service;
  state.modal.busy = true; state.modal.error = ""; renderModal();
  try {
    await api(`/enhancer/api/service-categories/${encodeURIComponent(service.id)}`, {method:"DELETE"});
    state.modal = null;
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) {
    state.modal.busy = false; state.modal.error = error.message; renderModal();
  }
}

async function createCategory(event) {
  event.preventDefault();
  const name = String(new FormData(event.currentTarget).get("name") || "").trim();
  const serviceId = state.modal.serviceId || "";
  state.modal.busy = true; state.modal.error = ""; renderModal();
  try {
    const created = await api("/enhancer/api/categories", {method:"POST", body:JSON.stringify({name})});
    state.categories = [...new Set([...state.categories, created.name])];
    state.customCategories = [...new Set([...state.customCategories, created.name])];
    if (serviceId) {
      await api(`/enhancer/api/service-categories/${encodeURIComponent(serviceId)}`, {method:"PUT", body:JSON.stringify({category:created.name})});
      state.modal = null;
      await loadAll(true);
      toast(t("saved"), "good");
      return;
    }
    state.modal.busy = false;
    render();
    toast(t("categoryCreated"), "good");
  } catch (error) {
    state.modal.busy = false; state.modal.error = error.message; renderModal();
  }
}

async function deleteCategory(category) {
  if (!confirm(t("categoryDeleteConfirm"))) return;
  try {
    await api("/enhancer/api/categories", {method:"DELETE", body:JSON.stringify({name:category})});
    state.categories = state.categories.filter(item => item !== category);
    state.customCategories = state.customCategories.filter(item => item !== category);
    render();
    toast(t("categoryDeleted"), "good");
  } catch (error) {
    state.modal.error = error.message; renderModal();
  }
}

async function applyManagedCategory(category) {
  const serviceId = state.modal.serviceId;
  if (!serviceId) return;
  state.modal.busy = true; state.modal.error = ""; renderModal();
  try {
    await api(`/enhancer/api/service-categories/${encodeURIComponent(serviceId)}`, {method:"PUT", body:JSON.stringify({category})});
    state.modal = null;
    await loadAll(true);
    toast(t("saved"), "good");
  } catch (error) {
    state.modal.busy = false; state.modal.error = error.message; renderModal();
  }
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
  const query = state.modal.serviceSearch;
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
    const targetResult = serviceResult(targetConfig, service);
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
  if (!modal || modal.type !== "ip-form") return;
  if (!Object.keys(modal.routes).length) return;
  const editing = !!modal.config;
  const payload = {...modal.draft, routes:modal.routes};
  try {
    if (!editing) {
      const existing = (Array.isArray(state.ipConfigs) ? state.ipConfigs : [])
        .find(config => normalizeIP(config.ip) === normalizeIP(modal.draft.ip));
      if (existing) {
        modal.config = existing;
        modal.routes = {...(existing.routes || {}), ...modal.routes};
        return saveIPConfig();
      }
    }
    const path = editing ? `/enhancer/api/ip-configs/${encodeURIComponent(modal.config.id)}` : "/enhancer/api/ip-configs";
    const config = await api(path, {method:editing ? "PUT" : "POST", body:JSON.stringify(payload)});
    const linkedRoutes = Object.keys(config.routes || {}).filter(serviceId => payload.routes?.[serviceId] && nodeID(payload.routes[serviceId]) !== nodeID(config.routes[serviceId])).length;
    state.modal = editing ? null : {type:"ip-script", config, command:ipScriptCommand(config)};
    await loadAll(true);
    renderModal();
    toast(linkedRoutes ? (state.lang === "zh" ? `保存成功；${linkedRoutes} 个共享域名服务已自动联动到同一解锁机` : `Saved; ${linkedRoutes} overlapping-domain services were linked to one proxy`) : t("saved"), "good");
  } catch (error) {
    if (!editing && /IP.*存在|already exists/i.test(error.message || "")) {
      try {
        const configs = await api("/enhancer/api/ip-configs");
        const existing = (Array.isArray(configs) ? configs : []).find(config => normalizeIP(config.ip) === normalizeIP(modal.draft.ip));
        if (existing) {
          modal.config = existing;
          modal.routes = {...(existing.routes || {}), ...modal.routes};
          return saveIPConfig();
        }
      } catch (recoveryError) {
        error = recoveryError;
      }
    }
    if (state.modal === modal) {
      modal.error = error.message;
      renderModal();
    } else {
      toast(error.message, "error");
    }
  }
}

function updateIPServiceAuditUI(config) {
  document.querySelectorAll(".ip-service-option").forEach(option => {
    const serviceId = option.dataset.serviceId;
    const service = state.catalog.find(item => item.id === serviceId);
    const badge = option.querySelector(".service-audit .badge");
    if (!badge) return;
    const audit = auditResultState(serviceResult(config, service));
    badge.className = `badge ${audit.kind}`;
    badge.textContent = audit.label;
    badge.title = audit.detail;
  });
}

async function triggerIPServiceAudit(serviceId) {
  const modal = state.modal;
  const configID = modal?.config?.id;
  if (!configID) return;
  const buttons = [...document.querySelectorAll(".ip-service-audit")];
  buttons.forEach(button => {
    button.disabled = true;
    if (button.dataset.serviceId === serviceId) button.textContent = state.lang === "zh" ? "实测中..." : "Testing...";
  });
  try {
    const requested = await api(`/enhancer/api/ip-configs/${encodeURIComponent(configID)}/audit`, {method:"POST"});
    const requestedAt = new Date(requested.service_audit_requested_at || Date.now()).getTime();
    const started = Date.now();
    let latest = requested;
    while (Date.now() - started < 600000) {
      await delay(2500);
      const configs = await api("/enhancer/api/ip-configs");
      state.ipConfigs = Array.isArray(configs) ? configs : state.ipConfigs;
      latest = state.ipConfigs.find(config => config.id === configID) || latest;
      const activeModal = state.modal;
      if (activeModal?.type !== "ip-form" || activeModal.config?.id !== configID) return;
      activeModal.config = latest;
      activeModal.routes = {...(latest.routes || activeModal.routes)};
      updateIPServiceAuditUI(latest);
      const auditedAt = new Date(latest.service_audited_at || 0).getTime();
      const service = state.catalog.find(item => item.id === serviceId);
      const latestResult = serviceResult(latest, service);
      if (auditedAt >= requestedAt && latestResult) {
        const audit = auditResultState(latestResult);
        toast(audit.label || t("testDone"), audit.kind === "bad" ? "error" : audit.kind || "good");
        return;
      }
    }
    throw new Error(state.lang === "zh" ? "目标机尚未返回实测结果，请检查客户端在线状态" : "The target did not return an audit result");
  } catch (error) {
    toast(error.message, "error");
  } finally {
    document.querySelectorAll(".ip-service-audit").forEach(button => {
      button.disabled = false;
      button.textContent = state.lang === "zh" ? "立即实测" : "Test now";
    });
  }
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
  const managedConfig = state.ipConfigs.find(config => nodeID(config.dns_node_id) === nodeID(node.id));
  if (node.role === "dns" && managedConfig) return ipScriptCommand(managedConfig);
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
  if (!["services", "nodes", "ips", "sync", "audit", "alerts", "settings"].includes(state.tab)) state.tab = "services";
  bindNavigation();
  await loadBranding();
  render();
  if (state.token) loadAll();
}
initialize();
