<?php
/**
 * Shared security-analysis helpers.
 *
 * security-data.json holds automated VirusTotal results (keyed by build
 * SHA-256) plus manually added analyses from other providers. Both the admin
 * API (api/admin.php) and the public /security page render securityEntries(),
 * which flattens the two sources into one list ordered newest-checked first —
 * satisfying "order them based off which ones have been checked latest".
 */

declare(strict_types=1);

/** Loads security-data.json with the expected shape, tolerating a missing file. */
function securityLoad(string $path): array
{
    $data = ['virustotal' => [], 'manual' => []];
    if (is_file($path)) {
        $decoded = json_decode((string) file_get_contents($path), true);
        if (is_array($decoded)) {
            $data['virustotal'] = is_array($decoded['virustotal'] ?? null) ? $decoded['virustotal'] : [];
            $data['manual']     = is_array($decoded['manual'] ?? null) ? $decoded['manual'] : [];
        }
    }
    return $data;
}

/**
 * Flattens VirusTotal + manual analyses into a single list of display rows,
 * sorted by lastChecked descending (newest first). Each row:
 *   ['kind','id','provider','file','verdict','detail','url','reportFile','lastChecked']
 * verdict is one of clean|flagged|malicious|suspicious|informational|pending|error.
 */
function securityEntries(array $data): array
{
    $rows = [];

    foreach (($data['virustotal'] ?? []) as $hash => $v) {
        if (!is_array($v)) {
            continue;
        }
        $stats   = is_array($v['stats'] ?? null) ? $v['stats'] : [];
        $engines = (int) ($v['engines'] ?? array_sum(array_map('intval', $stats)));
        $flagged = (int) ($stats['malicious'] ?? 0) + (int) ($stats['suspicious'] ?? 0);
        $status  = (string) ($v['status'] ?? 'error');

        $verdict = 'error';
        $detail  = (string) ($v['note'] ?? '');
        if ($status === 'clean' || $status === 'flagged') {
            $verdict = $status === 'flagged' ? 'flagged' : 'clean';
            $detail  = $flagged . '/' . max($engines, 1) . ' engines flagged';
        } elseif ($status === 'not_found') {
            $verdict = 'pending';
            $detail  = 'not yet on VirusTotal';
        }

        $rows[] = [
            'kind'        => 'virustotal',
            'id'          => (string) ($v['sha256'] ?? $hash),
            'provider'    => 'VirusTotal',
            'file'        => (string) ($v['file'] ?? ''),
            'verdict'     => $verdict,
            'detail'      => $detail,
            'url'         => (string) ($v['permalink'] ?? ''),
            'reportFile'  => '',
            'lastChecked' => (string) ($v['lastChecked'] ?? ($v['analysisDate'] ?? '')),
        ];
    }

    foreach (($data['manual'] ?? []) as $m) {
        if (!is_array($m)) {
            continue;
        }
        $rows[] = [
            'kind'        => 'manual',
            'id'          => (string) ($m['id'] ?? ''),
            'provider'    => (string) ($m['provider'] ?? 'Unknown'),
            'file'        => (string) ($m['file'] ?? ''),
            'verdict'     => (string) ($m['verdict'] ?? 'clean'),
            'detail'      => (string) ($m['notes'] ?? ''),
            'url'         => (string) ($m['url'] ?? ''),
            'reportFile'  => (string) ($m['reportFile'] ?? ''),
            'lastChecked' => (string) ($m['lastChecked'] ?? ''),
        ];
    }

    // Newest-checked first; strcmp on ISO-8601 sorts chronologically.
    usort($rows, fn($a, $b) => strcmp($b['lastChecked'], $a['lastChecked']));
    return $rows;
}
