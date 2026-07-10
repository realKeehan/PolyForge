# PolyForge modpack format (.polypack)

> **Status: implemented.** Pack layout, packager metadata resolution, and
> per-launcher instance generation are in place; the per-launcher schemas
> were captured from real installs on the test machine
> (`TemporaryDetectRef/MachineTest_01`, `scripts/dump-launcher-trees.ps1`).
> Remaining gaps are listed under "Open items" at the bottom.

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
      "id": "sodium",
      "name": "Sodium",
      "version": "0.5.3",
      "sha256": "…",
      "sha1": "…",
      "source": {
        "provider": "modrinth",
        "projectId": "AANobbMI",
        "versionId": "…",
        "url": "https://cdn.modrinth.com/data/…/sodium-fabric-0.5.3.jar"
      }
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
`internal/kumi/packformat.go`). Entries are keyed by `id` — the authoritative
mod id read from the loader metadata inside each jar — falling back to `name`
for packs built before ids were emitted.

### Mod identity + source (how the fields are filled)

Both packagers read each mod's metadata from inside the archive rather than
trusting the filename:

| File inside the jar | Loader | Fields |
|---------------------|--------|--------|
| `fabric.mod.json` | Fabric | `id`, `version`, display `name` |
| `quilt.mod.json` (`quilt_loader`) | Quilt | `id`, `version`, `metadata.name` |
| `META-INF/mods.toml` / `META-INF/neoforge.mods.toml` | Forge / NeoForge | `modId`, `version` (`${file.jarVersion}` resolves via `MANIFEST.MF` `Implementation-Version`), `displayName` |
| `litemod.json` | LiteLoader (`.litemod` files, legacy) | `name` (used as id), `version` |

The filename split (`name-version.jar` at the last hyphen-digit) is only the
fallback when no metadata is readable.

`sha1` keys a bulk Modrinth lookup (`POST /v2/version_files`): mods found on
Modrinth get a `source` block (`provider`, `projectId`, `versionId`, direct
`url`), which makes packs traceable to their upstream projects and lets a
future update path re-fetch a mod from Modrinth instead of shipping the
bytes. CurseForge has an equivalent (`POST /v1/fingerprints`, murmur2) but
requires a partner `x-api-key` — planned once a key is provisioned. Mods
without a match (private/renamed jars) simply have no `source`.

### Loader version resolution

The PowerShell packager resolves and validates `loader.version` against each
loader's official metadata (`-LoaderVersion latest` — or empty — picks the
newest stable for `-McVersion`; an explicit version is validated with a
warning):

| Loader | Source of truth |
|--------|-----------------|
| Fabric | `meta.fabricmc.net/v2/versions/loader/{mc}` (stable flag) |
| Quilt | `meta.quiltmc.org/v3/versions/loader/{mc}` |
| Forge | `files.minecraftforge.net/.../promotions_slim.json` (recommended → latest), validated against the Forge Maven |
| NeoForge | `maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml` (MC `1.X.Y` → NeoForge `X.Y.*`; 1.20.1 is the legacy `net.neoforged:forge` artifact and must be passed explicitly) |
| LiteLoader / vanilla | recorded verbatim, never resolved |

`-McVersion` itself is sanity-checked against Mojang's
`piston-meta.mojang.com/mc/game/version_manifest_v2.json`. Every network step
degrades to a warning when unreachable, and `-Offline` skips them all — only
`-LoaderVersion latest` hard-requires the network.

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
where overrides go (`InstanceSubdirFor`) and a `Generate` writer
(`internal/kumi/gen_launchers.go`). At install time the app derives the
instance layout from the chosen path (`PlanInstallDirs`: launcher root →
`instances/<name>/<gamedir>`, `profiles/<name>`, ...), extracts + verifies
the overrides, then writes the launcher's own files. Every schema was
captured from a real install on the reference machine
(`TemporaryDetectRef/MachineTest_01/INSTANCES`):

| Launcher | Generated at install time |
|----------|---------------------------|
| vanilla | `launcher_profiles.json` entry (gameDir, icon, `-Xmx` args); Fabric/Quilt version JSON fetched from the loader meta API into `versions\` so the profile launches immediately (Forge/NeoForge: run their installer once) |
| MultiMC / PolyMC / UltimMC | legacy headerless `instance.cfg` + `mmc-pack.json` (components from `minecraft` + `loader`) |
| Prism / ShatteredPrism / PineconeMC (elyprism) / Fjord / Freesm | `[General]`-style `instance.cfg` + `mmc-pack.json` |
| Modrinth App (Theseus) | profile row registered directly in `app.db` (pure-Go SQLite; app-version-specific columns templated from an existing profile row) |
| CurseForge | `minecraftinstance.json` (guid, `baseModLoader`, memory; the app repairs the loader `versionJson` on first open) |
| GDLauncher Carbon | `instance.json` |
| XMCL | `instance.json` (runtime block) |
| Dawn | `profile.json` (schemaVersion 3) + `content-index.json` |
| Polymerium | `profile.json` (Trident purl-style loader id) |
| ATLauncher / Technic / QWERTZ / BakaXL / SKLauncher / HMCL | no generator (embedded-manifest format, uncaptured schema, or unsupported) — overrides still extract; the log says to add the instance manually |

Generation never fails a verified install: every writer degrades to a log
note (missing launcher state, unreachable meta API, ...) so worst case the
files are in place and the user adds the instance in the launcher UI.

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
so **both** packagers read the loader metadata from inside each jar (see
"Mod identity + source" above); the filename split is only the fallback.

## Open items

- Generators for the remaining launchers: ATLauncher (its `instance.json`
  embeds full Mojang + loader manifests, so it needs the meta APIs at install
  time), QWERTZ (capture its `profiles.json` master-list schema from the test
  machine), Technic / BakaXL / SKLauncher / HMCL (unsupported or untested).
- CurseForge fingerprint matching in the packagers (`POST /v1/fingerprints`,
  murmur2) once a partner `x-api-key` is provisioned.
