package markdown

import "testing"

func TestExtractHTMLTitle(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"plain", `<html><head><title>Hello</title></head></html>`, "Hello"},
		{"whitespace trimmed", "<title>\n  Spaced  \n</title>", "Spaced"},
		{"entities decoded", `<title>A &amp; B</title>`, "A & B"},
		{"nested markup stripped", `<title>A <b>bold</b> title</title>`, "A bold title"},
		{"case-insensitive tag", `<TITLE>Caps</TITLE>`, "Caps"},
		{"none", `<html><head></head></html>`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTMLTitle(tc.in); got != tc.want {
				t.Errorf("extractHTMLTitle(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractHTMLMetaFormat(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"basic", `<meta name="pdf-format" content="PDF/UA-2">`, "PDF/UA-2"},
		{"single quotes", `<meta name='pdf-format' content='PDF/UA'>`, "PDF/UA"},
		{"content before name", `<meta content="PDF/A-3b" name="pdf-format">`, "PDF/A-3b"},
		{"case-insensitive", `<META NAME="pdf-format" CONTENT="PDF/UA-2">`, "PDF/UA-2"},
		{"other meta ignored", `<meta name="viewport" content="width=device-width">`, ""},
		{"none", `<head><title>x</title></head>`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTMLMetaFormat(tc.in); got != tc.want {
				t.Errorf("extractHTMLMetaFormat(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
