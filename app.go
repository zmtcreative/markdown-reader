package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	dateparse "github.com/araddon/dateparse"
	encodingUnicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"md-reader/internal/markdown"
	mdrutils "md-reader/internal/utils"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
)

// App struct
type App struct {
    ctx context.Context
    currentFile string
	appName string // Store the application name without extension
	appNameWithExt string // Store the application name with extension
    stripH1 bool
	allowInlineHTML bool
	theme string // Store display mode (e.g., "light", "dark")
	docTypes []string // List of document types (e.g., "techdoc", "mydoc")
	showHelp bool // Flag to indicate if help should be shown
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
		showHelp: false, // Default to false, can be set via CLI flag
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
	if a.currentFile != "" {
		log.Printf("Loading initial file from command line: %s", a.currentFile)
		err := a.LoadAndDisplayMarkdown(a.currentFile)
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
		runtime.EventsEmit(a.ctx, "markdown-rendered", "<h2>No file loaded</h2>" + welcomeHTML)
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
	fileMenu.AddText("Open", keys.CmdOrCtrl("o"), a.OpenFileMenuHandler)
    fileMenu.AddSeparator()
    fileMenu.AddText("Print", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
        a.PrintContent()
    })
    // fileMenu.AddText("Save as PDF", keys.CmdOrCtrl("e"), func(_ *menu.CallbackData) {
    //     app.PrintContentToPDF()
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

    if selection != "" {
        log.Printf("User selected file: %s", selection)

        // Check if the selected file is binary
        isBinary, err := a.isBinaryFile(selection)
        if err != nil {
            log.Printf("Error checking if file is binary %q: %v", selection, err)
            runtime.EventsEmit(a.ctx, "error", "Failed to read selected file: "+err.Error())
            return
        }

        if isBinary {
            log.Printf("User selected a binary file: %s", selection)
            runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Cannot open binary file: %s\n\nPlease select a text-based Markdown file (.md or .markdown).", filepath.Base(selection)))
            return
        }

        err = a.LoadAndDisplayMarkdown(selection)
        if err != nil {
            log.Printf("Error loading selected Markdown file %q: %v", selection, err)
            runtime.EventsEmit(a.ctx, "error", "Failed to load selected file: "+err.Error())
        } else {
            log.Printf("Successfully loaded Markdown file: %s", selection)
            a.currentFile = selection // Update currentFile to the newly opened file
        }
    } else {
        log.Println("No file selected. User cancelled the dialog.")
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

    // Detect and handle UTF-16 BOMs, and convert to UTF-8 if necessary
    if len(mdContent) >= 2 {
        bom := mdContent[:2]
        var transformer transform.Transformer
        if bom[0] == 0xFE && bom[1] == 0xFF { // UTF-16 BE
            transformer = encodingUnicode.UTF16(encodingUnicode.BigEndian, encodingUnicode.IgnoreBOM).NewDecoder()
        } else if bom[0] == 0xFF && bom[1] == 0xFE { // UTF-16 LE
            transformer = encodingUnicode.UTF16(encodingUnicode.LittleEndian, encodingUnicode.IgnoreBOM).NewDecoder()
        }

        if transformer != nil {
            utf8Content, _, err := transform.Bytes(transformer, mdContent)
            if err == nil {
                mdContent = utf8Content
            } else {
                log.Printf("Warning: Failed to convert from UTF-16 to UTF-8: %v", err)
            }
        }
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
		tmpDocTitle = mdrutils.GetValueFromMap(docFrontmatter, "Title")
		tmpDocDate = mdrutils.GetValueFromMap(docFrontmatter, "Date")
		docType = strings.ToLower(mdrutils.GetValueFromMap(docFrontmatter, "Type"))
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


// isBinaryFile checks if a file is binary by reading the first 8192 bytes
// and using multiple detection methods including UTF-8 validation
func (a *App) isBinaryFile(filePath string) (bool, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return false, err
    }
    defer file.Close()

    // Read first 8KB to check for binary content (increased from 512 bytes)
    buffer := make([]byte, 8192)
    n, err := file.Read(buffer)
    if err != nil && err.Error() != "EOF" {
        return false, err
    }

    if n == 0 {
        return false, nil // Empty file is considered text
    }

    // Trim buffer to actual read size
    buffer = buffer[:n]

    // Check for UTF-16/UTF-32 BOMs first. If a BOM is present, it's a text file.
    if (n >= 2 && ((buffer[0] == 0xFE && buffer[1] == 0xFF) || (buffer[0] == 0xFF && buffer[1] == 0xFE))) ||
        (n >= 4 && ((buffer[0] == 0x00 && buffer[1] == 0x00 && buffer[2] == 0xFE && buffer[3] == 0xFF) ||
            (buffer[0] == 0xFF && buffer[1] == 0xFE && buffer[2] == 0x00 && buffer[3] == 0x00))) {
        return false, nil
    }

    // Check for null bytes (but allow some exceptions for UTF-16/UTF-32)
    nullCount := 0
    for i := 0; i < n; i++ {
        if buffer[i] == 0 {
            nullCount++
        }
    }

    // If more than 1% null bytes, likely binary (unless it's UTF-16/UTF-32)
    if float64(nullCount)/float64(n) > 0.01 {
        // Check if it might be UTF-16 or UTF-32 by looking for BOM or patterns
        if !a.isLikelyUTF16or32(buffer) {
            return true, nil
        }
    }

    // Check if the content is valid UTF-8
    if !a.isValidUTF8(buffer) {
        // If not valid UTF-8, check if it might be other text encodings
        if !a.isLikelyTextEncoding(buffer) {
            return true, nil
        }
    }

    // Check for high percentage of control characters (excluding common whitespace)
    controlChars := 0
    for i := 0; i < n; i++ {
        b := buffer[i]
        // Count control characters but exclude common text characters:
        // 9 (tab), 10 (LF), 13 (CR), and anything >= 32 (printable ASCII)
        if b < 32 && b != 9 && b != 10 && b != 13 {
            controlChars++
        }
    }

    // If more than 5% control characters, likely binary
    if n > 0 && float64(controlChars)/float64(n) > 0.05 {
        return true, nil
    }

    // Check for common binary file signatures/magic numbers
    if a.hasBinarySignature(buffer) {
        return true, nil
    }

    return false, nil
}

// isValidUTF8 checks if the buffer contains valid UTF-8 text
func (a *App) isValidUTF8(buffer []byte) bool {
    // Check if the entire buffer is valid UTF-8
    return utf8.Valid(buffer)
}

// isLikelyUTF16or32 checks for UTF-16 or UTF-32 patterns
func (a *App) isLikelyUTF16or32(buffer []byte) bool {
    if len(buffer) < 4 {
        return false
    }

    // Check for UTF-16 BOM (Byte Order Mark)
    if (buffer[0] == 0xFF && buffer[1] == 0xFE) || // UTF-16 LE
        (buffer[0] == 0xFE && buffer[1] == 0xFF) { // UTF-16 BE
        return true
    }

    // Check for UTF-32 BOM
    if len(buffer) >= 4 {
        if (buffer[0] == 0xFF && buffer[1] == 0xFE && buffer[2] == 0x00 && buffer[3] == 0x00) || // UTF-32 LE
            (buffer[0] == 0x00 && buffer[1] == 0x00 && buffer[2] == 0xFE && buffer[3] == 0xFF) { // UTF-32 BE
            return true
        }
    }

    // Look for UTF-16 patterns (every other byte might be null for ASCII in UTF-16)
    if len(buffer) >= 100 { // Need reasonable sample size
        evenNulls, oddNulls := 0, 0
        for i := 0; i < len(buffer)-1; i += 2 {
            if buffer[i] == 0 {
                evenNulls++
            }
            if buffer[i+1] == 0 {
                oddNulls++
            }
        }

        // If predominantly even or odd positioned nulls, might be UTF-16
        total := len(buffer) / 2
        if total > 0 {
            evenRatio := float64(evenNulls) / float64(total)
            oddRatio := float64(oddNulls) / float64(total)
            // Lowered threshold to 0.4 to catch more cases of mixed ASCII/non-ASCII
            if evenRatio > 0.4 || oddRatio > 0.4 {
                return true
            }
        }
    }

    return false
}

// isLikelyTextEncoding checks for other common text encodings
func (a *App) isLikelyTextEncoding(buffer []byte) bool {
    // Check for common text file patterns even if not UTF-8
    textIndicators := 0
    totalChars := len(buffer)

    if totalChars == 0 {
        return true
    }

    for _, b := range buffer {
        // Count characters that are likely to appear in text files
        if (b >= 32 && b <= 126) || // Printable ASCII
            b == 9 || b == 10 || b == 13 || // Tab, LF, CR
            (b >= 128) { // Extended ASCII/Latin-1
            textIndicators++
        }
    }

    // If more than 70% of characters look like text, consider it text
    return float64(textIndicators)/float64(totalChars) > 0.70
}

// hasBinarySignature checks for common binary file magic numbers/signatures
func (a *App) hasBinarySignature(buffer []byte) bool {
    if len(buffer) < 4 {
        return false
    }

    // Common binary file signatures
    signatures := [][]byte{
        {0x89, 0x50, 0x4E, 0x47}, // PNG
        {0xFF, 0xD8, 0xFF},       // JPEG
        {0x47, 0x49, 0x46},       // GIF
        {0x25, 0x50, 0x44, 0x46}, // PDF
        {0x50, 0x4B, 0x03, 0x04}, // ZIP/DOCX/etc
        {0x50, 0x4B, 0x05, 0x06}, // ZIP (empty)
        {0x50, 0x4B, 0x07, 0x08}, // ZIP (spanned)
        {0x52, 0x61, 0x72, 0x21}, // RAR
        {0x7F, 0x45, 0x4C, 0x46}, // ELF (Linux executable)
        {0x4D, 0x5A},             // Windows PE executable
        {0xCA, 0xFE, 0xBA, 0xBE}, // Java class file
        {0xFE, 0xED, 0xFA},       // Mach-O binary (macOS)
        {0x89, 0x48, 0x44, 0x46}, // HDF5
    }

    for _, sig := range signatures {
        if len(buffer) >= len(sig) {
            match := true
            for i, b := range sig {
                if buffer[i] != b {
                    match = false
                    break
                }
            }
            if match {
                return true
            }
        }
    }

    return false
}
