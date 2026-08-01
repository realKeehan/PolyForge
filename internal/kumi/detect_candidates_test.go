package kumi

import (
	"context"
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

// ── Real-install preference (machine test 2) ─────

// TestScanForExesPrefersMarkedDir pins the machine test 2 defect: a leftover
// copy of ATLauncher.exe in a downloads stash must lose to the actual
// portable install, whose folder carries the launcher's data markers —
// regardless of walk order. Without a marked hit anywhere, the bare copy is
// still returned as the fallback (a fresh portable unzip has no data yet).
func TestScanForExesPrefersMarkedDir(t *testing.T) {
	root := t.TempDir()
	stash := filepath.Join(root, "AAA_INSTALLERS") // sorts (and walks) first
	install := filepath.Join(root, "ZZZ_LAUNCHERS", "ATLauncher")
	for _, dir := range []string{stash, filepath.Join(install, "instances")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	stashExe := filepath.Join(stash, "ATLauncher.exe")
	installExe := filepath.Join(install, "ATLauncher.exe")
	for _, exe := range []string{stashExe, installExe} {
		if err := os.WriteFile(exe, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	wanted := map[string]string{"atlauncher.exe": "atlauncher"}
	got := scanForExes(context.Background(), []string{root}, wanted, 6, 1)
	if got["atlauncher"] != installExe {
		t.Fatalf("scan picked %q, want marked install %q", got["atlauncher"], installExe)
	}

	// Fallback: with the marked install gone, the bare copy must still win.
	if err := os.RemoveAll(filepath.Join(root, "ZZZ_LAUNCHERS")); err != nil {
		t.Fatal(err)
	}
	got = scanForExes(context.Background(), []string{root}, wanted, 6, 1)
	if got["atlauncher"] != stashExe {
		t.Fatalf("scan fallback picked %q, want bare copy %q", got["atlauncher"], stashExe)
	}
}

func TestTrustCachedExeCandidate(t *testing.T) {
	bare := t.TempDir()
	marked := t.TempDir()
	if err := os.MkdirAll(filepath.Join(marked, "instances"), 0o755); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		id   string
		cand Candidate
		want bool
	}{
		{"scan hit, bare dir", "atlauncher", Candidate{Evidence: EvScan, Path: filepath.Join(bare, "ATLauncher.exe")}, false},
		{"scan hit, marked dir", "atlauncher", Candidate{Evidence: EvScan, Path: filepath.Join(marked, "ATLauncher.exe")}, true},
		{"shortcut hit, bare dir", "atlauncher", Candidate{Evidence: EvStartMenuLnk, Path: filepath.Join(bare, "ATLauncher.exe")}, true},
		{"user pick, bare dir", "atlauncher", Candidate{Evidence: EvScan, UserPicked: true, Path: filepath.Join(bare, "ATLauncher.exe")}, true},
		{"launcher without markers", "ultimmc", Candidate{Evidence: EvScan, Path: filepath.Join(bare, "UltimMC.exe")}, true},
	}
	for _, tc := range cases {
		if got := trustCachedExeCandidate(tc.id, &tc.cand); got != tc.want {
			t.Errorf("%s: trustCachedExeCandidate = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestDiscoverLauncherDirsDisplacesStrayExeCopy replays the machine test 2
// cache state end-to-end: the cache validates (the stray exe still exists)
// but its folder carries no data markers, so discovery must re-run and land
// on the real portable install under a scan root.
func TestDiscoverLauncherDirsDisplacesStrayExeCopy(t *testing.T) {
	appData := isolateDiscoveryEnv(t)
	base := filepath.Dir(appData)
	// Keep the scan away from the real Program Files on this machine.
	t.Setenv("ProgramFiles", filepath.Join(base, "PF"))
	t.Setenv("ProgramFiles(x86)", filepath.Join(base, "PF86"))

	// Stray leftover exe copy outside every scan root, cached as a scan hit.
	stray := t.TempDir()
	strayExe := filepath.Join(stray, "ATLauncher.exe")
	if err := os.WriteFile(strayExe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cache, _ := LoadCache()
	UpsertCandidate(cache, &Candidate{
		Launcher: LauncherID("atlauncher"), Path: strayExe, Kind: "exe",
		Evidence: EvScan, Confidence: "low",
		LastUsed: time.Now(), LastOK: time.Now(), HashHint: PathHint(strayExe),
	})
	if err := SaveCache(cache); err != nil {
		t.Fatal(err)
	}

	// The real portable install, marked by its instances\ folder, sits under
	// a common scan root (Downloads).
	install := filepath.Join(base, "Downloads", "LAUNCHERS", "ATLauncher")
	if err := os.MkdirAll(filepath.Join(install, "instances"), 0o755); err != nil {
		t.Fatal(err)
	}
	installExe := filepath.Join(install, "ATLauncher.exe")
	if err := os.WriteFile(installExe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := discoverLauncherDirs("atlauncher")
	if len(got) != 1 || got[0] != install {
		t.Fatalf("discoverLauncherDirs = %v, want [%s]", got, install)
	}
}
