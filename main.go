package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	wd := os.Getenv("WD")
	pkg := os.Getenv("PKG")
	// Create an instance of the app structure
	app := NewApp()
	if pkg != "" {
		app.state.TestParams.Pkg = pkg
	} else {
		app.state.TestParams.Pkg = "./..."
	}
	app.state.CWD = wd

	if wd == "" {
		var err error
		app.state.CWD, err = os.Getwd()
		if err != nil {
			panic("cant get working directory")
		}
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "gogreen",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
