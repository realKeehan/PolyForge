package install

import (
	"fmt"
	"os"
	"path/filepath"

	"polyforge/internal/kumi/fsutil"
	ktypes "polyforge/internal/kumi/types"
)

type Dependencies struct {
	DownloadAndExtract func(url, destination, explicitName string) error
	AddLauncherProfile func(path string) error
}

func Vanilla(deps Dependencies, vanillaZipURL, quiltLoaderZipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}

	vanillaDir := filepath.Join(home, "AppData", "Roaming", "KUMIProfiles", "TurtelSMP5")
	if err := fsutil.EnsureDir(vanillaDir); err != nil {
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
	if fsutil.PathExists(versionsDir) {
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

func MultiMC(deps Dependencies, explicitRoot, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	root := fsutil.FirstExisting([]string{
		explicitRoot,
		filepath.Join(os.Getenv("USERPROFILE"), "MultiMC"),
		filepath.Join(os.Getenv("ProgramFiles")),
		filepath.Join(os.Getenv("ProgramFiles(x86)")),
	}, "MultiMC.exe")

	if root == "" {
		result.Warning("Unable to locate MultiMC.exe. Please provide the MultiMC root directory.")
		result.Success = false
		return result, nil
	}

	instanceDir := filepath.Join(filepath.Dir(root), "instances", "TurtelSMP5")
	if fsutil.PathExists(instanceDir) {
		result.Warning("MultiMC instance already exists - skipping download")
		result.Success = true
		return result, nil
	}

	if err := fsutil.EnsureDir(instanceDir); err != nil {
		result.Error(fmt.Sprintf("failed to create instance directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, instanceDir, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision MultiMC instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed MultiMC instance to %s", instanceDir))
	result.Success = true
	return result, nil
}

func CurseForge(deps Dependencies, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}
	target := filepath.Join(home, "curseforge", "minecraft", "Instances", "TurtelSMP5")
	if fsutil.PathExists(target) {
		result.Warning("CurseForge instance already present - skipping")
		result.Success = true
		return result, nil
	}
	if err := fsutil.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create CurseForge directory: %v", err))
		result.Success = false
		return result, nil
	}
	if err := deps.DownloadAndExtract(zipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to install CurseForge instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed CurseForge instance to %s", target))
	result.Success = true
	return result, nil
}

func Modrinth(deps Dependencies, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}
	target := filepath.Join(home, "AppData", "Roaming", "com.modrinth.theseus", "profiles", "TurtelSMP5")
	if fsutil.PathExists(target) {
		result.Warning("Modrinth instance already present - skipping")
		result.Success = true
		return result, nil
	}
	if err := fsutil.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create Modrinth directory: %v", err))
		result.Success = false
		return result, nil
	}
	if err := deps.DownloadAndExtract(zipURL, target, "TurtelModrinth.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install Modrinth instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed Modrinth instance to %s", target))
	result.Success = true
	return result, nil
}

func InstanceWithOptionalZip(deps Dependencies, label string, candidates []string, subDir string, instanceName string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	root := fsutil.FirstExistingDirectory(candidates)
	if root == "" {
		result.Error(fmt.Sprintf("Unable to locate %s root. Please provide a valid path.", label))
		result.Success = false
		return result, nil
	}

	target := filepath.Join(root, subDir, instanceName)
	if fsutil.PathExists(target) {
		result.Warning(fmt.Sprintf("%s instance already exists at %s", label, target))
		result.Success = true
		return result, nil
	}

	if err := fsutil.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create %s target directory: %v", label, err))
		result.Success = false
		return result, nil
	}

	if zipURL == "" {
		result.Warning(warning)
		result.Success = true
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision %s instance: %v", label, err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Installed %s instance to %s", label, target))
	result.Success = true
	return result, nil
}

func CustomMods(deps Dependencies, modsDir, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	if modsDir == "" {
		result.Error("mods directory is required")
		result.Success = false
		return result, nil
	}
	if err := fsutil.EnsureDir(modsDir); err != nil {
		result.Error(fmt.Sprintf("failed to prepare mods directory: %v", err))
		result.Success = false
		return result, nil
	}

	bypass := filepath.Join(modsDir, "bypass.turtel")
	if fsutil.PathExists(bypass) {
		result.Warning("bypass.turtel present — skipping mod installation")
		result.Success = true
		return result, nil
	}

	result.Info("Moving existing files to 'not-turtel'")
	notTurtel := filepath.Join(modsDir, "not-turtel")
	if err := fsutil.EnsureDir(notTurtel); err != nil {
		result.Error(fmt.Sprintf("failed to create not-turtel directory: %v", err))
		result.Success = false
		return result, nil
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		result.Error(fmt.Sprintf("failed to enumerate mods directory: %v", err))
		result.Success = false
		return result, nil
	}
	for _, entry := range entries {
		if entry.Name() == "not-turtel" {
			continue
		}
		source := filepath.Join(modsDir, entry.Name())
		dest := filepath.Join(notTurtel, entry.Name())
		if err := os.Rename(source, dest); err != nil {
			result.Warning(fmt.Sprintf("failed to move %s: %v", entry.Name(), err))
		}
	}

	if err := deps.DownloadAndExtract(zipURL, modsDir, "TurtelCustom.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install custom mods: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info("Installed custom mod files")
	result.Success = true
	return result, nil
}

func Manual(deps Dependencies, target, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()
	var err error
	if target == "" {
		target, err = os.Getwd()
		if err != nil {
			result.Error(fmt.Sprintf("failed to determine working directory: %v", err))
			result.Success = false
			return result, nil
		}
	}
	if err := fsutil.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to prepare target directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, target, "TurtelManual.zip"); err != nil {
		result.Error(fmt.Sprintf("manual install failed: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Manual package extracted to %s", target))
	result.Success = true
	return result, nil
}
