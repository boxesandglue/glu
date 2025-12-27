package frontend

import (
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

const documentMetaTable = "Document"

// Document wraps the boxesandglue frontend.Document type
type Document struct {
	Value *frontend.Document
}

// checkDocument retrieves a Document userdata from the stack
func checkDocument(l *lua.State, index int) *Document {
	ud := lua.CheckUserData(l, index, documentMetaTable)
	if d, ok := ud.(*Document); ok {
		return d
	}
	lua.Errorf(l, "Document expected")
	return nil
}

// documentNew creates a new document: frontend.new(filename)
func documentNew(l *lua.State) int {
	filename := lua.CheckString(l, 1)

	doc, err := frontend.New(filename)
	if err != nil {
		lua.Errorf(l, "failed to create document: %s", err.Error())
		return 0
	}

	l.PushUserData(&Document{Value: doc})
	lua.SetMetaTableNamed(l, documentMetaTable)
	return 1
}

// documentFinish finalizes the document: doc:finish()
func documentFinish(l *lua.State) int {
	d := checkDocument(l, 1)
	if err := d.Value.Finish(); err != nil {
		lua.Errorf(l, "failed to finish document: %s", err.Error())
		return 0
	}
	return 0
}

// documentNewFontFamily creates a new font family: doc:new_font_family(name)
func documentNewFontFamily(l *lua.State) int {
	d := checkDocument(l, 1)
	name := lua.CheckString(l, 2)

	ff := d.Value.NewFontFamily(name)

	l.PushUserData(&FontFamily{Value: ff, doc: d.Value})
	lua.SetMetaTableNamed(l, fontFamilyMetaTable)
	return 1
}

// documentFindFontFamily finds an existing font family: doc:find_font_family(name)
func documentFindFontFamily(l *lua.State) int {
	d := checkDocument(l, 1)
	name := lua.CheckString(l, 2)

	ff := d.Value.FindFontFamily(name)
	if ff == nil {
		l.PushNil()
		return 1
	}

	l.PushUserData(&FontFamily{Value: ff, doc: d.Value})
	lua.SetMetaTableNamed(l, fontFamilyMetaTable)
	return 1
}

// documentLoadFace loads a font face: doc:load_face(fontsource)
func documentLoadFace(l *lua.State) int {
	d := checkDocument(l, 1)
	fs := checkFontSource(l, 2)

	face, err := d.Value.LoadFace(fs.Value)
	if err != nil {
		lua.Errorf(l, "failed to load face: %s", err.Error())
		return 0
	}

	l.PushUserData(&Face{Value: face})
	lua.SetMetaTableNamed(l, faceMetaTable)
	return 1
}

// documentCreateText creates a new Text object: doc:create_text()
func documentCreateText(l *lua.State) int {
	te := frontend.NewText()
	l.PushUserData(&Text{Value: te})
	lua.SetMetaTableNamed(l, textMetaTable)
	return 1
}

// documentFormatParagraph formats a paragraph: doc:format_paragraph(text, width, [options])
// width can be a number (points) or string with unit ("400pt", "15cm")
func documentFormatParagraph(l *lua.State) int {
	d := checkDocument(l, 1)
	te := checkText(l, 2)
	hsize := checkDimension(l, 3)

	// Collect options if provided
	var opts []frontend.TypesettingOption
	if l.Top() >= 4 && l.IsTable(4) {
		opts = tableToTypesettingOptions(l, 4, d.Value)
	}

	vlist, info, err := d.Value.FormatParagraph(te.Value, hsize, opts...)
	if err != nil {
		lua.Errorf(l, "format paragraph failed: %s", err.Error())
		return 0
	}

	// Return vlist and info
	l.PushUserData(&VList{Value: vlist})
	lua.SetMetaTableNamed(l, vlistMetaTable)

	// Push paragraph info as table
	l.NewTable()
	l.PushNumber(info.Height.ToPT())
	l.SetField(-2, "height")
	l.PushNumber(info.Depth.ToPT())
	l.SetField(-2, "depth")

	return 2
}

// documentBuildTable builds a table: doc:build_table(table)
func documentBuildTable(l *lua.State) int {
	d := checkDocument(l, 1)
	tbl := checkTable(l, 2)

	vlists, err := d.Value.BuildTable(tbl.Value)
	if err != nil {
		lua.Errorf(l, "build table failed: %s", err.Error())
		return 0
	}

	// Return array of vlists
	l.NewTable()
	for i, vl := range vlists {
		l.PushUserData(&VList{Value: vl})
		lua.SetMetaTableNamed(l, vlistMetaTable)
		l.RawSetInt(-2, i+1)
	}

	return 1
}

// documentDefineColor defines a named color: doc:define_color(name, color)
func documentDefineColor(l *lua.State) int {
	d := checkDocument(l, 1)
	name := lua.CheckString(l, 2)
	col := checkColor(l, 3)

	d.Value.DefineColor(name, col.Value)
	return 0
}

// documentGetColor gets a color by name or CSS: doc:get_color(spec)
func documentGetColor(l *lua.State) int {
	d := checkDocument(l, 1)
	spec := lua.CheckString(l, 2)

	col := d.Value.GetColor(spec)
	if col == nil {
		l.PushNil()
		return 1
	}

	l.PushUserData(&Color{Value: col})
	lua.SetMetaTableNamed(l, colorMetaTable)
	return 1
}

// documentGetLanguage gets a language: doc:get_language(name)
func documentGetLanguage(l *lua.State) int {
	langname := lua.CheckString(l, 2)

	lang, err := frontend.GetLanguage(langname)
	if err != nil {
		lua.Errorf(l, "language not found: %s", err.Error())
		return 0
	}

	l.PushUserData(&Language{Value: lang})
	lua.SetMetaTableNamed(l, languageMetaTable)
	return 1
}

// documentNewPage creates a new page: doc:new_page()
func documentNewPage(l *lua.State) int {
	d := checkDocument(l, 1)

	page := d.Value.Doc.NewPage()

	l.PushUserData(&Page{Value: page})
	lua.SetMetaTableNamed(l, pageMetaTable)
	return 1
}

// documentIndex handles attribute access (__index metamethod)
func documentIndex(l *lua.State) int {
	key := lua.CheckString(l, 2)

	switch key {
	// Methods
	case "finish":
		l.PushGoFunction(documentFinish)
		return 1
	case "new_font_family":
		l.PushGoFunction(documentNewFontFamily)
		return 1
	case "find_font_family":
		l.PushGoFunction(documentFindFontFamily)
		return 1
	case "load_face":
		l.PushGoFunction(documentLoadFace)
		return 1
	case "create_text":
		l.PushGoFunction(documentCreateText)
		return 1
	case "format_paragraph":
		l.PushGoFunction(documentFormatParagraph)
		return 1
	case "build_table":
		l.PushGoFunction(documentBuildTable)
		return 1
	case "define_color":
		l.PushGoFunction(documentDefineColor)
		return 1
	case "get_color":
		l.PushGoFunction(documentGetColor)
		return 1
	case "get_language":
		l.PushGoFunction(documentGetLanguage)
		return 1
	case "new_page":
		l.PushGoFunction(documentNewPage)
		return 1
	}

	return 0
}

// registerDocumentMetaTable creates the Document metatable
func registerDocumentMetaTable(l *lua.State) {
	lua.NewMetaTable(l, documentMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: documentIndex},
	}, 0)
	l.Pop(1)
}
