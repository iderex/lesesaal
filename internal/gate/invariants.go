package gate

import (
	"fmt"
	"regexp"
	"strings"
)

// An invariant is a rule of this project that is a plain text fact about the
// tree. Issue #94 is where the class was argued: every rule in it is one a
// person currently has to remember, and a rule nothing refuses is a rule that
// holds until the first person who has not read the document that states it.
//
// ADDING ONE IS ONE FILE. A rule is an entry in invariantRules and the
// function it names, both here. The leg picks it up with no dispatch line, no
// workflow step and no second place to keep level. What it owes is what every
// guard in this package owes: a near miss in invariants_test.go, watched
// failing with the rule taken out.
//
// WHAT DOES NOT BELONG HERE, and both halves have cost something already. A
// rule with no subject in this tree examines nothing, and a pattern that
// matches nothing reads exactly like a clean tree, so an empty examination is
// refused below rather than reported as a pass. And a rule harness_test.go
// already refuses does not get a second refusal here: that file reads the
// source with the toolchain's own parser rather than as text, it is the
// stronger of the two, and a second implementation beside it would drift the
// way the workflow shell drifted against everything else before #150.
type invariant struct {
	// id names the rule in a refusal, so somebody who trips one can search
	// for it rather than guess which paragraph it came from.
	id string
	// rule is the sentence a refusal quotes back.
	rule string
	// decided is where the rule was argued, because a contributor who trips a
	// rule needs to be able to disagree with it somewhere.
	decided string
	// judge reads the tree and says what it refused and how much it read.
	judge func(env Env, all []string) (verdict, error)
}

// verdict is what one invariant found.
type verdict struct {
	// refused is one line per violation, each naming where it is.
	refused []string
	// examined is how many of subject the rule read.
	examined int
	// subject is what those were.
	subject string
}

// invariantRules is the whole set. Two rules rather than the seven #94 lists,
// and the four that are absent are absent for reasons written in that issue
// rather than forgotten: two are refused by harness_test.go already, one has
// to name the directory the configuration loader lands in and #82 has not
// chosen it, and three have no asset, no log call and no operator-facing
// message in this tree to examine.
func invariantRules() []invariant {
	return []invariant{
		{
			id:      "suppression-names-a-check-and-a-reason",
			rule:    "an analyser suppression names the check identifier it silences and says why it is right here",
			decided: "staticcheck.conf, for issue #22",
			judge:   suppressions,
		},
		{
			id:      "check-listing-reproduces-the-workflows",
			rule:    "a document pasting the check-run names reproduces the command it quotes above them",
			decided: "issue #174, and the closing section of test/gate-refusals.md",
			judge:   checkListings,
		},
	}
}

// invariants refuses a tree that breaks one of the rules above.
//
// It fails closed twice over. A rule that examined nothing is a failure rather
// than a pass, because a rule pointed at a subject that has moved away reports
// exactly what a clean tree reports. And a set with no rules in it is a
// failure for the same reason, which is the state this leg would reach if
// somebody emptied the table rather than fixing what it caught.
func invariants(env Env) Result {
	all, listing, err := tracked(env, "")
	if err != nil {
		return failure("The tracked files could not be listed.", listing)
	}
	rules := invariantRules()
	if len(rules) == 0 {
		return failure("No invariant was examined, because none is declared. Failing closed rather than reporting a clean tree.", "")
	}

	found := []string{}
	parts := []string{}
	for _, rule := range rules {
		got, err := rule.judge(env, all)
		if err != nil {
			return failure(fmt.Sprintf("The invariant %s could not be examined.", rule.id), err.Error())
		}
		parts = append(parts, fmt.Sprintf("%d %s under %s", got.examined, got.subject, rule.id))
		if got.examined == 0 {
			found = append(found, fmt.Sprintf("FAIL  %s examined no %s, so a tree that holds this rule and a rule looking at nothing report the same thing. Failing closed.\n      The rule is that %s. Decided in %s.", rule.id, got.subject, rule.rule, rule.decided))
			continue
		}
		for _, refusal := range got.refused {
			found = append(found, fmt.Sprintf("FAIL  %s\n      The rule is that %s. Decided in %s.", refusal, rule.rule, rule.decided))
		}
	}

	examined := fmt.Sprintf("Examined %d invariant(s): %s.", len(rules), strings.Join(parts, ", "))
	if len(found) > 0 {
		return failure(examined, strings.Join(found, "\n"))
	}
	return Result{Outcome: Passed, Examined: examined}
}

// identifier is the shape of a check identifier this project can look up: two
// or three letters and four digits, which is what staticcheck names its checks
// and what `staticcheck -list-checks` answers for. The point of requiring one
// is written in staticcheck.conf: an exclusion is worth nothing until its
// identifier has been checked against that list, and a typo leaves the check
// enabled and leaves a comment behind saying it is not.
var identifier = regexp.MustCompile(`^[A-Z]{1,3}[0-9]{4}$`)

// suppressionMarkers is the two spellings a suppression is written in. They
// are assembled rather than written out because this rule reads Go source as
// text, and a literal here would be a suppression this file carries and this
// rule refuses. A checker refusing its own documentation is a shape that has
// already cost a repair elsewhere.
func suppressionMarkers() []string {
	return []string{"//" + "lint:ignore", "//" + "nolint"}
}

// suppressions refuses a suppression that silences an analyser without saying
// which check it silences or why.
//
// The subject is every tracked Go file rather than every suppression, which is
// the difference between this rule and one that cannot fail closed. There is
// no suppression in this tree today, so the rule reads sixteen files and
// refuses none, and that is a pass that says what it looked at.
func suppressions(env Env, all []string) (verdict, error) {
	files := []string{}
	for _, file := range all {
		if strings.HasSuffix(file, ".go") {
			files = append(files, file)
		}
	}
	got := verdict{examined: len(files), subject: "tracked Go file(s)"}
	for _, file := range files {
		content, err := env.Read(file)
		if err != nil {
			return verdict{}, err
		}
		for number, line := range records(content) {
			for _, marker := range suppressionMarkers() {
				at := strings.Index(line, marker)
				if at < 0 {
					continue
				}
				names, reason := suppressionParts(marker, line[at+len(marker):])
				where := fmt.Sprintf("%s:%d", file, number+1)
				switch {
				case names == "":
					got.refused = append(got.refused, fmt.Sprintf("%s silences an analyser and names no check identifier, so nothing can be looked up and nothing would notice the check it meant to silence still running", where))
				case !namesChecks(names):
					got.refused = append(got.refused, fmt.Sprintf("%s names %q, which is not a check identifier this project can look up against `staticcheck -list-checks`", where, names))
				case reason == "":
					got.refused = append(got.refused, fmt.Sprintf("%s silences %s and gives no reason, so a later reader cannot tell whether the finding was wrong or merely inconvenient", where, names))
				}
			}
		}
	}
	return got, nil
}

// suppressionParts splits what follows a marker into the checks it names and
// the reason it gives. The two spellings put the identifiers in different
// places: one takes them as the first word after a space, the other after a
// colon, and a marker with neither names nothing.
func suppressionParts(marker string, rest string) (names string, reason string) {
	if strings.HasSuffix(marker, "nolint") {
		if !strings.HasPrefix(rest, ":") {
			return "", ""
		}
		rest = rest[1:]
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return "", ""
	}
	return fields[0], strings.TrimSpace(strings.Join(fields[1:], " "))
}

// namesChecks reports whether every identifier in a comma-separated list is
// one this project could look up. All of them rather than the first, because a
// suppression naming one real check and one typed one silences the real one
// and leaves the other running under a comment saying it does not.
func namesChecks(names string) bool {
	for _, name := range strings.Split(names, ",") {
		if !identifier.MatchString(strings.TrimSpace(name)) {
			return false
		}
	}
	return true
}

// listingCommand is the line a document quotes above a pasted block of
// check-run names. The pattern the document quoted is taken out of it rather
// than assumed, because the two blocks in this tree quote two different
// patterns and only one of them keeps the upload step names.
var listingCommand = regexp.MustCompile(`^git grep -n '([^']*)' -- \.github/workflows/$`)

// checkListings refuses a document whose pasted block of check-run names is
// not what the command above it produces.
//
// This is the rule the tree has already broken twice in two days. A check
// landed, four documents that enumerate the check-run names stayed one short,
// and the workflow rewrite that followed moved five line numbers in every one
// of them. Both were found by a person running the commands by hand. The
// document that costs the most while it is wrong is docs/required-checks.md,
// because a required check is matched by its literal name and a name missing
// from that list is a check the branch would not require with nothing saying
// so.
//
// WHAT THIS DOES NOT DO. The pattern the document quotes is evaluated here
// with Go's regular expressions rather than by git, so a pattern that means
// two different things to the two engines would be judged by this one. The two
// patterns quoted in this tree are a literal and a repeated space, and those
// mean the same thing to both. A document quoting a third would want checking
// by hand.
// THE SUBJECT IS EVERY DOCUMENT RATHER THAN EVERY PASTED BLOCK, and the
// difference is what keeps the rule able to fail closed. A count of blocks
// goes to nothing the day the last document drops one, and nothing to compare
// is what a rule pointed at a subject that moved away also reports. Counting
// documents makes the examination empty only when there are no documents,
// which is a tree this leg should refuse, and the sentence it prints carries
// the number of blocks it actually compared so a reader can see it fall.
func checkListings(env Env, all []string) (verdict, error) {
	files := documents(all)
	got := verdict{examined: len(files)}
	compared := 0
	for _, file := range files {
		content, err := env.Read(file)
		if err != nil {
			return verdict{}, err
		}
		read := records(content)
		for number, line := range read {
			match := listingCommand.FindStringSubmatch(strings.TrimSpace(line))
			if match == nil {
				continue
			}
			compared++
			produced, err := workflowNames(env, all, match[1])
			if err != nil {
				return verdict{}, err
			}
			pasted := pastedBlock(read[number+1:])
			got.refused = append(got.refused, listingDifferences(file, number+1, pasted, produced)...)
		}
	}
	got.subject = fmt.Sprintf("tracked document(s), in which %d pasted check listing(s) were compared against the workflows", compared)
	return got, nil
}

// pastedBlock is the run of indented lines naming a workflow file that follows
// a quoted command. It stops at the first line that is not one, which is what
// ends a transcript in every document here.
func pastedBlock(rest []string) []string {
	block := []string{}
	for _, line := range rest {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, ".github/workflows/") {
			break
		}
		block = append(block, trimmed)
	}
	return block
}

// workflowNames is what the quoted command produces, read out of the tracked
// workflow files in the order git lists them.
func workflowNames(env Env, all []string, pattern string) ([]string, error) {
	expression, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	produced := []string{}
	for _, file := range all {
		if !strings.HasPrefix(file, ".github/workflows/") {
			continue
		}
		content, err := env.Read(file)
		if err != nil {
			return nil, err
		}
		for number, line := range records(content) {
			if expression.MatchString(line) {
				produced = append(produced, fmt.Sprintf("%s:%d:%s", file, number+1, line))
			}
		}
	}
	return produced, nil
}

// listingDifferences says what is wrong with one pasted block, line by line,
// so a reader repairs the block rather than being told it disagrees.
func listingDifferences(file string, at int, pasted []string, produced []string) []string {
	differences := []string{}
	for index := range pasted {
		if index >= len(produced) {
			differences = append(differences, fmt.Sprintf("%s:%d pastes %q, which the command it quotes does not produce at all", file, at+1+index, pasted[index]))
			continue
		}
		if pasted[index] != produced[index] {
			differences = append(differences, fmt.Sprintf("%s:%d pastes %q where the command it quotes produces %q", file, at+1+index, pasted[index], produced[index]))
		}
	}
	for index := len(pasted); index < len(produced); index++ {
		differences = append(differences, fmt.Sprintf("%s:%d pastes %d line(s) where the command it quotes produces %d, and %q is one it is missing", file, at, len(pasted), len(produced), produced[index]))
	}
	return differences
}
