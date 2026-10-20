package common

import (
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/speedata/go-lua"
)

// ParseCrop reads the value at index as an image crop, a table
// { x = ..., y = ..., width = ..., height = ... } in the units of the
// image's natural size, measured from its top left corner. nil removes
// the crop. name is the key the value was given under, for the error
// messages.
func ParseCrop(l *lua.State, index int, name string) *node.ImageCrop {
	index = l.AbsIndex(index)
	if l.IsNil(index) {
		return nil
	}
	if !l.IsTable(index) {
		lua.Errorf(l, "%s: table expected", name)
	}
	var c node.ImageCrop
	for _, f := range []struct {
		key string
		v   *float64
	}{{"x", &c.X}, {"y", &c.Y}, {"width", &c.Width}, {"height", &c.Height}} {
		l.Field(index, f.key)
		if !l.IsNil(-1) {
			n, ok := l.ToNumber(-1)
			if !ok {
				lua.Errorf(l, "%s.%s: number expected", name, f.key)
			}
			*f.v = n
		}
		l.Pop(1)
	}
	if c.Width <= 0 || c.Height <= 0 {
		lua.Errorf(l, "%s: width and height must be greater than 0", name)
	}
	return &c
}

// PushCrop pushes c as a table, or nil.
func PushCrop(l *lua.State, c *node.ImageCrop) {
	if c == nil {
		l.PushNil()
		return
	}
	l.CreateTable(0, 4)
	l.PushNumber(c.X)
	l.SetField(-2, "x")
	l.PushNumber(c.Y)
	l.SetField(-2, "y")
	l.PushNumber(c.Width)
	l.SetField(-2, "width")
	l.PushNumber(c.Height)
	l.SetField(-2, "height")
}
