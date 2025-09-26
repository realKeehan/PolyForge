package install

import ktypes "polyforge/internal/kumi/types"

func PrismLauncher(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "PrismLauncher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
