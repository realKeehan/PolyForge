package kumi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ══════════════════════════════════════════════════
// Self-destruct: remote removal of marked mods
//
// When a pack is installed, its target folder is recorded here. On launch the
// app checks the remote manifest for a pack's `removeMods` list and deletes
// those mod files from every recorded install — the mechanism for pulling
// proprietary mods back after a pack has shipped.
//
// Safety: only exact filenames located directly inside an install's mods/
// folder are removed, only for folders that still carry our
// .polyforge-pack.json marker, and each file is removed at most once.
// ══════════════════════════════════════════════════

// installedPack records where a pack was installed so its mods can be managed
// (self-destruct, and later updates) without re-selecting the launcher.
type installedPack struct {
	ID          string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Target      string   `json:"target"`
	InstalledAt string   `json:"installedAt,omitempty"`
	Removed     []string `json:"removed,omitempty"` // mods already self-destructed here
}

type installedPacksFile struct {
	Installs []installedPack `json:"installs"`
}

func installedPacksPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "installed-packs.json"), nil
}

func loadInstalledPacks() installedPacksFile {
	var f installedPacksFile
	path, err := installedPacksPath()
	if err != nil {
		return f
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return f
	}
	_ = json.Unmarshal(data, &f)
	return f
}

func saveInstalledPacks(f installedPacksFile) {
	path, err := installedPacksPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	if data, err := json.MarshalIndent(f, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}
}

// recordInstalledPack remembers that packID was installed into target so its
// mods can be managed later. Dedupes on (id, target); a fresh install clears
// the per-target "removed" set so self-destruct can re-apply if needed.
func recordInstalledPack(id, name, version, target string) {
	if id == "" || target == "" {
		return
	}
	clean := filepath.Clean(target)
	f := loadInstalledPacks()
	for i := range f.Installs {
		if f.Installs[i].ID == id && filepath.Clean(f.Installs[i].Target) == clean {
			f.Installs[i].Name = name
			f.Installs[i].Version = version
			f.Installs[i].InstalledAt = time.Now().UTC().Format(time.RFC3339)
			f.Installs[i].Removed = nil
			saveInstalledPacks(f)
			return
		}
	}
	f.Installs = append(f.Installs, installedPack{
		ID:          id,
		Name:        name,
		Version:     version,
		Target:      clean,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	})
	saveInstalledPacks(f)
}

// applySelfDestruct deletes every mod named in each pack's manifest RemoveMods
// list from that pack's recorded installs, returning notes for the log.
func applySelfDestruct(manifest *RemoteManifest) []string {
	if manifest == nil || len(manifest.Modpacks) == 0 {
		return nil
	}
	remove := make(map[string][]string)
	for _, p := range manifest.Modpacks {
		if len(p.RemoveMods) > 0 {
			remove[p.ID] = p.RemoveMods
		}
	}
	if len(remove) == 0 {
		return nil
	}

	f := loadInstalledPacks()
	var notes []string
	changed := false
	for i := range f.Installs {
		inst := &f.Installs[i]
		wanted, ok := remove[inst.ID]
		if !ok {
			continue
		}
		// Only touch folders that are still one of our installs.
		if _, err := os.Stat(filepath.Join(inst.Target, ".polyforge-pack.json")); err != nil {
			continue
		}
		modsDir := filepath.Join(inst.Target, "mods")
		already := make(map[string]bool, len(inst.Removed))
		for _, r := range inst.Removed {
			already[r] = true
		}
		for _, name := range wanted {
			base := filepath.Base(strings.TrimSpace(name))
			if base == "" || base == "." || base == ".." || strings.ContainsAny(base, `/\`) {
				continue // never traverse out of mods/
			}
			if already[base] {
				continue
			}
			path := filepath.Join(modsDir, base)
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				// Not present (already gone / never installed) — mark handled so
				// it isn't re-checked every launch.
				inst.Removed = append(inst.Removed, base)
				already[base] = true
				changed = true
				continue
			}
			if err := os.Remove(path); err == nil {
				notes = append(notes, "Self-destruct: removed "+base+" from "+installLabel(inst))
				inst.Removed = append(inst.Removed, base)
				already[base] = true
				changed = true
			}
		}
	}
	if changed {
		saveInstalledPacks(f)
	}
	return notes
}

func installLabel(inst *installedPack) string {
	if inst.Name != "" {
		return inst.Name
	}
	return inst.ID
}

// RunSelfDestruct fetches the manifest and removes any mods marked for removal
// from recorded installs. Best-effort: offline / manifest errors fall back to
// the disk cache, and otherwise it is a no-op. Returns a note for the log.
func (s *Service) RunSelfDestruct() string {
	manifest, err := fetchRemoteManifest(s.client)
	if err != nil {
		cached, cerr := readCachedManifest()
		if cerr != nil {
			return ""
		}
		manifest = cached
	}
	notes := applySelfDestruct(manifest)
	if len(notes) == 0 {
		return ""
	}
	return strings.Join(notes, "; ")
}
