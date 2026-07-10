<?php
/**
 * Minimal VirusTotal API v3 client (file report lookup by hash).
 *
 * Public/free-tier keys are rate limited (~4 requests/min, 500/day), so this is
 * only ever driven by an explicit admin action — never per visitor. Results are
 * cached in security-data.json and the public /security page reads that cache.
 *
 * We look a build up by its SHA-256; VirusTotal already has a report for any
 * file the community has scanned (our public downloads usually qualify). When a
 * hash is unknown VT returns 404 — surfaced as status "not_found" with a link to
 * submit it manually, rather than uploading the (potentially large) binary here.
 */

declare(strict_types=1);

/**
 * Looks up a file report by SHA-256. Returns a normalized array:
 *   ['status' => clean|flagged|not_found|error, 'stats' => [...], 'engines' => N,
 *    'permalink' => url, 'analysisDate' => 'c'|null, 'note' => '']
 * status is 'error' with a human note on auth/quota/transport failures.
 */
function vtLookupHash(string $sha256, string $apiKey): array
{
    $permalink = 'https://www.virustotal.com/gui/file/' . $sha256;
    $out = [
        'status'       => 'error',
        'stats'        => [],
        'engines'      => 0,
        'permalink'    => $permalink,
        'analysisDate' => null,
        'note'         => '',
    ];

    if ($apiKey === '') {
        $out['note'] = 'no VirusTotal API key configured';
        return $out;
    }
    if (!preg_match('/^[a-f0-9]{64}$/i', $sha256)) {
        $out['note'] = 'invalid sha256';
        return $out;
    }

    [$code, $raw, $err] = vtHttpGet(
        'https://www.virustotal.com/api/v3/files/' . strtolower($sha256),
        $apiKey
    );

    if ($err !== '') {
        $out['note'] = 'request failed: ' . $err;
        return $out;
    }
    if ($code === 404) {
        $out['status'] = 'not_found';
        $out['note']   = 'not yet on VirusTotal — submit the file to generate a report';
        return $out;
    }
    if ($code === 401) {
        $out['note'] = 'VirusTotal rejected the API key (401)';
        return $out;
    }
    if ($code === 429) {
        $out['note'] = 'VirusTotal rate limit hit (429) — wait a minute and retry';
        return $out;
    }
    if ($code !== 200) {
        $out['note'] = 'unexpected VirusTotal response (' . $code . ')';
        return $out;
    }

    $data = json_decode($raw, true);
    $attr = $data['data']['attributes'] ?? null;
    if (!is_array($attr) || !is_array($attr['last_analysis_stats'] ?? null)) {
        $out['note'] = 'VirusTotal response missing analysis stats';
        return $out;
    }

    $stats = array_map('intval', $attr['last_analysis_stats']);
    $malicious  = (int) ($stats['malicious'] ?? 0);
    $suspicious = (int) ($stats['suspicious'] ?? 0);
    $engines = array_sum($stats);

    $out['status']       = ($malicious + $suspicious) > 0 ? 'flagged' : 'clean';
    $out['stats']        = $stats;
    $out['engines']      = $engines;
    $out['analysisDate'] = isset($attr['last_analysis_date'])
        ? date('c', (int) $attr['last_analysis_date'])
        : null;
    return $out;
}

/**
 * Performs the authenticated GET, preferring cURL and falling back to the
 * streams API. Returns [httpCode, body, errorString]; errorString is non-empty
 * only on a transport failure (DNS/TLS/timeout), never on an HTTP error status.
 */
function vtHttpGet(string $url, string $apiKey): array
{
    if (function_exists('curl_init')) {
        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER     => ['x-apikey: ' . $apiKey, 'Accept: application/json'],
            CURLOPT_TIMEOUT        => 20,
            CURLOPT_CONNECTTIMEOUT => 10,
            CURLOPT_USERAGENT      => 'PolyForge-Security/1.0',
        ]);
        $body = curl_exec($ch);
        $err  = $body === false ? curl_error($ch) : '';
        $code = (int) curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        return [$code, is_string($body) ? $body : '', $err];
    }

    // Streams fallback (ignore_errors keeps the body on 4xx/5xx).
    $ctx = stream_context_create(['http' => [
        'method'        => 'GET',
        'header'        => "x-apikey: {$apiKey}\r\nAccept: application/json\r\nUser-Agent: PolyForge-Security/1.0\r\n",
        'timeout'       => 20,
        'ignore_errors' => true,
    ]]);
    $body = @file_get_contents($url, false, $ctx);
    if ($body === false) {
        return [0, '', 'stream request failed'];
    }
    $code = 0;
    foreach ($http_response_header ?? [] as $h) {
        if (preg_match('#^HTTP/\S+\s+(\d{3})#', $h, $m)) {
            $code = (int) $m[1];
        }
    }
    return [$code, $body, ''];
}
