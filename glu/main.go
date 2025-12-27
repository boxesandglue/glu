package main

import (
	"fmt"
	"os"
	"time"

	"github.com/speedata/go-lua"
	"github.com/speedata/optionparser"

	luabackend "github.com/speedata/glu/lua/backend"
	luafrontend "github.com/speedata/glu/lua/frontend"
	luapdf "github.com/speedata/glu/lua/pdf"
)

// Version is the version of the program.
var Version string

func dothings() error {
	defaults := map[string]string{
		"loglevel": "info",
	}
	op := optionparser.NewOptionParser()
	op.Banner = "glu - Lua typesetting with boxes and glue"
	op.Coda = "\nUsage: glu [options] <filename.lua>"
	op.On("--loglevel LVL", "Set the log level (debug, info, warn, error)", defaults)
	op.Command("help", "Show the help message")
	op.Command("version", "Print version and exit")
	if err := op.Parse(); err != nil {
		return err
	}

	var mainfile string
	for _, arg := range op.Extra {
		switch arg {
		case "version":
			fmt.Printf("glu version %s\n", Version)
			return nil
		case "help":
			op.Help()
			return nil
		default:
			mainfile = arg
		}
	}

	if mainfile == "" {
		return fmt.Errorf("usage: %s <filename.lua>", os.Args[0])
	}

	// Create Lua state
	l := lua.NewState()
	lua.OpenLibraries(l)

	// Register modules
	luapdf.Open(l)
	luafrontend.Open(l)
	luabackend.Open(l)

	// Execute the Lua file
	if err := lua.DoFile(l, mainfile); err != nil {
		return fmt.Errorf("lua error: %v", err)
	}

	return nil
}

func main() {
	now := time.Now()
	if err := dothings(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	elapsed := time.Since(now)
	fmt.Printf("glu finished in %s\n", elapsed)
}
