---
title: My Report
author: Your Name
lang: en
papersize: a4
css: style.css
---

# Introduction

A report-style document with title, author, and a custom stylesheet.
Page numbers come from the CSS `@bottom-center` margin box —
everything that's purely visual lives in `style.css`.

## Markdown features

- **Bold**, *italic*, `code`
- [Links](https://example.com)
- Inline Lua: pi is approximately {= string.format("%.4f", math.pi) =}.

| Column A | Column B |
| -------- | -------- |
| one      | two      |
| three    | four     |

## Next steps

* Edit `style.css` for visual tweaks.
* Run `glu index.md` (or `glu --watch index.md`) to rebuild.
* Add an `index.lua` next to this file to register lifecycle
  callbacks (page-init backgrounds, post-element decorations, …).
  See <https://boxesandglue.dev/glu/callbacks>.
