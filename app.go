package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	dateparse "github.com/araddon/dateparse"
	flag "github.com/spf13/pflag"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

// App struct
type App struct {
    ctx context.Context
    initialFile string
    stripH1 bool
	allowInlineHTML bool
	cmdlineOptions string // Store command line options here
	sanitizeHTML bool // Flag to control sanitization of HTML and URL links
    frontMatter map[string]string // Store frontmatter data here
    mdConverter goldmark.Markdown
}

// NewApp creates a new App application struct
func NewApp() *App {
    app := &App{
        frontMatter: map[string]string{},
        stripH1: false,
		allowInlineHTML: true, // Default to true, can be set via CLI flag
		sanitizeHTML: true, // Default to true, can be set via CLI flag
    }
	if err := app.GetArgs(); err != nil {
		log.Printf("Error processing command line arguments: %v", err)
		return nil // Return nil if there was an error processing command line arguments
	}
	app.mdConverter = app.CreateGoldmarkInstance()
	return app
}

// startup is called when the app starts. The context is created
// after the app is created, but before the event loop starts.
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
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


func (a *App) GetArgs() (err error) {
    // Using flag.Parsed() prevents re-parsing, which can cause panics.
    if !flag.Parsed() {
        // This logic seems to be for finding a single file from command-line args.
        initialFile := flag.StringP("file", "f", "", "Path to the initial Markdown file")
		allowInlineHTML := flag.Bool("nohtml", true, "No inline HTML rendering (default: allow inline HTML)")
		sanitizeHTML := flag.Bool("nosanitize", true, "No sanitizing of HTML and URL output (default: sanitize HTML and URLs)")
		// showHelp := flag.BoolP("help", "h", true, "Display help message")
		flag.Lookup("nohtml").NoOptDefVal = "false" // Allow --nohtml to be used without a value
		flag.Lookup("nosanitize").NoOptDefVal = "false" // Allow --nosanitize to be used without a value
		// flag.Lookup("help").NoOptDefVal = "false" // Allow --help to be used without a value
        flag.Parse()
        if *initialFile == "" && len(flag.Args()) > 0 {
            a.initialFile = string(flag.Args()[0])
        }
        if *initialFile != "" {
            a.initialFile = string(*initialFile)
        }
		a.allowInlineHTML = *allowInlineHTML
		a.sanitizeHTML = *sanitizeHTML
		if !*allowInlineHTML {
			fmt.Println("--nohtml option provided: Inline HTML rendering is disabled.")
		}
		if !*sanitizeHTML {
			fmt.Println("--nosanitize option provided: HTML and URL sanitization is disabled.")
		}
        if err := flag.ErrHelp; err != nil {
            log.Println("Error parsing flags:", err)
            flag.Usage()
        }

        // 1. Build the usage string first
        var usageText strings.Builder
        usageText.WriteString(fmt.Sprintf("Usage: %s [options] [file]\n\n", os.Args[0]))
        usageText.WriteString("Options:\n")

        flag.VisitAll(func(f *flag.Flag) {
            if !f.Hidden {
                line := fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name)
                line += fmt.Sprintf("\t%s", f.Usage)
                if f.DefValue != "" {
                    line += fmt.Sprintf(" (default: %s)", f.DefValue)
                }
                usageText.WriteString(line + "\n")
            }
        })
        a.cmdlineOptions = usageText.String()

        // 2. Assign a function to flag.Usage that prints the string
        flag.Usage = func() {
            fmt.Fprint(os.Stderr, a.cmdlineOptions)
        }
    }

    return err
}

func (a *App) CreateGoldmarkInstance() goldmark.Markdown {
    options := []goldmark.Option{
        goldmark.WithParserOptions(
            parser.WithAutoHeadingID(), // Automatically generate IDs for headings
            parser.WithAttribute(),      // Enable attributes for nodes
        ),
        goldmark.WithExtensions(
            &frontmatter.Extender{}, // Add the frontmatter extension
            extension.GFM,
            extension.DefinitionList,
            extension.Footnote,
            extension.Typographer,
        ),
    }

    // Conditionally add renderer options based on allowInlineHTML setting
    if a.allowInlineHTML {
        options = append(options,
			goldmark.WithRendererOptions(
            	html.WithUnsafe(), // Allow unsafe HTML rendering
    	    ),
		)
    }

	if a.sanitizeHTML {
		options = append(options,
			goldmark.WithExtensions(
				&SanitizeHTMLExtension{}, // Custom extension to sanitize HTML
			),
		)
	}

    return goldmark.New(options...)
}

// // ProcessMarkdown reads a markdown file, renders it to HTML using Goldmark, and returns the HTML string.
// func (a *App) ProcessMarkdown(filepath string) (string, error) {
//     content, err := os.ReadFile(filepath)
//     if err != nil {
//         return "", fmt.Errorf("could not read file: %w", err)
//     }

//     md := goldmark.New(
//         goldmark.WithExtensions(
//             &frontmatter.Extender{}, // Add the frontmatter extension
//         ),
//     )

//     var buf bytes.Buffer
//     var meta map[string]string
//     context := parser.NewContext() // Create a context for parsing
//     if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
//         return "", fmt.Errorf("could not convert markdown: %w", err)
//     }

//     // Extract frontmatter data from the context
//     fm := frontmatter.Get(context)
//     if fm != nil {
//         if err := fm.Decode(&meta); err == nil {
//             a.frontMatter = meta
//             log.Printf("Frontmatter data: %v", a.frontMatter) // Log the frontmatter data
//         }
//     }

//     return buf.String(), nil
// }

func (a *App) OpenFileMenuHandler(_ *menu.CallbackData) { // Corrected: Use *menu.CallbackData
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
    var thisDocumentTitle string
    if a.stripH1 {
	    thisDocumentTitle, mdContent, _ = ExtractH1(string(mdContent))
    }
	// thisDocumentTitle := ""
	// if err != nil {
	// 	return fmt.Errorf("failed to extract document title: %w", err) // Corrected: Use fmt.Errorf
	// }

	// Convert Markdown content to HTML
	htmlContent, docFrontmatter, err := a.ConvertMarkdownToHTML(mdContent)
	if err != nil {
		return fmt.Errorf("failed to convert Markdown to HTML: %w", err) // Corrected: Use fmt.Errorf
	}

	// Emit the converted HTML to the frontend.
	var docTitle, docDate, tmpDocTitle, tmpDocDate string
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
		tmpDocTitle = GetValueFromMap(docFrontmatter, "Title")
		tmpDocDate = GetValueFromMap(docFrontmatter, "Date")
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

	runtime.EventsEmit(a.ctx, "markdown-rendered", string(htmlContent), docTitle, docDate)
	return nil
}

// convertMarkdownToHTML converts a byte slice of Markdown content
// into a byte slice of HTML content using the configured Goldmark converter.
func (a *App) ConvertMarkdownToHTML(markdown []byte) ([]byte, map[string]string, error) {
	var buf strings.Builder
	var meta map[string]string
	cntxt := parser.NewContext()
	err := a.mdConverter.Convert(markdown, &buf, parser.WithContext(cntxt))
	if err != nil {
		return nil, nil, err
	}
	html := buf.String()
	fm := frontmatter.Get(cntxt)
	// var docTitle, docDate interface{}
	if fm == nil {
		return []byte(html), nil, nil
	}
	if err := fm.Decode(&meta); err != nil {
		return []byte(html), nil, nil
	}
	return []byte(html), meta, nil
}

func ExtractH1(md string) (string, []byte, error) {
	source := []byte(md)
	mdParser := goldmark.DefaultParser()
	reader := text.NewReader(source)
	doc := mdParser.Parse(reader)

	var h1Node *ast.Heading
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || h1Node != nil {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
			h1Node = heading
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})

	if h1Node == nil {
		// No H1 found - return empty title and original source
		return "", []byte(md), nil
	}

	// Extract title text
	title := ExtractTextContent(h1Node, source)
	if strings.TrimSpace(title) == "" {
		// Empty title - return original source
		return title, []byte(md), nil
	}

	// Remove H1 from source
	modifiedSource := RemoveNodeFromSource(source, h1Node)
	return title, []byte(modifiedSource), nil
}

func ExtractTextContent(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			segment := textNode.Segment
			value := segment.Value(source)
			buf.Write(value)
			if textNode.SoftLineBreak() {
				buf.WriteByte(' ')
			}
		} else {
			buf.WriteString(ExtractTextContent(child, source))
		}
	}
	return buf.String()
}

func RemoveNodeFromSource(source []byte, node ast.Node) string {
	segment := node.Lines().At(0)
	start := segment.Start
	end := segment.Stop

	// Extend removal range to include trailing newline if present
	if end < len(source) && source[end] == '\n' {
		end++
	} else if end < len(source)-1 && source[end] == '\r' && source[end+1] == '\n' {
		end += 2
	}

	// Remove the node's segment from the source
	return string(source[:start]) + string(source[end:])
}

func InitAlertIcons() map[string]string {
	// Set core list of icons
	var ai = map[string]string{
		"bug": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-bug-icon lucide-bug"><path d="m8 2 1.88 1.88"/><path d="M14.12 3.88 16 2"/><path d="M9 7.13v-1a3.003 3.003 0 1 1 6 0v1"/><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6"/><path d="M12 20v-9"/><path d="M6.53 9C4.6 8.8 3 7.1 3 5"/><path d="M6 13H2"/><path d="M3 21c0-2.1 1.7-3.9 3.8-4"/><path d="M20.97 5c0 2.1-1.6 3.8-3.5 4"/><path d="M22 13h-4"/><path d="M17.2 17c2.1.1 3.8 1.9 3.8 4"/></svg>`,
		"danger": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-octagon-alert-icon lucide-octagon-alert"><path d="M12 16h.01"/><path d="M12 8v4"/><path d="M15.312 2a2 2 0 0 1 1.414.586l4.688 4.688A2 2 0 0 1 22 8.688v6.624a2 2 0 0 1-.586 1.414l-4.688 4.688a2 2 0 0 1-1.414.586H8.688a2 2 0 0 1-1.414-.586l-4.688-4.688A2 2 0 0 1 2 15.312V8.688a2 2 0 0 1 .586-1.414l4.688-4.688A2 2 0 0 1 8.688 2z"/></svg>`,
		"example": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-notebook-pen-icon lucide-notebook-pen"><path d="M13.4 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-7.4"/><path d="M2 6h4"/><path d="M2 10h4"/><path d="M2 14h4"/><path d="M2 18h4"/><path d="M21.378 5.626a1 1 0 1 0-3.004-3.004l-5.01 5.012a2 2 0 0 0-.506.854l-.837 2.87a.5.5 0 0 0 .62.62l2.87-.837a2 2 0 0 0 .854-.506z"/></svg>`,
		"failure": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-circle-x-icon lucide-circle-x"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>`,
		"important": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-message-square-warning-icon lucide-message-square-warning"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><path d="M12 7v2"/><path d="M12 13h.01"/></svg>`,
		"info": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-info-icon lucide-info"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>`,
		"question": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-message-circle-question-icon lucide-message-circle-question"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>`,
		"quote": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-quote-icon lucide-quote"><path d="M16 3a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2 1 1 0 0 1 1 1v1a2 2 0 0 1-2 2 1 1 0 0 0-1 1v2a1 1 0 0 0 1 1 6 6 0 0 0 6-6V5a2 2 0 0 0-2-2z"/><path d="M5 3a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2 1 1 0 0 1 1 1v1a2 2 0 0 1-2 2 1 1 0 0 0-1 1v2a1 1 0 0 0 1 1 6 6 0 0 0 6-6V5a2 2 0 0 0-2-2z"/></svg>`,
		"success": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-circle-check-big-icon lucide-circle-check-big"><path d="M21.801 10A10 10 0 1 1 17 3.335"/><path d="m9 11 3 3L22 4"/></svg>`,
		"summary": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-clipboard-list-icon lucide-clipboard-list"><rect width="8" height="4" x="8" y="2" rx="1" ry="1"/><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><path d="M12 11h4"/><path d="M12 16h4"/><path d="M8 11h.01"/><path d="M8 16h.01"/></svg>  `,
		"tip": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-lightbulb-icon lucide-lightbulb"><path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.7 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/></svg>`,
		"todo": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-list-todo-icon lucide-list-todo"><rect x="3" y="5" width="6" height="6" rx="1"/><path d="m3 17 2 2 4-4"/><path d="M13 6h8"/><path d="M13 12h8"/><path d="M13 18h8"/></svg>`,
		"warning": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-triangle-alert-icon lucide-triangle-alert"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>`,
		"scroll": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-scroll-text-icon lucide-scroll-text"><path d="M15 12h-5"/><path d="M15 8h-5"/><path d="M19 17V5a2 2 0 0 0-2-2H4"/><path d="M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v1a2 2 0 1 1-4 0V5a2 2 0 1 0-4 0v2a1 1 0 0 0 1 1h3"/></svg>`,
	}

	// Set aliases
	ai["abstract"] = ai["summary"]
	ai["attention"] = ai["warning"]
	ai["caution"] = ai["danger"]
	ai["check"] = ai["success"]
	ai["cite"] = ai["quote"]
	ai["done"] = ai["success"]
	ai["error"] = ai["danger"]
	ai["fail"] = ai["failure"]
	ai["faq"] = ai["question"]
	ai["help"] = ai["question"]
	ai["hint"] = ai["tip"]
	ai["history"] = ai["scroll"]
	ai["missing"] = ai["failure"]
	ai["note"] = ai["info"]
	ai["tldr"] = ai["scroll"]
	ai["warn"] = ai["warning"]

	return ai
}