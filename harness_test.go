package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This is the guard behind docs/harness.md. That document says that time,
// randomness, identifiers and outbound connections are taken as dependencies
// rather than read out of the runtime, and that only the wiring supplies the
// real ones. A document alone is an explanation of a rule rather than the rule:
// a handler calling the wall clock compiles, formats and lints exactly like one
// taking a clock, and the suite that then depends on the hour it ran at fails
// once a week for no reason anybody can reproduce.
//
// It reads the source with the toolchain's own parser, so it needs no
// subprocess, no network and no dependency, which are the three things the unit
// suite is not allowed to have. It is the same shape as layout_test.go and sits
// beside it on purpose.
//
// WHAT IT CANNOT SEE, so a green run is not read as more than it is. It matches
// on the name a package is imported under, and it does not resolve scopes: a
// local variable called time shadowing the import would be judged as the
// package. It reads this repository's own source and nothing a dependency does,
// which is a promise the empty dependency set keeps for now rather than one
// this test makes. And it judges the call that is written rather than the call
// that is reached, so a runtime read behind a function this tree does not own
// is invisible to it.

// wiringPackage is the one directory allowed to read the runtime. Everything it
// produces reaches the rest of the program as a field of campaign.Depends, and
// docs/layout.md lets only the entry point import it.
const wiringPackage = "internal/system"

// doublesPackage holds the fakes. Only a test file may import it, so that
// nothing an operator runs can carry a clock that does not move.
const doublesPackage = "github.com/iderex/lesesaal/internal/campaign/campaigntest"

// read is one way of reaching the runtime directly, with the injected field
// that replaces it. An empty name refuses every identifier on the package,
// which is what a package that exists only to produce randomness deserves.
//
// everywhere marks the reads that are refused in the wiring package and in test
// files as well. There is exactly one, and the reason is at it.
type read struct {
	path       string
	name       string
	instead    string
	everywhere bool
}

// refused is the table docs/harness.md describes in prose. The two are meant to
// be read side by side.
var refused = []read{
	// A sleep is refused in every file in the tree, including the wiring and
	// including a test. In a test it is a statement that the author could not
	// express what they were waiting for, and it is the commonest single cause
	// of a suite that fails once a week; in the program it is a wait nobody can
	// move a clock past.
	{path: "time", name: "Sleep", instead: "a clock the caller can move, or the thing you are actually waiting for", everywhere: true},

	{path: "time", name: "Now", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "Since", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "Until", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "After", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "Tick", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "NewTimer", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "NewTicker", instead: "the Now field of campaign.Depends"},
	{path: "time", name: "AfterFunc", instead: "the Now field of campaign.Depends"},

	{path: "math/rand", instead: "the Intn field of campaign.Depends"},
	{path: "math/rand/v2", instead: "the Intn field of campaign.Depends"},
	{path: "crypto/rand", instead: "the Intn and NewID fields of campaign.Depends"},

	{path: "net", name: "Dial", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "DialTimeout", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "DialIP", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "DialTCP", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "DialUDP", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "DialUnix", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "Dialer", instead: "the Dial field of campaign.Depends"},
	{path: "net", name: "Listen", instead: "the entry point, which is where a listener is opened"},
	{path: "net", name: "ListenPacket", instead: "the entry point, which is where a listener is opened"},

	{path: "net/http", name: "Get", instead: "the Dial field of campaign.Depends"},
	{path: "net/http", name: "Head", instead: "the Dial field of campaign.Depends"},
	{path: "net/http", name: "Post", instead: "the Dial field of campaign.Depends"},
	{path: "net/http", name: "PostForm", instead: "the Dial field of campaign.Depends"},
	{path: "net/http", name: "DefaultClient", instead: "the Dial field of campaign.Depends"},
	{path: "net/http", name: "DefaultTransport", instead: "the Dial field of campaign.Depends"},
}

// TestOnlyTheWiringReadsTheRuntime is the one that earns the file. It refuses a
// direct read of the clock, of a random source or of a socket anywhere but the
// wiring package, and it names the file, the line, the call and what to use
// instead, so the failure says which rule was broken rather than only that one
// was.
func TestOnlyTheWiringReadsTheRuntime(t *testing.T) {
	files := 0
	selectors := 0

	forEachGoFile(t, func(file string, dir string, fset *token.FileSet, parsed *ast.File) {
		files++
		imported := importsByName(t, file, parsed)
		inWiring := dir == wiringPackage
		isTest := strings.HasSuffix(filepath.Base(file), "_test.go")

		ast.Inspect(parsed, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			qualifier, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			path, ok := imported[qualifier.Name]
			if !ok {
				return true
			}
			selectors++
			for _, rule := range refused {
				if rule.path != path {
					continue
				}
				if rule.name != "" && rule.name != selector.Sel.Name {
					continue
				}
				if inWiring && !rule.everywhere {
					continue
				}
				where := "outside " + wiringPackage
				if rule.everywhere {
					where = "anywhere in this tree"
					if isTest {
						where = "in a test"
					}
				}
				t.Errorf("%s:%d calls %s.%s, which reads the runtime and is refused %s. Take %s instead.",
					file, fset.Position(selector.Pos()).Line, qualifier.Name, selector.Sel.Name, where, rule.instead)
			}
			return true
		})
	})

	// A run that examined nothing is not a run that examined everything and
	// found nothing. Both print the same word otherwise.
	if files == 0 {
		t.Fatal("no Go file was examined, so nothing was judged")
	}
	if selectors == 0 {
		t.Fatal("no qualified identifier was examined, so the table above judged nothing")
	}
	t.Logf("examined %d Go file(s) and %d qualified identifier(s) against %d refused read(s)",
		files, selectors, len(refused))
}

// TestOnlyATestImportsTheFakes keeps the doubles out of the program. A fake
// clock reaching the binary an operator runs is worse than no fake at all: it
// would be a deployment whose time does not move, and nothing else here would
// notice.
func TestOnlyATestImportsTheFakes(t *testing.T) {
	files := 0
	importers := 0

	forEachGoFile(t, func(file string, _ string, _ *token.FileSet, parsed *ast.File) {
		files++
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Errorf("%s carries an import path that is not a quoted string: %s", file, spec.Path.Value)
				continue
			}
			if path != doublesPackage {
				continue
			}
			importers++
			if strings.HasSuffix(filepath.Base(file), "_test.go") {
				continue
			}
			t.Errorf("%s imports %s and is not a test file. The fakes are for tests; a program that carries them carries a clock that does not move.",
				file, path)
		}
	})

	if files == 0 {
		t.Fatal("no Go file was examined, so nothing was judged")
	}
	t.Logf("examined %d Go file(s); %d of them import the fakes", files, importers)
}

// TestNoDotImport refuses an import with no name. Every rule above is decided
// by the name a package is imported under, and a dot import puts identifiers
// into the file with no qualifier at all, so it would walk through this whole
// table unseen.
func TestNoDotImport(t *testing.T) {
	files := 0

	forEachGoFile(t, func(file string, _ string, _ *token.FileSet, parsed *ast.File) {
		files++
		for _, spec := range parsed.Imports {
			if spec.Name != nil && spec.Name.Name == "." {
				t.Errorf("%s imports %s with a dot, and every rule in this file is decided by the name a package is imported under.",
					file, spec.Path.Value)
			}
		}
	})

	if files == 0 {
		t.Fatal("no Go file was examined, so nothing was judged")
	}
	t.Logf("examined %d Go file(s) for a dot import", files)
}

// forEachGoFile parses every tracked Go file in the tree and hands it to visit
// with its slashed path and the directory it sits in. It fails the test rather
// than returning an error, because a tree this cannot be walked is a tree
// nothing here judged.
func forEachGoFile(t *testing.T, visit func(file string, dir string, fset *token.FileSet, parsed *ast.File)) {
	t.Helper()
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			// The set of directories holding no source is layout_test.go's,
			// read from there rather than repeated here: two copies of it would
			// let one test judge a directory the other does not.
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
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Errorf("could not read %s: %v", file, err)
			return nil
		}
		visit(file, dir, fset, parsed)
		return nil
	})
	if err != nil {
		t.Fatalf("could not walk the tree: %v", err)
	}
}

// importsByName maps the name a package is used under in this file to its
// import path. A blank import binds no name and is left out; a dot import is
// left out here and refused by its own test above.
func importsByName(t *testing.T, file string, parsed *ast.File) map[string]string {
	t.Helper()
	byName := make(map[string]string, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Errorf("%s carries an import path that is not a quoted string: %s", file, spec.Path.Value)
			continue
		}
		name := path
		if slash := strings.LastIndex(name, "/"); slash >= 0 {
			name = name[slash+1:]
		}
		if spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				continue
			}
			name = spec.Name.Name
		}
		byName[name] = path
	}
	return byName
}
