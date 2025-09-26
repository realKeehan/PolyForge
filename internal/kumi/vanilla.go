package kumi

import "polyforge/internal/kumi/install"

func (s *Service) installVanilla() (*ActionResult, error) {
	return install.Vanilla(s.installDependencies(), vanillaZipURL, quiltLoaderZipURL)
}
