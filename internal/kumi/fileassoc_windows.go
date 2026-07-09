//go:build windows

package kumi

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// setupFileAssociation registers the .polypack extension with PolyForge in
// the per-user registry hive (HKCU\Software\Classes), so double-clicking a
// pack opens it in the app. Per-user means no admin prompt and no drivers.
// A generated icon distinguishes pack files from the app itself.
func setupFileAssociation() string {
	exe, err := os.Executable()
	if err != nil {
		return "First-run: could not resolve executable path for file association."
	}

	const progID = "PolyForge.polypack"

	// Generate the distinct pack icon; fall back to the app icon if writing
	// it fails for any reason.
	iconRef := fmt.Sprintf(`"%s",0`, exe)
	if iconPath, ierr := writePackIcon(); ierr == nil {
		iconRef = fmt.Sprintf(`"%s",0`, iconPath)
	}

	// .polypack → ProgID
	if err := setRegString(`Software\Classes\`+PackExtension, "", progID); err != nil {
		return "First-run: file association skipped (" + err.Error() + ")."
	}

	// ProgID description + icon
	_ = setRegString(`Software\Classes\`+progID, "", "PolyForge modpack")
	_ = setRegString(`Software\Classes\`+progID+`\DefaultIcon`, "", iconRef)

	// Open command: PolyForge.exe "%1"
	if err := setRegString(`Software\Classes\`+progID+`\shell\open\command`, "", fmt.Sprintf(`"%s" "%%1"`, exe)); err != nil {
		return "First-run: file association partially set (" + err.Error() + ")."
	}

	// Tell the shell associations changed so Explorer refreshes cached icons
	// immediately instead of after a reboot / icon-cache rebuild.
	notifyShellAssocChanged()

	return "First-run: registered .polypack packs to open in PolyForge."
}

// notifyShellAssocChanged raises SHCNE_ASSOCCHANGED so Explorer drops its
// cached .polypack icon and shows the freshly registered one right away.
func notifyShellAssocChanged() {
	const (
		SHCNE_ASSOCCHANGED = 0x08000000
		SHCNF_IDLIST       = 0x0000
	)
	shell32 := windows.NewLazySystemDLL("shell32.dll")
	proc := shell32.NewProc("SHChangeNotify")
	_, _, _ = proc.Call(uintptr(SHCNE_ASSOCCHANGED), uintptr(SHCNF_IDLIST), 0, 0)
}

// writePackIcon generates the .polypack file-type icon and writes it next to
// the app config, returning its path.
func writePackIcon() (string, error) {
	path, err := packIconPath()
	if err != nil {
		return "", err
	}
	ico, err := packIconICO()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, ico, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func setRegString(path, name, value string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(name, value)
}
