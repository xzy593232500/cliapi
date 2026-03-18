package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const quotaPanelHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CLIProxyAPI Quota Console</title>
  <style>
    :root {
      --bg: #0b1020;
      --panel: #141b2d;
      --panel-2: #0f1526;
      --border: #24314d;
      --text: #e7ecf5;
      --muted: #9cabbe;
      --primary: #315efb;
      --primary-2: #6f8cff;
      --success: #1fb978;
      --danger: #e25563;
      --warn: #ffb648;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
      background: linear-gradient(180deg, #0b1020 0%, #0d1324 100%);
      color: var(--text);
    }
    a { color: #9db7ff; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .wrap { max-width: 1320px; margin: 0 auto; padding: 24px; }
    .hero {
      background: linear-gradient(135deg, rgba(49,94,251,.18), rgba(111,140,255,.08));
      border: 1px solid rgba(111,140,255,.25);
      border-radius: 18px;
      padding: 20px;
      margin-bottom: 18px;
      box-shadow: 0 16px 40px rgba(0,0,0,.18);
    }
    .hero-top {
      display: flex;
      gap: 16px;
      justify-content: space-between;
      align-items: flex-start;
      flex-wrap: wrap;
    }
    .hero h1 { margin: 0 0 8px; font-size: 28px; }
    .muted { color: var(--muted); }
    .hero-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
    }
    .link-btn, button {
      appearance: none;
      border: none;
      border-radius: 12px;
      padding: 10px 14px;
      cursor: pointer;
      font-weight: 700;
      font-size: 14px;
    }
    button.primary, .link-btn.primary {
      background: var(--primary);
      color: white;
    }
    button.ghost, .link-btn.ghost {
      background: #1a243a;
      color: var(--text);
      border: 1px solid #33415f;
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(12, minmax(0, 1fr));
      gap: 16px;
    }
    .card {
      background: var(--panel);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 16px;
      box-shadow: 0 10px 30px rgba(0,0,0,.12);
    }
    .span-12 { grid-column: span 12; }
    .span-8 { grid-column: span 8; }
    .span-6 { grid-column: span 6; }
    .span-4 { grid-column: span 4; }
    .span-3 { grid-column: span 3; }
    @media (max-width: 1100px) {
      .span-8, .span-6, .span-4, .span-3 { grid-column: span 12; }
    }
    h2 { margin: 0 0 12px; font-size: 18px; }
    .stats {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
    }
    @media (max-width: 900px) {
      .stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    }
    @media (max-width: 560px) {
      .stats { grid-template-columns: 1fr; }
    }
    .stat {
      background: var(--panel-2);
      border: 1px solid #2b3958;
      border-radius: 14px;
      padding: 14px;
    }
    .stat .label { color: var(--muted); font-size: 13px; margin-bottom: 8px; }
    .stat .value { font-size: 28px; font-weight: 800; letter-spacing: -.02em; }
    .row {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 10px;
    }
    input, select {
      width: 100%;
      padding: 11px 12px;
      border-radius: 12px;
      border: 1px solid #33415f;
      background: #0d1423;
      color: #fff;
      outline: none;
    }
    input:focus { border-color: var(--primary-2); box-shadow: 0 0 0 3px rgba(49,94,251,.16); }
    .toolbar {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 10px;
    }
    .status {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      color: var(--muted);
    }
    .dot {
      width: 10px;
      height: 10px;
      border-radius: 999px;
      background: var(--warn);
      box-shadow: 0 0 0 4px rgba(255,182,72,.12);
    }
    .dot.ok { background: var(--success); box-shadow: 0 0 0 4px rgba(31,185,120,.12); }
    .dot.err { background: var(--danger); box-shadow: 0 0 0 4px rgba(226,85,99,.12); }
    .feedback {
      border-radius: 12px;
      padding: 12px 14px;
      margin-top: 12px;
      font-size: 14px;
      display: none;
    }
    .feedback.show { display: block; }
    .feedback.ok { background: rgba(31,185,120,.12); border: 1px solid rgba(31,185,120,.28); color: #aef0d0; }
    .feedback.err { background: rgba(226,85,99,.12); border: 1px solid rgba(226,85,99,.28); color: #ffc3ca; }
    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 14px;
    }
    th, td {
      text-align: left;
      padding: 10px 8px;
      border-bottom: 1px solid #22304b;
      vertical-align: top;
    }
    th { color: var(--muted); font-weight: 700; font-size: 12px; text-transform: uppercase; letter-spacing: .04em; }
    code, pre {
      background: #0d1423;
      border: 1px solid #22304b;
      border-radius: 12px;
      padding: 10px 12px;
      display: block;
      white-space: pre-wrap;
      word-break: break-word;
    }
    .section-sub { margin-top: -4px; margin-bottom: 14px; color: var(--muted); font-size: 14px; }
    .split {
      display: grid;
      grid-template-columns: repeat(2, minmax(0,1fr));
      gap: 12px;
    }
    @media (max-width: 900px) {
      .split { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
<div class="wrap">
  <div class="hero">
    <div class="hero-top">
      <div>
        <h1>Quota Console</h1>
        <div class="muted">A product-style quota center for CLIProxyAPI: token balances, redeem-code top-up, account operations, and audit history.</div>
        <div class="muted" style="margin-top:8px;">Fixed exchange rate: <b>1 USD = 500000 tokens</b></div>
      </div>
      <div class="hero-actions">
        <a class="link-btn ghost" href="/management.html" target="_blank" rel="noreferrer">Open Original Management</a>
        <button class="ghost" onclick="toggleAutoRefresh()" id="autoRefreshBtn">Auto Refresh: Off</button>
        <button class="primary" onclick="refreshAll()">Refresh Now</button>
      </div>
    </div>
  </div>

  <div class="grid">
    <div class="card span-12">
      <div class="toolbar">
        <div>
          <h2>Management Authentication</h2>
          <div class="section-sub">Use the same management key you already use for the original management API.</div>
        </div>
        <div class="status"><span id="statusDot" class="dot"></span><span id="statusText">Waiting for first refresh</span></div>
      </div>
      <div class="row">
        <input id="mgmtKey" placeholder="Management key / bearer token" />
      </div>
      <div id="globalFeedback" class="feedback"></div>
    </div>

    <div class="card span-12">
      <h2>Overview</h2>
      <div class="section-sub">Live summary across quota accounts, redeem codes, and transaction history.</div>
      <div class="stats">
        <div class="stat"><div class="label">Accounts</div><div class="value" id="statAccounts">0</div></div>
        <div class="stat"><div class="label">Total Balance Tokens</div><div class="value" id="statTokens">0</div></div>
        <div class="stat"><div class="label">Total Balance USD</div><div class="value" id="statUsd">$0</div></div>
        <div class="stat"><div class="label">Unused Redeem Codes</div><div class="value" id="statUnusedCodes">0</div></div>
      </div>
    </div>

    <div class="card span-6">
      <h2>Create Redeem Codes</h2>
      <div class="section-sub">Generate productized top-up codes priced in USD and converted to token balances.</div>
      <div class="row">
        <input id="usdAmount" type="number" min="0.01" step="0.01" value="1" placeholder="USD amount" />
        <input id="codeCount" type="number" min="1" step="1" value="1" placeholder="Count" />
        <input id="expiresAt" placeholder="Expires at RFC3339, optional" />
        <button class="primary" onclick="createCodes()">Create Codes</button>
      </div>
      <div id="createCodesFeedback" class="feedback"></div>
      <pre id="codesOut">Generated codes will appear here</pre>
    </div>

    <div class="card span-6">
      <h2>Balance Operations</h2>
      <div class="section-sub">Adjust client balance or redeem a code for a specific API key.</div>
      <div class="split">
        <div>
          <div class="muted" style="margin-bottom:8px;">Manual balance adjustment</div>
          <div class="row">
            <input id="adjustApiKey" placeholder="Client API key" />
            <input id="adjustUsd" type="number" step="0.01" placeholder="USD amount (optional)" />
            <input id="adjustTokens" type="number" step="1" placeholder="Token amount (optional)" />
            <input id="adjustNote" placeholder="Note" />
            <button class="primary" onclick="adjustBalance()">Apply Adjustment</button>
          </div>
          <div id="adjustFeedback" class="feedback"></div>
        </div>
        <div>
          <div class="muted" style="margin-bottom:8px;">Redeem code to target API key</div>
          <div class="row">
            <input id="redeemApiKey" placeholder="Client API key" />
            <input id="redeemCode" placeholder="Redeem code" />
            <button class="primary" onclick="redeemForApiKey()">Redeem Code</button>
          </div>
          <div id="redeemFeedback" class="feedback"></div>
        </div>
      </div>
    </div>

    <div class="card span-12">
      <div class="toolbar">
        <div>
          <h2>Accounts</h2>
          <div class="section-sub">Per-client quota accounts keyed by authenticated API key.</div>
        </div>
      </div>
      <div id="accountsBox" class="muted">No data yet</div>
    </div>

    <div class="card span-12">
      <div class="toolbar">
        <div>
          <h2>Redeem Codes</h2>
          <div class="section-sub">Track lifecycle of top-up codes and see who used them.</div>
        </div>
      </div>
      <div id="redeemCodesBox" class="muted">No data yet</div>
    </div>

    <div class="card span-12">
      <div class="toolbar">
        <div>
          <h2>Transactions</h2>
          <div class="section-sub">Audit quota consumption, redemptions, and manual adjustments.</div>
        </div>
      </div>
      <div id="transactionsBox" class="muted">No data yet</div>
    </div>
  </div>
</div>
<script>
let autoRefreshTimer = null;
function headers() {
  const key = document.getElementById('mgmtKey').value.trim();
  const h = { 'Content-Type': 'application/json' };
  if (key) h['Authorization'] = 'Bearer ' + key;
  return h;
}
async function api(path, options) {
  const res = await fetch(path, { ...(options || {}), headers: { ...headers(), ...(((options || {}).headers) || {}) } });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || ('HTTP_' + res.status));
  return data;
}
function fmtNum(n){ return new Intl.NumberFormat('en-US').format(Number(n || 0)); }
function fmtUsd(n){ return '$' + Number(n || 0).toFixed(6).replace(/0+$/,'').replace(/\.$/,''); }
function setStatus(kind, text) {
  const dot = document.getElementById('statusDot');
  const label = document.getElementById('statusText');
  dot.className = 'dot' + (kind ? ' ' + kind : '');
  label.textContent = text;
}
function showFeedback(id, kind, text) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = 'feedback show ' + kind;
  el.textContent = text;
}
function clearFeedback(id) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = 'feedback';
  el.textContent = '';
}
function setOverview(accounts, codes) {
  const totalAccounts = accounts.length;
  const totalTokens = accounts.reduce((sum, x) => sum + Number(x.token_balance || 0), 0);
  const totalUsd = accounts.reduce((sum, x) => sum + Number(x.balance_usd_equivalent || 0), 0);
  const unusedCodes = codes.filter(x => x.status === 'unused').length;
  document.getElementById('statAccounts').textContent = fmtNum(totalAccounts);
  document.getElementById('statTokens').textContent = fmtNum(totalTokens);
  document.getElementById('statUsd').textContent = fmtUsd(totalUsd);
  document.getElementById('statUnusedCodes').textContent = fmtNum(unusedCodes);
}
async function refreshAccounts(){
  const data = await api('/v0/management/quota/accounts');
  const items = data.items || [];
  if (!items.length) { accountsBox.textContent = 'No accounts'; return items; }
  accountsBox.innerHTML = '<table><thead><tr><th>API key</th><th>Balance tokens</th><th>Balance USD</th><th>Used tokens</th><th>Used USD</th><th>Status</th><th>Updated</th></tr></thead><tbody>' +
    items.map(x => '<tr><td><code>' + x.api_key + '</code></td><td>' + fmtNum(x.token_balance) + '</td><td>' + fmtUsd(x.balance_usd_equivalent) + '</td><td>' + fmtNum(x.token_used_total) + '</td><td>' + fmtUsd(x.usd_used_total) + '</td><td>' + x.status + '</td><td>' + (x.updated_at || '-') + '</td></tr>').join('') + '</tbody></table>';
  return items;
}
async function refreshRedeemCodes(){
  const data = await api('/v0/management/quota/redeem-codes');
  const items = data.items || [];
  if (!items.length) { redeemCodesBox.textContent = 'No redeem codes'; return items; }
  redeemCodesBox.innerHTML = '<table><thead><tr><th>Code</th><th>USD</th><th>Tokens</th><th>Status</th><th>Used by</th><th>Used at</th><th>Expires at</th></tr></thead><tbody>' +
    items.map(x => '<tr><td><code>' + x.code + '</code></td><td>' + fmtUsd(x.usd_amount) + '</td><td>' + fmtNum(x.token_amount) + '</td><td>' + x.status + '</td><td>' + (x.used_by_api_key || '-') + '</td><td>' + (x.used_at || '-') + '</td><td>' + (x.expires_at || '-') + '</td></tr>').join('') + '</tbody></table>';
  return items;
}
async function refreshTransactions(){
  const data = await api('/v0/management/quota/transactions');
  const items = data.items || [];
  if (!items.length) { transactionsBox.textContent = 'No transactions'; return items; }
  transactionsBox.innerHTML = '<table><thead><tr><th>Time</th><th>API key</th><th>Type</th><th>Token delta</th><th>USD delta</th><th>Balance after</th><th>Model</th><th>Code</th><th>Note</th></tr></thead><tbody>' +
    items.map(x => '<tr><td>' + (x.created_at || '-') + '</td><td><code>' + x.api_key + '</code></td><td>' + x.type + '</td><td>' + (x.token_delta > 0 ? '+' : '') + fmtNum(x.token_delta) + '</td><td>' + (x.usd_delta > 0 ? '+' : '') + fmtUsd(x.usd_delta) + '</td><td>' + fmtNum(x.balance_after_tokens) + '</td><td>' + (x.model || '-') + '</td><td>' + (x.redemption_code || '-') + '</td><td>' + (x.note || '-') + '</td></tr>').join('') + '</tbody></table>';
  return items;
}
async function createCodes(){
  clearFeedback('createCodesFeedback');
  try {
    const data = await api('/v0/management/quota/redeem-codes', { method: 'POST', body: JSON.stringify({ usd_amount: Number(usdAmount.value), count: Number(codeCount.value), expires_at: expiresAt.value || '' }) });
    codesOut.textContent = data.codes.join('\n');
    showFeedback('createCodesFeedback', 'ok', 'Redeem codes created successfully.');
    await refreshAll();
  } catch (e) {
    showFeedback('createCodesFeedback', 'err', e.message);
  }
}
async function adjustBalance(){
  clearFeedback('adjustFeedback');
  try {
    const payload = { api_key: adjustApiKey.value.trim(), note: adjustNote.value.trim() };
    if (adjustUsd.value.trim()) payload.usd_amount = Number(adjustUsd.value);
    if (adjustTokens.value.trim()) payload.token_amount = Number(adjustTokens.value);
    await api('/v0/management/quota/adjust', { method: 'POST', body: JSON.stringify(payload) });
    showFeedback('adjustFeedback', 'ok', 'Balance adjusted successfully.');
    await refreshAll();
  } catch (e) {
    showFeedback('adjustFeedback', 'err', e.message);
  }
}
async function redeemForApiKey(){
  clearFeedback('redeemFeedback');
  try {
    await api('/v0/management/quota/redeem', { method: 'POST', body: JSON.stringify({ api_key: redeemApiKey.value.trim(), code: redeemCode.value.trim() }) });
    showFeedback('redeemFeedback', 'ok', 'Redeem code applied successfully.');
    await refreshAll();
  } catch (e) {
    showFeedback('redeemFeedback', 'err', e.message);
  }
}
async function refreshAll(){
  clearFeedback('globalFeedback');
  try {
    const [accounts, codes] = await Promise.all([refreshAccounts(), refreshRedeemCodes(), refreshTransactions()]);
    setOverview(accounts || [], codes || []);
    setStatus('ok', 'Connected and refreshed just now');
  } catch (e) {
    setStatus('err', 'Refresh failed');
    showFeedback('globalFeedback', 'err', e.message);
  }
}
function toggleAutoRefresh() {
  const btn = document.getElementById('autoRefreshBtn');
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer);
    autoRefreshTimer = null;
    btn.textContent = 'Auto Refresh: Off';
    return;
  }
  autoRefreshTimer = setInterval(refreshAll, 15000);
  btn.textContent = 'Auto Refresh: On';
}
</script>
</body>
</html>`

func (s *Server) serveQuotaControlPanel(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(quotaPanelHTML))
}
