package backend

import (
	"fmt"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/boxesandglue/glu/lua/common"
	"github.com/speedata/go-lua"
)

// toSP reads a length the way the node module takes one: a number in
// scaled points, a string with a unit or a ScaledPoint.
func toSP(l *lua.State, index int) (bag.ScaledPoint, error) {
	if sp := common.TestScaledPoint(l, index); sp != nil {
		return sp.Value, nil
	}
	if l.IsNumber(index) {
		n, _ := l.ToNumber(index)
		return bag.ScaledPoint(n), nil
	}
	if l.IsString(index) {
		s, _ := l.ToString(index)
		return bag.SP(s)
	}
	return 0, fmt.Errorf("expected scaled points, a string with a unit or a ScaledPoint")
}

// nodeLinebreak breaks a node list into lines:
// node.linebreak(head, settings) returns the VList of lines and a list
// with one table per line: its adjustment ratio, demerits, fitness class
// and whether no feasible break could end it.
func nodeLinebreak(l *lua.State) int {
	head := getNode(l, 1)
	lua.CheckType(l, 2, lua.TypeTable)
	ls := node.NewLinebreakSettings()
	l.PushNil()
	for l.Next(2) {
		key, ok := l.ToString(-2)
		if !ok {
			lua.Errorf(l, "linebreak: settings keys must be strings")
		}
		setLinebreakSetting(l, ls, key)
		l.Pop(1)
	}
	if ls.HSize <= 0 {
		lua.Errorf(l, "linebreak: hsize must be greater than 0")
	}
	// Without a forced break at its end, the line breaker loses the last
	// node of the list.
	if p, ok := node.Tail(head).(*node.Penalty); !ok || p.Penalty > -10000 {
		lua.Errorf(l, "linebreak: the list must end with a forced break, see node.append_lineend")
	}
	vl, breakpoints := node.Linebreak(head, ls)
	pushNode(l, vl)
	l.CreateTable(len(breakpoints), 0)
	for i, bp := range breakpoints {
		l.CreateTable(0, 4)
		l.PushNumber(bp.R)
		l.SetField(-2, "ratio")
		l.PushInteger(bp.Demerits)
		l.SetField(-2, "demerits")
		l.PushInteger(bp.Fitness)
		l.SetField(-2, "fitness")
		l.PushBoolean(bp.Emergency)
		l.SetField(-2, "emergency")
		l.RawSetInt(-2, i+1)
	}
	return 2
}

// setLinebreakSetting sets the field key of ls from the value on top of
// the stack.
func setLinebreakSetting(l *lua.State, ls *node.LinebreakSettings, key string) {
	length := func() bag.ScaledPoint {
		sp, err := toSP(l, -1)
		if err != nil {
			lua.Errorf(l, "linebreak: %s: %s", key, err.Error())
		}
		return sp
	}
	integer := func() int {
		n, ok := l.ToInteger(-1)
		if !ok {
			lua.Errorf(l, "linebreak: %s: integer expected", key)
		}
		return n
	}
	number := func() float64 {
		f, ok := l.ToNumber(-1)
		if !ok {
			lua.Errorf(l, "linebreak: %s: number expected", key)
		}
		return f
	}
	glue := func() *node.Glue {
		g, ok := getNode(l, -1).(*node.Glue)
		if !ok {
			lua.Errorf(l, "linebreak: %s: glue node expected", key)
		}
		return g
	}
	switch key {
	case "hsize":
		ls.HSize = length()
	case "line_height":
		ls.LineHeight = length()
	case "tolerance":
		if ls.Tolerance = number(); ls.Tolerance < 0 {
			lua.Errorf(l, "linebreak: tolerance must not be negative")
		}
	case "emergency_stretch":
		ls.EmergencyStretch = length()
	case "hyphen_penalty":
		ls.Hyphenpenalty = integer()
	case "line_penalty":
		ls.LinePenalty = integer()
	case "double_hyphen_demerits":
		ls.DoublehyphenDemerits = integer()
	case "final_hyphen_demerits":
		ls.FinalHyphenDemerits = integer()
	case "demerits_fitness":
		ls.DemeritsFitness = integer()
	case "font_expansion":
		ls.FontExpansion = number()
	case "indent":
		ls.Indent = length()
	case "indent_rows":
		ls.IndentRows = integer()
	case "indent_right":
		ls.IndentRight = length()
	case "indent_right_rows":
		ls.IndentRightRows = integer()
	case "indent_steps":
		ls.IndentSteps = common.ParseIndentSteps(l, -1, key, toSP)
	case "indent_right_steps":
		ls.IndentRightSteps = common.ParseIndentSteps(l, -1, key, toSP)
	case "line_start_glue":
		ls.LineStartGlue = glue()
	case "line_end_glue":
		ls.LineEndGlue = glue()
	case "squeeze_overfull_boxes":
		ls.SqueezeOverfullBoxes = l.ToBoolean(-1)
	case "hanging_punctuation_end":
		ls.HangingPunctuationEnd = l.ToBoolean(-1)
	case "omit_last_leading":
		ls.OmitLastLeading = l.ToBoolean(-1)
	case "half_leading":
		ls.HalfLeading = l.ToBoolean(-1)
	case "font":
		ls.Font = checkFont(l, -1).Value
	case "direction":
		switch s, _ := l.ToString(-1); s {
		case "ltr":
			ls.TextDirection = node.TextDirLTR
		case "rtl":
			ls.TextDirection = node.TextDirRTL
		default:
			lua.Errorf(l, "linebreak: direction must be ltr or rtl, got %s", s)
		}
	default:
		lua.Errorf(l, "linebreak: unknown setting %s", key)
	}
}

// nodeAppendLineend ends a paragraph for node.linebreak:
// node.append_lineend(head, tail) appends a penalty that forbids a break,
// a glue of fil stretch that fills the last line and a forced break after
// tail, and returns the head and the new tail.
func nodeAppendLineend(l *lua.State) int {
	head, tail := node.AppendLineEndAfter(getNode(l, 1), getNode(l, 2))
	pushNode(l, head)
	pushNode(l, tail)
	return 2
}
