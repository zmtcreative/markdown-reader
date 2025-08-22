package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"text/template"

	"md-reader/internal/app"

	"md-reader/internal/cli"

	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed wails.json
var wailsConfig string

//go:embed frontend/src/assets/html/about.gohtml
var aboutTemplate string

//go:embed frontend/src/assets/html/license-short.html
var licenseShort string

// App struct
type App struct {
    ctx                 context.Context
    currentFile         string
    appProgName         string // Store the application name without extension
    appProgNameWithExt  string // Store the application name with extension
    showHelp            bool   // Flag to indicate if help should be shown
    frontMatter         map[string]string // Store frontmatter data
    cmdlineOptions      string // Store command line options here
    versionInfo         string // Store version information

    // Managers
    themeManager        *app.ThemeManager
    printManager        *app.PrintManager
    fileManager         *app.FileManager
    documentProcessor   *app.DocumentProcessor
    binaryDetector      *app.BinaryDetector
    configManager       *app.ConfigManager
}

// NewApp creates a new App application struct
func NewApp(cliArgs *cli.CliArgs) *App {
    // Initialize configuration manager first
    configManager := app.NewConfigManager()

    // Apply CLI overrides to configuration
    //   (Removed these from CliArgs [2025-08-20])
    //   No need to deal with overrides for now -- leave as is in case we need it in the future
    // configManager.ApplyCliOverrides(cliArgs.AllowInlineHTML, cliArgs.SanitizeHTML, cliArgs.StripH1)

    // Get final configuration after overrides
    // finalConfig := configManager.GetConfig()

    // Handle app name
    appProgNameWithExt := stringFromPtr(cliArgs.AppProgNameWithExt, "md-reader")
    setAboutString := setAbout(appProgNameWithExt)

    return &App{
        frontMatter:        map[string]string{},                                // Initialize an empty map for frontmatter
        currentFile:        stringFromPtr(cliArgs.InitialFile, ""),            // Default to empty, can be set via CLI flag
        appProgName:        stringFromPtr(cliArgs.AppProgName, "md-reader"),   // Store the application name without extension
        appProgNameWithExt: appProgNameWithExt,                                // Store the application name with extension
        showHelp:           boolFromPtr(cliArgs.ShowHelp, false),              // Default to false, can be set via CLI flag
        versionInfo:        setAboutString,                                     // Set version info using the application name with extension
        cmdlineOptions:     stringFromPtr(cliArgs.CmdlineOptions, ""),         // Store the command line options for help display
        configManager:      configManager,                                     // Store the config manager
    }
}

// boolFromPtr safely dereferences a *bool, returning a default value if the pointer is nil.
func boolFromPtr(p *bool, defaultValue bool) bool {
    if p != nil {
        return *p
    }
    return defaultValue
}

func stringFromPtr(p *string, defaultValue string) string {
    if p != nil {
        return *p
    }
    return defaultValue
}

func setAbout(appProgNameWithExt string) string {
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
		AppName:  appProgNameWithExt,
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

// startup is called when the app starts. The context is created
// after the app is created, but before the event loop starts.
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx

    // Initialize managers
    a.themeManager = app.NewThemeManager(ctx)
    a.printManager = app.NewPrintManager(ctx)
    a.binaryDetector = app.NewBinaryDetector()

    // Get the current configuration for document processor
    a.documentProcessor = app.NewDocumentProcessorWithStyle(ctx, a.configManager)
    a.fileManager = app.NewFileManager(ctx, a.binaryDetector, a.documentProcessor)
}

// domReady is called after the frontend loads the DOM.
// This is where we load and display the initial Markdown file if provided via CLI.
func (a *App) domReady(ctx context.Context) {
    if a.currentFile != "" {
        log.Printf("Loading initial file from command line: %s", a.currentFile)
        err := a.documentProcessor.LoadAndDisplayMarkdown(a.currentFile)
        if err != nil {
            log.Printf("Error loading initial Markdown file %q: %v", a.currentFile, err)
            // Emit an error event to the frontend
            runtime.EventsEmit(a.ctx, "error", "Failed to load initial file: "+err.Error())
        }
    } else {
        // Emit a welcome message if no initial file is provided
        welcomeHTML := "<h1>Welcome to Markdown Reader!</h1>" +
            "<p>Open a Markdown file using the <code>File &gt; Open</code> menu option or provide a path via the command line (e.g., <code>./markdown-reader.exe --file path/to/your/file.md</code>).</p>" +
            "<p>This reader supports GitHub Flavored Markdown (GFM).</p>"
        runtime.EventsEmit(a.ctx, "markdown-rendered", "<h2>No file loaded</h2>"+welcomeHTML)
    }
    if a.showHelp {
        runtime.EventsEmit(a.ctx, "show-help", "Command-Line Options", a.cmdlineOptions)
    }
}

// shutdown is called when the app is about to exit.
// Perform any cleanup here if necessary.
func (a *App) shutdown(ctx context.Context) {
    log.Println("Application is shutting down.")
}

func (a *App) menu() *menu.Menu {
    // Create the application menu
    appMenu := menu.NewMenu()

    fileMenu := appMenu.AddSubmenu("File")
    fileMenu.AddText("Open", keys.CmdOrCtrl("o"), func(data *menu.CallbackData) {
        a.fileManager.OpenFileMenuHandler(data, &a.currentFile)
    })
    fileMenu.AddSeparator()
    fileMenu.AddText("Print", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
        a.printManager.PrintContent()
    })
    // fileMenu.AddText("Save as PDF", keys.CmdOrCtrl("e"), func(_ *menu.CallbackData) {
    //     a.printManager.PrintContentToPDF()
    // })
    fileMenu.AddSeparator()
    fileMenu.AddText("Settings", keys.CmdOrCtrl("comma"), func(_ *menu.CallbackData) {
        runtime.EventsEmit(a.ctx, "show-settings")
    })
    fileMenu.AddSeparator()
    fileMenu.AddText("Exit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
        runtime.Quit(a.ctx)
    })

    // --- Add a new Help menu ---
    helpMenu := appMenu.AddSubmenu("Help")
    helpMenu.AddText("Command-Line Options", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
        // Emit an event to the frontend, sending the help text as data.
        runtime.EventsEmit(a.ctx, "show-help", "Command-Line Options", a.cmdlineOptions)
    })
    helpMenu.AddSeparator()
    helpMenu.AddText("About", keys.CmdOrCtrl("a"), func(_ *menu.CallbackData) {
        // Emit an event to the frontend, sending the version information as data.
        runtime.EventsEmit(a.ctx, "show-help", "About", a.versionInfo)
    })

    return appMenu
}

// Delegate methods to managers

// PrintContent prints the current HTML content
func (a *App) PrintContent() error {
    return a.printManager.PrintContent()
}

// PrintContentToPDF exports the current content to PDF (Windows-specific)
func (a *App) PrintContentToPDF() error {
    return a.printManager.PrintContentToPDF()
}

// GetTheme returns the current theme
func (a *App) GetTheme() string {
    return a.themeManager.GetTheme()
}

// SetTheme sets the theme and emits an event to the frontend
func (a *App) SetTheme(theme string) {
    a.themeManager.SetTheme(theme)
}

// LoadAndDisplayMarkdown loads and displays a markdown file
func (a *App) LoadAndDisplayMarkdown(filePath string) error {
    err := a.documentProcessor.LoadAndDisplayMarkdown(filePath)
    if err == nil {
        a.currentFile = filePath
    }
    return err
}

// ReloadCurrentDocument reloads and regenerates the currently opened markdown document
func (a *App) ReloadCurrentDocument() error {
    if a.currentFile == "" {
        return fmt.Errorf("no document currently loaded")
    }

    // Reload the current file using the updated settings
    err := a.documentProcessor.LoadAndDisplayMarkdown(a.currentFile)
    if err != nil {
        return fmt.Errorf("failed to reload document %s: %w", a.currentFile, err)
    }

    return nil
}

// TODO: CLEANUP - These wrapper methods are never called from the frontend
// The frontend uses its own JavaScript functions instead
//
// Probably don't need these any longer, but leave them commented out for now.
//
// // AddDocClass adds the class to html and body elements
// func (a *App) AddDocClass(thisClass ...string) {
//     a.documentProcessor.AddDocClass(thisClass...)
// }
//
// // RemoveDocClass removes the class from html and body elements
// func (a *App) RemoveDocClass(thisClass ...string) {
//     a.documentProcessor.RemoveDocClass(thisClass...)
// }
//
// // ToggleDocClass toggles the class on html and body elements
// func (a *App) ToggleDocClass(thisClass ...string) {
//     a.documentProcessor.ToggleDocClass(thisClass...)
// }
//
// // OpenFileMenuHandler handles the File > Open menu action
// func (a *App) OpenFileMenuHandler(data *menu.CallbackData) {
//     a.fileManager.OpenFileMenuHandler(data, &a.currentFile)
// }

// Settings-related methods

// GetSettings returns the current application settings
func (a *App) GetSettings() *app.Config {
    return a.configManager.GetConfig()
}

// GetAlertCalloutStyles returns the available alert callout styles
func (a *App) GetAlertCalloutStyles() map[string]string {
    return app.AlertCalloutStyles
}

// SaveSettings saves the provided settings configuration
func (a *App) SaveSettings(settings *app.Config) error {
    // Validate the alert callout style
    settings.AlertCallouts.AlertCalloutStyle = a.configManager.ValidateAlertCalloutStyle(settings.AlertCallouts.AlertCalloutStyle)

    // Update the configuration
    a.configManager.SetConfig(settings)

    // Save to file
    if err := a.configManager.SaveConfig(); err != nil {
        return err
    }

    // Update the current app settings

    // Recreate the document processor with new settings
    a.documentProcessor = app.NewDocumentProcessorWithStyle(a.ctx, a.configManager)

    // Update the file manager with the new document processor
    a.fileManager = app.NewFileManager(a.ctx, a.binaryDetector, a.documentProcessor)

    // Automatically reload the current document to apply new settings
    if a.currentFile != "" {
        if reloadErr := a.ReloadCurrentDocument(); reloadErr != nil {
            log.Printf("Warning: Failed to reload document after settings change: %v", reloadErr)
            // Don't return the reload error, as settings were successfully saved
        }
    }

    return nil
}

// SaveSettingsSessionOnly applies the provided settings to the current session without saving to disk
func (a *App) SaveSettingsSessionOnly(settings *app.Config) error {
    // Validate the alert callout style
    settings.AlertCallouts.AlertCalloutStyle = a.configManager.ValidateAlertCalloutStyle(settings.AlertCallouts.AlertCalloutStyle)

    // Update the configuration in memory only (do NOT call SaveConfig)
    a.configManager.SetConfig(settings)

    // Update the current app settings

    // Recreate the document processor with new settings
    a.documentProcessor = app.NewDocumentProcessorWithStyle(a.ctx, a.configManager)

    // Update the file manager with the new document processor
    a.fileManager = app.NewFileManager(a.ctx, a.binaryDetector, a.documentProcessor)

    // Automatically reload the current document to apply new settings
    if a.currentFile != "" {
        if reloadErr := a.ReloadCurrentDocument(); reloadErr != nil {
            log.Printf("Warning: Failed to reload document after settings change: %v", reloadErr)
            // Don't return the reload error, as settings were successfully applied
        }
    }

    return nil
}
