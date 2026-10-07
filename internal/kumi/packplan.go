package kumi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// packInstallPlan contains only PolyForge layout, identity and mode decisions.
// Launcher-native validation happens when executing the prepared plan.
type packInstallPlan struct {
	InstanceDir string
	ContentDir  string
	RecordDir   string
	Mode        InstallMode
	Previous    *PackManifest
	Note        string
}

func planPackInstall(launcherID, chosen string, manifest *PackManifest, launchers *LaunchersFile, requested InstallMode) (packInstallPlan, error) {
	target := launcherTargets[launcherID]
	name := instanceNameFor(launcherID, manifest, launchers)
	if target.InstanceKey != nil {
		name = target.InstanceKey(manifest, launchers)
	}
	instance, content := PlanInstallDirs(launcherID, chosen, name)
	recordDir := content
	if target.RecordAtInstanceRoot {
		var err error
		instance, err = managedInstanceDirectory(launcherID, chosen, name, manifest.ID)
		if err != nil {
			return packInstallPlan{}, err
		}
		content = filepath.Join(instance, target.InstanceSubdir)
		recordDir = instance
	}
	mode, previous, note := resolveInstallMode(requested, recordDir)
	return packInstallPlan{
		InstanceDir: instance, ContentDir: content, RecordDir: recordDir,
		Mode: mode, Previous: previous, Note: note,
	}, nil
}

// Reuse only a matching PolyForge record. Occupied paths stay untouched, and
// suffixes give new instances unique keys without consulting native profiles.
func managedInstanceDirectory(launcherID, chosen, name, packID string) (string, error) {
	key := sanitizeInstanceName(name)
	if key == "" {
		return "", fmt.Errorf("instance name is empty after sanitizing")
	}
	chosen = filepath.Clean(chosen)
	instance, _ := PlanInstallDirs(launcherID, chosen, key)
	container := instancesDirName(launcherID)
	if (container != "" && strings.EqualFold(filepath.Base(filepath.Dir(chosen)), container)) ||
		pathExists(filepath.Join(chosen, installedManifestName)) {
		instance = chosen
	}
	instance, err := filepath.Abs(instance)
	if err != nil {
		return "", err
	}
	exists, matches, err := managedInstanceMatches(instance, packID, launcherID)
	if err != nil {
		return "", err
	}
	if !exists || matches {
		return instance, nil
	}
	parent := filepath.Dir(instance)
	for candidateKey := key; ; candidateKey += "_" {
		candidate := filepath.Join(parent, candidateKey)
		exists, matches, err := managedInstanceMatches(candidate, packID, launcherID)
		if err != nil {
			return "", err
		}
		if !exists || matches {
			return candidate, nil
		}
	}
}

func managedInstanceMatches(instance, packID, launcherID string) (exists, matches bool, err error) {
	info, err := os.Lstat(instance)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if !info.IsDir() {
		return true, false, nil
	}
	if info, err := os.Lstat(filepath.Join(instance, installedManifestName)); os.IsNotExist(err) {
		return true, false, nil
	} else if err != nil {
		return true, false, err
	} else if !info.Mode().IsRegular() {
		return true, false, nil
	}
	record, err := readInstalledPackRecord(instance)
	if err != nil {
		return true, false, fmt.Errorf("cannot identify existing instance %s: %w", instance, err)
	}
	return true, record.ID == packID && (record.Launcher == "" || record.Launcher == launcherID), nil
}
