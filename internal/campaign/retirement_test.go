package campaign

import (
	"strings"
	"testing"
)

// plateSweep is the task the empty answer cases are written against: a sweep
// where most plates carry nothing, which is what docs/retirement.md says the
// early exit is for.
func plateSweep(t *testing.T) Task {
	t.Helper()
	return taskFrom(t, DeclaredTask{
		ID:       "sweep",
		Question: "Is there anything on this plate?",
		Type:     "single-choice",
		Options: []DeclaredOption{
			{ID: "nothing", Text: "Nothing"},
			{ID: "something", Text: "Something"},
		},
	})
}

// labelled is the label a task gets when the given answers are counted under
// the default agreement threshold, which is where every share and every
// statement about agreement in these cases comes from.
func labelled(t *testing.T, task Task, given ...string) Label {
	t.Helper()
	return Consensus(task, answers(given...), DefaultAgreement())
}

// agreementRule is the default rule at the numbers docs/retirement.md suggests,
// built through the constructor so that every case here runs against a rule the
// constructor admits.
func agreementRule(t *testing.T) Retirement {
	t.Helper()
	rule, err := StopWhenTheyAgree(SuggestedFloor, SuggestedCeiling)
	if err != nil {
		t.Fatalf("the suggested floor and ceiling were refused: %v", err)
	}
	return rule
}

// TestEachRuleInTheVocabularyRecordsWhatFired is the condition the issue puts
// first. Every rule docs/retirement.md offers a campaign owner has a case
// here, and each one is asserted to record which rule fired and the numbers it
// fired on rather than only that the subject retired.
func TestEachRuleInTheVocabularyRecordsWhatFired(t *testing.T) {
	task := plateCondition(t)
	sweep := plateSweep(t)

	fixed, err := EverySubjectTheSameNumberOfTimes(5)
	if err != nil {
		t.Fatalf("a fixed count of 5 was refused: %v", err)
	}

	empty, err := agreementRule(t).StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, SuggestedEmptyAnswerFloor-1)
	if err != nil {
		t.Fatalf("the empty answer exit was refused: %v", err)
	}

	shortened, err := agreementRule(t).LettingTheModelShorten(2)
	if err != nil {
		t.Fatalf("the model exit was refused: %v", err)
	}

	cases := []struct {
		name            string
		rule            Retirement
		labels          []Label
		classifications int
		proposals       []Answer
		fired           string
		numbers         string
	}{
		{
			name:            "every subject the same number of times",
			rule:            fixed,
			labels:          []Label{labelled(t, task, "clear", "marked", "clear", "unusable", "clear")},
			classifications: 5,
			fired:           "every subject the same number of times",
			numbers:         "5 classification(s), and the rule asks for 5",
		},
		{
			name:            "stop as soon as the volunteers agree",
			rule:            agreementRule(t),
			labels:          []Label{labelled(t, task, "clear", "clear", "clear")},
			classifications: 3,
			fired:           "stop as soon as the volunteers agree",
			numbers:         "3 classification(s), the floor is 3, and every task has a label",
		},
		{
			name:            "give up after a while",
			rule:            agreementRule(t),
			labels:          []Label{labelled(t, task, "clear", "clear", "marked", "marked", "unusable", "clear", "marked")},
			classifications: 7,
			fired:           "give up after a while",
			numbers:         "7 classification(s) and the ceiling is 7, with 1 task(s) still without a label",
		},
		{
			name:            "stop early when enough volunteers say there is nothing here",
			rule:            empty,
			labels:          []Label{labelled(t, sweep, "nothing", "nothing")},
			classifications: 2,
			fired:           "stop early when enough volunteers say there is nothing here",
			numbers:         "2 classification(s), and the empty answer reached a label at the empty answer floor of 2",
		},
		{
			name:            "let the model shorten the easy ones",
			rule:            shortened,
			labels:          []Label{labelled(t, task, "clear", "clear")},
			classifications: 2,
			proposals:       []Answer{{"clear"}},
			fired:           "let the model shorten the easy ones",
			numbers:         "2 classification(s), the volunteers agreed on every task, the model proposed the same, and the reduced floor is 2",
		},
	}

	for _, c := range cases {
		got := c.rule.Decide(c.labels, c.classifications, c.proposals)
		if !got.Retired() {
			t.Errorf("%s: the subject did not retire: %s", c.name, got)
			continue
		}
		if got.Rule() != c.fired {
			t.Errorf("%s: recorded %q as the rule that fired", c.name, got.Rule())
		}
		if got.Numbers() != c.numbers {
			t.Errorf("%s: recorded the numbers as %q, want %q", c.name, got.Numbers(), c.numbers)
		}
	}
}

// TestTheFixedCountRuleConsultsNoAgreement holds the property that separates
// the first rule from the second. A subject the volunteers never agreed about
// retires exactly like one they did, and it is not unresolved for it: the
// campaign owner asked for a fixed weight of evidence and got it.
func TestTheFixedCountRuleConsultsNoAgreement(t *testing.T) {
	rule, err := EverySubjectTheSameNumberOfTimes(3)
	if err != nil {
		t.Fatalf("a fixed count of 3 was refused: %v", err)
	}
	task := plateCondition(t)

	disagreed := rule.Decide([]Label{labelled(t, task, "clear", "marked", "unusable")}, 3, nil)
	if !disagreed.Retired() {
		t.Errorf("a subject with the fixed count of answers did not retire: %s", disagreed)
	}
	if disagreed.Unresolved() {
		t.Errorf("the fixed count rule marked a subject unresolved, and it consults no agreement: %s", disagreed)
	}

	short := rule.Decide([]Label{labelled(t, task, "clear", "clear")}, 2, nil)
	if short.Retired() {
		t.Errorf("a subject two answers into a count of three retired: %s", short)
	}
}

// TestTheDefaultRuleNeedsBothTheFloorAndALabelOnEveryTask holds the two halves
// of the default rule against each other. A rule that dropped either half
// would pass a table of subjects that satisfy both.
func TestTheDefaultRuleNeedsBothTheFloorAndALabelOnEveryTask(t *testing.T) {
	rule := agreementRule(t)
	task := plateCondition(t)

	// Agreement below the floor: two unanimous answers, and the floor is 3.
	early := rule.Decide([]Label{labelled(t, task, "clear", "clear")}, 2, nil)
	if early.Retired() {
		t.Errorf("a subject retired two answers into a floor of three: %s", early)
	}

	// The floor with no label: four answers and nothing near the threshold.
	spread := rule.Decide([]Label{labelled(t, task, "clear", "clear", "marked", "unusable")}, 4, nil)
	if spread.Retired() {
		t.Errorf("a subject past the floor with no label retired: %s", spread)
	}

	// Both.
	both := rule.Decide([]Label{labelled(t, task, "clear", "clear", "clear")}, 3, nil)
	if !both.Retired() {
		t.Errorf("a subject at the floor with a label did not retire: %s", both)
	}
	if both.Unresolved() {
		t.Errorf("a subject that retired with a label was marked unresolved: %s", both)
	}
}

// TestASubjectThatCannotReachAgreementRetiresUnresolved is the destination
// docs/retirement.md sends it to: retired, so it stops costing volunteer time,
// and marked, so nobody reads the absence of a label as an oversight.
func TestASubjectThatCannotReachAgreementRetiresUnresolved(t *testing.T) {
	rule := agreementRule(t)
	task := plateCondition(t)

	stuck := rule.Decide([]Label{labelled(t, task,
		"clear", "clear", "marked", "marked", "unusable", "clear", "marked")}, 7, nil)
	if !stuck.Retired() {
		t.Fatalf("a subject at the ceiling with no label kept circulating: %s", stuck)
	}
	if !stuck.Unresolved() {
		t.Errorf("a subject that retired at the ceiling with no label was not marked unresolved: %s", stuck)
	}
	if !strings.Contains(stuck.Numbers(), "without a label") {
		t.Errorf("the record reads %q and does not say what was missing", stuck.Numbers())
	}

	// A subject that collected as many answers as the ceiling and does have a
	// label is retired and is not unresolved. It never reaches the ceiling
	// branch at all: the agreement rule takes it first, which is why the
	// ceiling branch marks unresolved unconditionally.
	settled := rule.Decide([]Label{labelled(t, task,
		"marked", "clear", "clear", "clear", "clear", "clear", "clear")}, 7, nil)
	if !settled.Retired() {
		t.Fatalf("a subject at the ceiling with a label did not retire: %s", settled)
	}
	if settled.Unresolved() {
		t.Errorf("a subject that retired with a label was marked unresolved: %s", settled)
	}
}

// TestTheEmptyAnswerExitFiresOnlyWhereEveryTaskNamesIt holds the third rule.
// It reads the label rather than the answers, so the agreement it requires is
// the consensus rule's and this rule declares no second threshold.
func TestTheEmptyAnswerExitFiresOnlyWhereEveryTaskNamesIt(t *testing.T) {
	sweep := plateSweep(t)
	rule, err := agreementRule(t).StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, 2)
	if err != nil {
		t.Fatalf("the empty answer exit was refused: %v", err)
	}

	early := rule.Decide([]Label{labelled(t, sweep, "nothing", "nothing")}, 2, nil)
	if !early.Retired() {
		t.Errorf("two volunteers agreeing there is nothing here did not retire the subject: %s", early)
	}

	// The other option reaching a label is not the empty answer, so the
	// ordinary floor applies and two answers are not enough.
	something := rule.Decide([]Label{labelled(t, sweep, "something", "something")}, 2, nil)
	if something.Retired() {
		t.Errorf("a subject two answers into a floor of three retired on a label that is not the empty answer: %s", something)
	}

	// No label at the smaller floor is no early exit either.
	split := rule.Decide([]Label{labelled(t, sweep, "nothing", "something")}, 2, nil)
	if split.Retired() {
		t.Errorf("a subject the volunteers disagreed about retired early: %s", split)
	}
}

// TestTheModelNeverBreaksATieAndNeverRescuesADisagreement is the constraint
// docs/retirement.md puts on the fourth rule and the reason it is allowed at
// all: what the model changes is how much evidence was collected rather than
// what it says. The case where a model would help most is exactly the case
// where using it would be inventing evidence.
func TestTheModelNeverBreaksATieAndNeverRescuesADisagreement(t *testing.T) {
	task := plateCondition(t)
	rule, err := agreementRule(t).LettingTheModelShorten(2)
	if err != nil {
		t.Fatalf("the model exit was refused: %v", err)
	}

	// The volunteers agree and the model agrees with them, which is the only
	// case the exit fires in.
	shortened := rule.Decide([]Label{labelled(t, task, "clear", "clear")}, 2, []Answer{{"clear"}})
	if !shortened.Retired() {
		t.Fatalf("a subject the volunteers and the model agreed on did not retire early: %s", shortened)
	}
	if !shortened.Shortened() {
		t.Errorf("a subject retired at the reduced floor was not marked as shortened by the model: %s", shortened)
	}

	// A tie the model would break.
	tie := rule.Decide([]Label{labelled(t, task, "clear", "marked")}, 2, []Answer{{"clear"}})
	if tie.Retired() {
		t.Errorf("the model retired a subject the volunteers were level on: %s", tie)
	}

	// A disagreement the model would rescue.
	spread := rule.Decide([]Label{labelled(t, task, "clear", "marked", "unusable")}, 3, []Answer{{"clear"}})
	if spread.Shortened() {
		t.Errorf("the model shortened a subject the volunteers disagreed about: %s", spread)
	}

	// The volunteers agree and the model does not, so there is nothing to
	// shorten and the ordinary floor applies.
	disagreed := rule.Decide([]Label{labelled(t, task, "clear", "clear")}, 2, []Answer{{"marked"}})
	if disagreed.Retired() {
		t.Errorf("a subject retired at the reduced floor while the model proposed something else: %s", disagreed)
	}

	// A model that proposes nothing at all. This is the case the agreement
	// condition is load bearing for: an unlabelled task derives no options, so
	// a comparison of the label against the proposal alone would find the two
	// equal and would let an empty proposal retire a subject the volunteers
	// disagree about.
	nothing := rule.Decide([]Label{labelled(t, task, "clear", "marked", "unusable")}, 3, []Answer{{}})
	if nothing.Retired() {
		t.Errorf("an empty proposal retired a subject the volunteers disagreed about: %s", nothing)
	}
	if nothing.Shortened() {
		t.Errorf("an empty proposal was recorded as having shortened a subject: %s", nothing)
	}

	// With no model there is no proposal, and the campaign runs on the
	// ordinary floor rather than on a special case, which is #60 by
	// construction.
	cold := rule.Decide([]Label{labelled(t, task, "clear", "clear")}, 2, nil)
	if cold.Retired() {
		t.Errorf("a subject retired at the reduced floor with no proposal at all: %s", cold)
	}
}

// TestACampaignWithSeveralTasksRetiresOnlyWhenEveryTaskIsSatisfied holds the
// sentence that there is no per task retirement and no subject that is half
// retired.
func TestACampaignWithSeveralTasksRetiresOnlyWhenEveryTaskIsSatisfied(t *testing.T) {
	rule := agreementRule(t)
	first := plateCondition(t)
	second := whatIsVisible(t)

	settled := labelled(t, first, "clear", "clear", "clear")
	unsettled := Consensus(second, []Answer{{"star"}, {"galaxy"}, {"plate_defect"}}, DefaultAgreement())

	half := rule.Decide([]Label{settled, unsettled}, 3, nil)
	if half.Retired() {
		t.Errorf("a subject with one task labelled and one not retired: %s", half)
	}
	if !strings.Contains(half.Numbers(), "1 task(s) are without a label") {
		t.Errorf("the record reads %q and does not say how many tasks are missing one", half.Numbers())
	}

	both := Consensus(second, []Answer{{"star"}, {"star"}, {"star"}}, DefaultAgreement())
	whole := rule.Decide([]Label{settled, both}, 3, nil)
	if !whole.Retired() {
		t.Errorf("a subject with a label on every task did not retire: %s", whole)
	}
}

// TestTheNumbersEachConstructorRefuses holds the numbers that would produce a
// rule doing something other than what its name says. Each of these is a
// campaign owner's typo that would otherwise run for a whole campaign.
func TestTheNumbersEachConstructorRefuses(t *testing.T) {
	if _, err := EverySubjectTheSameNumberOfTimes(0); err == nil {
		t.Error("a fixed count of zero was admitted, and it retires a subject before anybody sees it")
	}
	if _, err := StopWhenTheyAgree(0, 7); err == nil {
		t.Error("a floor of zero was admitted")
	}
	if _, err := StopWhenTheyAgree(5, 3); err == nil {
		t.Error("a ceiling below the floor was admitted, and under it no subject reaches the floor")
	}

	rule := agreementRule(t)
	if _, err := rule.StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, SuggestedFloor); err == nil {
		t.Error("an empty answer floor level with the campaign's floor was admitted, and it retires nothing early")
	}
	if _, err := rule.StoppingEarlyOnTheEmptyAnswer(2, []string{"nothing"}, 2); err == nil {
		t.Error("an empty answer naming one option for a campaign of two tasks was admitted")
	}
	if _, err := rule.LettingTheModelShorten(SuggestedFloor); err == nil {
		t.Error("a reduced floor level with the campaign's floor was admitted, and the model would shorten nothing")
	}
	if _, err := rule.LettingTheModelShorten(0); err == nil {
		t.Error("a reduced floor of zero was admitted, and a subject retired there carries no human evidence")
	}

	fixed, err := EverySubjectTheSameNumberOfTimes(3)
	if err != nil {
		t.Fatalf("a fixed count of 3 was refused: %v", err)
	}
	if _, err := fixed.StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, 2); err == nil {
		t.Error("the empty answer exit was put on the rule that consults no agreement")
	}
	if _, err := fixed.LettingTheModelShorten(2); err == nil {
		t.Error("the model exit was put on the rule that consults no agreement")
	}
}

// TestTheSuggestedNumbersForTheTwoExitsAreRefusedBesideTheSuggestedFloor is a
// finding held as a test rather than as a sentence somewhere. Both exits take
// a floor the rule calls smaller or reduced, so both are refused at a number
// level with the campaign's floor. docs/retirement.md suggests a floor of 3
// and an empty answer floor of 3, which are the same number, so its own
// starting numbers cannot be declared together and the exit they describe
// would retire nothing early. #192 is where that is settled, and this case
// exists so the state is in the suite rather than only in a comment.
func TestTheSuggestedNumbersForTheTwoExitsAreRefusedBesideTheSuggestedFloor(t *testing.T) {
	rule := agreementRule(t)

	if _, err := rule.StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, SuggestedEmptyAnswerFloor); err == nil {
		t.Fatalf("the suggested empty answer floor of %d was admitted beside the suggested floor of %d", SuggestedEmptyAnswerFloor, SuggestedFloor)
	}

	// One below it is what a campaign owner has to declare instead, and it is
	// admitted, so the refusal above is about the number rather than about the
	// exit being unavailable.
	if _, err := rule.StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, SuggestedEmptyAnswerFloor-1); err != nil {
		t.Errorf("an empty answer floor one below the campaign's floor was refused: %v", err)
	}
}

// TestARuleCarryingNoNumbersRetiresNothing holds the zero value, which any
// package can build because the type is exported even though its fields are
// not. A ceiling of zero read as a ceiling retires every subject at its first
// read, which is a whole campaign of one-answer labels.
func TestARuleCarryingNoNumbersRetiresNothing(t *testing.T) {
	got := Retirement{}.Decide([]Label{labelled(t, plateCondition(t), "clear")}, 1, nil)
	if got.Retired() {
		t.Fatalf("a rule carrying no numbers retired a subject: %s", got)
	}
	if !strings.Contains(got.Numbers(), "no rule") {
		t.Errorf("the record reads %q and does not say that no rule was declared", got.Numbers())
	}
}

// TestTheRuleWritesItselfTheWayARetirementRecordsIt holds the identity a
// retirement is written with, so that the rule and its numbers cannot reach an
// export as columns that later disagree.
func TestTheRuleWritesItselfTheWayARetirementRecordsIt(t *testing.T) {
	fixed, err := EverySubjectTheSameNumberOfTimes(5)
	if err != nil {
		t.Fatalf("a fixed count of 5 was refused: %v", err)
	}
	if fixed.String() != "same-count@5" {
		t.Errorf("the fixed count rule writes itself as %q", fixed.String())
	}

	rule := agreementRule(t)
	if rule.String() != "agreement@3-7" {
		t.Errorf("the default rule writes itself as %q", rule.String())
	}

	empty, err := rule.StoppingEarlyOnTheEmptyAnswer(1, []string{"nothing"}, 2)
	if err != nil {
		t.Fatalf("the empty answer exit was refused: %v", err)
	}
	if empty.String() != "agreement@3-7+empty@2{nothing}" {
		t.Errorf("the empty answer exit writes itself as %q", empty.String())
	}

	shortened, err := empty.LettingTheModelShorten(2)
	if err != nil {
		t.Fatalf("the model exit was refused: %v", err)
	}
	if shortened.String() != "agreement@3-7+empty@2{nothing}+model@2" {
		t.Errorf("both exits write themselves as %q", shortened.String())
	}
}
