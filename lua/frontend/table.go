package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/speedata/go-lua"
)

const tableMetaTable = "Table"
const tableRowMetaTable = "TableRow"
const tableCellMetaTable = "TableCell"

// Table wraps the boxesandglue frontend.Table type
type Table struct {
	Value *frontend.Table
}

// TableRow wraps the boxesandglue frontend.TableRow type
type TableRow struct {
	Value *frontend.TableRow
}

// TableCell wraps the boxesandglue frontend.TableCell type
type TableCell struct {
	Value *frontend.TableCell
}

// checkTable retrieves a Table userdata from the stack
func checkTable(l *lua.State, index int) *Table {
	ud := lua.CheckUserData(l, index, tableMetaTable)
	if t, ok := ud.(*Table); ok {
		return t
	}
	lua.Errorf(l, "Table expected")
	return nil
}

// tableNew creates a new Table: table.new(options)
// Dimensions (max_width, font_size, leading) can be numbers (points) or strings ("12pt", "1cm")
func tableNew(l *lua.State) int {
	tbl := &frontend.Table{}

	if l.IsTable(1) {
		l.Field(1, "max_width")
		if sp, err := toDimension(l, -1); err == nil {
			tbl.MaxWidth = sp
		}
		l.Pop(1)

		l.Field(1, "stretch")
		if l.IsBoolean(-1) {
			tbl.Stretch = l.ToBoolean(-1)
		}
		l.Pop(1)

		l.Field(1, "font_size")
		if sp, err := toDimension(l, -1); err == nil {
			tbl.FontSize = sp
		}
		l.Pop(1)

		l.Field(1, "leading")
		if sp, err := toDimension(l, -1); err == nil {
			tbl.Leading = sp
		}
		l.Pop(1)

		l.Field(1, "font_family")
		if ud := lua.TestUserData(l, -1, fontFamilyMetaTable); ud != nil {
			if ff, ok := ud.(*FontFamily); ok {
				tbl.FontFamily = ff.Value
			}
		}
		l.Pop(1)
	}

	l.PushUserData(&Table{Value: tbl})
	lua.SetMetaTableNamed(l, tableMetaTable)
	return 1
}

// tableAddRow adds a row to the table: tbl:add_row()
func tableAddRow(l *lua.State) int {
	tbl := checkTable(l, 1)

	row := &frontend.TableRow{}
	tbl.Value.Rows = append(tbl.Value.Rows, row)

	l.PushUserData(&TableRow{Value: row})
	lua.SetMetaTableNamed(l, tableRowMetaTable)
	return 1
}

// tableSetColSpec sets column specifications: tbl:set_columns({width1, width2, ...})
// Widths can be numbers (points) or strings ("100pt", "3cm")
func tableSetColSpec(l *lua.State) int {
	tbl := checkTable(l, 1)

	if !l.IsTable(2) {
		lua.Errorf(l, "table expected")
		return 0
	}

	var specs []frontend.ColSpec
	l.PushNil()
	for l.Next(2) {
		if sp, err := toDimension(l, -1); err == nil {
			glue := node.NewGlue()
			glue.Width = sp
			specs = append(specs, frontend.ColSpec{
				ColumnWidth: glue,
			})
		}
		l.Pop(1)
	}

	tbl.Value.ColSpec = specs

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// tableIndex handles attribute access (__index metamethod)
func tableIndex(l *lua.State) int {
	tbl := checkTable(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "max_width":
		l.PushNumber(tbl.Value.MaxWidth.ToPT())
		return 1
	case "stretch":
		l.PushBoolean(tbl.Value.Stretch)
		return 1
	case "add_row":
		l.PushGoFunction(tableAddRow)
		return 1
	case "set_columns":
		l.PushGoFunction(tableSetColSpec)
		return 1
	}

	return 0
}

// tableNewIndex handles attribute setting (__newindex metamethod)
// Dimensions can be numbers (points) or strings ("12pt", "1cm")
func tableNewIndex(l *lua.State) int {
	tbl := checkTable(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "max_width":
		tbl.Value.MaxWidth = checkDimension(l, 3)
	case "stretch":
		tbl.Value.Stretch = l.ToBoolean(3)
	case "font_size":
		tbl.Value.FontSize = checkDimension(l, 3)
	case "leading":
		tbl.Value.Leading = checkDimension(l, 3)
	case "font_family":
		if ud := lua.TestUserData(l, 3, fontFamilyMetaTable); ud != nil {
			if ff, ok := ud.(*FontFamily); ok {
				tbl.Value.FontFamily = ff.Value
			}
		}
	}
	return 0
}

// rowAddCell adds a cell to the row: row:add_cell()
func rowAddCell(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, tableRowMetaTable)
	row, ok := ud.(*TableRow)
	if !ok {
		lua.Errorf(l, "TableRow expected")
		return 0
	}

	cell := &frontend.TableCell{}
	row.Value.Cells = append(row.Value.Cells, cell)

	l.PushUserData(&TableCell{Value: cell})
	lua.SetMetaTableNamed(l, tableCellMetaTable)
	return 1
}

// rowIndex handles attribute access (__index metamethod)
func rowIndex(l *lua.State) int {
	key := lua.CheckString(l, 2)

	switch key {
	case "add_cell":
		l.PushGoFunction(rowAddCell)
		return 1
	}

	return 0
}

// cellSetContents sets cell contents: cell:set_contents(item, ...)
func cellSetContents(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, tableCellMetaTable)
	cell, ok := ud.(*TableCell)
	if !ok {
		lua.Errorf(l, "TableCell expected")
		return 0
	}

	n := l.Top()
	for i := 2; i <= n; i++ {
		item := luaValueToItem(l, i)
		if item != nil {
			cell.Value.Contents = append(cell.Value.Contents, item)
		}
	}

	// Return self for chaining
	l.PushValue(1)
	return 1
}

// cellIndex handles attribute access (__index metamethod)
func cellIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, tableCellMetaTable)
	cell, ok := ud.(*TableCell)
	if !ok {
		return 0
	}

	key := lua.CheckString(l, 2)

	switch key {
	case "set_contents":
		l.PushGoFunction(cellSetContents)
		return 1
	case "halign":
		l.PushString(halignToString(cell.Value.HAlign))
		return 1
	case "valign":
		l.PushString(valignToString(cell.Value.VAlign))
		return 1
	case "colspan":
		l.PushInteger(cell.Value.ExtraColspan + 1)
		return 1
	case "rowspan":
		l.PushInteger(cell.Value.ExtraRowspan + 1)
		return 1
	}

	return 0
}

// cellNewIndex handles attribute setting (__newindex metamethod)
// Dimensions can be numbers (points) or strings ("2pt", "5mm")
func cellNewIndex(l *lua.State) int {
	ud := lua.CheckUserData(l, 1, tableCellMetaTable)
	cell, ok := ud.(*TableCell)
	if !ok {
		return 0
	}

	key := lua.CheckString(l, 2)

	switch key {
	case "halign":
		s, _ := l.ToString(3)
		cell.Value.HAlign = parseHAlign(s)
	case "valign":
		s, _ := l.ToString(3)
		cell.Value.VAlign = parseVAlign(s)
	case "colspan":
		n, _ := l.ToInteger(3)
		cell.Value.ExtraColspan = n - 1
	case "rowspan":
		n, _ := l.ToInteger(3)
		cell.Value.ExtraRowspan = n - 1
	case "padding_left":
		cell.Value.PaddingLeft = checkDimension(l, 3)
	case "padding_right":
		cell.Value.PaddingRight = checkDimension(l, 3)
	case "padding_top":
		cell.Value.PaddingTop = checkDimension(l, 3)
	case "padding_bottom":
		cell.Value.PaddingBottom = checkDimension(l, 3)
	case "border_left_width":
		cell.Value.BorderLeftWidth = checkDimension(l, 3)
	case "border_right_width":
		cell.Value.BorderRightWidth = checkDimension(l, 3)
	case "border_top_width":
		cell.Value.BorderTopWidth = checkDimension(l, 3)
	case "border_bottom_width":
		cell.Value.BorderBottomWidth = checkDimension(l, 3)
	}

	return 0
}

func halignToString(h frontend.HorizontalAlignment) string {
	switch h {
	case frontend.HAlignLeft:
		return "left"
	case frontend.HAlignRight:
		return "right"
	case frontend.HAlignCenter:
		return "center"
	case frontend.HAlignJustified:
		return "justified"
	default:
		return "default"
	}
}

func valignToString(v frontend.VerticalAlignment) string {
	switch v {
	case frontend.VAlignTop:
		return "top"
	case frontend.VAlignMiddle:
		return "middle"
	case frontend.VAlignBottom:
		return "bottom"
	default:
		return "default"
	}
}

// registerTableMetaTable creates the Table metatable
func registerTableMetaTable(l *lua.State) {
	lua.NewMetaTable(l, tableMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: tableIndex},
		{Name: "__newindex", Function: tableNewIndex},
	}, 0)
	l.Pop(1)
}

// registerTableRowMetaTable creates the TableRow metatable
func registerTableRowMetaTable(l *lua.State) {
	lua.NewMetaTable(l, tableRowMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: rowIndex},
	}, 0)
	l.Pop(1)
}

// registerTableCellMetaTable creates the TableCell metatable
func registerTableCellMetaTable(l *lua.State) {
	lua.NewMetaTable(l, tableCellMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: cellIndex},
		{Name: "__newindex", Function: cellNewIndex},
	}, 0)
	l.Pop(1)
}
