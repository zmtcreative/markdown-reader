package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"text/template"

	"md-reader/internal/cli"

	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed wails.json
var wailsConfig string

//go:embed frontend/src/assets/html/about.gohtml
var aboutTemplate string

//go:embed frontend/src/assets/html/license-short.html
var licenseShort string

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
    // if err != nil {
    //     // Handle parsing error if any
    //     fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
    //     // os.Exit(1)
    // }

    // Create an instance of the app structure
    app := NewApp()

	// Pass the parsed arguments to the app instance
    app.initialFile = cliArgs.InitialFile
    app.allowInlineHTML = cliArgs.AllowInlineHTML
    app.sanitizeHTML = cliArgs.SanitizeHTML
    app.cmdlineOptions = cliArgs.CmdlineOptions
	app.appName = cliArgs.AppName // Store the application name without extension
	app.appNameWithExt = cliArgs.AppNameWithExt // Store the application name with extension
	app.showHelp = cliArgs.ShowHelp // Store the help flag

	app.versionInfo = app.getAbout()

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
		Menu:             app.menu(), // Add the menu here
		Bind: []interface{}{
			app,
		},
	})

	if werr != nil {
		println("Error:", err.Error())
	}
}

func (a *App) getAbout() string {
	var versionText bytes.Buffer

	authorName := gjson.Get(wailsConfig, "author.name").String()
	// authorEmail := gjson.Get(wailsConfig, "author.email").String()
	productName := gjson.Get(wailsConfig, "info.productName").String()

	tplData := struct {
		ProductName string
		AppName  string
		Version  string
		BuildDate string
		Copyright string
		License   string
	}{
		ProductName: productName,
		AppName:  a.appNameWithExt,
		Version:  Version,
		BuildDate: Date,
		Copyright: fmt.Sprintf("Copyright 2025 %s", authorName),
		License:   licenseShort,
	}
	tpl, err := template.New("about").Parse(aboutTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing about template: %v\n", err)
		os.Exit(1)
	}
	err = tpl.Execute(&versionText, tplData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing about template: %v\n", err)
		os.Exit(1)
	}

	return versionText.String()
}
