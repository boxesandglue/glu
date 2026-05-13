package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/boxesandglue/glu/internal/errkind"
	"github.com/boxesandglue/glu/markdown"
)

// debounceInterval collapses bursts of filesystem events (editor
// writes typically fire create+rename+write back-to-back) into a
// single rebuild. 200 ms is the practical sweet spot.
const debounceInterval = 200 * time.Millisecond

// watchPaths is the set of files --watch monitors: input, the
// companion <stem>.lua (if it exists), any --css argument, and the
// `css:` field from a Markdown file's frontmatter (parsed once
// before the watch loop starts). The list is computed once and not
// updated mid-session — if the user changes `css: a.css` to
// `css: b.css` in the frontmatter, b.css won't be watched until
// the next glu invocation.
func watchPaths(input, cssFlag string) []string {
	paths := []string{input}
	companion := companionLuaPath(input)
	if companion != "" {
		if _, err := os.Stat(companion); err == nil {
			paths = append(paths, companion)
		}
	}
	if cssFlag != "" {
		if _, err := os.Stat(cssFlag); err == nil {
			paths = append(paths, cssFlag)
		}
	}
	if css := frontmatterCSS(input); css != "" {
		paths = append(paths, css)
	}
	return paths
}

// frontmatterCSS reads the input file (Markdown only) and returns
// the resolved path to the `css:` frontmatter entry, or "" if the
// input is not Markdown, has no frontmatter, the entry is missing,
// or the referenced file does not exist on disk. The path is
// resolved by trying cwd-relative first (matching cb.ReadCSSFile's
// behaviour in the Markdown render path) and falling back to
// input-directory-relative.
func frontmatterCSS(input string) string {
	if ext := filepath.Ext(input); ext != ".md" {
		return ""
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return ""
	}
	fm, _ := markdown.ExtractFrontmatter(string(data))
	if fm.CSS == "" {
		return ""
	}
	if filepath.IsAbs(fm.CSS) {
		if _, err := os.Stat(fm.CSS); err == nil {
			return fm.CSS
		}
		return ""
	}
	if _, err := os.Stat(fm.CSS); err == nil {
		return fm.CSS
	}
	rel := filepath.Join(filepath.Dir(input), fm.CSS)
	if _, err := os.Stat(rel); err == nil {
		return rel
	}
	return ""
}

// companionLuaPath returns <stem>.lua for .md/.html inputs, or ""
// for .lua inputs (the input itself is already in the watch list).
func companionLuaPath(input string) string {
	ext := filepath.Ext(input)
	switch ext {
	case ".md", ".html", ".htm":
		return input[0:len(input)-len(ext)] + ".lua"
	}
	return ""
}

// runWatch watches paths and re-invokes build on any write. Errors
// from build are logged but never returned — the watcher keeps
// running until ctx is cancelled (SIGINT / SIGTERM). The first
// build runs synchronously before the watch loop starts.
//
// We watch parent directories rather than the files themselves
// because most editors save via temp-file + rename, and fsnotify on
// macOS does not always surface the post-rename event when the
// watch target was the old inode. Filtering by basename inside the
// event handler restores the per-file semantics.
func runWatch(ctx context.Context, paths []string, build func() error) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("%w: creating watcher: %s", errkind.IO, err.Error())
	}
	defer w.Close()

	wanted := map[string]bool{}
	watchedDirs := map[string]bool{}
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("%w: resolving %s: %s", errkind.IO, p, err.Error())
		}
		wanted[abs] = true
		dir := filepath.Dir(abs)
		if !watchedDirs[dir] {
			if err := w.Add(dir); err != nil {
				return fmt.Errorf("%w: watching %s: %s", errkind.IO, dir, err.Error())
			}
			watchedDirs[dir] = true
		}
	}

	slog.Info("Watch mode", "files", paths, "debounce_ms", debounceInterval.Milliseconds())
	if err := build(); err != nil {
		slog.Error("Initial build failed", "err", err.Error())
	} else {
		slog.Info("Initial build OK, watching for changes")
	}

	// debounce + coalesce-while-building: any event during a build
	// schedules exactly one more build after the current one
	// finishes. Without that, fast save bursts queue up and you
	// rebuild N times when the user only meant to rebuild once.
	var timer *time.Timer
	var pending bool
	building := false

	var rebuild func()
	rebuild = func() {
		building = true
		slog.Info("Rebuilding")
		if err := build(); err != nil {
			slog.Error("Build failed", "err", err.Error())
		} else {
			slog.Info("Rebuild OK, watching for changes")
		}
		building = false
		if pending {
			pending = false
			rebuild()
		}
	}

	schedule := func() {
		if building {
			pending = true
			return
		}
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceInterval, rebuild)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			if !wanted[ev.Name] {
				continue
			}
			// CHMOD-only on macOS fires when fonts refresh metadata
			// without content change. Ignore.
			if ev.Op == fsnotify.Chmod {
				continue
			}
			schedule()
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			slog.Warn("Watcher error", "err", err.Error())
		}
	}
}
