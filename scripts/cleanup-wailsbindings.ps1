[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$bindingPath = Join-Path ([System.IO.Path]::GetTempPath()) 'wailsbindings.exe'
if (Test-Path $bindingPath) {
    try {
        Remove-Item $bindingPath -Force -ErrorAction Stop
        Write-Host "Removed stale Wails binding helper at $bindingPath" -ForegroundColor DarkGray
    }
    catch {
        Write-Error "Failed to remove cached wailsbindings helper at $bindingPath: $_"
        exit 1
    }
} else {
    Write-Host "No cached Wails binding helper found at $bindingPath" -ForegroundColor DarkGray
}
