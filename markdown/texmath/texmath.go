// Package texmath converts a small, scope-controlled subset of TeX math
// notation into MathML. The output is deliberately limited to the element
// set understood by the boxesandglue MathML reader
// (boxesandglue/frontend/math/mathml): <mi> <mn> <mo> <mrow> <mfrac> <msqrt>
// <mroot> <msup> <msub> <msubsup>. Feeding the existing MathML entry point —
// rather than building math atoms directly — means TeX formulas inherit the
// whole rendering and accessibility pipeline for free (Formula tagging,
// MathML associated file under PDF/UA-2).
//
// This is a proof-of-concept subset, not a full LaTeX math engine. It covers
// what typical Markdown prose needs: variables, numbers, operators, super/
// subscripts, \frac, \sqrt (with optional index), Greek letters, common
// relation/operator symbols and named operators (\sin, \log, …). Unknown
// macros degrade to their literal name as an upright identifier rather than
// erroring out, so a document never fails to render over one stray command.
package texmath

import (
	"fmt"
	"strings"
)

// ToMathML converts a TeX math fragment into a MathML string. When display is
// true the root <math> carries display="block", which the boxesandglue MathML
// reader routes to DisplayMath (centred, larger operators); otherwise it is
// laid out inline. The returned string always has a single <math> root, so it
// drops straight into HTML where htmlbag picks it up as a formula.
func ToMathML(tex string, display bool) (string, error) {
	p := &parser{src: []rune(tex), display: display}
	items, err := p.parseList(true)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if display {
		b.WriteString(`<math display="block">`)
	} else {
		b.WriteString(`<math>`)
	}
	// A single child needs no wrapping <mrow>; multiple children are wrapped
	// so super/subscript and fraction logic upstream always sees one nucleus.
	b.WriteString(wrapRow(items))
	b.WriteString(`</math>`)
	return b.String(), nil
}

// parser walks the rune slice with a single read cursor. There is no separate
// lexer: TeX math tokenisation is simple enough (single chars, brace groups
// and backslash macros) that a hand-rolled recursive-descent reader stays
// clearer than a token stream.
type parser struct {
	src []rune
	pos int
	// display records whether the formula is display math; limit
	// operators then take stacked scripts (munderover) instead of side
	// scripts, mirroring TeX's movablelimits behavior.
	display bool
}

func (p *parser) atEnd() bool { return p.pos >= len(p.src) }

func (p *parser) peek() rune {
	if p.atEnd() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) next() rune {
	r := p.src[p.pos]
	p.pos++
	return r
}

func (p *parser) skipSpace() {
	for !p.atEnd() && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t' || p.src[p.pos] == '\n' || p.src[p.pos] == '\r') {
		p.pos++
	}
}

// parseList parses a run of atoms until end-of-input (topLevel) or a closing
// brace (nested). Each base atom is immediately offered to parseScripts so
// that ^ and _ bind to the atom on their left, exactly as in TeX.
func (p *parser) parseList(topLevel bool) ([]string, error) {
	var out []string
	for {
		p.skipSpace()
		if p.atEnd() {
			if !topLevel {
				return nil, fmt.Errorf("unexpected end of input: missing }")
			}
			break
		}
		if p.peek() == '}' {
			if topLevel {
				return nil, fmt.Errorf("unexpected }")
			}
			break
		}
		base, err := p.parseAtomBase()
		if err != nil {
			return nil, err
		}
		if base == "" {
			continue
		}
		base, err = p.parseScripts(base)
		if err != nil {
			return nil, err
		}
		out = append(out, base)
	}
	return out, nil
}

// parseScripts consumes any ^/_ suffixes that follow base and wraps it in the
// matching MathML script element. Both orders (x^2_i and x_i^2) collapse to a
// single <msubsup>, which is what TeX does too.
func (p *parser) parseScripts(base string) (string, error) {
	var sub, sup string
	primes := 0
	for {
		p.skipSpace()
		c := p.peek()
		if c == '\'' {
			// TeX prime shorthand: f' is f^{\prime}, f'' doubles it, and
			// f'^2 merges into one superscript (f^{\prime 2}).
			p.next()
			primes++
			continue
		}
		if c != '^' && c != '_' {
			break
		}
		p.next()
		arg, err := p.parseScriptArg()
		if err != nil {
			return "", err
		}
		if c == '^' {
			if sup != "" {
				return "", fmt.Errorf("double superscript")
			}
			sup = arg
		} else {
			if sub != "" {
				return "", fmt.Errorf("double subscript")
			}
			sub = arg
		}
	}
	if primes > 0 {
		pr := "<mo>" + primeRun(primes) + "</mo>"
		if sup == "" {
			sup = pr
		} else {
			sup = "<mrow>" + pr + sup + "</mrow>"
		}
	}
	// Display-style limit operators (∑, lim, …) stack their scripts
	// above/below via munderover; TeX calls this movablelimits. Inline
	// they keep side scripts. Integrals are deliberately not in the set:
	// TeX treats \int as nolimits, so msubsup side scripts are correct
	// in both modes.
	if p.display && limitBases[base] {
		switch {
		case sub != "" && sup != "":
			return "<munderover>" + base + sub + sup + "</munderover>", nil
		case sup != "":
			return "<mover>" + base + sup + "</mover>", nil
		case sub != "":
			return "<munder>" + base + sub + "</munder>", nil
		}
	}
	switch {
	case sub != "" && sup != "":
		return "<msubsup>" + base + sub + sup + "</msubsup>", nil
	case sup != "":
		return "<msup>" + base + sup + "</msup>", nil
	case sub != "":
		return "<msub>" + base + sub + "</msub>", nil
	default:
		return base, nil
	}
}

// limitBases is the set of operator atoms that take movable limits: their
// scripts stack in display style. Keyed by the emitted MathML so brace
// groups and macros compare equal.
var limitBases = map[string]bool{
	"<mo>∑</mo>":            true,
	"<mo>∏</mo>":            true,
	"<mo>∐</mo>":            true,
	"<mo>⋃</mo>":            true,
	"<mo>⋂</mo>":            true,
	"<mo>⋁</mo>":            true,
	"<mo>⋀</mo>":            true,
	"<mo>⨁</mo>":            true,
	"<mo>⨂</mo>":            true,
	"<mo>lim</mo>":          true,
	"<mo>lim\u2009sup</mo>": true,
	"<mo>lim\u2009inf</mo>": true,
	"<mo>max</mo>":          true,
	"<mo>min</mo>":          true,
	"<mo>sup</mo>":          true,
	"<mo>inf</mo>":          true,
}

// parseScriptArg reads the single atom that a ^ or _ applies to: either a
// brace group {…} or one bare token. Per TeX, a bare script argument is one
// token only — x^12 is x^{1}2, not x^{12} — so parseSingleToken is used rather
// than parseAtomBase (which would greedily swallow the whole digit run).
func (p *parser) parseScriptArg() (string, error) {
	return p.parseSingleToken()
}

// parseSingleToken reads exactly one TeX token as a MathML node: a brace group
// {…}, a macro, or a single character (one digit, one letter, one operator).
// It is the bare-argument reader for scripts and for required arguments given
// without braces (e.g. \frac12).
func (p *parser) parseSingleToken() (string, error) {
	p.skipSpace()
	if p.atEnd() {
		return "", fmt.Errorf("missing argument")
	}
	c := p.peek()
	switch {
	case c == '{':
		return p.parseAtomBase() // brace group handled there
	case c == '\\':
		return p.parseMacro()
	case c >= '0' && c <= '9':
		p.next()
		return "<mn>" + string(c) + "</mn>", nil
	case isLetter(c):
		p.next()
		return "<mi>" + string(c) + "</mi>", nil
	default:
		p.next()
		return "<mo>" + escapeXML(string(c)) + "</mo>", nil
	}
}

// parseAtomBase parses exactly one base atom without trailing scripts: a brace
// group, a macro, a number run, a single letter, or an operator/punctuation
// character. Whitespace returns "" so the caller skips it.
func (p *parser) parseAtomBase() (string, error) {
	p.skipSpace()
	if p.atEnd() {
		return "", nil
	}
	c := p.peek()
	switch {
	case c == '{':
		p.next()
		items, err := p.parseList(false)
		if err != nil {
			return "", err
		}
		if p.peek() != '}' {
			return "", fmt.Errorf("missing }")
		}
		p.next()
		return wrapRow(items), nil
	case c == '\\':
		return p.parseMacro()
	case c >= '0' && c <= '9':
		start := p.pos
		for !p.atEnd() && ((p.peek() >= '0' && p.peek() <= '9') || p.peek() == '.') {
			p.next()
		}
		return "<mn>" + string(p.src[start:p.pos]) + "</mn>", nil
	case isLetter(c):
		p.next()
		return "<mi>" + string(c) + "</mi>", nil
	default:
		p.next()
		return "<mo>" + escapeXML(string(c)) + "</mo>", nil
	}
}

// parseMacro handles a backslash command. The leading backslash is consumed by
// the caller's peek; here we read the command name (a run of letters, or a
// single non-letter for commands like \, or \{).
func (p *parser) parseMacro() (string, error) {
	p.next() // consume backslash
	if p.atEnd() {
		return "", fmt.Errorf("trailing backslash")
	}
	// Single non-letter command, e.g. \{ \} \, \; \  (escaped delimiters/spaces).
	if !isLetter(p.peek()) {
		c := p.next()
		switch c {
		case ' ', ',', ';', ':', '!':
			return "", nil // spacing commands: no visual atom in this subset
		case '{', '}', '%', '$', '#', '&', '_':
			return "<mo>" + escapeXML(string(c)) + "</mo>", nil
		default:
			return "<mo>" + escapeXML(string(c)) + "</mo>", nil
		}
	}
	start := p.pos
	for !p.atEnd() && isLetter(p.peek()) {
		p.next()
	}
	name := string(p.src[start:p.pos])

	switch name {
	case "frac":
		num, err := p.parseRequiredArg()
		if err != nil {
			return "", err
		}
		den, err := p.parseRequiredArg()
		if err != nil {
			return "", err
		}
		return "<mfrac>" + num + den + "</mfrac>", nil
	case "sqrt":
		// Optional index: \sqrt[n]{x} → <mroot>x n</mroot>.
		p.skipSpace()
		if p.peek() == '[' {
			idx, err := p.parseOptionalArg()
			if err != nil {
				return "", err
			}
			rad, err := p.parseRequiredArg()
			if err != nil {
				return "", err
			}
			return "<mroot>" + rad + idx + "</mroot>", nil
		}
		rad, err := p.parseRequiredArg()
		if err != nil {
			return "", err
		}
		return "<msqrt>" + rad + "</msqrt>", nil
	case "left", "right", "bigl", "bigr", "Bigl", "Bigr":
		// Delimiter sizing is ignored in this subset; emit the following
		// delimiter as a plain operator. \left. / \right. are invisible.
		p.skipSpace()
		if p.atEnd() {
			return "", nil
		}
		d := p.next()
		if d == '.' {
			return "", nil
		}
		if d == '\\' {
			// e.g. \left\{ — re-read as a macro delimiter.
			p.pos-- // step back onto backslash
			return p.parseMacro()
		}
		return "<mo>" + escapeXML(string(d)) + "</mo>", nil
	case "mathrm", "mathbf", "mathit", "text", "operatorname", "mathsf", "mathtt":
		// Take the group's textual content as one upright identifier. A
		// multi-char <mi> is upright by MathML convention, which the reader
		// honours, so named operators like \text{if} render correctly.
		txt := p.readGroupText()
		if txt == "" {
			return "", nil
		}
		return "<mi>" + escapeXML(txt) + "</mi>", nil
	}

	// Symbol or named-operator lookup.
	if sym, ok := symbols[name]; ok {
		return sym, nil
	}

	// Unknown macro: degrade to its literal name as an upright identifier so a
	// single unsupported command never aborts the whole render.
	return "<mi>" + escapeXML(name) + "</mi>", nil
}

// parseRequiredArg reads a mandatory {…} (or single-token) argument and returns
// it as a single MathML node (wrapped in <mrow> when it has several children).
func (p *parser) parseRequiredArg() (string, error) {
	p.skipSpace()
	if p.atEnd() {
		return "", fmt.Errorf("missing argument")
	}
	if p.peek() == '{' {
		p.next()
		items, err := p.parseList(false)
		if err != nil {
			return "", err
		}
		if p.peek() != '}' {
			return "", fmt.Errorf("missing }")
		}
		p.next()
		return wrapRow(items), nil
	}
	// Bare single-token argument, e.g. \frac12 == \frac{1}{2}.
	return p.parseSingleToken()
}

// parseOptionalArg reads a bracketed [...] argument (used by \sqrt[n]).
func (p *parser) parseOptionalArg() (string, error) {
	p.next() // consume '['
	var items []string
	for {
		p.skipSpace()
		if p.atEnd() {
			return "", fmt.Errorf("missing ]")
		}
		if p.peek() == ']' {
			p.next()
			break
		}
		base, err := p.parseAtomBase()
		if err != nil {
			return "", err
		}
		if base == "" {
			continue
		}
		base, err = p.parseScripts(base)
		if err != nil {
			return "", err
		}
		items = append(items, base)
	}
	return wrapRow(items), nil
}

// readGroupText returns the raw textual content of a {…} group, ignoring any
// markup — used for \text/\mathrm where we only want the letters.
func (p *parser) readGroupText() string {
	p.skipSpace()
	if p.peek() != '{' {
		// single token
		if p.atEnd() {
			return ""
		}
		return string(p.next())
	}
	p.next() // consume '{'
	start := p.pos
	depth := 1
	for !p.atEnd() && depth > 0 {
		switch p.peek() {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				txt := string(p.src[start:p.pos])
				p.next() // consume '}'
				return strings.TrimSpace(txt)
			}
		}
		p.next()
	}
	return strings.TrimSpace(string(p.src[start:p.pos]))
}

// wrapRow returns a single MathML node for a list of atoms: the atom itself
// when there is exactly one, an <mrow> wrapper for several, and an empty
// <mrow/> for none (so a slot is never structurally absent).
func wrapRow(items []string) string {
	switch len(items) {
	case 0:
		return "<mrow></mrow>"
	case 1:
		return items[0]
	default:
		return "<mrow>" + strings.Join(items, "") + "</mrow>"
	}
}

func isLetter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// escapeXML escapes the five XML predefined entities so operator characters
// like < > & survive serialisation into the MathML string.
func escapeXML(s string) string {
	if !strings.ContainsAny(s, `<>&"'`) {
		return s
	}
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

// primeRun returns the Unicode character for a run of n TeX prime marks:
// the dedicated double/triple/quadruple prime characters when they exist,
// a repetition of U+2032 beyond that.
func primeRun(n int) string {
	switch n {
	case 1:
		return "′"
	case 2:
		return "″"
	case 3:
		return "‴"
	case 4:
		return "⁗"
	}
	return strings.Repeat("′", n)
}
