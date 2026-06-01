package app

// This file was created to expand test coverage without adding more tests to document_test.go.

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	encodingUnicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type capturedDocumentEvent struct {
	ctx  context.Context
	name string
	data []interface{}
}

func captureDocumentRuntime(t *testing.T) *[]capturedDocumentEvent {
	t.Helper()

	events := []capturedDocumentEvent{}
	originalEventsEmit := documentEventsEmit
	originalReadFile := documentReadFile

	documentEventsEmit = func(ctx context.Context, eventName string, optionalData ...interface{}) {
		dataCopy := append([]interface{}{}, optionalData...)
		events = append(events, capturedDocumentEvent{ctx: ctx, name: eventName, data: dataCopy})
	}
	documentReadFile = os.ReadFile

	t.Cleanup(func() {
		documentEventsEmit = originalEventsEmit
		documentReadFile = originalReadFile
	})

	return &events
}

func writeTempMarkdownFile(t *testing.T, name string, contents []byte) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(filePath, contents, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return filePath
}

func encodeUTF16WithBOM(t *testing.T, input string, endianness encodingUnicode.Endianness) []byte {
	t.Helper()

	encoded, _, err := transform.Bytes(
		encodingUnicode.UTF16(endianness, encodingUnicode.UseBOM).NewEncoder(),
		[]byte(input),
	)
	if err != nil {
		t.Fatalf("transform.Bytes() error = %v", err)
	}
	return encoded
}

func renderedPayloads(t *testing.T, events []capturedDocumentEvent) []MarkdownRenderData {
	t.Helper()

	payloads := []MarkdownRenderData{}
	for _, event := range events {
		if event.name != "markdown-rendered" {
			continue
		}
		payload, ok := event.data[0].(MarkdownRenderData)
		if !ok {
			t.Fatalf("markdown-rendered payload type = %T, want MarkdownRenderData", event.data[0])
		}
		payloads = append(payloads, payload)
	}
	return payloads
}

func lastRenderedPayload(t *testing.T, events []capturedDocumentEvent) MarkdownRenderData {
	t.Helper()

	payloads := renderedPayloads(t, events)
	if len(payloads) == 0 {
		t.Fatal("no markdown-rendered payloads captured")
	}
	return payloads[len(payloads)-1]
}

func TestNewDocumentProcessorInitializesConverter(t *testing.T) {
	cm := newTestConfigManager(t)
	ctx := context.Background()

	processor := NewDocumentProcessor(ctx, cm)
	processorWithStyle := NewDocumentProcessorWithStyle(ctx, cm)

	for name, dp := range map[string]*DocumentProcessor{
		"NewDocumentProcessor":          processor,
		"NewDocumentProcessorWithStyle": processorWithStyle,
	} {
		if dp == nil {
			t.Fatalf("%s() returned nil", name)
		}
		if dp.ctx != ctx {
			t.Fatalf("%s() ctx mismatch", name)
		}
		if dp.configManager != cm {
			t.Fatalf("%s() configManager mismatch", name)
		}
		if dp.mdConverter == nil {
			t.Fatalf("%s() did not initialize markdown converter", name)
		}
		if len(dp.docTypes) != 0 {
			t.Fatalf("%s() docTypes length = %d, want 0", name, len(dp.docTypes))
		}
	}
}

func TestDocumentProcessorLoadAndDisplayMarkdownFileNotFound(t *testing.T) {
	cm := newTestConfigManager(t)
	dp := NewDocumentProcessorWithStyle(context.Background(), cm)
	events := captureDocumentRuntime(t)
	missingPath := filepath.Join(t.TempDir(), "missing.md")

	err := dp.LoadAndDisplayMarkdown(missingPath)
	if err == nil {
		t.Fatal("LoadAndDisplayMarkdown() error = nil, want file not found error")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Fatalf("LoadAndDisplayMarkdown() error = %q, want file not found", err)
	}
	if len(*events) != 0 {
		t.Fatalf("captured %d events, want 0 for read failure", len(*events))
	}
}

func TestDocumentProcessorLoadAndDisplayMarkdownPermissionDenied(t *testing.T) {
	cm := newTestConfigManager(t)
	dp := NewDocumentProcessorWithStyle(context.Background(), cm)
	events := captureDocumentRuntime(t)
	documentReadFile = func(path string) ([]byte, error) {
		return nil, &os.PathError{Op: "open", Path: path, Err: fs.ErrPermission}
	}

	err := dp.LoadAndDisplayMarkdown(`C:\secret\restricted.md`)
	if err == nil {
		t.Fatal("LoadAndDisplayMarkdown() error = nil, want permission denied error")
	}
	if !strings.Contains(err.Error(), "permission denied to read file") {
		t.Fatalf("LoadAndDisplayMarkdown() error = %q, want permission denied", err)
	}
	if len(*events) != 0 {
		t.Fatalf("captured %d events, want 0 for read failure", len(*events))
	}
}

func TestDocumentProcessorLoadAndDisplayMarkdownHandlesUTF16Input(t *testing.T) {
	tests := []struct {
		name       string
		endianness encodingUnicode.Endianness
	}{
		{name: "utf16 little endian", endianness: encodingUnicode.LittleEndian},
		{name: "utf16 big endian", endianness: encodingUnicode.BigEndian},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := newTestConfigManager(t)
			cm.config.Application.UseStripH1 = true
			ctx := context.Background()
			dp := NewDocumentProcessorWithStyle(ctx, cm)
			events := captureDocumentRuntime(t)

			contents := "# Visible Heading\r\n\r\nBody paragraph.\r\n"
			filePath := writeTempMarkdownFile(t, "utf16.md", encodeUTF16WithBOM(t, contents, tt.endianness))

			if err := dp.LoadAndDisplayMarkdown(filePath); err != nil {
				t.Fatalf("LoadAndDisplayMarkdown() error = %v", err)
			}

			payloads := renderedPayloads(t, *events)
			if len(payloads) != 2 {
				t.Fatalf("markdown-rendered payload count = %d, want 2", len(payloads))
			}
			if !strings.Contains(payloads[0].HTML, "Loading document") {
				t.Fatalf("initial payload HTML = %q, want loading message", payloads[0].HTML)
			}

			finalPayload := payloads[1]
				if finalPayload.Title == "" {
					t.Fatal("final payload title is empty")
			}
			if strings.Contains(finalPayload.HTML, "Visible Heading") {
				t.Fatalf("final payload HTML still contains stripped H1: %q", finalPayload.HTML)
			}
			if !strings.Contains(finalPayload.HTML, "Body paragraph.") {
				t.Fatalf("final payload HTML missing body paragraph: %q", finalPayload.HTML)
			}
			if !strings.Contains(finalPayload.Date, "Last Modified:") {
				t.Fatalf("final payload Date missing last-modified date: %q", finalPayload.Date)
			}
				if !strings.Contains(finalPayload.FrontmatterHTML, "No ACTIVE frontmatter") {
					t.Fatalf("final payload FrontmatterHTML = %q, want commented-out frontmatter placeholder", finalPayload.FrontmatterHTML)
			}
		})
	}
}

func TestDocumentProcessorLoadAndDisplayMarkdownProcessesMetadataAndDocTypes(t *testing.T) {
	cm := newTestConfigManager(t)
	cm.config.Application.UseStripH1 = true
	ctx := context.Background()
	dp := NewDocumentProcessorWithStyle(ctx, cm)
	events := captureDocumentRuntime(t)
	contents := []byte("---\ntitle: Frontmatter Title\ndate: 2026-05-31T14:00:00Z\ndoctype: techdoc api\n---\n\n# Visible Heading\n\nBody paragraph.\n")
	filePath := writeTempMarkdownFile(t, "metadata.md", contents)

	if err := dp.LoadAndDisplayMarkdown(filePath); err != nil {
		t.Fatalf("LoadAndDisplayMarkdown() error = %v", err)
	}

	finalPayload := lastRenderedPayload(t, *events)
	if finalPayload.Title != "Frontmatter Title" {
		t.Fatalf("final payload title = %q, want %q", finalPayload.Title, "Frontmatter Title")
	}
	if !strings.Contains(finalPayload.Date, "Document Date:") {
		t.Fatalf("final payload Date missing document date: %q", finalPayload.Date)
	}
	if !strings.Contains(finalPayload.Date, "Last Modified:") {
		t.Fatalf("final payload Date missing last-modified date: %q", finalPayload.Date)
	}
	if !strings.Contains(finalPayload.FrontmatterHTML, "Frontmatter Title") {
		t.Fatalf("final payload FrontmatterHTML missing title: %q", finalPayload.FrontmatterHTML)
	}
	if !strings.Contains(finalPayload.FrontmatterHTML, "techdoc api") {
		t.Fatalf("final payload FrontmatterHTML missing doctype: %q", finalPayload.FrontmatterHTML)
	}

	addedClasses := []string{}
	for _, event := range *events {
		if event.name != "add-doc-class" {
			continue
		}
		classes, ok := event.data[0].([]string)
		if !ok {
			t.Fatalf("add-doc-class payload type = %T, want []string", event.data[0])
		}
		addedClasses = append(addedClasses, classes...)
	}
	if len(addedClasses) != 2 || addedClasses[0] != "techdoc" || addedClasses[1] != "api" {
		t.Fatalf("add-doc-class events = %#v, want [techdoc api]", addedClasses)
	}
}

func TestDocumentProcessorLoadAndDisplayMarkdownKeepsH1WhenConfigured(t *testing.T) {
	cm := newTestConfigManager(t)
	cm.config.Application.UseStripH1 = false
	dp := NewDocumentProcessorWithStyle(context.Background(), cm)
	events := captureDocumentRuntime(t)
	filePath := writeTempMarkdownFile(t, "crlf.md", []byte("# Heading\r\n\r\nParagraph line.\r\n"))

	if err := dp.LoadAndDisplayMarkdown(filePath); err != nil {
		t.Fatalf("LoadAndDisplayMarkdown() error = %v", err)
	}

	finalPayload := lastRenderedPayload(t, *events)
	if !strings.Contains(finalPayload.HTML, "<h1") {
		t.Fatalf("final payload HTML missing H1: %q", finalPayload.HTML)
	}
	if !strings.Contains(finalPayload.HTML, "Heading") {
		t.Fatalf("final payload HTML missing heading text: %q", finalPayload.HTML)
	}
	if finalPayload.Title != "Heading" {
		t.Fatalf("final payload title = %q, want %q", finalPayload.Title, "Heading")
	}
}

func TestDocumentProcessorLoadAndDisplayMarkdownReplacesAbbreviations(t *testing.T) {
	cm := newTestConfigManager(t)
	cm.config.Markdown.UseAbbreviations = true
	dp := NewDocumentProcessorWithStyle(context.Background(), cm)
	events := captureDocumentRuntime(t)
	filePath := writeTempMarkdownFile(t, "abbr.md", []byte("*[HTML]: Hyper Text Markup Language\n\n# Title\n\nHTML body.\n"))

	if err := dp.LoadAndDisplayMarkdown(filePath); err != nil {
		t.Fatalf("LoadAndDisplayMarkdown() error = %v", err)
	}

	finalPayload := lastRenderedPayload(t, *events)
	if !strings.Contains(finalPayload.HTML, `<abbr title="Hyper Text Markup Language">HTML</abbr> body.`) {
		t.Fatalf("final payload HTML missing abbreviation replacement: %q", finalPayload.HTML)
	}
	if !strings.Contains(finalPayload.FrontmatterHTML, "No ACTIVE frontmatter") {
		t.Fatalf("final payload FrontmatterHTML = %q, want commented-out frontmatter placeholder", finalPayload.FrontmatterHTML)
	}
}