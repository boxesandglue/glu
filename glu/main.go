package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/speedata/go-lua"
	"github.com/speedata/optionparser"

	luabackend "github.com/boxesandglue/glu/lua/backend"
	luacxpath "github.com/boxesandglue/glu/lua/cxpath"
	luafrontend "github.com/boxesandglue/glu/lua/frontend"
	luahtmlbag "github.com/boxesandglue/glu/lua/htmlbag"
	luajson "github.com/boxesandglue/glu/lua/json"
	lualog "github.com/boxesandglue/glu/lua/log"
	luapdf "github.com/boxesandglue/glu/lua/pdf"
	luatextshape "github.com/boxesandglue/glu/lua/textshape"
	"github.com/boxesandglue/glu/markdown"

	"github.com/boxesandglue/glu/internal/errkind"
	"github.com/boxesandglue/hobby"
)

// Version is the version of the program.
var (
	Version string
	logger  *slog.Logger
)

// registerGluModules adds the glu Lua module surface to a state. The
// state has already had its standard libraries opened (full or safe
// subset) by the caller.
func registerGluModules(l *lua.State) {
	luapdf.Open(l)
	luafrontend.Open(l)
	luabackend.Open(l)
	luacxpath.Open(l)
	luahtmlbag.Open(l)
	luatextshape.Open(l)
	luajson.Open(l)
	lualog.Open(l)
	hobby.Open(l)
}

// makeSetupLua returns a SetupLua callback that opens the chosen set
// of standard libraries and then registers the glu modules. Called
// once per pass for the Markdown/HTML pipelines (which need a fresh
// state per re-run), or once total for the .lua entry path.
func makeSetupLua(safe bool) func(l *lua.State) {
	return func(l *lua.State) {
		if safe {
			openSafeLibraries(l)
		} else {
			lua.OpenLibraries(l)
		}
		registerGluModules(l)
	}
}

// pushArgTable populates the Lua `arg` global PUC-Rio style:
// arg[-1]=interpreter, arg[0]=script, arg[1..n]=positional args.
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

// outputPathFor resolves the user-requested output path against the
// input file. If outputFlag is empty, the output sits next to the
// input with extension replaced by .pdf. If outputFlag is a path
// without .pdf extension, the extension is appended.
func outputPathFor(input, outputFlag string) string {
	if outputFlag == "" {
		ext := filepath.Ext(input)
		return input[0:len(input)-len(ext)] + ".pdf"
	}
	if filepath.Ext(outputFlag) == "" {
		return outputFlag + ".pdf"
	}
	return outputFlag
}

// logPathFor returns the log file path co-located with the output PDF.
func logPathFor(outputPath string) string {
	ext := filepath.Ext(outputPath)
	return outputPath[0:len(outputPath)-len(ext)] + ".log"
}

// stdinToTempFile reads all of os.Stdin into a fresh temp file with
// the requested extension (".md" etc.) and returns its path. Caller
// is responsible for cleanup via os.Remove.
func stdinToTempFile(ext string) (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("%w: reading stdin: %s", errkind.IO, err.Error())
	}
	f, err := os.CreateTemp("", "glu-stdin-*"+ext)
	if err != nil {
		return "", fmt.Errorf("%w: creating temp file: %s", errkind.IO, err.Error())
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", fmt.Errorf("%w: writing temp file: %s", errkind.IO, err.Error())
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("%w: closing temp file: %s", errkind.IO, err.Error())
	}
	return f.Name(), nil
}

func dothings() error {
	now := time.Now()
	var loglevel string = "info"
	var quiet bool
	var showVersion bool
	var useTemplate bool
	var cssFile string
	var debugMarkdown bool
	var debugHTML bool
	var clean bool
	var cpuprofile string
	var outputPath string
	var maxPassesStr string
	var safe bool
	var logFile string
	var logFormat string = "text"
	var inputFormat string
	var manifestPath string
	var sourceDateEpochStr string
	op := optionparser.NewOptionParser()
	op.Banner = "glu - typesetting with boxes and glue"
	op.Coda = helpCoda()
	op.On("--loglevel LVL", "Set the log level (debug, info, warn, error)", &loglevel)
	op.On("-q", "--quiet", "Suppress output on console", &quiet)
	op.On("-v", "--version", "Print version and exit", &showVersion)
	op.On("-o", "--output FILE", "Write PDF to FILE (default: <input>.pdf)", &outputPath)
	op.On("--max-passes N", "Maximum auto-rerun passes for aux convergence (default: 3)", &maxPassesStr)
	op.On("--safe", "Run Lua in a sandbox (no io/os/debug, no file loaders)", &safe)
	op.On("--log-file FILE", "Write log to FILE (default: <output>.log; '-' disables)", &logFile)
	op.On("--log-format FMT", "Log format: text or json (default: text)", &logFormat)
	op.On("--input-format FMT", "Force input format for stdin: md, html, or lua", &inputFormat)
	op.On("--manifest FILE", "Write a JSON manifest (pages, passes, headings, duration)", &manifestPath)
	op.On("--source-date-epoch SECONDS", "Override PDF CreationDate (also honours $SOURCE_DATE_EPOCH)", &sourceDateEpochStr)
	op.On("--template", "Apply Go template expansion (Markdown mode)", &useTemplate)
	op.On("--css FILE", "Additional CSS file", &cssFile)
	op.On("--markdown", "Print expanded Markdown to stdout (debug)", &debugMarkdown)
	op.On("--html", "Print generated HTML to stdout (debug, Markdown mode)", &debugHTML)
	op.On("--clean", "Remove auxiliary files before processing", &clean)
	op.On("--cpuprofile FILE", "Write CPU profile to file", &cpuprofile)
	op.Command("help", "Show the help message")
	op.Command("version", "Print version and exit")
	op.Command("doctor", "Run environment self-checks")
	if err := op.Parse(); err != nil {
		// --help / -h prints the help text and returns ErrHelp — not
		// an actual error from the user's perspective.
		if errors.Is(err, optionparser.ErrHelp) {
			return nil
		}
		return fmt.Errorf("%w: %s", errkind.Usage, err.Error())
	}

	if showVersion {
		fmt.Printf("glu version %s\n", Version)
		return nil
	}

	maxPasses := 3
	if maxPassesStr != "" {
		n, err := strconv.Atoi(maxPassesStr)
		if err != nil || n < 1 {
			return fmt.Errorf("%w: --max-passes must be a positive integer, got %q", errkind.Usage, maxPassesStr)
		}
		maxPasses = n
	}

	switch logFormat {
	case "text", "json":
	default:
		return fmt.Errorf("%w: --log-format must be 'text' or 'json', got %q", errkind.Usage, logFormat)
	}

	// SOURCE_DATE_EPOCH: --source-date-epoch flag overrides env, both
	// must be a positive integer seconds-since-Unix-epoch.
	var sourceDateEpoch time.Time
	if sourceDateEpochStr == "" {
		sourceDateEpochStr = os.Getenv("SOURCE_DATE_EPOCH")
	}
	if sourceDateEpochStr != "" {
		secs, err := strconv.ParseInt(sourceDateEpochStr, 10, 64)
		if err != nil || secs < 0 {
			return fmt.Errorf("%w: --source-date-epoch must be a non-negative integer, got %q", errkind.Usage, sourceDateEpochStr)
		}
		sourceDateEpoch = time.Unix(secs, 0).UTC()
	}

	// Multi-call binary: when invoked under any name other than "glu" —
	// typically via a symlink — look for a Lua script of that name and
	// run it with all positional arguments forwarded. Search order:
	//   1. current working directory
	//   2. directory of the symlink (os.Args[0])
	// Flags that glu owns (--loglevel, --css, etc.) are still handled
	// up front; everything else goes to the script.
	binBase := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	if binBase != "glu" {
		scriptName := binBase + ".lua"
		scriptPath := ""
		if _, err := os.Stat(scriptName); err == nil {
			if abs, absErr := filepath.Abs(scriptName); absErr == nil {
				scriptPath = abs
			}
		}
		if scriptPath == "" {
			symDir := filepath.Dir(os.Args[0])
			if symDir != "" {
				candidate := filepath.Join(symDir, scriptName)
				if _, err := os.Stat(candidate); err == nil {
					scriptPath = candidate
				}
			}
		}
		if scriptPath == "" {
			return fmt.Errorf("%w: multi-call %q: no %s found in current directory or next to %s", errkind.IO, binBase, scriptName, os.Args[0])
		}
		op.Extra = append([]string{scriptPath}, op.Extra...)
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			return fmt.Errorf("%w: creating CPU profile: %s", errkind.IO, err.Error())
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("starting CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Configure logger based on loglevel
	var level slog.Level
	switch loglevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	if len(op.Extra) == 0 {
		return fmt.Errorf("%w: usage: %s <filename.lua|filename.md>", errkind.Usage, os.Args[0])
	}
	switch op.Extra[0] {
	case "version":
		fmt.Printf("glu version %s\n", Version)
		return nil
	case "help":
		op.Help()
		return nil
	case "doctor":
		if n := runDoctor(os.Stdout); n > 0 {
			return fmt.Errorf("doctor reported %d failure(s)", n)
		}
		return nil
	}
	mainfile, scriptArgs := op.Extra[0], op.Extra[1:]
	originalInput := mainfile

	// Stdin input: `-` reads from os.Stdin and writes to a temp file
	// so the rest of the pipeline (extension dispatch, companion lua
	// lookup, aux file placement) keeps working. Companion .lua won't
	// match the tempfile's unique name — that's by design, stdin
	// callers are expected to be one-shot.
	if mainfile == "-" {
		if inputFormat == "" {
			return fmt.Errorf("%w: stdin input requires --input-format (md, html, lua)", errkind.Usage)
		}
		ext := "." + inputFormat
		switch inputFormat {
		case "md", "html", "htm", "lua":
		default:
			return fmt.Errorf("%w: --input-format must be md, html, or lua, got %q", errkind.Usage, inputFormat)
		}
		if outputPath == "" && inputFormat != "lua" {
			return fmt.Errorf("%w: stdin input requires -o / --output", errkind.Usage)
		}
		tmpFile, err := stdinToTempFile(ext)
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile)
		mainfile = tmpFile
	}

	// Resolve output path. For .lua entry points the PDF is created by
	// the script itself, so outputPath is informational only (used to
	// place the log file). For .md / .html the resolved path is passed
	// through to the markdown package.
	resolvedOutput := outputPathFor(mainfile, outputPath)

	// Resolve log path: explicit --log-file wins; "-" disables the
	// log file entirely. Otherwise the log sits next to the PDF.
	noLogFile := logFile == "-"
	logfilename := logFile
	if logfilename == "" {
		logfilename = logPathFor(resolvedOutput)
	}

	// Log handler chain.
	var fileSink slog.Handler
	if !noLogFile {
		logfileFD, err := os.Create(logfilename)
		if err != nil {
			return fmt.Errorf("%w: failed to create log file: %s", errkind.IO, err.Error())
		}
		defer logfileFD.Close()
		if logFormat == "json" {
			fileSink = NewJSONHandler(logfileFD, level)
		} else {
			fileSink = NewFileHandler(logfileFD, level)
		}
	}

	var handler slog.Handler
	if quiet {
		if fileSink != nil {
			handler = fileSink
		} else {
			// No file, quiet → no logging at all. Use a level above
			// error to discard everything.
			handler = NewConsoleHandler(io.Discard, slog.LevelError+1)
		}
	} else {
		var consoleSink slog.Handler
		if logFormat == "json" {
			consoleSink = NewJSONHandler(os.Stderr, level)
		} else {
			consoleSink = NewConsoleHandler(os.Stdout, level)
		}
		if fileSink != nil {
			handler = NewMultiHandler(fileSink, consoleSink)
		} else {
			handler = consoleSink
		}
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
	bag.SetLogger(logger)
	pdf.Logger = logger

	if clean {
		auxFile := strings.TrimSuffix(resolvedOutput, filepath.Ext(resolvedOutput)) + "-aux.json"
		if err := os.Remove(auxFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("%w: removing %s: %s", errkind.IO, auxFile, err.Error())
		}
	}

	logger.Info("Start processing", "file", mainfile, "glu version", Version, "date", time.Now().Format(time.RFC3339))

	setupLua := makeSetupLua(safe)
	if safe {
		logger.Info("Running in --safe mode (Lua sandboxed: no io/os/debug)")
	}

	var result *markdown.Result
	if manifestPath != "" {
		result = &markdown.Result{}
	}

	ext := filepath.Ext(mainfile)
	switch ext {
	case ".lua":
		// Direct Lua entry: single-shot, fresh state, no aux loop.
		l := lua.NewState()
		setupLua(l)
		pushArgTable(l, mainfile, scriptArgs)
		if err := lua.DoFile(l, mainfile); err != nil {
			return fmt.Errorf("%w: %s", errkind.Lua, err.Error())
		}
	case ".md":
		opts := markdown.Options{
			Template:        useTemplate,
			CSSFile:         cssFile,
			DebugMarkdown:   debugMarkdown,
			DebugHTML:       debugHTML,
			OutputPath:      resolvedOutput,
			MaxPasses:       maxPasses,
			SetupLua:        setupLua,
			ScriptArgs:      scriptArgs,
			Result:          result,
			SourceDateEpoch: sourceDateEpoch,
		}
		if err := markdown.ProcessFile(mainfile, opts); err != nil {
			return err
		}
	case ".html", ".htm":
		opts := markdown.Options{
			CSSFile:         cssFile,
			OutputPath:      resolvedOutput,
			MaxPasses:       maxPasses,
			SetupLua:        setupLua,
			ScriptArgs:      scriptArgs,
			Result:          result,
			SourceDateEpoch: sourceDateEpoch,
		}
		if err := markdown.ProcessHTMLFile(mainfile, opts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: unsupported file extension: %s (expected .lua, .md, or .html)", errkind.Usage, ext)
	}

	if manifestPath != "" {
		if err := writeManifest(manifestPath, originalInput, resolvedOutput, time.Since(now), result); err != nil {
			return err
		}
		logger.Info("Manifest written", "file", manifestPath)
	}

	elapsed := time.Since(now)
	logger.Info("Transcript written", "file", logfilename)
	logger.Info("Total duration", "duration", elapsed.String())

	return nil
}

func main() {
	if err := dothings(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}
