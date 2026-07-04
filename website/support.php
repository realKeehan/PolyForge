<?php
$pageTitle       = 'Support - PolyForge';
$pageDescription = 'Get help with PolyForge. Browse FAQ, documentation, and report issues.';
$pageSlug        = 'support';
require __DIR__ . '/partials/header.php';
?>

  <main id="main">
    <div class="container page-hero">
      <h1>Support</h1>
      <p>Need help? Browse the resources below to find answers or report an issue.</p>
    </div>

    <section class="container section" style="padding-top:0">
      <div class="grid grid-3">
        <a class="card" href="./faq" style="text-decoration:none">
          <div class="card-icon"><svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/><path d="M9.5 9.5a2.5 2.5 0 0 1 4.8.8c0 1.7-2.5 2.3-2.5 2.3M12 17h.01" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>
          <h3>FAQ</h3>
          <p>Answers to the most common questions about PolyForge, modpack installs, and launcher compatibility.</p>
        </a>
        <a class="card" href="https://docs.polyforge.dev" target="_blank" rel="noopener noreferrer" style="text-decoration:none">
          <div class="card-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>
          <h3>Documentation</h3>
          <p>Detailed guides on setup, configuration, launcher adapters, and building from source. Hosted at docs.polyforge.dev.</p>
        </a>
        <a class="card" href="https://github.com/realKeehan/PolyForge/issues" target="_blank" rel="noopener noreferrer" style="text-decoration:none">
          <div class="card-icon"><svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg></div>
          <h3>Report an issue</h3>
          <p>Found a bug or have a feature request? Open an issue on GitHub and the team will triage it.</p>
        </a>
      </div>
    </section>

    <section class="container section">
      <div class="divider"></div>
      <header class="section-head" style="margin-top:24px"><h2>Get in touch</h2><p>For questions not covered by the FAQ or docs.</p></header>
      <div class="grid grid-2">
        <article class="card">
          <h3>GitHub Issues</h3>
          <p>The primary channel for bug reports, feature requests, and technical discussions. Search existing issues before creating a new one.</p>
          <a class="btn btn-ghost btn-sm" href="https://github.com/realKeehan/PolyForge/issues" target="_blank" rel="noopener noreferrer" style="margin-top:12px">Open issues</a>
        </article>
        <article class="card">
          <h3>Email</h3>
          <p>For private inquiries, partnerships, or security concerns that shouldn't be discussed publicly.</p>
          <a class="btn btn-ghost btn-sm" href="mailto:contact@polyforge.dev" style="margin-top:12px">contact@polyforge.dev</a>
        </article>
      </div>
    </section>
  </main>

<?php require __DIR__ . '/partials/footer.php'; ?>
