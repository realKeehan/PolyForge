# .polypack container helpers ("slime" is the internal codec name) — must
# stay byte-for-byte compatible with internal/kumi/slime.go and
# website/api/slime-lib.php.
#
# Layout: "SLIME" + 0x01 (version) + 0x00 (flags) + 0x00 (reserved) + payload
# payload[i] = zip[i] XOR key[i % 32] XOR (i & 0xFF)
# key = SHA-256("PolyForge-Slime-v1")   (32 bytes)
#
# This is obfuscation for branding/format-obscurity, NOT encryption.

function Get-SlimeKey {
    $sha = [Security.Cryptography.SHA256]::Create()
    try {
        return $sha.ComputeHash([Text.Encoding]::ASCII.GetBytes('PolyForge-Slime-v1'))
    } finally {
        $sha.Dispose()
    }
}

function Invoke-SlimeTransform {
    param([byte[]]$Bytes)
    $key = Get-SlimeKey
    $out = New-Object byte[] $Bytes.Length
    for ($i = 0; $i -lt $Bytes.Length; $i++) {
        $out[$i] = $Bytes[$i] -bxor $key[$i % 32] -bxor ($i -band 0xFF)
    }
    return $out
}

function ConvertTo-Slime {
    param(
        [Parameter(Mandatory)][string]$InputPath,
        [Parameter(Mandatory)][string]$OutputPath
    )
    $zip = [IO.File]::ReadAllBytes($InputPath)
    $payload = Invoke-SlimeTransform -Bytes $zip
    $header = [byte[]]@(0x53, 0x4C, 0x49, 0x4D, 0x45, 0x01, 0x00, 0x00) # "SLIME" v1
    $out = New-Object byte[] ($header.Length + $payload.Length)
    [Array]::Copy($header, 0, $out, 0, $header.Length)
    [Array]::Copy($payload, 0, $out, $header.Length, $payload.Length)
    [IO.File]::WriteAllBytes($OutputPath, $out)
}

function ConvertFrom-Slime {
    param(
        [Parameter(Mandatory)][string]$InputPath,
        [Parameter(Mandatory)][string]$OutputPath
    )
    $data = [IO.File]::ReadAllBytes($InputPath)
    if ($data.Length -lt 8 -or $data[0] -ne 0x53 -or $data[1] -ne 0x4C -or $data[2] -ne 0x49 -or $data[3] -ne 0x4D -or $data[4] -ne 0x45) {
        throw "Not a .slime file (bad magic): $InputPath"
    }
    $payload = New-Object byte[] ($data.Length - 8)
    [Array]::Copy($data, 8, $payload, 0, $payload.Length)
    [IO.File]::WriteAllBytes($OutputPath, (Invoke-SlimeTransform -Bytes $payload))
}
