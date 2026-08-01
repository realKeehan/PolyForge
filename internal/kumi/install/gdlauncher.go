package install

import ktypes "polyforge/internal/kumi/types"

// TODO(matrix #1): GDLauncher Carbon nests instances under data\instances
// (machine test 2: gdlauncher_carbon\data\instances\<name>\instance.json),
// so both this subdir and kumi's instancesDirName("gdlauncher") target the
// wrong folder when handed the detected gdlauncher_carbon root. Fix both
// together — see docs/launcher-install-matrix.md.
func GDLauncher(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "GDLauncher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
