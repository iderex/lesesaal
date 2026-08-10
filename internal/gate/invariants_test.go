package gate

import (
	"strings"
	"testing"
)

// The markers are assembled here for the same reason they are assembled in
// invariants.go: this file is a tracked Go file, the rule below reads tracked
// Go files as text, and a literal in a fixture would be a suppression the
// suite's own source carries and the leg refuses.
var (
	ignore = "//" + "lint:ignore"
	nolint = "//" + "nolint"
)

// suppressionTree is a two-file tree with one Go file a test spoils. The
// second file is there so the count the leg reports is a count rather than a
// constant.
func suppressionTree(source string) *fake {
	return &fake{
		files: map[string]string{
			"main.go":         "package main\n\nfunc main() {}\n",
			"internal/a/a.go": source,
			"README.md":       "# A tree\n",
		},
		answers: map[string]answer{
			"git ls-files .": {output: "README.md\ninternal/a/a.go\nmain.go\n"},
		},
	}
}

// TestASuppressionThatNamesItsCheckAndItsReasonIsNotRefused is the near miss's
// other half. Without it the three tests below would pass against a rule that
// refuses every suppression, which is a different rule from the one written
// down and one nobody could satisfy.
func TestASuppressionThatNamesItsCheckAndItsReasonIsNotRefused(t *testing.T) {
	f := suppressionTree("package a\n\n" + ignore + " ST1003 the field name follows the column it stores\nfunc A() {}\n")

	got, err := suppressions(f.env(), []string{"README.md", "internal/a/a.go", "main.go"})
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 0 {
		t.Errorf("a suppression naming its check and its reason was refused: %v", got.refused)
	}
	if got.examined != 2 {
		t.Errorf("the rule reports examining %d Go file(s) where the tree carries 2", got.examined)
	}
}

func TestASuppressionWithNoCheckIdentifierIsRefused(t *testing.T) {
	f := suppressionTree("package a\n\n" + ignore + "\nfunc A() {}\n")

	got, err := suppressions(f.env(), []string{"README.md", "internal/a/a.go", "main.go"})
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a suppression naming no check was not refused exactly once: %v", got.refused)
	}
	if !strings.Contains(got.refused[0], "internal/a/a.go:3") {
		t.Errorf("the refusal does not name where the suppression is: %q", got.refused[0])
	}
}

// TestASuppressionNamingSomethingThatIsNotACheckIsRefused is the one-character
// mistake staticcheck.conf says is worth refusing: an identifier that was
// never checked against the tool's own list leaves the check enabled and
// leaves a comment behind saying it is not.
func TestASuppressionNamingSomethingThatIsNotACheckIsRefused(t *testing.T) {
	f := suppressionTree("package a\n\n" + ignore + " ST103 the digit that was dropped\nfunc A() {}\n")

	got, err := suppressions(f.env(), []string{"README.md", "internal/a/a.go", "main.go"})
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a suppression naming ST103 was not refused exactly once: %v", got.refused)
	}
	if !strings.Contains(got.refused[0], "ST103") {
		t.Errorf("the refusal does not quote what was named: %q", got.refused[0])
	}
}

// TestASuppressionNamingOneRealCheckAndOneTypedOneIsRefused is the case a rule
// checking only the first identifier would pass. The real check is silenced
// and the typed one keeps running under a comment saying it does not.
func TestASuppressionNamingOneRealCheckAndOneTypedOneIsRefused(t *testing.T) {
	f := suppressionTree("package a\n\n" + ignore + " ST1003,ST103 two checks, one of them typed\nfunc A() {}\n")

	got, err := suppressions(f.env(), []string{"README.md", "internal/a/a.go", "main.go"})
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a suppression naming one typed check beside a real one was not refused: %v", got.refused)
	}
}

func TestASuppressionWithNoReasonIsRefused(t *testing.T) {
	f := suppressionTree("package a\n\n" + ignore + " ST1003\nfunc A() {}\n")

	got, err := suppressions(f.env(), []string{"README.md", "internal/a/a.go", "main.go"})
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a suppression with no reason was not refused exactly once: %v", got.refused)
	}
	if !strings.Contains(got.refused[0], "no reason") {
		t.Errorf("the refusal does not say what is missing: %q", got.refused[0])
	}
}

// TestTheOtherSuppressionSpellingIsJudgedToo covers the marker that puts its
// identifiers after a colon. A rule that knew only one spelling would be a
// rule anybody could walk around by writing the other.
func TestTheOtherSuppressionSpellingIsJudgedToo(t *testing.T) {
	f := suppressionTree("package a\n\n" + nolint + " because it is fine\nfunc A() {}\n")

	got, err := suppressions(f.env(), []string{"README.md", "internal/a/a.go", "main.go"})
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a suppression in the second spelling naming no check was not refused: %v", got.refused)
	}
}

// listingPaths is the fixture tree for the second rule, in the order git lists
// it. Two workflow files rather than one, because the rule has to hold the
// order the command produces as well as the lines.
var listingPaths = []string{".github/workflows/one.yml", ".github/workflows/two.yml", "docs/checks.md", "main.go"}

const (
	oneWorkflow = "name: One\n" +
		"\n" +
		"jobs:\n" +
		"  one:\n" +
		"    name: The first check\n"
	twoWorkflow = "jobs:\n" +
		"  two:\n" +
		"    name: The second check\n"
)

// listingTree is a document quoting the command and pasting a block under it.
// The block is supplied by the caller, so each test below spoils exactly one
// thing about a document that is otherwise right.
func listingTree(block string) *fake {
	return &fake{
		files: map[string]string{
			".github/workflows/one.yml": oneWorkflow,
			".github/workflows/two.yml": twoWorkflow,
			"docs/checks.md": "# The checks\n" +
				"\n" +
				"    git grep -n '^    name:' -- .github/workflows/\n" +
				block,
			"main.go": "package main\n\nfunc main() {}\n",
		},
		answers: map[string]answer{
			"git ls-files .": {output: strings.Join(listingPaths, "\n") + "\n"},
		},
	}
}

const rightBlock = "    .github/workflows/one.yml:5:    name: The first check\n" +
	"    .github/workflows/two.yml:3:    name: The second check\n"

func TestAPastedCheckListingThatReproducesIsNotRefused(t *testing.T) {
	got, err := checkListings(listingTree(rightBlock).env(), listingPaths)
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 0 {
		t.Errorf("a block that reproduces its own command was refused: %v", got.refused)
	}
	if !strings.Contains(got.subject, "1 pasted check listing(s)") {
		t.Errorf("the rule does not say how many blocks it compared: %q", got.subject)
	}
}

// TestAPastedCheckListingWithAStaleLineNumberIsRefused is the mistake this
// rule exists for. A workflow was rewritten around a job name, every document
// quoting the command kept the old number, and the block stopped reproducing
// without a word in it changing.
func TestAPastedCheckListingWithAStaleLineNumberIsRefused(t *testing.T) {
	stale := strings.Replace(rightBlock, "one.yml:5:", "one.yml:4:", 1)

	got, err := checkListings(listingTree(stale).env(), listingPaths)
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a block with one stale line number was not refused exactly once: %v", got.refused)
	}
	if !strings.Contains(got.refused[0], "docs/checks.md:4") {
		t.Errorf("the refusal does not name the line of the document to repair: %q", got.refused[0])
	}
	if !strings.Contains(got.refused[0], "one.yml:5:") {
		t.Errorf("the refusal does not say what the command produces instead: %q", got.refused[0])
	}
}

// TestAPastedCheckListingMissingANameIsRefused is the other half and the one
// that cost the most: a check landed, four documents that enumerate the names
// stayed one short, and one of them is the list a required check request is
// set from.
func TestAPastedCheckListingMissingANameIsRefused(t *testing.T) {
	short := "    .github/workflows/one.yml:5:    name: The first check\n"

	got, err := checkListings(listingTree(short).env(), listingPaths)
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a block one name short was not refused exactly once: %v", got.refused)
	}
	if !strings.Contains(got.refused[0], "The second check") {
		t.Errorf("the refusal does not name what is missing: %q", got.refused[0])
	}
}

// TestAPastedCheckListingCarryingANameTheWorkflowsDoNotProduceIsRefused is the
// direction a comparison by count alone would pass: a block the same length as
// the command's output and wrong in one entry is caught by the line, not by
// the total.
func TestAPastedCheckListingCarryingAnExtraNameIsRefused(t *testing.T) {
	extra := rightBlock + "    .github/workflows/three.yml:9:    name: A check that is not there\n"

	got, err := checkListings(listingTree(extra).env(), listingPaths)
	if err != nil {
		t.Fatalf("the rule could not read the tree: %v", err)
	}
	if len(got.refused) != 1 {
		t.Fatalf("a block naming a check the workflows do not produce was not refused: %v", got.refused)
	}
}

// TestTheInvariantLegFailsClosedOnAnEmptyExamination is the fourth condition
// of #94. A rule whose subject has moved out from under it reports what a tree
// holding the rule reports, so the empty case is a failure rather than a pass.
func TestTheInvariantLegFailsClosedOnAnEmptyExamination(t *testing.T) {
	f := &fake{
		files:   map[string]string{},
		answers: map[string]answer{"git ls-files .": {output: ""}},
	}

	result := invariants(f.env())
	if result.Outcome != Failed {
		t.Fatalf("a tree with nothing in it passed the invariant leg: %+v", result)
	}
	for _, rule := range invariantRules() {
		if !strings.Contains(result.Output, rule.id) {
			t.Errorf("the refusal does not name the rule %s that examined nothing:\n%s", rule.id, result.Output)
		}
	}
}

// TestAnInvariantRefusalNamesTheRuleAndWhereItWasDecided is the second
// condition of #94. A contributor who trips one of these has to be able to
// find out why rather than working around it.
func TestAnInvariantRefusalNamesTheRuleAndWhereItWasDecided(t *testing.T) {
	f := listingTree(strings.Replace(rightBlock, "one.yml:5:", "one.yml:4:", 1))

	result := invariants(f.env())
	if result.Outcome != Failed {
		t.Fatalf("a stale pasted listing passed the invariant leg: %+v", result)
	}
	rule, found := ruleByID("check-listing-reproduces-the-workflows")
	if !found {
		t.Fatal("the rule this test is about is not in the table")
	}
	if !strings.Contains(result.Output, rule.rule) {
		t.Errorf("the refusal does not quote the rule:\n%s", result.Output)
	}
	if !strings.Contains(result.Output, rule.decided) {
		t.Errorf("the refusal does not say where the rule was decided:\n%s", result.Output)
	}
}

// TestTheInvariantLegSaysWhatEachRuleExamined keeps the leg honest about the
// difference between a rule that read the tree and found nothing and a rule
// that read nothing.
func TestTheInvariantLegSaysWhatEachRuleExamined(t *testing.T) {
	result := invariants(listingTree(rightBlock).env())
	if result.Outcome != Passed {
		t.Fatalf("a tree holding both rules did not pass: %+v", result)
	}
	for _, rule := range invariantRules() {
		if !strings.Contains(result.Examined, rule.id) {
			t.Errorf("the examined sentence does not account for %s: %q", rule.id, result.Examined)
		}
	}
}

// TestEveryInvariantRuleIsDistinctAndSaysWhereItCameFrom is the guard behind
// the promise that adding a rule is one file change: a rule added without an
// id, without a sentence or without the place it was decided would otherwise
// produce a refusal nobody can act on.
func TestEveryInvariantRuleIsDistinctAndSaysWhereItCameFrom(t *testing.T) {
	seen := map[string]bool{}
	for _, rule := range invariantRules() {
		if rule.id == "" || rule.rule == "" || rule.decided == "" || rule.judge == nil {
			t.Errorf("the rule %q is missing one of its four parts", rule.id)
		}
		if seen[rule.id] {
			t.Errorf("two rules are called %q, so a refusal names neither of them", rule.id)
		}
		seen[rule.id] = true
	}
}

// ruleByID finds a rule the way a test names one.
func ruleByID(id string) (invariant, bool) {
	for _, rule := range invariantRules() {
		if rule.id == id {
			return rule, true
		}
	}
	return invariant{}, false
}
