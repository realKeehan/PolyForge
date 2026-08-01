package install

import ktypes "polyforge/internal/kumi/types"

// Parked (2026-07-11, language barrier). The instances\ subdir below is a
// KUMI.ps1-era guess that no dump has confirmed — BakaXL points its game
// dirs through CoreDirectory.json, whose schema is still uncaptured (the
// file was scrubbed from the machine test 2 dump; dump script v4 captures
// it next run). See docs/launcher-install-matrix.md.
func BakaXL(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "BakaXL", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
