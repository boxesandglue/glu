package frontend

import (
	"strings"
	"testing"

	"github.com/speedata/go-lua"
)

func runLua(t *testing.T, script string) error {
	t.Helper()
	l := lua.NewState()
	lua.OpenLibraries(l)
	Open(l)
	return lua.DoString(l, `local frontend = require("glu.frontend")
`+script)
}

// The setter works on a color of its own: get_color hands out a shared color
// for a name, and setting overprint on it must not reach other users.
func TestColorOverprint(t *testing.T) {
	err := runLua(t, `
local doc = frontend.new("/dev/null")
local first = doc:get_color("black")
local ink = doc:get_color("black")
assert(ink.overprint == false)
ink.overprint = true
assert(ink.overprint == true)
assert(first.overprint == false)
assert(doc:get_color("black").overprint == false)

local cmyk = doc:get_color("device-cmyk(0 0 0 1)")
cmyk.overprint = true
doc:define_color("ink", cmyk)
assert(doc:get_color("ink").overprint == true)

-- a defined color stays until it is defined again
cmyk.overprint = false
assert(doc:get_color("ink").overprint == true)
doc:define_color("ink", cmyk)
assert(doc:get_color("ink").overprint == false)
`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestColorSetterErrors(t *testing.T) {
	for script, want := range map[string]string{
		`frontend.color(0, 0, 0).overprint = "yes"`: "boolean expected",
		`frontend.color(0, 0, 0).r = 1`:             "no writable attribute r",
	} {
		err := runLua(t, script)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: error %v, want %q", script, err, want)
		}
	}
}
