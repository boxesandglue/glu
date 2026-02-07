---@meta

-- LuaCATS type definitions for glu.frontend
-- High-level typesetting API

---@alias Dimension ScaledPoint|string|number

--------------------------------------------------------------------------------
-- glu.frontend module
--------------------------------------------------------------------------------

---The glu.frontend module provides a high-level interface for creating PDF documents.
---@class glu.frontend
local frontend = {}

---Create a new PDF document
---@param filename string Output filename
---@return Document
function frontend.new(filename) end

---Create a new Text object
---@param settings? TextSettingsTable Optional initial settings
---@return Text
function frontend.text(settings) end

---@class TextSettingsTable
---@field font_family? FontFamily Font family
---@field fontfamily? FontFamily Font family (alias)
---@field font_size? Dimension Font size
---@field fontsize? Dimension Font size (alias)
---@field font_weight? string|integer Font weight: "regular", "bold", or 100-900
---@field fontweight? string|integer Font weight (alias)
---@field font_style? "normal"|"italic" Font style
---@field fontstyle? "normal"|"italic" Font style (alias)
---@field color? string|Color Text color
---@field leading? Dimension Line height
---@field halign? "left"|"right"|"center"|"justified" Horizontal alignment
---@field valign? "top"|"middle"|"bottom" Vertical alignment
---@field margin_left? Dimension Left margin
---@field margin_right? Dimension Right margin
---@field margin_top? Dimension Top margin
---@field margin_bottom? Dimension Bottom margin
---@field padding_left? Dimension Left padding
---@field padding_right? Dimension Right padding
---@field padding_top? Dimension Top padding
---@field padding_bottom? Dimension Bottom padding
---@field background_color? string|Color Background color
---@field hyperlink? string URL for hyperlink
---@field underline? boolean Underline text
---@field line_through? boolean Strikethrough text

---Create a new font source
---@param options FontSourceOptions
---@return FontSource
function frontend.fontsource(options) end

---Create a new Table object
---@param options? TableOptions
---@return Table
function frontend.table(options) end

---Create a new Color object
---@param options ColorOptions
---@return Color
function frontend.color(options) end

---Create a ScaledPoint from a number of DTP points
---@param points number Number of DTP points
---@return ScaledPoint
function frontend.sp(points) end

---Create a ScaledPoint from a string with unit
---@param dimension string Dimension string like "12pt", "1cm", "10mm", "1in"
---@return ScaledPoint
function frontend.sp_string(dimension) end

--------------------------------------------------------------------------------
-- Document
--------------------------------------------------------------------------------

---A PDF document
---@class Document
---@field title string Document title
---@field author string Document author
---@field subject string Document subject
---@field creator string Creator application
---@field keywords string Document keywords
---@field format string PDF format: "PDF", "PDF/A-3b", "PDF/X-3", "PDF/X-4", "PDF/UA"
---@field language Language Default document language (write-only)
---@field additional_xml_metadata string Additional XMP metadata
local Document = {}

---Create a new font family
---@param name string Font family name
---@return FontFamily
function Document:new_font_family(name) end

---Find an existing font family
---@param name string Font family name
---@return FontFamily?
function Document:find_font_family(name) end

---Create a new Text object
---@return Text
function Document:create_text() end

---Format text into a paragraph
---@param text Text Text to format
---@param width Dimension Paragraph width
---@param options? ParagraphOptions Optional settings
---@return VList vlist The formatted paragraph
---@return table info Information about the paragraph
function Document:format_paragraph(text, width, options) end

---Build a table into VLists
---@param tbl Table The table to build
---@return VList[] vlists Array of VLists (one per page)
function Document:build_table(tbl) end

---Define a named color
---@param name string Color name
---@param color Color Color object
function Document:define_color(name, color) end

---Get a color by name or CSS value
---@param spec string Color name or CSS value (e.g., "#ff0000", "rebeccapurple")
---@return Color
function Document:get_color(spec) end

---Get a language for hyphenation
---@param name string Language code (e.g., "en", "de", "fr")
---@return Language
function Document:get_language(name) end

---Create a new page
---@return Page
function Document:new_page() end

---Load an image or PDF file
---@param filename string Path to image or PDF
---@param page? integer Page number (for PDFs, default 1)
---@param box? string PDF box: "/MediaBox", "/CropBox", etc.
---@return Imagefile
function Document:load_imagefile(filename, page, box) end

---Create an image node for placement
---@param imagefile Imagefile Loaded image file
---@param page? integer Page number (for PDFs)
---@param box? string PDF box
---@return ImageNode
function Document:create_image_node(imagefile, page, box) end

---Load an ICC color profile
---@param filename string Path to ICC profile
---@return ColorProfile
function Document:load_colorprofile(filename) end

---Attach a file to the PDF
---@param options AttachmentOptions
function Document:attach_file(options) end

---Finalize the PDF document
function Document:finish() end

---@class ParagraphOptions
---@field leading? Dimension Line height
---@field font_size? Dimension Font size
---@field fontsize? Dimension Font size (alias)
---@field font_family? FontFamily Font family
---@field fontfamily? FontFamily Font family (alias)
---@field language? Language Language for hyphenation
---@field halign? "left"|"right"|"center"|"justified" Horizontal alignment
---@field indent_left? number Left indentation in points
---@field indent_left_rows? integer Number of rows to indent

---@class AttachmentOptions
---@field filename string Path to file
---@field name? string Visible name
---@field description? string File description
---@field mimetype? string MIME type

--------------------------------------------------------------------------------
-- Page
--------------------------------------------------------------------------------

---A PDF page
---@class Page
---@field width Dimension Page width (read: ScaledPoint, write: Dimension)
---@field height Dimension Page height (read: ScaledPoint, write: Dimension)
local Page = {}

---Place a VList at a position on the page
---@param x Dimension X coordinate (from left)
---@param y Dimension Y coordinate (from bottom)
---@param vlist VList The VList to place
---@return Page self
function Page:output_at(x, y, vlist) end

---Finalize and output the page
function Page:shipout() end

--------------------------------------------------------------------------------
-- Text
--------------------------------------------------------------------------------

---A text object for building styled content
---@class Text
---@field items table Content items (strings, Text, VList, Table)
---@field settings TextSettings Settings proxy
local Text = {}

---Append content items
---@param ... string|Text|VList|Table Items to append
---@return Text self
function Text:append(...) end

---Set a single setting
---@param key string Setting name
---@param value any Setting value
---@return Text self
function Text:set(key, value) end

---Apply multiple settings from a table
---@param tbl table<string, any> Settings table
---@return Text self
function Text:apply(tbl) end

---Text settings proxy
---@class TextSettings
---@field font_family FontFamily Font family
---@field fontfamily FontFamily Font family (alias)
---@field font_size Dimension Font size
---@field fontsize Dimension Font size (alias)
---@field font_weight string|integer Font weight: "regular", "bold", or 100-900
---@field fontweight string|integer Font weight (alias)
---@field font_style "normal"|"italic" Font style
---@field fontstyle "normal"|"italic" Font style (alias)
---@field color string|Color Text color
---@field leading Dimension Line height
---@field halign "left"|"right"|"center"|"justified" Horizontal alignment
---@field valign "top"|"middle"|"bottom" Vertical alignment
---@field margin_left Dimension Left margin
---@field margin_right Dimension Right margin
---@field margin_top Dimension Top margin
---@field margin_bottom Dimension Bottom margin
---@field padding_left Dimension Left padding
---@field padding_right Dimension Right padding
---@field padding_top Dimension Top padding
---@field padding_bottom Dimension Bottom padding
---@field background_color string|Color Background color
---@field hyperlink string URL for hyperlink
---@field underline boolean Underline text
---@field line_through boolean Strikethrough text

--------------------------------------------------------------------------------
-- FontFamily and FontSource
--------------------------------------------------------------------------------

---A font family containing multiple font faces
---@class FontFamily
local FontFamily = {}

---Add a font member to the family
---@param source FontSource|FontMemberOptions Font source or options table
---@param weight? string|integer Font weight (if not using options table)
---@param style? "normal"|"italic" Font style (if not using options table)
---@return FontFamily self
function FontFamily:add_member(source, weight, style) end

---@class FontMemberOptions
---@field source FontSource Font source
---@field weight integer|string Font weight: 100-900 or "regular", "bold"
---@field style "normal"|"italic" Font style

---@class FontSourceOptions
---@field location string Path to font file
---@field index? integer Font index (for collections)
---@field size_adjust? number Size adjustment factor
---@field features? string[] OpenType features (e.g., {"kern", "liga"})

---A font source (loaded font file)
---@class FontSource

---A font face instance
---@class Face

--------------------------------------------------------------------------------
-- VList
--------------------------------------------------------------------------------

---A vertical list (formatted content ready for placement)
---@class VList
---@field width ScaledPoint Width of the VList
---@field height ScaledPoint Height of the VList
---@field depth ScaledPoint Depth of the VList

--------------------------------------------------------------------------------
-- Table
--------------------------------------------------------------------------------

---@class TableOptions
---@field max_width? Dimension Maximum table width
---@field stretch? boolean Stretch to fill width
---@field font_family? FontFamily Default font family
---@field font_size? Dimension Default font size
---@field leading? Dimension Default line height

---A table for building tabular content
---@class Table
---@field max_width Dimension Maximum table width
---@field stretch boolean Stretch to fill width
local Table = {}

---Set column widths
---@param widths Dimension[] Array of column widths
---@return Table self
function Table:set_columns(widths) end

---Add a new row
---@return TableRow
function Table:add_row() end

---Create a new row (alias for add_row)
---@return TableRow
function Table:new_row() end

---A table row
---@class TableRow
local TableRow = {}

---Add a new cell
---@return TableCell
function TableRow:add_cell() end

---Create a new cell (alias for add_cell)
---@return TableCell
function TableRow:new_cell() end

---A table cell
---@class TableCell
---@field halign "left"|"right"|"center" Horizontal alignment
---@field valign "top"|"middle"|"bottom" Vertical alignment
---@field colspan integer Column span
---@field rowspan integer Row span
---@field padding_left Dimension Left padding
---@field padding_right Dimension Right padding
---@field padding_top Dimension Top padding
---@field padding_bottom Dimension Bottom padding
---@field border_left_width Dimension Left border width
---@field border_right_width Dimension Right border width
---@field border_top_width Dimension Top border width
---@field border_bottom_width Dimension Bottom border width
local TableCell = {}

---Set cell contents
---@param ... string|Text Items to add
---@return TableCell self
function TableCell:set_contents(...) end

---Append content to cell
---@param item string|Text Item to append
function TableCell:append(item) end

--------------------------------------------------------------------------------
-- Color
--------------------------------------------------------------------------------

---@class ColorOptions
---@field model "rgb"|"cmyk"|"gray" Color model
---@field r? number Red (0-1, for RGB)
---@field g? number Green (0-1, for RGB) or Gray (0-1, for gray)
---@field b? number Blue (0-1, for RGB)
---@field c? number Cyan (0-1, for CMYK)
---@field m? number Magenta (0-1, for CMYK)
---@field y? number Yellow (0-1, for CMYK)
---@field k? number Black (0-1, for CMYK)

---A color value
---@class Color

--------------------------------------------------------------------------------
-- Language
--------------------------------------------------------------------------------

---A language for hyphenation
---@class Language

--------------------------------------------------------------------------------
-- Imagefile and ImageNode
--------------------------------------------------------------------------------

---A loaded image or PDF file
---@class Imagefile
---@field filename string Original filename
---@field format string Image format
---@field width integer Image width in pixels
---@field height integer Image height in pixels
---@field scale_x number X scale factor
---@field scale_y number Y scale factor
---@field number_of_pages integer Number of pages (for PDFs)
---@field internal_name string Internal PDF name

---An image node ready for placement
---@class ImageNode
---@field width ScaledPoint Image width
---@field height ScaledPoint Image height

--------------------------------------------------------------------------------
-- ColorProfile
--------------------------------------------------------------------------------

---An ICC color profile
---@class ColorProfile
---@field identifier string Profile identifier
---@field registry string Registry name
---@field info string Profile description
---@field condition string Color condition (e.g., "RGB")
---@field colors integer Number of color components

return frontend
