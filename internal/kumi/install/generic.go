package install

import (
	"fmt"
	"path/filepath"
	"strings"

	ktypes "polyforge/internal/kumi/types"
)

func InstanceWithOptionalZip(deps Dependencies, label string, candidates []string, subDir string, instanceName string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	root := deps.FirstExistingDirectory(candidates)
	if root == "" {
		result.Error(fmt.Sprintf("Unable to locate %s root. Please provide a valid path.", label))
		result.Success = false
		return result, nil
	}

	target := filepath.Join(root, subDir, instanceName)
	if deps.PathExists(target) {
		result.Warning(fmt.Sprintf("%s instance already exists at %s", label, target))
		result.Success = true
		return result, nil
	}

	if err := deps.EnsureDir(target); err != nil {
		result.Error(fmt.Sprintf("failed to create %s target directory: %v", label, err))
		result.Success = false
		return result, nil
	}

	if strings.TrimSpace(zipURL) == "" {
		result.Warning(warning)
		result.Info(fmt.Sprintf("%s install complete.", label))
		result.Success = true
		return result, nil
	}

	if err := deps.DownloadAndExtract(zipURL, target, ""); err != nil {
		result.Error(fmt.Sprintf("failed to provision %s instance: %v", label, err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("Installed %s instance to %s", label, target))
	result.Info(fmt.Sprintf("%s install complete.", label))
	result.Success = true
	return result, nil
}
