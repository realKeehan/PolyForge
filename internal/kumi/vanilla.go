package kumi

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Service) installVanilla() (*ActionResult, error) {
	result := newResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}

	vanillaDir := filepath.Join(home, "AppData", "Roaming", "KUMIProfiles", "TurtelSMP5")
	if err := ensureDir(vanillaDir); err != nil {
		result.Error(fmt.Sprintf("failed to create vanilla directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := s.downloadAndExtract(vanillaZipURL, vanillaDir, ""); err != nil {
		result.Error(fmt.Sprintf("vanilla download failed: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed vanilla files to %s", vanillaDir))

	versionsDir := filepath.Join(home, "AppData", "Roaming", ".minecraft", "versions")
	if pathExists(versionsDir) {
		if err := s.downloadAndExtract(quiltLoaderZipURL, versionsDir, ""); err != nil {
			result.Warning(fmt.Sprintf("Quilt loader download failed: %v", err))
		} else {
			result.Info("Installed Quilt Loader profile")
		}
	} else {
		result.Warning(fmt.Sprintf("Minecraft versions directory not found at %s", versionsDir))
	}

	launcherJSON := filepath.Join(home, "AppData", "Roaming", ".minecraft", "launcher_profiles.json")
	if err := addLauncherProfile(launcherJSON); err != nil {
		result.Warning(fmt.Sprintf("Failed to update launcher profiles: %v", err))
	} else {
		result.Info("Ensured launcher profile 'turtelsmp'")
	}

	result.Success = true
	return result, nil
}
