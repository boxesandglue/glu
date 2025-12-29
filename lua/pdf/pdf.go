package pdf

import (
	"os"

	pdf "github.com/boxesandglue/baseline-pdf"
	"github.com/speedata/go-lua"
)

const pdfMetaTable = "PDF"

// PDF wraps the baseline-pdf PDF type
type PDF struct {
	Value *pdf.PDF
	file  *os.File
}

// checkPDF retrieves a PDF userdata from the stack
func checkPDF(l *lua.State, index int) *PDF {
	ud := lua.CheckUserData(l, index, pdfMetaTable)
	if p, ok := ud.(*PDF); ok {
		return p
	}
	lua.Errorf(l, "PDF expected")
	return nil
}

// pdfNew creates a new PDF writer: pdf.new(filename)
func pdfNew(l *lua.State) int {
	filename := lua.CheckString(l, 1)

	f, err := os.Create(filename)
	if err != nil {
		lua.Errorf(l, "failed to create file: %s", err.Error())
		return 0
	}

	p := &PDF{
		Value: pdf.NewPDFWriter(f),
		file:  f,
	}

	l.PushUserData(p)
	lua.SetMetaTableNamed(l, pdfMetaTable)
	return 1
}

// pdfFinish finalizes the PDF: pdf:finish()
func pdfFinish(l *lua.State) int {
	p := checkPDF(l, 1)
	if err := p.Value.Finish(); err != nil {
		lua.Errorf(l, "failed to finish PDF: %s", err.Error())
		return 0
	}
	if p.file != nil {
		p.file.Close()
	}
	return 0
}

// pdfNewObject creates a new PDF object: pdf:new_object()
func pdfNewObject(l *lua.State) int {
	p := checkPDF(l, 1)
	obj := p.Value.NewObject()

	l.PushUserData(&Object{Value: obj})
	lua.SetMetaTableNamed(l, objectMetaTable)
	return 1
}

// pdfAddPage adds a page: pdf:add_page(stream_object)
func pdfAddPage(l *lua.State) int {
	p := checkPDF(l, 1)
	obj := checkObject(l, 2)

	page := p.Value.AddPage(obj.Value, 0)

	l.PushUserData(&Page{Value: page})
	lua.SetMetaTableNamed(l, pageMetaTable)
	return 1
}

// pdfLoadFace loads a font: pdf:load_face(filename, [index])
func pdfLoadFace(l *lua.State) int {
	p := checkPDF(l, 1)
	filename := lua.CheckString(l, 2)
	idx := lua.OptInteger(l, 3, 0)

	face, err := p.Value.LoadFace(filename, idx)
	if err != nil {
		lua.Errorf(l, "failed to load face: %s", err.Error())
		return 0
	}

	l.PushUserData(&Face{Value: face})
	lua.SetMetaTableNamed(l, faceMetaTable)
	return 1
}

// pdfLoadImage loads an image: pdf:load_image(filename)
func pdfLoadImage(l *lua.State) int {
	p := checkPDF(l, 1)
	filename := lua.CheckString(l, 2)

	img, err := p.Value.LoadImageFile(filename)
	if err != nil {
		lua.Errorf(l, "failed to load image: %s", err.Error())
		return 0
	}

	l.PushUserData(&Imagefile{Value: img})
	lua.SetMetaTableNamed(l, imagefileMetaTable)
	return 1
}

// pdfIndex handles attribute access (__index metamethod)
func pdfIndex(l *lua.State) int {
	p := checkPDF(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "default_page_width":
		l.PushNumber(p.Value.DefaultPageWidth)
		return 1
	case "default_page_height":
		l.PushNumber(p.Value.DefaultPageHeight)
		return 1
	case "default_offset_x":
		l.PushNumber(p.Value.DefaultOffsetX)
		return 1
	case "default_offset_y":
		l.PushNumber(p.Value.DefaultOffsetY)
		return 1
	// Methods
	case "add_page":
		l.PushGoFunction(pdfAddPage)
		return 1
	case "finish":
		l.PushGoFunction(pdfFinish)
		return 1
	case "load_face":
		l.PushGoFunction(pdfLoadFace)
		return 1
	case "load_image":
		l.PushGoFunction(pdfLoadImage)
		return 1
	case "new_object":
		l.PushGoFunction(pdfNewObject)
		return 1
	}

	return 0
}

// pdfNewIndex handles attribute setting (__newindex metamethod)
func pdfNewIndex(l *lua.State) int {
	p := checkPDF(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "default_page_width":
		p.Value.DefaultPageWidth = lua.CheckNumber(l, 3)
	case "default_page_height":
		p.Value.DefaultPageHeight = lua.CheckNumber(l, 3)
	case "default_offset_x":
		p.Value.DefaultOffsetX = lua.CheckNumber(l, 3)
	case "default_offset_y":
		p.Value.DefaultOffsetY = lua.CheckNumber(l, 3)
	default:
		lua.Errorf(l, "cannot set attribute %s on PDF", key)
	}
	return 0
}

// openPDF creates the pdf module table for require("glu.pdf")
func openPDF(l *lua.State) int {
	// Create PDF metatable
	lua.NewMetaTable(l, pdfMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: pdfIndex},
		{Name: "__newindex", Function: pdfNewIndex},
	}, 0)
	l.Pop(1)

	// Create Object, Page, Face, Imagefile metatables
	registerObjectMetaTable(l)
	registerPageMetaTable(l)
	registerFaceMetaTable(l)
	registerImagefileMetaTable(l)

	// Create the pdf module table
	lua.NewLibrary(l, []lua.RegistryFunction{
		{Name: "new", Function: pdfNew},
	})
	return 1
}

// Open registers the pdf module for require() in the Lua state.
func Open(l *lua.State) {
	lua.Require(l, "glu.pdf", openPDF, false)
	l.Pop(1)
}
