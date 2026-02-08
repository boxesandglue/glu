package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/csshtml"
	"github.com/boxesandglue/htmlbag"
	"github.com/speedata/go-lua"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"

	luafrontend "github.com/speedata/glu/lua/frontend"
)

// auxHeading stores a heading entry in the aux file.
type auxHeading struct {
	Level string `json:"level"`
	Text  string `json:"text"`
	Page  int    `json:"page"`
}

// auxData holds cross-run information that requires multiple passes.
type auxData struct {
	Pages    int          `json:"pages"`
	Headings []auxHeading `json:"headings,omitempty"`
}

// readAuxFile reads a previously written aux file. If the file does not exist
// it returns zero-valued auxData and no error.
func readAuxFile(auxPath string) (auxData, error) {
	var ad auxData
	data, err := os.ReadFile(auxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ad, nil
		}
		return ad, err
	}
	if err := json.Unmarshal(data, &ad); err != nil {
		return ad, fmt.Errorf("parsing %s: %w", auxPath, err)
	}
	return ad, nil
}

// writeAuxFile writes aux data to disk and returns true if any value changed
// compared to the old data (i.e. a rerun is needed).
func writeAuxFile(auxPath string, old, cur auxData) (bool, error) {
	data, err := json.Marshal(cur)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(auxPath, data, 0644); err != nil {
		return false, err
	}
	changed := old.Pages != cur.Pages || len(old.Headings) != len(cur.Headings)
	if !changed {
		for i := range old.Headings {
			if old.Headings[i] != cur.Headings[i] {
				changed = true
				break
			}
		}
	}
	return changed, nil
}

// setTOCGlobal pushes the _toc Lua global from the aux heading data.
func setTOCGlobal(l *lua.State, headings []auxHeading) {
	l.NewTable()
	for i, h := range headings {
		l.NewTable()
		l.PushString(h.Level)
		l.SetField(-2, "level")
		l.PushString(h.Text)
		l.SetField(-2, "text")
		l.PushInteger(h.Page)
		l.SetField(-2, "page")
		l.RawSetInt(-2, i+1)
	}
	l.SetGlobal("_toc")
}

// Options controls the Markdown processing pipeline.
type Options struct {
	Template      bool   // apply Go template expansion before processing
	CSSFile       string // additional CSS file to load
	DebugMarkdown bool   // print expanded Markdown to stdout instead of generating PDF
	DebugHTML     bool   // print generated HTML to stdout instead of generating PDF
}

// ProcessFile reads a Markdown file and produces a PDF.
func ProcessFile(l *lua.State, filename string, opts Options) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}
	source := string(data)

	// Step 1: Optional Go template expansion
	if opts.Template {
		tmpl, err := template.New(filepath.Base(filename)).Parse(source)
		if err != nil {
			return fmt.Errorf("template parse: %w", err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			return fmt.Errorf("template execute: %w", err)
		}
		source = buf.String()
	}

	// Step 2: Extract YAML front matter
	fm, body := extractFrontmatter(source)
	slog.Debug("Frontmatter", "title", fm.Title, "author", fm.Author, "papersize", fm.Papersize, "css", fm.CSS)

	// Read aux data early so _toc is available to Lua blocks.
	outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".pdf"
	auxPath := strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename)) + ".aux"
	earlyAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("reading aux file: %w", err)
	}
	setTOCGlobal(l, earlyAux.Headings)

	// Step 3: Load companion Lua file (e.g. example.lua for example.md)
	luaFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".lua"
	if _, err := os.Stat(luaFile); err == nil {
		slog.Info("Loading companion Lua file", "file", luaFile)
		if err := lua.DoFile(l, luaFile); err != nil {
			return fmt.Errorf("companion lua file %s: %w", luaFile, err)
		}
	}

	// Step 4: Execute Lua blocks
	body, err = extractAndRunLuaBlocks(l, body)
	if err != nil {
		return err
	}

	// Step 4: Expand inline expressions
	body, err = expandInlineExpressions(l, body)
	if err != nil {
		return err
	}

	// Debug mode: print expanded Markdown and exit
	if opts.DebugMarkdown {
		fmt.Print(body)
		return nil
	}

	// Step 5: Markdown → HTML
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)
	var htmlBuf bytes.Buffer
	if err := gm.Convert([]byte(body), &htmlBuf); err != nil {
		return fmt.Errorf("goldmark convert: %w", err)
	}
	htmlStr := htmlBuf.String()
	slog.Debug("HTML generated", "length", len(htmlStr))

	// Debug mode: print generated HTML and exit
	if opts.DebugHTML {
		fmt.Print(htmlStr)
		return nil
	}

	// Step 6: HTML → PDF via htmlbag
	return generatePDF(l, outputFilename, htmlStr, fm, opts)
}

// ProcessHTMLFile reads an HTML file and produces a PDF.
// Unlike Markdown mode, no default CSS is applied — styling comes from
// <link>, <style>, inline styles, or the --css flag.
func ProcessHTMLFile(l *lua.State, filename string, opts Options) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}
	htmlStr := string(data)

	// Load companion Lua file (e.g. page.lua for page.html)
	luaFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".lua"
	if _, err := os.Stat(luaFile); err == nil {
		slog.Info("Loading companion Lua file", "file", luaFile)
		if err := lua.DoFile(l, luaFile); err != nil {
			return fmt.Errorf("companion lua file %s: %w", luaFile, err)
		}
	}

	outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".pdf"

	// Read aux data from a previous run (for forward references like counter(pages)).
	auxPath := strings.TrimSuffix(outputFilename, ".pdf") + ".aux"
	oldAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("reading aux file: %w", err)
	}

	// Set _toc global from previous run's heading data.
	setTOCGlobal(l, oldAux.Headings)

	fe, err := frontend.New(outputFilename)
	if err != nil {
		return fmt.Errorf("creating document: %w", err)
	}

	cssParser := csshtml.NewCSSParserWithDefaults()
	// Set the directory so relative CSS paths (in <link> or @import) resolve correctly.
	abs, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	cssParser.PushDir(abs)

	cb, err := htmlbag.New(fe, cssParser)
	if err != nil {
		return fmt.Errorf("creating CSS builder: %w", err)
	}
	cb.Counters["pages"] = oldAux.Pages

	// Install callbacks
	if cr := luafrontend.GetRegistry(); cr != nil {
		cr.SetCSSBuilder(cb)
		cr.InstallPreShipout(fe)
		cr.InstallPostElement()
		cr.InstallPageInit()
	}

	// Load CSS from CLI flag
	if opts.CSSFile != "" {
		if err := cb.ReadCSSFile(opts.CSSFile); err != nil {
			return fmt.Errorf("reading CSS file %s: %w", opts.CSSFile, err)
		}
	}

	// Convert HTML to text elements first — this processes <link> and <style>
	// tags, making @page rules available for page initialization.
	te, err := cb.HTMLToText(htmlStr)
	if err != nil {
		return fmt.Errorf("HTML to text: %w", err)
	}

	// Initialize the page (now @page rules from <style> are available)
	pd, err := cb.PageSize()
	if err != nil {
		return fmt.Errorf("getting page size: %w", err)
	}

	// Build vertical list
	vl, err := cb.CreateVlist(te, pd.ContentWidth)
	if err != nil {
		return fmt.Errorf("creating vlist: %w", err)
	}

	// Distribute content across pages
	if err := cb.OutputPages(vl); err != nil {
		return fmt.Errorf("outputting pages: %w", err)
	}

	// Write aux file with values from this run.
	curAux := auxData{Pages: len(fe.Doc.Pages)}
	for _, h := range cb.Headings {
		curAux.Headings = append(curAux.Headings, auxHeading{Level: h.Level, Text: h.Text, Page: h.Page})
	}
	if changed, err := writeAuxFile(auxPath, oldAux, curAux); err != nil {
		return fmt.Errorf("writing aux file: %w", err)
	} else if changed {
		slog.Info("Aux data changed — rerun to update cross-references", "pages", curAux.Pages)
	}

	if err := fe.Finish(); err != nil {
		return fmt.Errorf("finishing document: %w", err)
	}

	slog.Info("PDF written", "file", outputFilename, "pages", len(fe.Doc.Pages))
	return nil
}

func generatePDF(l *lua.State, outputFilename string, htmlStr string, fm Frontmatter, opts Options) error {
	// Read aux data from a previous run (for forward references like counter(pages)).
	auxPath := strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename)) + ".aux"
	oldAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("reading aux file: %w", err)
	}

	// Set _toc global from previous run's heading data.
	setTOCGlobal(l, oldAux.Headings)

	fe, err := frontend.New(outputFilename)
	if err != nil {
		return fmt.Errorf("creating document: %w", err)
	}

	if fm.Title != "" {
		fe.Doc.Title = fm.Title
	}
	if fm.Author != "" {
		fe.Doc.Author = fm.Author
	}

	cssParser := csshtml.NewCSSParserWithDefaults()
	cb, err := htmlbag.New(fe, cssParser)
	if err != nil {
		return fmt.Errorf("creating CSS builder: %w", err)
	}
	cb.Counters["pages"] = oldAux.Pages

	// Install pre_shipout callback hook so Lua callbacks fire on Shipout.
	// Must happen after CSSBuilder creation so PageInfo is available.
	if cr := luafrontend.GetRegistry(); cr != nil {
		cr.SetCSSBuilder(cb)
		cr.InstallPreShipout(fe)
		cr.InstallPostElement()
		cr.InstallPageInit()
	}

	// Load default Markdown CSS
	if err := cb.AddCSS(defaultCSS); err != nil {
		return fmt.Errorf("adding default CSS: %w", err)
	}

	// Apply papersize from front matter
	if fm.Papersize != "" {
		pageSizeCSS := fmt.Sprintf("@page { size: %s; }", fm.Papersize)
		if err := cb.AddCSS(pageSizeCSS); err != nil {
			return fmt.Errorf("applying papersize: %w", err)
		}
	}

	// Load CSS from front matter
	if fm.CSS != "" {
		if err := cb.ReadCSSFile(fm.CSS); err != nil {
			return fmt.Errorf("reading CSS file %s: %w", fm.CSS, err)
		}
	}

	// Load CSS from CLI flag
	if opts.CSSFile != "" {
		if err := cb.ReadCSSFile(opts.CSSFile); err != nil {
			return fmt.Errorf("reading CSS file %s: %w", opts.CSSFile, err)
		}
	}

	// Initialize the page
	if err := cb.InitPage(); err != nil {
		return fmt.Errorf("initializing page: %w", err)
	}

	pd, err := cb.PageSize()
	if err != nil {
		return fmt.Errorf("getting page size: %w", err)
	}

	// Convert HTML to text elements
	te, err := cb.HTMLToText(htmlStr)
	if err != nil {
		return fmt.Errorf("HTML to text: %w", err)
	}

	// Build vertical list
	vl, err := cb.CreateVlist(te, pd.ContentWidth)
	if err != nil {
		return fmt.Errorf("creating vlist: %w", err)
	}

	// Distribute content across pages (with automatic page breaks)
	if err := cb.OutputPages(vl); err != nil {
		return fmt.Errorf("outputting pages: %w", err)
	}

	// Write aux file with values from this run.
	curAux := auxData{Pages: len(fe.Doc.Pages)}
	for _, h := range cb.Headings {
		curAux.Headings = append(curAux.Headings, auxHeading{Level: h.Level, Text: h.Text, Page: h.Page})
	}
	if changed, err := writeAuxFile(auxPath, oldAux, curAux); err != nil {
		return fmt.Errorf("writing aux file: %w", err)
	} else if changed {
		slog.Info("Aux data changed — rerun to update cross-references", "pages", curAux.Pages)
	}

	if err := fe.Finish(); err != nil {
		return fmt.Errorf("finishing document: %w", err)
	}

	slog.Info("PDF written", "file", outputFilename, "pages", len(fe.Doc.Pages))
	return nil
}
