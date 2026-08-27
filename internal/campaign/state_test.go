package campaign

import (
	"strings"
	"testing"
)

// campaignOf judges a declared campaign and hands back what it becomes, so
// that every fixture below is a campaign the definition boundary accepts.
func campaignOf(t *testing.T, declared DeclaredCampaign) Campaign {
	t.Helper()
	defined, err := Define(declared)
	if err != nil {
		t.Fatalf("the fixture campaign is refused by the boundary: %v", err)
	}
	return defined
}

// plateSurvey is the fixture the edit cases move away from: one single choice
// over three options, which is the shape docs/consensus.md works by hand.
func plateSurvey(t *testing.T) DeclaredCampaign {
	t.Helper()
	return DeclaredCampaign{
		ID:           "plate-survey",
		Title:        "The plate survey",
		Instructions: "Look at the plate and say what condition it is in.",
		Tasks: []DeclaredTask{{
			ID:       "t1",
			Question: "What condition is this plate in?",
			Type:     "single-choice",
			Options: []DeclaredOption{
				{ID: "clear", Text: "Clear"},
				{ID: "marked", Text: "Marked"},
				{ID: "unusable", Text: "Unusable"},
			},
		}},
	}
}

// TestEveryStateAndTransitionPairIsDecided is the condition the issue puts
// first: the machine is written down with every transition marked. A pair
// nobody decided is what this refuses, and it refuses it by asking for all of
// them rather than by trusting a table in a document to have been kept level
// with the code.
//
// It asserts the shape of the answer and not which answer it is. What each
// pair decides is the tables below and the document; what this one holds is
// that a pair cannot fall through a switch into whatever the last branch did.
func TestEveryStateAndTransitionPairIsDecided(t *testing.T) {
	for _, from := range States() {
		for _, transition := range Transitions() {
			effect, err := Move(from, transition, Progress{Subjects: 1})
			if err == nil {
				if effect.State() == "" {
					t.Errorf("%s + %q: allowed, and left the campaign in no state", from, transition)
				}
				continue
			}

			refused, ok := err.(TransitionRefused)
			if !ok {
				t.Errorf("%s + %q: refused with %T rather than TransitionRefused", from, transition, err)
				continue
			}
			if refused.Lost == "" {
				t.Errorf("%s + %q: refused and said nothing about what would have been lost", from, transition)
			}
			if refused.From != from || refused.Transition != transition {
				t.Errorf("%s + %q: refusal names %s + %q", from, transition, refused.From, refused.Transition)
			}
			if effect.State() != "" {
				t.Errorf("%s + %q: refused and still returned the state %s", from, transition, effect.State())
			}
		}
	}
}

// TestTheOrdinaryLifeOfACampaign walks the path docs/campaign-states.md tables
// as the one an owner takes: draft, open, paused, open again, closed.
func TestTheOrdinaryLifeOfACampaign(t *testing.T) {
	held := Progress{Subjects: 2000}

	state := Draft
	for _, step := range []struct {
		transition Transition
		to         State
	}{
		{Opening, Open},
		{Pausing, Paused},
		{Resuming, Open},
		{Closing, Closed},
	} {
		effect, err := Move(state, step.transition, held)
		if err != nil {
			t.Fatalf("%s + %q was refused: %v", state, step.transition, err)
		}
		if effect.State() != step.to {
			t.Fatalf("%s + %q left the campaign %s, want %s", state, step.transition, effect.State(), step.to)
		}
		state = effect.State()
	}
}

// TestOpeningACampaignWithNoSubjectsIsRefused holds the one condition on the
// transition that lets volunteers in. An empty campaign opens without error
// and then shows the first volunteer to arrive the message a finished campaign
// shows, which is the failure that is invisible from the owner's side.
func TestOpeningACampaignWithNoSubjectsIsRefused(t *testing.T) {
	if _, err := Move(Draft, Opening, Progress{Subjects: 0}); err == nil {
		t.Fatal("a campaign holding no subject was opened to volunteers")
	}

	effect, err := Move(Draft, Opening, Progress{Subjects: 1})
	if err != nil {
		t.Fatalf("a campaign holding one subject was refused: %v", err)
	}
	if effect.State() != Open {
		t.Errorf("opening left the campaign %s, want %s", effect.State(), Open)
	}
}

// TestAddingSubjectsToARunningCampaignWorks is the condition the issue states
// for the common case. It is allowed while a campaign is being filled, while
// it is open and while it is paused, and it leaves the state where it was.
func TestAddingSubjectsToARunningCampaignWorks(t *testing.T) {
	for _, from := range []State{Draft, Open, Paused} {
		effect, err := Move(from, AddingSubjects, Progress{Subjects: 2000, Classifications: 4000})
		if err != nil {
			t.Errorf("a campaign in the %s state refused more subjects: %v", from, err)
			continue
		}
		if effect.State() != from {
			t.Errorf("adding subjects to a campaign in the %s state left it %s", from, effect.State())
		}
	}

	// Open is the one of the three that carries a consequence, because the
	// added subjects are at the campaign's first day while the rest of it is
	// not, which is what #60 is about.
	effect, err := Move(Open, AddingSubjects, Progress{Subjects: 2000, Classifications: 4000})
	if err != nil {
		t.Fatalf("an open campaign refused more subjects: %v", err)
	}
	if effect.Consequence() == "" {
		t.Error("adding subjects to an open campaign was reported as having no consequence")
	}

	if _, err := Move(Closed, AddingSubjects, Progress{Subjects: 2000}); err == nil {
		t.Error("a closed campaign took more subjects, which would collect nothing")
	}
}

// TestEditingATaskDefinitionAfterAnswersExistIsRefused is the transition the
// issue names as the dangerous one. The chosen behaviour is refusal rather
// than versioning, which docs/retirement.md already states in its own words,
// and this is that behaviour rather than a comment about it.
func TestEditingATaskDefinitionAfterAnswersExistIsRefused(t *testing.T) {
	for _, from := range []State{Draft, Open, Paused} {
		if _, err := Move(from, Redefining, Progress{Subjects: 10, Classifications: 1}); err == nil {
			t.Errorf("a campaign in the %s state with one classification was redefined", from)
		}

		effect, err := Move(from, Redefining, Progress{Subjects: 10})
		if err != nil {
			t.Errorf("a campaign in the %s state with no classification refused a redefinition: %v", from, err)
			continue
		}
		if effect.State() != from {
			t.Errorf("redefining a campaign in the %s state left it %s", from, effect.State())
		}
	}

	if _, err := Move(Closed, Redefining, Progress{Subjects: 10}); err == nil {
		t.Error("a closed campaign was redefined, and it is a record of what was asked")
	}
}

// TestRewordingIsAllowedWhileACampaignRuns is the other half of that
// transition, and the half a machine that refused every edit would get wrong.
// docs/task-grammar.md expects a campaign owner to rewrite an option's wording
// while a campaign is running, and separates the text from the identifier so
// that the edit changes nothing about what was answered.
func TestRewordingIsAllowedWhileACampaignRuns(t *testing.T) {
	running := Progress{Subjects: 2000, Classifications: 4000}

	for _, from := range []State{Draft, Open, Paused} {
		effect, err := Move(from, Rewording, running)
		if err != nil {
			t.Errorf("a campaign in the %s state with answers behind it refused a rewording: %v", from, err)
			continue
		}
		if effect.State() != from {
			t.Errorf("rewording a campaign in the %s state left it %s", from, effect.State())
		}
	}

	// Open and paused carry the consequence and draft does not, because there
	// is nobody who read the old words.
	if consequence := effectOf(t, Open, Rewording, running).Consequence(); consequence == "" {
		t.Error("rewording an open campaign was reported as having no consequence")
	}
	if consequence := effectOf(t, Draft, Rewording, Progress{Subjects: 1}).Consequence(); consequence != "" {
		t.Errorf("rewording a draft was given the consequence %q, and no volunteer has read it", consequence)
	}

	if _, err := Move(Closed, Rewording, running); err == nil {
		t.Error("a closed campaign was reworded, which changes only what the export claims volunteers were shown")
	}
}

// TestRemovingSubjectsThatCarryAnswersIsRefused is where
// docs/retirement.md's rule that nothing is deleted and no subject is quietly
// dropped reaches this machine. A subject nobody has answered is a different
// case and is the ordinary correction of a collection ingested wrong.
func TestRemovingSubjectsThatCarryAnswersIsRefused(t *testing.T) {
	for _, from := range []State{Draft, Open, Paused} {
		if _, err := Move(from, RemovingSubjects, Progress{Subjects: 2000, Classifications: 4000, Answered: 1}); err == nil {
			t.Errorf("a campaign in the %s state gave up a subject that carries a classification", from)
		}

		// The campaign has thousands of answers and none of them is against
		// the subjects being removed, which is the case the count of answered
		// subjects exists to separate from the total.
		effect, err := Move(from, RemovingSubjects, Progress{Subjects: 2000, Classifications: 4000})
		if err != nil {
			t.Errorf("a campaign in the %s state refused to give up subjects nobody has answered: %v", from, err)
			continue
		}
		if effect.State() != from {
			t.Errorf("removing subjects from a campaign in the %s state left it %s", from, effect.State())
		}
	}

	if _, err := Move(Closed, RemovingSubjects, Progress{Subjects: 2000}); err == nil {
		t.Error("a closed campaign gave up a subject, and it is a record of what it measured")
	}
}

// TestReopeningIsAllowedAndSaysWhatItCosts is the transition that undoes
// something. docs/retirement.md allows a subject to come back under exactly
// these conditions, only by the campaign owner and only as an explicit action,
// and rejects anything that leaves an export no longer describing the campaign
// with nothing marking the moment it changed.
func TestReopeningIsAllowedAndSaysWhatItCosts(t *testing.T) {
	effect, err := Move(Closed, Reopening, Progress{Subjects: 2000, Classifications: 4000})
	if err != nil {
		t.Fatalf("a closed campaign refused to reopen: %v", err)
	}
	if effect.State() != Open {
		t.Errorf("reopening left the campaign %s, want %s", effect.State(), Open)
	}
	if !strings.Contains(effect.Consequence(), "export") {
		t.Errorf("reopening said %q, and what it costs is the export", effect.Consequence())
	}

	for _, from := range []State{Draft, Open, Paused} {
		if _, err := Move(from, Reopening, Progress{Subjects: 1}); err == nil {
			t.Errorf("a campaign in the %s state was reopened, and it was never closed", from)
		}
	}
}

// TestARefusalNamesWhatWouldHaveBeenLost holds the last condition of the
// issue over the refusals that protect something. Each of these is a
// transition that would have destroyed a thing the campaign owner can name,
// and the message has to name it rather than saying the transition is not
// allowed.
func TestARefusalNamesWhatWouldHaveBeenLost(t *testing.T) {
	cases := []struct {
		name       string
		from       State
		transition Transition
		progress   Progress
		says       string
	}{
		{
			"redefining under answers",
			Open, Redefining, Progress{Classifications: 4000},
			"4000 classification(s)",
		},
		{
			"removing answered subjects",
			Open, RemovingSubjects, Progress{Classifications: 4000, Answered: 3},
			"3 subject(s)",
		},
		{
			"redefining a record",
			Closed, Redefining, Progress{},
			"questions nobody was asked",
		},
		{
			"adding subjects to a record",
			Closed, AddingSubjects, Progress{},
			"collecting nothing",
		},
		{
			"opening a campaign with nothing in it",
			Draft, Opening, Progress{Subjects: 0},
			"holds no subject",
		},
	}

	for _, c := range cases {
		_, err := Move(c.from, c.transition, c.progress)
		if err == nil {
			t.Errorf("%s: was allowed", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: the refusal reads %q and does not name %q", c.name, err, c.says)
		}
	}
}

// TestARefusalThatDestroysNothingSaysSoAndSaysWhereTheTransitionLives is the
// other kind of refusal, and it is held apart on purpose. Pausing a draft
// destroys nothing, so a message claiming a loss would be inventing one, and
// what the campaign owner needs instead is where the transition does live.
func TestARefusalThatDestroysNothingSaysSoAndSaysWhereTheTransitionLives(t *testing.T) {
	_, err := Move(Draft, Pausing, Progress{Subjects: 1})
	if err == nil {
		t.Fatal("a draft was paused, and no volunteer has seen it")
	}
	if !strings.Contains(err.Error(), "nothing would have been lost") {
		t.Errorf("the refusal reads %q and claims a loss", err)
	}
	if !strings.Contains(err.Error(), string(Open)) {
		t.Errorf("the refusal reads %q and does not say which state pausing is a move from", err)
	}
}

// TestWhatChangesTellsARewordingFromARedefinition is the comparison that
// decides which of the two edits a campaign owner asked for, and it is what
// makes the refusal above land on the right one. Every case here is a real
// edit somebody would make.
func TestWhatChangesTellsARewordingFromARedefinition(t *testing.T) {
	cases := []struct {
		name string
		edit func(DeclaredCampaign) DeclaredCampaign
		want Transition
	}{
		{
			"the title",
			func(c DeclaredCampaign) DeclaredCampaign { c.Title = "The plate survey, second attempt"; return c },
			Rewording,
		},
		{
			"the instructions",
			func(c DeclaredCampaign) DeclaredCampaign { c.Instructions = "Look closely first."; return c },
			Rewording,
		},
		{
			"the question a task asks",
			func(c DeclaredCampaign) DeclaredCampaign {
				c.Tasks[0].Question = "How clear is this plate?"
				return c
			},
			Rewording,
		},
		{
			"the words on an option",
			func(c DeclaredCampaign) DeclaredCampaign {
				c.Tasks[0].Options[1].Text = "Marked, scratched or written on"
				return c
			},
			Rewording,
		},
		{
			"an option identifier",
			func(c DeclaredCampaign) DeclaredCampaign {
				c.Tasks[0].Options[1].ID = "annotated"
				return c
			},
			Redefining,
		},
		{
			"the order of the options",
			func(c DeclaredCampaign) DeclaredCampaign {
				c.Tasks[0].Options[0], c.Tasks[0].Options[2] = c.Tasks[0].Options[2], c.Tasks[0].Options[0]
				return c
			},
			Redefining,
		},
		{
			"an option removed",
			func(c DeclaredCampaign) DeclaredCampaign {
				c.Tasks[0].Options = c.Tasks[0].Options[:2]
				return c
			},
			Redefining,
		},
		{
			"a task added",
			func(c DeclaredCampaign) DeclaredCampaign {
				c.Tasks = append(c.Tasks, DeclaredTask{
					ID:       "t2",
					Question: "Which of these can you see on the plate?",
					Type:     "choice-of-several",
					Options: []DeclaredOption{
						{ID: "star", Text: "A star"},
						{ID: "galaxy", Text: "A galaxy"},
					},
					AtLeast: bound(1),
					AtMost:  bound(2),
				})
				return c
			},
			Redefining,
		},
		{
			"a task identifier",
			func(c DeclaredCampaign) DeclaredCampaign { c.Tasks[0].ID = "condition"; return c },
			Redefining,
		},
	}

	for _, c := range cases {
		was := campaignOf(t, plateSurvey(t))
		now := campaignOf(t, c.edit(plateSurvey(t)))
		if got := WhatChanges(was, now); got != c.want {
			t.Errorf("changing %s is %q, want %q", c.name, got, c.want)
		}
	}

	// A definition compared with itself has changed nothing, and nothing is a
	// rewording rather than a redefinition: an owner who saved the form
	// without touching it is not starting a new campaign.
	was := campaignOf(t, plateSurvey(t))
	if got := WhatChanges(was, was); got != Rewording {
		t.Errorf("a definition compared with itself is %q, want %q", got, Rewording)
	}
}

// TestATransitionThisRuleDoesNotKnowIsRefused holds the branch a caller
// reaches by building the string itself. It is a refusal rather than a panic,
// because a typo that kills the process is worse than one that is reported.
func TestATransitionThisRuleDoesNotKnowIsRefused(t *testing.T) {
	if _, err := Move(Open, Transition("be archived"), Progress{Subjects: 1}); err == nil {
		t.Fatal("a transition this rule does not know was allowed")
	}
}

// effectOf is Move where the test has already asserted the transition is
// allowed and wants the effect.
func effectOf(t *testing.T, from State, transition Transition, progress Progress) Effect {
	t.Helper()
	effect, err := Move(from, transition, progress)
	if err != nil {
		t.Fatalf("%s + %q was refused: %v", from, transition, err)
	}
	return effect
}
