package kumi

import (
	"os"
	"path/filepath"
)

// ══════════════════════════════════════════════════
// First-run setup
//
// On first launch PolyForge performs one-time preliminary setup — chiefly
// registering the .polypack file type (with its own icon) so packs
// double-click into the app. A stamp file records that setup ran so it only
// happens once. No admin rights or drivers are required; the Windows
// implementation writes only to the per-user registry hive (see
// fileassoc_windows.go).
// ══════════════════════════════════════════════════

// PackExtension is PolyForge's pack file extension. The file itself is the
// obfuscated container implemented in slime.go — .polypack is just the
// user-facing name/extension for it.
const PackExtension = ".polypack"

func firstRunStampPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "first-run-done"), nil
}

// packIconPath is where the generated .polypack file-type icon is written.
func packIconPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "polypack.ico"), nil
}

// NeedsFirstRunSetup reports whether first-run setup has not yet completed.
func NeedsFirstRunSetup() bool {
	path, err := firstRunStampPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return os.IsNotExist(err)
}

func markFirstRunDone() {
	path, err := firstRunStampPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte("done"), 0o644)
}

// RunFirstRunSetup registers the .polypack association and returns a
// human-readable note on the first run (for logs / UI), or "" afterwards.
//
// The registration itself runs on EVERY startup, not just the first: it is
// idempotent (it only rewrites per-user registry values and regenerates the
// icon), and re-running it self-heals a stale association — e.g. after the exe
// is renamed/moved (release builds are versioned, so the path changes) or when
// an earlier run registered against a broken build. Only the first run is
// announced so we don't spam the log every launch.
func RunFirstRunSetup() string {
	first := NeedsFirstRunSetup()
	note := setupFileAssociation() // platform-specific (see fileassoc_*.go)
	if first {
		markFirstRunDone()
		return note
	}
	return ""
}

// LaunchedPackPath returns a .polypack/.zip path passed on the command line
// (e.g. from double-clicking a pack), or "" if none. Lets the app jump
// straight into installing a pack the user opened.
func LaunchedPackPath() string {
	for _, arg := range os.Args[1:] {
		ext := filepath.Ext(arg)
		if ext == PackExtension || ext == ".zip" {
			if fileExistsR(arg) {
				return arg
			}
		}
	}
	return ""
}
