---@meta

-- LuaCATS type definitions for glu.node
-- Low-level node manipulation

--------------------------------------------------------------------------------
-- Node types
--------------------------------------------------------------------------------

---Base node type
---@class Node
---@field next Node? Next node in list
---@field prev Node? Previous node in list
---@field type string Node type name
---@field id integer Unique node ID

---A glyph node representing a character
---@class GlyphNode: Node
---@field codepoint integer Font-specific glyph ID
---@field components string Character composition
---@field width integer Width in scaled points
---@field height integer Height in scaled points
---@field depth integer Depth in scaled points
---@field yoffset integer Vertical offset in scaled points
---@field hyphenate boolean Part of hyphenatable word

---A glue node representing elastic space
---@class GlueNode: Node
---@field width integer Natural width in scaled points
---@field stretch integer Maximum stretch in scaled points
---@field shrink integer Maximum shrink in scaled points
---@field stretch_order integer Stretch infinity level (0-3)
---@field shrink_order integer Shrink infinity level (0-3)

---A kern node representing fixed space
---@class KernNode: Node
---@field kern integer Kern width in scaled points

---A penalty node indicating a break point
---@class PenaltyNode: Node
---@field penalty integer Break cost (-10000 to 10000)
---@field width integer Width if broken here, in scaled points

---A rule node drawing a rectangle
---@class RuleNode: Node
---@field width integer Rule width in scaled points
---@field height integer Rule height in scaled points
---@field depth integer Rule depth in scaled points
---@field pre string PDF code before rule
---@field post string PDF code after rule
---@field hide boolean Don't draw the rectangle

---A horizontal list node
---@class HListNode: Node
---@field width integer Total width in scaled points
---@field height integer Maximum height in scaled points
---@field depth integer Maximum depth in scaled points
---@field list Node? First node in list
---@field glue_set number Glue ratio (stretch/shrink)
---@field glue_sign integer 0=normal, 1=stretch, 2=shrink
---@field glue_order integer Infinity level
---@field shift integer Vertical shift in scaled points
---@field badness integer Line badness

---A vertical list node
---@class VListNode: Node
---@field width integer Total width in scaled points
---@field height integer Total height in scaled points
---@field depth integer Depth in scaled points
---@field list Node? First node in list
---@field glue_set number Glue ratio
---@field glue_sign integer 0=normal, 1=stretch, 2=shrink
---@field shift_x integer Horizontal shift in scaled points

---A discretionary break node for hyphenation
---@class DiscNode: Node
---@field pre Node? Nodes before break
---@field post Node? Nodes after break
---@field replace Node? Replacement if no break
---@field penalty integer Additional penalty

---An image node
---@class ImageNodeBackend: Node
---@field width integer Image width in scaled points
---@field height integer Image height in scaled points
---@field page integer Page number (for PDFs)
---@field used boolean Already output

---A language node
---@class LangNode: Node

---A start/stop node
---@class StartStopNode: Node
---@field action integer Action type

---Any node type
---@alias AnyNode GlyphNode|GlueNode|KernNode|PenaltyNode|RuleNode|HListNode|VListNode|DiscNode|ImageNodeBackend|LangNode|StartStopNode

--------------------------------------------------------------------------------
-- glu.node module
--------------------------------------------------------------------------------

---The glu.node module provides low-level node manipulation.
---@class glu.node
---@field normal integer Finite glue (0)
---@field fil integer First order infinity (1)
---@field fill integer Second order infinity (2)
---@field filll integer Third order infinity (3)
local node = {}

---Create a new node
---@param nodetype "glyph"|"glue"|"kern"|"disc"|"penalty"|"rule"|"hlist"|"vlist"|"image"|"lang"|"startstop"
---@return AnyNode
function node.new(nodetype) end

---Insert a node after another
---@param head Node? Head of list
---@param current Node Node to insert after
---@param new_node Node Node to insert
---@return Node head New head of list
function node.insert_after(head, current, new_node) end

---Insert a node before another
---@param head Node? Head of list
---@param current Node Node to insert before
---@param new_node Node Node to insert
---@return Node head New head of list
function node.insert_before(head, current, new_node) end

---Delete a node from a list
---@param head Node? Head of list
---@param node_to_remove Node Node to remove
---@return Node? head New head of list
function node.delete(head, node_to_remove) end

---Create a deep copy of a node list
---@param head Node? Head of list to copy
---@return Node? copy Copied list
function node.copy_list(head) end

---Get the last node in a list
---@param head Node? Head of list
---@return Node? tail Last node
function node.tail(head) end

---Pack nodes into a horizontal list (natural width)
---@param head Node? First node
---@return HListNode
function node.hpack(head) end

---Pack nodes into a horizontal list with specific width
---@param head Node? First node
---@param width ScaledPoint|integer Target width
---@return HListNode
function node.hpack_to(head, width) end

---Pack nodes into a vertical list
---@param head Node? First node
---@return VListNode
function node.vpack(head) end

---Get dimensions of a node list
---@param head Node? First node
---@return integer width Width in scaled points
---@return integer height Height in scaled points
---@return integer depth Depth in scaled points
function node.dimensions(head) end

---Get a string representation of a node list
---@param head Node? First node
---@return string
function node.string(head) end

return node
