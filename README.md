[![Explore in Constellation](https://img.shields.io/badge/Explore%20in-Constellation-blue)](https://constellation.speedata.de) [![Homepage](https://img.shields.io/badge/homepage-boxesandglue.dev-blue)](https://boxesandglue.dev/glu)

# glu


**Markdown to PDF. With typographic precision.**

glu turns Markdown into beautifully typeset PDFs — with real line breaking, hyphenation, font shaping, and all the things you'd expect from a proper typesetting engine. Under the hood it uses [boxes and glue](https://github.com/boxesandglue/boxesandglue), a Go library that implements TeX's algorithms.

```bash
glu document.md        # → document.pdf
```


## What makes glu different?

- **Markdown first** — Write Markdown, get a PDF with proper typography. No LaTeX, no XML.
- **CSS for styling** — Page size, fonts, margins, borders, background colors — all via CSS.
- **Lua for logic** — Embed Lua blocks for computed content, table of contents, dynamic data.
- **Callbacks** — Hook into page creation for headers, footers, decorative frames, watermarks.
- **TeX quality** — Knuth-Plass line breaking, optical margin alignment, microtypography.

## Quick example

````markdown
---
title: The Frog King
css: custom.css
---

# Once upon a time

In olden times when wishing still helped one,
there lived a king whose daughters were all beautiful.

```{lua}
return string.format("5! = %d", factorial(5))
```
````

```bash
glu story.md    # → story.pdf
```

## More

glu also supports raw HTML mode, low-level Lua scripting for full control, and direct access to the typesetting engine's node lists.

Full documentation: **[boxesandglue.dev/glu](https://boxesandglue.dev/glu)**

## Installation

```bash
rake build      # creates bin/glu
```

## Ecosystem

glu is part of a broader ecosystem of PDF, typesetting and publishing technologies.
**[Explore the constellation →](https://constellation.speedata.de)**

## License

MIT
