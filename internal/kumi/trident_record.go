package kumi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Keep the full mod list for PolyForge's update/removal flows, but verify only
// files still installed in import/. Repository packages are verified by Trident.
func tridentInstallRecord(manifest *PackManifest, entries []any) installedPackRecord {
	record := installedPackRecord{PackManifest: *manifest, Launcher: "polymerium"}
	refs := make(map[string]bool, len(entries))
	for _, item := range entries {
		entry := item.(map[string]any)
		refs[entry["pref"].(string)] = true
	}
	converted := make(map[string]bool)
	for _, mod := range manifest.Mods {
		relative, pref := tridentModReference(mod)
		if refs[pref] {
			converted[tridentPathKey(relative)] = true
		}
	}
	record.Overrides.Files = make([]PackFile, 0, len(manifest.Overrides.Files))
	for _, file := range manifest.Overrides.Files {
		if converted[tridentPathKey(file.Path)] {
			record.Overrides.TotalBytes -= file.Size
		} else {
			record.Overrides.Files = append(record.Overrides.Files, file)
		}
	}
	record.Overrides.FileCount = max(0, record.Overrides.FileCount-len(converted))
	record.Overrides.TotalBytes = max(0, record.Overrides.TotalBytes)
	return record
}

// Remove only a recorded pack's import file or exact native package entry.
// Shared caches, persist data and packages from other sources stay untouched.
func removeTridentPackMod(instance string, manifest *PackManifest, name string) (bool, error) {
	profilePath, err := tridentRegularPath(instance, "profile.json")
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return false, err
	}
	var profile map[string]any
	if err := json.Unmarshal(data, &profile); err != nil {
		return false, err
	}
	setup, ok := profile["setup"].(map[string]any)
	if !ok {
		return false, fmt.Errorf("profile.json is missing setup")
	}
	source := tridentModpackSource(manifest)
	if setup["source"] != source {
		return false, fmt.Errorf("profile.json is no longer bound to the recorded pack")
	}
	packages, ok := setup["packages"].([]any)
	if !ok && setup["packages"] != nil {
		return false, fmt.Errorf("profile.json has invalid packages")
	}
	relative := "mods/" + name
	if err := validateTridentPath(relative); err != nil {
		return false, err
	}
	recordedFile := false
	refs := make(map[string]bool)
	for _, file := range manifest.Overrides.Files {
		if tridentPathKey(file.Path) == tridentPathKey(relative) {
			recordedFile = true
		}
	}
	for _, mod := range manifest.Mods {
		file := strings.TrimPrefix(mod.File, "mods/")
		if file == name {
			_, pref := tridentModReference(mod)
			if pref != "" {
				refs[pref] = true
			} else {
				recordedFile = true
			}
		}
	}
	if !recordedFile && len(refs) == 0 {
		return false, nil
	}
	kept := make([]any, 0, len(packages))
	removedEntry := false
	for _, item := range packages {
		entry, ok := item.(map[string]any)
		if !ok {
			return false, fmt.Errorf("profile.json has an invalid package entry")
		}
		pref, _ := entry["pref"].(string)
		if entry["source"] == source && refs[pref] {
			removedEntry = true
		} else {
			kept = append(kept, item)
		}
	}
	var importPath, livePath string
	if recordedFile {
		importPath, err = tridentRegularPath(filepath.Join(instance, "import"), relative)
		if err != nil {
			return false, err
		}
		if info, statErr := os.Lstat(importPath); statErr == nil && !info.Mode().IsRegular() {
			return false, fmt.Errorf("pack mod is not a regular file: %s", importPath)
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return false, statErr
		}
		livePath, err = tridentLiveFile(instance, relative)
		if err != nil {
			return false, err
		}
	}
	if removedEntry {
		setup["packages"] = kept
		// Same-directory promotion preserves a valid profile on write failure.
		file, err := os.CreateTemp(instance, ".polyforge-profile-")
		if err != nil {
			return false, err
		}
		temp := file.Name()
		defer os.Remove(temp)
		if err := file.Close(); err != nil {
			return false, err
		}
		if err := writeJSONFile(temp, profile); err != nil {
			return false, err
		}
		if err := os.Rename(temp, profilePath); err != nil {
			return false, err
		}
	}
	removed := removedEntry
	for _, path := range []string{importPath, livePath} {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}
		} else {
			removed = true
		}
	}
	return removed, nil
}
