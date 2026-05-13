---
title: Lua Pipeline
---

# Lua Pipeline

This document exercises the `{= expr =}` inline expression and a
fenced `{lua}` block. Two plus two is {= 2 + 2 =}; the square root
of 144 is {= math.sqrt(144) =}.

```{lua}
items = { "alpha", "beta", "gamma" }
return "<ul><li>" .. table.concat(items, "</li><li>") .. "</li></ul>"
```

End of document.
