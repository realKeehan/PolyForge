package kumi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ══════════════════════════════════════════════════
// Install modes
//
// How a pack install treats an existing game directory. The mode constants
// live in internal/kumi/types (the UI requests one via
// ExecutionPayload.Extra["mode"]); this file implements their semantics:
//
//   auto      — update when the dir carries our .polyforge-pack.json record,
//               plain install otherwise
//   install   — legacy lay-down: shipped files overwrite same-path files,
//               everything else is left alone (KUMI.ps1 behavior)
//   update    — sweep files the OLD pack shipped that the NEW pack no longer
//               ships (only paths from the old manifest's overrides list are
//               ever candidates — user saves/options/extra mods are not),
//               then extract
//   reinstall — repair: same sweep, then the full re-extract restores files
//               that went missing or were modified (VerifyInstalledPack
//               reports which)
//   clean     — quarantine EVERYTHING in the game dir into
//               .polyforge-quarantine\<timestamp>\ and extract a pristine
//               copy; nothing is deleted. Generalizes the original KUMI.ps1
//               custom install's "not-turtel" quarantine.
//
// The sweep + quarantine run between layout planning and extraction in
// extractAndVerifyPack; the clean quarantine covers only the game dir, not
// the launcher's instance metadata (instance.cfg etc.), which the config
// writers regenerate anyway.
// ══════════════════════════════════════════════════

// installedManifestName is the manifest copy an install leaves in its game
// dir — the record that update diffs, repair verification, and self-destruct
// all key off.
const installedManifestName = ".polyforge-pack.json"

// quarantineDirName holds cleaned-out content inside the game dir. Inert for
// every launcher (loaders only read top-level mods\), and kept out of its own
// sweep so repeated cleans nest snapshots instead of eating older ones.
const quarantineDirName = ".polyforge-quarantine"

// ParseInstallMode normalizes a UI-provided mode string. Unknown or empty
// values resolve to ModeAuto so an outdated frontend can never break installs.
func ParseInstallMode(s string) InstallMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "install", "fresh":
		return ModeInstall
	case "update", "upgrade":
		return ModeUpdate
	case "reinstall", "repair":
		return ModeReinstall
	case "clean", "clean-install", "cleaninstall":
		return ModeClean
	default:
		return ModeAuto
	}
}

// readInstalledPackManifest loads the manifest copy a previous install left
// in dir, or an error when the dir carries no record.
func readInstalledPackManifest(dir string) (*PackManifest, error) {
	data, err := os.ReadFile(filepath.Join(filepath.Clean(dir), installedManifestName))
	if err != nil {
		return nil, fmt.Errorf("no installed pack manifest here: %w", err)
	}
	return ParsePackManifest(data)
}

// resolveInstallMode turns the requested mode into the effective one for
// gameDir, returning the previous install's manifest when one exists and an
// optional note for the install log.
func resolveInstallMode(requested InstallMode, gameDir string) (InstallMode, *PackManifest, string) {
	old, err := readInstalledPackManifest(gameDir)
	if err != nil {
		old = nil
	}

	switch requested {
	case ModeAuto:
		if old != nil {
			return ModeUpdate, old, fmt.Sprintf("Found %s v%s here — updating it.", old.Name, old.Version)
		}
		return ModeInstall, nil, ""
	case ModeUpdate, ModeReinstall:
		if old == nil {
			return ModeInstall, nil, "No previous install record here — doing a fresh install instead."
		}
	}
	return requested, old, ""
}

// prepareGameDir applies the pre-extraction half of a mode's semantics.
// Extraction itself (overwriting every shipped file) is mode-independent.
func prepareGameDir(mode InstallMode, gameDir string, old, latest *PackManifest) ([]string, error) {
	switch mode {
	case ModeUpdate, ModeReinstall:
		if old == nil {
			return nil, nil
		}
		if len(old.Overrides.Files) == 0 {
			return []string{"Previous install predates per-file manifests; stale files cannot be swept and are left in place."}, nil
		}
		removed := removeStalePackFiles(gameDir, old, latest)
		if removed > 0 {
			return []string{fmt.Sprintf("Removed %d file(s) the pack no longer ships.", removed)}, nil
		}
		return nil, nil
	case ModeClean:
		moved, quarantine, err := quarantineGameDir(gameDir)
		if err != nil {
			return nil, fmt.Errorf("clean install could not quarantine the existing files: %w", err)
		}
		if moved > 0 {
			return []string{fmt.Sprintf("Moved %d existing item(s) to %s — delete that folder once you are happy with the install.", moved, quarantine)}, nil
		}
		return nil, nil
	}
	return nil, nil
}

// removeStalePackFiles deletes files the old manifest shipped that the new
// one no longer does, pruning directories the sweep leaves empty. Only paths
// from the old overrides list are candidates, so user data is untouchable by
// construction. Missing files are fine (the user may have removed them).
func removeStalePackFiles(gameDir string, old, latest *PackManifest) int {
	still := make(map[string]struct{}, len(latest.Overrides.Files))
	for _, f := range latest.Overrides.Files {
		still[strings.ToLower(f.Path)] = struct{}{}
	}

	clean := filepath.Clean(gameDir)
	removed := 0
	for _, f := range old.Overrides.Files {
		if _, ok := still[strings.ToLower(f.Path)]; ok {
			continue
		}
		path := filepath.Join(clean, filepath.FromSlash(f.Path))
		// Zip-slip-style guard, same as extraction: never step outside.
		if !strings.HasPrefix(path, clean+string(os.PathSeparator)) {
			continue
		}
		if err := os.Remove(path); err != nil {
			continue
		}
		removed++
		// Prune now-empty parents up to (never including) the game dir.
		for dir := filepath.Dir(path); dir != clean; dir = filepath.Dir(dir) {
			if os.Remove(dir) != nil {
				break
			}
		}
	}
	return removed
}

// quarantineGameDir moves every top-level entry of gameDir into a fresh
// .polyforge-quarantine\<timestamp>\ snapshot, skipping the quarantine root
// itself so earlier snapshots survive repeated cleans.
func quarantineGameDir(gameDir string) (moved int, quarantinePath string, err error) {
	clean := filepath.Clean(gameDir)
	entries, err := os.ReadDir(clean)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", nil
		}
		return 0, "", err
	}

	// Second-granularity timestamps collide when two cleans run back to back
	// (renaming onto the existing snapshot then fails as access-denied on
	// Windows), so bump a suffix until the snapshot path is fresh.
	base := filepath.Join(clean, quarantineDirName, time.Now().Format("2006-01-02_150405"))
	quarantinePath = base
	for n := 2; ; n++ {
		if _, err := os.Stat(quarantinePath); os.IsNotExist(err) {
			break
		}
		quarantinePath = fmt.Sprintf("%s-%d", base, n)
	}
	for _, e := range entries {
		if strings.EqualFold(e.Name(), quarantineDirName) {
			continue
		}
		if moved == 0 {
			if err := os.MkdirAll(quarantinePath, 0o755); err != nil {
				return 0, "", err
			}
		}
		if err := os.Rename(filepath.Join(clean, e.Name()), filepath.Join(quarantinePath, e.Name())); err != nil {
			return moved, quarantinePath, err
		}
		moved++
	}
	if moved == 0 {
		return 0, "", nil
	}
	return moved, quarantinePath, nil
}

// PlanPackUpdate diffs the pack installed under installDir against the
// latest manifest, powering "update available" checks and the update mode's
// preview. The error means "nothing is installed here".
func PlanPackUpdate(installDir string, latest *PackManifest) (PackModDiff, *PackManifest, error) {
	installed, err := readInstalledPackManifest(installDir)
	if err != nil {
		return PackModDiff{}, nil, err
	}
	return ComparePackMods(installed.Mods, latest.Mods), installed, nil
}
