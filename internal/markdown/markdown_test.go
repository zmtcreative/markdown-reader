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

func TestInitAlertCalloutsIcons(t *testing.T) {
    icons := InitAlertCalloutsIcons(alertCalloutsCustomData)

    // Test that core icons are present
    coreIcons := []string{"bug", "danger", "example", "failure", "important", "info", "question", "quote", "success", "summary", "tip", "todo", "warning", "scroll"}
    for _, icon := range coreIcons {
        if _, exists := icons[icon]; !exists {
            t.Errorf("InitAlertIcons() missing core icon: %s", icon)
        }
        if len(icons[icon]) == 0 {
            t.Errorf("InitAlertIcons() core icon %s has empty SVG content", icon)
        }
        if !strings.Contains(icons[icon], "<svg") {
            t.Errorf("InitAlertIcons() core icon %s does not contain SVG content", icon)
        }
    }

    // Test that aliases are present and point to correct icons
    aliasTests := map[string]string{
        "abstract":  "summary",
        "attention": "warning",
        "caution":   "danger",
        "check":     "success",
        "cite":      "quote",
        "done":      "success",
        "error":     "danger",
        "fail":      "failure",
        "faq":       "question",
        "help":      "question",
        "hint":      "tip",
        "history":   "scroll",
        "missing":   "failure",
        "note":      "info",
        "tldr":      "scroll",
        "warn":      "warning",
    }

    for alias, primary := range aliasTests {
        if aliasIcon, exists := icons[alias]; !exists {
            t.Errorf("InitAlertIcons() missing alias: %s", alias)
        } else if primaryIcon, exists := icons[primary]; !exists {
            t.Errorf("InitAlertIcons() missing primary icon for alias %s: %s", alias, primary)
        } else if aliasIcon != primaryIcon {
            t.Errorf("InitAlertIcons() alias %s does not match primary %s", alias, primary)
        }
    }

    // Test that we have a reasonable number of total icons (core + aliases)
    if len(icons) < 5 {
        t.Errorf("InitAlertIcons() returned %d icons, expected at least 5", len(icons))
    }
}