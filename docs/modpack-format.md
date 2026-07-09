# PolyForge modpack format (.polypack)

> **Status: scaffold.** The layout and schemas below are the working
> structure; exact per-launcher fields and default folder locations will be
> filled in once real launcher trees from the test machine are provided
> (`scripts/dump-launcher-trees.ps1`).

## The .polypack container

A pack ships as `<id>-<version>.polypack` — PolyForge's branded container. It
is a standard ZIP archive wrapped by a simple, reversible transform so the
file isn't a plain openable zip and gets its own extension + double-click
handler:

```text
[0:5] "SLIME"  [5] version 0x01  [6] flags 0x00  [7] reserved
[8:]  zip bytes, out[i] = zip[i] XOR key[i%32] XOR (i & 0xFF)
      key = SHA-256("PolyForge-Slime-v1")
```

The identical transform lives in three places and is covered by a pinned
test vector (`df5758d227b17001` for 8 zero bytes) so they can't drift:

- Go: `internal/kumi/slime.go` (the app reads packs)
- PowerShell: `scripts/slime-lib.ps1` (local packager)
- PHP: `website/api/slime-lib.php` (admin online packager)

**This is obfuscation, not encryption.** The key is a constant, so anyone
with the source can reverse it — its purpose is format obscurity and
branding. Real access control for private packs is the server-side password
gate (`api/pack-access.php`), never the container.

On first launch the app registers `.polypack` with Windows (per-user registry,
no admin/drivers) so double-clicking a pack opens it in PolyForge — see
`internal/kumi/firstrun.go` + `fileassoc_windows.go`.

## Archive layout (inside the container)

Unwrapped, a pack is a zip:

```text
turtel-smp-1.0.0.polypack   (contains, once unwrapped:)
├── pack-manifest.json     identity + mod versions (drives updates)
├── launchers.json         per-launcher info fields (installer generates
│                          the actual launcher files from these)
└── overrides/             copied into the instance/.minecraft as-is
    ├── mods/
    ├── config/
    ├── resourcepacks/
    ├── shaderpacks/
    └── ...any other minecraft folders included at pack time
```

The packager also emits `<id>-<version>.manifest.json` (a standalone,
un-obfuscated copy of `pack-manifest.json`) next to the `.polypack`, so the
website can host just the manifest for update checks without clients
downloading the full pack.

## pack-manifest.json

```json
{
  "schemaVersion": 1,
  "id": "turtel-smp",
  "name": "Turtel SMP",
  "version": "1.0.0",
  "minecraft": "1.20.1",
  "loader": { "type": "quilt", "version": "0.22.0" },
  "created": "2026-07-04T00:00:00Z",
  "mods": [
    {
      "file": "sodium-fabric-0.5.3.jar",
      "name": "sodium-fabric",
      "version": "0.5.3",
      "sha256": "…"
    }
  ],
  "overrides": {
    "folders": ["mods", "config"],
    "fileCount": 123,
    "totalBytes": 456789,
    "files": [
      { "path": "mods/sodium-fabric-0.5.3.jar", "sha256": "…", "size": 456789 },
      { "path": "config/sodium.properties", "sha256": "…", "size": 42 }
    ]
  }
}
```

**The `mods` array is the only thing used for update decisions.** The
installer compares the installed manifest against the hosted one and
computes added / removed / version-changed mods (see `ComparePackMods` in
`internal/kumi/packformat.go`).

### `overrides.files` — per-file integrity

`overrides.files` lists **every** file the pack ships (path relative to
`overrides/`, forward slashes) with its SHA-256 and size. Both packagers emit
it; the installer re-hashes each extracted file against this list and refuses
the install if anything is wrong (`VerifyInstalledPack` /
`verifyFilesOnDisk` in `internal/kumi/packformat.go`). This catches
**corruption, tampering, and truncated/incomplete downloads before they cause
problems in-game.** Packs built before this field existed simply omit it and
verification is skipped (backward compatible). This same per-file hash list is
the foundation for file-level delta updates (below).

Mod `name`/`version` are currently derived from the jar filename
(best-effort). TODO: read `fabric.mod.json` / `META-INF/mods.toml` from
inside each jar for authoritative metadata.

## launchers.json — one pack, every launcher

The pack is **launcher-agnostic**: `launchers.json` carries info fields for
*every* supported launcher, and the installer generates each launcher's real
files (profiles, instance configs) from those fields + the manifest. The
packager never targets a single launcher and never ships launcher-specific
files, so the same `.polypack` installs everywhere.

```json
{
  "schemaVersion": 1,
  "defaults": { "minMemoryMb": 2048, "recommendedMemoryMb": 4096, "javaArgs": "", "iconPath": "" },
  "launchers": {
    "vanilla":       { "profileName": "...", "instanceName": "..." },
    "multimc":       { "profileName": "...", "instanceName": "..." },
    "prismlauncher": { "profileName": "...", "instanceName": "..." },
    "modrinth":      { "profileName": "...", "instanceName": "..." },
    "curseforge":    { "profileName": "...", "instanceName": "..." }
    /* ...all 21 supported launchers... */
  }
}
```

Generation is data-driven through the registry in
`internal/kumi/packformat.go` (`launcherTargets`): each launcher declares
where overrides go (`InstanceSubdirFor`) and a `Generate` writer. Overrides
are always extracted; the `Generate` writers are stubs until the real
schemas are captured. Rough plan per family:

| Launcher family | Generated at install time |
|-----------------|---------------------------|
| vanilla | `launcher_profiles.json` entry (profile id, icon, args) |
| MultiMC / PolyMC / Prism forks | `instance.cfg` + `mmc-pack.json` (components from `minecraft` + `loader`) |
| Modrinth (Theseus) | profile entry / `profile.json` |
| CurseForge | `minecraftinstance.json` |
| others | TBD from `dump-launcher-trees.ps1` output |

## Update flow

1. App fetches the hosted `<id>-<latest>.manifest.json` (URL comes from the
   remote content manifest / pack-access endpoint).
2. Diffs the latest `overrides.files` against the locally installed manifest
   copy (`.polyforge-pack.json`) **by path + sha256**:
   - unchanged (same path + hash) → left on disk, never re-downloaded,
   - changed / added → fetched and written,
   - removed → deleted.
   User files (not listed in the manifest) are never touched.
3. Every fetched file is hash-verified against the manifest before it is
   swapped into place — the same integrity guarantee as a fresh install.

### Delta updates (bandwidth)

Because step 2 already knows exactly which files changed, an update only needs
to *transfer* the changed ones. Two hosting models:

- **File-level (recommended):** host each shipped file content-addressed by its
  hash (`/packs/<id>/objects/<sha256>`). The client fetches only the objects it
  doesn't already have; unchanged files (and files shared across versions or
  packs) are free. Verification is automatic — the object *is* its hash.
- **Whole-pack (today):** host the full `<id>-<version>.polypack`. The diff
  still avoids rewriting unchanged files on disk, but the whole pack is
  downloaded. Simplest to host; no per-object endpoint.

Byte-level binary diffs (bsdiff/xdelta) are intentionally **not** used: pack
payloads are mostly jars that change wholesale between versions, so file-level
delta captures nearly all the savings without a pure-Go patch codec.

## What ships vs. what never ships

Based on analysis of a real Modrinth/Theseus profile
(`Modrinth/profiles/<name>` is a full .minecraft-style instance dir):

| Ships in overrides/ | Never ships (user data / caches) |
|---------------------|----------------------------------|
| `mods/` | `saves/`, `screenshots/`, `logs/`, `crash-reports/` |
| `config/`, `defaultconfigs/` | `journeymap/` (map cache, thousands of files) |
| `resourcepacks/`, `shaderpacks/` | `essential/`, `emotes/` (mod caches) |
| `datapacks/`, `scripts/`, `kubejs/` | `.fabric/`, `debug/`, `downloads/`, `schematics/` |
| root: `options.txt`, `servers.dat` | root: `usercache.json`, `ops.json`, `whitelist.json`, `hotbar.nbt`, `hs_err_*.log`, `replay_*.log`, `command_history.txt` |

Both packagers (the PowerShell script and the admin panel's online tool)
apply this filter automatically.

Mod filenames in the wild are too inconsistent for reliable parsing
(`entity_texture_features_26.1-fabric-7.1.jar`, `Gamma-Utils-3.0.0+mc26.1.jar`),
so the online packager reads `fabric.mod.json` / `quilt.mod.json` /
`META-INF/mods.toml` from inside each jar; the filename split is only the
fallback (and the PowerShell script's current method).

## Open items (waiting on test-machine pack structures)

- Default install locations per launcher (instances dir, profiles dir).
- Exact launcher file schemas + required fields for generation.
