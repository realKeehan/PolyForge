package kumi

import (
	"fmt"
	"os"
	"strings"
)

// installHostedPack downloads a website-hosted .polypack and installs it into
// the chosen target directory, streaming download + install progress live. The
// download URL comes from the website (VerifyPackAccess / the pack registry);
// a root-relative "/packs/..." URL is resolved against the site base in the
// download layer, so the app never 404s on a path it has no host for.
func (s *Service) installHostedPack(payload ExecutionPayload) (*ActionResult, error) {
	result := NewResult()
	packURL := strings.TrimSpace(payload.Extra["packUrl"])
	packName := strings.TrimSpace(payload.Extra["packName"])
	target := strings.TrimSpace(payload.Path)
	if packName == "" {
		packName = "pack"
	}

	if packURL == "" {
		s.logStep(result, "error", "No download URL was provided for this pack.")
		result.Success = false
		return result, nil
	}
	if target == "" {
		s.logStep(result, "error", "No target directory selected.")
		result.Success = false
		return result, nil
	}
	if err := ensureDir(target); err != nil {
		s.logStep(result, "error", fmt.Sprintf("cannot create target directory: %v", err))
		result.Success = false
		return result, nil
	}

	s.logStep(result, "info", fmt.Sprintf("Downloading %s from the website…", packName))
	s.emitStage(fmt.Sprintf("Downloading %s…", packName))
	tmpPath, err := s.downloadToTemp(packURL, "Downloading "+packName)
	if err != nil {
		s.logStep(result, "error", fmt.Sprintf("download failed: %v", err))
		result.Success = false
		return result, nil
	}
	defer os.Remove(tmpPath)
	s.logStep(result, "info", "Download complete.")

	if !s.extractAndVerifyPack(result, tmpPath, target) {
		result.Success = false
		return result, nil
	}
	result.Success = true
	return result, nil
}

// extractAndVerifyPack extracts a pack file's overrides into target, verifies
// every extracted file against the manifest checksums, and records the install
// for future update/self-destruct passes. It streams each step live and returns
// whether the install succeeded (integrity failures return false). Shared by the
// hosted-pack and local-pack install paths.
func (s *Service) extractAndVerifyPack(result *ActionResult, packPath, target string) bool {
	s.emitStage("Extracting pack contents…")
	s.logStep(result, "info", "Extracting pack contents…")
	files, manifest, report, err := installLocalPack(packPath, target)
	if err != nil {
		s.logStep(result, "error", fmt.Sprintf("pack install failed: %v", err))
		return false
	}
	s.logStep(result, "info", fmt.Sprintf("Extracted %d files to %s", files, target))
	s.logStep(result, "info", fmt.Sprintf("Installed %s v%s (%d mods)", manifest.Name, manifest.Version, len(manifest.Mods)))

	// Integrity: verify every extracted file against the manifest checksums so
	// corruption, tampering, or a truncated download is caught before the user
	// launches the game. Older packs carry no per-file hashes (Total == 0).
	s.emitStage("Verifying integrity…")
	if report.Total == 0 {
		s.logStep(result, "info", "This pack predates per-file checksums; skipping integrity verification.")
	} else if report.OK() {
		s.logStep(result, "info", fmt.Sprintf("Integrity verified: %d/%d files match the manifest.", report.Checked, report.Total))
	} else {
		s.logStep(result, "error", fmt.Sprintf("Integrity check FAILED: %d of %d files did not match the manifest.", len(report.Issues), report.Total))
		for i, issue := range report.Issues {
			if i >= 20 {
				s.logStep(result, "warning", fmt.Sprintf("  …and %d more.", len(report.Issues)-20))
				break
			}
			s.logStep(result, "warning", fmt.Sprintf("  %s — %s", issue.Path, issue.Reason))
		}
		s.logStep(result, "error", "Re-download the pack and reinstall; do not launch this install.")
		return false
	}

	// Remember where this pack landed so remote mod removal (self-destruct) and
	// future update checks can find it without re-selecting the launcher.
	recordInstalledPack(manifest.ID, manifest.Name, manifest.Version, target)
	s.emitProgress(100, "Done")
	s.logStep(result, "warning", "Launcher profile generation is not wired up yet - add the instance to your launcher manually.")
	return true
}
