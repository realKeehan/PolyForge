<?php
/**
 * Admin panel configuration TEMPLATE.
 *
 * Copy this file to admin-config.php (same folder) and fill in real values.
 * admin-config.php is gitignored — it holds secrets — and is blocked from
 * direct web access via .htaccess.
 *
 * passwordHash is the SHA-256 hex of the admin password:
 *   php -r "echo hash('sha256', 'your-new-password');"
 * Consider something long; this guards release uploads and manifest edits.
 */

declare(strict_types=1);

return [
    'passwordHash'   => 'CHANGE-ME-sha256-hex-of-admin-password',
    'sessionName'    => 'pf_admin',
    'sessionTtl'     => 60 * 60 * 8, // 8 hours
    'maxLoginTries'  => 5,           // per IP...
    'loginWindowSec' => 300,         // ...per 5 minutes

    // VirusTotal API key for automated release scans (Security tab / security
    // page). Prefer the VIRUSTOTAL_API_KEY environment variable (set it in
    // cPanel); the literal is only a fallback for hosts without env control.
    'virusTotalApiKey' => getenv('VIRUSTOTAL_API_KEY') ?: '',
];
