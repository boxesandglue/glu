// Package cxpath provides Lua bindings for XPath XML querying using cxpath.
package cxpath

import (
	"github.com/speedata/cxpath"
	"github.com/speedata/go-lua"
)

const contextMetaTable = "XPathContext"

// Context wraps the cxpath.Context type
type Context struct {
	Value *cxpath.Context
}

// checkContext retrieves a Context userdata from the stack
func checkContext(l *lua.State, index int) *Context {
	ud := lua.CheckUserData(l, index, contextMetaTable)
	if ctx, ok := ud.(*Context); ok {
		return ctx
	}
	lua.Errorf(l, "XPathContext expected")
	return nil
}

// xpathOpen opens an XML file: xpath.open(filename)
func xpathOpen(l *lua.State) int {
	filename := lua.CheckString(l, 1)

	ctx, err := cxpath.NewFromFile(filename)
	if err != nil {
		lua.Errorf(l, "failed to open XML file: %s", err.Error())
		return 0
	}

	l.PushUserData(&Context{Value: ctx})
	lua.SetMetaTableNamed(l, contextMetaTable)
	return 1
}

// contextSetNamespace sets a namespace prefix: ctx:set_namespace(prefix, uri)
func contextSetNamespace(l *lua.State) int {
	ctx := checkContext(l, 1)
	prefix := lua.CheckString(l, 2)
	uri := lua.CheckString(l, 3)

	ctx.Value.SetNamespace(prefix, uri)

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// contextEval evaluates an XPath expression: ctx:eval(xpath)
func contextEval(l *lua.State) int {
	ctx := checkContext(l, 1)
	xpath := lua.CheckString(l, 2)

	result := ctx.Value.Eval(xpath)
	if result.Error != nil {
		lua.Errorf(l, "XPath error: %s", result.Error.Error())
		return 0
	}

	l.PushUserData(&Context{Value: result})
	lua.SetMetaTableNamed(l, contextMetaTable)
	return 1
}

// contextString returns the string value: ctx:string() or ctx.string
func contextString(l *lua.State) int {
	ctx := checkContext(l, 1)
	l.PushString(ctx.Value.String())
	return 1
}

// contextInt returns the integer value: ctx:int()
func contextInt(l *lua.State) int {
	ctx := checkContext(l, 1)
	l.PushInteger(ctx.Value.Int())
	return 1
}

// contextBool returns the boolean value: ctx:bool()
func contextBool(l *lua.State) int {
	ctx := checkContext(l, 1)
	l.PushBoolean(ctx.Value.Bool())
	return 1
}

// contextEach returns an iterator: for item in ctx:each(xpath) do ... end
func contextEach(l *lua.State) int {
	ctx := checkContext(l, 1)
	xpath := lua.CheckString(l, 2)

	// Collect all results into a table
	results := make([]*cxpath.Context, 0)
	for item := range ctx.Value.Each(xpath) {
		results = append(results, item)
	}

	// Create iterator closure
	index := 0
	iterator := func(l *lua.State) int {
		if index >= len(results) {
			l.PushNil()
			return 1
		}
		l.PushUserData(&Context{Value: results[index]})
		lua.SetMetaTableNamed(l, contextMetaTable)
		index++
		return 1
	}

	l.PushGoFunction(iterator)
	return 1
}

// contextRoot returns the root element: ctx:root()
func contextRoot(l *lua.State) int {
	ctx := checkContext(l, 1)
	result := ctx.Value.Root()

	l.PushUserData(&Context{Value: result})
	lua.SetMetaTableNamed(l, contextMetaTable)
	return 1
}

// contextIndex handles attribute access (__index metamethod)
func contextIndex(l *lua.State) int {
	ctx := checkContext(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	// Properties
	case "string":
		l.PushString(ctx.Value.String())
		return 1
	// Methods
	case "set_namespace":
		l.PushGoFunction(contextSetNamespace)
		return 1
	case "eval":
		l.PushGoFunction(contextEval)
		return 1
	case "each":
		l.PushGoFunction(contextEach)
		return 1
	case "root":
		l.PushGoFunction(contextRoot)
		return 1
	case "int":
		l.PushGoFunction(contextInt)
		return 1
	case "bool":
		l.PushGoFunction(contextBool)
		return 1
	}
	return 0
}

// registerContextMetaTable creates the XPathContext metatable
func registerContextMetaTable(l *lua.State) {
	lua.NewMetaTable(l, contextMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: contextIndex},
	}, 0)
	l.Pop(1)
}

// openCXPath creates the cxpath module table for require("glu.cxpath")
func openCXPath(l *lua.State) int {
	registerContextMetaTable(l)

	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "open", Function: xpathOpen},
	})
	return 1
}

// Open registers the cxpath module for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "xml.cxpath", openCXPath, false)
	l.Pop(1)
}
