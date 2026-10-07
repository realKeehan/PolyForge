package kumi

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unicode/utf16"
)

// The expected paths come from Trident PathDef, including its platform-specific
// brand names and authoritative home precedence (not a list of fallbacks).
func TestTridentDataHome(t *testing.T) {
	home, cwd := t.TempDir(), t.TempDir()
	env := map[string]string{"LOCALAPPDATA": filepath.Join(home, "Local")}
	getenv := func(key string) string { return env[key] }
	for _, tc := range []struct{ platform, relative string }{
		{"windows", "Local/Trident"},
		{"darwin", "Library/Application Support/dev.dearain.trident"},
		{"linux", ".local/share/trident"},
	} {
		if got := tridentDataHome(home, cwd, tc.platform, getenv); got != filepath.Join(home, filepath.FromSlash(tc.relative)) {
			t.Errorf("%s data home = %q", tc.platform, got)
		}
	}
	legacy := filepath.Join(home, ".trident")
	if err := os.Mkdir(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := tridentDataHome(home, cwd, "darwin", getenv); got != legacy {
		t.Fatal(got)
	}
	override := filepath.Join(home, "custom") // need not already exist
	if err := os.WriteFile(filepath.Join(home, ".trident.home"), []byte(override+"\nignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := tridentDataHome(home, cwd, "darwin", getenv); got != override {
		t.Fatal(got)
	}
	workspace := filepath.Join(cwd, ".trident")
	if err := os.Mkdir(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := tridentDataHome(home, filepath.Join(cwd, "child"), "darwin", getenv); got != workspace {
		t.Fatal(got)
	}
	env["TRIDENT_HOME"] = filepath.Join(home, "env-home")
	if got := tridentDataHome(home, cwd, "darwin", getenv); got != env["TRIDENT_HOME"] {
		t.Fatal(got)
	}
}

func tridentTestWrite(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func tridentTestInstall(t *testing.T, root string, files map[string]string, mode InstallMode, configure ...func(*PackManifest)) string {
	t.Helper()
	pack := buildTestPack(t, files, true, configure...)
	reader, err := openPackReader(pack)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := readZipPackManifest(reader)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := planPackInstall("polymerium", root, manifest, nil, mode)
	if err != nil {
		t.Fatal(err)
	}
	_, err = installTridentPack(pack, plan, manifest, nil)
	if err != nil {
		t.Fatal(err)
	}
	return plan.InstanceDir
}

// Filesystem assertions check the adapter's ownership boundary. Actual Trident
// deployment and game launch remain separate acceptance checks.
func TestTridentPackUpdateLiveOwnership(t *testing.T) {
	root := t.TempDir()
	instance := tridentTestInstall(t, root, map[string]string{
		"mods/old.jar": "old", "config/settings.txt": "v1", "options.txt": "default",
	}, ModeAuto)
	oldSource := "modpack://local/Test%20Pack"
	tridentTestWrite(t, instance, "build/mods/old.jar", "old")
	tridentTestWrite(t, instance, "build/config/settings.txt", "runtime edit")
	tridentTestWrite(t, instance, "build/mods/user.jar", "user")
	tridentTestWrite(t, instance, "persist/options.txt", "user settings")
	if err := os.Symlink(filepath.Join(instance, "persist", "options.txt"), filepath.Join(instance, "build", "options.txt")); err != nil {
		t.Fatal(err)
	}
	// The manifest can retain paths whose import source was removed manually.
	tridentTestWrite(t, instance, "build/config/orphan.txt", "stale live")
	tridentTestWrite(t, instance, "build/trident.import.json", `{"version":1,"files":["mods/old.jar","config/settings.txt","options.txt","config/orphan.txt"]}`)
	profile := decodeJSONFile(t, filepath.Join(instance, "profile.json"))
	profile["name"], profile["custom"] = "My instance", "keep"
	setup := profile["setup"].(map[string]any)
	setup["sourceOrders"] = []any{oldSource, "collection://by-name/MyMods"}
	setup["packages"] = []any{
		map[string]any{"pref": "pref://modrinth/old@v1", "enabled": true, "source": oldSource},
		map[string]any{"pref": "pref://modrinth/manual@v1", "enabled": false},
		map[string]any{"pref": "pref://modrinth/other@v1", "enabled": true, "source": "modpack://local/Other"},
	}
	if err := writeJSONFile(filepath.Join(instance, "profile.json"), profile); err != nil {
		t.Fatal(err)
	}
	// The same pack can change its display name without changing identity.
	got := tridentTestInstall(t, instance, map[string]string{
		"config/settings.txt": "v2", "mods/new.jar": "managed", "mods/unknown.jar": "bundled", "options.txt": "pack settings",
	}, ModeAuto, func(m *PackManifest) {
		m.Name, m.Version = "Other: Pack", "2.0"
		m.Mods = []PackMod{
			{File: "new.jar", Source: &PackModSource{Provider: "modrinth", ProjectID: "managed", VersionID: "v2"}},
			{File: "unknown.jar", Source: &PackModSource{Provider: "modrinth", ProjectID: "unknown"}},
			{File: "absent.jar", Source: &PackModSource{Provider: "modrinth", ProjectID: "absent", VersionID: "v1"}},
		}
	})
	if got != instance {
		t.Fatalf("created a different instance: %s", got)
	}
	for _, relative := range []string{
		"import/mods/old.jar", "build/mods/old.jar", "build/config/settings.txt", "build/config/orphan.txt",
		"import/mods/new.jar", "import/" + installedManifestName, "build/trident.import.json",
	} {
		if _, err := os.Lstat(filepath.Join(instance, filepath.FromSlash(relative))); !os.IsNotExist(err) {
			t.Errorf("unexpected file remains: %s (%v)", relative, err)
		}
	}
	for relative, want := range map[string]string{
		"import/config/settings.txt": "v2", "import/mods/unknown.jar": "bundled",
		"build/mods/user.jar": "user", "persist/options.txt": "user settings", "build/options.txt": "user settings",
	} {
		data, err := os.ReadFile(filepath.Join(instance, filepath.FromSlash(relative)))
		if err != nil || string(data) != want {
			t.Errorf("%s = %q, %v", relative, data, err)
		}
	}
	profile = decodeJSONFile(t, filepath.Join(instance, "profile.json"))
	setup = profile["setup"].(map[string]any)
	packages := setup["packages"].([]any)
	if profile["name"] != "My instance" || profile["custom"] != "keep" || len(packages) != 3 {
		t.Fatalf("profile preservation: %v", profile)
	}
	if setup["source"] != "modpack://local/Other%3A%20Pack" || setup["sourceOrders"].([]any)[0] != setup["source"] {
		t.Fatalf("wrong active source: %v", setup)
	}
	entry := packages[2].(map[string]any)
	if entry["pref"] != "pref://modrinth/managed@v2" || entry["source"] != setup["source"] {
		t.Fatalf("wrong managed entry: %v", entry)
	}
	if packages[0].(map[string]any)["enabled"] != false || packages[1].(map[string]any)["source"] != "modpack://local/Other" {
		t.Fatal("unrelated package entries changed")
	}
	if matches, _ := filepath.Glob(filepath.Join(instance, ".polyforge-install-*")); len(matches) != 0 {
		t.Fatalf("staging files remain: %v", matches)
	}
}

func TestTridentCleanQuarantineOutsideImport(t *testing.T) {
	root := t.TempDir()
	instance := tridentTestInstall(t, root, map[string]string{"config/foo/a.txt": "original"}, ModeInstall)
	oldProfile, err := os.ReadFile(filepath.Join(instance, "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	oldRecord, err := os.ReadFile(filepath.Join(instance, installedManifestName))
	if err != nil {
		t.Fatal(err)
	}
	tridentTestWrite(t, instance, "build/config/foo/a.txt", "live")
	tridentTestWrite(t, instance, "build/trident.import.json", `{"version":1,"files":["config/foo/a.txt"]}`)
	tridentTestWrite(t, instance, "build/saves/world.txt", "world")
	tridentTestInstall(t, root, map[string]string{"config/foo": "replacement"}, ModeClean)
	matches, err := filepath.Glob(filepath.Join(instance, quarantineDirName, "*", "import", "config", "foo", "a.txt"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("quarantine = %v, %v", matches, err)
	}
	backups, err := filepath.Glob(filepath.Join(instance, quarantineDirName, "*"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("backup directories = %v, %v", backups, err)
	}
	for relative, want := range map[string]string{
		"import/config/foo/a.txt":   "original",
		"build/config/foo/a.txt":    "live",
		"build/trident.import.json": `{"version":1,"files":["config/foo/a.txt"]}`,
		"profile.json":              string(oldProfile),
		installedManifestName:       string(oldRecord),
	} {
		data, err := os.ReadFile(filepath.Join(backups[0], filepath.FromSlash(relative)))
		if err != nil || string(data) != want {
			t.Errorf("backup %s = %q, %v", relative, data, err)
		}
	}
	if _, err := os.Stat(filepath.Join(instance, "build", "config", "foo")); !os.IsNotExist(err) {
		t.Fatal("old live directory would block the replacement file")
	}
	if data, err := os.ReadFile(filepath.Join(instance, "build", "saves", "world.txt")); err != nil || string(data) != "world" {
		t.Fatal("unowned live file changed")
	}
}

func TestTridentRejectsUnsafeProjectionBeforeUpdate(t *testing.T) {
	root := t.TempDir()
	instance := tridentTestInstall(t, root, map[string]string{"config/a.txt": "original"}, ModeInstall)
	tridentTestWrite(t, instance, "build/trident.import.json", `{"version":1,"files":["../outside.txt"]}`)
	pack := buildTestPack(t, map[string]string{"config/a.txt": "replacement"}, true)
	reader, _ := openPackReader(pack)
	manifest, _ := readZipPackManifest(reader)
	plan, err := planPackInstall("polymerium", root, manifest, nil, ModeAuto)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := installTridentPack(pack, plan, manifest, nil); err == nil {
		t.Fatal("accepted traversal in projection manifest")
	}
	if data, err := os.ReadFile(filepath.Join(instance, "import", "config", "a.txt")); err != nil || string(data) != "original" {
		t.Fatal("modified import before validating projection ownership")
	}
}

func TestTridentRejectsUnsupportedLoader(t *testing.T) {
	manifest := genTestManifest("liteloader", "1.0")
	if _, err := tridentProfile(t.TempDir(), manifest, nil, PackLauncherDefaults{}); err == nil {
		t.Fatal("unsupported loader silently became vanilla")
	}
}

// Trident updates retain entry settings when the project stays in the pack,
// even when its version changes. Other source groups remain independent.
func TestTridentPackageMergePreservesEntrySettings(t *testing.T) {
	for _, mode := range []InstallMode{ModeUpdate, ModeInstall} {
		t.Run(string(mode), func(t *testing.T) {
			oldSource := "modpack://local/Old"
			manifest := &PackManifest{Name: "New"}
			newSource := tridentModpackSource(manifest)
			kept := map[string]any{
				"pref": "pref://modrinth/kept@v1", "source": oldSource,
				"enabled": false, "tags": []any{"optional"}, "custom": "keep",
			}
			retired := map[string]any{"pref": "pref://modrinth/retired@v1", "source": oldSource, "enabled": true}
			manual := map[string]any{"pref": "pref://modrinth/kept@manual", "enabled": false}
			other := map[string]any{"pref": "pref://modrinth/kept@other", "source": "collection://by-name/Other", "enabled": true}
			added := map[string]any{"pref": "pref://curseforge/123@456", "source": newSource, "enabled": true}
			profile := map[string]any{"setup": map[string]any{
				"source": oldSource, "sourceOrders": []any{oldSource},
				"packages": []any{kept, retired, manual, other},
			}}
			incoming := []any{
				map[string]any{"pref": "pref://modrinth/kept@v2", "source": newSource, "enabled": true},
				added,
			}
			if err := updateTridentPackages(profile, manifest, incoming, mode); err != nil {
				t.Fatal(err)
			}
			wantKept := map[string]any{
				"pref": "pref://modrinth/kept@v2", "source": newSource,
				"enabled": false, "tags": []any{"optional"}, "custom": "keep",
			}
			want := []any{wantKept, manual, other, added}
			if mode == ModeInstall {
				want = []any{wantKept,
					map[string]any{"pref": "pref://modrinth/retired@v1", "source": newSource, "enabled": true},
					manual, other, added}
			}
			setup := profile["setup"].(map[string]any)
			if !reflect.DeepEqual(setup["packages"], want) {
				t.Errorf("packages = %#v, want %#v", setup["packages"], want)
			}
			if setup["source"] != newSource || !reflect.DeepEqual(setup["sourceOrders"], []any{newSource}) {
				t.Errorf("wrong source binding: %v", setup)
			}
		})
	}
}

func TestTridentInstallMergesActiveEntries(t *testing.T) {
	root := t.TempDir()
	instance := tridentTestInstall(t, root, map[string]string{"mods/first.jar": "first"}, ModeInstall, func(m *PackManifest) {
		m.Mods = []PackMod{{File: "first.jar", Source: &PackModSource{Provider: "curseforge", ProjectID: "123", VersionID: "456"}}}
	})
	tridentTestWrite(t, instance, "build/mods/orphan.jar", "old projection")
	tridentTestWrite(t, instance, "build/trident.import.json", `{"version":1,"files":["mods/orphan.jar"]}`)
	tridentTestInstall(t, instance, map[string]string{"mods/second.jar": "second"}, ModeInstall, func(m *PackManifest) {
		m.Name = "Merged"
		m.Mods = []PackMod{{File: "second.jar", Source: &PackModSource{Provider: "modrinth", ProjectID: "second", VersionID: "v1"}}}
	})
	profile := decodeJSONFile(t, filepath.Join(instance, "profile.json"))
	setup := profile["setup"].(map[string]any)
	entries := setup["packages"].([]any)
	if len(entries) != 2 {
		t.Fatalf("missing merged entries: %v", entries)
	}
	for _, item := range entries {
		if item.(map[string]any)["source"] != setup["source"] {
			t.Fatal("merged entry retained previous group")
		}
	}
	if matches, _ := filepath.Glob(filepath.Join(instance, "import/mods/*.jar")); len(matches) != 0 {
		t.Fatalf("converted jars remain: %v", matches)
	}
	tridentTestInstall(t, instance, map[string]string{"config.txt": "replacement"}, ModeUpdate)
	if _, err := os.Stat(filepath.Join(instance, "build/mods/orphan.jar")); !os.IsNotExist(err) {
		t.Fatal("merge lost ownership of an orphaned runtime mod")
	}
}

func TestTridentRollbackRestoresSourceAndLive(t *testing.T) {
	instance := t.TempDir()
	work := filepath.Join(instance, ".work")
	tridentTestWrite(t, instance, "import/config.txt", "old source")
	tridentTestWrite(t, instance, "build/config.txt", "edited live")
	tridentTestWrite(t, instance, "build/trident.import.json", `{"version":1,"files":["config.txt"]}`)
	tridentTestWrite(t, work, "import/config.txt", "new source")
	tridentTestWrite(t, instance, installedManifestName, `{"id":"old","version":"1"}`)
	tridentTestWrite(t, work, installedManifestName, `{"id":"new","version":"2"}`)
	// A directory at the profile destination makes the final promotion fail.
	if err := os.Mkdir(filepath.Join(instance, "profile.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, retain, err := commitTridentInstall(instance, work, map[string]any{"name": "New"}, ModeUpdate, map[string]string{"config.txt": "config.txt"})
	if err == nil || retain {
		t.Fatalf("expected a fully rolled-back promotion failure: %v, retain=%v", err, retain)
	}
	for relative, want := range map[string]string{
		"import/config.txt": "old source", "build/config.txt": "edited live",
		"build/trident.import.json": `{"version":1,"files":["config.txt"]}`,
		installedManifestName:       `{"id":"old","version":"1"}`,
	} {
		data, err := os.ReadFile(filepath.Join(instance, relative))
		if err != nil || string(data) != want {
			t.Fatalf("rollback %s = %q, %v", relative, data, err)
		}
	}
}

func TestTridentHomeBOM(t *testing.T) {
	home, cwd := t.TempDir(), t.TempDir()
	want := filepath.Join(home, "自定义")
	utf8Data := append([]byte{0xef, 0xbb, 0xbf}, []byte(want+"\r\n")...)
	utf16Data := []byte{0xff, 0xfe}
	for _, code := range utf16.Encode([]rune(want + "\r\n")) {
		utf16Data = binary.LittleEndian.AppendUint16(utf16Data, code)
	}
	for _, data := range [][]byte{utf8Data, utf16Data} {
		if err := os.WriteFile(filepath.Join(home, ".trident.home"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if got := tridentDataHome(home, cwd, "darwin", func(string) string { return "" }); got != want {
			t.Fatalf("BOM override resolved to %q, want %q", got, want)
		}
	}
}
