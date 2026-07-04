<?php
$pageTitle       = 'Team - PolyForge';
$pageDescription = 'Meet the team behind PolyForge and learn how to contribute.';
$pageSlug        = 'team';
require __DIR__ . '/partials/header.php';
?>

  <main id="main">
    <div class="container page-hero">
      <h1>The team</h1>
      <p>The people behind PolyForge - and how you can get involved.</p>
    </div>

    <section class="container section" style="padding-top:0">
      <header class="section-head"><h2>Maintainers</h2></header>
      <div class="team-grid">
        <div class="team-card">
          <img class="team-avatar" src="https://avatars.githubusercontent.com/u/41766284?v=4" alt="Keehan" loading="lazy" />
          <h3>Keehan</h3>
          <div class="team-role">Creator & Lead Maintainer</div>
          <p>Designed and built PolyForge from scratch. Responsible for core architecture, installer logic, launcher adapters, and the ecosystem vision.</p>
          <div class="team-links">
            <a href="https://keehan.co" target="_blank" rel="noopener noreferrer">keehan.co</a>
            <a href="https://github.com/realKeehan" target="_blank" rel="noopener noreferrer">GitHub</a>
          </div>
        </div>
      </div>
    </section>

    <section class="container section">
      <div class="divider"></div>
      <header class="section-head" style="margin-top:24px" id="contributors"><h2>Contributors</h2><p>People who have contributed code, testing, or feedback.</p></header>
      <div class="contributor-grid">
        <div class="contributor-card">
          <div class="contributor-avatar">AB</div>
          <h4>Alex B.</h4>
          <div class="contributor-role">Contributor</div>
          <div class="contributor-links">
            <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>
          </div>
        </div>
        <div class="contributor-card">
          <div class="contributor-avatar">JM</div>
          <h4>Jordan M.</h4>
          <div class="contributor-role">Contributor</div>
          <div class="contributor-links">
            <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>
          </div>
        </div>
        <div class="contributor-card">
          <div class="contributor-avatar">SK</div>
          <h4>Sam K.</h4>
          <div class="contributor-role">Contributor</div>
          <div class="contributor-links">
            <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>
          </div>
        </div>
        <div class="contributor-card">
          <div class="contributor-avatar">You?</div>
          <h4>Your name here</h4>
          <div class="contributor-role">Contributor</div>
          <div class="contributor-links">
            <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">Contribute</a>
          </div>
        </div>
      </div>
    </section>

    <section class="container section">
      <div class="divider"></div>
      <header class="section-head" style="margin-top:24px" id="sponsors"><h2>Sponsors & special thanks</h2><p>Thank you to everyone who supports the project.</p></header>
      <div class="sponsor-grid">
        <div class="sponsor-card">
          <div class="sponsor-avatar">⭐</div>
          <div class="sponsor-info">
            <h4>Taylor R.</h4>
            <p>Early financial supporter who helped fund development time.</p>
            <span class="sponsor-tag">Sponsor</span>
          </div>
        </div>
        <div class="sponsor-card">
          <div class="sponsor-avatar">⭐</div>
          <div class="sponsor-info">
            <h4>Chris D.</h4>
            <p>Ongoing supporter and community advocate.</p>
            <span class="sponsor-tag">Sponsor</span>
          </div>
        </div>
        <div class="sponsor-card">
          <div class="sponsor-avatar">&hearts;</div>
          <div class="sponsor-info">
            <h4>Community supporters</h4>
            <p>Everyone who tested, reported bugs, or supported the project.</p>
            <span class="sponsor-tag">Community</span>
          </div>
        </div>
      </div>

      <div class="hint-box" style="margin-top:24px">
        <div class="hint-box-title">Want to sponsor PolyForge?</div>
        <p>If you'd like to support continued development, consider sponsoring through GitHub Sponsors or reaching out directly. Sponsors may be featured on this page and receive early access to new features.</p>
      </div>
    </section>

    <section class="container section">
      <div class="cta-block">
        <div><h2>Join the project</h2><p>PolyForge is built in the open. Whether you want to contribute code, test launcher adapters, write docs, or just provide feedback - the project benefits from your involvement.</p></div>
        <div class="cta-actions">
          <a class="btn btn-primary btn-sm" href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">View on GitHub</a>
          <a class="btn btn-ghost btn-sm" href="mailto:contact@polyforge.dev">Get in touch</a>
        </div>
      </div>
    </section>
  </main>

<?php require __DIR__ . '/partials/footer.php'; ?>
