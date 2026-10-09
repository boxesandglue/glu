package mdext

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindTableCaption is the NodeKind of the TableCaption node.
var KindTableCaption = ast.NewNodeKind("TableCaption")

// A TableCaption is the first child of a table and holds the inline content
// of pandoc's caption paragraph, "Table: text" or ": text".
type TableCaption struct{ ast.BaseBlock }

// Kind implements ast.Node.Kind.
func (n *TableCaption) Kind() ast.NodeKind { return KindTableCaption }

// Dump implements ast.Node.Dump.
func (n *TableCaption) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }

// captionPrefixes are the paragraph starts pandoc takes as a table caption.
var captionPrefixes = [][]byte{[]byte("Table:"), []byte("table:"), []byte(":")}

// tableCaptionTransformer turns a paragraph that starts with a caption prefix
// and stands directly after or before a table into that table's caption.
// After wins, as in pandoc. Attributes on the paragraph (an attribute line
// below it) go to the table, so the id of a table can sit at its caption.
type tableCaptionTransformer struct{}

func (t *tableCaptionTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	var tables []ast.Node
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == east.KindTable {
			tables = append(tables, n)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	for _, table := range tables {
		para, text, cut := captionParagraph(table.NextSibling(), source)
		if para == nil {
			para, text, cut = captionParagraph(table.PreviousSibling(), source)
		}
		if para == nil {
			continue
		}
		// Strip the prefix and the space after it. Linkify splits text at
		// spaces, so the space can sit in the next text node, and a prefix
		// alone on the first line leaves an empty node behind.
		text.Segment = text.Segment.WithStart(text.Segment.Start + cut)
		for text != nil {
			seg := text.Segment
			text.Segment = seg.TrimLeftSpace(source)
			if !text.Segment.IsEmpty() {
				break
			}
			next, _ := text.NextSibling().(*ast.Text)
			para.RemoveChild(para, text)
			text = next
		}

		caption := &TableCaption{}
		for c := para.FirstChild(); c != nil; c = para.FirstChild() {
			para.RemoveChild(para, c)
			caption.AppendChild(caption, c)
		}
		for _, attr := range para.Attributes() {
			if _, exists := table.Attribute(attr.Name); !exists {
				table.SetAttribute(attr.Name, attr.Value)
			}
		}
		para.Parent().RemoveChild(para.Parent(), para)
		table.InsertBefore(table, table.FirstChild(), caption)
	}
}

// captionParagraph returns n, its first text node and the length of the
// caption prefix in that node if n is a paragraph that starts with a prefix
// followed by a space, a tab or the end of the line.
func captionParagraph(n ast.Node, source []byte) (*ast.Paragraph, *ast.Text, int) {
	para, ok := n.(*ast.Paragraph)
	if !ok {
		return nil, nil, 0
	}
	text, ok := para.FirstChild().(*ast.Text)
	if !ok {
		return nil, nil, 0
	}
	v := text.Segment.Value(source)
	indent := len(v) - len(bytes.TrimLeft(v, " \t"))
	for _, p := range captionPrefixes {
		if !bytes.HasPrefix(v[indent:], p) {
			continue
		}
		if rest := v[indent+len(p):]; len(rest) == 0 || rest[0] == ' ' || rest[0] == '\t' {
			return para, text, indent + len(p)
		}
	}
	return nil, nil, 0
}

type tableCaptionRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *tableCaptionRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindTableCaption, func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			_, _ = w.WriteString("<caption>")
		} else {
			_, _ = w.WriteString("</caption>\n")
		}
		return ast.WalkContinue, nil
	})
}

type tableCaptions struct{}

// TableCaptions is the goldmark extender for pandoc's table_captions: a
// paragraph "Table: text" (or "table: text", ": text") after or before a
// table becomes its <caption>.
var TableCaptions goldmark.Extender = &tableCaptions{}

func (e *tableCaptions) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		// After the block attributes (100), so an attribute line between
		// table and caption is gone and the table holds its id.
		util.Prioritized(&tableCaptionTransformer{}, 200),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&tableCaptionRenderer{}, 500),
	))
}
