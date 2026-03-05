package install

import ktypes "polyforge/internal/kumi/types"

// Planned launcher adapters — these all follow the MultiMC-family instance
// pattern via InstanceWithOptionalZip. ZIP URLs and instance subdirectories
// will be refined as each adapter matures.

func SKLauncher(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "SK Launcher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func Freesm(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Freesm Launcher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func ElyPrism(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "ElyPrism", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func ShatteredPrism(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "ShatteredPrism", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func QWERTZ(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "QWERTZ", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func Fjord(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Fjord Launcher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func HMCL(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "HMCL", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

func UltimMC(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "UltimMC", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

// Polymerium — a modern Minecraft launcher for Windows with a clean UI
// and modpack management capabilities.
// https://github.com/d3ara1n/Polymerium
func Polymerium(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Polymerium", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

// XMCL (X Minecraft Launcher) — an open-source Minecraft launcher
// supporting multiple accounts, modpacks, and resource management.
// https://github.com/Voxelum/x-minecraft-launcher
func XMCL(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "X Minecraft Launcher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
