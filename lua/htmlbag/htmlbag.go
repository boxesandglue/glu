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
	"github.com/boxesandglue/glu/lua/common"
	"github.com/boxesandglue/glu/markdown"
	"github.com/speedata/go-lua"
)

// luaRender is the Lua-callable entry point.
//
//	render(html_string, output_pdf [, base_dir_or_options])
//
// The third argument is either:
//
//   - a string: treated as base_dir for resolving relative CSS paths
//     (e.g. <link rel="stylesheet" href="…">) inside the HTML payload;
//
//   - a table: an options dict with the following recognised keys:
//
//     base_dir = "."          -- relative CSS path root
//     format   = "PDF/UA"     -- PDF conformance level
//     lang     = "en-US"      -- BCP47, written to PDF /Lang
//     title    = "Showcase"   -- PDF /Title (also XMP dc:title)
//
// Unknown keys are silently ignored. base_dir defaults to ".".
func luaRender(l *lua.State) int {
	htmlStr := lua.CheckString(l, 1)
	outputPDF := lua.CheckString(l, 2)
	baseDir := "."
	opts := markdown.Options{}
	if l.Top() >= 3 {
		switch {
		case l.IsString(3):
			s, _ := l.ToString(3)
			baseDir = s
		case l.IsTable(3):
			if m, ok := common.LuaTableToGo(l, 3).(map[string]any); ok {
				if v, ok := m["base_dir"].(string); ok && v != "" {
					baseDir = v
				}
				if v, ok := m["format"].(string); ok {
					opts.Format = v
				}
				if v, ok := m["lang"].(string); ok {
					opts.Lang = v
				}
				if v, ok := m["title"].(string); ok {
					opts.Title = v
				}
			}
		}
	}
	if err := markdown.ProcessHTMLString(l, htmlStr, baseDir, outputPDF, opts); err != nil {
		// lua.Errorf forwards to lua_pushfstring which does NOT support %v;
		// only %s, %d, %f, %p, %c, %U are honoured. Stick to %s with the
		// pre-stringified message.
		lua.Errorf(l, "glu.htmlbag.render: %s", err.Error())
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
