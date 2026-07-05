<?php
/**
 * Counting download gateway.
 *
 * Primary — latest file for a download type (stable URLs, set once):
 *   GET /api/download?type=windows
 *   → serves the newest file in releases/windows/, so publishing a release
 *     is just dropping the new build into the type folder.
 *
 * Secondary — exact file (pinned/older versions):
 *   GET /api/download?f=windows/PolyForge-5.5.2-windows-amd64.exe
 *
 * Both validate against releases/, increment the counters in
 * stats-data.json (total + per type), then 302-redirect so Apache serves
 * the file itself. Nothing about the visitor is stored — only totals.
 *
 * Doc files (.md/.txt/.json/.html) inside type folders are ignored when
 * picking the latest, so SHA256SUMS.txt and READMEs are safe to keep there.
 */

declare(strict_types=1);

const STATS_FILE   = __DIR__ . '/stats-data.json';
const RELEASES_DIR = __DIR__ . '/../releases';
const DOC_EXTENSIONS = ['md', 'txt', 'json', 'html'];

function fail(int $status, string $message): never
{
    http_response_code($status);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode(['error' => $message]);
    exit;
}

function recordDownload(?string $type): void
{
    $stats = [];
    if (is_file(STATS_FILE)) {
        $decoded = json_decode((string) file_get_contents(STATS_FILE), true);
        if (is_array($decoded)) {
            $stats = $decoded;
        }
    }
    $stats['downloads'] = max(0, (int) ($stats['downloads'] ?? 0)) + 1;
    if ($type !== null) {
        $byType = is_array($stats['byType'] ?? null) ? $stats['byType'] : [];
        $byType[$type] = max(0, (int) ($byType[$type] ?? 0)) + 1;
        $stats['byType'] = $byType;
    }
    $stats['updated'] = gmdate('c');
    file_put_contents(STATS_FILE, json_encode($stats, JSON_PRETTY_PRINT), LOCK_EX);
}

function redirectToRelease(string $relative): never
{
    $encoded = implode('/', array_map('rawurlencode', explode('/', $relative)));
    header('Cache-Control: no-store');
    header('Location: /releases/' . $encoded, true, 302);
    exit;
}

$releasesRoot = realpath(RELEASES_DIR);
if ($releasesRoot === false) {
    fail(500, 'releases folder missing');
}

// ── Latest file for a download type ──────────────
$type = (string) ($_GET['type'] ?? '');
if ($type !== '') {
    if (!preg_match('#^[A-Za-z0-9._-]+$#', $type) || str_contains($type, '..')) {
        fail(400, 'invalid type');
    }
    $dir = realpath($releasesRoot . '/' . $type);
    if ($dir === false || !str_starts_with($dir, $releasesRoot . DIRECTORY_SEPARATOR) || !is_dir($dir)) {
        fail(404, 'unknown download type');
    }

    $latest = null;
    $latestTime = -1;
    foreach (scandir($dir) as $entry) {
        $path = $dir . DIRECTORY_SEPARATOR . $entry;
        if ($entry[0] === '.' || !is_file($path)) {
            continue;
        }
        $ext = strtolower(pathinfo($entry, PATHINFO_EXTENSION));
        if (in_array($ext, DOC_EXTENSIONS, true)) {
            continue; // hashes/readmes never count as builds
        }
        $mtime = filemtime($path);
        // Newest file wins; name breaks ties for deterministic results.
        if ($mtime > $latestTime || ($mtime === $latestTime && strcmp($entry, (string) $latest) > 0)) {
            $latest = $entry;
            $latestTime = $mtime;
        }
    }
    if ($latest === null) {
        fail(404, 'no builds available for this type yet');
    }

    recordDownload($type);
    redirectToRelease($type . '/' . $latest);
}

// ── Exact file (pinned links, older versions) ────
$file = (string) ($_GET['f'] ?? '');
if ($file === '') {
    fail(400, 'missing type or f parameter');
}
if (!preg_match('#^[A-Za-z0-9._-]+/[A-Za-z0-9 ._()-]+$#', $file) || str_contains($file, '..')) {
    fail(400, 'invalid file parameter');
}
$path = realpath($releasesRoot . '/' . $file);
if ($path === false || !str_starts_with($path, $releasesRoot . DIRECTORY_SEPARATOR) || !is_file($path)) {
    fail(404, 'file not found');
}

recordDownload(explode('/', $file)[0]);
redirectToRelease($file);
