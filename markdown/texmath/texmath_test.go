package texmath

import "testing"

func TestToMathML(t *testing.T) {
	cases := []struct {
		name    string
		tex     string
		display bool
		want    string
	}{
		{
			name: "single variable italic",
			tex:  "x",
			want: "<math><mi>x</mi></math>",
		},
		{
			name: "number stays upright",
			tex:  "42",
			want: "<math><mn>42</mn></math>",
		},
		{
			name: "decimal number is one token",
			tex:  "3.14",
			want: "<math><mn>3.14</mn></math>",
		},
		{
			name: "superscript",
			tex:  "x^2",
			want: "<math><msup><mi>x</mi><mn>2</mn></msup></math>",
		},
		{
			name: "subscript",
			tex:  "a_i",
			want: "<math><msub><mi>a</mi><mi>i</mi></msub></math>",
		},
		{
			name: "sub and sup collapse to msubsup regardless of order",
			tex:  "x_i^2",
			want: "<math><msubsup><mi>x</mi><mi>i</mi><mn>2</mn></msubsup></math>",
		},
		{
			name: "braced exponent becomes mrow",
			tex:  "e^{x+1}",
			want: "<math><msup><mi>e</mi><mrow><mi>x</mi><mo>+</mo><mn>1</mn></mrow></msup></math>",
		},
		{
			name: "Einstein inline",
			tex:  "E=mc^2",
			want: "<math><mrow><mi>E</mi><mo>=</mo><mi>m</mi><msup><mi>c</mi><mn>2</mn></msup></mrow></math>",
		},
		{
			name: "fraction",
			tex:  "\\frac{a}{b}",
			want: "<math><mfrac><mi>a</mi><mi>b</mi></mfrac></math>",
		},
		{
			name: "fraction with bare args",
			tex:  "\\frac12",
			want: "<math><mfrac><mn>1</mn><mn>2</mn></mfrac></math>",
		},
		{
			name: "square root",
			tex:  "\\sqrt{2}",
			want: "<math><msqrt><mn>2</mn></msqrt></math>",
		},
		{
			name: "nth root",
			tex:  "\\sqrt[3]{x}",
			want: "<math><mroot><mi>x</mi><mn>3</mn></mroot></math>",
		},
		{
			name: "greek letter",
			tex:  "\\alpha",
			want: "<math><mi>α</mi></math>",
		},
		{
			name: "named operator",
			tex:  "\\sin",
			want: "<math><mi>sin</mi></math>",
		},
		{
			name:    "display flag sets attribute",
			tex:     "x",
			display: true,
			want:    `<math display="block"><mi>x</mi></math>`,
		},
		{
			name: "less-than is escaped",
			tex:  "a<b",
			want: "<math><mrow><mi>a</mi><mo>&lt;</mo><mi>b</mi></mrow></math>",
		},
		{
			name: "quadratic formula",
			tex:  "x=\\frac{-b\\pm\\sqrt{b^2-4ac}}{2a}",
			want: "<math><mrow><mi>x</mi><mo>=</mo><mfrac><mrow><mo>-</mo><mi>b</mi><mo>±</mo><msqrt><mrow><msup><mi>b</mi><mn>2</mn></msup><mo>-</mo><mn>4</mn><mi>a</mi><mi>c</mi></mrow></msqrt></mrow><mrow><mn>2</mn><mi>a</mi></mrow></mfrac></mrow></math>",
		},
		{
			name: "unknown macro degrades to identifier",
			tex:  "\\foobar",
			want: "<math><mi>foobar</mi></math>",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToMathML(tc.tex, tc.display)
			if err != nil {
				t.Fatalf("ToMathML(%q) error: %v", tc.tex, err)
			}
			if got != tc.want {
				t.Errorf("ToMathML(%q)\n got: %s\nwant: %s", tc.tex, got, tc.want)
			}
		})
	}
}

func TestToMathMLErrors(t *testing.T) {
	cases := []string{
		"{a",        // unclosed brace
		"a}",        // stray close brace
		"\\frac{a}", // missing second fraction argument
		"x^2^3",     // double superscript
	}
	for _, tc := range cases {
		if _, err := ToMathML(tc, false); err == nil {
			t.Errorf("ToMathML(%q): expected error, got nil", tc)
		}
	}
}

func TestFloorCeilDelimiters(t *testing.T) {
	testdata := []struct{ in, want string }{
		{`\lfloor x \rfloor`, "<math><mrow><mo>⌊</mo><mi>x</mi><mo>⌋</mo></mrow></math>"},
		{`\lceil x \rceil`, "<math><mrow><mo>⌈</mo><mi>x</mi><mo>⌉</mo></mrow></math>"},
		{`\langle a \rangle`, "<math><mrow><mo>⟨</mo><mi>a</mi><mo>⟩</mo></mrow></math>"},
	}
	for _, td := range testdata {
		got, err := ToMathML(td.in, false)
		if err != nil {
			t.Fatalf("ToMathML(%q): %v", td.in, err)
		}
		if got != td.want {
			t.Errorf("ToMathML(%q) = %s, want %s", td.in, got, td.want)
		}
	}
}

func TestPrimeShorthand(t *testing.T) {
	testdata := []struct{ in, want string }{
		{`f'`, "<math><msup><mi>f</mi><mo>′</mo></msup></math>"},
		{`f''`, "<math><msup><mi>f</mi><mo>″</mo></msup></math>"},
		{`f'''`, "<math><msup><mi>f</mi><mo>‴</mo></msup></math>"},
		{`f'(x)`, "<math><mrow><msup><mi>f</mi><mo>′</mo></msup><mo>(</mo><mi>x</mi><mo>)</mo></mrow></math>"},
		{`f'^2`, "<math><msup><mi>f</mi><mrow><mo>′</mo><mn>2</mn></mrow></msup></math>"},
		{`a_i'`, "<math><msubsup><mi>a</mi><mi>i</mi><mo>′</mo></msubsup></math>"},
	}
	for _, td := range testdata {
		got, err := ToMathML(td.in, false)
		if err != nil {
			t.Fatalf("ToMathML(%q): %v", td.in, err)
		}
		if got != td.want {
			t.Errorf("ToMathML(%q) = %s, want %s", td.in, got, td.want)
		}
	}
}

// TestDisplayLimitOperators — display math stacks scripts of movable-limit
// operators via munderover; inline keeps side scripts; integrals keep side
// scripts in both modes (TeX nolimits).
func TestDisplayLimitOperators(t *testing.T) {
	testdata := []struct {
		in      string
		display bool
		want    string
	}{
		{`\sum_{i=1}^{n}`, true, "<math display=\"block\"><munderover><mo>∑</mo><mrow><mi>i</mi><mo>=</mo><mn>1</mn></mrow><mi>n</mi></munderover></math>"},
		{`\sum_{i=1}^{n}`, false, "<math><msubsup><mo>∑</mo><mrow><mi>i</mi><mo>=</mo><mn>1</mn></mrow><mi>n</mi></msubsup></math>"},
		{`\int_a^b`, true, "<math display=\"block\"><msubsup><mo>∫</mo><mi>a</mi><mi>b</mi></msubsup></math>"},
		{`\lim_{x \to 0}`, true, "<math display=\"block\"><munder><mo>lim</mo><mrow><mi>x</mi><mo>→</mo><mn>0</mn></mrow></munder></math>"},
	}
	for _, td := range testdata {
		got, err := ToMathML(td.in, td.display)
		if err != nil {
			t.Fatalf("ToMathML(%q): %v", td.in, err)
		}
		if got != td.want {
			t.Errorf("ToMathML(%q, display=%v) =\n%s, want\n%s", td.in, td.display, got, td.want)
		}
	}
}
