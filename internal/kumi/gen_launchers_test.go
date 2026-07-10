package kumi

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func genTestManifest(loaderType, loaderVersion string) *PackManifest {
	return &PackManifest{
		SchemaVersion: 1,
		ID:            "smoke-pack",
		Name:          "Smoke Pack",
		Version:       "1.0.0",
		Minecraft:     "26.2",
		Loader:        PackLoader{Type: loaderType, Version: loaderVersion},
	}
}

func decodeJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("%s is not valid JSON: %v", path, err)
	}
	return v
}

func TestPlanInstallDirs(t *testing.T) {
	sep := string(os.PathSeparator)
	cases := []struct {
		launcher, chosen, name string
		wantInstance, wantGame string
	}{
		// Launcher root chosen: instances/<name>/<subdir> is derived.
		{"prismlauncher", `C:\Users\u\AppData\Roaming\PrismLauncher`, "Smoke Pack",
			`C:\Users\u\AppData\Roaming\PrismLauncher\instances\Smoke Pack`,
			`C:\Users\u\AppData\Roaming\PrismLauncher\instances\Smoke Pack\minecraft`},
		// The instances folder itself chosen.
		{"multimc", `D:\MultiMC\instances`, "Smoke Pack",
			`D:\MultiMC\instances\Smoke Pack`,
			`D:\MultiMC\instances\Smoke Pack\.minecraft`},
		// The instance folder itself chosen: no double nesting.
		{"prismlauncher", `C:\PrismLauncher\instances\Smoke Pack`, "Smoke Pack",
			`C:\PrismLauncher\instances\Smoke Pack`,
			`C:\PrismLauncher\instances\Smoke Pack\minecraft`},
		// CurseForge capitalises its instances dir.
		{"curseforge", `C:\Users\u\curseforge\minecraft`, "Smoke Pack",
			`C:\Users\u\curseforge\minecraft\Instances\Smoke Pack`,
			`C:\Users\u\curseforge\minecraft\Instances\Smoke Pack`},
		// Modrinth keeps profiles, and the profile root is the game dir.
		{"modrinth", `C:\Users\u\AppData\Roaming\ModrinthApp`, "Smoke Pack",
			`C:\Users\u\AppData\Roaming\ModrinthApp\profiles\Smoke Pack`,
			`C:\Users\u\AppData\Roaming\ModrinthApp\profiles\Smoke Pack`},
		// GDLauncher nests the game dir under instance\.
		{"gdlauncher", `C:\Users\u\AppData\Roaming\gdlauncher_carbon`, "Smoke Pack",
			`C:\Users\u\AppData\Roaming\gdlauncher_carbon\instances\Smoke Pack`,
			`C:\Users\u\AppData\Roaming\gdlauncher_carbon\instances\Smoke Pack\instance`},
		// Unknown/empty launcher: the chosen path is used untouched.
		{"", `C:\somewhere`, "Smoke Pack", `C:\somewhere`, `C:\somewhere`},
		{"custom", `C:\somewhere`, "Smoke Pack", `C:\somewhere`, `C:\somewhere`},
	}
	for _, c := range cases {
		gotInstance, gotGame := PlanInstallDirs(c.launcher, c.chosen, c.name)
		if gotInstance != filepath.Clean(c.wantInstance) || gotGame != filepath.Clean(c.wantGame) {
			t.Errorf("PlanInstallDirs(%q, %q, %q) = (%q, %q), want (%q, %q)",
				c.launcher, c.chosen, c.name, gotInstance, gotGame, c.wantInstance, c.wantGame)
		}
	}
	_ = sep
}

func TestSanitizeInstanceName(t *testing.T) {
	if got := sanitizeInstanceName(`Sm:o*ke?  `); got != "Sm_o_ke_" {
		t.Errorf("sanitizeInstanceName = %q", got)
	}
}

func TestGenMMCInstance(t *testing.T) {
	dir := t.TempDir()
	m := genTestManifest("fabric", "0.19.3")
	notes, err := genMMCInstance(true)(dir, m, nil, PackLauncherDefaults{})
	if err != nil {
		t.Fatalf("genMMCInstance: %v", err)
	}
	if len(notes) == 0 {
		t.Error("expected a note")
	}

	cfg, err := os.ReadFile(filepath.Join(dir, "instance.cfg"))
	if err != nil {
		t.Fatalf("instance.cfg: %v", err)
	}
	if !strings.HasPrefix(string(cfg), "[General]\nConfigVersion=1.3\n") {
		t.Errorf("prism-style cfg missing header:\n%s", cfg)
	}
	if !strings.Contains(string(cfg), "name=Smoke Pack\n") {
		t.Errorf("cfg missing name:\n%s", cfg)
	}

	pack := decodeJSONFile(t, filepath.Join(dir, "mmc-pack.json"))
	if pack["formatVersion"] != float64(1) {
		t.Errorf("formatVersion = %v", pack["formatVersion"])
	}
	uids := map[string]string{}
	for _, c := range pack["components"].([]any) {
		comp := c.(map[string]any)
		version, _ := comp["version"].(string)
		uids[comp["uid"].(string)] = version
	}
	if uids["net.minecraft"] != "26.2" || uids["net.fabricmc.intermediary"] != "26.2" || uids["net.fabricmc.fabric-loader"] != "0.19.3" {
		t.Errorf("components = %v", uids)
	}

	// Legacy MultiMC flavor: headerless cfg, quilt loader component.
	dir2 := t.TempDir()
	if _, err := genMMCInstance(false)(dir2, genTestManifest("quilt", "0.22.0"), nil, PackLauncherDefaults{}); err != nil {
		t.Fatalf("legacy flavor: %v", err)
	}
	cfg2, _ := os.ReadFile(filepath.Join(dir2, "instance.cfg"))
	if strings.Contains(string(cfg2), "[General]") {
		t.Error("legacy cfg must not have a [General] header")
	}
	pack2 := decodeJSONFile(t, filepath.Join(dir2, "mmc-pack.json"))
	found := false
	for _, c := range pack2["components"].([]any) {
		if c.(map[string]any)["uid"] == "org.quiltmc.quilt-loader" {
			found = true
		}
	}
	if !found {
		t.Error("quilt pack missing org.quiltmc.quilt-loader component")
	}
}

// TestMMCComponentUIDs pins the loader component uid for every loader the
// pack format supports — Prism-family launchers resolve these against their
// metadata service, so a wrong uid means a broken instance.
func TestMMCComponentUIDs(t *testing.T) {
	cases := map[string]string{
		"fabric":     "net.fabricmc.fabric-loader",
		"quilt":      "org.quiltmc.quilt-loader",
		"forge":      "net.minecraftforge",
		"neoforge":   "net.neoforged",
		"liteloader": "com.mumfrey.liteloader",
	}
	for loader, wantUID := range cases {
		comps := mmcComponents(genTestManifest(loader, "1.0.0"))
		found := false
		for _, c := range comps {
			if c.UID == wantUID && c.Version == "1.0.0" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: component %s@1.0.0 missing; got %+v", loader, wantUID, comps)
		}
	}
}

func TestGenVanillaProfile(t *testing.T) {
	appData := t.TempDir()
	t.Setenv("APPDATA", appData)
	mcDir := filepath.Join(appData, ".minecraft")
	if err := os.MkdirAll(mcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{"profiles":{"existing":{"name":"Keep Me"}},"settings":{"keepLauncherOpen":true}}`
	profilesPath := filepath.Join(mcDir, "launcher_profiles.json")
	if err := os.WriteFile(profilesPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	// Vanilla loader: no network needed, lastVersionId is the MC version.
	m := genTestManifest("vanilla", "")
	instanceDir := filepath.Join(appData, "PolyForgeProfiles", "Smoke Pack")
	notes, err := genVanillaProfile(instanceDir, m, nil, PackLauncherDefaults{RecommendedMemoryMB: 4096})
	if err != nil {
		t.Fatalf("genVanillaProfile: %v", err)
	}
	if len(notes) == 0 {
		t.Error("expected a note")
	}

	doc := decodeJSONFile(t, profilesPath)
	profiles := doc["profiles"].(map[string]any)
	if _, ok := profiles["existing"]; !ok {
		t.Error("pre-existing profile was dropped")
	}
	entry, ok := profiles["polyforge-smoke-pack"].(map[string]any)
	if !ok {
		t.Fatalf("profile entry missing; profiles = %v", profiles)
	}
	if entry["lastVersionId"] != "26.2" || entry["gameDir"] != instanceDir || entry["javaArgs"] != "-Xmx4096M" {
		t.Errorf("entry = %v", entry)
	}
	if doc["settings"].(map[string]any)["keepLauncherOpen"] != true {
		t.Error("unrelated settings were not preserved")
	}
}

func TestVanillaVersionIDs(t *testing.T) {
	cases := map[string]*PackManifest{
		"fabric-loader-0.19.3-26.2": genTestManifest("fabric", "0.19.3"),
		"quilt-loader-0.22.0-26.2":  genTestManifest("quilt", "0.22.0"),
		"26.2-forge-47.4.0":         genTestManifest("forge", "47.4.0"),
		"neoforge-21.4.111":         genTestManifest("neoforge", "21.4.111"),
		"26.2-LiteLoader26.2":       genTestManifest("liteloader", "1.2"),
		"26.2":                      genTestManifest("", ""),
	}
	for want, m := range cases {
		if got := vanillaVersionID(m); got != want {
			t.Errorf("vanillaVersionID(%s %s) = %q, want %q", m.Loader.Type, m.Loader.Version, got, want)
		}
	}
}

func TestGenCurseForgeInstance(t *testing.T) {
	dir := t.TempDir()
	if _, err := genCurseForgeInstance(dir, genTestManifest("fabric", "0.19.3"), nil, PackLauncherDefaults{RecommendedMemoryMB: 6144}); err != nil {
		t.Fatalf("genCurseForgeInstance: %v", err)
	}
	inst := decodeJSONFile(t, filepath.Join(dir, "minecraftinstance.json"))
	if inst["gameVersion"] != "26.2" || inst["gameTypeID"] != float64(432) || inst["name"] != "Smoke Pack" {
		t.Errorf("instance identity fields wrong: %v", inst)
	}
	if inst["guid"] == "" || len(inst["guid"].(string)) != 36 {
		t.Errorf("guid = %v", inst["guid"])
	}
	if inst["allocatedMemory"] != float64(6144) {
		t.Errorf("allocatedMemory = %v", inst["allocatedMemory"])
	}
	bml := inst["baseModLoader"].(map[string]any)
	if bml["type"] != float64(4) || bml["name"] != "fabric-0.19.3-26.2" || bml["forgeVersion"] != "0.19.3" {
		t.Errorf("baseModLoader = %v", bml)
	}
}

func TestGenGDLauncherInstance(t *testing.T) {
	dir := t.TempDir()
	if _, err := genGDLauncherInstance(dir, genTestManifest("neoforge", "21.4.111"), nil, PackLauncherDefaults{}); err != nil {
		t.Fatalf("genGDLauncherInstance: %v", err)
	}
	inst := decodeJSONFile(t, filepath.Join(dir, "instance.json"))
	if inst["_version"] != "1" {
		t.Errorf("_version = %v", inst["_version"])
	}
	gc := inst["game_configuration"].(map[string]any)["version"].(map[string]any)
	if gc["release"] != "26.2" {
		t.Errorf("release = %v", gc["release"])
	}
	ml := gc["modloaders"].([]any)[0].(map[string]any)
	if ml["type"] != "Neoforge" || ml["version"] != "21.4.111" {
		t.Errorf("modloader = %v", ml)
	}

	// LiteLoader is outside Carbon's modloader enum — it must degrade to a
	// vanilla instance (empty modloaders) rather than an unparseable type.
	dir2 := t.TempDir()
	notes, err := genGDLauncherInstance(dir2, genTestManifest("liteloader", "1.2"), nil, PackLauncherDefaults{})
	if err != nil {
		t.Fatalf("liteloader flavor: %v", err)
	}
	inst2 := decodeJSONFile(t, filepath.Join(dir2, "instance.json"))
	gc2 := inst2["game_configuration"].(map[string]any)["version"].(map[string]any)
	if len(gc2["modloaders"].([]any)) != 0 {
		t.Errorf("liteloader pack must write no modloaders, got %v", gc2["modloaders"])
	}
	if len(notes) < 2 {
		t.Errorf("expected a degradation note, got %v", notes)
	}
}

func TestGenXMCLInstance(t *testing.T) {
	dir := t.TempDir()
	if _, err := genXMCLInstance(dir, genTestManifest("fabric", "0.19.3"), nil, PackLauncherDefaults{}); err != nil {
		t.Fatalf("genXMCLInstance: %v", err)
	}
	inst := decodeJSONFile(t, filepath.Join(dir, "instance.json"))
	rt := inst["runtime"].(map[string]any)
	if rt["minecraft"] != "26.2" || rt["fabricLoader"] != "0.19.3" || rt["forge"] != "" {
		t.Errorf("runtime = %v", rt)
	}
	if inst["path"] != dir || inst["edition"] != "java" {
		t.Errorf("instance = %v", inst)
	}
}

func TestGenDawnProfile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "smoke-pack")
	if _, err := genDawnProfile(dir, genTestManifest("fabric", "0.19.3"), nil, PackLauncherDefaults{MinMemoryMB: 3072, RecommendedMemoryMB: 5120}); err != nil {
		t.Fatalf("genDawnProfile: %v", err)
	}
	doc := decodeJSONFile(t, filepath.Join(dir, "profile.json"))
	if doc["schemaVersion"] != float64(3) {
		t.Errorf("schemaVersion = %v", doc["schemaVersion"])
	}
	profile := doc["profile"].(map[string]any)
	if profile["id"] != "smoke-pack" || profile["kind"] != "Custom" || profile["minecraftVersion"] != "26.2" {
		t.Errorf("profile = %v", profile)
	}
	loader := profile["loader"].(map[string]any)
	if loader["kind"] != "Fabric" || loader["version"] != "0.19.3" {
		t.Errorf("loader = %v", loader)
	}
	mem := profile["settings"].(map[string]any)["memoryDefaults"].(map[string]any)
	if mem["minimumMiB"] != float64(3072) || mem["recommendedMiB"] != float64(5120) {
		t.Errorf("memoryDefaults = %v", mem)
	}
	index := decodeJSONFile(t, filepath.Join(dir, "content-index.json"))
	if index["schemaVersion"] != float64(1) {
		t.Errorf("content-index = %v", index)
	}
}

func TestGenPolymeriumProfile(t *testing.T) {
	dir := t.TempDir()
	if _, err := genPolymeriumProfile(dir, genTestManifest("fabric", "0.19.3"), nil, PackLauncherDefaults{}); err != nil {
		t.Fatalf("genPolymeriumProfile: %v", err)
	}
	doc := decodeJSONFile(t, filepath.Join(dir, "profile.json"))
	setup := doc["setup"].(map[string]any)
	if setup["version"] != "26.2" || setup["loader"] != "net.fabricmc:0.19.3" {
		t.Errorf("setup = %v", setup)
	}
}

// modrinthProfileColumns matches the schema CloneModrinthProfile and
// genModrinthProfile touch; the test DB carries every referenced column.
const modrinthProfileColumns = `
  path TEXT PRIMARY KEY, install_stage TEXT, name TEXT, icon_path TEXT,
  game_version TEXT, mod_loader TEXT, mod_loader_version TEXT,
  groups TEXT, linked_project_id TEXT, linked_version_id TEXT, locked INTEGER,
  created INTEGER, modified INTEGER, last_played INTEGER,
  submitted_time_played INTEGER, recent_time_played INTEGER,
  override_java_path TEXT, override_extra_launch_args TEXT, override_custom_env_vars TEXT,
  override_mc_memory_max INTEGER, override_mc_force_fullscreen INTEGER,
  override_mc_game_resolution_x INTEGER, override_mc_game_resolution_y INTEGER,
  override_hook_pre_launch TEXT, override_hook_wrapper TEXT, override_hook_post_exit TEXT,
  protocol_version INTEGER, launcher_feature_version INTEGER`

func TestGenModrinthProfile(t *testing.T) {
	appData := t.TempDir()
	t.Setenv("APPDATA", appData)
	appDir := filepath.Join(appData, "ModrinthApp")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(appDir, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE profiles (" + modrinthProfileColumns + ")"); err != nil {
		t.Fatal(err)
	}
	// Template row, as if the user already has one profile in the app.
	if _, err := db.Exec(`INSERT INTO profiles (path, install_stage, name, game_version, mod_loader, mod_loader_version, groups, locked, created, modified, submitted_time_played, recent_time_played, override_custom_env_vars, override_mc_memory_max, override_mc_force_fullscreen, protocol_version, launcher_feature_version)
VALUES ('Existing', 'installed', 'Existing', '26.1', 'fabric', '0.19.0', '[]', 0, 1, 1, 0, 0, NULL, NULL, NULL, 2, 1)`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	instanceDir := filepath.Join(appDir, "profiles", "Smoke Pack")
	notes, err := genModrinthProfile(instanceDir, genTestManifest("fabric", "0.19.3"), nil, PackLauncherDefaults{})
	if err != nil {
		t.Fatalf("genModrinthProfile: %v", err)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "Registered") {
		t.Fatalf("notes = %v", notes)
	}

	check, err := sql.Open("sqlite", filepath.Join(appDir, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer check.Close()
	var name, gv, ml, mlv, stage string
	row := check.QueryRow("SELECT name, game_version, mod_loader, mod_loader_version, install_stage FROM profiles WHERE path = 'Smoke Pack'")
	if err := row.Scan(&name, &gv, &ml, &mlv, &stage); err != nil {
		t.Fatalf("inserted row: %v", err)
	}
	if name != "Smoke Pack" || gv != "26.2" || ml != "fabric" || mlv != "0.19.3" || stage != "installed" {
		t.Errorf("row = %s %s %s %s %s", name, gv, ml, mlv, stage)
	}

	// Second run against the same folder updates instead of duplicating.
	m2 := genTestManifest("fabric", "0.19.4")
	if notes, err = genModrinthProfile(instanceDir, m2, nil, PackLauncherDefaults{}); err != nil {
		t.Fatalf("update pass: %v", err)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "Updated") {
		t.Fatalf("update notes = %v", notes)
	}
	var count int
	if err := check.QueryRow("SELECT COUNT(1) FROM profiles WHERE path = 'Smoke Pack'").Scan(&count); err != nil || count != 1 {
		t.Errorf("count = %d, err = %v", count, err)
	}
}

func TestGenerateLauncherFilesUnknownLauncher(t *testing.T) {
	generated, notes, err := GenerateLauncherFiles("qwertz", t.TempDir(), genTestManifest("fabric", "0.19.3"), nil)
	if generated || notes != nil || err != nil {
		t.Errorf("expected no-op for launcher without a generator, got (%v, %v, %v)", generated, notes, err)
	}
}
