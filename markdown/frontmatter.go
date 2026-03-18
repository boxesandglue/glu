package markdown

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds metadata extracted from the YAML front matter block.
type Frontmatter struct {
	Title     string         `yaml:"title"`
	Author    string         `yaml:"author"`
	CSS       string         `yaml:"css"`
	Papersize string         `yaml:"papersize"`
	Format    string         `yaml:"format"`
	Lang      string         `yaml:"lang"`
	Extra     map[string]any `yaml:"-"` // all key-value pairs (including the known ones)
}

// extractFrontmatter separates YAML front matter from the Markdown body.
// The front matter must be delimited by "---" lines at the beginning of the file.
// Returns the parsed front matter and the remaining Markdown content.
func extractFrontmatter(source string) (Frontmatter, string) {
	var fm Frontmatter

	trimmed := strings.TrimLeft(source, " \t\r\n")
	if !strings.HasPrefix(trimmed, "---") {
		return fm, source
	}

	// Find the closing "---"
	rest := trimmed[3:]
	// Skip the newline after opening ---
	if idx := strings.IndexByte(rest, '\n'); idx >= 0 {
		rest = rest[idx+1:]
	} else {
		return fm, source
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return fm, source
	}

	yamlBlock := rest[:endIdx]
	body := rest[endIdx+4:] // skip "\n---"
	// Skip the newline after closing ---
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	_ = yaml.Unmarshal([]byte(yamlBlock), &fm)
	// Also parse into a generic map so all keys are available.
	var extra map[string]any
	_ = yaml.Unmarshal([]byte(yamlBlock), &extra)
	if extra == nil {
		extra = make(map[string]any)
	}
	fm.Extra = extra
	return fm, body
}
