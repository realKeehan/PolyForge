package kumi

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ══════════════════════════════════════════════════
// PolyForge modpack format (.polypack)
//
// Scaffold mirroring docs/modpack-format.md. Packs are built by
// scripts/package-modpack.ps1 and contain:
//   pack-manifest.json  — identity + mod versions (drives updates)
//   launchers.json      — per-launcher info fields
//   overrides/          — files copied into the instance
//
// The installer generates the actual launcher files (profiles, instance
// configs) from LaunchersFile + PackManifest at install time — see
// gen_launchers.go for the per-launcher writers and install-layout planning.
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
	Type    string `json:"type,omitempty"` // "fabric", "forge", "neoforge", "quilt", "liteloader", "vanilla"
	Version string `json:"version,omitempty"`
}

// PackMod is one mod entry. ModID is the authoritative mod id the packager
// read from the loader metadata inside the jar (fabric.mod.json /
// quilt.mod.json / mods.toml / litemod.json) and is the stable identity for
// update comparison; Name is the display name (filename-derived when no
// metadata was readable). SHA256 doubles as integrity verification, SHA1
// keys Modrinth hash lookups.
type PackMod struct {
	File    string         `json:"file"`
	ModID   string         `json:"id,omitempty"`
	Name    string         `json:"name"`
	Version string         `json:"version,omitempty"`
	SHA256  string         `json:"sha256,omitempty"`
	SHA1    string         `json:"sha1,omitempty"`
	Source  *PackModSource `json:"source,omitempty"`
}

// PackModSource records the upstream platform a mod was matched to (the
// packager resolves it by file hash), so packs are traceable to their
// projects and updates can re-fetch a mod from the platform instead of
// shipping the bytes.
type PackModSource struct {
	Provider  string `json:"provider,omitempty"`  // "modrinth" | "curseforge"
	ProjectID string `json:"projectId,omitempty"`
	VersionID string `json:"versionId,omitempty"` // curseforge: the file id
	URL       string `json:"url,omitempty"`       // direct download URL
}

// key is the identity used for update comparison: the mod id when the
// packager could read one, otherwise the name. The prefixes keep an id from
// ever colliding with a name-keyed entry from an older pack.
func (m PackMod) key() string {
	if m.ModID != "" {
		return "id:" + m.ModID
	}
	return "name:" + m.Name
}

// PackOverrides summarizes the overrides/ payload. Files lists every shipped
// file with its checksum so an install can be verified byte-for-byte against
// the manifest (and, later, updated file-by-file).
type PackOverrides struct {
	Folders    []string   `json:"folders"`
	FileCount  int        `json:"fileCount"`
	TotalBytes int64      `json:"totalBytes"`
	Files      []PackFile `json:"files,omitempty"`
}

// PackFile is one shipped file: its path relative to overrides/ (forward
// slashes), the SHA-256 of its contents, and its size. This is the authority
// for integrity verification and file-level delta updates.
type PackFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size,omitempty"`
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

// ComparePackMods diffs mod lists by mod id (falling back to name for packs
// built before ids were emitted). Version is compared first, falling back to
// hash so re-built jars with identical versions still register as changed.
// A pack whose mods gained ids since the installed copy diffs as
// removed+added once; the files on disk still reconcile by path+hash.
func ComparePackMods(installed, latest []PackMod) PackModDiff {
	var diff PackModDiff

	installedByKey := make(map[string]PackMod, len(installed))
	for _, m := range installed {
		installedByKey[m.key()] = m
	}
	latestByKey := make(map[string]PackMod, len(latest))
	for _, m := range latest {
		latestByKey[m.key()] = m
	}

	for _, m := range latest {
		old, ok := installedByKey[m.key()]
		if !ok {
			diff.Added = append(diff.Added, m)
			continue
		}
		if old.Version != m.Version || (m.SHA256 != "" && old.SHA256 != "" && old.SHA256 != m.SHA256) {
			diff.Changed = append(diff.Changed, m)
		}
	}
	for _, m := range installed {
		if _, ok := latestByKey[m.key()]; !ok {
			diff.Removed = append(diff.Removed, m)
		}
	}
	return diff
}

// ── Local pack files (manual profile mode) ───────

// PolyPackInfo is the summary shown in the UI after inspecting a local
// .polypack chosen by the user.
type PolyPackInfo struct {
	Path          string `json:"path"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Minecraft     string `json:"minecraft,omitempty"`
	LoaderType    string `json:"loaderType,omitempty"`
	LoaderVersion string `json:"loaderVersion,omitempty"`
	ModCount      int    `json:"modCount"`
}

// openPackReader reads a pack file (either a .polypack container or a plain
// zip) and returns a zip.Reader over its contents.
func openPackReader(path string) (*zip.Reader, error) {
	data, err := readPackArchive(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open pack: %w", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("pack is not a valid archive: %w", err)
	}
	return reader, nil
}

// InspectPolyPack opens a local pack and returns its manifest summary.
func InspectPolyPack(path string) (*PolyPackInfo, error) {
	reader, err := openPackReader(path)
	if err != nil {
		return nil, err
	}

	manifest, err := readZipPackManifest(reader)
	if err != nil {
		return nil, err
	}

	return &PolyPackInfo{
		Path:          path,
		ID:            manifest.ID,
		Name:          manifest.Name,
		Version:       manifest.Version,
		Minecraft:     manifest.Minecraft,
		LoaderType:    manifest.Loader.Type,
		LoaderVersion: manifest.Loader.Version,
		ModCount:      len(manifest.Mods),
	}, nil
}

func readZipPackManifest(reader *zip.Reader) (*PackManifest, error) {
	for _, f := range reader.File {
		if f.Name != "pack-manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
		rc.Close()
		if err != nil {
			return nil, err
		}
		return ParsePackManifest(data)
	}
	return nil, fmt.Errorf("not a PolyForge pack: pack-manifest.json missing")
}

// readZipLaunchersFile returns the pack's launchers.json, or nil when the
// pack doesn't carry one (older packs) or it is unreadable — launcher
// generation then falls back to manifest-derived defaults.
func readZipLaunchersFile(reader *zip.Reader) *LaunchersFile {
	for _, f := range reader.File {
		if f.Name != "launchers.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil
		}
		data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
		rc.Close()
		if err != nil {
			return nil
		}
		l, err := ParseLaunchersFile(data)
		if err != nil {
			return nil
		}
		return l
	}
	return nil
}

// ── Integrity verification ───────────────────────
//
// Every file a pack ships is listed in overrides.files with its SHA-256, so an
// install can be checked byte-for-byte against the manifest. This catches
// corruption, tampering, and incomplete/truncated downloads before they cause
// problems in-game. Packs built before per-file hashes existed simply have an
// empty Files list, and verification is skipped (Total == 0).

// IntegrityIssue is one file that failed verification.
type IntegrityIssue struct {
	Path   string `json:"path"`
	Reason string `json:"reason"` // "hash mismatch", "wrong size", "missing", "unreadable"
}

// IntegrityReport is the outcome of verifying an install against a manifest.
type IntegrityReport struct {
	Checked int              `json:"checked"` // files that matched their manifest hash
	Total   int              `json:"total"`   // files the manifest declares
	Issues  []IntegrityIssue `json:"issues,omitempty"`
}

// OK reports whether every declared file verified. A report with no declared
// files (older pack format) is trivially OK.
func (r IntegrityReport) OK() bool { return len(r.Issues) == 0 }

// verifyFilesOnDisk re-hashes each declared file under dir and compares it to
// the manifest. Missing, truncated, or altered files are reported.
func verifyFilesOnDisk(dir string, files []PackFile) IntegrityReport {
	report := IntegrityReport{Total: len(files)}
	for _, pf := range files {
		path := filepath.Join(dir, filepath.FromSlash(pf.Path))
		sum, size, err := hashFile(path)
		if err != nil {
			report.Issues = append(report.Issues, IntegrityIssue{Path: pf.Path, Reason: "missing"})
			continue
		}
		switch {
		case !strings.EqualFold(sum, pf.SHA256):
			report.Issues = append(report.Issues, IntegrityIssue{Path: pf.Path, Reason: "hash mismatch"})
		case pf.Size != 0 && size != pf.Size:
			report.Issues = append(report.Issues, IntegrityIssue{Path: pf.Path, Reason: "wrong size"})
		default:
			report.Checked++
		}
	}
	return report
}

// hashFile returns the hex SHA-256 and byte size of a file.
func hashFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// VerifyInstalledPack re-checks an existing install against the manifest copy
// (.polyforge-pack.json) left at install time. Powers a "verify / repair" pass
// and the pre-flight check before an update.
func VerifyInstalledPack(installDir string) (*PackManifest, IntegrityReport, error) {
	data, err := os.ReadFile(filepath.Join(filepath.Clean(installDir), ".polyforge-pack.json"))
	if err != nil {
		return nil, IntegrityReport{}, fmt.Errorf("no installed pack manifest here: %w", err)
	}
	manifest, err := ParsePackManifest(data)
	if err != nil {
		return nil, IntegrityReport{}, err
	}
	return manifest, verifyFilesOnDisk(filepath.Clean(installDir), manifest.Overrides.Files), nil
}

// ── Local pack install (manual profile mode) ─────

// installLocalPack extracts a pack's overrides/ into targetDir, writes the
// manifest copy used for future update diffs, and verifies every extracted
// file against the manifest's checksums. Returns counts + an integrity report.
func installLocalPack(packPath, targetDir string) (files int, manifest *PackManifest, report IntegrityReport, err error) {
	reader, err := openPackReader(packPath)
	if err != nil {
		return 0, nil, IntegrityReport{}, err
	}

	manifest, err = readZipPackManifest(reader)
	if err != nil {
		return 0, nil, IntegrityReport{}, err
	}

	cleanTarget := filepath.Clean(targetDir)
	for _, f := range reader.File {
		rel, ok := strings.CutPrefix(f.Name, "overrides/")
		if !ok || rel == "" || strings.HasSuffix(f.Name, "/") {
			continue
		}
		dest := filepath.Join(cleanTarget, filepath.FromSlash(rel))
		// Zip-slip guard: extracted paths must stay inside the target.
		if !strings.HasPrefix(dest, cleanTarget+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return files, manifest, IntegrityReport{}, err
		}
		rc, err := f.Open()
		if err != nil {
			return files, manifest, IntegrityReport{}, err
		}
		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			return files, manifest, IntegrityReport{}, err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return files, manifest, IntegrityReport{}, copyErr
		}
		files++
	}

	// Installed manifest copy — the future update check diffs against this.
	if data, jsonErr := json.MarshalIndent(manifest, "", "  "); jsonErr == nil {
		_ = os.WriteFile(filepath.Join(cleanTarget, ".polyforge-pack.json"), data, 0o644)
	}

	// Verify what actually landed on disk against the manifest's checksums.
	report = verifyFilesOnDisk(cleanTarget, manifest.Overrides.Files)
	return files, manifest, report, nil
}

// ── Dynamic per-launcher generation ──────────────
//
// A pack is launcher-agnostic: overrides/ + a manifest + info fields. The
// installer turns those into the concrete files a given launcher needs, so
// one pack installs everywhere. Generation is data-driven through this
// registry; the writers live in gen_launchers.go and follow the schemas
// captured from real installs (TemporaryDetectRef/MachineTest_01/INSTANCES).

// LauncherTarget describes where and how a launcher expects an instance.
type LauncherTarget struct {
	ID string
	// InstanceSubdir is where overrides/ land relative to the launcher's
	// instance root, e.g. "minecraft" or ".minecraft" or "" (root).
	InstanceSubdir string
	// Generate writes the launcher-specific files (profile/instance configs)
	// from the pack manifest + info fields into instanceDir, returning
	// human-readable notes for the install log. nil = not implemented
	// (overrides still get extracted; the profile is added manually).
	Generate func(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error)
}

// launcherTargets is the generation registry. Subdirs reflect the layouts
// captured from the MachineTest_01 reference dump: MultiMC and the older
// forks keep `.minecraft`, the modern Prism family writes `minecraft` (no
// dot), and several launchers use the instance root itself as the game dir.
// Launchers without a Generate writer and why:
//   - atlauncher: its instance.json embeds the full Mojang + loader version
//     manifests (~70 KB); use ATLauncher's own "Add pack" instead.
//   - technic: custom packs couldn't be produced for reference (Notes.txt).
//   - qwertz: profiles.json master-list schema not captured yet.
//   - bakaxl / hmcl / ultimmc / sklauncher: unsupported or untested on the
//     reference machine (language barrier / missing download / uses the
//     vanilla .minecraft directly).
var launcherTargets = map[string]LauncherTarget{
	"vanilla":        {ID: "vanilla", InstanceSubdir: "", Generate: genVanillaProfile}, // profile in launcher_profiles.json → chosen game dir
	"multimc":        {ID: "multimc", InstanceSubdir: ".minecraft", Generate: genMMCInstance(false)},
	"polymc":         {ID: "polymc", InstanceSubdir: ".minecraft", Generate: genMMCInstance(false)},
	"prismlauncher":  {ID: "prismlauncher", InstanceSubdir: "minecraft", Generate: genMMCInstance(true)},
	"shatteredprism": {ID: "shatteredprism", InstanceSubdir: "minecraft", Generate: genMMCInstance(true)},
	"elyprism":       {ID: "elyprism", InstanceSubdir: "minecraft", Generate: genMMCInstance(true)},
	"ultimmc":        {ID: "ultimmc", InstanceSubdir: ".minecraft", Generate: genMMCInstance(false)},
	"fjord":          {ID: "fjord", InstanceSubdir: "minecraft", Generate: genMMCInstance(true)},
	"modrinth":       {ID: "modrinth", InstanceSubdir: "", Generate: genModrinthProfile},
	"curseforge":     {ID: "curseforge", InstanceSubdir: "", Generate: genCurseForgeInstance},
	"atlauncher":     {ID: "atlauncher", InstanceSubdir: ""}, // mods/ etc. sit at the instance root
	"gdlauncher":     {ID: "gdlauncher", InstanceSubdir: "instance", Generate: genGDLauncherInstance},
	"technic":        {ID: "technic", InstanceSubdir: "bin"},
	"dawn":           {ID: "dawn", InstanceSubdir: ".minecraft", Generate: genDawnProfile},
	"bakaxl":         {ID: "bakaxl", InstanceSubdir: ""},
	"sklauncher":     {ID: "sklauncher", InstanceSubdir: ""},
	"freesm":         {ID: "freesm", InstanceSubdir: "minecraft", Generate: genMMCInstance(true)},
	"qwertz":         {ID: "qwertz", InstanceSubdir: ""}, // profiles\<name> is the game dir
	"hmcl":           {ID: "hmcl", InstanceSubdir: ""},
	"polymerium":     {ID: "polymerium", InstanceSubdir: "", Generate: genPolymeriumProfile},
	"xmcl":           {ID: "xmcl", InstanceSubdir: "", Generate: genXMCLInstance}, // instance root is the game dir (.minecraftx\instances\<name>)
}

// InstanceSubdirFor returns where a launcher expects the pack's overrides,
// relative to the instance root ("" = the root itself).
func InstanceSubdirFor(launcherID string) string {
	if t, ok := launcherTargets[launcherID]; ok {
		return t.InstanceSubdir
	}
	return ""
}

// GenerateLauncherFiles writes the launcher-specific configuration for an
// installed pack, returning notes for the install log. Returns
// (false, nil, nil) when no generator exists for the launcher — the caller
// should tell the user to add the instance manually.
func GenerateLauncherFiles(launcherID, instanceDir string, m *PackManifest, l *LaunchersFile) (generated bool, notes []string, err error) {
	target, ok := launcherTargets[launcherID]
	if !ok || target.Generate == nil {
		return false, nil, nil
	}
	var info map[string]any
	var defaults PackLauncherDefaults
	if l != nil {
		info = l.Launchers[launcherID]
		defaults = l.Defaults
	}
	notes, err = target.Generate(instanceDir, m, info, defaults)
	return true, notes, err
}

// TODO: CheckPackUpdate(installedManifest, hostedManifestURL) using
// ComparePackMods to decide whether and what to update.
