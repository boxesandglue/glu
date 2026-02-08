package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/boxesandglue/htmlbag"
	"github.com/speedata/go-lua"
)

const elementInfoMetaTable = "ElementInfo"

// ElementInfo wraps an htmlbag.ElementEvent to expose element information
// to Lua callbacks.
type ElementInfo struct {
	event htmlbag.ElementEvent
}

func checkElementInfo(l *lua.State, index int) *ElementInfo {
	ud := lua.CheckUserData(l, index, elementInfoMetaTable)
	if ei, ok := ud.(*ElementInfo); ok {
		return ei
	}
	lua.Errorf(l, "ElementInfo expected")
	return nil
}

func elementInfoIndex(l *lua.State) int {
	ei := checkElementInfo(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "tag_name":
		l.PushString(ei.event.TagName)
		return 1
	case "text_content":
		l.PushString(ei.event.TextContent)
		return 1
	case "width":
		if ei.event.VList != nil {
			pushScaledPoint(l, ei.event.VList.Width)
			return 1
		}
	case "height":
		if ei.event.VList != nil {
			pushScaledPoint(l, ei.event.VList.Height)
			return 1
		}
	case "depth":
		if ei.event.VList != nil {
			pushScaledPoint(l, ei.event.VList.Depth)
			return 1
		}
	case "set_background":
		l.PushGoFunction(elementInfoSetBackground)
		return 1
	}

	l.PushNil()
	return 1
}

// elementInfoSetBackground inserts a background node behind the element's
// content using the negative-kern overlay technique:
//
//	bgNode → Kern(-bgHeight) → original content
//
// The background is rendered first (behind), then the kern moves back up,
// so the original content is drawn on top.
// Accepts SVGNode or VList as the background argument.
func elementInfoSetBackground(l *lua.State) int {
	ei := checkElementInfo(l, 1)
	vl := ei.event.VList
	if vl == nil || vl.List == nil {
		return 0
	}

	var bgNode node.Node
	var bgHeight bag.ScaledPoint

	if ud := lua.TestUserData(l, 2, svgNodeMetaTable); ud != nil {
		if v, ok := ud.(*SVGNode); ok {
			bgNode = v.Value
			bgHeight = v.Value.Height + v.Value.Depth
		}
	} else if ud := lua.TestUserData(l, 2, vlistMetaTable); ud != nil {
		if v, ok := ud.(*VList); ok {
			bgNode = v.Value
			bgHeight = v.Value.Height + v.Value.Depth
		}
	}

	if bgNode == nil {
		lua.Errorf(l, "set_background: SVGNode or VList expected")
		return 0
	}

	// Insert negative kern to compensate for background height
	k := node.NewKern()
	k.Kern = -bgHeight

	// Prepend: bgNode → kern(-H) → original list content
	vl.List = node.InsertBefore(vl.List, vl.List, k)
	vl.List = node.InsertBefore(vl.List, vl.List, bgNode)

	return 0
}

func registerElementInfoMetaTable(l *lua.State) {
	lua.NewMetaTable(l, elementInfoMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: elementInfoIndex},
	}, 0)
	l.Pop(1)
}
