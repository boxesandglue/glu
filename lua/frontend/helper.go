package frontend

import (
	"fmt"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

// toDimension converts a Lua value (ScaledPoint userdata, number, or string) to bag.ScaledPoint
// ScaledPoint userdata is used directly, numbers are interpreted as points,
// strings can have units like "12pt", "1cm", "10mm"
func toDimension(l *lua.State, index int) (bag.ScaledPoint, error) {
	// Check for ScaledPoint userdata first
	if sp := testScaledPoint(l, index); sp != nil {
		return sp.Value, nil
	}
	if l.IsNumber(index) {
		n, _ := l.ToNumber(index)
		return bag.ScaledPointFromFloat(n), nil
	}
	if l.IsString(index) {
		s, _ := l.ToString(index)
		return bag.SP(s)
	}
	return 0, fmt.Errorf("expected ScaledPoint, number or string with unit")
}

// checkDimension is like toDimension but calls lua.Errorf on error
func checkDimension(l *lua.State, index int) bag.ScaledPoint {
	sp, err := toDimension(l, index)
	if err != nil {
		lua.Errorf(l, "invalid dimension: %s", err.Error())
		return 0
	}
	return sp
}

// optDimension returns a dimension or a default value
func optDimension(l *lua.State, index int, def bag.ScaledPoint) bag.ScaledPoint {
	if l.IsNoneOrNil(index) {
		return def
	}
	sp, err := toDimension(l, index)
	if err != nil {
		return def
	}
	return sp
}

// tableToTypesettingOptions converts a Lua table to TypesettingOptions
func tableToTypesettingOptions(l *lua.State, index int, doc *frontend.Document) []frontend.TypesettingOption {
	var opts []frontend.TypesettingOption
	absIndex := l.AbsIndex(index)

	l.Field(absIndex, "leading")
	if sp, err := toDimension(l, -1); err == nil {
		opts = append(opts, frontend.Leading(sp))
	}
	l.Pop(1)

	l.Field(absIndex, "font_size")
	if sp, err := toDimension(l, -1); err == nil {
		opts = append(opts, frontend.FontSize(sp))
	} else {
		l.Pop(1)
		l.Field(absIndex, "fontsize")
		if sp, err := toDimension(l, -1); err == nil {
			opts = append(opts, frontend.FontSize(sp))
		}
	}
	l.Pop(1)

	l.Field(absIndex, "font_family")
	if ud := lua.TestUserData(l, -1, fontFamilyMetaTable); ud != nil {
		if ff, ok := ud.(*FontFamily); ok {
			opts = append(opts, frontend.Family(ff.Value))
		}
	} else {
		l.Pop(1)
		l.Field(absIndex, "fontfamily")
		if ud := lua.TestUserData(l, -1, fontFamilyMetaTable); ud != nil {
			if ff, ok := ud.(*FontFamily); ok {
				opts = append(opts, frontend.Family(ff.Value))
			}
		}
	}
	l.Pop(1)

	l.Field(absIndex, "language")
	if ud := lua.TestUserData(l, -1, languageMetaTable); ud != nil {
		if lang, ok := ud.(*Language); ok {
			opts = append(opts, frontend.Language(lang.Value))
		}
	}
	l.Pop(1)

	l.Field(absIndex, "halign")
	if l.IsString(-1) {
		s, _ := l.ToString(-1)
		opts = append(opts, frontend.HorizontalAlign(parseHAlign(s)))
	}
	l.Pop(1)

	for _, side := range []struct {
		key    string
		option func(bag.ScaledPoint, int) frontend.TypesettingOption
	}{{"indent_left", frontend.IndentLeft}, {"indent_right", frontend.IndentRight}} {
		l.Field(absIndex, side.key)
		if !l.IsNil(-1) {
			sp, err := toDimension(l, -1)
			if err != nil {
				lua.Errorf(l, "%s: %s", side.key, err.Error())
			}
			rows := 1
			l.Field(absIndex, side.key+"_rows")
			if l.IsNumber(-1) {
				rows, _ = l.ToInteger(-1)
			}
			l.Pop(1)
			opts = append(opts, side.option(sp, rows))
		}
		l.Pop(1)
	}

	// The line breaker's values. 0 keeps the default here; the text
	// settings of the same names take 0 as a value.
	l.Field(absIndex, "linebreak_tolerance")
	if f, ok := l.ToNumber(-1); ok {
		opts = append(opts, frontend.Tolerance(f))
	}
	l.Pop(1)

	l.Field(absIndex, "linebreak_emergency_stretch")
	if !l.IsNil(-1) {
		sp, err := toDimension(l, -1)
		if err != nil {
			lua.Errorf(l, "linebreak_emergency_stretch: %s", err.Error())
		}
		opts = append(opts, frontend.EmergencyStretch(sp))
	}
	l.Pop(1)

	l.Field(absIndex, "linebreak_hyphen_penalty")
	if n, ok := l.ToInteger(-1); ok {
		opts = append(opts, frontend.HyphenPenalty(n))
	}
	l.Pop(1)

	return opts
}
