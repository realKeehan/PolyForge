# Dumps folder structures for every launcher PolyForge targets, so the pack
# format and per-launcher install generators can be finalized from real data.
#
# Run this on a test machine that has the launchers installed with at least
# one instance/profile each, then send the output folder back for analysis.
#
#   pwsh scripts/dump-launcher-trees.ps1
#   pwsh scripts/dump-launcher-trees.ps1 -MaxDepth 4 -OutDir C:\pf-trees
#
# For each launcher it records the candidate data directories, and for the
# ones that exist, a depth-limited tree plus the small config/metadata files
# that describe instances/profiles (verbatim, so their schemas are captured).
#
# Beyond the fixed candidates (mirroring internal/kumi/detect.go) it also
# mirrors the app's discovery pipeline so portable installs show up:
#   - Start Menu / taskbar / Desktop shortcuts are resolved and matched
#     against each launcher's executable names (internal/kumi/service.go
#     launcherExeNames); hits add the exe's folder as a candidate dir, but
#     only when no fixed data dir was found — exe dirs of launchers that keep
#     data elsewhere are pure Electron/JRE noise in the dump.
#   - Modrinth's app.db is located and settings.custom_dir extracted (via the
#     sqlite3 CLI when available), because profiles live under that custom
#     directory when set — not under %APPDATA%\ModrinthApp.
#   - PolyForge's own launcher_cache.json is copied into the dump.
#
# Candidate paths were validated against the MachineTest_01 reference dump
# (TemporaryDetectRef/MachineTest_01): gdlauncher_carbon, .dawn (Feather →
# Dawn rebrand), QWERTZ-Launcher, .minecraftx (XMCL) and Trident (Polymerium)
# were all missed by earlier revisions of this script.

[CmdletBinding()]
param(
    [int]$MaxDepth = 3,
    [string]$OutDir = '',
    # Extra folders to sweep for .lnk shortcuts (e.g. a test machine's
    # hand-made shortcut collection for portable launchers).
    [string[]]$ExtraShortcutRoots = @()
)

$ErrorActionPreference = 'SilentlyContinue'

$root = Split-Path -Parent $PSScriptRoot
if (-not $OutDir) {
    $OutDir = Join-Path $root ("launcher-trees-" + (Get-Date -Format 'yyyy-MM-dd_HHmm'))
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

$APPDATA  = $env:APPDATA
$LOCAL    = $env:LOCALAPPDATA
$USER     = $env:USERPROFILE
$LOCALLOW = if ($LOCAL) { Join-Path (Split-Path $LOCAL -Parent) 'LocalLow' } else { $null }

# Candidate data directories per launcher (mirrors internal/kumi/detect.go).
$launchers = [ordered]@{
    'vanilla'        = @( (Join-Path $APPDATA '.minecraft') )
    'multimc'        = @(
        (Join-Path $USER 'MultiMC'),
        (Join-Path $USER 'Desktop\MultiMC'),
        (Join-Path $USER 'Downloads\MultiMC'),
        (Join-Path $LOCAL 'MultiMC'),
        'C:\MultiMC', 'C:\Games\MultiMC', 'C:\Programs\MultiMC',
        'D:\MultiMC', 'D:\Games\MultiMC', 'D:\Programs\MultiMC'
    )
    'polymc'         = @( (Join-Path $APPDATA 'PolyMC'), (Join-Path $APPDATA 'polymc') )
    'prismlauncher'  = @( (Join-Path $APPDATA 'PrismLauncher') )
    'shatteredprism' = @( (Join-Path $APPDATA 'ShatteredPrism') )
    'elyprism'       = @( (Join-Path $APPDATA 'PineconeMC'), (Join-Path $APPDATA 'PineconeMCLauncher'), (Join-Path $APPDATA 'ElyPrism'), (Join-Path $APPDATA 'ElyPrismLauncher') )
    'ultimmc'        = @( (Join-Path $APPDATA 'UltimMC') )
    'fjord'          = @( (Join-Path $APPDATA 'FjordLauncher') )
    'modrinth'       = @( (Join-Path $APPDATA 'ModrinthApp'), (Join-Path $APPDATA 'com.modrinth.theseus') )
    'curseforge'     = @( (Join-Path $USER 'curseforge\minecraft'), (Join-Path $APPDATA 'CurseForge') )
    'atlauncher'     = @( (Join-Path $APPDATA 'ATLauncher'), 'C:\ATLauncher' )
    'gdlauncher'     = @( (Join-Path $APPDATA 'gdlauncher_carbon'), (Join-Path $APPDATA 'GDLauncher Carbon'), (Join-Path $APPDATA 'gdlauncher_next'), (Join-Path $APPDATA 'gdlauncher') )
    'technic'        = @( (Join-Path $APPDATA '.technic'), 'C:\.technic' )
    # Dawn = rebranded Feather: exe under %LOCALAPPDATA%\Dawn, profiles in
    # %APPDATA%\.dawn. Legacy Feather dirs kept for pre-rebrand installs.
    'dawn'           = @( (Join-Path $APPDATA '.dawn'), (Join-Path $LOCAL 'Dawn'), (Join-Path $APPDATA 'feather'), (Join-Path $APPDATA 'FeatherClient'), (Join-Path $LOCALLOW 'Feather') )
    'bakaxl'         = @( (Join-Path $APPDATA 'BakaXL'), 'C:\BakaXL' )
    'sklauncher'     = @( (Join-Path $APPDATA 'SKLauncher'), (Join-Path $APPDATA '.sklauncher') )
    'freesm'         = @( (Join-Path $APPDATA 'FreesmLauncher'), (Join-Path $APPDATA 'freesmlauncher') )
    # QWERTZ keeps the exe inside its data dir; instances live in profiles\.
    'qwertz'         = @( (Join-Path $APPDATA 'QWERTZ-Launcher'), (Join-Path $APPDATA 'QWERTZ'), (Join-Path $APPDATA 'qwertz') )
    'hmcl'           = @( (Join-Path $APPDATA '.hmcl'), (Join-Path $USER '.hmcl') )
    # Polymerium: settings in Roaming\Polymerium, instances in Local\Trident.
    'polymerium'     = @( (Join-Path $LOCAL 'Trident'), (Join-Path $APPDATA 'Polymerium'), (Join-Path $LOCAL 'Polymerium') )
    # XMCL: config + instances.json registry in Roaming\xmcl, instance
    # folders under %USERPROFILE%\.minecraftx\instances.
    'xmcl'           = @( (Join-Path $USER '.minecraftx'), (Join-Path $APPDATA 'xmcl'), (Join-Path $APPDATA 'X Minecraft Launcher'), (Join-Path $LOCAL 'xmcl') )
}

# Executable names per launcher (mirrors internal/kumi/service.go
# launcherExeNames). Used to match resolved shortcut targets so portable
# installs are discovered wherever they live.
$exeNames = @{
    'curseforge'     = @('CurseForge.exe')
    'modrinth'       = @('Modrinth App.exe')
    'multimc'        = @('MultiMC.exe')
    'gdlauncher'     = @('GDLauncher.exe', 'GDLauncher Carbon.exe')
    'atlauncher'     = @('ATLauncher.exe')
    'prismlauncher'  = @('prismlauncher.exe')
    'bakaxl'         = @('BakaXL.exe')
    'dawn'           = @('Dawn (Feather).exe', 'Feather Launcher.exe', 'Feather.exe')
    'technic'        = @('TechnicLauncher.exe', 'technic-launcher.exe')
    'polymc'         = @('polymc.exe')
    'sklauncher'     = @('SKlauncher.exe')
    'freesm'         = @('freesmlauncher.exe')
    'elyprism'       = @('PineconeMC.exe', 'ElyPrismLauncher.exe', 'elyprism.exe')
    'shatteredprism' = @('shatteredprism.exe')
    'qwertz'         = @('QWERTZ Launcher.exe', 'QWERTZLauncher.exe', 'QWERTZ_Launcher.exe')
    'fjord'          = @('fjordlauncher.exe')
    'hmcl'           = @('HMCL.exe')
    'ultimmc'        = @('UltimMC.exe')
    'polymerium'     = @('Polymerium.exe')
    'xmcl'           = @('xmcl.exe', 'X Minecraft Launcher.exe')
}

# Small text files worth capturing verbatim (instance/profile schemas).
$schemaNames = @(
    'instance.cfg',
    'mmc-pack.json',
    'instance.json',
    'launcher_profiles.json',
    'profile.json',
    'minecraftinstance.json',
    'manifest.json',
    'pack.json',
    'instances.json',
    'modpack.json',
    'settings.json',
    '.minecraft.json',
    'profiles.json',      # QWERTZ profile registry (launcher root)
    'content-index.json', # Dawn per-profile content index
    'data.lock.json',     # Polymerium per-instance lock
    'CoreDirectory.json', # BakaXL game-dir pointers
    'installedPacks',     # Technic installed-pack registry
    '.curseclient'        # CurseForge per-instance marker
)

# ── Shortcut discovery (mirrors internal/kumi/resolver.go) ──────────────────

# exe basename (lowercase) -> launcher id
$exeToLauncher = @{}
foreach ($id in $exeNames.Keys) {
    foreach ($exe in $exeNames[$id]) { $exeToLauncher[$exe.ToLower()] = $id }
}

$shortcutRoots = @(
    (Join-Path $USER 'AppData\Roaming\Microsoft\Internet Explorer\Quick Launch'),
    (Join-Path $APPDATA 'Microsoft\Windows\Start Menu\Programs'),
    (Join-Path $env:ProgramData 'Microsoft\Windows\Start Menu\Programs'),
    (Join-Path $USER 'Desktop'),
    # Test-machine convention: hand-made shortcuts for portable launchers
    # (MultiMC/ATLauncher/XMCL zips) that never register a Start Menu entry.
    (Join-Path $USER 'Downloads\PolyForge\SHORTCUTS')
) + $ExtraShortcutRoots | Where-Object { $_ -and (Test-Path $_ -PathType Container) }

# launcher id -> list of exe-holding dirs discovered via shortcuts
$discovered = @{}
$shortcutLog = [System.Text.StringBuilder]::new()
[void]$shortcutLog.AppendLine("Shortcut discovery ($(Get-Date -Format 'u'))")

$wsh = New-Object -ComObject WScript.Shell
foreach ($lnkRoot in $shortcutRoots) {
    $lnks = Get-ChildItem -LiteralPath $lnkRoot -Filter '*.lnk' -File -Recurse -ErrorAction SilentlyContinue
    foreach ($lnk in $lnks) {
        $target = $null
        try { $target = $wsh.CreateShortcut($lnk.FullName).TargetPath } catch {}
        if (-not $target -or -not (Test-Path -LiteralPath $target -PathType Leaf)) { continue }

        $id = $exeToLauncher[([IO.Path]::GetFileName($target)).ToLower()]
        if (-not $id) { continue }

        $dir = Split-Path -Parent $target
        if (-not $discovered[$id]) { $discovered[$id] = @() }
        if ($discovered[$id] -notcontains $dir) {
            $discovered[$id] += $dir
            [void]$shortcutLog.AppendLine("$id : $($lnk.FullName) -> $target")
        }
    }
}

# ── Modrinth custom_dir extraction ──────────────────────────────────────────
# Profiles live under settings.custom_dir (app.db) when the user relocated
# the data directory; without it they sit next to app.db. Needs the sqlite3
# CLI — record the manual query if it's unavailable.

$modrinthNotes = [System.Text.StringBuilder]::new()
$modrinthDb = @( (Join-Path $APPDATA 'ModrinthApp\app.db'), (Join-Path $APPDATA 'com.modrinth.theseus\app.db') ) |
    Where-Object { Test-Path $_ -PathType Leaf } | Select-Object -First 1

if ($modrinthDb) {
    [void]$modrinthNotes.AppendLine("app.db: $modrinthDb")
    $sqlite = Get-Command sqlite3 -ErrorAction SilentlyContinue
    if ($sqlite) {
        $customDir = (& $sqlite.Source $modrinthDb 'SELECT custom_dir FROM settings;' 2>$null |
            Select-Object -First 1)
        if ($customDir) { $customDir = $customDir.Trim() }
        if ($customDir) {
            [void]$modrinthNotes.AppendLine("custom_dir: $customDir")
            if (Test-Path $customDir -PathType Container) {
                $launchers['modrinth'] = @($customDir) + $launchers['modrinth']
            }
        } else {
            [void]$modrinthNotes.AppendLine('custom_dir: (not set - profiles live next to app.db)')
        }
    } else {
        [void]$modrinthNotes.AppendLine('custom_dir: UNKNOWN - sqlite3 CLI not on PATH.')
        [void]$modrinthNotes.AppendLine("Run manually: sqlite3 `"$modrinthDb`" `"SELECT custom_dir FROM settings;`"")
    }
} else {
    [void]$modrinthNotes.AppendLine('app.db not found - Modrinth not installed or never launched.')
}

# Fold shortcut-discovered exe dirs into the candidate tables — but only for
# launchers whose fixed data dirs all came up empty (mirrors detect.go's
# withDirDiscovery). Most launchers keep data away from the exe (Dawn,
# GDLauncher, CurseForge, the Prism family), so dumping the exe dir when the
# data dir was already found only buries the useful trees under megabytes of
# Electron/JRE noise. When nothing else was found the exe dir is still worth
# dumping: it proves the launcher is installed and, for portable installs
# (MultiMC family, ATLauncher), it *is* the data dir.
foreach ($id in $discovered.Keys) {
    $hasFixedHit = @($launchers[$id] | Where-Object { $_ -and (Test-Path $_ -PathType Container) }).Count -gt 0
    if (-not $hasFixedHit) {
        $launchers[$id] = @($launchers[$id]) + $discovered[$id]
    }
}

function Write-Tree {
    param(
        [string]$Path,
        [int]$Depth,
        [System.Text.StringBuilder]$Sb,
        [int]$Indent = 0
    )

    if ($Depth -lt 0) {
        return
    }

    $prefix = ('  ' * $Indent)

    $items = Get-ChildItem -LiteralPath $Path -Force -ErrorAction SilentlyContinue |
        Sort-Object { -not $_.PSIsContainer }, Name

    foreach ($item in $items) {
        if ($item.PSIsContainer) {
            [void]$Sb.AppendLine("$prefix[$($item.Name)]/")
            if ($Depth -gt 0) {
                Write-Tree -Path $item.FullName -Depth ($Depth - 1) -Sb $Sb -Indent ($Indent + 1)
            }
        } else {
            $size = if ($item.Length -ge 1MB) {
                "{0:N1}MB" -f ($item.Length / 1MB)
            } elseif ($item.Length -ge 1KB) {
                "{0:N0}KB" -f ($item.Length / 1KB)
            } else {
                "$($item.Length)B"
            }

            [void]$Sb.AppendLine("$prefix$($item.Name)  ($size)")
        }
    }
}

$summary = [System.Text.StringBuilder]::new()
[void]$summary.AppendLine("PolyForge launcher tree dump - $(Get-Date -Format 'u')")
[void]$summary.AppendLine("MaxDepth=$MaxDepth`n")

$foundCount = 0

foreach ($name in $launchers.Keys) {
    $existing = @(
        $launchers[$name] |
            Where-Object { $_ -and (Test-Path $_ -PathType Container) } |
            Select-Object -Unique
    )

    $viaShortcut = if ($discovered[$name]) { ' (+shortcut hit)' } else { '' }
    $status = if ($existing.Count -gt 0) { "FOUND$viaShortcut" } else { 'missing' }
    [void]$summary.AppendLine(("{0,-16} {1}" -f $name, $status))

    if ($existing.Count -eq 0) {
        continue
    }

    $foundCount++

    $sb = [System.Text.StringBuilder]::new()
    [void]$sb.AppendLine("=== $name ===")

    if ($name -eq 'modrinth' -and $modrinthNotes.Length -gt 0) {
        [void]$sb.AppendLine("`n# app.db resolution")
        [void]$sb.AppendLine($modrinthNotes.ToString())
    }

    foreach ($dir in $existing) {
        [void]$sb.AppendLine("`n# $dir")
        Write-Tree -Path $dir -Depth $MaxDepth -Sb $sb

        # Capture instance/profile schema files verbatim (small ones only).
        $schemaFiles = Get-ChildItem -LiteralPath $dir -Recurse -Depth $MaxDepth -File -Force -ErrorAction SilentlyContinue |
            Where-Object { $schemaNames -contains $_.Name -and $_.Length -lt 256KB } |
            Select-Object -First 25

        foreach ($sf in $schemaFiles) {
            [void]$sb.AppendLine("`n--- FILE: $($sf.FullName) ---")
            $content = Get-Content -LiteralPath $sf.FullName -Raw -ErrorAction SilentlyContinue

            if ($content) {
                [void]$sb.AppendLine($content)
            }
        }
    }

    [IO.File]::WriteAllText((Join-Path $OutDir "$name.txt"), $sb.ToString())
    Write-Host "  $name : dumped" -ForegroundColor Green
}

# Extra diagnostics: shortcut hits, Modrinth resolution, PolyForge's own cache.
[IO.File]::WriteAllText((Join-Path $OutDir '_shortcuts.txt'), $shortcutLog.ToString())
[IO.File]::WriteAllText((Join-Path $OutDir '_modrinth-db.txt'), $modrinthNotes.ToString())
$pfCache = Join-Path $APPDATA 'PolyForge\launcher_cache.json'
if (Test-Path $pfCache -PathType Leaf) {
    Copy-Item $pfCache (Join-Path $OutDir '_polyforge-launcher-cache.json') -Force
}

[IO.File]::WriteAllText((Join-Path $OutDir '_summary.txt'), $summary.ToString())

Write-Host ''
Write-Host "Dumped $foundCount launcher(s) to: $OutDir" -ForegroundColor Cyan
Write-Host 'Zip that folder and send it over for pack-format analysis.' -ForegroundColor Yellow
Write-Host 'Note: schema files may contain usernames/paths - review before sharing.' -ForegroundColor DarkYellow
