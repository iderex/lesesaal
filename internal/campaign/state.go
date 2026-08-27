package campaign

import "fmt"

// The states a campaign passes through and the transitions that are refused,
// which is docs/campaign-states.md and nothing beyond it.
//
// The rule is here rather than in the store for the same reason the consensus
// rule is: it reads no store, no clock and no volunteer, so a table of cases is
// a complete test of it. What the store knows is handed in as counts, the way
// Label.Stale takes the number of classifications behind a label rather than
// going to look for them.

// State is where a campaign is. The four are exhaustive and a campaign is
// always in exactly one of them.
//
// Closed rather than finished, complete or done: docs/vocabulary.md gives
// retired to a subject that needs no more classifications, and the campaign as
// a whole needs a word of its own that does not claim anybody is finished with
// what it collected.
type State string

const (
	// Draft is a campaign no volunteer has seen. Its definition may still
	// change in any way, because nothing has been answered against it.
	Draft State = "draft"

	// Open is a campaign volunteers are classifying.
	Open State = "open"

	// Paused is a campaign that was open and is not taking answers. The
	// campaign owner is still there and the campaign is expected back.
	Paused State = "paused"

	// Closed is a campaign that is a record. It takes no further answer, and
	// what it collected is what it will always have collected unless the
	// campaign owner deliberately reopens it.
	Closed State = "closed"
)

// States is every state, in the order a campaign meets them. It is a function
// rather than a variable because a package-level slice can be written through
// by any caller that holds it.
func States() []State { return []State{Draft, Open, Paused, Closed} }

// Transition is one thing a campaign owner asks for. The value is the verb
// phrase a refusal reads with, so a message is assembled rather than written
// out once per pair and kept level by hand.
type Transition string

const (
	// Opening lets volunteers in for the first time.
	Opening Transition = "be opened to volunteers"

	// Pausing stops a campaign taking answers without ending it.
	Pausing Transition = "be paused"

	// Resuming lets volunteers back in after a pause.
	Resuming Transition = "be resumed"

	// Closing ends a campaign.
	Closing Transition = "be closed"

	// Reopening takes a closed campaign back to open, and is the one
	// transition that takes something back.
	Reopening Transition = "be reopened"

	// AddingSubjects puts more subjects into a campaign, which is the common
	// case and the one an owner does most.
	AddingSubjects Transition = "take more subjects"

	// RemovingSubjects takes subjects back out.
	RemovingSubjects Transition = "give up subjects"

	// Rewording changes what a volunteer reads and nothing about what was
	// answered: a title, the instructions, a question, an option's text.
	// docs/task-grammar.md separates the text from the identifier for exactly
	// this edit and expects a campaign owner to make it while a campaign runs.
	Rewording Transition = "have its wording changed"

	// Redefining changes what an answer means: which tasks a campaign asks,
	// their type, their bounds, or which option identifiers they offer.
	Redefining Transition = "have its tasks redefined"
)

// Transitions is every transition, in the order docs/campaign-states.md tables
// them.
func Transitions() []Transition {
	return []Transition{
		Opening, Pausing, Resuming, Closing, Reopening,
		AddingSubjects, RemovingSubjects, Rewording, Redefining,
	}
}

// Progress is what the store knows about a campaign at the moment a transition
// is asked for. It is counts rather than the things they count, so this rule
// runs with no store behind it, which docs/layout.md requires of the core.
type Progress struct {
	// Subjects is how many subjects the campaign holds.
	Subjects int

	// Classifications is how many classifications have been recorded against
	// the campaign, which is what decides whether an edit is still free.
	Classifications int

	// Answered is how many of the subjects a removal concerns already carry at
	// least one classification. It is read by RemovingSubjects alone, and a
	// campaign owner dropping subjects nobody has seen is not the case that
	// rule exists against.
	Answered int
}

// Effect is what a transition does: the state it leaves the campaign in, and
// the consequence a campaign owner is told about before it happens.
//
// A consequence is not a warning that something may go wrong. It is a thing
// that will happen, stated because the campaign owner is the only person who
// can decide whether it is worth it.
type Effect struct {
	to          State
	consequence string
}

// State is the state the campaign is in afterwards.
func (e Effect) State() State { return e.to }

// Consequence is what the transition does beside moving the state, and is
// empty where it does nothing beside that.
func (e Effect) Consequence() string { return e.consequence }

// TransitionRefused is what Move returns where a transition may not happen.
//
// Lost is what the campaign owner would have lost had it been allowed, which
// is the sentence docs/campaign-states.md requires of a refusal. Where the
// transition would have destroyed nothing and is simply not a move from that
// state, Lost says so and says where the transition does live: naming a loss
// that does not exist would be inventing one, and a message a reader learns to
// discount is worse than a shorter one.
type TransitionRefused struct {
	From       State
	Transition Transition
	Lost       string
}

// Error names the state, the transition and what would have gone.
func (t TransitionRefused) Error() string {
	return fmt.Sprintf("a campaign in the %s state cannot %s: %s", t.From, t.Transition, t.Lost)
}

// Move asks for a transition and says what it does, or refuses it.
//
// Every pair of a state and a transition has an answer here. A pair nobody
// decided would otherwise be settled by whichever branch fell through last,
// which is how a transition that destroys data arrives without anybody
// choosing it.
func Move(from State, transition Transition, progress Progress) (Effect, error) {
	switch transition {
	case Opening:
		return opening(from, progress)
	case Pausing:
		return onlyFrom(from, Open, Pausing, Paused)
	case Resuming:
		return onlyFrom(from, Paused, Resuming, Open)
	case Closing:
		return closing(from, progress)
	case Reopening:
		return reopening(from)
	case AddingSubjects:
		return addingSubjects(from)
	case RemovingSubjects:
		return removingSubjects(from, progress)
	case Rewording:
		return rewording(from)
	case Redefining:
		return redefining(from, progress)
	}

	// A transition of no known name cannot reach here from Transitions, which
	// lists them. It can reach here from a caller that built the string
	// itself, and a rule that panicked on that would turn a typo into a dead
	// process rather than into a refusal somebody can read.
	return Effect{}, TransitionRefused{
		From:       from,
		Transition: transition,
		Lost:       "nothing, because that is not a transition this rule knows; docs/campaign-states.md tables the ones that exist",
	}
}

// onlyFrom is the shape of a transition that is available from exactly one
// state and does nothing but move.
func onlyFrom(from State, available State, transition Transition, to State) (Effect, error) {
	if from != available {
		return notAMoveFromHere(from, transition, available)
	}
	return Effect{to: to}, nil
}

// notAMoveFromHere is the refusal for a transition that destroys nothing and is
// simply not available in this state.
func notAMoveFromHere(from State, transition Transition, available ...State) (Effect, error) {
	names := make([]string, 0, len(available))
	for _, state := range available {
		names = append(names, string(state))
	}
	return Effect{}, TransitionRefused{
		From:       from,
		Transition: transition,
		Lost:       fmt.Sprintf("nothing would have been lost, and this is a move from %s", join(names)),
	}
}

// join writes a short list the way a sentence wants it.
func join(names []string) string {
	switch len(names) {
	case 0:
		return "no state"
	case 1:
		return names[0] + " and from nowhere else"
	}
	out := ""
	for i, name := range names[:len(names)-1] {
		if i > 0 {
			out += ", "
		}
		out += name
	}
	return out + " and " + names[len(names)-1]
}

// opening lets volunteers in. A campaign with no subjects is refused rather
// than opened empty: the selection rule in docs/subject-selection.md draws from
// the eligible set, and a volunteer arriving at a campaign with nothing in it
// is shown what a finished campaign shows.
func opening(from State, progress Progress) (Effect, error) {
	if from != Draft {
		return notAMoveFromHere(from, Opening, Draft)
	}
	if progress.Subjects == 0 {
		return Effect{}, TransitionRefused{
			From:       from,
			Transition: Opening,
			Lost:       "nothing yet, and that is the refusal: this campaign holds no subject, so the first volunteer to arrive would be told there is nothing left for them",
		}
	}
	return Effect{to: Open}, nil
}

// closing ends a campaign. It is available from every state a campaign can be
// worked in, because a campaign owner is entitled to stop, and it carries its
// consequence rather than a confirmation somebody clicks through.
func closing(from State, progress Progress) (Effect, error) {
	switch from {
	case Draft:
		return Effect{
			to:          Closed,
			consequence: "this campaign never opened, so it closes as a definition and a set of subjects with no classification behind them",
		}, nil
	case Open, Paused:
		return Effect{
			to:          Closed,
			consequence: fmt.Sprintf("the %d classification(s) already recorded are what this campaign will have collected, and a subject that had not reached its retirement rule stays unresolved rather than being decided by the closing", progress.Classifications),
		}, nil
	}
	return notAMoveFromHere(from, Closing, Draft, Open, Paused)
}

// reopening is the one transition that takes something back, and it is allowed for
// the reason docs/retirement.md allows a subject to come back: only by the
// campaign owner, only as an explicit action, and never on its own.
//
// What it costs is the export. docs/retirement.md rejects anything that would
// leave a campaign owner holding an export that no longer describes the
// campaign with nothing marking the moment it changed, so the moment is this
// transition and the campaign owner is told what it does.
func reopening(from State) (Effect, error) {
	if from != Closed {
		return notAMoveFromHere(from, Reopening, Closed)
	}
	return Effect{
		to:          Open,
		consequence: "an export already produced from this campaign stops describing it from here, and has to be produced again for the two to tell the same story",
	}, nil
}

// addingSubjects is the common case and is deliberately cheap. A subject added
// to a running campaign starts with no classification and no label, which moves
// nothing about the subjects already there: the consensus rule counts per
// subject and the retirement rule counts per subject, so neither reads a total.
func addingSubjects(from State) (Effect, error) {
	switch from {
	case Draft, Paused:
		return Effect{to: from}, nil
	case Open:
		return Effect{
			to:          Open,
			consequence: "the added subjects carry no classification and no model score, so they are in the position of the campaign's first day until they collect both",
		}, nil
	}
	return Effect{}, TransitionRefused{
		From:       from,
		Transition: AddingSubjects,
		Lost:       "nothing, and that is why it is refused: a closed campaign takes no answer, so the added subjects would sit in it collecting nothing while appearing to be part of what it measured",
	}
}

// removingSubjects is the transition docs/retirement.md answers rather than
// this rule. Nothing is deleted and no subject is quietly dropped, so a subject
// that has been classified stays and the campaign owner retires it instead,
// which stops it costing volunteer time and keeps the answers people gave.
//
// A subject nobody has answered carries no such evidence, and removing it is
// the ordinary correction of a collection that was ingested wrong.
func removingSubjects(from State, progress Progress) (Effect, error) {
	if from == Closed {
		return Effect{}, TransitionRefused{
			From:       from,
			Transition: RemovingSubjects,
			Lost:       "the campaign's own account of what it measured: a closed campaign is a record, and a record that loses a subject after the fact describes a collection nobody classified",
		}
	}
	if progress.Answered > 0 {
		return Effect{}, TransitionRefused{
			From:       from,
			Transition: RemovingSubjects,
			Lost:       fmt.Sprintf("every classification behind the %d subject(s) in this removal that already carry one, and the labels derived from them; docs/retirement.md is what a campaign owner uses instead, which stops a subject being handed out and keeps its answers as evidence", progress.Answered),
		}
	}
	return Effect{to: from}, nil
}

// rewording changes what a volunteer reads and nothing about what was
// answered, which is the separation docs/task-grammar.md makes the option
// identifier for. A campaign owner will rewrite the wording of an option while
// a campaign is running, because the first ten volunteers will misread it, and
// that edit is expected rather than tolerated.
func rewording(from State) (Effect, error) {
	switch from {
	case Draft:
		return Effect{to: Draft}, nil
	case Open, Paused:
		return Effect{
			to:          from,
			consequence: "the volunteers before this edit and the volunteers after it read different words for the same option identifier, and nothing in the export says which of the two a given answer was made against",
		}, nil
	}
	return Effect{}, TransitionRefused{
		From:       from,
		Transition: Rewording,
		Lost:       "the campaign's account of the question it asked: no volunteer will read the new wording, so the only thing this edit changes is what the export claims they were shown",
	}
}

// redefining changes what an answer means, and once an answer exists it is
// refused rather than accommodated. docs/retirement.md already takes that position
// in its own words, where a campaign owner who rewrites the question is told
// the redefinition is a new campaign rather than an edit to this one, and
// docs/task-grammar.md refuses changing an option identifier or removing an
// option that has answers behind it for the same reason.
//
// The alternative, keeping every definition and carrying a version on every
// answer, is defensible and is rejected: it puts a version into the export,
// into the consensus rule and into every count that today reads the answers to
// one task, in return for saving a campaign owner from starting a campaign
// whose question they have just discovered was the wrong one.
func redefining(from State, progress Progress) (Effect, error) {
	if from == Closed {
		return Effect{}, TransitionRefused{
			From:       from,
			Transition: Redefining,
			Lost:       "the campaign's account of the question it asked: a closed campaign is a record of what volunteers answered, and redefining its tasks would make that record describe questions nobody was asked",
		}
	}
	if progress.Classifications > 0 {
		return Effect{}, TransitionRefused{
			From:       from,
			Transition: Redefining,
			Lost:       fmt.Sprintf("the meaning of the %d classification(s) already recorded, which answered the definition as it stands; docs/retirement.md says the redefinition is a new campaign rather than an edit to this one", progress.Classifications),
		}
	}
	return Effect{to: from}, nil
}

// WhatChanges says which of the two edits a campaign owner is asking for: a
// rewording, which changes what a volunteer reads, or a redefinition, which
// changes what an answer means.
//
// It is a comparison of two definitions rather than a flag somebody sets,
// because the two are told apart by what actually moved and not by what the
// person editing believed they were doing. Everything an answer carries is
// structural: which tasks exist and in what order, each task's type and bounds,
// and each task's option identifiers and their order. docs/consensus.md derives
// a label in the order the task declares its options, so a change to that order
// moves the label's own representation and is not a rewording.
func WhatChanges(from Campaign, to Campaign) Transition {
	if len(from.tasks) != len(to.tasks) {
		return Redefining
	}
	for i, was := range from.tasks {
		now := to.tasks[i]
		if was.id != now.id || was.kind != now.kind || was.atLeast != now.atLeast || was.atMost != now.atMost {
			return Redefining
		}
		if len(was.options) != len(now.options) {
			return Redefining
		}
		for j, option := range was.options {
			if option.id != now.options[j].id {
				return Redefining
			}
		}
	}
	return Rewording
}
