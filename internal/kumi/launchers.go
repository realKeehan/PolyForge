package kumi

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Service) installMultiMC(explicitRoot string) (*ActionResult, error) {
	result := newResult()
	root := firstExisting([]string{
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
	if pathExists(instanceDir) {
		result.Warning("MultiMC instance already exists - skipping download")
		result.Success = true
		return result, nil
	}

	if err := ensureDir(instanceDir); err != nil {
		result.Error(fmt.Sprintf("failed to create instance directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := s.downloadAndExtract(multimcZipURL, instanceDir, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision MultiMC instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed MultiMC instance to %s", instanceDir))
	result.Success = true
	return result, nil
}

func (s *Service) installCurseForge() (*ActionResult, error) {
	result := newResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}
	target := filepath.Join(home, "curseforge", "minecraft", "Instances", "TurtelSMP5")
	if pathExists(target) {
		result.Warning("CurseForge instance already present - skipping")
		result.Success = true
		return result, nil
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create CurseForge directory: %v", err))
		result.Success = false
		return result, nil
	}
	if err := s.downloadAndExtract(curseforgeZipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to install CurseForge instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed CurseForge instance to %s", target))
	result.Success = true
	return result, nil
}

func (s *Service) installModrinth() (*ActionResult, error) {
	result := newResult()
	home, err := os.UserHomeDir()
	if err != nil {
		result.Error(fmt.Sprintf("unable to resolve user home: %v", err))
		result.Success = false
		return result, nil
	}
	target := filepath.Join(home, "AppData", "Roaming", "com.modrinth.theseus", "profiles", "TurtelSMP5")
	if pathExists(target) {
		result.Warning("Modrinth instance already present - skipping")
		result.Success = true
		return result, nil
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create Modrinth directory: %v", err))
		result.Success = false
		return result, nil
	}
	if err := s.downloadAndExtract(modrinthZipURL, target, "TurtelModrinth.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install Modrinth instance: %v", err))
		result.Success = false
		return result, nil
	}
	result.Info(fmt.Sprintf("Installed Modrinth instance to %s", target))
	result.Success = true
	return result, nil
}

func (s *Service) installGDLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher_next"),
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher"),
	}
	return s.installInstanceWithOptionalZip("GDLauncher", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installATLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "ATLauncher"),
		`C:\\ATLauncher`,
	}
	return s.installInstanceWithOptionalZip("ATLauncher", candidates, "Instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installPrismLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher"),
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher", "minecraft"),
	}
	return s.installInstanceWithOptionalZip("PrismLauncher", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installBakaXL(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "BakaXL"),
		`C:\\BakaXL`,
	}
	return s.installInstanceWithOptionalZip("BakaXL", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installFeather(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "feather"),
		filepath.Join(os.Getenv("APPDATA"), "FeatherClient"),
	}
	return s.installInstanceWithOptionalZip("Feather", candidates, "profiles", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installTechnic(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), ".technic"),
		`C:\\.technic`,
	}
	return s.installInstanceWithOptionalZip("Technic", candidates, "modpacks", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installPolyMC(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "PolyMC"),
		filepath.Join(os.Getenv("APPDATA"), "polymc"),
	}
	return s.installInstanceWithOptionalZip("PolyMC", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}
