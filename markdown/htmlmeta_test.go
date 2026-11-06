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
			if got := extractHTMLMeta(tc.in, "pdf-format"); got != tc.want {
				t.Errorf("extractHTMLMeta(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractHTMLLang(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"html", `<!DOCTYPE html><html lang="de"><head></head></html>`, "de"},
		{"single quotes", `<html lang='fr'>`, "fr"},
		{"unquoted", `<html lang=en-US class=x>`, "en-US"},
		{"among attributes", `<html class="a" lang="de" dir="ltr">`, "de"},
		{"case-insensitive", `<HTML LANG="de">`, "de"},
		{"xml:lang", `<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="de">`, "de"},
		{"lang wins over xml:lang", `<html xml:lang="fr" lang="de">`, "de"},
		{"body as fallback", `<html><body lang="de"><p>x</p></body></html>`, "de"},
		{"html wins over body", `<html lang="de"><body lang="fr"></body></html>`, "de"},
		{"data attribute ignored", `<html data-lang="de">`, ""},
		{"later element ignored", `<html><body><p lang="fr">x</p></body></html>`, ""},
		{"hgroup is no html element", `<hgroup lang="fr"></hgroup>`, ""},
		{"empty", `<html lang="">`, ""},
		{"none", `<html><head></head></html>`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTMLLang(tc.in); got != tc.want {
				t.Errorf("extractHTMLLang(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractHTMLMetaAuthor(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"basic", `<meta name="author" content="Jane Doe">`, "Jane Doe"},
		{"content before name", `<meta content="Jane Doe" name="author">`, "Jane Doe"},
		{"entities decoded", `<meta name="author" content="Smith &amp; Sons">`, "Smith & Sons"},
		{"case-insensitive", `<META NAME="AUTHOR" CONTENT="Jane Doe">`, "Jane Doe"},
		{"other meta ignored", `<meta name="pdf-format" content="PDF/UA">`, ""},
		{"none", `<head><title>x</title></head>`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTMLMeta(tc.in, "author"); got != tc.want {
				t.Errorf("extractHTMLMeta(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
