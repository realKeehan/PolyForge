<?php
/**
 * Counting download gateway.
 *
 *   GET /api/download?f=<version>/<filename>
 *   e.g. /api/download?f=5.6.0/PolyForge-5.6.0-windows-amd64.exe
 *
 * Validates the file against the releases/ folder, increments the download
 * counter in stats-data.json, then 302-redirects so Apache serves the file
 * itself. Point the downloads-page buttons and the manifest downloadUrl at
 * this gateway so every download is counted. Nothing about the visitor is
 * stored — only the total.
 */

declare(strict_types=1);

const STATS_FILE   = __DIR__ . '/stats-data.json';
const RELEASES_DIR = __DIR__ . '/../releases';

function fail(int $status, string $message): never
{
    http_response_code($status);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode(['error' => $message]);
    exit;
}

$file = (string) ($_GET['f'] ?? '');

// Strict shape: <version-ish folder>/<filename>, no traversal characters.
if (!preg_match('#^[A-Za-z0-9._-]+/[A-Za-z0-9 ._()-]+$#', $file) || str_contains($file, '..')) {
    fail(400, 'invalid file parameter');
}

$path = realpath(RELEASES_DIR . '/' . $file);
$releasesRoot = realpath(RELEASES_DIR);
if ($path === false || $releasesRoot === false || !str_starts_with($path, $releasesRoot . DIRECTORY_SEPARATOR) || !is_file($path)) {
    fail(404, 'file not found');
}

// ── Count ────────────────────────────────────────
$stats = ['downloads' => 0];
if (is_file(STATS_FILE)) {
    $decoded = json_decode((string) file_get_contents(STATS_FILE), true);
    if (is_array($decoded)) {
        $stats = $decoded;
    }
}
$stats['downloads'] = max(0, (int) ($stats['downloads'] ?? 0)) + 1;
$stats['updated']   = gmdate('c');
file_put_contents(STATS_FILE, json_encode($stats, JSON_PRETTY_PRINT), LOCK_EX);

// ── Hand off to Apache ───────────────────────────
$encoded = implode('/', array_map('rawurlencode', explode('/', $file)));
header('Cache-Control: no-store');
header('Location: /releases/' . $encoded, true, 302);
