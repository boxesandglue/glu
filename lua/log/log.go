package log

import (
	"fmt"
	"log/slog"

	"github.com/speedata/glu/lua/common"
	"github.com/speedata/go-lua"
)

// argsFromStack reads the message and optional key-value pairs from the Lua stack.
// Stack layout: (message [, key1, value1, key2, value2, ...])
func argsFromStack(l *lua.State) (string, []any) {
	msg := lua.CheckString(l, 1)
	n := l.Top()
	var attrs []any
	for i := 2; i <= n-1; i += 2 {
		key, _ := l.ToString(i)
		val := common.LuaToGo(l, i+1)
		attrs = append(attrs, key, val)
	}
	// Odd trailing argument without value: include as-is
	if n >= 2 && (n-1)%2 != 0 {
		key, _ := l.ToString(n)
		attrs = append(attrs, key, nil)
	}
	return msg, attrs
}

func luaDebug(l *lua.State) int {
	msg, attrs := argsFromStack(l)
	slog.Debug(msg, attrs...)
	return 0
}

func luaInfo(l *lua.State) int {
	msg, attrs := argsFromStack(l)
	slog.Info(msg, attrs...)
	return 0
}

func luaWarn(l *lua.State) int {
	msg, attrs := argsFromStack(l)
	slog.Warn(msg, attrs...)
	return 0
}

func luaError(l *lua.State) int {
	msg, attrs := argsFromStack(l)
	slog.Error(msg, attrs...)
	return 0
}

func openLog(l *lua.State) int {
	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "debug", Function: luaDebug},
		{Name: "info", Function: luaInfo},
		{Name: "warn", Function: luaWarn},
		{Name: "error", Function: luaError},
	})
	l.PushString(fmt.Sprintf("glu.log (%d functions)", 4))
	l.SetField(-2, "_description")
	return 1
}

// Open registers the log module for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "glu.log", openLog, false)
	l.Pop(1)
}
