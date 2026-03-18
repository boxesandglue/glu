package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/speedata/go-lua"
)

const structureElementMetaTable = "StructureElement"

// StructureElement wraps a document.StructureElement for Lua access.
type StructureElement struct {
	Value *document.StructureElement
}

func checkStructureElement(l *lua.State, index int) *StructureElement {
	ud := lua.CheckUserData(l, index, structureElementMetaTable)
	if se, ok := ud.(*StructureElement); ok {
		return se
	}
	lua.Errorf(l, "StructureElement expected")
	return nil
}

// seAddChild adds a child: se:add_child(child)
func seAddChild(l *lua.State) int {
	se := checkStructureElement(l, 1)
	child := checkStructureElement(l, 2)
	se.Value.AddChild(child.Value)
	return 0
}

// seIndex handles attribute access (__index metamethod)
func seIndex(l *lua.State) int {
	se := checkStructureElement(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "role":
		l.PushString(se.Value.Role)
		return 1
	case "alt":
		l.PushString(se.Value.Alt)
		return 1
	case "actual_text":
		l.PushString(se.Value.ActualText)
		return 1
	case "lang":
		l.PushString(se.Value.Lang)
		return 1
	case "add_child":
		l.PushGoFunction(seAddChild)
		return 1
	}
	return 0
}

// seNewIndex handles attribute setting (__newindex metamethod)
func seNewIndex(l *lua.State) int {
	se := checkStructureElement(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "role":
		se.Value.Role = lua.CheckString(l, 3)
	case "alt":
		se.Value.Alt = lua.CheckString(l, 3)
	case "actual_text":
		se.Value.ActualText = lua.CheckString(l, 3)
	case "lang":
		se.Value.Lang = lua.CheckString(l, 3)
	default:
		lua.Errorf(l, "cannot set attribute %s on StructureElement", key)
	}
	return 0
}

func registerStructureElementMetaTable(l *lua.State) {
	lua.NewMetaTable(l, structureElementMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: seIndex},
		{Name: "__newindex", Function: seNewIndex},
	}, 0)
	l.Pop(1)
}
