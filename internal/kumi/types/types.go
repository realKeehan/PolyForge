package types

import "time"

type OptionDescriptor struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	RequiresPath bool   `json:"requiresPath"`
	PathLabel    string `json:"pathLabel,omitempty"`
	DetectedPath string `json:"detectedPath,omitempty"`
	Found        bool   `json:"found"`
	// Info is reference text shown behind an ⓘ icon on the launcher row
	// (e.g. where Modrinth actually keeps its profiles). May be multi-line.
	Info string `json:"info,omitempty"`
}

type ExecutionPayload struct {
	Path  string            `json:"path,omitempty"`
	Extra map[string]string `json:"extra,omitempty"`
}

// InstallMode selects how a pack install treats an existing game directory.
// The UI requests one via ExecutionPayload.Extra["mode"]; empty/unknown
// values resolve as ModeAuto. Semantics live in kumi/installmodes.go.
type InstallMode string

const (
	// ModeAuto picks ModeUpdate when the target carries an installed-pack
	// record, ModeInstall otherwise.
	ModeAuto InstallMode = "auto"
	// ModeInstall lays the pack down over whatever is there (same-path files
	// are overwritten, everything else is left alone) — the legacy behavior.
	ModeInstall InstallMode = "install"
	// ModeUpdate removes files the previous pack version shipped that the
	// new one no longer ships, then extracts. User data is never a candidate.
	ModeUpdate InstallMode = "update"
	// ModeReinstall is repair: the same stale-file sweep as ModeUpdate plus a
	// full re-extract, fixing files that went missing or were modified.
	ModeReinstall InstallMode = "reinstall"
	// ModeClean quarantines the entire game dir into
	// .polyforge-quarantine\<timestamp>\ before extracting a pristine copy.
	// Nothing is deleted.
	ModeClean InstallMode = "clean"
)

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type ActionResult struct {
	Success   bool       `json:"success"`
	Messages  []LogEntry `json:"messages"`
	Timestamp time.Time  `json:"timestamp"`
}

type ModrinthCloneRequest struct {
	DBPath            string `json:"dbPath"`
	SourcePath        string `json:"sourcePath"`
	NewPath           string `json:"newPath"`
	NewName           string `json:"newName"`
	GameVersion       string `json:"gameVersion"`
	ModLoader         string `json:"modLoader"`
	ModLoaderVersion  string `json:"modLoaderVersion"`
	ResetLastPlayed   bool   `json:"resetLastPlayed"`
	ResetPlayCounters bool   `json:"resetPlayCounters"`
}

type ExecutableSearchRequest struct {
	Query           string `json:"query"`
	SearchAllDrives bool   `json:"searchAllDrives"`
}

type ApplicationInfo struct {
	Name            string `json:"name"`
	Kind            string `json:"kind"`
	TargetPath      string `json:"targetPath"`
	AppUserModelID  string `json:"appUserModelId"`
	PackageFullName string `json:"packageFullName"`
	LaunchCommand   string `json:"launchCommand"`
	Type            string `json:"type"`
}

func NewResult() *ActionResult {
	return &ActionResult{Messages: make([]LogEntry, 0, 4), Timestamp: time.Now().UTC()}
}

func (r *ActionResult) Info(message string) {
	r.Messages = append(r.Messages, LogEntry{Level: "info", Message: message})
}

func (r *ActionResult) Warning(message string) {
	r.Messages = append(r.Messages, LogEntry{Level: "warning", Message: message})
}

func (r *ActionResult) Error(message string) {
	r.Messages = append(r.Messages, LogEntry{Level: "error", Message: message})
}
