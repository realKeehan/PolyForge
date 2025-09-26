package install

import (
	"fmt"

	ktypes "polyforge/internal/kumi/types"
)

func CurseForge(deps Dependencies, target string, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	if deps.PathExists(target) {
		result.Warning("CurseForge instance already present - skipping")
		result.Success = true
		return result, nil
	}

	if err := deps.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create CurseForge directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to install CurseForge instance: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Installed CurseForge instance to %s", target))
	result.Success = true
	return result, nil
}
