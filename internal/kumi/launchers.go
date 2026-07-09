package kumi

import (
	"fmt"

	"polyforge/internal/kumi/install"
)

// Install-time candidate lists get discovery applied (cache → shortcuts →
// bounded scan, see detect.go) for launchers that are commonly portable and
// keep their instances next to the executable. The plain candidate functions
// stay cheap because startup detection probes all of them.

func (s *Service) installMultiMC(explicitRoot string) (*ActionResult, error) {
	candidates := withExeDiscovery("multimc", multiMCCandidates(explicitRoot))
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
	candidates := withDirDiscovery("gdlauncher", gdLauncherCandidates(explicitRoot))
	return install.GDLauncher(s.installDependencies(), candidates, "", emptyZipWarning)
}

func (s *Service) installATLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := withDirDiscovery("atlauncher", atLauncherCandidates(explicitRoot))
	return install.ATLauncher(s.installDependencies(), candidates, "", emptyZipWarning)
}

func (s *Service) installPrismLauncher(explicitRoot string) (*ActionResult, error) {
	candidates := withDirDiscovery("prismlauncher", prismLauncherCandidates(explicitRoot))
	return install.PrismLauncher(s.installDependencies(), candidates, "", emptyZipWarning)
}

func (s *Service) installBakaXL(explicitRoot string) (*ActionResult, error) {
	candidates := withDirDiscovery("bakaxl", bakaXLCandidates(explicitRoot))
	return install.BakaXL(s.installDependencies(), candidates, "", emptyZipWarning)
}

func (s *Service) installFeather(explicitRoot string) (*ActionResult, error) {
	return install.Feather(s.installDependencies(), featherCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installTechnic(explicitRoot string) (*ActionResult, error) {
	return install.Technic(s.installDependencies(), technicCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installPolyMC(explicitRoot string) (*ActionResult, error) {
	candidates := withDirDiscovery("polymc", polyMCCandidates(explicitRoot))
	return install.PolyMC(s.installDependencies(), candidates, "", emptyZipWarning)
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
	candidates := withDirDiscovery("fjord", fjordCandidates(explicitRoot))
	return install.Fjord(s.installDependencies(), candidates, "", emptyZipWarning)
}

func (s *Service) installHMCL(explicitRoot string) (*ActionResult, error) {
	return install.HMCL(s.installDependencies(), hmclCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installUltimMC(explicitRoot string) (*ActionResult, error) {
	candidates := withDirDiscovery("ultimmc", ultimMCCandidates(explicitRoot))
	return install.UltimMC(s.installDependencies(), candidates, "", emptyZipWarning)
}

func (s *Service) installPolymerium(explicitRoot string) (*ActionResult, error) {
	return install.Polymerium(s.installDependencies(), polymeriumCandidates(explicitRoot), "", emptyZipWarning)
}

func (s *Service) installXMCL(explicitRoot string) (*ActionResult, error) {
	candidates := withDirDiscovery("xmcl", xmclCandidates(explicitRoot))
	return install.XMCL(s.installDependencies(), candidates, "", emptyZipWarning)
}
