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

header('Content-Type: application/json; charset=utf-8');
header('Cache-Control: no-store');

const STATE_FILE   = __DIR__ . '/pack-access-state.json';
const MAX_ATTEMPTS = 10;     // per IP...
const WINDOW_S     = 300;    // ...per 5 minutes

function respond(int $status, array $body) // exits; no `: never` (PHP 7.4 host)
{
    http_response_code($status);
    echo json_encode($body);
    exit;
}

if (($_SERVER['REQUEST_METHOD'] ?? '') !== 'POST') {
    respond(405, ['granted' => false, 'error' => 'method not allowed']);
}

$raw = file_get_contents('php://input', false, null, 0, 4096);
$body = json_decode($raw === false ? '' : $raw, true);
if (!is_array($body)) {
    respond(400, ['granted' => false, 'error' => 'invalid JSON body']);
}

$packId   = (string) ($body['packId'] ?? '');
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
require __DIR__ . '/packs-registry.php';
$packs = loadPackRegistry();
$pack = $packs[$packId] ?? null;

if ($pack === null) {
    respond(404, ['granted' => false, 'error' => 'unknown pack']);
}

if (empty($pack['requiresPassword'])) {
    respond(200, ['granted' => true, 'url' => $pack['downloadUrl']]);
}

$expected = (string) ($pack['passwordHash'] ?? '');
if ($expected === '' || $password === '') {
    respond(403, ['granted' => false, 'error' => 'incorrect password']);
}

if (!hash_equals($expected, hash('sha256', $password))) {
    respond(403, ['granted' => false, 'error' => 'incorrect password']);
}

respond(200, ['granted' => true, 'url' => $pack['downloadUrl']]);
