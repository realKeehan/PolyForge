package install

import (
	"fmt"
	"os"
	"strings"

	ktypes "polyforge/internal/kumi/types"
)

func Manual(deps Dependencies, target, zipURL string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	target = strings.Trim(target, "\"")

	var err error
	if strings.TrimSpace(target) == "" {
		target, err = os.Getwd()
		if err != nil {
			result.Error(fmt.Sprintf("failed to determine working directory: %v", err))
			result.Success = false
			return result, nil
		}
	}

	if err := deps.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to prepare target directory: %v", err))
		result.Success = false
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, target, "TurtelManual.zip"); err != nil {
		result.Error(fmt.Sprintf("manual install failed: %v", err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Manual package extracted to %s", target))
	result.Success = true
	return result, nil
}
