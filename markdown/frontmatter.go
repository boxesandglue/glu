package markdown

import (
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds metadata extracted from the YAML front matter block.
type Frontmatter struct {
	Title       string           `yaml:"title"`
	Author      string           `yaml:"author"`
	CSS         string           `yaml:"css"`
	Papersize   string           `yaml:"papersize"`
	Format      string           `yaml:"format"`
	Lang        string           `yaml:"lang"`
	Math        bool             `yaml:"math"`
	Extensions  ExtensionList    `yaml:"extensions"`
	Attachments []AttachmentSpec `yaml:"attachments"`
	Extra       map[string]any   `yaml:"-"` // all key-value pairs (including the known ones)
}

// ExtensionList is the value of the "extensions" frontmatter key: Markdown
// extension names in pandoc spelling (fenced_divs, smart, …), a leading
// "-" disables a default-enabled extension. Both a comma- or
// space-separated scalar and a YAML sequence are accepted:
//
//	extensions: smart, auto_identifiers, -footnotes
//	extensions: [smart, -footnotes]
type ExtensionList []string

// UnmarshalYAML implements yaml.Unmarshaler.
func (e *ExtensionList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*e = splitExtensionNames(value.Value)
	case yaml.SequenceNode:
		var raw []string
		if err := value.Decode(&raw); err != nil {
			return err
		}
		var out ExtensionList
		for _, entry := range raw {
			out = append(out, splitExtensionNames(entry)...)
		}
		*e = out
	}
	// Other node kinds are ignored, matching the lenient frontmatter
	// parsing elsewhere in this file.
	return nil
}

// splitExtensionNames splits a scalar extension spec on commas and
// whitespace and drops empty entries.
func splitExtensionNames(s string) ExtensionList {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	return ExtensionList(fields)
}

// AttachmentSpec is a single entry in the Frontmatter Attachments list. File is
// the only required field; the path is resolved relative to the Markdown
// source file at render time. Name defaults to basename(File); MimeType
// defaults to "application/octet-stream".
type AttachmentSpec struct {
	File        string `yaml:"file"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	MimeType    string `yaml:"mimetype"`
}

// ExtractFrontmatter separates YAML front matter from the Markdown body.
// The front matter must be delimited by "---" lines at the beginning of the file.
// Returns the parsed front matter and the remaining Markdown content.
func ExtractFrontmatter(source string) (Frontmatter, string) {
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
