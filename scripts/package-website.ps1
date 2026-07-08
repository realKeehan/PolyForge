# Packages the website/ folder into a cPanel-ready zip using 7-Zip.
#
# The archive root maps directly onto public_html: upload the zip in the
# cPanel File Manager, right-click -> Extract into public_html, and it
# overwrites the deployed files in place.
#
# Runtime + admin-managed state files are excluded so a deploy never clobbers
# live data. That includes the content manifest (app version control, modpack
# overrides / self-destruct marks, disabled options, option overrides), its
# history, and the pack registry — all edited on the live server through the
# admin panel, never from the repo. Windows PowerShell 5.1 compatible.

$ErrorActionPreference = 'Stop'

$root    = Split-Path -Parent $PSScriptRoot
$webDir  = Join-Path $root 'website'
$outDir  = Join-Path $root 'build'
$stamp   = Get-Date -Format 'yyyy-MM-dd_HHmm'
$outZip  = Join-Path $outDir "polyforge-website-$stamp.zip"

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

if (-not (Test-Path $webDir)) {
    Write-Error "Website folder not found: $webDir"
    exit 1
}

# ── Refresh downloadable tool copies ─────────────
# The site serves copies of the packager scripts (repo-root scripts/ is not part
# of the deploy). Refresh them here so the download on the admin Packager tab
# always matches the source of truth.
$toolsDir = Join-Path $webDir 'tools'
if (-not (Test-Path $toolsDir)) { New-Item -ItemType Directory -Path $toolsDir | Out-Null }
foreach ($tool in @('package-modpack.ps1', 'slime-lib.ps1')) {
    $src = Join-Path $PSScriptRoot $tool
    if (Test-Path $src) { Copy-Item $src (Join-Path $toolsDir $tool) -Force }
    else { Write-Warning "Packager tool not found, skipping: $src" }
}
if (-not (Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir | Out-Null
}
if (Test-Path $outZip) {
    Remove-Item $outZip -Force
}

# ── Build the archive ────────────────────────────
# Run from inside website/ so entries sit at the archive root.
# Exclusions: live server data + the dev-only router.
Push-Location $webDir
try {
    & $sevenZip a -tzip -mx=9 $outZip * `
        '-xr!tetris-scores.json' `
        '-xr!pack-access-state.json' `
        '-xr!stats-data.json' `
        '-xr!manifest.json' `
        '-xr!manifest-history.json' `
        '-xr!packs-data.json' `
        '-xr!admin-state.json' `
        '-x!router.php'
    if ($LASTEXITCODE -ne 0) {
        Write-Error "7-Zip exited with code $LASTEXITCODE"
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}

$sizeKB = [math]::Round((Get-Item $outZip).Length / 1KB)
Write-Host ''
Write-Host "Packaged website -> $outZip ($sizeKB KB)" -ForegroundColor Green
Write-Host ''
Write-Host 'Deploy: cPanel File Manager -> upload the zip into public_html ->' -ForegroundColor Yellow
Write-Host 'right-click -> Extract. Existing files are overwritten in place.'  -ForegroundColor Yellow
