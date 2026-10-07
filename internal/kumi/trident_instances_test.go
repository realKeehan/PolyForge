package kumi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTridentInstallKeepsUnmanagedCollisions(t *testing.T) {
	root := t.TempDir()
	original := `{"setup":{"source":"modpack://local/Test%20Pack"},"custom":"user"}`
	tridentTestWrite(t, root, "instances/test-pack/profile.json", original)
	tridentTestWrite(t, root, "instances/test-pack/import/user.txt", "user")
	tridentTestWrite(t, root, "instances/test-pack_", "occupied file")

	instance := tridentTestInstall(t, root, map[string]string{"config.txt": "v1"}, ModeAuto)
	want := filepath.Join(root, "instances", "test-pack__")
	if instance != want {
		t.Fatalf("new instance = %s, want %s", instance, want)
	}
	// A subsequent import through either the data root or an unrelated instance
	// finds the managed suffix instead of allocating yet another directory.
	for _, chosen := range []string{root, filepath.Join(root, "instances", "test-pack")} {
		got := tridentTestInstall(t, chosen, map[string]string{"config.txt": "v2"}, ModeAuto)
		if got != instance {
			t.Fatalf("reimport through %s used %s, want %s", chosen, got, instance)
		}
	}
	for relative, expected := range map[string]string{
		"instances/test-pack/profile.json":    original,
		"instances/test-pack/import/user.txt": "user",
		"instances/test-pack_":                "occupied file",
	} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil || string(data) != expected {
			t.Errorf("unmanaged path %s changed: %q, %v", relative, data, err)
		}
	}
}

func TestTridentInstallDistinguishesPackIDsAfterSanitizing(t *testing.T) {
	root := t.TempDir()
	original := tridentTestInstall(t, root, map[string]string{"old.txt": "original"}, ModeAuto, func(m *PackManifest) {
		m.ID = "Same:Pack"
	})
	if original != filepath.Join(root, "instances", "same_pack") {
		t.Fatalf("unsanitized key: %s", original)
	}
	replacement := tridentTestInstall(t, original, map[string]string{"new.txt": "new"}, ModeAuto, func(m *PackManifest) {
		m.ID = "Same?Pack"
	})
	if replacement != filepath.Join(root, "instances", "same_pack_") {
		t.Fatalf("different pack reused a colliding key: %s", replacement)
	}
	old, err := readInstalledPackManifest(original)
	if err != nil || old.ID != "Same:Pack" {
		t.Fatalf("original record changed: %+v, %v", old, err)
	}
	if data, err := os.ReadFile(filepath.Join(original, "import", "old.txt")); err != nil || string(data) != "original" {
		t.Fatalf("original files changed: %q, %v", data, err)
	}
}

func TestTridentManagedInstanceMissingSourceFails(t *testing.T) {
	instance := tridentTestInstall(t, t.TempDir(), map[string]string{"config.txt": "old"}, ModeAuto)
	record, err := readInstalledPackRecord(instance)
	if err != nil {
		t.Fatal(err)
	}
	// Reproduce the legacy acceptance instance: a valid PolyForge record and a
	// native profile whose source is null. It must fail instead of migrating.
	if err := writeJSONFile(filepath.Join(instance, installedManifestName), &record.PackManifest); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(instance, "profile.json")
	profile := decodeJSONFile(t, profilePath)
	profile["setup"].(map[string]any)["source"] = nil
	if err := writeJSONFile(profilePath, profile); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	pack := buildTestPack(t, map[string]string{"config.txt": "new"}, true)
	reader, err := openPackReader(pack)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := readZipPackManifest(reader)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := planPackInstall("polymerium", instance, manifest, nil, ModeAuto)
	if err != nil || plan.Mode != ModeUpdate || plan.InstanceDir != instance {
		t.Fatalf("PolyForge planning must accept the managed record without validating source: %+v, %v", plan, err)
	}
	if _, err := installTridentPack(pack, plan, manifest, nil); err == nil {
		t.Fatal("missing source was silently migrated or bypassed with a new instance")
	}
	after, err := os.ReadFile(profilePath)
	if err != nil || string(after) != string(before) {
		t.Fatalf("failed update changed the profile: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(instance, "import", "config.txt")); err != nil || string(data) != "old" {
		t.Fatalf("failed update changed source files: %q, %v", data, err)
	}
}
