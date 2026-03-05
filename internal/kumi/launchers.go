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

func (s *Service) installSKLauncher(explicitRoot string) (*ActionResult, error) {
	return install.SKLauncher(s.installDependencies(), skLauncherCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installFreesm(explicitRoot string) (*ActionResult, error) {
	return install.Freesm(s.installDependencies(), freesmCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installElyPrism(explicitRoot string) (*ActionResult, error) {
	return install.ElyPrism(s.installDependencies(), elyPrismCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installShatteredPrism(explicitRoot string) (*ActionResult, error) {
	return install.ShatteredPrism(s.installDependencies(), shatteredPrismCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installQWERTZ(explicitRoot string) (*ActionResult, error) {
	return install.QWERTZ(s.installDependencies(), qwertzCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installFjord(explicitRoot string) (*ActionResult, error) {
	return install.Fjord(s.installDependencies(), fjordCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installHMCL(explicitRoot string) (*ActionResult, error) {
	return install.HMCL(s.installDependencies(), hmclCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installUltimMC(explicitRoot string) (*ActionResult, error) {
	return install.UltimMC(s.installDependencies(), ultimMCCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installPolymerium(explicitRoot string) (*ActionResult, error) {
	return install.Polymerium(s.installDependencies(), polymeriumCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installXMCL(explicitRoot string) (*ActionResult, error) {
	return install.XMCL(s.installDependencies(), xmclCandidates(explicitRoot), "", emptyZipWarning)
}
