package kumi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func modeTestManifest(version string, files ...PackFile) *PackManifest {
	return &PackManifest{
		SchemaVersion: 1,
		ID:            "smoke-pack",
		Name:          "Smoke Pack",
		Version:       version,
		Overrides:     PackOverrides{Files: files, FileCount: len(files)},
	}
}

func writeInstalledRecord(t *testing.T, dir string, m *PackManifest) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, installedManifestName), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeGameFile(t *testing.T, dir, rel string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseInstallMode(t *testing.T) {
	cases := map[string]InstallMode{
		"":             ModeAuto,
		"auto":         ModeAuto,
		"Install":      ModeInstall,
		"fresh":        ModeInstall,
		"UPDATE":       ModeUpdate,
		"upgrade":      ModeUpdate,
		"reinstall":    ModeReinstall,
		"repair":       ModeReinstall,
		"clean":        ModeClean,
		"cleaninstall": ModeClean,
		"lolwhat":      ModeAuto, // unknown values must never break installs
	}
	for in, want := range cases {
		if got := ParseInstallMode(in); got != want {
			t.Errorf("ParseInstallMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveInstallMode(t *testing.T) {
	empty := t.TempDir()

	installed := t.TempDir()
	writeInstalledRecord(t, installed, modeTestManifest("1.0.0"))

	cases := []struct {
		name      string
		requested InstallMode
		dir       string
		want      InstallMode
		wantOld   bool
	}{
		{"auto on empty dir", ModeAuto, empty, ModeInstall, false},
		{"auto on installed dir", ModeAuto, installed, ModeUpdate, true},
		{"update without record falls back", ModeUpdate, empty, ModeInstall, false},
		{"reinstall without record falls back", ModeReinstall, empty, ModeInstall, false},
		{"update with record", ModeUpdate, installed, ModeUpdate, true},
		{"reinstall with record", ModeReinstall, installed, ModeReinstall, true},
		{"clean stays clean either way", ModeClean, empty, ModeClean, false},
		{"install stays install on installed dir", ModeInstall, installed, ModeInstall, true},
	}
	for _, tc := range cases {
		got, old, _ := resolveInstallMode(tc.requested, tc.dir)
		if got != tc.want {
			t.Errorf("%s: mode = %q, want %q", tc.name, got, tc.want)
		}
		if (old != nil) != tc.wantOld {
			t.Errorf("%s: old manifest present = %v, want %v", tc.name, old != nil, tc.wantOld)
		}
	}
}

// TestRemoveStalePackFiles pins the update/reinstall sweep: only files the
// OLD pack shipped and the NEW pack dropped are removed; user data and
// user-added mods are untouchable by construction, and directories the sweep
// empties are pruned.
func TestRemoveStalePackFiles(t *testing.T) {
	game := t.TempDir()

	old := modeTestManifest("1.0.0",
		PackFile{Path: "mods/dropped-mod.jar"},
		PackFile{Path: "mods/kept-mod.jar"},
		PackFile{Path: "config/dropped/nested.toml"},
		PackFile{Path: "already-gone.txt"}, // user deleted it — not an error
	)
	latest := modeTestManifest("2.0.0",
		PackFile{Path: "mods/kept-mod.jar"},
		PackFile{Path: "mods/new-mod.jar"},
	)

	for _, rel := range []string{
		"mods/dropped-mod.jar",
		"mods/kept-mod.jar",
		"mods/user-added.jar",
		"config/dropped/nested.toml",
		"saves/world/level.dat",
	} {
		writeGameFile(t, game, rel)
	}

	removed := removeStalePackFiles(game, old, latest)
	if removed != 2 {
		t.Fatalf("removed = %d, want 2 (dropped-mod.jar + nested.toml)", removed)
	}

	for _, gone := range []string{"mods/dropped-mod.jar", "config/dropped/nested.toml", "config/dropped", "config"} {
		if _, err := os.Stat(filepath.Join(game, filepath.FromSlash(gone))); !os.IsNotExist(err) {
			t.Errorf("%s should have been removed/pruned", gone)
		}
	}
	for _, kept := range []string{"mods/kept-mod.jar", "mods/user-added.jar", "saves/world/level.dat"} {
		if _, err := os.Stat(filepath.Join(game, filepath.FromSlash(kept))); err != nil {
			t.Errorf("%s should have survived the sweep: %v", kept, err)
		}
	}
}

// TestQuarantineGameDir pins clean mode: everything moves into a timestamped
// snapshot under .polyforge-quarantine, nothing is deleted, and earlier
// snapshots survive a second clean.
func TestQuarantineGameDir(t *testing.T) {
	game := t.TempDir()
	for _, rel := range []string{"mods/a.jar", "options.txt", "saves/world/level.dat"} {
		writeGameFile(t, game, rel)
	}

	moved, quarantine, err := quarantineGameDir(game)
	if err != nil {
		t.Fatal(err)
	}
	if moved != 3 {
		t.Fatalf("moved = %d, want 3 top-level entries", moved)
	}
	if !strings.HasPrefix(quarantine, filepath.Join(game, quarantineDirName)) {
		t.Fatalf("quarantine landed at %s, want under %s", quarantine, quarantineDirName)
	}

	// Game dir now holds only the quarantine root; the content is inside it.
	entries, err := os.ReadDir(game)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != quarantineDirName {
		t.Fatalf("game dir entries = %v, want only %s", entries, quarantineDirName)
	}
	if _, err := os.Stat(filepath.Join(quarantine, "saves", "world", "level.dat")); err != nil {
		t.Errorf("quarantined save missing: %v", err)
	}

	// Second clean over new content must not disturb the first snapshot.
	writeGameFile(t, game, "mods/b.jar")
	if _, _, err := quarantineGameDir(game); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(quarantine, "options.txt")); err != nil {
		t.Errorf("first snapshot was disturbed by the second clean: %v", err)
	}

	// Empty dir (or one holding only quarantine) is a no-op.
	moved, path, err := quarantineGameDir(game)
	if err != nil || moved != 0 || path != "" {
		t.Errorf("quarantine of clean dir = (%d, %q, %v), want no-op", moved, path, err)
	}
}

func TestPrepareGameDirCleanAndUpdateNotes(t *testing.T) {
	// Update against a pre-checksum record: nothing to sweep, but say so.
	game := t.TempDir()
	oldNoFiles := modeTestManifest("0.9.0")
	notes, err := prepareGameDir(ModeUpdate, game, oldNoFiles, modeTestManifest("1.0.0"))
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "predates per-file manifests") {
		t.Errorf("notes = %v, want the pre-checksum limitation note", notes)
	}

	// Clean over existing content reports the quarantine location.
	writeGameFile(t, game, "mods/a.jar")
	notes, err = prepareGameDir(ModeClean, game, nil, modeTestManifest("1.0.0"))
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], quarantineDirName) {
		t.Errorf("notes = %v, want the quarantine location note", notes)
	}

	// Plain install prepares nothing.
	notes, err = prepareGameDir(ModeInstall, game, nil, modeTestManifest("1.0.0"))
	if err != nil || len(notes) != 0 {
		t.Errorf("install prepare = (%v, %v), want no-op", notes, err)
	}
}

func TestPlanPackUpdate(t *testing.T) {
	game := t.TempDir()
	installed := modeTestManifest("1.0.0")
	installed.Mods = []PackMod{
		{File: "a.jar", ModID: "a", Name: "A", Version: "1"},
		{File: "b.jar", ModID: "b", Name: "B", Version: "1"},
	}
	writeInstalledRecord(t, game, installed)

	latest := modeTestManifest("2.0.0")
	latest.Mods = []PackMod{
		{File: "a.jar", ModID: "a", Name: "A", Version: "2"}, // changed
		{File: "c.jar", ModID: "c", Name: "C", Version: "1"}, // added
	}

	diff, old, err := PlanPackUpdate(game, latest)
	if err != nil {
		t.Fatal(err)
	}
	if old.Version != "1.0.0" {
		t.Errorf("installed version = %s, want 1.0.0", old.Version)
	}
	if len(diff.Added) != 1 || len(diff.Removed) != 1 || len(diff.Changed) != 1 {
		t.Errorf("diff = %+v, want 1 added / 1 removed / 1 changed", diff)
	}

	// No record → the caller learns nothing is installed here.
	if _, _, err := PlanPackUpdate(t.TempDir(), latest); err == nil {
		t.Error("PlanPackUpdate on empty dir should error")
	}
}
