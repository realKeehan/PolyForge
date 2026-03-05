package kumi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── Resolver framework ───────────────────────────
// Multi-strategy launcher resolver. Discovers installed launchers through:
//   1. Cache - previously validated paths
//   2. Known paths - common install locations
//   3. Registry - Windows uninstall keys
//   4. Shell AppsFolder - UWP/Store apps via PowerShell
//   5. Running processes - detect launchers currently open
//   6. Start Menu shortcuts - resolve .lnk targets
//   7. Targeted scan - depth-limited concurrent filesystem scan

// ResolveResult holds the output of a launcher resolution attempt.
type ResolveResult struct {
	ExePath   string
	DataDir   string
	Evidence  Evidence
	Notes     []string
	FromCache bool
}

// LauncherSpec defines the parameters for resolving a specific launcher.
type LauncherSpec struct {
	ID         LauncherID
	ExeNames   []string // e.g. ["MinecraftLauncher.exe"]
	DataDirFn  func() string
	KnownPaths func() []string
	Validate   func(string) error
	// Registry display name substring for uninstall key search (Windows)
	RegistryDisplayName string
	// Process name without extension for running process detection
	ProcessName string
}

// ResolveLauncher runs through all detection strategies for the given spec.
func ResolveLauncher(ctx context.Context, cache *LauncherCache, spec *LauncherSpec) (*ResolveResult, error) {
	var notes []string

	// 0) Data dir check
	dataDir := ""
	if spec.DataDirFn != nil {
		dataDir = spec.DataDirFn()
		if dirExists(dataDir) {
			notes = append(notes, "Found data directory")
		} else {
			notes = append(notes, "Data directory not found (may be normal)")
		}
	}

	// 1) Cache first
	if cand := BestValidCachedCandidate(cache, spec.ID, spec.Validate); cand != nil {
		cand.LastUsed = time.Now()
		notes = append(notes, "Using cached path (validated)")
		return &ResolveResult{
			ExePath:   cand.Path,
			DataDir:   dataDir,
			Evidence:  EvCache,
			Notes:     notes,
			FromCache: true,
		}, nil
	}

	// 2) Known paths
	if spec.KnownPaths != nil {
		for _, p := range spec.KnownPaths() {
			if spec.Validate(p) == nil {
				notes = append(notes, "Found via known paths")
				UpsertCandidate(cache, &Candidate{
					Launcher:   spec.ID,
					Path:       p,
					Kind:       "exe",
					Evidence:   EvKnownPaths,
					Confidence: "high",
					LastUsed:   time.Now(),
					LastOK:     time.Now(),
					HashHint:   PathHint(p),
				})
				return &ResolveResult{ExePath: p, DataDir: dataDir, Evidence: EvKnownPaths, Notes: notes}, nil
			}
		}
	}

	// 3-7) Platform-specific detection strategies would go here.
	// These are scaffolded as extension points.
	// On Windows: registry, appsfolder, running processes, start menu, targeted scan.
	// On Linux/macOS: known paths, running processes, XDG paths, targeted scan.

	// 7) Targeted scan (last resort)
	if len(spec.ExeNames) > 0 {
		roots := commonScanRoots()
		for _, exeName := range spec.ExeNames {
			if p := scanForExe(ctx, roots, exeName, 5, 8); p != "" && spec.Validate(p) == nil {
				notes = append(notes, fmt.Sprintf("Found via targeted scan: %s", exeName))
				UpsertCandidate(cache, &Candidate{
					Launcher:   spec.ID,
					Path:       p,
					Kind:       "exe",
					Evidence:   EvScan,
					Confidence: "low",
					LastUsed:   time.Now(),
					LastOK:     time.Now(),
					HashHint:   PathHint(p),
				})
				return &ResolveResult{ExePath: p, DataDir: dataDir, Evidence: EvScan, Notes: notes}, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid executable found for %s via any detection strategy", spec.ID)
}

// SaveManualChoice caches a user-selected path with highest priority.
func SaveManualChoice(cache *LauncherCache, id LauncherID, exePath string, validate func(string) error) error {
	if err := validate(exePath); err != nil {
		return err
	}
	UpsertCandidate(cache, &Candidate{
		Launcher:   id,
		Path:       exePath,
		Kind:       "exe",
		Evidence:   EvManual,
		Confidence: "high",
		UserPicked: true,
		LastUsed:   time.Now(),
		LastOK:     time.Now(),
		HashHint:   PathHint(exePath),
	})
	return SaveCache(cache)
}

// ── Targeted concurrent scan ─────────────────────

func scanForExe(ctx context.Context, roots []string, exeName string, maxDepth int, workers int) string {
	type job struct {
		dir   string
		depth int
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan job, 256)
	found := make(chan string, 1)

	ignoredDir := func(name string) bool {
		n := strings.ToLower(name)
		return n == "windows" ||
			n == "$recycle.bin" ||
			n == "system volume information" ||
			n == "node_modules" ||
			n == ".git" ||
			strings.HasPrefix(n, "msocache")
	}

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for j := range jobs {
			select {
			case <-ctx.Done():
				return
			default:
			}
			entries, err := os.ReadDir(j.dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				select {
				case <-ctx.Done():
					return
				default:
				}
				name := e.Name()
				full := filepath.Join(j.dir, name)
				if e.Type().IsRegular() {
					if strings.EqualFold(name, exeName) {
						select {
						case found <- full:
							cancel()
							return
						default:
							cancel()
							return
						}
					}
					continue
				}
				if e.IsDir() {
					if ignoredDir(name) {
						continue
					}
					if j.depth < maxDepth {
						select {
						case jobs <- job{dir: full, depth: j.depth + 1}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	go func() {
		defer close(jobs)
		for _, r := range roots {
			r = filepath.Clean(r)
			if r == "" || !dirExists(r) {
				continue
			}
			select {
			case jobs <- job{dir: r, depth: 0}:
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case p := <-found:
		wg.Wait()
		return p
	case <-ctx.Done():
		wg.Wait()
		return ""
	}
}

func commonScanRoots() []string {
	home, _ := os.UserHomeDir()
	local := os.Getenv("LOCALAPPDATA")
	roam := os.Getenv("APPDATA")
	prog := os.Getenv("ProgramFiles")
	prog86 := os.Getenv("ProgramFiles(x86)")

	var roots []string
	add := func(p string) {
		if p != "" {
			roots = append(roots, p)
		}
	}
	add(prog)
	add(prog86)
	add(local)
	add(roam)
	add(filepath.Join(home, "Desktop"))
	add(filepath.Join(home, "Downloads"))
	add(filepath.Join(home, "Documents"))
	add(filepath.Join(home, "Games"))

	return uniqueStrings(roots)
}

// ── FS helpers ───────────────────────────────────

func fileExistsR(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		k := strings.ToLower(filepath.Clean(s))
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, s)
	}
	return out
}

// ── Convenience: validate exe by name ────────────

// ValidateExeByName returns a validator that checks path existence and
// base filename match (case-insensitive).
func ValidateExeByName(allowedNames ...string) func(string) error {
	return func(p string) error {
		if p == "" {
			return errors.New("empty path")
		}
		if !fileExistsR(p) {
			return fmt.Errorf("not found: %s", p)
		}
		base := strings.ToLower(filepath.Base(p))
		for _, n := range allowedNames {
			if base == strings.ToLower(n) {
				return nil
			}
		}
		return fmt.Errorf("unexpected filename: %s", base)
	}
}
