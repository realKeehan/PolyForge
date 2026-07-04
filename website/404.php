<?php
http_response_code(404);
$pageTitle       = 'Page not found - PolyForge';
$pageDescription = "The page you're looking for doesn't exist.";
$pageSlug        = '404';
$noIndex         = true;
require __DIR__ . '/partials/header.php';
?>

  <main id="main">
    <div class="container page-hero" style="min-height:55vh;display:flex;flex-direction:column;justify-content:center;align-items:center;text-align:center">
      <p class="eyebrow mono">Error 404</p>
      <h1 class="h1" style="margin-bottom:12px">Page not found.</h1>
      <p class="lead" style="max-width:520px">
        This page got lost somewhere between launchers. Check the URL,
        or head back to safety below.
      </p>
      <div class="cta-row" style="justify-content:center">
        <a class="btn btn-primary" href="./">Back to home</a>
        <a class="btn btn-ghost" href="./downloads">Downloads</a>
      </div>
      <p class="mono muted" style="margin-top:48px;font-size:.7rem;opacity:.45;letter-spacing:.2em" title="Old cheat codes never die">
        &uarr; &uarr; &darr; &darr; &larr; &rarr; &larr; &rarr; B A
      </p>
    </div>
  </main>

  <!-- ─── Konami Tetris easter egg ────────────────── -->
  <div class="tetris-overlay" id="tetrisOverlay" hidden>
    <div class="tetris-shell" role="dialog" aria-modal="true" aria-label="PolyTris">
      <div class="tetris-head">
        <span class="tetris-title mono">POLY<span>TRIS</span></span>
        <button class="tetris-close" id="tetrisClose" type="button" aria-label="Close game">&times;</button>
      </div>
      <div class="tetris-layout">
        <div class="tetris-side">
          <div class="tetris-panel">
            <div class="tetris-panel-label mono">HOLD</div>
            <canvas id="tetrisHold" width="80" height="80"></canvas>
          </div>
          <div class="tetris-panel tetris-stats mono">
            <div><span>SCORE</span><b id="tetrisScore">0</b></div>
            <div><span>LINES</span><b id="tetrisLines">0</b></div>
            <div><span>LEVEL</span><b id="tetrisLevel">1</b></div>
            <div><span>BEST</span><b id="tetrisBest">-</b></div>
          </div>
        </div>
        <div class="tetris-board-wrap">
          <canvas id="tetrisBoard" width="240" height="480"></canvas>
          <div class="tetris-msg mono" id="tetrisMsg" hidden></div>
        </div>
        <div class="tetris-side">
          <div class="tetris-panel">
            <div class="tetris-panel-label mono">NEXT</div>
            <canvas id="tetrisNext" width="80" height="80"></canvas>
          </div>
          <div class="tetris-panel tetris-keys mono">
            <div>&larr; &rarr; move</div>
            <div>&darr; soft drop</div>
            <div>&uarr;/X rotate &middot; Z ccw</div>
            <div>Space hard drop</div>
            <div>C hold &middot; P pause</div>
            <div>Esc quit</div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <style>
    /* ── PolyTris (404 easter egg) ──────────────── */
    .tetris-overlay {
      position: fixed;
      inset: 0;
      z-index: 200;
      display: flex;
      align-items: center;
      justify-content: center;
      background: color-mix(in srgb, var(--bg) 78%, transparent);
      backdrop-filter: blur(8px);
    }
    .tetris-shell {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 16px;
      box-shadow: 0 24px 70px rgba(0, 0, 0, .45);
      padding: 14px 18px 18px;
      max-width: calc(100vw - 32px);
    }
    .tetris-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
    }
    .tetris-title { font-weight: 600; letter-spacing: .25em; }
    .tetris-title span { color: var(--pf-purple); }
    .tetris-close {
      background: none;
      border: 1px solid var(--border-2);
      border-radius: 8px;
      color: var(--text);
      width: 30px; height: 30px;
      font-size: 1.05rem;
      cursor: pointer;
      line-height: 1;
    }
    .tetris-close:hover { border-color: var(--pf-purple); color: var(--pf-purple); }
    .tetris-layout { display: flex; gap: 14px; align-items: stretch; }
    .tetris-side { display: flex; flex-direction: column; gap: 12px; width: 104px; }
    .tetris-panel {
      background: var(--surface-3);
      border: 1px solid var(--border-2);
      border-radius: 10px;
      padding: 8px;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 6px;
    }
    .tetris-panel-label { font-size: .62rem; letter-spacing: .3em; color: var(--text-muted); }
    .tetris-stats { align-items: stretch; gap: 8px; font-size: .68rem; }
    .tetris-stats div { display: flex; justify-content: space-between; gap: 6px; }
    .tetris-stats span { color: var(--text-muted); }
    .tetris-keys { align-items: flex-start; gap: 4px; font-size: .6rem; color: var(--text-muted); }
    .tetris-board-wrap { position: relative; }
    #tetrisBoard {
      display: block;
      background: var(--surface-3);
      border: 1px solid var(--border);
      border-radius: 10px;
    }
    #tetrisHold, #tetrisNext { display: block; }
    .tetris-msg {
      position: absolute;
      inset: 0;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 12px;
      text-align: center;
      background: color-mix(in srgb, var(--bg) 76%, transparent);
      border-radius: 10px;
      font-size: .85rem;
      letter-spacing: .12em;
      padding: 12px;
    }
    .tetris-msg-title {
      font-size: .92rem;
      letter-spacing: .3em;
      color: var(--pf-purple);
      font-weight: 600;
    }
    .tetris-msg-sub { font-size: .68rem; color: var(--text-muted); letter-spacing: .18em; }
    .tetris-msg-hint { font-size: .6rem; color: var(--text-muted); letter-spacing: .14em; line-height: 1.7; }

    /* Retro name entry */
    .tetris-entry-slots { display: flex; gap: 8px; }
    .tetris-slot {
      width: 28px; height: 38px;
      display: flex; align-items: center; justify-content: center;
      border: 1px solid var(--border);
      border-bottom-width: 3px;
      border-radius: 6px;
      background: var(--surface-2);
      font-size: 1.15rem;
      font-weight: 600;
    }
    .tetris-slot.is-cursor {
      border-color: var(--pf-purple);
      box-shadow: 0 0 12px color-mix(in srgb, var(--pf-purple) 45%, transparent);
      animation: tetrisBlink 1s steps(2, start) infinite;
    }
    @keyframes tetrisBlink {
      50% { background: color-mix(in srgb, var(--pf-purple) 30%, var(--surface-2)); }
    }

    /* High-score billboard */
    .tetris-lb { border-collapse: collapse; font-size: .7rem; letter-spacing: .14em; }
    .tetris-lb td { padding: 2px 7px; text-align: left; white-space: pre; }
    .tetris-lb td.num { text-align: right; }
    .tetris-lb tr.is-you td { color: var(--pf-purple); font-weight: 700; }
    .tetris-lb tr.is-top td:nth-child(2) { color: var(--pf-warning); }

    @media (max-width: 620px) {
      .tetris-layout { flex-direction: column; align-items: center; }
      .tetris-side { flex-direction: row; width: auto; }
    }
  </style>

  <script>
    (() => {
      "use strict";

      // ── Konami code listener ─────────────────────
      const KONAMI = ["ArrowUp","ArrowUp","ArrowDown","ArrowDown","ArrowLeft","ArrowRight","ArrowLeft","ArrowRight","KeyB","KeyA"];
      let konamiIdx = 0;
      let game = null;

      document.addEventListener("keydown", (e) => {
        if (game) return; // game handles its own keys while open
        konamiIdx = (e.code === KONAMI[konamiIdx]) ? konamiIdx + 1 : (e.code === KONAMI[0] ? 1 : 0);
        if (konamiIdx === KONAMI.length) {
          konamiIdx = 0;
          openGame();
        }
      });

      // ── Leaderboard store (server + local fallback) ──
      const SCORE_API = "/api/scores";
      const LS_SCORES = "pf-tetris-scores";
      const LS_NAME = "pf-tetris-name";
      const LB_SIZE = 10;

      function localScores() {
        try { return JSON.parse(localStorage.getItem(LS_SCORES) || "[]"); } catch { return []; }
      }

      async function fetchScores() {
        try {
          const res = await fetch(SCORE_API, { cache: "no-store" });
          if (res.ok) {
            const data = await res.json();
            if (Array.isArray(data.scores)) return data.scores;
          }
        } catch { /* offline / no backend */ }
        return localScores();
      }

      async function submitScore(entry) {
        try {
          const res = await fetch(SCORE_API, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(entry),
          });
          if (res.ok) {
            const data = await res.json();
            if (Array.isArray(data.scores)) return data.scores;
          }
        } catch { /* offline / no backend */ }
        // Local fallback so the billboard still works without the API
        const list = localScores();
        list.push({ ...entry, date: new Date().toISOString() });
        list.sort((a, b) => b.score - a.score);
        const top = list.slice(0, LB_SIZE);
        try { localStorage.setItem(LS_SCORES, JSON.stringify(top)); } catch { /* storage full/blocked */ }
        return top;
      }

      function escHtml(s) {
        return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
      }

      // ── PolyTris ─────────────────────────────────
      const COLS = 10, ROWS = 20, CELL = 24;
      const SHAPES = {
        I: [[1,0],[1,1],[1,2],[1,3]],
        J: [[0,0],[1,0],[1,1],[1,2]],
        L: [[0,2],[1,0],[1,1],[1,2]],
        O: [[0,1],[0,2],[1,1],[1,2]],
        S: [[0,1],[0,2],[1,0],[1,1]],
        T: [[0,1],[1,0],[1,1],[1,2]],
        Z: [[0,0],[0,1],[1,1],[1,2]],
      };
      const NAMES = Object.keys(SHAPES);
      const ENTRY_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 ";
      const NAME_LEN = 5;

      function cssVar(name, fallback) {
        const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
        return v || fallback;
      }

      function palette() {
        const accent = cssVar("--pf-purple", "#8f00ff");
        return {
          I: accent,
          T: cssVar("--pf-success", "#37d29c"),
          O: cssVar("--pf-warning", "#ffc14d"),
          S: "#4dc9ff",
          Z: cssVar("--pf-danger", "#ff5c8f"),
          J: "#7a7dff",
          L: "#d98cff",
          grid: cssVar("--border-2", "rgba(255,255,255,.06)"),
          ghost: cssVar("--text-muted", "#6e6490"),
        };
      }

      function rotate(cells) {
        // rotate within 4x4 box: (r,c) -> (c, 3-r), then normalize into top-left
        const r = cells.map(([y, x]) => [x, 3 - y]);
        const minY = Math.min(...r.map(c => c[0]));
        const minX = Math.min(...r.map(c => c[1]));
        return r.map(([y, x]) => [y - minY, x - minX]);
      }

      function openGame() {
        const overlay = document.getElementById("tetrisOverlay");
        overlay.hidden = false;
        game = createGame(() => {
          overlay.hidden = true;
          game = null;
        });
      }

      function createGame(onExit) {
        const board = document.getElementById("tetrisBoard");
        const bctx = board.getContext("2d");
        const nctx = document.getElementById("tetrisNext").getContext("2d");
        const hctx = document.getElementById("tetrisHold").getContext("2d");
        const scoreEl = document.getElementById("tetrisScore");
        const linesEl = document.getElementById("tetrisLines");
        const levelEl = document.getElementById("tetrisLevel");
        const bestEl = document.getElementById("tetrisBest");
        const msgEl = document.getElementById("tetrisMsg");
        const closeBtn = document.getElementById("tetrisClose");
        const colors = palette();

        let grid, bag, current, next, hold, holdUsed;
        let score, lines, level, dropMs, paused;
        let lastDrop = 0, raf = 0;

        // mode: "play" | "entry" (name input) | "over" (billboard shown)
        let mode = "play";
        let leaderboard = [];
        let entryName, entryCursor;

        fetchScores().then((scores) => {
          leaderboard = scores;
          updateBest();
        });

        function updateBest() {
          const top = leaderboard[0];
          bestEl.textContent = top ? String(top.score) : "-";
        }

        function qualifies(s) {
          if (s <= 0) return false;
          if (leaderboard.length < LB_SIZE) return true;
          return s > leaderboard[leaderboard.length - 1].score;
        }

        function refillBag() {
          const b = [...NAMES];
          for (let i = b.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [b[i], b[j]] = [b[j], b[i]];
          }
          return b;
        }

        function takePiece() {
          if (bag.length === 0) bag = refillBag();
          const name = bag.pop();
          return { name, cells: SHAPES[name].map(c => [...c]), row: 0, col: 3 };
        }

        function reset() {
          grid = Array.from({ length: ROWS }, () => Array(COLS).fill(null));
          bag = refillBag();
          current = takePiece();
          next = takePiece();
          hold = null;
          holdUsed = false;
          score = 0; lines = 0; level = 1;
          dropMs = 800;
          paused = false;
          mode = "play";
          msgEl.hidden = true;
          updateStats();
          drawPreviews();
        }

        function collides(piece, dRow, dCol, cells) {
          const body = cells || piece.cells;
          return body.some(([y, x]) => {
            const r = piece.row + y + dRow;
            const c = piece.col + x + dCol;
            return c < 0 || c >= COLS || r >= ROWS || (r >= 0 && grid[r][c]);
          });
        }

        function lockPiece() {
          current.cells.forEach(([y, x]) => {
            const r = current.row + y, c = current.col + x;
            if (r >= 0) grid[r][c] = current.name;
          });
          let cleared = 0;
          for (let r = ROWS - 1; r >= 0; r--) {
            if (grid[r].every(Boolean)) {
              grid.splice(r, 1);
              grid.unshift(Array(COLS).fill(null));
              cleared++;
              r++;
            }
          }
          if (cleared) {
            lines += cleared;
            score += [0, 100, 300, 500, 800][cleared] * level;
            const newLevel = Math.floor(lines / 10) + 1;
            if (newLevel !== level) {
              level = newLevel;
              dropMs = Math.max(90, 800 - (level - 1) * 70);
            }
          }
          current = next;
          next = takePiece();
          holdUsed = false;
          drawPreviews();
          updateStats();
          if (collides(current, 0, 0)) endGame();
        }

        // ── Game over → name entry → billboard ─────
        function endGame() {
          if (qualifies(score)) {
            startNameEntry();
          } else {
            mode = "over";
            renderBillboard(null);
          }
        }

        function startNameEntry() {
          mode = "entry";
          const saved = (localStorage.getItem(LS_NAME) || "AAAAA").toUpperCase();
          entryName = Array.from({ length: NAME_LEN }, (_, i) =>
            ENTRY_CHARS.includes(saved[i] || " ") ? (saved[i] || " ") : "A");
          entryCursor = 0;
          renderEntry();
        }

        function renderEntry() {
          const slots = entryName.map((ch, i) =>
            `<span class="tetris-slot${i === entryCursor ? " is-cursor" : ""}">${ch === " " ? "&nbsp;" : escHtml(ch)}</span>`
          ).join("");
          msgEl.innerHTML = `
            <div class="tetris-msg-title">NEW HIGH SCORE</div>
            <div class="tetris-msg-sub">SCORE ${score}</div>
            <div class="tetris-entry-slots">${slots}</div>
            <div class="tetris-msg-hint">TYPE OR &uarr;&darr; TO PICK &middot; &larr;&rarr; MOVE<br>ENTER SAVES &middot; ESC SKIPS</div>
          `;
          msgEl.hidden = false;
        }

        function cycleChar(dir) {
          const idx = ENTRY_CHARS.indexOf(entryName[entryCursor]);
          const nextIdx = (idx + dir + ENTRY_CHARS.length) % ENTRY_CHARS.length;
          entryName[entryCursor] = ENTRY_CHARS[nextIdx];
          renderEntry();
        }

        function confirmEntry() {
          const name = entryName.join("").replace(/\s+$/, "");
          if (!name) { cycleChar(0); return; } // require at least one character
          try { localStorage.setItem(LS_NAME, name); } catch { /* ignore */ }
          mode = "over";
          msgEl.innerHTML = `<div class="tetris-msg-title">SAVING...</div>`;
          const entry = { name, score, lines, level };
          submitScore(entry).then((scores) => {
            leaderboard = scores;
            updateBest();
            renderBillboard(entry);
          });
        }

        function renderBillboard(you) {
          const rows = leaderboard.slice(0, LB_SIZE).map((s, i) => {
            const isYou = you && !s._marked && s.name === you.name && s.score === you.score;
            if (isYou) s._marked = true; // highlight only the first match
            const cls = [isYou ? "is-you" : "", i === 0 ? "is-top" : ""].filter(Boolean).join(" ");
            return `<tr${cls ? ` class="${cls}"` : ""}>` +
              `<td class="num">${i + 1}.</td>` +
              `<td>${escHtml(String(s.name || "?").padEnd(NAME_LEN))}</td>` +
              `<td class="num">${escHtml(String(s.score))}</td></tr>`;
          }).join("");

          msgEl.innerHTML = `
            <div class="tetris-msg-title">HIGH SCORES</div>
            ${rows ? `<table class="tetris-lb"><tbody>${rows}</tbody></table>` : `<div class="tetris-msg-sub">NO SCORES YET</div>`}
            <div class="tetris-msg-sub">GAME OVER &middot; SCORE ${score}</div>
            <div class="tetris-msg-hint">R RESTART &middot; ESC QUIT</div>
          `;
          msgEl.hidden = false;
        }

        // ── Moves ───────────────────────────────────
        function move(dCol) {
          if (!collides(current, 0, dCol)) { current.col += dCol; }
        }

        function softDrop() {
          if (!collides(current, 1, 0)) { current.row++; score += 1; updateStats(); }
          else lockPiece();
        }

        function hardDrop() {
          let dist = 0;
          while (!collides(current, dist + 1, 0)) dist++;
          current.row += dist;
          score += dist * 2;
          lockPiece();
        }

        function tryRotate(ccw) {
          let cells = rotate(current.cells);
          if (ccw) { cells = rotate(rotate(cells)); }
          for (const kick of [0, -1, 1, -2, 2]) {
            if (!collides(current, 0, kick, cells)) {
              current.cells = cells;
              current.col += kick;
              return;
            }
          }
        }

        function doHold() {
          if (holdUsed) return;
          holdUsed = true;
          const prev = hold;
          hold = { name: current.name, cells: SHAPES[current.name].map(c => [...c]) };
          if (prev) {
            current = { name: prev.name, cells: prev.cells, row: 0, col: 3 };
          } else {
            current = next;
            next = takePiece();
          }
          if (collides(current, 0, 0)) endGame();
          drawPreviews();
        }

        function updateStats() {
          scoreEl.textContent = String(score);
          linesEl.textContent = String(lines);
          levelEl.textContent = String(level);
        }

        // ── Rendering ───────────────────────────────
        function drawCell(ctx, x, y, size, color) {
          ctx.fillStyle = color;
          ctx.fillRect(x, y, size - 1, size - 1);
          ctx.fillStyle = "rgba(255,255,255,.25)";
          ctx.fillRect(x, y, size - 1, 2);
          ctx.fillStyle = "rgba(0,0,0,.25)";
          ctx.fillRect(x, y + size - 3, size - 1, 2);
        }

        function drawMini(ctx, piece) {
          ctx.clearRect(0, 0, 80, 80);
          if (!piece) return;
          const cells = SHAPES[piece.name];
          const ys = cells.map(c => c[0]), xs = cells.map(c => c[1]);
          const h = Math.max(...ys) - Math.min(...ys) + 1;
          const w = Math.max(...xs) - Math.min(...xs) + 1;
          const size = 16;
          const offX = (80 - w * size) / 2 - Math.min(...xs) * size;
          const offY = (80 - h * size) / 2 - Math.min(...ys) * size;
          cells.forEach(([y, x]) => drawCell(ctx, offX + x * size, offY + y * size, size, colors[piece.name]));
        }

        function drawPreviews() {
          drawMini(nctx, next);
          drawMini(hctx, hold);
        }

        function draw() {
          bctx.clearRect(0, 0, board.width, board.height);

          bctx.fillStyle = colors.grid;
          for (let r = 1; r < ROWS; r++)
            for (let c = 1; c < COLS; c++)
              bctx.fillRect(c * CELL, r * CELL, 1, 1);

          for (let r = 0; r < ROWS; r++)
            for (let c = 0; c < COLS; c++)
              if (grid[r][c]) drawCell(bctx, c * CELL, r * CELL, CELL, colors[grid[r][c]]);

          if (mode === "play") {
            let dist = 0;
            while (!collides(current, dist + 1, 0)) dist++;
            bctx.globalAlpha = 0.22;
            current.cells.forEach(([y, x]) => {
              const r = current.row + y + dist, c = current.col + x;
              if (r >= 0) drawCell(bctx, c * CELL, r * CELL, CELL, colors.ghost);
            });
            bctx.globalAlpha = 1;

            current.cells.forEach(([y, x]) => {
              const r = current.row + y, c = current.col + x;
              if (r >= 0) drawCell(bctx, c * CELL, r * CELL, CELL, colors[current.name]);
            });
          }
        }

        function frame(now) {
          if (mode === "play" && !paused && now - lastDrop >= dropMs) {
            lastDrop = now;
            if (!collides(current, 1, 0)) current.row++;
            else lockPiece();
          }
          draw();
          raf = requestAnimationFrame(frame);
        }

        function togglePause() {
          paused = !paused;
          msgEl.innerHTML = `<div class="tetris-msg-title">PAUSED</div>`;
          msgEl.hidden = !paused;
        }

        // ── Input ───────────────────────────────────
        function entryKey(e) {
          if (e.code === "Escape") { mode = "over"; renderBillboard(null); return; }
          if (e.code === "Enter" || e.code === "NumpadEnter") { confirmEntry(); return; }
          if (e.code === "ArrowLeft") { entryCursor = (entryCursor + NAME_LEN - 1) % NAME_LEN; renderEntry(); return; }
          if (e.code === "ArrowRight") { entryCursor = (entryCursor + 1) % NAME_LEN; renderEntry(); return; }
          if (e.code === "ArrowUp") { cycleChar(1); return; }
          if (e.code === "ArrowDown") { cycleChar(-1); return; }
          if (e.code === "Backspace") {
            entryName[entryCursor] = " ";
            entryCursor = Math.max(0, entryCursor - 1);
            renderEntry();
            return;
          }
          const ch = e.key.length === 1 ? e.key.toUpperCase() : "";
          if (ch && ENTRY_CHARS.includes(ch)) {
            entryName[entryCursor] = ch;
            entryCursor = Math.min(NAME_LEN - 1, entryCursor + 1);
            renderEntry();
          }
        }

        function onKey(e) {
          if (mode === "entry") {
            e.preventDefault();
            e.stopPropagation();
            entryKey(e);
            return;
          }

          if (e.code === "Escape") { e.preventDefault(); exit(); return; }

          if (mode === "over") {
            if (e.code === "KeyR") { e.preventDefault(); reset(); }
            return;
          }

          switch (e.code) {
            case "ArrowLeft": e.preventDefault(); if (!paused) move(-1); break;
            case "ArrowRight": e.preventDefault(); if (!paused) move(1); break;
            case "ArrowDown": e.preventDefault(); if (!paused) softDrop(); break;
            case "ArrowUp":
            case "KeyX": e.preventDefault(); if (!paused) tryRotate(false); break;
            case "KeyZ": e.preventDefault(); if (!paused) tryRotate(true); break;
            case "Space": e.preventDefault(); if (!paused) hardDrop(); break;
            case "KeyC":
            case "ShiftLeft":
            case "ShiftRight": e.preventDefault(); if (!paused) doHold(); break;
            case "KeyP": e.preventDefault(); togglePause(); break;
          }
        }

        function exit() {
          cancelAnimationFrame(raf);
          document.removeEventListener("keydown", onKey, true);
          closeBtn.removeEventListener("click", exit);
          onExit();
        }

        document.addEventListener("keydown", onKey, true);
        closeBtn.addEventListener("click", exit);

        reset();
        lastDrop = performance.now();
        raf = requestAnimationFrame(frame);

        return { exit };
      }
    })();
  </script>

<?php require __DIR__ . '/partials/footer.php'; ?>
