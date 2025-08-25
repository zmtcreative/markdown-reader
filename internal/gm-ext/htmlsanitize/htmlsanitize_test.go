package htmlsanitize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestHTMLSanitizeAsExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic superscript",
			input:    `<sup>10</sup>`,
			expected: `<sup>10</sup>`,
		},
		{
			name:     "Basic subscript",
			input:    `<sub>2</sub>`,
			expected: `<sub>2</sub>`,
		},
		{
			name:     "Superscript in text",
			input:    `This line contains a superscript<sup>10</sup> test.`,
			expected: `This line contains a superscript<sup>10</sup> test.`,
		},
		{
			name:     "Subscript in text",
			input:    `This line contains a subscript<sub>2</sub> test.`,
			expected: `This line contains a subscript<sub>2</sub> test.`,
		},
		{
			name:     "Multiple inline elements",
			input:    `H<sub>2</sub>O and E=mc<sup>2</sup>`,
			expected: `H<sub>2</sub>O and E=mc<sup>2</sup>`,
		},
		{
			name:     "Nested elements",
			input:    `<strong>Bold<sup>10</sup></strong>`,
			expected: `<strong>Bold<sup>10</sup></strong>`,
		},
		{
			name:     "URL A HREF Python script should be disallowed",
			input:    `<a href="https://www.example.com/myapp.py">Python Script</a>`,
			expected: `<a bad-href="https://www.example.com/myapp.py" class="disallowed bad-href" title="(Link to 'https://www.example.com/myapp.py' is disallowed by policy)">Python Script</a>`,
		},
		{
			name:     "URL A HREF Python script in GET string should be disallowed",
			input:    `<a href="https://www.example.com/?prog=myapp.py&arg1=hello&arg2=world">Python Script</a>`,
			expected: `<a bad-href="https://www.example.com/?prog=myapp.py&arg1=hello&arg2=world" class="disallowed bad-href" title="(Link to 'https://www.example.com/?prog=myapp.py&arg1=hello&arg2=world' is disallowed by policy)">Python Script</a>`,
		},
		{
			name:     "URL A HREF Markdown document should be allowed",
			input:    `<a href="https://www.example.com/myapp.md">MyApp.md</a>`,
			expected: `<a href="https://www.example.com/myapp.md" target="_blank">MyApp.md</a>`,
		},
		{
			name:     "URL A HREF with GET parameters should be allowed",
			input:    `<a href="https://www.example.com/myapp?q=42&s=hello">MyApp.md</a>`,
			expected: `<a href="https://www.example.com/myapp?q=42&s=hello" target="_blank">MyApp.md</a>`,
		},
		{
			name:     "URL IMG SRC with executable link should be disallowed",
			input:    `<img src="https://www.example.com/myapp.dll">`,
			expected: `<img bad-src="https://www.example.com/myapp.dll" class="disallowed bad-src" title="(Link to 'https://www.example.com/myapp.dll' is disallowed by policy)">`,
		},
		{
			name:     "URL IMG SRC with image should be allowed",
			input:    `<img src="https://www.example.com/img/foo.png">`,
			expected: `<img src="https://www.example.com/img/foo.png">`,
		},
		{
			name:     "URL IMG SRC with GET parameters should be allowed",
			input:    `<img src="https://www.example.com/myapp?q=42&s=hello">`,
			expected: `<img src="https://www.example.com/myapp?q=42&s=hello">`,
		},
		{
			name:     "URL IMG SRC Check that self-closing tags work properly",
			input:    `<img src="https://www.example.com/img/foo.png"/>`,
			expected: `<img src="https://www.example.com/img/foo.png"/>`,
		},
		{
			name:     "BR - Check that self-closing tags work (no space)",
			input:    `<br/>`,
			expected: `<br/>`,
		},
		{
			name:     "BR - Check that self-closing tags work (extra space removed)",
			input:    `<br />`,
			expected: `<br/>`,
		},
		{
			name:     "Remove FORM element",
			input:    `<form><input type="text"></form>`,
			expected: `<!-- Removed by HTML sanitizer: form -->`,
		},
		{
			name:     "Leave SCRIPT element in HEAD",
			input:    `<head><script>alert("XSS");</script></head>`,
			expected: ``,
		},
		{
			name:     "Remove SCRIPT element in BODY",
			input:    `<body><script>alert("XSS");</script></body>`,
			expected: `<!-- Removed by HTML sanitizer: script -->`,
		},
		{
			name:     "Remove DIALOG element",
			input:    `<dialog><input></dialog>`,
			expected: `<!-- Removed by HTML sanitizer: dialog -->`,
		},
		{
			name:     "Keep allowed elements",
			input:    `<div><span>text</span></div>`,
			expected: `<div><span>text</span></div>`,
		},
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&SanitizeHTMLExtension{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			result := buf.String()

			// Remove the paragraph tags for inline tests since Goldmark wraps in <p>
			if strings.HasPrefix(result, "<p>") && strings.HasSuffix(result, "</p>\n") {
				result = strings.TrimSuffix(strings.TrimPrefix(result, "<p>"), "</p>\n")
			}

			if result != tt.expected {
				t.Errorf("Expected: %s", tt.expected)
				t.Errorf("Got:      %s", result)
			}
		})
	}
}

func TestFilterHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple superscript",
			input:    `<sup>10</sup>`,
			expected: `<sup>10</sup>`,
		},
		{
			name:     "Simple subscript",
			input:    `<sub>2</sub>`,
			expected: `<sub>2</sub>`,
		},
		{
			name:     "Remove form element",
			input:    `<form><input type="text"></form>`,
			expected: `<!-- Removed by HTML sanitizer: form -->`,
		},
		{
			name:     "Keep allowed elements",
			input:    `<div><span>text</span></div>`,
			expected: `<div><span>text</span></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHTML(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain: %s", tt.expected)
				t.Errorf("Got: %s", result)
			}
		})
	}
}

func TestDisallowedLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		shouldContain string
	}{
		{
			name:     "Python script should be disallowed",
			input:    `<a href="https://www.example.com/myapp.py">Python Script</a>`,
			shouldContain: `class="disallowed bad-href"`,
		},
		{
			name:     "Markdown document should be allowed",
			input:    `<a href="https://www.example.com/myapp.md">MyApp.md</a>`,
			shouldContain: `href="https://www.example.com/myapp.md"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHTML(tt.input)
			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("Expected result to contain: %s", tt.shouldContain)
				t.Errorf("Got: %s", result)
			}
		})
	}
}
