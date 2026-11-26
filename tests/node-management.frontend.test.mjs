import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import {test} from "node:test";
import {runInNewContext} from "node:vm";

const appSource = readFileSync(new URL("../enhancer/web/app.js", import.meta.url), "utf8");
const startTime = Date.parse("2026-07-24T12:00:00Z");
const release = {version:"v1.5.40"};
const checkedInAt = new Date(startTime).toISOString();
const node = {id:"fixture-node", role:"proxy", name:"Fixture node", public_ip:"192.0.2.25", last_heartbeat:checkedInAt};
const managed = {...node, management:{version:"v1.5.9", transport_version:"2.5.2", agent_version:"v1.2.1", ipv4:"198.51.100.9", last_seen:checkedInAt, update_supported:true}};
const nodesPath = "/enhancer/api/nodes";
const releasePath = "/enhancer/api/node-release";
const updatesPath = "/enhancer/api/node-updates";

function json(data, status = 200) {
  return new Response(JSON.stringify(data), {status, headers:{"Content-Type":"application/json"}});
}

function htmlError(status, message) {
  return new Response(`<!doctype html><html><title>${message}</title></html>`, {status, headers:{"Content-Type":"text/html"}});
}

function fixture(context, lang = "zh") {
  let now = startTime;
  const routes = new Map();
  const requests = [];
  const unexpected = [];
  const confirmations = [];
  const notices = [];
  const status = {textContent:"", classList:{toggle(name, enabled) { status[name] = enabled; }}};
  const list = {innerHTML:"", contains:() => false};
  const document = {hidden:false, addEventListener() {}, querySelector:() => null, getElementById:id => ({"node-inventory-status":status, "node-list":list}[id])};
  const app = runInNewContext(appSource.replace(/\ninitialize\(\);\s*$/, "") + `
    render = () => {};
    connectSSE = () => {};
    toast = (...args) => notify(...args);
    syncMediaCheck = () => {};
    ({pollNodeInventory, nodesHTML, requestNodeUpdate, loadAll,
      node: () => state.nodes[0],
      setNode: value => { state.nodes = [value]; },
      releaseDue: () => { nextNodeReleaseCheck = 0; }})`, {
    localStorage:{getItem:key => ({prism_token:"fixture-token", enhancer_tab:"nodes", enhancer_lang:lang}[key]), setItem() {}},
    document, window:{scrollX:0, scrollY:0}, AbortController, FormData, setTimeout, clearTimeout,
    MutationObserver:class { observe() {} },
    Date:class extends Date { static now() { return now; } },
    notify:(...args) => notices.push(args),
    confirm:message => { confirmations.push(message); return context.confirmUpdate; },
    fetch:async (path, options) => {
      const method = options.method || "GET";
      requests.push({path, method, body:options.body});
      const respond = routes.get(`${method} ${path}`);
      if (!respond) {
        unexpected.push(`${method} ${path}`);
        throw new Error("Unexpected fixture request");
      }
      return respond();
    }
  });
  context.after(() => assert.deepEqual(unexpected, []));
  return {app, status, list, document, confirmations, notices, requests,
    route:(path, respond, method = "GET") => routes.set(`${method} ${path}`, respond),
    advance:milliseconds => { now += milliseconds; }};
}

function nodeFixtures(view, nodes = [managed]) {
  view.route(nodesPath, () => json(nodes));
  view.route(releasePath, () => json(release));
}

function updateDisabled(html) {
  return /<button[^>]+node-update-trigger[^>]+\bdisabled\b/.test(html);
}

test("an online but unenrolled node has no inferred versions, heartbeat or reported IPv4", async context => {
  const view = fixture(context);
  nodeFixtures(view, [node]);
  await view.app.pollNodeInventory();
  const page = view.app.nodesHTML();
  assert.match(page, /ONLINE/);
  assert.match(page, /尚未纳管/);
  assert.match(page, /原有安装或升级入口/);
  assert.match(page, /连接 IP \/ DNS 地址<\/span><strong>192\.0\.2\.25/);
  assert.equal((page.match(/<dd>未上报<\/dd>/g) || []).length, 5);
  assert.doesNotMatch(page, /<dd>v1\.5\.40|<dd>192\.0\.2\.25/);
  assert.ok(updateDisabled(page));
  await view.app.requestNodeUpdate(view.app.node());
  assert.equal(view.confirmations.length, 0);
});

test("reported fields stay separate from the configured address and survive worker expiration", async context => {
  const view = fixture(context, "en");
  nodeFixtures(view);
  await view.app.pollNodeInventory();
  assert.match(view.list.innerHTML, /Heartbeat current/);
  assert.match(view.list.innerHTML, /Prism version<\/dt><dd>v1\.5\.9/);
  assert.match(view.list.innerHTML, /Transport version<\/dt><dd>2\.5\.2/);
  assert.match(view.list.innerHTML, /Agent version<\/dt><dd>v1\.2\.1/);
  assert.match(view.list.innerHTML, /Reported IPv4<\/dt><dd>198\.51\.100\.9/);
  assert.equal(updateDisabled(view.list.innerHTML), false);

  view.advance(90000);
  const page = view.app.nodesHTML();
  assert.match(page, /Heartbeat expired/);
  assert.match(page, /Last update worker check-in<\/dt><dd>[^<]+/);
  assert.match(page, /Reported IPv4<\/dt><dd>198\.51\.100\.9/);
  assert.ok(updateDisabled(page));
  await view.app.requestNodeUpdate(view.app.node());
  assert.equal(view.confirmations.length, 0);
});

test("worker flags and invalid or non-upgrade data fail closed", async context => {
  const view = fixture(context, "en");
  nodeFixtures(view);
  await view.app.pollNodeInventory();
  for (const [changes, reason] of [
    [{update_supported:false}, /Heartbeat expired/],
    [{last_seen:"invalid"}, /Heartbeat unverified/],
    [{version:""}, /did not report a comparable Prism version/],
    [{version:"1.5.40"}, /No update needed/],
    [{version:"v1.6.0"}, /No update needed/]
  ]) {
    view.app.setNode({...managed, management:{...managed.management, ...changes}});
    const page = view.app.nodesHTML();
    assert.match(page, reason);
    assert.ok(updateDisabled(page));
    assert.doesNotMatch(page, /Invalid Date/);
  }
  view.app.setNode({...managed, management:{version:"v1.5.9", last_seen:checkedInAt, update_supported:true, ipv4:"2001:db8::1"}});
  const page = view.app.nodesHTML();
  assert.equal((page.match(/<dd>Not reported<\/dd>/g) || []).length, 3);
  assert.doesNotMatch(page, /2001:db8/);
});

test("first-load denial is distinct from a successful empty inventory", async context => {
  const view = fixture(context);
  assert.match(view.app.nodesHTML(), /正在读取节点/);
  view.route(nodesPath, () => json({error:"Forbidden"}, 403));
  view.route("/api/rules", () => json([]));
  view.route("/enhancer/api/catalog", () => json({services:[]}));
  view.route("/enhancer/api/ip-configs", () => json([]));
  view.route("/enhancer/api/categories", () => json({categories:[]}));
  await view.app.loadAll(true);
  assert.match(view.status.textContent, /节点刷新失败.*HTTP 403/);
  assert.match(view.list.innerHTML, /暂无可显示的节点数据/);
  assert.doesNotMatch(view.list.innerHTML, /还没有节点/);
  assert.equal(view.status["has-error"], true);

  nodeFixtures(view, []);
  await view.app.pollNodeInventory();
  assert.match(view.list.innerHTML, /还没有节点/);
  assert.equal(view.status["has-error"], false);
});

test("polling keeps last reports on a 502 and independently shows a release 403 until recovery", async context => {
  const view = fixture(context);
  nodeFixtures(view);
  await view.app.pollNodeInventory();
  view.app.releaseDue();
  view.route(nodesPath, () => htmlError(502, "502 Bad Gateway"));
  view.route(releasePath, () => json({error:"Forbidden"}, 403));
  await view.app.pollNodeInventory();
  assert.match(view.status.textContent, /节点刷新失败.*HTTP 502.*正在显示上次取得的数据/);
  assert.match(view.status.textContent, /暂时无法查询可用版本.*HTTP 403/);
  assert.doesNotMatch(view.status.textContent, /v1\.5\.40|<html>/);
  assert.match(view.list.innerHTML, /198\.51\.100\.9/);
  assert.ok(updateDisabled(view.list.innerHTML));

  view.advance(10000);
  nodeFixtures(view);
  await view.app.pollNodeInventory();
  assert.equal(view.status.textContent, "可用版本: v1.5.40");
  assert.equal(view.status["has-error"], false);
  assert.equal(updateDisabled(view.list.innerHTML), false);
});

test("a release gateway error, unsupported endpoint and invalid version are distinct", async context => {
  const view = fixture(context, "en");
  nodeFixtures(view);
  for (const [respond, reason] of [
    [() => htmlError(502, "502 Bad Gateway"), /Could not check the available release.*HTTP 502/],
    [() => json({}, 404), /does not support node updates/],
    [() => json({version:"latest"}), /Could not check the available release.*did not return valid data/]
  ]) {
    view.route(releasePath, respond);
    view.app.releaseDue();
    await view.app.pollNodeInventory();
    assert.match(view.status.textContent, reason);
    assert.doesNotMatch(view.status.textContent, /Available release: latest/);
    assert.ok(updateDisabled(view.list.innerHTML));
  }
});

test("submit only sends the confirmed node and release; errors never fabricate a queued job", async context => {
  const view = fixture(context);
  nodeFixtures(view);
  await view.app.pollNodeInventory();
  view.document.hidden = true;
  await view.app.requestNodeUpdate(view.app.node());
  assert.equal(view.requests.filter(request => request.method === "POST").length, 0);
  context.confirmUpdate = true;
  for (const [respond, detail] of [
    [() => json({error:"Forbidden"}, 403), /HTTP 403/],
    [() => htmlError(502, "502 Bad Gateway"), /HTTP 502/],
    [() => json({version:release.version, status:"queued"}, 202), /未返回有效数据/],
    [() => new Response(null, {status:204}), /未返回有效数据/]
  ]) {
    view.route(updatesPath, respond, "POST");
    await view.app.requestNodeUpdate(view.app.node());
    assert.match(view.list.innerHTML, /提交未获确认/);
    assert.match(view.list.innerHTML, detail);
    assert.doesNotMatch(view.list.innerHTML, /更新已排队|<html>/);
    assert.equal(view.app.node().management.update, undefined);
    assert.deepEqual(view.requests.at(-1), {path:updatesPath, method:"POST", body:JSON.stringify({node_id:node.id, version:release.version})});
  }
  assert.match(view.confirmations.at(-1), /Fixture node\nv1\.5\.9 → v1\.5\.40/);
});

test("confirmed job state comes from the server, not from the requested version", async context => {
  context.confirmUpdate = true;
  const view = fixture(context);
  nodeFixtures(view);
  await view.app.pollNodeInventory();
  const job = {id:"fixture-job", version:release.version, status:"queued", requested_at:checkedInAt, updated_at:checkedInAt};
  view.document.hidden = true;
  view.route(updatesPath, () => json(job, 202), "POST");
  await view.app.requestNodeUpdate(view.app.node());
  assert.match(view.list.innerHTML, /更新已排队 · v1\.5\.40/);
  assert.match(view.list.innerHTML, /提交时间: .*最后更新: /);
  assert.ok(updateDisabled(view.list.innerHTML));
  assert.equal(view.app.node().management.version, "v1.5.9");

  view.document.hidden = false;
  for (const [status, label, disabled] of [["running", /正在更新/, true], ["failed", /更新失败/, false], ["succeeded", /更新成功/, false]]) {
    view.route(nodesPath, () => json([{...managed, management:{...managed.management, update:{...job, status}}}]));
    await view.app.pollNodeInventory();
    assert.match(view.list.innerHTML, label);
    assert.equal(updateDisabled(view.list.innerHTML), disabled);
    assert.equal(view.app.node().management.version, "v1.5.9");
  }
});
