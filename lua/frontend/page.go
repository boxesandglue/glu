package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
	"github.com/speedata/glu/lua/backend"
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
// vlist can be either a frontend VList or a backend node.VList
func pageOutputAt(l *lua.State) int {
	p := checkPage(l, 1)
	x := checkDimension(l, 2)
	y := checkDimension(l, 3)

	// Try to get the VList - accept both frontend VList and backend node.VList
	var vl *node.VList
	if ud := lua.TestUserData(l, 4, vlistMetaTable); ud != nil {
		if v, ok := ud.(*VList); ok {
			vl = v.Value
		}
	} else if ud := lua.TestUserData(l, 4, "node.VList"); ud != nil {
		// Backend NodeVList wrapper
		if v, ok := ud.(*backend.NodeVList); ok {
			vl = v.Value
		}
	}

	if vl == nil {
		lua.Errorf(l, "VList expected")
		return 0
	}

	p.Value.OutputAt(x, y, vl)

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
		pushScaledPoint(l, p.Value.Width)
		return 1
	case "height":
		pushScaledPoint(l, p.Value.Height)
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
