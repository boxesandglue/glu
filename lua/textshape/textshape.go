// Package textshape provides Lua bindings for the textshape ot package,
// exposing low-level text shaping (HarfBuzz-port) to Lua scripts.
package textshape

import (
	"os"

	"github.com/boxesandglue/textshape/ot"
	"github.com/speedata/go-lua"
)

const (
	fontMetaTable    = "textshape.Font"
	faceMetaTable    = "textshape.Face"
	shaperMetaTable  = "textshape.Shaper"
	bufferMetaTable  = "textshape.Buffer"
	featureMetaTable = "textshape.Feature"
)

// --- Font type ---

// Font wraps an ot.Font for Lua.
type Font struct {
	Value *ot.Font
	data  []byte // keep reference to font data
}

func checkFont(l *lua.State, index int) *Font {
	ud := lua.CheckUserData(l, index, fontMetaTable)
	if f, ok := ud.(*Font); ok {
		return f
	}
	lua.Errorf(l, "textshape.Font expected")
	return nil
}

func fontIndex(l *lua.State) int {
	f := checkFont(l, 1)
	key := lua.CheckString(l, 2)
	switch key {
	case "num_glyphs":
		l.PushInteger(f.Value.NumGlyphs())
		return 1
	}
	return 0
}

func registerFontMetaTable(l *lua.State) {
	lua.NewMetaTable(l, fontMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: fontIndex},
	}, 0)
	l.Pop(1)
}

// --- Face type ---

// Face wraps an ot.Face for Lua.
type Face struct {
	Value *ot.Face
}

func checkFace(l *lua.State, index int) *Face {
	ud := lua.CheckUserData(l, index, faceMetaTable)
	if f, ok := ud.(*Face); ok {
		return f
	}
	lua.Errorf(l, "textshape.Face expected")
	return nil
}

func faceVariationAxes(l *lua.State) int {
	f := checkFace(l, 1)
	axes := f.Value.VariationAxes()
	l.NewTable()
	for i, ax := range axes {
		l.NewTable()
		l.PushString(ax.Tag.String())
		l.SetField(-2, "tag")
		l.PushNumber(float64(ax.MinValue))
		l.SetField(-2, "min")
		l.PushNumber(float64(ax.DefaultValue))
		l.SetField(-2, "default")
		l.PushNumber(float64(ax.MaxValue))
		l.SetField(-2, "max")
		l.RawSetInt(-2, i+1)
	}
	return 1
}

func faceHasVariations(l *lua.State) int {
	f := checkFace(l, 1)
	l.PushBoolean(f.Value.HasVariations())
	return 1
}

// faceGlyphOutline returns the outline segments for a glyph: face:glyph_outline(gid)
// Returns a table of segments on success, or nil if the glyph has no outline.
// Each segment is a table with an "op" field ("M", "L", "Q", or "C") and coordinate fields.
func faceGlyphOutline(l *lua.State) int {
	f := checkFace(l, 1)
	gid := lua.CheckInteger(l, 2)

	outline, ok := f.Value.GlyphOutline(ot.GlyphID(gid))
	if !ok {
		l.PushNil()
		return 1
	}

	l.NewTable()
	for i, seg := range outline.Segments {
		l.NewTable()
		switch seg.Op {
		case ot.SegmentMoveTo:
			l.PushString("M")
			l.SetField(-2, "op")
			l.PushNumber(float64(seg.Args[0].X))
			l.SetField(-2, "x")
			l.PushNumber(float64(seg.Args[0].Y))
			l.SetField(-2, "y")
		case ot.SegmentLineTo:
			l.PushString("L")
			l.SetField(-2, "op")
			l.PushNumber(float64(seg.Args[0].X))
			l.SetField(-2, "x")
			l.PushNumber(float64(seg.Args[0].Y))
			l.SetField(-2, "y")
		case ot.SegmentQuadTo:
			l.PushString("Q")
			l.SetField(-2, "op")
			l.PushNumber(float64(seg.Args[0].X))
			l.SetField(-2, "x1")
			l.PushNumber(float64(seg.Args[0].Y))
			l.SetField(-2, "y1")
			l.PushNumber(float64(seg.Args[1].X))
			l.SetField(-2, "x")
			l.PushNumber(float64(seg.Args[1].Y))
			l.SetField(-2, "y")
		case ot.SegmentCubeTo:
			l.PushString("C")
			l.SetField(-2, "op")
			l.PushNumber(float64(seg.Args[0].X))
			l.SetField(-2, "x1")
			l.PushNumber(float64(seg.Args[0].Y))
			l.SetField(-2, "y1")
			l.PushNumber(float64(seg.Args[1].X))
			l.SetField(-2, "x2")
			l.PushNumber(float64(seg.Args[1].Y))
			l.SetField(-2, "y2")
			l.PushNumber(float64(seg.Args[2].X))
			l.SetField(-2, "x")
			l.PushNumber(float64(seg.Args[2].Y))
			l.SetField(-2, "y")
		}
		l.RawSetInt(-2, i+1)
	}
	return 1
}

func faceIndex(l *lua.State) int {
	f := checkFace(l, 1)
	key := lua.CheckString(l, 2)
	switch key {
	case "upem":
		l.PushInteger(int(f.Value.Upem()))
		return 1
	case "ascender":
		l.PushInteger(int(f.Value.Ascender()))
		return 1
	case "descender":
		l.PushInteger(int(f.Value.Descender()))
		return 1
	case "cap_height":
		l.PushInteger(int(f.Value.CapHeight()))
		return 1
	case "x_height":
		l.PushInteger(int(f.Value.XHeight()))
		return 1
	case "postscript_name":
		l.PushString(f.Value.PostscriptName())
		return 1
	case "family_name":
		l.PushString(f.Value.FamilyName())
		return 1
	case "weight_class":
		l.PushInteger(int(f.Value.WeightClass()))
		return 1
	case "is_italic":
		l.PushBoolean(f.Value.IsItalic())
		return 1
	case "is_fixed_pitch":
		l.PushBoolean(f.Value.IsFixedPitch())
		return 1
	case "is_cff":
		l.PushBoolean(f.Value.IsCFF())
		return 1
	case "has_variations":
		l.PushGoFunction(faceHasVariations)
		return 1
	case "variation_axes":
		l.PushGoFunction(faceVariationAxes)
		return 1
	case "glyph_outline":
		l.PushGoFunction(faceGlyphOutline)
		return 1
	}
	return 0
}

func registerFaceMetaTable(l *lua.State) {
	lua.NewMetaTable(l, faceMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: faceIndex},
	}, 0)
	l.Pop(1)
}

// --- Feature type ---

// Feature wraps an ot.Feature for Lua.
type Feature struct {
	Value ot.Feature
}

func checkFeature(l *lua.State, index int) *Feature {
	ud := lua.CheckUserData(l, index, featureMetaTable)
	if f, ok := ud.(*Feature); ok {
		return f
	}
	lua.Errorf(l, "textshape.Feature expected")
	return nil
}

func featureIndex(l *lua.State) int {
	f := checkFeature(l, 1)
	key := lua.CheckString(l, 2)
	switch key {
	case "tag":
		l.PushString(f.Value.Tag.String())
		return 1
	case "value":
		l.PushInteger(int(f.Value.Value))
		return 1
	}
	return 0
}

func registerFeatureMetaTable(l *lua.State) {
	lua.NewMetaTable(l, featureMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: featureIndex},
	}, 0)
	l.Pop(1)
}

// --- Buffer type ---

// Buffer wraps an ot.Buffer for Lua.
type Buffer struct {
	Value *ot.Buffer
}

func checkBuffer(l *lua.State, index int) *Buffer {
	ud := lua.CheckUserData(l, index, bufferMetaTable)
	if b, ok := ud.(*Buffer); ok {
		return b
	}
	lua.Errorf(l, "textshape.Buffer expected")
	return nil
}

func bufferAddString(l *lua.State) int {
	b := checkBuffer(l, 1)
	s := lua.CheckString(l, 2)
	b.Value.AddString(s)
	return 0
}

func bufferAddCodepoints(l *lua.State) int {
	b := checkBuffer(l, 1)
	lua.CheckType(l, 2, lua.TypeTable)
	n := l.RawLength(2)
	cps := make([]ot.Codepoint, 0, n)
	for i := 1; i <= n; i++ {
		l.RawGetInt(2, i)
		cp, ok := l.ToInteger(-1)
		if !ok {
			lua.Errorf(l, "codepoint at index %d is not an integer", i)
			return 0
		}
		cps = append(cps, ot.Codepoint(cp))
		l.Pop(1)
	}
	b.Value.AddCodepoints(cps)
	return 0
}

func bufferGuessSegmentProperties(l *lua.State) int {
	b := checkBuffer(l, 1)
	b.Value.GuessSegmentProperties()
	return 0
}

func bufferSetDirection(l *lua.State) int {
	b := checkBuffer(l, 1)
	dirStr := lua.CheckString(l, 2)
	var dir ot.Direction
	switch dirStr {
	case "ltr":
		dir = ot.DirectionLTR
	case "rtl":
		dir = ot.DirectionRTL
	case "ttb":
		dir = ot.DirectionTTB
	case "btt":
		dir = ot.DirectionBTT
	default:
		lua.Errorf(l, "invalid direction %q, expected ltr/rtl/ttb/btt", dirStr)
		return 0
	}
	b.Value.SetDirection(dir)
	return 0
}

func bufferSetScript(l *lua.State) int {
	b := checkBuffer(l, 1)
	s := lua.CheckString(l, 2)
	if len(s) != 4 {
		lua.Errorf(l, "script tag must be exactly 4 characters, got %q", s)
		return 0
	}
	b.Value.Script = ot.MakeTag(s[0], s[1], s[2], s[3])
	return 0
}

func bufferSetLanguage(l *lua.State) int {
	b := checkBuffer(l, 1)
	s := lua.CheckString(l, 2)
	if len(s) < 2 || len(s) > 4 {
		lua.Errorf(l, "language tag should be 2-4 characters, got %q", s)
		return 0
	}
	// Pad to 4 bytes with spaces
	padded := s + "    "
	b.Value.Language = ot.MakeTag(padded[0], padded[1], padded[2], padded[3])
	return 0
}

func bufferClear(l *lua.State) int {
	b := checkBuffer(l, 1)
	b.Value.Clear()
	return 0
}

func bufferReverse(l *lua.State) int {
	b := checkBuffer(l, 1)
	b.Value.Reverse()
	return 0
}

func bufferLen(l *lua.State) int {
	b := checkBuffer(l, 1)
	l.PushInteger(b.Value.Len())
	return 1
}

func directionString(d ot.Direction) string {
	switch d {
	case ot.DirectionLTR:
		return "ltr"
	case ot.DirectionRTL:
		return "rtl"
	case ot.DirectionTTB:
		return "ttb"
	case ot.DirectionBTT:
		return "btt"
	default:
		return "invalid"
	}
}

func bufferIndex(l *lua.State) int {
	b := checkBuffer(l, 1)
	key := lua.CheckString(l, 2)
	switch key {
	case "add_string":
		l.PushGoFunction(bufferAddString)
		return 1
	case "add_codepoints":
		l.PushGoFunction(bufferAddCodepoints)
		return 1
	case "guess_segment_properties":
		l.PushGoFunction(bufferGuessSegmentProperties)
		return 1
	case "set_direction":
		l.PushGoFunction(bufferSetDirection)
		return 1
	case "set_script":
		l.PushGoFunction(bufferSetScript)
		return 1
	case "set_language":
		l.PushGoFunction(bufferSetLanguage)
		return 1
	case "clear":
		l.PushGoFunction(bufferClear)
		return 1
	case "reverse":
		l.PushGoFunction(bufferReverse)
		return 1
	case "direction":
		l.PushString(directionString(b.Value.Direction))
		return 1
	case "info":
		info := b.Value.Info
		l.NewTable()
		for i, gi := range info {
			l.NewTable()
			l.PushInteger(int(gi.GlyphID))
			l.SetField(-2, "glyph_id")
			l.PushInteger(gi.Cluster)
			l.SetField(-2, "cluster")
			l.PushInteger(int(gi.Codepoint))
			l.SetField(-2, "codepoint")
			l.RawSetInt(-2, i+1)
		}
		return 1
	case "pos":
		pos := b.Value.Pos
		l.NewTable()
		for i, gp := range pos {
			l.NewTable()
			l.PushInteger(int(gp.XAdvance))
			l.SetField(-2, "x_advance")
			l.PushInteger(int(gp.YAdvance))
			l.SetField(-2, "y_advance")
			l.PushInteger(int(gp.XOffset))
			l.SetField(-2, "x_offset")
			l.PushInteger(int(gp.YOffset))
			l.SetField(-2, "y_offset")
			l.RawSetInt(-2, i+1)
		}
		return 1
	}
	return 0
}

func registerBufferMetaTable(l *lua.State) {
	lua.NewMetaTable(l, bufferMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: bufferIndex},
		{Name: "__len", Function: bufferLen},
	}, 0)
	l.Pop(1)
}

// --- Shaper type ---

// Shaper wraps an ot.Shaper for Lua.
type Shaper struct {
	Value *ot.Shaper
}

func checkShaper(l *lua.State, index int) *Shaper {
	ud := lua.CheckUserData(l, index, shaperMetaTable)
	if s, ok := ud.(*Shaper); ok {
		return s
	}
	lua.Errorf(l, "textshape.Shaper expected")
	return nil
}

// collectFeatures reads features from the Lua stack at the given index.
// The argument can be a table of strings/Feature userdata, or nil.
func collectFeatures(l *lua.State, index int) []ot.Feature {
	if l.IsNoneOrNil(index) {
		return nil
	}
	lua.CheckType(l, index, lua.TypeTable)
	n := l.RawLength(index)
	features := make([]ot.Feature, 0, n)
	for i := 1; i <= n; i++ {
		l.RawGetInt(index, i)
		switch {
		case l.IsString(-1):
			s, _ := l.ToString(-1)
			if feat, ok := ot.FeatureFromString(s); ok {
				features = append(features, feat)
			}
		case l.IsUserData(-1):
			ud := l.ToUserData(-1)
			if f, ok := ud.(*Feature); ok {
				features = append(features, f.Value)
			}
		}
		l.Pop(1)
	}
	return features
}

func shaperShape(l *lua.State) int {
	s := checkShaper(l, 1)
	b := checkBuffer(l, 2)
	features := collectFeatures(l, 3)
	s.Value.Shape(b.Value, features)
	return 0
}

func shaperHasVariations(l *lua.State) int {
	s := checkShaper(l, 1)
	l.PushBoolean(s.Value.HasVariations())
	return 1
}

func shaperSetVariation(l *lua.State) int {
	s := checkShaper(l, 1)
	tagStr := lua.CheckString(l, 2)
	val := lua.CheckNumber(l, 3)
	if len(tagStr) != 4 {
		lua.Errorf(l, "variation tag must be exactly 4 characters, got %q", tagStr)
		return 0
	}
	tag := ot.MakeTag(tagStr[0], tagStr[1], tagStr[2], tagStr[3])
	s.Value.SetVariation(tag, float32(val))
	return 0
}

func shaperSetVariations(l *lua.State) int {
	s := checkShaper(l, 1)
	lua.CheckType(l, 2, lua.TypeTable)
	var variations []ot.Variation
	l.PushNil()
	for l.Next(2) {
		tagStr, ok := l.ToString(-2)
		if !ok || len(tagStr) != 4 {
			lua.Errorf(l, "variation key must be a 4-character string")
			return 0
		}
		val, ok := l.ToNumber(-1)
		if !ok {
			lua.Errorf(l, "variation value must be a number")
			return 0
		}
		tag := ot.MakeTag(tagStr[0], tagStr[1], tagStr[2], tagStr[3])
		variations = append(variations, ot.Variation{Tag: tag, Value: float32(val)})
		l.Pop(1)
	}
	s.Value.SetVariations(variations)
	return 0
}

func shaperSetSyntheticBold(l *lua.State) int {
	s := checkShaper(l, 1)
	x := float32(lua.CheckNumber(l, 2))
	y := float32(lua.CheckNumber(l, 3))
	inPlace := !l.IsNoneOrNil(4) && l.ToBoolean(4)
	s.Value.SetSyntheticBold(x, y, inPlace)
	return 0
}

func shaperSetSyntheticSlant(l *lua.State) int {
	s := checkShaper(l, 1)
	slant := float32(lua.CheckNumber(l, 2))
	s.Value.SetSyntheticSlant(slant)
	return 0
}

func shaperHasGSUB(l *lua.State) int {
	s := checkShaper(l, 1)
	l.PushBoolean(s.Value.HasGSUB())
	return 1
}

func shaperHasGPOS(l *lua.State) int {
	s := checkShaper(l, 1)
	l.PushBoolean(s.Value.HasGPOS())
	return 1
}

func shaperSetDefaultFeatures(l *lua.State) int {
	s := checkShaper(l, 1)
	features := collectFeatures(l, 2)
	s.Value.SetDefaultFeatures(features)
	return 0
}

func shaperIndex(l *lua.State) int {
	_ = checkShaper(l, 1)
	key := lua.CheckString(l, 2)
	switch key {
	case "shape":
		l.PushGoFunction(shaperShape)
		return 1
	case "has_variations":
		l.PushGoFunction(shaperHasVariations)
		return 1
	case "set_variation":
		l.PushGoFunction(shaperSetVariation)
		return 1
	case "set_variations":
		l.PushGoFunction(shaperSetVariations)
		return 1
	case "set_synthetic_bold":
		l.PushGoFunction(shaperSetSyntheticBold)
		return 1
	case "set_synthetic_slant":
		l.PushGoFunction(shaperSetSyntheticSlant)
		return 1
	case "has_gsub":
		l.PushGoFunction(shaperHasGSUB)
		return 1
	case "has_gpos":
		l.PushGoFunction(shaperHasGPOS)
		return 1
	case "set_default_features":
		l.PushGoFunction(shaperSetDefaultFeatures)
		return 1
	}
	return 0
}

func registerShaperMetaTable(l *lua.State) {
	lua.NewMetaTable(l, shaperMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: shaperIndex},
	}, 0)
	l.Pop(1)
}

// --- Module functions ---

// parseFont loads a font file: parse_font(filename, [index])
func parseFont(l *lua.State) int {
	filename := lua.CheckString(l, 1)
	idx := lua.OptInteger(l, 2, 0)

	data, err := os.ReadFile(filename)
	if err != nil {
		lua.Errorf(l, "failed to read font file: %s", err.Error())
		return 0
	}

	font, err := ot.ParseFont(data, idx)
	if err != nil {
		lua.Errorf(l, "failed to parse font: %s", err.Error())
		return 0
	}

	l.PushUserData(&Font{Value: font, data: data})
	lua.SetMetaTableNamed(l, fontMetaTable)
	return 1
}

// newShaper creates a shaper from a font: new_shaper(font)
func newShaper(l *lua.State) int {
	f := checkFont(l, 1)

	shaper, err := ot.NewShaper(f.Value)
	if err != nil {
		lua.Errorf(l, "failed to create shaper: %s", err.Error())
		return 0
	}

	l.PushUserData(&Shaper{Value: shaper})
	lua.SetMetaTableNamed(l, shaperMetaTable)
	return 1
}

// newFace creates a face from a font: new_face(font)
func newFace(l *lua.State) int {
	f := checkFont(l, 1)

	face, err := ot.NewFace(f.Value)
	if err != nil {
		lua.Errorf(l, "failed to create face: %s", err.Error())
		return 0
	}

	l.PushUserData(&Face{Value: face})
	lua.SetMetaTableNamed(l, faceMetaTable)
	return 1
}

// newBuffer creates an empty buffer: new_buffer()
func newBuffer(l *lua.State) int {
	buf := ot.NewBuffer()
	l.PushUserData(&Buffer{Value: buf})
	lua.SetMetaTableNamed(l, bufferMetaTable)
	return 1
}

// featureFunc parses a single feature string: feature(str)
func featureFunc(l *lua.State) int {
	s := lua.CheckString(l, 1)
	feat, ok := ot.FeatureFromString(s)
	if !ok {
		lua.Errorf(l, "invalid feature string: %q", s)
		return 0
	}
	l.PushUserData(&Feature{Value: feat})
	lua.SetMetaTableNamed(l, featureMetaTable)
	return 1
}

// featuresFunc parses comma-separated features: features(str)
func featuresFunc(l *lua.State) int {
	s := lua.CheckString(l, 1)
	feats := ot.ParseFeatures(s)
	l.NewTable()
	for i, feat := range feats {
		l.PushUserData(&Feature{Value: feat})
		lua.SetMetaTableNamed(l, featureMetaTable)
		l.RawSetInt(-2, i+1)
	}
	return 1
}

// openTextshape creates the module table for require("glu.textshape")
func openTextshape(l *lua.State) int {
	registerFontMetaTable(l)
	registerFaceMetaTable(l)
	registerShaperMetaTable(l)
	registerBufferMetaTable(l)
	registerFeatureMetaTable(l)

	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "parse_font", Function: parseFont},
		{Name: "new_shaper", Function: newShaper},
		{Name: "new_face", Function: newFace},
		{Name: "new_buffer", Function: newBuffer},
		{Name: "feature", Function: featureFunc},
		{Name: "features", Function: featuresFunc},
	})
	return 1
}

// Open registers the textshape module for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "glu.textshape", openTextshape, false)
	l.Pop(1)
}
