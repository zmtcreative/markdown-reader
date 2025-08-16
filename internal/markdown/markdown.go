package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"md-reader/internal/gm-ext/htmlsanitize"
	// "md-reader/internal/gm-ext/sectionwrapper"
	sectionwrapper "github.com/ZMT-Creative/gm-sectionwrapper"

	alertcallouts "github.com/ZMT-Creative/gm-alert-callouts"
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

var GlobalAttributeFilter = util.NewBytesFilterString(`accesskey,autocapitalize,autofocus,class,contenteditable,dir,draggable,enterkeyhint,hidden,id,inert,inputmode,is,itemid,itemprop,itemref,itemscope,itemtype,lang,part,role,slot,spellcheck,style,tabindex,title,translate`) // nolint:lll
var CodeBlockAttributeFilter = GlobalAttributeFilter.ExtendString(`nolabel,nolable,label,lable`)
var dataPrefix = []byte("data-")

type GoldmarkInstanceOptions struct {
	AllowInlineHTML bool
	SanitizeHTML    bool
}

// CreateGoldmarkInstance creates and configures a new Goldmark instance.
func CreateGoldmarkInstance(opt GoldmarkInstanceOptions) goldmark.Markdown {
    myIcons := InitAlertIcons() // Initialize alert icons
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
            &alertcallouts.AlertCallouts{
                Icons: myIcons,
				DisableFolding: false,
            },
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
			&fancylists.FancyLists{},
        ),
    }

    if opt.AllowInlineHTML {
        options = append(options,
            goldmark.WithRendererOptions(
                html.WithUnsafe(), // Allow unsafe HTML rendering
            ),
        )
    }

    if opt.SanitizeHTML {
        options = append(options,
            goldmark.WithExtensions(
                &htmlsanitize.SanitizeHTMLExtension{}, // Custom extension to sanitize HTML
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