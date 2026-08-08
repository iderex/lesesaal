package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This is the guard behind docs/layout.md. That document says which way a
// dependency may point between the parts of this tree, and a document alone is
// an explanation of a rule rather than the rule: a handler reaching into the
// store, or the core importing the surface, compiles perfectly and is invisible
// in a diff that only shows the import line.
//
// It reads the import declarations out of the source rather than asking the
// toolchain, so it needs no subprocess, no network and no dependency, which are
// the three things the unit suite in #20 is not allowed to have.

// part is one directory the layout names, with the parts inside this module its
// files are permitted to import. A part not listed in mayImport is refused.
type part struct {
	dir       string
	mayImport []string
}

// layout is the table in docs/layout.md, in the order that document lists it.
// The two are meant to be read side by side, and a change to one that is not
// made in the other is the drift this test exists to make expensive.
var layout = []part{
	{dir: ".", mayImport: []string{
		"internal/campaign",
		"internal/model",
		"internal/store",
		"internal/store/migration",
		"internal/web",
	}},
	{dir: "internal/campaign", mayImport: nil},
	{dir: "internal/model", mayImport: []string{"internal/campaign"}},
	{dir: "internal/store", mayImport: []string{"internal/campaign"}},
	{dir: "internal/store/migration", mayImport: []string{"internal/campaign"}},
	{dir: "internal/web", mayImport: []string{"internal/campaign", "internal/model"}},
	{dir: "deploy", mayImport: nil},
	{dir: "test", mayImport: nil},
	{dir: "docs", mayImport: nil},
}

// skipped directories hold no part of the layout. A vendored or generated tree
// would be judged by whoever generated it, and .git is not source.
var skipped = map[string]bool{".git": true, ".github": true, "testdata": true}

// TestEveryDirectoryTheLayoutNamesExists refuses a layout table that has drifted
// away from the tree it describes. Without this, a renamed directory would leave
// every rule about it unenforced and every check below still green, because
// nothing would be found in it to judge.
func TestEveryDirectoryTheLayoutNamesExists(t *testing.T) {
	for _, p := range layout {
		info, err := os.Stat(p.dir)
		if err != nil {
			t.Errorf("the layout names %s and the tree does not carry it: %v", p.dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("the layout names %s as a directory and it is not one", p.dir)
		}
	}
}

// TestDependenciesPointTheWayTheLayoutSays is the one that earns the file. It
// refuses an import that crosses a boundary in the direction docs/layout.md
// refuses, and it names both ends so the failure says which rule was broken
// rather than only that one was.
func TestDependenciesPointTheWayTheLayoutSays(t *testing.T) {
	module := modulePath(t)

	permitted := make(map[string]map[string]bool, len(layout))
	for _, p := range layout {
		allowed := make(map[string]bool, len(p.mayImport))
		for _, target := range p.mayImport {
			allowed[target] = true
		}
		permitted[p.dir] = allowed
	}

	files := 0
	imports := 0
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skipped[entry.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		// Slashed for the report as well as for the lookup, so a failure reads
		// the same on every platform and can be pasted into an issue as it is.
		file := filepath.ToSlash(path)
		dir := filepath.ToSlash(filepath.Dir(path))
		allowed, known := permitted[dir]
		if !known {
			t.Errorf("%s sits in %s, which docs/layout.md does not name", file, dir)
			return nil
		}
		files++

		// ImportsOnly: the declarations are the whole subject here, and parsing
		// the bodies would make this slower for nothing.
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("could not read the imports of %s: %v", file, err)
			return nil
		}
		for _, spec := range parsed.Imports {
			line, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Errorf("%s carries an import path that is not a quoted string: %s", file, spec.Path.Value)
				continue
			}
			// Anything outside this module is the standard library or a
			// dependency, and neither is this test's subject. What the module
			// is allowed to depend on at all is the build check's, #19.
			if !strings.HasPrefix(line, module+"/") {
				continue
			}
			target := strings.TrimPrefix(line, module+"/")
			imports++
			if target == dir || allowed[target] {
				continue
			}
			t.Errorf("%s imports %s, and docs/layout.md does not let %s depend on %s",
				file, line, dir, target)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("could not walk the tree: %v", err)
	}

	// A run that examined nothing is not a run that examined everything and
	// found nothing. Both print the same word otherwise, and only one of them is
	// worth a green tick.
	if files == 0 {
		t.Fatal("no Go file was examined, so nothing was judged")
	}
	t.Logf("examined %d Go file(s) and %d import(s) of this module against the layout", files, imports)
}

// modulePath reads the module path from go.mod rather than repeating it here,
// so a module rename fails this test loudly instead of leaving it matching
// nothing and passing.
func modulePath(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("could not read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if rest, found := strings.CutPrefix(line, "module "); found {
			path := strings.TrimSpace(rest)
			if path == "" {
				break
			}
			return path
		}
	}
	t.Fatal("go.mod declares no module path")
	return ""
}
