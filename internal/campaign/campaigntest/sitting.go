package campaigntest

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/iderex/lesesaal/internal/campaign"
)

// Sitting is one whole campaign run in memory, from the first answer to the
// last retirement.
//
// It exists because a rule tested one call at a time is tested against the
// inputs somebody thought of. A campaign run end to end reaches the states
// nobody wrote a case for: the subject that reaches the ceiling without a
// label, the volunteer who has answered everything left, the moment the last
// subject retires and the campaign has nothing for anybody.
//
// Nothing here touches an image, a browser, a model, a store or a network. The
// clock is the one a test moves, the draw is the fixture sequence, and a
// campaign of two thousand subjects finishes in the time it takes to count
// them.
type Sitting struct {
	// Definition is the campaign as its owner wrote it. Fixtures returns one
	// per task type the grammar declares.
	Definition campaign.DeclaredCampaign

	// Subjects is how many subjects the campaign holds, and Volunteers how
	// many people are answering.
	Subjects   int
	Volunteers int

	// Agreement is how often, in ten, a volunteer names the subject's own
	// answer. Ten is a campaign where everybody agrees about everything, which
	// no real one is; below about six the disagreement is what the run is
	// about and subjects reach the ceiling unresolved.
	Agreement int

	// Retirement is the rule that decides when a subject has had enough, and
	// Rule is the consensus rule that turns its answers into a label. The zero
	// value of each is refused rather than filled in, because a run whose
	// numbers came from a default nobody wrote down proves nothing about the
	// numbers a campaign owner would set.
	Retirement campaign.Retirement
	Rule       campaign.Rule

	// Hold is how long a subject stays out with the volunteer it was handed
	// to, and Between is how much the clock moves between two
	// classifications. Between is what makes the elapsed time in the summary a
	// number rather than zero.
	Hold    time.Duration
	Between time.Duration

	// Seed fixes the draw. Two sittings with one seed produce identical runs,
	// which is what makes a change to a rule comparable against the previous
	// behaviour on the same input.
	Seed uint64
}

// Classification is one volunteer's submission about one subject: one answer
// per task in the campaign's task order, and when it arrived.
type Classification struct {
	Volunteer string
	Subject   string
	Answers   []campaign.Answer
	At        time.Time
}

// Result is the whole of a run: the campaign it was, what arrived, what that
// became, and the numbers on top.
//
// The classifications and the labels are here rather than only the summary
// because this is what an export is written from. #68 reads a finished campaign
// out of a value of this shape instead of building a second driver to produce
// one.
type Result struct {
	Campaign        campaign.Campaign
	Subjects        []campaign.Subject
	Classifications []Classification

	// Labels is one label per task, in the campaign's task order, per subject
	// identifier. A subject nobody answered is absent rather than present with
	// an empty label, which is docs/vocabulary.md's "none until the first
	// answer".
	Labels map[string][]campaign.Label

	// Decisions is the last retirement decision taken about each subject that
	// retired, which is where the rule that fired and the numbers it fired on
	// are recorded.
	Decisions map[string]campaign.Decision

	Summary Summary
}

// Summary is what a campaign owner's progress view reads, which is #49's list:
// how many subjects are retired, how fast that is moving, whether the
// volunteers agree with each other, and which subjects are stuck.
type Summary struct {
	Subjects        int
	Retired         int
	Unresolved      int
	Shortened       int
	Classifications int

	// Labelled is the number of subject and task pairs carrying a label, which
	// is not the number of subjects: a campaign with two tasks produces two
	// labels per subject.
	Labelled int

	// Level is how many labels came back level at the top, which
	// docs/retirement.md reads differently from a plain spread.
	Level int

	// Confidence is the mean confidence over every label that exists, and is
	// zero where none does. It is the "whether the volunteers agree with each
	// other" of #49's list, and it is a mean over labels rather than anything
	// about a person.
	Confidence float64

	// Elapsed is how far the clock moved between the first classification and
	// the last.
	Elapsed time.Duration
}

// Stuck is the subjects that retired without every task reaching a label, which
// is the list docs/retirement.md sends a campaign owner to look at.
func (r Result) Stuck() []string {
	var stuck []string
	for id, decision := range r.Decisions {
		if decision.Unresolved() {
			stuck = append(stuck, id)
		}
	}
	sort.Strings(stuck)
	return stuck
}

// String writes the summary one number per line, in a fixed order.
//
// It is the form the determinism condition of #41 is read off: two runs under
// one seed produce the same string, and a difference is a line rather than a
// diff of two structures somebody has to read.
func (s Summary) String() string {
	return strings.Join([]string{
		fmt.Sprintf("subjects %d", s.Subjects),
		fmt.Sprintf("retired %d", s.Retired),
		fmt.Sprintf("unresolved %d", s.Unresolved),
		fmt.Sprintf("shortened %d", s.Shortened),
		fmt.Sprintf("classifications %d", s.Classifications),
		fmt.Sprintf("labelled %d", s.Labelled),
		fmt.Sprintf("level %d", s.Level),
		fmt.Sprintf("confidence %.4f", s.Confidence),
		fmt.Sprintf("elapsed %s", s.Elapsed),
	}, "\n")
}

// Run plays the whole campaign and reports what happened.
//
// The loop is volunteers in a fixed order asking for work until nobody can be
// given any. A round in which nothing was handed out is where it stops, which
// covers both endings: every subject retired, and subjects left that everybody
// present has already answered.
func (s Sitting) Run() (Result, error) {
	if err := s.check(); err != nil {
		return Result{}, err
	}

	defined, err := campaign.Define(s.Definition)
	if err != nil {
		return Result{}, fmt.Errorf("the fixture definition does not start: %w", err)
	}
	subjects, err := Plates(s.Subjects, Epoch)
	if err != nil {
		return Result{}, err
	}
	handouts, err := campaign.KeepHoldsFor(s.Hold)
	if err != nil {
		return Result{}, err
	}

	clock := NewClock(Epoch)
	draw := NewDraw(s.Seed)
	volunteers := VolunteerNames(s.Volunteers)
	tasks := defined.Tasks()

	answered := map[string]map[string]bool{}
	for _, volunteer := range volunteers {
		answered[volunteer] = map[string]bool{}
	}
	// collected is subject identifier to one answer set per task, in task
	// order, which is what the consensus rule reads.
	collected := map[string][]([]campaign.Answer){}
	counted := map[string]int{}
	retired := map[string]bool{}

	result := Result{
		Campaign:  defined,
		Subjects:  subjects,
		Labels:    map[string][]campaign.Label{},
		Decisions: map[string]campaign.Decision{},
	}

	started, ended := time.Time{}, time.Time{}
	for {
		handedOut := 0
		for _, volunteer := range volunteers {
			pool := campaign.Pool{
				Subjects: subjects,
				Retired:  retired,
				Answered: answered[volunteer],
			}
			handout, err := handouts.Next(volunteer, pool, clock.Now(), draw.Intn)
			if err != nil {
				// Nothing for this volunteer is an ordinary answer rather
				// than a failure: they have answered everything left, or
				// what is left is out with somebody else, or the campaign
				// is finished. The round ends and the loop stops when a
				// whole round hands out nothing.
				var nothing campaign.NoSubject
				if errors.As(err, &nothing) {
					continue
				}
				return Result{}, fmt.Errorf("the draw for %s failed: %w", volunteer, err)
			}

			subject := handout.Subject()
			id := subject.ID()
			if _, holding := collected[id]; !holding {
				collected[id] = make([][]campaign.Answer, len(tasks))
			}

			at := clock.Now()
			if started.IsZero() {
				started = at
			}
			ended = at

			answers := make([]campaign.Answer, 0, len(tasks))
			for i, task := range tasks {
				answer := s.answerTo(task, id, draw)
				collected[id][i] = append(collected[id][i], answer)
				answers = append(answers, answer)
			}
			result.Classifications = append(result.Classifications, Classification{
				Volunteer: volunteer,
				Subject:   id,
				Answers:   answers,
				At:        at,
			})
			answered[volunteer][id] = true
			counted[id]++
			handouts.Release(volunteer, id)

			labels := make([]campaign.Label, 0, len(tasks))
			for i, task := range tasks {
				labels = append(labels, campaign.Consensus(task, collected[id][i], s.Rule))
			}
			result.Labels[id] = labels

			decision := s.Retirement.Decide(labels, counted[id], nil)
			if decision.Retired() {
				retired[id] = true
				result.Decisions[id] = decision
			}

			clock.Advance(s.Between)
			handedOut++
		}
		if handedOut == 0 {
			break
		}
	}

	result.Summary = summarise(result, len(subjects), started, ended)
	return result, nil
}

// check refuses a sitting that would run and mean nothing.
func (s Sitting) check() error {
	var wrong []string
	if s.Subjects <= 0 {
		wrong = append(wrong, fmt.Sprintf("it holds %d subject(s)", s.Subjects))
	}
	if s.Volunteers <= 0 {
		wrong = append(wrong, fmt.Sprintf("it has %d volunteer(s)", s.Volunteers))
	}
	if s.Agreement < 0 || s.Agreement > 10 {
		wrong = append(wrong, fmt.Sprintf("its agreement is %d, which is not a number in ten", s.Agreement))
	}
	if s.Hold <= 0 {
		wrong = append(wrong, fmt.Sprintf("its hold is %s", s.Hold))
	}
	if s.Rule.Threshold() <= 0 {
		wrong = append(wrong, fmt.Sprintf("its consensus rule holds a threshold of %v, which is the zero value rather than a rule: campaign.DefaultAgreement is what a campaign gets where its owner declares none", s.Rule.Threshold()))
	}
	if len(wrong) == 0 {
		return nil
	}
	return fmt.Errorf("this sitting would run and mean nothing: %s", strings.Join(wrong, ", "))
}

// answerTo is what one volunteer says about one task for one subject.
//
// The subject has an answer of its own, derived from its identifier, and a
// volunteer names it as often as Agreement says and something else otherwise.
// This is a fixture and not a model of people: what it has to produce is
// agreement that reaches a threshold and disagreement that does not, so that a
// run reaches both the labelled and the unresolved ending.
func (s Sitting) answerTo(task campaign.Task, subject string, draw *Draw) campaign.Answer {
	options := task.Options()
	if len(options) == 0 {
		return nil
	}
	own := index(subject) % len(options)
	agreeing := draw.Intn(10) < s.Agreement

	if task.Type() == campaign.SingleChoice {
		at := own
		if !agreeing {
			at = (own + 1 + draw.Intn(len(options)-1)) % len(options)
		}
		return campaign.Answer{options[at].ID()}
	}

	atLeast, atMost := task.Bounds()
	size := atLeast
	if atMost > atLeast {
		size = atLeast + index(subject)%(atMost-atLeast+1)
	}
	if size > len(options) {
		size = len(options)
	}
	if !agreeing {
		// One option away from the subject's own set, in whichever direction
		// the bounds allow. That is what makes a near miss rather than a
		// different answer entirely, and a near miss is what a threshold has
		// to be able to reject.
		switch {
		case size < atMost && size < len(options):
			size++
		case size > atLeast:
			size--
		default:
			own = (own + 1) % len(options)
		}
	}
	answer := make(campaign.Answer, 0, size)
	for i := range size {
		answer = append(answer, options[(own+i)%len(options)].ID())
	}
	return answer
}

// index turns a fixture subject identifier into a number, so that a subject's
// own answer is a property of the subject rather than of the order it was
// drawn in. Anything unparsable counts as zero, which no fixture produces.
func index(subject string) int {
	at := strings.LastIndex(subject, "-")
	if at < 0 {
		return 0
	}
	n := 0
	for _, digit := range subject[at+1:] {
		if digit < '0' || digit > '9' {
			return 0
		}
		n = n*10 + int(digit-'0')
	}
	return n
}

// summarise counts the run.
func summarise(result Result, subjects int, started time.Time, ended time.Time) Summary {
	summary := Summary{
		Subjects:        subjects,
		Classifications: len(result.Classifications),
	}
	if !started.IsZero() {
		summary.Elapsed = ended.Sub(started)
	}
	for _, decision := range result.Decisions {
		summary.Retired++
		if decision.Unresolved() {
			summary.Unresolved++
		}
		if decision.Shortened() {
			summary.Shortened++
		}
	}
	total := 0.0
	for _, labels := range result.Labels {
		for _, label := range labels {
			if label.Level() {
				summary.Level++
			}
			if !label.Labelled() {
				continue
			}
			summary.Labelled++
			total += label.Confidence()
		}
	}
	if summary.Labelled > 0 {
		summary.Confidence = total / float64(summary.Labelled)
	}
	return summary
}
