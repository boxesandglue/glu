package frontend

import (
	"github.com/speedata/go-lua"
)

// openFrontend creates the frontend module table for require("glu.frontend")
func openFrontend(l *lua.State) int {
	// Register all metatables
	registerScaledPointMetaTable(l)
	registerDocumentMetaTable(l)
	registerPageMetaTable(l)
	registerTextMetaTable(l)
	registerTextSettingsMetaTable(l)
	registerFontFamilyMetaTable(l)
	registerFontSourceMetaTable(l)
	registerFaceMetaTable(l)
	registerVListMetaTable(l)
	registerTableMetaTable(l)
	registerTableRowMetaTable(l)
	registerTableCellMetaTable(l)
	registerColorMetaTable(l)
	registerLanguageMetaTable(l)
	registerImagefileMetaTable(l)
	registerImageNodeMetaTable(l)
	registerColorProfileMetaTable(l)
	registerSVGDocumentMetaTable(l)
	registerSVGNodeMetaTable(l)
	registerPageInfoMetaTable(l)
	registerElementInfoMetaTable(l)

	// Create the frontend module table
	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "new", Function: documentNew},
		{Name: "text", Function: textNew},
		{Name: "fontsource", Function: fontSourceNew},
		{Name: "color", Function: colorNew},
		{Name: "table", Function: tableNew},
		{Name: "sp", Function: spNew},
		{Name: "sp_string", Function: spFromString},
		{Name: "parse_svg", Function: parseSVGFromFile},
		{Name: "parse_svg_string", Function: parseSVGFromString},
		{Name: "add_callback", Function: luaAddCallback},
		{Name: "remove_callback", Function: luaRemoveCallback},
		{Name: "list_callbacks", Function: luaListCallbacks},
	})
	return 1
}

// Open registers the frontend module for require() in the Lua state.
func Open(l *lua.State) {
	InitRegistry(l)
	lua.Require(l, "glu.frontend", openFrontend, false)
	l.Pop(1)
}
