package install

import (
	"fmt"

	ktypes "polyforge/internal/kumi/types"
)

// Pack-driven install scaffold — the migration target for the legacy
// InstanceWithOptionalZip modules. Per-launcher status and ordering live in
// docs/launcher-install-matrix.md.
//
// The legacy modules (ports of TemporaryDetectRef/OriginalKumiScripts/
// KUMI.ps1) hardcode the "TurtelSMP5" instance name plus a per-launcher
// subdir and download a fixed ZIP whose Discord CDN link has expired. The
// hosted-pack flow already knows every launcher's reference-verified layout
// (kumi.PlanInstallDirs) and writes the real config files
// (kumi.GenerateLauncherFiles) — but this package cannot import kumi (kumi
// imports install), so that knowledge arrives here as closures.
//
// Migration recipe, per launcher:
//
//  1. In kumi/launchers.go build the closures over the hosted pack:
//
//     plan := func(root string) (string, string) {
//         return PlanInstallDirs(id, root, instanceNameFor(id, m, l))
//     }
//     stage := func(mode ktypes.InstallMode, instanceDir, gameDir string) ([]string, error) {
//         // extractAndVerifyPack (hostedpack.go) already implements the
//         // whole mode-aware pipeline: resolveInstallMode + prepareGameDir
//         // (update sweep / reinstall repair / clean quarantine), extract,
//         // verify, record, GenerateLauncherFiles.
//     }
//
//  2. Swap install.X(deps, candidates, zipURL, warning) for
//     install.PackInstance(deps, label, candidates, mode, plan, stage).
//
//  3. Delete the launcher's ZIP const from service.go once nothing
//     references it; emptyZipWarning goes last.

type (
	// PlanFunc maps the resolved launcher root to the instance directory and
	// the game directory overrides are extracted into (they coincide for
	// launchers whose instance root is the game dir).
	PlanFunc func(root string) (instanceDir, gameDir string)

	// StageFunc materializes the pack: overrides into gameDir, launcher
	// config files into instanceDir, honoring the requested install mode
	// (kumi's resolveInstallMode/prepareGameDir implement the semantics —
	// update sweeps stale pack files, reinstall repairs, clean quarantines).
	// Returned notes land in the install log. Staging must be update-safe:
	// an existing instance is the update/repair path, not an error.
	StageFunc func(mode ktypes.InstallMode, instanceDir, gameDir string) (notes []string, err error)
)

// PackInstance is the pack-driven counterpart of InstanceWithOptionalZip:
// resolve the launcher root from the detection candidates, let the planner
// pick the real instance layout, and hand materialization to the stager
// along with the user's requested install mode. Unlike the legacy helper it
// does not skip when the instance already exists — modes own that decision.
func PackInstance(deps Dependencies, label string, candidates []string, mode ktypes.InstallMode, plan PlanFunc, stage StageFunc) (*ktypes.ActionResult, error) {
	result := ktypes.NewResult()

	root := deps.FirstExistingDirectory(candidates)
	if root == "" {
		result.Error(fmt.Sprintf("Unable to locate %s root. Please provide a valid path.", label))
		result.Success = false
		return result, nil
	}

	instanceDir, gameDir := plan(root)
	if err := deps.EnsureDir(gameDir); err != nil {
		result.Error(fmt.Sprintf("failed to create %s game directory: %v", label, err))
		result.Success = false
		return result, nil
	}

	notes, err := stage(mode, instanceDir, gameDir)
	for _, note := range notes {
		result.Info(note)
	}
	if err != nil {
		result.Error(fmt.Sprintf("failed to install %s pack: %v", label, err))
		result.Success = false
		return result, nil
	}

	result.Info(fmt.Sprintf("%s install complete.", label))
	result.Success = true
	return result, nil
}
