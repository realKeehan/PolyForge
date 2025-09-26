package kumi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	osruntime "runtime"
	"sort"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (s *Service) SearchExecutable(query ExecutableSearchRequest) (*ActionResult, error) {
	result := newResult()
	exeName, args := splitExecutableQuery(query.Query)
	if exeName == "" {
		result.Error("no executable name found in query")
		result.Success = false
		return result, nil
	}

	preferredRoots := s.collectPreferredRoots()
	matches := s.scanRoots(preferredRoots, exeName, args)

	if len(matches) == 0 && query.SearchAllDrives {
		drives := s.enumerateDrives()
		matches = s.scanRoots(drives, exeName, args)
	}

	if len(matches) == 0 {
		result.Warning(fmt.Sprintf("No match for '%s' (exe: '%s'%s)", query.Query, exeName, formatArgs(args)))
		result.Success = false
		return result, nil
	}

	sort.Strings(matches)
	for _, match := range matches {
		result.Info(match)
	}
	result.Success = true
	return result, nil
}

func (s *Service) EnumerateApplications() (*ActionResult, error) {
	result := newResult()
	infos, err := enumerateApplications()
	if err != nil {
		result.Error(fmt.Sprintf("failed to enumerate applications: %v", err))
		result.Success = false
		return result, nil
	}
	if len(infos) == 0 {
		result.Warning("no applications detected")
		result.Success = true
		return result, nil
	}

	for _, info := range infos {
		result.Info(fmt.Sprintf("%s [%s] -> %s", info.Name, info.Kind, info.LaunchCommand))
	}
	result.Success = true
	return result, nil
}

func splitExecutableQuery(query string) (string, string) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "", ""
	}
	if strings.HasPrefix(trimmed, "\"") {
		parts := strings.SplitN(trimmed, "\"", 3)
		if len(parts) >= 3 {
			return parts[1], strings.TrimSpace(parts[2])
		}
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", ""
	}
	exe := fields[0]
	args := strings.Join(fields[1:], " ")
	return exe, args
}

func formatArgs(args string) string {
	if strings.TrimSpace(args) == "" {
		return ""
	}
	return fmt.Sprintf(", args contain: '%s'", args)
}

func (s *Service) collectPreferredRoots() []string {
	roots := []string{}
	add := func(path string) {
		if path != "" && pathExists(path) {
			roots = append(roots, path)
		}
	}

	user := os.Getenv("USERPROFILE")
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")
	programData := os.Getenv("ProgramData")

	add(filepath.Join(user, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch", "User Pinned"))
	add(filepath.Join(user, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch"))
	add(filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"))
	add(filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs"))
	add(filepath.Join(user, "curseforge", "minecraft", "Install"))
	add(filepath.Join(user, "Desktop"))
	add(os.Getenv("ProgramFiles"))
	add(os.Getenv("ProgramFiles(x86)"))
	add(programData)
	add("D:\\Program Files")
	add("D:\\Program Files (x86)")
	add("D:\\Programs")
	add(localAppData)
	if localAppData != "" {
		add(filepath.Join(filepath.Dir(localAppData), "LocalLow"))
	}
	add(filepath.Join(user, "Downloads"))
	add(appData)
	return roots
}

func (s *Service) scanRoots(roots []string, exeName string, args string) []string {
	var matches []string
	for _, root := range roots {
		if root == "" {
			continue
		}
		if s.ctx != nil {
			wailsruntime.EventsEmit(s.ctx, "search:progress", fmt.Sprintf("Scanning %s", root))
		}
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !strings.EqualFold(filepath.Base(path), exeName) {
				if strings.HasSuffix(strings.ToLower(path), ".lnk") {
					if resolved := resolveShortcut(path, exeName, args); resolved != "" {
						matches = append(matches, resolved)
						return filepath.SkipDir
					}
				}
				return nil
			}
			matches = append(matches, path)
			return filepath.SkipDir
		})
		if len(matches) > 0 {
			break
		}
	}
	return unique(matches)
}

func (s *Service) enumerateDrives() []string {
	drives := []string{}
	if osruntime.GOOS != "windows" {
		return drives
	}
	for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		root := fmt.Sprintf("%c:\\", letter)
		if pathExists(root) {
			drives = append(drives, root)
		}
	}
	return drives
}

func resolveShortcut(path, exeName, args string) string {
	// Placeholder: resolving Windows shortcuts requires COM and is left for a dedicated Windows build.
	return ""
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func enumerateApplications() ([]ApplicationInfo, error) {
	if osruntime.GOOS != "windows" {
		return []ApplicationInfo{}, nil
	}
	return []ApplicationInfo{}, errors.New("application enumeration is not implemented on this platform build")
}
