package dashboard

import (
	"net/http"
)

// Routes returns the HTTP mux for the dashboard.
func Routes(h *Hub) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.ServeWS)
	mux.HandleFunc("/", serveIndex)
	return mux
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>⚔️ Chaincode Carnival — Live Arena</title>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link href="https://fonts.googleapis.com/css2?family=Share+Tech+Mono&family=Rajdhani:wght@500;700&display=swap" rel="stylesheet" />
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    :root {
      --bg:        #090c13;
      --surface:   #0d1117;
      --border:    #1e2a3a;
      --analyzer:  #00d4ff;
      --redteam:   #ff4757;
      --blueteam:  #2ed573;
      --skeptic:   #a855f7;
      --system:    #ffa502;
      --text:      #c9d1d9;
      --dim:       #556272;
      --glow-r:    0 0 12px #ff475766;
      --glow-b:    0 0 12px #2ed57366;
    }

    body {
      background: var(--bg);
      color: var(--text);
      font-family: 'Share Tech Mono', monospace;
      min-height: 100vh;
      display: grid;
      grid-template-rows: auto 1fr auto;
    }

    /* ── HEADER ── */
    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 16px 28px;
      border-bottom: 1px solid var(--border);
      background: linear-gradient(90deg, #0d1117 0%, #111827 100%);
    }
    header h1 {
      font-family: 'Rajdhani', sans-serif;
      font-size: 1.6rem;
      font-weight: 700;
      letter-spacing: 2px;
      background: linear-gradient(90deg, var(--redteam), var(--analyzer));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }
    #conn-status {
      font-size: .75rem;
      padding: 4px 10px;
      border-radius: 20px;
      border: 1px solid var(--dim);
      color: var(--dim);
      transition: all .3s;
    }
    #conn-status.live { border-color: #2ed573; color: #2ed573; }

    /* ── MAIN GRID ── */
    main {
      display: grid;
      grid-template-columns: 1fr 320px;
      gap: 0;
      overflow: hidden;
    }

    /* ── FEED ── */
    #feed-panel {
      border-right: 1px solid var(--border);
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    #feed-panel h2 {
      font-family: 'Rajdhani', sans-serif;
      padding: 12px 20px;
      border-bottom: 1px solid var(--border);
      font-size: 1rem;
      letter-spacing: 1px;
      color: var(--dim);
    }
    #feed {
      flex: 1;
      overflow-y: auto;
      padding: 16px 20px;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }
    #feed::-webkit-scrollbar { width: 4px; }
    #feed::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }

    .msg {
      display: flex;
      gap: 10px;
      align-items: flex-start;
      animation: slideIn .25s ease;
    }
    @keyframes slideIn {
      from { opacity: 0; transform: translateY(8px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    .avatar {
      width: 36px;
      height: 36px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.1rem;
      flex-shrink: 0;
      border: 1px solid transparent;
    }
    .bubble {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 4px 12px 12px 12px;
      padding: 8px 14px;
      max-width: 90%;
    }
    .bubble .agent-name {
      font-family: 'Rajdhani', sans-serif;
      font-size: .75rem;
      font-weight: 700;
      letter-spacing: 1px;
      margin-bottom: 3px;
    }
    .bubble .text { font-size: .85rem; line-height: 1.5; }
    .bubble .ts   { font-size: .68rem; color: var(--dim); margin-top: 4px; }

    /* Per-agent colours */
    .agent-Analyzer  .agent-name  { color: var(--analyzer); }
    .agent-Analyzer  .avatar      { background: #00d4ff11; border-color: #00d4ff44; }
    .agent-RedTeam   .agent-name  { color: var(--redteam); }
    .agent-RedTeam   .avatar      { background: #ff475711; border-color: #ff475744; }
    .agent-BlueTeam  .agent-name  { color: var(--blueteam); }
    .agent-BlueTeam  .avatar      { background: #2ed57311; border-color: #2ed57344; }
    .agent-Skeptic   .agent-name  { color: var(--skeptic); }
    .agent-Skeptic   .avatar      { background: #a855f711; border-color: #a855f744; }
    .agent-System    .agent-name  { color: var(--system); }
    .agent-System    .avatar      { background: #ffa50211; border-color: #ffa50244; }

    /* Event-type badges */
    .evt-badge {
      display: inline-block;
      font-size: .62rem;
      padding: 1px 6px;
      border-radius: 3px;
      background: var(--border);
      color: var(--dim);
      margin-left: 6px;
      vertical-align: middle;
      font-family: 'Rajdhani', sans-serif;
      letter-spacing: .5px;
    }

    /* ── RIGHT PANEL ── */
    #right-panel {
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    /* ── SCOREBOARD ── */
    #scoreboard {
      padding: 16px;
      border-bottom: 1px solid var(--border);
    }
    #scoreboard h2 {
      font-family: 'Rajdhani', sans-serif;
      font-size: .85rem;
      letter-spacing: 1px;
      color: var(--dim);
      margin-bottom: 12px;
    }
    .score-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 10px;
      gap: 8px;
    }
    .score-label {
      font-family: 'Rajdhani', sans-serif;
      font-size: .9rem;
      font-weight: 700;
      letter-spacing: .5px;
      flex: 1;
    }
    .score-bar-wrap {
      flex: 2;
      background: var(--border);
      border-radius: 4px;
      height: 6px;
      overflow: hidden;
    }
    .score-bar {
      height: 100%;
      border-radius: 4px;
      width: 0%;
      transition: width .6s ease;
    }
    .score-val {
      font-size: .85rem;
      min-width: 36px;
      text-align: right;
    }

    #verdict-box {
      margin-top: 12px;
      padding: 10px 14px;
      border-radius: 8px;
      border: 1px solid var(--border);
      font-family: 'Rajdhani', sans-serif;
      font-size: 1rem;
      font-weight: 700;
      letter-spacing: 1px;
      text-align: center;
      color: var(--dim);
      transition: all .4s;
    }
    #verdict-box.approved { border-color: var(--blueteam); color: var(--blueteam); background: #2ed57311; }
    #verdict-box.rejected { border-color: var(--redteam);  color: var(--redteam);  background: #ff475711; }

    /* ── TIMELINE ── */
    #timeline {
      flex: 1;
      overflow-y: auto;
      padding: 16px;
    }
    #timeline h2 {
      font-family: 'Rajdhani', sans-serif;
      font-size: .85rem;
      letter-spacing: 1px;
      color: var(--dim);
      margin-bottom: 12px;
    }
    .tl-item {
      display: flex;
      gap: 10px;
      align-items: flex-start;
      margin-bottom: 14px;
      animation: slideIn .25s ease;
    }
    .tl-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      margin-top: 4px;
      flex-shrink: 0;
      border: 2px solid;
    }
    .tl-content { font-size: .78rem; line-height: 1.5; }
    .tl-evt     { font-family: 'Rajdhani', sans-serif; font-weight: 700; font-size: .8rem; }
    .tl-ts      { color: var(--dim); font-size: .68rem; }

    /* ── FOOTER ── */
    footer {
      padding: 10px 28px;
      border-top: 1px solid var(--border);
      font-size: .72rem;
      color: var(--dim);
      display: flex;
      justify-content: space-between;
    }
    #event-count { font-family: 'Rajdhani', sans-serif; color: var(--analyzer); }
  </style>
</head>
<body>

<header>
  <h1>⚔️ CHAINCODE CARNIVAL — LIVE ARENA</h1>
  <div id="conn-status">⬤ CONNECTING…</div>
</header>

<main>
  <!-- Live feed (left) -->
  <section id="feed-panel">
    <h2>📡 LIVE AGENT FEED</h2>
    <div id="feed"></div>
  </section>

  <!-- Scoreboard + Timeline (right) -->
  <aside id="right-panel">
    <div id="scoreboard">
      <h2>🏆 SCOREBOARD</h2>

      <div class="score-row">
        <div class="score-label" style="color:var(--redteam)">👺 RED TEAM</div>
        <div class="score-bar-wrap"><div id="bar-red" class="score-bar" style="background:var(--redteam)"></div></div>
        <div id="val-red" class="score-val">0</div>
      </div>

      <div class="score-row">
        <div class="score-label" style="color:var(--blueteam)">🛡️ BLUE TEAM</div>
        <div class="score-bar-wrap"><div id="bar-blue" class="score-bar" style="background:var(--blueteam)"></div></div>
        <div id="val-blue" class="score-val">0</div>
      </div>

      <div class="score-row">
        <div class="score-label" style="color:var(--skeptic)">⚖️ SKEPTIC SCORE</div>
        <div class="score-bar-wrap"><div id="bar-skeptic" class="score-bar" style="background:var(--skeptic)"></div></div>
        <div id="val-skeptic" class="score-val">—</div>
      </div>

      <div id="verdict-box">AWAITING VERDICT</div>
    </div>

    <div id="timeline">
      <h2>📋 EVENT TIMELINE</h2>
    </div>
  </aside>
</main>

<footer>
  <span>Chaincode Carnival · Multi-Agent Adversarial Arena</span>
  <span id="event-count">0 events</span>
</footer>

<script>
const AGENT_ICONS = {
  Analyzer: '🔍', RedTeam: '👺', BlueTeam: '🛡️', Skeptic: '⚖️', System: '🏟️'
};
const AGENT_COLORS = {
  Analyzer: '#00d4ff', RedTeam: '#ff4757', BlueTeam: '#2ed573', Skeptic: '#a855f7', System: '#ffa502'
};
const EVT_LABELS = {
  ARENA_START: 'Match Start', FINDINGS_READY: 'Findings Ready',
  EXPLOIT_LAUNCHED: 'Exploit Launched', EXPLOIT_CONFIRMED: 'Exploit Confirmed',
  PATCH_SUBMITTED: 'Patch Submitted', RETEST_RESULT: 'Retest',
  VERDICT_RENDERED: 'Verdict', BANTER: 'Live', ARENA_OVER: 'Match Over'
};

let eventCount = 0;
let redScore   = 0;
let blueScore  = 0;

const feed     = document.getElementById('feed');
const timeline = document.getElementById('timeline');
const status   = document.getElementById('conn-status');
const evtCount = document.getElementById('event-count');

function formatTime(ts) {
  const d = new Date(ts);
  return d.toLocaleTimeString('en-US', {hour12: false, hour:'2-digit', minute:'2-digit', second:'2-digit'});
}

function addFeedMessage(agent, evtType, message) {
  const icon  = AGENT_ICONS[agent]  || '🤖';
  const color = AGENT_COLORS[agent] || '#888';

  const div = document.createElement('div');
  div.className = 'msg agent-' + agent;
  div.innerHTML = ` + "`" + `
    <div class="avatar">${icon}</div>
    <div class="bubble">
      <div class="agent-name">${agent.toUpperCase()} <span class="evt-badge">${EVT_LABELS[evtType] || evtType}</span></div>
      <div class="text">${message}</div>
      <div class="ts">${formatTime(Date.now())}</div>
    </div>` + "`" + `;
  feed.appendChild(div);
  feed.scrollTop = feed.scrollHeight;
}

function addTimeline(agent, evtType, ts) {
  const color = AGENT_COLORS[agent] || '#888';
  const div = document.createElement('div');
  div.className = 'tl-item';
  div.innerHTML = ` + "`" + `
    <div class="tl-dot" style="border-color:${color};background:${color}33"></div>
    <div class="tl-content">
      <div class="tl-evt" style="color:${color}">${EVT_LABELS[evtType] || evtType}</div>
      <div style="color:#888;font-size:.72rem">by ${agent}</div>
      <div class="tl-ts">${formatTime(ts)}</div>
    </div>` + "`" + `;
  timeline.appendChild(div);
  timeline.scrollTop = timeline.scrollHeight;
}

function setBar(id, valId, val, max) {
  document.getElementById(id).style.width = Math.min(100, (val / max) * 100) + '%';
  document.getElementById(valId).textContent = val;
}

function handleEvent(evt) {
  eventCount++;
  evtCount.textContent = eventCount + ' events';

  addTimeline(evt.agent, evt.type, evt.ts);

  // Extract message from payload
  let msg = null;
  if (evt.type === 'BANTER' && evt.payload && evt.payload.message) {
    msg = evt.payload.message;
  } else if (evt.type === 'ARENA_START') {
    msg = 'Arena initialized. Match starting…';
  } else if (evt.type === 'FINDINGS_READY') {
    const count = evt.payload?.report?.findings?.length || 0;
    msg = count + ' vulnerabilities found in the chaincode. Red Team and Blue Team are reacting simultaneously.';
  } else if (evt.type === 'EXPLOIT_CONFIRMED') {
    redScore++;
    setBar('bar-red', 'val-red', redScore, Math.max(redScore + blueScore, 1));
    msg = 'Exploit confirmed in sandbox — chaincode state corrupted.';
  } else if (evt.type === 'PATCH_SUBMITTED') {
    blueScore++;
    setBar('bar-blue', 'val-blue', blueScore, Math.max(redScore + blueScore, 1));
    msg = 'Patch submitted after ' + (evt.payload?.rounds || '?') + ' round(s). Awaiting retest…';
  } else if (evt.type === 'VERDICT_RENDERED') {
    const score = evt.payload?.score || 0;
    setBar('bar-skeptic', 'val-skeptic', score, 100);
    const box = document.getElementById('verdict-box');
    if (evt.payload?.approved) {
      box.textContent = '✅ APPROVED — ' + score + '/100 — DEPLOYING TO FABRIC';
      box.className = 'approved';
    } else {
      box.textContent = '❌ REJECTED — ' + score + '/100';
      box.className = 'rejected';
    }
    msg = 'Verdict: ' + score + '/100 — ' + (evt.payload?.approved ? 'AUTHORIZED' : 'REJECTED');
  } else if (evt.type === 'ARENA_OVER') {
    msg = '=== ARENA MATCH CONCLUDED ===';
  }

  if (msg) addFeedMessage(evt.agent, evt.type, msg);
}

// WebSocket connection with auto-reconnect
function connect() {
  const ws = new WebSocket('ws://' + location.host + '/ws');

  ws.onopen = () => {
    status.textContent = '⬤ LIVE';
    status.classList.add('live');
  };

  ws.onmessage = (e) => {
    try { handleEvent(JSON.parse(e.data)); } catch {}
  };

  ws.onclose = () => {
    status.textContent = '⬤ DISCONNECTED';
    status.classList.remove('live');
    setTimeout(connect, 2000);
  };
}

connect();
</script>
</body>
</html>`
