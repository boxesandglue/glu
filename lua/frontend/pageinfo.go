package frontend

import (
	"github.com/boxesandglue/htmlbag"
	"github.com/speedata/go-lua"
)

const pageInfoMetaTable = "PageInfo"

// PageInfo wraps an htmlbag.CSSBuilder to expose page layout and CSS
// information to Lua callbacks.
type PageInfo struct {
	cssBuilder *htmlbag.CSSBuilder
	dimensions htmlbag.PageDimensions
}

func checkPageInfo(l *lua.State, index int) *PageInfo {
	ud := lua.CheckUserData(l, index, pageInfoMetaTable)
	if pi, ok := ud.(*PageInfo); ok {
		return pi
	}
	lua.Errorf(l, "PageInfo expected")
	return nil
}

func pageInfoIndex(l *lua.State) int {
	pi := checkPageInfo(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "margin_top":
		pushScaledPoint(l, pi.dimensions.MarginTop)
		return 1
	case "margin_bottom":
		pushScaledPoint(l, pi.dimensions.MarginBottom)
		return 1
	case "margin_left":
		pushScaledPoint(l, pi.dimensions.MarginLeft)
		return 1
	case "margin_right":
		pushScaledPoint(l, pi.dimensions.MarginRight)
		return 1
	case "content_width":
		pushScaledPoint(l, pi.dimensions.ContentWidth)
		return 1
	case "content_height":
		pushScaledPoint(l, pi.dimensions.ContentHeight)
		return 1
	case "page_area_left":
		pushScaledPoint(l, pi.dimensions.PageAreaLeft)
		return 1
	case "page_area_top":
		pushScaledPoint(l, pi.dimensions.PageAreaTop)
		return 1
	case "page_areas":
		areas := pi.dimensions.PageAreas()
		if areas == nil {
			l.PushNil()
			return 1
		}
		// Push as nested Lua table: { ["@top-center"] = { content = "...", ... }, ... }
		l.NewTable()
		for areaName, props := range areas {
			l.NewTable()
			for k, v := range props {
				l.PushString(v)
				l.SetField(-2, k)
			}
			l.SetField(-2, areaName)
		}
		return 1
	}

	return 0
}

func registerPageInfoMetaTable(l *lua.State) {
	lua.NewMetaTable(l, pageInfoMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: pageInfoIndex},
	}, 0)
	l.Pop(1)
}
