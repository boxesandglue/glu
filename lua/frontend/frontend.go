package frontend

import (
	"github.com/speedata/go-lua"
)

// Open registers the frontend module with the Lua state
func Open(l *lua.State) {
	// Register all metatables
	registerDocumentMetaTable(l)
	registerPageMetaTable(l)
	registerTextMetaTable(l)
	registerFontFamilyMetaTable(l)
	registerFontSourceMetaTable(l)
	registerFaceMetaTable(l)
	registerVListMetaTable(l)
	registerTableMetaTable(l)
	registerTableRowMetaTable(l)
	registerTableCellMetaTable(l)
	registerColorMetaTable(l)
	registerLanguageMetaTable(l)

	// Create the frontend module table
	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "new", Function: documentNew},
		{Name: "text", Function: textNew},
		{Name: "fontsource", Function: fontSourceNew},
		{Name: "color", Function: colorNew},
		{Name: "table", Function: tableNew},
		{Name: "sp", Function: spNew},
		{Name: "sp_string", Function: spFromString},
	})
	l.SetGlobal("frontend")
}
