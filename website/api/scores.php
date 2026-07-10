<?php
/**
 * PolyTris leaderboard endpoint.
 *
 *   GET  /api/scores  → {"scores":[{"name":"AAAAA","score":1234,"lines":5,"level":2,"date":"..."}]}
 *   POST /api/scores  → body {"name":"AAAAA","score":1234,"lines":5,"level":2}
 *
 * Scores live in tetris-scores.json next to this file (blocked from direct
 * web access via .htaccess). Top 10 is returned; top 50 is kept on disk.
 * Submissions are rate-limited per IP.
 */

declare(strict_types=1);

require __DIR__ . '/php-compat.php';

header('Content-Type: application/json; charset=utf-8');
header('Cache-Control: no-store');

const DATA_FILE     = __DIR__ . '/tetris-scores.json';
const MAX_KEPT      = 50;
const MAX_RETURNED  = 10;
const RATE_LIMIT_S  = 10;
const MAX_SCORE     = 9999999;
const MAX_LINES     = 9999;
const MAX_LEVEL     = 100;

function loadData(): array
{
    if (!is_file(DATA_FILE)) {
        return ['scores' => [], 'ips' => []];
    }
    $raw = file_get_contents(DATA_FILE);
    $data = json_decode($raw === false ? '' : $raw, true);
    if (!is_array($data)) {
        return ['scores' => [], 'ips' => []];
    }
    $data['scores'] = is_array($data['scores'] ?? null) ? $data['scores'] : [];
    $data['ips']    = is_array($data['ips'] ?? null) ? $data['ips'] : [];
    return $data;
}

function saveData(array $data): void
{
    file_put_contents(DATA_FILE, json_encode($data, JSON_PRETTY_PRINT), LOCK_EX);
}

function topScores(array $scores, int $limit): array
{
    usort($scores, fn(array $a, array $b): int => ($b['score'] <=> $a['score']) ?: strcmp($a['date'] ?? '', $b['date'] ?? ''));
    return array_slice($scores, 0, $limit);
}

function respond(int $status, array $body) // exits; no `: never` (PHP 7.4 host)
{
    http_response_code($status);
    echo json_encode($body);
    exit;
}

$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';

if ($method === 'GET') {
    $data = loadData();
    respond(200, ['scores' => topScores($data['scores'], MAX_RETURNED)]);
}

if ($method !== 'POST') {
    respond(405, ['error' => 'method not allowed']);
}

// ── Submission ───────────────────────────────────

$raw = file_get_contents('php://input', false, null, 0, 4096);
$body = json_decode($raw === false ? '' : $raw, true);
if (!is_array($body)) {
    respond(400, ['error' => 'invalid JSON body']);
}

$name  = strtoupper(trim((string) ($body['name'] ?? '')));
$score = $body['score'] ?? null;
$lines = $body['lines'] ?? null;
$level = $body['level'] ?? null;

if (!preg_match('/^[A-Z0-9 ]{1,5}$/', $name) || trim($name) === '') {
    respond(400, ['error' => 'name must be 1-5 characters (A-Z, 0-9)']);
}
if (!is_int($score) || $score < 1 || $score > MAX_SCORE) {
    respond(400, ['error' => 'invalid score']);
}
if (!is_int($lines) || $lines < 0 || $lines > MAX_LINES) {
    respond(400, ['error' => 'invalid lines']);
}
if (!is_int($level) || $level < 1 || $level > MAX_LEVEL) {
    respond(400, ['error' => 'invalid level']);
}

$data = loadData();

// Per-IP rate limit
$ip  = $_SERVER['REMOTE_ADDR'] ?? 'unknown';
$now = time();
$last = $data['ips'][$ip] ?? 0;
if ($now - (int) $last < RATE_LIMIT_S) {
    respond(429, ['error' => 'slow down']);
}
// Keep the IP map from growing unbounded
$data['ips'] = array_filter($data['ips'], fn($ts) => $now - (int) $ts < 3600);
$data['ips'][$ip] = $now;

$data['scores'][] = [
    'name'  => $name,
    'score' => $score,
    'lines' => $lines,
    'level' => $level,
    'date'  => date('c'),
];
$data['scores'] = topScores($data['scores'], MAX_KEPT);

saveData($data);

respond(200, ['scores' => topScores($data['scores'], MAX_RETURNED)]);
