package kumi

import (
	"fmt"

	"polyforge/internal/kumi/install"
)

func (s *Service) installMultiMC(explicitRoot string) (*ActionResult, error) {
	candidates := multiMCCandidates(explicitRoot)
	return install.MultiMC(s.installDependencies(), candidates, multimcZipURL)
}

func (s *Service) installCurseForge() (*ActionResult, error) {
	target, err := curseForgeTarget()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve CurseForge path: %w", err)
	}
	return install.CurseForge(s.installDependencies(), target, curseforgeZipURL)
}

func (s *Service) installModrinth() (*ActionResult, error) {
	target, err := modrinthTarget()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve Modrinth path: %w", err)
	}
	return install.Modrinth(s.installDependencies(), target, modrinthZipURL)
}

func (s *Service) installGDLauncher(explicitRoot string) (*ActionResult, error) {
	return install.GDLauncher(s.installDependencies(), gdLauncherCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installATLauncher(explicitRoot string) (*ActionResult, error) {
	return install.ATLauncher(s.installDependencies(), atLauncherCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installPrismLauncher(explicitRoot string) (*ActionResult, error) {
	return install.PrismLauncher(s.installDependencies(), prismLauncherCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installBakaXL(explicitRoot string) (*ActionResult, error) {
	return install.BakaXL(s.installDependencies(), bakaXLCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installFeather(explicitRoot string) (*ActionResult, error) {
	return install.Feather(s.installDependencies(), featherCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installTechnic(explicitRoot string) (*ActionResult, error) {
	return install.Technic(s.installDependencies(), technicCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installPolyMC(explicitRoot string) (*ActionResult, error) {
	return install.PolyMC(s.installDependencies(), polyMCCandidates(explicitRoot), "", emptyZipWarning)
}
