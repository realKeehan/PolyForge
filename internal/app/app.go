package app

import (
	"context"
	"fmt"

	"polyforge/internal/kumi"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx  context.Context
	kumi *kumi.Service
}

func New(service *kumi.Service) *App {
	return &App{kumi: service}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.kumi.SetContext(ctx)

	// One-time preliminary setup (registers the .polypack file type, etc.).
	kumi.CleanupOldBinary()
	if note := kumi.RunFirstRunSetup(); note != "" {
		runtime.LogInfo(ctx, note)
	}
}

// FirstRunNote runs first-run setup if needed and returns a note for the UI.
func (a *App) FirstRunNote() string {
	return kumi.RunFirstRunSetup()
}

// LaunchedPackPath returns a pack file passed on the command line (e.g. from
// double-clicking a .polypack), or "" if the app was launched normally.
func (a *App) LaunchedPackPath() string {
	return kumi.LaunchedPackPath()
}

func (a *App) Shutdown(context.Context) {
	// Placeholder for graceful shutdown hooks.
}

func (a *App) GetMenuOptions() []kumi.OptionDescriptor {
	return a.kumi.Options()
}

// GetRemoteContent returns the remote content manifest (modpacks, option
// overrides, update info) fetched from the website, with disk-cache fallback.
func (a *App) GetRemoteContent() kumi.RemoteContentResult {
	return a.kumi.RemoteContent()
}

// VerifyPackAccess checks a modpack password against the website endpoint.
func (a *App) VerifyPackAccess(packID, password string) kumi.PackAccessResult {
	return a.kumi.VerifyPackAccess(packID, password)
}

func (a *App) Execute(optionID string, payload kumi.ExecutionPayload) (*kumi.ActionResult, error) {
	return a.kumi.Execute(optionID, payload)
}

func (a *App) SelectDirectory(title string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not available")
	}

	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
	if err != nil {
		return "", err
	}

	return path, nil
}

// SelectPackFile opens a file picker for a local .polypack file.
func (a *App) SelectPackFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not available")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a PolyForge pack",
		Filters: []runtime.FileFilter{
			{DisplayName: "PolyForge packs (*.polypack;*.zip)", Pattern: "*.polypack;*.zip"},
		},
	})
}

// InspectPolyPack reads a local pack's manifest for display before install.
func (a *App) InspectPolyPack(path string) (*kumi.PolyPackInfo, error) {
	return kumi.InspectPolyPack(path)
}

func (a *App) CloneModrinthProfile(request kumi.ModrinthCloneRequest) (*kumi.ActionResult, error) {
	return a.kumi.CloneModrinthProfile(request)
}

func (a *App) SearchExecutable(query kumi.ExecutableSearchRequest) (*kumi.ActionResult, error) {
	return a.kumi.SearchExecutable(query)
}

func (a *App) EnumerateApplications() (*kumi.ActionResult, error) {
	return a.kumi.EnumerateApplications()
}
