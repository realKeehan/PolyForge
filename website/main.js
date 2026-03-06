// ══════════════════════════════════════════════════
// PolyForge - Shared JS (all pages)
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
    { name:"Vanilla Launcher", status:"supported", note:"The official Mojang/Microsoft launcher - the baseline for standard Minecraft installs.", url:"https://www.minecraft.net/en-us/download" },
    { name:"MultiMC", status:"supported", note:"A lightweight multi-instance launcher focused on custom modded setups and clean instance management.", url:"https://multimc.org/" },
    { name:"CurseForge", status:"supported", note:"A popular ecosystem for modpacks with built-in browsing, installs, and updates through the CurseForge platform.", url:"https://www.curseforge.com/" },
    { name:"Modrinth (Theseus)", status:"supported", note:"Modrinth's launcher/profile system - modern pack distribution with a fast-growing mod ecosystem.", url:"https://modrinth.com/" },
    { name:"Custom Path", status:"supported", note:"For nonstandard installs, portable environments, or advanced setups where you want full control of location.", url:null },
    { name:"Manual Install", status:"supported", note:"For users who prefer to manage placement themselves or need a pack output for custom workflows.", url:null },

    // In progress
    { name:"Prism Launcher", status:"working", note:"A modern MultiMC fork with broader platform support, active development, and power-user features.", url:"https://prismlauncher.org/" },
    { name:"ATLauncher", status:"working", note:"A long-running launcher built around curated packs and easy modded profiles.", url:"https://atlauncher.com/" },
    { name:"GDLauncher", status:"working", note:"A sleek launcher that emphasizes a friendly UI and integrated pack browsing/management.", url:"https://gdlauncher.com/" },
    { name:"Technic", status:"working", note:"One of the classic launcher platforms - known for legacy packs and older modpack history.", url:"https://www.technicpack.net/" },
    { name:"PolyMC", status:"working", note:"A Prism/MultiMC-family launcher - similar instance philosophy with community-driven tooling.", url:"https://polymc.org/" },
    { name:"Feather", status:"working", note:"A performance-focused launcher often used for competitive play and client-side enhancements.", url:"https://feathermc.com/" },
    { name:"BakaXL", status:"working", note:"A launcher favored by some modded communities, especially in regions where it's widely adopted.", url:"https://www.bakaxl.com/" },

    // Planned
    { name:"Polymerium", status:"planned", note:"A modern Minecraft launcher for Windows with a clean UI and modpack management capabilities.", url:"https://github.com/d3ara1n/Polymerium" },
    { name:"X Minecraft Launcher", status:"planned", note:"An open-source Minecraft launcher supporting multiple accounts, modpacks, and resource management.", url:"https://github.com/Voxelum/x-minecraft-launcher" },
    { name:"SK Launcher", status:"planned", note:"An all-in-one Minecraft hub with built-in modloaders, modpack support, and skin management.", url:"https://skmedix.pl/" },
    { name:"Freesm Launcher", status:"planned", note:"A Prism-based launcher that removes offline account restrictions and adds custom auth server support.", url:"https://freesmlauncher.org/" },
    { name:"ElyPrism", status:"planned", note:"A Prism Launcher fork with Ely.by authentication integration for alternative account systems.", url:"https://elyprismlauncher.github.io/" },
    { name:"ShatteredPrism", status:"planned", note:"A community-maintained Prism Launcher fork focused on extended features and flexibility.", url:"https://github.com/Noctilune/ShatteredPrism" },
    { name:"QWERTZ Launcher", status:"planned", note:"A launcher from the QWERTZ project ecosystem with streamlined Minecraft instance management.", url:"https://qwertz.app/projects/" },
    { name:"Fjord Launcher", status:"planned", note:"An Unmojang project - a Prism-family launcher with its own community-driven direction.", url:"https://github.com/unmojang/FjordLauncher" },
    { name:"HMCL", status:"planned", note:"A cross-platform Minecraft launcher popular in the Chinese community, supporting multiple auth and mod sources.", url:"https://hmcl.huangyuhui.net/" },
    { name:"UltimMC", status:"planned", note:"A MultiMC fork focused on offline play support and community-driven development.", url:"https://github.com/UltimMC/Launcher" },

    // Unsupported
    { name:"TLauncher", status:"unsupported", note:"Not supported due to documented privacy and security concerns raised by the Minecraft community.", url:null },

    // Catch-all
    { name:"Additional ecosystems", status:"planned", note:"More launcher adapters as the ecosystem evolves - prioritized by demand and stability.", url:null },
  ];

  const meta = {
    supported:   { label:"Supported",    badge:"badge-supported" },
    working:     { label:"In progress",  badge:"badge-working" },
    planned:     { label:"Planned",      badge:"badge-planned" },
    unsupported: { label:"Unsupported",  badge:"badge-unsupported" },
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
    const raw = location.pathname.split("/").pop() || "";
    const current = raw.replace(/\.html$/, "") || "index";
    $$(".nav a, .mobile-menu a").forEach(a => {
      const href = (a.getAttribute("href") || "").replace(/^\.\//, "").replace(/\.html$/, "") || "index";
      if (href === current || (current === "index" && (href === "" || href === "index"))) {
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

  // ── Download/user counters ─────────────────────
  async function fetchStats() {
    // Primary: site-hosted stats.json
    try {
      const res = await fetch("./stats.json?" + Date.now());
      if (res.ok) {
        const data = await res.json();
        return { downloads: data.downloads || 0, users: data.users || 0 };
      }
    } catch {}

    // Fallback: GitHub API release download counts
    try {
      const res = await fetch("https://api.github.com/repos/realKeehan/PolyForge/releases");
      if (res.ok) {
        const releases = await res.json();
        let total = 0;
        releases.forEach(r => {
          (r.assets || []).forEach(a => { total += a.download_count || 0; });
        });
        return { downloads: total, users: 0 };
      }
    } catch {}

    return { downloads: 0, users: 0 };
  }

  function renderStats() {
    fetchStats().then(stats => {
      const dlEl = $("#statDownloads");
      const usEl = $("#statUsers");
      if (dlEl) dlEl.textContent = stats.downloads.toLocaleString();
      if (usEl) usEl.textContent = stats.users.toLocaleString();
    });
  }

  // ── Compute launcher counts (always available) ──
  function computeLauncherCounts() {
    const counts = { supported:0, working:0, planned:0, unsupported:0 };
    launchers.forEach(l => { if (counts[l.status] !== undefined) counts[l.status]++; });
    return counts;
  }

  function renderLauncherCounts() {
    const counts = computeLauncherCounts();
    const set = (id, val) => { const el = $(id); if (el) el.textContent = String(val); };
    set("#countSupported", counts.supported);
    set("#countWorking", counts.working);
    set("#countPlanned", counts.planned);
    set("#btnSupported", counts.supported);
    set("#btnWorking", counts.working);
    set("#btnPlanned", counts.planned);
    set("#btnUnsupported", counts.unsupported);
    set("#btnAll", counts.supported + counts.working + counts.planned + counts.unsupported);
  }

  // ── Render launchers (used by supported.html & index.html) ──
  function renderLaunchers(gridId, opts = {}) {
    const grid = $(gridId || "#launcherGrid");
    if (!grid) return;

    let list = opts.subset ? launchers.filter(l => opts.subset.includes(l.status)) : launchers;

    // Search filter
    const searchInput = $("#launcherSearch");
    if (searchInput) {
      const query = searchInput.value.trim().toLowerCase();
      if (query) {
        list = list.filter(l => l.name.toLowerCase().includes(query) || l.note.toLowerCase().includes(query));
      }
    }

    grid.innerHTML = list.map(l => {
      const m = meta[l.status];
      if (!m) return "";
      const tag = l.url ? "a" : "article";
      const hrefAttr = l.url ? ` href="${esc(l.url)}" target="_blank" rel="noopener noreferrer"` : "";

      return `
        <${tag} class="launcher" data-status="${l.status}"${hrefAttr}>
          <div class="launcher-top">
            <div class="launcher-info">
              <div class="launcher-name">${esc(l.name)}</div>
            </div>
            <div class="launcher-actions">
              <span class="badge ${m.badge} mono">${esc(m.label)}</span>
            </div>
          </div>
          <p class="launcher-note">${esc(l.note)}</p>
          ${l.url ? `<span class="launcher-link">Visit site &rarr;</span>` : ""}
        </${tag}>
      `;
    }).join("");

    renderLauncherCounts();
  }

  // ── Filter dropdown ────────────────────────────
  function initFilterDropdown() {
    const wrapper = $("#filterDropdownWrapper");
    const toggle = $("#filterDropdownToggle");
    const menu = $("#filterDropdownMenu");
    const searchInput = $("#launcherSearch");
    if (!wrapper || !toggle || !menu) return;

    let currentFilter = "all";

    function setFilter(key) {
      currentFilter = key;
      // Update toggle pill
      const m = key === "all" ? { label:"All", badge:"" } : (meta[key] || { label:key, badge:"" });
      const pill = toggle.querySelector(".filter-pill");
      if (pill) {
        pill.className = "filter-pill" + (key !== "all" ? " filter-pill--" + key : "");
        pill.textContent = m.label;
      }

      // Apply filter
      $$("#launcherGrid .launcher").forEach(card => {
        const status = card.getAttribute("data-status");
        card.classList.toggle("is-hidden", key !== "all" && status !== key);
      });

      menu.classList.remove("is-open");
    }

    toggle.addEventListener("click", (e) => {
      e.stopPropagation();
      menu.classList.toggle("is-open");
    });

    menu.querySelectorAll("[data-filter]").forEach(btn => {
      btn.addEventListener("click", () => setFilter(btn.dataset.filter));
    });

    document.addEventListener("click", (e) => {
      if (!wrapper.contains(e.target)) menu.classList.remove("is-open");
    });

    // Search functionality
    if (searchInput) {
      searchInput.addEventListener("input", () => {
        renderLaunchers("#launcherGrid");
        // Re-apply current filter after re-render
        if (currentFilter !== "all") {
          $$("#launcherGrid .launcher").forEach(card => {
            const status = card.getAttribute("data-status");
            card.classList.toggle("is-hidden", status !== currentFilter);
          });
        }
      });
    }

    setFilter("all");
  }

  // ── Legacy filters (fallback for pages without dropdown) ──
  function initFilters() {
    const buttons = $$("[data-filter]");
    if (!buttons.length) return;
    // Skip if dropdown is present
    if ($("#filterDropdownWrapper")) return;

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

  // ── Carousel (Coming Soon) ─────────────────────
  function initCarousel() {
    const carousel = $("#comingSoonCarousel");
    if (!carousel) return;

    const slides = $$(".carousel-slide", carousel);
    const dots = $$(".carousel-dot", carousel);
    const timerBar = $(".carousel-timer-bar", carousel);
    if (slides.length === 0) return;

    let current = 0;
    let interval = null;
    let paused = false;
    const DURATION = 5000;

    function showSlide(index) {
      slides.forEach((s, i) => {
        s.classList.toggle("is-active", i === index);
      });
      dots.forEach((d, i) => {
        d.classList.toggle("is-active", i === index);
      });
      current = index;
      resetTimer();
    }

    function nextSlide() {
      showSlide((current + 1) % slides.length);
    }

    function resetTimer() {
      if (timerBar) {
        timerBar.style.transition = "none";
        timerBar.style.width = "0%";
        // Force reflow
        void timerBar.offsetWidth;
        timerBar.style.transition = `width ${DURATION}ms linear`;
        timerBar.style.width = "100%";
      }
      clearInterval(interval);
      if (!paused) {
        interval = setInterval(nextSlide, DURATION);
      }
    }

    dots.forEach((dot, i) => {
      dot.addEventListener("click", () => showSlide(i));
    });

    carousel.addEventListener("mouseenter", () => {
      paused = true;
      clearInterval(interval);
      if (timerBar) {
        const w = timerBar.getBoundingClientRect().width;
        const pw = timerBar.parentElement.getBoundingClientRect().width;
        timerBar.style.transition = "none";
        timerBar.style.width = (pw > 0 ? (w / pw * 100) : 0) + "%";
      }
    });

    carousel.addEventListener("mouseleave", () => {
      paused = false;
      resetTimer();
    });

    showSlide(0);
  }

  // ── FAQ accordion ──────────────────────────────
  function initFaqAccordion() {
    $$(".faq-question").forEach(btn => {
      btn.addEventListener("click", () => {
        const item = btn.closest(".faq-item");
        const answer = item.querySelector(".faq-answer");
        const isOpen = item.classList.contains("is-open");

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

  // ── Security provider dropdown accordion ───────
  function initSecurityAccordion() {
    $$(".provider-accordion-toggle").forEach(btn => {
      btn.addEventListener("click", () => {
        const item = btn.closest(".provider-accordion");
        const body = item.querySelector(".provider-accordion-body");
        const isOpen = item.classList.contains("is-open");

        // Close all
        $$(".provider-accordion.is-open").forEach(open => {
          open.classList.remove("is-open");
          open.querySelector(".provider-accordion-body").style.maxHeight = "0";
        });

        if (!isOpen) {
          item.classList.add("is-open");
          body.style.maxHeight = body.scrollHeight + "px";
        }
      });
    });
  }

  // ── Download page: file type dropdowns ─────────
  function initDownloadDropdowns() {
    $$(".dl-more-toggle").forEach(btn => {
      btn.addEventListener("click", () => {
        const extra = btn.closest(".platform-card").querySelector(".dl-extra");
        if (!extra) return;
        const open = !extra.classList.contains("is-open");
        extra.classList.toggle("is-open", open);
        btn.textContent = open ? "Show fewer options" : "Show all formats";
      });
    });
  }

  // ── Dot field canvas (performance-optimized) ───
  function initDotField() {
    const canvas = document.getElementById("dotCanvas");
    if (!(canvas instanceof HTMLCanvasElement)) return;
    const ctx = canvas.getContext("2d", { alpha: true });
    if (!ctx) return;
    if (window.matchMedia?.("(prefers-reduced-motion: reduce)").matches) return;

    let w = 0, h = 0;
    const dpr = Math.min(1.5, window.devicePixelRatio || 1); // Cap DPR for performance

    let spacing = 22, baseR = 1.0;
    const waveAmp = 0.6, waveSpeed = 0.25, noiseScale = 0.01, noiseOctaves = 2;
    const ripples = [];
    const maxRipples = 6; // Reduced for performance

    // Throttle resize
    let resizeTimer = null;
    function resize() {
      const rect = canvas.getBoundingClientRect();
      w = Math.floor(rect.width);
      h = Math.floor(rect.height);
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      spacing = Math.max(18, Math.min(28, Math.round(Math.min(w, h) / 40)));
      baseR = 0.9;
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
        ? { r:80, g:50, b:120, a:0.22 }
        : { r:210, g:190, b:255, a:0.28 };
    }

    function themeAccentColor() {
      return (document.documentElement.dataset.theme || "dark") === "light"
        ? { r:143, g:0, b:255, a:0.12 }
        : { r:143, g:0, b:255, a:0.18 };
    }

    function addRipple(x, y) {
      ripples.unshift({ x, y, t:0, amp:1.0, freq:8, speed:0.7, decay:2.0 });
      if (ripples.length > maxRipples) ripples.pop();
    }

    // Throttle pointer events
    let lastPointer = 0;
    window.addEventListener("pointermove", e => {
      const now = performance.now();
      if (now - lastPointer < 100) return;
      lastPointer = now;
      if (Math.random() < 0.2) addRipple(e.clientX / window.innerWidth, e.clientY / window.innerHeight);
    }, { passive: true });

    window.addEventListener("pointerdown", e => {
      const x = e.clientX / window.innerWidth, y = e.clientY / window.innerHeight;
      addRipple(x, y);
    }, { passive: true });

    const start = performance.now();
    let lastFrame = 0;

    function draw(now) {
      // Frame rate limiter: target ~30fps for performance
      if (now - lastFrame < 32) {
        requestAnimationFrame(draw);
        return;
      }
      lastFrame = now;

      const t = (now - start) / 1000;
      ctx.clearRect(0, 0, w, h);
      const base = themeDotColor();
      const accent = themeAccentColor();
      for (const r of ripples) r.t += 0.032;

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

          const rad = Math.max(0.3, baseR + wave * waveAmp + rippleSum * 0.8);
          const a = base.a * (0.8 + 0.3 * n);

          ctx.beginPath();
          ctx.fillStyle = `rgba(${base.r},${base.g},${base.b},${a.toFixed(3)})`;
          ctx.arc(x, y, rad, 0, Math.PI * 2);
          ctx.fill();

          if (rippleGlow > 0.02) {
            const glow = Math.min(0.25, rippleGlow * 0.4);
            ctx.beginPath();
            ctx.fillStyle = `rgba(${accent.r},${accent.g},${accent.b},${(accent.a + glow).toFixed(3)})`;
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

    let roTimer = null;
    new ResizeObserver(() => {
      clearTimeout(roTimer);
      roTimer = setTimeout(resize, 150);
    }).observe(canvas);
    resize();
    requestAnimationFrame(draw);
  }

  // ── Info bubble toggles (download hints) ───────
  function initInfoBubbles() {
    $$(".info-bubble-toggle").forEach(btn => {
      btn.addEventListener("click", (e) => {
        e.stopPropagation();
        const bubble = btn.closest(".info-bubble-wrapper").querySelector(".info-bubble-content");
        if (!bubble) return;
        const isOpen = bubble.classList.contains("is-open");
        // Close all open bubbles first
        $$(".info-bubble-content.is-open").forEach(b => b.classList.remove("is-open"));
        if (!isOpen) bubble.classList.add("is-open");
      });
    });
    document.addEventListener("click", () => {
      $$(".info-bubble-content.is-open").forEach(b => b.classList.remove("is-open"));
    });
  }

  // ── Eye-toggle for hash sections ───────────────
  function initEyeToggles() {
    $$(".hash-eye-toggle").forEach(btn => {
      btn.addEventListener("click", () => {
        const hashBlock = btn.closest(".dl-hash-wrapper").querySelector(".dl-hash");
        if (!hashBlock) return;
        const isOpen = hashBlock.classList.contains("is-visible");
        hashBlock.classList.toggle("is-visible", !isOpen);
        btn.classList.toggle("is-active", !isOpen);
        btn.setAttribute("aria-expanded", String(!isOpen));
      });
    });
  }

  // ── Page transition (morph) ────────────────────
  function initPageTransition() {
    // Morph-in on load
    document.body.classList.add("page-loaded");

    // Intercept internal navigation for smooth morph-out
    $$("a[href]").forEach(a => {
      const href = a.getAttribute("href") || "";
      // Only internal links (same-origin, not anchors, not external)
      if (href.startsWith("http") || href.startsWith("#") || href.startsWith("mailto:") || a.target === "_blank") return;
      a.addEventListener("click", (e) => {
        e.preventDefault();
        document.body.classList.add("page-leaving");
        setTimeout(() => { window.location.href = href; }, 220);
      });
    });
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

    // Launcher counts (always, even without grid)
    renderLauncherCounts();

    // Launchers
    renderLaunchers("#launcherGrid");
    initFilterDropdown();
    initFilters();

    // Navigation
    highlightActiveNav();
    initMobileMenu();
    initAnchorHandling();

    // FAQ
    initFaqAccordion();

    // Security
    initSecurityAccordion();

    // Download dropdowns + info bubbles + eye toggles
    initDownloadDropdowns();
    initInfoBubbles();
    initEyeToggles();

    // Carousel
    initCarousel();

    // Stats
    renderStats();

    // Page transition
    initPageTransition();

    // Dot field
    initDotField();
  });
})();
