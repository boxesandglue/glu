package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	luafrontend "github.com/boxesandglue/glu/lua/frontend"
	"github.com/speedata/go-lua"
)

// renderWithFlag renders a document with a --css file and returns the PDF.
func renderWithFlag(t *testing.T, ext, doc, flagCSS string) []byte {
	t.Helper()
	dir := t.TempDir()
	name := filepath.Join(dir, "doc"+ext)
	cssFile := filepath.Join(dir, "flag.css")
	for file, content := range map[string]string{name: doc, cssFile: flagCSS} {
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	output := filepath.Join(dir, "doc.pdf")
	opts := Options{
		OutputPath: output,
		MaxPasses:  3,
		CSSFile:    cssFile,
		SetupLua: func(l *lua.State) {
			lua.OpenLibraries(l)
			luafrontend.Open(l)
		},
	}
	var err error
	if ext == ".md" {
		err = ProcessFile(name, opts)
	} else {
		err = ProcessHTMLFile(name, opts)
	}
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// The --css file comes last: it overrides the stylesheets of the document,
// in rules as well as in @page. The size shows up uncompressed in /MediaBox.
func TestCSSFlagOverridesTheDocument(t *testing.T) {
	flag := "@page { size: 100pt 200pt; margin: 10pt; }\n"
	for _, tc := range []struct{ name, ext, doc string }{
		{"html style", ".html", `<!DOCTYPE html><html><head><style>@page { size: 300pt 400pt; }</style></head><body><p>x</p></body></html>`},
		{"markdown raw style", ".md", "<style>@page { size: 300pt 400pt; }</style>\n\nx\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pdf := renderWithFlag(t, tc.ext, tc.doc, flag)
			if !bytes.Contains(pdf, []byte("/MediaBox [0 0 100 200]")) {
				t.Errorf("the page does not have the size from --css")
			}
		})
	}
}
