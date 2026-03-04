// main.js
(() => {
  // Launchers: descriptions are ABOUT the launchers (not how PolyForge installs)
  const launchers = [
    {
      name: "Vanilla Launcher",
      status: "supported",
      note: "The official Mojang/Microsoft launcher — the baseline for standard Minecraft installs."
    },
    {
      name: "MultiMC",
      status: "supported",
      note: "A lightweight multi-instance launcher focused on custom modded setups and clean instance management."
    },
    {
      name: "CurseForge",
      status: "supported",
      note: "A popular ecosystem for modpacks with built-in browsing, installs, and updates through the CurseForge platform."
    },
    {
      name: "Modrinth (Theseus)",
      status: "supported",
      note: "Modrinth’s launcher/profile system — modern pack distribution with a fast-growing mod ecosystem."
    },
    {
      name: "Custom Path",
      status: "supported",
      note: "For nonstandard installs, portable environments, or advanced setups where you want full control of location."
    },
    {
      name: "Manual Install",
      status: "supported",
      note: "For users who prefer to manage placement themselves or need a pack output for custom workflows."
    },

    {
      name: "Prism Launcher",
      status: "working",
      note: "A modern MultiMC fork with broader platform support, active development, and power-user features."
    },
    {
      name: "ATLauncher",
      status: "working",
      note: "A long-running launcher built around curated packs and easy modded profiles."
    },
    {
      name: "GDLauncher",
      status: "working",
      note: "A sleek launcher that emphasizes a friendly UI and integrated pack browsing/management."
    },
    {
      name: "Technic",
      status: "working",
      note: "One of the classic launcher platforms — known for legacy packs and older modpack history."
    },
    {
      name: "PolyMC",
      status: "working",
      note: "A Prism/MultiMC-family launcher — similar instance philosophy with community-driven tooling."
    },
    {
      name: "Feather",
      status: "working",
      note: "A performance-focused launcher often used for competitive play and client-side enhancements."
    },
    {
      name: "BakaXL",
      status: "working",
      note: "A launcher favored by some modded communities, especially in regions where it’s widely adopted."
    },

    // Planned: keep a slot for “future ecosystems”
    {
      name: "Additional ecosystems",
      status: "planned",
      note: "More launcher adapters as the ecosystem evolves — prioritized by demand and stability."
    }
  ];

  const meta = {
    supported: { label: "Supported", badge: "badge-supported" },
    working: { label: "In progress", badge: "badge-working" },
    planned: { label: "Planned", badge: "badge-planned" }
  };

  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

  function escapeHTML(s) {
    return String(s)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function getPreferredTheme() {
    const saved = localStorage.getItem("pf-theme");
    if (saved === "light" || saved === "dark") return saved;
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
  }

  function applyTheme(theme) {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem("pf-theme", theme);
  }

  function toggleTheme() {
    const now = document.documentElement.dataset.theme === "light" ? "dark" : "light";
    applyTheme(now);
  }

  function getHeaderOffset() {
    const header = $("#header");
    if (!header) return 0;
    const h = header.getBoundingClientRect().height;
    return Math.ceil(h + 10);
  }

  function scrollToHash(hash) {
    const el = document.querySelector(hash);
    if (!el) return;
    const y = window.scrollY + el.getBoundingClientRect().top - getHeaderOffset();
    window.scrollTo({ top: Math.max(0, y), behavior: "smooth" });
  }

  function renderLaunchers() {
    const grid = $("#launcherGrid");
    if (!grid) return;

    grid.innerHTML = launchers.map((l) => {
      const m = meta[l.status];
      return `
        <article class="launcher" data-status="${l.status}">
          <div class="launcher-top">
            <div class="launcher-name">${escapeHTML(l.name)}</div>
            <span class="badge ${m.badge} mono">${escapeHTML(m.label)}</span>
          </div>
          <p class="launcher-note">${escapeHTML(l.note)}</p>
        </article>
      `;
    }).join("");

    const counts = {
      supported: launchers.filter(x => x.status === "supported").length,
      working: launchers.filter(x => x.status === "working").length,
      planned: launchers.filter(x => x.status === "planned").length
    };

    $("#countSupported").textContent = String(counts.supported);
    $("#countWorking").textContent = String(counts.working);
    $("#countPlanned").textContent = String(counts.planned);

    $("#btnSupported").textContent = String(counts.supported);
    $("#btnWorking").textContent = String(counts.working);
    $("#btnPlanned").textContent = String(counts.planned);
    $("#btnAll").textContent = String(counts.supported + counts.working + counts.planned);
  }

  function initFilters() {
    const buttons = $$("[data-filter]");

    const setActive = (key) => {
      buttons.forEach(b => {
        const active = b.dataset.filter === key;
        b.classList.toggle("is-active", active);
        b.setAttribute("aria-selected", active ? "true" : "false");
      });

      $$("#launcherGrid .launcher").forEach(card => {
        const status = card.getAttribute("data-status");
        const show = key === "all" || status === key;
        card.classList.toggle("is-hidden", !show);
      });
    };

    buttons.forEach(btn => btn.addEventListener("click", () => setActive(btn.dataset.filter)));
    setActive("all");
  }

  function initMobileMenu() {
    const btn = $("#menuBtn");
    const menu = $("#mobileMenu");
    if (!btn || !menu) return;

    const close = () => {
      menu.classList.remove("is-open");
      btn.setAttribute("aria-expanded", "false");
    };

    btn.addEventListener("click", () => {
      const open = !menu.classList.contains("is-open");
      menu.classList.toggle("is-open", open);
      btn.setAttribute("aria-expanded", open ? "true" : "false");
    });

    menu.addEventListener("click", (e) => {
      const a = e.target.closest("a");
      if (a) close();
    });
  }

  function initAnchorHandling() {
    $$(".nav a, .mobile-menu a, [data-jump]").forEach(el => {
      el.addEventListener("click", (e) => {
        const href = el.getAttribute("href") || el.getAttribute("data-jump") || "";
        if (!href.startsWith("#")) return;
        e.preventDefault();
        scrollToHash(href);
      });
    });
  }

  // ===== Dot grid like your CodePen, but with perlin-ish waves + mouse ripple waves =====
  function initDotField() {
    const canvas = document.getElementById("dotCanvas");
    if (!(canvas instanceof HTMLCanvasElement)) return;

    const ctx = canvas.getContext("2d", { alpha: true });
    if (!ctx) return;

    const prefersReduce = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (prefersReduce) return;

    let dpr = Math.max(1, Math.min(2, window.devicePixelRatio || 1));
    let w = 0, h = 0;

    // Grid config (tweak to taste)
    let spacing = 18;           // px (CSS pixels)
    let baseR = 1.15;           // base dot radius
    let waveAmp = 0.85;         // noise-driven radius amplitude
    let waveSpeed = 0.32;       // time speed
    let noiseScale = 0.012;     // spatial scale for noise
    let noiseOctaves = 3;

    // Mouse ripple waves
    const ripples = [];
    const maxRipples = 10;

    // Pointer state
    let pointerX = 0.5;
    let pointerY = 0.5;

    function resize() {
      const rect = canvas.getBoundingClientRect();
      w = Math.floor(rect.width);
      h = Math.floor(rect.height);
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

      // adapt spacing slightly to viewport
      spacing = Math.max(14, Math.min(22, Math.round(Math.min(w, h) / 55)));
      baseR = spacing <= 16 ? 1.0 : 1.2;
    }

    // Deterministic hash noise
    function hash2(x, y) {
      // integer-ish hash
      let n = x * 374761393 + y * 668265263; // large primes
      n = (n ^ (n >> 13)) * 1274126177;
      return ((n ^ (n >> 16)) >>> 0) / 4294967295;
    }

    function smoothstep(t) {
      return t * t * (3 - 2 * t);
    }

    // Value noise (2D) with bilinear interpolation
    function valueNoise(x, y) {
      const xi = Math.floor(x);
      const yi = Math.floor(y);
      const xf = x - xi;
      const yf = y - yi;

      const r00 = hash2(xi, yi);
      const r10 = hash2(xi + 1, yi);
      const r01 = hash2(xi, yi + 1);
      const r11 = hash2(xi + 1, yi + 1);

      const u = smoothstep(xf);
      const v = smoothstep(yf);

      const a = r00 + (r10 - r00) * u;
      const b = r01 + (r11 - r01) * u;
      return a + (b - a) * v; // 0..1
    }

    // Fractal noise (fbm)
    function fbm(x, y, octaves) {
      let amp = 0.5;
      let freq = 1.0;
      let sum = 0.0;
      let norm = 0.0;
      for (let i = 0; i < octaves; i++) {
        sum += amp * valueNoise(x * freq, y * freq);
        norm += amp;
        amp *= 0.5;
        freq *= 2.0;
      }
      return sum / Math.max(1e-6, norm); // 0..1
    }

    function themeDotColor() {
      const theme = document.documentElement.dataset.theme || "dark";
      if (theme === "light") {
        // subtle gray dots for light
        return { r: 0, g: 0, b: 0, a: 0.25 };
      }
      // subtle white dots for dark
      return { r: 255, g: 255, b: 255, a: 0.25 };
    }

    function themeAccentColor() {
      // purple-tinted response (still subtle)
      const theme = document.documentElement.dataset.theme || "dark";
      if (theme === "light") return { r: 143, g: 0, b: 255, a: 0.12 };
      return { r: 143, g: 0, b: 255, a: 0.16 };
    }

    function addRipple(x, y) {
      // store in normalized space (0..1)
      ripples.unshift({
        x,
        y,
        t: 0,
        // tuned so it looks like a “wave field”
        amp: 1.35,
        freq: 10.5,
        speed: 0.85,
        decay: 1.9
      });
      if (ripples.length > maxRipples) ripples.pop();
    }

    function onPointerMove(e) {
      pointerX = e.clientX / window.innerWidth;
      pointerY = e.clientY / window.innerHeight;

      // create ripples occasionally so it feels alive, not spammy
      // also add a ripple on click/touch separately below for stronger impact
      if (Math.random() < 0.14) addRipple(pointerX, pointerY);
    }

    function onPointerDown(e) {
      const x = e.clientX / window.innerWidth;
      const y = e.clientY / window.innerHeight;
      addRipple(x, y);
      addRipple(x, y); // double for a crisp “pulse”
    }

    window.addEventListener("pointermove", onPointerMove, { passive: true });
    window.addEventListener("pointerdown", onPointerDown, { passive: true });

    // Render loop
    let start = performance.now();

    function draw(now) {
      const t = (now - start) / 1000;

      ctx.clearRect(0, 0, w, h);

      const base = themeDotColor();
      const accent = themeAccentColor();

      // Fade ripples over time
      for (const r of ripples) r.t += 0.016; // approx frame step

      // grid bounds
      const cols = Math.ceil(w / spacing);
      const rows = Math.ceil(h / spacing);

      // small time offsets for moving waves
      const tx = t * waveSpeed;
      const ty = t * waveSpeed * 0.86;

      // draw dots
      for (let iy = 0; iy <= rows; iy++) {
        const y = iy * spacing + (spacing * 0.5);
        for (let ix = 0; ix <= cols; ix++) {
          const x = ix * spacing + (spacing * 0.5);

          const nx = x * noiseScale + tx;
          const ny = y * noiseScale + ty;

          // 0..1
          const n = fbm(nx, ny, noiseOctaves);

          // center around 0
          const wave = (n - 0.5) * 2.0;

          // ripple field contribution
          let rippleSum = 0;
          let rippleGlow = 0;

          // convert dot position to normalized
          const px = x / Math.max(1, w);
          const py = y / Math.max(1, h);

          for (const r of ripples) {
            const dx = px - r.x;
            const dy = py - r.y;
            const dist = Math.sqrt(dx * dx + dy * dy);

            // traveling sine ring: sin(dist*freq - t*speed)
            const phase = dist * r.freq - (t * r.speed) - (r.t * 1.2);
            const ring = Math.sin(phase);

            // exponential falloff with distance and age
            const atten = Math.exp(-dist * r.decay * 6.0) * Math.exp(-r.t * 0.9);

            rippleSum += ring * atten * r.amp;
            rippleGlow += Math.max(0, ring) * atten; // for subtle accent
          }

          // final radius modulation
          const r = Math.max(
            0.35,
            baseR + wave * waveAmp + rippleSum * 1.15
          );

          // base dot alpha slightly wave-modulated
          const a = base.a * (0.8 + 0.4 * (n));

          // subtle accent when ripple is strong
          const glow = Math.min(0.35, rippleGlow * 0.55);

          // Draw base dot
          ctx.beginPath();
          ctx.fillStyle = `rgba(${base.r},${base.g},${base.b},${a.toFixed(4)})`;
          ctx.arc(x, y, r, 0, Math.PI * 2);
          ctx.fill();

          // Draw accent overlay (only if ripple nearby)
          if (glow > 0.01) {
            ctx.beginPath();
            ctx.fillStyle = `rgba(${accent.r},${accent.g},${accent.b},${(accent.a + glow).toFixed(4)})`;
            ctx.arc(x, y, r * 1.05, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      }

      // prune dead ripples
      for (let i = ripples.length - 1; i >= 0; i--) {
        if (ripples[i].t > 3.5) ripples.splice(i, 1);
      }

      requestAnimationFrame(draw);
    }

    // Handle resize
    const ro = new ResizeObserver(() => {
      resize();
    });
    ro.observe(canvas);

    resize();
    requestAnimationFrame(draw);
  }

  document.addEventListener("DOMContentLoaded", () => {
    applyTheme(getPreferredTheme());
    $("#themeToggle")?.addEventListener("click", toggleTheme);
    $("#themeToggleSm")?.addEventListener("click", toggleTheme);

    $("#year").textContent = String(new Date().getFullYear());

    renderLaunchers();
    initFilters();

    initMobileMenu();
    initAnchorHandling();

    initDotField();
  });

  // Mobile/menu helpers
  function initFilters() {
    const buttons = $$("[data-filter]");

    const setActive = (key) => {
      buttons.forEach(b => {
        const active = b.dataset.filter === key;
        b.classList.toggle("is-active", active);
        b.setAttribute("aria-selected", active ? "true" : "false");
      });

      $$("#launcherGrid .launcher").forEach(card => {
        const status = card.getAttribute("data-status");
        const show = key === "all" || status === key;
        card.classList.toggle("is-hidden", !show);
      });
    };

    buttons.forEach(btn => btn.addEventListener("click", () => setActive(btn.dataset.filter)));
    setActive("all");
  }

  function initMobileMenu() {
    const btn = $("#menuBtn");
    const menu = $("#mobileMenu");
    if (!btn || !menu) return;

    const close = () => {
      menu.classList.remove("is-open");
      btn.setAttribute("aria-expanded", "false");
    };

    btn.addEventListener("click", () => {
      const open = !menu.classList.contains("is-open");
      menu.classList.toggle("is-open", open);
      btn.setAttribute("aria-expanded", open ? "true" : "false");
    });

    menu.addEventListener("click", (e) => {
      const a = e.target.closest("a");
      if (a) close();
    });
  }

  function initAnchorHandling() {
    $$(".nav a, .mobile-menu a, [data-jump]").forEach(el => {
      el.addEventListener("click", (e) => {
        const href = el.getAttribute("href") || el.getAttribute("data-jump") || "";
        if (!href.startsWith("#")) return;
        e.preventDefault();
        scrollToHash(href);
      });
    });
  }

  function getHeaderOffset() {
    const header = $("#header");
    if (!header) return 0;
    const h = header.getBoundingClientRect().height;
    return Math.ceil(h + 10);
  }

  function scrollToHash(hash) {
    const el = document.querySelector(hash);
    if (!el) return;
    const y = window.scrollY + el.getBoundingClientRect().top - getHeaderOffset();
    window.scrollTo({ top: Math.max(0, y), behavior: "smooth" });
  }

  function renderLaunchers() {
    const grid = $("#launcherGrid");
    if (!grid) return;

    const meta = {
      supported: { label: "Supported", badge: "badge-supported" },
      working: { label: "In progress", badge: "badge-working" },
      planned: { label: "Planned", badge: "badge-planned" }
    };

    grid.innerHTML = launchers.map((l) => {
      const m = meta[l.status];
      return `
        <article class="launcher" data-status="${l.status}">
          <div class="launcher-top">
            <div class="launcher-name">${escapeHTML(l.name)}</div>
            <span class="badge ${m.badge} mono">${escapeHTML(m.label)}</span>
          </div>
          <p class="launcher-note">${escapeHTML(l.note)}</p>
        </article>
      `;
    }).join("");

    const counts = {
      supported: launchers.filter(x => x.status === "supported").length,
      working: launchers.filter(x => x.status === "working").length,
      planned: launchers.filter(x => x.status === "planned").length
    };

    $("#countSupported").textContent = String(counts.supported);
    $("#countWorking").textContent = String(counts.working);
    $("#countPlanned").textContent = String(counts.planned);

    $("#btnSupported").textContent = String(counts.supported);
    $("#btnWorking").textContent = String(counts.working);
    $("#btnPlanned").textContent = String(counts.planned);
    $("#btnAll").textContent = String(counts.supported + counts.working + counts.planned);
  }
})();