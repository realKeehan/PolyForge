<?php
/**
 * Site download statistics.
 *
 *   GET /api/stats → {"downloads":123,"updated":"..."}
 *
 * Counters live in stats-data.json next to this file (blocked from direct
 * web access) and are incremented by download.php whenever a release file
 * is fetched through the counting gateway. No personal data is stored —
 * just totals, in keeping with the privacy policy.
 */

declare(strict_types=1);

header('Content-Type: application/json; charset=utf-8');
header('Cache-Control: no-store');

const STATS_FILE = __DIR__ . '/stats-data.json';

$stats = ['downloads' => 0, 'updated' => null];
if (is_file(STATS_FILE)) {
    $decoded = json_decode((string) file_get_contents(STATS_FILE), true);
    if (is_array($decoded)) {
        $stats['downloads'] = max(0, (int) ($decoded['downloads'] ?? 0));
        $stats['updated']   = $decoded['updated'] ?? null;
    }
}

echo json_encode($stats);
