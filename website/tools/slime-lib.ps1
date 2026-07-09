# .polypack container helpers ("slime" is the internal codec name) — must
# stay byte-for-byte compatible with internal/kumi/slime.go and
# website/api/slime-lib.php.
#
# Layout: "SLIME" + 0x01 (version) + 0x00 (flags) + 0x00 (reserved) + payload
# payload[i] = zip[i] XOR key[i % 32] XOR (i & 0xFF)
# key = SHA-256("PolyForge-Slime-v1")   (32 bytes)
#
# This is obfuscation for branding/format-obscurity, NOT encryption.
#
# The transform is streamed in fixed-size chunks so multi-GB packs (Distant
# Horizons LODs, uncompressed resource packs, ...) don't blow up memory or hit
# .NET's ~2 GB single-array limit — the reason a large pack could previously
# fail the wrap step and leave a stray ".polypack.zip" behind.

function Get-SlimeKey {
    $sha = [Security.Cryptography.SHA256]::Create()
    try {
        return $sha.ComputeHash([Text.Encoding]::ASCII.GetBytes('PolyForge-Slime-v1'))
    } finally {
        $sha.Dispose()
    }
}

# The keystream key[i % 32] XOR (i & 0xFF) repeats every 256 bytes (LCM of the
# 32-byte key and the 256-value counter), so one 256-byte period is the whole
# pad. XORing against it reproduces the per-byte formula exactly.
function Get-SlimePad {
    $key = Get-SlimeKey
    $pad = New-Object byte[] 256
    for ($i = 0; $i -lt 256; $i++) {
        $pad[$i] = $key[$i % 32] -bxor $i
    }
    return $pad
}

$script:SlimeHeader = [byte[]]@(0x53, 0x4C, 0x49, 0x4D, 0x45, 0x01, 0x00, 0x00) # "SLIME" v1

# XOR a stream against the repeating slime pad. Symmetric — same call both
# encodes and decodes. $startPos is the global byte offset of the first byte
# read from $In (the header is not part of the transform, so payload starts 0).
function Invoke-SlimeStream {
    param(
        [Parameter(Mandatory)][IO.Stream]$In,
        [Parameter(Mandatory)][IO.Stream]$Out
    )
    $pad = Get-SlimePad
    $buf = New-Object byte[] (1 -shl 20) # 1 MiB chunks
    $pos = 0                             # global index mod 256
    while (($read = $In.Read($buf, 0, $buf.Length)) -gt 0) {
        for ($j = 0; $j -lt $read; $j++) {
            $buf[$j] = $buf[$j] -bxor $pad[$pos]
            $pos = ($pos + 1) -band 0xFF
        }
        $Out.Write($buf, 0, $read)
    }
}

function ConvertTo-Slime {
    param(
        [Parameter(Mandatory)][string]$InputPath,
        [Parameter(Mandatory)][string]$OutputPath
    )
    $in = [IO.File]::OpenRead($InputPath)
    try {
        $out = [IO.File]::Create($OutputPath)
        try {
            $out.Write($script:SlimeHeader, 0, $script:SlimeHeader.Length)
            Invoke-SlimeStream -In $in -Out $out
        } finally { $out.Dispose() }
    } finally { $in.Dispose() }
}

function ConvertFrom-Slime {
    param(
        [Parameter(Mandatory)][string]$InputPath,
        [Parameter(Mandatory)][string]$OutputPath
    )
    $in = [IO.File]::OpenRead($InputPath)
    try {
        $magic = New-Object byte[] 8
        $got = $in.Read($magic, 0, 8)
        if ($got -lt 8 -or $magic[0] -ne 0x53 -or $magic[1] -ne 0x4C -or $magic[2] -ne 0x49 -or $magic[3] -ne 0x4D -or $magic[4] -ne 0x45) {
            throw "Not a .polypack/.slime file (bad magic): $InputPath"
        }
        $out = [IO.File]::Create($OutputPath)
        try {
            Invoke-SlimeStream -In $in -Out $out
        } finally { $out.Dispose() }
    } finally { $in.Dispose() }
}
