package htmlsanitize

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	gohtml "golang.org/x/net/html"
)

// //go:embed assets/images/danger-square.svg
// var dangerSquare string

//go:embed assets/images/broken-file.svg
var BrokenFile string

// SanitizeHTMLExtension implements Goldmark extension for markdown and raw HTML filtering
type SanitizeHTMLExtension struct{}

// Extend registers our custom renderer with Goldmark
func (e *SanitizeHTMLExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&FilteredHTMLRenderer{}, 500),
	))
}

// FilteredHTMLRenderer handles HTML block and span rendering
type FilteredHTMLRenderer struct{}

// RegisterFuncs registers rendering functions
func (r *FilteredHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindImage, r.renderImage)
}

// renderHTMLBlock processes HTML block elements
func (r *FilteredHTMLRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.HTMLBlock)
	var sb strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		sb.Write(line.Value(source))
	}

	filtered := filterHTML(sb.String())
	_, _ = w.WriteString(filtered)
	return ast.WalkContinue, nil
}

// renderRawHTML processes inline HTML elements
func (r *FilteredHTMLRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.RawHTML)
	var sb strings.Builder
	for i := 0; i < n.Segments.Len(); i++ {
		segment := n.Segments.At(i)
		sb.Write(segment.Value(source))
	}

	filtered := filterHTML(sb.String())
	_, _ = w.WriteString(filtered)
	return ast.WalkContinue, nil
}

// renderLink processes markdown links
func (r *FilteredHTMLRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	link := node.(*ast.Link)
	if !entering {
		_, _ = w.WriteString("</a>")
		return ast.WalkContinue, nil
	}

	newSVG := ""
	titleText := ""

	// Filter the destination URL
	dest := string(link.Destination)
	allowed := isAllowedAttributeValue(dest)

	// Check if link is internal anchor
	isInternalAnchor := strings.HasPrefix(dest, "#")

	// Check if link has existing target attribute
	hasTarget := false
	if link.Attributes() != nil {
		for _, attr := range link.Attributes() {
			if string(attr.Name) == "target" {
				hasTarget = true
				break
			}
		}
	}

	// Add start of <a> tag withhref if URL is allowed
	if allowed {
		newSVG = ""
		_, _ = w.WriteString("<a href=\"")
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(link.Destination, true)))
		_ = w.WriteByte('"')
	} else {
		newSVG = resizeSVG(BrokenFile, 16, 16)
		_, _ = w.WriteString(`<a class="disallowed bad-href" badhref="`)
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(link.Destination, true)))
		_ = w.WriteByte('"')
		titleText = fmt.Sprintf("(Link to '%s' is disallowed by policy)", string(link.Destination))
	}

	// Add target="_blank" for external links without existing target
	if allowed && !isInternalAnchor && !hasTarget {
		_, _ = w.WriteString(` target="_blank"`)
	}

	// Render title if present
	if link.Title != nil {
		if titleText != "" {
			titleText = string(link.Title) + " " + titleText
		} else {
			titleText = string(link.Title)
		}
	}

	// Render title if present
	if titleText != "" {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(util.EscapeHTML([]byte(titleText)))
		_ = w.WriteByte('"')
	}

	// Render additional attributes
	if link.Attributes() != nil {
		for _, attr := range link.Attributes() {
			_ = w.WriteByte(' ')
			_, _ = w.Write(attr.Name)
			_, _ = w.WriteString(`="`)
			_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
			_ = w.WriteByte('"')
		}
	}
	_, _ = w.WriteString(`>` + newSVG + `&nbsp;`)
	return ast.WalkContinue, nil
}

// renderImage processes markdown images
func (r *FilteredHTMLRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	img := node.(*ast.Image)
	_, _ = w.WriteString("<img")

	// Filter and render src attribute if URL is allowed
	dest := string(img.Destination)
	if isAllowedAttributeValue(dest) {
		_, _ = w.WriteString(` src="`)
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(img.Destination, true)))
		_ = w.WriteByte('"')
	} else {
		// Render with  badsrc attribute if URL is disallowed and set class for css
		_, _ = w.WriteString(` class="disallowed bad-src" badsrc="`)
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(img.Destination, true)))
		_ = w.WriteByte('"')
	}

	// Render alt text
	_, _ = w.WriteString(` alt="`)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			_, _ = w.Write(textNode.Value(source))
		}
	}
	_ = w.WriteByte('"')

	// Render title if present
	if img.Title != nil {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(img.Title)
		_ = w.WriteByte('"')
	}

	// Render additional attributes
	if img.Attributes() != nil {
		for _, attr := range img.Attributes() {
			_ = w.WriteByte(' ')
			_, _ = w.Write(attr.Name)
			_, _ = w.WriteString(`="`)
			_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
			_ = w.WriteByte('"')
		}
	}
	_, _ = w.WriteString(" />")
	return ast.WalkSkipChildren, nil
}

// filterHTML processes raw HTML through our security filters
func filterHTML(raw string) string {
	doc, err := gohtml.ParseFragment(strings.NewReader(raw), nil)
	if err != nil {
		return raw // Return original on parse error
	}

	filteredNodes := filterNodeList(doc)
	var buf bytes.Buffer
	for _, node := range filteredNodes {
		_ = gohtml.Render(&buf, node)
	}
	return buf.String()
}

// filterNodeList processes a list of HTML nodes
func filterNodeList(nodes []*gohtml.Node) []*gohtml.Node {
	var result []*gohtml.Node
	for _, node := range nodes {
		if filtered := filterNode(node); filtered != nil {
			result = append(result, filtered)
		}
	}
	return result
}

// filterNode applies security filters to an HTML node
func filterNode(node *gohtml.Node) *gohtml.Node {
	// Remove forbidden elements and their contents
	if node.Type == gohtml.ElementNode {
		forbiddenTags := []string{
			"script", "dialog", "embed", "iframe",
			"form", "button", "input", "select",
		}
		for _, tag := range forbiddenTags {
			if strings.EqualFold(node.Data, tag) {
				return nil // Completely remove the element
			}
		}
	}

	// Filter attributes in allowed elements
	if node.Type == gohtml.ElementNode {
		filterAttributes(node)
	}

	// Add target="_blank" to external <a> tags
	if node.Type == gohtml.ElementNode && node.Data == "a" {
		addTargetBlank(node)
	}

	// Collect child nodes to process
	var childNodes []*gohtml.Node
	for child := node.FirstChild; child != nil; {
		// Save next sibling before processing since filtering might detach nodes
		next := child.NextSibling
		childNodes = append(childNodes, child)
		child = next
	}

	// Process child nodes
	var filteredChildren []*gohtml.Node
	for _, child := range childNodes {
		if filtered := filterNode(child); filtered != nil {
			filteredChildren = append(filteredChildren, filtered)
		} else {
			tmpCommentNode := child
			tmpCommentNode.Data = fmt.Sprintf(" Removed by HTML sanitizer: %s ", child.Data)
			tmpCommentNode.Type = gohtml.CommentNode
			filteredChildren = append(filteredChildren, tmpCommentNode)
		}
	}

	// Remove all existing children
	node.FirstChild = nil
	node.LastChild = nil

	// Append filtered children in order
	for _, child := range filteredChildren {
		// Clear sibling references before appending
		child.Parent = nil
		child.PrevSibling = nil
		child.NextSibling = nil
		node.AppendChild(child)
	}

	return node
}

// addTargetBlank adds target="_blank" to external links
func addTargetBlank(node *gohtml.Node) {
	// Check if this is an internal anchor link
	isInternalAnchor := false
	hasTarget := false
	hasHref := false

	for _, attr := range node.Attr {
		switch attr.Key {
		case "href":
			hasHref = true
			if strings.HasPrefix(attr.Val, "#") {
				isInternalAnchor = true
			}
		case "target":
			hasTarget = true
		}
	}

	// Add target="_blank" for external links without existing target
	if hasHref && !isInternalAnchor && !hasTarget {
		node.Attr = append(node.Attr, gohtml.Attribute{
			Key: "target",
			Val: "_blank",
		})
	}
}

// filterAttributes removes unsafe attributes from elements
func filterAttributes(node *gohtml.Node) {
	var attrs []gohtml.Attribute
	for _, attr := range node.Attr {
		if !isSensitiveAttribute(attr.Key) || isAllowedAttributeValue(attr.Val) {
			attrs = append(attrs, attr)
		} else {
			var newAttr gohtml.Attribute
			newAttr.Key = "bad-" + strings.ToLower(attr.Key)
			newAttr.Val = attr.Val
			attrs = append(attrs, newAttr)
			newAttr.Key = "class"
			newAttr.Val = "disallowed bad-" + strings.ToLower(attr.Key)
			attrs = append(attrs, newAttr)
		}
	}
	node.Attr = attrs
}

// isSensitiveAttribute checks if attribute needs filtering
func isSensitiveAttribute(attrName string) bool {
	sensitiveAttrs := []string{"action", "cite", "data", "href", "src"}
	for _, attr := range sensitiveAttrs {
		if strings.EqualFold(attrName, attr) {
			return true
		}
	}
	return false
}

// isAllowedAttributeValue determines if attribute value is safe
func isAllowedAttributeValue(value string) bool {
	if value == "" {
		return true
	}

	// Allow fragment-only identifiers
	// (this only checks for values starting with #, not full URLs)
	if value[0] == '#' {
		return isValidFragment(value[1:])
	}

	// Allow directory paths
	if strings.HasSuffix(value, "/") {
		return true
	}

	// Extract and analyze URL components
	u, err := url.Parse(value)
	if err != nil {
		// Fallback to simple path extraction for invalid URLs
		_, ext := extractExtension(value)
		return isExtensionAllowed(ext)
	}

	// Handle special case: domain-only URLs (e.g., "example.com")
	if isDomainOnlyURL(u) {
		return true
	}

	// Check to see if the URL has a fragment
	if u.Fragment != "" && isValidFragment(u.Fragment) {
		return true
	}

	// // Check to see if the URL has a query string
	// if u.RawQuery != "" {
	// 	return true
	// }

	// Extract file extension from path
	base := path.Base(u.Path)
	ext := path.Ext(base)

	// If no extension found, allow the URL
	if ext == "" {
		return true
	}

	// Check if extension is allowed
	return isExtensionAllowed(ext)
}

// isDomainOnlyURL checks if URL represents a domain only (no path or file)
func isDomainOnlyURL(u *url.URL) bool {
	// Valid domain-only cases:
	// 1. Empty path (e.g., "http://example.com")
	// 2. Single slash path (e.g., "http://example.com/")
	return u.Path == "" || u.Path == "/"
}

// isExtensionAllowed checks if a file extension is permitted
func isExtensionAllowed(ext string) bool {
	ext = strings.ToLower(ext)
	if (ext == "") || (ext == ".") {
		return true
	}

	// Check against allowed extensions
	allowedExts := []string{
		// Archive types
		".zip", ".rar", ".7z", ".tar", ".gz", ".bz2",
		// Plain text Document types
		".css", ".htm", ".html", ".md", ".txt", ".xml",
		// Non-text Document types
		".docx", ".xlsx", ".pptx", ".pdf",
		// Images
		".png", ".apng", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp", ".ico", ".tiff", ".tif", ".icns",
		// Videos
		".mp4", ".ogg", ".webm", ".mov", ".avi", ".wmv", ".flv", ".mkv", ".m4v",
		// Audio
		".mp3", ".ogg", ".wav", ".flac", ".aac", ".m4a", ".wma", ".opus",
	}

	// If the extension is in the list of allowed extensions, return true
	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}

	return false // Forbid all other extensions by default
}

// extractExtension retrieves file extension from a string
func extractExtension(rawValue string) (string, string) {
	base := path.Base(rawValue)
	ext := path.Ext(base)
	return base, ext
}

// isValidFragment checks fragment identifier format
func isValidFragment(fragment string) bool {
	if fragment == "" {
		return true
	}
	validFragment := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validFragment.MatchString(fragment)
}

func resizeSVG(svg string, width, height int) string {
	if width == 0 || height == 0 {
		if width == 0 {
			width = height
		}
		if height == 0 {
			height = width
		}
	}
	newSVG := regexp.MustCompile(`\s+width="[^"]+"`).ReplaceAllString(svg, fmt.Sprintf(` width="%dpx"`, width))
	newSVG = regexp.MustCompile(`\s+height="[^"]+"`).ReplaceAllString(newSVG, fmt.Sprintf(` height="%dpx"`, height))
	return newSVG
}