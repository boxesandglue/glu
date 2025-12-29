package frontend

import (
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

const textMetaTable = "Text"
const textSettingsMetaTable = "TextSettings"

// Text wraps the boxesandglue frontend.Text type
type Text struct {
	Value *frontend.Text
}

// TextSettings provides access to Text.Settings
type TextSettings struct {
	text *frontend.Text
}

// checkText retrieves a Text userdata from the stack
func checkText(l *lua.State, index int) *Text {
	ud := lua.CheckUserData(l, index, textMetaTable)
	if t, ok := ud.(*Text); ok {
		return t
	}
	lua.Errorf(l, "Text expected")
	return nil
}

// textNew creates a new Text object: text.new() or text.new({settings})
func textNew(l *lua.State) int {
	te := frontend.NewText()

	// If a table is passed, apply settings
	if l.Top() >= 1 && l.IsTable(1) {
		te.Settings = make(frontend.TypesettingSettings)
		l.PushNil()
		for l.Next(1) {
			if l.IsString(-2) {
				key, _ := l.ToString(-2)
				settingType, value := parseSettingKeyValue(l, key, l.AbsIndex(-1))
				if settingType != 0 {
					te.Settings[settingType] = value
				}
			}
			l.Pop(1)
		}
	}

	l.PushUserData(&Text{Value: te})
	lua.SetMetaTableNamed(l, textMetaTable)
	return 1
}

// textAppend appends content to text: text:append(item, ...)
func textAppend(l *lua.State) int {
	te := checkText(l, 1)
	n := l.Top()

	for i := 2; i <= n; i++ {
		item := luaValueToItem(l, i)
		if item != nil {
			te.Value.Items = append(te.Value.Items, item)
		}
	}

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// textSet sets a setting: text:set(key, value)
func textSet(l *lua.State) int {
	te := checkText(l, 1)
	key := lua.CheckString(l, 2)

	if te.Value.Settings == nil {
		te.Value.Settings = make(frontend.TypesettingSettings)
	}

	settingType, value := parseSettingKeyValue(l, key, 3)
	if settingType != 0 {
		te.Value.Settings[settingType] = value
	}

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// textSetSettings sets multiple settings from a table: text:settings(table)
func textSetSettings(l *lua.State) int {
	te := checkText(l, 1)

	if !l.IsTable(2) {
		lua.Errorf(l, "table expected")
		return 0
	}

	if te.Value.Settings == nil {
		te.Value.Settings = make(frontend.TypesettingSettings)
	}

	l.PushNil()
	for l.Next(2) {
		if l.IsString(-2) {
			key, _ := l.ToString(-2)
			settingType, value := parseSettingKeyValue(l, key, l.AbsIndex(-1))
			if settingType != 0 {
				te.Value.Settings[settingType] = value
			}
		}
		l.Pop(1)
	}

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// textIndex handles attribute access (__index metamethod)
func textIndex(l *lua.State) int {
	te := checkText(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "append":
		l.PushGoFunction(textAppend)
		return 1
	case "set":
		l.PushGoFunction(textSet)
		return 1
	case "settings":
		// Return a proxy table for settings access (txt.settings.font_family = ...)
		l.PushUserData(&TextSettings{text: te.Value})
		lua.SetMetaTableNamed(l, textSettingsMetaTable)
		return 1
	case "apply":
		// Set multiple settings from a table: txt:apply({font_family = ff, ...})
		l.PushGoFunction(textSetSettings)
		return 1
	}

	return 0
}

// textNewIndex handles attribute setting (__newindex metamethod)
func textNewIndex(l *lua.State) int {
	te := checkText(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "items":
		// Set items from a Lua table
		if l.IsTable(3) {
			te.Value.Items = nil
			l.PushNil()
			for l.Next(3) {
				item := luaValueToItem(l, -1)
				if item != nil {
					te.Value.Items = append(te.Value.Items, item)
				}
				l.Pop(1)
			}
		}
	}
	return 0
}

// textSettingsIndex handles settings attribute access (__index metamethod)
func textSettingsIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, textSettingsMetaTable)
	ts, ok := ud.(*TextSettings)
	if !ok {
		return 0
	}

	key := lua.CheckString(l, 2)
	if ts.text.Settings == nil {
		return 0
	}

	// Look up the setting
	settingType := settingKeyToType(key)
	if settingType == 0 {
		return 0
	}

	if val, exists := ts.text.Settings[settingType]; exists {
		pushSettingValue(l, settingType, val)
		return 1
	}
	return 0
}

// textSettingsNewIndex handles settings attribute setting (__newindex metamethod)
func textSettingsNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, textSettingsMetaTable)
	ts, ok := ud.(*TextSettings)
	if !ok {
		return 0
	}

	key := lua.CheckString(l, 2)

	if ts.text.Settings == nil {
		ts.text.Settings = make(frontend.TypesettingSettings)
	}

	settingType, value := parseSettingKeyValue(l, key, 3)
	if settingType != 0 {
		ts.text.Settings[settingType] = value
	}
	return 0
}

func settingKeyToType(key string) frontend.SettingType {
	switch key {
	case "fontfamily", "font_family":
		return frontend.SettingFontFamily
	case "fontweight", "font_weight":
		return frontend.SettingFontWeight
	case "fontstyle", "font_style":
		return frontend.SettingStyle
	case "size", "fontsize", "font_size":
		return frontend.SettingSize
	case "color":
		return frontend.SettingColor
	case "leading":
		return frontend.SettingLeading
	case "halign", "align":
		return frontend.SettingHAlign
	case "valign":
		return frontend.SettingVAlign
	}
	return 0
}

func pushSettingValue(l *lua.State, settingType frontend.SettingType, val any) {
	switch settingType {
	case frontend.SettingHAlign:
		if h, ok := val.(frontend.HorizontalAlignment); ok {
			l.PushString(halignToString(h))
		}
	case frontend.SettingVAlign:
		if v, ok := val.(frontend.VerticalAlignment); ok {
			l.PushString(valignToString(v))
		}
	case frontend.SettingFontWeight:
		if w, ok := val.(frontend.FontWeight); ok {
			l.PushInteger(int(w))
		}
	default:
		l.PushNil()
	}
}

// registerTextMetaTable creates the Text metatable
func registerTextMetaTable(l *lua.State) {
	lua.NewMetaTable(l, textMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: textIndex},
		{Name: "__newindex", Function: textNewIndex},
	}, 0)
	l.Pop(1)
}

// registerTextSettingsMetaTable creates the TextSettings metatable
func registerTextSettingsMetaTable(l *lua.State) {
	lua.NewMetaTable(l, textSettingsMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: textSettingsIndex},
		{Name: "__newindex", Function: textSettingsNewIndex},
	}, 0)
	l.Pop(1)
}

// luaValueToItem converts a Lua value to a Text item
func luaValueToItem(l *lua.State, index int) any {
	switch {
	case l.IsString(index):
		s, _ := l.ToString(index)
		return s
	case l.IsUserData(index):
		// Check for Text
		if ud := lua.TestUserData(l, index, textMetaTable); ud != nil {
			if t, ok := ud.(*Text); ok {
				return t.Value
			}
		}
		// Check for Table
		if ud := lua.TestUserData(l, index, tableMetaTable); ud != nil {
			if t, ok := ud.(*Table); ok {
				return t.Value
			}
		}
		// Check for VList
		if ud := lua.TestUserData(l, index, vlistMetaTable); ud != nil {
			if v, ok := ud.(*VList); ok {
				return v.Value
			}
		}
	}
	return nil
}

// parseSettingKeyValue parses a setting key and value from Lua
func parseSettingKeyValue(l *lua.State, key string, valueIndex int) (frontend.SettingType, any) {
	switch key {
	case "fontfamily", "font_family":
		if ud := lua.TestUserData(l, valueIndex, fontFamilyMetaTable); ud != nil {
			if ff, ok := ud.(*FontFamily); ok {
				return frontend.SettingFontFamily, ff.Value
			}
		}
	case "fontweight", "font_weight":
		if l.IsString(valueIndex) {
			s, _ := l.ToString(valueIndex)
			return frontend.SettingFontWeight, frontend.ResolveFontWeight(s, frontend.FontWeight400)
		} else if l.IsNumber(valueIndex) {
			n, _ := l.ToInteger(valueIndex)
			return frontend.SettingFontWeight, frontend.FontWeight(n)
		}
	case "fontstyle", "font_style":
		if l.IsString(valueIndex) {
			s, _ := l.ToString(valueIndex)
			return frontend.SettingStyle, frontend.ResolveFontStyle(s)
		}
	case "size", "fontsize", "font_size":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingSize, sp
		}
	case "color":
		if l.IsString(valueIndex) {
			// Color will be resolved later
			s, _ := l.ToString(valueIndex)
			return frontend.SettingColor, s
		}
		if ud := lua.TestUserData(l, valueIndex, colorMetaTable); ud != nil {
			if c, ok := ud.(*Color); ok {
				return frontend.SettingColor, c.Value
			}
		}
	case "leading":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingLeading, sp
		}
	case "halign", "align":
		if l.IsString(valueIndex) {
			s, _ := l.ToString(valueIndex)
			return frontend.SettingHAlign, parseHAlign(s)
		}
	case "valign":
		if l.IsString(valueIndex) {
			s, _ := l.ToString(valueIndex)
			return frontend.SettingVAlign, parseVAlign(s)
		}
	case "marginleft", "margin_left":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingMarginLeft, sp
		}
	case "marginright", "margin_right":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingMarginRight, sp
		}
	case "margintop", "margin_top":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingMarginTop, sp
		}
	case "marginbottom", "margin_bottom":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingMarginBottom, sp
		}
	case "paddingleft", "padding_left":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingPaddingLeft, sp
		}
	case "paddingright", "padding_right":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingPaddingRight, sp
		}
	case "paddingtop", "padding_top":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingPaddingTop, sp
		}
	case "paddingbottom", "padding_bottom":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingPaddingBottom, sp
		}
	case "backgroundcolor", "background_color":
		if l.IsString(valueIndex) {
			s, _ := l.ToString(valueIndex)
			return frontend.SettingBackgroundColor, s
		}
		if ud := lua.TestUserData(l, valueIndex, colorMetaTable); ud != nil {
			if c, ok := ud.(*Color); ok {
				return frontend.SettingBackgroundColor, c.Value
			}
		}
	case "indentleft", "indent_left":
		if sp, err := toDimension(l, valueIndex); err == nil {
			return frontend.SettingIndentLeft, sp
		}
	case "hyperlink":
		if l.IsString(valueIndex) {
			s, _ := l.ToString(valueIndex)
			return frontend.SettingHyperlink, s
		}
	case "underline":
		if l.ToBoolean(valueIndex) {
			return frontend.SettingTextDecorationLine, frontend.TextDecorationUnderline
		}
	case "linethrough", "line_through":
		if l.ToBoolean(valueIndex) {
			return frontend.SettingTextDecorationLine, frontend.TextDecorationLineThrough
		}
	}
	return 0, nil
}

func parseHAlign(s string) frontend.HorizontalAlignment {
	switch s {
	case "left":
		return frontend.HAlignLeft
	case "right":
		return frontend.HAlignRight
	case "center":
		return frontend.HAlignCenter
	case "justified", "justify":
		return frontend.HAlignJustified
	default:
		return frontend.HAlignDefault
	}
}

func parseVAlign(s string) frontend.VerticalAlignment {
	switch s {
	case "top":
		return frontend.VAlignTop
	case "middle", "center":
		return frontend.VAlignMiddle
	case "bottom":
		return frontend.VAlignBottom
	default:
		return frontend.VAlignDefault
	}
}
