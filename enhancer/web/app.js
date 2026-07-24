const translations = {
  zh: {
    title: "Prism DNS 中文增强版", online: "控制器已连接", loginTitle: "登录 Prism Gateway", loginHint: "使用 Controller 管理员账号登录",
    username: "用户名", password: "密码", login: "登录", loggingIn: "正在登录...", services: "服务编排", nodes: "节点管理",
    search: "搜索服务或域名", allCategories: "全部分类", dnsClient: "DNS 节点", selectDNS: "请选择 DNS 节点", addService: "新增服务",
    addNode: "新增节点", refresh: "刷新", logout: "退出", totalServices: "服务数量", configured: "已配置", proxyNodes: "解锁机",
    customServices: "自定义服务", notConfigured: "未配置", auto: "自动选择", manual: "手动", open: "配置", domains: "域名",
    serviceConfig: "配置服务", currentRoute: "当前路由", targetServer: "目标解锁机", assign: "保存并切换", reset: "恢复自动",
    connectivity: "连通性测试", testing: "正在测试...", triggerUnlock: "运行解锁检测", customEdit: "编辑服务", customDelete: "删除服务",
    serviceName: "服务名称", category: "分类", domainList: "域名列表", save: "保存", cancel: "取消", deleteConfirm: "确认删除这个自定义服务？",
    nodeName: "节点名称", role: "节点类型", proxy: "解锁机", dns: "DNS 客户端", address: "连接 IP / DNS 地址", country: "地区",
    group: "分组", priority: "优先级", smartMode: "智能模式", createNode: "创建节点", installCommand: "Agent 安装命令",
    noDNS: "尚未创建 DNS 节点", noProxy: "尚未添加解锁机", noServices: "没有匹配的服务", sourceUpdated: "名单更新时间",
    supported: "解锁成功", unavailable: "不可用", unknown: "未检测", active: "正在使用", manualOffline: "手动目标离线", ruleCreated: "服务规则已创建",
    saved: "保存成功", deleted: "删除成功", switched: "切换成功", resetDone: "已恢复自动选择", testDone: "测试完成", failed: "操作失败",
    catalogRefresh: "同步域名名单", sourceList: "细化域名库", theme: "切换主题", language: "English", originalStatus: "原生检测",
  },
  en: {
    title: "Prism DNS Enhanced", online: "Controller connected", loginTitle: "Sign in to Prism Gateway", loginHint: "Use your Controller administrator account",
    username: "Username", password: "Password", login: "Sign in", loggingIn: "Signing in...", services: "Service Routing", nodes: "Nodes",
    search: "Search services or domains", allCategories: "All categories", dnsClient: "DNS client", selectDNS: "Select a DNS client", addService: "Add service",
    addNode: "Add node", refresh: "Refresh", logout: "Sign out", totalServices: "Services", configured: "Configured", proxyNodes: "Proxy nodes",
    customServices: "Custom services", notConfigured: "Not configured", auto: "Automatic", manual: "Manual", open: "Configure", domains: "Domains",
    serviceConfig: "Configure service", currentRoute: "Current route", targetServer: "Target proxy", assign: "Save and switch", reset: "Reset to auto",
    connectivity: "Connectivity test", testing: "Testing...", triggerUnlock: "Run unlock check", customEdit: "Edit service", customDelete: "Delete service",
    serviceName: "Service name", category: "Category", domainList: "Domain list", save: "Save", cancel: "Cancel", deleteConfirm: "Delete this custom service?",
    nodeName: "Node name", role: "Node role", proxy: "Proxy agent", dns: "DNS client", address: "Connect IP / DNS address", country: "Region",
    group: "Group", priority: "Priority", smartMode: "Smart mode", createNode: "Create node", installCommand: "Agent install command",
    noDNS: "No DNS client has been created", noProxy: "No proxy agent has been added", noServices: "No matching services", sourceUpdated: "Catalog updated",
    supported: "Unlocked", unavailable: "Unavailable", unknown: "Not checked", active: "Active", manualOffline: "Manual target offline", ruleCreated: "Service rule created",
    saved: "Saved", deleted: "Deleted", switched: "Switched", resetDone: "Automatic selection restored", testDone: "Test completed", failed: "Operation failed",
    catalogRefresh: "Sync domain catalog", sourceList: "Detailed domain catalog", theme: "Toggle theme", language: "简体中文", originalStatus: "Native check",
  }
};

const commonNames = {
  "OpenAI": "ChatGPT / OpenAI", "Disney+": "Disney+", "Prime Video": "Prime Video", "Apple TV+": "Apple TV+",
  "BiliBili": "哔哩哔哩", "YouTube": "YouTube", "Gemini": "Gemini", "Claude": "Claude", "Netflix": "Netflix",
  "TikTok": "TikTok", "Spotify": "Spotify", "Meta AI": "Meta AI", "Suno": "Suno"
};

const categoryNames = {
  "AI Platform": "AI 服务", "Canada Media": "加拿大媒体", "China Media": "中国大陆媒体", "Europe Media": "欧洲媒体",
  "Global Plaform": "全球服务", "Global Platform": "全球服务", "Hong Kong Media": "香港媒体", "Indian Media": "印度媒体",
  "Japan Media": "日本媒体", "Korean Media": "韩国媒体", "North America Media": "北美媒体", "Oceania Media": "大洋洲媒体",
  "Others": "其他服务", "South America Media": "南美媒体", "SouthEastAsia media": "东南亚媒体", "Taiwan Media": "台湾媒体"
};

const serviceIcons = ["◈", "◆", "●", "▲", "■", "✦", "◎", "◇"];
const state = {
  token: localStorage.getItem("prism_token") || "",
  user: JSON.parse(localStorage.getItem("prism_user") || "{}"),
  lang: localStorage.getItem("enhancer_lang") || "zh",
  theme: localStorage.getItem("prism_theme") || "dark",
  tab: "services", nodes: [], rules: [], catalog: [], catalogMeta: {}, dnsNodeId: localStorage.getItem("enhancer_dns") || "",
  overrides: {}, activeSelections: {}, search: "", category: "", loading: false, modal: null, testResults: null
};

function t(key) { return translations[state.lang][key] || key; }
function escapeHTML(value = "") { return String(value).replace(/[&<>'"]/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"}[char])); }
function nodeID(value) { return value == null ? "" : String(value); }
function selectedDNS() { return state.nodes.find(node => nodeID(node.id) === nodeID(state.dnsNodeId)); }
function proxyNodes() { return state.nodes.filter(node => node.role === "proxy"); }
function dnsNodes() { return state.nodes.filter(node => node.role === "dns"); }
function parseUnlock(node) {
  if (!node) return {};
  if (node.unlock_data && typeof node.unlock_data === "object") return node.unlock_data;
  try { return JSON.parse(node.unlock_json || "{}"); } catch { return {}; }
}
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
  const key = serviceDetectorKey(service);
  const value = key ? parseUnlock(node)[key] || "" : "";
  if (value.startsWith("Yes")) return {kind:"good", label:t("supported"), raw:value};
  if (value) return {kind:"bad", label:t("unavailable"), raw:value};
  return {kind:"", label:t("unknown"), raw:""};
}
function iconFor(service) { return service.custom ? "✚" : serviceIcons[Math.abs(hashCode(service.id)) % serviceIcons.length]; }
function hashCode(value) { return Array.from(value).reduce((hash, char) => ((hash << 5) - hash + char.charCodeAt(0)) | 0, 0); }
function formatDate(value) { if (!value) return "-"; try { return new Date(value).toLocaleString(state.lang === "zh" ? "zh-CN" : "en-US"); } catch { return value; } }
function displayCategory(value) { return state.lang === "zh" ? categoryNames[value] || value : value; }
function isOnline(node) { if (!node.last_heartbeat) return false; return Date.now() - new Date(node.last_heartbeat).getTime() < 90000; }

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
    state.token = data.token; state.user = data.user || {};
    localStorage.setItem("prism_token", state.token); localStorage.setItem("prism_user", JSON.stringify(state.user));
    await loadAll();
  } catch (loginError) { error.textContent = loginError.message; }
  finally { button.disabled = false; button.textContent = t("login"); }
}

function logout(callAPI = true) {
  if (callAPI && state.token) fetch("/api/logout", {method:"POST", headers:{Authorization:`Bearer ${state.token}`}}).catch(() => {});
  state.token = ""; state.user = {}; state.nodes = []; state.rules = [];
  localStorage.removeItem("prism_token"); localStorage.removeItem("prism_user");
  render();
}

async function loadAll(silent = false) {
  if (!state.token) return render();
  state.loading = !silent; render();
  try {
    const [nodes, rules, catalog] = await Promise.all([api("/api/nodes"), api("/api/rules"), api("/enhancer/api/catalog")]);
    state.nodes = Array.isArray(nodes) ? nodes : [];
    state.rules = Array.isArray(rules) ? rules : [];
    state.catalog = catalog.services || [];
    state.catalogMeta = catalog;
    if (!dnsNodes().some(node => nodeID(node.id) === nodeID(state.dnsNodeId))) state.dnsNodeId = dnsNodes()[0]?.id || "";
    localStorage.setItem("enhancer_dns", nodeID(state.dnsNodeId));
    await loadRoutingState();
    connectSSE();
  } catch (error) { toast(error.message, "error"); }
  finally { state.loading = false; render(); }
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
function connectSSE() {
  if (eventSource || !state.token) return;
  eventSource = new EventSource(`/api/sse?token=${encodeURIComponent(state.token)}`);
  eventSource.onmessage = () => { clearTimeout(reloadTimer); reloadTimer = setTimeout(() => loadAll(true), 700); };
  eventSource.onerror = () => { eventSource.close(); eventSource = null; setTimeout(connectSSE, 3500); };
}

function render() {
  document.documentElement.dataset.theme = state.theme === "light" ? "light" : "dark";
  document.documentElement.lang = state.lang === "zh" ? "zh-CN" : "en";
  document.title = t("title");
  const app = document.getElementById("app");
  if (!state.token) { app.innerHTML = loginHTML(); bindLogin(); return; }
  app.innerHTML = shellHTML();
  bindShell();
  renderModal();
}

function loginHTML() {
  return `<main class="login-wrap"><section class="login-box panel"><h1>${escapeHTML(t("loginTitle"))}</h1><p>${escapeHTML(t("loginHint"))}</p>
    <form class="form-stack" id="login-form"><div class="field"><label>${t("username")}</label><input class="input" name="username" autocomplete="username" required></div>
    <div class="field"><label>${t("password")}</label><input class="input" type="password" name="password" autocomplete="current-password" required></div>
    <div class="form-error"></div><button class="btn primary" type="submit">${t("login")}</button></form></section></main>`;
}
function bindLogin() { document.getElementById("login-form").addEventListener("submit", login); }

function shellHTML() {
  return `<div class="shell"><header class="topbar"><div class="brand"><h1>${escapeHTML(t("title"))}</h1><div class="brand-meta"><span class="status-dot"></span><span>${escapeHTML(t("online"))}</span></div></div>
    <div class="top-actions"><button class="btn icon" id="theme-toggle" title="${t("theme")}">${state.theme === "light" ? "☾" : "☀"}</button>
    <button class="btn" id="lang-toggle">${t("language")}</button><span class="user-label">${escapeHTML(state.user.username || "admin")}</span><button class="btn icon danger" id="logout" title="${t("logout")}">↪</button></div></header>
    <main class="container"><div class="toolbar"><div class="toolbar-left"><nav class="tabs"><button class="tab ${state.tab === "services" ? "active" : ""}" data-tab="services">${t("services")}</button><button class="tab ${state.tab === "nodes" ? "active" : ""}" data-tab="nodes">${t("nodes")}</button></nav></div>
    <div class="toolbar-right">${state.tab === "services" ? `<button class="btn" id="refresh-catalog">↻ ${t("catalogRefresh")}</button><button class="btn primary" id="add-service">＋ ${t("addService")}</button>` : `<button class="btn primary" id="add-node">＋ ${t("addNode")}</button>`}<button class="btn icon" id="refresh" title="${t("refresh")}">↻</button></div></div>
    ${state.loading ? `<div class="loading panel"><div class="spinner"></div></div>` : state.tab === "services" ? servicesHTML() : nodesHTML()}</main></div>`;
}

function servicesHTML() {
  const dns = dnsNodes(); const proxies = proxyNodes();
  const categories = [...new Set(state.catalog.map(service => service.category))].sort((a,b) => a.localeCompare(b));
  const filtered = state.catalog.filter(service => {
    const query = state.search.trim().toLowerCase();
    const matchesSearch = !query || `${service.name} ${service.category} ${service.domains.join(" ")}`.toLowerCase().includes(query);
    return matchesSearch && (!state.category || service.category === state.category);
  });
  const configured = state.catalog.filter(service => serviceRule(service)).length;
  return `${dns.length === 0 ? `<div class="panel empty"><strong>${t("noDNS")}</strong><button class="btn primary" id="empty-add-node">＋ ${t("addNode")}</button></div>` : ""}
    <section class="stats"><div class="panel stat"><div class="stat-label">${t("totalServices")}</div><div class="stat-value">${state.catalog.length}</div><div class="stat-sub">${t("sourceList")}</div></div>
    <div class="panel stat"><div class="stat-label">${t("configured")}</div><div class="stat-value">${configured}</div><div class="stat-sub">${dns.length} ${t("dnsClient")}</div></div>
    <div class="panel stat"><div class="stat-label">${t("proxyNodes")}</div><div class="stat-value">${proxies.length}</div><div class="stat-sub">${proxies.filter(isOnline).length} online</div></div>
    <div class="panel stat"><div class="stat-label">${t("customServices")}</div><div class="stat-value">${state.catalog.filter(item => item.custom).length}</div><div class="stat-sub">${t("sourceUpdated")}: ${formatDate(state.catalogMeta.updated_at)}</div></div></section>
    <section class="filterbar panel"><input class="input" id="service-search" value="${escapeHTML(state.search)}" placeholder="${t("search")}">
    <select class="select" id="category-filter"><option value="">${t("allCategories")}</option>${categories.map(category => `<option value="${escapeHTML(category)}" ${state.category === category ? "selected" : ""}>${escapeHTML(displayCategory(category))}</option>`).join("")}</select>
    <select class="select" id="dns-select"><option value="">${t("selectDNS")}</option>${dns.map(node => `<option value="${nodeID(node.id)}" ${nodeID(node.id) === nodeID(state.dnsNodeId) ? "selected" : ""}>${escapeHTML(node.name)}</option>`).join("")}</select>
    <button class="btn" id="clear-filter">× ${state.lang === "zh" ? "清除筛选" : "Clear"}</button></section>
    ${filtered.length ? `<section class="service-grid">${filtered.map(serviceCardHTML).join("")}</section>` : `<div class="panel empty"><strong>${t("noServices")}</strong></div>`}`;
}

function serviceCardHTML(service) {
  const rule = serviceRule(service); const node = activeNodeFor(rule); const status = serviceStatus(service, node); const manual = !!overrideFor(rule);
  const displayName = state.lang === "zh" ? commonNames[service.name] || service.name : service.name;
  return `<article class="panel service-card ${rule ? "configured" : ""} ${service.custom ? "custom" : ""}" data-service-id="${escapeHTML(service.id)}">
    <div class="service-head"><div class="service-icon">${iconFor(service)}</div><span class="badge ${status.kind}">${status.label}</span></div>
    <div class="service-name" title="${escapeHTML(displayName)}">${escapeHTML(displayName)}</div><div class="service-category">${escapeHTML(displayCategory(service.category))} · ${service.domains.length} ${t("domains")}</div>
    <div class="service-route"><span>${manual ? "●" : "○"}</span><strong>${escapeHTML(node?.name || (rule ? t("auto") : t("notConfigured")))}</strong></div>
    <div class="service-actions"><span class="badge ${manual ? "warn" : ""}">${manual ? t("manual") : t("auto")}</span><button class="btn small primary service-open" data-service-id="${escapeHTML(service.id)}">${t("open")}</button></div></article>`;
}

function nodesHTML() {
  if (!state.nodes.length) return `<div class="panel empty"><strong>${state.lang === "zh" ? "还没有节点" : "No nodes"}</strong><button class="btn primary" id="empty-add-node">＋ ${t("addNode")}</button></div>`;
  return `<section class="node-grid">${state.nodes.map(node => {
    const online = isOnline(node); const unlock = parseUnlock(node); const values = Object.values(unlock).filter(Boolean); const successes = values.filter(value => String(value).startsWith("Yes")).length;
    return `<article class="panel node-card"><div class="node-title"><div><h3>${escapeHTML(node.name)}</h3><div class="brand-meta"><span class="status-dot" style="background:${online ? "var(--good)" : "var(--bad)"}"></span>${node.role === "proxy" ? t("proxy") : t("dns")}</div></div><span class="badge ${online ? "good" : "bad"}">${online ? "ONLINE" : "OFFLINE"}</span></div>
      <div class="node-meta"><div><span>${t("address")}</span><strong>${escapeHTML(node.public_ip || node.address || "-")}</strong></div><div><span>${t("country")}</span><strong>${escapeHTML(node.country || "-")}</strong></div><div><span>${t("priority")}</span><strong>${node.priority || "-"}</strong></div><div><span>${t("originalStatus")}</span><strong>${successes}/${values.length}</strong></div></div>
      <div class="inline" style="margin-top:14px"><button class="btn small node-test" data-node-id="${nodeID(node.id)}" ${node.role !== "proxy" ? "disabled" : ""}>${t("triggerUnlock")}</button></div></article>`;
  }).join("")}</section>`;
}

function bindShell() {
  document.querySelectorAll("[data-tab]").forEach(button => button.addEventListener("click", () => { state.tab = button.dataset.tab; render(); }));
  document.getElementById("theme-toggle").onclick = () => { state.theme = state.theme === "light" ? "dark" : "light"; localStorage.setItem("prism_theme", state.theme); render(); };
  document.getElementById("lang-toggle").onclick = () => { state.lang = state.lang === "zh" ? "en" : "zh"; localStorage.setItem("enhancer_lang", state.lang); render(); };
  document.getElementById("logout").onclick = () => logout(true);
  document.getElementById("refresh").onclick = () => loadAll();
  document.getElementById("add-service")?.addEventListener("click", () => openServiceForm());
  document.getElementById("add-node")?.addEventListener("click", openNodeForm);
  document.getElementById("empty-add-node")?.addEventListener("click", openNodeForm);
  document.getElementById("refresh-catalog")?.addEventListener("click", refreshCatalog);
  document.getElementById("service-search")?.addEventListener("input", event => { state.search = event.target.value; render(); document.getElementById("service-search")?.focus(); });
  document.getElementById("category-filter")?.addEventListener("change", event => { state.category = event.target.value; render(); });
  document.getElementById("dns-select")?.addEventListener("change", async event => { state.dnsNodeId = event.target.value; localStorage.setItem("enhancer_dns", state.dnsNodeId); await loadRoutingState(); render(); });
  document.getElementById("clear-filter")?.addEventListener("click", () => { state.search = ""; state.category = ""; render(); });
  document.querySelectorAll(".service-open").forEach(button => button.addEventListener("click", event => { event.stopPropagation(); openService(button.dataset.serviceId); }));
  document.querySelectorAll(".service-card").forEach(card => card.addEventListener("click", () => openService(card.dataset.serviceId)));
  document.querySelectorAll(".node-test").forEach(button => button.addEventListener("click", () => triggerNodeCheck(button.dataset.nodeId)));
}

async function refreshCatalog() {
  try { await api("/enhancer/api/catalog?refresh=1"); toast(t("saved"), "good"); await loadAll(); } catch (error) { toast(error.message, "error"); }
}

function openService(id) { const service = state.catalog.find(item => item.id === id); if (!service) return; state.modal = {type:"service", service, proxyId:nodeID(activeNodeFor(serviceRule(service))?.id)}; state.testResults = null; renderModal(); }
function openServiceForm(service = null) { state.modal = {type:"service-form", service}; renderModal(); }
function openNodeForm() { state.modal = {type:"node-form"}; renderModal(); }
function closeModal() { state.modal = null; state.testResults = null; renderModal(); }

function renderModal() {
  const root = document.getElementById("modal-root");
  if (!state.modal) { root.innerHTML = ""; return; }
  if (state.modal.type === "service") root.innerHTML = serviceModalHTML(state.modal.service);
  if (state.modal.type === "service-form") root.innerHTML = serviceFormHTML(state.modal.service);
  if (state.modal.type === "node-form") root.innerHTML = nodeFormHTML();
  bindModal();
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
  return `<div class="test-results">${results.map(result => `<div class="test-row"><span title="${escapeHTML(result.error || result.addresses?.join(", ") || "")}">${escapeHTML(result.domain)} · ${result.error ? escapeHTML(result.error) : escapeHTML((result.addresses || []).join(", "))}</span><strong style="color:${result.success ? "var(--good)" : "var(--bad)"}">${result.success ? `TLS ${result.tls_ms}ms` : "FAIL"}</strong></div>`).join("")}</div>`;
}

function serviceFormHTML(service) {
  return `<div class="modal-backdrop"><form class="modal medium panel" id="service-form"><header class="modal-head"><div><h2>${service ? t("customEdit") : t("addService")}</h2></div><button class="btn icon modal-close" type="button">×</button></header>
    <div class="modal-body form-stack"><div class="field"><label>${t("serviceName")}</label><input class="input" name="name" value="${escapeHTML(service?.name || "")}" required maxlength="80"></div>
    <div class="field"><label>${t("category")}</label><input class="input" name="category" value="${escapeHTML(service?.category || (state.lang === "zh" ? "自定义服务" : "Custom services"))}" maxlength="80"></div>
    <div class="field"><label>${t("domainList")}</label><textarea class="textarea" name="domains" placeholder="example.com&#10;cdn.example.com" required>${escapeHTML((service?.domains || []).join("\n"))}</textarea><span class="hint">${state.lang === "zh" ? "每行一个域名，也支持逗号或分号分隔。" : "One domain per line; commas and semicolons are also accepted."}</span></div></div>
    <footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit">${t("save")}</button></div></footer></form></div>`;
}

function nodeFormHTML() {
  return `<div class="modal-backdrop"><form class="modal medium panel" id="node-form"><header class="modal-head"><div><h2>${t("addNode")}</h2></div><button class="btn icon modal-close" type="button">×</button></header>
    <div class="modal-body form-stack"><div class="field"><label>${t("nodeName")}</label><input class="input" name="name" required maxlength="32"></div>
    <div class="field"><label>${t("role")}</label><select class="select" name="role"><option value="proxy">${t("proxy")}</option><option value="dns">${t("dns")}</option></select></div>
    <div class="field"><label>${t("address")}</label><input class="input" name="public_ip" placeholder="203.0.113.10"></div>
    <div class="field"><label>${t("country")}</label><input class="input" name="country" placeholder="SG / JP / US"></div>
    <div class="field"><label>${t("group")}</label><input class="input" name="group" value="unlock-all"></div>
    <div class="field"><label>${t("priority")}</label><input class="input" type="number" name="priority" min="1" max="100" value="50"></div></div>
    <footer class="modal-foot"><div></div><div class="modal-foot-right"><button class="btn modal-close" type="button">${t("cancel")}</button><button class="btn primary" type="submit">${t("createNode")}</button></div></footer></form></div>`;
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
  document.getElementById("node-form")?.addEventListener("submit", createNode);
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
    await api(`/api/nodes/${node.id}/check_unlock`, {method:"POST"}).catch(() => null);
    const data = await api("/enhancer/api/connectivity", {method:"POST", body:JSON.stringify({dns_server:serverHost(node), domains:service.domains.slice(0, 5), timeout_ms:6000})});
    state.testResults = data.results || []; toast(t("testDone"), "good"); renderModal();
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

async function createNode(event) {
  event.preventDefault(); const form = new FormData(event.currentTarget);
  const payload = {name:form.get("name"), role:form.get("role"), public_ip:form.get("public_ip"), country:form.get("country"), group:form.get("group"), priority:Number(form.get("priority")) || 50, secret:""};
  try {
    const node = await api("/api/nodes", {method:"POST", body:JSON.stringify(payload)});
    await loadAll(true);
    const smart = payload.role === "dns" ? " --smart" : "";
    const ip = payload.role === "proxy" && payload.public_ip ? ` --ip \"${payload.public_ip}\"` : "";
    const command = `curl -sL https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh | bash -s -- --master ${location.origin} --secret ${node.secret || "<secret>"}${smart}${ip}`;
    document.getElementById("modal-root").innerHTML = `<div class="modal-backdrop"><section class="modal medium panel"><header class="modal-head"><div><h2>${t("installCommand")}</h2></div><button class="btn icon modal-close">×</button></header><div class="modal-body"><div class="code">${escapeHTML(command)}</div></div></section></div>`;
    document.querySelector(".modal-close").onclick = closeModal; toast(t("saved"), "good");
  } catch (error) { toast(error.message, "error"); }
}

async function triggerNodeCheck(id) {
  try { await api(`/api/nodes/${id}/check_unlock`, {method:"POST"}); toast(t("testing"), "good"); setTimeout(() => loadAll(true), 2500); }
  catch (error) { toast(error.message, "error"); }
}

function toast(message, type = "") {
  const root = document.getElementById("toast-root"); const item = document.createElement("div"); item.className = `toast ${type}`; item.textContent = message; root.appendChild(item); setTimeout(() => item.remove(), 4200);
}

render();
if (state.token) loadAll();
