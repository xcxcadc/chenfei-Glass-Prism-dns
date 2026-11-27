import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import {test} from "node:test";
import {runInNewContext} from "node:vm";

const appSource = readFileSync(new URL("../enhancer/web/app.js", import.meta.url), "utf8");
const collection = "/enhancer/api/media-checks";
const payload = {node_id:"fixture-node"};
const query = `${collection}?${new URLSearchParams(payload)}`;
const job = {id:"fixture-job", status:"running", results:{Netflix:"Yes (Region: US)"}};
const jobPath = `${collection}/${job.id}`;

function json(data, status = 200) {
  return new Response(JSON.stringify(data), {status, headers:{"Content-Type":"application/json"}});
}

function gatewayError() {
  return new Response("<!doctype html><html><title>502 Bad Gateway</title></html>", {status:502, headers:{"Content-Type":"text/html"}});
}

function fixture(context, lang = "zh") {
  const requests = [];
  const routes = new Map();
  const unexpected = [];
  const timers = new Map();
  let nextTimer = 0;
  const document = {addEventListener() {}, getElementById:() => null, querySelectorAll:() => [], createElement:() => {
    const elements = new Map();
    return {dataset:{}, querySelector:selector => {
      if (!elements.has(selector)) elements.set(selector, {addEventListener(event, callback) { this[event] = callback; }});
      return elements.get(selector);
    }};
  }};
  const app = runInNewContext(appSource.replace(/\ninitialize\(\);\s*$/, "") + `
    state.mediaCheckView = {position:{details:{}}};
    ({mediaCheckRecord, mediaJobCard, mediaRequestStatus, syncMediaCheck})`, {
    localStorage:{getItem:key => ({prism_token:"fixture-token", enhancer_lang:lang}[key])},
    document, AbortController, FormData, URLSearchParams,
    MutationObserver:class { observe() {} },
    setTimeout:(callback, delay) => { timers.set(++nextTimer, {callback, delay}); return nextTimer; },
    clearTimeout:id => timers.delete(id),
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
  const record = app.mediaCheckRecord(payload);
  context.after(() => assert.deepEqual(unexpected, []));
  return {record, requests,
    sync:operation => app.syncMediaCheck(record, operation),
    lookupButton:() => app.mediaJobCard(record).querySelector(".media-reconnect"),
    status:() => app.mediaRequestStatus(record),
    route:(path, respond, method = "GET") => routes.set(`${method} ${path}`, respond),
    delays:() => [...timers.values()].map(timer => timer.delay),
    tick:async () => {
      assert.equal(timers.size, 1);
      const [id, timer] = timers.entries().next().value;
      timers.delete(id);
      await timer.callback();
    }};
}

test("a failed creation is explicit and neither reconnect nor reopening can resend its POST", async context => {
  const view = fixture(context);
  view.route(query, () => json(null));
  view.route(collection, gatewayError, "POST");
  await view.sync("latest");
  assert.equal(view.status().label, "媒体任务建单失败（未确认）");
  assert.match(view.status().messages.join(" "), /HTTP 502.*不会自动重发建单请求/);
  assert.deepEqual(view.delays(), []);
  assert.deepEqual(view.requests.at(-1), {path:collection, method:"POST", body:JSON.stringify(payload)});

  await view.sync("read");
  const lookup = view.lookupButton();
  assert.equal(lookup.textContent, "重新查询任务");
  await lookup.click();
  await view.sync("latest");
  assert.deepEqual(view.requests.map(request => request.method), ["GET", "POST", "GET", "GET"]);
  assert.equal(view.status().label, "媒体任务建单失败（未确认）");
  assert.deepEqual(view.delays(), []);

  view.route(query, () => json(job));
  await view.sync("resume");
  assert.equal(view.status().label, "");
  assert.equal(view.record.job.id, job.id);
  assert.deepEqual(view.delays(), [1000]);
});

test("an invalid successful POST response also leaves creation unconfirmed", async context => {
  const view = fixture(context, "en");
  view.route(collection, () => new Response(null, {status:204}), "POST");
  await view.sync("create");
  assert.equal(view.record.job, null);
  assert.match(view.status().messages.join(" "), /Media job creation failed.*Invalid media job response/);
  assert.deepEqual(view.delays(), []);
});

test("ongoing jobs poll every second when healthy, then back off and pause on persistent 502", async context => {
  const view = fixture(context);
  view.route(query, () => json(job));
  view.route(jobPath, gatewayError);
  await view.sync("latest");
  assert.deepEqual(view.delays(), [1000]);
  for (const delay of [2000, 5000, 15000]) {
    await view.tick();
    assert.match(view.status().messages.join(" "), /媒体任务轮询失败.*HTTP 502.*稍后自动重查/);
    assert.deepEqual(view.delays(), [delay]);
  }
  await view.tick();
  assert.equal(view.status().label, "媒体任务轮询失败");
  assert.match(view.status().messages.join(" "), /HTTP 502.*已暂停自动重查/);
  assert.equal(view.record.job.results.Netflix, job.results.Netflix);
  assert.deepEqual(view.delays(), []);
  await view.sync("read");
  assert.equal(view.requests.length, 5);

  view.route(query, () => json({...job, status:"completed", results:{YouTube:"Yes"}}));
  await view.sync("resume");
  assert.ok(view.requests.every(request => request.method === "GET"));
  assert.equal(view.record.job.status, "completed");
  assert.equal(view.record.job.results.Netflix, job.results.Netflix);
  assert.equal(view.status().label, "");
  assert.deepEqual(view.delays(), []);
});

test("a denied task lookup pauses immediately and never falls back to creating a job", async context => {
  const view = fixture(context, "en");
  view.route(query, () => json({error:"Forbidden"}, 403));
  await view.sync("latest");
  assert.match(view.status().messages.join(" "), /Media job polling failed.*HTTP 403.*Automatic checks paused/);
  assert.deepEqual(view.delays(), []);
  assert.deepEqual(view.requests.map(request => request.method), ["GET"]);
});

test("recovering a missing job looks up the latest without creating a replacement", async context => {
  const view = fixture(context);
  view.route(query, () => json(job));
  view.route(jobPath, () => json({error:"Missing"}, 404));
  await view.sync("latest");
  view.route(query, () => json(null));
  await view.tick();
  assert.deepEqual(view.delays(), [2000]);
  await view.tick();
  assert.equal(view.status().label, "任务记录丢失");
  assert.match(view.status().messages.join(" "), /已保留上次结果/);
  assert.equal(view.record.job.results.Netflix, job.results.Netflix);
  assert.deepEqual(view.requests.map(request => request.path), [query, jobPath, query]);
  assert.deepEqual(view.delays(), []);
});
