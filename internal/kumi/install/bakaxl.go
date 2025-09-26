package install

import ktypes "polyforge/internal/kumi/types"

func BakaXL(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "BakaXL", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
