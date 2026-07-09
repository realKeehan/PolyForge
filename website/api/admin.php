<?php
/**
 * Admin API — everything the admin panel (/admin) does goes through here.
 *
 * Auth: session cookie after password login (hash in admin-config.php),
 * login attempts rate-limited per IP. All state-changing requests must be
 * POSTs carrying the X-PolyForge-Admin header (CSRF guard).
 *
 * Actions (action=... in the query string):
 *   login, logout, session
 *   releases-list, release-upload, release-delete, release-type-create
 *   manifest-get, manifest-save, manifest-history
 *   packs-list, pack-save-meta, pack-selfdestruct-save, pack-delete,
 *     pack-build (online packager)
 *   stats
 */

declare(strict_types=1);

require __DIR__ . '/php-compat.php';
require __DIR__ . '/packs-registry.php';

const RELEASES_DIR   = __DIR__ . '/../releases';
const PACKS_DIR      = __DIR__ . '/../packs';
const MANIFEST_FILE  = __DIR__ . '/manifest.json';
const HISTORY_FILE   = __DIR__ . '/manifest-history.json';
const ADMIN_STATE    = __DIR__ . '/admin-state.json';
const STATS_FILE     = __DIR__ . '/stats-data.json';
const DOC_EXTENSIONS = ['md', 'txt', 'json', 'html'];

// Folders shippable in a pack (from real profile analysis); everything else
// (saves, journeymap, essential, logs, ...) is user data and never packed.
const PACK_FOLDERS = ['mods', 'config', 'resourcepacks', 'shaderpacks', 'datapacks', 'defaultconfigs', 'scripts', 'kubejs'];
const PACK_ROOT_FILES = ['options.txt', 'servers.dat'];

$config = require __DIR__ . '/admin-config.php';

header('Cache-Control: no-store');

// No `: never` return type here: the host runs PHP 7.4, where that 8.1-only
// syntax is a parse error and turns every request into a blank 500. This
// function still exits; the missing type is only a static hint.
function respond(int $status, array $body)
{
    http_response_code($status);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode($body);
    exit;
}

// Surface real errors as JSON instead of a bare 500, so the admin panel can
// show what actually went wrong (session path, permissions, PHP version, ...).
set_exception_handler(function (Throwable $e) {
    respond(500, ['error' => 'server error: ' . $e->getMessage()]);
});
register_shutdown_function(function () {
    $e = error_get_last();
    if ($e && in_array($e['type'], [E_ERROR, E_PARSE, E_CORE_ERROR, E_COMPILE_ERROR], true)) {
        respond(500, ['error' => 'fatal: ' . $e['message']]);
    }
});

function loadJson(string $path, array $fallback = []): array
{
    if (!is_file($path)) {
        return $fallback;
    }
    $decoded = json_decode((string) file_get_contents($path), true);
    return is_array($decoded) ? $decoded : $fallback;
}

function saveJson(string $path, array $data): bool
{
    return file_put_contents($path, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES), LOCK_EX) !== false;
}

function safeName(string $name): bool
{
    return (bool) preg_match('#^[A-Za-z0-9 ._()-]+$#', $name) && !str_contains($name, '..');
}

function safeType(string $type): bool
{
    return (bool) preg_match('#^[A-Za-z0-9._-]+$#', $type) && !str_contains($type, '..');
}

// ── Session ──────────────────────────────────────
session_name($config['sessionName']);
session_set_cookie_params([
    'lifetime' => $config['sessionTtl'],
    'path'     => '/',
    'httponly' => true,
    'samesite' => 'Strict',
    'secure'   => !empty($_SERVER['HTTPS']),
]);
session_start();

$action = (string) ($_GET['action'] ?? '');
$isPost = ($_SERVER['REQUEST_METHOD'] ?? '') === 'POST';

// CSRF guard: state changes must be POST + custom header.
if ($isPost && ($_SERVER['HTTP_X_POLYFORGE_ADMIN'] ?? '') !== '1') {
    respond(403, ['error' => 'missing admin header']);
}

$body = [];
if ($isPost && str_starts_with((string) ($_SERVER['CONTENT_TYPE'] ?? ''), 'application/json')) {
    $decoded = json_decode((string) file_get_contents('php://input', false, null, 0, 1 << 20), true);
    if (is_array($decoded)) {
        $body = $decoded;
    }
}

// ── Auth actions ─────────────────────────────────
if ($action === 'login') {
    if (!$isPost) {
        respond(405, ['error' => 'POST required']);
    }
    $state = loadJson(ADMIN_STATE);
    $ip = $_SERVER['REMOTE_ADDR'] ?? 'unknown';
    $now = time();
    $tries = array_values(array_filter(
        is_array($state['logins'][$ip] ?? null) ? $state['logins'][$ip] : [],
        fn($t) => $now - (int) $t < $config['loginWindowSec']
    ));
    if (count($tries) >= $config['maxLoginTries']) {
        respond(429, ['error' => 'too many attempts, try again later']);
    }

    $password = (string) ($body['password'] ?? '');
    if ($password === '' || !hash_equals($config['passwordHash'], hash('sha256', $password))) {
        $tries[] = $now;
        $state['logins'][$ip] = $tries;
        saveJson(ADMIN_STATE, $state);
        respond(403, ['error' => 'wrong password']);
    }

    unset($state['logins'][$ip]);
    saveJson(ADMIN_STATE, $state);
    session_regenerate_id(true);
    $_SESSION['admin'] = true;
    $_SESSION['since'] = $now;
    respond(200, ['ok' => true]);
}

if ($action === 'logout') {
    session_destroy();
    respond(200, ['ok' => true]);
}

if ($action === 'session') {
    respond(200, ['authenticated' => !empty($_SESSION['admin'])]);
}

// Everything below requires auth.
if (empty($_SESSION['admin'])) {
    respond(401, ['error' => 'not authenticated']);
}

switch ($action) {

// ── Releases ─────────────────────────────────────
case 'releases-list': {
    $root = realpath(RELEASES_DIR);
    $types = [];
    if ($root !== false) {
        foreach (scandir($root) as $entry) {
            $dir = $root . DIRECTORY_SEPARATOR . $entry;
            if ($entry[0] === '.' || !is_dir($dir)) {
                continue;
            }
            $files = [];
            $latest = null;
            $latestTime = -1;
            foreach (scandir($dir) as $f) {
                $p = $dir . DIRECTORY_SEPARATOR . $f;
                if ($f[0] === '.' || !is_file($p)) {
                    continue;
                }
                $ext = strtolower(pathinfo($f, PATHINFO_EXTENSION));
                $isDoc = in_array($ext, DOC_EXTENSIONS, true);
                $m = filemtime($p);
                $files[] = ['name' => $f, 'size' => filesize($p), 'mtime' => gmdate('c', $m), 'doc' => $isDoc];
                if (!$isDoc && ($m > $latestTime || ($m === $latestTime && strcmp($f, (string) $latest) > 0))) {
                    $latest = $f;
                    $latestTime = $m;
                }
            }
            usort($files, fn($a, $b) => strcmp($b['mtime'], $a['mtime']));
            $types[] = ['type' => $entry, 'latest' => $latest, 'files' => $files];
        }
    }
    respond(200, ['types' => $types]);
}

case 'release-type-create': {
    $type = (string) ($body['type'] ?? '');
    if (!$isPost || !safeType($type)) {
        respond(400, ['error' => 'invalid type name']);
    }
    $dir = RELEASES_DIR . '/' . $type;
    if (!is_dir($dir) && !mkdir($dir, 0755, true)) {
        respond(500, ['error' => 'could not create folder']);
    }
    respond(200, ['ok' => true]);
}

case 'release-upload': {
    if (!$isPost) {
        respond(405, ['error' => 'POST required']);
    }
    $type = (string) ($_POST['type'] ?? '');
    if (!safeType($type) || !is_dir(RELEASES_DIR . '/' . $type)) {
        respond(400, ['error' => 'unknown type folder']);
    }
    $up = $_FILES['file'] ?? null;
    if (!$up || $up['error'] !== UPLOAD_ERR_OK) {
        respond(400, ['error' => 'upload failed (check post_max_size/upload_max_filesize)']);
    }
    $name = basename((string) $up['name']);
    if (!safeName($name) || strtolower(pathinfo($name, PATHINFO_EXTENSION)) === 'php') {
        respond(400, ['error' => 'invalid filename']);
    }
    $dest = RELEASES_DIR . '/' . $type . '/' . $name;
    if (!move_uploaded_file($up['tmp_name'], $dest)) {
        respond(500, ['error' => 'could not store file']);
    }
    // Refresh the checksum manifest so hashes are always current for this type.
    writeReleaseSums(RELEASES_DIR . '/' . $type);
    respond(200, ['ok' => true, 'sha256' => hash_file('sha256', $dest)]);
}

case 'release-delete': {
    $type = (string) ($body['type'] ?? '');
    $name = (string) ($body['name'] ?? '');
    if (!$isPost || !safeType($type) || !safeName($name)) {
        respond(400, ['error' => 'invalid parameters']);
    }
    $path = realpath(RELEASES_DIR . '/' . $type . '/' . $name);
    $root = realpath(RELEASES_DIR);
    if ($path === false || $root === false || !str_starts_with($path, $root) || !is_file($path)) {
        respond(404, ['error' => 'file not found']);
    }
    unlink($path);
    // Keep the checksum manifest in sync after a build is removed.
    writeReleaseSums(RELEASES_DIR . '/' . $type);
    respond(200, ['ok' => true]);
}

// ── Manifest (version control + app/installer visibility) ──
case 'manifest-get': {
    respond(200, ['manifest' => loadJson(MANIFEST_FILE)]);
}

case 'manifest-save': {
    $manifest = $body['manifest'] ?? null;
    if (!$isPost || !is_array($manifest) || (int) ($manifest['schemaVersion'] ?? 0) < 1) {
        respond(400, ['error' => 'invalid manifest (schemaVersion required)']);
    }
    // Snapshot the previous manifest for history/rollback.
    $history = loadJson(HISTORY_FILE, ['entries' => []]);
    $history['entries'][] = [
        'saved'    => gmdate('c'),
        'manifest' => loadJson(MANIFEST_FILE),
    ];
    $history['entries'] = array_slice($history['entries'], -100);
    saveJson(HISTORY_FILE, $history);

    if (!saveJson(MANIFEST_FILE, $manifest)) {
        respond(500, ['error' => 'could not write manifest']);
    }
    respond(200, ['ok' => true]);
}

case 'manifest-history': {
    $history = loadJson(HISTORY_FILE, ['entries' => []]);
    // Newest first, summarized for the UI.
    $entries = array_reverse($history['entries']);
    $out = array_map(fn($e) => [
        'saved'         => $e['saved'] ?? '',
        'latestVersion' => $e['manifest']['app']['latestVersion'] ?? '',
        'minSupported'  => $e['manifest']['app']['minSupportedVersion'] ?? '',
        'manifest'      => $e['manifest'] ?? [],
    ], $entries);
    respond(200, ['entries' => $out]);
}

// ── Packs ────────────────────────────────────────
// Auto-discovers packs from the packs/ folder: each <id>-<ver>.manifest.json is
// the source of truth for identity + mod list. The registry supplies editable
// metadata (password, download URL), the public manifest supplies self-destruct
// marks (removeMods), and stats supply per-pack download counts.
case 'packs-list': {
    $registry = loadPackRegistry();
    $stats    = loadJson(STATS_FILE);
    $byPack   = is_array($stats['byPack'] ?? null) ? $stats['byPack'] : [];

    $discovered = [];
    $hosted = [];
    if (is_dir(PACKS_DIR)) {
        foreach (scandir(PACKS_DIR) as $f) {
            $p = PACKS_DIR . '/' . $f;
            if ($f[0] === '.' || !is_file($p) || str_ends_with($f, '.md')) {
                continue;
            }
            $hosted[] = ['name' => $f, 'size' => filesize($p), 'mtime' => gmdate('c', filemtime($p))];
            if (!str_ends_with($f, '.manifest.json')) {
                continue;
            }
            $pm = json_decode((string) file_get_contents($p), true);
            if (!is_array($pm) || !isset($pm['id'])) {
                continue;
            }
            $pid = (string) $pm['id'];
            $mods = [];
            foreach (($pm['mods'] ?? []) as $mod) {
                if (is_array($mod) && isset($mod['file'])) {
                    $mods[] = ['file' => (string) $mod['file'], 'name' => (string) ($mod['name'] ?? $mod['file'])];
                }
            }
            $packFile = $pid . '-' . ($pm['version'] ?? '') . '.polypack';
            $discovered[$pid] = [
                'name'    => (string) ($pm['name'] ?? $pid),
                'version' => (string) ($pm['version'] ?? ''),
                'file'    => is_file(PACKS_DIR . '/' . $packFile) ? $packFile : null,
                'mods'    => $mods,
            ];
        }
    }

    // Auto-publish any pack that is hosted on disk (built through the online
    // packager, or packed locally and uploaded/dropped into packs/) but not yet
    // listed in the public manifest, so it reaches the app on its next launch
    // without the admin having to open and re-save it. Existing entries — and
    // any admin edits to them — are left untouched.
    reconcileManifestPacks($discovered, $registry);

    $manifest = loadJson(MANIFEST_FILE);
    $manifestPacks = [];
    foreach (($manifest['modpacks'] ?? []) as $mp) {
        if (is_array($mp) && isset($mp['id'])) {
            $manifestPacks[(string) $mp['id']] = $mp;
        }
    }

    $ids = array_unique(array_merge(array_keys($registry), array_keys($discovered), array_keys($manifestPacks)));
    sort($ids);
    $packs = [];
    foreach ($ids as $id) {
        $reg  = is_array($registry[$id] ?? null) ? $registry[$id] : [];
        $disc = $discovered[$id] ?? [];
        $mp   = $manifestPacks[$id] ?? [];
        $packs[] = [
            'id'               => $id,
            'name'             => $disc['name'] ?? ($reg['name'] ?? ($mp['name'] ?? $id)),
            'version'          => $disc['version'] ?? '',
            'file'             => $disc['file'] ?? null,
            'inFolder'         => isset($discovered[$id]),
            'inManifest'       => isset($manifestPacks[$id]),
            'requiresPassword' => !empty($reg['requiresPassword']),
            'hasPassword'      => !empty($reg['passwordHash']),
            // Fall back to the hosted file so a pack dropped straight into
            // packs/ (no registry URL) still shows its real download link —
            // the same URL pack-access derives and hands to the app.
            'downloadUrl'      => ($reg['downloadUrl'] ?? null) ?: (!empty($disc['file']) ? '/packs/' . $disc['file'] : null),
            'mods'             => $disc['mods'] ?? [],
            'removeMods'       => array_values($mp['removeMods'] ?? []),
            'downloads'        => (int) ($byPack[$id] ?? 0),
        ];
    }
    respond(200, ['packs' => $packs, 'hosted' => $hosted]);
}

// Arms/updates the "self-destruct" mod removal list for a pack. Marks live in
// the public manifest's modpacks[] entry (removeMods); the app deletes those
// files from existing installs on next launch. Disarming clears the list.
case 'pack-selfdestruct-save': {
    $id = normalizePackId((string) ($body['id'] ?? ''));
    if (!$isPost || !preg_match('#^[a-z0-9-]+$#', $id)) {
        respond(400, ['error' => 'invalid pack id']);
    }
    $removeMods = [];
    if (!empty($body['armed']) && is_array($body['removeMods'] ?? null)) {
        foreach ($body['removeMods'] as $f) {
            $f = basename((string) $f); // filename only — never a path
            if ($f !== '' && $f !== '.' && $f !== '..') {
                $removeMods[] = $f;
            }
        }
        $removeMods = array_values(array_unique($removeMods));
    }

    $manifest = loadJson(MANIFEST_FILE);
    if ((int) ($manifest['schemaVersion'] ?? 0) < 1) {
        $manifest['schemaVersion'] = 1;
    }
    $modpacks = is_array($manifest['modpacks'] ?? null) ? $manifest['modpacks'] : [];
    $found = false;
    foreach ($modpacks as &$mp) {
        if (is_array($mp) && (string) ($mp['id'] ?? '') === $id) {
            if ($removeMods) {
                $mp['removeMods'] = $removeMods;
            } else {
                unset($mp['removeMods']);
            }
            $found = true;
            break;
        }
    }
    unset($mp);
    if (!$found) {
        $reg = loadPackRegistry();
        $new = ['id' => $id, 'name' => $reg[$id]['name'] ?? $id];
        if (!empty($reg[$id]['requiresPassword'])) {
            $new['requiresPassword'] = true;
        }
        if ($removeMods) {
            $new['removeMods'] = $removeMods;
        }
        $modpacks[] = $new;
    }
    $manifest['modpacks'] = array_values($modpacks);
    $manifest['updated'] = gmdate('c');

    // Snapshot for history/rollback, mirroring manifest-save.
    $history = loadJson(HISTORY_FILE, ['entries' => []]);
    $history['entries'][] = ['saved' => gmdate('c'), 'manifest' => loadJson(MANIFEST_FILE)];
    $history['entries'] = array_slice($history['entries'], -100);
    saveJson(HISTORY_FILE, $history);

    if (!saveJson(MANIFEST_FILE, $manifest)) {
        respond(500, ['error' => 'could not write manifest']);
    }
    respond(200, ['ok' => true, 'armed' => (bool) ($removeMods !== []), 'removeMods' => $removeMods]);
}

case 'pack-save-meta': {
    $id = normalizePackId((string) ($body['id'] ?? ''));
    if (!$isPost || !preg_match('#^[a-z0-9-]+$#', $id)) {
        respond(400, ['error' => 'invalid pack id — letters, numbers, and hyphens only (spaces become hyphens; other symbols are rejected)']);
    }
    $registry = loadPackRegistry();
    $entry = $registry[$id] ?? ['name' => $id, 'requiresPassword' => false, 'passwordHash' => null, 'downloadUrl' => null];
    if (array_key_exists('name', $body)) {
        $name = trim((string) $body['name']);
        if ($name !== '') {
            $entry['name'] = $name;
        }
    }
    // Never persist a blank name — fall back to the discovered pack manifest
    // (locally-packed uploads carry the real name there) and finally the id.
    if (empty($entry['name'])) {
        $entry['name'] = discoveredPackName($id) ?: $id;
    }
    if (array_key_exists('requiresPassword', $body)) {
        $entry['requiresPassword'] = (bool) $body['requiresPassword'];
    }
    if (!empty($body['password'])) {
        $entry['passwordHash'] = hash('sha256', (string) $body['password']);
    }
    if (array_key_exists('downloadUrl', $body)) {
        $entry['downloadUrl'] = $body['downloadUrl'] !== '' ? (string) $body['downloadUrl'] : null;
    }
    $registry[$id] = $entry;
    savePackRegistry($registry);

    // Mirror identity into the PUBLIC manifest so the running app actually sees
    // name/visibility edits — it only reads manifest.json, never the registry.
    // downloadUrl stays registry-only (the app resolves it through pack-access;
    // RemotePack has no url field).
    $set = ['name' => (string) $entry['name']];
    $unset = [];
    if (!empty($entry['requiresPassword'])) {
        $set['requiresPassword'] = true;
    } else {
        $unset[] = 'requiresPassword';
    }
    upsertManifestPack($id, $set, $unset);
    respond(200, ['ok' => true]);
}

case 'pack-delete': {
    $id = normalizePackId((string) ($body['id'] ?? ''));
    if (!$isPost || !preg_match('#^[a-z0-9-]+$#', $id)) {
        respond(400, ['error' => 'invalid pack id']);
    }
    $registry = loadPackRegistry();
    unset($registry[$id]);
    savePackRegistry($registry);
    if (!empty($body['deleteFiles']) && is_dir(PACKS_DIR)) {
        foreach (glob(PACKS_DIR . '/' . $id . '-*') ?: [] as $f) {
            if (is_file($f)) {
                unlink($f);
            }
        }
    }
    respond(200, ['ok' => true]);
}

// ── Online packager ──────────────────────────────
// Upload a zip of the pack source folder (a .minecraft-style profile dir);
// this extracts it, keeps only pack-worthy folders/files, reads mod
// metadata from the jars, and produces <id>-<version>.polypack plus
// the standalone update manifest in packs/.
case 'pack-build': {
    if (!$isPost) {
        respond(405, ['error' => 'POST required']);
    }
    if (!class_exists('ZipArchive')) {
        respond(501, ['error' => 'PHP zip extension is not enabled on this server - use scripts/package-modpack.ps1 locally instead']);
    }
    // Fold the id to lowercase and turn spaces into hyphens so casual input
    // ("Turtel SMP") just works; other symbols still fail loudly below so the
    // packer knows to fix them before it ships.
    $id      = normalizePackId((string) ($_POST['id'] ?? ''));
    $name    = (string) ($_POST['name'] ?? '');
    $version = (string) ($_POST['version'] ?? '');
    $mc      = (string) ($_POST['minecraft'] ?? '');
    $loader  = (string) ($_POST['loader'] ?? '');
    $loaderV = (string) ($_POST['loaderVersion'] ?? '');
    if (!preg_match('#^[a-z0-9-]+$#', $id)) {
        respond(400, ['error' => 'pack id may only contain letters, numbers, and hyphens — no spaces or symbols (letters are lowercased automatically)']);
    }
    if ($name === '' || !preg_match('#^[\w.+-]+$#', $version)) {
        respond(400, ['error' => 'name and version are required (version: letters, numbers, . + -)']);
    }
    $up = $_FILES['source'] ?? null;
    if (!$up || $up['error'] !== UPLOAD_ERR_OK) {
        respond(400, ['error' => 'source zip upload failed (check upload_max_filesize)']);
    }

    $src = new ZipArchive();
    if ($src->open($up['tmp_name']) !== true) {
        respond(400, ['error' => 'source is not a readable zip']);
    }

    // Some zips nest everything under a single top folder — detect prefix.
    $prefix = null;
    for ($i = 0; $i < min($src->numFiles, 50); $i++) {
        $n = $src->getNameIndex($i);
        $top = explode('/', $n)[0];
        if ($prefix === null) {
            $prefix = $top;
        } elseif ($prefix !== $top) {
            $prefix = '';
            break;
        }
    }
    $prefix = $prefix ? $prefix . '/' : '';

    if (!is_dir(PACKS_DIR)) {
        mkdir(PACKS_DIR, 0755, true);
    }
    $outZipPath = PACKS_DIR . "/$id-$version.pack.tmp.zip";
    @unlink($outZipPath);
    $out = new ZipArchive();
    if ($out->open($outZipPath, ZipArchive::CREATE) !== true) {
        respond(500, ['error' => 'could not create output zip']);
    }

    $mods = [];
    $files = [];
    $folders = [];
    $fileCount = 0;
    $totalBytes = 0;

    for ($i = 0; $i < $src->numFiles; $i++) {
        $entry = $src->getNameIndex($i);
        if (str_ends_with($entry, '/')) {
            continue;
        }
        $rel = $prefix !== '' && str_starts_with($entry, $prefix) ? substr($entry, strlen($prefix)) : $entry;
        if ($rel === '' || str_contains($rel, '..')) {
            continue;
        }
        $parts = explode('/', $rel);
        $isRootFile = count($parts) === 1 && in_array($parts[0], PACK_ROOT_FILES, true);
        $isPackFolder = count($parts) > 1 && in_array($parts[0], PACK_FOLDERS, true);
        if (!$isRootFile && !$isPackFolder) {
            continue; // user data (saves, journeymap, logs, ...) never ships
        }

        $data = $src->getFromIndex($i);
        if ($data === false) {
            continue;
        }
        $out->addFromString('overrides/' . $rel, $data);
        $fileCount++;
        $totalBytes += strlen($data);
        // Per-file checksum so the installer can verify every file against the
        // manifest (corruption / tampering / truncated download detection).
        $files[] = ['path' => $rel, 'sha256' => hash('sha256', $data), 'size' => strlen($data)];
        if ($isPackFolder && !in_array($parts[0], $folders, true)) {
            $folders[] = $parts[0];
        }

        // Mod metadata: filename heuristic, refined from inside the jar.
        if ($parts[0] === 'mods' && count($parts) === 2 && str_ends_with(strtolower($rel), '.jar')) {
            $mods[] = packBuildModEntry($parts[1], $data);
        }
    }
    $src->close();

    if ($fileCount === 0) {
        $out->close();
        @unlink($outZipPath);
        respond(400, ['error' => 'no pack-worthy folders found in the zip (expected mods/, config/, ...)']);
    }

    usort($mods, fn($a, $b) => strcmp($a['file'], $b['file']));
    $manifest = [
        'schemaVersion' => 1,
        'id'            => $id,
        'name'          => $name,
        'version'       => $version,
        'minecraft'     => $mc,
        'loader'        => ['type' => $loader, 'version' => $loaderV],
        'created'       => gmdate('c'),
        'mods'          => $mods,
        'overrides'     => ['folders' => $folders, 'fileCount' => $fileCount, 'totalBytes' => $totalBytes, 'files' => $files],
    ];
    // Launcher-agnostic info fields for every supported launcher, so one
    // pack installs everywhere (the installer generates each launcher's
    // real files from these + the manifest).
    $launcherIds = [
        'vanilla', 'multimc', 'polymc', 'prismlauncher', 'shatteredprism', 'elyprism',
        'ultimmc', 'fjord', 'modrinth', 'curseforge', 'atlauncher', 'gdlauncher',
        'technic', 'feather', 'bakaxl', 'sklauncher', 'freesm', 'qwertz', 'hmcl',
        'polymerium', 'xmcl',
    ];
    $launcherEntries = [];
    foreach ($launcherIds as $lid) {
        $launcherEntries[$lid] = ['profileName' => $name, 'instanceName' => $name];
    }
    $launchers = [
        'schemaVersion' => 1,
        'defaults'      => ['minMemoryMb' => 2048, 'recommendedMemoryMb' => 4096, 'javaArgs' => '', 'iconPath' => ''],
        'launchers'     => $launcherEntries,
    ];
    $manifestJson = json_encode($manifest, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
    $out->addFromString('pack-manifest.json', $manifestJson);
    $out->addFromString('launchers.json', json_encode($launchers, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));
    $out->close();

    // Wrap the built zip into a .polypack container (same transform the app
    // and the PowerShell packager use), then drop the intermediate zip.
    require __DIR__ . '/slime-lib.php';
    $zipBytes = file_get_contents($outZipPath);
    $packPath = PACKS_DIR . "/$id-$version.polypack";
    file_put_contents($packPath, slime_wrap((string) $zipBytes), LOCK_EX);
    @unlink($outZipPath);

    file_put_contents(PACKS_DIR . "/$id-$version.manifest.json", $manifestJson, LOCK_EX);

    // Register/refresh the pack entry with its hosted URL.
    $registry = loadPackRegistry();
    $entry = $registry[$id] ?? ['name' => $name, 'requiresPassword' => false, 'passwordHash' => null];
    $entry['name'] = $name;
    $entry['downloadUrl'] = "/packs/$id-$version.polypack";
    $registry[$id] = $entry;
    savePackRegistry($registry);

    // Publish it to the app's manifest too, so a freshly built pack shows up
    // (with the right name) on the next launch without a second manual step.
    upsertManifestPack($id, ['name' => $name]);

    respond(200, [
        'ok'       => true,
        'pack'     => "/packs/$id-$version.polypack",
        'manifest' => "/packs/$id-$version.manifest.json",
        'mods'     => count($mods),
        'files'    => $fileCount,
        'bytes'    => $totalBytes,
        'folders'  => $folders,
    ]);
}

// ── Stats ────────────────────────────────────────
case 'stats': {
    respond(200, ['stats' => loadJson(STATS_FILE, ['downloads' => 0])]);
}

default:
    respond(400, ['error' => 'unknown action']);
}

// ── Helpers ──────────────────────────────────────

/**
 * Upserts a pack's identity into the PUBLIC manifest's modpacks[] list — the
 * only pack source the app reads at launch. $set fields are written onto the
 * entry (created if missing); $unset keys are removed. A pre-change snapshot is
 * recorded first, mirroring manifest-save / pack-selfdestruct-save. Throws on a
 * write failure (surfaced as a JSON 500 by the global exception handler).
 */
function upsertManifestPack(string $id, array $set, array $unset = []): void
{
    $manifest = loadJson(MANIFEST_FILE);
    if ((int) ($manifest['schemaVersion'] ?? 0) < 1) {
        $manifest['schemaVersion'] = 1;
    }
    $modpacks = is_array($manifest['modpacks'] ?? null) ? $manifest['modpacks'] : [];
    $found = false;
    foreach ($modpacks as &$mp) {
        if (is_array($mp) && (string) ($mp['id'] ?? '') === $id) {
            foreach ($set as $k => $v) {
                $mp[$k] = $v;
            }
            foreach ($unset as $k) {
                unset($mp[$k]);
            }
            $found = true;
            break;
        }
    }
    unset($mp);
    if (!$found) {
        $modpacks[] = array_merge(['id' => $id], $set);
    }
    $manifest['modpacks'] = array_values($modpacks);
    $manifest['updated'] = gmdate('c');

    $history = loadJson(HISTORY_FILE, ['entries' => []]);
    $history['entries'][] = ['saved' => gmdate('c'), 'manifest' => loadJson(MANIFEST_FILE)];
    $history['entries'] = array_slice($history['entries'], -100);
    saveJson(HISTORY_FILE, $history);

    if (!saveJson(MANIFEST_FILE, $manifest)) {
        throw new RuntimeException('could not write manifest');
    }
}

/**
 * Publishes hosted-but-unlisted packs into the PUBLIC manifest so a pack that
 * was dropped into packs/ (or built locally and uploaded) reaches the app on
 * its next launch without the admin having to open and re-save it. Only packs
 * missing from modpacks[] are added — existing entries, and any admin edits to
 * them, are left untouched. Writes the manifest at most once (with a single
 * history snapshot) and only when something was actually added, so repeat calls
 * — e.g. every time the admin opens the packs tab — are cheap no-ops.
 *
 * @param array $discovered pack-id => ['name' => ..., ...] from the packs/ scan
 * @param array $registry   pack registry, consulted for requiresPassword
 */
function reconcileManifestPacks(array $discovered, array $registry): void
{
    if ($discovered === []) {
        return;
    }
    $manifest = loadJson(MANIFEST_FILE);
    $modpacks = is_array($manifest['modpacks'] ?? null) ? $manifest['modpacks'] : [];
    $known = [];
    foreach ($modpacks as $mp) {
        if (is_array($mp) && isset($mp['id'])) {
            $known[(string) $mp['id']] = true;
        }
    }

    $added = false;
    foreach ($discovered as $pid => $disc) {
        $pid = (string) $pid;
        if (isset($known[$pid])) {
            continue;
        }
        $entry = ['id' => $pid, 'name' => (string) ($disc['name'] ?? $pid)];
        if (!empty($registry[$pid]['requiresPassword'])) {
            $entry['requiresPassword'] = true;
        }
        $modpacks[] = $entry;
        $added = true;
    }
    if (!$added) {
        return;
    }

    if ((int) ($manifest['schemaVersion'] ?? 0) < 1) {
        $manifest['schemaVersion'] = 1;
    }
    $manifest['modpacks'] = array_values($modpacks);
    $manifest['updated'] = gmdate('c');

    // Snapshot the pre-change manifest for history/rollback, mirroring the
    // other manifest writers.
    $history = loadJson(HISTORY_FILE, ['entries' => []]);
    $history['entries'][] = ['saved' => gmdate('c'), 'manifest' => loadJson(MANIFEST_FILE)];
    $history['entries'] = array_slice($history['entries'], -100);
    saveJson(HISTORY_FILE, $history);

    if (!saveJson(MANIFEST_FILE, $manifest)) {
        throw new RuntimeException('could not write manifest');
    }
}

/**
 * Best-effort display name for a pack read from its discovered manifest in
 * packs/ (<id>-<version>.manifest.json), so locally-packed uploads keep their
 * real name instead of collapsing to the id. Empty string if none is found.
 */
function discoveredPackName(string $id): string
{
    if (!is_dir(PACKS_DIR)) {
        return '';
    }
    foreach (glob(PACKS_DIR . '/' . $id . '-*.manifest.json') ?: [] as $f) {
        $pm = json_decode((string) file_get_contents($f), true);
        if (is_array($pm) && (string) ($pm['id'] ?? '') === $id && !empty($pm['name'])) {
            return (string) $pm['name'];
        }
    }
    return '';
}

/**
 * (Re)generates SHA256SUMS.txt for a release type folder so download hashes
 * stay current automatically on every upload/delete. Standard coreutils
 * format (`<hex>␠␠<filename>`, one per line); doc files — including the sums
 * file itself — are skipped, matching how download.php picks the latest build.
 */
function writeReleaseSums(string $typeDir): void
{
    if (!is_dir($typeDir)) {
        return;
    }
    $lines = [];
    foreach (scandir($typeDir) ?: [] as $f) {
        $p = $typeDir . DIRECTORY_SEPARATOR . $f;
        if ($f[0] === '.' || !is_file($p)) {
            continue;
        }
        if (in_array(strtolower(pathinfo($f, PATHINFO_EXTENSION)), DOC_EXTENSIONS, true)) {
            continue; // hashes/readmes are not builds
        }
        $hash = hash_file('sha256', $p);
        if ($hash !== false) {
            $lines[] = $hash . '  ' . $f;
        }
    }
    sort($lines);
    file_put_contents($typeDir . '/SHA256SUMS.txt', $lines ? implode("\n", $lines) . "\n" : '', LOCK_EX);
}

/**
 * Builds a mod entry from a jar: filename heuristic first, then refined by
 * fabric.mod.json / quilt.mod.json / META-INF/mods.toml inside the jar.
 */
function packBuildModEntry(string $filename, string $jarBytes): array
{
    $base = preg_replace('/\.jar$/i', '', $filename);
    $name = $base;
    $version = '';
    if (preg_match('/^(.*?)[-_](v?\d[\w.+-]*)$/', $base, $m)) {
        $name = $m[1];
        $version = $m[2];
    }

    // Read authoritative metadata from inside the jar when possible.
    $tmp = tempnam(sys_get_temp_dir(), 'pfjar');
    if ($tmp !== false) {
        file_put_contents($tmp, $jarBytes);
        $jar = new ZipArchive();
        if ($jar->open($tmp) === true) {
            foreach (['fabric.mod.json', 'quilt.mod.json'] as $metaFile) {
                $raw = $jar->getFromName($metaFile);
                if ($raw !== false) {
                    $meta = json_decode($raw, true);
                    $info = $metaFile === 'quilt.mod.json' ? ($meta['quilt_loader'] ?? []) : $meta;
                    if (is_array($info)) {
                        if (!empty($info['id'])) {
                            $name = (string) $info['id'];
                        }
                        if (!empty($info['version'])) {
                            $version = (string) $info['version'];
                        }
                    }
                    break;
                }
            }
            // Forge/NeoForge mods.toml: light regex parse for modId/version.
            if ($version === '' || $name === $base) {
                $toml = $jar->getFromName('META-INF/mods.toml');
                if ($toml === false) {
                    $toml = $jar->getFromName('META-INF/neoforge.mods.toml');
                }
                if ($toml !== false) {
                    if (preg_match('/modId\s*=\s*"([^"]+)"/', $toml, $m)) {
                        $name = $m[1];
                    }
                    if (preg_match('/version\s*=\s*"(?!\$)([^"]+)"/', $toml, $m)) {
                        $version = $m[1];
                    }
                }
            }
            $jar->close();
        }
        unlink($tmp);
    }

    return [
        'file'    => $filename,
        'name'    => $name,
        'version' => $version,
        'sha256'  => hash('sha256', $jarBytes),
    ];
}
