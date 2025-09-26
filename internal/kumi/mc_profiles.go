package kumi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func addLauncherProfile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	profiles, ok := data["profiles"].(map[string]interface{})
	if !ok {
		profiles = make(map[string]interface{})
		data["profiles"] = profiles
	}

	if _, exists := profiles["turtelsmp"]; exists {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	gameDir := filepath.Join(home, "AppData", "Roaming", "KUMIProfiles", "TurtelSMP5")
	profile := map[string]interface{}{
		"gameDir":       gameDir,
		"icon":          launcherIconData,
		"javaArgs":      "-Xmx4G -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1NewSizePercent=20 -XX:G1ReservePercent=20 -XX:MaxGCPauseMillis=50 -XX:G1HeapRegionSize=32M",
		"lastUsed":      time.Now().UTC().Format(time.RFC3339Nano),
		"lastVersionId": "quilt-loader-0.22.0-beta.1-1.20.1",
		"name":          "TurtelSMP5",
		"type":          "",
	}

	profiles["turtelsmp"] = profile

	updated, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, updated, 0o644)
}
