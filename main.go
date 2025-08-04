package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

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

var (
	Version = "dev-build"
	Commit  = "none"
	Date    = "unknown"
)

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

	authorName := gjson.Get(wailsConfig, "author.name").String()
	authorEmail := gjson.Get(wailsConfig, "author.email").String()
	var versionText strings.Builder
	versionText.WriteString("<pre>\n")
	versionText.WriteString("  Application: " + cliArgs.AppName + "\n")
	versionText.WriteString("      Version: " + Version + "\n")
	versionText.WriteString("   Build Date: " + Date + "\n\n")
	versionText.WriteString("Copyright 2025 " + authorName + " <" + authorEmail + ">\n\n")
	versionText.WriteString("Licensed under the Apache License, Version 2.0 (the \"License\");\n")
	versionText.WriteString("you may not use this file except in compliance with the License.\n")
	versionText.WriteString("You may obtain a copy of the License at\n\n")
	versionText.WriteString("    https://www.apache.org/licenses/LICENSE-2.0\n\n")
	versionText.WriteString("Unless required by applicable law or agreed to in writing, software\n")
	versionText.WriteString("distributed under the License is distributed on an \"AS IS\" BASIS,\n")
	versionText.WriteString("WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n")
	versionText.WriteString("See the License for the specific language governing permissions and\n")
	versionText.WriteString("limitations under the License.\n")
	versionText.WriteString("</pre>\n")

	// versionText.WriteString("Commit Hash: " + Commit + "\n")

	app.versionInfo = versionText.String()


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
