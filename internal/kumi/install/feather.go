package install

import ktypes "polyforge/internal/kumi/types"

func Feather(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Feather", candidates, "profiles", "TurtelSMP5", zipURL, warning)
}
