package pdf

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/speedata/go-lua"
)

// tableToDict converts a Lua table to a PDF Dict
func tableToDict(l *lua.State, index int) pdf.Dict {
	dict := make(pdf.Dict)
	absIndex := l.AbsIndex(index)

	l.PushNil() // First key
	for l.Next(absIndex) {
		// Key is at -2, value at -1
		if l.IsString(-2) {
			keyStr, _ := l.ToString(-2)
			key := pdf.Name(keyStr)

			switch {
			case l.IsString(-1):
				val, _ := l.ToString(-1)
				dict[key] = val
			case l.IsNumber(-1):
				val, _ := l.ToNumber(-1)
				dict[key] = val
			case l.IsBoolean(-1):
				dict[key] = l.ToBoolean(-1)
			case l.IsTable(-1):
				dict[key] = tableToDict(l, l.AbsIndex(-1))
			}
		}
		l.Pop(1) // Remove value, keep key for next iteration
	}

	return dict
}

// tableToInts converts a Lua array to []int
func tableToInts(l *lua.State, index int) []int {
	var result []int
	absIndex := l.AbsIndex(index)

	l.PushNil()
	for l.Next(absIndex) {
		if l.IsNumber(-1) {
			val, _ := l.ToNumber(-1)
			result = append(result, int(val))
		}
		l.Pop(1)
	}

	return result
}

// tableToFaces converts a Lua array to []*pdf.Face
func tableToFaces(l *lua.State, index int) []*pdf.Face {
	var result []*pdf.Face
	absIndex := l.AbsIndex(index)

	l.PushNil()
	for l.Next(absIndex) {
		if ud := lua.TestUserData(l, -1, faceMetaTable); ud != nil {
			if f, ok := ud.(*Face); ok {
				result = append(result, f.Value)
			}
		}
		l.Pop(1)
	}

	return result
}

// tableToImages converts a Lua array to []*pdf.Imagefile
func tableToImages(l *lua.State, index int) []*pdf.Imagefile {
	var result []*pdf.Imagefile
	absIndex := l.AbsIndex(index)

	l.PushNil()
	for l.Next(absIndex) {
		if ud := lua.TestUserData(l, -1, imagefileMetaTable); ud != nil {
			if img, ok := ud.(*Imagefile); ok {
				result = append(result, img.Value)
			}
		}
		l.Pop(1)
	}

	return result
}
