package install

import ktypes "polyforge/internal/kumi/types"

func GDLauncher(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "GDLauncher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
