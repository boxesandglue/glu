# LuaCATS Type Definitions for glu

This directory contains [LuaCATS](https://luals.github.io/wiki/annotations/) type definitions for the glu Lua typesetting library. These definitions provide IDE support including:

- Autocomplete
- Type checking
- Hover documentation
- Go to definition

## Setup

### VS Code with Lua Language Server

1. Install the [Lua extension](https://marketplace.visualstudio.com/items?itemName=sumneko.lua) by sumneko

2. Add this to your `.vscode/settings.json`:

```json
{
    "Lua.workspace.library": [
        "/path/to/glu/types"
    ],
    "Lua.runtime.version": "Lua 5.4"
}
```

Or create a `.luarc.json` in your project root:

```json
{
    "$schema": "https://raw.githubusercontent.com/sumneko/vscode-lua/master/setting/schema.json",
    "workspace.library": [
        "/path/to/glu/types"
    ],
    "runtime.version": "Lua 5.4"
}
```

### Neovim with nvim-lspconfig

```lua
require('lspconfig').lua_ls.setup {
    settings = {
        Lua = {
            workspace = {
                library = {
                    "/path/to/glu/types"
                }
            }
        }
    }
}
```

## Files

| File | Module | Description |
|------|--------|-------------|
| `glu.lua` | `glu` | Base module: ScaledPoint, logging |
| `glu/frontend.lua` | `glu.frontend` | High-level typesetting API |
| `glu/node.lua` | `glu.node` | Low-level node manipulation |
| `glu/font.lua` | `glu.font` | Font loading and shaping |
| `glu/pdf.lua` | `glu.pdf` | Low-level PDF writing |
| `xml/cxpath.lua` | `xml.cxpath` | XPath XML querying |

## Example

With the type definitions configured, you get full IDE support:

```lua
local frontend = require("glu.frontend")
local glu = require("glu")

-- Autocomplete and type checking work
local doc = frontend.new("output.pdf")  -- Returns Document
local sp = glu.sp("12pt")               -- Returns ScaledPoint

-- Hover shows documentation
local page = doc:new_page()
page.width = frontend.sp_string("210mm")
page.height = frontend.sp_string("297mm")

-- Arithmetic operations are typed
local margin = glu.sp("2cm")
local y = glu.sp("28cm") - margin  -- ScaledPoint

-- Error detection
page:output_at(margin, y, vlist)
```
