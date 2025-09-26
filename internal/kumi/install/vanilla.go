package install

import (
	"fmt"
	"os"
	"path/filepath"

	ktypes "polyforge/internal/kumi/types"
)

func Vanilla(deps Dependencies, vanillaZipURL, quiltLoaderZipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}

	vanillaDir := filepath.Join(home, "AppData", "Roaming", "KUMIProfiles", "TurtelSMP5")
	if err := deps.EnsureDir(vanillaDir); err != nil {
		result.Error(fmt.Sprintf("failed to create vanilla directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := deps.DownloadAndExtract(vanillaZipURL, vanillaDir, ""); err != nil {
		result.Error(fmt.Sprintf("vanilla download failed: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed vanilla files to %s", vanillaDir))

	versionsDir := filepath.Join(home, "AppData", "Roaming", ".minecraft", "versions")
	if deps.PathExists(versionsDir) {
		if err := deps.DownloadAndExtract(quiltLoaderZipURL, versionsDir, ""); err != nil {
			result.Warning(fmt.Sprintf("Quilt loader download failed: %v", err))
		} else {
			result.Info("Installed Quilt Loader profile")
		}
	} else {
		result.Warning(fmt.Sprintf("Minecraft versions directory not found at %s", versionsDir))
	}

	if deps.AddLauncherProfile != nil {
		launcherJSON := filepath.Join(home, "AppData", "Roaming", ".minecraft", "launcher_profiles.json")
		if err := deps.AddLauncherProfile(launcherJSON); err != nil {
			result.Warning(fmt.Sprintf("Failed to update launcher profiles: %v", err))
		} else {
			result.Info("Ensured launcher profile 'turtelsmp'")
		}
	}

	result.Success = true
	return result, nil
}
