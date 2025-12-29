package backend

import (
	"log/slog"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/speedata/glu/lua/common"
	"github.com/speedata/go-lua"
)

// spFromString creates a ScaledPoint from string: glu.sp("12pt")
// Returns a ScaledPoint userdata
func spFromString(l *lua.State) int {
	return common.SpFromString(l)
}

// spFromPT creates a ScaledPoint from points: glu.sp_from_pt(12)
// Returns a ScaledPoint userdata
func spFromPT(l *lua.State) int {
	return common.SpNew(l)
}

// spToPT converts a ScaledPoint to points: glu.sp_to_pt(sp)
// Accepts ScaledPoint userdata, string, or number
func spToPT(l *lua.State) int {
	sp, ok := common.ToScaledPointValue(l, 1)
	if !ok {
		lua.Errorf(l, "expected ScaledPoint, string with unit, or number")
		return 0
	}
	l.PushNumber(sp.ToPT())
	return 1
}

// spToUnit converts a ScaledPoint to a unit: glu.sp_to_unit(sp, "cm")
// Accepts ScaledPoint userdata, string, or number
func spToUnit(l *lua.State) int {
	sp, ok := common.ToScaledPointValue(l, 1)
	if !ok {
		lua.Errorf(l, "expected ScaledPoint, string with unit, or number")
		return 0
	}
	unit := lua.CheckString(l, 2)
	val, err := sp.ToUnit(unit)
	if err != nil {
		lua.Errorf(l, "invalid unit: %s", err.Error())
		return 0
	}
	l.PushNumber(val)
	return 1
}

// spMax returns the maximum of two ScaledPoints: glu.max(sp1, sp2)
// Returns a ScaledPoint userdata
func spMax(l *lua.State) int {
	return common.SpMax(l)
}

// spMin returns the minimum of two ScaledPoints: glu.min(sp1, sp2)
// Returns a ScaledPoint userdata
func spMin(l *lua.State) int {
	return common.SpMin(l)
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

// openGlu creates the glu module table for require("glu")
func openGlu(l *lua.State) int {
	// Register ScaledPoint metatable (shared with frontend)
	common.RegisterScaledPointMetaTable(l)

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

	return 1
}
