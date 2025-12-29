package common

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/speedata/go-lua"
)

const ScaledPointMetaTable = "ScaledPoint"

// ScaledPoint wraps bag.ScaledPoint as Lua userdata
type ScaledPoint struct {
	Value bag.ScaledPoint
}

// CheckScaledPoint retrieves a ScaledPoint userdata from the stack
func CheckScaledPoint(l *lua.State, index int) *ScaledPoint {
	ud := lua.CheckUserData(l, index, ScaledPointMetaTable)
	if sp, ok := ud.(*ScaledPoint); ok {
		return sp
	}
	lua.Errorf(l, "ScaledPoint expected")
	return nil
}

// TestScaledPoint checks if the value at index is a ScaledPoint userdata
func TestScaledPoint(l *lua.State, index int) *ScaledPoint {
	ud := lua.TestUserData(l, index, ScaledPointMetaTable)
	if sp, ok := ud.(*ScaledPoint); ok {
		return sp
	}
	return nil
}

// PushScaledPoint creates a new ScaledPoint userdata and pushes it onto the stack
func PushScaledPoint(l *lua.State, sp bag.ScaledPoint) {
	ud := &ScaledPoint{Value: sp}
	l.PushUserData(ud)
	lua.SetMetaTableNamed(l, ScaledPointMetaTable)
}

// ToScaledPointValue converts a Lua value (ScaledPoint, string, or number) to bag.ScaledPoint
// Returns the value and true on success, or 0 and false on failure
func ToScaledPointValue(l *lua.State, index int) (bag.ScaledPoint, bool) {
	if sp := TestScaledPoint(l, index); sp != nil {
		return sp.Value, true
	}
	// Check number BEFORE string, because Lua can coerce numbers to strings
	if l.IsNumber(index) {
		n, _ := l.ToNumber(index)
		return bag.ScaledPointFromFloat(n), true
	}
	if l.IsString(index) {
		s, _ := l.ToString(index)
		sp, err := bag.SP(s)
		if err != nil {
			return 0, false
		}
		return sp, true
	}
	return 0, false
}

// scaledPointAdd implements __add: sp + sp, sp + "1cm", sp + number
func scaledPointAdd(l *lua.State) int {
	sp1, ok1 := ToScaledPointValue(l, 1)
	sp2, ok2 := ToScaledPointValue(l, 2)
	if !ok1 || !ok2 {
		lua.Errorf(l, "invalid operands for + (expected ScaledPoint, string with unit, or number)")
		return 0
	}
	PushScaledPoint(l, sp1+sp2)
	return 1
}

// scaledPointSub implements __sub: sp - sp, sp - "1cm", sp - number
func scaledPointSub(l *lua.State) int {
	sp1, ok1 := ToScaledPointValue(l, 1)
	sp2, ok2 := ToScaledPointValue(l, 2)
	if !ok1 || !ok2 {
		lua.Errorf(l, "invalid operands for - (expected ScaledPoint, string with unit, or number)")
		return 0
	}
	PushScaledPoint(l, sp1-sp2)
	return 1
}

// scaledPointMul implements __mul: sp * number or number * sp
func scaledPointMul(l *lua.State) int {
	// Check which operand is the ScaledPoint
	if sp := TestScaledPoint(l, 1); sp != nil {
		n := lua.CheckNumber(l, 2)
		PushScaledPoint(l, bag.ScaledPoint(float64(sp.Value)*n))
		return 1
	}
	if sp := TestScaledPoint(l, 2); sp != nil {
		n := lua.CheckNumber(l, 1)
		PushScaledPoint(l, bag.ScaledPoint(float64(sp.Value)*n))
		return 1
	}
	lua.Errorf(l, "ScaledPoint expected for multiplication")
	return 0
}

// scaledPointDiv implements __div: sp / number or sp / sp (returns ratio)
func scaledPointDiv(l *lua.State) int {
	sp1 := CheckScaledPoint(l, 1)
	// Check if dividing by another ScaledPoint (returns ratio as number)
	if sp2 := TestScaledPoint(l, 2); sp2 != nil {
		if sp2.Value == 0 {
			lua.Errorf(l, "division by zero")
			return 0
		}
		l.PushNumber(float64(sp1.Value) / float64(sp2.Value))
		return 1
	}
	// Dividing by a number
	n := lua.CheckNumber(l, 2)
	if n == 0 {
		lua.Errorf(l, "division by zero")
		return 0
	}
	PushScaledPoint(l, bag.ScaledPoint(float64(sp1.Value)/n))
	return 1
}

// scaledPointUnm implements __unm: -sp
func scaledPointUnm(l *lua.State) int {
	sp := CheckScaledPoint(l, 1)
	PushScaledPoint(l, -sp.Value)
	return 1
}

// scaledPointEq implements __eq: sp1 == sp2
func scaledPointEq(l *lua.State) int {
	sp1 := CheckScaledPoint(l, 1)
	sp2 := CheckScaledPoint(l, 2)
	l.PushBoolean(sp1.Value == sp2.Value)
	return 1
}

// scaledPointLt implements __lt: sp1 < sp2
func scaledPointLt(l *lua.State) int {
	sp1 := CheckScaledPoint(l, 1)
	sp2 := CheckScaledPoint(l, 2)
	l.PushBoolean(sp1.Value < sp2.Value)
	return 1
}

// scaledPointLe implements __le: sp1 <= sp2
func scaledPointLe(l *lua.State) int {
	sp1 := CheckScaledPoint(l, 1)
	sp2 := CheckScaledPoint(l, 2)
	l.PushBoolean(sp1.Value <= sp2.Value)
	return 1
}

// scaledPointToString implements __tostring
func scaledPointToString(l *lua.State) int {
	sp := CheckScaledPoint(l, 1)
	l.PushString(sp.Value.String() + "pt")
	return 1
}

// scaledPointIndex handles attribute access (__index metamethod)
func scaledPointIndex(l *lua.State) int {
	sp := CheckScaledPoint(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "pt":
		// Return value in points as number
		l.PushNumber(sp.Value.ToPT())
		return 1
	case "sp":
		// Return raw scaled point value as number
		l.PushNumber(float64(sp.Value))
		return 1
	case "to_pt":
		// Method to convert to points
		l.PushGoFunction(func(l *lua.State) int {
			sp := CheckScaledPoint(l, 1)
			l.PushNumber(sp.Value.ToPT())
			return 1
		})
		return 1
	case "to_mm":
		// Method to convert to millimeters
		l.PushGoFunction(func(l *lua.State) int {
			sp := CheckScaledPoint(l, 1)
			mm, _ := sp.Value.ToUnit("mm")
			l.PushNumber(mm)
			return 1
		})
		return 1
	case "to_cm":
		// Method to convert to centimeters
		l.PushGoFunction(func(l *lua.State) int {
			sp := CheckScaledPoint(l, 1)
			cm, _ := sp.Value.ToUnit("cm")
			l.PushNumber(cm)
			return 1
		})
		return 1
	case "to_in":
		// Method to convert to inches
		l.PushGoFunction(func(l *lua.State) int {
			sp := CheckScaledPoint(l, 1)
			in, _ := sp.Value.ToUnit("in")
			l.PushNumber(in)
			return 1
		})
		return 1
	}

	return 0
}

// RegisterScaledPointMetaTable creates the ScaledPoint metatable
func RegisterScaledPointMetaTable(l *lua.State) {
	lua.NewMetaTable(l, ScaledPointMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__add", Function: scaledPointAdd},
		{Name: "__sub", Function: scaledPointSub},
		{Name: "__mul", Function: scaledPointMul},
		{Name: "__div", Function: scaledPointDiv},
		{Name: "__unm", Function: scaledPointUnm},
		{Name: "__eq", Function: scaledPointEq},
		{Name: "__lt", Function: scaledPointLt},
		{Name: "__le", Function: scaledPointLe},
		{Name: "__tostring", Function: scaledPointToString},
		{Name: "__index", Function: scaledPointIndex},
	}, 0)
	l.Pop(1)
}

// SpNew creates a ScaledPoint userdata from pt: sp(points)
func SpNew(l *lua.State) int {
	pt := lua.CheckNumber(l, 1)
	sp := bag.ScaledPointFromFloat(pt)
	PushScaledPoint(l, sp)
	return 1
}

// SpFromString creates a ScaledPoint userdata from string: sp_string("12pt")
func SpFromString(l *lua.State) int {
	s := lua.CheckString(l, 1)
	sp, err := bag.SP(s)
	if err != nil {
		lua.Errorf(l, "invalid dimension: %s", err.Error())
		return 0
	}
	PushScaledPoint(l, sp)
	return 1
}

// SpMax returns the maximum of two ScaledPoints
func SpMax(l *lua.State) int {
	sp1, ok1 := ToScaledPointValue(l, 1)
	sp2, ok2 := ToScaledPointValue(l, 2)
	if !ok1 || !ok2 {
		lua.Errorf(l, "invalid arguments for max (expected ScaledPoint, string with unit, or number)")
		return 0
	}
	PushScaledPoint(l, bag.Max(sp1, sp2))
	return 1
}

// SpMin returns the minimum of two ScaledPoints
func SpMin(l *lua.State) int {
	sp1, ok1 := ToScaledPointValue(l, 1)
	sp2, ok2 := ToScaledPointValue(l, 2)
	if !ok1 || !ok2 {
		lua.Errorf(l, "invalid arguments for min (expected ScaledPoint, string with unit, or number)")
		return 0
	}
	PushScaledPoint(l, bag.Min(sp1, sp2))
	return 1
}
