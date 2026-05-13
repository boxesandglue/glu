package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/boxesandglue/boxesandglue/frontend"
)

// runDoctor reports environment information useful for debugging glu
// installations. Each check prints [OK] / [WARN] / [FAIL]. Returns
// the number of FAILs so the caller can use it as the exit code; 0
// means everything passed (WARNs do not contribute).
func runDoctor(out io.Writer) int {
	d := &doctor{out: out}

	d.section("Build info")
	d.checkBuildInfo()

	d.section("Filesystem")
	d.checkCwdWritable()

	d.section("Hyphenation patterns")
	d.checkHyphenation()

	d.section("Font directories")
	d.checkFontDirs()

	d.section("External tools (optional)")
	d.checkExternalTools()

	fmt.Fprintln(out)
	if d.fails == 0 {
		fmt.Fprintf(out, "%d check(s) OK, %d warning(s).\n", d.oks, d.warns)
	} else {
		fmt.Fprintf(out, "%d check(s) OK, %d warning(s), %d FAILURE(s).\n", d.oks, d.warns, d.fails)
	}
	return d.fails
}

type doctor struct {
	out               io.Writer
	oks, warns, fails int
}

func (d *doctor) section(title string) {
	fmt.Fprintf(d.out, "\n== %s ==\n", title)
}

func (d *doctor) ok(msg string, args ...any) {
	d.oks++
	fmt.Fprintf(d.out, "  [OK]   "+msg+"\n", args...)
}

func (d *doctor) warn(msg string, args ...any) {
	d.warns++
	fmt.Fprintf(d.out, "  [WARN] "+msg+"\n", args...)
}

func (d *doctor) fail(msg string, args ...any) {
	d.fails++
	fmt.Fprintf(d.out, "  [FAIL] "+msg+"\n", args...)
}

func (d *doctor) checkBuildInfo() {
	d.ok("glu version=%s", Version)
	d.ok("Go runtime=%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok {
		var rev, vcsTime string
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.time":
				vcsTime = s.Value
			}
		}
		if rev != "" {
			d.ok("vcs revision=%s time=%s", rev, vcsTime)
		}
	}
}

func (d *doctor) checkCwdWritable() {
	cwd, err := os.Getwd()
	if err != nil {
		d.fail("getwd: %s", err)
		return
	}
	f, err := os.CreateTemp(cwd, ".glu-doctor-*")
	if err != nil {
		d.fail("cwd %s is not writable: %s", cwd, err)
		return
	}
	f.Close()
	os.Remove(f.Name())
	d.ok("cwd %s is writable", cwd)
}

func (d *doctor) checkHyphenation() {
	// Sample a spread of language families. The list mirrors what
	// boxesandglue's pattern table is most likely to carry.
	tags := []string{"en", "de", "fr", "es", "it", "nl", "pl", "pt", "ru", "cs", "hu", "tr", "ar", "he", "zh", "ja", "ko"}
	var supported, missing []string
	for _, t := range tags {
		if frontend.IsHyphenationSupported(t) {
			supported = append(supported, t)
		} else {
			missing = append(missing, t)
		}
	}
	d.ok("supported: %s", strings.Join(supported, " "))
	if len(missing) > 0 {
		d.warn("no-op (no patterns, hyphenation disabled): %s", strings.Join(missing, " "))
	}
}

func (d *doctor) checkFontDirs() {
	dirs := []string{}
	switch runtime.GOOS {
	case "darwin":
		dirs = append(dirs, "/System/Library/Fonts", "/Library/Fonts")
		if home, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
		}
	case "linux":
		dirs = append(dirs, "/usr/share/fonts", "/usr/local/share/fonts")
		if home, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(home, ".local", "share", "fonts"), filepath.Join(home, ".fonts"))
		}
	case "windows":
		if win := os.Getenv("WINDIR"); win != "" {
			dirs = append(dirs, filepath.Join(win, "Fonts"))
		}
	}
	any := false
	for _, dir := range dirs {
		n := countFonts(dir)
		if n > 0 {
			d.ok("%s — %d font file(s)", dir, n)
			any = true
		} else if _, err := os.Stat(dir); err == nil {
			d.warn("%s — exists but no .ttf/.otf found", dir)
		}
	}
	if !any {
		d.warn("no fonts discovered in standard locations — text may render as fallback")
	}
}

func countFonts(dir string) int {
	n := 0
	filepath.WalkDir(dir, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if e.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".ttf" || ext == ".otf" || ext == ".ttc" {
			n++
		}
		return nil
	})
	return n
}

func (d *doctor) checkExternalTools() {
	tools := []struct{ name, why string }{
		{"qpdf", "PDF inspection, normalisation (--qdf)"},
		{"pdfinfo", "PDF metadata diagnostics"},
		{"verapdf", "PDF/UA conformance validation"},
		{"pdftoppm", "PDF→PNG rendering for visual diffs"},
	}
	for _, t := range tools {
		p, err := exec.LookPath(t.name)
		if err != nil {
			d.warn("%s not on PATH (%s)", t.name, t.why)
			continue
		}
		// Version probe is tool-specific; just report the path.
		d.ok("%s at %s — %s", t.name, p, t.why)
	}
}
