<?php
/**
 * Shared site footer. Requires no variables.
 * Closes the document — include after the page's </main>.
 */
?>
  <!-- ─── Footer ──────────────────────────────── -->
  <footer class="container footer">
    <div class="footer-inner">
      <div class="footer-brand">
        <a class="brand" href="./">
          <span class="brand-dot" aria-hidden="true"></span>
          <span class="brand-name">PolyForge</span>
        </a>
        <p>Keehan's Universal Modpack Installer for Minecraft launchers. One workflow, many ecosystems.</p>
      </div>
      <div class="footer-cols">
        <div class="footer-col">
          <h4>Product</h4>
          <a href="./downloads">Downloads</a>
          <a href="./supported">Launchers</a>
          <a href="./security">Security</a>
          <a href="./faq">FAQ</a>
        </div>
        <div class="footer-col">
          <h4>Resources</h4>
          <a href="./support">Support</a>
          <a href="./team">Team</a>
          <a href="https://docs.polyforge.dev" target="_blank" rel="noopener noreferrer">Docs</a>
          <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>
        </div>
        <div class="footer-col">
          <h4>Legal</h4>
          <a href="./privacy">Privacy</a>
          <a href="./terms">Terms</a>
        </div>
        <div class="footer-col">
          <h4>Contact</h4>
          <a href="mailto:contact@polyforge.dev">contact@polyforge.dev</a>
        </div>
      </div>
    </div>
    <div class="footer-bottom">
      <div class="mono muted">&copy; <?= date('Y') ?> PolyForge</div>
      <span class="footer-creator">Created by <a href="https://keehan.co" target="_blank" rel="noopener noreferrer">Keehan</a></span>
      <a class="footer-status" href="https://status.polyforge.dev" target="_blank" rel="noopener noreferrer">
        <span class="footer-status-dot"></span>
        All systems operational
      </a>
    </div>
  </footer>

  <script src="./main.js" defer></script>
</body>
</html>
