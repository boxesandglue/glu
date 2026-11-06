package markdown

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boxesandglue/glu/internal/errkind"
	luafrontend "github.com/boxesandglue/glu/lua/frontend"
	"github.com/speedata/go-lua"
)

// renderWithCompanion renders a one paragraph document with the given
// companion Lua file and reports the error and whether a PDF exists.
func renderWithCompanion(t *testing.T, ext, doc, companion string) (error, bool) {
	t.Helper()
	dir := t.TempDir()
	name := filepath.Join(dir, "doc"+ext)
	if err := os.WriteFile(name, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "doc.lua"), []byte(companion), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "doc.pdf")
	opts := Options{
		OutputPath: output,
		MaxPasses:  3,
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
	_, statErr := os.Stat(output)
	return err, statErr == nil
}

// A failed run must not leave a file at the output path. Creating the document
// creates the file, so what stays behind would be the start of a PDF.
func TestFailedRunLeavesNoFile(t *testing.T) {
	err, hasPDF := renderWithCompanion(t, ".html", "<!DOCTYPE html><html><head><link rel=\"stylesheet\" href=\"missing.css\"></head><body><p>x</p></body></html>", "")
	if err == nil {
		t.Fatal("want an error for the missing stylesheet")
	}
	if hasPDF {
		t.Error("a file stayed at the output path")
	}
}

// A callback that fails must fail the run. The hooks of bag and htmlbag
// cannot return an error, so a failure used to end up in the log only: the
// PDF came out without what the callback was meant to do, with exit code 0.
func TestCallbackErrorFailsTheRun(t *testing.T) {
	for _, event := range []string{"page_init", "post_element", "pre_shipout"} {
		for _, input := range []struct{ ext, doc string }{
			{".html", "<!DOCTYPE html><html><body><p>x</p></body></html>"},
			{".md", "x\n"},
		} {
			t.Run(event+input.ext, func(t *testing.T) {
				companion := fmt.Sprintf(`local frontend = require("glu.frontend")
frontend.add_callback(%q, "fails", function() error("broken callback") end)
`, event)
				err, hasPDF := renderWithCompanion(t, input.ext, input.doc, companion)
				if err == nil {
					t.Fatal("want an error, got none")
				}
				if !errors.Is(err, errkind.Lua) {
					t.Errorf("error %q is not a Lua error", err)
				}
				if !strings.Contains(err.Error(), "broken callback") {
					t.Errorf("error %q does not name the cause", err)
				}
				if hasPDF {
					t.Error("a PDF was written in spite of the failure")
				}
			})
		}
	}
}

func TestWorkingCallbackStillRenders(t *testing.T) {
	companion := `local frontend = require("glu.frontend")
frontend.add_callback("page_init", "works", function() end)
`
	err, hasPDF := renderWithCompanion(t, ".html", "<!DOCTYPE html><html><body><p>x</p></body></html>", companion)
	if err != nil {
		t.Fatal(err)
	}
	if !hasPDF {
		t.Error("no PDF written")
	}
}
