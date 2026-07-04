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

# ── Frontend build ─────────────────────────────────
if (-not $SkipFrontend) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
    Push-Location (Join-Path $PSScriptRoot '..' 'frontend')
    try {
        & npm ci
        if ($LASTEXITCODE -ne 0) { throw "npm ci failed with exit code $LASTEXITCODE" }

        Write-Host "Building frontend bundle..." -ForegroundColor Cyan
        & npm run build
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
if ($Obfuscated) {
    Write-Host "Obfuscation enabled - bound Wails methods will be garbled." -ForegroundColor Cyan
    if (-not (Get-Command 'garble' -ErrorAction SilentlyContinue)) {
        Write-Warning "garble was not found in PATH. Install it first:"
        Write-Warning "  go install mvdan.cc/garble@latest"
    }
    $ArgsToForward += '-obfuscated'
}

# ── Build ─────────────────────────────────────────
Write-Host "Executing: wails build $ArgsToForward" -ForegroundColor Cyan
& wails build @ArgsToForward
