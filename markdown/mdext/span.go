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

// KindBracketedSpan is the NodeKind of the BracketedSpan node.
var KindBracketedSpan = ast.NewNodeKind("BracketedSpan")

// A BracketedSpan node represents a pandoc bracketed span:
// [inline content]{.class key=value}.
type BracketedSpan struct {
	ast.BaseInline
}

// Kind implements ast.Node.Kind.
func (n *BracketedSpan) Kind() ast.NodeKind { return KindBracketedSpan }

// Dump implements ast.Node.Dump.
func (n *BracketedSpan) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// kindSpanOpener is the NodeKind of the spanOpener placeholder.
var kindSpanOpener = ast.NewNodeKind("BracketedSpanOpener")

// spanOpener is a placeholder inline node marking a '[' that lookahead
// confirmed will close as ']{attributes}'. It mirrors goldmark's internal
// linkLabelState: inline parsing continues normally after it, and the ']'
// handler wraps everything that follows it into a BracketedSpan. A
// placeholder that never gets matched (the confirming ']' was consumed by
// other inline syntax) is turned back into literal text in CloseBlock, so
// it never reaches a renderer.
type spanOpener struct {
	ast.BaseInline

	Segment text.Segment
	// bottom is pc.LastDelimiter() at open time; ProcessDelimiters gets it
	// when the span closes so emphasis inside the brackets resolves inside
	// the span (same dance as goldmark's link parser).
	bottom ast.Node
}

func (n *spanOpener) Kind() ast.NodeKind { return kindSpanOpener }

func (n *spanOpener) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// bracketedSpanParser parses pandoc bracketed spans. It runs before the
// link parser (priority 150 < 200): on '[' it claims the bracket only when
// lookahead proves a matching ']' immediately followed by parseable
// attributes, on ']' it only acts when its own opener stack is non-empty
// and '{' follows; in every other case it returns nil and the link parser
// proceeds untouched.
type bracketedSpanParser struct{}

// spanStackKey holds the stack of open spanOpener placeholders in the
// parser context.
var spanStackKey = parser.NewContextKey()

func (s *bracketedSpanParser) Trigger() []byte {
	return []byte{'[', ']'}
}

// spanFindClosureOptions: nesting so inner bracket pairs don't end the
// span, newline because inline content may wrap within the paragraph,
// codespan so a ']' inside `code` doesn't count as the closer.
var spanFindClosureOptions = text.FindClosureOptions{
	Nesting:  true,
	Newline:  true,
	CodeSpan: true,
	Advance:  true,
}

func (s *bracketedSpanParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()

	if line[0] == '[' {
		// Lookahead on the real reader; on the nil returns the framework
		// restores the position, on success we reset it ourselves.
		savedLine, savedPos := block.Position()
		block.Advance(1)
		if _, found := block.FindClosure('[', ']', spanFindClosureOptions); !found {
			return nil
		}
		if block.Peek() != '{' {
			return nil
		}
		if _, ok := parser.ParseAttributes(block); !ok {
			return nil
		}
		block.SetPosition(savedLine, savedPos)
		block.Advance(1)

		opener := &spanOpener{
			Segment: text.NewSegment(segment.Start, segment.Start+1),
			bottom:  pc.LastDelimiter(),
		}
		stack, _ := pc.Get(spanStackKey).([]*spanOpener)
		pc.Set(spanStackKey, append(stack, opener))
		return opener
	}

	// line[0] == ']'
	stack, _ := pc.Get(spanStackKey).([]*spanOpener)
	if len(stack) == 0 {
		return nil
	}
	if len(line) < 2 || line[1] != '{' {
		return nil
	}
	block.Advance(1)
	attrs, ok := parser.ParseAttributes(block)
	if !ok {
		return nil
	}

	opener := stack[len(stack)-1]
	pc.Set(spanStackKey, stack[:len(stack)-1])
	parser.ProcessDelimiters(opener.bottom, pc)

	span := &BracketedSpan{}
	for _, attr := range attrs {
		span.SetAttribute(attr.Name, attr.Value)
	}
	openerParent := opener.Parent()
	for c := opener.NextSibling(); c != nil; {
		next := c.NextSibling()
		openerParent.RemoveChild(openerParent, c)
		span.AppendChild(span, c)
		c = next
	}
	openerParent.RemoveChild(openerParent, opener)
	return span
}

// CloseBlock implements parser.CloseBlocker: pending openers whose
// confirming ']' was consumed by other inline syntax become literal text
// again, exactly like dangling link labels.
func (s *bracketedSpanParser) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	stack, _ := pc.Get(spanStackKey).([]*spanOpener)
	if stack == nil {
		return
	}
	for _, opener := range stack {
		if p := opener.Parent(); p != nil {
			p.ReplaceChild(p, opener, ast.NewTextSegment(opener.Segment))
		}
	}
	pc.Set(spanStackKey, nil)
}

// BracketedSpanHTMLRenderer renders a BracketedSpan as a span element
// carrying all authored attributes.
type BracketedSpanHTMLRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *BracketedSpanHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindBracketedSpan, r.renderBracketedSpan)
	reg.Register(kindSpanOpener, r.renderSpanOpener)
}

func (r *BracketedSpanHTMLRenderer) renderBracketedSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<span")
		html.RenderAttributes(w, node, nil)
		_, _ = w.WriteString(">")
	} else {
		_, _ = w.WriteString("</span>")
	}
	return ast.WalkContinue, nil
}

// renderSpanOpener is a safety net: CloseBlock replaces every dangling
// opener, so this should be unreachable. Render the literal '[' rather
// than crashing if it ever isn't.
func (r *BracketedSpanHTMLRenderer) renderSpanOpener(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("[")
	}
	return ast.WalkContinue, nil
}

type bracketedSpans struct{}

// BracketedSpans is the goldmark extender enabling pandoc bracketed spans.
var BracketedSpans goldmark.Extender = &bracketedSpans{}

func (e *bracketedSpans) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&bracketedSpanParser{}, 150),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&BracketedSpanHTMLRenderer{}, 500),
	))
}
