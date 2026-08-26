package mdext

import (
	"fmt"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// pandocIDs generates heading identifiers following pandoc's
// auto_identifiers algorithm instead of goldmark's default. The important
// difference: non-ASCII letters survive, so "Über uns" becomes "über-uns"
// rather than "ber-uns".
type pandocIDs struct {
	used map[string]bool
}

// NewPandocIDs returns a parser.IDs implementation with pandoc's
// auto_identifiers algorithm: keep letters, digits, '_', '-' and '.',
// lowercase everything, turn spaces into hyphens, strip everything before
// the first letter, fall back to "section" for letterless headings.
func NewPandocIDs() parser.IDs {
	return &pandocIDs{used: map[string]bool{}}
}

func (s *pandocIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	var runes []rune
	for _, r := range string(value) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.':
			runes = append(runes, unicode.ToLower(r))
		case unicode.IsSpace(r):
			runes = append(runes, '-')
		}
	}
	start := 0
	for start < len(runes) && !unicode.IsLetter(runes[start]) {
		start++
	}
	id := string(runes[start:])
	if id == "" {
		id = "section"
	}
	if !s.used[id] {
		s.used[id] = true
		return []byte(id)
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", id, i)
		if !s.used[candidate] {
			s.used[candidate] = true
			return []byte(candidate)
		}
	}
}

func (s *pandocIDs) Put(value []byte) {
	s.used[string(value)] = true
}
