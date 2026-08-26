// Package mdext provides pandoc-flavoured goldmark extensions for the glu
// Markdown pipeline: fenced divs (::: {.class} … :::), bracketed spans
// ([text]{.class}) and a pandoc-style heading identifier generator that
// keeps non-ASCII letters.
package mdext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindFencedDiv is the NodeKind of the FencedDiv node.
var KindFencedDiv = ast.NewNodeKind("FencedDiv")

// A FencedDiv node represents one pandoc fenced div: a line of at least
// three colons carrying attributes opens a block container, a line of
// colons alone closes the innermost open one.
type FencedDiv struct {
	ast.BaseBlock
}

// Kind implements ast.Node.Kind.
func (n *FencedDiv) Kind() ast.NodeKind { return KindFencedDiv }

// Dump implements ast.Node.Dump.
func (n *FencedDiv) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// fencedDivParser parses pandoc fenced divs. The opening fence must carry
// attributes (braced or a bare class word): a bare colon run is always a
// closing fence, never an opener. That rule (from pandoc) is what makes
// nesting unambiguous without counting fence lengths.
type fencedDivParser struct{}

// divStackKey holds the stack of currently open FencedDiv nodes in the
// parser context. A closing fence line may only be consumed by the
// innermost open div; outer divs see the same line first (goldmark calls
// Continue outermost-first) and must pass it on.
var divStackKey = parser.NewContextKey()

func (p *fencedDivParser) Trigger() []byte {
	return []byte{':'}
}

func (p *fencedDivParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || line[pos] != ':' {
		return nil, parser.NoChildren
	}
	i := pos
	for ; i < len(line) && line[i] == ':'; i++ {
	}
	if i-pos < 3 {
		return nil, parser.NoChildren
	}

	// Everything after the colons, minus trailing whitespace and the
	// optional decorative colon run ("::: {.x} :::::").
	rest := util.TrimRightSpace(line[i:])
	j := len(rest)
	for j > 0 && rest[j-1] == ':' {
		j--
	}
	rest = util.TrimRightSpace(rest[:j])
	rest = rest[util.TrimLeftSpaceLength(rest):]
	if len(rest) == 0 {
		// No attributes: this is a closing fence (or stray colons), not
		// an opener.
		return nil, parser.NoChildren
	}

	node := &FencedDiv{}
	if rest[0] == '{' {
		// Parse the attributes on a throwaway reader: goldmark does not
		// restore the main reader's position when Open returns nil, so
		// nothing may be consumed before this line is fully validated.
		attrReader := text.NewReader(rest)
		attrs, ok := parser.ParseAttributes(attrReader)
		if !ok {
			return nil, parser.NoChildren
		}
		if leftover, _ := attrReader.PeekLine(); !util.IsBlank(leftover) {
			return nil, parser.NoChildren
		}
		for _, attr := range attrs {
			node.SetAttribute(attr.Name, attr.Value)
		}
	} else {
		// pandoc's shorthand: a single bare word is the class.
		for _, c := range rest {
			if util.IsSpace(c) || c == '{' || c == '}' {
				return nil, parser.NoChildren
			}
		}
		cls := make([]byte, len(rest))
		copy(cls, rest)
		node.SetAttributeString("class", cls)
	}

	stack, _ := pc.Get(divStackKey).([]ast.Node)
	pc.Set(divStackKey, append(stack, node))

	// Consume the whole fence line (up to the newline) so its remainder is
	// not offered to child block parsers.
	newline := 1
	if line[len(line)-1] != '\n' {
		newline = 0
	}
	reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
	return node, parser.HasChildren
}

func (p *fencedDivParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if util.IsBlank(line) {
		return parser.Continue | parser.HasChildren
	}
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 {
		return parser.Continue | parser.HasChildren
	}
	i := pos
	for ; i < len(line) && line[i] == ':'; i++ {
	}
	if i-pos < 3 || !util.IsBlank(line[i:]) {
		return parser.Continue | parser.HasChildren
	}

	// Closing fence. Only the innermost open div consumes it; an outer div
	// stays open so the next closing fence can close it in turn.
	stack, _ := pc.Get(divStackKey).([]ast.Node)
	if len(stack) == 0 || stack[len(stack)-1] != node {
		return parser.Continue | parser.HasChildren
	}
	pc.Set(divStackKey, stack[:len(stack)-1])
	newline := 1
	if line[len(line)-1] != '\n' {
		newline = 0
	}
	reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
	return parser.Close
}

func (p *fencedDivParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// A div can be force-closed (EOF, enclosing block ends) without its
	// closing fence: drop it from the stack wherever it sits.
	stack, _ := pc.Get(divStackKey).([]ast.Node)
	for idx := len(stack) - 1; idx >= 0; idx-- {
		if stack[idx] == node {
			pc.Set(divStackKey, append(stack[:idx:idx], stack[idx+1:]...))
			return
		}
	}
}

func (p *fencedDivParser) CanInterruptParagraph() bool {
	return true
}

func (p *fencedDivParser) CanAcceptIndentedLine() bool {
	return false
}

// FencedDivHTMLRenderer renders a FencedDiv as a div element carrying all
// authored attributes.
type FencedDivHTMLRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *FencedDivHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFencedDiv, r.renderFencedDiv)
}

func (r *FencedDivHTMLRenderer) renderFencedDiv(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<div")
		html.RenderAttributes(w, node, nil)
		_, _ = w.WriteString(">\n")
	} else {
		_, _ = w.WriteString("</div>\n")
	}
	return ast.WalkContinue, nil
}

type fencedDivs struct{}

// FencedDivs is the goldmark extender enabling pandoc fenced divs.
var FencedDivs goldmark.Extender = &fencedDivs{}

func (e *fencedDivs) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&fencedDivParser{}, 799),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&FencedDivHTMLRenderer{}, 500),
	))
}
