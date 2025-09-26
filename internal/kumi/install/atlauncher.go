package install

import ktypes "polyforge/internal/kumi/types"

func ATLauncher(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "ATLauncher", candidates, "Instances", "TurtelSMP5", zipURL, warning)
}
