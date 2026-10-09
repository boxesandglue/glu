---@meta

-- LuaCATS type definitions for xml.cxpath
-- XPath XML querying

--------------------------------------------------------------------------------
-- Types
--------------------------------------------------------------------------------

---An XPath context: the document, or the result of an evaluation (a node,
---a sequence of nodes or an atomic value). Every query runs relative to it.
---@class XPathContext
---@field string string String value of the context (read-only)
local XPathContext = {}

---Register a namespace prefix for XPath expressions. The prefix `xml` is
---predeclared.
---@param prefix string Namespace prefix
---@param uri string Namespace URI
---@return XPathContext self The same context, for chaining
function XPathContext:set_namespace(prefix, uri) end

---Evaluate an XPath expression relative to this context. Raises an error
---on an invalid expression.
---@param xpath string XPath expression
---@return XPathContext result Context holding the result
function XPathContext:eval(xpath) end

---Iterate over the items an XPath expression selects, in document order.
---@param xpath string XPath expression
---@return fun(): XPathContext? iterator Yields one context per item
function XPathContext:each(xpath) end

---Get the root element of the document.
---@return XPathContext root
function XPathContext:root() end

---Get the value of this context as an integer.
---@return integer
function XPathContext:int() end

---Get the value of this context as a boolean.
---@return boolean
function XPathContext:bool() end

--------------------------------------------------------------------------------
-- xml.cxpath module
--------------------------------------------------------------------------------

---The xml.cxpath module provides XPath XML querying.
---@class xml.cxpath
local cxpath = {}

---Open and parse an XML file. Raises an error if the file cannot be read or
---parsed.
---@param filename string Path to XML file
---@return XPathContext doc Context of the document
function cxpath.open(filename) end

return cxpath
