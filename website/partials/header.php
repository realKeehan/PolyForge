<?php
/**
 * Shared site header.
 *
 * Set these variables before requiring this file:
 *   $pageTitle       (string)  Document title.
 *   $pageDescription (string)  Meta description.
 *   $pageSlug        (string)  Page identifier without extension ('index', 'downloads', ...).
 *                              Used for the canonical URL and active nav highlighting.
 *   $ogImage         (string)  Optional absolute URL for the social preview image.
 *   $hideDownloadBtn (bool)    Optional. Hides the header Download button (e.g. on the downloads page).
 *   $noIndex         (bool)    Optional. Emits robots noindex and skips the canonical tag (error pages).
 */

declare(strict_types=1);

$siteUrl = 'https://polyforge.dev';

$pageTitle       = $pageTitle       ?? "PolyForge - Keehan's Universal Modpack Installer";
$pageDescription = $pageDescription ?? "PolyForge is Keehan's Universal Modpack Installer for Minecraft launchers. One workflow. Many ecosystems.";
$pageSlug        = $pageSlug        ?? 'index';
$ogImage         = $ogImage         ?? $siteUrl . '/images/CHOOSE%20MODPACK.png';
$hideDownloadBtn = $hideDownloadBtn ?? false;
$noIndex         = $noIndex         ?? false;

$canonical = $siteUrl . ($pageSlug === 'index' ? '/' : '/' . $pageSlug);

$navLinks = [
  'index'     => 'Home',
  'supported' => 'Launchers',
  'downloads' => 'Downloads',
  'security'  => 'Security',
  'faq'       => 'FAQ',
];

function pf_href(string $slug): string
{
  return $slug === 'index' ? './' : './' . $slug;
}

function pf_nav(array $links, string $activeSlug): void
{
  foreach ($links as $slug => $label) {
    $active = $slug === $activeSlug ? ' class="is-active" aria-current="page"' : '';
    echo '        <a href="' . pf_href($slug) . '"' . $active . '>' . htmlspecialchars($label) . "</a>\n";
  }
}
?>
<!doctype html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title><?= htmlspecialchars($pageTitle) ?></title>
  <meta name="description" content="<?= htmlspecialchars($pageDescription) ?>" />
<?php if ($noIndex): ?>
  <meta name="robots" content="noindex" />
<?php else: ?>
  <link rel="canonical" href="<?= htmlspecialchars($canonical) ?>" />
<?php endif; ?>

  <!-- Open Graph / Social Link Preview -->
  <meta property="og:type" content="website" />
  <meta property="og:title" content="<?= htmlspecialchars($pageTitle) ?>" />
  <meta property="og:description" content="<?= htmlspecialchars($pageDescription) ?>" />
  <meta property="og:url" content="<?= htmlspecialchars($canonical) ?>" />
  <meta property="og:image" content="<?= htmlspecialchars($ogImage) ?>" />
  <meta property="og:site_name" content="PolyForge" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="<?= htmlspecialchars($pageTitle) ?>" />
  <meta name="twitter:description" content="<?= htmlspecialchars($pageDescription) ?>" />
  <meta name="twitter:image" content="<?= htmlspecialchars($ogImage) ?>" />

  <meta name="theme-color" content="#0c0914" media="(prefers-color-scheme: dark)" />
  <meta name="theme-color" content="#f5f0ff" media="(prefers-color-scheme: light)" />

  <link rel="icon" href="./favicon.ico" />
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
  <link rel="stylesheet" href="./styles.css" />

  <!-- Apply the saved theme before first paint to avoid a flash of the wrong theme -->
  <script>
    (function () {
      try {
        var t = localStorage.getItem("pf-theme");
        if (t !== "light" && t !== "dark") {
          t = window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
        }
        document.documentElement.dataset.theme = t;
      } catch (e) { /* default stays dark */ }
    })();
  </script>
</head>
<body>
  <!-- Animated dot background -->
  <div class="bg" aria-hidden="true">
    <canvas class="dot-canvas" id="dotCanvas"></canvas>
  </div>

  <!-- ─── Header ──────────────────────────────── -->
  <header class="site-header" id="header">
    <div class="container header-row">
      <a class="brand" href="./" aria-label="PolyForge home">
        <span class="brand-dot" aria-hidden="true"></span>
        <span class="brand-name">PolyForge</span>
      </a>

      <nav class="nav" aria-label="Primary">
<?php pf_nav($navLinks, $pageSlug); ?>
      </nav>

      <div class="header-actions">
<?php if (!$hideDownloadBtn): ?>
        <a class="btn btn-primary btn-sm" href="./downloads">Download</a>
<?php endif; ?>
        <button class="icon-toggle" id="themeToggle" type="button" aria-label="Toggle theme">
          <span class="icon-toggle__sun" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none"><path d="M12 18a6 6 0 1 0 0-12 6 6 0 0 0 0 12Z" stroke="currentColor" stroke-width="1.8"/><path d="M12 2v2M12 20v2M4 12H2M22 12h-2M5.2 5.2 3.8 3.8M20.2 20.2l-1.4-1.4M18.8 5.2l1.4-1.4M3.8 20.2l1.4-1.4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          </span>
          <span class="icon-toggle__moon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none"><path d="M21 14.2A7.6 7.6 0 0 1 9.8 3a6.9 6.9 0 1 0 11.2 11.2Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
          </span>
        </button>
        <button class="iconbtn" id="menuBtn" type="button" aria-label="Open menu" aria-expanded="false">
          <span></span><span></span>
        </button>
      </div>
    </div>

    <!-- Mobile menu -->
    <div class="container mobile-menu" id="mobileMenu">
<?php pf_nav($navLinks, $pageSlug); ?>
      <div class="mobile-row">
        <button class="icon-toggle" id="themeToggleSm" type="button" aria-label="Toggle theme">
          <span class="icon-toggle__sun" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none"><path d="M12 18a6 6 0 1 0 0-12 6 6 0 0 0 0 12Z" stroke="currentColor" stroke-width="1.8"/><path d="M12 2v2M12 20v2M4 12H2M22 12h-2M5.2 5.2 3.8 3.8M20.2 20.2l-1.4-1.4M18.8 5.2l1.4-1.4M3.8 20.2l1.4-1.4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          </span>
          <span class="icon-toggle__moon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none"><path d="M21 14.2A7.6 7.6 0 0 1 9.8 3a6.9 6.9 0 1 0 11.2 11.2Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
          </span>
        </button>
<?php if (!$hideDownloadBtn): ?>
        <a class="btn btn-primary" href="./downloads">Download</a>
<?php endif; ?>
      </div>
    </div>
  </header>
