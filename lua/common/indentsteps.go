package common

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
)

// ParseIndentSteps reads the table at index as a list of indent steps,
// each a table { inset = <dimension>, rows = <integer> } or the pair
// { <dimension>, <integer> }. dim reads a dimension the way the calling
// module does. name is the key the table was given under, for the error
// messages.
func ParseIndentSteps(l *lua.State, index int, name string, dim func(*lua.State, int) (bag.ScaledPoint, error)) []node.IndentStep {
	index = l.AbsIndex(index)
	if !l.IsTable(index) {
		lua.Errorf(l, "%s: table expected", name)
	}
	n := l.RawLength(index)
	steps := make([]node.IndentStep, 0, n)
	for i := 1; i <= n; i++ {
		l.RawGetInt(index, i)
		if !l.IsTable(-1) {
			lua.Errorf(l, "%s[%d]: table expected", name, i)
		}
		var step node.IndentStep
		l.Field(-1, "inset")
		if l.IsNil(-1) {
			l.Pop(1)
			l.RawGetInt(-1, 1)
		}
		inset, err := dim(l, -1)
		if err != nil {
			lua.Errorf(l, "%s[%d].inset: %s", name, i, err.Error())
		}
		step.Inset = inset
		l.Pop(1)
		l.Field(-1, "rows")
		if l.IsNil(-1) {
			l.Pop(1)
			l.RawGetInt(-1, 2)
		}
		if !l.IsNil(-1) {
			rows, ok := l.ToInteger(-1)
			if !ok {
				lua.Errorf(l, "%s[%d].rows: integer expected", name, i)
			}
			step.Rows = rows
		}
		l.Pop(1)
		steps = append(steps, step)
		l.Pop(1)
	}
	return steps
}

// PushIndentSteps pushes steps as a list of { inset = ..., rows = ... }
// tables, the inset pushed by pushDim.
func PushIndentSteps(l *lua.State, steps []node.IndentStep, pushDim func(*lua.State, bag.ScaledPoint)) {
	l.CreateTable(len(steps), 0)
	for i, s := range steps {
		l.CreateTable(0, 2)
		pushDim(l, s.Inset)
		l.SetField(-2, "inset")
		l.PushInteger(s.Rows)
		l.SetField(-2, "rows")
		l.RawSetInt(-2, i+1)
	}
}
