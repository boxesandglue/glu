---@meta

-- LuaCATS type definitions for glu
-- These definitions provide IDE support (autocomplete, type checking)
-- for the glu Lua typesetting library.

--------------------------------------------------------------------------------
-- ScaledPoint
--------------------------------------------------------------------------------

---A type-safe dimension value representing scaled points (1/65536 of a DTP point).
---Supports arithmetic operations with other ScaledPoints, strings with units, and numbers.
---@class ScaledPoint
---@field pt number Value in DTP points (read-only)
---@field sp number Raw scaled point value (read-only)
---@operator add(ScaledPoint|string|number): ScaledPoint
---@operator sub(ScaledPoint|string|number): ScaledPoint
---@operator mul(number): ScaledPoint
---@operator div(number): ScaledPoint
---@operator div(ScaledPoint): number
---@operator unm: ScaledPoint
---@operator eq(ScaledPoint): boolean
---@operator lt(ScaledPoint): boolean
---@operator le(ScaledPoint): boolean
local ScaledPoint = {}

---Convert to DTP points
---@return number
function ScaledPoint:to_pt() end

---Convert to millimeters
---@return number
function ScaledPoint:to_mm() end

---Convert to centimeters
---@return number
function ScaledPoint:to_cm() end

---Convert to inches
---@return number
function ScaledPoint:to_in() end

--------------------------------------------------------------------------------
-- glu module
--------------------------------------------------------------------------------

---The glu base module provides scaled point operations and logging.
---@class glu
---@field factor integer Scaled points per DTP point (65535)
local glu = {}

---Create a ScaledPoint from a string with unit
---@param dimension string Dimension string like "12pt", "1cm", "10mm", "1in"
---@return ScaledPoint
function glu.sp(dimension) end

---Create a ScaledPoint from a number of DTP points
---@param points number Number of DTP points
---@return ScaledPoint
function glu.sp_from_pt(points) end

---Convert a ScaledPoint to DTP points
---@param sp ScaledPoint|string|number The dimension to convert
---@return number
function glu.sp_to_pt(sp) end

---Convert a ScaledPoint to a specific unit
---@param sp ScaledPoint|string|number The dimension to convert
---@param unit string Target unit: "pt", "mm", "cm", "in", "sp"
---@return number
function glu.sp_to_unit(sp, unit) end

---Return the maximum of two dimensions
---@param a ScaledPoint|string|number First dimension
---@param b ScaledPoint|string|number Second dimension
---@return ScaledPoint
function glu.max(a, b) end

---Return the minimum of two dimensions
---@param a ScaledPoint|string|number First dimension
---@param b ScaledPoint|string|number Second dimension
---@return ScaledPoint
function glu.min(a, b) end

---Log a debug message with optional key-value pairs
---@param message string The message to log
---@param ... any Key-value pairs
function glu.debug(message, ...) end

---Log an info message with optional key-value pairs
---@param message string The message to log
---@param ... any Key-value pairs
function glu.info(message, ...) end

---Log a warning message with optional key-value pairs
---@param message string The message to log
---@param ... any Key-value pairs
function glu.warn(message, ...) end

---Log an error message with optional key-value pairs
---@param message string The message to log
---@param ... any Key-value pairs
function glu.error(message, ...) end

return glu
