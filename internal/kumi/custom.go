package kumi

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Service) installCustomMods(modsDir string) (*ActionResult, error) {
	result := newResult()
	if modsDir == "" {
		result.Error("mods directory is required")
		result.Success = false
		return result, nil
	}
	if err := ensureDir(modsDir); err != nil {
		result.Error(fmt.Sprintf("failed to prepare mods directory: %v", err))
		result.Success = false
		return result, nil
	}

	bypass := filepath.Join(modsDir, "bypass.turtel")
	if pathExists(bypass) {
		result.Warning("bypass.turtel present — skipping mod installation")
		result.Success = true
		return result, nil
	}

	result.Info("Moving existing files to 'not-turtel'")
	notTurtel := filepath.Join(modsDir, "not-turtel")
	if err := ensureDir(notTurtel); err != nil {
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

	if err := s.downloadAndExtract(customZipURL, modsDir, "TurtelCustom.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install custom mods: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info("Installed custom mod files")
	result.Success = true
	return result, nil
}

func (s *Service) installManual(target string) (*ActionResult, error) {
	result := newResult()
	var err error
	if target == "" {
		target, err = os.Getwd()
		if err != nil {
			result.Error(fmt.Sprintf("failed to determine working directory: %v", err))
			result.Success = false
			return result, nil
		}
	}
	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to prepare target directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := s.downloadAndExtract(manualZipURL, target, "TurtelManual.zip"); err != nil {
		result.Error(fmt.Sprintf("manual install failed: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Manual package extracted to %s", target))
	result.Success = true
	return result, nil
}

func (s *Service) installInstanceWithOptionalZip(label string, candidates []string, subDir string, instanceName string, zipURL string, warning string) (*ActionResult, error) {
	result := newResult()
	root := firstExistingDirectory(candidates)
	if root == "" {
		result.Error(fmt.Sprintf("Unable to locate %s root. Please provide a valid path.", label))
		result.Success = false
		return result, nil
	}

	target := filepath.Join(root, subDir, instanceName)
	if pathExists(target) {
		result.Warning(fmt.Sprintf("%s instance already exists at %s", label, target))
		result.Success = true
		return result, nil
	}

	if err := ensureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create %s target directory: %v", label, err))
		result.Success = false
		return result, nil
	}

	if zipURL == "" {
		result.Warning(warning)
		result.Success = true
		return result, nil
	}

	if err := s.downloadAndExtract(zipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision %s instance: %v", label, err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Installed %s instance to %s", label, target))
	result.Success = true
	return result, nil
}
