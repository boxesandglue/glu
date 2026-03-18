package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/speedata/go-lua"
	"github.com/speedata/optionparser"

	luabackend "github.com/speedata/glu/lua/backend"
	luacxpath "github.com/speedata/glu/lua/cxpath"
	luafrontend "github.com/speedata/glu/lua/frontend"
	luajson "github.com/speedata/glu/lua/json"
	lualog "github.com/speedata/glu/lua/log"
	luapdf "github.com/speedata/glu/lua/pdf"
	luatextshape "github.com/speedata/glu/lua/textshape"
	"github.com/speedata/glu/markdown"

	"github.com/boxesandglue/hobby"
)

// Version is the version of the program.
var (
	Version string
	logger  *slog.Logger
)

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
	op := optionparser.NewOptionParser()
	op.Banner = "glu - typesetting with boxes and glue"
	op.Coda = "\nUsage: glu [options] <filename.lua|filename.md|filename.html>"
	op.On("--loglevel LVL", "Set the log level (debug, info, warn, error)", &loglevel)
	op.On("-q", "--quiet", "Suppress output on console", &quiet)
	op.On("-v", "--version", "Print version and exit", &showVersion)
	op.On("--template", "Apply Go template expansion (Markdown mode)", &useTemplate)
	op.On("--css FILE", "Additional CSS file", &cssFile)
	op.On("--markdown", "Print expanded Markdown to stdout (debug)", &debugMarkdown)
	op.On("--html", "Print generated HTML to stdout (debug, Markdown mode)", &debugHTML)
	op.On("--clean", "Remove auxiliary files before processing", &clean)
	op.On("--cpuprofile FILE", "Write CPU profile to file", &cpuprofile)
	op.Command("help", "Show the help message")
	op.Command("version", "Print version and exit")
	if err := op.Parse(); err != nil {
		return err
	}

	if showVersion {
		fmt.Printf("glu version %s\n", Version)
		return nil
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			return fmt.Errorf("creating CPU profile: %w", err)
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

	var mainfile string
	var scriptArgs []string
	for i, a := range op.Extra {
		if a == "version" {
			fmt.Printf("glu version %s\n", Version)
			return nil
		}
		if a == "help" {
			op.Help()
			return nil
		}
		mainfile = a
		scriptArgs = op.Extra[i+1:]
		break
	}

	if mainfile == "" {
		return fmt.Errorf("usage: %s <filename.lua|filename.md>", os.Args[0])
	}
	// logfile is main file with .log extension
	ext := filepath.Ext(mainfile)
	logfilename := mainfile[0:len(mainfile)-len(ext)] + ".log"

	// Open log file
	logfile, err := os.Create(logfilename)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logfile.Close()

	// Create handlers
	fileHandler := NewFileHandler(logfile, level)
	var handler slog.Handler
	if quiet {
		handler = fileHandler
	} else {
		consoleHandler := NewConsoleHandler(os.Stdout, level)
		handler = NewMultiHandler(fileHandler, consoleHandler)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
	bag.SetLogger(logger)
	pdf.Logger = logger

	if clean {
		auxFile := mainfile[0:len(mainfile)-len(ext)] + "-aux.json"
		if err := os.Remove(auxFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", auxFile, err)
		}
	}

	logger.Info("Start processing", "file", mainfile, "glu version", Version, "date", time.Now().Format(time.RFC3339))

	// Create Lua state
	l := lua.NewState()
	lua.OpenLibraries(l)

	// Register modules
	luapdf.Open(l)
	luafrontend.Open(l)
	luabackend.Open(l)
	luacxpath.Open(l)
	luatextshape.Open(l)
	luajson.Open(l)
	lualog.Open(l)
	hobby.Open(l)

	// Set up arg table (PUC-Rio Lua compatible):
	//   arg[-1] = interpreter name
	//   arg[0]  = script name
	//   arg[1], arg[2], ... = script arguments
	l.NewTable()
	l.PushString(os.Args[0])
	l.RawSetInt(-2, -1)
	l.PushString(mainfile)
	l.RawSetInt(-2, 0)
	for i, a := range scriptArgs {
		l.PushString(a)
		l.RawSetInt(-2, i+1)
	}
	l.SetGlobal("arg")

	switch ext {
	case ".lua":
		if err := lua.DoFile(l, mainfile); err != nil {
			return fmt.Errorf("lua error: %v", err)
		}
	case ".md":
		opts := markdown.Options{
			Template:      useTemplate,
			CSSFile:       cssFile,
			DebugMarkdown: debugMarkdown,
			DebugHTML:     debugHTML,
		}
		if err := markdown.ProcessFile(l, mainfile, opts); err != nil {
			return err
		}
	case ".html", ".htm":
		opts := markdown.Options{
			CSSFile: cssFile,
		}
		if err := markdown.ProcessHTMLFile(l, mainfile, opts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported file extension: %s (expected .lua, .md, or .html)", ext)
	}

	elapsed := time.Since(now)
	logger.Info("Transcript written", "file", logfilename)
	logger.Info("Total duration", "duration", elapsed.String())

	return nil
}

func main() {
	if err := dothings(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
