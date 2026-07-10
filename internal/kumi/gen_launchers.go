package kumi

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ══════════════════════════════════════════════════
// Per-launcher instance/profile file generation
//
// A .polypack is launcher-agnostic; these writers turn the pack manifest +
// launchers.json info fields into the real files each launcher needs. Every
// schema below was captured from a real install on the reference machine
// (TemporaryDetectRef/MachineTest_01/INSTANCES) — when a writer needs to
// change, re-dump a fresh instance there rather than guessing.
//
// Writers return human-readable notes (surfaced in the install log) and an
// error only for hard failures; anything best-effort (network fetches,
// missing launcher state) degrades to a note so a verified pack install is
// never failed by profile generation.
// ══════════════════════════════════════════════════

// ── Install layout planning ──────────────────────

// instancesDirName is the folder a launcher keeps its instances/profiles in,
// relative to its data root ("" = the chosen directory is used as-is).
func instancesDirName(launcherID string) string {
	switch launcherID {
	case "multimc", "polymc", "prismlauncher", "shatteredprism", "elyprism",
		"ultimmc", "fjord", "freesm",
		"gdlauncher", "xmcl", "atlauncher", "polymerium":
		return "instances"
	case "curseforge":
		return "Instances"
	case "modrinth", "dawn", "qwertz":
		return "profiles"
	case "technic":
		return "modpacks"
	}
	return ""
}

// PlanInstallDirs maps the user-chosen path (usually the detected launcher
// data root) to the instance directory and the game directory overrides are
// extracted into. The chosen path may be the launcher root, its instances
// folder, or the instance folder itself — all three land in the same place.
// Unknown launchers (and manual/custom installs) use the chosen path as-is.
func PlanInstallDirs(launcherID, chosenPath, instanceName string) (instanceDir, gameDir string) {
	chosen := filepath.Clean(chosenPath)
	name := sanitizeInstanceName(instanceName)
	sub := instancesDirName(launcherID)
	switch {
	case sub == "" || name == "":
		instanceDir = chosen
	case strings.EqualFold(filepath.Base(chosen), name):
		instanceDir = chosen // the instance folder itself was chosen
	case strings.EqualFold(filepath.Base(chosen), sub):
		instanceDir = filepath.Join(chosen, name)
	default:
		instanceDir = filepath.Join(chosen, sub, name)
	}
	if s := InstanceSubdirFor(launcherID); s != "" {
		return instanceDir, filepath.Join(instanceDir, s)
	}
	return instanceDir, instanceDir
}

// sanitizeInstanceName strips characters Windows forbids in folder names.
func sanitizeInstanceName(name string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(name) {
		if r < 0x20 || strings.ContainsRune(`<>:"/\|?*`, r) {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), " .")
}

// instanceNameFor picks the instance/profile name: the pack's per-launcher
// info field first, then the pack display name, then its id.
func instanceNameFor(launcherID string, m *PackManifest, l *LaunchersFile) string {
	if l != nil {
		if info, ok := l.Launchers[launcherID]; ok {
			if v, ok := info["instanceName"].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	if m == nil {
		return ""
	}
	if m.Name != "" {
		return m.Name
	}
	return m.ID
}

// displayNameFrom mirrors instanceNameFor for a writer that already has the
// info map in hand.
func displayNameFrom(info map[string]any, m *PackManifest) string {
	for _, key := range []string{"instanceName", "profileName"} {
		if v, ok := info[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if m.Name != "" {
		return m.Name
	}
	return m.ID
}

// loaderKindTitle is the capitalised loader name several launchers use in
// their schemas ("Fabric", "NeoForge", ...). "" for vanilla/unknown.
func loaderKindTitle(t string) string {
	switch strings.ToLower(t) {
	case "fabric":
		return "Fabric"
	case "quilt":
		return "Quilt"
	case "forge":
		return "Forge"
	case "neoforge":
		return "NeoForge"
	case "liteloader":
		return "LiteLoader"
	}
	return ""
}

// writeJSONFile marshals v with 4-space indentation (matching the captured
// launcher files) and writes it under instanceDir.
func writeJSONFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// newGUID returns a random RFC 4122 v4 UUID string (no external dependency).
func newGUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Timestamp fallback keeps the id unique enough for a local file.
		return fmt.Sprintf("00000000-0000-4000-8000-%012x", time.Now().UnixNano()&0xffffffffffff)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ── MultiMC family (MultiMC, PolyMC, Prism + forks) ─
//
// One format covers seven launchers: instance.cfg + mmc-pack.json in the
// instance root, game files in .minecraft/minecraft. Captured from the
// MachineTest_01 instances: MultiMC/PolyMC/UltimMC write a headerless cfg,
// the Prism family adds [General] + ConfigVersion.

type mmcRequire struct {
	Equals   string `json:"equals,omitempty"`
	Suggests string `json:"suggests,omitempty"`
	UID      string `json:"uid"`
}

type mmcComponent struct {
	CachedName     string       `json:"cachedName,omitempty"`
	CachedRequires []mmcRequire `json:"cachedRequires,omitempty"`
	CachedVersion  string       `json:"cachedVersion,omitempty"`
	CachedVolatile bool         `json:"cachedVolatile,omitempty"`
	DependencyOnly bool         `json:"dependencyOnly,omitempty"`
	Important      bool         `json:"important,omitempty"`
	UID            string       `json:"uid"`
	Version        string       `json:"version,omitempty"`
}

type mmcPack struct {
	Components    []mmcComponent `json:"components"`
	FormatVersion int            `json:"formatVersion"`
}

// mmcComponents builds the component list for mmc-pack.json. The launcher
// fills in library/dependency caches on first launch; uid + version is what
// actually matters.
func mmcComponents(m *PackManifest) []mmcComponent {
	mc := m.Minecraft
	comps := []mmcComponent{{
		CachedName:    "Minecraft",
		CachedVersion: mc,
		Important:     true,
		UID:           "net.minecraft",
		Version:       mc,
	}}
	requiresMC := []mmcRequire{{Equals: mc, UID: "net.minecraft"}}
	intermediary := mmcComponent{
		CachedName:     "Intermediary Mappings",
		CachedRequires: requiresMC,
		CachedVersion:  mc,
		CachedVolatile: true,
		DependencyOnly: true,
		UID:            "net.fabricmc.intermediary",
		Version:        mc,
	}
	v := m.Loader.Version
	switch strings.ToLower(m.Loader.Type) {
	case "fabric":
		comps = append(comps, intermediary, mmcComponent{
			CachedName:     "Fabric Loader",
			CachedRequires: []mmcRequire{{UID: "net.fabricmc.intermediary"}},
			CachedVersion:  v,
			UID:            "net.fabricmc.fabric-loader",
			Version:        v,
		})
	case "quilt":
		comps = append(comps, intermediary, mmcComponent{
			CachedName:     "Quilt Loader",
			CachedRequires: []mmcRequire{{UID: "net.fabricmc.intermediary"}},
			CachedVersion:  v,
			UID:            "org.quiltmc.quilt-loader",
			Version:        v,
		})
	case "forge":
		comps = append(comps, mmcComponent{
			CachedName:     "Forge",
			CachedRequires: requiresMC,
			CachedVersion:  v,
			UID:            "net.minecraftforge",
			Version:        v,
		})
	case "neoforge":
		comps = append(comps, mmcComponent{
			CachedName:     "NeoForge",
			CachedRequires: requiresMC,
			CachedVersion:  v,
			UID:            "net.neoforged",
			Version:        v,
		})
	case "liteloader":
		comps = append(comps, mmcComponent{
			CachedName:     "LiteLoader",
			CachedRequires: requiresMC,
			CachedVersion:  v,
			UID:            "com.mumfrey.liteloader",
			Version:        v,
		})
	}
	return comps
}

// genMMCInstance returns a writer for the MultiMC family. prismStyle selects
// the modern cfg header ([General] + ConfigVersion, used by Prism and its
// forks) over the legacy headerless MultiMC/PolyMC/UltimMC style.
func genMMCInstance(prismStyle bool) func(string, *PackManifest, map[string]any, PackLauncherDefaults) ([]string, error) {
	return func(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
		if m.Minecraft == "" {
			return []string{"Pack has no Minecraft version; skipped instance.cfg/mmc-pack.json — add the instance manually."}, nil
		}
		name := displayNameFrom(info, m)
		var cfg strings.Builder
		if prismStyle {
			cfg.WriteString("[General]\nConfigVersion=1.3\n")
		}
		cfg.WriteString("InstanceType=OneSix\n")
		cfg.WriteString("iconKey=default\n")
		cfg.WriteString("name=" + name + "\n")
		if !prismStyle {
			cfg.WriteString("notes=\n")
		}
		if err := os.MkdirAll(instanceDir, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(instanceDir, "instance.cfg"), []byte(cfg.String()), 0o644); err != nil {
			return nil, err
		}
		pack := mmcPack{Components: mmcComponents(m), FormatVersion: 1}
		if err := writeJSONFile(filepath.Join(instanceDir, "mmc-pack.json"), pack); err != nil {
			return nil, err
		}
		return []string{"Wrote instance.cfg + mmc-pack.json; the launcher fetches the loader on first launch."}, nil
	}
}

// ── Vanilla launcher ─────────────────────────────
//
// A profile entry in .minecraft\launcher_profiles.json pointing at the
// install dir, plus (for Fabric/Quilt) the loader's launcher version JSON
// under versions\ — fetched straight from the loader's meta API, which
// serves ready-made launcher profiles for exactly this purpose.

// vanillaVersionID is the launcher version id a loader install registers.
func vanillaVersionID(m *PackManifest) string {
	mc, v := m.Minecraft, m.Loader.Version
	switch strings.ToLower(m.Loader.Type) {
	case "fabric":
		return fmt.Sprintf("fabric-loader-%s-%s", v, mc)
	case "quilt":
		return fmt.Sprintf("quilt-loader-%s-%s", v, mc)
	case "forge":
		return fmt.Sprintf("%s-forge-%s", mc, v)
	case "neoforge":
		return "neoforge-" + v
	case "liteloader":
		// The LiteLoader installer registers "<mc>-LiteLoader<mc>"
		// regardless of the loader's own version.
		return fmt.Sprintf("%s-LiteLoader%s", mc, mc)
	}
	if mc != "" {
		return mc
	}
	return "latest-release"
}

// loaderProfileJSONURL is the meta endpoint serving a ready launcher version
// JSON for the loader, or "" when the loader has none (Forge/NeoForge ship
// installers instead).
func loaderProfileJSONURL(m *PackManifest) string {
	mc, v := m.Minecraft, m.Loader.Version
	if mc == "" || v == "" {
		return ""
	}
	switch strings.ToLower(m.Loader.Type) {
	case "fabric":
		return fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", mc, v)
	case "quilt":
		return fmt.Sprintf("https://meta.quiltmc.org/v3/versions/loader/%s/%s/profile/json", mc, v)
	}
	return ""
}

var metaHTTPClient = &http.Client{Timeout: 30 * time.Second}

// fetchMetaJSON GETs a loader-meta endpoint with the app's User-Agent.
func fetchMetaJSON(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PolyForge/"+version+" (+https://polyforge.dev)")
	resp, err := metaHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

func genVanillaProfile(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	var notes []string
	mcDir := defaultMinecraftDir()
	if mcDir == "" {
		return []string{"Could not resolve the .minecraft directory; add a launcher profile manually."}, nil
	}
	profilesPath := filepath.Join(mcDir, "launcher_profiles.json")
	if !pathExists(profilesPath) {
		return []string{"launcher_profiles.json not found — run the Minecraft Launcher once, then reinstall."}, nil
	}

	versionID := vanillaVersionID(m)

	// Ensure the loader's version JSON exists so the profile can actually
	// launch. Fabric/Quilt meta serve it directly; Forge/NeoForge need their
	// own installer run once.
	loaderType := strings.ToLower(m.Loader.Type)
	versionDir := filepath.Join(mcDir, "versions", versionID)
	versionJSON := filepath.Join(versionDir, versionID+".json")
	switch {
	case loaderType == "" || loaderType == "vanilla":
		// Plain Minecraft: the launcher provisions the version itself.
	case pathExists(versionJSON):
		notes = append(notes, fmt.Sprintf("Loader version %s is already installed.", versionID))
	default:
		if url := loaderProfileJSONURL(m); url != "" {
			if data, err := fetchMetaJSON(url); err == nil {
				if err := os.MkdirAll(versionDir, 0o755); err != nil {
					return notes, err
				}
				if err := os.WriteFile(versionJSON, data, 0o644); err != nil {
					return notes, err
				}
				notes = append(notes, fmt.Sprintf("Installed %s into the vanilla launcher (from the loader's meta service).", versionID))
			} else {
				notes = append(notes, fmt.Sprintf("Could not fetch the %s launcher profile (%v) — install the loader manually.", m.Loader.Type, err))
			}
		} else {
			notes = append(notes, fmt.Sprintf("Run the %s installer once so version %s exists in the vanilla launcher.", m.Loader.Type, versionID))
		}
	}

	javaArgs := defaults.JavaArgs
	if javaArgs == "" && defaults.RecommendedMemoryMB > 0 {
		javaArgs = fmt.Sprintf("-Xmx%dM", defaults.RecommendedMemoryMB)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	profile := map[string]any{
		"name":          displayNameFrom(info, m),
		"type":          "custom",
		"icon":          launcherIconData,
		"gameDir":       instanceDir,
		"lastVersionId": versionID,
		"created":       now,
		"lastUsed":      now,
	}
	if javaArgs != "" {
		profile["javaArgs"] = javaArgs
	}
	if err := upsertLauncherProfileEntry(profilesPath, "polyforge-"+m.ID, profile); err != nil {
		return notes, err
	}
	notes = append(notes, "Added a profile to the Minecraft Launcher.")
	return notes, nil
}

// upsertLauncherProfileEntry adds or replaces one entry in a vanilla
// launcher_profiles.json, preserving everything else in the file.
func upsertLauncherProfileEntry(path, key string, profile map[string]any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("launcher_profiles.json is not valid JSON: %w", err)
	}
	profiles, ok := data["profiles"].(map[string]any)
	if !ok {
		profiles = map[string]any{}
		data["profiles"] = profiles
	}
	profiles[key] = profile
	updated, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, updated, 0o644)
}

// ── CurseForge app ───────────────────────────────
//
// minecraftinstance.json in the instance folder (captured from
// PolyforgeCurseforgeTest). The versionJson blob the app normally embeds is
// omitted — CurseForge restores it via its own repair/scan on first open.

// cfLoaderType maps a loader to CurseForge's modloader enum.
func cfLoaderType(t string) int {
	switch strings.ToLower(t) {
	case "forge":
		return 1
	case "liteloader":
		return 3
	case "fabric":
		return 4
	case "quilt":
		return 5
	case "neoforge":
		return 6
	}
	return 0
}

func genCurseForgeInstance(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	const zeroDate = "0001-01-01T00:00:00"
	loaderType := cfLoaderType(m.Loader.Type)
	var baseModLoader map[string]any
	if loaderType != 0 && m.Loader.Version != "" {
		lname := strings.ToLower(m.Loader.Type) + "-" + m.Loader.Version
		if loaderType == 4 || loaderType == 5 {
			lname += "-" + m.Minecraft
		}
		baseModLoader = map[string]any{
			"forgeVersion":     m.Loader.Version,
			"name":             lname,
			"type":             loaderType,
			"installMethod":    loaderType,
			"latest":           false,
			"recommended":      false,
			"minecraftVersion": m.Minecraft,
		}
	}
	memory := defaults.RecommendedMemoryMB
	if memory <= 0 {
		memory = 4096
	}
	instance := map[string]any{
		"baseModLoader":                        baseModLoader,
		"isUnlocked":                           true,
		"javaArgsOverride":                     nil,
		"lastPlayed":                           zeroDate,
		"playedCount":                          0,
		"timePlayed":                           0,
		"manifest":                             nil,
		"fileDate":                             zeroDate,
		"installedModpack":                     nil,
		"projectID":                            0,
		"fileID":                               0,
		"customAuthor":                         nil,
		"modpackOverrides":                     []any{},
		"isMemoryOverride":                     false,
		"allocatedMemory":                      memory,
		"profileImagePath":                     nil,
		"groupId":                              nil,
		"isVanilla":                            loaderType == 0,
		"guid":                                 newGUID(),
		"gameTypeID":                           432,
		"installPath":                          instanceDir + string(os.PathSeparator),
		"name":                                 displayNameFrom(info, m),
		"cachedScans":                          []any{},
		"isValid":                              true,
		"lastPreviousMatchUpdate":              zeroDate,
		"lastRefreshAttempt":                   zeroDate,
		"isEnabled":                            true,
		"gameVersion":                          m.Minecraft,
		"gameVersionFlavor":                    nil,
		"gameVersionTypeId":                    nil,
		"preferenceAlternateFile":              false,
		"preferenceAutoInstallUpdates":         false,
		"preferenceDeleteOrphanedDependencies": false,
		"preferenceDeleteSavedVariables":       false,
		"preferenceReleaseType":                1,
		"preferenceModdingFolderPath":          nil,
		"installDate":                          time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z"),
		"installedAddons":                      []any{},
		"installedGamePrerequisites":           []any{},
		"wasNameManuallyChanged":               false,
	}
	if err := writeJSONFile(filepath.Join(instanceDir, "minecraftinstance.json"), instance); err != nil {
		return nil, err
	}
	return []string{"Wrote minecraftinstance.json; CurseForge provisions the loader when the profile is first opened."}, nil
}

// ── GDLauncher (Carbon) ──────────────────────────
// instance.json (captured from "Polyforge GD Launcher Test").

func genGDLauncherInstance(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	var notes []string
	modloaders := []any{}
	kind := loaderKindTitle(m.Loader.Type)
	if kind == "LiteLoader" {
		// Carbon's modloader enum is Forge/Fabric/Quilt/Neoforge only; an
		// unknown type would fail the app's instance.json parse, so a
		// LiteLoader pack installs as a plain instance instead.
		kind = ""
		notes = append(notes, "GDLauncher has no LiteLoader support; wrote a vanilla instance (mods are still in place).")
	}
	if kind != "" && m.Loader.Version != "" {
		// Carbon spells it "Neoforge" (only the first letter capitalised).
		if kind == "NeoForge" {
			kind = "Neoforge"
		}
		modloaders = append(modloaders, map[string]any{"type": kind, "version": m.Loader.Version})
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	instance := map[string]any{
		"_version":       "1",
		"name":           displayNameFrom(info, m),
		"icon":           nil,
		"created_at":     now,
		"updated_at":     now,
		"last_played":    nil,
		"seconds_played": 0,
		"modpack":        nil,
		"game_configuration": map[string]any{
			"version": map[string]any{
				"release":    m.Minecraft,
				"modloaders": modloaders,
			},
			"global_java_args": true,
		},
		"mod_sources": nil,
		"notes":       "",
	}
	if err := writeJSONFile(filepath.Join(instanceDir, "instance.json"), instance); err != nil {
		return nil, err
	}
	return append(notes, "Wrote GDLauncher instance.json."), nil
}

// ── X Minecraft Launcher ─────────────────────────
// instance.json (captured from "Polyforge XMCL Test").

func genXMCLInstance(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	runtime := map[string]any{
		"minecraft":    m.Minecraft,
		"forge":        "",
		"neoForged":    "",
		"fabricLoader": "",
		"quiltLoader":  "",
		"optifine":     "",
		"labyMod":      "",
	}
	versionLabel := m.Minecraft
	switch strings.ToLower(m.Loader.Type) {
	case "fabric":
		runtime["fabricLoader"] = m.Loader.Version
		versionLabel += "-fabric" + m.Loader.Version
	case "quilt":
		runtime["quiltLoader"] = m.Loader.Version
		versionLabel += "-quilt" + m.Loader.Version
	case "forge":
		runtime["forge"] = m.Loader.Version
		versionLabel += "-forge" + m.Loader.Version
	case "neoforge":
		runtime["neoForged"] = m.Loader.Version
		versionLabel += "-neoforged" + m.Loader.Version
	}
	nowMs := time.Now().UnixMilli()
	instance := map[string]any{
		"name":           displayNameFrom(info, m),
		"author":         "",
		"description":    "",
		"version":        versionLabel,
		"edition":        "java",
		"runtime":        runtime,
		"java":           "",
		"url":            "",
		"icon":           "",
		"fileApi":        "",
		"server":         nil,
		"lastAccessDate": nowMs,
		"lastPlayedDate": 0,
		"playtime":       0,
		"creationDate":   nowMs,
		"path":           instanceDir,
	}
	if err := writeJSONFile(filepath.Join(instanceDir, "instance.json"), instance); err != nil {
		return nil, err
	}
	return []string{"Wrote XMCL instance.json."}, nil
}

// ── Dawn (formerly Feather) ──────────────────────
// profile.json + content-index.json (captured from polyforge-dawn-test).

func genDawnProfile(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	minMiB, recMiB, maxMiB := 2048, 4096, 8192
	if defaults.MinMemoryMB > 0 {
		minMiB = defaults.MinMemoryMB
	}
	if defaults.RecommendedMemoryMB > 0 {
		recMiB = defaults.RecommendedMemoryMB
		if maxMiB < recMiB*2 {
			maxMiB = recMiB * 2
		}
	}
	profile := map[string]any{
		"id":               filepath.Base(instanceDir),
		"kind":             "Custom",
		"name":             displayNameFrom(info, m),
		"minecraftVersion": m.Minecraft,
		"settings": map[string]any{
			"memoryDefaults": map[string]any{
				"minimumMiB":     minMiB,
				"recommendedMiB": recMiB,
				"maximumMiB":     maxMiB,
			},
			"jvmOptions": map[string]any{"extraArguments": []any{}},
			"javaOverrides": map[string]any{
				"enabled":              false,
				"javaInstallationPath": map[string]any{"enabled": false},
				"memoryAllocationMiB":  map[string]any{"enabled": false},
				"javaArguments":        map[string]any{"enabled": false},
				"environmentVariables": map[string]any{"enabled": false},
				"prelaunchHook":        map[string]any{"enabled": false},
				"wrapperHook":          map[string]any{"enabled": false},
				"postExitHook":         map[string]any{"enabled": false},
			},
			"tags": []any{},
		},
		"runtimePolicy": map[string]any{
			"modAuthority":  "UserManaged",
			"requiredHooks": []any{"DawnCoreMods"},
		},
	}
	// Dawn's loader kinds cover the modern loaders only; LiteLoader (legacy)
	// is omitted so the profile still parses — mods stay in .minecraft/mods.
	if kind := loaderKindTitle(m.Loader.Type); kind != "" && kind != "LiteLoader" && m.Loader.Version != "" {
		profile["loader"] = map[string]any{"kind": kind, "version": m.Loader.Version}
	}
	doc := map[string]any{"schemaVersion": 3, "profile": profile}
	if err := writeJSONFile(filepath.Join(instanceDir, "profile.json"), doc); err != nil {
		return nil, err
	}
	index := map[string]any{"schemaVersion": 1, "entries": []any{}}
	if err := writeJSONFile(filepath.Join(instanceDir, "content-index.json"), index); err != nil {
		return nil, err
	}
	return []string{"Wrote Dawn profile.json + content-index.json."}, nil
}

// ── Polymerium (Trident) ─────────────────────────
// profile.json (captured from polyforge_polymerium_test); data.lock.json is
// produced by the launcher itself on first deploy.

func genPolymeriumProfile(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	loader := ""
	if m.Loader.Version != "" {
		switch strings.ToLower(m.Loader.Type) {
		case "fabric":
			loader = "net.fabricmc:" + m.Loader.Version
		case "quilt":
			loader = "org.quiltmc:" + m.Loader.Version
		case "forge":
			loader = "net.minecraftforge:" + m.Loader.Version
		case "neoforge":
			loader = "net.neoforged:" + m.Loader.Version
		}
	}
	profile := map[string]any{
		"name": displayNameFrom(info, m),
		"setup": map[string]any{
			"source":       nil,
			"sourceOrders": []any{},
			"version":      m.Minecraft,
			"loader":       loader,
			"packages":     []any{},
			"rules":        []any{},
		},
		"overrides": map[string]any{},
	}
	if err := writeJSONFile(filepath.Join(instanceDir, "profile.json"), profile); err != nil {
		return nil, err
	}
	return []string{"Wrote Polymerium profile.json; the launcher deploys the instance on first open."}, nil
}

// ── Modrinth app (Theseus) ───────────────────────
//
// Profiles live as folders under profiles\, but the app only shows rows from
// its app.db (SQLite). The pure-Go driver the Modrinth clone tool already
// uses lets us register the profile directly; app-version-specific columns
// (install_stage, protocol_version, ...) are copied from an existing row so
// we never guess values the running app version expects.

func genModrinthProfile(instanceDir string, m *PackManifest, info map[string]any, defaults PackLauncherDefaults) ([]string, error) {
	dbPath := modrinthDBPath()
	if dbPath == "" {
		return []string{"Modrinth app.db not found — open the Modrinth App once, then reinstall (or import the folder in-app)."}, nil
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return []string{fmt.Sprintf("Could not open Modrinth app.db (%v) — import the profile in-app.", err)}, nil
	}
	defer db.Close()

	folder := filepath.Base(instanceDir)
	loader := strings.ToLower(m.Loader.Type)
	switch loader {
	case "fabric", "forge", "quilt", "neoforge":
	default:
		// Theseus' mod_loader enum stops at the modern loaders; anything
		// else (liteloader, "") registers as vanilla so the row still loads.
		loader = "vanilla"
	}
	name := displayNameFrom(info, m)
	now := time.Now().UTC().Unix()

	var count int
	if err := db.QueryRow("SELECT COUNT(1) FROM profiles WHERE path = ?", folder).Scan(&count); err != nil {
		return []string{fmt.Sprintf("Could not query Modrinth profiles (%v) — import the profile in-app.", err)}, nil
	}
	if count == 0 {
		// Template a new row from any existing profile so app-version-specific
		// defaults stay valid, then overwrite the identity fields below.
		var template int
		if err := db.QueryRow("SELECT COUNT(1) FROM profiles").Scan(&template); err != nil || template == 0 {
			return []string{"No existing Modrinth profile to template from — create any profile in the app once, then reinstall."}, nil
		}
		insertSQL := `INSERT INTO profiles (
  path, install_stage, name, icon_path,
  game_version, mod_loader, mod_loader_version,
  groups, linked_project_id, linked_version_id, locked,
  created, modified, last_played,
  submitted_time_played, recent_time_played,
  override_java_path, override_extra_launch_args, override_custom_env_vars,
  override_mc_memory_max, override_mc_force_fullscreen,
  override_mc_game_resolution_x, override_mc_game_resolution_y,
  override_hook_pre_launch, override_hook_wrapper, override_hook_post_exit,
  protocol_version, launcher_feature_version
)
SELECT
  ?, install_stage, ?, NULL,
  ?, ?, ?,
  groups, NULL, NULL, locked,
  ?, ?, NULL,
  0, 0,
  NULL, NULL, override_custom_env_vars,
  override_mc_memory_max, override_mc_force_fullscreen,
  NULL, NULL,
  NULL, NULL, NULL,
  protocol_version, launcher_feature_version
FROM profiles LIMIT 1`
		if _, err := db.Exec(insertSQL, folder, name, m.Minecraft, loader, m.Loader.Version, now, now); err != nil {
			return []string{fmt.Sprintf("Could not register the profile in Modrinth's app.db (%v) — import it in-app.", err)}, nil
		}
		return []string{"Registered the profile in the Modrinth App."}, nil
	}

	// Same folder already registered (an update): refresh its identity fields.
	updateSQL := `UPDATE profiles SET name = ?, game_version = ?, mod_loader = ?, mod_loader_version = ?, modified = ? WHERE path = ?`
	if _, err := db.Exec(updateSQL, name, m.Minecraft, loader, m.Loader.Version, now, folder); err != nil {
		return []string{fmt.Sprintf("Could not update the existing Modrinth profile (%v).", err)}, nil
	}
	return []string{"Updated the existing Modrinth App profile."}, nil
}
