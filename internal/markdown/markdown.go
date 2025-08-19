package markdown

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"md-reader/internal/gm-ext/htmlsanitize"
	// "md-reader/internal/gm-ext/sectionwrapper"
	alertcallouts "github.com/ZMT-Creative/gm-alert-callouts"
	sectionwrapper "github.com/ZMT-Creative/gm-sectionwrapper"

	fancylists "github.com/ZMT-Creative/gm-fancy-lists"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	figure "github.com/mangoumbrella/goldmark-figure"

	blockattr "github.com/mdigger/goldmark-attributes"
	bracketedspan "github.com/nemunaire/goldmark-inline-attributes"
	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/frontmatter"
	mermaid "go.abhg.dev/goldmark/mermaid"
)

//go:embed assets/alertcallouts-gfmstrict.icons
var alertCalloutsGFMStrictData string

var _ = alertCalloutsGFMStrictData

const (
	ALERT_NOICONS = iota
	ALERT_GFM_STRICT
	ALERT_GFM_WITH_ALIASES
	ALERT_GFM_PLUS
	ALERT_OBSIDIAN
)


var GlobalAttributeFilter = util.NewBytesFilterString(`accesskey,autocapitalize,autofocus,class,contenteditable,dir,draggable,enterkeyhint,hidden,id,inert,inputmode,is,itemid,itemprop,itemref,itemscope,itemtype,lang,part,role,slot,spellcheck,style,tabindex,title,translate`) // nolint:lll
var CodeBlockAttributeFilter = GlobalAttributeFilter.ExtendString(`nolabel,nolable,label,lable`)
var dataPrefix = []byte("data-")

type GoldmarkInstanceOptions struct {
	AllowInlineHTML   bool
	SanitizeHTML      bool
	AlertCalloutStyle string
}

// CreateGoldmarkInstance creates and configures a new Goldmark instance.
func CreateGoldmarkInstance(opt GoldmarkInstanceOptions) goldmark.Markdown {
    // Select alert callout icons based on style
    alertIconID := ALERT_NOICONS
    switch opt.AlertCalloutStyle {
	case "GFMStrict":
		alertIconID = ALERT_GFM_STRICT
    case "GFMWithAliases":
        alertIconID = ALERT_GFM_WITH_ALIASES
    case "GFMPlus":
        alertIconID = ALERT_GFM_PLUS
    case "Obsidian":
        alertIconID = ALERT_OBSIDIAN
    default:
        alertIconID = ALERT_GFM_STRICT // Default to Strict GFM
    }

    // myAlertCalloutsIcons := InitAlertCalloutsIcons(alertIconData)
    // var _ = myAlertCalloutsIcons
    options := []goldmark.Option{
        blockattr.Enable,
        bracketedspan.Enable,
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
            &mermaid.Extender{}, // Add Mermaid support for diagrams
            emoji.Emoji,
            figure.Figure.WithSkipNoCaption(),
            &anchor.Extender{
                Position: anchor.Before,
                Texter:   anchor.Text("#"),
            },
            &fences.Extender{},
            sectionwrapper.NewSectionWrapper(
				sectionwrapper.WithHeadingClass(true),
			),
            highlighting.NewHighlighting(
                highlighting.WithStyle("monokailight"),
                highlighting.WithWrapperRenderer(highlightingCustomWrapperRenderer),
                highlighting.WithFormatOptions(
                    chromahtml.WithClasses(false),
                    chromahtml.PreventSurroundingPre(true), // Let WithWrapperRenderer handle the <pre> tag
                    chromahtml.WithAllClasses(false),      // Use all classes for syntax highlighting
                    chromahtml.Standalone(true),           // Set to false to prevent a full HTML document
                ),
            ),
        ),
    }

	// Alert Callouts are always enabled, it's just the icon sets that change
	if alertIconID != ALERT_NOICONS {
		// Folding is enabled by default here
		options = append(options,
			goldmark.WithExtensions(
				alertcallouts.NewAlertCallouts(
					alertcallouts.WithFolding(true),
				),
			),
		)

		// Add alert callouts based on selected style
		switch alertIconID {
		case ALERT_GFM_STRICT:
			// Use strict GFM icons
			options = append(options,
				goldmark.WithExtensions(
					alertcallouts.NewAlertCallouts(
						alertcallouts.WithIcons(alertcallouts.CreateIconsMap(alertCalloutsGFMStrictData)),
						alertcallouts.WithFolding(false),
					),
				),
			)
		case ALERT_GFM_WITH_ALIASES:
			// Use standard GFM icons but with aliases for similar alert names (e.g., notes->note)
			options = append(options,
				goldmark.WithExtensions(
					alertcallouts.NewAlertCallouts(
						alertcallouts.UseGFMIcons(),
						// alertcallouts.WithFolding(true),
					),
				),
			)
		case ALERT_GFM_PLUS:
			// Use plus GFM icons
			options = append(options,
				goldmark.WithExtensions(
					alertcallouts.NewAlertCallouts(
						alertcallouts.UseGFMPlusIcons(),
						// alertcallouts.WithFolding(true),
					),
				),
			)
		case ALERT_OBSIDIAN:
			// Use Obsidian icons
			options = append(options,
				goldmark.WithExtensions(
					alertcallouts.NewAlertCallouts(
						alertcallouts.UseObsidianIcons(),
						// alertcallouts.WithFolding(true),
					),
				),
			)
		}
	}

    if opt.AllowInlineHTML {
        options = append(options,
            goldmark.WithRendererOptions(
                html.WithUnsafe(), // Allow unsafe HTML rendering
            ),
        )
    }

	// Sanitize HTML
	// Note to Self: There is some kind of parsing priority issue between the Sanitizer,
	//               Alert-Callouts and the FancyLists extension (these are all my extensions).
	//               The load order seems to matter here. Something to debug later.
	//
	//               THIS ordering seems to be working properly for now.
    if opt.SanitizeHTML {
        options = append(options,
            goldmark.WithExtensions(
                &htmlsanitize.SanitizeHTMLExtension{}, // Custom extension to sanitize HTML
				&fancylists.FancyListsOptions{},
            ),
        )
    } else {
        options = append(options,
            goldmark.WithExtensions(
				&fancylists.FancyListsOptions{},
            ),
        )
	}

    return goldmark.New(options...)
}

// ConvertMarkdownToHTML converts a byte slice of Markdown content into HTML.
func ConvertMarkdownToHTML(mdConverter goldmark.Markdown, markdown []byte) ([]byte, map[string]string, error) {
    var buf strings.Builder
    var meta map[string]string
    cntxt := parser.NewContext()
    err := mdConverter.Convert(markdown, &buf, parser.WithContext(cntxt))
    if err != nil {
        return nil, nil, err
    }
    html := buf.String()
    fm := frontmatter.Get(cntxt)
    if fm == nil {
        return []byte(html), nil, nil
    }
    if err := fm.Decode(&meta); err != nil {
        return []byte(html), nil, nil
    }
    return []byte(html), meta, nil
}

// CleanupHTMLContent refines the generated HTML for better rendering.
func CleanupHTMLContent(htmlContent []byte) []byte {
    htmlString := string(htmlContent)

    re := regexp.MustCompile("(?si)" + `(>)\s*(<p>|<p\s+[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(?:(?:</body>|</html>)?\s)+(</code>)`)
    htmlString = re.ReplaceAllString(htmlString, "\r\n$1")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)(<code[^>]*>)(?:<html>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)\s*(<code[^>]*>)\s*(?:<body[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    re = regexp.MustCompile("(?si)" + `(<pre[^>]*>)\s*(<code[^>]*>)(\S+)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2\r\n$3")

    re = regexp.MustCompile("(?si)" + `(>)\s*(<section[^>]*>)`)
    htmlString = re.ReplaceAllString(htmlString, "$1\r\n$2")

    return []byte(htmlString)
}

// ExtractH1 finds and removes the first H1 heading from Markdown source.
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
        return "", []byte(md), nil
    }

    title := ExtractTextContent(h1Node, source)
    if strings.TrimSpace(title) == "" {
        return title, []byte(md), nil
    }

    modifiedSource := RemoveNodeFromSource(source, h1Node)
    return title, []byte(modifiedSource), nil
}

func highlightingCustomWrapperRenderer(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
	if entering {
		lang, _ := c.Language()
		lang = getLanguageByAlias(lang)
        // Add language class to the <pre> tag
		fmt.Fprintf(w, `<pre language="%s"`, lang)
		if c.Attributes() != nil {
			renderCodeBlockAttributes(w, c, CodeBlockAttributeFilter)
		}
		fmt.Fprintf(w, `>`)
		// Add language class to the <code> tag
		fmt.Fprintf(w, `<code class="chroma" language="%s">`, lang)
	} else {
		_, _ = w.WriteString(`</code></pre>`)
	}
}

func renderCodeBlockAttributes(w util.BufWriter, c highlighting.CodeBlockContext, filter util.BytesFilter) {
	for _, attr := range c.Attributes().All() {
		if filter != nil && !filter.Contains(attr.Name) {
			if !bytes.HasPrefix(attr.Name, dataPrefix) {
				continue
			}
		}
		_, _ = w.WriteString(" ")
		_, _ = w.Write(attr.Name)
		_, _ = w.WriteString(`="`)
		var value []byte
		switch typed := attr.Value.(type) {
			case []byte:
				value = typed
			case string:
				value = util.StringToReadOnlyBytes(typed)
		}
		_, _ = w.Write(util.EscapeHTML(value))
		_ = w.WriteByte('"')
	}
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
    lines := node.Lines()
    if lines.Len() == 0 {
        return string(source)
    }

    // Get the full range from first line start to last line end
    firstLine := lines.At(0)
    lastLine := lines.At(lines.Len() - 1)

    // Find the actual start of the line (including any preceding whitespace and # markers)
    start := firstLine.Start
    // Go backwards to find the beginning of the line
    for start > 0 && source[start-1] != '\n' && source[start-1] != '\r' {
        start--
    }

    end := lastLine.Stop

    // Extend removal range to include trailing newline if present
    if end < len(source) && source[end] == '\n' {
        end++
    } else if end < len(source)-1 && source[end] == '\r' && source[end+1] == '\n' {
        end += 2
    }

    // Remove the node's segment from the source
    return string(source[:start]) + string(source[end:])
}

func InitAlertCalloutsIcons(icondata string) map[string]string {
	ai := make(map[string]string)

	// Parse the embedded alert callouts data
	lines := strings.Split(icondata, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if it's an alias definition (contains ->)
		if strings.Contains(line, "->") {
			parts := strings.SplitN(line, "->", 2)
			if len(parts) == 2 {
				alias := strings.TrimSpace(parts[0])
				primary := strings.TrimSpace(parts[1])
				// Set alias to reference the primary icon (will be set after core icons are loaded)
				if svg, exists := ai[primary]; exists {
					ai[alias] = svg
				} else {
					// Store for later processing if primary doesn't exist yet
					// This handles the case where aliases are defined before their primary keys
					defer func(alias, primary string) {
						if svg, exists := ai[primary]; exists {
							ai[alias] = svg
						}
					}(alias, primary)
				}
			}
			continue
		}

		// Parse core icon definition (key|svg)
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			svg := strings.TrimSpace(parts[1])
			ai[key] = svg
		}
	}

	// Second pass to handle any aliases that couldn't be resolved in first pass
	lines = strings.Split(icondata, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "->") {
			parts := strings.SplitN(line, "->", 2)
			if len(parts) == 2 {
				alias := strings.TrimSpace(parts[0])
				primary := strings.TrimSpace(parts[1])
				if svg, exists := ai[primary]; exists {
					ai[alias] = svg
				}
			}
		}
	}

	return ai
}

type CodeLanguage struct {
	Name string
	Aliases []string // Aliases for the language
}

var CodeLanguages = []CodeLanguage{
	{Name: "actionscript", Aliases: []string{"actionscript", "as", "as3", "actionscript3"}},
	{Name: "awk", Aliases: []string{"awk", "gawk", "nawk", "mawk"}},
	{Name: "bash", Aliases: []string{"bash", "bashrc", "ksh", "sh", "shell", "zsh", "zshrc"}},
	{Name: "basic", Aliases: []string{"basic", "bas", "qbasic"}},
	{Name: "batchfile", Aliases: []string{"batchfile", "bat", "cmd", "dos", "dosbatch", "windowsbatch", "winbatch"}},
	{Name: "c", Aliases: []string{"c", "c89", "c90", "c99", "c11", "h", "idc"}},
	{Name: "c++", Aliases: []string{"c++", "cpp", "cxx", "cc", "c++11", "c++14", "c++17", "c++20", "h++", "hpp", "hxx"}},
	{Name: "c#", Aliases: []string{"c#", "csharp", "cs"}},
	{Name: "clojure", Aliases: []string{"clojure", "clj"}},
	{Name: "coffeescript", Aliases: []string{"coffeescript", "coffee", "coffee-script", "cjs"}},
	{Name: "diff", Aliases: []string{"diff", "patch", "gitdiff", "gdiff", "udiff"}},
	{Name: "f#", Aliases: []string{"f#", "fsharp", "fs"}},
	{Name: "fortran", Aliases: []string{"fortran", "f77", "f90", "f95", "f03", "f08"}},
	{Name: "gnuplot", Aliases: []string{"gnuplot", "gp", "plot"}},
	{Name: "go", Aliases: []string{"go", "golang"}},
	{Name: "groovy", Aliases: []string{"groovy", "gradle", "gvy", "gy"}},
	{Name: "haskell", Aliases: []string{"haskell", "hs"}},
	{Name: "html", Aliases: []string{"html", "htm", "xhtml"}},
	{Name: "ini", Aliases: []string{"ini", "dosini", "winini", "inf"}},
	{Name: "javascript", Aliases: []string{"javascript", "js", "jsm"}},
	{Name: "julia", Aliases: []string{"julia", "jl"}},
	{Name: "kotlin", Aliases: []string{"kotlin", "kt"}},
	{Name: "lisp", Aliases: []string{"lisp", "cl", "common-lisp"}},
	{Name: "makefile", Aliases: []string{"makefile", "mf", "mk", "mak", "make"}},
	{Name: "markdown", Aliases: []string{"markdown", "md"}},
	{Name: "perl", Aliases: []string{"perl", "pl", "pm", "pod"}},
	{Name: "postscript", Aliases: []string{"postscript", "ps", "eps"}},
	{Name: "powershell", Aliases: []string{"powershell", "ps1", "psm1", "psd1", "pwsh", "posh"}},
	{Name: "python", Aliases: []string{"python", "py", "py3", "py2", "python2", "python3"}},
	{Name: "rexx", Aliases: []string{"rexx", "rex", "arexx", "rx"}},
	{Name: "ruby", Aliases: []string{"ruby", "rb", "rbw", "rbx", "rake", "gemspec", "gemfile"}},
	{Name: "rust", Aliases: []string{"rust", "rs"}},
	{Name: "typescript", Aliases: []string{"typescript", "ts", "tsx", "tsm"}},
}

func getLanguageByAlias(language []byte) []byte {
	for _, lang := range CodeLanguages {
		if lang.Name == string(language) || slices.Contains(lang.Aliases, string(language)) {
			return []byte(lang.Name)
		}
	}
	return language
}