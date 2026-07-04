package kumi

import (
	"encoding/json"
	"fmt"
)

// ══════════════════════════════════════════════════
// PolyForge modpack format (.polypack.zip)
//
// Scaffold mirroring docs/modpack-format.md. Packs are built by
// scripts/package-modpack.ps1 and contain:
//   pack-manifest.json  — identity + mod versions (drives updates)
//   launchers.json      — per-launcher info fields
//   overrides/          — files copied into the instance
//
// The installer generates the actual launcher files (profiles, instance
// configs) from LaunchersFile + PackManifest at install time.
// TODO: per-launcher generators and default install locations will be
// implemented once the test-machine pack structures are provided.
// ══════════════════════════════════════════════════

// PackManifest identifies a pack and lists its mods. The Mods slice is the
// only input to update decisions.
type PackManifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Version       string        `json:"version"`
	Minecraft     string        `json:"minecraft,omitempty"`
	Loader        PackLoader    `json:"loader"`
	Created       string        `json:"created,omitempty"`
	Mods          []PackMod     `json:"mods"`
	Overrides     PackOverrides `json:"overrides"`
}

// PackLoader names the mod loader a pack targets.
type PackLoader struct {
	Type    string `json:"type,omitempty"` // "fabric", "forge", "neoforge", "quilt", "vanilla"
	Version string `json:"version,omitempty"`
}

// PackMod is one mod entry; name+version drive update comparison and
// sha256 doubles as integrity verification.
type PackMod struct {
	File    string `json:"file"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	SHA256  string `json:"sha256,omitempty"`
}

// PackOverrides summarizes the overrides/ payload.
type PackOverrides struct {
	Folders    []string `json:"folders"`
	FileCount  int      `json:"fileCount"`
	TotalBytes int64    `json:"totalBytes"`
}

// LaunchersFile carries per-launcher info fields. The installer turns these
// into real launcher files; the pack never ships launcher-specific files.
type LaunchersFile struct {
	SchemaVersion int                       `json:"schemaVersion"`
	Defaults      PackLauncherDefaults      `json:"defaults"`
	Launchers     map[string]map[string]any `json:"launchers"`
}

// PackLauncherDefaults are shared install-time settings.
type PackLauncherDefaults struct {
	MinMemoryMB         int    `json:"minMemoryMb,omitempty"`
	RecommendedMemoryMB int    `json:"recommendedMemoryMb,omitempty"`
	JavaArgs            string `json:"javaArgs,omitempty"`
	IconPath            string `json:"iconPath,omitempty"`
}

// ParsePackManifest decodes and minimally validates a pack-manifest.json.
func ParsePackManifest(data []byte) (*PackManifest, error) {
	var m PackManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid pack manifest: %w", err)
	}
	if m.ID == "" || m.Version == "" {
		return nil, fmt.Errorf("pack manifest missing id or version")
	}
	return &m, nil
}

// ParseLaunchersFile decodes a launchers.json.
func ParseLaunchersFile(data []byte) (*LaunchersFile, error) {
	var l LaunchersFile
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("invalid launchers file: %w", err)
	}
	return &l, nil
}

// ── Update comparison ────────────────────────────

// PackModDiff describes what changed between an installed pack and the
// latest hosted manifest.
type PackModDiff struct {
	Added   []PackMod // in latest, not installed
	Removed []PackMod // installed, not in latest
	Changed []PackMod // same name, different version/hash (latest entry)
}

// HasChanges reports whether an update would modify anything.
func (d PackModDiff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Changed) > 0
}

// ComparePackMods diffs mod lists by name. Version is compared first,
// falling back to hash so re-built jars with identical versions still
// register as changed.
func ComparePackMods(installed, latest []PackMod) PackModDiff {
	var diff PackModDiff

	installedByName := make(map[string]PackMod, len(installed))
	for _, m := range installed {
		installedByName[m.Name] = m
	}
	latestByName := make(map[string]PackMod, len(latest))
	for _, m := range latest {
		latestByName[m.Name] = m
	}

	for _, m := range latest {
		old, ok := installedByName[m.Name]
		if !ok {
			diff.Added = append(diff.Added, m)
			continue
		}
		if old.Version != m.Version || (m.SHA256 != "" && old.SHA256 != "" && old.SHA256 != m.SHA256) {
			diff.Changed = append(diff.Changed, m)
		}
	}
	for _, m := range installed {
		if _, ok := latestByName[m.Name]; !ok {
			diff.Removed = append(diff.Removed, m)
		}
	}
	return diff
}

// ── Installer integration stubs ──────────────────
// TODO (pending test-machine pack structures):
//   - InstallPack(zipPath, launcherID, targetDir): extract overrides/,
//     write the installed manifest copy for future update diffs, and call
//     the per-launcher generator.
//   - Per-launcher generators driven by LaunchersFile info fields:
//       vanilla    → launcher_profiles.json entry
//       multimc    → instance.cfg + mmc-pack.json (components from Loader)
//       modrinth   → Theseus profile entry
//       curseforge → minecraftinstance.json
//   - CheckPackUpdate(installedManifest, hostedManifestURL) using
//     ComparePackMods to decide whether and what to update.
