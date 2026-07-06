//go:build !windows

package kumi

// setupFileAssociation is a no-op on non-Windows platforms for now.
// TODO: Linux — write a .desktop entry + MIME type (application/x-polyforge-slime)
// via xdg-mime; macOS — declare the UTI in the .app bundle's Info.plist.
func setupFileAssociation() string {
	return ""
}
