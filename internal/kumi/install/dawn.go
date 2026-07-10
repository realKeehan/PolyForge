package install

import ktypes "polyforge/internal/kumi/types"

// Dawn (formerly Feather) keeps profiles under <root>/profiles, where each
// profile holds profile.json, content-index.json and a .minecraft game dir.
func Dawn(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Dawn", candidates, "profiles", "TurtelSMP5", zipURL, warning)
}
