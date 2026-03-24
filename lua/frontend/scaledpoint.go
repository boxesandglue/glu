package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/glu/lua/common"
	"github.com/speedata/go-lua"
)

// Re-export common functions for use within frontend package
func pushScaledPoint(l *lua.State, sp bag.ScaledPoint) {
	common.PushScaledPoint(l, sp)
}

func testScaledPoint(l *lua.State, index int) *common.ScaledPoint {
	return common.TestScaledPoint(l, index)
}

func registerScaledPointMetaTable(l *lua.State) {
	common.RegisterScaledPointMetaTable(l)
}

func spNew(l *lua.State) int {
	return common.SpNew(l)
}

func spFromString(l *lua.State) int {
	return common.SpFromString(l)
}
