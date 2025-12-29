package backend

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/font"
	"github.com/boxesandglue/textlayout/harfbuzz"
	"github.com/speedata/go-lua"
)

const fontMetaTable = "font.Font"
const atomMetaTable = "font.Atom"

// Font wraps the boxesandglue font.Font type
type Font struct {
	Value *font.Font
}

// Atom wraps the boxesandglue font.Atom type
type Atom struct {
	Value font.Atom
}

// checkFont retrieves a Font userdata from the stack
func checkFont(l *lua.State, index int) *Font {
	ud := lua.CheckUserData(l, index, fontMetaTable)
	if f, ok := ud.(*Font); ok {
		return f
	}
	lua.Errorf(l, "Font expected")
	return nil
}

// fontNew creates a new Font from a Face and size: font.new(face, size)
func fontNew(l *lua.State) int {
	// Get face from pdf module's Face type
	ud := l.ToUserData(1)
	var face *pdf.Face

	// Try to get face from different userdata types
	switch v := ud.(type) {
	case *struct{ Value *pdf.Face }:
		face = v.Value
	default:
		// Try to access Value field via reflection-like approach
		lua.Errorf(l, "Face expected as first argument")
		return 0
	}

	size := bag.ScaledPoint(lua.CheckInteger(l, 2))
	fnt := font.NewFont(face, size)

	l.PushUserData(&Font{Value: fnt})
	lua.SetMetaTableNamed(l, fontMetaTable)
	return 1
}

// fontShape shapes text with the font: font:shape(text, features...)
func fontShape(l *lua.State) int {
	f := checkFont(l, 1)
	text := lua.CheckString(l, 2)

	// Collect features
	var features []harfbuzz.Feature
	n := l.Top()
	for i := 3; i <= n; i++ {
		if l.IsString(i) {
			featStr, _ := l.ToString(i)
			feat, err := harfbuzz.ParseFeature(featStr)
			if err == nil {
				features = append(features, feat)
			}
		}
	}

	atoms := f.Value.Shape(text, features)

	// Return as array of atoms
	l.NewTable()
	for i, a := range atoms {
		l.PushUserData(&Atom{Value: a})
		lua.SetMetaTableNamed(l, atomMetaTable)
		l.RawSetInt(-2, i+1)
	}
	return 1
}

// fontIndex handles attribute access
func fontIndex(l *lua.State) int {
	f := checkFont(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "shape":
		l.PushGoFunction(fontShape)
		return 1
	case "size":
		l.PushInteger(int(f.Value.Size))
		return 1
	case "space":
		l.PushInteger(int(f.Value.Space))
		return 1
	case "space_stretch":
		l.PushInteger(int(f.Value.SpaceStretch))
		return 1
	case "space_shrink":
		l.PushInteger(int(f.Value.SpaceShrink))
		return 1
	}
	return 0
}

// atomIndex handles Atom attribute access
func atomIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, atomMetaTable)
	a := ud.(*Atom).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "advance":
		l.PushInteger(int(a.Advance))
		return 1
	case "height":
		l.PushInteger(int(a.Height))
		return 1
	case "depth":
		l.PushInteger(int(a.Depth))
		return 1
	case "codepoint":
		l.PushInteger(a.Codepoint)
		return 1
	case "components":
		l.PushString(a.Components)
		return 1
	case "is_space":
		l.PushBoolean(a.IsSpace)
		return 1
	case "hyphenate":
		l.PushBoolean(a.Hyphenate)
		return 1
	case "kern_after":
		l.PushInteger(int(a.Kernafter))
		return 1
	}
	return 0
}

// atomToString returns a string representation
func atomToString(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, atomMetaTable)
	a := ud.(*Atom).Value
	l.PushString(a.Components)
	return 1
}

// registerFontMetaTables creates font metatables
func registerFontMetaTables(l *lua.State) {
	// Font metatable
	lua.NewMetaTable(l, fontMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: fontIndex},
	}, 0)
	l.Pop(1)

	// Atom metatable
	lua.NewMetaTable(l, atomMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: atomIndex},
		{Name: "__tostring", Function: atomToString},
	}, 0)
	l.Pop(1)
}

// openFont creates the font module table for require("glu.font")
func openFont(l *lua.State) int {
	registerFontMetaTables(l)

	l.NewTable()
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "new", Function: fontNew},
	}, 0)
	return 1
}
