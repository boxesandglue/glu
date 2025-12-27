package frontend

import (
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

const textMetaTable = "Text"

// Text wraps the boxesandglue frontend.Text type
type Text struct {
	Value *frontend.Text
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

// textNew creates a new Text object: text.new()
func textNew(l *lua.State) int {
	te := frontend.NewText()
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
	key := lua.CheckString(l, 2)

	switch key {
	case "append":
		l.PushGoFunction(textAppend)
		return 1
	case "set":
		l.PushGoFunction(textSet)
		return 1
	case "settings":
		l.PushGoFunction(textSetSettings)
		return 1
	}

	return 0
}

// registerTextMetaTable creates the Text metatable
func registerTextMetaTable(l *lua.State) {
	lua.NewMetaTable(l, textMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: textIndex},
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
