<?php
$pageTitle       = 'Security - PolyForge';
$pageDescription = 'PolyForge security scans, SHA256 hashes, and analysis provider results.';
$pageSlug        = 'security';

// Live analysis results recorded through the admin panel (automated VirusTotal
// scans + manually added provider reports), newest check first.
require __DIR__ . '/api/security-lib.php';
$pfSecurityRows = securityEntries(securityLoad(__DIR__ . '/api/security-data.json'));

/** Verdict → [css class, label] for the results table. */
function pf_sec_verdict(string $verdict): array
{
    switch ($verdict) {
        case 'clean':         return ['scan-result--clean', 'Clean'];
        case 'flagged':       return ['scan-result--flagged', 'Flagged'];
        case 'malicious':     return ['scan-result--flagged', 'Malicious'];
        case 'suspicious':    return ['scan-result--flagged', 'Suspicious'];
        case 'informational': return ['scan-result--pending', 'Informational'];
        case 'pending':       return ['scan-result--pending', 'Pending'];
        default:              return ['scan-result--pending', 'Unavailable'];
    }
}

require __DIR__ . '/partials/header.php';
?>

  <main id="main">
    <div class="container page-hero">
      <h1>Security & transparency</h1>
      <p>PolyForge is scanned across multiple trusted malware analysis platforms. We believe in full transparency about our security posture.</p>
    </div>

    <!-- Security overview -->
    <section class="container section" style="padding-top:0">
      <div class="grid grid-3">
        <article class="card">
          <div class="card-icon">
            <svg viewBox="0 0 24 24" fill="none"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
          </div>
          <h3>Multi-scanner verified</h3>
          <p>Every release is submitted to multiple independent security analysis platforms before distribution.</p>
        </article>
        <article class="card">
          <div class="card-icon">
            <svg viewBox="0 0 24 24" fill="none"><rect x="3" y="3" width="18" height="18" rx="4" stroke="currentColor" stroke-width="1.8"/><path d="m9 12 2 2 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </div>
          <h3>Open source</h3>
          <p>The full source code is available on GitHub for anyone to audit, review, and build from source.</p>
        </article>
        <article class="card">
          <div class="card-icon">
            <svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          </div>
          <h3>About false positives</h3>
          <p>Go-compiled binaries are sometimes flagged by heuristic engines. This is a known false-positive pattern across the Go ecosystem, not a sign of malware.</p>
        </article>
      </div>
    </section>

    <!-- Latest analysis results (recorded via the admin panel) -->
    <?php if ($pfSecurityRows !== []): ?>
    <section class="container section" style="padding-top:12px">
      <header class="section-head"><h2>Latest scan results</h2><p>Recorded analysis results across providers, most recently checked first.</p></header>
      <div class="card" style="overflow-x:auto">
        <table class="scan-table">
          <thead><tr><th>Checked</th><th>Provider</th><th>File</th><th>Result</th><th>Detail</th><th>Report</th></tr></thead>
          <tbody>
          <?php foreach ($pfSecurityRows as $row):
              [$vClass, $vLabel] = pf_sec_verdict((string) $row['verdict']);
              $checked = $row['lastChecked'] !== '' ? substr($row['lastChecked'], 0, 10) : '-';
              $links = [];
              if ($row['url'] !== '') {
                  $links[] = '<a href="' . htmlspecialchars($row['url'], ENT_QUOTES) . '" target="_blank" rel="noopener noreferrer">View report &rarr;</a>';
              }
              if ($row['reportFile'] !== '') {
                  $links[] = '<a href="/' . htmlspecialchars($row['reportFile'], ENT_QUOTES) . '" target="_blank" rel="noopener noreferrer">Evidence</a>';
              }
          ?>
            <tr>
              <td class="mono muted"><?= htmlspecialchars($checked, ENT_QUOTES) ?></td>
              <td><?= htmlspecialchars((string) $row['provider'], ENT_QUOTES) ?></td>
              <td class="mono"><?= htmlspecialchars((string) ($row['file'] !== '' ? $row['file'] : '-'), ENT_QUOTES) ?></td>
              <td class="<?= $vClass ?>"><?= $vLabel ?></td>
              <td><?= htmlspecialchars((string) $row['detail'], ENT_QUOTES) ?></td>
              <td><?= $links !== [] ? implode(' &middot; ', $links) : '<span class="muted">-</span>' ?></td>
            </tr>
          <?php endforeach; ?>
          </tbody>
        </table>
      </div>
    </section>
    <?php endif; ?>

    <!-- Scan providers -->
    <section class="container section" style="padding-top:12px">
      <header class="section-head"><h2>Analysis platforms</h2><p>PolyForge is submitted to these security analysis providers. Click a provider to expand scan details and verification info.</p></header>

      <!-- VirusTotal -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg></div>
          <div class="provider-accordion-info"><h3>VirusTotal</h3><p>Multi-engine antivirus scanning - 70+ engines</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>VirusTotal scans each binary against 70+ antivirus engines. Results are published with each release.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Engines</th><th>Result</th><th>Hash (SHA256)</th></tr></thead>
              <tbody>
                <tr><td>PolyForge-Setup.exe</td><td>70+</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr>
              </tbody>
            </table>
            <p style="margin-top:12px"><a href="https://www.virustotal.com/" target="_blank" rel="noopener noreferrer">Visit VirusTotal &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- Hybrid Analysis -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/><path d="M12 8v4l3 3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>
          <div class="provider-accordion-info"><h3>Hybrid Analysis</h3><p>Behavioral sandbox analysis by CrowdStrike</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>Hybrid Analysis runs the binary in a sandbox environment to detect behavioral indicators of malware.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Verdict</th><th>Hash (SHA256)</th></tr></thead>
              <tbody>
                <tr><td>PolyForge-Setup.exe</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr>
              </tbody>
            </table>
            <p style="margin-top:12px"><a href="https://www.hybrid-analysis.com/" target="_blank" rel="noopener noreferrer">Visit Hybrid Analysis &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- ANY.RUN -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><rect x="3" y="3" width="18" height="18" rx="4" stroke="currentColor" stroke-width="1.8"/><path d="M8 12h8M12 8v8" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg></div>
          <div class="provider-accordion-info"><h3>ANY.RUN</h3><p>Interactive sandbox for malware analysis</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>ANY.RUN provides interactive malware sandboxing with visual reports of process execution.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Verdict</th><th>Hash (SHA256)</th></tr></thead>
              <tbody>
                <tr><td>PolyForge-Setup.exe</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr>
              </tbody>
            </table>
            <p style="margin-top:12px"><a href="https://any.run/" target="_blank" rel="noopener noreferrer">Visit ANY.RUN &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- Joe Sandbox -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M4 4h16v16H4z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="m9 12 2 2 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>
          <div class="provider-accordion-info"><h3>Joe Sandbox</h3><p>Deep malware analysis with detailed reports</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>Joe Sandbox performs automated dynamic analysis with behavioral and network analysis reports.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Verdict</th><th>Hash (SHA256)</th></tr></thead>
              <tbody>
                <tr><td>PolyForge-Setup.exe</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr>
              </tbody>
            </table>
            <p style="margin-top:12px"><a href="https://www.joesandbox.com/" target="_blank" rel="noopener noreferrer">Visit Joe Sandbox &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- MetaDefender -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="m9 12 2 2 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>
          <div class="provider-accordion-info"><h3>MetaDefender (OPSWAT)</h3><p>Multi-scanning with 30+ engines</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>MetaDefender by OPSWAT scans files against 30+ antivirus engines with content disarm and reconstruction capabilities.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Engines</th><th>Result</th><th>Hash (SHA256)</th></tr></thead>
              <tbody>
                <tr><td>PolyForge-Setup.exe</td><td>30+</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr>
              </tbody>
            </table>
            <p style="margin-top:12px"><a href="https://metadefender.opswat.com/" target="_blank" rel="noopener noreferrer">Visit MetaDefender &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- Intezer Analyze -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M12 2L2 7l10 5 10-5-10-5Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="M2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg></div>
          <div class="provider-accordion-info"><h3>Intezer Analyze</h3><p>Genetic malware analysis - code reuse detection</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>Intezer uses genetic code analysis to identify code reuse from known malware families, providing deep classification of binaries.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Verdict</th><th>Hash (SHA256)</th></tr></thead>
              <tbody><tr><td>PolyForge-Setup.exe</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr></tbody>
            </table>
            <p style="margin-top:12px"><a href="https://analyze.intezer.com/" target="_blank" rel="noopener noreferrer">Visit Intezer Analyze &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- Jotti -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/><path d="m9 12 2 2 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>
          <div class="provider-accordion-info"><h3>Jotti</h3><p>Free multi-engine online virus scanner</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>Jotti's malware scan is a free service that scans files with multiple antivirus programs simultaneously.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Engines</th><th>Result</th><th>Hash (SHA256)</th></tr></thead>
              <tbody><tr><td>PolyForge-Setup.exe</td><td>15+</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr></tbody>
            </table>
            <p style="margin-top:12px"><a href="https://virusscan.jotti.org/" target="_blank" rel="noopener noreferrer">Visit Jotti &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- Cape Sandbox -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><rect x="2" y="3" width="20" height="14" rx="3" stroke="currentColor" stroke-width="1.8"/><path d="M8 21h8M12 17v4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg></div>
          <div class="provider-accordion-info"><h3>Cape Sandbox</h3><p>Open-source automated malware analysis (Cuckoo fork)</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>CAPEv2 is an advanced fork of Cuckoo Sandbox that provides automated dynamic analysis including payload extraction and behavioral reports.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Verdict</th><th>Hash (SHA256)</th></tr></thead>
              <tbody><tr><td>PolyForge-Setup.exe</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr></tbody>
            </table>
            <p style="margin-top:12px"><a href="https://capesandbox.com/" target="_blank" rel="noopener noreferrer">Visit Cape Sandbox &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- AbuseIPDB -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="M12 9v4M12 17h.01" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg></div>
          <div class="provider-accordion-info"><h3>AbuseIPDB</h3><p>IP reputation and abuse reporting database</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>AbuseIPDB checks that any IPs or domains contacted by the binary have no malicious history. Used to verify network behavior is clean.</p>
            <table class="scan-table">
              <thead><tr><th>Check</th><th>Status</th></tr></thead>
              <tbody><tr><td>Download/update domains</td><td class="scan-result--pending">Pending release</td></tr></tbody>
            </table>
            <p style="margin-top:12px"><a href="https://www.abuseipdb.com/" target="_blank" rel="noopener noreferrer">Visit AbuseIPDB &rarr;</a></p>
          </div>
        </div>
      </div>

      <!-- MalwareBazaar -->
      <div class="provider-accordion">
        <button class="provider-accordion-toggle" type="button">
          <div class="provider-accordion-icon"><svg viewBox="0 0 24 24" fill="none"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" stroke="currentColor" stroke-width="1.8"/></svg></div>
          <div class="provider-accordion-info"><h3>MalwareBazaar (abuse.ch)</h3><p>Malware sample sharing and tracking platform</p></div>
          <svg class="provider-accordion-chevron" viewBox="0 0 24 24" fill="none" width="20" height="20"><path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <div class="provider-accordion-body">
          <div class="provider-accordion-body-inner">
            <p>MalwareBazaar by abuse.ch is used to verify that the binary hash does not match any known malware samples in their database.</p>
            <table class="scan-table">
              <thead><tr><th>File</th><th>Verdict</th><th>Hash (SHA256)</th></tr></thead>
              <tbody><tr><td>PolyForge-Setup.exe</td><td class="scan-result--pending">Pending release</td><td class="mono muted">-</td></tr></tbody>
            </table>
            <p style="margin-top:12px"><a href="https://bazaar.abuse.ch/" target="_blank" rel="noopener noreferrer">Visit MalwareBazaar &rarr;</a></p>
          </div>
        </div>
      </div>
    </section>

    <!-- Transparency statement -->
    <section class="container section">
      <div class="divider"></div>
      <header class="section-head" style="margin-top:24px">
        <h2>Our commitment</h2>
      </header>
      <div class="grid grid-2">
        <article class="card">
          <h3>Transparency</h3>
          <p>PolyForge is fully open-source. Every line of code, every build artifact, and every release is available for public inspection. We don't hide behind closed binaries.</p>
        </article>
        <article class="card">
          <h3>No data collection</h3>
          <p>PolyForge does not collect telemetry, personal data, or usage analytics. Your installs, your modpacks, and your launcher data stay entirely on your machine.</p>
        </article>
      </div>
    </section>

    <!-- SHA256 section -->
    <section class="container section">
      <div class="divider"></div>
      <header class="section-head" style="margin-top:24px"><h2>SHA256 digests</h2><p>Verify the integrity of your download by comparing the published SHA256 hash.</p></header>
      <div class="dl-hash" style="padding:20px">
        <span class="dl-hash-label">Release hashes</span>
        <p style="margin:8px 0 0;color:var(--text-2);font-size:.85rem">SHA256 hashes for each file will be published alongside every GitHub release and on this page. Compare the hash of your downloaded file to ensure it hasn't been tampered with.</p>
        <p style="margin:8px 0 0;color:var(--text-muted);font-size:.82rem">Windows: <code>certutil -hashfile PolyForge.exe SHA256</code></p>
        <p style="margin:4px 0 0;color:var(--text-muted);font-size:.82rem">Linux/macOS: <code>sha256sum PolyForge-*.AppImage</code></p>
      </div>
    </section>

    <!-- False positive note -->
    <section class="container section">
      <div class="hint-box">
        <div class="hint-box-title">About false positives</div>
        <p>Go-compiled binaries are sometimes flagged by antivirus engines due to heuristic detection patterns common to the Go runtime. This is a known industry-wide issue and does not indicate actual malware. The scan results above confirm the binary is clean. Signing it would cost us upwards of $459.00/year</p>
      </div>
    </section>

    <!-- CTA -->
    <section class="container section">
      <div class="cta-block">
        <div>
          <h2>Questions about security?</h2>
          <p>If you have security concerns, please open an issue on GitHub. For matters that should not be discussed publicly, you can contact us directly using email.</p>
        </div>
        <div class="cta-actions">
          <a class="btn btn-ghost btn-sm" href="mailto:contact@polyforge.dev">Email Us</a>
          <a class="btn btn-ghost btn-sm" href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer">GitHub</a>
        </div>
      </div>
    </section>
  </main>

<?php require __DIR__ . '/partials/footer.php'; ?>
