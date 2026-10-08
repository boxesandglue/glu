package mdext

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindSuperscript and KindSubscript are the NodeKinds of the Superscript and
// Subscript nodes.
var (
	KindSuperscript = ast.NewNodeKind("Superscript")
	KindSubscript   = ast.NewNodeKind("Subscript")
)

// A Superscript node is pandoc's 2^10^, a Subscript node H~2~O. The text
// between the delimiters is the node's child.
type Superscript struct{ ast.BaseInline }

// Kind implements ast.Node.Kind.
func (n *Superscript) Kind() ast.NodeKind { return KindSuperscript }

// Dump implements ast.Node.Dump.
func (n *Superscript) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }

// Subscript is the node of H~2~O, see Superscript.
type Subscript struct{ ast.BaseInline }

// Kind implements ast.Node.Kind.
func (n *Subscript) Kind() ast.NodeKind { return KindSubscript }

// Dump implements ast.Node.Dump.
func (n *Subscript) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }

// scriptParser parses text between two delim characters on one line. As in
// pandoc, the text must not be empty or hold a space; a double ~ is left to
// strikethrough.
type scriptParser struct {
	delim byte
}

func (s *scriptParser) Trigger() []byte { return []byte{s.delim} }

func (s *scriptParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	if len(line) < 3 || line[1] == s.delim {
		return nil
	}
	end := bytes.IndexByte(line[1:], s.delim) + 1
	if end < 2 || bytes.ContainsAny(line[1:end], " \t\r\n") {
		return nil
	}
	var n ast.Node = &Superscript{}
	if s.delim == '~' {
		n = &Subscript{}
	}
	n.AppendChild(n, ast.NewTextSegment(text.NewSegment(segment.Start+1, segment.Start+end)))
	block.Advance(end + 1)
	return n
}

type scriptRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *scriptRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSuperscript, r.render("sup"))
	reg.Register(KindSubscript, r.render("sub"))
}

func (r *scriptRenderer) render(tag string) renderer.NodeRendererFunc {
	return func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			_, _ = w.WriteString("<" + tag + ">")
		} else {
			_, _ = w.WriteString("</" + tag + ">")
		}
		return ast.WalkContinue, nil
	}
}

type script struct{ delim byte }

// Superscripts is the goldmark extender for pandoc superscripts, 2^10^.
var Superscripts goldmark.Extender = &script{'^'}

// Subscripts is the goldmark extender for pandoc subscripts, H~2~O. It
// runs before strikethrough, which keeps ~~text~~.
var Subscripts goldmark.Extender = &script{'~'}

func (e *script) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		// Before strikethrough (500), after links (200).
		util.Prioritized(&scriptParser{delim: e.delim}, 400),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&scriptRenderer{}, 500),
	))
}
