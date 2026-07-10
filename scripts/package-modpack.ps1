# Modpack packager scaffold — builds a .polypack the installer consumes.
# Format spec: docs/modpack-format.md
#
# Takes a source folder containing minecraft-style folders (mods/, config/,
# resourcepacks/, ...) and produces:
#   <OutDir>\<id>-<version>.polypack        (obfuscated zip: manifest + launchers + overrides/)
#   <OutDir>\<id>-<version>.manifest.json   (standalone manifest for hosted update checks)
#
# The mods list in pack-manifest.json (ids + versions + hashes) is the
# only thing used for update decisions. launchers.json carries per-launcher
# info fields; the installer generates the actual launcher files from them.
#
# Mod identity is read from the loader metadata inside each jar
# (fabric.mod.json / quilt.mod.json / META-INF/[neoforge.]mods.toml /
# litemod.json); the filename split is only the fallback. Loader versions
# resolve/validate against the loaders' official metadata (Fabric Meta,
# Quilt Meta, Forge/NeoForge Maven), and mods are matched to their Modrinth
# source by hash. All network steps degrade to warnings; -Offline skips them.
#
# Example:
#   pwsh scripts/package-modpack.ps1 -SourceDir C:\packs\turtel-src `
#     -PackId turtel-smp -PackName "Turtel SMP" -PackVersion 1.0.0 `
#     -McVersion 1.20.1 -Loader quilt -LoaderVersion latest

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$SourceDir,
    [Parameter(Mandatory)][string]$PackId,
    [Parameter(Mandatory)][string]$PackName,
    [Parameter(Mandatory)][string]$PackVersion,
    [string]$McVersion = '',
    [ValidateSet('', 'fabric', 'forge', 'neoforge', 'quilt', 'liteloader', 'vanilla')]
    [string]$Loader = '',
    # Empty or 'latest' resolves the newest stable release for -McVersion from
    # the loader's official metadata service; an explicit version is validated
    # against the same source (warning only, never fails the build).
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

    # Skip every network call (loader resolution/validation, Minecraft version
    # check, Modrinth source matching). The pack still builds fine offline.
    [switch]$Offline,

    [string]$OutDir = ''
)

$ErrorActionPreference = 'Stop'

# ── Metadata service access ──────────────────────
# The loader/mod metadata APIs need TLS 1.2+ (PS 5.1 doesn't enable it by
# default) and a User-Agent identifying the tool - Quilt Meta and Modrinth
# both require a descriptive UA.
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
$script:UserAgent = 'PolyForge-Packager/2.0 (+https://polyforge.dev)'

# GET/POST a metadata endpoint (JSON or XML). Returns $null on any failure so
# callers degrade gracefully when offline - packaging never needs the network.
function Invoke-MetaApi {
    param(
        [Parameter(Mandatory)][string]$Uri,
        [string]$Method = 'Get',
        [string]$Body = $null
    )
    if ($Offline) { return $null }
    try {
        $req = @{
            Uri        = $Uri
            Method     = $Method
            UserAgent  = $script:UserAgent
            TimeoutSec = 20
        }
        if ($Body) {
            $req.Body        = $Body
            $req.ContentType = 'application/json'
        }
        return Invoke-RestMethod @req
    } catch {
        Write-Verbose "Request failed: $Uri ($_)"
        return $null
    }
}

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

# ── Mod metadata from inside the archive ─────────
# Mod filenames in the wild are too inconsistent for reliable parsing, so the
# loader metadata files inside each jar are the authority (same approach as
# the online packager in website/api/admin.php): fabric.mod.json /
# quilt.mod.json / META-INF/[neoforge.]mods.toml / litemod.json. The filename
# split stays as the fallback only.
Add-Type -AssemblyName System.IO.Compression.FileSystem

function Get-ModArchiveMetadata {
    param([Parameter(Mandatory)][string]$Path)
    $meta = @{ Id = ''; Name = ''; Version = '' }
    $zip = $null
    try {
        $zip = [IO.Compression.ZipFile]::OpenRead($Path)
        # Reads one zip entry as text; $null when the entry doesn't exist.
        $readEntry = {
            param($entryName)
            $entry = $zip.GetEntry($entryName)
            if (-not $entry) { return $null }
            $sr = New-Object IO.StreamReader($entry.Open())
            try { return $sr.ReadToEnd() } finally { $sr.Dispose() }
        }

        # Fabric / Quilt (JSON): id, version, display name.
        foreach ($jsonName in 'fabric.mod.json', 'quilt.mod.json') {
            $raw = & $readEntry $jsonName
            if (-not $raw) { continue }
            try {
                $json = $raw | ConvertFrom-Json
                $info = if ($jsonName -eq 'quilt.mod.json') { $json.quilt_loader } else { $json }
                if ($info.id)      { $meta.Id = [string]$info.id }
                if ($info.version) { $meta.Version = [string]$info.version }
                $display = if ($jsonName -eq 'quilt.mod.json') { $info.metadata.name } else { $json.name }
                if ($display) { $meta.Name = [string]$display }
            } catch { }
            break
        }

        # Forge / NeoForge (TOML): light regex parse, like the online packager.
        if (-not $meta.Id) {
            $toml = & $readEntry 'META-INF/mods.toml'
            if (-not $toml) { $toml = & $readEntry 'META-INF/neoforge.mods.toml' }
            if ($toml) {
                if ($toml -match '(?m)^\s*modId\s*=\s*"([^"]+)"')       { $meta.Id = $matches[1] }
                if ($toml -match '(?m)^\s*displayName\s*=\s*"([^"]+)"') { $meta.Name = $matches[1] }
                if ($toml -match '(?m)^\s*version\s*=\s*"([^"]+)"')     { $meta.Version = $matches[1] }
                # "${file.jarVersion}" defers to the jar's own manifest.
                if ($meta.Version -like '*${*') {
                    $meta.Version = ''
                    $mf = & $readEntry 'META-INF/MANIFEST.MF'
                    if ($mf -and $mf -match '(?m)^Implementation-Version:\s*(.+?)\s*$') {
                        $meta.Version = $matches[1]
                    }
                }
            }
        }

        # LiteLoader (.litemod): legacy, but the metadata is one JSON away.
        if (-not $meta.Id) {
            $raw = & $readEntry 'litemod.json'
            if ($raw) {
                try {
                    $json = $raw | ConvertFrom-Json
                    if ($json.name)    { $meta.Id = [string]$json.name; $meta.Name = [string]$json.name }
                    if ($json.version) { $meta.Version = [string]$json.version }
                } catch { }
            }
        }
    } catch { } finally {
        if ($zip) { $zip.Dispose() }
    }
    return $meta
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

# ── Resolve loader + game versions ───────────────
# Each loader's official metadata source (see docs/modpack-format.md): Fabric
# Meta and Quilt Meta are JSON APIs built for launchers/tools; Forge and
# NeoForge publish Maven metadata (their installer jars are the install-time
# source of truth, but versions resolve from the same Maven). LiteLoader is
# legacy - recorded verbatim, never resolved.

function Get-MavenVersions {
    param([Parameter(Mandatory)][string]$Uri)
    $xml = Invoke-MetaApi $Uri
    if (-not $xml) { return @() }
    try { return @($xml.metadata.versioning.versions.version) } catch { return @() }
}

# Returns the version to record in the manifest, or $null when the metadata
# service is unreachable. Unknown explicit versions warn but pass through.
function Resolve-LoaderVersion {
    param([string]$Type, [string]$Mc, [string]$Requested)

    $wantLatest = ($Requested -eq '' -or $Requested -eq 'latest')
    switch ($Type) {
        'fabric' {
            $list = Invoke-MetaApi "https://meta.fabricmc.net/v2/versions/loader/$Mc"
            if (-not $list) { return $null }
            if ($wantLatest) {
                $stable = @($list | Where-Object { $_.loader.stable })
                if ($stable.Count -gt 0) { return [string]$stable[0].loader.version }
                return [string]$list[0].loader.version
            }
            if (@($list | Where-Object { $_.loader.version -eq $Requested }).Count -eq 0) {
                Write-Warning "Fabric loader $Requested is not listed for Minecraft $Mc on meta.fabricmc.net."
            }
            return $Requested
        }
        'quilt' {
            $list = Invoke-MetaApi "https://meta.quiltmc.org/v3/versions/loader/$Mc"
            if (-not $list) { return $null }
            if ($wantLatest) {
                # Quilt Meta has no stable flag (a hyphen marks a beta/pre) and,
                # unlike Fabric Meta, the list is NOT newest-first — verified
                # 2026-07: stable 0.24.0 sat mid-list while 0.29.2 was newest.
                # Sort by parsed version instead of trusting list order.
                $stable = @($list | Where-Object { $_.loader.version -notmatch '-' })
                $pool = if ($stable.Count -gt 0) { $stable } else { @($list) }
                $newest = $pool | Sort-Object {
                    $base = ($_.loader.version -split '[-+]')[0]
                    try { [version]$base } catch { [version]'0.0' }
                } | Select-Object -Last 1
                return [string]$newest.loader.version
            }
            if (@($list | Where-Object { $_.loader.version -eq $Requested }).Count -eq 0) {
                Write-Warning "Quilt loader $Requested is not listed for Minecraft $Mc on meta.quiltmc.org."
            }
            return $Requested
        }
        'forge' {
            # promotions_slim.json carries Forge's own recommended/latest per
            # MC version; the Maven list validates explicit versions.
            if ($wantLatest) {
                $promos = Invoke-MetaApi 'https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json'
                if ($promos) {
                    $pick = $promos.promos."$Mc-recommended"
                    if (-not $pick) { $pick = $promos.promos."$Mc-latest" }
                    if ($pick) { return [string]$pick }
                }
            }
            $versions = Get-MavenVersions 'https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml'
            if ($versions.Count -eq 0) { return $null }
            $matching = @($versions | Where-Object { $_ -like "$Mc-*" })
            if ($wantLatest) {
                if ($matching.Count -eq 0) { return $null }
                return ($matching[-1] -split '-', 2)[1]   # maven list runs oldest -> newest
            }
            if ($matching -notcontains "$Mc-$Requested") {
                Write-Warning "Forge $Requested is not on the Forge Maven for Minecraft $Mc."
            }
            return $Requested
        }
        'neoforge' {
            if ($Mc -eq '1.20.1') {
                # 1.20.1 NeoForge lives under the legacy net.neoforged:forge
                # artifact, which upstream advises against targeting now.
                Write-Warning 'NeoForge for Minecraft 1.20.1 uses the legacy net.neoforged:forge artifact; pass its version explicitly.'
                return $Requested
            }
            # Two NeoForge version schemes (verified against the Maven 2026-07):
            #   1.x era:  MC without the leading "1." + build  (1.21.4 -> 21.4.x)
            #   26.x era: MC padded to three segments + build  (26.2 -> 26.2.0.x,
            #             26.1.2 -> 26.1.2.x)
            if ($Mc -match '^1\.(\d+)(?:\.(\d+))?$') {
                $prefix = "$($matches[1]).$([int]$matches[2])."
            } elseif ($Mc -match '^\d+(\.\d+){1,2}$') {
                $parts = @($Mc -split '\.')
                while ($parts.Count -lt 3) { $parts += '0' }
                $prefix = ($parts -join '.') + '.'
            } else {
                return $Requested
            }
            $versions = Get-MavenVersions 'https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml'
            if ($versions.Count -eq 0) { return $null }
            $matching = @($versions | Where-Object { $_ -like "$prefix*" })
            if ($wantLatest) {
                if ($matching.Count -eq 0) { return $null }
                $stable = @($matching | Where-Object { $_ -notmatch '-' })
                if ($stable.Count -gt 0) { return $stable[-1] }
                return $matching[-1]
            }
            if ($matching -notcontains $Requested) {
                Write-Warning "NeoForge $Requested is not on the NeoForge Maven (expected a $prefix* version for Minecraft $Mc)."
            }
            return $Requested
        }
    }
    return $Requested
}

if ($Loader -and $Loader -notin @('vanilla', 'liteloader')) {
    if (-not $McVersion) {
        if ($LoaderVersion -eq '' -or $LoaderVersion -eq 'latest') {
            Write-Warning "-McVersion is required to resolve the latest $Loader version; leaving the loader version empty."
            $LoaderVersion = ''
        }
    } elseif ($Offline) {
        if ($LoaderVersion -eq 'latest') {
            Write-Error 'Cannot resolve -LoaderVersion latest with -Offline; pass an explicit version.'
            exit 1
        }
    } else {
        $resolved = Resolve-LoaderVersion -Type $Loader -Mc $McVersion -Requested $LoaderVersion
        if ($null -eq $resolved) {
            if ($LoaderVersion -eq '' -or $LoaderVersion -eq 'latest') {
                Write-Error "Could not resolve the latest $Loader version for Minecraft $McVersion (offline, or no $Loader release for that Minecraft version). Pass an explicit -LoaderVersion or -Offline."
                exit 1
            }
            Write-Warning "Could not reach the $Loader metadata service to validate loader version $LoaderVersion; continuing."
        } elseif ($resolved -ne $LoaderVersion) {
            Write-Host "Resolved $Loader loader version: $resolved (Minecraft $McVersion)" -ForegroundColor Cyan
            $LoaderVersion = $resolved
        }
    }
}

# Sanity-check the Minecraft version against Mojang's manifest (warn only).
if ($McVersion) {
    $mojang = Invoke-MetaApi 'https://piston-meta.mojang.com/mc/game/version_manifest_v2.json'
    if ($mojang -and (@($mojang.versions | Where-Object { $_.id -eq $McVersion }).Count -eq 0)) {
        Write-Warning "Minecraft version '$McVersion' is not in Mojang's version manifest - check for a typo."
    }
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
# Identity comes from the loader metadata inside each jar; the filename split
# ("sodium-fabric-0.5.3.jar" -> name + version at the last hyphen-digit) is
# only the fallback. sha256 drives integrity verification; sha1 keys the
# Modrinth hash lookup below.
$mods = @()
$modsDir = Join-Path $SourceDir 'mods'
if (Test-Path $modsDir -PathType Container) {
    $modFiles = @(Get-ChildItem $modsDir -File |
        Where-Object { $_.Extension -in @('.jar', '.litemod') } | Sort-Object Name)
    $n = 0
    foreach ($modFile in $modFiles) {
        $n++
        Update-Spinner "Hashing mods ($n/$($modFiles.Count)): $($modFile.Name)"
        $base = [IO.Path]::GetFileNameWithoutExtension($modFile.Name)
        $name = $base
        $version = ''
        if ($base -match '^(.*?)[-_](v?\d[\w.+-]*)$') {
            $name = $matches[1]
            $version = $matches[2]
        }
        $meta = Get-ModArchiveMetadata -Path $modFile.FullName
        if ($meta.Name)        { $name = $meta.Name }
        elseif ($meta.Id)      { $name = $meta.Id }
        if ($meta.Version)     { $version = $meta.Version }
        $mods += [ordered]@{
            file    = $modFile.Name
            id      = $meta.Id
            name    = $name
            version = $version
            sha256  = (Get-FileHash $modFile.FullName -Algorithm SHA256).Hash.ToLower()
            sha1    = (Get-FileHash $modFile.FullName -Algorithm SHA1).Hash.ToLower()
        }
    }
    if ($modFiles.Count -gt 0) { Complete-Spinner "Hashed $($modFiles.Count) mods" }
}
Write-Host "Found $($mods.Count) mods" -ForegroundColor Cyan

# ── Modrinth source annotation ───────────────────
# One bulk POST resolves every mod by sha1 (docs.modrinth.com, version_files).
# Matched mods carry source { provider, projectId, versionId, url } so packs
# are traceable to their upstream project and the update flow can re-fetch a
# mod from Modrinth instead of shipping the bytes. CurseForge has an
# equivalent (POST /v1/fingerprints, murmur2) but requires a partner
# x-api-key - left out until a key is provisioned.
if ($mods.Count -gt 0 -and -not $Offline) {
    Update-Spinner 'Matching mods on Modrinth...'
    $body = @{
        hashes    = @($mods | ForEach-Object { $_.sha1 })
        algorithm = 'sha1'
    } | ConvertTo-Json
    $mrVersions = Invoke-MetaApi 'https://api.modrinth.com/v2/version_files' -Method Post -Body $body
    if ($mrVersions) {
        $hits = 0
        foreach ($mod in $mods) {
            $ver = $mrVersions.($mod.sha1)
            if (-not $ver) { continue }
            $verFile = @($ver.files | Where-Object { $_.hashes.sha1 -eq $mod.sha1 }) | Select-Object -First 1
            $mod.source = [ordered]@{
                provider  = 'modrinth'
                projectId = [string]$ver.project_id
                versionId = [string]$ver.id
                url       = if ($verFile) { [string]$verFile.url } else { '' }
            }
            $hits++
        }
        Complete-Spinner "Matched $hits/$($mods.Count) mods to Modrinth projects"
    } else {
        Complete-Spinner 'Modrinth lookup unavailable (offline?) - packing without sources'
    }
}

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
    'technic', 'dawn', 'bakaxl', 'sklauncher', 'freesm', 'qwertz', 'hmcl',
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
