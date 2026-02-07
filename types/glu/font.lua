---@meta

-- LuaCATS type definitions for glu.font
-- Font loading and text shaping

--------------------------------------------------------------------------------
-- Types
--------------------------------------------------------------------------------

---A shaped atom representing a glyph or space
---@class Atom
---@field advance integer Advance width in scaled points
---@field height integer Glyph height in scaled points
---@field depth integer Glyph depth in scaled points
---@field codepoint integer Font-specific glyph ID
---@field components string Character(s) represented
---@field is_space boolean Is a space character
---@field hyphenate boolean Part of word
---@field kern_after integer Kerning after this glyph in scaled points

---A font instance with a specific size
---@class Font
---@field size integer Font size in scaled points
---@field space integer Space width in scaled points
---@field space_stretch integer Space stretchability in scaled points
---@field space_shrink integer Space shrinkability in scaled points
local Font = {}

---Shape text into atoms
---@param text string Text to shape
---@param ... string OpenType features (e.g., "+kern", "+liga", "-liga")
---@return Atom[] atoms Array of shaped atoms
function Font:shape(text, ...) end

--------------------------------------------------------------------------------
-- glu.font module
--------------------------------------------------------------------------------

---The glu.font module provides font instances and text shaping.
---@class glu.font
local font = {}

---Create a new font instance
---@param face Face Loaded font face (from glu.pdf)
---@param size ScaledPoint|integer Font size
---@return Font
function font.new(face, size) end

return font
