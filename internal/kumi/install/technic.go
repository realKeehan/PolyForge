package install

import ktypes "polyforge/internal/kumi/types"

func Technic(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Technic", candidates, "modpacks", "TurtelSMP5", zipURL, warning)
}
