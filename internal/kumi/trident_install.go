package kumi

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func copyTridentFile(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	return errors.Join(copyErr, out.Close())
}

// Swap the staged pack source, invalidate its ordinary runtime copies and save
// the profile last. Backups exist only for rollback (or explicit clean mode),
// never as an alternate description of package ownership.
func commitTridentInstall(instance, work string, profile map[string]any, mode InstallMode, owned map[string]string) (backup string, retainWork bool, err error) {
	imports := filepath.Join(instance, "import")
	stage := filepath.Join(work, "import")
	oldImport := filepath.Join(work, "old-import")
	liveBackup := filepath.Join(work, "live")
	oldExists := directoryExists(imports)
	var live []string
	for _, relative := range owned {
		path, err := tridentLiveFile(instance, relative)
		if err != nil {
			return "", false, err
		}
		if path != "" {
			live = append(live, relative)
		}
	}
	stagedProfile := filepath.Join(work, "profile.json")
	if err := writeJSONFile(stagedProfile, profile); err != nil {
		return "", false, err
	}
	recordPath, err := tridentRegularPath(instance, installedManifestName)
	if err != nil {
		return "", false, err
	}
	oldRecordExists := false
	if info, statErr := os.Lstat(recordPath); statErr == nil {
		if !info.Mode().IsRegular() {
			return "", false, fmt.Errorf("install record is not a regular file: %s", recordPath)
		}
		oldRecordExists = true
	} else if !os.IsNotExist(statErr) {
		return "", false, statErr
	}
	oldRecord := filepath.Join(work, "previous-polyforge-pack.json")
	stagedRecord := filepath.Join(work, installedManifestName)
	if mode == ModeClean && (oldExists || oldRecordExists || len(live) > 0) {
		if _, err := tridentRegularPath(instance, quarantineDirName); err != nil {
			return "", false, err
		}
		backup = filepath.Join(instance, quarantineDirName, time.Now().UTC().Format("20060102T150405.000000000Z"))
		if err := os.MkdirAll(backup, 0o755); err != nil {
			return "", false, err
		}
		oldImport = filepath.Join(backup, "import")
		liveBackup = filepath.Join(backup, "build")
		oldRecord = filepath.Join(backup, installedManifestName)
		if err := copyTridentFile(filepath.Join(instance, "profile.json"), filepath.Join(backup, "profile.json")); err != nil && !os.IsNotExist(err) {
			_ = os.RemoveAll(backup)
			return "", false, err
		}
	}
	build := filepath.Join(instance, "build")
	projectionManifest := filepath.Join(build, tridentImportManifest)
	manifestBackup := filepath.Join(work, tridentImportManifest)
	if backup != "" {
		manifestBackup = filepath.Join(liveBackup, tridentImportManifest)
	}
	var movedLive []string
	movedImport, promoted, movedManifest := false, false, false
	movedRecord, promotedRecord := false, false
	defer func() {
		if err == nil {
			return
		}
		var recovery []error
		restore := func(e error) {
			if e != nil {
				recovery = append(recovery, e)
			}
		}
		if promotedRecord {
			restore(os.Rename(recordPath, stagedRecord))
		}
		if movedRecord {
			restore(os.Rename(oldRecord, recordPath))
		}
		if promoted {
			restore(os.Rename(imports, stage))
		}
		if movedImport {
			restore(os.Rename(oldImport, imports))
		}
		for _, relative := range movedLive {
			restore(os.Rename(filepath.Join(liveBackup, filepath.FromSlash(relative)), filepath.Join(build, filepath.FromSlash(relative))))
		}
		if movedManifest {
			restore(os.Rename(manifestBackup, projectionManifest))
		}
		if len(recovery) != 0 {
			retainWork = true
			err = errors.Join(err, fmt.Errorf("rollback incomplete; recovery files retained at %s (previous import: %s): %w", work, oldImport, errors.Join(recovery...)))
		}
		if backup != "" && len(recovery) == 0 {
			_ = os.RemoveAll(backup)
		}
	}()
	for _, relative := range live {
		target := filepath.Join(liveBackup, filepath.FromSlash(relative))
		if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return
		}
		if err = os.Rename(filepath.Join(build, filepath.FromSlash(relative)), target); err != nil {
			return
		}
		movedLive = append(movedLive, relative)
	}
	if oldExists {
		if err = os.Rename(imports, oldImport); err != nil {
			return
		}
		movedImport = true
	}
	if err = os.Rename(stage, imports); err != nil {
		return
	}
	promoted = true
	// Merge-install leaves some old runtime copies intact. Keep their native
	// ownership manifest, including paths already removed from import/, until
	// deployment can reconcile them.
	if mode != ModeInstall {
		if _, statErr := os.Lstat(projectionManifest); statErr == nil {
			if err = os.MkdirAll(filepath.Dir(manifestBackup), 0o755); err != nil {
				return
			}
			if err = os.Rename(projectionManifest, manifestBackup); err != nil {
				return
			}
			movedManifest = true
		} else if !os.IsNotExist(statErr) {
			err = statErr
			return
		}
	}
	if oldRecordExists {
		if err = os.Rename(recordPath, oldRecord); err != nil {
			return
		}
		movedRecord = true
	}
	if err = os.Rename(stagedRecord, recordPath); err != nil {
		return
	}
	promotedRecord = true
	if err = os.Rename(stagedProfile, filepath.Join(instance, "profile.json")); err != nil {
		return
	}
	// The profile now describes the new source. Only empty runtime directories
	// are pruned; user files and persist/package links stop the removal walk.
	for _, relative := range movedLive {
		for parent := filepath.Dir(filepath.Join(build, filepath.FromSlash(relative))); parent != build; parent = filepath.Dir(parent) {
			if os.Remove(parent) != nil {
				break
			}
		}
	}
	return
}
