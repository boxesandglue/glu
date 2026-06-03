package markdown

import (
	"testing"
)

func TestExtractFrontmatter_KnownKeys(t *testing.T) {
	src := `---
title: My Document
author: Jane Doe
lang: en
papersize: a4
format: PDF/UA
---
# Body
`
	fm, body := ExtractFrontmatter(src)
	if fm.Title != "My Document" {
		t.Errorf("Title: got %q", fm.Title)
	}
	if fm.Author != "Jane Doe" {
		t.Errorf("Author: got %q", fm.Author)
	}
	if fm.Lang != "en" {
		t.Errorf("Lang: got %q", fm.Lang)
	}
	if fm.Papersize != "a4" {
		t.Errorf("Papersize: got %q", fm.Papersize)
	}
	if fm.Format != "PDF/UA" {
		t.Errorf("Format: got %q", fm.Format)
	}
	if body != "# Body\n" {
		t.Errorf("body: got %q", body)
	}
}

func TestExtractFrontmatter_Attachments(t *testing.T) {
	src := `---
attachments:
  - file: spec.pdf
    name: Spezifikation.pdf
    description: technische Beilage
    mimetype: application/pdf
  - file: data.json
---
body
`
	fm, _ := ExtractFrontmatter(src)
	if len(fm.Attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(fm.Attachments))
	}
	a0 := fm.Attachments[0]
	if a0.File != "spec.pdf" || a0.Name != "Spezifikation.pdf" ||
		a0.Description != "technische Beilage" || a0.MimeType != "application/pdf" {
		t.Errorf("attachment[0] mismatch: %+v", a0)
	}
	a1 := fm.Attachments[1]
	if a1.File != "data.json" || a1.Name != "" || a1.MimeType != "" {
		t.Errorf("attachment[1] should have defaults empty (caller fills): %+v", a1)
	}
}

func TestExtractFrontmatter_NoFrontmatter(t *testing.T) {
	src := "# Title\n\nbody\n"
	fm, body := ExtractFrontmatter(src)
	if fm.Title != "" || len(fm.Attachments) != 0 {
		t.Errorf("expected empty frontmatter, got %+v", fm)
	}
	if body != src {
		t.Errorf("body should be source verbatim")
	}
}

func TestExtractFrontmatter_ExtraPopulated(t *testing.T) {
	src := `---
title: t
custom-key: custom-value
---
body
`
	fm, _ := ExtractFrontmatter(src)
	if v, ok := fm.Extra["custom-key"].(string); !ok || v != "custom-value" {
		t.Errorf("Extra[custom-key] = %v (%T)", fm.Extra["custom-key"], fm.Extra["custom-key"])
	}
	if v, ok := fm.Extra["title"].(string); !ok || v != "t" {
		t.Errorf("Extra should also carry known keys; title = %v", fm.Extra["title"])
	}
}
