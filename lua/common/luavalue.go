package common

import (
	"fmt"

	"github.com/speedata/go-lua"
)

// PushAny pushes a Go value onto the Lua stack as the corresponding Lua type.
// Supported types: nil, bool, float64, int, string, map[string]any, []any.
func PushAny(l *lua.State, v any) {
	switch val := v.(type) {
	case nil:
		l.PushNil()
	case bool:
		l.PushBoolean(val)
	case float64:
		if val == float64(int(val)) {
			l.PushInteger(int(val))
		} else {
			l.PushNumber(val)
		}
	case int:
		l.PushInteger(val)
	case string:
		l.PushString(val)
	case map[string]any:
		l.NewTable()
		for k, sub := range val {
			PushAny(l, sub)
			l.SetField(-2, k)
		}
	case []any:
		l.NewTable()
		for i, sub := range val {
			PushAny(l, sub)
			l.RawSetInt(-2, i+1)
		}
	default:
		l.PushNil()
	}
}

// LuaToGo reads the Lua value at index and returns the Go equivalent.
// Tables become map[string]any or []any, numbers become int or float64,
// strings and booleans map directly.
func LuaToGo(l *lua.State, index int) any {
	if l.IsNoneOrNil(index) {
		return nil
	}
	if l.IsBoolean(index) {
		return l.ToBoolean(index)
	}
	if l.IsNumber(index) {
		n, _ := l.ToNumber(index)
		if n == float64(int(n)) {
			return int(n)
		}
		return n
	}
	if l.IsString(index) {
		s, _ := l.ToString(index)
		return s
	}
	if l.IsTable(index) {
		return LuaTableToGo(l, index)
	}
	return nil
}

// LuaTableToGo converts a Lua table to a Go slice (for array-like tables with
// consecutive integer keys 1..n) or map[string]any (otherwise).
func LuaTableToGo(l *lua.State, index int) any {
	// Convert negative index to absolute (stack-safe for recursive calls).
	if index < 0 {
		index = l.Top() + index + 1
	}

	type entry struct {
		intKey int
		strKey string
		isInt  bool
		val    any
	}

	var entries []entry
	l.PushNil()
	for l.Next(index) {
		val := LuaToGo(l, -1)

		if l.IsNumber(-2) {
			n, _ := l.ToNumber(-2)
			entries = append(entries, entry{intKey: int(n), isInt: true, val: val})
		} else {
			// Copy key before ToString to avoid corrupting the iteration.
			l.PushValue(-2)
			key, _ := l.ToString(-1)
			l.Pop(1)
			entries = append(entries, entry{strKey: key, val: val})
		}

		l.Pop(1) // pop value, keep key for Next
	}

	// Pure array if all keys are consecutive integers 1..n.
	allInt := len(entries) > 0
	maxN := 0
	for _, e := range entries {
		if !e.isInt {
			allInt = false
			break
		}
		if e.intKey > maxN {
			maxN = e.intKey
		}
	}
	if allInt && maxN == len(entries) {
		arr := make([]any, maxN)
		for _, e := range entries {
			arr[e.intKey-1] = e.val
		}
		return arr
	}

	m := make(map[string]any)
	for _, e := range entries {
		if e.isInt {
			m[fmt.Sprintf("%d", e.intKey)] = e.val
		} else {
			m[e.strKey] = e.val
		}
	}
	return m
}
