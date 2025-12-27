package pdf

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/speedata/go-lua"
)

const pageMetaTable = "PDFPage"

// Page wraps the baseline-pdf Page type
type Page struct {
	Value *pdf.Page
}

// checkPage retrieves a Page userdata from the stack
func checkPage(l *lua.State, index int) *Page {
	ud := lua.CheckUserData(l, index, pageMetaTable)
	if p, ok := ud.(*Page); ok {
		return p
	}
	lua.Errorf(l, "PDFPage expected")
	return nil
}

// pageIndex handles attribute access (__index metamethod)
func pageIndex(l *lua.State) int {
	pg := checkPage(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushNumber(pg.Value.Width)
		return 1
	case "height":
		l.PushNumber(pg.Value.Height)
		return 1
	case "offset_x":
		l.PushNumber(pg.Value.OffsetX)
		return 1
	case "offset_y":
		l.PushNumber(pg.Value.OffsetY)
		return 1
	case "object_number":
		l.PushInteger(int(pg.Value.Objnum))
		return 1
	}
	return 0
}

// pageNewIndex handles attribute setting (__newindex metamethod)
func pageNewIndex(l *lua.State) int {
	pg := checkPage(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		pg.Value.Width = lua.CheckNumber(l, 3)
	case "height":
		pg.Value.Height = lua.CheckNumber(l, 3)
	case "offset_x":
		pg.Value.OffsetX = lua.CheckNumber(l, 3)
	case "offset_y":
		pg.Value.OffsetY = lua.CheckNumber(l, 3)
	case "faces":
		if l.IsTable(3) {
			pg.Value.Faces = tableToFaces(l, 3)
		}
	case "images":
		if l.IsTable(3) {
			pg.Value.Images = tableToImages(l, 3)
		}
	default:
		lua.Errorf(l, "cannot set attribute %s on PDFPage", key)
	}
	return 0
}

// registerPageMetaTable creates the PDFPage metatable
func registerPageMetaTable(l *lua.State) {
	lua.NewMetaTable(l, pageMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: pageIndex},
		{Name: "__newindex", Function: pageNewIndex},
	}, 0)
	l.Pop(1)
}
