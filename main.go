package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// Define a struct to hold the parsed command-line arguments
type CliArgs struct {
    InitialFile     string
    AllowInlineHTML bool
    SanitizeHTML    bool
    ShowHelp        bool
    CmdlineOptions  string
}

// GetArgs becomes a standalone function that parses flags and returns them.
func GetArgs() (*CliArgs, error) {
    // Use a new, local FlagSet to avoid conflicts with the global one,
    // which is a best practice for robust CLI parsing.
    fs := flag.NewFlagSet("markdown-reader", flag.ContinueOnError)

    args := &CliArgs{}

    // Define flags on the local FlagSet
    fs.StringVarP(&args.InitialFile, "file", "f", "", "Path to the initial Markdown file")
    fs.BoolVar(&args.AllowInlineHTML, "nohtml", true, "No inline HTML rendering (default: allow inline HTML)")
    fs.BoolVar(&args.SanitizeHTML, "nosanitize", true, "No sanitizing of HTML and URL output (default: sanitize HTML and URLs)")
    // fs.BoolVarP(&args.ShowHelp, "help", "h", false, "Display help message")

    // Configure flags that can be used without a value
    fs.Lookup("nohtml").NoOptDefVal = "false"
    fs.Lookup("nosanitize").NoOptDefVal = "false"

    // Parse the application's arguments (os.Args[1:])
    // ContinueOnError prevents the flagset from exiting the app on its own.
    err := fs.Parse(os.Args[1:])
    if err != nil {
        // If parsing fails for any reason (e.g., bad flag), return the error.
        return nil, err
    }

    // Handle non-flag argument as the initial file
    if args.InitialFile == "" && len(fs.Args()) > 0 {
        args.InitialFile = fs.Args()[0]
    }

    // Build the usage string
    var usageText strings.Builder
	appName := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
    usageText.WriteString(fmt.Sprintf("Usage: %s [options] [filepath]\n\n", appName))
	usageText.WriteString("  filepath             Path to the initial Markdown file\n")
	usageText.WriteString("                         (if not specified with --file)\n\n")
    usageText.WriteString("Options:\n")
	usageText.WriteString("  -f, --file=<path>    Path to the initial Markdown file\n")
	usageText.WriteString("      --nohtml         Disable inline HTML rendering\n")
	usageText.WriteString("                         (default: allow inline HTML)\n")
	usageText.WriteString("      --nosanitize     Disable sanitizing of HTML and URL output\n")
	usageText.WriteString("                         (default: sanitize HTML and URLs)\n")
    args.CmdlineOptions = usageText.String()

    // Assign the usage function to the global flag set for the main function to call.
    flag.Usage = func() {
        fmt.Fprint(os.Stderr, args.CmdlineOptions)
    }

    return args, nil
}

func main() {
    // Handle command-line arguments FIRST ---
    cliArgs, err := GetArgs()
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
