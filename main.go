package main

import (
	"embed"
	"fmt"
	"os"

	"markdown-reader/pkg/cli"

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
    // Handle command-line arguments FIRST ---
    cliArgs, err := cli.GetArgs()
    if err != nil {
        // Handle parsing error if any
        fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
        os.Exit(1)
    }

    // Create an instance of the app structure
    app := NewApp()
    // Pass the parsed arguments to the app instance
    app.initialFile = cliArgs.InitialFile
    app.allowInlineHTML = cliArgs.AllowInlineHTML
    app.sanitizeHTML = cliArgs.SanitizeHTML
    app.cmdlineOptions = cliArgs.CmdlineOptions

	// Create the application menu
	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open", keys.CmdOrCtrl("o"), app.OpenFileMenuHandler)
    fileMenu.AddSeparator()
    fileMenu.AddText("Print", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
        app.PrintContent()
    })
    // fileMenu.AddText("Save as PDF", keys.CmdOrCtrl("e"), func(_ *menu.CallbackData) {
    //     app.PrintContentToPDF()
    // })
	fileMenu.AddSeparator()
	fileMenu.AddText("Exit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

    // --- Add a new Help menu ---
    helpMenu := appMenu.AddSubmenu("Help")
    helpMenu.AddText("Command-Line Options", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
        // Emit an event to the frontend, sending the help text as data.
        runtime.EventsEmit(app.ctx, "show-help", app.cmdlineOptions)
    })

	// Create application with options
	werr := wails.Run(&options.App{
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

	if werr != nil {
		println("Error:", err.Error())
	}
}
