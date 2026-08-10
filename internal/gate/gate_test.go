package gate

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// The suite runs no command. Every leg reaches the machine through Env, so a
// test decides what `go build` returned instead of building, which is what
// keeps this package inside the rule that the unit suite has no subprocess, no
// network and no dependency.

// fake answers a command from a table and a file from another. A command the
// table does not name succeeds with the empty string, so a test writes down
// only the commands it is actually about; a file the tree does not carry is an
// error, because a leg reading one is asking about something it was told
// exists.
type fake struct {
	answers map[string]answer
	tools   map[string]bool
	files   map[string]string
	calls   []string
}

type answer struct {
	output string
	err    error
}

func (f *fake) run(name string, args ...string) (string, error) {
	line := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, line)
	if given, ok := f.answers[line]; ok {
		return given.output, given.err
	}
	return "", nil
}

func (f *fake) look(name string) bool { return f.tools[name] }

func (f *fake) read(name string) (string, error) {
	if content, ok := f.files[name]; ok {
		return content, nil
	}
	return "", errors.New("open " + name + ": no such file or directory")
}

func (f *fake) env() Env { return Env{Run: f.run, Look: f.look, Read: f.read} }

// paths is the fixture tree, in the order git lists it. The text and document
// legs are judged against this rather than against this repository, because a
// leg pointed at the tree it lives in proves the state of that tree on the day
// it ran and never the leg.
var paths = []string{".gitattributes", "README.md", "docs/thing.md", "internal/gate/gate.go", "main.go"}

// tree is what those paths hold. Every document in it is clean and each test
// below spoils exactly one thing about it.
//
// The two documents carry between them everything the three document legs have
// to get right: a path that resolves from the root, one that resolves from the
// directory of the document naming it, a directory, a backticked word that is
// not a path at all, a link with a fragment, a link that leaves the
// repository, and two dangling paths that are only safe because one sits in an
// indented block and the other inside a fence.
func tree() map[string]string {
	return map[string]string{
		".gitattributes": "* text=auto eol=lf\n",
		"README.md": "# A tree\n" +
			"\n" +
			"The entry point is `main.go` and the attributes are in\n" +
			"`.gitattributes`, which carries no extension any leg here knows.\n" +
			"\n" +
			"See [the thing](docs/thing.md) and [a heading](docs/thing.md#the-thing),\n" +
			"and [a page](https://example.invalid/) that no leg follows.\n",
		"docs/thing.md": "# The thing\n" +
			"\n" +
			"It points back at `README.md`, and the directory it sits in is `docs/`.\n" +
			"\n" +
			"    a transcript naming `nowhere/at/all.md`, which nothing resolves\n" +
			"\n" +
			"```\n" +
			"and a fence naming `also/nowhere.md`\n" +
			"```\n",
		"internal/gate/gate.go": "package gate\n",
		"main.go":               "package main\n\nfunc main() {}\n",
	}
}

// eolListing is the shape `git ls-files --eol` prints: the index attribute
// first, the fields padded with spaces, and the path after a tab. The fixture
// writes that shape rather than a tidier one, because it is what the parser
// reads.
func eolListing(stored map[string]string) string {
	var listing strings.Builder
	for _, file := range paths {
		kind := stored[file]
		if kind == "" {
			kind = "i/lf"
		}
		fmt.Fprintf(&listing, "%-8s%-8sattr/text=auto eol=lf \t%s\n", kind, "w/lf", file)
	}
	return listing.String()
}

// green is a machine where everything the legs ask for answers the way a clean
// tree answers. Each test spoils exactly one entry, so the failure is the
// spoiling rather than the setup.
func green() *fake {
	return &fake{
		tools: map[string]bool{"staticcheck": true},
		files: tree(),
		answers: map[string]answer{
			"git ls-files .":                                    {output: strings.Join(paths, "\n") + "\n"},
			"git ls-files --eol":                                {output: eolListing(nil)},
			"git ls-files go.sum":                               {output: ""},
			"go mod tidy -diff":                                 {output: ""},
			"go list -mod=readonly -m all":                      {output: "github.com/iderex/lesesaal\n"},
			"go list -mod=readonly ./...":                       {output: "github.com/iderex/lesesaal\ngithub.com/iderex/lesesaal/internal/gate\n"},
			"go list -mod=readonly -deps ./...":                 {output: "fmt\nio\nstrings\ngithub.com/iderex/lesesaal\n"},
			"git ls-files *.go":                                 {output: "main.go\ninternal/gate/gate.go\n"},
			"gofmt -l main.go internal/gate/gate.go":            {output: ""},
			"staticcheck -version":                              {output: "staticcheck 2025.1 (v0.7.0)\n"},
			"staticcheck -list-checks":                          {output: "SA1000\nSA1001\nST1000\n"},
			"go test -mod=readonly -list .* ./...":              {output: "TestOne\nTestTwo\nok  \tgithub.com/iderex/lesesaal\t0.1s\n"},
			"go test -mod=readonly -count=1 -failfast -v ./...": {output: "=== RUN   TestOne\n--- PASS: TestOne (0.00s)\n--- PASS: TestTwo (0.00s)\nok  \tgithub.com/iderex/lesesaal\t0.1s\n"},
		},
	}
}

func runAll(t *testing.T, f *fake) (string, int) {
	t.Helper()
	var report bytes.Buffer
	code := Run(&report, f.env(), "")
	return report.String(), code
}

// TestAGreenTreeRunsEveryLeg is the baseline the rest of the file spoils. It
// also fixes the sentence a reader trusts: the count in the summary is the
// count of legs, not of commands.
func TestAGreenTreeRunsEveryLeg(t *testing.T) {
	report, code := runAll(t, green())
	if code != 0 {
		t.Errorf("a tree where every command answers cleanly exited %d, and 0 is the only right answer:\n%s", code, report)
	}
	for _, leg := range Legs() {
		if !strings.Contains(report, "PASS  "+leg.ID) {
			t.Errorf("leg %s is not reported as passing:\n%s", leg.ID, report)
		}
	}
	if want := "15 leg(s) passed, 0 failed, 0 did not run"; !strings.Contains(report, want) {
		t.Errorf("the summary does not say %q:\n%s", want, report)
	}
}

// TestTheRunStopsAtTheFirstFailure is the property #150 asks for by name. The
// near miss is a failure in the middle of the order rather than at the end,
// because a failure in the last leg stops nothing and would pass a command
// that never stopped at all.
func TestTheRunStopsAtTheFirstFailure(t *testing.T) {
	f := green()
	f.answers["go build -mod=readonly ./..."] = answer{output: "main.go:9:2: undefined: nothing", err: errors.New("exit status 1")}

	report, code := runAll(t, f)
	if code == 0 {
		t.Errorf("a tree that does not compile exited 0:\n%s", report)
	}
	if !strings.Contains(report, "FAIL  build") {
		t.Errorf("the failing leg is not named in the report:\n%s", report)
	}
	if !strings.Contains(report, "undefined: nothing") {
		t.Errorf("what the compiler said is not in the report, so the report says only that something is wrong:\n%s", report)
	}
	for _, later := range []string{"gofmt", "vet", "staticcheck", "test-selection", "test"} {
		if strings.Contains(report, "PASS  "+later) {
			t.Errorf("leg %s ran after the failure, so the run did not stop:\n%s", later, report)
		}
	}
	if !strings.Contains(report, "The run stopped early") {
		t.Errorf("a run that stopped early does not say so, so it reads as one that covered everything:\n%s", report)
	}
}

// TestALegThatCannotRunIsNotALegThatPassed is the third state. A missing
// linter is the case that actually happens on a fresh clone, and the failure
// this guards against is a green tick over a gate that skipped a leg.
func TestALegThatCannotRunIsNotALegThatPassed(t *testing.T) {
	f := green()
	f.tools["staticcheck"] = false

	report, code := runAll(t, f)
	if code != 0 {
		t.Errorf("a leg that did not run failed the whole run, and it is a disclosure rather than a verdict:\n%s", report)
	}
	if !strings.Contains(report, "NOT RUN  staticcheck") {
		t.Errorf("the leg that did not run is not reported:\n%s", report)
	}
	if !strings.Contains(report, staticcheckInstall) {
		t.Errorf("the report does not say what would make the leg run:\n%s", report)
	}
	if !strings.Contains(report, "1 did not run") {
		t.Errorf("the summary counts the leg as run:\n%s", report)
	}
	if strings.Contains(report, "PASS  staticcheck") {
		t.Errorf("a leg that did not run is reported as passing:\n%s", report)
	}
}

// TestNamingALegRequiresIt is the other half of the state above, and it is
// what the workflow steps depend on. The server asks for one leg by name, and
// a leg that cannot run there has to red rather than report a disclosure
// nobody reads.
func TestNamingALegRequiresIt(t *testing.T) {
	f := green()
	f.tools["staticcheck"] = false

	var report bytes.Buffer
	if code := Run(&report, f.env(), "staticcheck"); code == 0 {
		t.Errorf("a leg asked for by name and not run exited 0:\n%s", report.String())
	}
}

// TestNamingOneLegRunsOnlyThatLeg keeps the workflow contract honest: a step
// that asks for the formatter must not run the suite as well.
func TestNamingOneLegRunsOnlyThatLeg(t *testing.T) {
	f := green()

	var report bytes.Buffer
	if code := Run(&report, f.env(), "gofmt"); code != 0 {
		t.Errorf("a clean tree exited %d for one leg:\n%s", code, report.String())
	}
	for _, call := range f.calls {
		if strings.HasPrefix(call, "go test") || strings.HasPrefix(call, "go build") {
			t.Errorf("asking for gofmt also ran %q", call)
		}
	}
}

// TestAnUnknownLegIsRefused. Silently running everything on a mistyped name
// would make a workflow step that asks for a leg that no longer exists look
// like a step that passed.
func TestAnUnknownLegIsRefused(t *testing.T) {
	f := green()

	var report bytes.Buffer
	code := Run(&report, f.env(), "gofmtt")
	if code == 0 {
		t.Errorf("an unknown leg name exited 0:\n%s", report.String())
	}
	if !strings.Contains(report.String(), "gofmt") {
		t.Errorf("the refusal does not list the legs that do exist:\n%s", report.String())
	}
	if len(f.calls) != 0 {
		t.Errorf("an unknown leg name still ran %v", f.calls)
	}
}

// TestALegFailsClosedOnAnEmptySelection is the near miss that a check over a
// tree is most likely to get wrong. A selection that came back empty and a
// selection that came back clean print the same word, and only one of them is
// worth a green tick.
func TestALegFailsClosedOnAnEmptySelection(t *testing.T) {
	spoil := map[string]string{
		"gofmt":        "git ls-files *.go",
		"build":        "go list -mod=readonly ./...",
		"deps":         "go list -mod=readonly -deps ./...",
		"line-endings": "git ls-files --eol",
		"encoding":     "git ls-files --eol",
		"unicode":      "git ls-files --eol",
		"doc-format":   "git ls-files .",
		"doc-paths":    "git ls-files .",
		"doc-links":    "git ls-files .",
		"lock":         "go list -mod=readonly -m all",
	}
	for leg, command := range spoil {
		f := green()
		f.answers[command] = answer{output: "\n"}

		var report bytes.Buffer
		if code := Run(&report, f.env(), leg); code == 0 {
			t.Errorf("leg %s passed on an empty selection from %q:\n%s", leg, command, report.String())
		}
		if !strings.Contains(report.String(), "Failing closed") {
			t.Errorf("leg %s does not say it failed closed:\n%s", leg, report.String())
		}
	}
}

// TestEveryLegSaysWhatItExamined. The second condition of #150 is that the run
// says what it looked at, and a leg that passes silently makes the whole
// report unreadable in exactly the way the condition is about.
func TestEveryLegSaysWhatItExamined(t *testing.T) {
	f := green()
	for _, leg := range Legs() {
		result := leg.Judge(f.env())
		if result.Outcome != Passed {
			t.Fatalf("leg %s did not pass on a clean machine: %+v", leg.ID, result)
		}
		if !strings.HasPrefix(result.Examined, "Examined ") {
			t.Errorf("leg %s reports %q, which does not say what it examined", leg.ID, result.Examined)
		}
	}
}

// TestZeroRequirementsIsAnAnswerRatherThanAnEmptySelection. The module carries
// no dependency, so the leg above it would fail closed on the same shape. The
// difference is written down here rather than left to be rediscovered.
func TestZeroRequirementsIsAnAnswerRatherThanAnEmptySelection(t *testing.T) {
	result := modVerify(green().env())
	if result.Outcome != Passed {
		t.Fatalf("a module with no requirement did not pass: %+v", result)
	}
	if !strings.Contains(result.Examined, "carries none") {
		t.Errorf("the leg reports %q, which does not distinguish no requirement from none examined", result.Examined)
	}
}

// TestTheRunSaysWhatItDoesNotRun is the fourth condition of #150. A leg that is
// absent from the report is one a reader takes the run for having covered.
func TestTheRunSaysWhatItDoesNotRun(t *testing.T) {
	report, _ := runAll(t, green())
	if len(NotCovered()) == 0 {
		t.Fatal("nothing is declared uncovered, which would mean this command runs the whole gate")
	}
	for _, leg := range NotCovered() {
		if !strings.Contains(report, leg.Name) {
			t.Errorf("the run does not say that %q was not run:\n%s", leg.Name, report)
		}
		if leg.Reason == "" {
			t.Errorf("%q is declared uncovered with no reason, which is the silence this list exists against", leg.Name)
		}
	}
}

// TestASuiteThatSelectsNothingIsRefused. The toolchain calls a package with no
// test file a success, so a change that moves the suite somewhere this command
// does not look would leave the run green and the suite unrun. That is the
// failure that makes every other green here mean nothing.
func TestASuiteThatSelectsNothingIsRefused(t *testing.T) {
	f := green()
	f.answers["go test -mod=readonly -list .* ./..."] = answer{output: "ok  \tgithub.com/iderex/lesesaal\t0.1s\n"}

	var report bytes.Buffer
	if code := Run(&report, f.env(), "test-selection"); code == 0 {
		t.Errorf("a suite that selects no test passed:\n%s", report.String())
	}
	if !strings.Contains(report.String(), "selected no test at all") {
		t.Errorf("the refusal does not say what was wrong:\n%s", report.String())
	}
}

// TestAListingThatFailedIsNotASuiteOfZero. Those are different failures and
// they name different repairs, so reading one as the other sends somebody to
// look for a missing test that is not missing.
func TestAListingThatFailedIsNotASuiteOfZero(t *testing.T) {
	f := green()
	f.answers["go test -mod=readonly -list .* ./..."] = answer{output: "build failed", err: errors.New("exit status 2")}

	var report bytes.Buffer
	if code := Run(&report, f.env(), "test-selection"); code == 0 {
		t.Errorf("a listing that could not be produced passed:\n%s", report.String())
	}
	if !strings.Contains(report.String(), "could not list the tests") {
		t.Errorf("the refusal reads as a suite of zero rather than as a toolchain that could not answer:\n%s", report.String())
	}
}

// TestASuiteWhereNothingRanIsRefused is the same shape one step later. The
// toolchain can exit successfully having run nothing at all, and a transcript
// with no test in it is not a suite that passed.
func TestASuiteWhereNothingRanIsRefused(t *testing.T) {
	f := green()
	f.answers["go test -mod=readonly -count=1 -failfast -v ./..."] = answer{output: "?   \tgithub.com/iderex/lesesaal\t[no test files]\n"}

	var report bytes.Buffer
	if code := Run(&report, f.env(), "test"); code == 0 {
		t.Errorf("a run in which no test ran passed:\n%s", report.String())
	}
	if !strings.Contains(report.String(), "No test ran") {
		t.Errorf("the refusal does not say that nothing ran:\n%s", report.String())
	}
}

// TestTheFormatterSaysWhatItWouldChange. A list of file names sends somebody
// back to run the tool themselves, which is the round trip this command exists
// to remove.
func TestTheFormatterSaysWhatItWouldChange(t *testing.T) {
	f := green()
	f.answers["gofmt -l main.go internal/gate/gate.go"] = answer{output: "main.go\n"}
	f.answers["gofmt -d main.go"] = answer{output: "--- main.go.orig\n+++ main.go\n-\tx := 1\n+\tx := 1\n"}

	var report bytes.Buffer
	if code := Run(&report, f.env(), "gofmt"); code == 0 {
		t.Errorf("a tree the formatter would change passed:\n%s", report.String())
	}
	if !strings.Contains(report.String(), "+++ main.go") {
		t.Errorf("the report does not carry what the formatter would change it to:\n%s", report.String())
	}
}

// refuse runs one leg over a spoiled tree and returns what it reported. Every
// near miss below is one change away from green(), so what reddened is the
// change and not the setup.
func refuse(t *testing.T, f *fake, leg string) string {
	t.Helper()
	var report bytes.Buffer
	if code := Run(&report, f.env(), leg); code == 0 {
		t.Errorf("leg %s passed on a tree it should have refused:\n%s", leg, report.String())
	}
	return report.String()
}

// pass runs one leg over a tree that should survive it. It is the other half of
// every exemption: a leg that refused nothing proves nothing, and a leg that
// refuses the thing it is supposed to allow is worse than one that is absent.
func pass(t *testing.T, f *fake, leg string) string {
	t.Helper()
	var report bytes.Buffer
	if code := Run(&report, f.env(), leg); code != 0 {
		t.Errorf("leg %s refused a tree it should have allowed:\n%s", leg, report.String())
	}
	return report.String()
}

// TestAFileNotStoredWithLFIsRefused. The blob is what a fresh clone on any
// platform receives, so this is the byte that decides whether the repository
// depends on the machine a file was written on. Neither state is visible in a
// diff.
func TestAFileNotStoredWithLFIsRefused(t *testing.T) {
	for _, stored := range []string{"i/crlf", "i/mixed"} {
		f := green()
		f.answers["git ls-files --eol"] = answer{output: eolListing(map[string]string{"README.md": stored})}

		report := refuse(t, f, "line-endings")
		if !strings.Contains(report, "README.md is stored as "+strings.TrimPrefix(stored, "i/")) {
			t.Errorf("the report does not name the file and how it is stored:\n%s", report)
		}
		if !strings.Contains(report, "git add --renormalize") {
			t.Errorf("the report does not say what repairs it:\n%s", report)
		}
	}
}

// TestABinaryFileIsNotJudgedForItsLineEndings. A file git treats as binary has
// no line endings to declare, and refusing one would make the leg unusable the
// day this tree carries an image.
func TestABinaryFileIsNotJudgedForItsLineEndings(t *testing.T) {
	f := green()
	f.answers["git ls-files --eol"] = answer{output: eolListing(map[string]string{"main.go": "i/-text"})}

	report := pass(t, f, "line-endings")
	if !strings.Contains(report, "Examined 4 tracked text file(s) of 5 tracked file(s)") {
		t.Errorf("the leg does not distinguish what it examined from what it skipped:\n%s", report)
	}
}

// TestAFileThatIsNotUTF8IsRefused. The trip case is one byte of latin-1, which
// is what an editor that chose the machine's own code page leaves behind, and
// it renders as a replacement character rather than as an error.
func TestAFileThatIsNotUTF8IsRefused(t *testing.T) {
	f := green()
	f.files["docs/thing.md"] = tree()["docs/thing.md"] + "caf\xe9 is latin-1 here\n"

	report := refuse(t, f, "encoding")
	if !strings.Contains(report, "docs/thing.md does not decode as UTF-8") {
		t.Errorf("the report does not name the file that does not decode:\n%s", report)
	}
}

// TestAHiddenControlCharacterIsRefused. Both trip cases are invisible: a
// zero-width space in the middle of a word, which is what arrives when
// something is pasted in from a rendered page, and a right-to-left override in
// ordinary prose, which is the Trojan Source attack itself.
func TestAHiddenControlCharacterIsRefused(t *testing.T) {
	hidden := []struct {
		what  string
		line  string
		point string
	}{
		{"a zero-width space inside a word", "the retro\u200bfitted guard\n", "U+200B (ZERO WIDTH SPACE)"},
		{"a right-to-left override in prose", "the \u202eguard\n", "U+202E (RIGHT-TO-LEFT OVERRIDE)"},
	}
	for _, one := range hidden {
		f := green()
		f.files["docs/thing.md"] = tree()["docs/thing.md"] + one.line

		report := refuse(t, f, "unicode")
		if !strings.Contains(report, "docs/thing.md:10 carries "+one.point) {
			t.Errorf("%s: the report does not name the file, the line and the codepoint:\n%s", one.what, report)
		}
	}
}

// TestAByteOrderMarkIsNotRefused. It is the character the set above
// deliberately leaves out, and a leg that refused it would red a tree for
// something that is not the defect.
func TestAByteOrderMarkIsNotRefused(t *testing.T) {
	f := green()
	f.files["docs/thing.md"] = "\ufeff" + tree()["docs/thing.md"]

	pass(t, f, "unicode")
}

// TestADocumentThisTreeWouldNotStoreIsRefused walks the formatting leg through
// one mistake at a time. Each spoiling is a change somebody makes rather than a
// document that could not have been written by hand.
func TestADocumentThisTreeWouldNotStoreIsRefused(t *testing.T) {
	clean := tree()["docs/thing.md"]
	// One column over the limit, which is the only width mistake worth
	// proving: a line of two hundred columns would red on a leg that had lost
	// the boundary entirely.
	over := strings.Repeat("x", columns+1)
	spoiled := []struct {
		what    string
		content string
		says    string
	}{
		{"a line one column over the width", clean + over + "\n", "is 82 columns, over the 81 this tree wraps at"},
		{"a line that ends in a space", clean + "a sentence ending in a space \n", "has trailing whitespace"},
		{"a tab", clean + "a sentence\twith a tab in it\n", "contains a tab"},
		{"two blank lines in a row", clean + "\n\nafter two blank lines\n", "is the second of two blank lines"},
		{"no newline at the end", strings.TrimSuffix(clean, "\n"), "does not end with a newline"},
		{"a blank line at the end", clean + "\n", "ends with a blank line"},
		{"nothing at all", "", "is empty"},
	}
	for _, one := range spoiled {
		f := green()
		f.files["docs/thing.md"] = one.content

		report := refuse(t, f, "doc-format")
		if !strings.Contains(report, one.says) {
			t.Errorf("%s: the report does not say %q:\n%s", one.what, one.says, report)
		}
	}
}

// TestAPastedTranscriptKeepsItsWidth is the other direction, and it is why the
// leg cannot simply refuse every long line. A command and its output have to
// stay byte-exact, and a leg that wrapped one would produce a document whose
// commands nobody can run.
func TestAPastedTranscriptKeepsItsWidth(t *testing.T) {
	long := strings.Repeat("x", columns+1)
	f := green()
	f.files["docs/thing.md"] = tree()["docs/thing.md"] +
		"```\n" + long + "\n```\n" +
		"\n" +
		"    " + long + "\n" +
		"\n" +
		strings.Repeat("x", columns) + "\n"

	pass(t, f, "doc-format")
}

// TestADocumentNamingAPathThatIsNotThereIsRefused. This is the leg that earns
// the three: a pointer to something renamed is invisible until somebody
// follows it, and the fixture's dangling path is one that already sits in the
// tree inside a code block, so what reddens is the exemption ending rather
// than a new mistake.
func TestADocumentNamingAPathThatIsNotThereIsRefused(t *testing.T) {
	f := green()
	f.files["docs/thing.md"] = strings.Replace(tree()["docs/thing.md"],
		"```\nand a fence naming `also/nowhere.md`\n```\n",
		"and a sentence naming `also/nowhere.md`\n", 1)

	report := refuse(t, f, "doc-paths")
	if !strings.Contains(report, "docs/thing.md names also/nowhere.md, which does not exist") {
		t.Errorf("the report does not name the document and the path it promised:\n%s", report)
	}
}

// TestAPathInAnIndentedTranscriptIsNotJudged is the same fixture from the
// other side. A pasted transcript names paths in other repositories and on
// other people's machines, and this tree cannot resolve those.
func TestAPathInAnIndentedTranscriptIsNotJudged(t *testing.T) {
	f := green()
	f.files["docs/thing.md"] = strings.Replace(tree()["docs/thing.md"],
		"    a transcript naming `nowhere/at/all.md`",
		"a sentence naming `nowhere/at/all.md`", 1)

	report := refuse(t, f, "doc-paths")
	if !strings.Contains(report, "names nowhere/at/all.md") {
		t.Errorf("un-indenting the transcript did not make its path judged:\n%s", report)
	}
}

// TestALinkThatGoesNowhereIsRefused. The trip case is a transposition in a
// file name rather than an invented link, because that is the one that gets
// written.
func TestALinkThatGoesNowhereIsRefused(t *testing.T) {
	f := green()
	f.files["README.md"] = strings.Replace(tree()["README.md"], "](docs/thing.md)", "](docs/thnig.md)", 1)

	report := refuse(t, f, "doc-links")
	if !strings.Contains(report, "README.md links to docs/thnig.md, which does not exist") {
		t.Errorf("the report does not name the document and the target:\n%s", report)
	}
}

// TestTheDocumentLegsSayHowMuchTheyJudged pins what the token rules admit. The
// counts are the only thing standing between a leg that resolves everything a
// document names and one whose extractor quietly stopped matching, which would
// pass every tree.
func TestTheDocumentLegsSayHowMuchTheyJudged(t *testing.T) {
	// main.go from one document, README.md and docs/ from the other. The
	// backticked .gitattributes is deliberately not among them: it carries no
	// slash and no extension, so nothing here can tell it from a word.
	if report := pass(t, green(), "doc-paths"); !strings.Contains(report, "Examined 3 path reference(s) in 2 tracked document(s)") {
		t.Errorf("the path leg does not report the references the fixture carries:\n%s", report)
	}
	// The two local links. The one leaving the repository is not counted,
	// because following it is the weekly leg's job and not this one's.
	if report := pass(t, green(), "doc-links"); !strings.Contains(report, "Examined 2 internal link(s) in 2 tracked document(s)") {
		t.Errorf("the link leg does not report the links the fixture carries:\n%s", report)
	}
}

// TestACheckedOutCarriageReturnDoesNotMoveTheWidthVerdict. The document legs
// judge the working tree and a checkout filter rewrites line endings on the way
// out, so without this every line on a machine that checked out CRLF would be
// one column wider than the server sees it and the whole tree would red for a
// byte no document carries. The leg that owns that byte reads the index and is
// unaffected, which the second half asserts.
func TestACheckedOutCarriageReturnDoesNotMoveTheWidthVerdict(t *testing.T) {
	f := green()
	f.files["docs/thing.md"] = strings.ReplaceAll(
		tree()["docs/thing.md"]+strings.Repeat("x", columns)+"\n", "\n", "\r\n")

	pass(t, f, "doc-format")

	f.answers["git ls-files --eol"] = answer{output: eolListing(map[string]string{"docs/thing.md": "i/crlf"})}
	refuse(t, f, "line-endings")
}

// TestAModuleFileTheTreeDoesNotRequireIsRefused. The trip case is the residue
// somebody leaves behind after removing the last use of a dependency: a
// requirement nothing imports, which every other leg here reports as fine
// because every import still resolves.
func TestAModuleFileTheTreeDoesNotRequireIsRefused(t *testing.T) {
	f := green()
	f.answers["go mod tidy -diff"] = answer{
		output: "diff current/go.mod tidy/go.mod\n" +
			"--- current/go.mod\n+++ tidy/go.mod\n@@ -1,5 +1,3 @@\n-require golang.org/x/text v0.3.8\n",
		err: errors.New("exit status 1"),
	}

	report := refuse(t, f, "lock")
	if !strings.Contains(report, "-require golang.org/x/text v0.3.8") {
		t.Errorf("the report does not carry what a tidy run would change:\n%s", report)
	}
	if !strings.Contains(report, "There is no lock file") {
		t.Errorf("the report does not say which state the tree is in, so a reader cannot tell an empty set from a checked one:\n%s", report)
	}
}

// TestTheLockLegSaysWhatTheLockFileHolds. An empty dependency set and a lock
// file nobody read print the same word otherwise, and the first is this tree's
// state today.
func TestTheLockLegSaysWhatTheLockFileHolds(t *testing.T) {
	f := green()
	f.answers["git ls-files go.sum"] = answer{output: "go.sum\n"}
	f.files["go.sum"] = "example.com/a v1.0.0 h1:x=\nexample.com/a v1.0.0/go.mod h1:y=\n"

	report := pass(t, f, "lock")
	if !strings.Contains(report, "The lock file holds 2 line(s)") {
		t.Errorf("the report does not say what the lock file holds:\n%s", report)
	}
}

// TestADocumentThatCannotBeReadFailsClosed. A tracked file that is not on disk
// is a tree nothing judged, and skipping it would let a leg report every
// document clean by finding none of them.
func TestADocumentThatCannotBeReadFailsClosed(t *testing.T) {
	for _, leg := range []string{"doc-format", "doc-paths", "doc-links", "encoding", "unicode"} {
		f := green()
		delete(f.files, "docs/thing.md")

		report := refuse(t, f, leg)
		if !strings.Contains(report, "Failing closed") {
			t.Errorf("leg %s does not say it failed closed on a file it could not read:\n%s", leg, report)
		}
	}
}

// TestLegNamesAreDistinct. The names are the contract with the workflow files,
// and two legs sharing one would make a step run something other than what it
// asked for.
func TestLegNamesAreDistinct(t *testing.T) {
	seen := map[string]bool{}
	for _, leg := range Legs() {
		if seen[leg.ID] {
			t.Errorf("two legs are called %s", leg.ID)
		}
		seen[leg.ID] = true
		if leg.ID == "" || leg.Title == "" || leg.Judge == nil {
			t.Errorf("leg %+v is missing a name, a title or a judgement", leg)
		}
	}
}
