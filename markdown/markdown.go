package markdown

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	htmlpkg "html"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/csshtml"
	"github.com/boxesandglue/htmlbag"
	"github.com/speedata/go-lua"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"

	"github.com/boxesandglue/glu/internal/errkind"
	luacommon "github.com/boxesandglue/glu/lua/common"
	luafrontend "github.com/boxesandglue/glu/lua/frontend"
	"github.com/boxesandglue/glu/markdown/mathext"
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

// outlineDepth maps an HTML heading level to its position in the PDF
// outline tree. h1 and h2 share the top level; every level below sits
// one rung deeper than the previous one. The mapping is `max(0, n-2)`
// for hN (so h3=1, h4=2, …). Unknown levels collapse to top-level.
var outlineDepth = map[string]int{
	"h1": 0,
	"h2": 0,
	"h3": 1,
	"h4": 2,
	"h5": 3,
	"h6": 4,
}

// appendHeadingOutlines builds a nested PDF outline tree from the
// heading list collected during VList construction. h1/h2 form the
// top level (Open: true → their children show expanded in the
// reader), h3+ nest by depth and stay collapsed (Open: false). A
// stack tracks the most recent outline at each depth so siblings and
// children attach correctly. Heading-level jumps (e.g. h2 → h4 with
// no h3 between) are absorbed by clamping the depth to the current
// stack size, which prevents a missing-parent panic.
func appendHeadingOutlines(fe *frontend.Document, headings []htmlbag.HeadingEntry) {
	var stack []*pdf.Outline
	ua2 := fe.Doc.Format.IsPDFUA2()
	for _, h := range headings {
		if h.Page <= 0 || h.Page > len(fe.Doc.Pages) {
			continue
		}
		var dest string
		if ua2 && h.SE != nil {
			// PDF/UA-2 §8.8: intra-document destinations must be
			// structure destinations. Pre-allocate the SE object now
			// so Finish() reuses it instead of allocating a new one;
			// the outline /Dest then targets the StructElem directly.
			if h.SE.Obj == nil {
				h.SE.Obj = fe.Doc.PDFWriter.NewObject()
			}
			dest = fmt.Sprintf("[%s /Fit]", h.SE.Obj.ObjectNumber.Ref())
		} else {
			pg := fe.Doc.Pages[h.Page-1]
			dest = fmt.Sprintf("[%s /Fit]", pg.Objectnumber.Ref())
		}

		depth := min(outlineDepth[h.Level], len(stack))

		o := &pdf.Outline{
			Title: h.Text,
			Dest:  dest,
			Open:  depth == 0,
		}

		if depth == 0 {
			fe.Doc.PDFWriter.Outlines = append(fe.Doc.PDFWriter.Outlines, o)
		} else {
			parent := stack[depth-1]
			parent.Children = append(parent.Children, o)
		}
		stack = append(stack[:depth], o)
	}
}

// Result holds post-run statistics from ProcessFile / ProcessHTMLFile.
// Callers wanting a manifest pre-allocate one and pass it via
// Options.Result; the markdown package populates it at the end of
// each pass (the final pass overwrites earlier values).
type Result struct {
	Pages    int                    `json:"pages"`
	Passes   int                    `json:"passes"`
	Headings []htmlbag.HeadingEntry `json:"headings"`
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
	// OutputPath is the resolved PDF output path. If empty, falls back
	// to <input>.pdf alongside the input file.
	OutputPath string
	// MaxPasses caps the auto-rerun loop used to converge the aux file
	// (forward references like total page count). Values <1 are treated
	// as 1 (single pass, no rerun).
	MaxPasses int
	// SetupLua is called once per pass on a fresh lua.State to register
	// modules. May be nil for callers that have already initialised the
	// state (e.g. ProcessHTMLString from the htmlbag bridge — that path
	// reuses the caller's state and ignores SetupLua entirely).
	SetupLua func(l *lua.State)
	// ScriptArgs becomes arg[1..n] in the Lua state. arg[0] is the
	// input filename, arg[-1] is the interpreter (os.Args[0]).
	ScriptArgs []string
	// Result, if non-nil, is populated with end-of-pass statistics.
	// The final pass's data wins. Nil disables collection.
	Result *Result
	// SourceDateEpoch, when non-zero, overrides the PDF CreationDate.
	// Combined with baseline-pdf's already-deterministic /ID (MD5 of
	// xref content) this yields byte-stable PDFs across runs with the
	// same input — the SOURCE_DATE_EPOCH reproducible-builds protocol.
	SourceDateEpoch time.Time
}

// resolveOutput chooses the actual PDF output path: opts.OutputPath if
// set, otherwise <input-without-ext>.pdf. If OutputPath has no
// extension, .pdf is appended.
func resolveOutput(input, opt string) string {
	if opt == "" {
		ext := filepath.Ext(input)
		return input[0:len(input)-len(ext)] + ".pdf"
	}
	if filepath.Ext(opt) == "" {
		return opt + ".pdf"
	}
	return opt
}

// auxPathFor returns the aux JSON path derived from the PDF output path.
func auxPathFor(outputPath string) string {
	ext := filepath.Ext(outputPath)
	return outputPath[0:len(outputPath)-len(ext)] + "-aux.json"
}

// hashAuxBytes returns a short hex hash of marshalled aux data. Used
// for oscillation detection in the multi-pass loop.
func hashAuxBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

// pushArgTable populates Lua `arg` PUC-Rio style: arg[-1]=interpreter,
// arg[0]=script, arg[1..n]=positional args.
func pushArgTable(l *lua.State, scriptName string, scriptArgs []string) {
	l.NewTable()
	l.PushString(os.Args[0])
	l.RawSetInt(-2, -1)
	l.PushString(scriptName)
	l.RawSetInt(-2, 0)
	for i, a := range scriptArgs {
		l.PushString(a)
		l.RawSetInt(-2, i+1)
	}
	l.SetGlobal("arg")
}

// newPassState creates a fresh lua.State, registers modules via the
// caller's SetupLua hook, and seeds the arg table.
func newPassState(opts Options, scriptName string) *lua.State {
	l := lua.NewState()
	if opts.SetupLua != nil {
		opts.SetupLua(l)
	}
	pushArgTable(l, scriptName, opts.ScriptArgs)
	return l
}

// ProcessFile reads a Markdown file and produces a PDF.
// ProcessFile reads a Markdown file and produces a PDF. The Lua state
// is owned by this function: a fresh state is created for each pass
// via opts.SetupLua so that {lua} blocks and the companion script
// always see a clean environment. Auto-rerun for aux convergence is
// driven by opts.MaxPasses; oscillation is detected by hashing the
// written aux content.
func ProcessFile(filename string, opts Options) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%w: reading %s: %s", errkind.IO, filename, err.Error())
	}
	source := string(data)

	if opts.Template {
		tmpl, err := template.New(filepath.Base(filename)).Parse(source)
		if err != nil {
			return fmt.Errorf("%w: template parse: %s", errkind.IO, err.Error())
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			return fmt.Errorf("%w: template execute: %s", errkind.IO, err.Error())
		}
		source = buf.String()
	}

	fm, originalBody := ExtractFrontmatter(source)
	slog.Debug("Frontmatter", "title", fm.Title, "author", fm.Author, "papersize", fm.Papersize, "css", fm.CSS)

	outputFilename := resolveOutput(filename, opts.OutputPath)
	auxPath := auxPathFor(outputFilename)

	// Debug shortcuts (single pass, no PDF). Both still need a live
	// Lua state because {lua} blocks must run before the body is
	// emitted to stdout.
	sourceDir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return fmt.Errorf("%w: resolving source dir: %s", errkind.IO, err.Error())
	}

	if opts.DebugMarkdown || opts.DebugHTML {
		l := newPassState(opts, filename)
		return runMarkdownDebug(l, filename, sourceDir, originalBody, fm, auxPath, opts)
	}

	maxPasses := opts.MaxPasses
	if maxPasses < 1 {
		maxPasses = 1
	}
	seen := map[string]int{}
	for pass := 1; pass <= maxPasses; pass++ {
		if opts.Result != nil {
			opts.Result.Passes = pass
		}
		l := newPassState(opts, filename)
		changed, hash, err := runMarkdownPass(l, filename, sourceDir, originalBody, fm, outputFilename, auxPath, opts)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		if prev, ok := seen[hash]; ok {
			slog.Warn("Aux oscillates between states — giving up", "first_pass", prev, "this_pass", pass, "hash", hash)
			return fmt.Errorf("aux oscillates: %w", errkind.AuxNotConverged)
		}
		seen[hash] = pass
		if pass < maxPasses {
			slog.Info("Aux changed, re-running", "pass", pass, "max_passes", maxPasses)
		}
	}
	slog.Warn("Aux did not converge after max passes", "max_passes", maxPasses)
	return fmt.Errorf("aux file did not converge after %d passes: %w", maxPasses, errkind.AuxNotConverged)
}

// runMarkdownDebug runs the Markdown pipeline far enough to satisfy
// --markdown / --html debug flags, then returns without writing a PDF.
func runMarkdownDebug(l *lua.State, filename, sourceDir, body string, fm Frontmatter, auxPath string, opts Options) error {
	_ = sourceDir // unused in debug path (no PDF, no attachments)
	luacommon.PushAny(l, any(fm.Extra))
	l.SetGlobal("_frontmatter")

	if err := loadCompanionLua(l, filename); err != nil {
		return err
	}
	if err := fireCallback("document_start"); err != nil {
		return fmt.Errorf("%w: document_start callback: %s", errkind.Lua, err.Error())
	}
	earlyAux, err := readAuxFile(auxPath)
	if err != nil {
		return fmt.Errorf("%w: reading aux file: %s", errkind.IO, err.Error())
	}
	setAuxGlobal(l, earlyAux)
	if err := fireCallback("content_ready"); err != nil {
		return fmt.Errorf("%w: content_ready callback: %s", errkind.Lua, err.Error())
	}
	body, err = extractAndRunLuaBlocks(l, body)
	if err != nil {
		return fmt.Errorf("%w: %s", errkind.Lua, err.Error())
	}
	body, err = expandInlineExpressions(l, body)
	if err != nil {
		return fmt.Errorf("%w: %s", errkind.Lua, err.Error())
	}
	if opts.DebugMarkdown {
		fmt.Print(body)
		return nil
	}
	htmlStr, err := markdownToHTML(body, fm)
	if err != nil {
		return err
	}
	fmt.Print(htmlStr)
	return nil
}

// runMarkdownPass executes a single pass of the Markdown pipeline on
// the supplied (fresh) state. Returns whether the aux file changed
// vs. the previous run, a short hash of the new aux for oscillation
// detection, and any error.
func runMarkdownPass(l *lua.State, filename, sourceDir, body string, fm Frontmatter, outputFilename, auxPath string, opts Options) (bool, string, error) {
	luacommon.PushAny(l, any(fm.Extra))
	l.SetGlobal("_frontmatter")

	if err := loadCompanionLua(l, filename); err != nil {
		return false, "", err
	}
	if err := fireCallback("document_start"); err != nil {
		return false, "", fmt.Errorf("%w: document_start callback: %s", errkind.Lua, err.Error())
	}
	oldAux, err := readAuxFile(auxPath)
	if err != nil {
		return false, "", fmt.Errorf("%w: reading aux file: %s", errkind.IO, err.Error())
	}
	setAuxGlobal(l, oldAux)
	if err := fireCallback("content_ready"); err != nil {
		return false, "", fmt.Errorf("%w: content_ready callback: %s", errkind.Lua, err.Error())
	}
	body, err = extractAndRunLuaBlocks(l, body)
	if err != nil {
		return false, "", fmt.Errorf("%w: %s", errkind.Lua, err.Error())
	}
	body, err = expandInlineExpressions(l, body)
	if err != nil {
		return false, "", fmt.Errorf("%w: %s", errkind.Lua, err.Error())
	}
	htmlStr, err := markdownToHTML(body, fm)
	if err != nil {
		return false, "", err
	}
	return renderHTMLToPDF(l, htmlStr, sourceDir, outputFilename, auxPath, fm, opts, true, oldAux)
}

// markdownToHTML converts body to HTML via goldmark. Highlight style
// comes from front matter (highlight-style key), defaulting to github.
func markdownToHTML(body string, fm Frontmatter) (string, error) {
	highlightStyle := "github"
	if s, ok := fm.Extra["highlight-style"].(string); ok {
		highlightStyle = s
	}
	extensions := []goldmark.Extender{
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		highlighting.NewHighlighting(
			highlighting.WithStyle(highlightStyle),
		),
	}
	// Opt-in TeX math: `math: true` in the frontmatter turns on dollar-math
	// parsing ($…$ inline, $$…$$ display → MathML). It is off by default so a
	// document that uses $ as a currency sign is never reinterpreted.
	if fm.Math {
		extensions = append(extensions, mathext.Math)
	}
	gm := goldmark.New(
		goldmark.WithExtensions(extensions...),
		// WithAttribute lets a {.class #id key=value} suffix on
		// headings and block elements turn into HTML attributes —
		// e.g. "# Title {.right}" → <h1 class="right">. Combined
		// with the .left / .right / .center / .justify utility
		// classes shipped in defaultCSS this is how the Markdown
		// frontend exposes paragraph alignment, which CommonMark
		// itself does not have a syntax for.
		goldmark.WithParserOptions(parser.WithAttribute()),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)
	var htmlBuf bytes.Buffer
	if err := gm.Convert([]byte(body), &htmlBuf); err != nil {
		return "", fmt.Errorf("%w: goldmark convert: %s", errkind.Typeset, err.Error())
	}
	// Strip trailing newline inside <pre><code> blocks (chroma and
	// goldmark both add one).
	htmlStr := preTrailingNL.ReplaceAllString(htmlBuf.String(), "$1")
	slog.Debug("HTML generated", "length", len(htmlStr))
	return htmlStr, nil
}

// loadCompanionLua loads <stem>.lua next to the input file if it
// exists. Skipped silently when no companion is present.
func loadCompanionLua(l *lua.State, filename string) error {
	luaFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".lua"
	if _, err := os.Stat(luaFile); err != nil {
		return nil
	}
	slog.Info("Loading companion Lua file", "file", luaFile)
	if err := lua.DoFile(l, luaFile); err != nil {
		return fmt.Errorf("%w: companion lua file %s: %s", errkind.Lua, luaFile, err.Error())
	}
	return nil
}

// ProcessHTMLFile reads an HTML file and produces a PDF, driving the
// same multi-pass aux convergence loop as ProcessFile. Unlike Markdown
// mode no default CSS is applied — styling comes from <link>, <style>,
// inline styles, or the --css flag.
func ProcessHTMLFile(filename string, opts Options) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%w: reading %s: %s", errkind.IO, filename, err.Error())
	}
	htmlStr := string(data)
	outputFilename := resolveOutput(filename, opts.OutputPath)
	auxPath := auxPathFor(outputFilename)
	baseDir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return fmt.Errorf("%w: resolving base dir: %s", errkind.IO, err.Error())
	}

	maxPasses := opts.MaxPasses
	if maxPasses < 1 {
		maxPasses = 1
	}
	seen := map[string]int{}
	for pass := 1; pass <= maxPasses; pass++ {
		if opts.Result != nil {
			opts.Result.Passes = pass
		}
		l := newPassState(opts, filename)
		changed, hash, err := runHTMLPass(l, htmlStr, baseDir, outputFilename, auxPath, filename, opts)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		if prev, ok := seen[hash]; ok {
			slog.Warn("Aux oscillates between states — giving up", "first_pass", prev, "this_pass", pass, "hash", hash)
			return fmt.Errorf("aux oscillates: %w", errkind.AuxNotConverged)
		}
		seen[hash] = pass
		if pass < maxPasses {
			slog.Info("Aux changed, re-running", "pass", pass, "max_passes", maxPasses)
		}
	}
	slog.Warn("Aux did not converge after max passes", "max_passes", maxPasses)
	return fmt.Errorf("aux file did not converge after %d passes: %w", maxPasses, errkind.AuxNotConverged)
}

// runHTMLPass executes one HTML→PDF pass on the supplied state. If
// companionLuaFilename is non-empty, its sibling <stem>.lua is loaded
// before firing document_start. ProcessHTMLString passes "" to skip
// companion loading entirely.
func runHTMLPass(l *lua.State, htmlStr, baseDir, outputFilename, auxPath, companionLuaFilename string, opts Options) (bool, string, error) {
	if companionLuaFilename != "" {
		if err := loadCompanionLua(l, companionLuaFilename); err != nil {
			return false, "", err
		}
	}
	if err := fireCallback("document_start"); err != nil {
		return false, "", fmt.Errorf("%w: document_start callback: %s", errkind.Lua, err.Error())
	}
	oldAux, err := readAuxFile(auxPath)
	if err != nil {
		return false, "", fmt.Errorf("%w: reading aux file: %s", errkind.IO, err.Error())
	}
	setAuxGlobal(l, oldAux)
	if err := fireCallback("content_ready"); err != nil {
		return false, "", fmt.Errorf("%w: content_ready callback: %s", errkind.Lua, err.Error())
	}
	// HTML mode metadata comes from opts (Title/Lang/Format), not from
	// front matter; build a synthetic Frontmatter so renderHTMLToPDF
	// can use one signature for both paths. When the caller gave no title,
	// fall back to the document's <title> element — the natural source for
	// an HTML document's name and required for PDF/UA (dc:title in XMP).
	title := opts.Title
	if title == "" {
		title = extractHTMLTitle(htmlStr)
	}
	// The --format flag wins; otherwise let the HTML declare its own
	// conformance level via <meta name="pdf-format" content="…">, so a
	// self-contained document renders correctly with a bare `glu file.html`.
	format := opts.Format
	if format == "" {
		format = extractHTMLMetaFormat(htmlStr)
	}
	fm := Frontmatter{
		Title:  title,
		Lang:   opts.Lang,
		Format: format,
	}
	return renderHTMLToPDF(l, htmlStr, baseDir, outputFilename, auxPath, fm, opts, false, oldAux)
}

// titleElementRE captures the text content of the first <title> element,
// case-insensitively and across newlines.
var titleElementRE = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// extractHTMLTitle returns the trimmed text of the document's <title>
// element, or "" if there is none. Inner markup is stripped and HTML
// entities are decoded so the result is plain text suitable for the PDF
// document title.
func extractHTMLTitle(htmlStr string) string {
	m := titleElementRE.FindStringSubmatch(htmlStr)
	if m == nil {
		return ""
	}
	inner := stripTagsRE.ReplaceAllString(m[1], "")
	return strings.TrimSpace(htmlpkg.UnescapeString(inner))
}

// stripTagsRE removes any nested element tags from captured title content.
var stripTagsRE = regexp.MustCompile(`<[^>]*>`)

// Matching <meta> elements with arbitrary attribute order: scan each meta
// tag, keep the one naming pdf-format, then pull its content. A single
// ordered regex would miss content-before-name authoring.
var (
	metaTagRE     = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	metaPDFNameRE = regexp.MustCompile(`(?is)\bname\s*=\s*["']pdf-format["']`)
	metaContentRE = regexp.MustCompile(`(?is)\bcontent\s*=\s*["']([^"']*)["']`)
)

// extractHTMLMetaFormat returns the PDF conformance level declared by a
// <meta name="pdf-format" content="PDF/UA-2"> element, or "" if none. This
// is the HTML equivalent of the Markdown frontmatter `format:` key, letting a
// standalone HTML document request its own conformance level without a CLI
// flag.
func extractHTMLMetaFormat(htmlStr string) string {
	for _, tag := range metaTagRE.FindAllString(htmlStr, -1) {
		if !metaPDFNameRE.MatchString(tag) {
			continue
		}
		if m := metaContentRE.FindStringSubmatch(tag); m != nil {
			return strings.TrimSpace(htmlpkg.UnescapeString(m[1]))
		}
	}
	return ""
}

// ProcessHTMLString takes an HTML payload (already in memory), the
// directory to resolve relative CSS @import / <link> paths against,
// and an output PDF filename, and runs the same pipeline as
// ProcessHTMLFile but using the caller-supplied Lua state. Single-
// pass — embedders that need aux convergence drive the loop
// themselves (typically: htmlbag Lua bridge from a user script).
func ProcessHTMLString(l *lua.State, htmlStr, baseDir, outputFilename string, opts Options) error {
	auxPath := auxPathFor(outputFilename)
	changed, _, err := runHTMLPass(l, htmlStr, baseDir, outputFilename, auxPath, "", opts)
	if err != nil {
		return err
	}
	if changed {
		slog.Info("Aux data changed — rerun to update cross-references")
	}
	return nil
}

// renderHTMLToPDF is the shared core for both Markdown and HTML
// paths. The caller has already loaded the companion Lua file, fired
// document_start / content_ready, and (for Markdown) run any {lua}
// blocks. useDefaultCSS=true loads the built-in Markdown stylesheet
// and applies front-matter papersize / css before InitPage(); HTML
// mode lets <style>/<link> in the document drive @page setup.
func renderHTMLToPDF(l *lua.State, htmlStr, baseDir, outputFilename, auxPath string, fm Frontmatter, opts Options, useDefaultCSS bool, oldAux map[string]any) (bool, string, error) {
	fe, err := frontend.New(outputFilename)
	if err != nil {
		return false, "", fmt.Errorf("%w: creating document: %s", errkind.Typeset, err.Error())
	}
	if !opts.SourceDateEpoch.IsZero() {
		// Override the library-level SOURCE_DATE_EPOCH default
		// (boxesandglue/frontend.New reads the env directly) with the
		// CLI flag --source-date-epoch or an explicit programmatic
		// caller value. SuppressInfo swaps the XMP DocumentID /
		// InstanceID UUIDs for stable hardcoded values; the InfoDict
		// CreationDate is written by the document Finish path from
		// Doc.CreationDate; the trailer /ID is an MD5 of the xref
		// byte content and falls out deterministically.
		t := opts.SourceDateEpoch.UTC()
		fe.Doc.CreationDate = t
		fe.Doc.SuppressInfo = true
	}
	if fm.Title != "" {
		fe.Doc.Title = fm.Title
	}
	if fm.Author != "" {
		fe.Doc.Author = fm.Author
	}
	// The --format CLI flag is a fallback: frontmatter (Markdown) and the
	// htmlbag.render options table (Lua) take precedence when they set a
	// format, but for plain HTML input — which has no frontmatter — the flag
	// is the only way to request a conformance level.
	if fm.Format == "" {
		fm.Format = opts.Format
	}
	if fm.Format != "" {
		// Frontmatter accepts a comma-separated list:
		//   format: PDF/A-3b, PDF/UA-1
		// composes both sub-conformances on the same document.
		f, err := document.ParseFormat(fm.Format)
		if err != nil {
			return false, "", fmt.Errorf("%w: frontmatter format: %s", errkind.Typeset, err.Error())
		}
		fe.Doc.Format = f
	}
	if err := applyAttachments(fe, baseDir, fm.Attachments); err != nil {
		return false, "", fmt.Errorf("%w: applying attachments: %s", errkind.IO, err.Error())
	}
	if fm.Lang != "" {
		fe.Doc.DefaultLanguageTag = fm.Lang
		// Tags without a TeX pattern (Arabic/Hebrew/CJK/unknown)
		// resolve to a no-op hyphenator — safe to call unconditionally.
		if l, err := frontend.GetLanguage(fm.Lang); err == nil {
			fe.Doc.DefaultLanguage = l
		}
	}

	cssParser := csshtml.NewCSSParserWithDefaults()
	if baseDir != "" {
		cssParser.PushDir(baseDir)
	}

	cb, err := htmlbag.New(fe, cssParser)
	if err != nil {
		return false, "", fmt.Errorf("%w: creating CSS builder: %s", errkind.Typeset, err.Error())
	}
	if pages, ok := oldAux["_pages"].(float64); ok {
		cb.Counters["pages"] = int(pages)
	}
	if anchorsRaw, ok := oldAux["_anchors"].(map[string]any); ok {
		anchorPages := make(map[string]int, len(anchorsRaw))
		anchorTexts := make(map[string]string, len(anchorsRaw))
		for id, raw := range anchorsRaw {
			switch v := raw.(type) {
			case float64:
				// Legacy aux format: _anchors was just id → page.
				anchorPages[id] = int(v)
			case map[string]any:
				if p, ok := v["page"].(float64); ok {
					anchorPages[id] = int(p)
				}
				if t, ok := v["text"].(string); ok {
					anchorTexts[id] = t
				}
			}
		}
		cb.SetAnchorPages(anchorPages)
		cb.SetAnchorTexts(anchorTexts)
	}

	// Install lifecycle callbacks. Must happen after CSSBuilder
	// creation so PageInfo is available to the pre_shipout hook.
	if cr := luafrontend.GetRegistry(); cr != nil {
		cr.SetCSSBuilder(cb)
		cr.InstallPreShipout(fe)
		cr.InstallPostElement()
		cr.InstallPageInit()
	}

	if useDefaultCSS {
		if err := cb.AddCSS(defaultCSS); err != nil {
			return false, "", fmt.Errorf("%w: adding default CSS: %s", errkind.Typeset, err.Error())
		}
		if fm.Papersize != "" {
			pageSizeCSS := fmt.Sprintf("@page { size: %s; }", fm.Papersize)
			if err := cb.AddCSS(pageSizeCSS); err != nil {
				return false, "", fmt.Errorf("%w: applying papersize: %s", errkind.Typeset, err.Error())
			}
		}
		if fm.CSS != "" {
			if err := cb.ReadCSSFile(fm.CSS); err != nil {
				return false, "", fmt.Errorf("%w: reading CSS file %s: %s", errkind.IO, fm.CSS, err.Error())
			}
		}
	}
	if opts.CSSFile != "" {
		if err := cb.ReadCSSFile(opts.CSSFile); err != nil {
			return false, "", fmt.Errorf("%w: reading CSS file %s: %s", errkind.IO, opts.CSSFile, err.Error())
		}
	}

	// Markdown path explicitly initialises the page before
	// HTMLToText runs because its default CSS already declares
	// @page. HTML mode relies on HTMLToText to discover @page from
	// the document's own <style>/<link>, so we skip InitPage there.
	if useDefaultCSS {
		if err := cb.InitPage(); err != nil {
			return false, "", fmt.Errorf("%w: initializing page: %s", errkind.Typeset, err.Error())
		}
	}

	te, err := cb.HTMLToText(htmlStr)
	if err != nil {
		return false, "", fmt.Errorf("%w: HTML to text: %s", errkind.Typeset, err.Error())
	}
	if err := cb.OutputPagesFromText(te); err != nil {
		return false, "", fmt.Errorf("%w: outputting pages: %s", errkind.Typeset, err.Error())
	}

	if len(cb.Headings) > 0 {
		appendHeadingOutlines(fe, cb.Headings)
	}

	if err := fe.Finish(); err != nil {
		return false, "", fmt.Errorf("%w: finishing document: %s", errkind.Typeset, err.Error())
	}
	if err := fireCallback("document_end"); err != nil {
		return false, "", fmt.Errorf("%w: document_end callback: %s", errkind.Lua, err.Error())
	}

	curAux := readAuxGlobal(l)
	curAux["_pages"] = len(fe.Doc.Pages)
	headings := make([]any, len(cb.Headings))
	for i, h := range cb.Headings {
		headings[i] = map[string]any{"level": h.Level, "text": h.Text, "page": h.Page}
	}
	curAux["_headings"] = headings
	anchors := make(map[string]any, len(cb.Anchors))
	for _, a := range cb.Anchors {
		if a.ID != "" && a.Page > 0 {
			anchors[a.ID] = map[string]any{
				"page": a.Page,
				"text": a.Text,
			}
		}
	}
	curAux["_anchors"] = anchors

	curBytes, err := json.MarshalIndent(curAux, "", "  ")
	if err != nil {
		return false, "", fmt.Errorf("%w: marshalling aux: %s", errkind.IO, err.Error())
	}
	if err := os.WriteFile(auxPath, curBytes, 0644); err != nil {
		return false, "", fmt.Errorf("%w: writing aux file: %s", errkind.IO, err.Error())
	}
	oldBytes, _ := json.MarshalIndent(oldAux, "", "  ")
	changed := !bytes.Equal(oldBytes, curBytes)

	if opts.Result != nil {
		opts.Result.Pages = len(fe.Doc.Pages)
		opts.Result.Headings = append([]htmlbag.HeadingEntry(nil), cb.Headings...)
	}

	slog.Info("PDF written", "file", outputFilename, "pages", len(fe.Doc.Pages))
	return changed, hashAuxBytes(curBytes), nil
}
