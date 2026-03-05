param(
  [string]$Platform,
  [switch]$Verbose,
  [switch]$SkipFrontend
)

$ErrorActionPreference = 'Stop'

$previousEnv = @{
  GOOS = $env:GOOS
  GOARCH = $env:GOARCH
}

function Ensure-Command {
  param([string]$CommandName)
  if (-not (Get-Command $CommandName -ErrorAction SilentlyContinue)) {
    throw "Required command '$CommandName' was not found in PATH."
  }
}

function Clear-GoEnv {
  Remove-Item env:GOOS -ErrorAction SilentlyContinue
  Remove-Item env:GOARCH -ErrorAction SilentlyContinue
}

function Restore-GoEnv {
  param($Saved)

  if ($Saved.GOOS) {
    Set-Item env:GOOS $Saved.GOOS
  } else {
    Remove-Item env:GOOS -ErrorAction SilentlyContinue
  }

  if ($Saved.GOARCH) {
    Set-Item env:GOARCH $Saved.GOARCH
  } else {
    Remove-Item env:GOARCH -ErrorAction SilentlyContinue
  }
}

function Resolve-GoPlatform {
  param([string]$Specified)

  if ($Specified) {
    $parts = $Specified.Split('/', 2, [System.StringSplitOptions]::RemoveEmptyEntries)
    if ($parts.Count -ne 2) {
      throw "Platform must be in the form <goos>/<goarch>."
    }

    return [PSCustomObject]@{
      Goos = $parts[0].ToLowerInvariant()
      Goarch = $parts[1].ToLowerInvariant()
    }
  }

  $arch = $env:PROCESSOR_ARCHITECTURE
  if ([string]::IsNullOrWhiteSpace($arch) -and $env:PROCESSOR_ARCHITEW6432) {
    $arch = $env:PROCESSOR_ARCHITEW6432
  }

  if (-not $arch) {
    $arch = ''
  }

  switch ($arch.ToUpperInvariant()) {
    'AMD64' { $goArch = 'amd64' }
    'X86' { $goArch = '386' }
    'ARM64' { $goArch = 'arm64' }
    'ARM' { $goArch = 'arm' }
    default { throw "Unsupported processor architecture '$arch'. Set -Platform explicitly." }
  }

  return [PSCustomObject]@{
    Goos = 'windows'
    Goarch = $goArch
  }
}

function Set-GoEnv {
  param(
    [string]$Goos,
    [string]$Goarch
  )

  Set-Item env:GOOS $Goos
  Set-Item env:GOARCH $Goarch
}

# ── Ensure required tools ─────────────────────────
Ensure-Command -CommandName 'go'
Ensure-Command -CommandName 'wails'
Ensure-Command -CommandName 'node'
Ensure-Command -CommandName 'npm'

# ── Frontend dependencies ─────────────────────────
# Wails dev mode handles the frontend dev server itself, but we need
# node_modules present so vite can resolve imports.
if (-not $SkipFrontend) {
  Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
  Push-Location (Join-Path $PSScriptRoot '..' 'frontend')
  try {
    & npm ci
    if ($LASTEXITCODE -ne 0) { throw "npm ci failed with exit code $LASTEXITCODE" }
  }
  finally {
    Pop-Location
  }
} else {
  Write-Host "Skipping frontend install (-SkipFrontend)" -ForegroundColor DarkGray
}

# ── Go environment ────────────────────────────────
Clear-GoEnv
$resolved = Resolve-GoPlatform -Specified $Platform
Set-GoEnv -Goos $resolved.Goos -Goarch $resolved.Goarch

try {
  Remove-Item -LiteralPath (Join-Path $env:TEMP 'wailsbindings.exe') -Force -ErrorAction SilentlyContinue

  if ($Verbose) {
    Write-Host 'go env (selected):'
    & go env GOHOSTOS GOHOSTARCH GOOS GOARCH | ForEach-Object { Write-Host ('  ' + $_) }
  }

  Write-Host ("Resolved platform: {0}/{1}" -f $resolved.Goos, $resolved.Goarch)
  Write-Host 'Executing: wails dev'
  & wails dev
}
finally {
  Restore-GoEnv -Saved $previousEnv
}
