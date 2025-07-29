package main

import (
	"strconv"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// SectionNode represents a <section> element in the AST
type SectionNode struct {
	ast.BaseBlock
	Level int
}

// KindSection is the node kind for SectionNode
var KindSection = ast.NewNodeKind("Section")

// Kind returns the node kind for SectionNode
func (n *SectionNode) Kind() ast.NodeKind {
	return KindSection
}

// Dump dumps the SectionNode for debugging
func (n *SectionNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// NewSectionNode creates a new SectionNode
func NewSectionNode(level int) *SectionNode {
	return &SectionNode{
		BaseBlock: ast.BaseBlock{},
		Level:     level,
	}
}

// sectionTransformer transforms the AST to wrap headings in sections
type sectionTransformer struct{}

// Transform implements the ASTTransformer interface
func (t *sectionTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	t.processChildren(doc, 0)
}

// processChildren processes child nodes to wrap headings in sections
func (t *sectionTransformer) processChildren(parent ast.Node, baseLevel int) {
	// Collect all direct children of the parent
	var children []ast.Node
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		children = append(children, child)
	}

	var newChildren []ast.Node // the new top-level nodes for the parent

	// Process nodes to group under sections
	i := 0
	for i < len(children) {
		child := children[i]
		if heading, ok := child.(*ast.Heading); ok && heading.Level > baseLevel {
			// Create a section for this heading
			section := NewSectionNode(heading.Level)
			section.AppendChild(section, heading) // Add the heading

			// Find the next heading with level <= current heading level
			j := i + 1
			for ; j < len(children); j++ {
				if nextHeading, ok := children[j].(*ast.Heading); ok && nextHeading.Level <= heading.Level {
					break
				}
			}

			// Add nodes between current and next heading to the section
			for k := i + 1; k < j; k++ {
				section.AppendChild(section, children[k])
			}

			newChildren = append(newChildren, section)
			i = j // Move to next unprocessed node
		} else {
			// Add non-heading or heading <= baseLevel directly
			newChildren = append(newChildren, child)
			i++
		}
	}

	// Replace parent's children with new structure
	parent.RemoveChildren(parent)
	for _, child := range newChildren {
		parent.AppendChild(parent, child)
	}

	// Recursively process sections
	for _, child := range newChildren {
		if section, ok := child.(*SectionNode); ok {
			t.processChildren(section, section.Level)
		}
	}
}

// sectionHTMLRenderer renders SectionNode to HTML
type sectionHTMLRenderer struct{}

// RegisterFuncs registers rendering functions
func (r *sectionHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSection, r.renderSection)
}

// renderSection renders a SectionNode
func (r *sectionHTMLRenderer) renderSection(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	sec := node.(*SectionNode)
	if entering {
		_, _ = w.WriteString(`<section class="section-h`)
		_, _ = w.WriteString(strconv.Itoa(sec.Level))
		_, _ = w.WriteString(`">`)
	} else {
		_, _ = w.WriteString("</section>")
	}
	return ast.WalkContinue, nil
}

// SectionExtension extends Goldmark to support sections
type SectionWrapperExtension struct{}

// Extend implements the goldmark.Extender interface
func (e *SectionWrapperExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&sectionTransformer{}, 100),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&sectionHTMLRenderer{}, 100),
	))
}
