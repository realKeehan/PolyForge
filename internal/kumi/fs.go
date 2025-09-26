package kumi

import (
	"os"
	"path/filepath"
	"strings"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func pathExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func firstExisting(candidates []string, exeName string) string {
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		probe := candidate
		if filepath.Ext(candidate) == "" {
			probe = filepath.Join(candidate, exeName)
		}
		if pathExists(probe) {
			return probe
		}
	}
	return ""
}

func firstExistingDirectory(candidates []string) string {
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
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
