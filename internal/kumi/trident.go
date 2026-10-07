package kumi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const tridentImportManifest = "trident.import.json"

// tridentInstanceKey formats the native key; PolyForge selects the destination.
func tridentInstanceKey(manifest *PackManifest, launchers *LaunchersFile) string {
	name := manifest.ID
	if launchers != nil {
		if value, ok := launchers.Launchers["polymerium"]["instanceName"].(string); ok && value != "" {
			name = value
		}
	}
	return strings.ToLower(sanitizeInstanceName(name))
}

// installTridentPack executes PolyForge's prepared plan. Native validation
// happens here, before any installation files are changed.
func installTridentPack(packPath string, plan packInstallPlan, manifest *PackManifest, launchers *LaunchersFile) ([]string, error) {
	instance, mode := plan.InstanceDir, plan.Mode
	if plan.Previous != nil {
		profilePath, err := tridentRegularPath(instance, "profile.json")
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(profilePath); err != nil {
			return nil, fmt.Errorf("managed instance has no readable profile.json: %w", err)
		}
	}
	var info map[string]any
	var defaults PackLauncherDefaults
	if launchers != nil {
		info, defaults = launchers.Launchers["polymerium"], launchers.Defaults
	}
	profile, err := tridentProfile(instance, manifest, info, defaults)
	if err != nil {
		return nil, err
	}
	reader, err := openPackReader(packPath)
	if err != nil {
		return nil, err
	}
	for _, file := range manifest.Overrides.Files {
		if err := validateTridentPath(file.Path); err != nil {
			return nil, err
		}
	}
	incoming := make(map[string]bool)
	for _, entry := range reader.File {
		if relative, ok := strings.CutPrefix(entry.Name, "overrides/"); ok && relative != "" {
			if err := validateTridentPath(strings.TrimSuffix(relative, "/")); err != nil {
				return nil, err
			}
			if !entry.FileInfo().IsDir() {
				incoming[tridentPathKey(relative)] = true
			}
		}
	}
	if err := os.MkdirAll(instance, 0o755); err != nil {
		return nil, err
	}
	work, err := os.MkdirTemp(instance, ".polyforge-install-")
	if err != nil {
		return nil, err
	}
	retainWork := false
	defer func() {
		if !retainWork {
			_ = os.RemoveAll(work)
		}
	}()
	stage := filepath.Join(work, "import")
	if err := os.Mkdir(stage, 0o755); err != nil {
		return nil, err
	}
	count, err := extractPackOverrides(reader, stage)
	if err != nil {
		return nil, err
	}
	if report := verifyFilesOnDisk(stage, manifest.Overrides.Files); !report.OK() {
		return nil, fmt.Errorf("pack integrity check failed: %v", report.Issues)
	}
	entries, err := tridentPackEntries(stage, manifest)
	if err != nil {
		return nil, err
	}
	if err := updateTridentPackages(profile, manifest, entries, mode); err != nil {
		return nil, err
	}

	imports := filepath.Join(instance, "import")
	previous, err := tridentSourceFiles(imports)
	if err != nil {
		return nil, err
	}
	owned, err := tridentImportOwnership(instance, previous)
	if err != nil {
		return nil, err
	}
	if mode == ModeInstall {
		// Explicit install merges files. Keep originals not supplied by this
		// archive, but never reintroduce a jar converted to a package entry.
		for _, relative := range previous {
			if !incoming[tridentPathKey(relative)] {
				if err := copyTridentFile(filepath.Join(imports, filepath.FromSlash(relative)), filepath.Join(stage, filepath.FromSlash(relative))); err != nil {
					return nil, err
				}
			}
		}
		for key := range owned {
			if !incoming[key] {
				delete(owned, key)
			}
		}
	}
	if err := writeJSONFile(filepath.Join(work, installedManifestName), tridentInstallRecord(manifest, entries)); err != nil {
		return nil, err
	}
	backup, keep, err := commitTridentInstall(instance, work, profile, mode, owned)
	retainWork = keep
	if err != nil {
		return nil, err
	}
	notes := []string{
		fmt.Sprintf("Installed %s v%s into Polymerium: %s (mode: %s).", manifest.Name, manifest.Version, instance, mode),
		fmt.Sprintf("Verified %d files; registered %d packages.", count, len(entries)),
		"Reopen Polymerium, then deploy the instance before launching.",
	}
	if backup != "" {
		notes = append(notes, "Previous pack files saved to "+backup)
	}
	return notes, nil
}

func tridentProfile(instance string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) (map[string]any, error) {
	if strings.TrimSpace(m.Minecraft) == "" {
		return nil, fmt.Errorf("pack %q is missing minecraft in pack-manifest.json", m.Name)
	}
	var loader any
	switch strings.ToLower(m.Loader.Type) {
	case "", "vanilla":
	case "fabric", "quilt", "forge", "neoforge":
		if strings.TrimSpace(m.Loader.Version) == "" {
			return nil, fmt.Errorf("pack %q is missing loader.version for %s in pack-manifest.json", m.Name, m.Loader.Type)
		}
		ids := map[string]string{"fabric": "net.fabricmc", "quilt": "org.quiltmc", "forge": "net.minecraftforge", "neoforge": "net.neoforged"}
		loader = ids[strings.ToLower(m.Loader.Type)] + ":" + m.Loader.Version
	default:
		return nil, fmt.Errorf("unsupported loader %q", m.Loader.Type)
	}
	profile := map[string]any{
		"name":      displayNameFrom(info, m),
		"setup":     map[string]any{"source": tridentModpackSource(m), "sourceOrders": []any{}, "packages": []any{}, "rules": []any{}},
		"overrides": map[string]any{},
	}
	path, err := tridentRegularPath(instance, "profile.json")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err == nil {
		profile = nil
		if err := json.Unmarshal(data, &profile); err != nil {
			return nil, fmt.Errorf("invalid profile.json: %w", err)
		}
		setup, ok := profile["setup"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("profile.json is missing setup")
		}
		source, _ := setup["source"].(string)
		if !isLocalTridentModpack(source) {
			return nil, fmt.Errorf("selected instance is not a local modpack")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	} else {
		overrides := profile["overrides"].(map[string]any)
		if defaults.RecommendedMemoryMB > 0 {
			overrides["java.max_memory"] = defaults.RecommendedMemoryMB
		}
		if defaults.JavaArgs != "" {
			overrides["java.additional_arguments"] = defaults.JavaArgs
		}
	}
	setup := profile["setup"].(map[string]any)
	setup["version"], setup["loader"] = m.Minecraft, loader
	overrides, ok := profile["overrides"].(map[string]any)
	if !ok {
		overrides = map[string]any{}
		profile["overrides"] = overrides
	}
	overrides["modpack.name"], overrides["modpack.version"] = m.Name, m.Version
	delete(overrides, "modpack.author") // this pack format does not declare an author
	return profile, nil
}

func validateTridentPath(relative string) error {
	if relative == "" || filepath.IsAbs(relative) || strings.ContainsAny(relative, `\:`) {
		return fmt.Errorf("invalid relative path %q", relative)
	}
	parts := strings.Split(relative, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("invalid relative path %q", relative)
		}
	}
	switch strings.ToLower(parts[0]) {
	case tridentImportManifest, "trident.persist.json", "allowed_symlinks.txt", ".trident-manifest-tmp", "natives", installedManifestName, quarantineDirName:
		return fmt.Errorf("pack uses reserved path %q", relative)
	}
	return nil
}

// tridentRegularPath permits missing components, but never follows a symlink.
func tridentRegularPath(root, relative string) (string, error) {
	path := filepath.Join(root, filepath.FromSlash(relative))
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("managed path is a symbolic link: %s", current)
		}
		if err == nil && current != path && !info.IsDir() {
			return "", fmt.Errorf("path parent is not a directory: %s", current)
		}
		if current == root {
			break
		}
		if filepath.Dir(current) == current {
			return "", fmt.Errorf("path escapes instance root: %s", path)
		}
	}
	return path, nil
}

func tridentSourceFiles(root string) ([]string, error) {
	if _, err := tridentRegularPath(root, "."); err != nil {
		return nil, err
	}
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if os.IsNotExist(err) && path == root {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("pack source is a symbolic link: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("pack source is not a regular file: %s", path)
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		if err := validateTridentPath(relative); err != nil {
			return err
		}
		paths = append(paths, relative)
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func tridentPathKey(path string) string {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.ToLower(path)
	}
	return path
}

func tridentImportOwnership(instance string, fallback []string) (map[string]string, error) {
	path, err := tridentRegularPath(filepath.Join(instance, "build"), tridentImportManifest)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	files := fallback
	if err == nil {
		var manifest struct {
			Version int      `json:"version"`
			Files   []string `json:"files"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("invalid trident.import.json: %w", err)
		}
		if manifest.Version != 1 || manifest.Files == nil {
			return nil, fmt.Errorf("unsupported trident.import.json format")
		}
		files = manifest.Files
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	owned := make(map[string]string, len(files))
	for _, relative := range files {
		if err := validateTridentPath(relative); err != nil {
			return nil, err
		}
		owned[tridentPathKey(relative)] = relative
	}
	return owned, nil
}

func tridentLiveFile(instance, relative string) (string, error) {
	root := filepath.Join(instance, "build")
	path := filepath.Join(root, filepath.FromSlash(relative))
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", nil // persist/package projections and their targets stay intact
		}
		if current == path && !info.Mode().IsRegular() {
			return "", fmt.Errorf("live file is occupied by a directory: %s", path)
		}
		if current == root {
			break
		}
	}
	return path, nil
}
