package pdf

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/speedata/go-lua"
)

const imagefileMetaTable = "PDFImagefile"

// Imagefile wraps the baseline-pdf Imagefile type
type Imagefile struct {
	Value *pdf.Imagefile
}

// checkImagefile retrieves an Imagefile userdata from the stack
func checkImagefile(l *lua.State, index int) *Imagefile {
	ud := lua.CheckUserData(l, index, imagefileMetaTable)
	if img, ok := ud.(*Imagefile); ok {
		return img
	}
	lua.Errorf(l, "PDFImagefile expected")
	return nil
}

// imagefileClose closes the image file: img:close()
func imagefileClose(l *lua.State) int {
	img := checkImagefile(l, 1)
	if err := img.Value.Close(); err != nil {
		lua.Errorf(l, "close error: %s", err.Error())
	}
	return 0
}

// imagefileIndex handles attribute access (__index metamethod)
func imagefileIndex(l *lua.State) int {
	img := checkImagefile(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "internal_name":
		l.PushString(img.Value.InternalName())
		return 1
	case "format":
		l.PushString(img.Value.Format)
		return 1
	case "filename":
		l.PushString(img.Value.Filename)
		return 1
	case "width":
		l.PushInteger(img.Value.W)
		return 1
	case "height":
		l.PushInteger(img.Value.H)
		return 1
	case "scale_x":
		l.PushNumber(img.Value.ScaleX)
		return 1
	case "scale_y":
		l.PushNumber(img.Value.ScaleY)
		return 1
	case "number_of_pages":
		l.PushInteger(img.Value.NumberOfPages)
		return 1
	// Methods
	case "close":
		l.PushGoFunction(imagefileClose)
		return 1
	}
	return 0
}

// registerImagefileMetaTable creates the PDFImagefile metatable
func registerImagefileMetaTable(l *lua.State) {
	lua.NewMetaTable(l, imagefileMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: imagefileIndex},
	}, 0)
	l.Pop(1)
}
