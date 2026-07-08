<?php
/**
 * Server-side pack password verification.
 *
 *   POST /api/pack-access  body: {"packId":"event-pack","password":"..."}
 *
 * Responses:
 *   200 {"granted":true,"url":"..."}    password correct (url may be null until the pack ships)
 *   403 {"granted":false,"error":"..."} wrong password
 *   404 {"granted":false,"error":"..."} unknown pack
 *   429 {"granted":false,"error":"..."} too many attempts from this IP
 *
 * Unlike the old client-side check, the password hash never leaves the
 * server, and the download URL is only revealed after verification.
 */

declare(strict_types=1);

require __DIR__ . '/php-compat.php';
require __DIR__ . '/packs-registry.php';

header('Content-Type: application/json; charset=utf-8');
header('Cache-Control: no-store');

const STATE_FILE   = __DIR__ . '/pack-access-state.json';
const STATS_FILE   = __DIR__ . '/stats-data.json';
const PACKS_DIR    = __DIR__ . '/../packs';
const MAX_ATTEMPTS = 10;     // per IP...
const WINDOW_S     = 300;    // ...per 5 minutes

function respond(int $status, array $body) // exits; no `: never` (PHP 7.4 host)
{
    http_response_code($status);
    echo json_encode($body);
    exit;
}

/**
 * Counts one download per granted pack access, into the same stats-data.json
 * the homepage/admin already read (total under `packDownloads`, per-pack under
 * `byPack`, plus daily history). Called only once access is granted.
 */
function recordPackDownload(string $packId): void
{
    $stats = [];
    if (is_file(STATS_FILE)) {
        $decoded = json_decode((string) file_get_contents(STATS_FILE), true);
        if (is_array($decoded)) {
            $stats = $decoded;
        }
    }
    $stats['packDownloads'] = max(0, (int) ($stats['packDownloads'] ?? 0)) + 1;
    $byPack = is_array($stats['byPack'] ?? null) ? $stats['byPack'] : [];
    $byPack[$packId] = max(0, (int) ($byPack[$packId] ?? 0)) + 1;
    $stats['byPack'] = $byPack;

    $today = gmdate('Y-m-d');
    $history = is_array($stats['history'] ?? null) ? $stats['history'] : [];
    $day = is_array($history[$today] ?? null) ? $history[$today] : [];
    $day['packs'] = max(0, (int) ($day['packs'] ?? 0)) + 1;
    $history[$today] = $day;
    $stats['history'] = $history;

    $stats['updated'] = gmdate('c');
    file_put_contents(STATS_FILE, json_encode($stats, JSON_PRETTY_PRINT), LOCK_EX);
}

/**
 * Derives the public download URL for a pack that is hosted on disk but has no
 * explicit URL in the registry (e.g. a .polypack dropped straight into packs/
 * rather than built through the online packager). Reads the newest matching
 * <id>-<version>.manifest.json and returns /packs/<id>-<version>.polypack when
 * that container actually exists, else null. The id is re-validated against the
 * pack-id charset so it can never escape packs/ through glob or traversal.
 */
function discoverPackDownloadUrl(string $packId): ?string
{
    if (!preg_match('#^[a-z0-9-]+$#', $packId) || !is_dir(PACKS_DIR)) {
        return null;
    }
    $best = null;
    $bestTime = -1;
    foreach (glob(PACKS_DIR . '/' . $packId . '-*.manifest.json') ?: [] as $manifestPath) {
        $pm = json_decode((string) file_get_contents($manifestPath), true);
        if (!is_array($pm) || (string) ($pm['id'] ?? '') !== $packId) {
            continue;
        }
        $version = (string) ($pm['version'] ?? '');
        $packName = $packId . '-' . $version . '.polypack';
        if ($version === '' || !is_file(PACKS_DIR . '/' . $packName)) {
            continue;
        }
        $mtime = (int) filemtime($manifestPath);
        if ($mtime > $bestTime) {
            $bestTime = $mtime;
            $best = '/packs/' . $packName;
        }
    }
    return $best;
}

/**
 * Turns a root-relative download URL (/packs/...) into an absolute one against
 * the current request origin, so the app can fetch it directly instead of
 * receiving a path it has no base for. URLs that are already absolute — or
 * null/empty — are returned unchanged.
 */
function absolutePackUrl(?string $url): ?string
{
    if ($url === null || $url === '' || preg_match('#^https?://#i', $url)) {
        return $url;
    }
    $scheme = (!empty($_SERVER['HTTPS']) && strtolower((string) $_SERVER['HTTPS']) !== 'off') ? 'https' : 'http';
    $host = (string) ($_SERVER['HTTP_HOST'] ?? 'polyforge.dev');
    return $scheme . '://' . $host . ($url[0] === '/' ? $url : '/' . $url);
}

if (($_SERVER['REQUEST_METHOD'] ?? '') !== 'POST') {
    respond(405, ['granted' => false, 'error' => 'method not allowed']);
}

$raw = file_get_contents('php://input', false, null, 0, 4096);
$body = json_decode($raw === false ? '' : $raw, true);
if (!is_array($body)) {
    respond(400, ['granted' => false, 'error' => 'invalid JSON body']);
}

// Pack ids are canonical (lowercase, spaces as hyphens); normalize the incoming
// id so a case/space mismatch from an older client still resolves to the right
// registry/disk entry.
$packId   = normalizePackId((string) ($body['packId'] ?? ''));
$password = (string) ($body['password'] ?? '');

if ($packId === '' || strlen($packId) > 64) {
    respond(400, ['granted' => false, 'error' => 'invalid packId']);
}

// ── Rate limit (attempts per IP per window) ──────
$ip  = $_SERVER['REMOTE_ADDR'] ?? 'unknown';
$now = time();
$state = [];
if (is_file(STATE_FILE)) {
    $decoded = json_decode((string) file_get_contents(STATE_FILE), true);
    if (is_array($decoded)) {
        $state = $decoded;
    }
}
// Drop expired windows
$state = array_filter($state, fn($s) => is_array($s) && $now - (int) ($s['start'] ?? 0) < WINDOW_S);

$slot = $state[$ip] ?? ['start' => $now, 'count' => 0];
if ($now - (int) $slot['start'] >= WINDOW_S) {
    $slot = ['start' => $now, 'count' => 0];
}
$slot['count']++;
$state[$ip] = $slot;
file_put_contents(STATE_FILE, json_encode($state), LOCK_EX);

if ($slot['count'] > MAX_ATTEMPTS) {
    respond(429, ['granted' => false, 'error' => 'too many attempts, try again later']);
}

// ── Verify ───────────────────────────────────────
$packs = loadPackRegistry();
$pack = $packs[$packId] ?? null;

// Resolve the download URL. An explicit registry URL wins, but a pack hosted by
// simply dropping <id>-<version>.polypack into packs/ (or a registered pack
// whose URL was never filled in, like the seed defaults) has none — derive it
// from the file on disk so the link still propagates to the app.
$downloadUrl = is_array($pack) ? ($pack['downloadUrl'] ?? null) : null;
if ($downloadUrl === null || $downloadUrl === '') {
    $downloadUrl = discoverPackDownloadUrl($packId);
}

// A pack is "known" if it's registered OR physically hosted on disk.
if ($pack === null && $downloadUrl === null) {
    respond(404, ['granted' => false, 'error' => 'unknown pack']);
}

// Open packs (no password) hand back the URL immediately.
if (empty($pack['requiresPassword'])) {
    recordPackDownload($packId);
    respond(200, ['granted' => true, 'url' => absolutePackUrl($downloadUrl)]);
}

$expected = (string) ($pack['passwordHash'] ?? '');
if ($expected === '' || $password === '') {
    respond(403, ['granted' => false, 'error' => 'incorrect password']);
}

if (!hash_equals($expected, hash('sha256', $password))) {
    respond(403, ['granted' => false, 'error' => 'incorrect password']);
}

recordPackDownload($packId);
respond(200, ['granted' => true, 'url' => absolutePackUrl($downloadUrl)]);
