package kumi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCleanCandidatesDropsEmptyAndDedups(t *testing.T) {
	got := cleanCandidates(
		"",
		"   ",
		`C:\A\B`,
		`C:\A\B\`, // same as above after Clean
		`c:\a\b`,  // same case-insensitively
		`C:\A\C`,
	)
	want := []string{filepath.Clean(`C:\A\B`), filepath.Clean(`C:\A\C`)}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLocalLowDirDerivesFromLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", filepath.Join("X:", "Users", "someone", "AppData", "Local"))
	got := localLowDir()
	want := filepath.Clean(filepath.Join("X:", "Users", "someone", "AppData", "LocalLow"))
	if filepath.Clean(got) != want {
		t.Errorf("localLowDir = %q, want %q", got, want)
	}

	t.Setenv("LOCALAPPDATA", "")
	if got := localLowDir(); got != "" {
		t.Errorf("localLowDir with unset LOCALAPPDATA = %q, want empty", got)
	}
}

// TestWithExeDiscoveryShortCircuits verifies that when a fixed candidate
// already holds the executable, the fixed list is returned unchanged (no
// cache read, no scan, no appended discovery entries).
func TestWithExeDiscoveryShortCircuits(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "MultiMC.exe"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fixed := cleanCandidates(dir, `C:\Does\Not\Exist\MultiMC`)
	got := withExeDiscovery("multimc", fixed)
	if len(got) != len(fixed) {
		t.Fatalf("expected fixed list unchanged, got %v", got)
	}
	if got[0] != filepath.Clean(dir) {
		t.Errorf("got[0] = %q, want %q", got[0], filepath.Clean(dir))
	}
}

// TestMultiMCCandidatesIncludesPortablePaths guards against the old bug where
// the candidate list was made of broad parent directories (USERPROFILE,
// ProgramFiles) rather than actual MultiMC locations.
func TestMultiMCCandidatesIncludesPortablePaths(t *testing.T) {
	t.Setenv("USERPROFILE", filepath.Join("X:", "Users", "tester"))
	t.Setenv("LOCALAPPDATA", filepath.Join("X:", "Users", "tester", "AppData", "Local"))

	got := multiMCCandidates("")
	joined := strings.ToLower(strings.Join(got, "|"))

	for _, want := range []string{
		strings.ToLower(filepath.Join("X:", "Users", "tester", "MultiMC")),
		strings.ToLower(filepath.Join("X:", "Users", "tester", "Desktop", "MultiMC")),
		strings.ToLower(filepath.Join("X:", "Users", "tester", "Downloads", "MultiMC")),
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("candidate list missing %q; got %v", want, got)
		}
	}

	// The bare profile root must never be offered as a candidate — that was
	// the false-positive source the reference flagged.
	for _, bad := range got {
		if filepath.Clean(bad) == filepath.Clean(filepath.Join("X:", "Users", "tester")) {
			t.Errorf("candidate list contains bare USERPROFILE root %q", bad)
		}
	}
}

// TestMachineTestCandidatePaths pins the data-dir names observed in the
// MachineTest_01 reference dump (TemporaryDetectRef/MachineTest_01). Each of
// these was missed by an earlier candidate list: gdlauncher_carbon (not
// "GDLauncher Carbon"), .dawn (Feather → Dawn rebrand), QWERTZ-Launcher,
// .minecraftx (XMCL instances) and Trident (Polymerium instances).
func TestMachineTestCandidatePaths(t *testing.T) {
	home := filepath.Join("X:", "Users", "tester")
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))

	cases := []struct {
		name string
		got  []string
		want string
	}{
		{"gdlauncher", gdLauncherCandidates(""), filepath.Join(home, "AppData", "Roaming", "gdlauncher_carbon")},
		{"dawn", dawnCandidates(""), filepath.Join(home, "AppData", "Roaming", ".dawn")},
		{"qwertz", qwertzCandidates(""), filepath.Join(home, "AppData", "Roaming", "QWERTZ-Launcher")},
		{"xmcl", xmclCandidates(""), filepath.Join(home, ".minecraftx")},
		{"polymerium", polymeriumCandidates(""), filepath.Join(home, "AppData", "Local", "Trident")},
	}
	for _, tc := range cases {
		joined := strings.ToLower(strings.Join(tc.got, "|"))
		if !strings.Contains(joined, strings.ToLower(tc.want)) {
			t.Errorf("%s candidates missing %q; got %v", tc.name, tc.want, tc.got)
		}
	}
}

func TestCommonScanRootsIncludesOneDriveDocuments(t *testing.T) {
	home := filepath.Join("X:", "Users", "vtori")
	oneDrive := filepath.Join(home, "OneDrive")
	t.Setenv("USERPROFILE", home)
	t.Setenv("OneDrive", oneDrive)
	t.Setenv("OneDriveConsumer", "")
	t.Setenv("OneDriveCommercial", "")

	got := commonScanRoots()
	joined := strings.ToLower(strings.Join(got, "|"))
	want := strings.ToLower(filepath.Join(oneDrive, "Documents"))
	if !strings.Contains(joined, want) {
		t.Fatalf("commonScanRoots missing OneDrive Documents root %q; got %v", want, got)
	}
}

func TestShortcutRootsIncludesOneDriveDesktop(t *testing.T) {
	base := t.TempDir()
	oneDrive := filepath.Join(base, "OneDrive")
	desktop := filepath.Join(oneDrive, "Desktop")
	if err := os.MkdirAll(desktop, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USERPROFILE", base)
	t.Setenv("APPDATA", filepath.Join(base, "AppData", "Roaming"))
	t.Setenv("ProgramData", filepath.Join(base, "ProgramData"))
	t.Setenv("OneDrive", oneDrive)
	t.Setenv("OneDriveConsumer", "")
	t.Setenv("OneDriveCommercial", "")

	got := shortcutRoots()
	found := false
	for _, root := range got {
		if root == desktop {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("shortcutRoots missing OneDrive Desktop root %q; got %v", desktop, got)
	}
}

// isolateDiscoveryEnv points every root the discovery pipeline touches
// (launcher cache, shortcut roots) at throwaway temp dirs so tests neither
// read nor write the real machine state.
func isolateDiscoveryEnv(t *testing.T) (appData string) {
	t.Helper()
	base := t.TempDir()
	appData = filepath.Join(base, "Roaming")
	for _, dir := range []string{appData, filepath.Join(base, "Local"), filepath.Join(base, "ProgramData")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("APPDATA", appData) // os.UserConfigDir → launcher cache location
	t.Setenv("LOCALAPPDATA", filepath.Join(base, "Local"))
	t.Setenv("USERPROFILE", base)
	t.Setenv("ProgramData", filepath.Join(base, "ProgramData"))
	return appData
}

// TestDiscoverLauncherDirsCacheHit verifies the boot-fast path: a previously
// found (and still valid) location is returned straight from the cache.
func TestDiscoverLauncherDirsCacheHit(t *testing.T) {
	isolateDiscoveryEnv(t)

	exeDir := t.TempDir()
	exe := filepath.Join(exeDir, "UltimMC.exe")
	if err := os.WriteFile(exe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cache, _ := LoadCache()
	UpsertCandidate(cache, &Candidate{
		Launcher: LauncherID("ultimmc"), Path: exe, Kind: "exe",
		Evidence: EvScan, Confidence: "low",
		LastUsed: time.Now(), LastOK: time.Now(), HashHint: PathHint(exe),
	})
	if err := SaveCache(cache); err != nil {
		t.Fatal(err)
	}

	got := discoverLauncherDirs("ultimmc")
	if len(got) != 1 || got[0] != exeDir {
		t.Fatalf("discoverLauncherDirs = %v, want [%s]", got, exeDir)
	}
}

// TestDiscoverLauncherDirsRevalidatesMovedInstall verifies the "user moved
// the folder" flow: the stale cache entry fails validation and discovery
// falls back to a fresh search (here fed by a Start Menu shortcut pointing
// at the new location), whose result is persisted for the next boot.
func TestDiscoverLauncherDirsRevalidatesMovedInstall(t *testing.T) {
	appData := isolateDiscoveryEnv(t)

	oldDir := t.TempDir()
	newDir := t.TempDir()
	oldExe := filepath.Join(oldDir, "UltimMC.exe")
	newExe := filepath.Join(newDir, "UltimMC.exe")
	if err := os.WriteFile(newExe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Cache still points at the old (now gone) location.
	cache, _ := LoadCache()
	UpsertCandidate(cache, &Candidate{
		Launcher: LauncherID("ultimmc"), Path: oldExe, Kind: "exe",
		Evidence: EvScan, Confidence: "low",
		LastUsed: time.Now(), LastOK: time.Now(), HashHint: PathHint(oldExe),
	})
	if err := SaveCache(cache); err != nil {
		t.Fatal(err)
	}

	// A Start Menu shortcut points at the new location.
	startMenu := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs")
	if err := os.MkdirAll(startMenu, 0o755); err != nil {
		t.Fatal(err)
	}
	lnk := filepath.Join(startMenu, "UltimMC.lnk")
	if err := os.WriteFile(lnk, buildLnkWithLinkInfo(newExe), 0o644); err != nil {
		t.Fatal(err)
	}

	got := discoverLauncherDirs("ultimmc")
	if len(got) != 1 || got[0] != newDir {
		t.Fatalf("discoverLauncherDirs = %v, want [%s]", got, newDir)
	}

	// The fresh find must be persisted so the next boot skips the search.
	cache, _ = LoadCache()
	cand := BestValidCachedCandidate(cache, LauncherID("ultimmc"), ValidateExeByName("UltimMC.exe"))
	if cand == nil || cand.Path != newExe {
		t.Fatalf("cache not updated with new location; got %+v", cand)
	}
}
