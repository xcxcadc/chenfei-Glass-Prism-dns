import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import {test} from "node:test";
import {runInNewContext} from "node:vm";

const source = readFileSync(new URL("../enhancer/web/app.js", import.meta.url), "utf8").replace(/\ninitialize\(\);\s*$/, "");
const service = {id:"netflix", name:"Netflix", domains:["netflix.com"]};
const config = {id:"ip-a", ip:"203.0.113.10", dns_node_id:"dns-a", routes:{netflix:"proxy-a"}};

function deferred() {
  let resolve;
  const promise = new Promise(done => { resolve = done; });
  return {promise, resolve};
}

function fixture() {
  const handlers = new Map();
  const renders = [];
  const requests = [];
  const document = {addEventListener() {}, getElementById:() => null, querySelector:() => null, querySelectorAll:() => []};
  const app = runInNewContext(`${source}
    render = () => onRender();
    renderModal = () => { renderedModal = state.modal; };
    connectSSE = () => {};
    toast = () => {};
    ({state, loadAll, setDNSNode, loadRoutingState, saveIPConfig, routedNodeForService,
      canAuditIPRoute, mediaCheckRecord, targetMediaServiceState, openServiceRoute})`, {
    document, onRender:() => renders.push("render"), AbortController, FormData, setTimeout, clearTimeout,
    MutationObserver:class { observe() {} },
    localStorage:{getItem:key => key === "prism_token" ? "fixture-session" : null, setItem() {}},
    fetch:async (path, options) => {
      const key = `${options.method || "GET"} ${path}`;
      requests.push({key, body:options.body});
      const response = await handlers.get(key)?.();
      return response instanceof Response ? response : Response.json(response ?? null);
    }
  });
  const route = (path, handler, method = "GET") => handlers.set(`${method} ${path}`, handler);
  route("/api/rules", () => []);
  route("/enhancer/api/catalog", () => ({services:[service]}));
  route("/enhancer/api/categories", () => ({categories:[]}));
  route("/enhancer/api/ip-configs", () => [config]);
  route("/enhancer/api/nodes", () => [{id:"dns-a", role:"dns", address:config.ip}, {id:"proxy-a", role:"proxy"}]);
  return {app, route, renders, requests};
}

test("an older panel refresh cannot overwrite a newer saved view or repaint an open draft", async () => {
  const {app, route, renders} = fixture();
  const slow = deferred();
  route("/enhancer/api/ip-configs", () => slow.promise);
  const old = app.loadAll(true);
  route("/enhancer/api/ip-configs", () => [{...config, note:"newer"}]);
  const modal = {type:"ip-form", draft:{note:"unsaved"}};
  app.state.modal = modal;
  await app.loadAll(true);
  slow.resolve([{...config, note:"older"}]);
  await old;
  assert.equal(app.state.ipConfigs[0].note, "newer");
  assert.equal(app.state.modal, modal);
  assert.equal(app.state.backgroundPending, true);
  assert.deepEqual(renders, []);
});

test("route refreshes are scoped to the current DNS node and failures preserve the latest known data", async () => {
  const {app, route} = fixture();
  const slow = deferred();
  route("/api/overrides?dns_node_id=dns-a", () => slow.promise);
  route("/api/active_selections?dns_node_id=dns-a", () => ({}));
  app.setDNSNode("dns-a");
  const old = app.loadRoutingState();
  app.setDNSNode("dns-b");
  route("/api/overrides?dns_node_id=dns-b", () => ({route:"proxy-b"}));
  route("/api/active_selections?dns_node_id=dns-b", () => ({}));
  assert.equal(await app.loadRoutingState(), true);
  slow.resolve({route:"proxy-a"});
  assert.equal(await old, false);
  route("/api/overrides?dns_node_id=dns-b", () => Response.json({error:"Gateway unavailable"}, {status:502}));
  assert.equal(await app.loadRoutingState(), false);
  assert.equal(app.state.overrides.route, "proxy-b");
  assert.ok(app.state.routingError);
});

test("managed routes never fall back to an unrelated Controller rule", async () => {
  const {app} = fixture();
  await app.loadAll(true);
  app.state.rules = [{id:"old-rule", name:"Stream · Netflix", value:"https://panel.test/enhancer/rules/netflix.list", target_type:"node", target_val:"proxy-a"}];
  assert.equal(app.routedNodeForService(service).id, "proxy-a");
  app.state.ipConfigs = [{...config, routes:{}}];
  assert.equal(app.routedNodeForService(service), null);
});

test("saving an empty existing route list sends one PUT and leaves a newer dialog open", async () => {
  const {app, requests, route} = fixture();
  await app.loadAll(true);
  const reply = deferred();
  const updated = {...config, routes:{}};
  route("/enhancer/api/ip-configs/ip-a", () => reply.promise, "PUT");
  route("/enhancer/api/ip-configs", () => [updated]);
  app.state.modal = {type:"ip-form", config, draft:{ip:config.ip}, routes:{}};
  const saving = app.saveIPConfig();
  await app.saveIPConfig();
  const otherModal = {type:"branding"};
  app.state.modal = otherModal;
  reply.resolve(updated);
  await saving;
  const writes = requests.filter(request => request.key.startsWith("PUT"));
  assert.equal(writes.length, 1);
  assert.deepEqual(JSON.parse(writes[0].body).routes, {});
  assert.equal(app.state.modal, otherModal);
  assert.equal(Object.keys(app.state.ipConfigs[0].routes).length, 0);
});

test("unsaved and previous-generation audit data cannot be attached to the new route", () => {
  const {app} = fixture();
  const initial = {...config, service_audit_requested_at:"2026-11-26T00:00:00Z"};
  app.state.ipConfigs = [initial];
  const record = app.mediaCheckRecord({config_id:config.id});
  record.job = {id:"old", generation:initial.service_audit_requested_at, status:"completed", results:{Netflix:"Yes"}};
  assert.equal(app.targetMediaServiceState(initial, service)?.kind, "good");
  assert.equal(app.canAuditIPRoute({config:initial, routes:{netflix:"proxy-b"}}, service), false);
  const updated = {...initial, routes:{netflix:"proxy-b"}, service_audit_requested_at:"2026-11-26T00:00:01Z"};
  app.state.ipConfigs = [updated];
  assert.equal(app.targetMediaServiceState(updated, service), null);
  assert.notEqual(app.mediaCheckRecord({config_id:config.id}), record);
});
