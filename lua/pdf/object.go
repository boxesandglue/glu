package pdf

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/speedata/go-lua"
)

const objectMetaTable = "PDFObject"

// Object wraps the baseline-pdf Object type
type Object struct {
	Value *pdf.Object
}

// checkObject retrieves an Object userdata from the stack
func checkObject(l *lua.State, index int) *Object {
	ud := lua.CheckUserData(l, index, objectMetaTable)
	if o, ok := ud.(*Object); ok {
		return o
	}
	lua.Errorf(l, "PDFObject expected")
	return nil
}

// objectSave saves the object: obj:save()
func objectSave(l *lua.State) int {
	obj := checkObject(l, 1)
	if err := obj.Value.Save(); err != nil {
		lua.Errorf(l, "save error: %s", err.Error())
	}
	return 0
}

// objectSetCompression sets compression level: obj:set_compression(level)
func objectSetCompression(l *lua.State) int {
	obj := checkObject(l, 1)
	level := lua.CheckInteger(l, 2)
	obj.Value.SetCompression(uint(level))
	return 0
}

// objectWriteString writes to the data buffer: obj:write(string)
func objectWriteString(l *lua.State) int {
	obj := checkObject(l, 1)
	str := lua.CheckString(l, 2)
	obj.Value.Data.WriteString(str)
	return 0
}

// objectIndex handles attribute access (__index metamethod)
func objectIndex(l *lua.State) int {
	obj := checkObject(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "object_number":
		l.PushInteger(int(obj.Value.ObjectNumber))
		return 1
	case "force_stream":
		l.PushBoolean(obj.Value.ForceStream)
		return 1
	case "raw":
		l.PushBoolean(obj.Value.Raw)
		return 1
	// Methods
	case "save":
		l.PushGoFunction(objectSave)
		return 1
	case "set_compression":
		l.PushGoFunction(objectSetCompression)
		return 1
	case "write":
		l.PushGoFunction(objectWriteString)
		return 1
	}
	return 0
}

// objectNewIndex handles attribute setting (__newindex metamethod)
func objectNewIndex(l *lua.State) int {
	obj := checkObject(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "force_stream":
		obj.Value.ForceStream = l.ToBoolean(3)
	case "raw":
		obj.Value.Raw = l.ToBoolean(3)
	case "dictionary":
		if l.IsTable(3) {
			dict := tableToDict(l, 3)
			obj.Value.Dictionary = dict
		}
	default:
		lua.Errorf(l, "cannot set attribute %s on PDFObject", key)
	}
	return 0
}

// registerObjectMetaTable creates the PDFObject metatable
func registerObjectMetaTable(l *lua.State) {
	lua.NewMetaTable(l, objectMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: objectIndex},
		{Name: "__newindex", Function: objectNewIndex},
	}, 0)
	l.Pop(1)
}
