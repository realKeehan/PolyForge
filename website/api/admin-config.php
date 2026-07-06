<?php
/**
 * Admin panel configuration. Blocked from direct web access via .htaccess.
 *
 * passwordHash is the SHA-256 hex of the admin password.
 * Default password: "keehanadmin" — CHANGE THIS before going live:
 *   php -r "echo hash('sha256', 'your-new-password');"
 * and paste the result below. Consider something long; this guards
 * release uploads and manifest edits.
 */

declare(strict_types=1);

return [
    'passwordHash'   => '50681a4974dc89b97b1c41722867ecc077ad09bd69a10d4a543464bfdb791a06', // keehanadmin
    'sessionName'    => 'pf_admin',
    'sessionTtl'     => 60 * 60 * 8, // 8 hours
    'maxLoginTries'  => 5,           // per IP...
    'loginWindowSec' => 300,         // ...per 5 minutes
];
