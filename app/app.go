package app

import (
	"context"
	"fmt"

	"polyforge/backend/installer"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	installer *installer.Service
}

func New() *App {
	return &App{installer: installer.NewService()}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.installer.SetContext(ctx)
}

func (a *App) Shutdown(ctx context.Context) {
	// stop any outstanding work if needed
}

func (a *App) GetMenuOptions() []installer.OptionDescriptor {
	return a.installer.Options()
}

func (a *App) Execute(optionID string, payload installer.ExecutionPayload) (*installer.ActionResult, error) {
	return a.installer.Execute(optionID, payload)
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

func (a *App) CloneModrinthProfile(request installer.ModrinthCloneRequest) (*installer.ActionResult, error) {
	return a.installer.CloneModrinthProfile(request)
}

func (a *App) SearchExecutable(query installer.ExecutableSearchRequest) (*installer.ActionResult, error) {
	return a.installer.SearchExecutable(query)
}

func (a *App) EnumerateApplications() (*installer.ActionResult, error) {
	return a.installer.EnumerateApplications()
}
