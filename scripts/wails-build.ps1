[CmdletBinding()]
param(
    [switch]$UPX,
    [switch]$Obfuscated,
    [switch]$SkipFrontend,
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsToForward
)

$ErrorActionPreference = 'Stop'

function Ensure-Command {
    param(
        [string]$CommandName
    )

    if (-not (Get-Command $CommandName -ErrorAction SilentlyContinue)) {
        throw "Required command '$CommandName' was not found in PATH."
    }
}

Ensure-Command -CommandName 'go'
Ensure-Command -CommandName 'wails'
Ensure-Command -CommandName 'node'
Ensure-Command -CommandName 'npm'

# A bare `npm` resolves to Node's npm.ps1 shim first. That shim's arg
# reconstruction is broken when npm is invoked with the `&` call operator from
# *inside* a script body (Windows PowerShell 5.1): $MyInvocation.InvocationName
# becomes "&" (length 1), so `npm ci` gets sliced to `pm ci` and npm dies with
# `Unknown command: "pm"`. Resolve npm.cmd directly to bypass the shim.
$npmCmd = Get-Command npm.cmd -ErrorAction SilentlyContinue
$npm = if ($npmCmd) { $npmCmd.Source } else { 'npm' }

# ── Frontend build ─────────────────────────────────
if (-not $SkipFrontend) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
    # Nested Join-Path: the 3-argument form is PowerShell 6+ only and throws
    # "positional parameter ... 'frontend'" under Windows PowerShell 5.1.
    Push-Location (Join-Path (Join-Path $PSScriptRoot '..') 'frontend')
    try {
        & $npm ci
        if ($LASTEXITCODE -ne 0) { throw "npm ci failed with exit code $LASTEXITCODE" }

        Write-Host "Building frontend bundle..." -ForegroundColor Cyan
        & $npm run build
        if ($LASTEXITCODE -ne 0) { throw "npm run build failed with exit code $LASTEXITCODE" }
    }
    finally {
        Pop-Location
    }
} else {
    Write-Host "Skipping frontend build (-SkipFrontend)" -ForegroundColor DarkGray
}

# ── Go environment normalisation ──────────────────
$expectedGoOS = 'windows'
$expectedGoArch = 'amd64'
$currentGoOS = (& go env GOOS).Trim()
$currentGoArch = (& go env GOARCH).Trim()

if ($currentGoOS -ne $expectedGoOS -or $currentGoArch -ne $expectedGoArch) {
    Write-Host "Normalising Go toolchain environment for Wails (GOOS=$expectedGoOS, GOARCH=$expectedGoArch)." -ForegroundColor Yellow
    $env:GOOS = $expectedGoOS
    $env:GOARCH = $expectedGoArch
} else {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
}

# ── Clean stale binding helper ────────────────────
$tempBindings = Join-Path ([System.IO.Path]::GetTempPath()) 'wailsbindings.exe'
if (Test-Path $tempBindings) {
    try {
        Remove-Item $tempBindings -ErrorAction Stop
        Write-Host "Removed stale binding helper at $tempBindings" -ForegroundColor DarkGray
    } catch {
        Write-Warning "Unable to remove existing binding helper ($tempBindings): $_"
    }
}

# ── UPX compression ──────────────────────────────
if ($UPX) {
    if (-not (Get-Command 'upx' -ErrorAction SilentlyContinue)) {
        Write-Warning "UPX requested but 'upx' was not found in PATH. The binary will NOT be compressed."
        Write-Warning "Install UPX from https://github.com/upx/upx/releases and add it to PATH."
    } else {
        Write-Host "UPX compression enabled - binary will be compressed after build." -ForegroundColor Cyan
        $ArgsToForward += '-upx'
    }
}

# ── Obfuscation (garble) ─────────────────────────
# Supported since Wails v2.x: obfuscates bound method names via garble.
# Requires garble on PATH (go install mvdan.cc/garble@latest).
#
# WARNING: this app's frontend calls bindings by their real names
# (window.go.app.App.GetMenuOptions, ...). Obfuscation garbles those names, so
# an obfuscated build fails at runtime with
#   "[ERROR] Unable to load installer options from backend."
# Leave this OFF unless/until the frontend is switched to the generated,
# obfuscation-aware wailsjs wrappers.
if ($Obfuscated) {
    Write-Warning "-------------------------------------------------------------------"
    Write-Warning "Obfuscation garbles the bound Go method names. This app's UI calls"
    Write-Warning "them directly (window.go.app.App.*), so the resulting build will show"
    Write-Warning "'Unable to load installer options from backend' on launch."
    Write-Warning "Build WITHOUT -Obfuscated unless you have a matching binding bridge."
    Write-Warning "-------------------------------------------------------------------"
    if (-not (Get-Command 'garble' -ErrorAction SilentlyContinue)) {
        Write-Warning "garble was not found in PATH. Install it first:"
        Write-Warning "  go install mvdan.cc/garble@latest"
    }
    $ArgsToForward += '-obfuscated'
}

# ── Output filename (what the website download gateway expects) ─
# The site serves the newest file in each releases/<type>/ folder and keys
# per-version stats off the filename, so a build must be named
# PolyForge-<version>-<os>-<arch>.exe (see website/releases/README.md). Read the
# version from the repo-root VERSION file; skip if the caller set -o explicitly.
if ($ArgsToForward -notcontains '-o') {
    $versionFile = Join-Path (Join-Path $PSScriptRoot '..') 'VERSION'
    $version = if (Test-Path $versionFile) { (Get-Content $versionFile -Raw).Trim() } else { '' }
    if ($version) {
        $outName = "PolyForge-$version-$expectedGoOS-$expectedGoArch.exe"
        Write-Host "Output binary: $outName (build/bin/$outName)" -ForegroundColor Cyan
        $ArgsToForward += @('-o', $outName)
    } else {
        Write-Warning "VERSION file is empty or missing; falling back to the wails.json output name (PolyForge.exe)."
    }
}

# ── Build ─────────────────────────────────────────
Write-Host "Executing: wails build $ArgsToForward" -ForegroundColor Cyan
& wails build @ArgsToForward
