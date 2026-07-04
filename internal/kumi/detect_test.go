package kumi

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// buildLnkWithLinkInfo synthesizes a minimal MS-SHLLINK file whose LinkInfo
// LocalBasePath points at target (ANSI variant, header size 0x1C).
func buildLnkWithLinkInfo(target string) []byte {
	base := append([]byte(target), 0)
	suffix := []byte{0}

	// LinkInfo: 28-byte header + strings
	liHeader := 28
	basePathOff := liHeader
	suffixOff := basePathOff + len(base)
	liSize := suffixOff + len(suffix)

	li := make([]byte, liSize)
	binary.LittleEndian.PutUint32(li[0:], uint32(liSize))
	binary.LittleEndian.PutUint32(li[4:], uint32(liHeader))
	binary.LittleEndian.PutUint32(li[8:], 0x1) // VolumeIDAndLocalBasePath
	binary.LittleEndian.PutUint32(li[16:], uint32(basePathOff))
	binary.LittleEndian.PutUint32(li[24:], uint32(suffixOff))
	copy(li[basePathOff:], base)
	copy(li[suffixOff:], suffix)

	header := make([]byte, lnkHeaderSize)
	binary.LittleEndian.PutUint32(header[0:], lnkHeaderSize)
	binary.LittleEndian.PutUint32(header[20:], lnkHasLinkInfo)

	return append(header, li...)
}

// buildLnkWithRelativePath synthesizes a shortcut that only carries an ANSI
// relative path string.
func buildLnkWithRelativePath(rel string) []byte {
	header := make([]byte, lnkHeaderSize)
	binary.LittleEndian.PutUint32(header[0:], lnkHeaderSize)
	binary.LittleEndian.PutUint32(header[20:], lnkHasRelativePath)

	str := make([]byte, 2+len(rel))
	binary.LittleEndian.PutUint16(str[0:], uint16(len(rel)))
	copy(str[2:], rel)

	return append(header, str...)
}

func TestParseShortcutLinkInfo(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Foo Launcher.exe")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	lnkPath := filepath.Join(dir, "foo.lnk")
	if err := os.WriteFile(lnkPath, buildLnkWithLinkInfo(target), 0o644); err != nil {
		t.Fatal(err)
	}

	got, args, err := parseShortcut(lnkPath)
	if err != nil {
		t.Fatalf("parseShortcut: %v", err)
	}
	if got != target {
		t.Errorf("target = %q, want %q", got, target)
	}
	if args != "" {
		t.Errorf("args = %q, want empty", args)
	}

	if resolved := resolveShortcut(lnkPath, "Foo Launcher.exe", ""); resolved != target {
		t.Errorf("resolveShortcut = %q, want %q", resolved, target)
	}
	if resolved := resolveShortcut(lnkPath, "Other.exe", ""); resolved != "" {
		t.Errorf("resolveShortcut with wrong exe = %q, want empty", resolved)
	}
}

func TestParseShortcutRelativePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "sub", "Bar.exe")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	lnkPath := filepath.Join(dir, "bar.lnk")
	rel := ".\\sub\\Bar.exe"
	if runtime.GOOS != "windows" {
		rel = "./sub/Bar.exe"
	}
	if err := os.WriteFile(lnkPath, buildLnkWithRelativePath(rel), 0o644); err != nil {
		t.Fatal(err)
	}

	got, _, err := parseShortcut(lnkPath)
	if err != nil {
		t.Fatalf("parseShortcut: %v", err)
	}
	if got != filepath.Clean(target) {
		t.Errorf("target = %q, want %q", got, target)
	}
}

func TestParseShortcutRejectsGarbage(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.lnk")
	if err := os.WriteFile(bad, []byte("this is not a shortcut"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := parseShortcut(bad); err == nil {
		t.Error("expected error for non-lnk file")
	}
}

// TestParseShortcutRealStartMenu opportunistically parses the shortcuts on
// the machine running the tests, ensuring the parser handles real-world
// .lnk files without errors on every target-bearing shortcut.
func TestParseShortcutRealStartMenu(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	parsed, withTarget := 0, 0
	for _, root := range shortcutRoots() {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".lnk") {
				return nil
			}
			parsed++
			if target, _, perr := parseShortcut(path); perr == nil && target != "" {
				withTarget++
			}
			return nil
		})
	}
	t.Logf("parsed %d shortcuts, %d yielded targets", parsed, withTarget)
	if parsed > 0 && withTarget == 0 {
		t.Errorf("no shortcut out of %d yielded a target — parser likely broken", parsed)
	}
}

func TestScanForExesFindsMultipleTargets(t *testing.T) {
	root := t.TempDir()
	mk := func(parts ...string) string {
		p := filepath.Join(append([]string{root}, parts...)...)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	a := mk("Programs", "AlphaLauncher", "Alpha.exe")
	b := mk("deep", "one", "two", "Beta.exe")
	mk("too", "deep", "one", "two", "three", "four", "Gamma.exe") // beyond depth 3

	wanted := map[string]string{
		"alpha.exe": "alpha",
		"beta.exe":  "beta",
		"gamma.exe": "gamma",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hits := scanForExes(ctx, []string{root}, wanted, 3, 4)

	if hits["alpha"] != a {
		t.Errorf("alpha = %q, want %q", hits["alpha"], a)
	}
	if hits["beta"] != b {
		t.Errorf("beta = %q, want %q", hits["beta"], b)
	}
	if _, ok := hits["gamma"]; ok {
		t.Error("gamma found beyond depth limit; depth cap not enforced")
	}
}

func TestScanForExesLargeTreeNoDeadlock(t *testing.T) {
	// Wide tree (many dirs) to exercise the bounded-concurrency walk; the
	// old channel-based version could deadlock when workers saturated.
	root := t.TempDir()
	for i := 0; i < 40; i++ {
		for j := 0; j < 20; j++ {
			dir := filepath.Join(root, "d"+string(rune('a'+i%26))+string(rune('0'+i/26)), "s"+string(rune('a'+j)))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	needle := filepath.Join(root, "dz1", "sx")
	if err := os.MkdirAll(needle, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(needle, "Needle.exe"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan map[string]string, 1)
	go func() {
		done <- scanForExes(ctx, []string{root}, map[string]string{"needle.exe": "n"}, 5, 8)
	}()

	select {
	case hits := <-done:
		if hits["n"] == "" {
			t.Error("needle not found")
		}
	case <-time.After(15 * time.Second):
		t.Fatal("scanForExes did not return — deadlock")
	}
}
