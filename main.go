package main

import (
	"embed"
	"log"

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

	// Create the application menu
	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("File")

	fileMenu.AddText("Open", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		// Open a file dialog
		filePath, err := runtime.OpenFileDialog(app.ctx, runtime.OpenDialogOptions{
			Title: "Select Markdown File",
			Filters: []runtime.FileFilter{
				{
					DisplayName: "Markdown Files (*.md, *.markdown)",
					Pattern:     "*.md;*.markdown",
				},
			},
		})
		if err != nil {
			log.Println("Error opening file dialog:", err)
			return
		}

		// If a file was selected, process it
		if filePath != "" {
			html, err := app.ProcessMarkdown(filePath)
			if err != nil {
				log.Println("Error processing markdown:", err)
				// Optionally, show an error dialog to the user
				runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "Error",
					Message: "Failed to render markdown file: " + err.Error(),
				})
				return
			}
			// Emit an event to the frontend with the rendered HTML
			runtime.EventsEmit(app.ctx, "markdown-rendered", html)
		}
	})

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
		Menu:             appMenu, // Add the menu here
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
