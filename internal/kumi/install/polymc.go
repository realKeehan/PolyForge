package install

import ktypes "polyforge/internal/kumi/types"

func PolyMC(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "PolyMC", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
