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
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0f1117;color:#e2e8f0;font-family:'SF Mono',Consolas,monospace;font-size:13px}
header{background:#1a1d2e;padding:12px 20px;display:flex;align-items:center;gap:16px;border-bottom:1px solid #2d3748}
header h1{font-size:16px;color:#a78bfa;letter-spacing:.05em}
.badge{font-size:11px;color:#64748b}
nav{display:flex;gap:2px;padding:8px 20px;background:#1a1d2e;border-bottom:1px solid #2d3748}
nav button{background:none;border:none;color:#94a3b8;cursor:pointer;padding:6px 14px;border-radius:4px;font-size:12px;font-family:inherit}
nav button.active{background:#2d3748;color:#e2e8f0}
nav button:hover:not(.active){background:#232638}
.tab{display:none;padding:16px 20px}
.tab.active{display:block}
table{width:100%;border-collapse:collapse;font-size:12px}
th{text-align:left;padding:6px 10px;color:#64748b;border-bottom:1px solid #2d3748;font-weight:500}
td{padding:6px 10px;border-bottom:1px solid #1e2333;vertical-align:top;max-width:400px;word-break:break-all}
tr:hover td{background:#1e2333}
.method{display:inline-block;padding:1px 6px;border-radius:3px;font-weight:bold;font-size:11px}
.GET{color:#4ade80}.POST{color:#60a5fa}.PUT{color:#f59e0b}.PATCH{color:#a78bfa}.DELETE{color:#f87171}
.tag{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;background:#1e2333;color:#94a3b8}
.kv{display:grid;grid-template-columns:auto 1fr;gap:4px 12px}
.kv .k{color:#64748b}
.empty{color:#4a5568;padding:20px 10px}
pre{white-space:pre-wrap;word-break:break-all;color:#94a3b8;max-height:80px;overflow-y:auto}
.toolbar{display:flex;gap:8px;margin-bottom:12px;align-items:center}
.toolbar button{background:#2d3748;border:none;color:#94a3b8;cursor:pointer;padding:5px 10px;border-radius:4px;font-family:inherit;font-size:12px}
.toolbar button:hover{background:#374151;color:#e2e8f0}
.state-box{display:inline-flex;align-items:center;gap:8px;background:#1e2333;padding:8px 14px;border-radius:6px;margin-bottom:16px}
.dot{width:8px;height:8px;border-radius:50%;background:#4ade80}
.dot.off{background:#374151}
h3{color:#94a3b8;font-size:12px;margin-bottom:8px;margin-top:16px}
h3:first-child{margin-top:0}
.store-name{color:#a78bfa;font-weight:bold;margin-bottom:4px}
.store-section{margin-bottom:20px}
</style>
</head>
<body>
<header>
  <h1>👻 specter</h1>
  <span class="badge" id="api-addr"></span>
  <span class="badge" id="last-update"></span>
</header>
<nav>
  <button class="active" onclick="showTab('requests',this)">Requests</button>
  <button onclick="showTab('routes',this)">Routes</button>
  <button onclick="showTab('state',this)">State &amp; Vars</button>
  <button onclick="showTab('stores',this)">Stores</button>
</nav>

<div class="tab active" id="tab-requests">
  <div class="toolbar">
    <button onclick="clearHistory()">Clear History</button>
    <span class="badge" id="req-count"></span>
  </div>
  <table>
    <thead><tr><th>#</th><th>Time</th><th>Method</th><th>Path</th><th>Query</th><th>Body</th></tr></thead>
    <tbody id="req-body"></tbody>
  </table>
</div>

<div class="tab" id="tab-routes">
  <table>
    <thead><tr><th>Method</th><th>Path</th><th>Source</th><th>State</th><th>Notes</th></tr></thead>
    <tbody id="routes-body"></tbody>
  </table>
</div>

<div class="tab" id="tab-state">
  <h3>State</h3>
  <div class="state-box"><span class="dot off" id="state-dot"></span><span id="state-val">—</span></div>
  <h3>Vars</h3>
  <div id="vars-body" class="kv"></div>
</div>

<div class="tab" id="tab-stores">
  <div id="stores-body"></div>
</div>

<script>
const API="{{API}}";
document.getElementById('api-addr').textContent=API;

function showTab(name,btn){
  document.querySelectorAll('.tab').forEach(t=>t.classList.remove('active'));
  document.querySelectorAll('nav button').forEach(b=>b.classList.remove('active'));
  document.getElementById('tab-'+name).classList.add('active');
  btn.classList.add('active');
}
function esc(s){return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')}
function mb(m){return '<span class="method '+esc(m)+'">'+esc(m)+'</span>'}
function fmtTime(t){const d=new Date(t);return d.toTimeString().slice(0,8)+'.'+String(d.getMilliseconds()).padStart(3,'0')}

async function loadRequests(){
  try{
    const data=await fetch(API+'/__specter/requests').then(r=>r.json());
    document.getElementById('req-count').textContent=data.length+' request'+(data.length!==1?'s':'');
    const tbody=document.getElementById('req-body');
    if(!data.length){tbody.innerHTML='<tr><td colspan="6" class="empty">no requests recorded</td></tr>';return}
    tbody.innerHTML=data.slice().reverse().map((r,i)=>{
      const q=r.query?Object.entries(r.query).map(([k,v])=>esc(k)+'='+esc(v)).join('&'):'';
      return '<tr><td>'+(data.length-i)+'</td><td>'+fmtTime(r.time)+'</td><td>'+mb(r.method)+'</td><td>'+esc(r.path)+'</td><td>'+q+'</td><td><pre>'+esc(r.body||'')+'</pre></td></tr>';
    }).join('');
  }catch(e){}
}

async function clearHistory(){
  await fetch(API+'/__specter/requests',{method:'DELETE'});
  loadRequests();
}

async function loadRoutes(){
  try{
    const data=await fetch(API+'/__specter/routes').then(r=>r.json());
    const tbody=document.getElementById('routes-body');
    if(!data.length){tbody.innerHTML='<tr><td colspan="5" class="empty">no routes configured</td></tr>';return}
    tbody.innerHTML=data.map(r=>{
      const rt=r.route||r;
      const notes=[];
      if(rt.delay)notes.push('delay:'+rt.delay+'ms');
      if(rt.rate_limit)notes.push('rate:'+rt.rate_limit);
      if(rt.error_rate)notes.push('err:'+(rt.error_rate*100).toFixed(0)+'%');
      if(rt.proxy)notes.push('proxy');
      if(rt.store_push||rt.store_list||rt.store_get||rt.store_put||rt.store_patch||rt.store_delete||rt.store_clear)notes.push('store');
      return '<tr><td>'+mb(rt.method)+'</td><td>'+esc(rt.path)+'</td><td><span class="tag">'+esc(r.source||'config')+'</span></td><td>'+esc(rt.state||'')+'</td><td>'+notes.map(n=>'<span class="tag">'+esc(n)+'</span>').join(' ')+'</td></tr>';
    }).join('');
  }catch(e){}
}

async function loadState(){
  try{
    const[sr,vr]=await Promise.all([
      fetch(API+'/__specter/state').then(r=>r.json()),
      fetch(API+'/__specter/vars').then(r=>r.json()),
    ]);
    const sv=sr.state||'';
    document.getElementById('state-val').textContent=sv||'(none)';
    document.getElementById('state-dot').className='dot'+(sv?'':' off');
    const entries=Object.entries(vr);
    const el=document.getElementById('vars-body');
    el.innerHTML=entries.length?entries.map(([k,v])=>'<span class="k">'+esc(k)+'</span><span>'+esc(v)+'</span>').join(''):'<span class="empty">no vars set</span>';
  }catch(e){}
}

async function loadStores(){
  try{
    const infos=await fetch(API+'/__specter/stores').then(r=>r.json());
    const el=document.getElementById('stores-body');
    if(!infos.length){el.innerHTML='<div class="empty">no store collections yet</div>';return}
    const sections=await Promise.all(infos.map(async info=>{
      const items=await fetch(API+'/__specter/stores/'+encodeURIComponent(info.name)).then(r=>r.json());
      const cols=items.length?Object.keys(items[0]).map(k=>'<th>'+esc(k)+'</th>').join(''):'<th></th>';
      const rows=items.length?items.map(item=>'<tr>'+Object.entries(item).map(([,v])=>'<td>'+esc(JSON.stringify(v))+'</td>').join('')+'</tr>').join(''):'<tr><td class="empty">empty</td></tr>';
      return '<div class="store-section"><div class="store-name">'+esc(info.name)+' <span class="badge">'+info.count+' item'+(info.count!==1?'s':'')+'</span></div><table><thead><tr>'+cols+'</tr></thead><tbody>'+rows+'</tbody></table></div>';
    }));
    el.innerHTML=sections.join('');
  }catch(e){}
}

function refresh(){
  loadRequests();loadRoutes();loadState();loadStores();
  document.getElementById('last-update').textContent='updated '+new Date().toTimeString().slice(0,8);
}
refresh();
setInterval(refresh,2000);
</script>
</body>
</html>`

// StartUI serves the web UI on uiAddr, pointing to the mock API server at apiAddr.
// It is non-blocking — call in a goroutine. A listen error (other than ErrServerClosed)
// is logged but does not terminate the process.
func StartUI(uiAddr, apiAddr string) {
	page := strings.ReplaceAll(uiHTML, "{{API}}", apiAddr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page)
	})
	log.Printf("UI running on http://%s", uiAddr)
	if err := http.ListenAndServe(uiAddr, mux); err != nil && err != http.ErrServerClosed {
		log.Printf("UI server stopped: %v", err)
	}
}
