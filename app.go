package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	dateparse "github.com/araddon/dateparse"

	"markdown-reader/pkg/markdown"
	"markdown-reader/pkg/util"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
)

// App struct
type App struct {
    ctx context.Context
    initialFile string
    stripH1 bool
	allowInlineHTML bool
	theme string // Store display mode (e.g., "light", "dark")
	docTypes []string // List of document types (e.g., "techdoc", "mydoc")
	cmdlineOptions string // Store command line options here
	versionInfo string // Store version information
	sanitizeHTML bool // Flag to control sanitization of HTML and URL links
    frontMatter map[string]string // Store frontmatter data here
    mdConverter goldmark.Markdown
}

// NewApp creates a new App application struct
func NewApp() *App {
    app := &App{
        frontMatter: map[string]string{},
        stripH1: true,
		allowInlineHTML: true, // Default to true, can be set via CLI flag
		sanitizeHTML: true, // Default to true, can be set via CLI flag
		theme: "light",
    }
	return app
}

// startup is called when the app starts. The context is created
// after the app is created, but before the event loop starts.
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
	a.mdConverter = markdown.CreateGoldmarkInstance(a.allowInlineHTML, a.sanitizeHTML)
}

// domReady is called after the frontend loads the DOM.
// This is where we load and display the initial Markdown file if provided via CLI.
func (a *App) domReady(ctx context.Context) {
	if a.initialFile != "" {
		log.Printf("Loading initial file from command line: %s", a.initialFile)
		err := a.LoadAndDisplayMarkdown(a.initialFile)
		if err != nil {
			log.Printf("Error loading initial Markdown file %q: %v", a.initialFile, err)
			// Emit an error event to the frontend
			runtime.EventsEmit(a.ctx, "error", "Failed to load initial file: "+err.Error())
		}
	} else {
		// Emit a welcome message if no initial file is provided
		welcomeHTML := "<h1>Welcome to Markdown Reader!</h1>" +
			"<p>Open a Markdown file using the <code>File &gt; Open</code> menu option or provide a path via the command line (e.g., <code>./markdown-reader.exe --file path/to/your/file.md</code>).</p>" +
			"<p>This reader supports GitHub Flavored Markdown (GFM).</p>"
		runtime.EventsEmit(a.ctx, "markdown-rendered", "<h2>No file loaded</h2>" + welcomeHTML)
	}
}

// shutdown is called when the app is about to exit.
// Perform any cleanup here if necessary.
func (a *App) shutdown(ctx context.Context) {
    log.Println("Application is shutting down.")
}

// PrintContent prints the current HTML content
func (a *App) PrintContent() error {
    // Emit event to frontend to trigger print
    runtime.EventsEmit(a.ctx, "print-content")
    return nil
}

// PrintContentToPDF exports the current content to PDF (Windows-specific)
func (a *App) PrintContentToPDF() error {
    // Get a save file path for the PDF
    filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
        Title: "Save as PDF",
        DefaultFilename: "document.pdf",
        Filters: []runtime.FileFilter{
            {DisplayName: "PDF Files (*.pdf)", Pattern: "*.pdf"},
        },
    })
    if err != nil {
        return err
    }

    if filePath == "" {
        return nil // User cancelled
    }

    // Emit event to frontend to save as PDF
    runtime.EventsEmit(a.ctx, "save-as-pdf", filePath)
    return nil
}

func (a *App) GetTheme() string {
	// Return the current theme (light or dark)
	return a.theme
}

// SetTheme sets the theme and emits an event to the frontend.
func (a *App) SetTheme(theme string) {
    a.theme = theme
    // Emit an event to notify the frontend of the change
    runtime.EventsEmit(a.ctx, "theme:changed", theme)
}

func (a *App) OpenFileMenuHandler(_ *menu.CallbackData) {
	log.Println("File -> Open menu item clicked. Opening file dialog...")

	// Open a file dialog to allow the user to select a Markdown file.
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md;*.markdown)", Pattern: "*.md;*.markdown"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "The user cancelled the dialog") || strings.Contains(err.Error(), "canceled") {
			log.Println("File dialog cancelled by user.")
			return
		}
		log.Printf("Error opening file dialog: %v", err)
		runtime.EventsEmit(a.ctx, "error", "Failed to open file dialog: "+err.Error())
		return
	}

	if selection == "" {
		log.Println("No file selected in dialog.")
		runtime.EventsEmit(a.ctx, "error", "No file was selected.")
		return
	}

	log.Printf("User selected file: %s", selection)
	err = a.LoadAndDisplayMarkdown(selection)
	if err != nil {
		log.Printf("Error loading selected Markdown file %q: %v", selection, err)
		runtime.EventsEmit(a.ctx, "error", "Failed to load selected file: "+err.Error())
	} else {
		log.Printf("Successfully loaded Markdown file: %s", selection)
		a.initialFile = selection // Update initialFile to the newly opened file
	}
}

// LoadAndDisplayMarkdown reads a Markdown file from the given path,
// converts its content to HTML using Goldmark, and then emits the HTML
// to the frontend via the "markdownLoaded" event.
func (a *App) LoadAndDisplayMarkdown(filePath string) error {
	mdContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath) // Corrected: Use fmt.Errorf
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied to read file: %s", filePath) // Corrected: Use fmt.Errorf
		}
		return fmt.Errorf("failed to read file %s: %w", filePath, err) // Corrected: Use fmt.Errorf
	}

	// Normalize line endings to Unix-style (LF)
	// Some extensions (e.g., goldmark-gh-alerts) rely on Unix-style line endings
	mdContent = []byte(strings.ReplaceAll(string(mdContent), "\r\n", "\n"))

	// Extract the document title from the H1 heading element if present
	// os.WriteFile("tmp/debug-before-extract-h1.md", mdContent, 0644) // Debugging: Write Markdown to file
    var thisDocumentTitle string
    if a.stripH1 {
		thisDocumentTitle, mdContent, _ = markdown.ExtractH1(string(mdContent))
    }
	// os.WriteFile("tmp/debug-after-extract-h1.md", mdContent, 0644) // Debugging: Write Markdown to file
	// thisDocumentTitle := ""
	// if err != nil {
	// 	return fmt.Errorf("failed to extract document title: %w", err) // Corrected: Use fmt.Errorf
	// }

	// Convert Markdown content to HTML
	htmlContent, docFrontmatter, err := markdown.ConvertMarkdownToHTML(a.mdConverter, mdContent)
	if err != nil {
		return fmt.Errorf("failed to convert Markdown to HTML: %w", err) // Corrected: Use fmt.Errorf
	}

	// Emit the converted HTML to the frontend.
	var docTitle, docDate, docType, tmpDocTitle, tmpDocDate string
	docFileTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	timeLayout := time.DateTime + " MST"
	docDateLM := ""
	docDateDD := ""

	if fileInfo, err := os.Stat(filePath); err == nil {
		fileModDate := fileInfo.ModTime()
		tz := fileModDate.Location()
		fmtLastModified := `<span class="date-label last-modified">Last Modified:</span> <span class="date-value last-modified">%s</span>`
		docDateLM = fmt.Sprintf(fmtLastModified, fileModDate.In(tz).Format(timeLayout))
	} else {
		docDate = ""
	}
	if docFrontmatter != nil {
		tmpDocTitle = util.GetValueFromMap(docFrontmatter, "Title")
		tmpDocDate = util.GetValueFromMap(docFrontmatter, "Date")
		docType = strings.ToLower(util.GetValueFromMap(docFrontmatter, "Type"))
	}

	if tmpDocTitle != "" {
		docTitle = tmpDocTitle
	} else if thisDocumentTitle != "" {
		docTitle = thisDocumentTitle
	} else {
		docTitle = fmt.Sprintf("File: %s", docFileTitle)
	}

	if tmpDocDate != "" {
		fmtDocDate := `<span class="date-label document-date">Document Date:</span> <span class="date-value document-date">%s</span>`
		tz := time.Now().Local().Location()
		dateString, err := dateparse.ParseIn(tmpDocDate, tz)
		if err == nil {
			docDateDD = fmt.Sprintf(fmtDocDate, dateString.Format(timeLayout))
		} else {
			docDateDD = fmt.Sprintf(fmtDocDate, tmpDocDate)
		}
	}

	if docDateDD != "" {
		docDate = docDateDD
	}
	if docDateLM != "" {
		if docDate == "" {
			docDate = docDateLM
		} else {
			docDate = docDate + "<br>" + docDateLM
		}
	}

	// os.WriteFile("tmp/debug-before.html", htmlContent, 0644) // Debugging: Write HTML to file

	// Cleanup HTML content by adjusting line breaks and removing unnecessary tags
	// This is necessary to ensure proper rendering in the frontend since some Markdown renderers
	// may produce inconsistent HTML output (this is a side effect of using some packages
	// like Highlighting/Chroma).
	htmlContent = markdown.CleanupHTMLContent(htmlContent)

	// os.WriteFile("tmp/debug-after.html", htmlContent, 0644) // Debugging: Write HTML to file

	runtime.EventsEmit(a.ctx, "markdown-rendered", string(htmlContent), docTitle, docDate)

	// if docType == "techdoc" {
	// 	a.AddDocClass(docType)
	// } else {
	// 	a.RemoveDocClass(docType)
	// }

	if a.docTypes != nil {
		for _, dt := range a.docTypes {
			a.RemoveDocClass(dt)
		}
	}
	if docType != "" {
		docTypeArray := strings.Fields(docType)
		for _, dt := range docTypeArray {
			a.AddDocClass(dt)
			a.docTypes = append(a.docTypes, dt)
		}
	}

	return nil
}

// AddDocClass adds the class to html and body elements
func (a *App) AddDocClass(thisClass ...string) {
    runtime.EventsEmit(a.ctx, "add-doc-class", thisClass)
}

// RemoveDocClass removes the class from html and body elements
func (a *App) RemoveDocClass(thisClass ...string) {
    runtime.EventsEmit(a.ctx, "remove-doc-class", thisClass)
}

// ToggleDocClass toggles the class on html and body elements
func (a *App) ToggleDocClass(thisClass ...string) {
    runtime.EventsEmit(a.ctx, "toggle-doc-class", thisClass)
}

