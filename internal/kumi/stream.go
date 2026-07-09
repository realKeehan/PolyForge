package kumi

// ══════════════════════════════════════════════════
// Live install streaming
//
// Long installs (download → extract → verify) used to run silently and only
// hand the UI a finished log at the end. To show progress as it happens, the
// service emits events while it works: one per log line, plus a progress event
// carrying a percentage + label for a live progress bar.
//
// The kumi package stays decoupled from Wails — the app layer injects an
// emitter (see App.Startup → SetEmitter) that forwards to runtime.EventsEmit.
// When no emitter is set (unit tests, `go run`), emitting is a no-op and the
// install still returns its full ActionResult as before.
// ══════════════════════════════════════════════════

// InstallEventName is the Wails event the frontend listens on for both live log
// lines and progress-bar updates during an install.
const InstallEventName = "install:event"

// InstallEvent is one streamed update. Kind is "log" for a log line or
// "progress" for a progress-bar update. Percent is 0–100, or negative to mean
// "indeterminate" (unknown total — show an animated bar).
type InstallEvent struct {
	Kind          string `json:"kind"`
	Level         string `json:"level,omitempty"`
	Message       string `json:"message,omitempty"`
	Percent       int    `json:"percent"`
	Label         string `json:"label,omitempty"`
	Indeterminate bool   `json:"indeterminate,omitempty"`
}

// SetEmitter installs the callback used to stream install events to the UI.
// Passing nil disables streaming.
func (s *Service) SetEmitter(fn func(event string, data ...interface{})) {
	s.emitFn = fn
}

func (s *Service) emit(ev InstallEvent) {
	if s.emitFn != nil {
		s.emitFn(InstallEventName, ev)
	}
}

// logStep records a log line on the result AND streams it live, so a caller can
// switch a block of result.Info/Warning/Error calls to streaming by swapping in
// this one helper. Level is "info", "warning", or "error".
func (s *Service) logStep(result *ActionResult, level, message string) {
	switch level {
	case "warning":
		result.Warning(message)
	case "error":
		result.Error(message)
	default:
		result.Info(message)
	}
	s.emit(InstallEvent{Kind: "log", Level: level, Message: message})
}

// emitProgress streams a determinate progress update (percent clamped 0–100).
// A negative percent is forwarded as an indeterminate update.
func (s *Service) emitProgress(percent int, label string) {
	if percent < 0 {
		s.emit(InstallEvent{Kind: "progress", Indeterminate: true, Label: label})
		return
	}
	if percent > 100 {
		percent = 100
	}
	s.emit(InstallEvent{Kind: "progress", Percent: percent, Label: label})
}

// emitStage streams an indeterminate stage change (e.g. "Installing…") so the
// bar keeps animating while a step with no byte total runs.
func (s *Service) emitStage(label string) {
	s.emit(InstallEvent{Kind: "progress", Indeterminate: true, Label: label})
}
