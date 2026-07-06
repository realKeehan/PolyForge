<?php
/**
 * .slime container helpers — must stay byte-for-byte compatible with
 * internal/kumi/slime.go and scripts/slime-lib.ps1.
 *
 * Layout: "SLIME" + 0x01 (version) + 0x00 (flags) + 0x00 (reserved) + payload
 * payload[i] = zip[i] XOR key[i % 32] XOR (i & 0xFF)
 * key = SHA-256("PolyForge-Slime-v1")   (32 bytes)
 *
 * Obfuscation for branding / format obscurity, NOT encryption.
 */

declare(strict_types=1);

function slime_key(): string
{
    // raw binary, 32 bytes
    return hash('sha256', 'PolyForge-Slime-v1', true);
}

function slime_transform(string $bytes): string
{
    $key = slime_key();
    $len = strlen($bytes);
    $out = '';
    for ($i = 0; $i < $len; $i++) {
        $out .= chr(ord($bytes[$i]) ^ ord($key[$i % 32]) ^ ($i & 0xFF));
    }
    return $out;
}

function slime_wrap(string $zipBytes): string
{
    return "SLIME\x01\x00\x00" . slime_transform($zipBytes);
}

function slime_unwrap(string $data): string
{
    if (strlen($data) < 8 || substr($data, 0, 5) !== 'SLIME') {
        throw new RuntimeException('not a .slime file (bad magic)');
    }
    return slime_transform(substr($data, 8));
}
