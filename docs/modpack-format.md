# PolyForge modpack format (.polypack.zip)

> **Status: scaffold.** The layout and schemas below are the working
> structure; exact per-launcher fields and default folder locations will be
> filled in once real pack structures from the test machine are provided.

## Archive layout

A pack is a plain zip named `<id>-<version>.polypack.zip`:

```text
turtel-smp-1.0.0.polypack.zip
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

The packager also emits `<id>-<version>.manifest.json` (a standalone copy of
`pack-manifest.json`) next to the zip, so the website can host just the
manifest for update checks without clients downloading the full archive.

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
    "totalBytes": 456789
  }
}
```

**The `mods` array is the only thing used for update decisions.** The
installer compares the installed manifest against the hosted one and
computes added / removed / version-changed mods (see `ComparePackMods` in
`internal/kumi/packformat.go`). `sha256` doubles as integrity verification.

Mod `name`/`version` are currently derived from the jar filename
(best-effort). TODO: read `fabric.mod.json` / `META-INF/mods.toml` from
inside each jar for authoritative metadata.

## launchers.json

Per-launcher **info fields only** — the installer dynamically generates the
real launcher files (profiles, instance configs) from these plus the pack
manifest. The packager never ships launcher-specific files.

```json
{
  "schemaVersion": 1,
  "defaults": {
    "minMemoryMb": 2048,
    "recommendedMemoryMb": 4096,
    "javaArgs": "",
    "iconPath": ""
  },
  "launchers": {
    "vanilla":    { "profileName": "Turtel SMP" },
    "multimc":    { "instanceName": "TurtelSMP5" },
    "modrinth":   { "profileName": "TurtelSMP5" },
    "curseforge": { "instanceName": "TurtelSMP5" }
  }
}
```

What the installer generates from this (all TODO, pending real structures):

| Launcher   | Generated at install time                                   |
|------------|-------------------------------------------------------------|
| vanilla    | `launcher_profiles.json` entry (profile id, icon, args)     |
| multimc/prism-family | `instance.cfg` + `mmc-pack.json` (components from `minecraft` + `loader`) |
| modrinth   | Theseus profile entry / `profile.json`                      |
| curseforge | `minecraftinstance.json`                                    |

## Update flow (planned)

1. App fetches the hosted `<id>-<latest>.manifest.json` (URL comes from the
   remote content manifest / pack-access endpoint).
2. Compares `mods` against the locally installed manifest copy.
3. Downloads only what changed (full zip for now; delta later), replaces
   managed mods, leaves user files alone.

## Open items (waiting on test-machine pack structures)

- Default folder locations per launcher (instances dir, profiles dir).
- Exact launcher file schemas + required fields.
- Which minecraft folders beyond the defaults should ship in overrides.
- Jar metadata extraction (fabric.mod.json / mods.toml).
