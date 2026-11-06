package markdown

import (
	"strings"
	"testing"
)

// Chroma lays out each line of a code block with display: flex, which htmlbag
// would report as if the document had asked for it.
func TestCodeBlockCarriesNoFlex(t *testing.T) {
	html, err := markdownToHTML("```go\nx := 1\ny := 2\n```\n", Frontmatter{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "display:flex") {
		t.Errorf("flex layout left in the code block: %s", html)
	}
	if !strings.Contains(html, "color:") {
		t.Errorf("the code block lost its highlighting: %s", html)
	}
}
