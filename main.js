// ══════════════════════════════════════════════════
// PolyForge — Shared JS (all pages)
// ══════════════════════════════════════════════════
(() => {
  "use strict";

  // ── Helpers ────────────────────────────────────
  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

  function esc(s) {
    const d = document.createElement("div");
    d.appendChild(document.createTextNode(s));
    return d.innerHTML;
  }

  // ── Launchers (shared data) ────────────────────
  const launchers = [
    // Supported
    { name:"Vanilla Launcher", status:"supported", note:"The official Mojang/Microsoft launcher — the baseline for standard Minecraft installs.", url:"https://www.minecraft.net/en-us/download" },
    { name:"MultiMC", status:"supported", note:"A lightweight multi-instance launcher focused on custom modded setups and clean instance management.", url:"https://multimc.org/" },
    { name:"CurseForge", status:"supported", note:"A popular ecosystem for modpacks with built-in browsing, installs, and updates through the CurseForge platform.", url:"https://www.curseforge.com/" },
    { name:"Modrinth (Theseus)", status:"supported", note:"Modrinth's launcher/profile system — modern pack distribution with a fast-growing mod ecosystem.", url:"https://modrinth.com/" },
    { name:"Custom Path", status:"supported", note:"For nonstandard installs, portable environments, or advanced setups where you want full control of location.", url:null },
    { name:"Manual Install", status:"supported", note:"For users who prefer to manage placement themselves or need a pack output for custom workflows.", url:null },

    // In progress
    { name:"Prism Launcher", status:"working", note:"A modern MultiMC fork with broader platform support, active development, and power-user features.", url:"https://prismlauncher.org/" },
    { name:"ATLauncher", status:"working", note:"A long-running launcher built around curated packs and easy modded profiles.", url:"https://atlauncher.com/" },
    { name:"GDLauncher", status:"working", note:"A sleek launcher that emphasizes a friendly UI and integrated pack browsing/management.", url:"https://gdlauncher.com/" },
    { name:"Technic", status:"working", note:"One of the classic launcher platforms — known for legacy packs and older modpack history.", url:"https://www.technicpack.net/" },
    { name:"PolyMC", status:"working", note:"A Prism/MultiMC-family launcher — similar instance philosophy with community-driven tooling.", url:"https://polymc.org/" },
    { name:"Feather", status:"working", note:"A performance-focused launcher often used for competitive play and client-side enhancements.", url:"https://feathermc.com/" },
    { name:"BakaXL", status:"working", note:"A launcher favored by some modded communities, especially in regions where it's widely adopted.", url:"https://www.bakaxl.com/" },

    // Planned — new launchers requested
    { name:"SK Launcher", status:"planned", note:"An all-in-one Minecraft hub with built-in modloaders, modpack support, and skin management.", url:"https://skmedix.pl/" },
    { name:"Freesm Launcher", status:"planned", note:"A Prism-based launcher that removes offline account restrictions and adds custom auth server support.", url:"https://freesmlauncher.org/" },
    { name:"ElyPrism", status:"planned", note:"A Prism Launcher fork with Ely.by authentication integration for alternative account systems.", url:"https://elyprismlauncher.github.io/" },
    { name:"ShatteredPrism", status:"planned", note:"A community-maintained Prism Launcher fork focused on extended features and flexibility.", url:"https://github.com/Noctilune/ShatteredPrism" },
    { name:"QWERTZ Launcher", status:"planned", note:"A launcher from the QWERTZ project ecosystem with streamlined Minecraft instance management.", url:"https://qwertz.app/projects/" },
    { name:"Fjord Launcher", status:"planned", note:"An Unmojang project — a Prism-family launcher with its own community-driven direction.", url:"https://github.com/unmojang/FjordLauncher" },
    { name:"HMCL", status:"planned", note:"A cross-platform Minecraft launcher popular in the Chinese community, supporting multiple auth and mod sources.", url:"https://hmcl.huangyuhui.net/" },
    { name:"UltimMC", status:"planned", note:"A MultiMC fork focused on offline play support and community-driven development.", url:"https://github.com/UltimMC/Launcher" },

    // Catch-all
    { name:"Additional ecosystems", status:"planned", note:"More launcher adapters as the ecosystem evolves — prioritized by demand and stability.", url:null },
  ];

  const meta = {
    supported: { label:"Supported", badge:"badge-supported" },
    working:   { label:"In progress", badge:"badge-working" },
    planned:   { label:"Planned", badge:"badge-planned" }
  };

  // ── Theme ──────────────────────────────────────
  function getPreferredTheme() {
    const saved = localStorage.getItem("pf-theme");
    if (saved === "light" || saved === "dark") return saved;
    return window.matchMedia?.("(prefers-color-scheme: light)").matches ? "light" : "dark";
  }

  function applyTheme(theme) {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem("pf-theme", theme);
  }

  function toggleTheme() {
    applyTheme(document.documentElement.dataset.theme === "light" ? "dark" : "light");
  }

  // ── Scroll ─────────────────────────────────────
  function getHeaderOffset() {
    const header = $("#header");
    return header ? Math.ceil(header.getBoundingClientRect().height + 10) : 0;
  }

  function scrollToHash(hash) {
    const el = document.querySelector(hash);
    if (!el) return;
    window.scrollTo({ top: Math.max(0, window.scrollY + el.getBoundingClientRect().top - getHeaderOffset()), behavior:"smooth" });
  }

  // ── Active nav highlighting ────────────────────
  function highlightActiveNav() {
    const current = location.pathname.split("/").pop() || "index.html";
    $$(".nav a, .mobile-menu a").forEach(a => {
      const href = a.getAttribute("href") || "";
      if (href === current || (current === "index.html" && href === "./")) {
        a.classList.add("is-active");
      }
    });
  }

  // ── Mobile menu ────────────────────────────────
  function initMobileMenu() {
    const btn = $("#menuBtn");
    const menu = $("#mobileMenu");
    if (!btn || !menu) return;
    btn.addEventListener("click", () => {
      const open = !menu.classList.contains("is-open");
      menu.classList.toggle("is-open", open);
      btn.setAttribute("aria-expanded", open ? "true" : "false");
    });
    menu.addEventListener("click", e => { if (e.target.closest("a")) { menu.classList.remove("is-open"); btn.setAttribute("aria-expanded","false"); }});
  }

  // ── Anchor jump handling ───────────────────────
  function initAnchorHandling() {
    $$("[data-jump]").forEach(el => {
      el.addEventListener("click", e => {
        const hash = el.getAttribute("data-jump") || "";
        if (!hash.startsWith("#")) return;
        e.preventDefault();
        scrollToHash(hash);
      });
    });
  }

  // ── Render launchers (used by supported.html & index.html) ──
  function renderLaunchers(gridId, opts = {}) {
    const grid = $(gridId || "#launcherGrid");
    if (!grid) return;

    const list = opts.subset ? launchers.filter(l => opts.subset.includes(l.status)) : launchers;

    grid.innerHTML = list.map(l => {
      const m = meta[l.status];
      const tag = l.url ? "a" : "article";
      const hrefAttr = l.url ? ` href="${esc(l.url)}" target="_blank" rel="noopener noreferrer"` : "";
      return `
        <${tag} class="launcher" data-status="${l.status}"${hrefAttr}>
          <div class="launcher-top">
            <div class="launcher-name">${esc(l.name)}</div>
            <span class="badge ${m.badge} mono">${esc(m.label)}</span>
          </div>
          <p class="launcher-note">${esc(l.note)}</p>
          ${l.url ? `<span class="launcher-link">Visit site →</span>` : ""}
        </${tag}>
      `;
    }).join("");

    // Update counts
    const counts = { supported:0, working:0, planned:0 };
    list.forEach(l => counts[l.status]++);

    const set = (id, val) => { const el = $(id); if (el) el.textContent = String(val); };
    set("#countSupported", counts.supported);
    set("#countWorking", counts.working);
    set("#countPlanned", counts.planned);
    set("#btnSupported", counts.supported);
    set("#btnWorking", counts.working);
    set("#btnPlanned", counts.planned);
    set("#btnAll", counts.supported + counts.working + counts.planned);
  }

  // ── Filters ────────────────────────────────────
  function initFilters() {
    const buttons = $$("[data-filter]");
    if (!buttons.length) return;

    const setActive = key => {
      buttons.forEach(b => {
        const active = b.dataset.filter === key;
        b.classList.toggle("is-active", active);
        b.setAttribute("aria-selected", String(active));
      });
      $$("#launcherGrid .launcher").forEach(card => {
        const status = card.getAttribute("data-status");
        card.classList.toggle("is-hidden", key !== "all" && status !== key);
      });
    };

    buttons.forEach(btn => btn.addEventListener("click", () => setActive(btn.dataset.filter)));
    setActive("all");
  }

  // ── FAQ accordion ──────────────────────────────
  function initFaqAccordion() {
    $$(".faq-question").forEach(btn => {
      btn.addEventListener("click", () => {
        const item = btn.closest(".faq-item");
        const answer = item.querySelector(".faq-answer");
        const isOpen = item.classList.contains("is-open");

        // close all
        $$(".faq-item.is-open").forEach(open => {
          open.classList.remove("is-open");
          open.querySelector(".faq-answer").style.maxHeight = "0";
        });

        if (!isOpen) {
          item.classList.add("is-open");
          answer.style.maxHeight = answer.scrollHeight + "px";
        }
      });
    });
  }

  // ── Dot field canvas (kept from original) ──────
  function initDotField() {
    const canvas = document.getElementById("dotCanvas");
    if (!(canvas instanceof HTMLCanvasElement)) return;
    const ctx = canvas.getContext("2d", { alpha: true });
    if (!ctx) return;
    if (window.matchMedia?.("(prefers-reduced-motion: reduce)").matches) return;

    let w = 0, h = 0;
    const dpr = Math.min(2, window.devicePixelRatio || 1);

    let spacing = 18, baseR = 1.15;
    const waveAmp = 0.85, waveSpeed = 0.32, noiseScale = 0.012, noiseOctaves = 3;
    const ripples = [];
    const maxRipples = 10;

    function resize() {
      const rect = canvas.getBoundingClientRect();
      w = Math.floor(rect.width);
      h = Math.floor(rect.height);
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      spacing = Math.max(14, Math.min(22, Math.round(Math.min(w, h) / 55)));
      baseR = spacing <= 16 ? 1.0 : 1.2;
    }

    function hash2(x, y) {
      let n = x * 374761393 + y * 668265263;
      n = (n ^ (n >> 13)) * 1274126177;
      return ((n ^ (n >> 16)) >>> 0) / 4294967295;
    }

    function smoothstep(t) { return t * t * (3 - 2 * t); }

    function valueNoise(x, y) {
      const xi = Math.floor(x), yi = Math.floor(y);
      const xf = x - xi, yf = y - yi;
      const u = smoothstep(xf), v = smoothstep(yf);
      const a = hash2(xi, yi) + (hash2(xi+1, yi) - hash2(xi, yi)) * u;
      const b = hash2(xi, yi+1) + (hash2(xi+1, yi+1) - hash2(xi, yi+1)) * u;
      return a + (b - a) * v;
    }

    function fbm(x, y, octaves) {
      let amp = 0.5, freq = 1, sum = 0, norm = 0;
      for (let i = 0; i < octaves; i++) {
        sum += amp * valueNoise(x * freq, y * freq);
        norm += amp; amp *= 0.5; freq *= 2;
      }
      return sum / Math.max(1e-6, norm);
    }

    function themeDotColor() {
      return (document.documentElement.dataset.theme || "dark") === "light"
        ? { r:60, g:40, b:80, a:0.18 }
        : { r:200, g:180, b:255, a:0.2 };
    }

    function themeAccentColor() {
      return (document.documentElement.dataset.theme || "dark") === "light"
        ? { r:143, g:0, b:255, a:0.1 }
        : { r:143, g:0, b:255, a:0.16 };
    }

    function addRipple(x, y) {
      ripples.unshift({ x, y, t:0, amp:1.35, freq:10.5, speed:0.85, decay:1.9 });
      if (ripples.length > maxRipples) ripples.pop();
    }

    window.addEventListener("pointermove", e => {
      if (Math.random() < 0.14) addRipple(e.clientX / window.innerWidth, e.clientY / window.innerHeight);
    }, { passive: true });

    window.addEventListener("pointerdown", e => {
      const x = e.clientX / window.innerWidth, y = e.clientY / window.innerHeight;
      addRipple(x, y); addRipple(x, y);
    }, { passive: true });

    const start = performance.now();

    function draw(now) {
      const t = (now - start) / 1000;
      ctx.clearRect(0, 0, w, h);
      const base = themeDotColor();
      const accent = themeAccentColor();
      for (const r of ripples) r.t += 0.016;

      const cols = Math.ceil(w / spacing), rows = Math.ceil(h / spacing);
      const tx = t * waveSpeed, ty = t * waveSpeed * 0.86;

      for (let iy = 0; iy <= rows; iy++) {
        const y = iy * spacing + spacing * 0.5;
        for (let ix = 0; ix <= cols; ix++) {
          const x = ix * spacing + spacing * 0.5;
          const n = fbm(x * noiseScale + tx, y * noiseScale + ty, noiseOctaves);
          const wave = (n - 0.5) * 2;
          let rippleSum = 0, rippleGlow = 0;
          const px = x / Math.max(1, w), py = y / Math.max(1, h);

          for (const r of ripples) {
            const dx = px - r.x, dy = py - r.y;
            const dist = Math.sqrt(dx*dx + dy*dy);
            const ring = Math.sin(dist * r.freq - t * r.speed - r.t * 1.2);
            const atten = Math.exp(-dist * r.decay * 6) * Math.exp(-r.t * 0.9);
            rippleSum += ring * atten * r.amp;
            rippleGlow += Math.max(0, ring) * atten;
          }

          const rad = Math.max(0.35, baseR + wave * waveAmp + rippleSum * 1.15);
          const a = base.a * (0.8 + 0.4 * n);
          const glow = Math.min(0.35, rippleGlow * 0.55);

          ctx.beginPath();
          ctx.fillStyle = `rgba(${base.r},${base.g},${base.b},${a.toFixed(4)})`;
          ctx.arc(x, y, rad, 0, Math.PI * 2);
          ctx.fill();

          if (glow > 0.01) {
            ctx.beginPath();
            ctx.fillStyle = `rgba(${accent.r},${accent.g},${accent.b},${(accent.a + glow).toFixed(4)})`;
            ctx.arc(x, y, rad * 1.05, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      }

      for (let i = ripples.length - 1; i >= 0; i--) {
        if (ripples[i].t > 3.5) ripples.splice(i, 1);
      }
      requestAnimationFrame(draw);
    }

    new ResizeObserver(() => resize()).observe(canvas);
    resize();
    requestAnimationFrame(draw);
  }

  // ── Init ───────────────────────────────────────
  document.addEventListener("DOMContentLoaded", () => {
    applyTheme(getPreferredTheme());

    // Theme toggles
    ["#themeToggle","#themeToggleSm"].forEach(id => {
      $(id)?.addEventListener("click", toggleTheme);
    });

    // Year
    const yearEl = $("#year");
    if (yearEl) yearEl.textContent = String(new Date().getFullYear());

    // Launchers
    renderLaunchers("#launcherGrid");
    initFilters();

    // Navigation
    highlightActiveNav();
    initMobileMenu();
    initAnchorHandling();

    // FAQ
    initFaqAccordion();

    // Dot field
    initDotField();
  });
})();
