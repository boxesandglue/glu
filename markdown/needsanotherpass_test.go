package markdown

import (
	"encoding/json"
	"testing"

	"github.com/boxesandglue/htmlbag"
)

// aux returns the map as it comes back from the aux file, with float64
// numbers.
func aux(t *testing.T, s string) map[string]any {
	t.Helper()
	m := map[string]any{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestNeedsAnotherPass(t *testing.T) {
	// What a pass writes: ints, as renderHTMLToPDF builds it.
	cur := map[string]any{
		"_pages":    2,
		"_headings": []any{map[string]any{"level": 1, "text": "Title", "page": 1}},
		"_anchors": map[string]any{
			"a": map[string]any{"page": 2, "text": "Title"},
		},
	}
	none := map[string]any{}
	same := aux(t, `{"_pages": 2, "_headings": [{"level": 1, "text": "Title", "page": 1}], "_anchors": {"a": {"page": 2, "text": "Title"}}}`)
	moved := aux(t, `{"_pages": 2, "_headings": [{"level": 1, "text": "Title", "page": 1}], "_anchors": {"a": {"page": 1, "text": "Title"}}}`)
	onePage := aux(t, `{"_pages": 1, "_headings": [{"level": 1, "text": "Title", "page": 1}], "_anchors": {"a": {"page": 2, "text": "Title"}}}`)

	readsA := htmlbag.PreviousPassReads{Anchors: map[string]bool{"a": true}}
	readsMissing := htmlbag.PreviousPassReads{Anchors: map[string]bool{"missing": true}}
	readsPages := htmlbag.PreviousPassReads{Pages: true}

	for _, tc := range []struct {
		name    string
		old     map[string]any
		reads   htmlbag.PreviousPassReads
		userLua bool
		want    bool
	}{
		{"first pass, nothing read", none, htmlbag.PreviousPassReads{}, false, false},
		{"first pass, anchor read", none, readsA, false, true},
		{"first pass, pages read", none, readsPages, false, true},
		{"first pass, user Lua", none, htmlbag.PreviousPassReads{}, true, true},
		{"unchanged, anchor read", same, readsA, false, false},
		{"unchanged, user Lua", same, htmlbag.PreviousPassReads{}, true, false},
		{"anchor moved, anchor read", moved, readsA, false, true},
		{"anchor moved, other anchor read", moved, readsMissing, false, false},
		{"anchor moved, nothing read", moved, htmlbag.PreviousPassReads{}, false, false},
		{"page count changed, pages read", onePage, readsPages, false, true},
		{"page count changed, anchor read", onePage, readsA, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := needsAnotherPass(tc.old, cur, tc.reads, tc.userLua); got != tc.want {
				t.Errorf("needsAnotherPass = %v, want %v", got, tc.want)
			}
		})
	}
}
