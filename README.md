# PolyForge

Keehan's Universal Modpack Installer (KUMI) rebuilt as a [Wails](https://wails.io/) desktop app. The backend mirrors the original PowerShell automation in Go and the frontend is a multi-step wizard implemented with Pug, SCSS, and TypeScript.

## Project layout

```
.
├── cmd/kumi/                  # Wails entrypoint
├── internal/
│   ├── app/                   # Lifecycle bindings exposed to the frontend
│   └── kumi/                  # Installer domain logic (downloads, profiles, search)
│       ├── assets/            # Embedded data (launcher icons, etc.)
│       └── install/           # Per-launcher installers
├── frontend/                  # Vite + Pug + SCSS + TypeScript UI
│   └── src/
│       ├── app/               # Client-side state + IPC helpers
│       ├── ui/                # Wizard shell and individual screens
│       ├── templates/         # Pug partials
│       └── styles.scss        # Global styling
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
