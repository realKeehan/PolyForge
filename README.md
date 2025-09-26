# PolyForge

Keehan's Universal Modpack Installer (KUMI) re-imagined as a cross-platform [Wails](https://wails.io/) desktop experience with a Pug/SCSS/TypeScript front-end and a Go backend.

## Project layout

```
.
├── app/                  # Wails binding entry-point
├── backend/installer     # Go service that performs installation logic
├── frontend/             # Vite + Pug + SCSS + TypeScript UI
├── main.go               # Wails bootstrap
└── go.mod
```

## Getting started

1. Install Go 1.21+, Node 18+ and the Wails CLI.
2. Install front-end dependencies:
   ```bash
   cd frontend
   npm install
   ```
3. Generate the front-end bundle for Wails:
   ```bash
   npm run build
   ```
4. From the project root, run the application in development mode:
   ```bash
   wails dev
   ```

The UI replicates the classic menu-driven installer with logging, launcher-specific options and helpers such as Modrinth profile cloning and executable search.
