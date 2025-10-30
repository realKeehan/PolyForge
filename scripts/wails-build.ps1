[CmdletBinding()]
param(
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

$tempBindings = Join-Path ([System.IO.Path]::GetTempPath()) 'wailsbindings.exe'
if (Test-Path $tempBindings) {
    try {
        Remove-Item $tempBindings -ErrorAction Stop
        Write-Host "Removed stale binding helper at $tempBindings" -ForegroundColor DarkGray
    } catch {
        Write-Warning "Unable to remove existing binding helper ($tempBindings): $_"
    }
}

Write-Host "Executing: wails build $ArgsToForward" -ForegroundColor Cyan
& wails build @ArgsToForward
