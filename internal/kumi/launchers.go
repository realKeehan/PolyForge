package kumi

import (
	"os"
	"path/filepath"

	"polyforge/internal/kumi/install"
)

func (s *Service) installMultiMC(explicitRoot string) (*ActionResult, error) {
	return install.MultiMC(s.installDependencies(), explicitRoot, multimcZipURL)
}

func (s *Service) installCurseForge() (*ActionResult, error) {
	return install.CurseForge(s.installDependencies(), curseforgeZipURL)
}

func (s *Service) installModrinth() (*ActionResult, error) {
	return install.Modrinth(s.installDependencies(), modrinthZipURL)
}

func (s *Service) installGDLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher_next"),
		filepath.Join(os.Getenv("APPDATA"), "gdlauncher"),
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "GDLauncher", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installATLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "ATLauncher"),
		`C:\\ATLauncher`,
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "ATLauncher", candidates, "Instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installPrismLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher"),
		filepath.Join(os.Getenv("APPDATA"), "PrismLauncher", "minecraft"),
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "PrismLauncher", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installBakaXL(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "BakaXL"),
		`C:\\BakaXL`,
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "BakaXL", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installFeather(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "feather"),
		filepath.Join(os.Getenv("APPDATA"), "FeatherClient"),
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "Feather", candidates, "profiles", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installTechnic(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), ".technic"),
		`C:\\.technic`,
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "Technic", candidates, "modpacks", "TurtelSMP5", "", emptyZipWarning)
}

func (s *Service) installPolyMC(explicitRoot string) (*ActionResult, error) {
	candidates := []string{
		explicitRoot,
		filepath.Join(os.Getenv("APPDATA"), "PolyMC"),
		filepath.Join(os.Getenv("APPDATA"), "polymc"),
	}
	return install.InstanceWithOptionalZip(s.installDependencies(), "PolyMC", candidates, "instances", "TurtelSMP5", "", emptyZipWarning)
}
