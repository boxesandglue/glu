package pdf

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/speedata/go-lua"
)

const faceMetaTable = "PDFFace"

// Face wraps the baseline-pdf Face type
type Face struct {
	Value *pdf.Face
}

// checkFace retrieves a Face userdata from the stack
func checkFace(l *lua.State, index int) *Face {
	ud := lua.CheckUserData(l, index, faceMetaTable)
	if f, ok := ud.(*Face); ok {
		return f
	}
	lua.Errorf(l, "PDFFace expected")
	return nil
}

// faceRegisterCodepoint registers a single codepoint: face:register_codepoint(cp)
func faceRegisterCodepoint(l *lua.State) int {
	face := checkFace(l, 1)
	cp := lua.CheckInteger(l, 2)
	face.Value.RegisterCodepoint(cp)
	return 0
}

// faceRegisterCodepoints registers multiple codepoints: face:register_codepoints(array)
func faceRegisterCodepoints(l *lua.State) int {
	face := checkFace(l, 1)
	if l.IsTable(2) {
		codepoints := tableToInts(l, 2)
		face.Value.RegisterCodepoints(codepoints)
	}
	return 0
}

// faceCodepoint returns the glyph index for a codepoint: face:codepoint(cp)
func faceCodepoint(l *lua.State) int {
	face := checkFace(l, 1)
	cp := lua.CheckInteger(l, 2)
	gid := face.Value.Codepoint(rune(cp))
	l.PushInteger(gid)
	return 1
}

// faceIndex handles attribute access (__index metamethod)
func faceIndex(l *lua.State) int {
	face := checkFace(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "internal_name":
		l.PushString(face.Value.InternalName())
		return 1
	case "units_per_em":
		l.PushInteger(int(face.Value.UnitsPerEM))
		return 1
	case "postscript_name":
		l.PushString(face.Value.PostscriptName)
		return 1
	case "face_id":
		l.PushInteger(face.Value.FaceID)
		return 1
	// Methods
	case "register_codepoint":
		l.PushGoFunction(faceRegisterCodepoint)
		return 1
	case "register_codepoints":
		l.PushGoFunction(faceRegisterCodepoints)
		return 1
	case "codepoint":
		l.PushGoFunction(faceCodepoint)
		return 1
	}
	return 0
}

// registerFaceMetaTable creates the PDFFace metatable
func registerFaceMetaTable(l *lua.State) {
	lua.NewMetaTable(l, faceMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: faceIndex},
	}, 0)
	l.Pop(1)
}
