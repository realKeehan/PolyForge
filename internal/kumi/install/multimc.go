package install

import (
	"fmt"
	"path/filepath"

	ktypes "polyforge/internal/kumi/types"
)

func MultiMC(deps Dependencies, candidates []string, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	root := deps.FirstExisting(candidates, "MultiMC.exe")

	if root == "" {
		result.Warning("Unable to locate MultiMC.exe. Please provide the MultiMC root directory.")
		result.Success = false
		return result, nil
	}

	instanceDir := filepath.Join(filepath.Dir(root), "instances", "TurtelSMP5")
	if deps.PathExists(instanceDir) {
		result.Warning("MultiMC instance already exists - skipping download")
		result.Success = true
		return result, nil
	}

	if err := deps.EnsureDir(instanceDir); err != nil {
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
