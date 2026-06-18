package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "IAD Console",
		Width:     1440,
		Height:    900,
		MinWidth:  1200,
		MinHeight: 720,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Near-black background to match the dark-first instrument theme and
		// avoid a white flash on launch.
		BackgroundColour: &options.RGBA{R: 14, G: 14, B: 16, A: 255},
		OnStartup: func(ctx context.Context) { app.startup(ctx) },
		Bind:      []interface{}{app},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
