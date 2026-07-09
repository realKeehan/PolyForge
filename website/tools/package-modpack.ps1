# Modpack packager scaffold — builds a .polypack the installer consumes.
# Format spec: docs/modpack-format.md
#
# Takes a source folder containing minecraft-style folders (mods/, config/,
# resourcepacks/, ...) and produces:
#   <OutDir>\<id>-<version>.polypack        (obfuscated zip: manifest + launchers + overrides/)
#   <OutDir>\<id>-<version>.manifest.json   (standalone manifest for hosted update checks)
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

    # Minecraft folders to include when present in SourceDir. Defaults are
    # based on real profile analysis - user data (saves, journeymap,
    # essential, emotes, logs, screenshots, ...) is intentionally absent.
    [string[]]$IncludeFolders = @(
        'mods', 'config', 'resourcepacks', 'shaderpacks',
        'datapacks', 'scripts', 'defaultconfigs', 'kubejs'
    ),
    # Additional folders beyond the defaults (e.g. 'journeymap').
    [string[]]$ExtraFolders = @(),
    # Root files packs commonly ship (default settings / server list).
    [string[]]$IncludeRootFiles = @('options.txt', 'servers.dat'),

    [string]$OutDir = ''
)

$ErrorActionPreference = 'Stop'

# ── 3x3 braille dot-matrix loader ────────────────
# A little 3x3 grid of dots that spins while the slow steps run (hashing,
# zipping). It's drawn as three braille cells, each showing one column's three
# stacked dots (braille dots 1/2/3), so together they read as a 3-wide x 3-tall
# matrix on a single line. The eight outer cells are lit (centre stays empty)
# except one "gap" that rotates around the ring, which reads as rotation. Chars
# are built from code points so the source stays pure ASCII — PowerShell 5.1
# reads a BOM-less .ps1 as ANSI and would otherwise mangle literal braille glyphs.
try { [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false) } catch {}
# Outer cells in clockwise order (each @(row, col), 0-based); the gap walks this.
$script:Spin3Ring       = @(@(0, 0), @(0, 1), @(0, 2), @(1, 2), @(2, 2), @(2, 1), @(2, 0), @(1, 0))
$script:Spin3Frame      = 0
$script:SpinnerLastTick = -1000

# Renders the 3x3 grid for a frame as three braille characters (one per column).
function Get-Spin3Grid {
    param([int]$Frame)
    $gap = $script:Spin3Ring[$Frame % $script:Spin3Ring.Count]
    $rowBit = @(0x01, 0x02, 0x04)   # grid row 0/1/2 -> braille dots 1/2/3 (left column)
    $grid = ''
    for ($c = 0; $c -lt 3; $c++) {
        $v = 0
        for ($r = 0; $r -lt 3; $r++) {
            $isCentre = ($r -eq 1 -and $c -eq 1)               # hollow centre
            $isGap    = ($r -eq $gap[0] -and $c -eq $gap[1])   # the rotating hole
            if (-not $isCentre -and -not $isGap) { $v = $v -bor $rowBit[$r] }
        }
        $grid += [char](0x2800 + $v)
    }
    return $grid
}

function Update-Spinner {
    param([string]$Label)
    # Silent when piped to a file/CI (no console to animate); throttled so the
    # spin speed is steady regardless of how fast the work arrives.
    if ([Console]::IsOutputRedirected) { return }
    $now = [Environment]::TickCount
    if (($now - $script:SpinnerLastTick) -lt 90) { return }
    $script:SpinnerLastTick = $now
    if ($Label.Length -gt 50) { $Label = $Label.Substring(0, 47) + '...' }
    $grid = Get-Spin3Grid -Frame $script:Spin3Frame
    $script:Spin3Frame++
    Write-Host ("`r{0}  {1}" -f $grid, $Label).PadRight(70) -NoNewline -ForegroundColor Cyan
}

function Complete-Spinner {
    param([string]$Label)
    $script:Spin3Frame      = 0
    $script:SpinnerLastTick = -1000
    if ([Console]::IsOutputRedirected) { Write-Host $Label -ForegroundColor Green; return }
    # Overwrite the grid with a green check + label, then break the line.
    Write-Host ("`r{0}  {1}" -f ([char]0x2713), $Label).PadRight(70) -ForegroundColor Green
}

if (-not (Test-Path $SourceDir -PathType Container)) {
    Write-Error "Source folder not found: $SourceDir"
    exit 1
}
# Normalize the id: fold to lowercase and turn spaces into hyphens so casual
# input ("Turtel SMP") just works, but reject any other symbol so a bad id is
# caught here instead of silently breaking the hosted download URL later.
$normalizedId = $PackId.Trim().ToLowerInvariant() -replace '\s+', '-' -replace '-+', '-'
$normalizedId = $normalizedId.Trim('-')
if ($normalizedId -notmatch '^[a-z0-9-]+$') {
    Write-Error "PackId may only contain letters, numbers, and hyphens (spaces become hyphens; other symbols are rejected) - got '$PackId'."
    exit 1
}
$PackId = $normalizedId

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
    $jars = @(Get-ChildItem $modsDir -Filter '*.jar' -File | Sort-Object Name)
    $n = 0
    foreach ($jar in $jars) {
        $n++
        Update-Spinner "Hashing mods ($n/$($jars.Count)): $($jar.Name)"
        $base = [IO.Path]::GetFileNameWithoutExtension($jar.Name)
        $name = $base
        $version = ''
        if ($base -match '^(.*?)-(v?\d[\w.+-]*)$') {
            $name = $matches[1]
            $version = $matches[2]
        }
        $mods += [ordered]@{
            file    = $jar.Name
            name    = $name
            version = $version
            sha256  = (Get-FileHash $jar.FullName -Algorithm SHA256).Hash.ToLower()
        }
    }
    if ($jars.Count -gt 0) { Complete-Spinner "Hashed $($jars.Count) mods" }
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
foreach ($rootFile in $IncludeRootFiles) {
    $srcFile = Join-Path $SourceDir $rootFile
    if (Test-Path $srcFile -PathType Leaf) {
        Copy-Item $srcFile -Destination $overrides
        $fileCount++
        $totalBytes += (Get-Item $srcFile).Length
        Write-Host "  root file: $rootFile" -ForegroundColor DarkGray
    }
}

# ── Per-file checksums (drive integrity verification) ─
# Hash every staged file so the installer can verify each one against the
# manifest and detect corruption, tampering, or a truncated download. Paths
# are relative to overrides/ with forward slashes so they match the archive.
$overridesPrefix = (Resolve-Path $overrides).Path.TrimEnd('\') + '\'
$fileEntries = @()
$allFiles = @(Get-ChildItem $overrides -Recurse -File | Sort-Object FullName)
$n = 0
foreach ($item in $allFiles) {
    $n++
    Update-Spinner "Hashing files ($n/$($allFiles.Count))"
    $rel = $item.FullName.Substring($overridesPrefix.Length) -replace '\\', '/'
    $fileEntries += [ordered]@{
        path   = $rel
        sha256 = (Get-FileHash $item.FullName -Algorithm SHA256).Hash.ToLower()
        size   = $item.Length
    }
}
Complete-Spinner "Hashed $($fileEntries.Count) files for integrity verification"

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
        files      = $fileEntries
    }
}
# WriteAllText writes UTF-8 without BOM - strict JSON parsers (incl. Go's)
# reject BOM-prefixed files, and PS 5.1's Out-File -Encoding utf8 adds one.
$manifestJson = $manifest | ConvertTo-Json -Depth 6
[IO.File]::WriteAllText((Join-Path $staging 'pack-manifest.json'), $manifestJson)

# ── launchers.json (launcher-agnostic info fields) ─
# The pack is generic: the installer generates each launcher's real files
# (profiles, instance configs) from these fields + the manifest. We emit an
# entry for EVERY supported launcher so one pack installs everywhere; the
# per-launcher install locations/schemas are filled in on the installer side
# (see internal/kumi/packformat.go and scripts/dump-launcher-trees.ps1).
$profileName = $PackName
$launcherIds = @(
    'vanilla', 'multimc', 'polymc', 'prismlauncher', 'shatteredprism', 'elyprism',
    'ultimmc', 'fjord', 'modrinth', 'curseforge', 'atlauncher', 'gdlauncher',
    'technic', 'feather', 'bakaxl', 'sklauncher', 'freesm', 'qwertz', 'hmcl',
    'polymerium', 'xmcl'
)
$launcherEntries = [ordered]@{}
foreach ($lid in $launcherIds) {
    $launcherEntries[$lid] = [ordered]@{ profileName = $profileName; instanceName = $profileName }
}
$launchers = [ordered]@{
    schemaVersion = 1
    defaults      = [ordered]@{
        minMemoryMb         = 2048
        recommendedMemoryMb = 4096
        javaArgs            = ''
        iconPath            = ''
    }
    launchers     = $launcherEntries
}
[IO.File]::WriteAllText((Join-Path $staging 'launchers.json'), ($launchers | ConvertTo-Json -Depth 6))

# ── Zip, then wrap into a .polypack container ────
# Future reference (heavy update): switch the payload from -tzip (DEFLATE) to
# LZMA for better ratio on Distant Horizons LODs / uncompressed packs. 7-Zip
# already supports it here (-t7z -m0=lzma2, or -tzip -mm=LZMA), but the
# READER side is the cost: the app would need a pure-Go xz codec
# (github.com/ulikunitz/xz, ~+284 KB) and a new slime flags byte. See the
# LZMA note in internal/kumi/slime.go before changing this.
# The intermediate zip is built in TEMP, never in $OutDir, so the output folder
# only ever ends up with the .polypack and its manifest — no stray .zip is left
# behind (on success or failure).
$tmpZip  = Join-Path ([IO.Path]::GetTempPath()) "polypack-$PackId-$PackVersion-$(Get-Random).zip"
$zLogOut = [IO.Path]::GetTempFileName()
$zLogErr = [IO.Path]::GetTempFileName()
try {
    # Run 7-Zip as a child process and spin the braille loader while it works.
    # Quote the (temp) output path for spaces; '*' is passed literally so 7-Zip
    # expands it against -WorkingDirectory ($staging).
    $proc = Start-Process -FilePath $sevenZip `
        -ArgumentList ('a -tzip -mx=9 "{0}" *' -f $tmpZip) `
        -WorkingDirectory $staging -NoNewWindow -PassThru `
        -RedirectStandardOutput $zLogOut -RedirectStandardError $zLogErr
    # Touch .Handle so the object keeps the process handle — otherwise
    # Start-Process -PassThru leaves .ExitCode null once the process exits.
    $null = $proc.Handle
    while (-not $proc.HasExited) {
        Update-Spinner 'Compressing pack...'
        Start-Sleep -Milliseconds 80
    }
    $proc.WaitForExit()
    if ($proc.ExitCode -ne 0) {
        Complete-Spinner 'Compression failed'
        $detail = ((Get-Content $zLogErr -Raw -ErrorAction SilentlyContinue), (Get-Content $zLogOut -Raw -ErrorAction SilentlyContinue)) -join "`n"
        Remove-Item $tmpZip -Force -ErrorAction SilentlyContinue
        Write-Error "7-Zip exited with code $($proc.ExitCode)`n$detail"
        exit 1
    }
    Complete-Spinner 'Compressed pack'
} finally {
    Remove-Item $staging -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item $zLogOut, $zLogErr -Force -ErrorAction SilentlyContinue
}

# Wrap the zip into a .polypack — PolyForge's branded, obfuscated container.
# Same transform the app reverses (internal/kumi/slime.go) and the PHP admin
# packager applies: magic header + XOR keystream. Obfuscation, not crypto.
. (Join-Path $PSScriptRoot 'slime-lib.ps1')
$outPack = Join-Path $OutDir "$PackId-$PackVersion.polypack"
try {
    Update-Spinner 'Wrapping into .polypack...'
    ConvertTo-Slime -InputPath $tmpZip -OutputPath $outPack
    Complete-Spinner 'Wrapped .polypack'
} catch {
    # Never leave a half-finished .polypack or the stray temp .zip behind.
    Remove-Item $outPack -Force -ErrorAction SilentlyContinue
    Remove-Item $tmpZip  -Force -ErrorAction SilentlyContinue
    Write-Error "Failed to wrap the pack into .polypack: $_"
    exit 1
}
Remove-Item $tmpZip -Force -ErrorAction SilentlyContinue

# Standalone manifest for hosted update checks (no pack download needed).
$manifestOut = Join-Path $OutDir "$PackId-$PackVersion.manifest.json"
[IO.File]::WriteAllText($manifestOut, $manifestJson)

$sizeMB = [math]::Round((Get-Item $outPack).Length / 1MB, 2)
Write-Host ''
Write-Host "Packaged modpack -> $outPack ($sizeMB MB)" -ForegroundColor Green
Write-Host "Update manifest  -> $manifestOut" -ForegroundColor Green
Write-Host ''
Write-Host 'Host both next to your releases (or behind api/pack-access for' -ForegroundColor Yellow
Write-Host 'password packs); the installer uses the manifest for update checks.' -ForegroundColor Yellow
