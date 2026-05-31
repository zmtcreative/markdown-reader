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

	// Unknown CLI flags are intentionally ignored so the GUI app can tolerate
	// wrapper- or launcher-added arguments while still honoring --file/-f.

    // Create an instance of the mdrApp structure, passing the parsed CLI arguments
	// This will initialize the application with the command-line arguments.
    mdrApp := NewApp(cliArgs)

	// fmt.Printf("##> DEBUG (MAIN): App initialized with CLI args\n")
	// fmt.Printf("####> Initial file: %s\n", mdrApp.currentFile)

	// TODO: Add options to NewApp and Config to set Width and Height and validate the sizes
	//       - Update the NewApp function to accept width and height parameters
	//       - Update the Config struct to include width and height fields
	//       - Update Settings.vue to bind the width and height fields to the UI
	//       - Set minimum width and height values (640, 480)
	//       - Use Wails ScreenGetAll(ctx context.Context) to get the current screen dimensions
	//         - See https://wails.io/docs/reference/runtime/screen/
	//         - Make sure Height and/or Width do not exceed the screen dimensions

	// Create application with options
	werr := wails.Run(&options.App{
		Title:  "Markdown Reader",
		MinWidth: 616,
		MinHeight: 539,
		// MaxWidth: 1296,
		// MaxHeight: 1020,
		Width:  1040,
		Height: 984,
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

