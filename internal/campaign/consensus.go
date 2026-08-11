package campaign

import (
	"fmt"
	"strings"
)

// The consensus rule, which is docs/consensus.md and nothing beyond it. It
// turns the answers one subject collected for one task into the single derived
// answer docs/vocabulary.md calls a label.
//
// Two properties are load bearing and both are visible in the signature below.
// It reads no store, no clock, no volunteer identifier and no proposal, so a
// table of cases is a complete test of it. And it is a pure count, so somebody
// holding the export can recompute every label this project published without
// running this program, which is the trade docs/consensus.md says the project
// is making.

// DefaultThreshold is the agreement threshold a campaign gets where its owner
// declares none. docs/consensus.md fixes it at 0.7 and says plainly that this
// is a starting position rather than a result: what moves it is the pilot in
// #115 and #116.
const DefaultThreshold = 0.7

// RuleID is the identity of the rule implemented here, written with every
// label it computes. A campaign owner may change the threshold while a
// campaign runs, and a label computed under the old one has to be readable as
// such rather than silently compared with labels computed under the new one.
// The identity is the half of that which survives a second rule ever existing.
const RuleID = "count-share"

// Rule is the consensus rule in force: its identity and the threshold a share
// has to reach. It is a value rather than a package-level setting because the
// threshold is per campaign and a label carries the one it was computed under.
//
// The campaign definition does not carry a threshold field today, so the rule
// is handed to the computation rather than read off the campaign. Where the
// declared definition grows one, Agreement is what it builds.
type Rule struct {
	threshold float64
}

// Agreement is the rule at the threshold a campaign owner declared.
//
// A threshold outside the interval is refused rather than clamped. Zero or
// below labels every subject with whatever was named most, including a single
// answer nobody agreed with, and above one labels nothing at all; both are a
// campaign that produces confident nonsense or no output, and neither reports
// anything at the moment it is set. docs/consensus.md offers a threshold of 1
// as unanimity, so one is inside the interval and zero is not.
func Agreement(threshold float64) (Rule, error) {
	if !(threshold > 0 && threshold <= 1) {
		return Rule{}, fmt.Errorf("an agreement threshold of %v is not a share: it has to be above 0 and at most 1, and docs/consensus.md offers 1 as unanimity and %v as the default", threshold, DefaultThreshold)
	}
	return Rule{threshold: threshold}, nil
}

// DefaultAgreement is the rule at the default threshold. It takes no error
// because DefaultThreshold is a constant inside the interval Agreement admits,
// and a caller forced to handle an impossible error learns to ignore errors.
func DefaultAgreement() Rule {
	return Rule{threshold: DefaultThreshold}
}

// Threshold is the share an option has to reach.
func (r Rule) Threshold() float64 { return r.threshold }

// ID is the identity of the rule, which is the same for every threshold. The
// threshold is the setting; this is what the setting is a setting of.
func (r Rule) ID() string { return RuleID }

// String is the rule as a label records it, identity and threshold together,
// so that the two cannot be written into an export as separate columns that
// later disagree.
func (r Rule) String() string { return fmt.Sprintf("%s@%g", RuleID, r.threshold) }

// Answer is what one volunteer said about one task for one subject: the option
// identifiers they named, which is what docs/vocabulary.md fixes an answer to
// be. It is a set, so an identifier named twice inside one answer is named
// once, and the order inside it means nothing.
//
// It is a plain slice of identifiers rather than a judged type because an
// answer is not judged here. What reaches this rule may be an answer that
// should never have been accepted, and docs/consensus.md requires the rule to
// be defined for that case rather than to assume clean input: a rule with an
// undefined branch is one that behaves differently depending on which
// unfinished part of the system reached it.
type Answer []string

// Label is the single derived answer for one subject and one task, together
// with the four things docs/consensus.md writes beside it. None of the four is
// derivable later, which is why they are fields rather than a comment about
// how the value was reached.
//
// The fields are unexported for the same reason the judged definition types
// hold theirs that way: the only way to hold a Label whose confidence matches
// its options is to have passed Consensus.
type Label struct {
	// options is the derived set, in the order the task declares its options.
	// A share is not a stable order, because two options at the same share
	// would then be ordered by whatever the count happened to do.
	options []string

	// labelled says whether a label exists at all, which an empty options
	// slice cannot: a choice of several with a lower bound of zero can derive
	// the empty set as its label, and that is a different finding from no
	// label at all.
	labelled bool

	confidence float64
	answers    int
	rule       Rule
	level      bool
}

// Options is the derived set. It returns a copy, so a caller cannot write
// through the result into a label that has already been computed.
func (l Label) Options() []string { return append([]string(nil), l.options...) }

// Labelled says whether the answers produced a label. Read it before reading
// Options or Confidence: a label that does not exist reports an empty set and
// a confidence of zero, and both of those are also legitimate values of a
// label that does.
func (l Label) Labelled() bool { return l.labelled }

// Confidence is the share of answers behind the label, which for a single
// choice is the chosen option's share and for a choice of several is the
// smallest share among the options in the label. A set label is only as well
// supported as its weakest member, and docs/consensus.md rejects the average
// because it would hide exactly the option a reader should doubt.
//
// It is not a probability that the label is correct and docs/consensus.md
// refuses to publish one.
func (l Label) Confidence() float64 { return l.confidence }

// Answers is the number of answers the label was computed from, which is every
// answer that arrived and not only the ones that agreed.
func (l Label) Answers() int { return l.answers }

// Rule is the rule that computed the label, with the threshold in force at the
// time.
func (l Label) Rule() Rule { return l.rule }

// Level says the answers were level at the top, which is the case a campaign
// owner reads differently from the rest. docs/retirement.md carries both to the
// same place for different reasons: a subject the volunteers were level on is
// one the collection is genuinely ambivalent about, and a subject whose answers
// were spread with nothing near the threshold is more often a broken image or a
// question read two ways.
//
// It is reported only where no label was derived. For a choice of several the
// top share is regularly shared by options that all reach the threshold and all
// enter the label, and calling that level would flag the campaign's healthiest
// subjects.
func (l Label) Level() bool { return l.level }

// Consensus counts the answers and derives the label, which is the rule
// docs/consensus.md fixes and the whole of it.
//
// It reads the task for its type, its options and its bounds, the answers for
// the identifiers they name, and the rule for its threshold. It reads nothing
// else: not who answered, not how often they have agreed with anybody before,
// not what the model proposed, not the order the answers arrived in and not the
// time they arrived. That is what makes the result recomputable by hand from
// the export, and it is why an answer arriving after the subject retired needs
// no branch here. A late answer is an answer; the label recomputed with it says
// so in its input count, and whether the subject should have taken it is
// docs/retirement.md's question rather than this one's.
//
// With no answers there is no label and nothing else either, which is
// docs/vocabulary.md's "none until the first answer".
func Consensus(task Task, answers []Answer, rule Rule) Label {
	label := Label{answers: len(answers), rule: rule}
	if len(answers) == 0 {
		return label
	}

	shares := sharePerOption(task, answers)
	switch task.kind {
	case SingleChoice:
		return single(task, shares, label)
	case ChoiceOfSeveral:
		return several(task, shares, label)
	}

	// A task of no known type cannot reach here from Define, which refuses
	// one. It can reach here from a zero value, which any package can build
	// because the type is exported even though its fields are not, and a rule
	// that panicked on it would turn a caller's uninitialised variable into a
	// dead process rather than into an unlabelled subject.
	return label
}

// sharePerOption is the share of answers naming each option the task declares,
// in the order the task declares them.
//
// Every answer that arrived is in the denominator, including one that names
// nothing the task offers. It arrived, so the subject has that many answers,
// and it names nothing the task offers, so it can support no label. Dropping
// such an answer instead would raise every other option's share, which is one
// malformed answer making the rest of them look more agreed than they were.
//
// A single choice admits exactly one option, so an answer to one that does not
// name exactly one of the task's own options supports none of them. A choice
// of several has no such condition, because docs/task-grammar.md decomposes it
// per option and the bounds are judged on the derived set rather than on each
// answer.
func sharePerOption(task Task, answers []Answer) []float64 {
	counted := make([]int, len(task.options))
	for _, answer := range answers {
		named := set(answer)
		if task.kind == SingleChoice && declared(task, named) != 1 {
			continue
		}
		for i, option := range task.options {
			if named[option.id] {
				counted[i]++
			}
		}
	}

	out := make([]float64, len(counted))
	for i, count := range counted {
		out[i] = float64(count) / float64(len(answers))
	}
	return out
}

// declared is how many of the task's own options an answer names.
func declared(task Task, named map[string]bool) int {
	count := 0
	for _, option := range task.options {
		if named[option.id] {
			count++
		}
	}
	return count
}

// set reads one answer as the set of identifiers it names, which is what
// docs/vocabulary.md says an answer is. An identifier repeated inside one
// answer is one identifier: counting it twice would let a single malformed
// submission carry an option past the threshold on its own.
func set(answer Answer) map[string]bool {
	named := make(map[string]bool, len(answer))
	for _, id := range answer {
		named[id] = true
	}
	return named
}

// single derives the label of a single choice: the option named by the largest
// share, if that share reaches the threshold and no other option is level with
// it.
//
// An answer that named no declared option, or named two, counts in the
// denominator and toward no option. A single choice admits exactly one option
// and Bounds reports 1 and 1 for it, so such an answer is one the surface and
// the definition boundary should never have let through, and here it is an
// answer that agreed with nothing.
func single(task Task, shares []float64, label Label) Label {
	top, level := highest(shares)
	if top == 0 {
		// No answer named exactly one of the task's own options. There is no
		// top to be level at, so this is spread rather than ambivalent.
		return label
	}
	if !level && top >= label.rule.threshold {
		label.options = optionsAt(task, shares, top)
		label.labelled = true
		label.confidence = top
		return label
	}
	label.level = level
	return label
}

// several derives the label of a choice of several: the set of options whose
// containing share reaches the threshold, if that set is inside the bounds the
// task declared.
//
// The decomposition is docs/task-grammar.md's. For every option the task
// declares, an answer either contains it or does not, so the count is per
// option and the options do not compete for one share.
func several(task Task, shares []float64, label Label) Label {
	var options []string
	confidence := 0.0
	for i, share := range shares {
		if share >= label.rule.threshold {
			options = append(options, task.options[i].id)
			if confidence == 0 || share < confidence {
				confidence = share
			}
		}
	}

	// A task whose lower bound is zero can derive the empty set as its label,
	// and that is a label rather than the absence of one: the volunteers
	// agreed that none of the options applies. Its confidence stays at zero
	// because docs/consensus.md defines confidence over the options in the
	// label and an empty set has none, so there is no smallest share to
	// report. Reporting anything higher would be this file inventing a number
	// the rule does not produce.

	atLeast, atMost := task.Bounds()
	if len(options) < atLeast || len(options) > atMost {
		// The derived set is not an answer the task admits, so it is not a
		// label either. docs/consensus.md requires this branch to exist: the
		// answers that produced it may include one the boundary should have
		// refused, and a rule assuming clean input is a rule with an undefined
		// branch.
		_, level := highest(shares)
		label.level = level
		return label
	}

	label.options = options
	label.labelled = true
	label.confidence = confidence
	return label
}

// highest is the largest share and whether more than one option holds it.
//
// The shares compared here are counts over one denominator, so two options
// with the same count give the same float and the comparison needs no
// tolerance. A tolerance would be the wrong instrument anyway: it would make
// two options with different counts level at the top.
//
// A top of zero is reported as not level however many options sit at it,
// because options nobody named are not options the volunteers were level on.
func highest(shares []float64) (float64, bool) {
	highest := 0.0
	holders := 0
	for _, share := range shares {
		switch {
		case share > highest:
			highest, holders = share, 1
		case share == highest && share > 0:
			holders++
		}
	}
	return highest, holders > 1
}

// optionsAt is the identifiers of the options holding the given share, in the
// order the task declares them.
func optionsAt(task Task, shares []float64, share float64) []string {
	var ids []string
	for i, each := range shares {
		if each == share {
			ids = append(ids, task.options[i].id)
		}
	}
	return ids
}

// Stale says why a stored label no longer describes what is behind it, and
// returns nothing where it still does.
//
// docs/vocabulary.md stores the label rather than recomputing it on read, and
// pays for that by making staleness detectable rather than assuming it away. A
// label whose recorded input count differs from the number of classifications
// now behind it, or whose recorded rule differs from the campaign's, is stale
// by inspection. This is that inspection, and it is here rather than in the
// store because it compares two things the store does not otherwise have to
// understand.
//
// It reports every reason rather than the first, because a label that is both
// short of answers and computed under an old threshold is recomputed once and
// the second reason would otherwise only surface on the next read.
func (l Label) Stale(classifications int, rule Rule) []string {
	var why []string
	if l.answers != classifications {
		why = append(why, fmt.Sprintf("computed from %d answer(s) and the subject now carries %d", l.answers, classifications))
	}
	if l.rule != rule {
		why = append(why, fmt.Sprintf("computed under %s and the campaign now runs %s", l.rule, rule))
	}
	return why
}

// String writes a label the way a reader checking one by hand wants it: the
// derived set, what it was computed from, and the rule that computed it.
func (l Label) String() string {
	if !l.labelled {
		reason := "spread"
		if l.level {
			reason = "level at the top"
		}
		if l.answers == 0 {
			reason = "no answers"
		}
		return fmt.Sprintf("no label (%s), %d answer(s), %s", reason, l.answers, l.rule)
	}
	return fmt.Sprintf("{%s}, confidence %.2f, %d answer(s), %s", strings.Join(l.options, " "), l.confidence, l.answers, l.rule)
}
