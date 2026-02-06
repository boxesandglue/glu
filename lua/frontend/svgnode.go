package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
)

const svgNodeMetaTable = "SVGNode"

// SVGNode wraps a node.Rule containing rendered SVG content.
// The Value field is discovered by extractNodeViaReflection, so SVGNode
// works automatically with node.vpack() and page:output_at().
type SVGNode struct {
	Value *node.Rule
}

// checkSVGNode retrieves an SVGNode userdata from the stack.
func checkSVGNode(l *lua.State, index int) *SVGNode {
	ud := lua.CheckUserData(l, index, svgNodeMetaTable)
	if n, ok := ud.(*SVGNode); ok {
		return n
	}
	lua.Errorf(l, "SVGNode expected")
	return nil
}

// svgNodeIndex handles attribute access (__index metamethod)
func svgNodeIndex(l *lua.State) int {
	n := checkSVGNode(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		pushScaledPoint(l, n.Value.Width)
		return 1
	case "height":
		pushScaledPoint(l, n.Value.Height)
		return 1
	}
	return 0
}

// svgNodeNewIndex handles attribute setting (__newindex metamethod)
func svgNodeNewIndex(l *lua.State) int {
	n := checkSVGNode(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		n.Value.Width = bag.ScaledPoint(checkDimension(l, 3))
	case "height":
		n.Value.Height = bag.ScaledPoint(checkDimension(l, 3))
	default:
		lua.Errorf(l, "cannot set attribute %s on SVGNode", key)
	}
	return 0
}

// registerSVGNodeMetaTable creates the SVGNode metatable.
func registerSVGNodeMetaTable(l *lua.State) {
	lua.NewMetaTable(l, svgNodeMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: svgNodeIndex},
		{Name: "__newindex", Function: svgNodeNewIndex},
	}, 0)
	l.Pop(1)
}
