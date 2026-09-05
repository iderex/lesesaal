package campaign

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// The write path: what one volunteer submits about one subject, and the rule
// that makes a repeated submission produce one classification rather than two.
// Decided in issue #34.
//
// The reason this is a rule rather than a schema constraint is what a duplicate
// looks like downstream. A second copy of one volunteer's answer is
// indistinguishable from a second volunteer agreeing, so it raises the count
// the retirement rule reads and the share the consensus rule computes, and both
// of them then say the campaign has more evidence than it collected. Nothing
// later can tell the two apart, which is why the refusal is here rather than in
// a cleanup somebody runs afterwards.
//
// It reads no clock and no store, for the same reason the consensus rule and
// the retirement rule do not. The moment is passed in from Depends.Now, and the
// write itself is a function the caller supplies, so this package holds the
// order the two things happen in without knowing what the second one is.

// NotRecorded is what Record returns where a submission is refused, which is a
// defect in what the client sent rather than a failure of anything downstream.
// It is a type of this rule's own rather than the definition boundary's
// Refused, because the two are read by the same campaign owner at different
// moments and a message opening "this campaign definition does not start" in
// answer to a posted answer sends them to the wrong file.
//
// Says carries every reason rather than the first, for the reason Define
// reports all of its refusals: a campaign owner writing their own client
// otherwise learns the shape one round trip at a time.
type NotRecorded struct {
	Says []string
}

// Error writes one reason per line.
func (n NotRecorded) Error() string {
	return fmt.Sprintf("this submission was not recorded:\n%s", strings.Join(n.Says, "\n"))
}

// Submission is what arrives from a volunteer: the answers, who gave them,
// which subject they are about, and the key that makes a retry safe.
//
// The key is the client's rather than this package's. A key minted here would
// be minted once per arrival, so the second arrival of one answer would get a
// second key and the two would be different classifications, which is the whole
// failure this type exists against.
type Submission struct {
	// Key is what the client sends with the answer and sends again with a
	// retry of it. docs/vocabulary.md makes the classification the unit that
	// arrives, so the key names one classification and not one answer.
	Key string

	// Volunteer is who answered. It is an identifier this package neither
	// reads nor interprets: what it stands for is #13's inventory and the
	// consensus rule reads none of it.
	Volunteer string

	// Subject is the subject identifier the volunteer was handed.
	Subject string

	// Answers is one answer per task in the campaign, in the campaign's own
	// task order. docs/vocabulary.md fixes a classification at exactly that,
	// so a submission carrying fewer is not a partial classification, it is
	// not one.
	Answers []Answer
}

// Classification is one volunteer's submission about one subject, recorded.
//
// Its fields are unexported for the reason the judged definition types hold
// theirs that way: the only way to hold one is to have been through Record, so
// a Classification in a caller's hand is one the rule admitted rather than one
// somebody assembled.
type Classification struct {
	key       string
	volunteer string
	subject   string
	answers   []Answer
	at        time.Time
}

// Key is the client's key this classification was recorded under.
func (c Classification) Key() string { return c.key }

// Volunteer is who answered.
func (c Classification) Volunteer() string { return c.volunteer }

// Subject is the subject it is about.
func (c Classification) Subject() string { return c.subject }

// Answers is one answer per task, in the campaign's task order. It returns a
// copy for the reason every other reader in this package does: a caller holding
// the slice could otherwise rewrite a recorded answer in place.
func (c Classification) Answers() []Answer {
	out := make([]Answer, len(c.answers))
	for i, answer := range c.answers {
		out[i] = append(Answer(nil), answer...)
	}
	return out
}

// At is the moment it was recorded, stamped by the caller from Depends.Now.
func (c Classification) At() time.Time { return c.at }

// String writes one the way somebody reading a log of them wants it.
func (c Classification) String() string {
	said := make([]string, 0, len(c.answers))
	for _, answer := range c.answers {
		said = append(said, "{"+strings.Join(answer, " ")+"}")
	}
	return fmt.Sprintf("%s answered %s with %s under key %s", c.volunteer, c.subject, strings.Join(said, " "), c.key)
}

// Classifications is the write path of one campaign: which keys have been
// recorded, what each one recorded, and which subjects each volunteer has
// answered.
//
// It is in memory here and it is not the durable copy. What makes a
// classification durable is the write the caller hands to Record, and this type
// is the order that write happens in and the decision about what a second
// arrival means. #33 chooses the store behind that write.
//
// WHAT DURABLE MEANS HERE, and it is the position issue #34 asks to be stated
// rather than left to be discovered. A submission is acknowledged only after
// the write returns without error. A write that fails registers nothing at all:
// the key stays free, so the volunteer's retry under the same key is a first
// arrival and writes. The alternative, registering the key before the write so
// that a retry is refused as a duplicate, was rejected because it turns a
// failed write into a permanently lost answer that the volunteer has been told
// nothing about and cannot resend.
//
// The cost of the position is stated rather than hidden. A write that succeeds
// and whose acknowledgement never reaches the volunteer is retried under the
// same key, and this rule returns the classification already recorded, so the
// volunteer sees their work counted. A write that succeeds and whose success is
// not observed by this process at all - the process dies between the store
// committing and Record returning - leaves the key registered nowhere and the
// row in the store, so a retry writes a second row under a key the store
// already holds. That window belongs to the store rather than to this rule, and
// #33 is where it is closed by making the key the store's own unique column.
type Classifications struct {
	// mu makes the read of the two registers and the write of the answer one
	// step. Two volunteers submitting at once is the ordinary case, and one
	// volunteer's retry racing their own first attempt is the case this type
	// exists for, so the decision has to be taken under a lock rather than
	// composed out of two lookups.
	mu sync.Mutex

	// repeats is whether this campaign lets one volunteer answer one subject
	// more than once. It is off unless the campaign says otherwise, which is
	// what docs/subject-selection.md assumes when it excludes what a volunteer
	// has answered from their eligible set.
	repeats bool

	// byKey is every classification recorded, under the client's key. It is
	// what makes a retry return the first result instead of writing again.
	byKey map[string]Classification

	// answered is volunteer and subject to the key that answered it. It is the
	// second rule rather than the first: a volunteer answering one subject
	// twice under two different keys is not a retry, it is a second answer,
	// and the two are refused for different reasons and with different words.
	answered map[pair]string

	// count is subject to how many classifications it carries, which is what
	// the retirement rule reads. It is kept here rather than counted over
	// byKey on every read because Record is on the path of every submission
	// and the retirement check runs on the same path.
	count map[string]int
}

// pair is one volunteer and one subject, which is what the second rule is about.
type pair struct {
	volunteer string
	subject   string
}

// OneAnswerPerVolunteerPerSubject builds the write path of a campaign where a
// volunteer answers each subject once, which is every campaign unless its owner
// says otherwise.
func OneAnswerPerVolunteerPerSubject() *Classifications {
	return &Classifications{
		byKey:    map[string]Classification{},
		answered: map[pair]string{},
		count:    map[string]int{},
	}
}

// LettingAVolunteerAnswerAgain builds the write path of a campaign that admits
// a second answer from the same volunteer about the same subject.
//
// It is a separate constructor rather than a flag a caller sets afterwards,
// because the difference is a campaign owner's decision made before anybody
// answers and never a thing that changes under a campaign that is running: the
// answers already collected were collected under one of the two rules, and a
// campaign that switched would carry both kinds with nothing marking which.
//
// What it does not switch off is the key. A retry is still a retry under either
// rule, so a second arrival of one submission is still one classification, and
// what this rule admits is a second submission with a key of its own.
func LettingAVolunteerAnswerAgain() *Classifications {
	c := OneAnswerPerVolunteerPerSubject()
	c.repeats = true
	return c
}

// Record takes one submission, or says why it was not taken.
//
// written is false where the key had already been recorded, and the
// classification returned is the one recorded the first time. That is the
// difference between this and a refusal: a retry is a success the caller
// acknowledges exactly as it would the first arrival, because from the
// volunteer's side those are the same click.
//
// write is what makes the classification durable and is called at most once per
// submission. It is a function rather than a store this package imports,
// because docs/layout.md points the dependency the other way: the store may
// read this package's words and this package may not read the store's.
//
// now is stamped by the caller from Depends.Now, for the reason nothing else
// here reads a clock.
func (c *Classifications) Record(submission Submission, now time.Time, write func(Classification) error) (recorded Classification, written bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if says := shapeOf(submission); len(says) > 0 {
		return Classification{}, false, NotRecorded{Says: says}
	}

	if already, taken := c.byKey[submission.Key]; taken {
		if !already.sameSubmissionAs(submission) {
			return Classification{}, false, NotRecorded{Says: []string{
				fmt.Sprintf("the key %q was recorded for a different answer, so this is a key reused rather than a submission retried, and recording it would replace an answer somebody gave", submission.Key),
			}}
		}
		return already, false, nil
	}

	who := pair{volunteer: submission.Volunteer, subject: submission.Subject}
	if under, answered := c.answered[who]; answered && !c.repeats {
		return Classification{}, false, NotRecorded{Says: []string{
			fmt.Sprintf("%q has already answered %q, under key %q, and a second answer from one volunteer would be counted as a second volunteer agreeing", submission.Volunteer, submission.Subject, under),
		}}
	}

	classification := Classification{
		key:       submission.Key,
		volunteer: submission.Volunteer,
		subject:   submission.Subject,
		answers:   copyAnswers(submission.Answers),
		at:        now,
	}

	// The write comes before anything is registered. A failure here leaves
	// every register as it was, so the retry that follows is a first arrival
	// and writes, which is the position the type comment states.
	if err := write(classification); err != nil {
		return Classification{}, false, fmt.Errorf("the answer %q gave about %q was not stored, and is not acknowledged: %w", submission.Volunteer, submission.Subject, err)
	}

	c.byKey[classification.key] = classification
	c.answered[who] = classification.key
	c.count[classification.subject]++

	return classification, true, nil
}

// Count is how many classifications a subject carries, which is the number the
// retirement rule takes.
func (c *Classifications) Count(subject string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count[subject]
}

// Answered names the subjects one volunteer has answered, which is the set
// Pool.Answered holds for the draw. It is built here rather than by the caller
// counting rows, so the exclusion the selection rule makes and the refusal this
// rule makes read the same register and cannot part.
func (c *Classifications) Answered(volunteer string) map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := map[string]bool{}
	for who := range c.answered {
		if who.volunteer == volunteer {
			out[who.subject] = true
		}
	}
	return out
}

// Of is every classification recorded about one subject, which is what the
// consensus rule's answers are taken from. The order is the order they were
// recorded in, so a campaign replayed from the same submissions produces the
// same reading.
func (c *Classifications) Of(subject string) []Classification {
	c.mu.Lock()
	defer c.mu.Unlock()

	var out []Classification
	for _, classification := range c.byKey {
		if classification.subject == subject {
			out = append(out, classification)
		}
	}
	sortByMomentThenKey(out)
	return out
}

// AnswersTo is the answer set of one task for one subject, which is exactly
// what Consensus takes. at is the task's position in the campaign's task order,
// which is the position its answer occupies in every classification.
func (c *Classifications) AnswersTo(subject string, at int) []Answer {
	classifications := c.Of(subject)

	out := make([]Answer, 0, len(classifications))
	for _, classification := range classifications {
		if at < 0 || at >= len(classification.answers) {
			continue
		}
		out = append(out, append(Answer(nil), classification.answers[at]...))
	}
	return out
}

// shapeOf refuses a submission that is not one, and reports every refusal
// rather than the first for the reason Define does.
//
// A campaign owner debugging a client of their own needs to know which of these
// it was, which is the requirement issue #40 states for the boundary this sits
// behind. What that boundary judges and this does not is whether the answers
// name options the tasks declare: that needs the campaign definition, and this
// rule reads none.
func shapeOf(submission Submission) []string {
	var says []string

	if submission.Key == "" {
		says = append(says, "it carries no key, and a submission with no key cannot be recorded once: a retry of it is indistinguishable from a second answer")
	}
	if submission.Volunteer == "" {
		says = append(says, "it names no volunteer, and the rule that keeps one person from answering one subject twice has nobody to hold it against")
	}
	if submission.Subject == "" {
		says = append(says, "it names no subject, so there is nothing for the retirement rule to count it towards")
	}
	if len(submission.Answers) == 0 {
		says = append(says, "it carries no answer, and docs/vocabulary.md makes a classification exactly one answer per task, so this is not a partial classification, it is not one")
	}

	return says
}

// sameSubmissionAs is whether a second arrival under one key is the retry it
// claims to be. It compares everything the classification carries except the
// moment, which moves between a first attempt and a retry by definition.
func (c Classification) sameSubmissionAs(submission Submission) bool {
	if c.volunteer != submission.Volunteer || c.subject != submission.Subject {
		return false
	}
	if len(c.answers) != len(submission.Answers) {
		return false
	}
	for i, answer := range c.answers {
		if !sameAnswer(answer, submission.Answers[i]) {
			return false
		}
	}
	return true
}

// sameAnswer compares two answers as the sets docs/vocabulary.md says they are,
// so a client that sends the same option identifiers in a different order on
// its retry is retrying rather than answering differently.
func sameAnswer(was Answer, now Answer) bool {
	wasSet, nowSet := set(was), set(now)
	if len(wasSet) != len(nowSet) {
		return false
	}
	for option := range wasSet {
		if !nowSet[option] {
			return false
		}
	}
	return true
}

// copyAnswers takes the submission's answers away from the caller, so a client
// struct reused for the next submission cannot rewrite a recorded one.
func copyAnswers(answers []Answer) []Answer {
	out := make([]Answer, len(answers))
	for i, answer := range answers {
		out[i] = append(Answer(nil), answer...)
	}
	return out
}

// sortByMomentThenKey puts classifications in the order they were recorded.
// The key breaks a tie, because two submissions can carry one moment from an
// injected clock that does not move, and a reading that depended on map order
// would make a campaign unreplayable.
func sortByMomentThenKey(classifications []Classification) {
	for i := 1; i < len(classifications); i++ {
		for j := i; j > 0 && earlier(classifications[j], classifications[j-1]); j-- {
			classifications[j], classifications[j-1] = classifications[j-1], classifications[j]
		}
	}
}

// earlier is the order sortByMomentThenKey sorts by.
func earlier(a Classification, b Classification) bool {
	if a.at.Equal(b.at) {
		return a.key < b.key
	}
	return a.at.Before(b.at)
}
