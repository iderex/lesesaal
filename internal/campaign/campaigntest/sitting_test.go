package campaigntest

import (
	"strings"
	"testing"
	"time"

	"github.com/iderex/lesesaal/internal/campaign"
)

// sitting is a run of the campaign given, at numbers that finish. Everything
// about it is written down rather than defaulted, because a number a run
// depended on and nobody wrote is a number the next reader has to go and find.
func sitting(declared campaign.DeclaredCampaign) Sitting {
	return Sitting{
		Definition: declared,
		Subjects:   40,
		Volunteers: 8,
		Agreement:  9,
		Retirement: campaign.SuggestedRetirement(),
		Rule:       campaign.DefaultAgreement(),
		Hold:       15 * time.Minute,
		Between:    20 * time.Second,
		Seed:       1,
	}
}

// TestAWholeCampaignRunsWithNoImagesNoBrowserNoModelAndNoNetwork is #41's first
// condition. What makes it true is not asserted here but in
// harness_test.go and layout_test.go at the module root, which refuse the core
// an import of anything and refuse every file of this tree a read of the
// runtime. What is asserted here is that a campaign of forty subjects reaches
// its end: every subject retired, and the numbers on top of that consistent
// with the classifications that produced them.
func TestAWholeCampaignRunsWithNoImagesNoBrowserNoModelAndNoNetwork(t *testing.T) {
	for _, declared := range Fixtures() {
		result, err := sitting(declared).Run()
		if err != nil {
			t.Fatalf("the campaign %q did not run: %v", declared.ID, err)
		}

		summary := result.Summary
		if summary.Subjects != 40 {
			t.Errorf("%s: the campaign held %d subject(s), want 40", declared.ID, summary.Subjects)
		}
		if summary.Retired != 40 {
			t.Errorf("%s: %d subject(s) retired, want 40. The campaign did not finish: %s", declared.ID, summary.Retired, summary)
		}
		if summary.Classifications != len(result.Classifications) {
			t.Errorf("%s: the summary counts %d classification(s) and the run recorded %d", declared.ID, summary.Classifications, len(result.Classifications))
		}
		if summary.Classifications < 40*campaign.SuggestedFloor {
			t.Errorf("%s: %d classification(s) for 40 subjects at a floor of %d", declared.ID, summary.Classifications, campaign.SuggestedFloor)
		}
		if summary.Elapsed <= 0 {
			t.Errorf("%s: the run took %s of campaign time", declared.ID, summary.Elapsed)
		}

		// Every classification is one volunteer, one subject, and one answer
		// per task, which is what docs/vocabulary.md fixes a classification to
		// be. And no volunteer answers one subject twice, which is what makes
		// the answers behind a label independent.
		tasks := len(result.Campaign.Tasks())
		seen := map[string]bool{}
		for _, one := range result.Classifications {
			if len(one.Answers) != tasks {
				t.Fatalf("%s: a classification carries %d answer(s) for %d task(s)", declared.ID, len(one.Answers), tasks)
			}
			pair := one.Volunteer + "/" + one.Subject
			if seen[pair] {
				t.Fatalf("%s: %s answered %s twice", declared.ID, one.Volunteer, one.Subject)
			}
			seen[pair] = true
		}
	}
}

// TestEveryQuestionTypeInTheGrammarHasAFixtureCampaign is #41's second
// condition. Fixtures is derived from campaign.TaskTypes rather than written
// beside it, so a type added to the grammar with no fixture is a failure here
// rather than a gap that leaves the driver covering two thirds of the grammar
// and reporting the same green.
func TestEveryQuestionTypeInTheGrammarHasAFixtureCampaign(t *testing.T) {
	fixtures := Fixtures()
	types := campaign.TaskTypes()
	if len(fixtures) != len(types) {
		t.Fatalf("the grammar declares %d task type(s) and there are %d fixture campaign(s)", len(types), len(fixtures))
	}
	for i, declared := range fixtures {
		defined, err := campaign.Define(declared)
		if err != nil {
			t.Fatalf("the fixture campaign %q does not start: %v", declared.ID, err)
		}
		asked := map[campaign.TaskType]bool{}
		for _, task := range defined.Tasks() {
			asked[task.Type()] = true
		}
		if !asked[types[i]] {
			t.Errorf("the fixture for %q asks no question of that type", types[i])
		}
	}
}

// TestTheRunIsIdenticalUnderOneSeed is #41's third condition. The summary is
// compared as a string so that a difference is a line rather than a diff of two
// structures, and the classifications are compared one by one so that a run
// that reached the same numbers by a different route is a failure rather than a
// pass.
func TestTheRunIsIdenticalUnderOneSeed(t *testing.T) {
	for _, declared := range Fixtures() {
		first, err := sitting(declared).Run()
		if err != nil {
			t.Fatalf("the first run of %q failed: %v", declared.ID, err)
		}
		second, err := sitting(declared).Run()
		if err != nil {
			t.Fatalf("the second run of %q failed: %v", declared.ID, err)
		}

		if first.Summary.String() != second.Summary.String() {
			t.Errorf("%s: two runs under one seed reported\n%s\nand\n%s", declared.ID, first.Summary, second.Summary)
		}
		if len(first.Classifications) != len(second.Classifications) {
			t.Fatalf("%s: two runs recorded %d and %d classification(s)", declared.ID, len(first.Classifications), len(second.Classifications))
		}
		for i := range first.Classifications {
			if transcript(first.Classifications[i]) != transcript(second.Classifications[i]) {
				t.Fatalf("%s: classification %d was %q and then %q", declared.ID, i,
					transcript(first.Classifications[i]), transcript(second.Classifications[i]))
			}
		}
		if strings.Join(first.Stuck(), ",") != strings.Join(second.Stuck(), ",") {
			t.Errorf("%s: the stuck list moved between runs", declared.ID)
		}
	}
}

// TestASecondSeedProducesADifferentRun is what the condition above cannot say
// on its own. A driver that ignored the seed entirely, or one whose draw was
// stuck, would pass every determinism assertion in this suite and would be
// worthless for comparing a rule against its previous behaviour.
func TestASecondSeedProducesADifferentRun(t *testing.T) {
	one := sitting(PlateCondition())
	other := sitting(PlateCondition())
	other.Seed = 2

	first, err := one.Run()
	if err != nil {
		t.Fatalf("the first run failed: %v", err)
	}
	second, err := other.Run()
	if err != nil {
		t.Fatalf("the second run failed: %v", err)
	}

	same := true
	for i := range min(len(first.Classifications), len(second.Classifications)) {
		if transcript(first.Classifications[i]) != transcript(second.Classifications[i]) {
			same = false
			break
		}
	}
	if same && len(first.Classifications) == len(second.Classifications) {
		t.Error("two seeds produced the same run, so the seed decides nothing")
	}
}

// TestTheSummaryAgreesWithWhatTheRunRecorded is #41's fourth condition as far
// as it can be decided here: the numbers a campaign owner would read are
// derived from the classifications and the labels rather than counted a second
// way. A summary computed beside the run rather than from it is the defect this
// exists against, and it is invisible until the two disagree in production.
//
// It runs over three sittings on purpose. In a campaign where everybody agrees,
// the number of labels, the number of subjects and the number of retirements
// are all the same number, so a summary counting any of them in place of
// another agrees with itself and proves nothing. The disagreeing run is where
// those three come apart.
func TestTheSummaryAgreesWithWhatTheRunRecorded(t *testing.T) {
	agreeable := sitting(PlateCondition())
	disagreeing := sitting(PlateCondition())
	disagreeing.Agreement = 3
	for _, run := range []Sitting{sitting(PlateDefects()), agreeable, disagreeing} {
		result, err := run.Run()
		if err != nil {
			t.Fatalf("the campaign %q did not run: %v", run.Definition.ID, err)
		}
		checkSummary(t, result)
	}
}

// checkSummary reads every number in the summary back out of what the run
// actually recorded.
func checkSummary(t *testing.T, result Result) {
	t.Helper()

	labelled, level := 0, 0
	total := 0.0
	for _, labels := range result.Labels {
		for _, label := range labels {
			if label.Level() {
				level++
			}
			if label.Labelled() {
				labelled++
				total += label.Confidence()
			}
		}
	}
	if result.Summary.Labelled != labelled {
		t.Errorf("the summary counts %d label(s) and the run holds %d", result.Summary.Labelled, labelled)
	}
	if result.Summary.Level != level {
		t.Errorf("the summary counts %d level label(s) and the run holds %d", result.Summary.Level, level)
	}
	if labelled > 0 {
		want := total / float64(labelled)
		if difference(result.Summary.Confidence, want) > 1e-9 {
			t.Errorf("the summary reports a confidence of %v, want %v", result.Summary.Confidence, want)
		}
	}

	retired, unresolved := 0, 0
	for _, decision := range result.Decisions {
		if decision.Retired() {
			retired++
		}
		if decision.Unresolved() {
			unresolved++
		}
	}
	if result.Summary.Retired != retired {
		t.Errorf("the summary counts %d retired and the run holds %d", result.Summary.Retired, retired)
	}
	if result.Summary.Unresolved != unresolved {
		t.Errorf("the summary counts %d unresolved and the run holds %d", result.Summary.Unresolved, unresolved)
	}
	if len(result.Stuck()) != unresolved {
		t.Errorf("the stuck list holds %d subject(s) and %d retired unresolved", len(result.Stuck()), unresolved)
	}
	if result.Summary.Shortened != 0 {
		t.Errorf("%d subject(s) were shortened by a model in a run with no model", result.Summary.Shortened)
	}
}

// TestVolunteersWhoDisagreeLeaveSubjectsUnresolved is the ending the driver
// exists to reach as much as the tidy one. docs/retirement.md sends a subject
// that reaches the ceiling without a label to the stuck list and marks it, and
// a driver that could only produce clean campaigns would never exercise that.
func TestVolunteersWhoDisagreeLeaveSubjectsUnresolved(t *testing.T) {
	run := sitting(PlateCondition())
	run.Agreement = 3
	run.Volunteers = 8

	result, err := run.Run()
	if err != nil {
		t.Fatalf("the campaign did not run: %v", err)
	}
	if result.Summary.Unresolved == 0 {
		t.Fatalf("no subject retired unresolved at an agreement of 3 in ten: %s", result.Summary)
	}
	if result.Summary.Retired != result.Summary.Subjects {
		t.Errorf("%d of %d subject(s) retired, and a campaign whose subjects disagree still ends at the ceiling",
			result.Summary.Retired, result.Summary.Subjects)
	}
	for _, id := range result.Stuck() {
		decision := result.Decisions[id]
		if !decision.Unresolved() {
			t.Errorf("%s is on the stuck list and its decision does not say unresolved: %s", id, decision)
		}
		if decision.Numbers() == "" {
			t.Errorf("%s retired unresolved and the decision reports no numbers", id)
		}
	}
}

// TestACampaignEndsWhenEverybodyHasAnsweredEverything is the second of the two
// endings docs/subject-selection.md separates, and the one a run that only ever
// finished campaigns would never reach. Three volunteers cannot retire a
// subject that needs a floor of five, so the run stops with work left rather
// than looping forever.
func TestACampaignEndsWhenEverybodyHasAnsweredEverything(t *testing.T) {
	stopping, err := campaign.StopWhenTheyAgree(5, 9)
	if err != nil {
		t.Fatalf("the retirement rule was refused: %v", err)
	}
	run := sitting(PlateCondition())
	run.Subjects = 4
	run.Volunteers = 3
	run.Retirement = stopping

	result, err := run.Run()
	if err != nil {
		t.Fatalf("the campaign did not run: %v", err)
	}
	if result.Summary.Retired != 0 {
		t.Errorf("%d subject(s) retired under a floor of 5 with 3 volunteers", result.Summary.Retired)
	}
	if want := 12; result.Summary.Classifications != want {
		t.Errorf("the run recorded %d classification(s), want %d: three volunteers answering four subjects each",
			result.Summary.Classifications, want)
	}
}

// TestASittingThatWouldMeanNothingIsRefused holds the constructor. Each of
// these runs to completion and reports numbers, and every one of those numbers
// is about a campaign nobody would have set up.
func TestASittingThatWouldMeanNothingIsRefused(t *testing.T) {
	whole := sitting(PlateCondition())
	broken := map[string]func(*Sitting){
		"no subjects":        func(s *Sitting) { s.Subjects = 0 },
		"no volunteers":      func(s *Sitting) { s.Volunteers = 0 },
		"agreement above 10": func(s *Sitting) { s.Agreement = 11 },
		"agreement below 0":  func(s *Sitting) { s.Agreement = -1 },
		"no hold":            func(s *Sitting) { s.Hold = 0 },
		"no consensus rule":  func(s *Sitting) { s.Rule = campaign.Rule{} },
	}
	for name, break_ := range broken {
		run := whole
		break_(&run)
		if _, err := run.Run(); err == nil {
			t.Errorf("a sitting with %s was accepted", name)
		}
	}
	if _, err := whole.Run(); err != nil {
		t.Errorf("the whole sitting was refused: %v", err)
	}
}

// TestAFixtureSubjectCarriesNoBytes is #41's "no image bytes" read off the
// fixtures rather than off the absence of a failure. A fixture that grew a
// payload would still run, and the suite would slow down for months before
// anybody asked why.
func TestAFixtureSubjectCarriesNoBytes(t *testing.T) {
	subjects, err := Plates(3, time.Time{})
	if err != nil {
		t.Fatalf("the fixture subjects were refused: %v", err)
	}
	if len(subjects) != 3 {
		t.Fatalf("%d fixture subject(s) were built, want 3", len(subjects))
	}
	for _, subject := range subjects {
		derived := subject.Derived()
		if derived.Bytes != 1 || derived.Width != 1 || derived.Height != 1 {
			t.Errorf("%s carries %d byte(s) at %dx%d, and a fixture subject stands for no file",
				subject.ID(), derived.Bytes, derived.Width, derived.Height)
		}
		if !derived.Entered.Equal(Epoch) {
			t.Errorf("%s entered at %s, want the epoch this harness starts at", subject.ID(), derived.Entered)
		}
	}
	if subjects[0].Digest() == subjects[1].Digest() {
		t.Error("two fixture subjects carry one digest, so nothing could tell them apart")
	}
}

// transcript writes one classification the way a comparison between two runs
// wants it: who, which subject, and what they said, in one line.
func transcript(one Classification) string {
	said := make([]string, 0, len(one.Answers))
	for _, answer := range one.Answers {
		said = append(said, strings.Join(answer, "+"))
	}
	return one.Volunteer + " " + one.Subject + " " + strings.Join(said, "|") + " at " + one.At.String()
}

// difference is the distance between two shares, so a comparison of means does
// not turn on the last bit of a division.
func difference(a float64, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
