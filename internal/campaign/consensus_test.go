package campaign

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

// taskFrom judges a declared task and hands back the task it becomes, so that
// every fixture below is a task the definition boundary accepts. A fixture the
// boundary would refuse would be a table of cases about campaigns nobody can
// run.
func taskFrom(t *testing.T, declared DeclaredTask) Task {
	t.Helper()
	defined, err := Define(DeclaredCampaign{
		ID:    "worked-cases",
		Title: "The worked cases in the consensus document",
		Tasks: []DeclaredTask{declared},
	})
	if err != nil {
		t.Fatalf("the fixture task is refused by the boundary: %v", err)
	}
	return defined.Tasks()[0]
}

// plateCondition is task t1 of the worked cases in docs/consensus.md, a single
// choice over three options.
func plateCondition(t *testing.T) Task {
	t.Helper()
	return taskFrom(t, DeclaredTask{
		ID:       "t1",
		Question: "What condition is this plate in?",
		Type:     "single-choice",
		Options: []DeclaredOption{
			{ID: "clear", Text: "Clear"},
			{ID: "marked", Text: "Marked"},
			{ID: "unusable", Text: "Unusable"},
		},
	})
}

// whatIsVisible is task t2 of the worked cases, a choice of one to three out
// of four options.
func whatIsVisible(t *testing.T) Task {
	t.Helper()
	return taskFrom(t, DeclaredTask{
		ID:       "t2",
		Question: "Which of these can you see on the plate?",
		Type:     "choice-of-several",
		Options: []DeclaredOption{
			{ID: "star", Text: "A star"},
			{ID: "galaxy", Text: "A galaxy"},
			{ID: "plate_defect", Text: "A defect in the plate"},
			{ID: "annotation_mark", Text: "A mark somebody drew on the plate"},
		},
		AtLeast: bound(1),
		AtMost:  bound(3),
	})
}

// exactlyOneOf is task t3 of the worked cases, the choice of several whose
// bounds admit one option and one only. It exists in the document to show what
// happens when the derived set breaks the bound the task declared.
func exactlyOneOf(t *testing.T) Task {
	t.Helper()
	return taskFrom(t, DeclaredTask{
		ID:       "t3",
		Question: "Which one of these is it?",
		Type:     "choice-of-several",
		Options: []DeclaredOption{
			{ID: "a", Text: "The first one"},
			{ID: "b", Text: "The second one"},
		},
		AtLeast: bound(1),
		AtMost:  bound(1),
	})
}

// mayNameNothing is a choice of several whose lower bound is zero, which is
// the shape that can derive the empty set as a label.
func mayNameNothing(t *testing.T) Task {
	t.Helper()
	return taskFrom(t, DeclaredTask{
		ID:       "t4",
		Question: "Which of these can you see, if any?",
		Type:     "choice-of-several",
		Options: []DeclaredOption{
			{ID: "star", Text: "A star"},
			{ID: "galaxy", Text: "A galaxy"},
		},
		AtLeast: bound(0),
		AtMost:  bound(2),
	})
}

// answers writes a list of single choice answers the way the document writes
// them, one identifier per answer.
func answers(ids ...string) []Answer {
	out := make([]Answer, 0, len(ids))
	for _, id := range ids {
		out = append(out, Answer{id})
	}
	return out
}

// sameShare reports whether two shares are the same number. The shares are
// counts over one denominator, so equal counts give an identical float, and
// this tolerance is for the literals written in the cases below rather than
// for the arithmetic.
func sameShare(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// check compares one computed label against what the case expects and says
// which field moved, because a label that is wrong in one field and right in
// four is the failure that is worth seconds rather than minutes.
func check(t *testing.T, name string, got Label, labelled bool, options []string, confidence float64, count int, level bool) {
	t.Helper()
	if got.Labelled() != labelled {
		t.Errorf("%s: labelled is %v, want %v (%s)", name, got.Labelled(), labelled, got)
	}
	if !reflect.DeepEqual(got.Options(), options) {
		t.Errorf("%s: options are %v, want %v", name, got.Options(), options)
	}
	if !sameShare(got.Confidence(), confidence) {
		t.Errorf("%s: confidence is %v, want %v", name, got.Confidence(), confidence)
	}
	if got.Answers() != count {
		t.Errorf("%s: input count is %d, want %d", name, got.Answers(), count)
	}
	if got.Level() != level {
		t.Errorf("%s: level at the top is %v, want %v", name, got.Level(), level)
	}
}

// TestTheSingleChoiceCasesWorkedInTheDocument runs the four single choice
// cases docs/consensus.md works by hand, with the numbers that document
// prints. It is the table the issue asks for and it is a copy of the decision
// rather than of the code, so a change to the rule that nobody meant reds here
// against the document rather than against itself.
func TestTheSingleChoiceCasesWorkedInTheDocument(t *testing.T) {
	task := plateCondition(t)
	rule := DefaultAgreement()

	// Agreeing: clear 4 of 5, share 0.80, which is at or above 0.70.
	check(t, "agreeing",
		Consensus(task, answers("clear", "clear", "marked", "clear", "clear"), rule),
		true, []string{"clear"}, 0.80, 5, false)

	// Level at the top, which the document calls the tie: clear 2, marked 2,
	// and unusable named by nobody.
	check(t, "level at the top",
		Consensus(task, answers("clear", "marked", "clear", "marked"), rule),
		false, nil, 0, 4, true)

	// Spread: clear 2, marked 2, unusable 1, top share 0.40. Two options are
	// level at 0.40 and a third was named, so the flag is off and the case is
	// recorded as spread rather than as ambivalence, which is the condition
	// the document now states in one place.
	check(t, "spread",
		Consensus(task, answers("clear", "clear", "marked", "unusable", "marked"), rule),
		false, nil, 0, 5, false)

	// Three answers naming three different options: every option at 0.33, so
	// three options share the top and the flag is off. This is the case that
	// decided which reading the flag means, and the document works it.
	check(t, "complete disagreement",
		Consensus(task, answers("clear", "marked", "unusable"), rule),
		false, nil, 0, 3, false)

	// One answer only: share 1.00, so a label at an input count of one.
	check(t, "one answer",
		Consensus(task, answers("marked"), rule),
		true, []string{"marked"}, 1.00, 1, false)
}

// TestTheChoiceOfSeveralCasesWorkedInTheDocument runs the two cases
// docs/consensus.md works for the other type in the grammar, so that every
// question type has a table rather than only the one that was easier.
func TestTheChoiceOfSeveralCasesWorkedInTheDocument(t *testing.T) {
	rule := DefaultAgreement()

	// Inside its bounds: star 5 of 5, galaxy 3 of 5 at 0.60 which does not
	// enter, plate_defect 1 of 5, annotation_mark none.
	check(t, "inside its bounds",
		Consensus(whatIsVisible(t), []Answer{
			{"star", "galaxy"},
			{"star"},
			{"star", "galaxy"},
			{"star", "plate_defect"},
			{"star", "galaxy"},
		}, rule),
		true, []string{"star"}, 1.00, 5, false)

	// Breaking the declared bound: a 2 of 3 and b 2 of 3, both at 0.67, so the
	// set at or above the threshold is empty and empty is below the declared
	// lower bound of one.
	check(t, "breaking the bound",
		Consensus(exactlyOneOf(t), []Answer{{"a"}, {"b"}, {"a", "b"}}, rule),
		false, nil, 0, 3, true)
}

// TestATieIsNeverLabelledEvenWhereBothOptionsReachTheThreshold is the second
// of the two conditions docs/consensus.md puts on a single choice, and it is
// only reachable below a threshold of 0.51. At the default threshold no tie
// can also clear the bar, so a rule that had dropped the tie condition would
// pass every case worked in the document and would publish a single choice
// label naming two options the first time a campaign owner declared 0.5.
func TestATieIsNeverLabelledEvenWhereBothOptionsReachTheThreshold(t *testing.T) {
	rule, err := Agreement(0.5)
	if err != nil {
		t.Fatalf("a threshold of 0.5 was refused: %v", err)
	}

	got := Consensus(plateCondition(t), answers("clear", "marked"), rule)
	if got.Labelled() {
		t.Fatalf("two options level at 0.50 under a threshold of 0.50 produced %s", got)
	}
	if !got.Level() {
		t.Errorf("two options level at the top were not reported level: %s", got)
	}
	if len(got.Options()) != 0 {
		t.Errorf("a single choice derived %v, and a single choice admits one option", got.Options())
	}
}

// TestCompleteDisagreementProducesNoLabel is the case the issue names beside
// the tie. Every volunteer named a different option, so nothing is anywhere
// near the threshold.
//
// It also holds the direction #180 decided the flag in. Three options share
// the top and the answers account for the whole of it, so the two readings
// this project rejected would both report the most spread result a campaign
// can produce as ambivalence.
func TestCompleteDisagreementProducesNoLabel(t *testing.T) {
	got := Consensus(plateCondition(t), answers("clear", "marked", "unusable"), DefaultAgreement())
	if got.Labelled() {
		t.Fatalf("three answers naming three different options produced %s", got)
	}
	if got.Answers() != 3 {
		t.Errorf("input count is %d, want 3", got.Answers())
	}
	if got.Level() {
		t.Errorf("three options sharing the top were reported level at the top: %s", got)
	}
}

// TestTheFlagReadsTwoOptionsAndNothingElseNamed is the condition #180 decided,
// held here as a table rather than inside one worked case, because the two
// readings it rejects each pass some of these rows and fail others.
//
// A reading of any largest share held by more than one option passes the first
// row and the fourth and fails the second and the third. A reading of the
// options at the top accounting for every answer passes the first three and
// fails the third. Only the condition the document states passes all four.
func TestTheFlagReadsTwoOptionsAndNothingElseNamed(t *testing.T) {
	task := plateCondition(t)
	rule := DefaultAgreement()

	cases := []struct {
		name    string
		answers []Answer
		level   bool
	}{
		{
			"two options level and no third named",
			answers("clear", "marked", "clear", "marked"),
			true,
		},
		{
			"two options level and a third named once",
			answers("clear", "clear", "marked", "unusable", "marked"),
			false,
		},
		{
			"three options sharing the top",
			answers("clear", "marked", "unusable"),
			false,
		},
		{
			"two options level beside an answer naming nothing declared",
			answers("clear", "marked", "nowhere"),
			true,
		},
	}

	for _, c := range cases {
		got := Consensus(task, c.answers, rule)
		if got.Labelled() {
			t.Fatalf("%s: produced a label, and every row here is a subject with none: %s", c.name, got)
		}
		if got.Level() != c.level {
			t.Errorf("%s: level at the top is %v, want %v: %s", c.name, got.Level(), c.level, got)
		}
	}
}

// TestTheSameAnswersInAnyOrderProduceTheSameLabel is the property the export
// rests on. The rule reads no order and no arrival time, so every permutation
// of one answer set has to produce a label equal in every field, including the
// order of the derived set itself.
func TestTheSameAnswersInAnyOrderProduceTheSameLabel(t *testing.T) {
	rule := DefaultAgreement()

	cases := []struct {
		name    string
		task    Task
		answers []Answer
	}{
		{"a single choice that labels", plateCondition(t), answers("clear", "clear", "marked", "clear", "clear")},
		{"a single choice that ties", plateCondition(t), answers("clear", "marked", "clear", "marked")},
		{"a choice of several", whatIsVisible(t), []Answer{
			{"star", "galaxy"},
			{"star"},
			{"galaxy", "star"},
			{"plate_defect", "star"},
			{"star", "galaxy"},
		}},
	}

	for _, one := range cases {
		want := Consensus(one.task, one.answers, rule)
		seen := 0
		permute(one.answers, func(order []Answer) {
			seen++
			got := Consensus(one.task, order, rule)
			if got.String() != want.String() {
				t.Fatalf("%s: the order %v produced %s, want %s", one.name, order, got, want)
			}
			if !reflect.DeepEqual(got.Options(), want.Options()) {
				t.Fatalf("%s: the order %v derived %v, want %v", one.name, order, got.Options(), want.Options())
			}
		})
		if seen != factorial(len(one.answers)) {
			t.Fatalf("%s: examined %d order(s) of %d answer(s), want %d", one.name, seen, len(one.answers), factorial(len(one.answers)))
		}
		t.Logf("%s: examined %d order(s) of %d answer(s)", one.name, seen, len(one.answers))
	}
}

// permute calls back with every order of the answers, on a copy each time so
// that a callback holding one cannot change the next.
func permute(given []Answer, yield func([]Answer)) {
	order := append([]Answer(nil), given...)
	var walk func(int)
	walk = func(from int) {
		if from == len(order) {
			yield(append([]Answer(nil), order...))
			return
		}
		for i := from; i < len(order); i++ {
			order[from], order[i] = order[i], order[from]
			walk(from + 1)
			order[from], order[i] = order[i], order[from]
		}
	}
	walk(0)
}

// factorial is how many orders a set of that size has, so the permutation test
// can refuse a run that examined fewer than all of them. A test that walked no
// order and a test that walked every one print the same word otherwise.
func factorial(n int) int {
	out := 1
	for i := 2; i <= n; i++ {
		out *= i
	}
	return out
}

// TestTheSmallestAgreeingCountThatReachesTheThreshold is the table
// docs/consensus.md prints so that a campaign owner reading 0.7 as roughly
// seventy percent is not surprised. At three answers the default threshold is
// unanimity, and that is the row worth having a test for.
func TestTheSmallestAgreeingCountThatReachesTheThreshold(t *testing.T) {
	task := plateCondition(t)
	rule := DefaultAgreement()

	// The answers that do not agree alternate between the other two options,
	// so that the disagreeing half can never reach the threshold itself and
	// the row being read is the agreeing count rather than a second majority.
	needed := []int{0, 1, 2, 3, 3, 4, 5, 5}
	for total := 1; total <= 7; total++ {
		want := needed[total]
		for agreeing := 1; agreeing <= total; agreeing++ {
			var given []string
			for i := 0; i < agreeing; i++ {
				given = append(given, "clear")
			}
			for i := agreeing; i < total; i++ {
				if (i-agreeing)%2 == 0 {
					given = append(given, "marked")
				} else {
					given = append(given, "unusable")
				}
			}
			got := Consensus(task, answers(given...), rule)
			if got.Labelled() != (agreeing >= want) {
				t.Errorf("%d of %d agreeing: labelled is %v, want %v (%s)", agreeing, total, got.Labelled(), agreeing >= want, got)
			}
			if got.Labelled() && !reflect.DeepEqual(got.Options(), []string{"clear"}) {
				t.Errorf("%d of %d agreeing: the label is %v, want the agreeing option", agreeing, total, got.Options())
			}
		}
	}
}

// TestAnAnswerNamingSomethingTheTaskDoesNotDeclareCountsAndSupportsNothing
// holds the branch docs/consensus.md requires to exist: a rule that assumes
// clean input is a rule with an undefined branch, and the surface and the
// definition boundary are not the only things that ever write to a store.
//
// The answer counts in the denominator, so it cannot raise the share of the
// options that were named. Dropping it would do exactly that.
func TestAnAnswerNamingSomethingTheTaskDoesNotDeclareCountsAndSupportsNothing(t *testing.T) {
	task := plateCondition(t)
	rule := DefaultAgreement()

	got := Consensus(task, answers("clear", "clear", "something-else"), rule)
	if got.Labelled() {
		t.Errorf("two of three is 0.67 but the third answer was dropped, giving %s", got)
	}
	if got.Answers() != 3 {
		t.Errorf("input count is %d, want 3", got.Answers())
	}

	all := Consensus(task, answers("nowhere", "nowhere"), rule)
	if all.Labelled() {
		t.Errorf("an option the task does not declare became a label: %s", all)
	}
	if all.Level() {
		t.Errorf("answers naming nothing the task declares were reported level at the top: %s", all)
	}
}

// TestAnIdentifierNamedTwiceInOneAnswerIsNamedOnce holds an answer as the set
// docs/vocabulary.md says it is. Counting a repeat twice would let one
// malformed submission carry an option past the threshold on its own.
func TestAnIdentifierNamedTwiceInOneAnswerIsNamedOnce(t *testing.T) {
	got := Consensus(whatIsVisible(t), []Answer{
		{"star", "star", "star"},
		{"galaxy"},
	}, DefaultAgreement())

	if got.Labelled() {
		t.Fatalf("star was named once by one answer of two, a share of 0.50, and produced %s", got)
	}
}

// TestASingleChoiceAnswerNamingTwoOptionsSupportsNeither is the other shape of
// malformed answer. A single choice reports bounds of one and one, so an
// answer naming two is one the boundary and the surface should have refused,
// and here it agrees with nothing rather than with both.
func TestASingleChoiceAnswerNamingTwoOptionsSupportsNeither(t *testing.T) {
	got := Consensus(plateCondition(t), []Answer{
		{"clear", "marked"},
		{"clear"},
	}, DefaultAgreement())

	if got.Labelled() {
		t.Fatalf("an answer naming two options of a single choice supported one of them: %s", got)
	}
	if got.Answers() != 2 {
		t.Errorf("input count is %d, want 2", got.Answers())
	}
}

// TestNoAnswersIsNoLabel is docs/vocabulary.md's cardinality, which gives a
// subject and a task no label until the first answer.
func TestNoAnswersIsNoLabel(t *testing.T) {
	got := Consensus(plateCondition(t), nil, DefaultAgreement())
	if got.Labelled() {
		t.Fatalf("no answers produced %s", got)
	}
	if got.Answers() != 0 {
		t.Errorf("input count is %d, want 0", got.Answers())
	}
	if got.Level() {
		t.Errorf("no answers were reported level at the top: %s", got)
	}
}

// TestAChoiceOfSeveralWithNoLowerBoundCanLabelTheEmptySet keeps the two
// different findings apart. A task whose owner declared that naming nothing is
// an answer can derive the empty set as its label, and that is not the same as
// having no label.
func TestAChoiceOfSeveralWithNoLowerBoundCanLabelTheEmptySet(t *testing.T) {
	got := Consensus(mayNameNothing(t), []Answer{{}, {}, {}}, DefaultAgreement())
	if !got.Labelled() {
		t.Fatalf("three answers naming nothing, on a task admitting nothing, produced %s", got)
	}
	if len(got.Options()) != 0 {
		t.Errorf("the label is %v, want the empty set", got.Options())
	}
}

// TestATaskOfNoKnownTypeProducesNoLabelRatherThanAPanic holds the branch a
// zero value reaches. Task is exported and its fields are not, so any package
// can build one that Define never judged, and a rule that panicked on it would
// turn a caller's uninitialised variable into a dead process.
func TestATaskOfNoKnownTypeProducesNoLabelRatherThanAPanic(t *testing.T) {
	got := Consensus(Task{}, answers("clear", "clear"), DefaultAgreement())
	if got.Labelled() {
		t.Fatalf("a task of no known type produced %s", got)
	}
	if got.Answers() != 2 {
		t.Errorf("input count is %d, want 2", got.Answers())
	}
}

// TestTheComputationTakesNothingButItsThreeArguments is the third condition of
// #37 read off the signature. A store, a clock or a volunteer identifier
// reaching this rule would be invisible in a table of cases, so the shape of
// the function is asserted rather than described.
func TestTheComputationTakesNothingButItsThreeArguments(t *testing.T) {
	shape := reflect.TypeOf(Consensus)
	if got := shape.NumIn(); got != 3 {
		t.Fatalf("Consensus takes %d argument(s), want 3", got)
	}
	want := []string{"campaign.Task", "[]campaign.Answer", "campaign.Rule"}
	for i, name := range want {
		if got := shape.In(i).String(); got != name {
			t.Errorf("argument %d is %s, want %s", i, got, name)
		}
	}
	if got := shape.NumOut(); got != 1 {
		t.Fatalf("Consensus returns %d value(s), want 1", got)
	}
	if got := shape.Out(0).String(); got != "campaign.Label" {
		t.Errorf("Consensus returns %s, want campaign.Label", got)
	}
}

// TestAThresholdThatIsNotAShareIsRefused holds the constructor. A threshold of
// zero labels every subject with whatever was named most, including a single
// answer nobody agreed with, and one above a share labels nothing at all.
// Neither reports anything at the moment it is set.
func TestAThresholdThatIsNotAShareIsRefused(t *testing.T) {
	for _, threshold := range []float64{-0.1, 0, 1.5, math.NaN()} {
		if _, err := Agreement(threshold); err == nil {
			t.Errorf("a threshold of %v was accepted", threshold)
		}
	}
	for _, threshold := range []float64{0.5, 0.7, 1} {
		rule, err := Agreement(threshold)
		if err != nil {
			t.Fatalf("a threshold of %v was refused: %v", threshold, err)
		}
		if rule.Threshold() != threshold {
			t.Errorf("the rule holds a threshold of %v, want %v", rule.Threshold(), threshold)
		}
	}
	if got := DefaultAgreement().Threshold(); got != DefaultThreshold {
		t.Errorf("the default rule holds a threshold of %v, want %v", got, DefaultThreshold)
	}
}

// TestUnanimityIsAvailableByDeclaringAThresholdOfOne is the option
// docs/consensus.md leaves open to a campaign owner, and it is the boundary of
// the interval the constructor admits.
func TestUnanimityIsAvailableByDeclaringAThresholdOfOne(t *testing.T) {
	rule, err := Agreement(1)
	if err != nil {
		t.Fatalf("unanimity was refused: %v", err)
	}
	task := plateCondition(t)

	if got := Consensus(task, answers("clear", "clear", "clear"), rule); !got.Labelled() {
		t.Errorf("three answers agreeing under unanimity produced %s", got)
	}
	if got := Consensus(task, answers("clear", "clear", "marked"), rule); got.Labelled() {
		t.Errorf("two of three under unanimity produced %s", got)
	}
}

// TestTheLabelCarriesTheRuleThatComputedIt is the third of the four things
// written beside a label. A campaign owner may change the threshold while a
// campaign runs, and a label computed under the old one has to be readable as
// such rather than silently compared with one computed under the new one.
func TestTheLabelCarriesTheRuleThatComputedIt(t *testing.T) {
	strict, err := Agreement(1)
	if err != nil {
		t.Fatalf("unanimity was refused: %v", err)
	}
	got := Consensus(plateCondition(t), answers("clear", "clear"), strict)

	if got.Rule().Threshold() != 1 {
		t.Errorf("the label records a threshold of %v, want 1", got.Rule().Threshold())
	}
	if got.Rule().ID() != RuleID {
		t.Errorf("the label records the rule %q, want %q", got.Rule().ID(), RuleID)
	}
	if !strings.Contains(got.String(), strict.String()) {
		t.Errorf("the label reads %q and does not carry its rule", got.String())
	}
}

// TestALabelIsStaleWhenTheAnswersOrTheRuleMovedUnderIt holds what
// docs/vocabulary.md pays for storing the label rather than recomputing it on
// read: staleness that is detectable by inspection rather than by suspicion.
func TestALabelIsStaleWhenTheAnswersOrTheRuleMovedUnderIt(t *testing.T) {
	rule := DefaultAgreement()
	stored := Consensus(plateCondition(t), answers("clear", "clear", "clear"), rule)

	if why := stored.Stale(3, rule); len(why) != 0 {
		t.Errorf("a label matching its inputs was called stale: %v", why)
	}
	if why := stored.Stale(4, rule); len(why) != 1 {
		t.Errorf("a label computed from three answers against four classifications reported %v, want one reason", why)
	}

	strict, err := Agreement(1)
	if err != nil {
		t.Fatalf("unanimity was refused: %v", err)
	}
	if why := stored.Stale(3, strict); len(why) != 1 {
		t.Errorf("a label computed under a threshold the campaign no longer runs reported %v, want one reason", why)
	}
	if why := stored.Stale(4, strict); len(why) != 2 {
		t.Errorf("a label stale in both ways reported %v, want two reasons", why)
	}
}

// TestAnAnswerArrivingAfterTheSubjectRetiredIsCountedAndMakesTheStoredLabelStale
// is the fourth awkward case the issue names. The rule reads no time and no
// retirement state, so a late answer is an answer: it changes the label it is
// recomputed into, and the label stored before it arrived says by inspection
// that it was computed from fewer.
func TestAnAnswerArrivingAfterTheSubjectRetiredIsCountedAndMakesTheStoredLabelStale(t *testing.T) {
	task := plateCondition(t)
	rule := DefaultAgreement()
	early := answers("clear", "clear", "clear")

	stored := Consensus(task, early, rule)
	if !stored.Labelled() {
		t.Fatalf("three agreeing answers produced %s", stored)
	}

	late := append(append([]Answer(nil), early...), Answer{"marked"})
	if why := stored.Stale(len(late), rule); len(why) == 0 {
		t.Fatal("a label computed before a late answer arrived was not stale against the count that includes it")
	}

	// Three of four is 0.75, so the label survives the late answer and its
	// confidence and input count both move. A stored label that was not
	// recomputed would keep saying 1.00 from three.
	recomputed := Consensus(task, late, rule)
	check(t, "one late answer", recomputed, true, []string{"clear"}, 0.75, 4, false)
	if stored.Confidence() == recomputed.Confidence() {
		t.Errorf("the late answer left the confidence at %v", stored.Confidence())
	}

	// A second one takes it to three of five, which is 0.60, and the subject
	// that had a label has none. That is the case a deployment ignoring
	// staleness would publish the old label for.
	later := append(append([]Answer(nil), late...), Answer{"marked"})
	if got := Consensus(task, later, rule); got.Labelled() {
		t.Fatalf("three of five is 0.60 under a threshold of 0.7 and produced %s", got)
	}
}

// TestALabelCannotBeWrittenThrough keeps the readers honest for the same
// reason the judged definition types do. A caller that could write into the
// derived set could change a label after it was computed and after its
// confidence was written beside it.
func TestALabelCannotBeWrittenThrough(t *testing.T) {
	got := Consensus(plateCondition(t), answers("clear", "clear", "clear"), DefaultAgreement())
	options := got.Options()
	if len(options) != 1 {
		t.Fatalf("the label is %v, want one option", options)
	}
	options[0] = "something-else"

	if again := got.Options(); again[0] != "clear" {
		t.Errorf("writing to the returned slice changed the label: %q", again[0])
	}
}

// TestALabelHasNoExportedField is how the constructor stays the only way to
// hold a label whose confidence matches its options.
func TestALabelHasNoExportedField(t *testing.T) {
	for _, one := range []any{Label{}, Rule{}} {
		shape := reflect.TypeOf(one)
		for i := 0; i < shape.NumField(); i++ {
			if field := shape.Field(i); field.IsExported() {
				t.Errorf("%s.%s is exported, so a %s can be built without passing the rule", shape.Name(), field.Name, shape.Name())
			}
		}
	}
}

// TestALabelReadsAsTheDocumentWritesOne is the reader a person checking a
// campaign by hand uses, and it is asserted rather than pasted so that the
// wording keeps holding.
func TestALabelReadsAsTheDocumentWritesOne(t *testing.T) {
	task := plateCondition(t)
	rule := DefaultAgreement()

	cases := []struct {
		given []Answer
		want  string
	}{
		{answers("clear", "clear", "marked", "clear", "clear"), "{clear}, confidence 0.80, 5 answer(s), count-share@0.7"},
		{answers("clear", "marked", "clear", "marked"), "no label (level at the top), 4 answer(s), count-share@0.7"},
		{nil, "no label (no answers), 0 answer(s), count-share@0.7"},
	}
	for _, one := range cases {
		if got := Consensus(task, one.given, rule).String(); got != one.want {
			t.Errorf("a label of %v reads %q, want %q", one.given, got, one.want)
		}
	}
}

// TestEveryOptionOfTheTaskIsCountedEvenWhereNobodyNamedIt is the difference
// between a zero and a missing value, which docs/consensus.md names as the
// reason the campaign definition travels with the export. An option nobody
// chose has a share of zero here rather than being absent from the count.
func TestEveryOptionOfTheTaskIsCountedEvenWhereNobodyNamedIt(t *testing.T) {
	task := whatIsVisible(t)
	got := sharePerOption(task, []Answer{{"star"}, {"star"}})

	if len(got) != len(task.Options()) {
		t.Fatalf("counted %d option(s) of %d", len(got), len(task.Options()))
	}
	for i, option := range task.Options() {
		want := 0.0
		if option.ID() == "star" {
			want = 1.0
		}
		if !sameShare(got[i], want) {
			t.Errorf("%s has a share of %v, want %v", option.ID(), got[i], want)
		}
	}
}
