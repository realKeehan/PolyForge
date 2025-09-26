package install

// Dependencies encapsulates the shared helpers required by the individual installers.
type Dependencies struct {
	DownloadAndExtract     func(url, destination, explicitName string) error
	AddLauncherProfile     func(path string) error
	EnsureDir              func(path string) error
	PathExists             func(path string) bool
	FirstExisting          func(candidates []string, exeName string) string
	FirstExistingDirectory func(candidates []string) string
}
