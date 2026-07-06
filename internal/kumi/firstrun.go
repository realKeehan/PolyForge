package kumi

import (
	"os"
	"path/filepath"
)

// ══════════════════════════════════════════════════
// First-run setup
//
// On first launch PolyForge performs one-time preliminary setup — chiefly
// registering the .slime file type so packs double-click into the app.
// A stamp file records that setup ran so it only happens once. No admin
// rights or drivers are required; the Windows implementation writes only to
// the per-user registry hive (see fileassoc_windows.go).
// ══════════════════════════════════════════════════

// SlimeExtension is PolyForge's pack file extension.
const SlimeExtension = ".slime"

func firstRunStampPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "first-run-done"), nil
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

// LaunchedPackPath returns a .slime/.polypack path passed on the command
// line (e.g. from double-clicking a pack), or "" if none. Lets the app jump
// straight into installing a pack the user opened.
func LaunchedPackPath() string {
	for _, arg := range os.Args[1:] {
		lower := filepath.Ext(arg)
		if lower == SlimeExtension || filepath.Ext(filepath.Base(arg)) == ".zip" {
			if fileExistsR(arg) {
				return arg
			}
		}
	}
	return ""
}
