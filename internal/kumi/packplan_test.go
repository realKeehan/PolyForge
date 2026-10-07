package kumi

import (
	"os"
	"path/filepath"
	"testing"
)

// File-based launchers keep the original directory and resolveInstallMode
// policy, including an install into an existing directory without a record.
func TestPlanPackInstallRetainsFileBasedBehavior(t *testing.T) {
	root := t.TempDir()
	manifest := &PackManifest{ID: "new-pack", Name: "Example", Version: "2"}
	instance, content := PlanInstallDirs("multimc", root, "Example")
	if err := os.MkdirAll(content, 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := planPackInstall("multimc", root, manifest, nil, ModeAuto)
	if err != nil || plan.Mode != ModeInstall || plan.InstanceDir != instance || plan.ContentDir != content || plan.RecordDir != content {
		t.Fatalf("unrecorded file-based plan = %+v, %v", plan, err)
	}
	// The original resolver accepts any valid existing record; the opt-in
	// managed-instance identity rule must not change that legacy behavior.
	if err := writeJSONFile(filepath.Join(content, installedManifestName), &PackManifest{ID: "old-pack", Version: "1"}); err != nil {
		t.Fatal(err)
	}
	plan, err = planPackInstall("multimc", root, manifest, nil, ModeAuto)
	if err != nil || plan.Mode != ModeUpdate || plan.InstanceDir != instance || plan.Previous == nil || plan.Previous.ID != "old-pack" {
		t.Fatalf("recorded file-based plan = %+v, %v", plan, err)
	}
}

func TestFileBasedInstallRetainsUpdateAndRegistration(t *testing.T) {
	config := t.TempDir()
	t.Setenv("APPDATA", config)
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("HOME", config)
	root := t.TempDir()
	service := NewService()
	first := buildTestPack(t, map[string]string{"old.txt": "v1"}, true)
	result := NewResult()
	if !service.extractAndVerifyPack(result, first, root, "multimc", ModeAuto) {
		t.Fatalf("first install failed: %+v", result.Messages)
	}
	instance, content := PlanInstallDirs("multimc", root, "Test Pack")
	tridentTestWrite(t, content, "user.txt", "user")
	second := buildTestPack(t, map[string]string{"new.txt": "v2"}, true, func(m *PackManifest) {
		m.Version = "2"
	})
	result = NewResult()
	if !service.extractAndVerifyPack(result, second, root, "multimc", ModeAuto) {
		t.Fatalf("update failed: %+v", result.Messages)
	}
	if _, err := os.Stat(filepath.Join(content, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("retired pack file was not removed: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(content, "user.txt")); err != nil || string(data) != "user" {
		t.Fatalf("untracked file changed: %q, %v", data, err)
	}
	record, err := readInstalledPackManifest(content)
	if err != nil || record.Version != "2" {
		t.Fatalf("updated record = %+v, %v", record, err)
	}
	installs := loadInstalledPacks().Installs
	if len(installs) != 1 || installs[0].Target != content || installs[0].Version != "2" {
		t.Fatalf("updated registration = %+v", installs)
	}
	if _, err := os.Stat(filepath.Join(instance, "mmc-pack.json")); err != nil {
		t.Fatalf("profile generation was bypassed: %v", err)
	}
}
