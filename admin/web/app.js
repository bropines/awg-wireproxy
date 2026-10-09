(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const H = { 'X-Requested-With': 'wireproxy-admin', 'Content-Type': 'application/json' };

  async function api(method, url, body) {
    const res = await fetch(url, {
      method, headers: H, credentials: 'same-origin',
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    let data = {};
    try { data = await res.json(); } catch (_) { /* empty body */ }
    if (res.status === 401 && url !== '/api/login') showLogin();
    if (!res.ok) throw Object.assign(new Error(data.error || res.statusText), { status: res.status });
    return data;
  }

  // ---- views ----
  function showLogin() {
    $('app').hidden = true; $('login').hidden = false; $('token').focus();
    stopTimers();
  }
  function showApp() {
    $('login').hidden = true; $('app').hidden = false;
    startTimers();
    refreshStatus();
  }

  $('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    $('login-error').textContent = '';
    try {
      await api('POST', '/api/login', { token: $('token').value });
      $('token').value = '';
      showApp();
    } catch (err) { $('login-error').textContent = err.message; }
  });
  $('logout').addEventListener('click', async () => {
    try { await api('POST', '/api/logout'); } catch (_) { /* ignore */ }
    showLogin();
  });

  document.querySelectorAll('nav [data-tab]').forEach((b) => b.addEventListener('click', () => {
    document.querySelectorAll('nav [data-tab]').forEach((x) => x.classList.toggle('active', x === b));
    ['status', 'config', 'logs'].forEach((t) => { $('tab-' + t).hidden = t !== b.dataset.tab; });
    if (b.dataset.tab === 'config' && !configLoaded) loadConfig();
    if (b.dataset.tab === 'logs') pollLogs();
  }));

  // ---- status ----
  const fmtBytes = (n) => {
    const u = ['B', 'KiB', 'MiB', 'GiB', 'TiB']; let i = 0;
    while (n >= 1024 && i < u.length - 1) { n /= 1024; i++; }
    return (i ? n.toFixed(1) : n) + ' ' + u[i];
  };
  const fmtDur = (s) => {
    const d = Math.floor(s / 86400), h = Math.floor(s % 86400 / 3600), m = Math.floor(s % 3600 / 60);
    return [d && d + 'd', (d || h) && h + 'h', (d || h || m) && m + 'm', s % 60 + 's'].filter(Boolean).join(' ');
  };
  const ago = (t) => {
    if (!t) return 'never';
    const s = Math.max(0, Math.floor(Date.now() / 1000 - t));
    return fmtDur(s) + ' ago';
  };
  function cell(tr, text, cls) {
    const td = document.createElement('td'); td.textContent = text; if (cls) td.className = cls; tr.appendChild(td);
  }

  async function refreshStatus() {
    let s;
    try { s = await api('GET', '/api/status'); } catch (_) { setState(false, 'unreachable'); return null; }
    $('version').textContent = s.version;
    $('uptime').textContent = fmtDur(s.uptime);
    $('cfgpath').textContent = s.config_path;
    setState(s.running, s.running ? 'running' : 'stopped');
    $('start-error').hidden = s.running || !s.start_error;
    $('start-error-text').textContent = s.start_error || '';
    $('restore').hidden = !s.backup;

    const tb = $('peers'); tb.textContent = '';
    if (!s.peers.length) {
      const tr = document.createElement('tr'); cell(tr, s.running ? 'No peers' : 'Tunnel is not running', 'muted'); tr.firstChild.colSpan = 5; tb.appendChild(tr);
    }
    for (const p of s.peers) {
      const tr = document.createElement('tr');
      cell(tr, p.public_key, 'key'); cell(tr, p.endpoint || '–');
      cell(tr, ago(p.last_handshake)); cell(tr, fmtBytes(p.rx_bytes)); cell(tr, fmtBytes(p.tx_bytes));
      tb.appendChild(tr);
    }
    const ul = $('listeners'); ul.textContent = '';
    if (!s.listeners.length) { const li = document.createElement('li'); li.className = 'muted'; li.textContent = 'None configured'; ul.appendChild(li); }
    for (const l of s.listeners) { const li = document.createElement('li'); li.textContent = l; ul.appendChild(li); }
    return s;
  }
  function setState(ok, text) {
    const el = $('state'); el.textContent = text; el.className = 'pill ' + (ok ? 'ok' : 'bad');
  }

  // ---- restart handling ----
  async function waitForRestart() {
    $('overlay').hidden = false; $('overlay-text').textContent = 'Restarting…';
    await new Promise((r) => setTimeout(r, 1200));
    for (let i = 0; i < 60; i++) {
      try {
        const s = await api('GET', '/api/status');
        $('overlay').hidden = true;
        configLoaded = false; if (!$('tab-config').hidden) loadConfig();
        refreshStatus();
        return s;
      } catch (err) { if (err.status === 401) { $('overlay').hidden = true; return; } }
      await new Promise((r) => setTimeout(r, 1000));
    }
    $('overlay-text').textContent = 'The daemon did not come back. Check the logs on the host.';
  }
  $('reload').addEventListener('click', async () => {
    if (!confirm('Restart wireproxy with the current config?')) return;
    try { await api('POST', '/api/reload'); await waitForRestart(); } catch (err) { alert(err.message); }
  });
  $('restore').addEventListener('click', async () => {
    if (!confirm('Replace the config with the previous saved version and restart?')) return;
    try { await api('POST', '/api/config/restore'); await waitForRestart(); } catch (err) { alert(err.message); }
  });

  // ---- config ----
  let configLoaded = false;
  const msg = (text, cls) => { const m = $('config-msg'); m.textContent = text; m.className = 'msg ' + (cls || ''); };

  async function loadConfig() {
    msg('');
    try {
      const c = await api('GET', '/api/config' + ($('reveal').checked ? '?reveal=1' : ''));
      $('editor').value = c.config;
      configLoaded = true;
      if (!c.exists) msg('No config file yet. Write one and press Save.', 'warn');
    } catch (err) { msg(err.message, 'error'); }
  }
  $('reveal').addEventListener('change', () => {
    if ($('editor').value && !confirm('Reloading the editor discards unsaved changes. Continue?')) { $('reveal').checked = !$('reveal').checked; return; }
    loadConfig();
  });

  async function submit(kind) {
    const body = { config: $('editor').value, apply: kind === 'apply' };
    msg(kind === 'validate' ? 'Validating…' : 'Saving…');
    try {
      if (kind === 'validate') {
        const r = await api('POST', '/api/validate', body);
        r.ok ? msg('Config is valid.', 'ok') : msg(r.error, 'error');
        return;
      }
      await api('PUT', '/api/config', body);
      if (kind === 'apply') { msg('Saved. Applying…', 'ok'); await waitForRestart(); msg('Applied.', 'ok'); }
      else msg('Saved. Press “Save & apply” or restart to use it.', 'ok');
    } catch (err) { msg(err.message, 'error'); }
  }
  $('validate').addEventListener('click', () => submit('validate'));
  $('save').addEventListener('click', () => submit('save'));
  $('apply').addEventListener('click', () => submit('apply'));

  function insertAtCursor(text) {
    const ta = $('editor'); const s = ta.selectionStart, e = ta.selectionEnd;
    ta.setRangeText(text, s, e, 'end'); ta.focus();
  }
  $('gen-key').addEventListener('click', () => {
    const b = crypto.getRandomValues(new Uint8Array(32));
    insertAtCursor('HeaderProtectionKey = ' + btoa(String.fromCharCode(...b)) + '\n');
    msg('Inserted a new key. It must match the server and S1–S4 must all be ≥ 12.', 'warn');
  });
  $('template').addEventListener('click', () => {
    insertAtCursor([
      '# AmneziaWG (remove what you do not need; S*/H* must match the server)',
      'Jc = 5', 'Jmin = 10', 'Jmax = 50',
      'S1 = 20', 'S2 = 30', 'S3 = 40', 'S4 = 50',
      'H1 = 100000-200000', 'H2 = 300000-400000', 'H3 = 500000-600000', 'H4 = 700000-800000',
      '#I1 = <b 0xc7000000010800000000000000000000><r 32><t>',
      '#ContentPaddingAddition = 4-16', '#RandomTrailers = true', '',
    ].join('\n'));
  });
  $('editor').addEventListener('keydown', (e) => {
    if (e.key === 'Tab') { e.preventDefault(); insertAtCursor('    '); }
    if ((e.ctrlKey || e.metaKey) && e.key === 's') { e.preventDefault(); submit('save'); }
  });

  // ---- logs ----
  let logNext = 0, logTimer = null, statusTimer = null;
  async function pollLogs() {
    try {
      const r = await api('GET', '/api/logs?after=' + logNext);
      if (r.next < logNext) { // the daemon restarted and numbering began again
        $('log').textContent = ''; logNext = 0;
        return pollLogs();
      }
      logNext = r.next;
      if (r.lines.length) {
        const el = $('log');
        el.textContent += r.lines.map((l) => l.text).join('\n') + '\n';
        if (el.textContent.length > 400000) el.textContent = el.textContent.slice(-300000);
        if ($('follow').checked) el.scrollTop = el.scrollHeight;
      }
    } catch (_) { /* shown via status */ }
  }
  $('clear-logs').addEventListener('click', () => { $('log').textContent = ''; });

  function startTimers() {
    stopTimers();
    statusTimer = setInterval(() => { if (!document.hidden) refreshStatus(); }, 3000);
    logTimer = setInterval(() => { if (!document.hidden && !$('tab-logs').hidden) pollLogs(); }, 1500);
  }
  function stopTimers() { clearInterval(statusTimer); clearInterval(logTimer); }

  // ---- boot ----
  api('GET', '/api/status').then(showApp).catch(() => showLogin());
})();
