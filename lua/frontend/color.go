package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/color"
	"github.com/speedata/go-lua"
)

const colorMetaTable = "Color"

// Color wraps the boxesandglue color.Color type
type Color struct {
	Value *color.Color
}

// checkColor retrieves a Color userdata from the stack
func checkColor(l *lua.State, index int) *Color {
	ud := lua.CheckUserData(l, index, colorMetaTable)
	if c, ok := ud.(*Color); ok {
		return c
	}
	lua.Errorf(l, "Color expected")
	return nil
}

// colorNew creates a color from RGB values: color.new(r, g, b, [a])
// Values can be 0-1 scale or 0-255 scale (auto-detected)
func colorNew(l *lua.State) int {
	col := &color.Color{Space: color.ColorRGB}

	if l.IsNumber(1) {
		// RGB or RGBA values (0-255 or 0-1 scale)
		r := lua.CheckNumber(l, 1)
		g := lua.CheckNumber(l, 2)
		b := lua.CheckNumber(l, 3)
		a := lua.OptNumber(l, 4, 1.0)

		// Assume 0-1 scale if all values are <= 1
		if r <= 1 && g <= 1 && b <= 1 {
			col.R = r
			col.G = g
			col.B = b
			col.A = a
		} else {
			// Assume 0-255 scale
			col.R = r / 255.0
			col.G = g / 255.0
			col.B = b / 255.0
			col.A = a
		}
	} else {
		lua.Errorf(l, "color.new requires numeric arguments (r, g, b, [a])")
		return 0
	}

	l.PushUserData(&Color{Value: col})
	lua.SetMetaTableNamed(l, colorMetaTable)
	return 1
}

// colorIndex handles attribute access (__index metamethod)
func colorIndex(l *lua.State) int {
	c := checkColor(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "r", "red":
		if c.Value.Space == color.ColorRGB || c.Value.Space == color.ColorNone {
			l.PushNumber(c.Value.R)
			return 1
		}
	case "g", "green":
		if c.Value.Space == color.ColorRGB || c.Value.Space == color.ColorNone {
			l.PushNumber(c.Value.G)
			return 1
		}
	case "b", "blue":
		if c.Value.Space == color.ColorRGB || c.Value.Space == color.ColorNone {
			l.PushNumber(c.Value.B)
			return 1
		}
	case "a", "alpha":
		l.PushNumber(c.Value.A)
		return 1
	case "c", "cyan":
		if c.Value.Space == color.ColorCMYK {
			l.PushNumber(c.Value.C)
			return 1
		}
	case "m", "magenta":
		if c.Value.Space == color.ColorCMYK {
			l.PushNumber(c.Value.M)
			return 1
		}
	case "y", "yellow":
		if c.Value.Space == color.ColorCMYK {
			l.PushNumber(c.Value.Y)
			return 1
		}
	case "k", "black":
		if c.Value.Space == color.ColorCMYK {
			l.PushNumber(c.Value.K)
			return 1
		}
	}

	return 0
}

// registerColorMetaTable creates the Color metatable
func registerColorMetaTable(l *lua.State) {
	lua.NewMetaTable(l, colorMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: colorIndex},
	}, 0)
	l.Pop(1)
}
