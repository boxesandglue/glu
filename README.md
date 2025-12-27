# glu

Lua typesetting with [boxes and glue](https://github.com/boxesandglue/boxesandglue).

glu is a command-line tool that executes Lua scripts for PDF typesetting. It provides Lua bindings for the boxesandglue typesetting library and the baseline-pdf low-level PDF writer.

## Installation

```bash
rake build
```

This creates the `bin/glu` binary.

## Usage

```bash
glu [options] <filename.lua>
```

### Commands

- `glu help` – Show help message
- `glu version` – Print version

## Quick Start

```lua
-- hello.lua
local doc = frontend.new("hello.pdf")

-- Load a font
local ff = doc:new_font_family("text")
local fs = frontend.fontsource({ location = "path/to/font.ttf" })
ff:add_member(fs, "regular", "normal")

-- Create text
local txt = frontend.text()
txt:settings({
    font_family = ff,
    font_size = "12pt",
    color = "black"
})
txt:append("Hello, World!")

-- Format and output
local vlist = doc:format_paragraph(txt, "15cm")

local page = doc:new_page()
page.width = "21cm"
page.height = "29.7cm"
page:output_at("2cm", "27cm", vlist)
page:shipout()

doc:finish()
```

Run with:
```bash
glu hello.lua
```

## Modules

glu provides the following Lua modules:

| Module     | Description                                    |
| ---------- | ---------------------------------------------- |
| `frontend` | High-level typesetting API                     |
| `pdf`      | Low-level PDF writing                          |
| `bag`      | Scaled points, unit conversion, logging        |
| `node`     | Node types and list operations                 |
| `font`     | Font instances and text shaping                |

### frontend

High-level typesetting API wrapping boxesandglue.

#### Document

```lua
local doc = frontend.new("output.pdf")

doc:new_font_family(name)      -- Create font family
doc:find_font_family(name)     -- Find existing font family
doc:create_text()              -- Create Text object
doc:format_paragraph(text, width, [options])  -- Format paragraph → VList
doc:build_table(table)         -- Build table → VList array
doc:define_color(name, color)  -- Define named color
doc:get_color(spec)            -- Get color by name or CSS
doc:get_language(name)         -- Get language for hyphenation
doc:new_page()                 -- Create new page
doc:finish()                   -- Finalize PDF
```

#### Text

```lua
local txt = frontend.text()

txt:append(item, ...)          -- Append string, Text, VList, or Table
txt:set(key, value)            -- Set single setting
txt:settings({ ... })          -- Set multiple settings
```

**Text settings:**
- `font_family` – FontFamily object
- `font_size` / `size` – e.g. `12` or `"12pt"`
- `font_weight` – `"regular"`, `"bold"`, or number (100-900)
- `font_style` – `"normal"`, `"italic"`
- `color` – Color name or CSS value
- `leading` – Line height
- `halign` / `align` – `"left"`, `"right"`, `"center"`, `"justified"`
- `margin_left`, `margin_right`, `margin_top`, `margin_bottom`
- `padding_left`, `padding_right`, `padding_top`, `padding_bottom`
- `background_color`
- `hyperlink` – URL string
- `underline`, `line_through` – boolean

#### FontFamily

```lua
local ff = doc:new_font_family("name")
local fs = frontend.fontsource({
    location = "path/to/font.ttf",
    index = 0,           -- optional, for font collections
    size_adjust = 1.0    -- optional
})
ff:add_member(fs, weight, style)
-- weight: "regular", "bold", "100"-"900"
-- style: "normal", "italic"
```

#### Page

```lua
local page = doc:new_page()

page.width = "21cm"            -- A4 width
page.height = "29.7cm"         -- A4 height
page:output_at(x, y, vlist)    -- Place VList at position
page:shipout()                 -- Finalize page
```

#### Table

```lua
local tbl = frontend.table({
    max_width = "15cm",
    stretch = true,
    font_family = ff,
    font_size = "10pt",
    leading = "12pt"
})

tbl:set_columns({ "5cm", "10cm" })  -- Column widths

local row = tbl:add_row()
local cell = row:add_cell()
cell:set_contents("Cell text")
cell.halign = "center"
cell.valign = "middle"
cell.colspan = 2
cell.padding_left = "2mm"

local vlists = doc:build_table(tbl)
```

#### Color

```lua
local col = frontend.color({
    model = "rgb",           -- "rgb", "cmyk", "gray"
    r = 1.0, g = 0.0, b = 0.0
})

doc:define_color("red", col)
local red = doc:get_color("red")
local blue = doc:get_color("#0000ff")
```

#### Language

```lua
local lang = doc:get_language("en")  -- English hyphenation
local lang = doc:get_language("de")  -- German hyphenation
```

### pdf

Low-level PDF API wrapping baseline-pdf.

```lua
local pw = pdf.new("output.pdf")

pw.default_page_width = 595    -- in points
pw.default_page_height = 842

local face = pw:load_face("font.ttf", 0)
local img = pw:load_image("image.png")

local stream = pw:new_object()
stream.force_stream = true
stream:write("BT /F1 12 Tf 100 700 Td (Hello) Tj ET")

local page = pw:add_page(stream)
page.width = 595
page.height = 842
page.faces = { face }

pw:finish()
```

### bag

Scaled point operations and logging.

```lua
-- Unit conversion
local sp = bag.sp("12pt")       -- String → ScaledPoint
local sp = bag.sp("1cm")
local sp = bag.sp_from_pt(12)   -- Points → ScaledPoint
local pt = bag.sp_to_pt(sp)     -- ScaledPoint → Points

-- Math
local max = bag.max(sp1, sp2)
local min = bag.min(sp1, sp2)

-- Logging
bag.info("Message", "key", value)
bag.debug/warn/error(...)
```

### node

Low-level node types and list operations.

```lua
-- Create nodes
local glyph = node.new("glyph")
local glue = node.new("glue")
local kern = node.new("kern")
local hlist = node.new("hlist")
local vlist = node.new("vlist")
-- Also: disc, penalty, rule, image, lang, startstop

-- Set attributes
glyph.codepoint = 65
glyph.width = bag.sp("10pt")
glue.width = bag.sp("12pt")
glue.stretch = bag.sp("6pt")

-- Link nodes
glyph.next = glue

-- List operations
node.insert_after(head, cur, new)
node.insert_before(head, cur, new)
node.delete(head, cur)
node.copy_list(head)
node.tail(head)

-- Packing
local hlist = node.hpack(head)
local hlist = node.hpack_to(head, width)
local vlist = node.vpack(head)

-- Dimensions
local w, h, d = node.dimensions(head)
```

### font

Font instances and text shaping.

```lua
-- Create font from face
local pw = pdf.new("out.pdf")
local face = pw:load_face("font.ttf")
local fnt = font.new(face, bag.sp("12pt"))

-- Shape text into atoms
local atoms = fnt:shape("Hello", "+liga", "+kern")

for _, atom in ipairs(atoms) do
    print(atom.components, atom.advance)
end
```

## Dimensions

All dimension parameters accept:

- **Numbers** – interpreted as points: `12`, `595.28`
- **Strings with units** – `"12pt"`, `"1cm"`, `"10mm"`, `"1in"`

Examples:
```lua
page.width = "21cm"
page:output_at("2cm", "27cm", vlist)
txt:set("size", "14pt")
doc:format_paragraph(txt, 400)  -- 400 points
```

## License

MIT
