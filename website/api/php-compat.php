<?php
/**
 * PHP 7.4 compatibility shims.
 *
 * The production host (polyforge.dev) runs PHP 7.4, which predates the PHP 8.0
 * string helpers used across the API. Define them when missing so every
 * endpoint behaves identically on 7.4 and on 8.x. Require this at the very top
 * of each API entry point, before the helpers are used.
 *
 * (Named-argument calls such as file_get_contents(..., length: N) are a
 * language feature that can't be polyfilled — those are written positionally
 * at the call sites instead.)
 */

declare(strict_types=1);

/**
 * App-wide timezone.
 *
 * The site is operated from Pacific time, and download/pack stats are bucketed
 * per calendar day. Bucketing in UTC rolled the "today" counter to tomorrow
 * during Pacific evenings (UTC is 7-8h ahead), so stats looked a day ahead.
 * Pinning a single fixed zone here — and using date()/mktime() instead of the
 * UTC-locked gmdate() — keeps every day bucket and timestamp on Pacific dates.
 * Fixed (not per-visitor) because the aggregates are shared across all users.
 * Keep this value in sync with APP_TIMEZONE in the admin page's JS.
 */
// defined() guard: this file is plain-`require`d and can load twice in one
// request (the dev router + an API endpoint), where a bare const would fatal.
if (!defined('APP_TIMEZONE')) {
    define('APP_TIMEZONE', 'America/Los_Angeles');
}
date_default_timezone_set(APP_TIMEZONE);

if (!function_exists('str_contains')) {
    function str_contains(string $haystack, string $needle): bool
    {
        return $needle === '' || strpos($haystack, $needle) !== false;
    }
}

if (!function_exists('str_starts_with')) {
    function str_starts_with(string $haystack, string $needle): bool
    {
        return strncmp($haystack, $needle, strlen($needle)) === 0;
    }
}

if (!function_exists('str_ends_with')) {
    function str_ends_with(string $haystack, string $needle): bool
    {
        return $needle === '' || substr($haystack, -strlen($needle)) === $needle;
    }
}
