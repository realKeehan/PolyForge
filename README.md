# PolyForge

Keehan's Universal Modpack Installer (KUMI) rebuilt as a [Wails](https://wails.io/) desktop app. The backend mirrors the original PowerShell automation in Go and the frontend is a multi-step wizard implemented with Pug, SCSS, and TypeScript.

Supported on Windows, Linux, and macOS.

## Project layout

```
.
├── cmd/
│   └── kumi/main.go           # Application entrypoint for Wails
├── internal/
│   ├── app/                   # Lifecycle bindings exposed to the frontend
│   │   ├── app.go
│   │   └── bind.go
│   └── kumi/                  # Installer domain logic (downloads, profiles, search)
│       ├── assets/            # Embedded data (launcher icon, etc.)
│       ├── install/           # Per-launcher installers and shared helpers
│       ├── types/             # Shared request/result structs
│       ├── cache.go           # Launcher detection cache (JSON persistence)
│       ├── resolver.go        # Multi-strategy launcher resolver framework
│       ├── updater.go         # Self-updater + content manifest scaffolding
│       ├── custom.go          # Custom + manual install wrappers
│       ├── detect.go          # Launcher candidate path heuristics
│       ├── fs.go              # File-system helpers shared across installers
│       ├── net.go             # HTTP download and zip extraction helpers
│       ├── launchers.go       # Switchboard into install package
│       ├── mc_profiles.go     # Minecraft launcher profile utilities
│       ├── modrinth.go        # Modrinth profile cloning helpers
│       ├── search.go          # Executable search + app enumeration
│       └── service.go         # KUMI service coordinating installs and actions
├── frontend/                  # Vite + Pug + SCSS + TypeScript UI
│   ├── package.json
│   └── src/
│       ├── app/               # Client-side state + IPC helpers
│       ├── ui/                # Wizard shell and individual screens
│       ├── templates/         # Pug partials
│       └── styles.scss        # Global styling
├── website/                   # Static marketing/docs site (polyforge.dev)
├── scripts/                   # Build and dev helper scripts
│   ├── wails-build.ps1        # Production build (npm + Go + UPX + obfuscation)
│   ├── wails-dev.ps1          # Dev mode launcher (npm + Go)
│   └── cleanup-wailsbindings.ps1
├── assets_embed.go            # Wails asset embedding helper
├── go.mod / go.sum            # Go module configuration
└── wails.json                 # Wails build configuration
```

## Prerequisites

| Tool | Minimum version | Notes |
|------|----------------|-------|
| Go | 1.21+ | Backend compilation |
| Node.js | 18+ | Frontend build toolchain |
| npm | (bundled with Node) | Dependency management |
| Wails CLI | v2 | Desktop shell + bindings |
| UPX | (optional) | Binary compression (`-UPX` flag) |

## Getting started

1. Install Go 1.21+, Node 18+, and the Wails CLI.

2. Install and build the frontend bundle:
   ```bash
   cd frontend
   npm ci
   npm run build
   ```

3. From the project root start the app in development mode:

   - macOS/Linux:
     ```bash
     wails dev
     ```

   - Windows (handles npm install, Go toolchain normalisation, and stale binding cleanup):
     ```powershell
     pwsh scripts/wails-dev.ps1
     ```

4. To produce a release build on Windows:
   ```powershell
   pwsh scripts/wails-build.ps1
   ```

### Build script options

| Flag | Description |
|------|-------------|
| `-UPX` | Compress the output binary with [UPX](https://github.com/upx/upx/releases). Requires `upx` on PATH. |
| `-Obfuscated` | Future preset for Wails v3 garble/obfuscation support. Currently a no-op that warns. |
| `-SkipFrontend` | Skip `npm ci` and `npm run build` if the frontend is already built. |

#### Examples

```powershell
# Standard build
pwsh scripts/wails-build.ps1

# Build with UPX compression
pwsh scripts/wails-build.ps1 -UPX

# Build with future obfuscation flag (warns, no-op until Wails 3)
pwsh scripts/wails-build.ps1 -Obfuscated

# Build with NSIS installer output
pwsh scripts/wails-build.ps1 -nsis

# Skip frontend rebuild (e.g. CI where frontend was built in a prior step)
pwsh scripts/wails-build.ps1 -SkipFrontend
```

### Troubleshooting

#### `assets_embed.go: pattern frontend/dist: no matching files found`

The Go compiler needs the frontend bundle to exist at `frontend/dist/` before building. The build scripts handle this automatically. If building manually or in CI, ensure you run the frontend build first:

```bash
cd frontend && npm ci && npm run build && cd ..
go build -v ./...
```

The GitHub Actions workflow (`.github/workflows/go.yml`) includes Node setup and frontend build steps.

#### Windows binding errors

If Wails reports `This version of %1 is not compatible with the version of Windows you're running` when generating bindings, the cached helper at `%TEMP%\wailsbindings.exe` is usually a stale helper compiled for the wrong architecture. The PowerShell scripts delete the cache automatically before each build. If you still encounter the error, remove the file manually and retry.

## Architecture

The wizard guides users through accepting the licence, selecting an action, choosing the modpack and launcher, and finally streams structured logs as the backend performs the installation. Utilities for Modrinth profile cloning, executable search, and launcher profile generation are exposed through the Go service for future UI integration.

### Launcher detection

PolyForge includes a multi-strategy launcher resolver that discovers installed launchers through:

1. **Cache** - previously validated paths (JSON persistence)
2. **Known paths** - common install locations per launcher
3. **Registry** - Windows uninstall keys (InstallLocation, DisplayIcon)
4. **Shell AppsFolder** - UWP/Store apps via PowerShell
5. **Running processes** - detect launchers that are currently open
6. **Start Menu shortcuts** - resolve `.lnk` targets
7. **Targeted scan** - depth-limited concurrent filesystem scan (last resort)

Users can also manually browse to select launcher paths, which are cached with highest priority.

### Self-updater (planned)

The updater architecture separates binary updates from content updates:

- **Binary updates**: checked against a version manifest; downloads, verifies (SHA256), replaces, and relaunches.
- **Content updates**: modpack manifests fetched independently so new packs appear without app updates.
- **Auth**: password-based access for private packs (no key system).

### Obfuscation roadmap

When Wails v3 releases with garble integration, the `-Obfuscated` build flag will enable code obfuscation for closed-source distribution builds. The script infrastructure is already in place.

## Distribution formats

PolyForge targets these distribution formats across platforms:

### Windows
- `.exe` (standard + NSIS installer)
- `.exe` (portable)
- `.exe` (UPX-compressed)
- `.zip` (portable archive)

### macOS
- `.app` bundle
- `.dmg` (disk image)
- `.zip` (portable)

### Linux
- AppImage
- `.deb` / `.rpm`
- `.tar.gz`
- Flatpak / Snap / AUR (community)

## License

See [LICENSE](LICENSE) for details.
