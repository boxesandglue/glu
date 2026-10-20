package common

import "github.com/boxesandglue/boxesandglue/backend/node"

// HasFlag reports whether attrs hold key, an attribute whose value is not
// read, such as document.PaintLast.
func HasFlag(attrs node.H, key string) bool {
	_, ok := attrs[key]
	return ok
}

// SetFlag sets the attribute key in attrs to true, or removes it.
func SetFlag(attrs *node.H, key string, on bool) {
	if !on {
		delete(*attrs, key)
		return
	}
	if *attrs == nil {
		*attrs = node.H{}
	}
	(*attrs)[key] = true
}
