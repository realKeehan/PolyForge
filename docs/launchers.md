# Supported launchers

PolyForge aims to install the same `.polypack` across every major Minecraft
launcher. Support is rolling out in stages.

| Status | Launchers |
|--------|-----------|
| **Supported** | Vanilla Launcher, MultiMC, CurseForge, Modrinth (Theseus), Custom Path, Manual Install |
| **In progress** | Prism Launcher, ATLauncher, GDLauncher, Technic, PolyMC, Feather, BakaXL |
| **Planned** | Polymerium, X Minecraft Launcher, SK Launcher, Freesm Launcher, PineconeMC (formerly ElyPrism), ShatteredPrism, QWERTZ, Fjord Launcher, HMCL, UltimMC |

## How one pack targets every launcher

A `.polypack` is **launcher-agnostic**. It carries a `launchers.json` with info
fields for every supported launcher, and the installer generates each
launcher's real files (profiles, instance configs) from those fields plus the
pack manifest at install time. The pack never ships launcher-specific files.

See the [modpack format](modpack-format.md) for the container layout and the
per-launcher generation plan.

## Launcher detection

Installed launchers are discovered through a multi-strategy resolver:

1. **Cache** — previously validated paths (JSON persistence).
2. **Known paths** — common install locations per launcher.
3. **Registry** — Windows uninstall keys (`InstallLocation`, `DisplayIcon`).
4. **Shell AppsFolder** — UWP/Store apps via PowerShell.
5. **Running processes** — launchers that are currently open.
6. **Start Menu shortcuts** — resolving `.lnk` targets.
7. **Targeted scan** — a depth-limited concurrent filesystem scan (last resort).

Users can also browse to a launcher path manually; manual selections are cached
with the highest priority.
