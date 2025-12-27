package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/speedata/go-lua"
)

const pageMetaTable = "Page"

// Page wraps the boxesandglue document.Page type
type Page struct {
	Value *document.Page
}

// checkPage retrieves a Page userdata from the stack
func checkPage(l *lua.State, index int) *Page {
	ud := lua.CheckUserData(l, index, pageMetaTable)
	if p, ok := ud.(*Page); ok {
		return p
	}
	lua.Errorf(l, "Page expected")
	return nil
}

// pageOutputAt places a VList at position: page:output_at(x, y, vlist)
// x, y can be numbers (points) or strings with units ("72pt", "1in", "2cm")
func pageOutputAt(l *lua.State) int {
	p := checkPage(l, 1)
	x := checkDimension(l, 2)
	y := checkDimension(l, 3)
	vl := checkVList(l, 4)

	p.Value.OutputAt(x, y, vl.Value)

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// pageShipout finalizes the page: page:shipout()
func pageShipout(l *lua.State) int {
	p := checkPage(l, 1)
	p.Value.Shipout()
	return 0
}

// pageIndex handles attribute access (__index metamethod)
func pageIndex(l *lua.State) int {
	p := checkPage(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "output_at":
		l.PushGoFunction(pageOutputAt)
		return 1
	case "shipout":
		l.PushGoFunction(pageShipout)
		return 1
	case "width":
		l.PushNumber(p.Value.Width.ToPT())
		return 1
	case "height":
		l.PushNumber(p.Value.Height.ToPT())
		return 1
	}

	return 0
}

// pageNewIndex handles attribute setting (__newindex metamethod)
func pageNewIndex(l *lua.State) int {
	p := checkPage(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		p.Value.Width = checkDimension(l, 3)
	case "height":
		p.Value.Height = checkDimension(l, 3)
	}

	return 0
}

// registerPageMetaTable creates the Page metatable
func registerPageMetaTable(l *lua.State) {
	lua.NewMetaTable(l, pageMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: pageIndex},
		{Name: "__newindex", Function: pageNewIndex},
	}, 0)
	l.Pop(1)
}
