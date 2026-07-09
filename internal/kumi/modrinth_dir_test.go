package kumi

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeModrinthDB creates <appDir>\app.db with a settings table whose
// custom_dir holds the given value ("" → NULL, mirroring a default install).
func makeModrinthDB(t *testing.T, appDir, customDir string) {
	t.Helper()
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(appDir, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE settings (id INTEGER PRIMARY KEY, custom_dir TEXT)`); err != nil {
		t.Fatal(err)
	}
	val := sql.NullString{String: customDir, Valid: customDir != ""}
	if _, err := db.Exec(`INSERT INTO settings (id, custom_dir) VALUES (0, ?)`, val); err != nil {
		t.Fatal(err)
	}
}

func TestModrinthProfilesRootHonoursCustomDir(t *testing.T) {
	appData := t.TempDir()
	t.Setenv("APPDATA", appData)

	custom := filepath.Join(t.TempDir(), "MovedModrinth")
	makeModrinthDB(t, filepath.Join(appData, "ModrinthApp"), custom)

	root, err := modrinthProfilesRoot()
	if err != nil {
		t.Fatalf("modrinthProfilesRoot: %v", err)
	}
	want := filepath.Join(custom, "profiles")
	if root != want {
		t.Errorf("root = %q, want %q", root, want)
	}

	info := modrinthProfilesInfo()
	if !strings.Contains(info, want) || !strings.Contains(info, "custom location") {
		t.Errorf("info = %q, want it to mention %q as a custom location", info, want)
	}
}

func TestModrinthProfilesRootDefaultsBesideAppDB(t *testing.T) {
	appData := t.TempDir()
	t.Setenv("APPDATA", appData)

	appDir := filepath.Join(appData, "ModrinthApp")
	makeModrinthDB(t, appDir, "") // custom_dir NULL

	root, err := modrinthProfilesRoot()
	if err != nil {
		t.Fatalf("modrinthProfilesRoot: %v", err)
	}
	want := filepath.Join(appDir, "profiles")
	if root != want {
		t.Errorf("root = %q, want %q", root, want)
	}
}

// TestModrinthRealAppDB opportunistically resolves the profiles root on the
// machine running the tests, ensuring the reader copes with a real app.db
// (real schema, WAL mode, possibly open in the Modrinth app).
func TestModrinthRealAppDB(t *testing.T) {
	dbPath := modrinthDBPath()
	if dbPath == "" {
		t.Skip("no Modrinth app.db on this machine")
	}
	root, err := modrinthProfilesRoot()
	if err != nil {
		t.Fatalf("modrinthProfilesRoot: %v", err)
	}
	t.Logf("app.db: %s", dbPath)
	t.Logf("profiles root: %s", root)
	t.Logf("info: %q", modrinthProfilesInfo())
	if !dirExists(root) {
		t.Logf("note: resolved profiles root does not exist yet")
	}
}

func TestModrinthProfilesRootWithoutAppDB(t *testing.T) {
	appData := t.TempDir()
	t.Setenv("APPDATA", appData)

	// Legacy dir exists, no app.db anywhere: fall back to the existing dir.
	legacy := filepath.Join(appData, "com.modrinth.theseus")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := modrinthProfilesRoot()
	if err != nil {
		t.Fatalf("modrinthProfilesRoot: %v", err)
	}
	want := filepath.Join(legacy, "profiles")
	if root != want {
		t.Errorf("root = %q, want %q", root, want)
	}
}
