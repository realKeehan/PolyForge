package main

import (
	"log"
	"polyforge/internal/app"
	"polyforge/internal/kumi"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func main() {
	kumiService := kumi.NewService()
	application := app.New(kumiService)
	err := wails.Run(&options.App{
		Title:            "PolyForge Installer",
		Width:            750,
		Height:           485,
		MinWidth:         750,
		MinHeight:        485,
		MaxWidth:         750,
		MaxHeight:        485,
		DisableResize:    true,
		Frameless:        true,
		BackgroundColour: options.NewRGBA(0, 0, 0, 0),
		Windows: &windows.Options{
			WindowIsTranslucent:  true,
			WebviewIsTransparent: true,
			DisableWindowIcon:    true,
		},
		AssetServer: &assetserver.Options{Assets: assetsFS()},
		OnStartup:   application.Startup,
		OnShutdown:  application.Shutdown,
		Bind:        app.Bindings(application),
		LogLevel:    logger.INFO,
	})
	if err != nil {
		log.Fatal(err)
	}
}
