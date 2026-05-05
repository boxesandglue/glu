// Package htmlbag exposes a thin Lua bridge that lets a Lua script
// hand an HTML payload directly to glu's HTML rendering pipeline,
// skipping the disk roundtrip of writing the HTML to a file just so
// `glu file.html` can read it back.
//
// The intended user is something like the XSL-FO walker
// (boxesandglue-examples/glu/xslfo/foproc.lua), which builds HTML
// in memory and now calls htmlbag.render(html_str, "out.pdf")
// instead of writing out.html and asking the user to invoke glu
// a second time.
package htmlbag

import (
	"github.com/boxesandglue/glu/markdown"
	"github.com/speedata/go-lua"
)

// luaRender is the Lua-callable entry point.
//   render(html_string, output_pdf [, base_dir])
//
// base_dir defaults to "." and is used to resolve relative CSS paths
// (e.g. <link rel="stylesheet" href="…">) inside the HTML payload.
func luaRender(l *lua.State) int {
	htmlStr := lua.CheckString(l, 1)
	outputPDF := lua.CheckString(l, 2)
	baseDir := "."
	if l.Top() >= 3 {
		if s, ok := l.ToString(3); ok {
			baseDir = s
		}
	}
	if err := markdown.ProcessHTMLString(l, htmlStr, baseDir, outputPDF, markdown.Options{}); err != nil {
		lua.Errorf(l, "glu.htmlbag.render: %v", err.Error())
	}
	return 0
}

func openHtmlbag(l *lua.State) int {
	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "render", Function: luaRender},
	})
	return 1
}

// Open registers the glu.htmlbag module for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "glu.htmlbag", openHtmlbag, false)
	l.Pop(1)
}
