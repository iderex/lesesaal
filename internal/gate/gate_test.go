package gate

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// The suite runs no command. Every leg reaches the machine through Env, so a
// test decides what `go build` returned instead of building, which is what
// keeps this package inside the rule that the unit suite has no subprocess, no
// network and no dependency.

// fake answers a command from a table. A command the table does not name
// succeeds with the empty string, so a test writes down only the commands it
// is actually about.
type fake struct {
	answers map[string]answer
	tools   map[string]bool
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

func (f *fake) env() Env { return Env{Run: f.run, Look: f.look} }

// green is a machine where everything the legs ask for answers the way a clean
// tree answers. Each test spoils exactly one entry, so the failure is the
// spoiling rather than the setup.
func green() *fake {
	return &fake{
		tools: map[string]bool{"staticcheck": true},
		answers: map[string]answer{
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
	if want := "8 leg(s) passed, 0 failed, 0 did not run"; !strings.Contains(report, want) {
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
		"gofmt": "git ls-files *.go",
		"build": "go list -mod=readonly ./...",
		"deps":  "go list -mod=readonly -deps ./...",
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
