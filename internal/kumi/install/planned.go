package install

import ktypes "polyforge/internal/kumi/types"

// Planned launcher adapters — these all follow the MultiMC-family instance
// pattern via InstanceWithOptionalZip. Layout facts below are verified
// against the machine test 2 dump (2026-07-11) unless marked otherwise;
// per-launcher status and the migration path to PackInstance (packflow.go)
// live in docs/launcher-install-matrix.md.

// TODO(matrix #3): wrong layout — SKLauncher has no instances\ of its own.
// It uses the vanilla .minecraft directly and registers instances as
// launcher_profiles.json entries with gameDir .minecraft\instances\<name>
// (machine test 2). Rewire to the vanilla flow instead of this module.
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

// QWERTZ stores instances under profiles\<name> (the profile root is the
// game dir) with a profiles.json registry at the launcher root — schema
// captured in machine test 2 and written by kumi's genQwertzProfile.
func QWERTZ(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "QWERTZ", candidates, "profiles", "TurtelSMP5", zipURL, warning)
}

func Fjord(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Fjord Launcher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

// Parked (language barrier); instances\ subdir unverified.
func HMCL(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "HMCL", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

// Parked (no obtainable download); layout assumed from its MultiMC lineage.
func UltimMC(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "UltimMC", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

// Polymerium (https://github.com/d3ara1n/Polymerium): instances live in
// %LOCALAPPDATA%\Trident\instances\<id>.
// TODO(matrix #4): the live game dir appears to be instances\<id>\build\
// after Polymerium deploys — files at the instance root may be invisible
// in-game. Verify on the reference machine before staging overrides here.
func Polymerium(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "Polymerium", candidates, "instances", "TurtelSMP5", zipURL, warning)
}

// XMCL (https://github.com/Voxelum/x-minecraft-launcher): instances live in
// %USERPROFILE%\.minecraftx\instances\<name> (the instance root is the game
// dir), and XMCL rescans that folder on boot, so its Roaming\xmcl
// instances.json registry needs no upsert.
func XMCL(deps Dependencies, candidates []string, zipURL string, warning string) (*ktypes.ActionResult, error) {
	return InstanceWithOptionalZip(deps, "X Minecraft Launcher", candidates, "instances", "TurtelSMP5", zipURL, warning)
}
