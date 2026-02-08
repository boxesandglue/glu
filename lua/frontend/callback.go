package frontend

import (
	"fmt"
	"log/slog"

	"github.com/boxesandglue/boxesandglue/backend/document"
	fe "github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/htmlbag"
	"github.com/speedata/go-lua"
)

// callbackEntry is a named callback stored in the registry.
type callbackEntry struct {
	name        string
	registryKey string // key in the Lua registry table
}

// CallbackRegistry manages named, ordered callbacks per event.
type CallbackRegistry struct {
	l          *lua.State
	feDoc      *fe.Document        // set when a document is created
	cssBuilder *htmlbag.CSSBuilder // set after CSS builder is created
	entries    map[string][]*callbackEntry
}

// SetCSSBuilder stores the CSSBuilder reference so that callbacks
// can access page dimensions and CSS information.
func (cr *CallbackRegistry) SetCSSBuilder(cb *htmlbag.CSSBuilder) {
	cr.cssBuilder = cb
}

// package-level registry, initialized in Open()
var registry *CallbackRegistry

// InitRegistry creates a new CallbackRegistry for the given Lua state.
func InitRegistry(l *lua.State) {
	registry = &CallbackRegistry{
		l:       l,
		entries: make(map[string][]*callbackEntry),
	}
}

// GetRegistry returns the package-level callback registry.
func GetRegistry() *CallbackRegistry {
	return registry
}

// registryKey builds a unique Lua registry key for a callback.
func registryKey(event, name string) string {
	return "glu.cb." + event + "." + name
}

// Add inserts a callback. position is "back" (default), "front", or a relative spec.
func (cr *CallbackRegistry) Add(event, name string, position callbackPosition) {
	// Remove existing entry with the same name first
	cr.Remove(event, name)

	key := registryKey(event, name)
	entry := &callbackEntry{name: name, registryKey: key}

	// The Lua function is on top of the stack — store it in the Lua registry
	cr.l.SetField(lua.RegistryIndex, key)

	list := cr.entries[event]
	switch position.kind {
	case posBack:
		cr.entries[event] = append(list, entry)
	case posFront:
		cr.entries[event] = append([]*callbackEntry{entry}, list...)
	case posBefore:
		cr.entries[event] = insertRelative(list, entry, position.target, true)
	case posAfter:
		cr.entries[event] = insertRelative(list, entry, position.target, false)
	}

	slog.Debug("Callback registered", "event", event, "name", name, "position", position.String())
}

// Remove removes a named callback from an event.
func (cr *CallbackRegistry) Remove(event, name string) {
	list := cr.entries[event]
	for i, e := range list {
		if e.name == name {
			// Clear the Lua registry entry
			cr.l.PushNil()
			cr.l.SetField(lua.RegistryIndex, e.registryKey)
			cr.entries[event] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// List returns the callback names in execution order for the given event.
func (cr *CallbackRegistry) List(event string) []string {
	list := cr.entries[event]
	names := make([]string, len(list))
	for i, e := range list {
		names[i] = e.name
	}
	return names
}

// Fire calls all registered Lua callbacks for an event.
// setupArgs is called to push arguments onto the Lua stack and returns the arg count.
func (cr *CallbackRegistry) Fire(event string, setupArgs func(*lua.State) int) error {
	list := cr.entries[event]
	if len(list) == 0 {
		return nil
	}
	for _, e := range list {
		cr.l.Field(lua.RegistryIndex, e.registryKey)
		if cr.l.IsNil(-1) {
			cr.l.Pop(1)
			continue
		}
		nargs := 0
		if setupArgs != nil {
			nargs = setupArgs(cr.l)
		}
		if err := cr.l.ProtectedCall(nargs, 0, 0); err != nil {
			return fmt.Errorf("callback %q (%s): %w", e.name, event, err)
		}
	}
	return nil
}

// InstallPreShipout registers a Go-level PreShipout callback on the document
// that fires all registered Lua "pre_shipout" callbacks.
func (cr *CallbackRegistry) InstallPreShipout(feDoc *fe.Document) {
	cr.feDoc = feDoc
	feDoc.Doc.RegisterCallback(document.CallbackPreShipout, func(page *document.Page) {
		pagenum := len(feDoc.Doc.Pages)
		err := cr.Fire("pre_shipout", func(l *lua.State) int {
			// arg 1: Document
			l.PushUserData(&Document{Value: feDoc})
			lua.SetMetaTableNamed(l, documentMetaTable)
			// arg 2: Page
			l.PushUserData(&Page{Value: page})
			lua.SetMetaTableNamed(l, pageMetaTable)
			// arg 3: page number
			l.PushInteger(pagenum)
			// arg 4: PageInfo (page dimensions + CSS access)
			if cr.cssBuilder != nil {
				if pd, err := cr.cssBuilder.PageSize(); err == nil {
					l.PushUserData(&PageInfo{
						cssBuilder: cr.cssBuilder,
						dimensions: pd,
					})
					lua.SetMetaTableNamed(l, pageInfoMetaTable)
				} else {
					l.PushNil()
				}
			} else {
				l.PushNil()
			}
			return 4
		})
		if err != nil {
			slog.Error("pre_shipout callback error", "error", err)
		}
	})
}

// InstallPostElement registers a Go-level ElementCallback on the CSSBuilder
// that fires all registered Lua "post_element" callbacks.
func (cr *CallbackRegistry) InstallPostElement() {
	cr.cssBuilder.ElementCallback = func(event htmlbag.ElementEvent) {
		err := cr.Fire("post_element", func(l *lua.State) int {
			// arg 1: ElementInfo
			l.PushUserData(&ElementInfo{event: event})
			lua.SetMetaTableNamed(l, elementInfoMetaTable)
			// arg 2: Document (for creating SVG nodes etc.)
			if cr.feDoc != nil {
				l.PushUserData(&Document{Value: cr.feDoc})
				lua.SetMetaTableNamed(l, documentMetaTable)
				return 2
			}
			return 1
		})
		if err != nil {
			slog.Error("post_element callback error", "error", err)
		}
	}
}

// InstallPageInit registers a Go-level PageInitCallback on the CSSBuilder
// that fires all registered Lua "page_init" callbacks.
func (cr *CallbackRegistry) InstallPageInit() {
	cr.cssBuilder.PageInitCallback = func() {
		page := cr.feDoc.Doc.CurrentPage
		pagenum := len(cr.feDoc.Doc.Pages)
		err := cr.Fire("page_init", func(l *lua.State) int {
			// arg 1: Document
			l.PushUserData(&Document{Value: cr.feDoc})
			lua.SetMetaTableNamed(l, documentMetaTable)
			// arg 2: Page
			l.PushUserData(&Page{Value: page})
			lua.SetMetaTableNamed(l, pageMetaTable)
			// arg 3: page number
			l.PushInteger(pagenum)
			// arg 4: PageInfo
			if pd, err := cr.cssBuilder.PageSize(); err == nil {
				l.PushUserData(&PageInfo{
					cssBuilder: cr.cssBuilder,
					dimensions: pd,
				})
				lua.SetMetaTableNamed(l, pageInfoMetaTable)
			} else {
				l.PushNil()
			}
			return 4
		})
		if err != nil {
			slog.Error("page_init callback error", "error", err)
		}
	}
}

// --- Position types ---

type positionKind int

const (
	posBack positionKind = iota
	posFront
	posBefore
	posAfter
)

type callbackPosition struct {
	kind   positionKind
	target string // for before/after
}

func (p callbackPosition) String() string {
	switch p.kind {
	case posFront:
		return "front"
	case posBefore:
		return "before:" + p.target
	case posAfter:
		return "after:" + p.target
	default:
		return "back"
	}
}

func insertRelative(list []*callbackEntry, entry *callbackEntry, target string, before bool) []*callbackEntry {
	for i, e := range list {
		if e.name == target {
			if before {
				return append(list[:i], append([]*callbackEntry{entry}, list[i:]...)...)
			}
			return append(list[:i+1], append([]*callbackEntry{entry}, list[i+1:]...)...)
		}
	}
	// Target not found: append
	return append(list, entry)
}

// --- Lua binding functions ---

// luaAddCallback implements frontend.add_callback(event, name, fn [, position])
func luaAddCallback(l *lua.State) int {
	event := lua.CheckString(l, 1)
	name := lua.CheckString(l, 2)
	lua.CheckType(l, 3, lua.TypeFunction)

	pos := parsePosition(l, 4)

	// Push the function to the top (it's at index 3)
	l.PushValue(3)
	// Add stores it from the top of the stack
	registry.Add(event, name, pos)

	return 0
}

// luaRemoveCallback implements frontend.remove_callback(event, name)
func luaRemoveCallback(l *lua.State) int {
	event := lua.CheckString(l, 1)
	name := lua.CheckString(l, 2)
	registry.Remove(event, name)
	return 0
}

// luaListCallbacks implements frontend.list_callbacks(event)
func luaListCallbacks(l *lua.State) int {
	event := lua.CheckString(l, 1)
	names := registry.List(event)
	l.NewTable()
	for i, n := range names {
		l.PushString(n)
		l.RawSetInt(-2, i+1)
	}
	return 1
}

// parsePosition reads the optional 4th argument for positioning.
func parsePosition(l *lua.State, index int) callbackPosition {
	if l.Top() < index || l.IsNil(index) {
		return callbackPosition{kind: posBack}
	}
	if l.IsString(index) {
		s := lua.CheckString(l, index)
		if s == "front" {
			return callbackPosition{kind: posFront}
		}
		return callbackPosition{kind: posBack}
	}
	if l.IsTable(index) {
		l.Field(index, "before")
		if !l.IsNil(-1) {
			target := lua.CheckString(l, -1)
			l.Pop(1)
			return callbackPosition{kind: posBefore, target: target}
		}
		l.Pop(1)

		l.Field(index, "after")
		if !l.IsNil(-1) {
			target := lua.CheckString(l, -1)
			l.Pop(1)
			return callbackPosition{kind: posAfter, target: target}
		}
		l.Pop(1)
	}
	return callbackPosition{kind: posBack}
}
