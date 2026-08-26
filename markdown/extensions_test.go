package markdown

import (
	"strings"
	"testing"
)

// TestExtensionListForms checks that the "extensions" frontmatter key
// accepts both a comma/space separated scalar and a YAML sequence.
func TestExtensionListForms(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want []string
	}{
		{"scalar commas", "extensions: smart, auto_identifiers, -footnotes", []string{"smart", "auto_identifiers", "-footnotes"}},
		{"scalar spaces", "extensions: smart -footnotes", []string{"smart", "-footnotes"}},
		{"sequence", "extensions:\n  - smart\n  - \"-footnotes\"", []string{"smart", "-footnotes"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fm, _ := ExtractFrontmatter("---\n" + c.yaml + "\n---\nbody\n")
			if len(fm.Extensions) != len(c.want) {
				t.Fatalf("got %v, want %v", fm.Extensions, c.want)
			}
			for i := range c.want {
				if fm.Extensions[i] != c.want[i] {
					t.Fatalf("got %v, want %v", fm.Extensions, c.want)
				}
			}
		})
	}
}

func TestResolveExtensions(t *testing.T) {
	fm := Frontmatter{Extensions: ExtensionList{"smart", "-footnotes", "+auto_identifiers"}}
	exts, err := resolveExtensions(fm)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{
		"smart":            true,
		"auto_identifiers": true,
		"footnotes":        false,
		"fenced_divs":      true,
		"bracketed_spans":  true,
		"definition_lists": true,
		"tex_math_dollars": false,
	} {
		if exts[name] != want {
			t.Errorf("extension %s: got %v, want %v", name, exts[name], want)
		}
	}

	// legacy math key
	exts, err = resolveExtensions(Frontmatter{Math: true})
	if err != nil {
		t.Fatal(err)
	}
	if !exts["tex_math_dollars"] {
		t.Error("math: true should enable tex_math_dollars")
	}

	// unknown names are an error
	if _, err = resolveExtensions(Frontmatter{Extensions: ExtensionList{"fancy_stuff"}}); err == nil {
		t.Error("unknown extension should be an error")
	}
}

// mustHTML converts body with the given frontmatter and fails the test on
// error.
func mustHTML(t *testing.T, body string, fm Frontmatter) string {
	t.Helper()
	htmlStr, err := markdownToHTML(body, fm)
	if err != nil {
		t.Fatal(err)
	}
	return htmlStr
}

func TestFencedDivs(t *testing.T) {
	fm := Frontmatter{}

	got := mustHTML(t, "::: {.note lang=en}\nEnglish content.\n:::\n", fm)
	if !strings.Contains(got, `<div class="note" lang="en">`) || !strings.Contains(got, "<p>English content.</p>") {
		t.Errorf("fenced div with attributes not rendered: %s", got)
	}

	got = mustHTML(t, "::: warning\nCareful.\n:::\n", fm)
	if !strings.Contains(got, `<div class="warning">`) {
		t.Errorf("bare word class not rendered: %s", got)
	}

	// nesting: the closing fence closes the innermost open div
	got = mustHTML(t, "::: {.outer}\nbefore\n\n::: {.inner}\ninside\n:::\n\nafter\n:::\n", fm)
	wantOrder := []string{`<div class="outer">`, "<p>before</p>", `<div class="inner">`, "<p>inside</p>", "</div>", "<p>after</p>", "</div>"}
	idx := 0
	for _, w := range wantOrder {
		p := strings.Index(got[idx:], w)
		if p < 0 {
			t.Fatalf("nested divs: %q not found after position %d in: %s", w, idx, got)
		}
		idx += p + len(w)
	}

	// decorative trailing colons on the opening fence
	got = mustHTML(t, "::: {.x} :::::\ncontent\n:::\n", fm)
	if !strings.Contains(got, `<div class="x">`) || strings.Contains(got, ":::") {
		t.Errorf("decorative colons not consumed: %s", got)
	}

	// a bare colon run without any open div stays paragraph text
	got = mustHTML(t, "just text\n\n:::\n", fm)
	if !strings.Contains(got, "<p>:::</p>") {
		t.Errorf("stray fence should stay literal: %s", got)
	}

	// disabled: the fences stay literal text
	got = mustHTML(t, "::: {.note}\ncontent\n:::\n", Frontmatter{Extensions: ExtensionList{"-fenced_divs"}})
	if strings.Contains(got, "<div") {
		t.Errorf("disabled fenced_divs still rendered a div: %s", got)
	}
}

func TestBracketedSpans(t *testing.T) {
	fm := Frontmatter{}

	got := mustHTML(t, "Deutscher Text mit [an *English* phrase]{lang=en .foreign} darin.\n", fm)
	if !strings.Contains(got, `<span lang="en" class="foreign">an <em>English</em> phrase</span>`) {
		t.Errorf("bracketed span not rendered: %s", got)
	}

	// nested spans
	got = mustHTML(t, "[outer [inner]{.i} rest]{.o}\n", fm)
	if !strings.Contains(got, `<span class="o">outer <span class="i">inner</span> rest</span>`) {
		t.Errorf("nested spans not rendered: %s", got)
	}

	// links inside a span, spans not interfering with plain links
	got = mustHTML(t, "[see [docs](https://example.com) here]{.ref}\n", fm)
	if !strings.Contains(got, `<span class="ref">see <a href="https://example.com">docs</a> here</span>`) {
		t.Errorf("link inside span not rendered: %s", got)
	}
	got = mustHTML(t, "a plain [link](https://example.com) stays\n", fm)
	if !strings.Contains(got, `<a href="https://example.com">link</a>`) || strings.Contains(got, "<span") {
		t.Errorf("plain link damaged: %s", got)
	}

	// bracket without attributes stays literal
	got = mustHTML(t, "not [a span] here\n", fm)
	if !strings.Contains(got, "not [a span] here") {
		t.Errorf("bracket without attributes changed: %s", got)
	}

	// disabled
	got = mustHTML(t, "no [span]{.x} here\n", Frontmatter{Extensions: ExtensionList{"-bracketed_spans"}})
	if strings.Contains(got, "<span") {
		t.Errorf("disabled bracketed_spans still rendered: %s", got)
	}
}

func TestFootnotesExtension(t *testing.T) {
	body := "Text with a note.[^1]\n\n[^1]: The note.\n"
	got := mustHTML(t, body, Frontmatter{})
	if !strings.Contains(got, "footnote") {
		t.Errorf("footnotes not rendered: %s", got)
	}
	got = mustHTML(t, body, Frontmatter{Extensions: ExtensionList{"-footnotes"}})
	if strings.Contains(got, `class="footnote`) {
		t.Errorf("disabled footnotes still rendered: %s", got)
	}
}

func TestDefinitionListsExtension(t *testing.T) {
	body := "Term\n: The definition.\n"
	got := mustHTML(t, body, Frontmatter{})
	if !strings.Contains(got, "<dl>") || !strings.Contains(got, "<dt>Term</dt>") {
		t.Errorf("definition list not rendered: %s", got)
	}
}

func TestSmartExtension(t *testing.T) {
	body := "Er sagte \"Hallo\" und ging.\n"

	// off by default
	got := mustHTML(t, body, Frontmatter{})
	if strings.Contains(got, "„") || strings.Contains(got, "&ldquo;") {
		t.Errorf("smart should be off by default: %s", got)
	}

	// German quotes with lang: de
	got = mustHTML(t, body, Frontmatter{Lang: "de", Extensions: ExtensionList{"smart"}})
	if !strings.Contains(got, "„Hallo“") {
		t.Errorf("German quotes not applied: %s", got)
	}

	// English default otherwise
	got = mustHTML(t, "He said \"hello\" and left.\n", Frontmatter{Lang: "en", Extensions: ExtensionList{"smart"}})
	if !strings.Contains(got, "&ldquo;hello&rdquo;") {
		t.Errorf("English quotes not applied: %s", got)
	}
}

func TestAutoIdentifiers(t *testing.T) {
	body := "# Über uns\n\n## Über uns\n"

	// off by default
	got := mustHTML(t, body, Frontmatter{})
	if strings.Contains(got, "id=") {
		t.Errorf("auto_identifiers should be off by default: %s", got)
	}

	got = mustHTML(t, body, Frontmatter{Extensions: ExtensionList{"auto_identifiers"}})
	if !strings.Contains(got, `<h1 id="über-uns">`) {
		t.Errorf("pandoc-style id missing: %s", got)
	}
	if !strings.Contains(got, `<h2 id="über-uns-1">`) {
		t.Errorf("duplicate id not disambiguated: %s", got)
	}

	// explicit {#id} still wins
	got = mustHTML(t, "# Title {#custom}\n", Frontmatter{Extensions: ExtensionList{"auto_identifiers"}})
	if !strings.Contains(got, `<h1 id="custom">`) {
		t.Errorf("explicit id overridden: %s", got)
	}
}
