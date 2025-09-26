package install

import (
	"fmt"

	ktypes "polyforge/internal/kumi/types"
)

func Modrinth(deps Dependencies, target string, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	if deps.PathExists(target) {
		result.Warning("Modrinth instance already present - skipping")
		result.Success = true
		return result, nil
	}

	if err := deps.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create Modrinth directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, target, "TurtelModrinth.zip"); err != nil {
		result.Error(fmt.Sprintf("failed to install Modrinth instance: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Installed Modrinth instance to %s", target))
	result.Success = true
	return result, nil
}
