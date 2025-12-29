package frontend

import (
	"fmt"

	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

const fontFamilyMetaTable = "FontFamily"
const fontSourceMetaTable = "FontSource"
const faceMetaTable = "Face"

// FontFamily wraps the boxesandglue frontend.FontFamily type
type FontFamily struct {
	Value *frontend.FontFamily
	doc   *frontend.Document
}

// FontSource wraps the boxesandglue frontend.FontSource type
type FontSource struct {
	Value *frontend.FontSource
}

// Face wraps the baseline-pdf Face type
type Face struct {
	Value *pdf.Face
}

// checkFontFamily retrieves a FontFamily userdata from the stack
func checkFontFamily(l *lua.State, index int) *FontFamily {
	ud := lua.CheckUserData(l, index, fontFamilyMetaTable)
	if ff, ok := ud.(*FontFamily); ok {
		return ff
	}
	lua.Errorf(l, "FontFamily expected")
	return nil
}

// checkFontSource retrieves a FontSource userdata from the stack
func checkFontSource(l *lua.State, index int) *FontSource {
	ud := lua.CheckUserData(l, index, fontSourceMetaTable)
	if fs, ok := ud.(*FontSource); ok {
		return fs
	}
	lua.Errorf(l, "FontSource expected")
	return nil
}

// fontSourceNew creates a new FontSource: fontsource.new(options)
func fontSourceNew(l *lua.State) int {
	fs := &frontend.FontSource{}

	if l.IsTable(1) {
		l.Field(1, "location")
		if l.IsString(-1) {
			fs.Location, _ = l.ToString(-1)
		}
		l.Pop(1)

		l.Field(1, "name")
		if l.IsString(-1) {
			fs.Name, _ = l.ToString(-1)
		}
		l.Pop(1)

		l.Field(1, "index")
		if l.IsNumber(-1) {
			idx, _ := l.ToInteger(-1)
			fs.Index = idx
		}
		l.Pop(1)

		l.Field(1, "size_adjust")
		if l.IsNumber(-1) {
			fs.SizeAdjust, _ = l.ToNumber(-1)
		}
		l.Pop(1)

		// Parse features as array of strings
		l.Field(1, "features")
		if l.IsTable(-1) {
			l.PushNil()
			for l.Next(-2) {
				if l.IsString(-1) {
					feature, _ := l.ToString(-1)
					fs.FontFeatures = append(fs.FontFeatures, feature)
				}
				l.Pop(1)
			}
		}
		l.Pop(1)
	} else if l.IsString(1) {
		// Simple case: just a filename
		fs.Location = lua.CheckString(l, 1)
	}

	l.PushUserData(&FontSource{Value: fs})
	lua.SetMetaTableNamed(l, fontSourceMetaTable)
	return 1
}

// fontFamilyAddMember adds a font member: ff:add_member(fontsource, weight, style)
// or ff:add_member({source = fs, weight = 400, style = "normal"})
func fontFamilyAddMember(l *lua.State) int {
	ff := checkFontFamily(l, 1)

	var fs *FontSource
	var weightStr string = "400"
	var styleStr string = "normal"

	if l.IsTable(2) {
		// Table-based call: {source = fs, weight = 400, style = "normal"}
		l.Field(2, "source")
		if ud := lua.TestUserData(l, -1, fontSourceMetaTable); ud != nil {
			if f, ok := ud.(*FontSource); ok {
				fs = f
			}
		}
		l.Pop(1)

		l.Field(2, "weight")
		if l.IsNumber(-1) {
			n, _ := l.ToInteger(-1)
			weightStr = fmt.Sprintf("%d", n)
		} else if l.IsString(-1) {
			weightStr, _ = l.ToString(-1)
		}
		l.Pop(1)

		l.Field(2, "style")
		if l.IsString(-1) {
			styleStr, _ = l.ToString(-1)
		}
		l.Pop(1)

		if fs == nil {
			lua.Errorf(l, "add_member: source is required")
			return 0
		}
	} else {
		// Positional call: ff:add_member(fontsource, weight, style)
		fs = checkFontSource(l, 2)
		weightStr = lua.OptString(l, 3, "400")
		styleStr = lua.OptString(l, 4, "normal")
	}

	weight := frontend.ResolveFontWeight(weightStr, frontend.FontWeight400)
	style := frontend.ResolveFontStyle(styleStr)

	if err := ff.Value.AddMember(fs.Value, weight, style); err != nil {
		lua.Errorf(l, "failed to add member: %s", err.Error())
		return 0
	}

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// fontFamilyIndex handles attribute access (__index metamethod)
func fontFamilyIndex(l *lua.State) int {
	ff := checkFontFamily(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "name":
		l.PushString(ff.Value.Name)
		return 1
	case "add_member":
		l.PushGoFunction(fontFamilyAddMember)
		return 1
	}

	return 0
}

// fontSourceIndex handles attribute access (__index metamethod)
func fontSourceIndex(l *lua.State) int {
	fs := checkFontSource(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "location":
		l.PushString(fs.Value.Location)
		return 1
	case "name":
		l.PushString(fs.Value.Name)
		return 1
	case "index":
		l.PushInteger(fs.Value.Index)
		return 1
	}

	return 0
}

// faceIndex handles attribute access (__index metamethod)
func faceIndex(l *lua.State) int {
	f := checkFace(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "internal_name":
		l.PushString(f.Value.InternalName())
		return 1
	case "postscript_name":
		l.PushString(f.Value.PostscriptName)
		return 1
	}

	return 0
}

func checkFace(l *lua.State, index int) *Face {
	ud := lua.CheckUserData(l, index, faceMetaTable)
	if f, ok := ud.(*Face); ok {
		return f
	}
	lua.Errorf(l, "Face expected")
	return nil
}

// registerFontFamilyMetaTable creates the FontFamily metatable
func registerFontFamilyMetaTable(l *lua.State) {
	lua.NewMetaTable(l, fontFamilyMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: fontFamilyIndex},
	}, 0)
	l.Pop(1)
}

// registerFontSourceMetaTable creates the FontSource metatable
func registerFontSourceMetaTable(l *lua.State) {
	lua.NewMetaTable(l, fontSourceMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: fontSourceIndex},
	}, 0)
	l.Pop(1)
}

// registerFaceMetaTable creates the Face metatable
func registerFaceMetaTable(l *lua.State) {
	lua.NewMetaTable(l, faceMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: faceIndex},
	}, 0)
	l.Pop(1)
}
