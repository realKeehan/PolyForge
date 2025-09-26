package kumi

import "polyforge/internal/kumi/install"

func (s *Service) installCustomMods(modsDir string) (*ActionResult, error) {
	return install.CustomMods(s.installDependencies(), modsDir, customZipURL)
}

func (s *Service) installManual(target string) (*ActionResult, error) {
	return install.Manual(s.installDependencies(), target, manualZipURL)
}

func (s *Service) installInstanceWithOptionalZip(label string, candidates []string, subDir string, instanceName string, zipURL string, warning string) (*ActionResult, error) {
	return install.InstanceWithOptionalZip(s.installDependencies(), label, candidates, subDir, instanceName, zipURL, warning)
}
