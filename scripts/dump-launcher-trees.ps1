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

[CmdletBinding()]
param(
    [int]$MaxDepth = 3,
    [string]$OutDir = ''
)

$ErrorActionPreference = 'SilentlyContinue'

$root = Split-Path -Parent $PSScriptRoot
if (-not $OutDir) { $OutDir = Join-Path $root ("launcher-trees-" + (Get-Date -Format 'yyyy-MM-dd_HHmm')) }
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

$APPDATA = $env:APPDATA
$LOCAL   = $env:LOCALAPPDATA
$USER    = $env:USERPROFILE
$LOCALLOW = if ($LOCAL) { Join-Path (Split-Path $LOCAL -Parent) 'LocalLow' } else { $null }

# Candidate data directories per launcher (mirrors internal/kumi/detect.go).
$launchers = [ordered]@{
    'vanilla'        = @( (Join-Path $APPDATA '.minecraft') )
    'multimc'        = @( (Join-Path $USER 'MultiMC'), 'C:\MultiMC', (Join-Path $LOCAL 'MultiMC') )
    'polymc'         = @( (Join-Path $APPDATA 'PolyMC'), (Join-Path $APPDATA 'polymc') )
    'prismlauncher'  = @( (Join-Path $APPDATA 'PrismLauncher') )
    'shatteredprism' = @( (Join-Path $APPDATA 'ShatteredPrism') )
    'elyprism'       = @( (Join-Path $APPDATA 'PineconeMC'), (Join-Path $APPDATA 'PineconeMCLauncher'), (Join-Path $APPDATA 'ElyPrism'), (Join-Path $APPDATA 'ElyPrismLauncher') )
    'ultimmc'        = @( (Join-Path $APPDATA 'UltimMC') )
    'fjord'          = @( (Join-Path $APPDATA 'FjordLauncher') )
    'modrinth'       = @( (Join-Path $APPDATA 'com.modrinth.theseus'), (Join-Path $APPDATA 'ModrinthApp') )
    'curseforge'     = @( (Join-Path $USER 'curseforge\minecraft'), (Join-Path $APPDATA 'CurseForge') )
    'atlauncher'     = @( (Join-Path $APPDATA 'ATLauncher'), 'C:\ATLauncher' )
    'gdlauncher'     = @( (Join-Path $APPDATA 'gdlauncher_next'), (Join-Path $APPDATA 'gdlauncher'), (Join-Path $APPDATA 'GDLauncher Carbon') )
    'technic'        = @( (Join-Path $APPDATA '.technic'), 'C:\.technic' )
    'feather'        = @( (Join-Path $APPDATA 'feather'), (Join-Path $APPDATA 'FeatherClient'), (Join-Path $LOCALLOW 'Feather') )
    'bakaxl'         = @( (Join-Path $APPDATA 'BakaXL'), 'C:\BakaXL' )
    'sklauncher'     = @( (Join-Path $APPDATA 'SKLauncher'), (Join-Path $APPDATA '.sklauncher') )
    'freesm'         = @( (Join-Path $APPDATA 'FreesmLauncher'), (Join-Path $APPDATA 'freesmlauncher') )
    'qwertz'         = @( (Join-Path $APPDATA 'QWERTZ'), (Join-Path $APPDATA 'qwertz') )
    'hmcl'           = @( (Join-Path $APPDATA '.hmcl'), (Join-Path $USER '.hmcl') )
    'polymerium'     = @( (Join-Path $APPDATA 'Polymerium'), (Join-Path $LOCAL 'Polymerium') )
    'xmcl'           = @( (Join-Path $APPDATA 'xmcl'), (Join-Path $APPDATA 'X Minecraft Launcher'), (Join-Path $LOCAL 'xmcl') )
}

# Small text files worth capturing verbatim (instance/profile schemas).
$schemaNames = @(
    'instance.cfg', 'mmc-pack.json', 'instance.json', 'launcher_profiles.json',
    'profile.json', 'minecraftinstance.json', 'manifest.json', 'pack.json',
    'instances.json', 'modpack.json', '.minecraft.json'
)

function Write-Tree {
    param([string]$Path, [int]$Depth, [System.Text.StringBuilder]$Sb, [int]$Indent = 0)
    if ($Depth -lt 0) { return }
    $prefix = ('  ' * $Indent)
    $items = Get-ChildItem -LiteralPath $Path -Force -ErrorAction SilentlyContinue | Sort-Object { -not $_.PSIsContainer }, Name
    foreach ($item in $items) {
        if ($item.PSIsContainer) {
            [void]$Sb.AppendLine("$prefix[$($item.Name)]/")
            if ($Depth -gt 0) { Write-Tree -Path $item.FullName -Depth ($Depth - 1) -Sb $Sb -Indent ($Indent + 1) }
        } else {
            $size = if ($item.Length -ge 1MB) { "{0:N1}MB" -f ($item.Length / 1MB) }
                    elseif ($item.Length -ge 1KB) { "{0:N0}KB" -f ($item.Length / 1KB) }
                    else { "$($item.Length)B" }
            [void]$Sb.AppendLine("$prefix$($item.Name)  ($size)")
        }
    }
}

$summary = [System.Text.StringBuilder]::new()
[void]$summary.AppendLine("PolyForge launcher tree dump — $(Get-Date -Format 'u')")
[void]$summary.AppendLine("MaxDepth=$MaxDepth`n")

$foundCount = 0
foreach ($name in $launchers.Keys) {
    $existing = @($launchers[$name] | Where-Object { $_ -and (Test-Path $_ -PathType Container) })
    $status = if ($existing.Count -gt 0) { "FOUND" } else { "missing" }
    [void]$summary.AppendLine(("{0,-16} {1}" -f $name, $status))
    if ($existing.Count -eq 0) { continue }
    $foundCount++

    $sb = [System.Text.StringBuilder]::new()
    [void]$sb.AppendLine("=== $name ===")
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
            if ($content) { [void]$sb.AppendLine($content) }
        }
    }
    [IO.File]::WriteAllText((Join-Path $OutDir "$name.txt"), $sb.ToString())
    Write-Host "  $name : dumped" -ForegroundColor Green
}

[IO.File]::WriteAllText((Join-Path $OutDir '_summary.txt'), $summary.ToString())

Write-Host ''
Write-Host "Dumped $foundCount launcher(s) to: $OutDir" -ForegroundColor Cyan
Write-Host 'Zip that folder and send it over for pack-format analysis.' -ForegroundColor Yellow
Write-Host 'Note: schema files may contain usernames/paths — review before sharing.' -ForegroundColor DarkYellow
