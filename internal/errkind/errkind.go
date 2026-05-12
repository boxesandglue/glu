// Package errkind defines sentinel errors for the glu CLI so that
// errors raised deep in the pipeline can be classified by main() and
// mapped to distinct process exit codes. Wrap concrete errors with
// fmt.Errorf("...: %w", errkind.Xxx) so the kind survives.
package errkind

import "errors"

var (
	// Usage: bad flags, unknown subcommand, missing argument.
	Usage = errors.New("usage error")
	// IO: file not found, permission denied, write failure.
	IO = errors.New("io error")
	// Lua: companion script load failure, Lua block runtime error.
	Lua = errors.New("lua error")
	// Typeset: boxesandglue / htmlbag failure during layout or PDF write.
	Typeset = errors.New("typesetting error")
	// AuxNotConverged: aux file kept changing past --max-passes.
	AuxNotConverged = errors.New("aux did not converge")
)
