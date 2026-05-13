package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boxesandglue/glu/internal/errkind"
)

//go:embed templates/*
var templatesFS embed.FS

// runInit handles `glu init` arguments. Pure positional grammar
// because the top-level optionparser would intercept any --foo
// flag (notably --template, which already means Go text/template
// expansion in Markdown mode).
//
//	glu init                       → scaffold "report" in cwd
//	glu init list                  → print available templates
//	glu init <dir>                 → scaffold "report" in <dir>
//	glu init <dir> <template>      → scaffold <template> in <dir>
//
// Edge case: a target directory literally named "list" must be
// passed as "./list" to dodge the list-subcommand special case.
func runInit(args []string) error {
	if len(args) == 1 && args[0] == "list" {
		return listTemplates(os.Stdout)
	}
	target := "."
	templateName := "report"
	if len(args) >= 1 {
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("%w: 'glu init' takes positional arguments only ('glu init [DIR] [TEMPLATE]'). Got flag %q", errkind.Usage, args[0])
		}
		target = args[0]
	}
	if len(args) >= 2 {
		if strings.HasPrefix(args[1], "-") {
			return fmt.Errorf("%w: 'glu init' takes positional arguments only. Got flag %q", errkind.Usage, args[1])
		}
		templateName = args[1]
	}
	if len(args) > 2 {
		return fmt.Errorf("%w: 'glu init' takes at most two arguments (dir, template), got %d", errkind.Usage, len(args))
	}
	return scaffold(target, templateName)
}

// listTemplates prints the embedded template names with a one-line
// description derived from each template's first H1.
func listTemplates(w io.Writer) error {
	names, err := templateNames()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "Available templates:")
	for _, n := range names {
		descr := templateDescription(n)
		if descr == "" {
			fmt.Fprintf(w, "  %s\n", n)
		} else {
			fmt.Fprintf(w, "  %-10s %s\n", n, descr)
		}
	}
	fmt.Fprintln(w, "\nUsage: glu init [DIR] [TEMPLATE]   (defaults: DIR=. TEMPLATE=report)")
	return nil
}

func templateNames() ([]string, error) {
	entries, err := fs.ReadDir(templatesFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("%w: reading embedded templates: %s", errkind.IO, err.Error())
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// templateDescription extracts the first H1 line from index.md as
// a one-line description, skipping any YAML frontmatter.
func templateDescription(name string) string {
	data, err := fs.ReadFile(templatesFS, "templates/"+name+"/index.md")
	if err != nil {
		return ""
	}
	s := strings.TrimLeft(string(data), " \t\r\n")
	if strings.HasPrefix(s, "---") {
		if i := strings.Index(s[3:], "\n---"); i >= 0 {
			s = s[3+i+4:]
		}
	}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

// scaffold copies the named template tree into dir, refusing to
// overwrite any existing file. dir is created (with parents) if it
// doesn't yet exist.
func scaffold(dir, templateName string) error {
	names, err := templateNames()
	if err != nil {
		return err
	}
	valid := false
	for _, n := range names {
		if n == templateName {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("%w: unknown template %q. Available: %s", errkind.Usage, templateName, strings.Join(names, ", "))
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: creating %s: %s", errkind.IO, dir, err.Error())
	}

	root := "templates/" + templateName
	var written []string
	err = fs.WalkDir(templatesFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, root+"/")
		dest := filepath.Join(dir, rel)
		if _, err := os.Stat(dest); err == nil {
			return fmt.Errorf("%w: refusing to overwrite existing file %s", errkind.IO, dest)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("%w: creating dir %s: %s", errkind.IO, filepath.Dir(dest), err.Error())
		}
		data, err := fs.ReadFile(templatesFS, p)
		if err != nil {
			return fmt.Errorf("%w: reading template %s: %s", errkind.IO, p, err.Error())
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("%w: writing %s: %s", errkind.IO, dest, err.Error())
		}
		written = append(written, rel)
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Created glu project (%s template) in %s:\n", templateName, dir)
	for _, w := range written {
		fmt.Fprintf(os.Stdout, "  %s\n", filepath.Join(dir, w))
	}
	hint := "glu index.md"
	if dir != "." {
		hint = fmt.Sprintf("cd %s && glu index.md", dir)
	}
	fmt.Fprintf(os.Stdout, "\nNext: %s\n", hint)
	return nil
}
