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

// RunFirstRunSetup performs the one-time setup if it has not run yet.
// Safe to call on every startup; it is a no-op after the first success.
// Returns a human-readable note about what happened (for logs / UI), or "".
func RunFirstRunSetup() string {
	if !NeedsFirstRunSetup() {
		return ""
	}
	note := setupFileAssociation() // platform-specific (see fileassoc_*.go)
	markFirstRunDone()
	return note
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
