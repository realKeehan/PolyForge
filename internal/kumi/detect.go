package kumi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── Path helpers ─────────────────────────────────

// localLowDir returns %USERPROFILE%\AppData\LocalLow, derived from LOCALAPPDATA
// so it tracks a redirected profile. Empty when LOCALAPPDATA is unset.
func localLowDir() string {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(local), "LocalLow")
}

// cleanCandidates drops empty entries, normalizes each path, and removes
// duplicates while preserving order. It keeps candidate lists honest: no blank
// explicit paths, no `C:\A` vs `C:\A\` twins probed twice.
func cleanCandidates(paths ...string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		clean := filepath.Clean(p)
		key := strings.ToLower(clean)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, clean)
	}
	return out
}

// ── Portable-install discovery ───────────────────
//
// The fixed candidate lists below cover conventional %APPDATA%/%LOCALAPPDATA%
// installs. Portable launchers (MultiMC, Prism forks, PolyMC, …) can live
// anywhere — Desktop, Downloads, D:\Games\…, a random folder on any drive.
// Discovery finds those, in cost order:
//
//	1. Launcher cache — locations found on previous runs, re-validated so a
//	   moved/uninstalled launcher falls through to a fresh search.
//	2. Shortcut resolution — Start Menu / taskbar / Desktop .lnk targets.
//	3. Bounded concurrent scan of the common install roots.
//
// Fresh finds are persisted to the cache, so the scan happens once per
// machine, not once per boot. The candidate functions themselves stay cheap
// (no I/O beyond env lookups) because Options() probes every launcher at
// startup; discovery is applied at install time in launchers.go via
// withExeDiscovery / withDirDiscovery.

// scanTimeout bounds the cold filesystem scan. The scan cancels early as soon
// as every wanted executable is found, so this ceiling only applies when a
// launcher genuinely is not installed.
const scanTimeout = 15 * time.Second

// discoverLauncherDirs locates the launcher's executable (names from
// launcherExeNames) and returns the directories holding it. Cache hits skip
// all filesystem searching; cache misses fall back to shortcuts, then a scan,
// and persist what they find.
func discoverLauncherDirs(id string) []string {
	exeNames := launcherExeNames[id]
	if len(exeNames) == 0 {
		return nil
	}

	cache, _ := LoadCache()

	// 1) Previously found and still valid (exe still exists at that path).
	if cand := BestValidCachedCandidate(cache, LauncherID(id), ValidateExeByName(exeNames...)); cand != nil {
		cand.LastUsed = time.Now()
		_ = SaveCache(cache)
		return []string{filepath.Dir(cand.Path)}
	}

	// Cache miss or the launcher moved — search again.
	wanted := make(map[string]string, len(exeNames))
	for _, name := range exeNames {
		wanted[strings.ToLower(name)] = id
	}

	// 2) Shortcut resolution — negligible cost, finds installs on any drive.
	exePath := resolveShortcutTargets(wanted)[id]
	evidence := EvStartMenuLnk
	confidence := "high"

	// 3) Bounded scan of the common install roots.
	if exePath == "" {
		ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
		defer cancel()
		exePath = scanForExes(ctx, commonScanRoots(), wanted, 6, 8)[id]
		evidence = EvScan
		confidence = "low"
	}
	if exePath == "" {
		return nil
	}

	UpsertCandidate(cache, &Candidate{
		Launcher:   LauncherID(id),
		Path:       exePath,
		Kind:       "exe",
		Evidence:   evidence,
		Confidence: confidence,
		LastUsed:   time.Now(),
		LastOK:     time.Now(),
		HashHint:   PathHint(exePath),
	})
	_ = SaveCache(cache)

	return []string{filepath.Dir(exePath)}
}

// withExeDiscovery is for launchers whose installer probes for an executable
// (firstExisting). If a fixed candidate already holds one of the launcher's
// executables it returns immediately — no cache read, no scan. Otherwise it
// appends the discovered install directories.
func withExeDiscovery(id string, fixed []string) []string {
	for _, name := range launcherExeNames[id] {
		if firstExisting(fixed, name) != "" {
			return fixed
		}
	}
	discovered := discoverLauncherDirs(id)
	if len(discovered) == 0 {
		return fixed
	}
	return cleanCandidates(append(append([]string{}, fixed...), discovered...)...)
}

// withDirDiscovery is for launchers whose installer probes for a data
// directory (firstExistingDirectory). If any fixed candidate already exists it
// returns immediately. Otherwise it appends the discovered install
// directories — only sensible for launchers that keep instances next to their
// executable when portable (the MultiMC/Prism family and kin).
func withDirDiscovery(id string, fixed []string) []string {
	if firstExistingDirectory(fixed) != "" {
		return fixed
	}
	discovered := discoverLauncherDirs(id)
	if len(discovered) == 0 {
		return fixed
	}
	return cleanCandidates(append(append([]string{}, fixed...), discovered...)...)
}

// ── Launcher candidate sets ──────────────────────
//
// These must stay cheap (env lookups only): Options() calls them for every
// launcher on startup via detectLauncherPath.

func multiMCCandidates(explicit string) []string {
	home := os.Getenv("USERPROFILE")
	local := os.Getenv("LOCALAPPDATA")
	return cleanCandidates(
		explicit,
		filepath.Join(home, "MultiMC"),
		filepath.Join(home, "Desktop", "MultiMC"),
		filepath.Join(home, "Downloads", "MultiMC"),
		filepath.Join(local, "MultiMC"),
		`C:\MultiMC`,
		`C:\Games\MultiMC`,
		`C:\Programs\MultiMC`,
		`D:\MultiMC`,
		`D:\Games\MultiMC`,
		`D:\Programs\MultiMC`,
	)
}

func curseForgeTarget() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "curseforge", "minecraft", "Instances", "TurtelSMP5"), nil
}

// modrinthTarget resolves where the Turtel profile should live: inside the
// profiles root, which honours the custom_dir the user may have set in the
// Modrinth app (stored in app.db — see modrinth_dir.go).
func modrinthTarget() (string, error) {
	root, err := modrinthProfilesRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "TurtelSMP5"), nil
}

func gdLauncherCandidates(explicit string) []string {
	appData := os.Getenv("APPDATA")
	return cleanCandidates(
		explicit,
		// Current GDLauncher (Carbon) data dir; the exe lives separately
		// under %LOCALAPPDATA%\Programs\@gddesktop (MachineTest_01).
		filepath.Join(appData, "gdlauncher_carbon"),
		filepath.Join(appData, "GDLauncher Carbon"),
		filepath.Join(appData, "gdlauncher_next"),
		filepath.Join(appData, "gdlauncher"),
	)
}

func atLauncherCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "ATLauncher"),
		`C:\ATLauncher`,
	)
}

func prismLauncherCandidates(explicit string) []string {
	appData := os.Getenv("APPDATA")
	return cleanCandidates(
		explicit,
		filepath.Join(appData, "PrismLauncher"),
		filepath.Join(appData, "PrismLauncher", "minecraft"),
	)
}

func bakaXLCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "BakaXL"),
		`C:\BakaXL`,
	)
}

// Dawn is the rebranded Feather client (acquired by InPVP, mid-2026). The
// exe installs under %LOCALAPPDATA%\Dawn but profiles live in
// %APPDATA%\.dawn (MachineTest_01). Legacy Feather dirs are kept for
// installs that predate the rebrand.
func dawnCandidates(explicit string) []string {
	appData := os.Getenv("APPDATA")
	return cleanCandidates(
		explicit,
		filepath.Join(appData, ".dawn"),
		filepath.Join(appData, "feather"),
		filepath.Join(appData, "FeatherClient"),
		filepath.Join(localLowDir(), "Feather"),
	)
}

func technicCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), ".technic"),
		`C:\.technic`,
	)
}

func polyMCCandidates(explicit string) []string {
	appData := os.Getenv("APPDATA")
	return cleanCandidates(
		explicit,
		filepath.Join(appData, "PolyMC"),
		filepath.Join(appData, "polymc"),
	)
}

func skLauncherCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "SKLauncher"),
		filepath.Join(os.Getenv("APPDATA"), ".sklauncher"),
	)
}

func freesmCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "FreesmLauncher"),
		filepath.Join(os.Getenv("APPDATA"), "freesmlauncher"),
	)
}

// PineconeMC is the rebranded ElyPrism; probe the new folders first but keep
// the legacy ElyPrism locations for existing installs.
func elyPrismCandidates(explicit string) []string {
	appData := os.Getenv("APPDATA")
	return cleanCandidates(
		explicit,
		filepath.Join(appData, "PineconeMC"),
		filepath.Join(appData, "PineconeMCLauncher"),
		filepath.Join(appData, "ElyPrism"),
		filepath.Join(appData, "ElyPrismLauncher"),
	)
}

func shatteredPrismCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "ShatteredPrism"),
	)
}

// QWERTZ keeps its exe *inside* the data dir (%APPDATA%\QWERTZ-Launcher)
// and stores instances under profiles\<name> next to a profiles.json
// registry (MachineTest_01).
func qwertzCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "QWERTZ-Launcher"),
		filepath.Join(os.Getenv("APPDATA"), "QWERTZ"),
		filepath.Join(os.Getenv("APPDATA"), "qwertz"),
	)
}

func fjordCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "FjordLauncher"),
	)
}

func hmclCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), ".hmcl"),
		filepath.Join(os.Getenv("USERPROFILE"), ".hmcl"),
	)
}

func ultimMCCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("APPDATA"), "UltimMC"),
	)
}

// Polymerium is inverted relative to most launchers: settings.json sits in
// %APPDATA%\Polymerium but instances live under %LOCALAPPDATA%\Trident
// (MachineTest_01), so Trident is probed first for install targeting.
func polymeriumCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Trident"),
		filepath.Join(os.Getenv("APPDATA"), "Polymerium"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Polymerium"),
	)
}

// XMCL splits config from game data: %APPDATA%\xmcl holds settings plus the
// instances.json registry, while the instance folders themselves live under
// %USERPROFILE%\.minecraftx\instances (MachineTest_01). .minecraftx comes
// first so installs land where the instances actually are.
func xmclCandidates(explicit string) []string {
	return cleanCandidates(
		explicit,
		filepath.Join(os.Getenv("USERPROFILE"), ".minecraftx"),
		filepath.Join(os.Getenv("APPDATA"), "xmcl"),
		filepath.Join(os.Getenv("APPDATA"), "X Minecraft Launcher"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "xmcl"),
	)
}
