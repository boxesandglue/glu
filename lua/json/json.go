package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boxesandglue/glu/lua/common"
	"github.com/speedata/go-lua"
)

// luaDecode implements json.decode(str) → table.
func luaDecode(l *lua.State) int {
	s := lua.CheckString(l, 1)
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		lua.Errorf(l, "json.decode: %s", err.Error())
		return 0
	}
	common.PushAny(l, v)
	return 1
}

// luaEncode implements json.encode(table [, indent]) → string.
// If indent is true or a string, the output is pretty-printed.
func luaEncode(l *lua.State) int {
	lua.CheckType(l, 1, lua.TypeTable)
	v := common.LuaToGo(l, 1)

	var data []byte
	var err error
	if !l.IsNoneOrNil(2) {
		indent := "  "
		if l.IsString(2) {
			indent, _ = l.ToString(2)
		}
		data, err = json.MarshalIndent(v, "", indent)
	} else {
		data, err = json.Marshal(v)
	}
	if err != nil {
		lua.Errorf(l, "json.encode: %s", err.Error())
		return 0
	}
	l.PushString(string(data))
	return 1
}

// luaRead implements json.read(filename) → table.
func luaRead(l *lua.State) int {
	filename := lua.CheckString(l, 1)
	data, err := os.ReadFile(filename)
	if err != nil {
		lua.Errorf(l, "json.read: %s", err.Error())
		return 0
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		lua.Errorf(l, "json.read: %s", err.Error())
		return 0
	}
	common.PushAny(l, v)
	return 1
}

// luaWrite implements json.write(filename, table [, indent]).
func luaWrite(l *lua.State) int {
	filename := lua.CheckString(l, 1)
	lua.CheckType(l, 2, lua.TypeTable)
	v := common.LuaToGo(l, 2)

	var data []byte
	var err error
	if !l.IsNoneOrNil(3) {
		indent := "  "
		if l.IsString(3) {
			indent, _ = l.ToString(3)
		}
		data, err = json.MarshalIndent(v, "", indent)
	} else {
		data, err = json.MarshalIndent(v, "", "  ")
	}
	if err != nil {
		lua.Errorf(l, "json.write: %s", err.Error())
		return 0
	}
	if err := os.WriteFile(filename, append(data, '\n'), 0644); err != nil {
		lua.Errorf(l, "json.write: %s", err.Error())
		return 0
	}
	return 0
}

func openJSON(l *lua.State) int {
	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "decode", Function: luaDecode},
		{Name: "encode", Function: luaEncode},
		{Name: "read", Function: luaRead},
		{Name: "write", Function: luaWrite},
	})
	// Add module description
	l.PushString(fmt.Sprintf("glu.json (%d functions)", 4))
	l.SetField(-2, "_description")
	return 1
}

// Open registers the json module for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "glu.json", openJSON, false)
	l.Pop(1)
}
