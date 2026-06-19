package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>specter</title>
<style>
:root{
  --bg:#08111f;
  --bg-soft:#101a2c;
  --panel:#121d31;
  --panel-2:#0e1728;
  --border:#22324d;
  --text:#e8f1ff;
  --muted:#88a0c1;
  --accent:#67e8f9;
  --accent-2:#22c55e;
  --warn:#fbbf24;
  --danger:#f87171;
  --shadow:0 18px 50px rgba(0,0,0,.28);
}
*{box-sizing:border-box;margin:0;padding:0}
body{
  min-height:100vh;
  background:
    radial-gradient(circle at top right, rgba(103,232,249,.12), transparent 28%),
    radial-gradient(circle at left 20%, rgba(34,197,94,.08), transparent 24%),
    linear-gradient(180deg, #091120 0%, #0b1423 100%);
  color:var(--text);
  font:13px/1.45 "Avenir Next","Segoe UI","Helvetica Neue",sans-serif;
}
button,input,textarea{font:inherit}
code,pre,.mono,.method{font-family:"SFMono-Regular","SF Mono",Consolas,Monaco,monospace}
button{
  border:1px solid var(--border);
  background:var(--panel);
  color:var(--text);
  cursor:pointer;
  border-radius:8px;
  padding:8px 12px;
  transition:background .15s ease,border-color .15s ease,transform .15s ease;
}
button:hover{background:#18253b;border-color:#314a70}
button:active{transform:translateY(1px)}
button.ghost{background:transparent}
button.warn{border-color:rgba(248,113,113,.45);color:#ffd0d0}
button.good{border-color:rgba(34,197,94,.35);color:#d5ffe3}
button.small{padding:6px 10px;font-size:12px}
button.active-toggle{background:#17324a;border-color:#3e6c91;color:#b8f5ff}
input,textarea{
  width:100%;
  background:var(--panel-2);
  color:var(--text);
  border:1px solid var(--border);
  border-radius:8px;
  padding:10px 12px;
}
textarea{min-height:220px;resize:vertical}
header{
  padding:22px 24px 14px;
  border-bottom:1px solid rgba(136,160,193,.12);
}
.brand-row{
  display:flex;
  flex-wrap:wrap;
  align-items:flex-end;
  justify-content:space-between;
  gap:16px;
}
.brand h1{
  font-size:28px;
  font-weight:700;
  letter-spacing:.03em;
}
.brand p{
  color:var(--muted);
  margin-top:4px;
}
.meta{
  display:flex;
  flex-wrap:wrap;
  gap:10px;
  align-items:center;
}
.pill,.metric{
  display:inline-flex;
  align-items:center;
  gap:8px;
  border:1px solid rgba(136,160,193,.14);
  background:rgba(18,29,49,.86);
  border-radius:999px;
  padding:7px 12px;
  color:var(--muted);
}
.metric strong,.pill strong{color:var(--text);font-weight:600}
.toolbar{
  display:flex;
  flex-wrap:wrap;
  gap:10px;
  align-items:center;
}
.topbar{
  display:flex;
  flex-wrap:wrap;
  gap:12px;
  justify-content:space-between;
  padding:14px 24px 0;
}
.flash{
  margin:14px 24px 0;
  min-height:22px;
  color:var(--muted);
}
.flash.error{color:#ffd0d0}
.flash.success{color:#c8ffe0}
.flash.info{color:#b8f5ff}
nav{
  display:flex;
  flex-wrap:wrap;
  gap:8px;
  padding:18px 24px 0;
}
nav button{
  background:transparent;
  color:var(--muted);
  border-color:transparent;
}
nav button.active{
  background:var(--panel);
  border-color:var(--border);
  color:var(--text);
}
main{padding:18px 24px 24px}
.tab{display:none}
.tab.active{display:block}
.split{
  display:grid;
  grid-template-columns:minmax(0,1.65fr) minmax(300px,1fr);
  gap:16px;
}
.stack{
  display:grid;
  gap:16px;
}
.panel{
  background:rgba(18,29,49,.88);
  border:1px solid rgba(136,160,193,.12);
  border-radius:12px;
  box-shadow:var(--shadow);
  overflow:hidden;
}
.panel-head{
  display:flex;
  justify-content:space-between;
  align-items:flex-start;
  gap:12px;
  padding:16px 18px;
  border-bottom:1px solid rgba(136,160,193,.1);
}
.panel-head h2{
  font-size:15px;
  font-weight:650;
}
.panel-head p{
  color:var(--muted);
  margin-top:4px;
}
.panel-body{padding:16px 18px}
.panel-body.tight{padding:0}
.stats{
  display:grid;
  grid-template-columns:repeat(auto-fit,minmax(140px,1fr));
  gap:12px;
}
.metric{
  min-height:74px;
  border-radius:12px;
  display:flex;
  flex-direction:column;
  align-items:flex-start;
  justify-content:center;
  padding:14px 16px;
}
.metric span{font-size:11px;text-transform:uppercase;letter-spacing:.08em}
.metric strong{font-size:22px}
table{width:100%;border-collapse:collapse}
th,td{text-align:left;padding:10px 12px;border-bottom:1px solid rgba(136,160,193,.08);vertical-align:top}
th{
  color:var(--muted);
  font-size:11px;
  text-transform:uppercase;
  letter-spacing:.08em;
  font-weight:600;
}
tbody tr{cursor:pointer;transition:background .15s ease}
tbody tr:hover{background:rgba(103,232,249,.05)}
tbody tr.active{background:rgba(103,232,249,.09)}
.scroll{max-height:560px;overflow:auto}
.empty{padding:24px 12px;color:var(--muted)}
.method{
  display:inline-block;
  min-width:60px;
  padding:4px 8px;
  border-radius:999px;
  font-size:11px;
  font-weight:700;
  text-align:center;
  border:1px solid rgba(136,160,193,.15);
}
.GET{color:#9ae6b4}
.POST{color:#9bd3ff}
.PUT{color:#ffd27d}
.PATCH{color:#d0b3ff}
.DELETE{color:#ffadad}
.tag{
  display:inline-flex;
  align-items:center;
  gap:6px;
  border-radius:999px;
  border:1px solid rgba(136,160,193,.12);
  background:rgba(255,255,255,.03);
  color:var(--muted);
  padding:4px 8px;
  font-size:11px;
}
.tag-row{display:flex;flex-wrap:wrap;gap:8px}
.detail{
  display:grid;
  gap:14px;
}
.detail h3{
  font-size:11px;
  text-transform:uppercase;
  letter-spacing:.08em;
  color:var(--muted);
  margin-bottom:6px;
}
.detail-card{
  background:rgba(10,18,32,.38);
  border:1px solid rgba(136,160,193,.08);
  border-radius:10px;
  padding:12px;
}
.detail-card pre{
  white-space:pre-wrap;
  word-break:break-word;
  color:#dbe9ff;
}
.kv{
  display:grid;
  grid-template-columns:140px 1fr;
  gap:8px 12px;
}
.kv .key{color:var(--muted)}
.field-grid{
  display:grid;
  grid-template-columns:1fr 1fr auto;
  gap:10px;
  align-items:end;
}
.field-grid.two{
  grid-template-columns:1fr auto auto;
}
.list{
  display:grid;
  gap:10px;
}
.list-item{
  border:1px solid rgba(136,160,193,.1);
  background:rgba(10,18,32,.25);
  border-radius:10px;
  padding:12px;
}
.list-item.active{border-color:#3e6c91;background:rgba(23,50,74,.45)}
.list-head{
  display:flex;
  justify-content:space-between;
  gap:10px;
  align-items:flex-start;
}
.list-head h4{font-size:14px}
.list-head p{color:var(--muted);margin-top:2px}
.table-actions{
  display:flex;
  gap:8px;
  align-items:center;
}
.hint{color:var(--muted);font-size:12px}
.status-dot{
  width:10px;
  height:10px;
  border-radius:50%;
  background:var(--accent-2);
  box-shadow:0 0 0 4px rgba(34,197,94,.12);
}
.status-dot.off{
  background:#42556f;
  box-shadow:none;
}
.inline{
  display:flex;
  gap:10px;
  align-items:center;
  flex-wrap:wrap;
}
.right{margin-left:auto}
@media (max-width:980px){
  .split{grid-template-columns:1fr}
}
@media (max-width:720px){
  header,.topbar,nav,main,.flash{padding-left:14px;padding-right:14px}
  .field-grid,.field-grid.two{grid-template-columns:1fr}
  .kv{grid-template-columns:1fr}
  .method{min-width:52px}
  th:nth-child(5),td:nth-child(5),th:nth-child(6),td:nth-child(6){display:none}
}
</style>
</head>
<body>
<header>
  <div class="brand-row">
    <div class="brand">
      <h1>👻 specter</h1>
      <p>Local mock API control room</p>
    </div>
    <div class="meta">
      <span class="pill"><strong>API</strong> <span id="api-addr"></span></span>
      <span class="pill"><strong>Updated</strong> <span id="last-update">never</span></span>
    </div>
  </div>
</header>

<div class="topbar">
  <div class="toolbar">
    <button onclick="refreshAll()">Refresh</button>
    <button id="autorefresh-btn" class="ghost active-toggle" onclick="toggleAutoRefresh()">Auto Refresh On</button>
    <button class="warn" onclick="resetAll()">Reset All</button>
  </div>
  <div class="toolbar">
    <span class="pill"><strong>Requests</strong> <span id="quick-requests">0</span></span>
    <span class="pill"><strong>Routes</strong> <span id="quick-routes">0</span></span>
    <span class="pill"><strong>Timelines</strong> <span id="quick-timelines">0</span></span>
    <span class="pill"><strong>Stores</strong> <span id="quick-stores">0</span></span>
    <span class="pill"><strong>State</strong> <span id="quick-state">(none)</span></span>
  </div>
</div>

<div id="flash" class="flash"></div>

<nav>
  <button id="nav-requests" class="active" onclick="showTab('requests')">Requests</button>
  <button id="nav-routes" onclick="showTab('routes')">Routes</button>
  <button id="nav-state" onclick="showTab('state')">State &amp; Vars</button>
  <button id="nav-stores" onclick="showTab('stores')">Stores</button>
  <button id="nav-config" onclick="showTab('config')">Config</button>
</nav>

<main>
  <section class="tab active" id="tab-requests">
    <div class="split">
      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Recorded Requests</h2>
            <p>Click any row to inspect headers, query, and body.</p>
          </div>
          <div class="table-actions">
            <span class="tag" id="req-count">0 requests</span>
            <button class="small warn" onclick="clearHistory()">Clear History</button>
          </div>
        </div>
        <div class="panel-body tight scroll">
          <table>
            <thead><tr><th>#</th><th>Time</th><th>Method</th><th>Path</th><th>Query</th><th>Body</th></tr></thead>
            <tbody id="req-body"></tbody>
          </table>
        </div>
      </div>

      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Request Detail</h2>
            <p>Selected request from history.</p>
          </div>
          <span class="tag" id="req-selected">none</span>
        </div>
        <div class="panel-body">
          <div id="request-detail" class="detail"></div>
        </div>
      </div>
    </div>
  </section>

  <section class="tab" id="tab-routes">
    <div class="split">
      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Registered Routes</h2>
            <p>Config routes and any dynamic routes created at runtime.</p>
          </div>
          <div class="table-actions">
            <span class="tag" id="route-count">0 routes</span>
            <button class="small good" onclick="newRoute()">New Route</button>
            <button class="small warn" onclick="clearDynamicRoutes()">Clear Dynamic</button>
          </div>
        </div>
        <div class="panel-body tight scroll">
          <table>
            <thead><tr><th>Method</th><th>Path</th><th>Source</th><th>State</th><th>Features</th></tr></thead>
            <tbody id="routes-body"></tbody>
          </table>
        </div>
      </div>

      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Route Detail</h2>
            <p>See the resolved route config as JSON.</p>
          </div>
          <div class="table-actions">
            <button id="edit-route-btn" class="small good" style="display:none" onclick="editSelectedRoute()">Edit</button>
            <button id="delete-route-btn" class="small warn" style="display:none" onclick="deleteSelectedRoute()">Delete</button>
          </div>
        </div>
        <div class="panel-body">
          <div id="route-detail" class="detail"></div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Dynamic Route Editor</h2>
            <p>Add or update runtime-only routes.</p>
          </div>
          <span class="tag" id="route-editor-mode">new route</span>
        </div>
        <div class="panel-body">
          <textarea id="route-json" spellcheck="false"></textarea>
          <div class="inline" style="margin-top:10px">
            <span class="hint" id="route-editor-hint">Create a dynamic route from JSON.</span>
            <button class="small ghost right" onclick="resetRouteEditor()">Reset</button>
            <button class="small good" onclick="saveRoute()">Save Route</button>
          </div>
        </div>
      </div>
    </div>
  </section>

  <section class="tab" id="tab-state">
    <div class="split">
      <div class="stack">
        <div class="panel">
          <div class="panel-head">
            <div>
              <h2>Current State</h2>
              <p>Override the active state used by stateful routes.</p>
            </div>
            <span class="pill"><span id="state-dot" class="status-dot off"></span><span id="state-val">(none)</span></span>
          </div>
          <div class="panel-body">
            <div class="field-grid two">
              <input id="state-input" type="text" placeholder="logged_in">
              <button class="good" onclick="saveState()">Save State</button>
              <button class="ghost" onclick="clearState()">Clear</button>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-head">
            <div>
              <h2>Vars</h2>
              <p>Select a row to edit or create a new key/value pair.</p>
            </div>
            <div class="table-actions">
              <span class="tag" id="var-count">0 vars</span>
              <button class="small warn" onclick="clearVars()">Clear Vars</button>
            </div>
          </div>
          <div class="panel-body">
            <div class="field-grid">
              <input id="var-key" type="text" placeholder="role">
              <input id="var-value" type="text" placeholder="admin">
              <button class="good" onclick="saveVar()">Save Var</button>
            </div>
            <div class="inline" style="margin-top:10px">
              <span class="hint" id="var-editor-hint">Create or update a var.</span>
              <button class="small ghost right" onclick="clearVarEditor()">Clear Selection</button>
            </div>
            <div id="vars-body" class="list" style="margin-top:14px"></div>
          </div>
        </div>
      </div>

      <div class="stack">
        <div class="panel">
          <div class="panel-head">
            <div>
              <h2>Quick Reset</h2>
              <p>Reset specific pieces without clearing everything.</p>
            </div>
          </div>
          <div class="panel-body">
            <div class="stats">
              <button onclick="resetTargets(['state'])">Reset State</button>
              <button onclick="resetTargets(['vars'])">Reset Vars</button>
              <button onclick="resetTargets(['history'])">Reset History</button>
              <button onclick="resetTargets(['stores'])">Reset Stores</button>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-head">
            <div>
              <h2>Timelines</h2>
              <p>Route progress for multi-step responses.</p>
            </div>
            <div class="table-actions">
              <span class="tag" id="timeline-count">0 timelines</span>
              <button class="small warn" onclick="resetTargets(['timelines'])">Reset Timelines</button>
            </div>
          </div>
          <div class="panel-body">
            <div id="timelines-body" class="list"></div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-head">
            <div>
              <h2>Current Snapshot</h2>
              <p>Useful when you want a quick read without leaving the tab.</p>
            </div>
          </div>
          <div class="panel-body">
            <div id="state-summary" class="detail"></div>
          </div>
        </div>
      </div>
    </div>
  </section>

  <section class="tab" id="tab-stores">
    <div class="split">
      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Collections</h2>
            <p>Inspect or select a collection to edit its JSON data.</p>
          </div>
          <span class="tag" id="store-count">0 stores</span>
        </div>
        <div class="panel-body">
          <div id="stores-list" class="list"></div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Store Editor</h2>
            <p>Replace the entire collection with a JSON array.</p>
          </div>
          <div class="table-actions">
            <button class="small ghost" onclick="newStore()">New</button>
            <button class="small warn" id="clear-store-btn" style="display:none" onclick="clearSelectedStore()">Clear</button>
          </div>
        </div>
        <div class="panel-body">
          <div class="stack" id="store-editor">
            <input id="store-name" type="text" placeholder="users">
            <textarea id="store-json" spellcheck="false">[]</textarea>
            <div class="inline">
              <button class="good" onclick="saveStore()">Save Collection</button>
              <span class="hint">The payload must be a JSON array of objects.</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>

  <section class="tab" id="tab-config">
    <div class="split">
      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Config Playground</h2>
            <p>Paste YAML to validate and preview registered routes.</p>
          </div>
          <button class="small good" onclick="validateConfigPlayground()">Validate</button>
        </div>
        <div class="panel-body">
          <textarea id="config-yaml" spellcheck="false">routes:
  - path: /hello
    method: GET
    response:
      message: Hello</textarea>
          <div class="inline" style="margin-top:10px">
            <span class="hint" id="config-validator-hint">Validation has not run yet.</span>
            <button class="small ghost right" onclick="loadCurrentConfigSample()">Sample</button>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-head">
          <div>
            <h2>Validation Result</h2>
            <p>Errors, routes, scenarios, and seeded stores.</p>
          </div>
          <span class="tag" id="config-validity">not checked</span>
        </div>
        <div class="panel-body">
          <div id="config-result" class="detail"></div>
        </div>
      </div>
    </div>
  </section>
</main>

<script>
const API='{{API}}';
const ui = {
  requests: [],
  routes: [],
  timelines: [],
  stores: [],
  autoRefresh: true,
  selectedRequest: null,
  selectedRoute: null,
  editingRouteID: '',
  selectedVar: '',
  selectedStore: '',
  flashTimer: null
};

document.getElementById('api-addr').textContent = API;

function esc(s){
  return String(s)
    .replace(/&/g,'&amp;')
    .replace(/</g,'&lt;')
    .replace(/>/g,'&gt;')
    .replace(/"/g,'&quot;')
    .replace(/'/g,'&#39;');
}

function pretty(v){
  if (v === null || v === undefined || v === '') return '(empty)';
  if (typeof v === 'string') {
    try {
      return JSON.stringify(JSON.parse(v), null, 2);
    } catch (e) {
      return v;
    }
  }
  return JSON.stringify(v, null, 2);
}

function methodBadge(method){
  return '<span class="method '+esc(method)+'">'+esc(method)+'</span>';
}

function fmtTime(t){
  const d = new Date(t);
  return d.toLocaleTimeString([], {hour12:false}) + '.' + String(d.getMilliseconds()).padStart(3,'0');
}

function showFlash(message, tone){
  const el = document.getElementById('flash');
  el.textContent = message || '';
  el.className = 'flash ' + (tone || 'info');
  clearTimeout(ui.flashTimer);
  if (message) {
    ui.flashTimer = setTimeout(function(){
      el.textContent = '';
      el.className = 'flash';
    }, 2800);
  }
}

async function fetchJSON(path, options){
  const res = await fetch(API + path, options);
  const text = await res.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch (e) {
      data = text;
    }
  }
  if (!res.ok) {
    let msg = res.status + ' ' + res.statusText;
    if (data && typeof data === 'object' && data.error) msg = data.error;
    throw new Error(msg);
  }
  return data;
}

async function sendJSON(path, method, body){
  return fetchJSON(path, {
    method: method,
    headers: {'Content-Type': 'application/json'},
    body: body === undefined ? undefined : JSON.stringify(body)
  });
}

function showTab(name){
  document.querySelectorAll('.tab').forEach(function(tab){ tab.classList.remove('active'); });
  document.querySelectorAll('nav button').forEach(function(btn){ btn.classList.remove('active'); });
  document.getElementById('tab-' + name).classList.add('active');
  document.getElementById('nav-' + name).classList.add('active');
}

function routeFeatures(rt){
  const notes = [];
  if (rt.delay || rt.delay_min || rt.delay_max) notes.push('delay');
  if (rt.rate_limit) notes.push('rate limit');
  if (rt.fault || rt.error_rate) notes.push('faults');
  if (rt.proxy) notes.push('proxy');
  if (rt.stream) notes.push('stream');
  if (rt.match && rt.match.length) notes.push('match');
  if (rt.timeline && rt.timeline.length) notes.push('timeline');
  if (rt.responses && rt.responses.length) notes.push('responses');
  if (rt.store_push || rt.store_list || rt.store_get || rt.store_put || rt.store_patch || rt.store_delete || rt.store_clear) notes.push('store');
  if (rt.redirect) notes.push('redirect');
  return notes;
}

function updateHeaderMetrics(){
  document.getElementById('quick-requests').textContent = String(ui.requests.length);
  document.getElementById('quick-routes').textContent = String(ui.routes.length);
  document.getElementById('quick-timelines').textContent = String(ui.timelines.length);
  document.getElementById('quick-stores').textContent = String(ui.stores.length);
  const stateValue = document.getElementById('state-val').textContent || '(none)';
  document.getElementById('quick-state').textContent = stateValue;
}

function renderRequestDetail(){
  const el = document.getElementById('request-detail');
  const badge = document.getElementById('req-selected');
  const index = ui.selectedRequest;
  if (index === null || !ui.requests[index]) {
    badge.textContent = 'none';
    el.innerHTML = '<div class="detail-card"><p class="hint">Select a request to inspect it here.</p></div>';
    return;
  }
  const req = ui.requests[index];
  badge.textContent = '#' + (index + 1);
  el.innerHTML =
    '<div class="detail-card"><div class="inline"><strong>' + methodBadge(req.method) + '</strong><span>' + esc(req.path) + '</span></div><div class="hint" style="margin-top:8px">' + esc(req.time) + '</div></div>' +
    '<div class="detail-card"><h3>Query</h3><pre>' + esc(pretty(req.query || {})) + '</pre></div>' +
    '<div class="detail-card"><h3>Headers</h3><pre>' + esc(pretty(req.headers || {})) + '</pre></div>' +
    '<div class="detail-card"><h3>Body</h3><pre>' + esc(pretty(req.body || '')) + '</pre></div>';
}

function selectRequest(index){
  ui.selectedRequest = index;
  renderRequests();
  renderRequestDetail();
}

function renderRequests(){
  const tbody = document.getElementById('req-body');
  document.getElementById('req-count').textContent = ui.requests.length + ' request' + (ui.requests.length === 1 ? '' : 's');
  if (!ui.requests.length) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">No requests recorded yet.</td></tr>';
    ui.selectedRequest = null;
    renderRequestDetail();
    updateHeaderMetrics();
    return;
  }
  tbody.innerHTML = ui.requests.slice().reverse().map(function(req, i){
    const originalIndex = ui.requests.length - 1 - i;
    const query = req.query && Object.keys(req.query).length ? Object.entries(req.query).map(function(entry){
      return esc(entry[0]) + '=' + esc(entry[1]);
    }).join('&') : '—';
    const body = req.body ? esc(pretty(req.body)) : '—';
    const active = ui.selectedRequest === originalIndex ? ' class="active"' : '';
    return '<tr' + active + ' onclick="selectRequest(' + originalIndex + ')">' +
      '<td>' + (originalIndex + 1) + '</td>' +
      '<td>' + fmtTime(req.time) + '</td>' +
      '<td>' + methodBadge(req.method) + '</td>' +
      '<td>' + esc(req.path) + '</td>' +
      '<td>' + query + '</td>' +
      '<td><pre>' + body + '</pre></td>' +
      '</tr>';
  }).join('');
  renderRequestDetail();
  updateHeaderMetrics();
}

async function loadRequests(){
  ui.requests = await fetchJSON('/__specter/requests') || [];
  if (ui.selectedRequest === null && ui.requests.length) ui.selectedRequest = ui.requests.length - 1;
  if (ui.selectedRequest !== null && !ui.requests[ui.selectedRequest]) ui.selectedRequest = ui.requests.length ? ui.requests.length - 1 : null;
  renderRequests();
}

async function clearHistory(){
  if (!confirm('Clear recorded request history?')) return;
  await fetchJSON('/__specter/requests', {method:'DELETE'});
  showFlash('Request history cleared.', 'success');
  await loadRequests();
}

function renderRouteDetail(){
  const el = document.getElementById('route-detail');
  const btn = document.getElementById('delete-route-btn');
  const editBtn = document.getElementById('edit-route-btn');
  const route = ui.routes.find(function(item){ return item._uiKey === ui.selectedRoute; });
  if (!route) {
    btn.style.display = 'none';
    editBtn.style.display = 'none';
    el.innerHTML = '<div class="detail-card"><p class="hint">Select a route to inspect the resolved config.</p></div>';
    return;
  }
  const rt = route.route || route;
  const features = routeFeatures(rt).map(function(note){ return '<span class="tag">' + esc(note) + '</span>'; }).join(' ');
  btn.style.display = route.source === 'dynamic' ? 'inline-flex' : 'none';
  editBtn.style.display = 'inline-flex';
  el.innerHTML =
    '<div class="detail-card"><div class="inline">' + methodBadge(rt.method) + '<strong>' + esc(rt.path) + '</strong><span class="tag">' + esc(route.source || 'config') + '</span></div><div class="tag-row" style="margin-top:10px">' + (features || '<span class="hint">No special route features.</span>') + '</div></div>' +
    '<div class="detail-card"><h3>Route JSON</h3><pre>' + esc(pretty(rt)) + '</pre></div>';
}

function selectRoute(key){
  ui.selectedRoute = key;
  renderRoutes();
  renderRouteDetail();
}

function renderRoutes(){
  const tbody = document.getElementById('routes-body');
  document.getElementById('route-count').textContent = ui.routes.length + ' route' + (ui.routes.length === 1 ? '' : 's');
  if (!ui.routes.length) {
    tbody.innerHTML = '<tr><td colspan="5" class="empty">No routes configured.</td></tr>';
    ui.selectedRoute = null;
    renderRouteDetail();
    updateHeaderMetrics();
    return;
  }
  tbody.innerHTML = ui.routes.map(function(item){
    const rt = item.route || item;
    const features = routeFeatures(rt).map(function(note){ return '<span class="tag">' + esc(note) + '</span>'; }).join(' ');
    const active = ui.selectedRoute === item._uiKey ? ' class="active"' : '';
    return '<tr' + active + ' onclick="selectRoute(\'' + esc(item._uiKey) + '\')">' +
      '<td>' + methodBadge(rt.method) + '</td>' +
      '<td>' + esc(rt.path) + '</td>' +
      '<td><span class="tag">' + esc(item.source || 'config') + '</span></td>' +
      '<td>' + esc(rt.state || '—') + '</td>' +
      '<td>' + (features || '<span class="hint">basic</span>') + '</td>' +
      '</tr>';
  }).join('');
  renderRouteDetail();
  updateHeaderMetrics();
}

async function loadRoutes(){
  const routes = await fetchJSON('/__specter/routes') || [];
  ui.routes = routes.map(function(item, index){
    item._uiKey = item.id || 'config-' + index;
    return item;
  });
  if (ui.selectedRoute === null && ui.routes.length) ui.selectedRoute = ui.routes[0]._uiKey;
  if (ui.selectedRoute !== null && !ui.routes.find(function(item){ return item._uiKey === ui.selectedRoute; })) {
    ui.selectedRoute = ui.routes.length ? ui.routes[0]._uiKey : null;
  }
  renderRoutes();
}

function routeTemplate(){
  return {
    path: '/example',
    method: 'GET',
    status: 200,
    response: {ok: true}
  };
}

function setRouteEditor(route, id, label){
  ui.editingRouteID = id || '';
  document.getElementById('route-json').value = JSON.stringify(route || routeTemplate(), null, 2);
  document.getElementById('route-editor-mode').textContent = label || (id ? 'editing dynamic' : 'new route');
  document.getElementById('route-editor-hint').textContent = id ? 'Editing dynamic route ' + id + '.' : 'Create a dynamic route from JSON.';
}

function resetRouteEditor(){
  setRouteEditor(routeTemplate(), '', 'new route');
}

function newRoute(){
  ui.selectedRoute = null;
  renderRoutes();
  setRouteEditor(routeTemplate(), '', 'new route');
}

function editSelectedRoute(){
  const route = ui.routes.find(function(item){ return item._uiKey === ui.selectedRoute; });
  if (!route) return;
  const rt = route.route || route;
  if (route.source === 'dynamic' && route.id) {
    setRouteEditor(rt, route.id, 'editing dynamic');
  } else {
    setRouteEditor(rt, '', 'copy as dynamic');
    document.getElementById('route-editor-hint').textContent = 'Config routes cannot be edited in memory; saving creates a dynamic copy.';
  }
}

function validateRoutePayload(route){
  if (!route || typeof route !== 'object' || Array.isArray(route)) return 'Route JSON must be an object.';
  if (!route.path || typeof route.path !== 'string') return 'Route JSON must include a string path.';
  if (!route.method || typeof route.method !== 'string') return 'Route JSON must include a string method.';
  return '';
}

async function saveRoute(){
  let route;
  try {
    route = JSON.parse(document.getElementById('route-json').value || '{}');
  } catch (e) {
    showFlash('Route JSON must be valid JSON.', 'error');
    return;
  }
  const validation = validateRoutePayload(route);
  if (validation) {
    showFlash(validation, 'error');
    return;
  }
  try {
    if (ui.editingRouteID) {
      await sendJSON('/__specter/routes/' + encodeURIComponent(ui.editingRouteID), 'PUT', route);
      showFlash('Dynamic route updated.', 'success');
    } else {
      const res = await sendJSON('/__specter/routes', 'POST', route);
      ui.editingRouteID = res && res.id ? res.id : '';
      showFlash('Dynamic route added.', 'success');
    }
  } catch (e) {
    showFlash(e.message || 'Route save failed.', 'error');
    return;
  }
  await loadRoutes();
  if (ui.editingRouteID) ui.selectedRoute = ui.editingRouteID;
  renderRoutes();
  editSelectedRoute();
}

async function clearDynamicRoutes(){
  if (!confirm('Remove all dynamic routes?')) return;
  await fetchJSON('/__specter/routes', {method:'DELETE'});
  resetRouteEditor();
  showFlash('Dynamic routes cleared.', 'success');
  await loadRoutes();
}

async function deleteSelectedRoute(){
  const route = ui.routes.find(function(item){ return item._uiKey === ui.selectedRoute; });
  if (!route || route.source !== 'dynamic' || !route.id) return;
  if (!confirm('Delete dynamic route ' + route.route.method + ' ' + route.route.path + '?')) return;
  await fetchJSON('/__specter/routes/' + encodeURIComponent(route.id), {method:'DELETE'});
  if (ui.editingRouteID === route.id) resetRouteEditor();
  showFlash('Dynamic route removed.', 'success');
  await loadRoutes();
}

function clearVarEditor(){
  ui.selectedVar = '';
  document.getElementById('var-key').value = '';
  document.getElementById('var-value').value = '';
  document.getElementById('var-editor-hint').textContent = 'Create or update a var.';
}

function editVar(key, value){
  ui.selectedVar = key;
  document.getElementById('var-key').value = key;
  document.getElementById('var-value').value = value;
  document.getElementById('var-editor-hint').textContent = 'Editing ' + key + '.';
}

async function deleteVar(key){
  if (!confirm('Delete var ' + key + '?')) return;
  await fetchJSON('/__specter/vars/' + encodeURIComponent(key), {method:'DELETE'});
  if (ui.selectedVar === key) clearVarEditor();
  showFlash('Var deleted.', 'success');
  await loadState();
}

function renderVars(vars){
  const entries = Object.entries(vars).sort(function(a,b){ return a[0].localeCompare(b[0]); });
  document.getElementById('var-count').textContent = entries.length + ' var' + (entries.length === 1 ? '' : 's');
  const body = document.getElementById('vars-body');
  if (!entries.length) {
    body.innerHTML = '<div class="list-item"><p class="hint">No vars set.</p></div>';
    return;
  }
  body.innerHTML = entries.map(function(entry){
    const key = entry[0];
    const value = entry[1];
    const active = ui.selectedVar === key ? ' active' : '';
    return '<div class="list-item' + active + '">' +
      '<div class="list-head">' +
      '<div onclick="editVar(\'' + esc(key) + '\', \'' + esc(value) + '\')">' +
      '<h4 class="mono">' + esc(key) + '</h4>' +
      '<p class="mono">' + esc(value) + '</p>' +
      '</div>' +
      '<button class="small warn" onclick="deleteVar(\'' + esc(key) + '\')">Delete</button>' +
      '</div>' +
      '</div>';
  }).join('');
}

function renderStateSummary(stateValue, vars){
  const el = document.getElementById('state-summary');
  const varEntries = Object.entries(vars || {});
  el.innerHTML =
    '<div class="detail-card"><h3>State</h3><pre>' + esc(stateValue || '(none)') + '</pre></div>' +
    '<div class="detail-card"><h3>Vars JSON</h3><pre>' + esc(pretty(vars || {})) + '</pre></div>' +
    '<div class="detail-card"><h3>Counts</h3><div class="kv"><span class="key">Vars</span><span>' + varEntries.length + '</span><span class="key">Requests</span><span>' + ui.requests.length + '</span><span class="key">Stores</span><span>' + ui.stores.length + '</span></div></div>';
}

function renderTimelines(){
  const body = document.getElementById('timelines-body');
  document.getElementById('timeline-count').textContent = ui.timelines.length + ' timeline' + (ui.timelines.length === 1 ? '' : 's');
  if (!ui.timelines.length) {
    body.innerHTML = '<div class="list-item"><p class="hint">No timelines configured.</p></div>';
    updateHeaderMetrics();
    return;
  }
  body.innerHTML = ui.timelines.map(function(item){
    const step = item.step || 0;
    const total = item.total || 0;
    const label = step ? step + ' / ' + total : 'not started';
    const complete = item.complete ? '<span class="tag">complete</span>' : '';
    return '<div class="list-item">' +
      '<div class="list-head">' +
      '<div>' +
      '<h4>' + methodBadge(item.method) + ' <span class="mono">' + esc(item.path) + '</span></h4>' +
      '<p>Step ' + esc(label) + ' · ' + esc(item.calls || 0) + ' request' + (item.calls === 1 ? '' : 's') + ' · ' + esc(item.source || 'config') + '</p>' +
      '</div>' +
      '<div class="table-actions">' + complete + '<button class="small warn" onclick="resetTimeline(\'' + esc(item.key) + '\')">Reset</button></div>' +
      '</div>' +
      '</div>';
  }).join('');
  updateHeaderMetrics();
}

async function loadTimelines(){
  ui.timelines = await fetchJSON('/__specter/timelines') || [];
  renderTimelines();
}

async function resetTimeline(key){
  await sendJSON('/__specter/timelines/' + encodeURIComponent(key) + '/reset', 'POST', {});
  showFlash('Timeline reset.', 'success');
  await loadTimelines();
}

function renderConfigValidation(result){
  const el = document.getElementById('config-result');
  const tag = document.getElementById('config-validity');
  const errors = result && result.errors ? result.errors : [];
  const routes = result && result.routes ? result.routes : [];
  const scenarios = result && result.scenarios ? result.scenarios : [];
  const stores = result && result.stores ? result.stores : [];
  tag.textContent = result && result.valid ? 'valid' : 'invalid';
  tag.className = 'tag';
  const errorHTML = errors.length
    ? '<ul>' + errors.map(function(err){ return '<li class="mono">' + esc(err) + '</li>'; }).join('') + '</ul>'
    : '<p class="hint">No validation errors.</p>';
  const routeHTML = routes.length
    ? '<div class="kv">' + routes.map(function(route){
        return '<span class="key">' + esc(route.method || '?') + '</span><span>' + esc(route.path || '(missing)') + (route.status ? ' · ' + esc(route.status) : '') + (route.state ? ' · state=' + esc(route.state) : '') + '</span>';
      }).join('') + '</div>'
    : '<p class="hint">No routes registered.</p>';
  el.innerHTML =
    '<div class="detail-card"><h3>Errors</h3>' + errorHTML + '</div>' +
    '<div class="detail-card"><h3>Registered Routes</h3>' + routeHTML + '</div>' +
    '<div class="detail-card"><h3>Scenarios</h3><pre>' + esc(pretty(scenarios)) + '</pre></div>' +
    '<div class="detail-card"><h3>Seeded Stores</h3><pre>' + esc(pretty(stores)) + '</pre></div>';
}

async function validateConfigPlayground(){
  const yaml = document.getElementById('config-yaml').value;
  try {
    const result = await sendJSON('/__specter/config/validate', 'POST', {yaml: yaml});
    renderConfigValidation(result || {});
    document.getElementById('config-validator-hint').textContent = result && result.valid ? 'Config is valid.' : 'Config has validation errors.';
  } catch (e) {
    document.getElementById('config-validator-hint').textContent = e.message || 'Validation failed.';
    showFlash(e.message || 'Validation failed.', 'error');
  }
}

function loadCurrentConfigSample(){
  document.getElementById('config-yaml').value = 'scenarios:\n  logged-in:\n    state: logged_in\n    vars:\n      role: admin\n    stores:\n      users:\n        - id: "1"\n          name: Alice\n\nroutes:\n  - path: /profile\n    method: GET\n    state: logged_in\n    response:\n      name: Alice\n';
  document.getElementById('config-validator-hint').textContent = 'Sample loaded.';
}

async function loadState(){
  const pair = await Promise.all([
    fetchJSON('/__specter/state'),
    fetchJSON('/__specter/vars')
  ]);
  const stateValue = pair[0] && pair[0].state ? pair[0].state : '';
  const vars = pair[1] || {};
  document.getElementById('state-val').textContent = stateValue || '(none)';
  document.getElementById('state-dot').className = 'status-dot' + (stateValue ? '' : ' off');
  if (document.activeElement !== document.getElementById('state-input')) {
    document.getElementById('state-input').value = stateValue;
  }
  renderVars(vars);
  renderStateSummary(stateValue, vars);
  updateHeaderMetrics();
}

async function saveState(){
  const value = document.getElementById('state-input').value;
  await sendJSON('/__specter/state', 'PUT', {state:value});
  showFlash('State updated.', 'success');
  await loadState();
}

async function clearState(){
  document.getElementById('state-input').value = '';
  await saveState();
}

async function saveVar(){
  const key = document.getElementById('var-key').value.trim();
  const value = document.getElementById('var-value').value;
  if (!key) {
    showFlash('Var key is required.', 'error');
    return;
  }
  await sendJSON('/__specter/vars/' + encodeURIComponent(key), 'PUT', {value:value});
  ui.selectedVar = key;
  showFlash('Var saved.', 'success');
  await loadState();
}

async function clearVars(){
  if (!confirm('Clear all vars?')) return;
  await fetchJSON('/__specter/vars', {method:'DELETE'});
  clearVarEditor();
  showFlash('Vars cleared.', 'success');
  await loadState();
}

async function resetTargets(targets){
  await sendJSON('/__specter/reset', 'POST', {targets:targets});
  if (targets.indexOf('vars') >= 0) clearVarEditor();
  if (targets.indexOf('stores') >= 0) newStore();
  showFlash('Reset ' + targets.join(', ') + '.', 'success');
  await refreshAll();
}

async function resetAll(){
  if (!confirm('Reset state, vars, history, stores, and timelines?')) return;
  await sendJSON('/__specter/reset', 'POST', {});
  clearVarEditor();
  newStore();
  showFlash('Everything reset.', 'success');
  await refreshAll();
}

function selectStore(name){
  ui.selectedStore = name;
  const store = ui.stores.find(function(item){ return item.name === name; });
  if (store) {
    if (!document.getElementById('store-name').matches(':focus')) document.getElementById('store-name').value = store.name;
    if (!document.getElementById('store-json').matches(':focus')) document.getElementById('store-json').value = JSON.stringify(store.items || [], null, 2);
  }
  renderStores();
}

function newStore(){
  ui.selectedStore = '';
  document.getElementById('store-name').value = '';
  document.getElementById('store-json').value = '[]';
  renderStores();
}

function renderStores(){
  document.getElementById('store-count').textContent = ui.stores.length + ' store' + (ui.stores.length === 1 ? '' : 's');
  const list = document.getElementById('stores-list');
  const clearBtn = document.getElementById('clear-store-btn');
  if (!ui.stores.length) {
    list.innerHTML = '<div class="list-item"><p class="hint">No store collections yet.</p></div>';
    clearBtn.style.display = 'none';
    updateHeaderMetrics();
    return;
  }
  list.innerHTML = ui.stores.map(function(store){
    const active = ui.selectedStore === store.name ? ' active' : '';
    return '<div class="list-item' + active + '">' +
      '<div class="list-head">' +
      '<div onclick="selectStore(\'' + esc(store.name) + '\')">' +
      '<h4>' + esc(store.name) + '</h4>' +
      '<p>' + store.count + ' item' + (store.count === 1 ? '' : 's') + '</p>' +
      '</div>' +
      '<button class="small warn" onclick="clearStore(\'' + esc(store.name) + '\')">Clear</button>' +
      '</div>' +
      '</div>';
  }).join('');
  clearBtn.style.display = ui.selectedStore ? 'inline-flex' : 'none';
  updateHeaderMetrics();
}

async function loadStores(){
  const infos = await fetchJSON('/__specter/stores') || [];
  const stores = await Promise.all(infos.map(async function(info){
    const items = await fetchJSON('/__specter/stores/' + encodeURIComponent(info.name)) || [];
    return {name:info.name, count:info.count, items:items};
  }));
  ui.stores = stores.sort(function(a,b){ return a.name.localeCompare(b.name); });
  if (ui.selectedStore && !ui.stores.find(function(item){ return item.name === ui.selectedStore; })) {
    ui.selectedStore = '';
  }
  if (!ui.selectedStore && ui.stores.length && document.getElementById('store-name').value === '') {
    selectStore(ui.stores[0].name);
  }
  renderStores();
}

async function saveStore(){
  const name = document.getElementById('store-name').value.trim();
  if (!name) {
    showFlash('Collection name is required.', 'error');
    return;
  }
  let payload;
  try {
    payload = JSON.parse(document.getElementById('store-json').value || '[]');
  } catch (e) {
    showFlash('Store JSON must be valid JSON.', 'error');
    return;
  }
  if (!Array.isArray(payload)) {
    showFlash('Store JSON must be an array.', 'error');
    return;
  }
  await sendJSON('/__specter/stores/' + encodeURIComponent(name), 'PUT', payload);
  ui.selectedStore = name;
  showFlash('Store collection saved.', 'success');
  await loadStores();
  selectStore(name);
}

async function clearStore(name){
  if (!confirm('Clear store collection ' + name + '?')) return;
  await fetchJSON('/__specter/stores/' + encodeURIComponent(name), {method:'DELETE'});
  if (ui.selectedStore === name) newStore();
  showFlash('Store collection cleared.', 'success');
  await loadStores();
}

async function clearSelectedStore(){
  if (!ui.selectedStore) return;
  await clearStore(ui.selectedStore);
}

function toggleAutoRefresh(force){
  if (typeof force === 'boolean') ui.autoRefresh = force;
  else ui.autoRefresh = !ui.autoRefresh;
  const btn = document.getElementById('autorefresh-btn');
  btn.textContent = ui.autoRefresh ? 'Auto Refresh On' : 'Auto Refresh Off';
  btn.classList.toggle('active-toggle', ui.autoRefresh);
}

async function refreshAll(){
  try {
    await Promise.all([loadRequests(), loadRoutes(), loadState(), loadStores(), loadTimelines()]);
    document.getElementById('last-update').textContent = new Date().toLocaleTimeString([], {hour12:false});
  } catch (e) {
    showFlash(e.message || 'Refresh failed.', 'error');
  }
}

document.addEventListener('focusin', function(e){
  if (e.target.matches('input, textarea') && ui.autoRefresh) {
    toggleAutoRefresh(false);
    showFlash('Auto refresh paused while editing.', 'info');
  }
});

resetRouteEditor();
refreshAll();
setInterval(function(){
  if (ui.autoRefresh) refreshAll();
}, 2500);
</script>
</body>
</html>`

func renderUI(apiAddr string) string {
	return strings.ReplaceAll(uiHTML, "{{API}}", apiAddr)
}

func uiHandler(apiAddr string) http.Handler {
	page := renderUI(apiAddr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page)
	})
	return mux
}

// StartUI serves the web UI on uiAddr, pointing to the mock API server at apiAddr.
// It is non-blocking; call in a goroutine. A listen error (other than ErrServerClosed)
// is logged but does not terminate the process.
func StartUI(uiAddr, apiAddr string) {
	log.Printf("UI running on http://%s", uiAddr)
	if err := http.ListenAndServe(uiAddr, uiHandler(apiAddr)); err != nil && err != http.ErrServerClosed {
		log.Printf("UI server stopped: %v", err)
	}
}
