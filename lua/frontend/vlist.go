package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
)

const vlistMetaTable = "VList"

// VList wraps the boxesandglue node.VList type
type VList struct {
	Value *node.VList
}

// checkVList retrieves a VList userdata from the stack
func checkVList(l *lua.State, index int) *VList {
	ud := lua.CheckUserData(l, index, vlistMetaTable)
	if v, ok := ud.(*VList); ok {
		return v
	}
	lua.Errorf(l, "VList expected")
	return nil
}

// vlistIndex handles attribute access (__index metamethod)
func vlistIndex(l *lua.State) int {
	vl := checkVList(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		pushScaledPoint(l, vl.Value.Width)
		return 1
	case "height":
		pushScaledPoint(l, vl.Value.Height)
		return 1
	case "depth":
		pushScaledPoint(l, vl.Value.Depth)
		return 1
	}

	return 0
}

// registerVListMetaTable creates the VList metatable
func registerVListMetaTable(l *lua.State) {
	lua.NewMetaTable(l, vlistMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: vlistIndex},
	}, 0)
	l.Pop(1)
}
