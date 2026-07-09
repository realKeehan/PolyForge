package kumi

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// buildTestPack writes a plain-zip .polypack with the given overrides-relative
// files. When withHashes is true the manifest carries overrides.files, so the
// installer performs integrity verification.
func buildTestPack(t *testing.T, files map[string]string, withHashes bool) string {
	t.Helper()
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	var packFiles []PackFile
	for rel, content := range files {
		w, err := zw.Create("overrides/" + rel)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256([]byte(content))
		packFiles = append(packFiles, PackFile{Path: rel, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(content))})
	}
	manifest := PackManifest{
		SchemaVersion: 1,
		ID:            "test-pack",
		Name:          "Test Pack",
		Version:       "1.0.0",
		Overrides:     PackOverrides{Folders: []string{"mods"}, FileCount: len(files)},
	}
	if withHashes {
		manifest.Overrides.Files = packFiles
	}
	mData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	mw, err := zw.Create("pack-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	mw.Write(mData)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	packPath := filepath.Join(t.TempDir(), "test.polypack")
	if err := os.WriteFile(packPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return packPath
}

func TestInstallLocalPackVerifiesIntegrity(t *testing.T) {
	files := map[string]string{"mods/a.jar": "pretend jar bytes", "config/x.toml": "key=value"}
	pack := buildTestPack(t, files, true)
	target := t.TempDir()

	n, m, report, err := installLocalPack(pack, target)
	if err != nil {
		t.Fatalf("installLocalPack: %v", err)
	}
	if n != 2 {
		t.Errorf("extracted %d files, want 2", n)
	}
	if m.ID != "test-pack" {
		t.Errorf("manifest id = %q", m.ID)
	}
	if !report.OK() || report.Checked != 2 || report.Total != 2 {
		t.Fatalf("report = %+v, want OK with 2/2 checked", report)
	}
}

func TestVerifyInstalledPackDetectsCorruptionAndMissing(t *testing.T) {
	files := map[string]string{"mods/a.jar": "genuine jar", "config/x.toml": "key=value"}
	pack := buildTestPack(t, files, true)
	target := t.TempDir()
	if _, _, _, err := installLocalPack(pack, target); err != nil {
		t.Fatalf("installLocalPack: %v", err)
	}

	// Tamper with one file, delete another.
	if err := os.WriteFile(filepath.Join(target, "mods", "a.jar"), []byte("TAMPERED"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(target, "config", "x.toml")); err != nil {
		t.Fatal(err)
	}

	_, report, err := VerifyInstalledPack(target)
	if err != nil {
		t.Fatalf("VerifyInstalledPack: %v", err)
	}
	if report.OK() {
		t.Fatal("expected integrity issues, got none")
	}
	reasons := map[string]string{}
	for _, issue := range report.Issues {
		reasons[issue.Path] = issue.Reason
	}
	if reasons["mods/a.jar"] != "hash mismatch" {
		t.Errorf("mods/a.jar reason = %q, want hash mismatch", reasons["mods/a.jar"])
	}
	if reasons["config/x.toml"] != "missing" {
		t.Errorf("config/x.toml reason = %q, want missing", reasons["config/x.toml"])
	}
}

func TestInstallLocalPackSkipsVerificationWhenNoHashes(t *testing.T) {
	// A pack built before per-file hashes existed: no overrides.files.
	pack := buildTestPack(t, map[string]string{"mods/a.jar": "jar"}, false)
	target := t.TempDir()
	_, _, report, err := installLocalPack(pack, target)
	if err != nil {
		t.Fatalf("installLocalPack: %v", err)
	}
	if report.Total != 0 || !report.OK() {
		t.Fatalf("report = %+v, want empty/OK (verification skipped)", report)
	}
}
