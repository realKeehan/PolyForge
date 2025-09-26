package kumi

import "time"

type OptionDescriptor struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	RequiresPath bool   `json:"requiresPath"`
	PathLabel    string `json:"pathLabel,omitempty"`
}

type ExecutionPayload struct {
	Path  string            `json:"path,omitempty"`
	Extra map[string]string `json:"extra,omitempty"`
}

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type ActionResult struct {
	Success  bool       `json:"success"`
	Messages []LogEntry `json:"messages"`
	// Timestamp marks when the result was produced which helps the frontend
	// merge log batches during long running actions.
	Timestamp time.Time `json:"timestamp"`
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

func newResult() *ActionResult {
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
