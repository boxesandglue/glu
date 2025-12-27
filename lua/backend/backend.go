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

// Open registers the backend modules (bag, node, font) as globals in the Lua state.
func Open(l *lua.State) {
	registerBagModule(l)
	registerNodeModule(l)
	registerFontModule(l)
}
