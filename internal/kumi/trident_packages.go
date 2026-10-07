package kumi

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func tridentModpackSource(m *PackManifest) string {
	name := m.Name
	if name == "" {
		name = m.ID
	}
	// Match URI component escaping, including colons and spaces in local names.
	return "modpack://local/" + strings.ReplaceAll(url.QueryEscape(name), "+", "%20")
}

func isLocalTridentModpack(source string) bool {
	u, err := url.Parse(source)
	return err == nil && u.Scheme == "modpack" && u.Host == "local" && u.User == nil &&
		u.RawQuery == "" && u.Fragment == "" && len(u.Path) > 1 && strings.Count(u.EscapedPath(), "/") == 1
}

func tridentModReference(mod PackMod) (relative, pref string) {
	if mod.Source == nil || mod.File == "" {
		return "", ""
	}
	source := mod.Source
	provider := strings.ToLower(source.Provider)
	if (provider != "modrinth" && provider != "curseforge") ||
		!tridentRepositoryID(source.ProjectID) || !tridentRepositoryID(source.VersionID) {
		return "", ""
	}
	relative = mod.File
	if !strings.HasPrefix(relative, "mods/") {
		relative = "mods/" + relative
	}
	if validateTridentPath(relative) != nil || !strings.EqualFold(filepath.Ext(relative), ".jar") {
		return "", ""
	}
	return relative, "pref://" + provider + "/" + source.ProjectID + "@" + source.VersionID
}

// A complete repository reference is useful only when its original jar exists
// in this archive. Verify before removing it from staging; incomplete/unknown
// source metadata continues to use normal file installation.
func tridentPackEntries(stage string, m *PackManifest) ([]any, error) {
	byFile := make(map[string]string)
	byProject := make(map[string]string)
	var entries []any
	for _, mod := range m.Mods {
		relative, pref := tridentModReference(mod)
		if pref == "" {
			continue
		}
		file, err := tridentRegularPath(stage, relative)
		if err != nil {
			return nil, err
		}
		info, err := os.Lstat(file)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if mod.SHA256 != "" {
			hash, _, err := hashFile(file)
			if err != nil {
				return nil, err
			}
			if !strings.EqualFold(hash, mod.SHA256) {
				return nil, fmt.Errorf("checksum mismatch for %s", relative)
			}
		}
		project := strings.SplitN(pref, "@", 2)[0]
		key := tridentPathKey(relative)
		if previous, ok := byFile[key]; ok && previous != pref {
			return nil, fmt.Errorf("conflicting package sources for %s", relative)
		}
		if previous, ok := byProject[project]; ok && previous != pref {
			return nil, fmt.Errorf("conflicting package versions for %s", project)
		}
		if _, exists := byProject[project]; !exists {
			entries = append(entries, map[string]any{"pref": pref, "enabled": true, "source": tridentModpackSource(m)})
		}
		byProject[project], byFile[key] = pref, pref
	}
	// Remove only after all candidates have been validated, including duplicate
	// metadata rows. Trident's profile remains authoritative for deployment.
	for _, relative := range m.Mods {
		file := relative.File
		if !strings.HasPrefix(file, "mods/") {
			file = "mods/" + file
		}
		if _, exists := byFile[tridentPathKey(file)]; exists {
			if err := os.Remove(filepath.Join(stage, filepath.FromSlash(file))); err != nil && !os.IsNotExist(err) {
				return nil, err
			}
		}
	}
	return entries, nil
}

func tridentRepositoryID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func updateTridentPackages(profile map[string]any, m *PackManifest, entries []any, mode InstallMode) error {
	setup := profile["setup"].(map[string]any)
	oldSource, _ := setup["source"].(string)
	if !isLocalTridentModpack(oldSource) {
		return fmt.Errorf("selected instance is not a local modpack")
	}
	newSource := tridentModpackSource(m)
	packages, ok := setup["packages"].([]any)
	if !ok && setup["packages"] != nil {
		return fmt.Errorf("profile.json has invalid packages")
	}
	// Match Trident's update merge: project identity excludes the version,
	// while the existing entry retains enabled, tags and other user metadata.
	pending := make(map[string]map[string]any, len(entries))
	for _, item := range entries {
		entry := item.(map[string]any)
		project := strings.SplitN(entry["pref"].(string), "@", 2)[0]
		pending[project] = entry
	}
	merged := make([]any, 0, len(packages)+len(entries))
	for _, item := range packages {
		entry, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("profile.json has an invalid package entry")
		}
		if entry["source"] != oldSource {
			merged = append(merged, item)
			continue
		}
		pref, _ := entry["pref"].(string)
		project := strings.SplitN(pref, "@", 2)[0]
		if replacement, exists := pending[project]; exists {
			entry["pref"] = replacement["pref"]
			entry["source"] = newSource
			merged = append(merged, item)
			delete(pending, project)
		} else if mode == ModeInstall {
			entry["source"] = newSource
			merged = append(merged, item)
		}
	}
	// Keep incoming order for newly added projects.
	for _, item := range entries {
		entry := item.(map[string]any)
		project := strings.SplitN(entry["pref"].(string), "@", 2)[0]
		if _, exists := pending[project]; exists {
			merged = append(merged, item)
			delete(pending, project)
		}
	}
	orders, ok := setup["sourceOrders"].([]any)
	if !ok && setup["sourceOrders"] != nil {
		return fmt.Errorf("profile.json has invalid sourceOrders")
	}
	for i, source := range orders {
		if source == oldSource {
			orders[i] = newSource
		}
	}
	setup["source"], setup["packages"] = newSource, merged
	return nil
}
