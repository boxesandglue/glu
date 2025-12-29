package frontend

import (
	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
)

const imagefileMetaTable = "FrontendImagefile"
const imageNodeMetaTable = "ImageNode"

// Imagefile wraps the baseline-pdf Imagefile type
type Imagefile struct {
	Value *pdf.Imagefile
}

// ImageNode wraps the node.Image type
type ImageNode struct {
	Value *node.Image
}

// checkImagefile retrieves an Imagefile userdata from the stack
func checkImagefile(l *lua.State, index int) *Imagefile {
	ud := lua.CheckUserData(l, index, imagefileMetaTable)
	if img, ok := ud.(*Imagefile); ok {
		return img
	}
	lua.Errorf(l, "Imagefile expected")
	return nil
}

// checkImageNode retrieves an ImageNode userdata from the stack
func checkImageNode(l *lua.State, index int) *ImageNode {
	ud := lua.CheckUserData(l, index, imageNodeMetaTable)
	if img, ok := ud.(*ImageNode); ok {
		return img
	}
	lua.Errorf(l, "ImageNode expected")
	return nil
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
	}
	return 0
}

// imageNodeIndex handles attribute access (__index metamethod)
func imageNodeIndex(l *lua.State) int {
	img := checkImageNode(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		pushScaledPoint(l, img.Value.Width)
		return 1
	case "height":
		pushScaledPoint(l, img.Value.Height)
		return 1
	}
	return 0
}

// imageNodeNewIndex handles attribute setting (__newindex metamethod)
func imageNodeNewIndex(l *lua.State) int {
	img := checkImageNode(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		img.Value.Width = bag.ScaledPoint(checkDimension(l, 3))
	case "height":
		img.Value.Height = bag.ScaledPoint(checkDimension(l, 3))
	default:
		lua.Errorf(l, "cannot set attribute %s on ImageNode", key)
	}
	return 0
}

// registerImagefileMetaTable creates the Imagefile metatable
func registerImagefileMetaTable(l *lua.State) {
	lua.NewMetaTable(l, imagefileMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: imagefileIndex},
	}, 0)
	l.Pop(1)
}

// registerImageNodeMetaTable creates the ImageNode metatable
func registerImageNodeMetaTable(l *lua.State) {
	lua.NewMetaTable(l, imageNodeMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: imageNodeIndex},
		{Name: "__newindex", Function: imageNodeNewIndex},
	}, 0)
	l.Pop(1)
}
