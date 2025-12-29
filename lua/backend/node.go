package backend

import (
	"reflect"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
)

const (
	glyphMetaTable     = "node.Glyph"
	glueMetaTable      = "node.Glue"
	kernMetaTable      = "node.Kern"
	discMetaTable      = "node.Disc"
	penaltyMetaTable   = "node.Penalty"
	ruleMetaTable      = "node.Rule"
	hlistMetaTable     = "node.HList"
	vlistMetaTable     = "node.VList"
	imageMetaTable     = "node.Image"
	langMetaTable      = "node.Lang"
	startStopMetaTable = "node.StartStop"
)

// Node wrapper types
type NodeGlyph struct{ Value *node.Glyph }
type NodeGlue struct{ Value *node.Glue }
type NodeKern struct{ Value *node.Kern }
type NodeDisc struct{ Value *node.Disc }
type NodePenalty struct{ Value *node.Penalty }
type NodeRule struct{ Value *node.Rule }
type NodeHList struct{ Value *node.HList }
type NodeVList struct{ Value *node.VList }
type NodeImage struct{ Value *node.Image }
type NodeLang struct{ Value *node.Lang }
type NodeStartStop struct{ Value *node.StartStop }

// pushNode pushes any node to the Lua stack with the correct metatable
func pushNode(l *lua.State, n node.Node) {
	if n == nil {
		l.PushNil()
		return
	}
	switch v := n.(type) {
	case *node.Glyph:
		l.PushUserData(&NodeGlyph{Value: v})
		lua.SetMetaTableNamed(l, glyphMetaTable)
	case *node.Glue:
		l.PushUserData(&NodeGlue{Value: v})
		lua.SetMetaTableNamed(l, glueMetaTable)
	case *node.Kern:
		l.PushUserData(&NodeKern{Value: v})
		lua.SetMetaTableNamed(l, kernMetaTable)
	case *node.Disc:
		l.PushUserData(&NodeDisc{Value: v})
		lua.SetMetaTableNamed(l, discMetaTable)
	case *node.Penalty:
		l.PushUserData(&NodePenalty{Value: v})
		lua.SetMetaTableNamed(l, penaltyMetaTable)
	case *node.Rule:
		l.PushUserData(&NodeRule{Value: v})
		lua.SetMetaTableNamed(l, ruleMetaTable)
	case *node.HList:
		l.PushUserData(&NodeHList{Value: v})
		lua.SetMetaTableNamed(l, hlistMetaTable)
	case *node.VList:
		l.PushUserData(&NodeVList{Value: v})
		lua.SetMetaTableNamed(l, vlistMetaTable)
	case *node.Image:
		l.PushUserData(&NodeImage{Value: v})
		lua.SetMetaTableNamed(l, imageMetaTable)
	case *node.Lang:
		l.PushUserData(&NodeLang{Value: v})
		lua.SetMetaTableNamed(l, langMetaTable)
	case *node.StartStop:
		l.PushUserData(&NodeStartStop{Value: v})
		lua.SetMetaTableNamed(l, startStopMetaTable)
	default:
		l.PushNil()
	}
}

// getNode extracts a node.Node from userdata at the given index
func getNode(l *lua.State, index int) node.Node {
	if l.IsNil(index) {
		return nil
	}
	ud := l.ToUserData(index)
	switch v := ud.(type) {
	case *NodeGlyph:
		return v.Value
	case *NodeGlue:
		return v.Value
	case *NodeKern:
		return v.Value
	case *NodeDisc:
		return v.Value
	case *NodePenalty:
		return v.Value
	case *NodeRule:
		return v.Value
	case *NodeHList:
		return v.Value
	case *NodeVList:
		return v.Value
	case *NodeImage:
		return v.Value
	case *NodeLang:
		return v.Value
	case *NodeStartStop:
		return v.Value
	default:
		// Try to extract node.Node via reflection for cross-package types
		// This handles frontend.ImageNode which has a Value *node.Image field
		if n := extractNodeViaReflection(ud); n != nil {
			return n
		}
	}
	return nil
}

// extractNodeViaReflection tries to extract a node.Node from a struct with a Value field
func extractNodeViaReflection(ud any) node.Node {
	v := reflect.ValueOf(ud)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	valueField := v.FieldByName("Value")
	if !valueField.IsValid() || valueField.IsNil() {
		return nil
	}
	if n, ok := valueField.Interface().(node.Node); ok {
		return n
	}
	return nil
}

// nodeNew creates a new node: node.new("glyph")
func nodeNew(l *lua.State) int {
	nodeType := lua.CheckString(l, 1)
	switch nodeType {
	case "glyph":
		l.PushUserData(&NodeGlyph{Value: node.NewGlyph()})
		lua.SetMetaTableNamed(l, glyphMetaTable)
	case "glue":
		l.PushUserData(&NodeGlue{Value: node.NewGlue()})
		lua.SetMetaTableNamed(l, glueMetaTable)
	case "kern":
		l.PushUserData(&NodeKern{Value: node.NewKern()})
		lua.SetMetaTableNamed(l, kernMetaTable)
	case "disc":
		l.PushUserData(&NodeDisc{Value: node.NewDisc()})
		lua.SetMetaTableNamed(l, discMetaTable)
	case "penalty":
		l.PushUserData(&NodePenalty{Value: node.NewPenalty()})
		lua.SetMetaTableNamed(l, penaltyMetaTable)
	case "rule":
		l.PushUserData(&NodeRule{Value: node.NewRule()})
		lua.SetMetaTableNamed(l, ruleMetaTable)
	case "hlist":
		l.PushUserData(&NodeHList{Value: node.NewHList()})
		lua.SetMetaTableNamed(l, hlistMetaTable)
	case "vlist":
		l.PushUserData(&NodeVList{Value: node.NewVList()})
		lua.SetMetaTableNamed(l, vlistMetaTable)
	case "image":
		l.PushUserData(&NodeImage{Value: node.NewImage()})
		lua.SetMetaTableNamed(l, imageMetaTable)
	case "lang":
		l.PushUserData(&NodeLang{Value: node.NewLang()})
		lua.SetMetaTableNamed(l, langMetaTable)
	case "startstop":
		l.PushUserData(&NodeStartStop{Value: node.NewStartStop()})
		lua.SetMetaTableNamed(l, startStopMetaTable)
	default:
		lua.Errorf(l, "unknown node type: %s", nodeType)
		return 0
	}
	return 1
}

// nodeInsertAfter inserts a node after another: node.insert_after(head, cur, new)
func nodeInsertAfter(l *lua.State) int {
	head := getNode(l, 1)
	cur := getNode(l, 2)
	insert := getNode(l, 3)
	result := node.InsertAfter(head, cur, insert)
	pushNode(l, result)
	return 1
}

// nodeInsertBefore inserts a node before another: node.insert_before(head, cur, new)
func nodeInsertBefore(l *lua.State) int {
	head := getNode(l, 1)
	cur := getNode(l, 2)
	insert := getNode(l, 3)
	result := node.InsertBefore(head, cur, insert)
	pushNode(l, result)
	return 1
}

// nodeDelete removes a node from a list: node.delete(head, cur)
func nodeDelete(l *lua.State) int {
	head := getNode(l, 1)
	cur := getNode(l, 2)
	result := node.DeleteFromList(head, cur)
	pushNode(l, result)
	return 1
}

// nodeCopyList creates a deep copy: node.copy_list(head)
func nodeCopyList(l *lua.State) int {
	head := getNode(l, 1)
	result := node.CopyList(head)
	pushNode(l, result)
	return 1
}

// nodeTail returns the last node: node.tail(head)
func nodeTail(l *lua.State) int {
	head := getNode(l, 1)
	result := node.Tail(head)
	pushNode(l, result)
	return 1
}

// nodeHpack packs nodes into an hlist: node.hpack(head)
func nodeHpack(l *lua.State) int {
	head := getNode(l, 1)
	result := node.Hpack(head)
	l.PushUserData(&NodeHList{Value: result})
	lua.SetMetaTableNamed(l, hlistMetaTable)
	return 1
}

// nodeHpackTo packs nodes to a specific width: node.hpack_to(head, width)
func nodeHpackTo(l *lua.State) int {
	head := getNode(l, 1)
	width := bag.ScaledPoint(lua.CheckInteger(l, 2))
	result := node.HpackTo(head, width)
	l.PushUserData(&NodeHList{Value: result})
	lua.SetMetaTableNamed(l, hlistMetaTable)
	return 1
}

// nodeVpack packs nodes into a vlist: node.vpack(head)
func nodeVpack(l *lua.State) int {
	head := getNode(l, 1)
	result := node.Vpack(head)
	l.PushUserData(&NodeVList{Value: result})
	lua.SetMetaTableNamed(l, vlistMetaTable)
	return 1
}

// nodeDimensions returns width, height, depth: node.dimensions(head)
func nodeDimensions(l *lua.State) int {
	head := getNode(l, 1)
	width, height, depth := node.Dimensions(head, nil, node.Horizontal)
	l.PushInteger(int(width))
	l.PushInteger(int(height))
	l.PushInteger(int(depth))
	return 3
}

// nodeString returns a string representation: node.string(head)
func nodeString(l *lua.State) int {
	head := getNode(l, 1)
	l.PushString(node.String(head))
	return 1
}

// Generic index function for all nodes - handles next/prev
func nodeGenericIndex(l *lua.State, n node.Node) int {
	key := lua.CheckString(l, 2)
	switch key {
	case "next":
		pushNode(l, n.Next())
		return 1
	case "prev":
		pushNode(l, n.Prev())
		return 1
	case "type":
		l.PushString(n.Type().String())
		return 1
	case "id":
		l.PushInteger(n.GetID())
		return 1
	}
	return 0
}

// Generic newindex function for all nodes - handles next/prev
func nodeGenericNewIndex(l *lua.State, n node.Node) int {
	key := lua.CheckString(l, 2)
	switch key {
	case "next":
		n.SetNext(getNode(l, 3))
		return 0
	case "prev":
		n.SetPrev(getNode(l, 3))
		return 0
	}
	return 0
}

// Glyph node index
func glyphIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, glyphMetaTable)
	g := ud.(*NodeGlyph).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "codepoint":
		l.PushInteger(g.Codepoint)
		return 1
	case "components":
		l.PushString(g.Components)
		return 1
	case "width":
		l.PushInteger(int(g.Width))
		return 1
	case "height":
		l.PushInteger(int(g.Height))
		return 1
	case "depth":
		l.PushInteger(int(g.Depth))
		return 1
	case "yoffset":
		l.PushInteger(int(g.YOffset))
		return 1
	case "hyphenate":
		l.PushBoolean(g.Hyphenate)
		return 1
	}
	return nodeGenericIndex(l, g)
}

func glyphNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, glyphMetaTable)
	g := ud.(*NodeGlyph).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "codepoint":
		g.Codepoint = lua.CheckInteger(l, 3)
	case "components":
		g.Components = lua.CheckString(l, 3)
	case "width":
		g.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "height":
		g.Height = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "depth":
		g.Depth = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "yoffset":
		g.YOffset = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "hyphenate":
		g.Hyphenate = l.ToBoolean(3)
	default:
		nodeGenericNewIndex(l, g)
	}
	return 0
}

// Glue node index
func glueIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, glueMetaTable)
	g := ud.(*NodeGlue).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushInteger(int(g.Width))
		return 1
	case "stretch":
		l.PushInteger(int(g.Stretch))
		return 1
	case "shrink":
		l.PushInteger(int(g.Shrink))
		return 1
	case "stretch_order":
		l.PushInteger(int(g.StretchOrder))
		return 1
	case "shrink_order":
		l.PushInteger(int(g.ShrinkOrder))
		return 1
	}
	return nodeGenericIndex(l, g)
}

func glueNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, glueMetaTable)
	g := ud.(*NodeGlue).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		g.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "stretch":
		g.Stretch = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "shrink":
		g.Shrink = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "stretch_order":
		g.StretchOrder = node.GlueOrder(lua.CheckInteger(l, 3))
	case "shrink_order":
		g.ShrinkOrder = node.GlueOrder(lua.CheckInteger(l, 3))
	default:
		nodeGenericNewIndex(l, g)
	}
	return 0
}

// Kern node index
func kernIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, kernMetaTable)
	k := ud.(*NodeKern).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "kern":
		l.PushInteger(int(k.Kern))
		return 1
	}
	return nodeGenericIndex(l, k)
}

func kernNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, kernMetaTable)
	k := ud.(*NodeKern).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "kern":
		k.Kern = bag.ScaledPoint(lua.CheckInteger(l, 3))
	default:
		nodeGenericNewIndex(l, k)
	}
	return 0
}

// Penalty node index
func penaltyIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, penaltyMetaTable)
	p := ud.(*NodePenalty).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "penalty":
		l.PushInteger(p.Penalty)
		return 1
	case "width":
		l.PushInteger(int(p.Width))
		return 1
	}
	return nodeGenericIndex(l, p)
}

func penaltyNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, penaltyMetaTable)
	p := ud.(*NodePenalty).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "penalty":
		p.Penalty = lua.CheckInteger(l, 3)
	case "width":
		p.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	default:
		nodeGenericNewIndex(l, p)
	}
	return 0
}

// Rule node index
func ruleIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, ruleMetaTable)
	r := ud.(*NodeRule).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushInteger(int(r.Width))
		return 1
	case "height":
		l.PushInteger(int(r.Height))
		return 1
	case "depth":
		l.PushInteger(int(r.Depth))
		return 1
	case "pre":
		l.PushString(r.Pre)
		return 1
	case "post":
		l.PushString(r.Post)
		return 1
	case "hide":
		l.PushBoolean(r.Hide)
		return 1
	}
	return nodeGenericIndex(l, r)
}

func ruleNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, ruleMetaTable)
	r := ud.(*NodeRule).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		r.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "height":
		r.Height = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "depth":
		r.Depth = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "pre":
		r.Pre = lua.CheckString(l, 3)
	case "post":
		r.Post = lua.CheckString(l, 3)
	case "hide":
		r.Hide = l.ToBoolean(3)
	default:
		nodeGenericNewIndex(l, r)
	}
	return 0
}

// HList node index
func hlistIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, hlistMetaTable)
	h := ud.(*NodeHList).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushInteger(int(h.Width))
		return 1
	case "height":
		l.PushInteger(int(h.Height))
		return 1
	case "depth":
		l.PushInteger(int(h.Depth))
		return 1
	case "list":
		pushNode(l, h.List)
		return 1
	case "glue_set":
		l.PushNumber(h.GlueSet)
		return 1
	case "glue_sign":
		l.PushInteger(int(h.GlueSign))
		return 1
	case "glue_order":
		l.PushInteger(int(h.GlueOrder))
		return 1
	case "shift":
		l.PushInteger(int(h.Shift))
		return 1
	case "badness":
		l.PushInteger(h.Badness)
		return 1
	}
	return nodeGenericIndex(l, h)
}

func hlistNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, hlistMetaTable)
	h := ud.(*NodeHList).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		h.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "height":
		h.Height = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "depth":
		h.Depth = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "list":
		h.List = getNode(l, 3)
	case "glue_set":
		h.GlueSet = lua.CheckNumber(l, 3)
	case "glue_sign":
		h.GlueSign = uint8(lua.CheckInteger(l, 3))
	case "glue_order":
		h.GlueOrder = node.GlueOrder(lua.CheckInteger(l, 3))
	case "shift":
		h.Shift = bag.ScaledPoint(lua.CheckInteger(l, 3))
	default:
		nodeGenericNewIndex(l, h)
	}
	return 0
}

// VList node index
func vlistIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, vlistMetaTable)
	v := ud.(*NodeVList).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushInteger(int(v.Width))
		return 1
	case "height":
		l.PushInteger(int(v.Height))
		return 1
	case "depth":
		l.PushInteger(int(v.Depth))
		return 1
	case "list":
		pushNode(l, v.List)
		return 1
	case "glue_set":
		l.PushNumber(v.GlueSet)
		return 1
	case "glue_sign":
		l.PushInteger(int(v.GlueSign))
		return 1
	case "shift_x":
		l.PushInteger(int(v.ShiftX))
		return 1
	}
	return nodeGenericIndex(l, v)
}

func vlistNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, vlistMetaTable)
	v := ud.(*NodeVList).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		v.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "height":
		v.Height = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "depth":
		v.Depth = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "list":
		v.List = getNode(l, 3)
	case "glue_set":
		v.GlueSet = lua.CheckNumber(l, 3)
	case "glue_sign":
		v.GlueSign = uint8(lua.CheckInteger(l, 3))
	case "shift_x":
		v.ShiftX = bag.ScaledPoint(lua.CheckInteger(l, 3))
	default:
		nodeGenericNewIndex(l, v)
	}
	return 0
}

// Disc node index
func discIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, discMetaTable)
	d := ud.(*NodeDisc).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "pre":
		pushNode(l, d.Pre)
		return 1
	case "post":
		pushNode(l, d.Post)
		return 1
	case "replace":
		pushNode(l, d.Replace)
		return 1
	case "penalty":
		l.PushInteger(d.Penalty)
		return 1
	}
	return nodeGenericIndex(l, d)
}

func discNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, discMetaTable)
	d := ud.(*NodeDisc).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "pre":
		d.Pre = getNode(l, 3)
	case "post":
		d.Post = getNode(l, 3)
	case "replace":
		d.Replace = getNode(l, 3)
	case "penalty":
		d.Penalty = lua.CheckInteger(l, 3)
	default:
		nodeGenericNewIndex(l, d)
	}
	return 0
}

// Image node index
func imageIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, imageMetaTable)
	img := ud.(*NodeImage).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		l.PushInteger(int(img.Width))
		return 1
	case "height":
		l.PushInteger(int(img.Height))
		return 1
	case "page":
		l.PushInteger(img.PageNumber)
		return 1
	case "used":
		l.PushBoolean(img.Used)
		return 1
	}
	return nodeGenericIndex(l, img)
}

func imageNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, imageMetaTable)
	img := ud.(*NodeImage).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "width":
		img.Width = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "height":
		img.Height = bag.ScaledPoint(lua.CheckInteger(l, 3))
	case "page":
		img.PageNumber = lua.CheckInteger(l, 3)
	case "used":
		img.Used = l.ToBoolean(3)
	default:
		nodeGenericNewIndex(l, img)
	}
	return 0
}

// Lang node index
func langIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, langMetaTable)
	ln := ud.(*NodeLang).Value
	return nodeGenericIndex(l, ln)
}

func langNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, langMetaTable)
	ln := ud.(*NodeLang).Value
	nodeGenericNewIndex(l, ln)
	return 0
}

// StartStop node index
func startStopIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, startStopMetaTable)
	ss := ud.(*NodeStartStop).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "action":
		l.PushInteger(int(ss.Action))
		return 1
	}
	return nodeGenericIndex(l, ss)
}

func startStopNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, startStopMetaTable)
	ss := ud.(*NodeStartStop).Value
	key := lua.CheckString(l, 2)

	switch key {
	case "action":
		ss.Action = node.ActionType(lua.CheckInteger(l, 3))
	default:
		nodeGenericNewIndex(l, ss)
	}
	return 0
}

// registerNodeMetaTables creates all node metatables
func registerNodeMetaTables(l *lua.State) {
	// Glyph
	lua.NewMetaTable(l, glyphMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: glyphIndex},
		{Name: "__newindex", Function: glyphNewIndex},
	}, 0)
	l.Pop(1)

	// Glue
	lua.NewMetaTable(l, glueMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: glueIndex},
		{Name: "__newindex", Function: glueNewIndex},
	}, 0)
	l.Pop(1)

	// Kern
	lua.NewMetaTable(l, kernMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: kernIndex},
		{Name: "__newindex", Function: kernNewIndex},
	}, 0)
	l.Pop(1)

	// Disc
	lua.NewMetaTable(l, discMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: discIndex},
		{Name: "__newindex", Function: discNewIndex},
	}, 0)
	l.Pop(1)

	// Penalty
	lua.NewMetaTable(l, penaltyMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: penaltyIndex},
		{Name: "__newindex", Function: penaltyNewIndex},
	}, 0)
	l.Pop(1)

	// Rule
	lua.NewMetaTable(l, ruleMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: ruleIndex},
		{Name: "__newindex", Function: ruleNewIndex},
	}, 0)
	l.Pop(1)

	// HList
	lua.NewMetaTable(l, hlistMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: hlistIndex},
		{Name: "__newindex", Function: hlistNewIndex},
	}, 0)
	l.Pop(1)

	// VList
	lua.NewMetaTable(l, vlistMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: vlistIndex},
		{Name: "__newindex", Function: vlistNewIndex},
	}, 0)
	l.Pop(1)

	// Image
	lua.NewMetaTable(l, imageMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: imageIndex},
		{Name: "__newindex", Function: imageNewIndex},
	}, 0)
	l.Pop(1)

	// Lang
	lua.NewMetaTable(l, langMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: langIndex},
		{Name: "__newindex", Function: langNewIndex},
	}, 0)
	l.Pop(1)

	// StartStop
	lua.NewMetaTable(l, startStopMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: startStopIndex},
		{Name: "__newindex", Function: startStopNewIndex},
	}, 0)
	l.Pop(1)
}

// openNode creates the node module table for require("glu.node")
func openNode(l *lua.State) int {
	registerNodeMetaTables(l)

	l.NewTable()
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "new", Function: nodeNew},
		{Name: "insert_after", Function: nodeInsertAfter},
		{Name: "insert_before", Function: nodeInsertBefore},
		{Name: "delete", Function: nodeDelete},
		{Name: "copy_list", Function: nodeCopyList},
		{Name: "tail", Function: nodeTail},
		{Name: "hpack", Function: nodeHpack},
		{Name: "hpack_to", Function: nodeHpackTo},
		{Name: "vpack", Function: nodeVpack},
		{Name: "dimensions", Function: nodeDimensions},
		{Name: "string", Function: nodeString},
	}, 0)

	// Add glue order constants
	l.PushInteger(int(node.StretchNormal))
	l.SetField(-2, "normal")
	l.PushInteger(int(node.StretchFil))
	l.SetField(-2, "fil")
	l.PushInteger(int(node.StretchFill))
	l.SetField(-2, "fill")
	l.PushInteger(int(node.StretchFilll))
	l.SetField(-2, "filll")

	return 1
}
