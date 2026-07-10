package kumi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ── Launcher discovery primitives ────────────────
// Shared building blocks for exe-based launcher detection. The detection
// pipeline itself lives in service.go (detectByExecutable) and detect.go
// (discoverLauncherDirs): validated cache entries first, then Start Menu /
// taskbar / Desktop shortcut resolution, then the bounded filesystem scan
// below as the throttled last resort.

// ── Targeted concurrent scan ─────────────────────
//
// Future reference: instant-search fast-path (heavy update). This bounded
// scan is the slow path (cold start ~seconds). It is already rare — the
// cache + Start Menu shortcut resolution catch most launchers first, and
// this is throttled to once per 12h — but for cold starts or launchers
// without a shortcut, two Windows-only options could make "find <exe> on
// any NTFS drive" near-instant:
//   1. voidtools Everything SDK — query its live MFT index over IPC
//      (Everything.dll / the IPC window). Millisecond lookups, but only
//      when Everything is installed and its service is running, so it can
//      only be an optional accelerator, never a dependency.
//   2. Read the NTFS USN journal / $MFT directly (what Everything does).
//      No third-party app, but needs elevation and is NTFS-only.
// Either would sit in front of this scan as the fallback; neither helps
// Linux/macOS. Only worth building if cold-scan latency becomes a real
// complaint — the shortcut+cache pipeline already covers the common case.

// scanForExes searches the given roots for multiple executables in a single
// traversal. wanted maps lowercase exe filenames to a caller-defined key;
// the result maps each key to the first matching path found. The scan is
// depth-limited, skips noise directories and reparse points (ReadDir never
// descends into junctions/symlinks), and cancels as soon as every key is
// resolved. Concurrency is bounded: when all worker slots are busy the
// walk continues inline instead of blocking, so it can never deadlock.
func scanForExes(ctx context.Context, roots []string, wanted map[string]string, maxDepth, workers int) map[string]string {
	results := map[string]string{}
	if len(wanted) == 0 {
		return results
	}

	keys := map[string]struct{}{}
	for _, key := range wanted {
		keys[key] = struct{}{}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu        sync.Mutex
		remaining = len(keys)
		sem       = make(chan struct{}, workers)
		wg        sync.WaitGroup
	)

	ignoredDir := func(name string) bool {
		n := strings.ToLower(name)
		return n == "windows" ||
			n == "$recycle.bin" ||
			n == "system volume information" ||
			n == "node_modules" ||
			n == ".git" ||
			strings.HasPrefix(n, "msocache")
	}

	record := func(key, path string) {
		mu.Lock()
		defer mu.Unlock()
		if _, exists := results[key]; exists {
			return
		}
		results[key] = path
		remaining--
		if remaining == 0 {
			cancel()
		}
	}

	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		select {
		case <-ctx.Done():
			return
		default:
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			select {
			case <-ctx.Done():
				return
			default:
			}
			name := e.Name()
			full := filepath.Join(dir, name)
			if e.Type().IsRegular() {
				if key, ok := wanted[strings.ToLower(name)]; ok {
					record(key, full)
				}
				continue
			}
			if !e.IsDir() || depth >= maxDepth || ignoredDir(name) {
				continue
			}
			select {
			case sem <- struct{}{}:
				wg.Add(1)
				go func(d string, dep int) {
					defer wg.Done()
					defer func() { <-sem }()
					walk(d, dep)
				}(full, depth+1)
			default:
				// All worker slots busy — recurse inline rather than block.
				walk(full, depth+1)
			}
		}
	}

	for _, r := range roots {
		r = filepath.Clean(r)
		if r == "" || !dirExists(r) {
			continue
		}
		walk(r, 0)
	}
	wg.Wait()
	return results
}

// ── Shortcut resolution pre-pass ─────────────────

// shortcutRoots returns the small, high-signal folders that hold .lnk files
// (taskbar pins, Start Menu, Desktop). Searching these first finds launchers
// installed on any drive at negligible cost.
func shortcutRoots() []string {
	user := os.Getenv("USERPROFILE")
	appData := os.Getenv("APPDATA")
	programData := os.Getenv("ProgramData")

	candidates := []string{
		filepath.Join(user, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch"),
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(user, "Desktop"),
	}
	for _, root := range oneDriveRoots(user) {
		candidates = append(candidates, filepath.Join(root, "Desktop"))
	}

	var roots []string
	for _, c := range candidates {
		if dirExists(c) {
			roots = append(roots, c)
		}
	}
	return roots
}

// resolveShortcutTargets walks the shortcut roots, parses the .lnk files in
// parallel (file reads can be slow under AV scanning), and returns the first
// existing target matching each wanted exe name. Stops as soon as every key
// is resolved. wanted maps lowercase exe filenames to a caller-defined key.
func resolveShortcutTargets(wanted map[string]string) map[string]string {
	results := map[string]string{}
	if len(wanted) == 0 {
		return results
	}

	// Collect shortcut paths first — directory enumeration is cheap.
	var lnks []string
	for _, root := range shortcutRoots() {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".lnk") {
				lnks = append(lnks, path)
			}
			return nil
		})
	}
	if len(lnks) == 0 {
		return results
	}

	keys := map[string]struct{}{}
	for _, key := range wanted {
		keys[key] = struct{}{}
	}

	var (
		mu        sync.Mutex
		remaining = len(keys)
		done      = make(chan struct{})
		closeOnce sync.Once
		jobs      = make(chan string)
		wg        sync.WaitGroup
	)

	const workers = 8
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for path := range jobs {
				target, _, err := parseShortcut(path)
				if err != nil || target == "" {
					continue
				}
				key, ok := wanted[strings.ToLower(filepath.Base(target))]
				if !ok || !fileExistsR(target) {
					continue
				}
				mu.Lock()
				if _, exists := results[key]; !exists {
					results[key] = target
					remaining--
					if remaining == 0 {
						closeOnce.Do(func() { close(done) })
					}
				}
				mu.Unlock()
			}
		}()
	}

feed:
	for _, path := range lnks {
		select {
		case jobs <- path:
		case <-done:
			break feed
		}
	}
	close(jobs)
	wg.Wait()
	return results
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
	add(os.Getenv("ProgramData"))
	add(local)
	add(roam)
	if local != "" {
		add(filepath.Join(filepath.Dir(local), "LocalLow"))
	}
	add(filepath.Join(home, "Desktop"))
	add(filepath.Join(home, "Downloads"))
	add(filepath.Join(home, "Documents"))
	add(filepath.Join(home, "Games"))
	for _, root := range oneDriveRoots(home) {
		add(filepath.Join(root, "Desktop"))
		add(filepath.Join(root, "Downloads"))
		add(filepath.Join(root, "Documents"))
		add(filepath.Join(root, "Games"))
	}

	// Program folders on secondary drives — installs are not constrained to
	// the system drive. Existence is checked by the scan itself.
	for _, drive := range fixedDriveRoots() {
		add(filepath.Join(drive, "Program Files"))
		add(filepath.Join(drive, "Program Files (x86)"))
		add(filepath.Join(drive, "Programs"))
		add(filepath.Join(drive, "Games"))
	}

	return uniqueStrings(roots)
}

func oneDriveRoots(home string) []string {
	roots := []string{
		os.Getenv("OneDrive"),
		os.Getenv("OneDriveConsumer"),
		os.Getenv("OneDriveCommercial"),
	}
	if home != "" {
		roots = append(roots, filepath.Join(home, "OneDrive"))
	}
	return uniqueStrings(roots)
}

// fixedDriveRoots lists existing drive roots (D:\ .. Z:\, plus C:\).
func fixedDriveRoots() []string {
	var drives []string
	for letter := 'C'; letter <= 'Z'; letter++ {
		root := string(letter) + ":\\"
		if dirExists(root) {
			drives = append(drives, root)
		}
	}
	return drives
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
