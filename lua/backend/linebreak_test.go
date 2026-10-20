package backend

import (
	"strings"
	"testing"

	"github.com/speedata/go-lua"
)

// runNode runs script in a Lua state with the backend modules.
func runNode(t *testing.T, script string) error {
	t.Helper()
	l := lua.NewState()
	lua.OpenLibraries(l)
	Open(l)
	return lua.DoString(l, `local node = require("glu.node")
-- a run of n rules of 10pt, each one a word, with glue between
local function words(n)
	local head, tail
	for i = 1, n do
		if i > 1 then
			local g = node.new("glue")
			g.width = 3 * 65536
			g.stretch = 2 * 65536
			g.shrink = 65536
			node.insert_after(head, tail, g)
			tail = g
		end
		local r = node.new("rule")
		r.width = 10 * 65536
		r.height = 8 * 65536
		if head == nil then head = r else node.insert_after(head, tail, r) end
		tail = r
	end
	return head, tail
end
local function count(vl)
	local n = 0
	local cur = vl.list
	while cur do n = n + 1; cur = cur.next end
	return n
end
`+script)
}

func TestNodeLinebreak(t *testing.T) {
	err := runNode(t, `
local head, tail = words(12)
node.append_lineend(head, tail)
-- four words of 10pt and three spaces of 3pt fit in 49pt
local vl, lines = node.linebreak(head, { hsize = "49pt", line_height = "12pt", line_penalty = 100, tolerance = 2 })
assert(#lines == 3, "lines: " .. #lines)
assert(lines[1].emergency == false and lines[1].demerits > 0)
assert(vl.type == "Vlist")
-- each line and the glue below it
assert(count(vl) == 6, "nodes in the vlist: " .. count(vl))
`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNodeLinebreakReplaceCount(t *testing.T) {
	err := runNode(t, `
-- a word of two rules with a disc between: a break there drops the 6pt
-- rule after the disc (the ligature) and sets pre and post instead
local head, tail = words(1)
local d = node.new("disc")
local pre = node.new("rule"); pre.width = 4 * 65536
local post = node.new("rule"); post.width = 5 * 65536
d.pre, d.post, d.replace_count = pre, post, 1
assert(d.replace_count == 1)
node.insert_after(head, tail, d)
local lig = node.new("rule"); lig.width = 6 * 65536
node.insert_after(head, d, lig)
local rest = node.new("rule"); rest.width = 10 * 65536
node.insert_after(head, lig, rest)
node.append_lineend(head, rest)
local vl, lines = node.linebreak(head, { hsize = "14pt", tolerance = 10 })
assert(#lines == 2, "lines: " .. #lines)
local widths = {}
local line = vl.list
while line do
	if line.type == "HList" then
		local cur = line.list
		while cur do
			if cur.type == "Rule" then widths[#widths + 1] = cur.width // 65536 end
			cur = cur.next
		end
	end
	line = line.next
end
-- 10 4 | 5 10: the 6pt rule is gone
assert(table.concat(widths, " ") == "10 4 5 10", table.concat(widths, " "))
`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNodeLinebreakErrors(t *testing.T) {
	for _, c := range []struct{ settings, want string }{
		{`{}`, "hsize must be greater than 0"},
		{`{ hsize = "10pt" }`, "must end with a forced break"},
		{`{ hsize = "10pt", hsized = 1 }`, "unknown setting hsized"},
		{`{ hsize = "10pt", tolerance = -1 }`, "tolerance must not be negative"},
		{`{ hsize = "10pt", direction = "ttb" }`, "direction must be ltr or rtl"},
		{`{ hsize = "10pt", indent_steps = { { "x", 1 } } }`, "indent_steps[1].inset"},
		{`{ hsize = "10pt", line_end_glue = 3 }`, "line_end_glue: glue node expected"},
	} {
		script := "local h, t = words(2)\nnode.append_lineend(h, t)\nnode.linebreak(h, " + c.settings + ")"
		if strings.Contains(c.want, "forced break") {
			script = "node.linebreak(words(2), " + c.settings + ")"
		}
		err := runNode(t, script)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, want %q", c.settings, err, c.want)
		}
	}
}
