---@meta

-- LuaCATS type definitions for glu.pdf
-- Low-level PDF writing

--------------------------------------------------------------------------------
-- Types
--------------------------------------------------------------------------------

---A PDF writer instance
---@class PDFWriter
local PDFWriter = {}

---Load a font face
---@param filename string Path to font file
---@param index? integer Font index (for collections)
---@return Face
function PDFWriter:load_face(filename, index) end

---Load an image file
---@param filename string Path to image file
---@return PDFImagefile
function PDFWriter:load_imagefile(filename) end

---Create a new page
---@return PDFPage
function PDFWriter:new_page() end

---Finish and close the PDF
function PDFWriter:finish() end

---A loaded font face
---@class Face

---A loaded image file (low-level)
---@class PDFImagefile
---@field filename string Original filename
---@field format string Image format
---@field width integer Image width in pixels
---@field height integer Image height in pixels

---A PDF page (low-level)
---@class PDFPage
---@field width integer Page width in scaled points
---@field height integer Page height in scaled points
local PDFPage = {}

---Output a VList at a position
---@param x integer X coordinate in scaled points
---@param y integer Y coordinate in scaled points
---@param vlist VListNode The VList to output
function PDFPage:output_at(x, y, vlist) end

---Finalize the page
function PDFPage:shipout() end

--------------------------------------------------------------------------------
-- glu.pdf module
--------------------------------------------------------------------------------

---The glu.pdf module provides low-level PDF writing capabilities.
---@class glu.pdf
local pdf = {}

---Create a new PDF writer
---@param filename string Output filename
---@return PDFWriter
function pdf.new(filename) end

return pdf
