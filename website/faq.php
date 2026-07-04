<?php
$pageTitle       = 'FAQ - PolyForge';
$pageDescription = 'Frequently asked questions about PolyForge, the universal modpack installer.';
$pageSlug        = 'faq';
require __DIR__ . '/partials/header.php';
?>

  <main id="main">
    <div class="container page-hero">
      <h1>Frequently asked questions</h1>
      <p>Quick answers to common questions about PolyForge.</p>
    </div>

    <section class="container section" style="padding-top:0;max-width:800px;margin:0 auto">

      <!-- Basics -->
      <div class="faq-category">
        <h3>Basics</h3>

        <div class="faq-item">
          <button class="faq-question" type="button">What is PolyForge?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              PolyForge is a universal modpack installer for Minecraft. It lets you install modpacks across multiple launchers — Vanilla, MultiMC, CurseForge, Modrinth, Prism, and more — using a single consistent workflow.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Is PolyForge free?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Yes. PolyForge is completely free and open-source. The full source code is available on <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">What platforms does PolyForge support?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Currently Windows 10+ (64-bit). Linux and macOS support is planned and on the roadmap. You can track progress on the <a href="./downloads">downloads page</a>.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Do I need to install anything else?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              No. PolyForge is a standalone executable — just download and run it. No additional runtimes, Java installations, or setup wizards are needed for PolyForge itself. Your target launcher should already be installed.
            </div>
          </div>
        </div>
      </div>

      <!-- Launcher -->
      <div class="faq-category">
        <h3>Launcher support</h3>

        <div class="faq-item">
          <button class="faq-question" type="button">Which launchers are supported?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              PolyForge currently supports the Vanilla Launcher, MultiMC, CurseForge, and Modrinth. Several more are in progress including Prism Launcher, ATLauncher, GDLauncher, and others. See the full list on the <a href="./supported">launchers page</a>.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">My launcher isn't listed. Can I still use PolyForge?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Yes! PolyForge includes a "Custom Path" option where you can point it to any launcher's instance directory, and a "Manual Install" option that lets you extract pack files for manual placement.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Will PolyForge modify my existing launcher profiles?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              PolyForge creates new profiles/instances for modpacks. It follows safe overwrite rules — existing data isn't deleted unless an update explicitly replaces files that belong to the managed modpack.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Can I request support for a new launcher?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Absolutely. Open an issue or discussion on <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a> and describe the launcher. Priority is based on demand, stability, and ecosystem fit.
            </div>
          </div>
        </div>
      </div>

      <!-- Security -->
      <div class="faq-category">
        <h3>Security</h3>

        <div class="faq-item">
          <button class="faq-question" type="button">Is PolyForge safe to use?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Yes. PolyForge is open-source and every release is submitted to multiple security analysis platforms including VirusTotal, Hybrid Analysis, ANY.RUN, and more. See the full list on the <a href="./security">security page</a>.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">My antivirus flagged PolyForge. Is it a virus?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              No. PolyForge is compiled with Go, which produces executables that some heuristic-based antivirus engines flag as suspicious. This is a well-known false positive pattern across the Go ecosystem. You can verify results yourself on any of the platforms listed on our <a href="./security">security page</a>, or build from source.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Does PolyForge collect any data?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              No. PolyForge does not collect telemetry, personal data, or usage analytics. Everything stays on your machine. See our <a href="./privacy">privacy policy</a> for details.
            </div>
          </div>
        </div>
      </div>

      <!-- Troubleshooting -->
      <div class="faq-category">
        <h3>Troubleshooting</h3>

        <div class="faq-item">
          <button class="faq-question" type="button">PolyForge can't find my launcher. What do I do?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              If your launcher is installed in a non-standard location (e.g., a portable install or custom directory), use the "Custom Path" option to manually point PolyForge to your launcher's instance folder. Check the status log for diagnostic details.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">The install failed. How do I troubleshoot?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Check the status log in the PolyForge UI — every step is logged with detailed progress and error messages. Common issues include network connectivity, file permissions, or the target launcher not being installed. If you're stuck, open an issue on <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a> with your log output.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Can I uninstall a modpack installed by PolyForge?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              PolyForge includes an uninstall/actions screen for managed installations. You can also manually remove the profile or instance from your launcher's UI.
            </div>
          </div>
        </div>
      </div>

      <!-- Other -->
      <div class="faq-category">
        <h3>Other</h3>

        <div class="faq-item">
          <button class="faq-question" type="button">Is PolyForge affiliated with Mojang or Microsoft?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              No. PolyForge is an independent, community-driven project. It is not affiliated with, endorsed by, or associated with Mojang Studios, Microsoft, or any launcher project.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Can I contribute to PolyForge?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              Yes! PolyForge is open-source and contributions are welcome. Check the <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub repository</a> for the codebase, issues, and contribution guidelines.
            </div>
          </div>
        </div>

        <div class="faq-item">
          <button class="faq-question" type="button">Still have questions?</button>
          <div class="faq-answer">
            <div class="faq-answer-inner">
              If your question isn't answered here, open a discussion or issue on <a href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>. We're happy to help.
            </div>
          </div>
        </div>
      </div>

    </section>

    <!-- CTA -->
    <section class="container section">
      <div class="cta-block">
        <div>
          <h2>Ready to try PolyForge?</h2>
          <p>Download the installer and get started in seconds.</p>
        </div>
        <div class="cta-actions">
          <a class="btn btn-primary btn-sm" href="./downloads">Download now</a>
        </div>
      </div>
    </section>
  </main>

<?php require __DIR__ . '/partials/footer.php'; ?>
