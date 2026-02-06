package frontend

import (
	"os"
	"strings"

	"github.com/boxesandglue/svgreader"
	"github.com/speedata/go-lua"
)

const svgDocumentMetaTable = "SVGDocument"

// SVGDocument wraps a parsed svgreader.Document for Lua.
type SVGDocument struct {
	Value *svgreader.Document
}

// checkSVGDocument retrieves an SVGDocument userdata from the stack.
func checkSVGDocument(l *lua.State, index int) *SVGDocument {
	ud := lua.CheckUserData(l, index, svgDocumentMetaTable)
	if d, ok := ud.(*SVGDocument); ok {
		return d
	}
	lua.Errorf(l, "SVGDocument expected")
	return nil
}

// parseSVGFromFile parses an SVG file: frontend.parse_svg(filename)
func parseSVGFromFile(l *lua.State) int {
	filename := lua.CheckString(l, 1)

	f, err := os.Open(filename)
	if err != nil {
		lua.Errorf(l, "failed to open SVG file: %s", err.Error())
		return 0
	}
	defer f.Close()

	doc, err := svgreader.Parse(f)
	if err != nil {
		lua.Errorf(l, "failed to parse SVG: %s", err.Error())
		return 0
	}

	l.PushUserData(&SVGDocument{Value: doc})
	lua.SetMetaTableNamed(l, svgDocumentMetaTable)
	return 1
}

// parseSVGFromString parses SVG from a string: frontend.parse_svg_string(str)
func parseSVGFromString(l *lua.State) int {
	str := lua.CheckString(l, 1)

	doc, err := svgreader.Parse(strings.NewReader(str))
	if err != nil {
		lua.Errorf(l, "failed to parse SVG string: %s", err.Error())
		return 0
	}

	l.PushUserData(&SVGDocument{Value: doc})
	lua.SetMetaTableNamed(l, svgDocumentMetaTable)
	return 1
}

// svgDocumentIndex handles attribute access (__index metamethod)
func svgDocumentIndex(l *lua.State) int {
	d := checkSVGDocument(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushNumber(d.Value.Width)
		return 1
	case "height":
		l.PushNumber(d.Value.Height)
		return 1
	}
	return 0
}

// registerSVGDocumentMetaTable creates the SVGDocument metatable.
func registerSVGDocumentMetaTable(l *lua.State) {
	lua.NewMetaTable(l, svgDocumentMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: svgDocumentIndex},
	}, 0)
	l.Pop(1)
}
