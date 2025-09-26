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
}

func (a *App) Shutdown(context.Context) {
	// Placeholder for graceful shutdown hooks.
}

func (a *App) GetMenuOptions() []kumi.OptionDescriptor {
	return a.kumi.Options()
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

func (a *App) CloneModrinthProfile(request kumi.ModrinthCloneRequest) (*kumi.ActionResult, error) {
	return a.kumi.CloneModrinthProfile(request)
}

func (a *App) SearchExecutable(query kumi.ExecutableSearchRequest) (*kumi.ActionResult, error) {
	return a.kumi.SearchExecutable(query)
}

func (a *App) EnumerateApplications() (*kumi.ActionResult, error) {
	return a.kumi.EnumerateApplications()
}
