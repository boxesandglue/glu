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
		"{a",       // unclosed brace
		"a}",       // stray close brace
		"\\frac{a}", // missing second fraction argument
		"x^2^3",    // double superscript
	}
	for _, tc := range cases {
		if _, err := ToMathML(tc, false); err == nil {
			t.Errorf("ToMathML(%q): expected error, got nil", tc)
		}
	}
}
