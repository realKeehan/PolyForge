# Release hosting

Built binaries live here on the **server only** — never commit them to git
(`*.exe` is already gitignored). Upload via cPanel File Manager or FTP.

## Folder layout

```
releases/
├── 5.6.0/
│   ├── PolyForge-5.6.0-windows-amd64.exe
│   └── SHA256SUMS.txt
├── 5.7.0/
│   └── ...
└── latest → keep old folders for rollback
```

Public URL: `https://polyforge.dev/releases/<version>/<file>`

## Publishing a release

1. **Bump the version** in both places, then build:
   - `internal/kumi/service.go` → `version` const
   - `frontend/src/app/constants.ts` → `APP_VERSION`
   - `pwsh scripts/wails-build.ps1` (output in `build/bin/`)
2. **Hash it**: `certutil -hashfile PolyForge-5.6.0-windows-amd64.exe SHA256`
3. **Upload** the exe (and hashes) to `releases/5.6.0/` on the server.
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

5. Optionally update the downloads page buttons and the security page with
   the new hashes/scan links.
