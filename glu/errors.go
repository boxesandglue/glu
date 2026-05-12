package main

import (
	"errors"

	"github.com/boxesandglue/glu/internal/errkind"
)

// exitCode maps an error to a process exit code. Unknown errors get 1;
// nil gets 0. Codes:
//
//	0  success
//	1  generic / unknown error
//	2  usage (bad flags, missing arguments)
//	3  io (file not found, permission denied)
//	4  lua (syntax, runtime in a Lua block or companion script)
//	5  typesetting (boxesandglue / htmlbag failure)
//	6  aux file did not converge after --max-passes
func exitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errkind.Usage):
		return 2
	case errors.Is(err, errkind.IO):
		return 3
	case errors.Is(err, errkind.Lua):
		return 4
	case errors.Is(err, errkind.Typeset):
		return 5
	case errors.Is(err, errkind.AuxNotConverged):
		return 6
	default:
		return 1
	}
}
