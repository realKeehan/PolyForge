package kumi

import (
	"os"
	"path/filepath"
	"testing"
)

// applySelfDestruct must delete only the exact marked files inside a recorded
// install's mods/ folder, leave everything else alone, and refuse path
// traversal.
func TestApplySelfDestruct(t *testing.T) {
	tmp := t.TempDir()
	// Redirect the installed-packs record + config dir into tmp.
	t.Setenv("APPDATA", tmp)                      // Windows UserConfigDir
	t.Setenv("XDG_CONFIG_HOME", tmp)              // Linux UserConfigDir
	t.Setenv("HOME", tmp)                         // macOS/Linux fallback

	target := filepath.Join(tmp, "instance")
	modsDir := filepath.Join(target, "mods")
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Marker that proves this is one of our installs.
	if err := os.WriteFile(filepath.Join(target, ".polyforge-pack.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(modsDir, "free-mod.jar")
	kill := filepath.Join(modsDir, "proprietary.jar")
	outside := filepath.Join(target, "secret.txt")
	for _, p := range []string{keep, kill, outside} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	recordInstalledPack("turtel-smp", "Turtel SMP", "1.0.0", target)

	manifest := &RemoteManifest{
		Modpacks: []RemotePack{{
			ID: "turtel-smp",
			// includes a traversal attempt that must be ignored
			RemoveMods: []string{"proprietary.jar", "../secret.txt", "not-there.jar"},
		}},
	}

	notes := applySelfDestruct(manifest)

	if _, err := os.Stat(kill); !os.IsNotExist(err) {
		t.Errorf("marked mod was not removed: %v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("unmarked mod should remain: %v", err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("traversal target must be untouched: %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 removal note, got %d: %v", len(notes), notes)
	}

	// Second pass is a no-op (already removed) — no notes, no panic.
	if notes2 := applySelfDestruct(manifest); len(notes2) != 0 {
		t.Errorf("second pass should remove nothing, got %v", notes2)
	}
}

// A folder without our marker must never be touched even if recorded.
func TestApplySelfDestructRequiresMarker(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	target := filepath.Join(tmp, "instance")
	modsDir := filepath.Join(target, "mods")
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	kill := filepath.Join(modsDir, "proprietary.jar")
	if err := os.WriteFile(kill, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	recordInstalledPack("turtel-smp", "Turtel SMP", "1.0.0", target)

	manifest := &RemoteManifest{Modpacks: []RemotePack{{ID: "turtel-smp", RemoveMods: []string{"proprietary.jar"}}}}
	applySelfDestruct(manifest)

	if _, err := os.Stat(kill); err != nil {
		t.Errorf("file removed despite missing marker: %v", err)
	}
}
