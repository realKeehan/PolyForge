package install

import ktypes "polyforge/internal/kumi/types"

// Parked (2026-07-11): whether PolyForge can target Technic at all hinges on
// it being a pack provider vs accepting local custom packs — installedPacks
// on the reference machine was empty, so there is no custom-pack entry to
// mirror. modpacks\<slug> is the confirmed instance location if it ever
// unparks. See docs/launcher-install-matrix.md.
func Technic(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Technic", candidates, "modpacks", "TurtelSMP5", zipURL, warning)
}
