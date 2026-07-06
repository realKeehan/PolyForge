<?php
/**
 * Shared pack registry helpers.
 *
 * The registry lives in packs-data.json (admin-managed, gitignored, blocked
 * from direct access). On first use it is seeded from packs.data.php so a
 * fresh deploy works before the admin ever touches it.
 *
 * Entry shape:
 *   "<pack-id>" => [
 *     'name'             => display name,
 *     'requiresPassword' => bool,
 *     'passwordHash'     => sha256 hex or null,
 *     'downloadUrl'      => public URL or null,
 *   ]
 */

declare(strict_types=1);

const PACKS_REGISTRY_FILE = __DIR__ . '/packs-data.json';

function loadPackRegistry(): array
{
    if (is_file(PACKS_REGISTRY_FILE)) {
        $decoded = json_decode((string) file_get_contents(PACKS_REGISTRY_FILE), true);
        if (is_array($decoded)) {
            return $decoded;
        }
    }
    // Seed from the committed defaults.
    $seedFile = __DIR__ . '/packs.data.php';
    $seed = is_file($seedFile) ? require $seedFile : [];
    return is_array($seed) ? $seed : [];
}

function savePackRegistry(array $registry): bool
{
    return file_put_contents(
        PACKS_REGISTRY_FILE,
        json_encode($registry, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES),
        LOCK_EX
    ) !== false;
}
