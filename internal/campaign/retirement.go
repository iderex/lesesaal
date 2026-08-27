package campaign

import (
	"fmt"
	"strings"
)

// The retirement rule, which is docs/retirement.md and nothing beyond it. It
// decides when a subject has had enough classifications and stops being handed
// out.
//
// It owns counts and nothing else. Every share, every threshold and every
// statement about whether the volunteers agree comes from the consensus rule
// and is read off the labels handed in here, so a campaign owner who changes
// the agreement threshold changes it in one place.
//
// It reads no store and no clock, for the same reason the consensus rule does
// not. The moment a subject retired is stamped by the caller from Depends.Now,
// because a rule that read the clock could not be a table of cases and would
// be quietly a test of the machine it ran on.

// SuggestedFloor, SuggestedCeiling and SuggestedEmptyAnswerFloor are the
// starting numbers docs/retirement.md offers a campaign owner. That document
// says plainly that they are starting positions rather than results, and that
// what moves them is the pilot in #115 and #116.
const (
	SuggestedFloor            = 3
	SuggestedCeiling          = 7
	SuggestedEmptyAnswerFloor = 3
)

// Retirement is the rule in force for one campaign: which of the two base
// rules docs/retirement.md offers, the numbers it takes, and the two additions
// that may sit on top of the agreement rule.
//
// It is a value rather than a package-level setting because the rule is per
// campaign and a retirement carries the one it fired under.
type Retirement struct {
	// count is the fixed number of classifications the same-number-of-times
	// rule asks for, and is zero under the agreement rule.
	count int

	// floor and ceiling are the agreement rule's own numbers, and are zero
	// under the fixed count rule.
	floor   int
	ceiling int

	// emptyAnswer names one option per task, in the campaign's task order,
	// that the campaign owner calls the empty answer. It is nil where the
	// early exit is off.
	emptyAnswer []string

	// emptyFloor is the smaller floor the empty answer retires at, and is zero
	// where the early exit is off.
	emptyFloor int

	// modelFloor is the reduced floor a subject the model agrees with may
	// retire at, and is zero where the model may not shorten anything, which
	// is the default docs/retirement.md sets.
	modelFloor int
}

// SuggestedRetirement is the default rule at the numbers docs/retirement.md
// suggests. It takes no error because both constants sit inside what
// StopWhenTheyAgree admits, and a caller forced to handle an impossible error
// learns to ignore errors.
func SuggestedRetirement() Retirement {
	return Retirement{floor: SuggestedFloor, ceiling: SuggestedCeiling}
}

// EverySubjectTheSameNumberOfTimes is the first rule docs/retirement.md
// offers: every subject is handed out until it has a fixed number of
// classifications and agreement is not consulted at all.
//
// It is the rule to choose when the campaign is measuring the volunteers as
// much as the plates, because every subject then carries the same weight of
// evidence.
func EverySubjectTheSameNumberOfTimes(count int) (Retirement, error) {
	if count < 1 {
		return Retirement{}, fmt.Errorf("a fixed count of %d is not a number of classifications: it has to be at least 1, and a subject that needs none would retire before anybody saw it", count)
	}
	return Retirement{count: count}, nil
}

// StopWhenTheyAgree is the second rule and the default: a subject retires when
// it has at least a floor of classifications and every task has a label, and
// it retires anyway once it reaches the ceiling.
//
// A ceiling below the floor is refused rather than resolved in either
// direction. Under it every subject retires at the ceiling without the floor
// ever being reached, so the campaign would collect a fixed count while its
// owner believed it was collecting agreement, which is the first rule wearing
// the second one's name.
func StopWhenTheyAgree(floor int, ceiling int) (Retirement, error) {
	if floor < 1 {
		return Retirement{}, fmt.Errorf("a floor of %d is not a number of classifications: it has to be at least 1, and a subject that needs none would retire before anybody saw it", floor)
	}
	if ceiling < floor {
		return Retirement{}, fmt.Errorf("a ceiling of %d is below the floor of %d, so no subject would ever reach the floor and every one of them would retire at the ceiling", ceiling, floor)
	}
	return Retirement{floor: floor, ceiling: ceiling}, nil
}

// StoppingEarlyOnTheEmptyAnswer is the third rule, which sits on the agreement
// rule rather than replacing it. The campaign owner names one option per task
// as the empty answer, and a subject whose empty answer reaches a label at a
// smaller floor retires there.
//
// It is for a collection where most plates carry nothing of interest, which is
// the common case in an archive sweep and is where a campaign's time is
// otherwise spent confirming absences.
//
// The named options are one per task in the campaign's own task order, so a
// list of a different length is refused rather than matched up by position and
// hoped for.
//
// THE SUGGESTED NUMBERS IN THAT DOCUMENT CANNOT BE USED TOGETHER, and this
// function is where a campaign owner meets it. The rule calls this a smaller
// floor, so a number level with the campaign's floor is refused here; the
// suggested starting numbers are a floor of 3 and an empty answer floor of 3,
// which are the same number and would retire nothing early. #192 holds both
// halves of that question, this one and the model exit's.
func (r Retirement) StoppingEarlyOnTheEmptyAnswer(tasks int, options []string, floor int) (Retirement, error) {
	if r.count != 0 {
		return Retirement{}, fmt.Errorf("the empty answer exit reads whether a task has a label, and the fixed count rule does not consult agreement at all, so the two cannot be combined")
	}
	if len(options) != tasks {
		return Retirement{}, fmt.Errorf("the empty answer names %d option(s) for a campaign of %d task(s), and it is one option per task in the campaign's own order", len(options), tasks)
	}
	if floor < 1 {
		return Retirement{}, fmt.Errorf("an empty answer floor of %d is not a number of classifications: it has to be at least 1", floor)
	}
	if floor >= r.floor {
		return Retirement{}, fmt.Errorf("an empty answer floor of %d is not below the campaign's floor of %d, so it would retire nothing early and would only look as though it did", floor, r.floor)
	}
	r.emptyAnswer = append([]string(nil), options...)
	r.emptyFloor = floor
	return r, nil
}

// LettingTheModelShorten is the fourth rule, and it is off unless a campaign
// owner switches it on. A subject whose model proposal agrees with the label
// the volunteers already produced may retire at a reduced floor.
//
// Two of the three constraints docs/retirement.md puts on it are in this
// function. The reduced floor has to be a reduction, so a number at or above
// the campaign's floor is refused rather than accepted and left doing nothing.
// And the rule reads the agreement rule's labels, so it cannot sit on the
// fixed count rule, which consults no agreement. The third constraint is in
// Decide: the model may only shorten a subject whose volunteers already agree,
// so a proposal never breaks a tie and never rescues a disagreement.
//
// WHICH FLOOR THE REDUCED ONE IS MEASURED AGAINST IS TWO SENTENCES IN THAT
// DOCUMENT AND THEY DO NOT AGREE. The rule calls the number a reduced floor,
// which is a number below the campaign's floor. The section on consulting the
// model says there is an absolute floor no model may retire below and that it
// is the campaign's own floor, which is that same number and leaves nothing to
// reduce. What is implemented here is the first, because the second read
// literally makes the whole rule incapable of firing, and #192 is where that
// is settled rather than being decided by this file. That issue carries the
// same question about the empty answer floor above.
func (r Retirement) LettingTheModelShorten(floor int) (Retirement, error) {
	if r.count != 0 {
		return Retirement{}, fmt.Errorf("the model exit reads the label the volunteers produced, and the fixed count rule does not consult agreement at all, so the two cannot be combined")
	}
	if floor < 1 {
		return Retirement{}, fmt.Errorf("a reduced floor of %d is not a number of classifications: it has to be at least 1, and a subject the model retired before any volunteer saw it carries no human evidence at all", floor)
	}
	if floor >= r.floor {
		return Retirement{}, fmt.Errorf("a reduced floor of %d is not below the campaign's floor of %d, so the model would shorten nothing and would only look as though it did", floor, r.floor)
	}
	r.modelFloor = floor
	return r, nil
}

// Decision is what the rule decided about one subject, and it is the record
// docs/retirement.md requires rather than a bare yes or no: which rule fired
// and the numbers it fired on.
//
// The moment is not here. It is stamped by the caller from Depends.Now,
// because this rule reads no clock.
type Decision struct {
	retired    bool
	rule       string
	numbers    string
	unresolved bool
	shortened  bool
}

// Retired says whether the subject stops being handed out.
func (d Decision) Retired() bool { return d.retired }

// Rule is the rule that fired, in the words docs/retirement.md uses for it, and
// is empty where nothing fired.
func (d Decision) Rule() string { return d.rule }

// Numbers is what it fired on, which is the half a campaign owner asks for
// first when a result surprises them.
func (d Decision) Numbers() string { return d.numbers }

// Unresolved says the subject retired without every task reaching a label.
// docs/retirement.md sends it to the stuck list rather than letting it
// circulate forever, and marks it so nobody reads the absence of a label as an
// oversight.
func (d Decision) Unresolved() bool { return d.unresolved }

// Shortened says the model's proposal is why this subject retired when it did.
// Every subject retired that way is marked, so a downstream reader can drop
// them and re-derive the campaign as if the model had not been there.
func (d Decision) Shortened() bool { return d.shortened }

// String writes the decision the way a campaign owner reading the stuck list
// wants it.
func (d Decision) String() string {
	if !d.retired {
		return fmt.Sprintf("not retired: %s", d.numbers)
	}
	out := fmt.Sprintf("retired under %s: %s", d.rule, d.numbers)
	if d.unresolved {
		out += ", unresolved"
	}
	if d.shortened {
		out += ", shortened by the model"
	}
	return out
}

// Decide applies the rule to one subject.
//
// labels is one label per task, in the campaign's task order, each computed by
// Consensus from the answers that subject collected. classifications is how
// many classifications the subject carries, which is the number of people who
// answered rather than the number of answers they gave: a classification
// carries one answer per task, which docs/vocabulary.md fixes and this rule
// takes from there.
//
// proposals is one model proposal per task, in the same order, and is nil
// where there is no model or where the campaign owner has not switched the
// model exit on. A campaign that has never trained a model is not a special
// case here: with no proposal the model exit cannot fire and the rest of the
// rule is unchanged, which is #60 answered by construction.
//
// A campaign with several tasks retires a subject when the rule is satisfied
// for every task. There is no per task retirement and no subject that is half
// retired.
func (r Retirement) Decide(labels []Label, classifications int, proposals []Answer) Decision {
	if r.count != 0 {
		return r.fixedCount(classifications)
	}
	if r.ceiling == 0 {
		// A rule carrying no number cannot reach here from either
		// constructor, both of which refuse one. It can reach here from a
		// zero value, which any package can build because the type is
		// exported even though its fields are not, and a rule that read a
		// ceiling of zero would retire every subject at its first read.
		return Decision{numbers: "no rule: this campaign carries no retirement numbers, and a subject cannot retire under a rule nobody declared"}
	}
	return r.agreement(labels, classifications, proposals)
}

// fixedCount is the rule that consults no agreement. It reads the count and
// nothing else, so a subject the volunteers never agreed about retires here
// exactly like one they did, and it is not unresolved for it: the campaign
// owner asked for a fixed weight of evidence and got it.
func (r Retirement) fixedCount(classifications int) Decision {
	if classifications < r.count {
		return Decision{numbers: fmt.Sprintf("%d of the %d classification(s) this rule asks for", classifications, r.count)}
	}
	return Decision{
		retired: true,
		rule:    "every subject the same number of times",
		numbers: fmt.Sprintf("%d classification(s), and the rule asks for %d", classifications, r.count),
	}
}

// agreement is the default rule and the two exits that may sit on it. The
// order the exits are tried in is the order of the floors they fire at, so a
// subject that satisfies two of them is recorded under the one that actually
// stopped it rather than under whichever branch was written first.
func (r Retirement) agreement(labels []Label, classifications int, proposals []Answer) Decision {
	agreed := everyTaskHasALabel(labels)

	if empty, at := r.emptyAnswerReached(labels, classifications); empty {
		return Decision{
			retired: true,
			rule:    "stop early when enough volunteers say there is nothing here",
			numbers: fmt.Sprintf("%d classification(s), and the empty answer reached a label at the empty answer floor of %d", classifications, at),
		}
	}

	if r.modelWouldShorten(labels, classifications, proposals, agreed) {
		return Decision{
			retired:   true,
			rule:      "let the model shorten the easy ones",
			numbers:   fmt.Sprintf("%d classification(s), the volunteers agreed on every task, the model proposed the same, and the reduced floor is %d", classifications, r.modelFloor),
			shortened: true,
		}
	}

	if agreed && classifications >= r.floor {
		return Decision{
			retired: true,
			rule:    "stop as soon as the volunteers agree",
			numbers: fmt.Sprintf("%d classification(s), the floor is %d, and every task has a label", classifications, r.floor),
		}
	}

	if classifications >= r.ceiling {
		// Unresolved is unconditional here rather than a second reading of
		// whether the tasks agreed, and that is a fact about the branch above
		// rather than a shortcut. A subject reaching this line is at or above
		// the ceiling, the ceiling is never below the floor, and the
		// agreement branch already took every subject at or above the floor
		// that has a label. So nothing with a label on every task can arrive
		// here, and a condition asking again would be one no case could make
		// false.
		return Decision{
			retired:    true,
			rule:       "give up after a while",
			numbers:    fmt.Sprintf("%d classification(s) and the ceiling is %d, with %d task(s) still without a label", classifications, r.ceiling, unlabelled(labels)),
			unresolved: true,
		}
	}

	return Decision{numbers: fmt.Sprintf("%d classification(s), the floor is %d, the ceiling is %d, and %d task(s) are without a label", classifications, r.floor, r.ceiling, unlabelled(labels))}
}

// emptyAnswerReached says whether every task's label is exactly the option the
// campaign owner named as that task's empty answer, at or above the smaller
// floor.
//
// It asks for the label rather than for the answers, so the agreement the
// early exit requires is the consensus rule's, unchanged, and this rule
// declares no second threshold.
func (r Retirement) emptyAnswerReached(labels []Label, classifications int) (bool, int) {
	if r.emptyFloor == 0 || classifications < r.emptyFloor {
		return false, 0
	}
	if len(labels) != len(r.emptyAnswer) {
		return false, 0
	}
	for i, label := range labels {
		options := label.Options()
		if !label.Labelled() || len(options) != 1 || options[0] != r.emptyAnswer[i] {
			return false, 0
		}
	}
	return true, r.emptyFloor
}

// modelWouldShorten says whether the model exit fires, and it fires under all
// three of the conditions docs/retirement.md puts on it: the campaign owner
// switched it on, the volunteers already agree on every task, and the model
// proposed the same label they derived.
//
// The second is the one to read carefully. A proposal never breaks a tie and
// never rescues a disagreement, because the case where a model would help most
// is exactly the case where using it would be inventing evidence.
func (r Retirement) modelWouldShorten(labels []Label, classifications int, proposals []Answer, agreed bool) bool {
	if r.modelFloor == 0 || !agreed || classifications < r.modelFloor {
		return false
	}
	if len(proposals) != len(labels) {
		return false
	}
	for i, label := range labels {
		if !sameSet(label.Options(), proposals[i]) {
			return false
		}
	}
	return true
}

// everyTaskHasALabel is the agreement the default rule asks for. A subject
// with no task has none, because a campaign with no task cannot reach this
// rule from Define and a subject that retired on an empty question would be
// this file inventing an answer.
func everyTaskHasALabel(labels []Label) bool {
	if len(labels) == 0 {
		return false
	}
	for _, label := range labels {
		if !label.Labelled() {
			return false
		}
	}
	return true
}

// unlabelled is how many tasks are still without a label, which is the number
// a campaign owner reading the stuck list wants beside the count.
func unlabelled(labels []Label) int {
	missing := 0
	for _, label := range labels {
		if !label.Labelled() {
			missing++
		}
	}
	return missing
}

// sameSet says whether a derived label and a model proposal name the same
// options. Both are sets, so an identifier named twice on either side is named
// once and the order means nothing.
func sameSet(label []string, proposal Answer) bool {
	named := set(proposal)
	for _, id := range label {
		if !named[id] {
			return false
		}
		delete(named, id)
	}
	return len(named) == 0
}

// String is the rule as a retirement records it, so that the rule and its
// numbers cannot be written into an export as columns that later disagree.
func (r Retirement) String() string {
	if r.count != 0 {
		return fmt.Sprintf("same-count@%d", r.count)
	}
	out := fmt.Sprintf("agreement@%d-%d", r.floor, r.ceiling)
	if r.emptyFloor != 0 {
		out += fmt.Sprintf("+empty@%d{%s}", r.emptyFloor, strings.Join(r.emptyAnswer, " "))
	}
	if r.modelFloor != 0 {
		out += fmt.Sprintf("+model@%d", r.modelFloor)
	}
	return out
}
