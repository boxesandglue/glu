package mdext

import (
	"log/slog"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindPageFootnote is the NodeKind of the PageFootnote node.
var KindPageFootnote = ast.NewNodeKind("PageFootnote")

// A PageFootnote node stands for a footnote reference whose note is set at
// the foot of the page. It has no children: the note's content stays in
// Footnote, which a note referenced twice shares.
type PageFootnote struct {
	ast.BaseInline
	Footnote *east.Footnote
}

// Kind implements ast.Node.Kind.
func (n *PageFootnote) Kind() ast.NodeKind { return KindPageFootnote }

// Dump implements ast.Node.Dump.
func (n *PageFootnote) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// pageFootnotes turns the notes of goldmark's footnote extension into page
// footnotes: each reference becomes <span class="footnote">, the content of
// the note, which htmlbag sets at the foot of the page, and the endnote list
// goes away. A note with more than one paragraph keeps its paragraphs apart
// with a line break. A note that holds another block, such as a list or
// code, cannot be set inline: then all notes of the document stay endnotes,
// with a warning.
type pageFootnotes struct {
	md goldmark.Markdown
}

// PageFootnotes is the extension that sets the notes of
// extension.Footnote as page footnotes. It needs extension.Footnote.
func PageFootnotes() goldmark.Extender { return &pageFootnotes{} }

// Extend implements goldmark.Extender.
func (e *pageFootnotes) Extend(m goldmark.Markdown) {
	e.md = m
	m.Parser().AddOptions(parser.WithASTTransformers(
		// After the footnote extension's transformer (999), which builds
		// the list of notes.
		util.Prioritized(e, 1000),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(e, 500),
	))
}

// Transform implements parser.ASTTransformer.
func (e *pageFootnotes) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	var list *east.FootnoteList
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		if l, ok := c.(*east.FootnoteList); ok {
			list = l
		}
	}
	if list == nil {
		return
	}
	notes := map[int]*east.Footnote{}
	for c := list.FirstChild(); c != nil; c = c.NextSibling() {
		fn := c.(*east.Footnote)
		for b := fn.FirstChild(); b != nil; b = b.NextSibling() {
			if b.Kind() != ast.KindParagraph {
				slog.Warn("A footnote holds a block that a page footnote cannot take, setting all notes as endnotes", "footnote", string(fn.Ref), "block", b.Kind().String())
				return
			}
		}
		notes[fn.Index] = fn
	}
	var links []*east.FootnoteLink
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if l, ok := n.(*east.FootnoteLink); ok && entering {
			links = append(links, l)
		}
		return ast.WalkContinue, nil
	})
	for _, l := range links {
		fn, ok := notes[l.Index]
		if !ok {
			continue
		}
		l.Parent().ReplaceChild(l.Parent(), l, &PageFootnote{Footnote: fn})
	}
	for _, fn := range notes {
		removeBacklinks(fn)
	}
	doc.RemoveChild(doc, list)
}

// removeBacklinks drops the links back to the reference, which the footnote
// extension puts at the end of a note's last paragraph.
func removeBacklinks(fn *east.Footnote) {
	for p := fn.FirstChild(); p != nil; p = p.NextSibling() {
		for c := p.FirstChild(); c != nil; {
			next := c.NextSibling()
			if c.Kind() == east.KindFootnoteBacklink {
				p.RemoveChild(p, c)
			}
			c = next
		}
	}
}

// RegisterFuncs implements renderer.NodeRenderer.
func (e *pageFootnotes) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindPageFootnote, e.renderPageFootnote)
}

// renderPageFootnote writes the note at the place of its reference. The
// note's inline content is rendered with the document's renderer, once for
// every reference to it.
func (e *pageFootnotes) renderPageFootnote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<span class="footnote">`)
	for p := node.(*PageFootnote).Footnote.FirstChild(); p != nil; p = p.NextSibling() {
		if p.PreviousSibling() != nil {
			_, _ = w.WriteString("<br>")
		}
		for c := p.FirstChild(); c != nil; c = c.NextSibling() {
			if err := e.md.Renderer().Render(w, source, c); err != nil {
				return ast.WalkStop, err
			}
		}
	}
	_, _ = w.WriteString("</span>")
	return ast.WalkSkipChildren, nil
}
