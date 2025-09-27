package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ktypes "polyforge/internal/kumi/types"
)

func CustomMods(deps Dependencies, modsDir, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	modsDir = strings.Trim(modsDir, "\"")

	if strings.TrimSpace(modsDir) == "" {
		result.Error("mods directory is required")
		result.Success = false
		return result, nil
	}

	if err := deps.EnsureDir(modsDir); err != nil {
		result.Error(fmt.Sprintf("failed to prepare mods directory: %v", err))
		result.Success = false
		return result, nil
	}

	bypass := filepath.Join(modsDir, "bypass.turtel")
	if deps.PathExists(bypass) {
		result.Warning("bypass.turtel present — skipping mod installation")
		result.Success = true
		return result, nil
	}

	result.Info("Moving existing files to 'not-turtel'")
	notTurtel := filepath.Join(modsDir, "not-turtel")
	if err := deps.EnsureDir(notTurtel); err != nil {
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
