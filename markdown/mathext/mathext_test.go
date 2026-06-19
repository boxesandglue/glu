package mathext

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func render(t *testing.T, md string) string {
	t.Helper()
	gm := goldmark.New(goldmark.WithExtensions(Math))
	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		t.Fatalf("convert %q: %v", md, err)
	}
	return buf.String()
}

func TestInlineMath(t *testing.T) {
	out := render(t, "energy is $E=mc^2$ today")
	if !strings.Contains(out, "<math><mrow><mi>E</mi><mo>=</mo>") {
		t.Errorf("inline math not converted:\n%s", out)
	}
	if strings.Contains(out, `display="block"`) {
		t.Errorf("inline math should not be display:\n%s", out)
	}
}

func TestDisplayMath(t *testing.T) {
	out := render(t, "see $$x^2+1$$ here")
	if !strings.Contains(out, `<math display="block">`) {
		t.Errorf("display math not converted:\n%s", out)
	}
}

func TestCurrencyIsNotMath(t *testing.T) {
	// Strict closing rule: "$5 and $6" — closing $ followed by a digit, and
	// opening $ followed by a digit then space; must stay literal text.
	out := render(t, "it costs $5 and $6 total")
	if strings.Contains(out, "<math") {
		t.Errorf("currency wrongly parsed as math:\n%s", out)
	}
	if !strings.Contains(out, "$5") || !strings.Contains(out, "$6") {
		t.Errorf("dollar signs lost:\n%s", out)
	}
}

func TestOpeningSpaceIsNotMath(t *testing.T) {
	out := render(t, "a $ b $ c")
	if strings.Contains(out, "<math") {
		t.Errorf("whitespace after opening $ should disable math:\n%s", out)
	}
}

func TestDollarInCodeSpanIsLiteral(t *testing.T) {
	// A real inline parser must not see dollars inside a code span.
	out := render(t, "use `$x$` verbatim and $y^2$ as math")
	if strings.Contains(out, "<math><mi>x</mi>") {
		t.Errorf("code-span dollars wrongly parsed as math:\n%s", out)
	}
	if !strings.Contains(out, "<math><msup><mi>y</mi>") {
		t.Errorf("real math after code span not parsed:\n%s", out)
	}
}

func TestUnclosedDollarIsLiteral(t *testing.T) {
	out := render(t, "a lone $ sign")
	if strings.Contains(out, "<math") {
		t.Errorf("lone dollar wrongly parsed:\n%s", out)
	}
}
