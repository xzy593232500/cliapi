package api

import (
	"bytes"
)

const managementQuotaInjection = `<script>
(function(){
  if (window.__CPA_QUOTA_PANEL__) return;
  window.__CPA_QUOTA_PANEL__ = true;

  const style = document.createElement('style');
  style.textContent = ` + "`" + `
    .cpa-quota-launcher {
      position: fixed;
      right: 20px;
      bottom: 20px;
      z-index: 99999;
      border: none;
      border-radius: 999px;
      background: linear-gradient(135deg, #315efb, #6f8cff);
      color: #fff;
      font-weight: 800;
      padding: 12px 18px;
      box-shadow: 0 12px 30px rgba(49,94,251,.28);
      cursor: pointer;
    }
    .cpa-quota-mask {
      position: fixed;
      inset: 0;
      background: rgba(7,10,18,.58);
      z-index: 99998;
      display: none;
    }
    .cpa-quota-panel {
      position: fixed;
      top: 0;
      right: 0;
      width: min(760px, 100vw);
      height: 100vh;
      overflow: auto;
      background: #0f1526;
      color: #e8edf8;
      z-index: 99999;
      transform: translateX(100%);
      transition: transform .25s ease;
      box-shadow: -20px 0 40px rgba(0,0,0,.28);
      font-family: Inter, system-ui, sans-serif;
    }
    .cpa-quota-panel.open { transform: translateX(0); }
    .cpa-quota-mask.open { display: block; }
    .cpa-quota-head { padding: 18px 18px 10px; border-bottom: 1px solid #24314d; }
    .cpa-quota-head-top { display:flex; justify-content:space-between; gap:12px; align-items:flex-start; }
    .cpa-quota-title { font-size: 22px; font-weight: 800; margin:0; }
    .cpa-quota-sub { color:#9cabbe; margin-top:6px; font-size:13px; }
    .cpa-quota-close { border:none; background:#1a243a; color:#fff; width:36px; height:36px; border-radius:10px; cursor:pointer; }
    .cpa-quota-body { padding: 16px 18px 26px; }
    .cpa-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:12px; }
    .cpa-card { background:#141b2d; border:1px solid #24314d; border-radius:14px; padding:14px; margin-bottom:12px; }
    .cpa-stat-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:10px; }
    .cpa-stat { background:#0d1423; border:1px solid #25314a; border-radius:12px; padding:12px; }
    .cpa-stat .label { font-size:12px; color:#9cabbe; margin-bottom:6px; }
    .cpa-stat .value { font-size:24px; font-weight:800; }
    .cpa-row { display:grid; grid-template-columns:repeat(auto-fit,minmax(140px,1fr)); gap:10px; }
    .cpa-quota-panel input, .cpa-quota-panel button {
      width:100%; box-sizing:border-box; border-radius:10px; padding:10px 12px; border:1px solid #33415f;
    }
    .cpa-quota-panel input { background:#0c1322; color:#fff; }
    .cpa-quota-panel button { background:#315efb; color:#fff; border:none; font-weight:700; cursor:pointer; }
    .cpa-muted { color:#9cabbe; font-size:13px; }
    .cpa-feedback { display:none; margin-top:10px; padding:10px 12px; border-radius:10px; font-size:13px; }
    .cpa-feedback.show { display:block; }
    .cpa-feedback.ok { background:rgba(31,185,120,.12); border:1px solid rgba(31,185,120,.26); color:#b2f2d4; }
    .cpa-feedback.err { background:rgba(226,85,99,.12); border:1px solid rgba(226,85,99,.26); color:#ffc4cb; }
    .cpa-table { width:100%; border-collapse:collapse; font-size:13px; }
    .cpa-table th, .cpa-table td { text-align:left; padding:9px 8px; border-bottom:1px solid #22304b; vertical-align:top; }
    .cpa-table th { color:#9cabbe; font-size:11px; text-transform:uppercase; letter-spacing:.04em; }
    .cpa-code { font-family: ui-monospace, SFMono-Regular, monospace; word-break:break-all; }
    .cpa-actions { display:flex; gap:10px; flex-wrap:wrap; }
    .cpa-actions button.secondary { background:#1a243a; }
    @media (max-width: 720px) {
      .cpa-grid, .cpa-stat-grid { grid-template-columns:1fr; }
    }
  ` + "`" + `;
  document.head.appendChild(style);

  const launcher = document.createElement('button');
  launcher.className = 'cpa-quota-launcher';
  launcher.textContent = 'Quota';

  const mask = document.createElement('div');
  mask.className = 'cpa-quota-mask';

  const panel = document.createElement('aside');
  panel.className = 'cpa-quota-panel';
  panel.innerHTML = ` + "`" + `
    <div class="cpa-quota-head">
      <div class="cpa-quota-head-top">
        <div>
          <div class="cpa-quota-title">Quota Management</div>
          <div class="cpa-quota-sub">Integrated token balance, redeem code and transaction management inside management.html</div>
        </div>
        <button class="cpa-quota-close" title="Close">×</button>
      </div>
    </div>
    <div class="cpa-quota-body">
      <div class="cpa-card">
        <div style="font-weight:700;margin-bottom:8px;">Management Key</div>
        <div class="cpa-row">
          <input id="cpaMgmtKey" placeholder="Enter management key" />
          <button id="cpaQuotaRefresh">Refresh</button>
        </div>
        <div class="cpa-muted" style="margin-top:8px;">This uses the same management API credentials as the original control panel.</div>
        <div id="cpaGlobalFeedback" class="cpa-feedback"></div>
      </div>
      <div class="cpa-card">
        <div style="font-weight:700;margin-bottom:10px;">Overview</div>
        <div class="cpa-stat-grid">
          <div class="cpa-stat"><div class="label">Accounts</div><div class="value" id="cpaStatAccounts">0</div></div>
          <div class="cpa-stat"><div class="label">Balance Tokens</div><div class="value" id="cpaStatTokens">0</div></div>
          <div class="cpa-stat"><div class="label">Balance USD</div><div class="value" id="cpaStatUsd">$0</div></div>
          <div class="cpa-stat"><div class="label">Unused Codes</div><div class="value" id="cpaStatCodes">0</div></div>
        </div>
      </div>
      <div class="cpa-grid">
        <div class="cpa-card">
          <div style="font-weight:700;margin-bottom:8px;">Create Redeem Codes</div>
          <div class="cpa-row">
            <input id="cpaUsdAmount" type="number" min="0.01" step="0.01" value="1" placeholder="USD amount" />
            <input id="cpaCodeCount" type="number" min="1" step="1" value="1" placeholder="Count" />
            <input id="cpaExpiresAt" placeholder="Expires at RFC3339 (optional)" />
            <button id="cpaCreateCodes">Create</button>
          </div>
          <div id="cpaCodesFeedback" class="cpa-feedback"></div>
          <pre id="cpaCodesOutput" style="margin-top:10px;background:#0c1322;border:1px solid #22304b;border-radius:10px;padding:10px;white-space:pre-wrap;word-break:break-word;">Generated codes will appear here</pre>
        </div>
        <div class="cpa-card">
          <div style="font-weight:700;margin-bottom:8px;">Balance Operations</div>
          <div class="cpa-row">
            <input id="cpaAdjustApiKey" placeholder="Client API key" />
            <input id="cpaAdjustUsd" type="number" step="0.01" placeholder="USD amount (optional)" />
            <input id="cpaAdjustTokens" type="number" step="1" placeholder="Token amount (optional)" />
            <input id="cpaAdjustNote" placeholder="Note" />
            <button id="cpaAdjustBalance">Adjust</button>
          </div>
          <div id="cpaAdjustFeedback" class="cpa-feedback"></div>
          <div style="height:12px"></div>
          <div class="cpa-row">
            <input id="cpaRedeemApiKey" placeholder="Client API key" />
            <input id="cpaRedeemCode" placeholder="Redeem code" />
            <button id="cpaRedeemBtn">Redeem</button>
          </div>
          <div id="cpaRedeemFeedback" class="cpa-feedback"></div>
        </div>
      </div>
      <div class="cpa-card">
        <div class="cpa-actions" style="justify-content:space-between; align-items:center; margin-bottom:10px;">
          <div style="font-weight:700;">Accounts</div>
          <button class="secondary" id="cpaRefreshAccounts">Refresh Accounts</button>
        </div>
        <div id="cpaAccountsBox" class="cpa-muted">No data yet</div>
      </div>
      <div class="cpa-card">
        <div class="cpa-actions" style="justify-content:space-between; align-items:center; margin-bottom:10px;">
          <div style="font-weight:700;">Redeem Codes</div>
          <button class="secondary" id="cpaRefreshCodes">Refresh Codes</button>
        </div>
        <div id="cpaCodesBox" class="cpa-muted">No data yet</div>
      </div>
      <div class="cpa-card">
        <div class="cpa-actions" style="justify-content:space-between; align-items:center; margin-bottom:10px;">
          <div style="font-weight:700;">Transactions</div>
          <button class="secondary" id="cpaRefreshTx">Refresh Transactions</button>
        </div>
        <div id="cpaTransactionsBox" class="cpa-muted">No data yet</div>
      </div>
    </div>
  ` + "`" + `;

  document.body.appendChild(mask);
  document.body.appendChild(panel);
  document.body.appendChild(launcher);

  const fmtNum = (n) => new Intl.NumberFormat('en-US').format(Number(n || 0));
  const fmtUsd = (n) => '$' + Number(n || 0).toFixed(6).replace(/0+$/,'').replace(/\.$/,'');
  const keyInput = panel.querySelector('#cpaMgmtKey');
  const storageKeys = ['cpa_quota_management_key', 'managementKey', 'management_key', 'apiKey', 'api_key'];
  const saved = storageKeys.map(k => localStorage.getItem(k) || '').find(Boolean) || '';
  if (saved) keyInput.value = saved;

  function storeKey(raw) {
    const value = String(raw || '').trim().replace(/^Bearer\s+/i, '');
    if (!value) return;
    keyInput.value = value;
    localStorage.setItem('cpa_quota_management_key', value);
  }

  function captureHeaders(headersLike) {
    if (!headersLike) return;
    try {
      if (headersLike instanceof Headers) {
        const auth = headersLike.get('Authorization') || headersLike.get('authorization');
        const mgmt = headersLike.get('X-Management-Key') || headersLike.get('x-management-key');
        if (auth) storeKey(auth);
        if (mgmt) storeKey(mgmt);
        return;
      }
      if (Array.isArray(headersLike)) {
        headersLike.forEach(([k, v]) => {
          if (!k) return;
          if (/^authorization$/i.test(k) || /^x-management-key$/i.test(k)) storeKey(v);
        });
        return;
      }
      Object.keys(headersLike || {}).forEach((k) => {
        if (/^authorization$/i.test(k) || /^x-management-key$/i.test(k)) storeKey(headersLike[k]);
      });
    } catch (_) {}
  }

  const originalFetch = window.fetch ? window.fetch.bind(window) : null;
  if (originalFetch) {
    window.fetch = function(input, init) {
      if (init && init.headers) captureHeaders(init.headers);
      if (input && input.headers) captureHeaders(input.headers);
      return originalFetch(input, init);
    };
  }

  const originalOpen = XMLHttpRequest.prototype.open;
  const originalSetHeader = XMLHttpRequest.prototype.setRequestHeader;
  XMLHttpRequest.prototype.open = function() {
    this.__cpaHeaders = {};
    return originalOpen.apply(this, arguments);
  };
  XMLHttpRequest.prototype.setRequestHeader = function(name, value) {
    this.__cpaHeaders = this.__cpaHeaders || {};
    this.__cpaHeaders[name] = value;
    captureHeaders(this.__cpaHeaders);
    return originalSetHeader.apply(this, arguments);
  };

  function openPanel(){ panel.classList.add('open'); mask.classList.add('open'); }
  function closePanel(){ panel.classList.remove('open'); mask.classList.remove('open'); }
  launcher.addEventListener('click', openPanel);
  mask.addEventListener('click', closePanel);
  panel.querySelector('.cpa-quota-close').addEventListener('click', closePanel);

  function getHeaders() {
    const key = keyInput.value.trim();
    if (key) localStorage.setItem('cpa_quota_management_key', key);
    const h = { 'Content-Type': 'application/json' };
    if (key) h.Authorization = 'Bearer ' + key;
    return h;
  }

  async function api(path, options) {
    const res = await fetch(path, { ...(options || {}), headers: { ...getHeaders(), ...(((options || {}).headers) || {}) } });
    const text = await res.text();
    let data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (_) { data = { raw: text }; }
    if (!res.ok) throw new Error(data.error || data.raw || ('HTTP_' + res.status));
    return data;
  }

  function feedback(id, kind, text) {
    const el = panel.querySelector('#' + id);
    if (!el) return;
    if (!text) { el.className = 'cpa-feedback'; el.textContent = ''; return; }
    el.className = 'cpa-feedback show ' + kind;
    el.textContent = text;
  }

  function renderAccounts(items) {
    const box = panel.querySelector('#cpaAccountsBox');
    if (!items.length) { box.textContent = 'No quota accounts yet'; return; }
    box.innerHTML = '<table class="cpa-table"><thead><tr><th>API key</th><th>Balance tokens</th><th>Balance USD</th><th>Used tokens</th><th>Used USD</th><th>Status</th></tr></thead><tbody>' +
      items.map(x => '<tr><td class="cpa-code">' + x.api_key + '</td><td>' + fmtNum(x.token_balance) + '</td><td>' + fmtUsd(x.balance_usd_equivalent) + '</td><td>' + fmtNum(x.token_used_total) + '</td><td>' + fmtUsd(x.usd_used_total) + '</td><td>' + x.status + '</td></tr>').join('') + '</tbody></table>';
  }

  function renderCodes(items) {
    const box = panel.querySelector('#cpaCodesBox');
    if (!items.length) { box.textContent = 'No redeem codes yet'; return; }
    box.innerHTML = '<table class="cpa-table"><thead><tr><th>Code</th><th>USD</th><th>Tokens</th><th>Status</th><th>Used by</th><th>Used at</th></tr></thead><tbody>' +
      items.map(x => '<tr><td class="cpa-code">' + x.code + '</td><td>' + fmtUsd(x.usd_amount) + '</td><td>' + fmtNum(x.token_amount) + '</td><td>' + x.status + '</td><td>' + (x.used_by_api_key || '-') + '</td><td>' + (x.used_at || '-') + '</td></tr>').join('') + '</tbody></table>';
  }

  function renderTransactions(items) {
    const box = panel.querySelector('#cpaTransactionsBox');
    if (!items.length) { box.textContent = 'No transactions yet'; return; }
    box.innerHTML = '<table class="cpa-table"><thead><tr><th>Time</th><th>API key</th><th>Type</th><th>Token delta</th><th>USD delta</th><th>Balance</th><th>Note</th></tr></thead><tbody>' +
      items.map(x => '<tr><td>' + (x.created_at || '-') + '</td><td class="cpa-code">' + x.api_key + '</td><td>' + x.type + '</td><td>' + (x.token_delta > 0 ? '+' : '') + fmtNum(x.token_delta) + '</td><td>' + (x.usd_delta > 0 ? '+' : '') + fmtUsd(x.usd_delta) + '</td><td>' + fmtNum(x.balance_after_tokens) + '</td><td>' + (x.note || '-') + '</td></tr>').join('') + '</tbody></table>';
  }

  function renderOverview(accounts, codes) {
    panel.querySelector('#cpaStatAccounts').textContent = fmtNum(accounts.length);
    panel.querySelector('#cpaStatTokens').textContent = fmtNum(accounts.reduce((a,b)=>a+Number(b.token_balance||0),0));
    panel.querySelector('#cpaStatUsd').textContent = fmtUsd(accounts.reduce((a,b)=>a+Number(b.balance_usd_equivalent||0),0));
    panel.querySelector('#cpaStatCodes').textContent = fmtNum(codes.filter(x => x.status === 'unused').length);
  }

  async function refreshAll() {
    feedback('cpaGlobalFeedback', '', '');
    try {
      const [accounts, codes, tx] = await Promise.all([
        api('/v0/management/quota/accounts').then(r => r.items || []),
        api('/v0/management/quota/redeem-codes').then(r => r.items || []),
        api('/v0/management/quota/transactions').then(r => r.items || []),
      ]);
      renderOverview(accounts, codes);
      renderAccounts(accounts);
      renderCodes(codes);
      renderTransactions(tx);
    } catch (e) {
      feedback('cpaGlobalFeedback', 'err', e.message);
    }
  }

  panel.querySelector('#cpaQuotaRefresh').addEventListener('click', refreshAll);
  panel.querySelector('#cpaRefreshAccounts').addEventListener('click', refreshAll);
  panel.querySelector('#cpaRefreshCodes').addEventListener('click', refreshAll);
  panel.querySelector('#cpaRefreshTx').addEventListener('click', refreshAll);

  panel.querySelector('#cpaCreateCodes').addEventListener('click', async function(){
    feedback('cpaCodesFeedback', '', '');
    try {
      const data = await api('/v0/management/quota/redeem-codes', {
        method: 'POST',
        body: JSON.stringify({
          usd_amount: Number(panel.querySelector('#cpaUsdAmount').value),
          count: Number(panel.querySelector('#cpaCodeCount').value),
          expires_at: panel.querySelector('#cpaExpiresAt').value || ''
        })
      });
      panel.querySelector('#cpaCodesOutput').textContent = (data.codes || []).join('\n');
      feedback('cpaCodesFeedback', 'ok', 'Redeem codes created successfully.');
      refreshAll();
    } catch (e) {
      feedback('cpaCodesFeedback', 'err', e.message);
    }
  });

  panel.querySelector('#cpaAdjustBalance').addEventListener('click', async function(){
    feedback('cpaAdjustFeedback', '', '');
    try {
      const payload = {
        api_key: panel.querySelector('#cpaAdjustApiKey').value.trim(),
        note: panel.querySelector('#cpaAdjustNote').value.trim()
      };
      const usd = panel.querySelector('#cpaAdjustUsd').value.trim();
      const tok = panel.querySelector('#cpaAdjustTokens').value.trim();
      if (usd) payload.usd_amount = Number(usd);
      if (tok) payload.token_amount = Number(tok);
      await api('/v0/management/quota/adjust', { method: 'POST', body: JSON.stringify(payload) });
      feedback('cpaAdjustFeedback', 'ok', 'Balance adjusted successfully.');
      refreshAll();
    } catch (e) {
      feedback('cpaAdjustFeedback', 'err', e.message);
    }
  });

  panel.querySelector('#cpaRedeemBtn').addEventListener('click', async function(){
    feedback('cpaRedeemFeedback', '', '');
    try {
      await api('/v0/management/quota/redeem', {
        method: 'POST',
        body: JSON.stringify({
          api_key: panel.querySelector('#cpaRedeemApiKey').value.trim(),
          code: panel.querySelector('#cpaRedeemCode').value.trim()
        })
      });
      feedback('cpaRedeemFeedback', 'ok', 'Redeem code applied successfully.');
      refreshAll();
    } catch (e) {
      feedback('cpaRedeemFeedback', 'err', e.message);
    }
  });
})();
</script>`

func injectManagementQuotaPanel(content []byte) []byte {
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) == 0 {
		return content
	}
	needle := []byte("</body>")
	idx := bytes.LastIndex(bytes.ToLower(content), needle)
	if idx < 0 {
		return append(content, []byte(managementQuotaInjection)...)
	}
	out := make([]byte, 0, len(content)+len(managementQuotaInjection)+1)
	out = append(out, content[:idx]...)
	out = append(out, []byte(managementQuotaInjection)...)
	out = append(out, content[idx:]...)
	return out
}
