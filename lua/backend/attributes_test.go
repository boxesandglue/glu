package backend

import (
	"strings"
	"testing"
)

func TestImageCrop(t *testing.T) {
	err := runNode(t, `
local img = node.new("image")
assert(img.crop == nil)
img.crop = { x = 10, y = 20, width = 100.5, height = 50 }
local c = img.crop
assert(c.x == 10 and c.y == 20 and c.width == 100.5 and c.height == 50)
img.crop = nil
assert(img.crop == nil)
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ crop, want string }{
		{`{ width = 0, height = 10 }`, "crop: width and height must be greater than 0"},
		{`{ x = "a", width = 1, height = 1 }`, "crop.x: number expected"},
		{`"all"`, "crop: table expected"},
	} {
		err := runNode(t, `node.new("image").crop = `+c.crop)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, want %q", c.crop, err, c.want)
		}
	}
}

func TestPaintLast(t *testing.T) {
	err := runNode(t, `
local vl = node.new("vlist")
assert(vl.paint_last == false and vl.paint_last_scope == false)
vl.paint_last, vl.paint_last_scope = true, true
assert(vl.paint_last and vl.paint_last_scope)
vl.paint_last = false
assert(not vl.paint_last and vl.paint_last_scope)
local r = node.new("rule")
r.paint_last = true
assert(r.paint_last)
`)
	if err != nil {
		t.Fatal(err)
	}
}
