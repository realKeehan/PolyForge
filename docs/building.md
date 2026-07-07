# Building from source

## Prerequisites

| Tool | Minimum version | Notes |
|------|-----------------|-------|
| Go | 1.21+ | Backend compilation |
| Node.js | 18+ | Frontend build toolchain |
| npm | (bundled with Node) | Dependency management |
| Wails CLI | v2 | Desktop shell + bindings |
| UPX | (optional) | Binary compression (`-UPX` flag) |
| garble | (optional) | Bound-method obfuscation (`-Obfuscated` flag) |

## Getting started

1. Install Go 1.21+, Node 18+, and the Wails CLI.

2. Install and build the frontend bundle:

    ```bash
    cd frontend
    npm ci
    npm run build
    ```

3. From the project root, start the app in development mode:

    === "Windows"

        ```powershell
        pwsh scripts/wails-dev.ps1
        ```

        Handles npm install, Go toolchain normalisation, and stale binding cleanup.

    === "macOS / Linux"

        ```bash
        wails dev
        ```

4. To produce a release build on Windows:

    ```powershell
    pwsh scripts/wails-build.ps1
    ```

!!! tip "Dev menu"
    `dev-menu.bat` wraps the common tasks (dev mode, release build with
    interactive flag selection, website localhost, packaging, version bump).

## Build script options

| Flag | Description |
|------|-------------|
| `-UPX` | Compress the output binary with [UPX](https://github.com/upx/upx/releases). Requires `upx` on PATH. |
| `-Obfuscated` | Obfuscate bound Wails methods via [garble](https://github.com/burrowers/garble) (`wails build -obfuscated`). Requires `go install mvdan.cc/garble@latest`. |
| `-SkipFrontend` | Skip `npm ci` and `npm run build` if the frontend is already built. |

Any other flags are forwarded straight to `wails build` (e.g. `-nsis`,
`-clean`, `-trimpath`, `-webview2 embed`, `-debug`).

### Examples

```powershell
# Standard build
pwsh scripts/wails-build.ps1

# UPX + obfuscation
pwsh scripts/wails-build.ps1 -UPX -Obfuscated

# NSIS installer output
pwsh scripts/wails-build.ps1 -nsis

# Skip the frontend rebuild (e.g. CI built it in a prior step)
pwsh scripts/wails-build.ps1 -SkipFrontend
```

!!! warning "Windows PowerShell 5.1"
    The build scripts are compatible with Windows PowerShell 5.1 (the default
    `powershell.exe`) as well as PowerShell 7 (`pwsh`). If you invoke them
    through the dev menu without `pwsh` installed, they fall back to 5.1.

## Troubleshooting

### `assets_embed.go: pattern frontend/dist: no matching files found`

The Go compiler needs the frontend bundle to exist at `frontend/dist/` before
building. The build scripts do this automatically. If building manually or in
CI, run the frontend build first:

```bash
cd frontend && npm ci && npm run build && cd ..
go build -v ./...
```

### Windows binding errors

If Wails reports *"This version of %1 is not compatible with the version of
Windows you're running"* when generating bindings, the cached helper at
`%TEMP%\wailsbindings.exe` is usually stale (compiled for the wrong
architecture). The PowerShell scripts delete it automatically before each
build; remove it manually and retry if the error persists.

## Distribution formats

PolyForge targets these formats across platforms (see the
[downloads page](https://polyforge.dev/downloads)):

=== "Windows"

    - `.exe` (standard + NSIS installer)
    - `.exe` (portable)
    - `.exe` (UPX-compressed)
    - `.zip` (portable archive)

=== "macOS"

    - `.app` bundle
    - `.dmg` (disk image)
    - `.zip` (portable)

=== "Linux"

    - AppImage
    - `.deb` / `.rpm`
    - `.tar.gz`
    - Flatpak / Snap / AUR (community)

Each release publishes a `SHA256SUMS.txt` alongside the builds so downloads can
be verified.
