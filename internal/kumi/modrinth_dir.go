package kumi

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ── Modrinth data-directory resolution ───────────
//
// The Modrinth app (Theseus) keeps its state in %APPDATA%\ModrinthApp (older
// builds: %APPDATA%\com.modrinth.theseus). Profiles normally live in
// <app dir>\profiles, BUT the user can relocate the whole data directory from
// the app's settings; that choice is stored in app.db (SQLite):
//
//	SELECT custom_dir FROM settings;
//
// When custom_dir is set, profiles live in <custom_dir>\profiles instead, so
// installing into the default location would put the pack somewhere the
// launcher never looks. app.db itself always stays in the roaming app dir.

// modrinthAppDirs returns the Modrinth app data dirs to probe, newest first.
func modrinthAppDirs() []string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		if home, err := os.UserHomeDir(); err == nil {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
	}
	return cleanCandidates(
		filepath.Join(appData, "ModrinthApp"),
		filepath.Join(appData, "com.modrinth.theseus"),
	)
}

// modrinthDBPath returns the path to the Modrinth app.db, or "" when no
// Modrinth data dir holds one.
func modrinthDBPath() string {
	for _, dir := range modrinthAppDirs() {
		db := filepath.Join(dir, "app.db")
		if fileExistsR(db) {
			return db
		}
	}
	return ""
}

// modrinthCustomDir reads settings.custom_dir from the given app.db. Returns
// "" (no error) when the setting is NULL/empty; errors are reserved for a
// missing/unreadable database so callers can fall back to the default dir.
func modrinthCustomDir(dbPath string) (string, error) {
	if dbPath == "" || !fileExistsR(dbPath) {
		return "", errors.New("modrinth app.db not found")
	}
	// mode=ro: never mutate the launcher's DB, and reading stays safe while
	// the Modrinth app itself is running.
	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var customDir sql.NullString
	if err := db.QueryRow("SELECT custom_dir FROM settings").Scan(&customDir); err != nil {
		return "", err
	}
	return strings.TrimSpace(customDir.String), nil
}

// modrinthDataRoot resolves the directory Modrinth actually keeps profiles
// under: the custom_dir from app.db when set, otherwise the app dir itself.
func modrinthDataRoot() (string, error) {
	if dbPath := modrinthDBPath(); dbPath != "" {
		if custom, err := modrinthCustomDir(dbPath); err == nil && custom != "" {
			// Use it even if the folder is currently missing — it's still
			// where the launcher will look (and recreate) on next start.
			return filepath.Clean(custom), nil
		}
		return filepath.Dir(dbPath), nil
	}
	// No app.db (Modrinth not installed / never launched): fall back to the
	// first conventional app dir so installs land where Modrinth will look.
	dirs := modrinthAppDirs()
	if len(dirs) == 0 {
		return "", errors.New("unable to resolve Modrinth data directory")
	}
	for _, dir := range dirs {
		if dirExists(dir) {
			return dir, nil
		}
	}
	return dirs[0], nil
}

// modrinthProfilesRoot returns the directory Modrinth stores profiles in.
func modrinthProfilesRoot() (string, error) {
	root, err := modrinthDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profiles"), nil
}

// modrinthProfilesInfo builds the human-readable note shown behind the info
// icon on the Modrinth row: where profiles will be installed and where that
// answer came from.
func modrinthProfilesInfo() string {
	root, err := modrinthProfilesRoot()
	if err != nil {
		return ""
	}
	dbPath := modrinthDBPath()
	if dbPath == "" {
		return "Profiles folder: " + root + "\n(Modrinth default — app.db not found)"
	}
	if custom, cerr := modrinthCustomDir(dbPath); cerr == nil && custom != "" {
		return "Profiles folder: " + root + "\n(custom location set in " + dbPath + ")"
	}
	return "Profiles folder: " + root + "\n(Modrinth default, per " + dbPath + ")"
}
