package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/csshtml"
	"github.com/boxesandglue/htmlbag"
	"github.com/speedata/go-lua"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"

	luacommon "github.com/boxesandglue/glu/lua/common"
	luafrontend "github.com/boxesandglue/glu/lua/frontend"
)

// preTrailingNL matches a trailing newline (possibly inside a chroma
// whitespace span) just before the closing </code></pre> sequence.
var preTrailingNL = regexp.MustCompile(`(?:<span[^>]*>)?\n((?:</span>)*</code></pre>)`)

// readAuxFile reads a previously written aux file. If the file does not exist
// it returns an empty map and no error.
func readAuxFile(auxPath string) (map[string]any, error) {
	data, err := os.ReadFile(auxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", auxPath, err)
	}
	if m == nil {
		m = make(map[string]any)
	}
	return m, nil
}

// writeAuxFile writes aux data to disk and returns true if any value changed
// compared to the old data (i.e. a rerun is needed).
func writeAuxFile(auxPath string, old, cur map[string]any) (bool, error) {
	data, err := json.MarshalIndent(cur, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(auxPath, data, 0644); err != nil {
		return false, err
	}
	oldData, _ := json.MarshalIndent(old, "", "  ")
	return !bytes.Equal(oldData, data), nil
}

// fireCallback fires a lifecycle callback if the registry is available.
func fireCallback(event string) error {
	if cr := luafrontend.GetRegistry(); cr != nil {
		return cr.Fire(event, nil)
	}
	return nil
}

// setAuxGlobal sets the _aux Lua global from the aux map and also sets
// _toc to _aux._headings for backward compatibility.
func setAuxGlobal(l *lua.State, data map[string]any) {
	luacommon.PushAny(l, any(data))
	l.SetGlobal("_aux")

	// Backward compat: _toc = _aux._headings
	l.Global("_aux")
	l.Field(-1, "_headings")
	l.SetGlobal("_toc")
	l.Pop(1) // pop _aux
}

// readAuxGlobal reads the _aux Lua global back into a Go map.
func readAuxGlobal(l *lua.State) map[string]any {
	l.Global("_aux")
	defer l.Pop(1)
	if l.IsTable(-1) {
		if m, ok := luacommon.LuaToGo(l, -1).(map[string]any); ok {
			return m
		}
	}
	return make(map[string]any)
}

// Options controls the Markdown processing pipeline.
type Options struct {
	Template      bool   // apply Go template expansion before processing
	CSSFile       string // additional CSS file to load
	DebugMarkdown bool   // print expanded Markdown to stdout instead of generating PDF
	DebugHTML     bool   // print generated HTML to stdout instead of generating PDF
	// Format selects a PDF conformance level. "PDF/UA" enables the
	// accessibility tagging pipeline (StructTreeRoot, MarkInfo, role-mapped
	// element tree, XMP pdfuaid:part 1). Empty means a plain PDF.
	Format string
	// Lang is the BCP47 language tag written to the PDF catalog (/Lang) and
	// used as the document-wide hyphenation default. Required for PDF/UA.
	Lang string
	// Title becomes the document /Title (catalog + XMP dc:title). PDF/UA
	// also auto-enables /DisplayDocTitle so PDF readers show the title in
	// the window chrome instead of the filename.
	Title string
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

	// Make the entire frontmatter available as _frontmatter Lua global.
	luacommon.PushAny(l, any(fm.Extra))
	l.SetGlobal("_frontmatter")

	// Step 3: Load companion Lua file (e.g. example.lua for example.md)
	// Loaded before aux so that lifecycle callbacks can be registered.
	luaFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".lua"
	if _, err := os.Stat(luaFile); err == nil {
		slog.Info("Loading companion Lua file", "file", luaFile)
		if err := lua.DoFile(l, luaFile); err != nil {
			return fmt.Errorf("companion lua file %s: %w", luaFile, err)
		}
	}

	// Fire document_start callback (before aux file is loaded).
	if err := fireCallback("document_start"); err != nil {
		return fmt.Errorf("document_start callback: %w", err)
	}

	// Read aux data so _aux/_toc are available to Lua blocks.
	outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".pdf"
	auxPath := strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename)) + "-aux.json"
	earlyAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("reading aux file: %w", err)
	}
	setAuxGlobal(l, earlyAux)

	// Fire content_ready callback (aux loaded, before content processing).
	if err := fireCallback("content_ready"); err != nil {
		return fmt.Errorf("content_ready callback: %w", err)
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
	highlightStyle := "github"
	if s, ok := fm.Extra["highlight-style"].(string); ok {
		highlightStyle = s
	}
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			highlighting.NewHighlighting(
				highlighting.WithStyle(highlightStyle),
			),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)
	var htmlBuf bytes.Buffer
	if err := gm.Convert([]byte(body), &htmlBuf); err != nil {
		return fmt.Errorf("goldmark convert: %w", err)
	}
	// Strip trailing newlines inside <pre><code> blocks to avoid an extra
	// blank line at the bottom (chroma and goldmark both add one).
	htmlStr := preTrailingNL.ReplaceAllString(htmlBuf.String(), "$1")
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
	baseDir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	return ProcessHTMLString(l, htmlStr, baseDir, outputFilename, opts)
}

// ProcessHTMLString takes an HTML payload (already in memory), the directory
// to resolve relative CSS @import / <link> paths against, and an output PDF
// filename, and runs the same pipeline as ProcessHTMLFile.
//
// Useful for embedders that produce HTML on the fly (e.g. an XSL-FO walker)
// and want to skip the disk-roundtrip of writing the HTML out and reading
// it back in.
func ProcessHTMLString(l *lua.State, htmlStr, baseDir, outputFilename string, opts Options) error {
	// Fire document_start callback (before aux file is loaded).
	if err := fireCallback("document_start"); err != nil {
		return fmt.Errorf("document_start callback: %w", err)
	}

	// Read aux data from a previous run (for forward references like counter(pages)).
	auxPath := strings.TrimSuffix(outputFilename, ".pdf") + "-aux.json"
	oldAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("reading aux file: %w", err)
	}

	// Set _aux and _toc globals from previous run's data.
	setAuxGlobal(l, oldAux)

	fe, err := frontend.New(outputFilename)
	if err != nil {
		return fmt.Errorf("creating document: %w", err)
	}

	// Apply caller-supplied document metadata. PDF/UA enables the
	// accessibility-tagging pipeline; htmlbag picks it up via
	// fe.Doc.Format == FormatPDFUA at cssbuilder.go:144 and emits
	// StructTreeRoot, MarkInfo, /DisplayDocTitle, /Lang, and the
	// per-element role mapping automatically. /Lang and /Title both
	// also surface in plain PDF output and the XMP metadata sidecar.
	if opts.Title != "" {
		fe.Doc.Title = opts.Title
	}
	if opts.Lang != "" {
		fe.Doc.DefaultLanguageTag = opts.Lang
	}
	if opts.Format == "PDF/UA" {
		fe.Doc.Format = document.FormatPDFUA
	}

	cssParser := csshtml.NewCSSParserWithDefaults()
	// Resolve relative CSS paths against baseDir.
	cssParser.PushDir(baseDir)

	cb, err := htmlbag.New(fe, cssParser)
	if err != nil {
		return fmt.Errorf("creating CSS builder: %w", err)
	}
	if pages, ok := oldAux["_pages"].(float64); ok {
		cb.Counters["pages"] = int(pages)
	}

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

	// Fire content_ready callback (aux loaded, before content processing).
	if err := fireCallback("content_ready"); err != nil {
		return fmt.Errorf("content_ready callback: %w", err)
	}

	// Convert HTML to text elements first — this processes <link> and <style>
	// tags, making @page rules available for page initialization.
	te, err := cb.HTMLToText(htmlStr)
	if err != nil {
		return fmt.Errorf("HTML to text: %w", err)
	}

	// Distribute content across pages. OutputPagesFromText splits the
	// Text tree at forced page breaks and formats each group with the
	// content width of its target page (respecting @page margins).
	if err := cb.OutputPagesFromText(te); err != nil {
		return fmt.Errorf("outputting pages: %w", err)
	}

	// PDF/UA: create bookmarks from headings
	if len(cb.Headings) > 0 {
		for _, h := range cb.Headings {
			if h.Page > 0 && h.Page <= len(fe.Doc.Pages) {
				pg := fe.Doc.Pages[h.Page-1]
				dest := fmt.Sprintf("[%s /Fit]", pg.Objectnumber.Ref())
				fe.Doc.PDFWriter.Outlines = append(fe.Doc.PDFWriter.Outlines, &pdf.Outline{
					Title: h.Text,
					Dest:  dest,
				})
			}
		}
	}

	if err := fe.Finish(); err != nil {
		return fmt.Errorf("finishing document: %w", err)
	}

	// Fire document_end callback (PDF has been written).
	if err := fireCallback("document_end"); err != nil {
		return fmt.Errorf("document_end callback: %w", err)
	}

	// Read _aux back from Lua (user may have modified it) and update system keys.
	curAux := readAuxGlobal(l)
	curAux["_pages"] = len(fe.Doc.Pages)
	headings := make([]any, len(cb.Headings))
	for i, h := range cb.Headings {
		headings[i] = map[string]any{"level": h.Level, "text": h.Text, "page": h.Page}
	}
	curAux["_headings"] = headings
	if changed, err := writeAuxFile(auxPath, oldAux, curAux); err != nil {
		return fmt.Errorf("writing aux file: %w", err)
	} else if changed {
		slog.Info("Aux data changed — rerun to update cross-references")
	}

	slog.Info("PDF written", "file", outputFilename, "pages", len(fe.Doc.Pages))
	return nil
}

func generatePDF(l *lua.State, outputFilename string, htmlStr string, fm Frontmatter, opts Options) error {
	// Read aux data from a previous run (for pages counter and change detection).
	// _aux/_toc globals are already set by ProcessFile.
	auxPath := strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename)) + "-aux.json"
	oldAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("reading aux file: %w", err)
	}

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
	if fm.Format == "PDF/UA" {
		fe.Doc.Format = document.FormatPDFUA
	}
	if fm.Lang != "" {
		fe.Doc.DefaultLanguageTag = fm.Lang
	}

	cssParser := csshtml.NewCSSParserWithDefaults()
	cb, err := htmlbag.New(fe, cssParser)
	if err != nil {
		return fmt.Errorf("creating CSS builder: %w", err)
	}
	if pages, ok := oldAux["_pages"].(float64); ok {
		cb.Counters["pages"] = int(pages)
	}

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

	// Convert HTML to text elements
	te, err := cb.HTMLToText(htmlStr)
	if err != nil {
		return fmt.Errorf("HTML to text: %w", err)
	}

	// Distribute content across pages. OutputPagesFromText splits the
	// Text tree at forced page breaks and formats each group with the
	// content width of its target page (respecting @page margins).
	if err := cb.OutputPagesFromText(te); err != nil {
		return fmt.Errorf("outputting pages: %w", err)
	}

	// PDF/UA: create bookmarks from headings
	if len(cb.Headings) > 0 {
		for _, h := range cb.Headings {
			if h.Page > 0 && h.Page <= len(fe.Doc.Pages) {
				pg := fe.Doc.Pages[h.Page-1]
				dest := fmt.Sprintf("[%s /Fit]", pg.Objectnumber.Ref())
				fe.Doc.PDFWriter.Outlines = append(fe.Doc.PDFWriter.Outlines, &pdf.Outline{
					Title: h.Text,
					Dest:  dest,
				})
			}
		}
	}

	if err := fe.Finish(); err != nil {
		return fmt.Errorf("finishing document: %w", err)
	}

	// Fire document_end callback (PDF has been written).
	if err := fireCallback("document_end"); err != nil {
		return fmt.Errorf("document_end callback: %w", err)
	}

	// Read _aux back from Lua (user may have modified it) and update system keys.
	curAux := readAuxGlobal(l)
	curAux["_pages"] = len(fe.Doc.Pages)
	headings := make([]any, len(cb.Headings))
	for i, h := range cb.Headings {
		headings[i] = map[string]any{"level": h.Level, "text": h.Text, "page": h.Page}
	}
	curAux["_headings"] = headings
	if changed, err := writeAuxFile(auxPath, oldAux, curAux); err != nil {
		return fmt.Errorf("writing aux file: %w", err)
	} else if changed {
		slog.Info("Aux data changed — rerun to update cross-references")
	}

	slog.Info("PDF written", "file", outputFilename, "pages", len(fe.Doc.Pages))
	return nil
}
