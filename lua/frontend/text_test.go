package frontend

import (
	"strings"
	"testing"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

// runText runs script in a Lua state with glu.frontend and returns the Text
// the script leaves in the global t.
func runText(t *testing.T, script string) (*frontend.Text, error) {
	t.Helper()
	l := lua.NewState()
	lua.OpenLibraries(l)
	Open(l)
	if err := lua.DoString(l, `local frontend = require("glu.frontend")
`+script); err != nil {
		return nil, err
	}
	l.Global("t")
	te := checkText(l, -1)
	return te.Value, nil
}

func TestTextLinebreakSettings(t *testing.T) {
	te, err := runText(t, `
t = frontend.text()
t:apply({
	linebreak_tolerance = 0,
	linebreak_emergency_stretch = "2pt",
	linebreak_hyphen_penalty = 0,
	indent_left_steps = { { inset = "10pt", rows = 2 }, { 0, 0 } },
	indent_right_steps = { { inset = 5, rows = 1 } },
	background_area = "ascent_descent",
	record_line_trims = true,
})
-- read back, the values must come as they were set
local s = t.settings
assert(s.linebreak_tolerance == 0)
assert(tostring(s.linebreak_emergency_stretch) == "2pt")
assert(s.linebreak_hyphen_penalty == 0)
assert(#s.indent_left_steps == 2 and s.indent_left_steps[1].rows == 2)
assert(tostring(s.indent_right_steps[1].inset) == "5pt")
assert(s.background_area == "ascent_descent")
assert(s.record_line_trims == true)
`)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := te.Settings[frontend.SettingLinebreakTolerance].(float64); !ok || v != 0 {
		t.Errorf("tolerance %v, want 0", te.Settings[frontend.SettingLinebreakTolerance])
	}
	if v := te.Settings[frontend.SettingLinebreakEmergencyStretch]; v != bag.MustSP("2pt") {
		t.Errorf("emergency stretch %v, want 2pt", v)
	}
	if v, ok := te.Settings[frontend.SettingHyphenPenalty].(int); !ok || v != 0 {
		t.Errorf("hyphen penalty %v, want 0", te.Settings[frontend.SettingHyphenPenalty])
	}
	want := []node.IndentStep{{Inset: bag.MustSP("10pt"), Rows: 2}, {Inset: 0, Rows: 0}}
	if got, _ := te.Settings[frontend.SettingIndentLeftSteps].([]node.IndentStep); len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("left steps %v, want %v", got, want)
	}
	if got, _ := te.Settings[frontend.SettingIndentRightSteps].([]node.IndentStep); len(got) != 1 || got[0] != (node.IndentStep{Inset: bag.MustSP("5pt"), Rows: 1}) {
		t.Errorf("right steps %v, want 5pt for 1 row", got)
	}
	if v := te.Settings[frontend.SettingBackgroundArea]; v != frontend.BackgroundAreaAscentDescent {
		t.Errorf("background area %v, want ascent_descent", v)
	}
	if v := te.Settings[frontend.SettingRecordLineTrims]; v != true {
		t.Errorf("record line trims %v, want true", v)
	}
}

func TestTextSettingErrors(t *testing.T) {
	for _, c := range []struct{ script, want string }{
		{`linebreak_tolerance = -1`, "linebreak_tolerance must not be negative"},
		{`background_area = "padding"`, "background_area must be em_box or ascent_descent"},
		{`indent_left_steps = { "10pt" }`, "indent_left_steps[1]: table expected"},
		{`indent_left_steps = { { inset = "x" } }`, "indent_left_steps[1].inset"},
	} {
		_, err := runText(t, "t = frontend.text()\nt:apply({ "+c.script+" })")
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, want %q", c.script, err, c.want)
		}
	}
}

func TestTextDecorationLines(t *testing.T) {
	te, err := runText(t, `
t = frontend.text({ underline = true, line_through = true })
assert(t.settings.underline and t.settings.line_through and not t.settings.overline)
t:set("overline", true):set("underline", false)
assert(not t.settings.underline and t.settings.overline)
`)
	if err != nil {
		t.Fatal(err)
	}
	want := frontend.TextDecorationOverline | frontend.TextDecorationLineThrough
	if got := te.Settings[frontend.SettingTextDecorationLine]; got != want {
		t.Errorf("decoration lines %v, want %v", got, want)
	}
	te, err = runText(t, `
t = frontend.text({ underline = true })
t.settings.underline = false
`)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := te.Settings[frontend.SettingTextDecorationLine]; ok {
		t.Errorf("decoration lines %v, want none", v)
	}
}
