glu(1)

# NAME

glu - typesetting with boxes and glue

# SYNOPSIS

*glu* [_options_] _FILE_++
*glu* [_options_] *--input-format* _FMT_ *-o* _PDF_ *-*++
*glu* *doctor*

# DESCRIPTION

*glu* turns Markdown, HTML, or a Lua script into a PDF using the
boxesandglue typesetting library (a Go port of the TeX line breaking,
hyphenation, and font shaping algorithms). The CLI dispatches on the
input file's extension: *.md*, *.html*/*.htm*, or *.lua*.

When invoked under a non-glu name via a symlink, *glu* runs the
adjacent _<basename>.lua_ script with all positional arguments
forwarded — see *MULTI-CALL BINARY*.

# OPTIONS

*-o*, *--output* _FILE_
	Write the PDF to _FILE_. The log file follows the same stem (so
	*-o /tmp/out.pdf* puts the log at */tmp/out.log* and the aux file
	at */tmp/out-aux.json*). Defaults to _<input>.pdf_ next to the input.

*--max-passes* _N_
	Maximum aux convergence passes (default *3*). Set to *1* to disable
	auto-rerun. Oscillating aux content aborts with exit code 6 even if
	*N* hasn't been reached.

*--safe*
	Run Lua in a hardened sandbox: no *io*, *os*, *debug*, *load*,
	*dofile*, *loadfile*, no file-system searchers in *require()*.
	Defence-in-depth, not hermetic — see CAVEAT in safelua.go source.

*--log-file* _FILE_
	Override the log file path. Pass *-* to disable the log file
	entirely (console-only logging).

*--log-format* _text_|_json_
	Log output format. JSON emits one record per line (slog default
	shape), useful for CI ingestion. Default *text*.

*--loglevel* _LVL_
	*debug*, *info* (default), *warn*, or *error*.

*-q*, *--quiet*
	Suppress console output. Log file is still written unless
	*--log-file -* is also passed.

*--input-format* _md_|_html_|_lua_
	Required when reading from stdin (input *-*). Picks the pipeline
	used for the stdin payload; companion .lua files are *not* loaded
	for stdin input.

*--manifest* _FILE_
	Write a JSON sidecar with pages, passes, duration_ms, generated_at,
	and the list of headings. Schema is versioned (*schema_version*
	field) so consumers can pin.

*--template*
	Apply Go *text/template* expansion to the input before front-matter
	extraction. Markdown mode only.

*--css* _FILE_
	Additional CSS file loaded after defaultCSS and any front-matter
	CSS reference.

*--clean*
	Remove _<output>-aux.json_ before the first pass. Useful when an
	aux file from an earlier crashed run is poisoning the rerun loop.

*--markdown*
	Print the body after Lua-block expansion to stdout and exit (no
	PDF). Debug aid; Markdown mode only.

*--html*
	Print the goldmark-rendered HTML to stdout and exit (no PDF).
	Debug aid; Markdown mode only.

*--cpuprofile* _FILE_
	Write a Go pprof CPU profile to _FILE_.

*-v*, *--version*
	Print version and exit.

*-h*, *--help*
	Print help and exit.

# SUBCOMMANDS

*doctor*
	Print environment self-checks (build info, font directories,
	hyphenation patterns, optional external tools). Exit code is the
	number of FAILures; WARNings do not affect exit status.

*version*
	Same as *--version*.

*help*
	Same as *--help*.

# FRONTMATTER

Markdown mode reads YAML front matter between two *---* lines:

[[ *key*
:[ *meaning*
|  title
:  document /Title
|  author
:  document /Author
|  css
:  extra CSS file path
|  papersize
:  e.g. *a4*, *letter*
|  format
:  *PDF/UA* enables tagged-PDF + StructTreeRoot pipeline
|  lang
:  BCP47 tag — /Lang + hyphenation default
|  highlight-style
:  chroma highlight name (default *github*)

# COMPANION SCRIPTS

For input _foo.md_ or _foo.html_, *glu* loads _foo.lua_ before running
the pipeline. The companion script typically registers lifecycle
callbacks via *glu.frontend.add_callback*. Stdin inputs have no
companion script (the temp filename never matches).

# AUX FILE AND AUTO-RERUN

Forward references (page counts, headings, custom Lua state in the
*_aux* global) are persisted to _<output>-aux.json_. Each run:

. reads the previous run's aux into _aux*/*_toc;
. runs the pipeline;
. writes the new aux back.

If the written aux differs from what was read, *glu* re-runs the
pipeline with a fresh Lua state. Up to *--max-passes* iterations are
attempted. Cycle detection: if aux content (SHA prefix) repeats,
abort early with exit code 6.

# LIFECYCLE CALLBACKS

Register via *glu.frontend.add_callback(name, fn)* in a companion or
.lua entry script:

[[ *event*
:[ *fires*
|  document_start
:  after companion .lua loads, before aux is read
|  content_ready
:  after _aux/_toc set, before {lua} blocks run
|  page_init
:  each new page after @page resolution
|  pre_shipout
:  each page right before PDF shipout
|  post_element
:  after each emitted block/inline element
|  document_end
:  after fe.Finish; aux is then read back from Lua

# MULTI-CALL BINARY

When *glu*'s basename is not literally "glu" (typically reached via a
symlink), *glu* searches for _<basename>.lua_ in:

. the current working directory;
. the directory of the symlink (*os.Args[0]*).

The first match is run with all positional arguments forwarded.
Flags that *glu* owns are still parsed first.

# EXIT CODES

[[ *code*
:[ *meaning*
|  0
:  success
|  1
:  generic / unknown error
|  2
:  usage (bad flags, missing arguments)
|  3
:  io (file not found, permission denied)
|  4
:  lua (syntax or runtime in {lua} or companion)
|  5
:  typesetting (boxesandglue / htmlbag failure)
|  6
:  aux file did not converge after --max-passes

# FILES

_<input>.lua_
	Companion Lua script auto-loaded for Markdown/HTML inputs.

_<output>-aux.json_
	Persisted aux/_toc state for forward-reference convergence.

_<output>.log_
	Transcript of the run. Override with *--log-file* or disable with
	*--log-file -*.

# ENVIRONMENT

*GOMAXPROCS*, *GOGC*
	Standard Go runtime knobs. boxesandglue is largely single-threaded
	for now; tuning *GOGC* (default 100) up may reduce GC pressure on
	large documents.

# SEE ALSO

*boxesandglue.dev/glu* for the full documentation and the Lua API
reference.

# AUTHORS

Patrick Gundlach and contributors. *glu* is part of the boxesandglue
ecosystem (boxesandglue.dev).
