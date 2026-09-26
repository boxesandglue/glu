package frontend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

// parseTabStops reads the tab_stops setting: a list whose entries are either
// a position (a left stop) or a table with position, align ("left",
// "right", "center", "decimal"), separator and leader.
func parseTabStops(l *lua.State, index int) []frontend.TabStop {
	index = l.AbsIndex(index)
	if !l.IsTable(index) {
		lua.Errorf(l, "tab_stops: table expected")
	}
	n := l.RawLength(index)
	stops := make([]frontend.TabStop, 0, n)
	for i := 1; i <= n; i++ {
		l.RawGetInt(index, i)
		stops = append(stops, parseTabStop(l, i))
		l.Pop(1)
	}
	return stops
}

// parseTabStop reads the tab stop on top of the stack, the i-th entry of
// tab_stops.
func parseTabStop(l *lua.State, i int) frontend.TabStop {
	var ts frontend.TabStop
	if !l.IsTable(-1) {
		if err := tabStopPosition(l, &ts); err != nil {
			lua.Errorf(l, "tab_stops[%d]: %s", i, err.Error())
		}
		return ts
	}
	l.Field(-1, "position")
	if err := tabStopPosition(l, &ts); err != nil {
		lua.Errorf(l, "tab_stops[%d].position: %s", i, err.Error())
	}
	l.Pop(1)

	l.Field(-1, "align")
	if !l.IsNil(-1) {
		s, _ := l.ToString(-1)
		switch s {
		case "left":
			ts.Align = node.TabAlignLeft
		case "right":
			ts.Align = node.TabAlignRight
		case "center":
			ts.Align = node.TabAlignCenter
		case "decimal":
			ts.Align = node.TabAlignDecimal
		default:
			lua.Errorf(l, "tab_stops[%d].align: unknown alignment '%s' (left, right, center, decimal)", i, s)
		}
	}
	l.Pop(1)

	l.Field(-1, "separator")
	if !l.IsNil(-1) {
		ts.Separator, _ = l.ToString(-1)
	}
	l.Pop(1)

	l.Field(-1, "leader")
	if !l.IsNil(-1) {
		ts.Leader, _ = l.ToString(-1)
	}
	l.Pop(1)
	return ts
}

// tabStopPosition reads the position on top of the stack: a dimension or a
// percentage of the line width such as "100%".
func tabStopPosition(l *lua.State, ts *frontend.TabStop) error {
	if s, ok := l.ToString(-1); ok && l.IsString(-1) && !l.IsNumber(-1) {
		if p, ok := strings.CutSuffix(s, "%"); ok {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return fmt.Errorf("invalid percentage %s", s)
			}
			ts.Fraction = f / 100
			return nil
		}
	}
	sp, err := toDimension(l, -1)
	if err != nil {
		return err
	}
	ts.Position = sp
	return nil
}

// pushTabStops pushes the stops as a list of tables, the long form of
// tab_stops.
func pushTabStops(l *lua.State, stops []frontend.TabStop) {
	l.CreateTable(len(stops), 0)
	for i, ts := range stops {
		l.CreateTable(0, 4)
		if ts.Fraction != 0 {
			l.PushString(strconv.FormatFloat(ts.Fraction*100, 'f', -1, 64) + "%")
		} else {
			pushScaledPoint(l, ts.Position)
		}
		l.SetField(-2, "position")
		l.PushString(tabAlignToString(ts.Align))
		l.SetField(-2, "align")
		if ts.Separator != "" {
			l.PushString(ts.Separator)
			l.SetField(-2, "separator")
		}
		if ts.Leader != "" {
			l.PushString(ts.Leader)
			l.SetField(-2, "leader")
		}
		l.RawSetInt(-2, i+1)
	}
}

func tabAlignToString(a node.TabAlign) string {
	switch a {
	case node.TabAlignRight:
		return "right"
	case node.TabAlignCenter:
		return "center"
	case node.TabAlignDecimal:
		return "decimal"
	default:
		return "left"
	}
}
