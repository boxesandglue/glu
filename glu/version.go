package main

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
)

// printVersion writes a multi-line version banner: tagged version,
// Go runtime + GOOS/GOARCH, VCS commit + timestamp + dirty flag.
// VCS info is best-effort — when glu is built from a non-VCS source
// tree (e.g. `go install` from a tarball, or `go run`) the build
// info has no vcs.revision and those lines are silently omitted.
func printVersion(w io.Writer) {
	v := Version
	if v == "" {
		// `go build` without ldflags leaves Version empty. Show
		// (devel) so the output is unambiguous instead of a bare
		// trailing space.
		v = "(devel)"
	}
	fmt.Fprintf(w, "glu version %s\n", v)
	fmt.Fprintf(w, "  go: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok {
		var rev, vcsTime, modified string
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.time":
				vcsTime = s.Value
			case "vcs.modified":
				modified = s.Value
			}
		}
		if rev != "" {
			if vcsTime != "" {
				fmt.Fprintf(w, "  commit: %s (%s)\n", rev, vcsTime)
			} else {
				fmt.Fprintf(w, "  commit: %s\n", rev)
			}
			if modified == "true" {
				fmt.Fprintln(w, "  dirty: true (uncommitted changes in build tree)")
			}
		}
	}
}
