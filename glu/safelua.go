package main

import (
	"github.com/speedata/go-lua"
)

// openSafeLibraries opens a hardened subset of the Lua standard
// libraries — string / table / math / utf8 / bit32 / coroutine plus
// the base library — and then nils out the base-library functions
// that can read or execute arbitrary code (loadfile, dofile, load,
// loadstring) and rips the file-searcher out of the package library
// so require() can only return modules already in _LOADED.
//
// Compared to lua.OpenLibraries this drops io, os, and debug
// entirely. Callers that need them in trusted mode should keep
// using OpenLibraries.
//
// CAVEAT: Lua sandboxing is best-effort. string.dump + load tricks,
// metatable shenanigans on registry tables, and any unsafe Go
// binding (including glu's own frontend.new) can still touch the
// filesystem. Treat --safe as a defence-in-depth knob, not a
// hermetic seal.
func openSafeLibraries(l *lua.State) {
	libs := []struct {
		name string
		f    lua.Function
	}{
		{"_G", lua.BaseOpen},
		{"package", lua.PackageOpen},
		{"coroutine", lua.CoroutineOpen},
		{"table", lua.TableOpen},
		{"string", lua.StringOpen},
		{"bit32", lua.Bit32Open},
		{"math", lua.MathOpen},
		{"utf8", lua.UTF8Open},
	}
	for _, lib := range libs {
		lua.Require(l, lib.name, lib.f, true)
		l.Pop(1)
	}

	// Neutralise base-library escape hatches.
	for _, name := range []string{"loadfile", "dofile", "load", "loadstring", "collectgarbage"} {
		l.PushNil()
		l.SetGlobal(name)
	}

	// Strip the file searcher and native-library hooks from `package`
	// so require() can only resolve modules that were already preloaded
	// into _LOADED (i.e. the glu modules registered by setupLua).
	l.Global("package")
	if l.IsTable(-1) {
		l.PushNil()
		l.SetField(-2, "loadlib")
		l.PushNil()
		l.SetField(-2, "searchpath")
		l.PushString("")
		l.SetField(-2, "cpath")
		l.PushString("")
		l.SetField(-2, "path")

		// Replace package.searchers with a single function that only
		// looks in _LOADED (set up by Require). This makes the
		// existing require() function behave like a cache-only lookup
		// without us having to write a custom require in Go.
		l.NewTable()
		l.PushGoFunction(loadedOnlySearcher)
		l.RawSetInt(-2, 1)
		l.SetField(-2, "searchers")
	}
	l.Pop(1) // pop package
}

// loadedOnlySearcher implements the single permitted package
// searcher in safe mode. require(name) calls each searcher in
// package.searchers until one returns a loader. We return the
// already-loaded module directly via the _LOADED table.
func loadedOnlySearcher(l *lua.State) int {
	name := lua.CheckString(l, 1)
	l.Field(lua.RegistryIndex, "_LOADED")
	l.Field(-1, name)
	if !l.IsNil(-1) {
		// Return a "loader" that just returns the cached module.
		// require() invokes the loader; we make that a closure that
		// pushes the captured module value.
		l.PushGoClosure(func(l *lua.State) int {
			l.PushValue(lua.UpValueIndex(1))
			return 1
		}, 1)
		return 1
	}
	l.Pop(2) // pop nil + _LOADED
	l.PushString("\n\tmodule '" + name + "' not available in --safe mode")
	return 1
}
