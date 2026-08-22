package markdown

import (
	"path/filepath"
	"testing"
)

// TestResolveFrontmatterCSS guards the resolution contract for the
// front-matter css: reference: relative paths resolve against the
// document's directory, never against the cwd (glu#3).
func TestResolveFrontmatterCSS(t *testing.T) {
	tests := []struct {
		name, baseDir, cssPath, want string
	}{
		{"document-relative", "/doc/dir", "style.css", filepath.Join("/doc/dir", "style.css")},
		{"subdirectory", "/doc/dir", "css/style.css", filepath.Join("/doc/dir", "css/style.css")},
		{"absolute passes through", "/doc/dir", "/other/style.css", "/other/style.css"},
		{"empty baseDir keeps raw path", "", "style.css", "style.css"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveFrontmatterCSS(tt.baseDir, tt.cssPath); got != tt.want {
				t.Errorf("ResolveFrontmatterCSS(%q,%q) = %q; want %q", tt.baseDir, tt.cssPath, got, tt.want)
			}
		})
	}
}
