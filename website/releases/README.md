# Release hosting

Built binaries live here on the **server only** — never commit them to git
(`*.exe` is already gitignored). Upload via cPanel File Manager or FTP.

## Folder layout: one folder per download type

```text
releases/
├── windows/            newest file here = what /api/download?type=windows serves
│   ├── PolyForge-5.5.2-windows-amd64.exe   (older, kept for rollback)
│   ├── PolyForge-5.6.0-windows-amd64.exe   (newest — this one is served)
│   └── SHA256SUMS.txt                       (doc files are ignored)
├── windows-arm64/
├── linux/
├── macos/
└── jar/
```

The download gateway serves **the newest file in the type folder** (by
modified time), skipping doc files (`.md/.txt/.json/.html`). Folder names
are free-form — create whatever types you need; the URL just mirrors the
folder name.

**Stable URLs — set the downloads-page buttons once and never touch them:**

```text
/api/download?type=windows
/api/download?type=linux
/api/download?type=macos
```

Pinned/older versions remain reachable exactly:

```text
/api/download?f=windows/PolyForge-5.5.2-windows-amd64.exe
```

Every gateway download increments the homepage counter (total + per type).

## Publishing a release

1. **Bump the version** — edit the repo-root `VERSION` file (or use
   dev-menu.bat → "Set app version"), then build with
   `pwsh scripts/wails-build.ps1` (output in `build/bin/`). The Go binary
   and the frontend both pick the version up automatically.
2. **Hash it**: `certutil -hashfile PolyForge-5.6.0-windows-amd64.exe SHA256`
3. **Upload** the new build into its type folder (e.g. `releases/windows/`).
   That's it — the stable URL now serves it. Keep old files for rollback
   (delete the newest to roll back).
4. **Edit `api/manifest.json`** and upload it:

   ```json
   "app": {
     "latestVersion": "5.6.0",          ← soft update: dismissible popup
     "minSupportedVersion": "5.0.0",    ← hard update: raise this instead
     "downloadUrl": "https://polyforge.dev/downloads",
     "notes": "What changed in this release"
   }
   ```

   - **Soft update** (most releases): bump `latestVersion` only. Every older
     app shows an "Update available" dialog with a Later button.
   - **Hard update** (breaking changes): also raise `minSupportedVersion`
     to the same version. Apps below it get "Update required" with no
     dismiss option.
   - **Content-only change** (new packs, renamed options): edit the
     `modpacks` / `optionOverrides` sections instead — no release needed,
     apps pick it up on next launch.

5. Update the security page with the new hashes/scan links.
