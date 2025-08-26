package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	dateparse "github.com/araddon/dateparse"
	encodingUnicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"md-reader/internal/markdown"
	"md-reader/internal/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
)

// MarkdownRenderData contains all data needed for the markdown-rendered event
type MarkdownRenderData struct {
	HTML        string `json:"html"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Frontmatter string `json:"frontmatter,omitempty"`
	// Future fields can be added here without breaking existing functionality
	// Type     string `json:"type,omitempty"`
	// Metadata map[string]string `json:"metadata,omitempty"`
}

// DocumentProcessor handles document processing and rendering
type DocumentProcessor struct {
    ctx               context.Context
    configManager     *ConfigManager
    docTypes          []string
    mdConverter       goldmark.Markdown
}

// NewDocumentProcessor creates a new DocumentProcessor
func NewDocumentProcessor(ctx context.Context, configManager *ConfigManager) *DocumentProcessor {
    return NewDocumentProcessorWithStyle(ctx, configManager)
}

// NewDocumentProcessorWithStyle creates a new DocumentProcessor with configuration manager
func NewDocumentProcessorWithStyle(ctx context.Context, configManager *ConfigManager) *DocumentProcessor {
    return &DocumentProcessor{
        ctx:               ctx,
        configManager:     configManager,
        docTypes:          []string{},
        mdConverter:       markdown.CreateGoldmarkInstance(configManager),
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

    // Convert Markdown content to HTML
    htmlContent, docFrontmatter, thisDocumentH1Title, err := markdown.ConvertMarkdownToHTML(dp.mdConverter, mdContent)
    if err != nil {
        return fmt.Errorf("failed to convert Markdown to HTML: %w", err)
    }

    // Process document metadata
    docTitle, docDate, docType := dp.processDocumentMetadata(filePath, thisDocumentH1Title, docFrontmatter)

    // Format frontmatter as HTML
    formatter := NewFrontmatterHTMLFormatter()
    frontmatterHTML := formatter.FormatAsHTML(docFrontmatter)

    // Cleanup HTML content
    htmlContent = CleanupHTMLContent(htmlContent)

    // Strip the first H1 element if configured
    if dp.configManager.StripH1() {
        htmlContent = stripFirstH1(htmlContent)
    }

    // Create structured data for the frontend
    renderData := MarkdownRenderData{
        HTML:        string(htmlContent),
        Title:       docTitle,
        Date:        docDate,
        Frontmatter: frontmatterHTML,
    }

    // Emit the converted HTML to the frontend
    runtime.EventsEmit(dp.ctx, "markdown-rendered", renderData)

    // Handle document type classes
    dp.updateDocumentClasses(docType)

    return nil
}

// processDocumentMetadata extracts and formats document metadata
func (dp *DocumentProcessor) processDocumentMetadata(filePath, extractedTitle string, docFrontmatter map[string]any) (string, string, string) {
    var docTitle, docDate, fmDocTitle, fmDocDate, fmDocType string
    // docFileTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
    docFileTitle := filepath.Base(filePath)
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

    // Extract frontmatter data - now using map[string]any the utils.GetValue generic to handle possible nil values
    if docFrontmatter != nil {
        if title, ok := utils.GetValue[string](docFrontmatter, "title"); ok {
            fmDocTitle = title
        }
        if date, ok := utils.GetValue[time.Time](docFrontmatter, "date"); ok {
            fmDocDate = date.String()
        }
        if docType, ok := utils.GetValue[string](docFrontmatter, "doctype"); ok {
            fmDocType = strings.ToLower(docType)
        }
    }

    // Determine document title
    if fmDocTitle != "" && dp.configManager.UseFrontmatterTitle() {
        docTitle = fmDocTitle
    } else if extractedTitle != "" {
        docTitle = extractedTitle
    } else {
        docTitle = fmt.Sprintf("File: %s", docFileTitle)
    }

    // Process document date from frontmatter
    if fmDocDate != "" {
        fmtDocDate := `<span class="date-label document-date">Document Date:</span> <span class="date-value document-date">%s</span>`
        tz := time.Now().Local().Location()
        dateString, err := dateparse.ParseIn(fmDocDate, tz)
        if err == nil {
            docDateDD = fmt.Sprintf(fmtDocDate, dateString.Format(timeLayout))
        } else {
            docDateDD = fmt.Sprintf(fmtDocDate, fmDocDate)
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

    return docTitle, docDate, fmDocType
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

// CleanupHTMLContent refines the generated HTML for better rendering.
func CleanupHTMLContent(htmlContent []byte) []byte {
    htmlString := string(htmlContent)

    re := regexp.MustCompile("(?si)" + `(>)\s*(<p>|<p\s+[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(?:(?:</body>|</html>)?\s)+(</code>)`)
    // htmlString = re.ReplaceAllString(htmlString, "\r\n$1")  // we do NOT want extra CRLF before closing </code>
    htmlString = re.ReplaceAllString(htmlString, "$1")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)(<code[^>]*>)(?:<html>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)\s*(<code[^>]*>)\s*(?:<body[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)\s*(<code[^>]*>)(?:\r\n|\n)*(\S+)`)
    // htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2\r\n$3") // we do NOT want extra CRLF after opening <code>
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2$3")

    re = regexp.MustCompile("(?si)" + `(>)\s*(<section[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    return []byte(htmlString)
}

// stripFirstH1 removes the first <h1> element and its contents from HTML content.
// If any other header elements (h2-h6) are encountered before the first h1,
// the original content is returned unchanged.
func stripFirstH1(htmlContent []byte) []byte {
    htmlString := string(htmlContent)

    // First pass: check if there are any h2-h6 elements before the first h1
    // This regex finds the first occurrence of any header element (h1-h6)
    headerPattern := regexp.MustCompile(`(?i)<h([1-6])[^>]*>`)
    headerMatch := headerPattern.FindStringSubmatch(htmlString)

    if headerMatch == nil {
        // No headers found, return original content
        return htmlContent
    }

    // Check if the first header found is not h1
    if headerMatch[1] != "1" {
        // First header is h2-h6, return original content unchanged
        return htmlContent
    }

    // Second pass: find and remove the first h1 element and its contents
    // This regex matches the opening h1 tag, its contents, and the closing tag
    // The (?s) flag makes . match newline characters as well
    h1Pattern := regexp.MustCompile(`(?is)<h1[^>]*>.*?</h1>`)

    // Find the first h1 match
    h1Match := h1Pattern.FindStringIndex(htmlString)
    if h1Match == nil {
        // No h1 found (shouldn't happen based on first pass, but safety check)
        return htmlContent
    }

    // Remove the first h1 element by concatenating the parts before and after it
    modifiedHTML := htmlString[:h1Match[0]] + htmlString[h1Match[1]:]

    return []byte(modifiedHTML)
}

