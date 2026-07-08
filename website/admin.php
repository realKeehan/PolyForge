<?php http_response_code(200); ?>
<!doctype html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>PolyForge Admin</title>
  <meta name="robots" content="noindex, nofollow" />
  <link rel="icon" href="./favicon.ico" />
  <link rel="stylesheet" href="./styles.css" />
  <style>
    body { padding: 24px; }
    .adm { max-width: 1100px; margin: 0 auto; }
    .adm-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
    .adm-head h1 { margin: 0; font-size: 1.3rem; }
    .adm-head h1 span { color: var(--pf-purple); }
    .adm-card {
      background: var(--surface); border: 1px solid var(--border);
      border-radius: 14px; padding: 20px; margin-bottom: 18px;
    }
    .adm-tabs { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 18px; }
    .adm-tab {
      padding: 8px 16px; border-radius: 999px; border: 1px solid var(--border-2);
      background: var(--surface-2); color: var(--text); cursor: pointer; font-size: .85rem;
    }
    .adm-tab.is-active { border-color: var(--pf-purple); color: var(--pf-purple); font-weight: 600; }
    .adm input[type=text], .adm input[type=password], .adm input[type=file], .adm select, .adm textarea {
      width: 100%; padding: 9px 12px; border-radius: 8px; border: 1px solid var(--border-2);
      background: var(--surface-3); color: var(--text); font-size: .85rem; font-family: inherit;
    }
    .adm textarea { font-family: "JetBrains Mono", monospace; font-size: .75rem; min-height: 260px; }
    .adm label { display: block; font-size: .72rem; letter-spacing: .08em; text-transform: uppercase; color: var(--text-muted); margin: 10px 0 4px; }
    .adm-row { display: flex; gap: 12px; flex-wrap: wrap; }
    .adm-row > div { flex: 1; min-width: 160px; }
    .adm-btn {
      display: inline-block; padding: 9px 18px; border-radius: 8px; border: none; cursor: pointer;
      background: var(--pf-purple); color: #fff; font-size: .85rem; font-weight: 600; margin-top: 12px;
    }
    .adm-btn--ghost { background: transparent; border: 1px solid var(--border); color: var(--text); }
    .adm-btn--danger { background: var(--pf-danger); }
    .adm-btn:disabled { opacity: .5; cursor: default; }
    .adm table { width: 100%; border-collapse: collapse; font-size: .8rem; }
    .adm th { text-align: left; color: var(--text-muted); font-size: .68rem; letter-spacing: .1em; text-transform: uppercase; padding: 6px 8px; }
    .adm td { padding: 6px 8px; border-top: 1px solid var(--border-2); }
    .adm-badge { display: inline-block; padding: 2px 8px; border-radius: 999px; font-size: .65rem; font-weight: 700; }
    .adm-badge--latest { background: rgba(55,210,156,.15); color: var(--pf-success); }
    .adm-badge--doc { background: rgba(255,255,255,.08); color: var(--text-muted); }
    .adm-badge--lock { background: rgba(143,0,255,.15); color: var(--pf-purple); }
    .adm-badge--dl { background: rgba(143,0,255,.12); color: var(--pf-purple); }
    .adm-badge--armed { background: rgba(255,92,143,.18); color: var(--pf-danger); margin-left: 6px; }
    .pack-card { border: 1px solid var(--border-2); border-radius: 10px; padding: 12px 14px; margin-bottom: 10px; background: var(--surface-2); }
    .pack-head { display: flex; justify-content: space-between; align-items: center; gap: 10px; flex-wrap: wrap; }
    .pack-head-right { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
    .pack-meta { margin: 6px 0 2px; }
    .pack-edit { margin-top: 10px; border-top: 1px solid var(--border-2); padding-top: 10px; }
    .sd-block { margin-top: 8px; border-top: 1px solid var(--border-2); padding-top: 8px; }
    .sd-block summary { cursor: pointer; font-size: .8rem; color: var(--text); }
    .sd-arm { display: flex; gap: 8px; align-items: flex-start; text-transform: none; letter-spacing: 0; font-size: .78rem; margin: 10px 0; color: var(--text-2); }
    .sd-arm input, .sd-mod input { width: auto; margin-top: 2px; }
    .sd-mods { display: flex; flex-direction: column; gap: 4px; max-height: 240px; overflow-y: auto; padding: 8px; background: var(--surface-3); border-radius: 8px; }
    .sd-mod { display: flex; gap: 8px; align-items: center; text-transform: none; letter-spacing: 0; font-size: .74rem; margin: 0; }
    .adm-msg { margin-top: 10px; font-size: .8rem; padding: 8px 12px; border-radius: 8px; display: none; }
    .adm-msg.ok { display: block; background: rgba(55,210,156,.12); color: var(--pf-success); }
    .adm-msg.err { display: block; background: rgba(255,92,143,.12); color: var(--pf-danger); }
    .adm-chart { width: 100%; height: 180px; background: var(--surface-3); border-radius: 10px; }
    .adm-bars { display: flex; flex-direction: column; gap: 6px; margin-top: 10px; }
    .adm-bar { display: grid; grid-template-columns: 130px 1fr 60px; gap: 10px; align-items: center; font-size: .78rem; }
    .adm-bar-track { height: 10px; border-radius: 999px; background: var(--surface-3); overflow: hidden; }
    .adm-bar-fill { height: 100%; border-radius: 999px; background: var(--pf-purple); }
    .adm-stat-tiles { display: flex; gap: 14px; flex-wrap: wrap; margin-bottom: 16px; }
    .adm-stat-tile { flex: 1; min-width: 140px; background: var(--surface-2); border: 1px solid var(--border-2); border-radius: 10px; padding: 14px; }
    .adm-stat-tile b { font-size: 1.4rem; display: block; }
    .adm-stat-tile span { font-size: .68rem; letter-spacing: .1em; text-transform: uppercase; color: var(--text-muted); }
    .adm-section-title { margin: 0 0 12px; font-size: 1rem; }
    .adm-inline { display: flex; gap: 8px; align-items: flex-end; flex-wrap: wrap; }
    .adm-inline > * { margin-top: 0 !important; }
    .adm-small { font-size: .72rem; color: var(--text-muted); }
    .adm details.adm-help { margin-top: 10px; }
    .adm details.adm-help summary { cursor: pointer; font-size: .78rem; color: var(--pf-purple); }
    .adm-example {
      margin: 10px 0 6px; padding: 12px 14px; border-radius: 8px;
      background: var(--surface-3); border: 1px solid var(--border-2);
      font-family: "JetBrains Mono", monospace; font-size: .72rem; line-height: 1.5;
      white-space: pre; overflow-x: auto; color: var(--text);
    }
    #loginView { max-width: 380px; margin: 12vh auto 0; }
    [hidden] { display: none !important; }
  </style>
</head>
<body>
  <div class="adm">

    <!-- ── Login ─────────────────────────────── -->
    <div id="loginView" hidden>
      <div class="adm-card">
        <h1 style="margin:0 0 6px">Poly<span style="color:var(--pf-purple)">Forge</span> Admin</h1>
        <p class="adm-small">Manage releases, versions, packs, and stats without cPanel.</p>
        <label for="loginPass">Password</label>
        <input type="password" id="loginPass" autocomplete="current-password" />
        <button class="adm-btn" id="loginBtn" type="button">Sign in</button>
        <div class="adm-msg" id="loginMsg"></div>
      </div>
    </div>

    <!-- ── Panel ─────────────────────────────── -->
    <div id="panelView" hidden>
      <div class="adm-head">
        <h1>Poly<span>Forge</span> Admin</h1>
        <button class="adm-btn adm-btn--ghost" id="logoutBtn" type="button" style="margin:0">Sign out</button>
      </div>

      <div class="adm-tabs" id="tabs">
        <button class="adm-tab is-active" data-tab="stats" type="button">Stats</button>
        <button class="adm-tab" data-tab="releases" type="button">Releases</button>
        <button class="adm-tab" data-tab="manifest" type="button">Version &amp; Manifest</button>
        <button class="adm-tab" data-tab="packs" type="button">Packs</button>
        <button class="adm-tab" data-tab="packager" type="button">Packager</button>
      </div>

      <!-- Stats -->
      <section data-panel="stats">
        <div class="adm-stat-tiles" id="statTiles"></div>
        <div class="adm-card">
          <h2 class="adm-section-title">Downloads per day</h2>
          <canvas class="adm-chart" id="historyChart" width="1040" height="180"></canvas>
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">By download type</h2>
          <div class="adm-bars" id="typeBars"></div>
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">Modpack downloads</h2>
          <div class="adm-bars" id="packBars"></div>
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">By file (per version)</h2>
          <table><thead><tr><th>File</th><th style="text-align:right">Downloads</th></tr></thead>
          <tbody id="fileRows"></tbody></table>
        </div>
      </section>

      <!-- Releases -->
      <section data-panel="releases" hidden>
        <div class="adm-card">
          <h2 class="adm-section-title">Upload a build</h2>
          <div class="adm-row">
            <div><label>Type folder</label><select id="upType"></select></div>
            <div><label>Or create new type</label>
              <div class="adm-inline">
                <input type="text" id="newType" placeholder="e.g. windows-arm64" />
                <button class="adm-btn adm-btn--ghost" id="newTypeBtn" type="button">Create</button>
              </div>
            </div>
          </div>
          <label>Build file</label>
          <input type="file" id="upFile" />
          <button class="adm-btn" id="upBtn" type="button">Upload</button>
          <div class="adm-msg" id="relMsg"></div>
          <p class="adm-small">The newest non-doc file in a type folder is what
          <code>/api/download?type=&lt;folder&gt;</code> serves. Older files stay for rollback.</p>
        </div>
        <div id="relTypes"></div>
      </section>

      <!-- Manifest -->
      <section data-panel="manifest" hidden>
        <div class="adm-card">
          <h2 class="adm-section-title">App version control</h2>
          <div class="adm-row">
            <div><label>Latest version (soft update)</label><input type="text" id="mLatest" /></div>
            <div><label>Min supported (hard update)</label><input type="text" id="mMinSup" /></div>
          </div>
          <label>Download URL</label><input type="text" id="mDlUrl" />
          <label>Release notes</label><input type="text" id="mNotes" />
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">Full manifest (packs, option overrides, visibility)</h2>
          <p class="adm-small">The app fetches this on every launch. <code>disabledOptions</code> hides
          launchers, <code>optionOverrides</code> renames them, <code>modpacks</code> is the pack list.</p>
          <details class="adm-help">
            <summary>Examples: optionOverrides &amp; disabledOptions</summary>
            <pre class="adm-example">"optionOverrides": [
  {
    "id": "vanilla",
    "title": "Vanilla (Turtel SMP)",
    "description": "Install into the official Minecraft launcher."
  },
  { "id": "modrinth", "title": "Modrinth App" }
],
"disabledOptions": ["technic", "bakaxl", "qwertz"]</pre>
            <p class="adm-small"><b>optionOverrides</b> — array of <code>{ "id", "title"?, "description"? }</code>.
            Only the fields you include are changed; omit <code>title</code>/<code>description</code> to keep
            the built-in text. <b>disabledOptions</b> — array of the same ids to hide from the installer menu.
            Both accept <code>[]</code> (empty) to change nothing.</p>
            <p class="adm-small">Valid ids (the launcher/installer options): <code>vanilla</code>,
            <code>multimc</code>, <code>curseforge</code>, <code>modrinth</code>, <code>gdlauncher</code>,
            <code>atlauncher</code>, <code>prismlauncher</code>, <code>bakaxl</code>, <code>feather</code>,
            <code>technic</code>, <code>polymc</code>, <code>sklauncher</code>, <code>freesm</code>,
            <code>elyprism</code>, <code>shatteredprism</code>, <code>qwertz</code>.</p>
          </details>
          <textarea id="mRaw" spellcheck="false"></textarea>
          <button class="adm-btn" id="mSaveBtn" type="button">Save manifest</button>
          <div class="adm-msg" id="mMsg"></div>
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">History</h2>
          <table><thead><tr><th>Saved</th><th>Latest</th><th>Min supported</th><th></th></tr></thead>
          <tbody id="mHistory"></tbody></table>
        </div>
      </section>

      <!-- Packs -->
      <section data-panel="packs" hidden>
        <div class="adm-card">
          <h2 class="adm-section-title">Pack registry <span class="adm-small">— read automatically from the packs/ folder</span></h2>
          <div id="packList"></div>
          <h2 class="adm-section-title" style="margin-top:20px">Pre-register a pack</h2>
          <p class="adm-small">Packs listed above have an inline <b>Edit</b> button — use it to change their
          name, password or URL. This form is only for pre-registering a pack that isn't hosted yet; the ID
          must match the pack's id. Saving here (or above) updates the public manifest automatically, so the
          app picks it up on next launch.</p>
          <div class="adm-row">
            <div><label>Pack ID</label><input type="text" id="pkId" placeholder="turtel-smp" /></div>
            <div><label>Name</label><input type="text" id="pkName" /></div>
          </div>
          <div class="adm-row">
            <div><label>Set password (blank = keep)</label><input type="text" id="pkPass" /></div>
            <div><label>Download URL (blank = none)</label><input type="text" id="pkUrl" /></div>
          </div>
          <div class="adm-inline" style="margin-top:10px">
            <label style="margin:0;display:flex;gap:6px;align-items:center;text-transform:none">
              <input type="checkbox" id="pkReq" style="width:auto" /> Requires password
            </label>
          </div>
          <button class="adm-btn" id="pkSaveBtn" type="button">Save pack</button>
          <div class="adm-msg" id="pkMsg"></div>
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">Hosted pack files</h2>
          <table><thead><tr><th>File</th><th>Size</th><th>Updated</th></tr></thead>
          <tbody id="hostedRows"></tbody></table>
        </div>
      </section>

      <!-- Packager -->
      <section data-panel="packager" hidden>
        <div class="adm-card">
          <h2 class="adm-section-title">Online modpack packager</h2>
          <p class="adm-small">Point it at your profile folder (the one containing <code>mods/</code>,
          <code>config/</code>, ...) — pick the folder directly, or upload a zip of it. Only pack-worthy
          folders are kept (mods, config, resourcepacks, shaderpacks, datapacks, defaultconfigs, scripts,
          kubejs + options.txt / servers.dat); saves, logs, journeymap and other user data are dropped
          automatically. When you pick a folder, the pack-worthy files are zipped in your browser first,
          so only those get uploaded. Mod names/versions are read from inside the jars.</p>
          <div class="adm-row">
            <div><label>Pack ID</label><input type="text" id="bId" placeholder="barebones-s5" /></div>
            <div><label>Name</label><input type="text" id="bName" placeholder="Barebones Season 5" /></div>
            <div><label>Version</label><input type="text" id="bVer" placeholder="1.0.0" /></div>
          </div>
          <div class="adm-row">
            <div><label>Minecraft</label><input type="text" id="bMc" placeholder="1.21.1" /></div>
            <div><label>Loader</label>
              <select id="bLoader">
                <option value="">(none)</option><option>fabric</option><option>quilt</option>
                <option>forge</option><option>neoforge</option><option>vanilla</option>
              </select>
            </div>
            <div><label>Loader version</label><input type="text" id="bLoaderV" /></div>
          </div>
          <div class="adm-row">
            <div><label>Source folder</label><input type="file" id="bFolder" webkitdirectory directory multiple /></div>
            <div><label>…or a source zip</label><input type="file" id="bZip" accept=".zip" /></div>
          </div>
          <button class="adm-btn" id="bBuildBtn" type="button">Build pack</button>
          <div class="adm-msg" id="bMsg"></div>
          <p class="adm-small">Large packs may exceed the host's upload limit — use
          the local packager script below for those and upload the result
          under Hosted pack files.</p>
        </div>
        <div class="adm-card">
          <h2 class="adm-section-title">Local packager script</h2>
          <p class="adm-small">Run this on your machine to build a <code>.polypack</code> with no
          upload limit:<br>
          <code>pwsh package-modpack.ps1 -SourceDir &lt;profile&gt; -PackId &lt;id&gt; -PackName "&lt;name&gt;" -PackVersion 1.0.0</code><br>
          Download <b>both</b> files into the <b>same folder</b> —
          <code>package-modpack.ps1</code> dot-sources <code>slime-lib.ps1</code> at runtime.</p>
          <div class="adm-inline">
            <a class="adm-btn" href="/tools/package-modpack.ps1" download>Download package-modpack.ps1</a>
            <a class="adm-btn adm-btn--ghost" href="/tools/slime-lib.ps1" download>Download slime-lib.ps1</a>
          </div>
        </div>
      </section>
    </div>
  </div>

  <script>
    (() => {
      "use strict";
      const $ = (s) => document.querySelector(s);
      // Extensionless: .php URLs get 301-redirected, which would turn POSTs
      // into GETs. The rewrite rules serve api/admin.php for this path.
      const API = "/api/admin";

      async function call(action, opts = {}) {
        const init = { method: opts.method || "GET", headers: { "X-PolyForge-Admin": "1" } };
        if (opts.json) {
          init.method = "POST";
          init.headers["Content-Type"] = "application/json";
          init.body = JSON.stringify(opts.json);
        } else if (opts.form) {
          init.method = "POST";
          init.body = opts.form;
        }
        const res = await fetch(`${API}?action=${action}`, init);
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
        return data;
      }

      function msg(el, text, ok) {
        el.textContent = text;
        el.className = "adm-msg " + (ok ? "ok" : "err");
        if (ok) setTimeout(() => { el.className = "adm-msg"; }, 4000);
      }
      const esc = (s) => String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
      const fmtSize = (b) => b > 1048576 ? (b / 1048576).toFixed(1) + " MB" : Math.round(b / 1024) + " KB";
      // Canonical pack id: lowercase, spaces -> hyphens, collapsed, trimmed —
      // mirrors normalizePackId() on the server so the UI and API agree.
      const normId = (s) => s.trim().toLowerCase().replace(/\s+/g, "-").replace(/-+/g, "-").replace(/^-+|-+$/g, "");

      // ── Auth ───────────────────────────────────
      async function checkSession() {
        try {
          const s = await call("session");
          show(s.authenticated);
        } catch { show(false); }
      }
      function show(authed) {
        $("#loginView").hidden = authed;
        $("#panelView").hidden = !authed;
        if (authed) { loadStats(); loadReleases(); loadManifest(); loadPacks(); }
      }
      $("#loginBtn").addEventListener("click", async () => {
        try {
          await call("login", { json: { password: $("#loginPass").value } });
          $("#loginPass").value = "";
          show(true);
        } catch (e) { msg($("#loginMsg"), e.message, false); }
      });
      $("#loginPass").addEventListener("keydown", (e) => { if (e.key === "Enter") $("#loginBtn").click(); });
      $("#logoutBtn").addEventListener("click", async () => { await call("logout", { json: {} }); show(false); });

      // ── Tabs ───────────────────────────────────
      $("#tabs").addEventListener("click", (e) => {
        const tab = e.target.closest(".adm-tab");
        if (!tab) return;
        document.querySelectorAll(".adm-tab").forEach(t => t.classList.toggle("is-active", t === tab));
        document.querySelectorAll("[data-panel]").forEach(p => { p.hidden = p.dataset.panel !== tab.dataset.tab; });
      });

      // ── Stats ──────────────────────────────────
      async function loadStats() {
        const { stats } = await call("stats");
        const total = stats.downloads || 0;
        const byType = stats.byType || {};
        const byFile = stats.byFile || {};
        const history = stats.history || {};

        const days = Object.keys(history).sort();
        const last30 = days.slice(-30);
        const todayKey = new Date().toISOString().slice(0, 10);
        $("#statTiles").innerHTML = `
          <div class="adm-stat-tile"><b>${total.toLocaleString()}</b><span>Total downloads</span></div>
          <div class="adm-stat-tile"><b>${(history[todayKey]?.total || 0).toLocaleString()}</b><span>Today</span></div>
          <div class="adm-stat-tile"><b>${Object.keys(byType).length}</b><span>Active types</span></div>
          <div class="adm-stat-tile"><b>${esc(stats.updated ? stats.updated.slice(0, 10) : "-")}</b><span>Last download</span></div>`;

        drawHistory($("#historyChart"), last30.map(d => ({ day: d, v: history[d].total || 0 })));

        const maxType = Math.max(1, ...Object.values(byType));
        $("#typeBars").innerHTML = Object.entries(byType).sort((a, b) => b[1] - a[1]).map(([t, v]) => `
          <div class="adm-bar"><span class="mono">${esc(t)}</span>
            <div class="adm-bar-track"><div class="adm-bar-fill" style="width:${(v / maxType * 100).toFixed(1)}%"></div></div>
            <b style="text-align:right">${v.toLocaleString()}</b></div>`).join("") || '<p class="adm-small">No downloads yet.</p>';

        const byPack = stats.byPack || {};
        const maxPack = Math.max(1, ...Object.values(byPack));
        $("#packBars").innerHTML = Object.entries(byPack).sort((a, b) => b[1] - a[1]).map(([t, v]) => `
          <div class="adm-bar"><span class="mono">${esc(t)}</span>
            <div class="adm-bar-track"><div class="adm-bar-fill" style="width:${(v / maxPack * 100).toFixed(1)}%"></div></div>
            <b style="text-align:right">${v.toLocaleString()}</b></div>`).join("") || '<p class="adm-small">No modpack downloads yet.</p>';

        $("#fileRows").innerHTML = Object.entries(byFile).sort((a, b) => b[1] - a[1]).map(([f, v]) =>
          `<tr><td class="mono">${esc(f)}</td><td style="text-align:right">${v.toLocaleString()}</td></tr>`).join("") ||
          '<tr><td colspan="2" class="adm-small">No downloads yet.</td></tr>';
      }

      function drawHistory(canvas, points) {
        const ctx = canvas.getContext("2d");
        const W = canvas.width, H = canvas.height, pad = 28;
        ctx.clearRect(0, 0, W, H);
        const accent = getComputedStyle(document.documentElement).getPropertyValue("--pf-purple").trim() || "#8f00ff";
        const muted = getComputedStyle(document.documentElement).getPropertyValue("--text-muted").trim() || "#888";
        if (points.length === 0) {
          ctx.fillStyle = muted; ctx.font = "13px monospace";
          ctx.fillText("No download history yet", pad, H / 2);
          return;
        }
        const max = Math.max(1, ...points.map(p => p.v));
        const step = (W - pad * 2) / Math.max(1, points.length - 1);
        ctx.strokeStyle = accent; ctx.lineWidth = 2; ctx.beginPath();
        points.forEach((p, i) => {
          const x = pad + i * step;
          const y = H - pad - (p.v / max) * (H - pad * 2);
          i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y);
        });
        ctx.stroke();
        ctx.fillStyle = accent;
        points.forEach((p, i) => {
          const x = pad + i * step;
          const y = H - pad - (p.v / max) * (H - pad * 2);
          ctx.beginPath(); ctx.arc(x, y, 3, 0, Math.PI * 2); ctx.fill();
        });
        ctx.fillStyle = muted; ctx.font = "10px monospace";
        ctx.fillText(points[0].day.slice(5), pad - 10, H - 8);
        ctx.fillText(points[points.length - 1].day.slice(5), W - pad - 24, H - 8);
        ctx.fillText(String(max), 4, pad + 4);
      }

      // ── Releases ───────────────────────────────
      async function loadReleases() {
        const { types } = await call("releases-list");
        $("#upType").innerHTML = types.map(t => `<option>${esc(t.type)}</option>`).join("");
        $("#relTypes").innerHTML = types.map(t => `
          <div class="adm-card">
            <h2 class="adm-section-title mono">${esc(t.type)}
              <span class="adm-small">— /api/download?type=${esc(t.type)}</span></h2>
            <table><thead><tr><th>File</th><th>Size</th><th>Updated</th><th></th><th></th></tr></thead><tbody>
            ${t.files.map(f => `<tr>
              <td class="mono">${esc(f.name)}</td>
              <td>${fmtSize(f.size)}</td>
              <td>${esc(f.mtime.slice(0, 16).replace("T", " "))}</td>
              <td>${f.name === t.latest ? '<span class="adm-badge adm-badge--latest">LATEST</span>' : (f.doc ? '<span class="adm-badge adm-badge--doc">doc</span>' : "")}</td>
              <td><button class="adm-btn adm-btn--danger" style="margin:0;padding:4px 10px;font-size:.7rem"
                data-del-type="${esc(t.type)}" data-del-name="${esc(f.name)}" type="button">Delete</button></td>
            </tr>`).join("") || '<tr><td colspan="5" class="adm-small">Empty — upload a build.</td></tr>'}
            </tbody></table>
          </div>`).join("") || '<div class="adm-card"><p class="adm-small">No type folders yet — create one above (e.g. windows, linux, macos).</p></div>';

        $("#relTypes").querySelectorAll("[data-del-type]").forEach(btn => btn.addEventListener("click", async () => {
          if (!confirm(`Delete ${btn.dataset.delName}?`)) return;
          try {
            await call("release-delete", { json: { type: btn.dataset.delType, name: btn.dataset.delName } });
            loadReleases();
          } catch (e) { msg($("#relMsg"), e.message, false); }
        }));
      }
      $("#newTypeBtn").addEventListener("click", async () => {
        try {
          await call("release-type-create", { json: { type: $("#newType").value.trim() } });
          $("#newType").value = "";
          msg($("#relMsg"), "Type folder created.", true);
          loadReleases();
        } catch (e) { msg($("#relMsg"), e.message, false); }
      });
      $("#upBtn").addEventListener("click", async () => {
        const file = $("#upFile").files[0];
        if (!file) { msg($("#relMsg"), "Choose a file first.", false); return; }
        const form = new FormData();
        form.append("type", $("#upType").value);
        form.append("file", file);
        $("#upBtn").disabled = true;
        try {
          const r = await call("release-upload", { form });
          msg($("#relMsg"), `Uploaded. SHA-256: ${r.sha256}`, true);
          $("#upFile").value = "";
          loadReleases();
        } catch (e) { msg($("#relMsg"), e.message, false); }
        finally { $("#upBtn").disabled = false; }
      });

      // ── Manifest ───────────────────────────────
      let manifestCache = null;
      async function loadManifest() {
        const { manifest } = await call("manifest-get");
        manifestCache = manifest;
        $("#mLatest").value = manifest.app?.latestVersion || "";
        $("#mMinSup").value = manifest.app?.minSupportedVersion || "";
        $("#mDlUrl").value = manifest.app?.downloadUrl || "";
        $("#mNotes").value = manifest.app?.notes || "";
        $("#mRaw").value = JSON.stringify(manifest, null, 2);
        loadManifestHistory();
      }
      async function loadManifestHistory() {
        const { entries } = await call("manifest-history");
        $("#mHistory").innerHTML = entries.slice(0, 20).map((e, i) => `<tr>
          <td>${esc((e.saved || "").slice(0, 16).replace("T", " "))}</td>
          <td class="mono">${esc(e.latestVersion)}</td>
          <td class="mono">${esc(e.minSupported)}</td>
          <td><button class="adm-btn adm-btn--ghost" style="margin:0;padding:4px 10px;font-size:.7rem"
            data-restore="${i}" type="button">Restore</button></td>
        </tr>`).join("") || '<tr><td colspan="4" class="adm-small">No saves yet.</td></tr>';
        $("#mHistory").querySelectorAll("[data-restore]").forEach(btn => btn.addEventListener("click", () => {
          $("#mRaw").value = JSON.stringify(entries[Number(btn.dataset.restore)].manifest, null, 2);
          msg($("#mMsg"), "Restored into the editor — review and press Save.", true);
        }));
      }
      // Keep the quick fields and the raw JSON in sync (fields win on save).
      $("#mSaveBtn").addEventListener("click", async () => {
        let manifest;
        try { manifest = JSON.parse($("#mRaw").value); }
        catch { msg($("#mMsg"), "Manifest is not valid JSON.", false); return; }
        manifest.app = manifest.app || {};
        if ($("#mLatest").value.trim()) manifest.app.latestVersion = $("#mLatest").value.trim();
        if ($("#mMinSup").value.trim()) manifest.app.minSupportedVersion = $("#mMinSup").value.trim();
        manifest.app.downloadUrl = $("#mDlUrl").value.trim();
        manifest.app.notes = $("#mNotes").value;
        manifest.updated = new Date().toISOString();
        try {
          await call("manifest-save", { json: { manifest } });
          msg($("#mMsg"), "Manifest saved — apps pick it up on next launch.", true);
          loadManifest();
        } catch (e) { msg($("#mMsg"), e.message, false); }
      });

      // ── Packs ──────────────────────────────────
      async function loadPacks() {
        const { packs, hosted } = await call("packs-list");
        $("#packList").innerHTML = packs.length
          ? packs.map(renderPackCard).join("")
          : '<p class="adm-small">No packs yet — build one in the Packager tab, or drop a .polypack + .manifest.json into packs/.</p>';
        $("#hostedRows").innerHTML = hosted.map(f => `<tr>
          <td class="mono">${esc(f.name)}</td><td>${fmtSize(f.size)}</td>
          <td>${esc(f.mtime.slice(0, 16).replace("T", " "))}</td></tr>`).join("") ||
          '<tr><td colspan="3" class="adm-small">No hosted pack files.</td></tr>';
        wirePackCards();
      }

      function renderPackCard(p) {
        const removeSet = new Set(p.removeMods || []);
        const armed = removeSet.size > 0;
        const pw = p.requiresPassword
          ? `<span class="adm-badge adm-badge--lock">${p.hasPassword ? "PASSWORD SET" : "PASSWORD MISSING"}</span>`
          : '<span class="adm-small">open</span>';
        const src = [p.inFolder ? "folder" : null, p.inManifest ? "manifest" : null].filter(Boolean).join(" + ") || "registry only";
        const mods = p.mods || [];
        const modRows = mods.length
          ? mods.map(m => `<label class="sd-mod"><input type="checkbox" data-sd-mod value="${esc(m.file)}" ${removeSet.has(m.file) ? "checked" : ""} ${armed ? "" : "disabled"} /><span class="mono">${esc(m.file)}</span></label>`).join("")
          : '<p class="adm-small">No mod list yet — upload this pack\'s .manifest.json to the packs/ folder.</p>';
        return `<div class="pack-card" data-pack="${esc(p.id)}">
          <div class="pack-head">
            <div><b class="mono">${esc(p.id)}</b> ${esc(p.name)} ${p.version ? `<span class="adm-small">v${esc(p.version)}</span>` : ""}</div>
            <div class="pack-head-right">
              <span class="adm-badge adm-badge--dl">${(p.downloads || 0).toLocaleString()} downloads</span>
              ${pw}
              <button class="adm-btn adm-btn--ghost pack-edit-toggle" style="margin:0;padding:4px 10px;font-size:.7rem" type="button">Edit</button>
              <button class="adm-btn adm-btn--danger pack-del" style="margin:0;padding:4px 10px;font-size:.7rem" type="button">Delete</button>
            </div>
          </div>
          <div class="adm-small pack-meta">in: ${esc(src)}${p.file ? ` &middot; <span class="mono">${esc(p.file)}</span>` : ""}${p.downloadUrl ? ` &middot; url: <span class="mono">${esc(p.downloadUrl)}</span>` : ""}</div>
          <div class="pack-edit" hidden>
            <div class="adm-row">
              <div><label>Name</label><input type="text" class="pe-name" value="${esc(p.name)}" /></div>
              <div><label>Download URL (blank = none)</label><input type="text" class="pe-url" value="${esc(p.downloadUrl || "")}" /></div>
            </div>
            <div class="adm-row">
              <div><label>Set password (blank = keep)</label><input type="text" class="pe-pass" placeholder="${p.hasPassword ? "password set — blank keeps it" : "no password set"}" /></div>
              <div style="display:flex;align-items:flex-end">
                <label style="margin:0;display:flex;gap:6px;align-items:center;text-transform:none">
                  <input type="checkbox" class="pe-req" style="width:auto" ${p.requiresPassword ? "checked" : ""} /> Requires password
                </label>
              </div>
            </div>
            <button class="adm-btn pe-save" type="button">Save changes</button>
            <div class="adm-msg pe-msg"></div>
          </div>
          <details class="sd-block">
            <summary>&#128163; Self-destruct — remove proprietary mods on next launch ${armed ? `<span class="adm-badge adm-badge--armed">ARMED &middot; ${removeSet.size}</span>` : ""}</summary>
            <label class="sd-arm"><input type="checkbox" data-sd-arm ${armed ? "checked" : ""} /> Arm self-destruct — the checked mods are deleted from every install of this pack the next time the app runs.</label>
            <div class="sd-mods">${modRows}</div>
            <button class="adm-btn sd-save" type="button">Save self-destruct</button>
            <div class="adm-msg sd-msg"></div>
          </details>
        </div>`;
      }

      function wirePackCards() {
        $("#packList").querySelectorAll(".pack-card").forEach(card => {
          const id = card.dataset.pack;
          const arm = card.querySelector("[data-sd-arm]");
          const modBoxes = () => [...card.querySelectorAll("[data-sd-mod]")];
          arm.addEventListener("change", () => modBoxes().forEach(m => { m.disabled = !arm.checked; }));

          // Inline metadata edit — pre-filled from the card, saved straight to
          // the registry AND the public manifest (the app reads the manifest).
          const editForm = card.querySelector(".pack-edit");
          card.querySelector(".pack-edit-toggle").addEventListener("click", () => { editForm.hidden = !editForm.hidden; });
          card.querySelector(".pe-save").addEventListener("click", async () => {
            const peMsg = card.querySelector(".pe-msg");
            try {
              await call("pack-save-meta", { json: {
                id,
                name: card.querySelector(".pe-name").value.trim(),
                requiresPassword: card.querySelector(".pe-req").checked,
                password: card.querySelector(".pe-pass").value,
                downloadUrl: card.querySelector(".pe-url").value.trim(),
              } });
              msg(peMsg, "Saved — manifest updated; apps pick it up on next launch.", true);
              loadPacks();
            } catch (e) { msg(peMsg, e.message, false); }
          });
          card.querySelector(".pack-del").addEventListener("click", async () => {
            if (!confirm(`Remove pack "${id}" from the registry (and its hosted files)?`)) return;
            try { await call("pack-delete", { json: { id, deleteFiles: true } }); loadPacks(); }
            catch (e) { msg($("#pkMsg"), e.message, false); }
          });
          card.querySelector(".sd-save").addEventListener("click", async () => {
            const removeMods = modBoxes().filter(m => m.checked).map(m => m.value);
            const sdMsg = card.querySelector(".sd-msg");
            if (arm.checked && removeMods.length &&
                !confirm(`Arm self-destruct for "${id}"?\n\nThese mods will be DELETED from every existing install of this pack the next time the app runs:\n\n${removeMods.join("\n")}`)) return;
            try {
              const r = await call("pack-selfdestruct-save", { json: { id, armed: arm.checked, removeMods } });
              msg(sdMsg, r.armed ? `Armed — ${r.removeMods.length} mod(s) removed on next launch.` : "Disarmed — nothing will be removed.", true);
              loadPacks();
            } catch (e) { msg(sdMsg, e.message, false); }
          });
        });
      }
      $("#pkSaveBtn").addEventListener("click", async () => {
        try {
          await call("pack-save-meta", { json: {
            id: normId($("#pkId").value),
            name: $("#pkName").value.trim(),
            requiresPassword: $("#pkReq").checked,
            password: $("#pkPass").value,
            downloadUrl: $("#pkUrl").value.trim(),
          } });
          msg($("#pkMsg"), "Pack saved — manifest updated automatically.", true);
          $("#pkPass").value = "";
          loadPacks();
        } catch (e) { msg($("#pkMsg"), e.message, false); }
      });

      // ── Packager ───────────────────────────────
      // A folder pick and a zip upload feed the same server endpoint: a folder is
      // filtered to pack-worthy files and zipped in the browser (below), so only
      // those bytes upload and the server sees an ordinary source zip either way.
      const PACK_FOLDERS = ["mods", "config", "resourcepacks", "shaderpacks", "datapacks", "defaultconfigs", "scripts", "kubejs"];
      const PACK_ROOT_FILES = ["options.txt", "servers.dat"];

      const CRC_TABLE = (() => {
        const t = new Uint32Array(256);
        for (let n = 0; n < 256; n++) { let c = n; for (let k = 0; k < 8; k++) c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1); t[n] = c >>> 0; }
        return t;
      })();
      const crc32 = (b) => { let c = 0xFFFFFFFF; for (let i = 0; i < b.length; i++) c = (c >>> 8) ^ CRC_TABLE[(c ^ b[i]) & 0xFF]; return (c ^ 0xFFFFFFFF) >>> 0; };
      // Little-endian record builder: fields are [byteWidth, value] pairs.
      const zrec = (fields) => {
        let len = 0; for (const [n] of fields) len += n;
        const b = new Uint8Array(len), dv = new DataView(b.buffer); let p = 0;
        for (const [n, v] of fields) { if (n === 2) dv.setUint16(p, v, true); else dv.setUint32(p, v >>> 0, true); p += n; }
        return b;
      };
      // Minimal store-only (no compression) ZIP writer — enough for ZipArchive to
      // read on the server, which recompresses into the .polypack anyway. Avoids
      // pulling in any external zip library (the site ships zero third-party JS).
      function buildStoreZip(entries) {
        const enc = new TextEncoder(), parts = [], central = []; let offset = 0;
        const push = (a) => { parts.push(a); offset += a.length; };
        const D = 0x0021, T = 0; // DOS 1980-01-01 00:00 — a valid, fixed timestamp
        for (const e of entries) {
          const name = enc.encode(e.name), data = e.data, crc = crc32(data), off = offset;
          push(zrec([[4, 0x04034b50], [2, 20], [2, 0], [2, 0], [2, T], [2, D], [4, crc], [4, data.length], [4, data.length], [2, name.length], [2, 0]]));
          push(name); push(data);
          central.push({ name, size: data.length, crc, off });
        }
        const cdStart = offset; let cdSize = 0; const cd = [];
        for (const c of central) {
          const h = zrec([[4, 0x02014b50], [2, 20], [2, 20], [2, 0], [2, 0], [2, T], [2, D], [4, c.crc], [4, c.size], [4, c.size], [2, c.name.length], [2, 0], [2, 0], [2, 0], [2, 0], [4, 0], [4, c.off]]);
          cd.push(h, c.name); cdSize += h.length + c.name.length;
        }
        const eocd = zrec([[4, 0x06054b50], [2, 0], [2, 0], [2, central.length], [2, central.length], [4, cdSize], [4, cdStart], [2, 0]]);
        return new Blob([...parts, ...cd, eocd], { type: "application/zip" });
      }
      // Turns a picked profile folder into a source zip: strips the top folder the
      // browser prepends, keeps only pack-worthy paths, and zips them. null = the
      // folder held nothing shippable.
      async function zipProfileFolder(fileList) {
        const picked = [];
        for (const f of fileList) {
          const rel = (f.webkitRelativePath || f.name).split("/").slice(1).join("/");
          if (!rel || rel.includes("..")) continue;
          const parts = rel.split("/");
          const keep = (parts.length > 1 && PACK_FOLDERS.includes(parts[0])) ||
                       (parts.length === 1 && PACK_ROOT_FILES.includes(parts[0]));
          if (keep) picked.push({ rel, file: f });
        }
        if (!picked.length) return null;
        const entries = [];
        for (const { rel, file } of picked) entries.push({ name: rel, data: new Uint8Array(await file.arrayBuffer()) });
        return buildStoreZip(entries);
      }

      $("#bBuildBtn").addEventListener("click", async () => {
        const folderFiles = [...($("#bFolder").files || [])];
        const zip = $("#bZip").files[0];
        if (!folderFiles.length && !zip) { msg($("#bMsg"), "Pick your profile folder (or a zip of it) first.", false); return; }

        const form = new FormData();
        form.append("id", normId($("#bId").value));
        form.append("name", $("#bName").value.trim());
        form.append("version", $("#bVer").value.trim());
        form.append("minecraft", $("#bMc").value.trim());
        form.append("loader", $("#bLoader").value);
        form.append("loaderVersion", $("#bLoaderV").value.trim());
        $("#bBuildBtn").disabled = true;
        try {
          let source = zip;
          if (folderFiles.length) {
            msg($("#bMsg"), "Packing folder in your browser…", true);
            source = await zipProfileFolder(folderFiles);
            if (!source) { msg($("#bMsg"), "No pack-worthy folders (mods/, config/, …) found in that folder.", false); return; }
          }
          form.append("source", source, "source.zip");
          msg($("#bMsg"), "Building… (large packs take a while)", true);
          const r = await call("pack-build", { form });
          msg($("#bMsg"), `Built ${r.pack} — ${r.mods} mods, ${r.files} files (${fmtSize(r.bytes)}). Folders: ${r.folders.join(", ")}`, true);
          loadPacks();
        } catch (e) { msg($("#bMsg"), e.message, false); }
        finally { $("#bBuildBtn").disabled = false; }
      });

      // Pack ids are always lowercase with spaces as hyphens — fold as the user
      // types so what they see matches what gets saved. Both transforms are 1:1
      // on length, so the caret stays put (full collapse/trim happens on send via
      // normId, and the server normalizes + rejects other symbols authoritatively).
      ["#pkId", "#bId"].forEach(sel => {
        const el = $(sel);
        if (el) el.addEventListener("input", () => {
          const start = el.selectionStart;
          el.value = el.value.toLowerCase().replace(/\s/g, "-");
          if (start !== null) el.setSelectionRange(start, start);
        });
      });

      checkSession();
    })();
  </script>
</body>
</html>
