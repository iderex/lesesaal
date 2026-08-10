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
// Nothing here starts a process and nothing here opens a file. Every leg
// reaches the machine through the functions in Env, which the entry point
// supplies from internal/system and a test supplies from its own table, so this
// package's suite decides what a command returned and what a file held rather
// than running one and reading the other. harness_test.go refuses os/exec
// outside the wiring and that refusal is what makes the split real.
package gate

import (
	"fmt"
	"io"
	"path"
	"strings"
	"unicode/utf8"
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

	// Read returns the bytes of one file in the working tree. The legs that
	// judge text need the bytes of every tracked file, and asking git for them
	// one blob at a time costs a process each; the measurement that decided
	// this is at system.Read.
	Read func(name string) (string, error)
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
// wants: what the tree is stored as, what its documents say, what the module
// claims, what it resolves to, whether it builds, how the code is written, what
// the analysers say, and finally whether it works. The cheap legs come first on
// purpose, because the run stops at the first failure and a contributor who
// wrapped a line at 90 columns should not wait out a compile to hear about it.
func Legs() []Leg {
	return []Leg{
		{ID: "line-endings", Title: "Every tracked text file is stored with LF", Judge: lineEndings},
		{ID: "encoding", Title: "Every tracked text file decodes as UTF-8", Judge: encoding},
		{ID: "unicode", Title: "No tracked text file carries a bidirectional or invisible control character", Judge: unicodeGuard},
		{ID: "doc-format", Title: "Every tracked document is formatted for readable diffs", Judge: docFormat},
		{ID: "doc-paths", Title: "Every path a document names resolves", Judge: docPaths},
		{ID: "doc-links", Title: "Every internal link a document carries resolves", Judge: docLinks},
		{ID: "mod-verify", Title: "The module's dependencies are what they claim to be", Judge: modVerify},
		{ID: "lock", Title: "The module file and the lock file are what the tree's imports require", Judge: lock},
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
		{"Documentation spelling", "needs a speller and a word list this tree does not carry, and takes them from the runner image"},
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
// An empty pattern means the whole tree, spelled as the pathspec git wants for
// it, because the empty string is the one thing git refuses to read as "all
// paths".
func tracked(env Env, pattern string) ([]string, string, error) {
	if pattern == "" {
		pattern = "."
	}
	output, err := env.Run("git", "ls-files", pattern)
	if err != nil {
		return nil, output, err
	}
	return lines(output), output, nil
}

// stored is one tracked file with the line ending git holds it under.
type stored struct {
	// eol is the index attribute `git ls-files --eol` prints first: i/lf,
	// i/crlf, i/mixed, i/none for a file with no line ending at all, and
	// i/-text for a file git treats as binary.
	eol string
	// path is what the repository stores the file as, slashed on every
	// platform because git prints it that way.
	path string
}

// storedFiles lists every tracked file with the line ending of its blob. It is
// the one place "a tracked text file" is decided, and three legs share it, so
// the tree cannot be text for one leg and binary for the next.
func storedFiles(env Env) ([]stored, string, error) {
	output, err := env.Run("git", "ls-files", "--eol")
	if err != nil {
		return nil, output, err
	}
	found := []stored{}
	for _, line := range lines(output) {
		// The attribute fields are padded with spaces and the path is
		// separated from them by a tab, so the last tab is the split.
		tab := strings.LastIndex(line, "\t")
		fields := strings.Fields(line)
		if tab < 0 || len(fields) == 0 {
			continue
		}
		found = append(found, stored{eol: fields[0], path: strings.TrimSpace(line[tab+1:])})
	}
	return found, output, nil
}

// text keeps the tracked files git does not treat as binary.
func text(files []stored) []stored {
	kept := []stored{}
	for _, file := range files {
		if file.eol == "i/-text" {
			continue
		}
		kept = append(kept, file)
	}
	return kept
}

// WHICH COPY OF A FILE THESE LEGS JUDGE, because there are three and they
// disagree. The line ending a file is STORED with is the blob's property and
// nothing else can answer it, so lineEndings reads the index and reads it from
// git. Everything below it judges the working tree, which is the same copy
// gofmt, the compiler and the suite already judge: a contributor who fixes a
// document, runs this and is told it is still wrong because the fix is not
// staged would stop running it.
//
// The one place that would make the two disagree is a checkout filter, which
// rewrites line endings on the way out and would otherwise make every line of
// every document one column wider on a machine that checked out CRLF. That is
// lineEndings' subject rather than these legs', so records below drops the
// carriage return and lets the leg that owns the question answer it.

// records splits a file the way a line-oriented tool reads it. A file whose
// last byte is a newline does not end with an empty record, because every leg
// below counts what it examined and an invented final record would make a
// clean document report one line more than it has.
func records(content string) []string {
	if content == "" {
		return nil
	}
	read := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	for number, line := range read {
		read[number] = strings.TrimSuffix(line, "\r")
	}
	return read
}

// lineEndings refuses a tracked text file whose blob is not stored with LF.
// .gitattributes declares that the tree is stored with LF, and a declaration
// changes nothing about bytes already in the repository: a contributor can put
// anything into a blob with git plumbing. The declaration and this leg are two
// different things and only this one bites.
func lineEndings(env Env) Result {
	files, listing, err := storedFiles(env)
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	if len(files) == 0 {
		return failure("No tracked file was examined. Failing closed rather than reporting a tree stored with LF.", listing)
	}
	examined := text(files)
	if len(examined) == 0 {
		return failure("No tracked text file was examined. Failing closed rather than reporting a tree stored with LF.", listing)
	}
	sentence := fmt.Sprintf("Examined %d tracked text file(s) of %d tracked file(s).", len(examined), len(files))
	found := []string{}
	for _, file := range examined {
		if file.eol == "i/crlf" || file.eol == "i/mixed" {
			found = append(found, fmt.Sprintf("FAIL  %s is stored as %s", file.path, strings.TrimPrefix(file.eol, "i/")))
		}
	}
	if len(found) != 0 {
		return failure(sentence, strings.Join(found, "\n")+
			"\nThis tree is stored with LF; see .gitattributes. Re-add them with 'git add --renormalize .'")
	}
	return Result{Outcome: Passed, Examined: sentence}
}

// encoding refuses a tracked text file whose blob does not decode as UTF-8.
// ASCII is a subset and passes, so what this catches is a file saved by an
// editor that chose the machine's own code page, which is invisible in a diff
// until the byte reaches somebody with a different one.
func encoding(env Env) Result {
	files, listing, err := storedFiles(env)
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	examined := text(files)
	if len(examined) == 0 {
		return failure("No tracked text file was examined. Failing closed rather than reporting a tree that decodes.", listing)
	}
	sentence := fmt.Sprintf("Examined %d tracked text file(s) for UTF-8 validity.", len(examined))
	found := []string{}
	for _, file := range examined {
		content, err := env.Read(file.path)
		if err != nil {
			return failure("A tracked text file could not be read. Failing closed rather than skipping it.", err.Error())
		}
		if !utf8.ValidString(content) {
			found = append(found, "FAIL  "+file.path+" does not decode as UTF-8")
		}
	}
	if len(found) != 0 {
		return failure(sentence, strings.Join(found, "\n")+
			"\nThis tree is UTF-8; convert the file rather than adding an exception.")
	}
	return Result{Outcome: Passed, Examined: sentence}
}

// deceptive is the set of characters that make a file read differently from
// how it behaves: the bidirectional controls of Trojan Source, CVE-2021-42574,
// and the zero-width characters that hide inside a word. Accents, em dashes and
// every other benign non-ASCII character are not in the set.
//
// U+FEFF is deliberately absent. A leading byte order mark is legitimate, the
// reordering attack does not rely on one, and refusing it would red a tree for
// something that is not the defect.
var deceptive = map[rune]string{
	0x061C: "ARABIC LETTER MARK",
	0x200B: "ZERO WIDTH SPACE",
	0x200C: "ZERO WIDTH NON-JOINER",
	0x200D: "ZERO WIDTH JOINER",
	0x200E: "LEFT-TO-RIGHT MARK",
	0x200F: "RIGHT-TO-LEFT MARK",
	0x202A: "LEFT-TO-RIGHT EMBEDDING",
	0x202B: "RIGHT-TO-LEFT EMBEDDING",
	0x202C: "POP DIRECTIONAL FORMATTING",
	0x202D: "LEFT-TO-RIGHT OVERRIDE",
	0x202E: "RIGHT-TO-LEFT OVERRIDE",
	0x2060: "WORD JOINER",
	0x2066: "LEFT-TO-RIGHT ISOLATE",
	0x2067: "RIGHT-TO-LEFT ISOLATE",
	0x2068: "FIRST STRONG ISOLATE",
	0x2069: "POP DIRECTIONAL ISOLATE",
}

// unicodeGuard refuses a character from that set anywhere in tracked text. It
// names the file, the line and the codepoint, because the whole property of
// these characters is that a reader cannot see them, so a report saying only
// which file is one nobody can act on.
func unicodeGuard(env Env) Result {
	files, listing, err := storedFiles(env)
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	examined := text(files)
	if len(examined) == 0 {
		return failure("No tracked text file was examined. Failing closed rather than reporting a tree with nothing hidden in it.", listing)
	}
	sentence := fmt.Sprintf("Examined %d tracked text file(s) against %d refused character(s).", len(examined), len(deceptive))
	found := []string{}
	for _, file := range examined {
		content, err := env.Read(file.path)
		if err != nil {
			return failure("A tracked text file could not be read. Failing closed rather than skipping it.", err.Error())
		}
		for number, line := range records(content) {
			for _, character := range line {
				name, refused := deceptive[character]
				if !refused {
					continue
				}
				found = append(found, fmt.Sprintf("FAIL  %s:%d carries U+%04X (%s)", file.path, number+1, character, name))
			}
		}
	}
	if len(found) != 0 {
		return failure(sentence, strings.Join(found, "\n")+
			"\nThese characters make source render differently from how it executes, which is the attack this refuses. Remove them.")
	}
	return Result{Outcome: Passed, Examined: sentence}
}

// columns is the width this tree wraps prose at. It is the tree's own limit
// rather than a number chosen here: it was measured over every tracked document
// outside code blocks on the day the check landed, and the one line that was
// over it was rewrapped in the same change.
const columns = 81

// documents keeps the tracked documents out of the tracked file list. They are
// what all three document legs judge and nothing else: the workflow files, the
// certificate and the attributes file carry prose in comments and no leg here
// reads it. It filters a list the caller already holds rather than asking git a
// second question, because every question costs a process.
func documents(all []string) []string {
	kept := []string{}
	for _, file := range all {
		if strings.HasSuffix(file, ".md") {
			kept = append(kept, file)
		}
	}
	return kept
}

// docFormat refuses a document that is not written the way this tree stores
// documents. The convention is hard wrapping, so a one-word edit produces a
// one-line diff instead of reflowing a paragraph and hiding the change inside
// it.
//
// Code blocks are exempt from the width in both spellings, fenced and indented.
// A pasted command and its output have to stay byte-exact, and wrapping one
// would make it a command nobody can run. They are not exempt from the tab, the
// trailing space or the second blank line, because none of those is load
// bearing in a transcript either.
func docFormat(env Env) Result {
	all, listing, err := tracked(env, "")
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	files := documents(all)
	if len(files) == 0 {
		return failure("No tracked document was examined. Failing closed rather than reporting a clean tree.", listing)
	}
	found := []string{}
	counted := 0
	for _, file := range files {
		content, err := env.Read(file)
		if err != nil {
			return failure("A tracked document could not be read. Failing closed rather than skipping it.", err.Error())
		}
		if content == "" {
			found = append(found, "FAIL  "+file+" is empty")
			continue
		}
		if !strings.HasSuffix(content, "\n") {
			found = append(found, "FAIL  "+file+" does not end with a newline")
		}
		read := records(content)
		if read[len(read)-1] == "" {
			found = append(found, "FAIL  "+file+" ends with a blank line")
		}
		infence := false
		// The line before this one, so two blank lines in a row can be told
		// from one. The first line of a file has no predecessor and the value
		// below is deliberately not the empty string, which would make a
		// document opening on a blank line report a second one.
		previous := "start of file"
		for number, line := range read {
			where := fmt.Sprintf("%s:%d", file, number+1)
			if strings.HasPrefix(line, "```") {
				infence = !infence
				previous = line
				continue
			}
			incode := infence || strings.HasPrefix(line, "    ")
			if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
				found = append(found, "FAIL  "+where+" has trailing whitespace")
			}
			if strings.Contains(line, "\t") {
				found = append(found, "FAIL  "+where+" contains a tab")
			}
			if line == "" && previous == "" {
				found = append(found, "FAIL  "+where+" is the second of two blank lines")
			}
			// Columns rather than bytes, so a document carrying a non-ASCII
			// character is judged by what a reader sees rather than by how many
			// bytes it took to store it.
			if width := utf8.RuneCountInString(line); !incode && width > columns {
				found = append(found, fmt.Sprintf("FAIL  %s is %d columns, over the %d this tree wraps at", where, width, columns))
			}
			counted++
			previous = line
		}
	}
	sentence := fmt.Sprintf("Examined %d line(s) in %d tracked document(s) for formatting.", counted, len(files))
	if counted == 0 {
		return failure("No line was examined. Failing closed rather than reporting documents that are formatted.", strings.Join(found, "\n"))
	}
	if len(found) != 0 {
		return failure(sentence, strings.Join(found, "\n")+
			"\nWrap prose at 81 columns, drop trailing whitespace and tabs, and end the file with one newline.")
	}
	return Result{Outcome: Passed, Examined: sentence}
}

// docPaths refuses a document naming a path that does not exist. This is the
// leg that earns the three: this plan depends on documents pointing at each
// other, at the workflows and at the attributes file, and a pointer to
// something renamed is invisible until somebody follows it.
//
// It resolves against what git tracks rather than against what is in the
// checkout, which is stricter in one direction on purpose: a document may not
// point at a file that exists on the author's machine and in no clone.
func docPaths(env Env) Result {
	all, listing, err := tracked(env, "")
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	files := documents(all)
	if len(files) == 0 {
		return failure("No tracked document was examined. Failing closed rather than reporting a clean tree.", listing)
	}
	exists := reachable(all)
	found := []string{}
	counted := 0
	for _, file := range files {
		content, err := env.Read(file)
		if err != nil {
			return failure("A tracked document could not be read. Failing closed rather than skipping it.", err.Error())
		}
		infence := false
		for _, line := range records(content) {
			if strings.HasPrefix(line, "```") {
				infence = !infence
				continue
			}
			// A pasted transcript names paths in other repositories and in
			// other people's machines, and this tree cannot resolve those.
			if infence || strings.HasPrefix(line, "    ") {
				continue
			}
			for _, token := range backticked(line) {
				if !pathLike(token) {
					continue
				}
				counted++
				if resolves(exists, file, token) {
					continue
				}
				found = append(found, "FAIL  "+file+" names "+token+", which does not exist")
			}
		}
	}
	sentence := fmt.Sprintf("Examined %d path reference(s) in %d tracked document(s).", counted, len(files))
	if len(found) != 0 {
		return failure(sentence, strings.Join(found, "\n")+
			"\nName no path you do not intend to resolve.")
	}
	return Result{Outcome: Passed, Examined: sentence}
}

// docLinks refuses a link that goes nowhere. A link in a document is a promise
// that the thing at the other end exists.
//
// The fragment on a local target is stripped and not checked. Following it
// means deriving the anchor slugs of somebody else's renderer, which is a guess
// rather than a check, and a wrong guess reds a document that is correct.
func docLinks(env Env) Result {
	all, listing, err := tracked(env, "")
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	files := documents(all)
	if len(files) == 0 {
		return failure("No tracked document was examined. Failing closed rather than reporting a clean tree.", listing)
	}
	exists := reachable(all)
	found := []string{}
	counted := 0
	for _, file := range files {
		content, err := env.Read(file)
		if err != nil {
			return failure("A tracked document could not be read. Failing closed rather than skipping it.", err.Error())
		}
		infence := false
		for _, line := range records(content) {
			if strings.HasPrefix(line, "```") {
				infence = !infence
				continue
			}
			if infence {
				continue
			}
			for _, target := range linkTargets(line) {
				if target == "" || strings.HasPrefix(target, "#") ||
					strings.HasPrefix(target, "http:") || strings.HasPrefix(target, "https:") ||
					strings.HasPrefix(target, "mailto:") {
					continue
				}
				counted++
				local := target
				if hash := strings.Index(local, "#"); hash >= 0 {
					local = local[:hash]
				}
				if local == "" || resolves(exists, file, local) {
					continue
				}
				found = append(found, "FAIL  "+file+" links to "+target+", which does not exist")
			}
		}
	}
	sentence := fmt.Sprintf("Examined %d internal link(s) in %d tracked document(s).", counted, len(files))
	if len(found) != 0 {
		return failure(sentence, strings.Join(found, "\n")+
			"\nA link in a document is a promise that the thing at the other end exists.")
	}
	return Result{Outcome: Passed, Examined: sentence}
}

// reachable is every path a document is allowed to name: the tracked files and
// the directories they sit in, because a document naming a directory is naming
// something that exists.
func reachable(all []string) map[string]bool {
	found := make(map[string]bool, len(all)*2)
	for _, file := range all {
		found[file] = true
		for dir := path.Dir(file); dir != "." && dir != "/"; dir = path.Dir(dir) {
			found[dir] = true
		}
	}
	return found
}

// resolves reports whether a token names something, relative to the directory
// of the document naming it first and then from the repository root, which is
// how the documents under docs/ already name their siblings.
func resolves(exists map[string]bool, from string, token string) bool {
	candidates := []string{token}
	if dir := path.Dir(from); dir != "." {
		candidates = append(candidates, dir+"/"+token)
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSuffix(candidate, "/")
		if candidate == "" {
			continue
		}
		if exists[path.Clean(candidate)] {
			return true
		}
	}
	return false
}

// backticked returns what a line writes between backticks, in order. A pair
// with nothing between them yields no token and the closing backtick is offered
// again as an opening one, which is what a line like “a“ requires.
func backticked(line string) []string {
	tokens := []string{}
	for {
		open := strings.Index(line, "`")
		if open < 0 {
			return tokens
		}
		rest := line[open+1:]
		shut := strings.Index(rest, "`")
		if shut < 0 {
			return tokens
		}
		if shut == 0 {
			line = rest
			continue
		}
		tokens = append(tokens, rest[:shut])
		line = rest[shut+1:]
	}
}

// extensions are the suffixes that make a backticked word a path even when it
// carries no slash.
var extensions = []string{".md", ".yml", ".yaml", ".json", ".toml", ".go", ".txt", ".sum", ".mod"}

// pathLike decides whether a backticked token is a path this tree should be
// able to resolve. A token holding a character no path here uses is a command,
// an identifier or a field name, and judging one of those as a missing file is
// how a check earns being ignored.
func pathLike(token string) bool {
	if token == "" {
		return false
	}
	for position := 0; position < len(token); position++ {
		character := token[position]
		alphanumeric := character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9'
		switch {
		case alphanumeric, character == '_', character == '.':
		case position != 0 && (character == '/' || character == '-'):
		default:
			return false
		}
	}
	if strings.Contains(token, "/") {
		return true
	}
	for _, extension := range extensions {
		if strings.HasSuffix(token, extension) {
			return true
		}
	}
	return false
}

// linkTargets returns what a line writes inside the parentheses of a markdown
// link, in order.
func linkTargets(line string) []string {
	targets := []string{}
	for {
		open := strings.Index(line, "](")
		if open < 0 {
			return targets
		}
		rest := line[open+2:]
		shut := strings.Index(rest, ")")
		if shut < 0 {
			return targets
		}
		targets = append(targets, rest[:shut])
		line = rest[shut+1:]
	}
}

// lock refuses a module file or a lock file that is not what the tree's imports
// require, in either direction: a requirement nothing imports, an import
// nothing requires, or a lock file that does not cover the graph.
//
// It reports what a tidy run WOULD change and changes nothing, which is the
// whole point. A leg that repaired the module file would let a drifted one land
// and then quietly correct it, and nobody would ever see the drift.
func lock(env Env) Result {
	graph, output, err := func() ([]string, string, error) {
		out, err := env.Run("go", "list", "-mod=readonly", "-m", "all")
		if err != nil {
			return nil, out, err
		}
		return lines(out), out, nil
	}()
	if err != nil {
		return failure("The module graph could not be listed.", output)
	}
	if len(graph) == 0 {
		return failure("The module graph came out empty, which not even a module with no requirement does. Failing closed rather than judging nothing.", output)
	}
	// Which state the tree is in, so a green run is not read as "the lock file
	// was checked" on a tree that has none. The toolchain writes the lock file
	// the moment a requirement needs one, so its absence and an empty
	// requirement set are the same fact rather than two.
	sentence := fmt.Sprintf("Examined a module graph of %d entry(s), this module included.", len(graph))
	held, listing, err := tracked(env, "go.sum")
	if err != nil {
		return failure("The tree could not be asked whether it holds a lock file.", listing)
	}
	if len(held) == 0 {
		sentence += " There is no lock file, because nothing outside this module is required yet."
	} else {
		content, err := env.Read("go.sum")
		if err != nil {
			return failure("The lock file is tracked and could not be read. Failing closed rather than reporting it absent.", err.Error())
		}
		sentence += fmt.Sprintf(" The lock file holds %d line(s).", len(lines(content)))
	}
	diff, err := env.Run("go", "mod", "tidy", "-diff")
	if err != nil {
		return failure(sentence, diff+
			"\nThe diff above is what 'go mod tidy' would change; run it and commit the result rather than letting a build resolve around it.")
	}
	return Result{Outcome: Passed, Examined: sentence}
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
