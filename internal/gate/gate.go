// Package gate holds the procedure that decides whether a change passes.
//
// It exists because the legs of that procedure were shell inside ten workflow
// files, so the only thing that could run them all was the server and the loop
// for a wrapped line in a document was a round trip rather than a second. Issue
// #150 is where that was argued.
//
// The point is one procedure rather than two. A second implementation beside
// the workflows would drift against them, and a local run that disagrees with
// the server is worse than no local run at all, so the workflow steps call this
// by leg name instead of restating what the leg does.
//
// Nothing here starts a process. Every leg reaches the machine through the two
// functions in Env, which the entry point supplies from internal/system and a
// test supplies from its own table, so this package's suite decides what a
// command returned rather than running one. harness_test.go refuses os/exec
// outside the wiring and that refusal is what makes the split real.
package gate

import (
	"fmt"
	"io"
	"strings"
)

// Env is what a leg is allowed to reach outside this process.
type Env struct {
	// Run starts one program and returns what it wrote and whether it
	// succeeded. Both streams arrive together, because a tool that writes its
	// findings on one and its summary on the other reads wrong split apart.
	Run func(name string, args ...string) (string, error)

	// Look reports whether a program is on the path. A leg asks before it
	// decides it could not run, so that a missing tool is reported as a leg
	// that did not run rather than as a tree that failed.
	Look func(name string) bool
}

// Outcome is what a leg concluded. Three values rather than two: a leg that
// could not run is not a leg that passed, and a run carrying one of them
// covered less than the whole set.
type Outcome int

const (
	// Passed means the leg ran and found nothing.
	Passed Outcome = iota
	// Failed means the leg ran and refused what it found.
	Failed
	// NotRun means the leg could not run here, and Result.Reason says why.
	NotRun
)

// Result is what one leg reports.
type Result struct {
	Outcome Outcome
	// Examined is the sentence saying what the leg looked at. A run that
	// judged nothing and a run that judged everything and found nothing print
	// the same word otherwise.
	Examined string
	// Reason is why a leg did not run, and is empty otherwise.
	Reason string
	// Output is what the command wrote, kept for a leg that failed so the
	// report says what is wrong rather than only that something is.
	Output string
}

// Leg is one part of the gate, with the name a workflow step calls it by.
type Leg struct {
	// ID is what `ci` takes on the command line and what a workflow step
	// names. It is part of the contract with the workflow files.
	ID string
	// Title is the line a reader sees.
	Title string
	// Judge runs the leg.
	Judge func(Env) Result
}

// Legs returns the gate in the order it runs, which is the order a reader
// wants: what the module claims, what it resolves to, whether it builds, how it
// is written, what the analysers say, and finally whether it works.
func Legs() []Leg {
	return []Leg{
		{ID: "mod-verify", Title: "The module's dependencies are what they claim to be", Judge: modVerify},
		{ID: "deps", Title: "Every import resolves with dependency resolution switched off", Judge: deps},
		{ID: "build", Title: "Every package compiles", Judge: build},
		{ID: "gofmt", Title: "No Go file is one the formatter would change", Judge: gofmt},
		{ID: "vet", Title: "The toolchain's own correctness analyser is clean", Judge: vet},
		{ID: "staticcheck", Title: "The linter set this project chose is clean", Judge: staticcheck},
		{ID: "test-selection", Title: "The suite selects at least one test", Judge: testSelection},
		{ID: "test", Title: "The suite passes", Judge: test},
	}
}

// Uncovered is a leg of the gate this command does not run, with the reason.
// It is printed by every full run, because a leg that is silently absent turns
// a partial run into one a reader takes for the whole set.
type Uncovered struct {
	Name   string
	Reason string
}

// NotCovered returns the legs the server runs that this command does not. The
// list is data here rather than prose in a document, so the run prints what is
// missing instead of a reader trusting a paragraph to have been kept level with
// the workflow files.
func NotCovered() []Uncovered {
	return []Uncovered{
		{"Line endings and encoding", "written as git plumbing inside .github/workflows/text-hygiene.yml and not yet in this language"},
		{"Reject Trojan Source Unicode", "written as git plumbing inside .github/workflows/unicode-guard.yml and not yet in this language"},
		{"Documentation formatting, paths and internal links", "written as awk inside .github/workflows/doc-hygiene.yml and not yet in this language"},
		{"Documentation spelling", "needs a speller and a word list this tree does not carry, and takes them from the runner image"},
		{"Dependency lock", "written as shell inside .github/workflows/dependency-lock.yml and not yet in this language"},
		{"Refuse a run that is elevated or has a display", "judges the machine a run landed on rather than the repository, and stays in .github/workflows/test.yml for that reason"},
		{"Bill of materials", "needs a generator this tree does not carry"},
		{"Audit workflows (zizmor)", "needs a runner this tree does not carry"},
		{"Code scanning (CodeQL)", "runs on the server against an analysis database built there"},
		{"Scorecard analysis", "scores against a service outside this repository"},
		{"dependency-review", "reads an advisory database outside this repository"},
		{"DCO sign-off", "judges the commits a pull request carries, which a working tree does not have"},
		{"Sweep the default branch for a non-success run", "reads this repository's run history on the server"},
	}
}

// Run runs the legs in order and stops at the first failure, which is what a
// contributor wants: the second finding is usually the first one again.
//
// Naming a leg runs that one and requires it. A leg that could not run is a
// failure when it was asked for by name and a disclosure when it was not,
// because asking for it by name is what makes its absence the answer to a
// question somebody asked.
func Run(out io.Writer, env Env, only string) int {
	legs := Legs()
	if only != "" {
		leg, found := byID(legs, only)
		if !found {
			fmt.Fprintf(out, "FAIL  there is no gate leg called %q. The legs are %s.\n", only, strings.Join(ids(legs), ", "))
			return 1
		}
		legs = []Leg{leg}
	}

	passed, failed, skipped := 0, 0, 0
	for _, leg := range legs {
		result := leg.Judge(env)
		switch result.Outcome {
		case Passed:
			passed++
			fmt.Fprintf(out, "PASS  %s\n      %s\n", leg.ID, result.Examined)
		case NotRun:
			skipped++
			fmt.Fprintf(out, "NOT RUN  %s\n         %s\n", leg.ID, result.Reason)
		case Failed:
			failed++
			fmt.Fprintf(out, "FAIL  %s\n      %s\n", leg.ID, result.Examined)
			if strings.TrimSpace(result.Output) != "" {
				fmt.Fprintf(out, "--- what %s reported ---\n%s\n", leg.ID, strings.TrimRight(result.Output, "\n"))
			}
		}
		if result.Outcome == Failed {
			break
		}
	}

	fmt.Fprintf(out, "\n%d leg(s) passed, %d failed, %d did not run, out of the %d this command runs.\n",
		passed, failed, skipped, len(legs))

	if only != "" {
		if failed != 0 || skipped != 0 {
			return 1
		}
		return 0
	}

	if passed+failed+skipped < len(Legs()) {
		fmt.Fprintf(out, "The run stopped early, so the legs after the failure above were not reached.\n")
	}
	fmt.Fprintf(out, "\nWhat this command does not run, and why:\n")
	for _, leg := range NotCovered() {
		fmt.Fprintf(out, "  %s: %s\n", leg.Name, leg.Reason)
	}
	if failed != 0 {
		return 1
	}
	return 0
}

// byID finds a leg by the name a workflow step calls it.
func byID(legs []Leg, id string) (Leg, bool) {
	for _, leg := range legs {
		if leg.ID == id {
			return leg, true
		}
	}
	return Leg{}, false
}

// ids is the leg names, for the message a wrong one produces.
func ids(legs []Leg) []string {
	names := make([]string, 0, len(legs))
	for _, leg := range legs {
		names = append(names, leg.ID)
	}
	return names
}

// lines splits command output into the lines that carry something. Several
// legs count what a listing returned, and a trailing newline would otherwise
// count as an entry.
func lines(output string) []string {
	kept := []string{}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			kept = append(kept, strings.TrimSpace(line))
		}
	}
	return kept
}

// failure is a leg that ran and refused.
func failure(examined string, output string) Result {
	return Result{Outcome: Failed, Examined: examined, Output: output}
}

// packages lists the packages of this module with dependency resolution
// switched off, so a file importing something the module file does not carry
// fails here rather than being resolved quietly.
func packages(env Env) ([]string, string, error) {
	output, err := env.Run("go", "list", "-mod=readonly", "./...")
	if err != nil {
		return nil, output, err
	}
	return lines(output), output, nil
}

// tracked lists what the repository stores. The file list comes from git
// rather than from a directory walk, because a build directory or an editor
// backup left in the checkout is not what is being judged.
func tracked(env Env, pattern string) ([]string, string, error) {
	output, err := env.Run("git", "ls-files", pattern)
	if err != nil {
		return nil, output, err
	}
	return lines(output), output, nil
}

func modVerify(env Env) Result {
	modules, output, err := func() ([]string, string, error) {
		out, err := env.Run("go", "list", "-mod=readonly", "-m", "all")
		if err != nil {
			return nil, out, err
		}
		return lines(out), out, nil
	}()
	if err != nil {
		return failure("The module graph could not be listed.", output)
	}
	// The first entry is this module. Zero requirements is a real answer here
	// rather than an empty selection: the module carries no dependency, and
	// saying so is the point.
	requirements := len(modules) - 1
	if requirements < 0 {
		requirements = 0
	}
	examined := fmt.Sprintf("Examined %d module requirement(s).", requirements)
	if requirements == 0 {
		examined += " This module carries none."
	}
	verify, err := env.Run("go", "mod", "verify")
	if err != nil {
		return failure(examined, verify)
	}
	return Result{Outcome: Passed, Examined: examined}
}

func deps(env Env) Result {
	output, err := env.Run("go", "list", "-mod=readonly", "-deps", "./...")
	if err != nil {
		return failure("Every import of every package was resolved.", output)
	}
	reached := lines(output)
	if len(reached) == 0 {
		return failure("No package was resolved. Failing closed rather than reporting a tree that resolves.", output)
	}
	return Result{Outcome: Passed, Examined: fmt.Sprintf("Examined %d package(s) reachable from this module, the standard library included.", len(reached))}
}

func build(env Env) Result {
	found, listing, err := packages(env)
	if err != nil {
		return failure("The packages of this module could not be listed.", listing)
	}
	if len(found) == 0 {
		return failure("No package was compiled. Failing closed rather than reporting a tree that builds.", listing)
	}
	examined := fmt.Sprintf("Examined %d package(s) of this module.", len(found))
	output, err := env.Run("go", "build", "-mod=readonly", "./...")
	if err != nil {
		return failure(examined, output)
	}
	return Result{Outcome: Passed, Examined: examined}
}

func gofmt(env Env) Result {
	files, listing, err := tracked(env, "*.go")
	if err != nil {
		return failure("The tracked Go files could not be listed.", listing)
	}
	if len(files) == 0 {
		return failure("No tracked Go file was examined. Failing closed rather than reporting a formatted tree.", listing)
	}
	examined := fmt.Sprintf("Examined %d tracked Go file(s).", len(files))
	// -l lists the files that differ from their formatted form. It is silent
	// and exits 0 on a clean tree, so the list rather than the exit code is
	// what decides this leg.
	output, err := env.Run("gofmt", append([]string{"-l"}, files...)...)
	if err != nil {
		return failure(examined, output)
	}
	if changed := lines(output); len(changed) != 0 {
		// What the formatter would change rather than only which files it
		// would touch. A list of names sends somebody back to run the tool
		// themselves, which is the round trip this command exists to remove.
		diff, _ := env.Run("gofmt", append([]string{"-d"}, changed...)...)
		return failure(examined, "the formatter would change:\n"+strings.Join(changed, "\n")+"\n--- what it would change them to ---\n"+diff)
	}
	return Result{Outcome: Passed, Examined: examined}
}

func vet(env Env) Result {
	found, listing, err := packages(env)
	if err != nil {
		return failure("The packages of this module could not be listed.", listing)
	}
	if len(found) == 0 {
		return failure("No package was analysed. Failing closed rather than reporting a clean tree.", listing)
	}
	examined := fmt.Sprintf("Examined %d package(s) of this module.", len(found))
	output, err := env.Run("go", "vet", "-mod=readonly", "./...")
	if err != nil {
		return failure(examined, output)
	}
	return Result{Outcome: Passed, Examined: examined}
}

// staticcheckInstall is the command that puts the linter where this leg can
// find it, at the version this project pinned. It is quoted at the reader
// rather than run for them: installing a tool is a decision about somebody's
// machine and not this command's to take.
const staticcheckInstall = "go install honnef.co/go/tools/cmd/staticcheck@v0.7.0"

func staticcheck(env Env) Result {
	if !env.Look("staticcheck") {
		return Result{
			Outcome: NotRun,
			Reason:  "staticcheck is not on the path. It is the one gate tool the toolchain does not ship. Install it with " + staticcheckInstall + " and run this again.",
		}
	}
	found, listing, err := packages(env)
	if err != nil {
		return failure("The packages of this module could not be listed.", listing)
	}
	if len(found) == 0 {
		return failure("No package was linted. Failing closed rather than reporting a clean tree.", listing)
	}
	// The enabled set is staticcheck.conf's, read from the module root, so this
	// command and a bare `staticcheck ./...` judge the same checks. No -checks
	// flag is passed on purpose. What is printed is which set that turned out
	// to be, because a linter that silently enabled nothing passes everything.
	version, _ := env.Run("staticcheck", "-version")
	available, _ := env.Run("staticcheck", "-list-checks")
	examined := fmt.Sprintf("Examined %d package(s) of this module with %s, %d check(s) available and the enabled set fixed by staticcheck.conf.",
		len(found), strings.TrimSpace(version), len(lines(available)))
	output, err := env.Run("staticcheck", "./...")
	if err != nil {
		return failure(examined, output)
	}
	return Result{Outcome: Passed, Examined: examined}
}

// testSelection is the leg that makes every later green mean something. The
// toolchain reports a package with no test file as a success, so a change that
// moves the whole suite somewhere this command does not look leaves the run
// green and the suite unrun. The selection is counted before anything is
// executed, and zero fails.
func testSelection(env Env) Result {
	found, listing, err := packages(env)
	if err != nil {
		return failure("The packages of this module could not be listed.", listing)
	}
	if len(found) == 0 {
		return failure("No package was examined. Failing closed rather than reporting a suite that selected nothing.", listing)
	}
	// -list takes an expression and prints the names it would select without
	// running any of them. A listing that could not be produced is a different
	// failure from a suite of zero, and reading one as the other would report
	// the wrong repair.
	output, err := env.Run("go", "test", "-mod=readonly", "-list", ".*", "./...")
	if err != nil {
		return failure("The toolchain could not list the tests. Failing closed rather than reading that as a suite of zero.", output)
	}
	selected := 0
	for _, line := range lines(output) {
		if strings.HasPrefix(line, "Test") || strings.HasPrefix(line, "Example") || strings.HasPrefix(line, "Fuzz") {
			selected++
		}
	}
	examined := fmt.Sprintf("Examined %d package(s) of this module, which select %d test(s).", len(found), selected)
	if selected == 0 {
		return failure("The suite selected no test at all. A run that selected nothing is not a run that passed; failing closed.", output)
	}
	return Result{Outcome: Passed, Examined: examined}
}

func test(env Env) Result {
	found, listing, err := packages(env)
	if err != nil {
		return failure("The packages of this module could not be listed.", listing)
	}
	if len(found) == 0 {
		return failure("No package was tested. Failing closed rather than reporting a passing suite.", listing)
	}
	// -count=1 defeats the result cache, so a green here is a run rather than
	// a memory of one. -failfast stops a package at its first failing test,
	// which is the one somebody repairs; the ones after it are frequently
	// consequences of it. -v is what makes the count below derived from what
	// ran rather than asserted.
	output, err := env.Run("go", "test", "-mod=readonly", "-count=1", "-failfast", "-v", "./...")
	ran := 0
	for _, line := range lines(output) {
		if strings.HasPrefix(line, "--- PASS") || strings.HasPrefix(line, "--- FAIL") || strings.HasPrefix(line, "--- SKIP") {
			ran++
		}
	}
	examined := fmt.Sprintf("Examined %d package(s) of this module, in which %d test(s) ran.", len(found), ran)
	if err != nil {
		return failure(examined, output)
	}
	if ran == 0 {
		return failure("No test ran. Failing closed rather than reading a toolchain that selected nothing as a suite that passed.", output)
	}
	return Result{Outcome: Passed, Examined: examined}
}
