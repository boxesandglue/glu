package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/speedata/go-lua"
)

const colorProfileMetaTable = "ColorProfile"

// ColorProfile wraps the document.ColorProfile type
type ColorProfile struct {
	Value *document.ColorProfile
}

// checkColorProfile retrieves a ColorProfile userdata from the stack
func checkColorProfile(l *lua.State, index int) *ColorProfile {
	ud := lua.CheckUserData(l, index, colorProfileMetaTable)
	if cp, ok := ud.(*ColorProfile); ok {
		return cp
	}
	lua.Errorf(l, "ColorProfile expected")
	return nil
}

// colorProfileIndex handles attribute access (__index metamethod)
func colorProfileIndex(l *lua.State) int {
	cp := checkColorProfile(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "identifier":
		l.PushString(cp.Value.Identifier)
		return 1
	case "registry":
		l.PushString(cp.Value.Registry)
		return 1
	case "info":
		l.PushString(cp.Value.Info)
		return 1
	case "condition":
		l.PushString(cp.Value.Condition)
		return 1
	case "colors":
		l.PushInteger(cp.Value.Colors)
		return 1
	}
	return 0
}

// colorProfileNewIndex handles attribute setting (__newindex metamethod)
func colorProfileNewIndex(l *lua.State) int {
	cp := checkColorProfile(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "identifier":
		cp.Value.Identifier = lua.CheckString(l, 3)
	case "registry":
		cp.Value.Registry = lua.CheckString(l, 3)
	case "info":
		cp.Value.Info = lua.CheckString(l, 3)
	case "condition":
		cp.Value.Condition = lua.CheckString(l, 3)
	case "colors":
		cp.Value.Colors = lua.CheckInteger(l, 3)
	default:
		lua.Errorf(l, "cannot set attribute %s on ColorProfile", key)
	}
	return 0
}

// registerColorProfileMetaTable creates the ColorProfile metatable
func registerColorProfileMetaTable(l *lua.State) {
	lua.NewMetaTable(l, colorProfileMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: colorProfileIndex},
		{Name: "__newindex", Function: colorProfileNewIndex},
	}, 0)
	l.Pop(1)
}
