---@meta

-- LuaCATS type definitions for xml.cxpath
-- XPath XML querying

--------------------------------------------------------------------------------
-- Types
--------------------------------------------------------------------------------

---An XML element node
---@class XMLElement
---@field name string Element name (local name)
---@field prefix string Namespace prefix
---@field namespace string Namespace URI
---@field attributes table<string, string> Element attributes
---@field children XMLNode[] Child nodes
local XMLElement = {}

---Get attribute value
---@param name string Attribute name
---@return string?
function XMLElement:get_attribute(name) end

---Get text content
---@return string
function XMLElement:text() end

---An XML text node
---@class XMLText
---@field content string Text content

---An XML comment node
---@class XMLComment
---@field content string Comment content

---Any XML node type
---@alias XMLNode XMLElement|XMLText|XMLComment

---An XPath context for evaluating expressions
---@class XPathContext
local XPathContext = {}

---Evaluate an XPath expression
---@param xpath string XPath expression
---@return XMLNode[] nodes Matching nodes
function XPathContext:query(xpath) end

---Evaluate an XPath expression and return first match
---@param xpath string XPath expression
---@return XMLNode? node First matching node
function XPathContext:query_first(xpath) end

---Evaluate an XPath expression and return string value
---@param xpath string XPath expression
---@return string
function XPathContext:string(xpath) end

---Evaluate an XPath expression and return number value
---@param xpath string XPath expression
---@return number
function XPathContext:number(xpath) end

---Evaluate an XPath expression and return boolean value
---@param xpath string XPath expression
---@return boolean
function XPathContext:boolean(xpath) end

--------------------------------------------------------------------------------
-- xml.cxpath module
--------------------------------------------------------------------------------

---The xml.cxpath module provides XPath XML querying.
---@class xml.cxpath
local cxpath = {}

---Parse XML from a string
---@param xml string XML content
---@return XMLElement root Root element
function cxpath.parse(xml) end

---Parse XML from a file
---@param filename string Path to XML file
---@return XMLElement root Root element
function cxpath.parse_file(filename) end

---Create an XPath context from an element
---@param element XMLElement Root element
---@return XPathContext
function cxpath.context(element) end

---Register a namespace for XPath queries
---@param ctx XPathContext XPath context
---@param prefix string Namespace prefix
---@param uri string Namespace URI
function cxpath.register_namespace(ctx, prefix, uri) end

return cxpath
