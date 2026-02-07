package markdown

import (
	"bytes"
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

	luafrontend "github.com/speedata/glu/lua/frontend"
)

// Options controls the Markdown processing pipeline.
type Options struct {
	Template    bool   // apply Go template expansion before processing
	CSSFile     string // additional CSS file to load
	DebugMarkdown bool // print expanded Markdown to stdout instead of generating PDF
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
	gm := goldmark.New(goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
	))
	var htmlBuf bytes.Buffer
	if err := gm.Convert([]byte(body), &htmlBuf); err != nil {
		return fmt.Errorf("goldmark convert: %w", err)
	}
	htmlStr := htmlBuf.String()
	slog.Debug("HTML generated", "length", len(htmlStr))

	// Step 6: HTML → PDF via htmlbag
	outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".pdf"
	return generatePDF(outputFilename, htmlStr, fm, opts)
}

func generatePDF(outputFilename string, htmlStr string, fm Frontmatter, opts Options) error {
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

	// Install pre_shipout callback hook so Lua callbacks fire on Shipout
	if cr := luafrontend.GetRegistry(); cr != nil {
		cr.InstallPreShipout(fe)
	}

	cssParser := csshtml.NewCSSParserWithDefaults()
	cb, err := htmlbag.New(fe, cssParser)
	if err != nil {
		return fmt.Errorf("creating CSS builder: %w", err)
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

	// Place on page
	fe.Doc.CurrentPage.OutputAt(pd.MarginLeft, pd.Height-pd.MarginTop, vl)

	// Finalize
	fe.Doc.CurrentPage.Shipout()
	if err := fe.Finish(); err != nil {
		return fmt.Errorf("finishing document: %w", err)
	}

	slog.Info("PDF written", "file", outputFilename)
	return nil
}
