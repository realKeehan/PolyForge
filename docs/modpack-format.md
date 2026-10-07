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

The `mods` array supplies mod-level change summaries: added, removed and
version-changed mods (see `ComparePackMods` in `internal/kumi/packformat.go`).
Entries are keyed by `id` — the authoritative mod id read from the loader
metadata inside each jar — falling back to `name` for older packs.
File removal during installation uses `overrides.files`, while extraction
replaces the files carried by the archive. Mod-level summaries and file-level
installation ownership are separate.

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
`url`). Supported launchers can register these mods with their repositories
instead of copying the bundled jars; deployment may then need network access.
CurseForge has an equivalent (`POST /v1/fingerprints`, murmur2) but
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
| Polymerium | `profile.json` (loader `identity:version`, e.g. `net.fabricmc:0.16.14`) |
| ATLauncher / Technic / QWERTZ / BakaXL / SKLauncher / HMCL | no generator (embedded-manifest format, uncaptured schema, or unsupported) — overrides still extract; the log says to add the instance manually |

The generic generation path degrades writer failures to a log
note (missing launcher state, unreachable meta API, ...) so worst case the
files are in place and the user adds the instance in the launcher UI.

## Update flow

Hosted installs download the whole `.polypack`; local installs use the selected
archive. For file-based installs, a previous `.polyforge-pack.json` record selects
an update. Update and repair remove paths listed in the old pack but absent from
the new one, then extract the shipped files and verify their checksums.
Files outside the pack's ownership remain, except for an explicitly requested
clean install. The current implementation re-extracts unchanged shipped files
too; per-file downloads and hash-based skipping are future work.

For Polymerium, PolyForge selects an existing managed instance using its install record and pack
ID, then resolves the installation mode with its normal mode resolver. Instance
planning does not read or validate launcher-native profiles. An occupied
directory without a matching install record is preserved; a new
instance uses a sanitized, lower-case key with underscores appended to avoid
collisions.

When executing an install or update, the Polymerium adapter validates the native
profile. Missing or invalid local modpack sources in a managed instance cause a
runtime installation error; they are not migrated or bypassed by allocating a
new instance. Package entries are merged by
repository and project, ignoring the version when matching: retained projects
keep their enabled state, tags and other entry settings while their references
are updated. New projects are added; retired projects are removed on update
or retained on an explicit merge install. Other package groups, instance settings
and persistent user data are preserved. Clean retains the previous import sources,
replaced ordinary runtime files, projection manifest, profile and PolyForge install record under
`.polyforge-quarantine/<timestamp>/`.

PolyForge maintains `.polyforge-pack.json` in the Polymerium instance root and
registers the instance in its usual `installed-packs.json`. The local record adds
a `launcher` field for adapter dispatch; the archive's manifest format is unchanged.
It retains the complete mod list for PolyForge's update previews and remote
removal, while its file verification list covers the files installed in `import/`.
Converted repository packages are verified by Trident. This record is outside
the import and runtime layers, and Trident does not use it to recognize updates.

PolyForge's remote removal uses the registered adapter to remove recorded import
mods and their ordinary runtime copies, or exact package entries belonging to
the recorded local modpack. Other source groups and persistent data remain.
Trident reconciles package projections on the next deployment; PolyForge does
not delete shared repository caches.

The launcher registry supplies the key format, content/record layout and native
execution callback. PolyForge owns plan selection, mode resolution, successful
install registration and completion. File-based launchers retain their existing
extraction, stale-file cleanup and profile-generation flow. The native callback
adapts that execution to Trident's import sources, package entries and projections.

### Delta updates (bandwidth)

The per-file manifest could support transferring only changed files. Two hosting
models:

- **File-level (recommended):** host each shipped file content-addressed by its
  hash (`/packs/<id>/objects/<sha256>`). The client fetches only the objects it
  doesn't already have; unchanged files (and files shared across versions or
  packs) are free. Verification is automatic — the object *is* its hash.
- **Whole-pack (today):** host the full `<id>-<version>.polypack`. The whole pack
  is downloaded and its shipped files are re-extracted. No per-object endpoint
  is required.

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
