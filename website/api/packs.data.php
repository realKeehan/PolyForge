<?php
/**
 * Private pack registry — consumed only by pack-access.php, never served
 * directly (blocked in .htaccess, and executing it outputs nothing).
 *
 * passwordHash is the SHA-256 hex of the pack password.
 * downloadUrl is only revealed after a successful password check.
 */

declare(strict_types=1);

return [
    'turtel-smp' => [
        'requiresPassword' => false,
        'passwordHash'     => null,
        'downloadUrl'      => null,
    ],
    'event-pack' => [
        'requiresPassword' => true,
        // SHA-256 of the pack password
        'passwordHash'     => '908baa40ef565d0d30fab71f76b9e73d4cf88101984c4f57c6c674804dc4006a',
        'downloadUrl'      => null, // set to the private zip URL when the pack ships
    ],
];
