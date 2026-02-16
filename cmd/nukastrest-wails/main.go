package main

import (
	"embed"
	"log"

	"github.com/tuusuario/NukaStress/wailsapp"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := wailsapp.New()

	err := wails.Run(&options.App{
		Title:     "Nukastrest",
		Width:     1280,
		Height:    820,
		Frameless: true,
		BackgroundColour: &options.RGBA{
			R: 7,
			G: 11,
			B: 7,
			A: 1,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
