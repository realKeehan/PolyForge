//go:build windows

package kumi

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

// setupFileAssociation registers the .slime extension with PolyForge in the
// per-user registry hive (HKCU\Software\Classes), so double-clicking a pack
// opens it in the app. Per-user means no admin prompt and no drivers.
func setupFileAssociation() string {
	exe, err := os.Executable()
	if err != nil {
		return "First-run: could not resolve executable path for file association."
	}

	const progID = "PolyForge.slime"

	// .slime → ProgID
	if err := setRegString(`Software\Classes\`+SlimeExtension, "", progID); err != nil {
		return "First-run: file association skipped (" + err.Error() + ")."
	}

	// ProgID description + friendly name
	_ = setRegString(`Software\Classes\`+progID, "", "PolyForge modpack")
	_ = setRegString(`Software\Classes\`+progID+`\DefaultIcon`, "", fmt.Sprintf(`"%s",0`, exe))

	// Open command: PolyForge.exe "%1"
	if err := setRegString(`Software\Classes\`+progID+`\shell\open\command`, "", fmt.Sprintf(`"%s" "%%1"`, exe)); err != nil {
		return "First-run: file association partially set (" + err.Error() + ")."
	}

	return "First-run: registered .slime packs to open in PolyForge."
}

func setRegString(path, name, value string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(name, value)
}
