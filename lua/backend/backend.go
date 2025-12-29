// Package backend provides Lua bindings for the boxes and glue backend.
//
// This includes:
//   - bag: ScaledPoint operations and logging
//   - node: Node types and list operations
//   - font: Font shaping
package backend

import (
	"github.com/speedata/go-lua"
)

// Open registers the backend modules for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "glu", openGlu, false)
	l.Pop(1)
	lua.Require(l, "glu.node", openNode, false)
	l.Pop(1)
	lua.Require(l, "glu.font", openFont, false)
	l.Pop(1)
}
