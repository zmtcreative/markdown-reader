package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	if app == nil {
		log.Println("Failed to create app instance")
		os.Exit(1)
	} else {

		// Create the application menu
		appMenu := menu.NewMenu()

		fileMenu := appMenu.AddSubmenu("File")
		fileMenu.AddText("Open", keys.CmdOrCtrl("o"), app.OpenFileMenuHandler)
		fileMenu.AddSeparator()
		fileMenu.AddText("Exit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
			runtime.Quit(app.ctx)
		})

		// Create application with options
		err := wails.Run(&options.App{
			Title:  "markdown-reader",
			Width:  1024,
			Height: 768,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
			OnStartup:        app.startup,
			OnDomReady:       app.domReady,
			OnShutdown:       app.shutdown,
			Menu:             appMenu, // Add the menu here
			Bind: []interface{}{
				app,
			},
		})

		if err != nil {
			println("Error:", err.Error())
		}
	}
}
