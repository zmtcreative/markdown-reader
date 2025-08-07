package main

import (
	"context"
	_ "embed"
	"log"

	"md-reader/internal/app"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
    ctx               context.Context
    currentFile       string
    appName           string // Store the application name without extension
    appNameWithExt    string // Store the application name with extension
    stripH1           bool
    allowInlineHTML   bool
    showHelp          bool   // Flag to indicate if help should be shown
    cmdlineOptions    string // Store command line options here
    versionInfo       string // Store version information
    sanitizeHTML      bool   // Flag to control sanitization of HTML and URL links
    frontMatter       map[string]string // Store frontmatter data here

    // Managers
    themeManager      *app.ThemeManager
    printManager      *app.PrintManager
    fileManager       *app.FileManager
    documentProcessor *app.DocumentProcessor
    binaryDetector    *app.BinaryDetector
}

// NewApp creates a new App application struct
func NewApp() *App {
    return &App{
        frontMatter:     map[string]string{},
        stripH1:         true,
        allowInlineHTML: true, // Default to true, can be set via CLI flag
        sanitizeHTML:    true, // Default to true, can be set via CLI flag
        showHelp:        false, // Default to false, can be set via CLI flag
    }
}

// startup is called when the app starts. The context is created
// after the app is created, but before the event loop starts.
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx

    // Initialize managers
    a.themeManager = app.NewThemeManager(ctx)
    a.printManager = app.NewPrintManager(ctx)
    a.binaryDetector = app.NewBinaryDetector()
    a.documentProcessor = app.NewDocumentProcessor(ctx, a.stripH1, a.allowInlineHTML, a.sanitizeHTML)
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

// AddDocClass adds the class to html and body elements
func (a *App) AddDocClass(thisClass ...string) {
    a.documentProcessor.AddDocClass(thisClass...)
}

// RemoveDocClass removes the class from html and body elements
func (a *App) RemoveDocClass(thisClass ...string) {
    a.documentProcessor.RemoveDocClass(thisClass...)
}

// ToggleDocClass toggles the class on html and body elements
func (a *App) ToggleDocClass(thisClass ...string) {
    a.documentProcessor.ToggleDocClass(thisClass...)
}

// OpenFileMenuHandler handles the File > Open menu action
func (a *App) OpenFileMenuHandler(data *menu.CallbackData) {
    a.fileManager.OpenFileMenuHandler(data, &a.currentFile)
}
