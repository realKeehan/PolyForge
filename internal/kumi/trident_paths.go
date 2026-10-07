package kumi

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
)

// polymeriumCandidates resolves the Trident data home used by both Polymerium
// (its desktop shell) and the Trident CLI. Executable/config directories are
// not instance roots. An explicit home is authoritative even before it exists.
func polymeriumCandidates(explicit string) []string {
	if explicit != "" {
		return []string{filepath.Clean(explicit)}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	cwd, _ := os.Getwd()
	root := tridentDataHome(home, cwd, runtime.GOOS, os.Getenv)
	if root == "" {
		return nil
	}
	return []string{root}
}

// Resolution order follows Trident's PathDef. Platform folder names are part
// of the data contract, not the Polymerium application brand.
func tridentDataHome(home, cwd, platform string, getenv func(string) string) string {
	isFile := func(path string) bool {
		info, err := os.Stat(path)
		return err == nil && !info.IsDir()
	}
	selectHome := func(root string) string {
		// PathDef uses the platform home when its legacy fallback was selected
		// but does not exist, even when an override spells that same path.
		if root == filepath.Join(home, ".trident") && !directoryExists(root) {
			return tridentPlatformHome(home, platform, getenv)
		}
		return filepath.Clean(root)
	}
	if root := getenv("TRIDENT_HOME"); root != "" && !isFile(root) {
		return selectHome(root)
	}
	for dir := cwd; dir != ""; dir = filepath.Dir(dir) {
		if root := filepath.Join(dir, ".trident"); directoryExists(root) {
			return root
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	if file, err := os.Open(filepath.Join(home, ".trident.home")); err == nil {
		reader := bufio.NewReader(file)
		prefix, _ := reader.Peek(4)
		decoder := unicode.BOMOverride(unicode.UTF8.NewDecoder())
		if bytes.HasPrefix(prefix, []byte{0xff, 0xfe, 0, 0}) {
			decoder = utf32.UTF32(utf32.LittleEndian, utf32.ExpectBOM).NewDecoder()
		} else if bytes.HasPrefix(prefix, []byte{0, 0, 0xfe, 0xff}) {
			decoder = utf32.UTF32(utf32.BigEndian, utf32.ExpectBOM).NewDecoder()
		}
		scanner := bufio.NewScanner(transform.NewReader(reader, decoder))
		if scanner.Scan() {
			root := scanner.Text()
			file.Close()
			if strings.TrimSpace(root) != "" && filepath.IsAbs(root) && !isFile(root) {
				return selectHome(root)
			}
		} else {
			file.Close()
		}
	}
	if root := filepath.Join(home, ".trident"); directoryExists(root) {
		return root
	}
	return tridentPlatformHome(home, platform, getenv)
}

func tridentPlatformHome(home, platform string, getenv func(string) string) string {
	switch platform {
	case "windows":
		if local := getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "Trident")
		}
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "dev.dearain.trident")
	default:
		if data := getenv("XDG_DATA_HOME"); filepath.IsAbs(data) {
			return filepath.Join(data, "trident")
		}
		return filepath.Join(home, ".local", "share", "trident")
	}
	return ""
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
