# PolyForge

**Keehan's Universal Modpack Installer (KUMI)** rebuilt as a [Wails](https://wails.io/)
desktop app. The backend mirrors the original PowerShell automation in Go, and
the frontend is a multi-step wizard built with Pug, SCSS, and TypeScript.

Supported on **Windows, Linux, and macOS**.

[Download](https://polyforge.dev/downloads){ .md-button .md-button--primary }
[View on GitHub](https://github.com/realKeehan/PolyForge){ .md-button }

## What it does

PolyForge installs modpacks across many Minecraft launchers from a single
workflow. It discovers installed launchers, guides you through picking an
action, modpack, and launcher, then streams structured logs while it performs
the install.

One pack format (`.polypack`) installs everywhere: packs are launcher-agnostic,
and the installer generates each launcher's real profile/instance files at
install time.

## Documentation

<div class="grid cards" markdown>

- :material-hammer-wrench: **[Building from source](building.md)**

    Prerequisites, dev mode, and the release build script options.

- :material-rocket-launch: **[Supported launchers](launchers.md)**

    Which launchers work today and what's in progress.

- :material-package-variant-closed: **[Modpack format](modpack-format.md)**

    The `.polypack` container, manifest schema, and update flow.

</div>

## Launcher detection

PolyForge uses a multi-strategy resolver that discovers installed launchers
through a cache of validated paths, common install locations, the Windows
registry, the Store AppsFolder, running processes, Start Menu shortcuts, and a
last-resort targeted filesystem scan. Users can also browse to a launcher
manually, which is cached with the highest priority.

## Project links

- Website: <https://polyforge.dev>
- Source: <https://github.com/realKeehan/PolyForge>
