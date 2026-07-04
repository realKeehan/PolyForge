<?php
$pageTitle       = 'Supported Launchers - PolyForge';
$pageDescription = "See which Minecraft launchers PolyForge supports, which are in progress, and what's planned.";
$pageSlug        = 'supported';
$ogImage         = 'https://polyforge.dev/images/CHOOSE%20LAUNCHER.png';
require __DIR__ . '/partials/header.php';
?>

  <main id="main">
    <div class="container page-hero">
      <h1>Launcher support</h1>
      <p>Filter by support level to see what's ready, what's in progress, and what's coming next.</p>
    </div>

    <section class="container section" style="padding-top:0">
      <!-- Search + Filter Dropdown -->
      <div class="section-head row">
        <div>
          <h2>All launchers</h2>
          <p>Select any launcher to learn more or visit its official page.</p>
        </div>
        <div class="filter-search-box">
          <svg viewBox="0 0 24 24" fill="none" width="18" height="18" style="flex-shrink:0;color:var(--text-muted)"><circle cx="11" cy="11" r="8" stroke="currentColor" stroke-width="1.8"/><path d="m21 21-4.35-4.35" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          <input type="text" id="launcherSearch" placeholder="Search launchers..." autocomplete="off" aria-label="Search launchers" />
          <div class="filter-dropdown-wrapper" id="filterDropdownWrapper">
            <button class="filter-dropdown-toggle" id="filterDropdownToggle" type="button" aria-haspopup="true">
              <span class="filter-pill">All</span>
              <svg viewBox="0 0 24 24" fill="none"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
            <div class="filter-dropdown-menu" id="filterDropdownMenu">
              <button type="button" data-filter="all">All <span class="mono muted" id="btnAll">-</span></button>
              <button type="button" data-filter="supported">Supported <span class="mono muted" id="btnSupported">-</span></button>
              <button type="button" data-filter="working">In progress <span class="mono muted" id="btnWorking">-</span></button>
              <button type="button" data-filter="planned">Planned <span class="mono muted" id="btnPlanned">-</span></button>
              <button type="button" data-filter="unsupported">Unsupported <span class="mono muted" id="btnUnsupported">-</span></button>
            </div>
          </div>
        </div>
      </div>

      <!-- Grid (populated by JS) -->
      <div class="launcher-grid" id="launcherGrid" aria-live="polite"></div>
    </section>

    <!-- How it works -->
    <section class="container section">
      <div class="divider"></div>
      <header class="section-head" style="margin-top:24px">
        <h2>How launcher support works</h2>
        <p>What each status level means for your experience.</p>
      </header>
      <div class="grid grid-3">
        <article class="card">
          <h3 style="color:var(--pf-success)">Supported</h3>
          <p>Fully tested and working. PolyForge can discover, target, and install modpacks to these launchers reliably.</p>
        </article>
        <article class="card">
          <h3 style="color:var(--pf-purple)">In progress</h3>
          <p>Adapter exists but is still being refined. Basic installs may work, but edge cases are being resolved.</p>
        </article>
        <article class="card">
          <h3 style="color:var(--pf-danger)">Planned</h3>
          <p>On the roadmap. These launchers will get adapters based on demand, stability, and ecosystem fit.</p>
        </article>
      </div>
    </section>

    <!-- CTA -->
    <section class="container section">
      <div class="cta-block">
        <div>
          <h2>Don't see your launcher?</h2>
          <p>PolyForge supports Custom Path and Manual Install modes for any launcher not listed here. You can also request support via GitHub.</p>
        </div>
        <div class="cta-actions">
          <a class="btn btn-primary btn-sm" href="./downloads">Download now</a>
          <a class="btn btn-ghost btn-sm" href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">Request on GitHub</a>
        </div>
      </div>
    </section>
  </main>

<?php require __DIR__ . '/partials/footer.php'; ?>
