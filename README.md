# PolyForge


Keehan's Universal Modpack Installer (KUMI) rebuilt as a [Wails](https://wails.io/) desktop app. The backend mirrors the original PowerShell automation in Go and the frontend is a multi-step wizard implemented with Pug, SCSS, and TypeScript.


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
│       ├── custom.go          # Custom + manual install wrappers
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
├── assets.go                  # Wails asset embedding helper
├── go.mod / go.sum            # Go module configuration
└── wails.json                 # Wails build configuration
```

## Getting started

1. Install Go 1.21+, Node 18+ and the Wails CLI.
2. Install and build the frontend bundle:
   ```bash
   cd frontend
   npm install
   npm run build
   ```
3. From the project root start the app in development mode:

   ```bash
   wails dev
   ```

The wizard guides users through accepting the licence, selecting an action, choosing the modpack and launcher, and finally streams structured logs as the backend performs the installation. Utilities for Modrinth profile cloning, executable search, and launcher profile generation are exposed through the Go service for future UI integration.
