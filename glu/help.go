package main

// helpCoda returns the long-form help text shown after the flag list
// by `glu --help`. Kept in a separate file so the option-parser
// boilerplate in main stays readable.
func helpCoda() string {
	return `
Usage: glu [options] <filename.lua|filename.md|filename.html>
       glu [options] --input-format md -o out.pdf -   (read from stdin)
       glu doctor                                     (environment checks)
       glu completion bash|zsh|fish                   (shell completion)

EXAMPLES

  glu story.md
      Render story.md to story.pdf next to the input. Companion
      story.lua (if present) is loaded for callbacks.

  glu -o /tmp/build.pdf --max-passes 5 story.md
      Write the PDF to /tmp/build.pdf and the log to /tmp/build.log.
      Allow up to 5 aux iterations for forward-reference convergence.

  glu --safe untrusted.md
      Render with the Lua sandbox enabled. {lua} blocks cannot reach
      os, io, debug, or load arbitrary files.

  cat story.md | glu --input-format md -o out.pdf -
      Read Markdown from stdin.

  glu --log-format json --log-file /var/log/glu.log story.md
      Write structured JSON logs for CI consumption.

  glu --manifest build.json story.md
      Write a JSON sidecar with pages, passes, duration, headings.

  glu --watch story.md
      Rebuild story.pdf whenever story.md, story.lua (if present), the
      --css file, or the frontmatter css: file changes. The list is
      computed once at watch start — editing 'css: a.css' to 'css:
      b.css' inside the frontmatter won't pick up b.css until next
      restart. Ctrl-C exits.

  glu completion zsh > ~/.zfunc/_glu
      Generate shell completion. Supported: bash, zsh, fish.
      For bash:  glu completion bash > /etc/bash_completion.d/glu
      For fish:  glu completion fish > ~/.config/fish/completions/glu.fish

FRONTMATTER (Markdown mode, YAML between '---' lines)

  title:           document /Title
  author:          document /Author
  css:             extra CSS file path (relative to cwd)
  papersize:       e.g. 'a4', 'letter'
  format:          'PDF/UA' enables the accessibility pipeline
  lang:            BCP47 tag (drives /Lang and hyphenation defaults)
  highlight-style: chroma highlight name (default 'github')

AUX FILE

  Forward references (counter(pages), TOC entries, custom Lua state in
  the _aux global) live in <output>-aux.json. glu auto-reruns up to
  --max-passes times to converge. Oscillating runs (aux flips between
  two states) abort with exit code 6.

EXIT CODES

  0  success
  1  generic / unknown error
  2  usage error (bad flags, missing arguments)
  3  io error (file not found, permission denied)
  4  lua error (syntax or runtime in a {lua} block or companion)
  5  typesetting error (boxesandglue or htmlbag failure)
  6  aux file did not converge after --max-passes

LIFECYCLE CALLBACKS (register from companion .lua via glu.frontend.add_callback)

  document_start    after companion .lua loads, before aux is read
  content_ready     after _aux/_toc are set, before {lua} blocks run
  page_init         each new page (post @page resolution)
  pre_shipout       each page about to be shipped to the PDF
  post_element      after each emitted block/inline element
  document_end      after fe.Finish; aux is then read back from Lua

See 'glu doctor' for environment diagnostics.
`
}
