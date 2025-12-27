package backend

import (
	"log/slog"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/speedata/go-lua"
)

// spFromString creates a ScaledPoint from string: bag.sp("12pt")
func spFromString(l *lua.State) int {
	s := lua.CheckString(l, 1)
	sp, err := bag.SP(s)
	if err != nil {
		lua.Errorf(l, "invalid dimension: %s", err.Error())
		return 0
	}
	l.PushInteger(int(sp))
	return 1
}

// spFromPT creates a ScaledPoint from points: bag.sp_from_pt(12)
func spFromPT(l *lua.State) int {
	pt := lua.CheckNumber(l, 1)
	sp := bag.ScaledPointFromFloat(pt)
	l.PushInteger(int(sp))
	return 1
}

// spToPT converts a ScaledPoint to points: bag.sp_to_pt(sp)
func spToPT(l *lua.State) int {
	sp := bag.ScaledPoint(lua.CheckInteger(l, 1))
	l.PushNumber(sp.ToPT())
	return 1
}

// spToUnit converts a ScaledPoint to a unit: bag.sp_to_unit(sp, "cm")
func spToUnit(l *lua.State) int {
	sp := bag.ScaledPoint(lua.CheckInteger(l, 1))
	unit := lua.CheckString(l, 2)
	val, err := sp.ToUnit(unit)
	if err != nil {
		lua.Errorf(l, "invalid unit: %s", err.Error())
		return 0
	}
	l.PushNumber(val)
	return 1
}

// spMax returns the maximum of two ScaledPoints: bag.max(sp1, sp2)
func spMax(l *lua.State) int {
	sp1 := bag.ScaledPoint(lua.CheckInteger(l, 1))
	sp2 := bag.ScaledPoint(lua.CheckInteger(l, 2))
	l.PushInteger(int(bag.Max(sp1, sp2)))
	return 1
}

// spMin returns the minimum of two ScaledPoints: bag.min(sp1, sp2)
func spMin(l *lua.State) int {
	sp1 := bag.ScaledPoint(lua.CheckInteger(l, 1))
	sp2 := bag.ScaledPoint(lua.CheckInteger(l, 2))
	l.PushInteger(int(bag.Min(sp1, sp2)))
	return 1
}

// logDebug logs a debug message: bag.debug(msg, key1, val1, ...)
func logDebug(l *lua.State) int {
	msg := lua.CheckString(l, 1)
	args := collectLogArgs(l, 2)
	slog.Debug(msg, args...)
	return 0
}

// logInfo logs an info message: bag.info(msg, key1, val1, ...)
func logInfo(l *lua.State) int {
	msg := lua.CheckString(l, 1)
	args := collectLogArgs(l, 2)
	slog.Info(msg, args...)
	return 0
}

// logWarn logs a warning message: bag.warn(msg, key1, val1, ...)
func logWarn(l *lua.State) int {
	msg := lua.CheckString(l, 1)
	args := collectLogArgs(l, 2)
	slog.Warn(msg, args...)
	return 0
}

// logError logs an error message: bag.error(msg, key1, val1, ...)
func logError(l *lua.State) int {
	msg := lua.CheckString(l, 1)
	args := collectLogArgs(l, 2)
	slog.Error(msg, args...)
	return 0
}

// collectLogArgs collects key-value pairs from the Lua stack
func collectLogArgs(l *lua.State, startIndex int) []any {
	n := l.Top()
	args := make([]any, 0, n-startIndex+1)
	for i := startIndex; i <= n; i++ {
		switch {
		case l.IsString(i):
			s, _ := l.ToString(i)
			args = append(args, s)
		case l.IsNumber(i):
			n, _ := l.ToNumber(i)
			args = append(args, n)
		case l.IsBoolean(i):
			args = append(args, l.ToBoolean(i))
		default:
			args = append(args, l.TypeOf(i).String())
		}
	}
	return args
}

// registerBagModule registers the bag module
func registerBagModule(l *lua.State) {
	l.NewTable()
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "sp", Function: spFromString},
		{Name: "sp_from_pt", Function: spFromPT},
		{Name: "sp_to_pt", Function: spToPT},
		{Name: "sp_to_unit", Function: spToUnit},
		{Name: "max", Function: spMax},
		{Name: "min", Function: spMin},
		{Name: "debug", Function: logDebug},
		{Name: "info", Function: logInfo},
		{Name: "warn", Function: logWarn},
		{Name: "error", Function: logError},
	}, 0)

	// Add constants
	l.PushInteger(int(bag.Factor))
	l.SetField(-2, "factor")

	l.SetGlobal("bag")
}
