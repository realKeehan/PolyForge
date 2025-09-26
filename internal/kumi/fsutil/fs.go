package fsutil

import (
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func PathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func FirstExisting(candidates []string, exeName string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		probe := candidate
		if filepath.Ext(candidate) == "" {
			probe = filepath.Join(candidate, exeName)
		}
		if PathExists(probe) {
			return probe
		}
	}
	return ""
}

func FirstExistingDirectory(candidates []string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			return candidate
		}
	}
	return ""
}
