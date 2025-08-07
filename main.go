package main

import (
	"embed"

	"md-reader/internal/cli"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	Version = "dev-build"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
    // Handle command-line arguments FIRST ---
    cliArgs, _ := cli.GetArgs()

	// We ignore the error from GetArgs, since the app will start with default values
	// if there are no command-line arguments or if parsing fails.

    // Create an instance of the mdrApp structure, passing the parsed CLI arguments
	// This will initialize the application with the command-line arguments.
    mdrApp := NewApp(cliArgs)

	// Create application with options
	werr := wails.Run(&options.App{
		Title:  "Markdown Reader",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        mdrApp.startup,
		OnDomReady:       mdrApp.domReady,
		OnShutdown:       mdrApp.shutdown,
		Menu:             mdrApp.menu(), // Add the menu here
		Bind: []interface{}{
			mdrApp,
		},
	})

	if werr != nil {
		println("Error:", werr.Error())
	}
}

