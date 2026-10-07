package kumi

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Registration and self-destruct use the real PolyForge pipeline, with its
// config redirected so these checks cannot touch the user's installed packs.
func TestTridentPolyForgeInstallRecord(t *testing.T) {
	config := t.TempDir()
	t.Setenv("APPDATA", config)
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("HOME", config)
	root := t.TempDir()
	pack := buildTestPack(t, map[string]string{
		"mods/managed.jar": "managed", "mods/bundled.jar": "bundled", "config.txt": "config",
	}, true, func(m *PackManifest) {
		m.Mods = []PackMod{
			{File: "managed.jar", ModID: "managed", Source: &PackModSource{Provider: "modrinth", ProjectID: "managed", VersionID: "v1"}},
			{File: "bundled.jar", ModID: "bundled"},
		}
	})
	result := NewResult()
	if !NewService().extractAndVerifyPack(result, pack, root, "polymerium", ModeAuto) {
		t.Fatalf("install failed: %+v", result.Messages)
	}
	instance := filepath.Join(root, "instances", "test-pack")
	record, err := readInstalledPackRecord(instance)
	if err != nil || record.Launcher != "polymerium" || len(record.Mods) != 2 || len(record.Overrides.Files) != 2 {
		t.Fatalf("install record = %+v, %v", record, err)
	}
	for _, file := range record.Overrides.Files {
		if file.Path == "mods/managed.jar" {
			t.Fatal("converted repository package was recorded as an import file")
		}
	}
	for _, relative := range []string{"import/" + installedManifestName, "build/" + installedManifestName} {
		if _, err := os.Stat(filepath.Join(instance, relative)); !os.IsNotExist(err) {
			t.Fatalf("PolyForge record leaked into Trident layers: %s, %v", relative, err)
		}
	}
	installs := loadInstalledPacks().Installs
	if len(installs) != 1 || installs[0].Target != instance || installs[0].ID != "test-pack" {
		t.Fatalf("install registration = %+v", installs)
	}
	if _, report, err := VerifyInstalledPack(instance); err != nil || !report.OK() || report.Checked != 2 {
		t.Fatalf("verify imports = %+v, %v", report, err)
	}
	tridentTestWrite(t, instance, "import/config.txt", "tampered")
	if _, report, err := VerifyInstalledPack(instance); err != nil || report.OK() || len(report.Issues) != 1 || report.Issues[0].Path != "config.txt" {
		t.Fatalf("verify tampered import = %+v, %v", report, err)
	}
	diff, old, err := PlanPackUpdate(instance, &PackManifest{Mods: []PackMod{{ModID: "managed", Version: "2"}}})
	if err != nil || old.ID != "test-pack" || len(diff.Removed) != 1 || len(diff.Changed) != 1 {
		t.Fatalf("update preview = %+v, %v", diff, err)
	}
	// A replacement updates the local record and the app's registration.
	replacement := buildTestPack(t, map[string]string{"new.txt": "new"}, true, func(m *PackManifest) {
		m.Version = "2"
	})
	result = NewResult()
	if !NewService().extractAndVerifyPack(result, replacement, instance, "polymerium", ModeAuto) {
		t.Fatalf("replacement failed: %+v", result.Messages)
	}
	record, err = readInstalledPackRecord(instance)
	installs = loadInstalledPacks().Installs
	if err != nil || record.Version != "2" || len(installs) != 1 || installs[0].Version != "2" {
		t.Fatalf("replacement record = %+v, registrations = %+v, %v", record, installs, err)
	}
}

func TestTridentPolyForgeRemoteRemoval(t *testing.T) {
	config := t.TempDir()
	t.Setenv("APPDATA", config)
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("HOME", config)
	instance := tridentTestInstall(t, t.TempDir(), map[string]string{
		"mods/managed.jar": "managed", "mods/bundled.jar": "bundled",
	}, ModeInstall, func(m *PackManifest) {
		m.Mods = []PackMod{
			{File: "managed.jar", Source: &PackModSource{Provider: "modrinth", ProjectID: "managed", VersionID: "v1"}},
			{File: "bundled.jar"},
		}
	})
	recordInstalledPack("test-pack", "Test Pack", "1.0.0", instance)
	tridentTestWrite(t, instance, "build/mods/bundled.jar", "runtime bundled")
	tridentTestWrite(t, instance, "import/mods/user.jar", "user")
	tridentTestWrite(t, instance, "build/mods/user.jar", "runtime user")
	tridentTestWrite(t, instance, "persist/mods/bundled.jar", "persist")
	profile := decodeJSONFile(t, filepath.Join(instance, "profile.json"))
	setup := profile["setup"].(map[string]any)
	manual := map[string]any{"pref": "pref://modrinth/managed@v1", "enabled": true}
	other := map[string]any{"pref": "pref://modrinth/managed@v1", "enabled": true, "source": "collection://by-name/Other"}
	setup["packages"] = append(setup["packages"].([]any), manual, other)
	if err := writeJSONFile(filepath.Join(instance, "profile.json"), profile); err != nil {
		t.Fatal(err)
	}
	remote := &RemoteManifest{Modpacks: []RemotePack{{
		ID: "test-pack", RemoveMods: []string{"managed.jar", "bundled.jar", "user.jar", "../user.jar"},
	}}}
	if notes := applySelfDestruct(remote); len(notes) != 2 {
		t.Fatalf("removal notes = %v", notes)
	}
	profile = decodeJSONFile(t, filepath.Join(instance, "profile.json"))
	packages := profile["setup"].(map[string]any)["packages"]
	if !reflect.DeepEqual(packages, []any{manual, other}) {
		t.Fatalf("unrelated package entries changed: %#v", packages)
	}
	for _, relative := range []string{"import/mods/bundled.jar", "build/mods/bundled.jar"} {
		if _, err := os.Stat(filepath.Join(instance, relative)); !os.IsNotExist(err) {
			t.Fatalf("pack file remains: %s, %v", relative, err)
		}
	}
	for _, relative := range []string{"import/mods/user.jar", "build/mods/user.jar", "persist/mods/bundled.jar"} {
		if _, err := os.Stat(filepath.Join(instance, relative)); err != nil {
			t.Fatalf("unowned file changed: %s, %v", relative, err)
		}
	}
	if notes := applySelfDestruct(remote); len(notes) != 0 {
		t.Fatalf("repeated removal = %v", notes)
	}
}

func TestTridentPolyForgeRemoteRemovalRequiresCurrentRecord(t *testing.T) {
	for _, changed := range []string{"marker removed", "pack replaced", "native source replaced"} {
		t.Run(changed, func(t *testing.T) {
			config := t.TempDir()
			t.Setenv("APPDATA", config)
			t.Setenv("XDG_CONFIG_HOME", config)
			t.Setenv("HOME", config)
			instance := tridentTestInstall(t, t.TempDir(), map[string]string{"mods/bundled.jar": "bundled"}, ModeInstall)
			recordInstalledPack("test-pack", "Test Pack", "1", instance)
			switch changed {
			case "marker removed":
				if err := os.Remove(filepath.Join(instance, installedManifestName)); err != nil {
					t.Fatal(err)
				}
			case "pack replaced":
				record, _ := readInstalledPackRecord(instance)
				record.ID = "replacement"
				if err := writeJSONFile(filepath.Join(instance, installedManifestName), record); err != nil {
					t.Fatal(err)
				}
			case "native source replaced":
				profile := decodeJSONFile(t, filepath.Join(instance, "profile.json"))
				profile["setup"].(map[string]any)["source"] = "modpack://local/Replacement"
				if err := writeJSONFile(filepath.Join(instance, "profile.json"), profile); err != nil {
					t.Fatal(err)
				}
			}
			applySelfDestruct(&RemoteManifest{Modpacks: []RemotePack{{ID: "test-pack", RemoveMods: []string{"bundled.jar"}}}})
			if _, err := os.Stat(filepath.Join(instance, "import/mods/bundled.jar")); err != nil {
				t.Fatalf("stale record allowed removal: %v", err)
			}
		})
	}
}
