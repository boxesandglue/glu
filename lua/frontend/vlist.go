package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/document"
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

// vlistSetTag tags a VList with a structure element: vl:set_tag(se)
func vlistSetTag(l *lua.State) int {
	vl := checkVList(l, 1)
	se := checkStructureElement(l, 2)
	if vl.Value.Attributes == nil {
		vl.Value.Attributes = node.H{}
	}
	vl.Value.Attributes["tag"] = se.Value
	return 0
}

// vlistSetArtifact marks a VList as artifact: vl:set_artifact([type])
func vlistSetArtifact(l *lua.State) int {
	vl := checkVList(l, 1)
	artType := lua.OptString(l, 2, "")
	if vl.Value.Attributes == nil {
		vl.Value.Attributes = node.H{}
	}
	vl.Value.Attributes["artifact"] = document.ArtifactType(artType)
	return 0
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
	case "set_tag":
		l.PushGoFunction(vlistSetTag)
		return 1
	case "set_artifact":
		l.PushGoFunction(vlistSetArtifact)
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
