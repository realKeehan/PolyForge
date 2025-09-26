package main

import (
	"log"

	"polyforge"
	"polyforge/internal/app"
	"polyforge/internal/kumi"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	kumiService := kumi.NewService()
	application := app.New(kumiService)

	err := wails.Run(&options.App{
		Title:       "PolyForge Installer",
		Width:       1100,
		Height:      720,
		AssetServer: &assetserver.Options{Assets: polyforge.Assets()},
		OnStartup:   application.Startup,
		OnShutdown:  application.Shutdown,
		Bind:        app.Bindings(application),
		LogLevel:    logger.INFO,
	})
	if err != nil {
		log.Fatal(err)
	}
}
