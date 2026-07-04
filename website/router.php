<?php
/**
 * Router for the PHP built-in dev server. Mirrors the production .htaccess:
 * extensionless URLs map to .php files, partials are blocked, unknown paths
 * get the 404 page.
 *
 *   php -S localhost:8080 router.php   (run from the website/ directory)
 */

declare(strict_types=1);

// Dev-server only — if this ever executes under Apache, pretend it isn't here.
if (PHP_SAPI !== 'cli-server') {
    http_response_code(404);
    require __DIR__ . '/404.php';
    exit;
}

$uri = urldecode((string) parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH));

// Never serve the shared partials directly
if (preg_match('#^/partials(/|$)#', $uri)) {
    http_response_code(403);
    exit('Forbidden');
}

// Serve existing static files (css, js, images, json, ico) as-is
$file = __DIR__ . $uri;
if ($uri !== '/' && is_file($file) && !str_ends_with($file, '.php')) {
    return false;
}

// /index → / (matches production)
if ($uri === '/index') {
    header('Location: /', true, 301);
    return true;
}

// Root → index.php
if ($uri === '/') {
    require __DIR__ . '/index.php';
    return true;
}

// Direct .php request → strip extension (matches the production 301)
if (str_ends_with($uri, '.php') && is_file($file)) {
    header('Location: ' . substr($uri, 0, -4), true, 301);
    return true;
}

// Legacy .html request → strip extension
if (str_ends_with($uri, '.html')) {
    header('Location: ' . substr($uri, 0, -5), true, 301);
    return true;
}

// Extensionless URL → serve the matching .php page
$page = __DIR__ . rtrim($uri, '/') . '.php';
if (is_file($page)) {
    require $page;
    return true;
}

// Anything else → 404 page
require __DIR__ . '/404.php';
return true;
