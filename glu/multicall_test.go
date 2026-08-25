package main

import "testing"

func TestMulticallName(t *testing.T) {
	testcases := []struct {
		argv0  string
		asName string
		want   string
	}{
		// Plain glu, with and without path, Windows suffix stripped.
		{"glu", "", "glu"},
		{"/usr/local/bin/glu", "", "glu"},
		{"glu.exe", "", "glu"},
		{"./glu.EXE", "", "glu"},
		// Symlink entrypoints dispatch on their basename.
		{"./foproc", "", "foproc"},
		{"foproc.exe", "", "foproc"},
		// Only .exe is a binary suffix: dotted names must survive
		// untouched (used to become "glu-0.0").
		{"./glu-0.0.31", "", "glu-0.0.31"},
		{"report.v2", "", "report.v2"},
		// --as overrides the detected name in both directions.
		{"./glu-nightly", "glu", "glu"},
		{"glu", "report", "report"},
	}
	for _, tc := range testcases {
		if got := multicallName(tc.argv0, tc.asName); got != tc.want {
			t.Errorf("multicallName(%q, %q) = %q, want %q", tc.argv0, tc.asName, got, tc.want)
		}
	}
}
