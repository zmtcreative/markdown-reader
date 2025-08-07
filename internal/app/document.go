package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dateparse "github.com/araddon/dateparse"
	encodingUnicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"md-reader/internal/markdown"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
)

// DocumentProcessor handles document processing and rendering
type DocumentProcessor struct {
    ctx             context.Context
    stripH1         bool
    allowInlineHTML bool
    sanitizeHTML    bool
    docTypes        []string
    mdConverter     goldmark.Markdown
}

// NewDocumentProcessor creates a new DocumentProcessor
func NewDocumentProcessor(ctx context.Context, stripH1, allowInlineHTML, sanitizeHTML bool) *DocumentProcessor {
    return &DocumentProcessor{
        ctx:             ctx,
        stripH1:         stripH1,
        allowInlineHTML: allowInlineHTML,
        sanitizeHTML:    sanitizeHTML,
        docTypes:        []string{},
        mdConverter:     markdown.CreateGoldmarkInstance(allowInlineHTML, sanitizeHTML),
    }
}

// LoadAndDisplayMarkdown reads a Markdown file from the given path,
// converts its content to HTML using Goldmark, and then emits the HTML
// to the frontend via the "markdownLoaded" event.
func (dp *DocumentProcessor) LoadAndDisplayMarkdown(filePath string) error {
    mdContent, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("file not found: %s", filePath)
        }
        if os.IsPermission(err) {
            return fmt.Errorf("permission denied to read file: %s", filePath)
        }
        return fmt.Errorf("failed to read file %s: %w", filePath, err)
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
            }
        }
    }

    // Normalize line endings to Unix-style (LF)
    mdContent = []byte(strings.ReplaceAll(string(mdContent), "\r\n", "\n"))

    // Extract the document title from the H1 heading element if present
    var thisDocumentTitle string
    if dp.stripH1 {
        thisDocumentTitle, mdContent, _ = markdown.ExtractH1(string(mdContent))
    }

    // Convert Markdown content to HTML
    htmlContent, docFrontmatter, err := markdown.ConvertMarkdownToHTML(dp.mdConverter, mdContent)
    if err != nil {
        return fmt.Errorf("failed to convert Markdown to HTML: %w", err)
    }

    // Process document metadata
    docTitle, docDate, docType := dp.processDocumentMetadata(filePath, thisDocumentTitle, docFrontmatter)

    // Cleanup HTML content
    htmlContent = markdown.CleanupHTMLContent(htmlContent)

    // Emit the converted HTML to the frontend
    runtime.EventsEmit(dp.ctx, "markdown-rendered", string(htmlContent), docTitle, docDate)

    // Handle document type classes
    dp.updateDocumentClasses(docType)

    return nil
}

// processDocumentMetadata extracts and formats document metadata
func (dp *DocumentProcessor) processDocumentMetadata(filePath, extractedTitle string, docFrontmatter map[string]string) (string, string, string) {
    var docTitle, docDate, docType, tmpDocTitle, tmpDocDate string
    docFileTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
    timeLayout := time.DateTime + " MST"
    docDateLM := ""
    docDateDD := ""

    // Get file modification date
    if fileInfo, err := os.Stat(filePath); err == nil {
        fileModDate := fileInfo.ModTime()
        tz := fileModDate.Location()
        fmtLastModified := `<span class="date-label last-modified">Last Modified:</span> <span class="date-value last-modified">%s</span>`
        docDateLM = fmt.Sprintf(fmtLastModified, fileModDate.In(tz).Format(timeLayout))
    }

    // Extract frontmatter data - now using map[string]string
    if docFrontmatter != nil {
        tmpDocTitle = docFrontmatter["Title"]
        if tmpDocTitle == "" {
            tmpDocTitle = docFrontmatter["title"] // Try lowercase
        }

        tmpDocDate = docFrontmatter["Date"]
        if tmpDocDate == "" {
            tmpDocDate = docFrontmatter["date"] // Try lowercase
        }

        docType = strings.ToLower(docFrontmatter["Type"])
        if docType == "" {
            docType = strings.ToLower(docFrontmatter["type"]) // Try lowercase
        }
    }

    // Determine document title
    if tmpDocTitle != "" {
        docTitle = tmpDocTitle
    } else if extractedTitle != "" {
        docTitle = extractedTitle
    } else {
        docTitle = fmt.Sprintf("File: %s", docFileTitle)
    }

    // Process document date from frontmatter
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

    // Combine dates
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

    return docTitle, docDate, docType
}

// updateDocumentClasses manages CSS classes on the document
func (dp *DocumentProcessor) updateDocumentClasses(docType string) {
    // Remove existing document type classes
    if dp.docTypes != nil {
        for _, dt := range dp.docTypes {
            dp.RemoveDocClass(dt)
        }
    }

    // Reset docTypes slice
    dp.docTypes = []string{}

    // Add new document type classes
    if docType != "" {
        docTypeArray := strings.Fields(docType)
        for _, dt := range docTypeArray {
            dp.AddDocClass(dt)
            dp.docTypes = append(dp.docTypes, dt)
        }
    }
}

// AddDocClass adds the class to html and body elements
func (dp *DocumentProcessor) AddDocClass(thisClass ...string) {
    runtime.EventsEmit(dp.ctx, "add-doc-class", thisClass)
}

// RemoveDocClass removes the class from html and body elements
func (dp *DocumentProcessor) RemoveDocClass(thisClass ...string) {
    runtime.EventsEmit(dp.ctx, "remove-doc-class", thisClass)
}

// ToggleDocClass toggles the class on html and body elements
func (dp *DocumentProcessor) ToggleDocClass(thisClass ...string) {
    runtime.EventsEmit(dp.ctx, "toggle-doc-class", thisClass)
}