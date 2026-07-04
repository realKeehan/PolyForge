# Modpack packager scaffold — builds a .polypack.zip the installer consumes.
# Format spec: docs/modpack-format.md
#
# Takes a source folder containing minecraft-style folders (mods/, config/,
# resourcepacks/, ...) and produces:
#   <OutDir>\<id>-<version>.polypack.zip   (pack-manifest.json + launchers.json + overrides/)
#   <OutDir>\<id>-<version>.manifest.json  (standalone manifest for hosted update checks)
#
# The mods list in pack-manifest.json (names + versions + hashes) is the
# only thing used for update decisions. launchers.json carries per-launcher
# info fields; the installer generates the actual launcher files from them.
#
# STATUS: scaffold. Folder defaults and per-launcher fields will be
# finalized once real pack structures from the test machine are provided.
#
# Example:
#   pwsh scripts/package-modpack.ps1 -SourceDir C:\packs\turtel-src `
#     -PackId turtel-smp -PackName "Turtel SMP" -PackVersion 1.0.0 `
#     -McVersion 1.20.1 -Loader quilt -LoaderVersion 0.22.0

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$SourceDir,
    [Parameter(Mandatory)][string]$PackId,
    [Parameter(Mandatory)][string]$PackName,
    [Parameter(Mandatory)][string]$PackVersion,
    [string]$McVersion = '',
    [ValidateSet('', 'fabric', 'forge', 'neoforge', 'quilt', 'vanilla')]
    [string]$Loader = '',
    [string]$LoaderVersion = '',

    # Minecraft folders to include when present in SourceDir.
    # TODO: finalize defaults from the test-machine pack structures.
    [string[]]$IncludeFolders = @(
        'mods', 'config', 'resourcepacks', 'shaderpacks',
        'datapacks', 'scripts', 'defaultconfigs', 'kubejs'
    ),
    # Additional folders beyond the defaults (e.g. 'journeymap').
    [string[]]$ExtraFolders = @(),

    [string]$OutDir = ''
)

$ErrorActionPreference = 'Stop'

if (-not (Test-Path $SourceDir -PathType Container)) {
    Write-Error "Source folder not found: $SourceDir"
    exit 1
}
if ($PackId -notmatch '^[a-z0-9-]+$') {
    Write-Error "PackId must be lowercase letters, digits, and hyphens (got '$PackId')"
    exit 1
}

$root = Split-Path -Parent $PSScriptRoot
if (-not $OutDir) { $OutDir = Join-Path $root 'build\packs' }
if (-not (Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

# ── Locate 7-Zip ─────────────────────────────────
$sevenZip = $null
$cmd = Get-Command 7z -ErrorAction SilentlyContinue
if ($cmd) { $sevenZip = $cmd.Source }
if (-not $sevenZip) {
    foreach ($candidate in @(
        (Join-Path $env:ProgramFiles '7-Zip\7z.exe'),
        (Join-Path ${env:ProgramFiles(x86)} '7-Zip\7z.exe')
    )) {
        if ($candidate -and (Test-Path $candidate)) { $sevenZip = $candidate; break }
    }
}
if (-not $sevenZip) {
    Write-Error '7-Zip not found. Install it or add 7z.exe to PATH.'
    exit 1
}

# ── Collect override folders ─────────────────────
$allFolders = @($IncludeFolders) + @($ExtraFolders) | Select-Object -Unique
$foundFolders = @()
foreach ($folder in $allFolders) {
    if (Test-Path (Join-Path $SourceDir $folder) -PathType Container) {
        $foundFolders += $folder
    }
}
if ($foundFolders.Count -eq 0) {
    Write-Error "No known minecraft folders found in $SourceDir (looked for: $($allFolders -join ', '))"
    exit 1
}
Write-Host "Including folders: $($foundFolders -join ', ')" -ForegroundColor Cyan

# ── Build mod list (drives updates) ──────────────
# Best-effort name/version split from the jar filename: the version starts
# at the last hyphen followed by a digit ("sodium-fabric-0.5.3.jar").
# TODO: read fabric.mod.json / META-INF/mods.toml inside the jar for
# authoritative metadata instead.
$mods = @()
$modsDir = Join-Path $SourceDir 'mods'
if (Test-Path $modsDir -PathType Container) {
    foreach ($jar in (Get-ChildItem $modsDir -Filter '*.jar' -File | Sort-Object Name)) {
        $base = [IO.Path]::GetFileNameWithoutExtension($jar.Name)
        $name = $base
        $version = ''
        if ($base -match '^(.*?)-(v?\d[\w.+-]*)$') {
            $name = $matches[1]
            $version = $matches[2]
        }
        Write-Host "  mod: $name $version" -ForegroundColor DarkGray
        $mods += [ordered]@{
            file    = $jar.Name
            name    = $name
            version = $version
            sha256  = (Get-FileHash $jar.FullName -Algorithm SHA256).Hash.ToLower()
        }
    }
}
Write-Host "Found $($mods.Count) mods" -ForegroundColor Cyan

# ── Stage the archive layout ─────────────────────
$staging = Join-Path ([IO.Path]::GetTempPath()) "polypack-$PackId-$(Get-Random)"
$overrides = Join-Path $staging 'overrides'
New-Item -ItemType Directory -Path $overrides | Out-Null

$fileCount = 0
$totalBytes = 0
foreach ($folder in $foundFolders) {
    Copy-Item (Join-Path $SourceDir $folder) -Destination $overrides -Recurse
    $items = Get-ChildItem (Join-Path $overrides $folder) -Recurse -File
    $fileCount += @($items).Count
    $totalBytes += (@($items) | Measure-Object Length -Sum).Sum
}

# ── pack-manifest.json ───────────────────────────
$manifest = [ordered]@{
    schemaVersion = 1
    id            = $PackId
    name          = $PackName
    version       = $PackVersion
    minecraft     = $McVersion
    loader        = [ordered]@{ type = $Loader; version = $LoaderVersion }
    created       = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
    mods          = $mods
    overrides     = [ordered]@{
        folders    = $foundFolders
        fileCount  = $fileCount
        totalBytes = $totalBytes
    }
}
# WriteAllText writes UTF-8 without BOM - strict JSON parsers (incl. Go's)
# reject BOM-prefixed files, and PS 5.1's Out-File -Encoding utf8 adds one.
$manifestJson = $manifest | ConvertTo-Json -Depth 6
[IO.File]::WriteAllText((Join-Path $staging 'pack-manifest.json'), $manifestJson)

# ── launchers.json (info fields only) ────────────
# The installer generates the actual launcher files (profiles, instance
# configs) from these fields plus the manifest. TODO: finalize fields per
# launcher once the test-machine structures are provided.
$launchers = [ordered]@{
    schemaVersion = 1
    defaults      = [ordered]@{
        minMemoryMb         = 2048
        recommendedMemoryMb = 4096
        javaArgs            = ''
        iconPath            = ''
    }
    launchers     = [ordered]@{
        vanilla    = [ordered]@{ profileName = $PackName }
        multimc    = [ordered]@{ instanceName = $PackName }
        modrinth   = [ordered]@{ profileName = $PackName }
        curseforge = [ordered]@{ instanceName = $PackName }
    }
}
[IO.File]::WriteAllText((Join-Path $staging 'launchers.json'), ($launchers | ConvertTo-Json -Depth 6))

# ── Zip it ───────────────────────────────────────
$outZip = Join-Path $OutDir "$PackId-$PackVersion.polypack.zip"
if (Test-Path $outZip) { Remove-Item $outZip -Force }

Push-Location $staging
try {
    & $sevenZip a -tzip -mx=9 $outZip *
    if ($LASTEXITCODE -ne 0) {
        Write-Error "7-Zip exited with code $LASTEXITCODE"
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
    Remove-Item $staging -Recurse -Force -ErrorAction SilentlyContinue
}

# Standalone manifest for hosted update checks (no zip download needed).
$manifestOut = Join-Path $OutDir "$PackId-$PackVersion.manifest.json"
[IO.File]::WriteAllText($manifestOut, $manifestJson)

$sizeMB = [math]::Round((Get-Item $outZip).Length / 1MB, 2)
Write-Host ''
Write-Host "Packaged modpack -> $outZip ($sizeMB MB)" -ForegroundColor Green
Write-Host "Update manifest  -> $manifestOut" -ForegroundColor Green
Write-Host ''
Write-Host 'Host both next to your releases (or behind api/pack-access for' -ForegroundColor Yellow
Write-Host 'password packs); the installer uses the manifest for update checks.' -ForegroundColor Yellow
