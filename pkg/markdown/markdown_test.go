package markdown

// This is a very rudimentary test for the ExtractH1 function.
// More tests need to be added for comprehensive coverage.

import (
	"strings"
	"testing"
)

func TestExtractH1(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        expectedTitle string
        expectedBody  string
        expectFound   bool
    }{
        {
            name:          "Simple H1",
            input:         "# My Title\n\nSome content.",
            expectedTitle: "My Title",
            expectedBody:  "\n\nSome content.",
            expectFound:   true,
        },
        {
            name:          "No H1",
            input:         "## Subtitle\n\nSome content.",
            expectedTitle: "",
            expectedBody:  "## Subtitle\n\nSome content.",
            expectFound:   true,
        },
        {
            name:          "H1 with surrounding text",
            input:         "pre-text\n# Title\npost-text",
            expectedTitle: "Title",
            expectedBody:  "pre-text\npost-text",
            expectFound:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            title, body, err := ExtractH1(tt.input) // Renamed 'found' to 'err' for clarity
            if title != tt.expectedTitle {
                t.Errorf("ExtractH1() title = %q, want %q", title, tt.expectedTitle)
            }
            if strings.TrimSpace(string(body)) != strings.TrimSpace(tt.expectedBody) {
                t.Errorf("ExtractH1() body = %q, want %q", string(body), tt.expectedBody)
            }

            // Check if an error was returned, which indicates whether the H1 was found.
            h1Found := err == nil
            if h1Found != tt.expectFound {
                t.Errorf("ExtractH1() found = %v, want %v (error: %v)", h1Found, tt.expectFound, err)
            }
        })
    }
}

func TestCleanupHTMLContent(t *testing.T) {
    input := `<p>Line 1</p><p>Line 2</p>`
    expected := "<p>Line 1</p>\r\n<p>Line 2</p>"
    output := CleanupHTMLContent([]byte(input))
    if string(output) != expected {
        t.Errorf("CleanupHTMLContent() = %q, want %q", string(output), expected)
    }
}