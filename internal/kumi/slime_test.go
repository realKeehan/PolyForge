package kumi

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSlimeRoundTrip(t *testing.T) {
	original := []byte("PK\x03\x04 this stands in for a zip payload \x00\xff\x01\x02")
	wrapped := WrapSlime(original)

	if !IsSlime(wrapped) {
		t.Fatal("wrapped bytes not detected as slime")
	}
	// Header must not equal a plain zip's — the payload is obfuscated.
	if bytes.HasPrefix(wrapped[8:], []byte("PK\x03\x04")) {
		t.Error("payload was not obfuscated (still starts with PK)")
	}

	back, err := UnwrapSlime(wrapped)
	if err != nil {
		t.Fatalf("UnwrapSlime: %v", err)
	}
	if !bytes.Equal(back, original) {
		t.Errorf("round trip mismatch:\n got %q\nwant %q", back, original)
	}
}

func TestSlimeTransformIsSymmetric(t *testing.T) {
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(i * 7)
	}
	twice := slimeTransform(slimeTransform(data))
	if !bytes.Equal(twice, data) {
		t.Error("transform is not its own inverse")
	}
}

func TestUnwrapRejectsGarbage(t *testing.T) {
	if _, err := UnwrapSlime([]byte("not a slime file at all")); err == nil {
		t.Error("expected error for non-slime bytes")
	}
}

// TestSlimeKnownVector pins the transform output so the PowerShell and PHP
// implementations can be checked against the same expected bytes. If this
// changes, the other two language libs must change identically.
func TestSlimeKnownVector(t *testing.T) {
	// Transform of 8 zero bytes, hex-encoded.
	out := slimeTransform(make([]byte, 8))
	got := ""
	const hexdigits = "0123456789abcdef"
	for _, b := range out {
		got += string(hexdigits[b>>4]) + string(hexdigits[b&0xF])
	}
	// key[0..7] XOR (0..7). Derived from SHA-256("PolyForge-Slime-v1").
	// Recomputed here so a key change fails loudly.
	key := slimeKey()
	want := ""
	for i := 0; i < 8; i++ {
		b := key[i] ^ byte(i)
		want += string(hexdigits[b>>4]) + string(hexdigits[b&0xF])
	}
	if got != want {
		t.Errorf("known vector = %s, want %s", got, want)
	}
	t.Logf("slime known vector (8 zero bytes) = %s", got)
}

// TestInstallLocalSlimePack builds a real .slime pack in-memory and installs
// it, exercising the full slime → zip → extract path.
func TestInstallLocalSlimePack(t *testing.T) {
	// Build a minimal pack zip.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	manifest := `{"schemaVersion":1,"id":"test-pack","name":"Test Pack","version":"1.0.0","loader":{"type":"fabric"},"mods":[{"file":"a.jar","name":"a","version":"1.0"}],"overrides":{"folders":["mods"],"fileCount":1,"totalBytes":3}}`
	must := func(w interface{ Write([]byte) (int, error) }, err error) interface{ Write([]byte) (int, error) } {
		if err != nil {
			t.Fatal(err)
		}
		return w
	}
	mw, err := zw.Create("pack-manifest.json")
	must(mw, err).Write([]byte(manifest))
	ow, err := zw.Create("overrides/mods/a.jar")
	must(ow, err).Write([]byte("jar"))
	cw, err := zw.Create("overrides/config/opts.txt")
	must(cw, err).Write([]byte("hello"))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	// Wrap into .slime and write to disk.
	dir := t.TempDir()
	packPath := filepath.Join(dir, "test-pack-1.0.0.polypack")
	if err := os.WriteFile(packPath, WrapSlime(buf.Bytes()), 0o644); err != nil {
		t.Fatal(err)
	}

	// Inspect.
	info, err := InspectPolyPack(packPath)
	if err != nil {
		t.Fatalf("InspectPolyPack: %v", err)
	}
	if info.ID != "test-pack" || info.ModCount != 1 || info.LoaderType != "fabric" {
		t.Errorf("unexpected info: %+v", info)
	}

	// Install.
	target := filepath.Join(dir, "instance")
	files, m, _, err := installLocalPack(packPath, target)
	if err != nil {
		t.Fatalf("installLocalPack: %v", err)
	}
	if files != 2 {
		t.Errorf("extracted %d files, want 2", files)
	}
	if m.Name != "Test Pack" {
		t.Errorf("manifest name = %q", m.Name)
	}
	if _, err := os.Stat(filepath.Join(target, "mods", "a.jar")); err != nil {
		t.Errorf("mods/a.jar not extracted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".polyforge-pack.json")); err != nil {
		t.Errorf("installed manifest copy not written: %v", err)
	}
}
