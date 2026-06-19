// Package mathext is a goldmark extension that recognises TeX math written
// between dollar delimiters and renders it as MathML. Inline math uses single
// dollars ($…$), display math uses double dollars ($$…$$). The actual
// TeX→MathML conversion is delegated to the texmath package; this package only
// owns the Markdown-level concerns: where a formula starts and ends, and how
// it serialises into the HTML stream that htmlbag consumes.
//
// Running as a real goldmark inline parser (rather than a regex pre-pass) buys
// two things for free: dollars inside `code spans` and fenced code blocks are
// never treated as math, and the strict opening/closing rules below disambiguate
// prose dollars ("it costs $5 and $6") from formulas — the same convention
// Pandoc uses and PLOS expects for Markdown submissions.
package mathext

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/boxesandglue/glu/markdown/texmath"
)

// KindMathInline is the AST node kind for a parsed math span.
var KindMathInline = ast.NewNodeKind("MathInline")

// MathInline is the AST node a math span parses into. Source is the raw TeX
// between the delimiters; Display selects $$…$$ (block) over $…$ (inline).
type MathInline struct {
	ast.BaseInline
	Source  []byte
	Display bool
}

// Kind implements ast.Node.
func (n *MathInline) Kind() ast.NodeKind { return KindMathInline }

// Dump implements ast.Node for debugging.
func (n *MathInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Source":  string(n.Source),
		"Display": map[bool]string{true: "block", false: "inline"}[n.Display],
	}, nil)
}

// mathParser is the inline parser triggered by '$'.
type mathParser struct{}

func (mathParser) Trigger() []byte { return []byte{'$'} }

// Parse implements the strict dollar-math rules. It only ever consumes input
// when it finds a well-formed formula; on any ambiguity it returns nil, which
// leaves the '$' as a literal text character.
func (mathParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	// Determine delimiter width: $$ is display, $ is inline.
	display := len(line) >= 2 && line[1] == '$'
	open := 1
	if display {
		open = 2
	}

	// Inline strict rule: the character right after the opening $ must not be
	// whitespace, so "$ x$" and a lone "$" in prose are not math.
	if !display {
		if open >= len(line) || isSpaceByte(line[open]) {
			return nil
		}
	}

	for i := open; i < len(line); i++ {
		c := line[i]
		if c == '\\' { // skip an escaped character (e.g. \$ inside the formula)
			i++
			continue
		}
		if c != '$' {
			continue
		}
		if display {
			if i+1 < len(line) && line[i+1] == '$' {
				return finish(block, line, open, i, i+2, true)
			}
			// A single $ inside a $$…$$ run is part of the formula.
			continue
		}
		// Inline closing rules: no whitespace immediately before the closing
		// $, and the closing $ is not directly followed by a digit (so
		// "$5 and $6" stays prose).
		if isSpaceByte(line[i-1]) {
			continue
		}
		if i+1 < len(line) && line[i+1] >= '0' && line[i+1] <= '9' {
			continue
		}
		return finish(block, line, open, i, i+1, false)
	}
	// No closing delimiter on this line: not math.
	return nil
}

// finish extracts the formula source between [open,close), advances the reader
// past the closing delimiter (which ends at end), and returns the node. Empty
// formulas are rejected (returns nil without consuming).
func finish(block text.Reader, line []byte, open, close, end int, display bool) ast.Node {
	src := line[open:close]
	if len(bytes.TrimSpace(src)) == 0 {
		return nil
	}
	block.Advance(end)
	n := &MathInline{Display: display}
	n.Source = append([]byte(nil), src...) // copy: the source buffer is shared
	return n
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// mathRenderer serialises a MathInline node to MathML in the HTML output.
type mathRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *mathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, r.render)
}

func (r *mathRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*MathInline)
	mathml, err := texmath.ToMathML(string(n.Source), n.Display)
	if err != nil {
		// Degrade gracefully: show the offending TeX verbatim as inline code
		// rather than failing the whole document render.
		_, _ = w.WriteString("<code>")
		_, _ = w.Write(util.EscapeHTML(n.Source))
		_, _ = w.WriteString("</code>")
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString(mathml)
	return ast.WalkSkipChildren, nil
}

// Extension is the goldmark extender. Add it to goldmark.New via
// goldmark.WithExtensions(mathext.Math).
type Extension struct{}

// Math is the ready-to-use extension instance.
var Math = &Extension{}

// Extend registers the dollar-math inline parser and its MathML renderer.
func (e *Extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(mathParser{}, 150),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&mathRenderer{}, 500),
	))
}
