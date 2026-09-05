package campaign

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Pool is what the store knows about a campaign at the moment a volunteer asks
// for work. It is the subjects together with two sets of identifiers rather
// than the things they stand for, so this rule runs with no store behind it,
// which docs/layout.md requires of the core. Progress does the same for the
// state machine and for the same reason.
type Pool struct {
	// Subjects is every subject in the campaign, in the order ingest took
	// them. The order matters: the draw is an index into the eligible set, so
	// a pool that came back in a different order on every call would make a
	// fixed draw unassertable and a campaign unreplayable.
	Subjects []Subject

	// Retired names the subjects the retirement rule has finished with.
	// docs/subject-selection.md excludes them at the point of the draw rather
	// than filtering afterwards, which is what makes "never" the honest word
	// there instead of "rarely".
	Retired map[string]bool

	// Answered names the subjects this volunteer has already answered. It is
	// the one condition of the three that is a property of the pair rather
	// than of the subject, and docs/subject-selection.md names it as the first
	// thing to give at a hundred times the design point.
	Answered map[string]bool

	// Uncertainty is how unsure the model is about a subject, higher meaning
	// less certain, for the subjects it has scored. It is the number #55
	// supplies and this rule does not compute or interpret it: what it does
	// with it is cut a share off the top, so any scale works as long as one
	// campaign uses one of them.
	//
	// A subject absent from this map is one the model has not scored, which is
	// every subject before the first model and every subject added since the
	// last training run. It stays fully available on the uniform draw and is
	// never in the uncertain share, because a subject with no score is not a
	// subject the model is uncertain about. #60 is where that second cold
	// start is held.
	//
	// An empty or absent map is the model having no opinion at all, and
	// docs/subject-selection.md answers that case by name: every draw is the
	// uniform one.
	Uncertainty map[string]float64
}

// Reason is why a volunteer asking for work was given none.
//
// The three are separate values rather than one because
// docs/subject-selection.md requires the surface to say three different things:
// a volunteer told the campaign is finished when it is not will believe the
// campaign is finished.
type Reason string

const (
	// CampaignFinished is every subject retired. There is no more work for
	// anybody and the page stops rather than waiting for some.
	CampaignFinished Reason = "every subject in this campaign is retired"

	// AnsweredEverything is work remaining that this volunteer has already
	// answered. Re-serving one would buy a classification and cost the
	// independence that makes the consensus rule mean anything.
	AnsweredEverything Reason = "this volunteer has answered every subject that is left"

	// OutWithOthers is the transient case, and the only one that resolves by
	// itself. Every subject this volunteer could still answer is held by
	// somebody else, and the shortest of those holds is how long that lasts.
	OutWithOthers Reason = "every subject left for this volunteer is out with somebody else"
)

// NoSubject is what a draw returns when the eligible set is empty. RetryAfter
// is how long until the position could change on its own, and it is set only
// for OutWithOthers: the other two do not resolve by waiting, and a surface
// that offered a retry for them would be asking a volunteer to sit and refresh.
type NoSubject struct {
	Because    Reason
	RetryAfter time.Duration
}

// Error writes the reason, with the wait where there is one.
func (n NoSubject) Error() string {
	if n.RetryAfter > 0 {
		return fmt.Sprintf("%s, and the first hold expires in %s", n.Because, n.RetryAfter)
	}
	return string(n.Because)
}

// Handout is one subject given to one volunteer, held for them until they
// answer or the hold expires.
type Handout struct {
	subject   Subject
	volunteer string
	until     time.Time
}

// Subject is the subject the volunteer was given.
func (h Handout) Subject() Subject { return h.subject }

// Volunteer is who it was given to.
func (h Handout) Volunteer() string { return h.volunteer }

// Until is when the hold expires and the subject returns to the eligible set
// for everybody. It is a moment rather than a duration because the caller
// stamping it and the caller reading it are not always the same one.
func (h Handout) Until() time.Time { return h.until }

// DefaultOneDrawIn is the proportion docs/subject-selection.md fixes: one draw
// in four is taken from the share of the eligible set the model is least
// certain about, and three in four are taken from the whole of it.
//
// That document says the mixture is ONE number and then uses the number twice,
// once for how often the model is consulted and once for how big the share it
// is consulted about is - one draw in four, and the quarter. This reads it as
// one setting rather than two, so a campaign that moves the proportion moves
// both together. Where they should part, that is a change to that document and
// not a second number added here.
const DefaultOneDrawIn = 4

// Mixture is how often the draw consults the model, and how much of the
// eligible set it consults it about.
//
// It is a setting with the document's proportion as its default rather than a
// constant, which is what #56 asks for: the number is a starting position and
// the pilot in #115 and #116 is what moves it, so an operator changing it
// should not be changing code.
type Mixture struct {
	// oneDrawIn is n in "one draw in n". It is unexported so that the only way
	// to hold a Mixture that is not the zero value is to have been through a
	// constructor that refused the numbers that mean nothing.
	oneDrawIn int
}

// DefaultMixture is the proportion docs/subject-selection.md fixes.
func DefaultMixture() Mixture { return Mixture{oneDrawIn: DefaultOneDrawIn} }

// OneDrawIn is the mixture that consults the model on one draw in n, over the
// 1/n of the eligible set it is least certain about.
//
// A n of 1 is refused rather than read as ordering by uncertainty. Every draw
// would come from the whole eligible set sorted by uncertainty, which is the
// ordering docs/subject-selection.md rejects by name: it hands the single
// hardest subject to everybody who asks, and an operator who typed 1 meaning
// "always consult the model" would get that without being told.
func OneDrawIn(n int) (Mixture, error) {
	if n < 2 {
		return Mixture{}, Refused{Refusals: []Refusal{{
			Says: fmt.Sprintf("a mixture of one draw in %d is not a mixture: every draw would come from the share the model is least certain about, which is the ordering docs/subject-selection.md rejects because it hands the hardest subject to everybody at once", n),
		}}}
	}
	return Mixture{oneDrawIn: n}, nil
}

// OneDrawIn is n in "one draw in n".
func (m Mixture) OneDrawIn() int { return m.oneDrawIn }

// String writes the mixture the way the configuration prints it, so an operator
// reads the proportion rather than a struct.
func (m Mixture) String() string {
	return fmt.Sprintf("1 draw in %d over the least certain 1/%d", m.oneDrawIn, m.oneDrawIn)
}

// Handouts is the holds of one campaign: which subjects are out with which
// volunteers and until when.
//
// The mechanism is the reservation with a timeout that docs/subject-selection.md
// chose, and the three failure modes it costs are these.
//
// A hold leaks. A volunteer who closes the tab is indistinguishable from one
// who is still looking, so the subject stays out until the timeout returns it.
// That timeout is a number somebody has to choose and neither direction is
// free: too long and a small campaign starves while its subjects sit in
// abandoned tabs, too short and a volunteer who reads the instructions
// carefully has their subject handed to somebody else while they are still
// looking at it.
//
// A restart forgets every hold, which is the same as every hold expiring at
// once. Holds are in memory rather than in the store, so a subject out with a
// volunteer at the moment a deployment restarts is immediately eligible again
// and can be handed to a second person. That costs at most one duplicate
// classification per subject that was out, which is a thing the consensus rule
// already takes, and it buys a write per draw that the store does not have to
// take. docs/deployment-ceiling.md is where the serialised write path is the
// resource being spent.
//
// A volunteer holds one subject at a time. Asking again while holding one hands
// back the one they hold rather than drawing a second, which holding says why,
// and the hold is not extended by the asking.
//
// Two draws in the same instant is the one this issue exists against, and it is
// closed rather than made rare. Every draw takes the same lock, so the read of
// the eligible set and the write that records the hold are one step and no
// second draw can see the subject as eligible in between.
// docs/subject-selection.md says "mostly rather than never is the honest word"
// about two volunteers being given one subject, names the write that records
// the hold as what closes the window, and leaves the proof to this mechanism.
// Inside one process the window is closed; across two it is not, and there is
// no second process to close it against.
type Handouts struct {
	// mu makes the read of the eligible set and the write of the hold one
	// step. It is a mutex rather than a channel or an atomic because what is
	// being protected is a decision over the whole set rather than one value.
	mu sync.Mutex

	// hold is how long a subject stays out with the volunteer it was given to.
	hold time.Duration

	// mixture is how often the draw consults the model. It is fixed for the
	// life of a campaign's handouts rather than passed per draw, because a
	// proportion that moved between two volunteers asking in the same minute
	// would make a campaign unreplayable.
	mixture Mixture

	// held is subject identifier to hold. Only subjects that are out are in
	// it, so its size is the number of volunteers working rather than the
	// number of subjects in the campaign.
	held map[string]held
}

// held is one subject out with one volunteer until one moment.
type held struct {
	volunteer string
	until     time.Time
}

// KeepHoldsFor builds the holds of one campaign, with the time a subject stays
// out after it is handed over.
//
// A hold of zero or less is refused rather than treated as no holding at all. A
// campaign that holds nothing hands one subject to everybody who asks at the
// same moment, which is the collision this whole mechanism exists against, and
// an operator who wants that has asked for something this project will not do
// silently.
func KeepHoldsFor(hold time.Duration) (*Handouts, error) {
	return KeepHoldsForDrawing(hold, DefaultMixture())
}

// KeepHoldsForDrawing is the same with the mixture named, for a campaign whose
// owner moved the proportion off the document's default.
func KeepHoldsForDrawing(hold time.Duration, mixture Mixture) (*Handouts, error) {
	if hold <= 0 {
		return nil, Refused{Refusals: []Refusal{{
			Says: fmt.Sprintf("a hold of %s keeps nothing: the subject would be eligible again before the volunteer it was given to had seen it, and every volunteer asking at once would be handed the same one", hold),
		}}}
	}
	if mixture.oneDrawIn < 2 {
		return nil, Refused{Refusals: []Refusal{{
			Says: fmt.Sprintf("a mixture of one draw in %d was not built by OneDrawIn, and the zero value would consult the model on every draw or on none depending on which branch read it first", mixture.oneDrawIn),
		}}}
	}
	return &Handouts{hold: hold, mixture: mixture, held: map[string]held{}}, nil
}

// Mixture is the proportion this campaign draws at, which is what the
// configuration prints for an operator.
func (h *Handouts) Mixture() Mixture {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.mixture
}

// Next gives one volunteer one subject and holds it for them, or says why there
// is nothing for them.
//
// The eligible set is what docs/subject-selection.md defines: every subject
// that is not retired, that this volunteer has not already answered, and that
// is not currently held by another volunteer. The draw over it is uniform. The
// mixture that takes one draw in four from the quarter the model is least
// certain about is #56's, and it is not here: this takes the whole eligible set
// every time, which is exactly what that document says every draw does where
// the model has no opinion, and a campaign with no model is the default.
//
// now and draw are passed in rather than read, from Depends.Now and
// Depends.Intn, because this package reads neither a clock nor a random source.
func (h *Handouts) Next(volunteer string, pool Pool, now time.Time, draw func(n int) int) (Handout, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.expire(now)

	if already, holding := h.holding(volunteer, pool); holding {
		return already, nil
	}

	eligible := make([]Subject, 0, len(pool.Subjects))
	answeredSomething, blocked := false, 0
	var soonest time.Time
	for _, subject := range pool.Subjects {
		if pool.Retired[subject.ID()] {
			continue
		}
		if pool.Answered[subject.ID()] {
			answeredSomething = true
			continue
		}
		if out, taken := h.held[subject.ID()]; taken && out.volunteer != volunteer {
			blocked++
			if soonest.IsZero() || out.until.Before(soonest) {
				soonest = out.until
			}
			continue
		}
		eligible = append(eligible, subject)
	}

	if len(eligible) == 0 {
		return Handout{}, nothingFor(answeredSomething, blocked, soonest, now)
	}

	from, err := h.lane(eligible, pool, draw)
	if err != nil {
		return Handout{}, err
	}

	at := draw(len(from))
	if at < 0 || at >= len(from) {
		return Handout{}, Refused{Refusals: []Refusal{{
			Says: fmt.Sprintf("the draw returned %d for a set of %d, which is outside the interval Depends.Intn declares, so no subject was handed out rather than one being chosen by whatever the number happened to reach", at, len(from)),
		}}}
	}

	chosen := from[at]
	until := now.Add(h.hold)
	h.held[chosen.ID()] = held{volunteer: volunteer, until: until}
	return Handout{subject: chosen, volunteer: volunteer, until: until}, nil
}

// lane is the set this draw is taken from: the whole eligible set, or the share
// of it the model is least certain about.
//
// This is the mixture docs/subject-selection.md decided, and the two parts that
// defuse uncertainty ordering are both here. Which lane is used is decided by a
// draw rather than by a counter, so ten volunteers asking in the same minute do
// not all take the uncertain one; and the draw INSIDE the uncertain lane is
// uniform rather than ordered, so ten volunteers who do take it land on ten
// different subjects instead of ten looks at the single hardest image.
//
// The caller holds the lock.
func (h *Handouts) lane(eligible []Subject, pool Pool, draw func(n int) int) ([]Subject, error) {
	if len(pool.Uncertainty) == 0 {
		return eligible, nil
	}

	which := draw(h.mixture.oneDrawIn)
	if which < 0 || which >= h.mixture.oneDrawIn {
		return nil, Refused{Refusals: []Refusal{{
			Says: fmt.Sprintf("the draw returned %d for a mixture of one in %d, which is outside the interval Depends.Intn declares, so no subject was handed out rather than the lane being chosen by whatever the number happened to reach", which, h.mixture.oneDrawIn),
		}}}
	}
	if which != 0 {
		return eligible, nil
	}

	uncertain := leastCertain(eligible, pool.Uncertainty, h.mixture.oneDrawIn)
	if len(uncertain) == 0 {
		// Every eligible subject is one the model has not scored, which is the
		// second cold start in #60 and is answered the same way as the first:
		// the share is undefined and this draw is the uniform one.
		return eligible, nil
	}
	return uncertain, nil
}

// leastCertain is the 1/share of the SCORED eligible subjects the model is
// least sure about, which is at least one where any subject is scored at all.
//
// It cuts the share from the scored subjects rather than from the whole
// eligible set, and that is a reading of docs/subject-selection.md rather than
// a sentence in it. The alternative, treating an unscored subject as maximally
// uncertain, would fill the uncertain lane with whatever was ingested most
// recently and never with what the model actually struggled on.
//
// The order is by uncertainty and then by identifier, so a campaign replayed
// from the same pool cuts the same share. Sorting by a float alone leaves
// subjects at equal uncertainty in whatever order the eligible set arrived in,
// which is stable here and would stop being stable the first time a caller
// built the pool from a map.
func leastCertain(eligible []Subject, uncertainty map[string]float64, share int) []Subject {
	scored := make([]Subject, 0, len(eligible))
	for _, subject := range eligible {
		if _, has := uncertainty[subject.ID()]; has {
			scored = append(scored, subject)
		}
	}
	if len(scored) == 0 {
		return nil
	}

	sort.SliceStable(scored, func(i, j int) bool {
		left, right := uncertainty[scored[i].ID()], uncertainty[scored[j].ID()]
		if left == right {
			return scored[i].ID() < scored[j].ID()
		}
		return left > right
	})

	cut := len(scored) / share
	if cut < 1 {
		cut = 1
	}
	return scored[:cut]
}

// holding is the subject this volunteer is already holding, where they are
// holding one. The caller holds the lock.
//
// docs/subject-selection.md excludes from the eligible set the subjects held by
// ANOTHER volunteer, so a subject this volunteer holds is still theirs to be
// given. What that document does not settle is what a second ask by the same
// volunteer means, and the two readings are far apart: drawing again leaves the
// first hold standing, so a volunteer who reloads the page four times takes four
// subjects out of the campaign and answers one, which is the reservation's
// leak arriving through the front door rather than through a closed tab.
//
// Handing back the one they already hold is the reading taken here. A reload
// then shows the same subject, which is what a volunteer expects, and a
// volunteer holds at most one subject at a time whatever they do to the page.
// The hold is NOT extended by asking again: a volunteer who reloads every
// minute would otherwise hold one subject for as long as they cared to, and the
// timeout that bounds an abandoned hold would bound nothing.
//
// A hold whose subject has since been retired, which the store now reports this
// volunteer as having answered, or which is no longer in the campaign at all is
// dropped rather than handed back. Each of the three means the hold has been
// overtaken by something the pool knows and this structure does not, and a hold
// left standing over one of them would keep a subject out of a campaign that no
// longer contains it.
func (h *Handouts) holding(volunteer string, pool Pool) (Handout, bool) {
	for id, out := range h.held {
		if out.volunteer != volunteer {
			continue
		}
		if !pool.Retired[id] && !pool.Answered[id] {
			for _, subject := range pool.Subjects {
				if subject.ID() == id {
					return Handout{subject: subject, volunteer: volunteer, until: out.until}, true
				}
			}
		}
		delete(h.held, id)
	}
	return Handout{}, false
}

// nothingFor decides which of the three empty cases this is.
//
// The order matters and is not the order they are declared in. Nothing held and
// nothing answered means every subject is retired, which is the only one of the
// three that is a fact about the campaign rather than about this volunteer.
func nothingFor(answeredSomething bool, blocked int, soonest time.Time, now time.Time) error {
	if blocked > 0 {
		wait := soonest.Sub(now)
		if wait < 0 {
			wait = 0
		}
		return NoSubject{Because: OutWithOthers, RetryAfter: wait}
	}
	if answeredSomething {
		return NoSubject{Because: AnsweredEverything}
	}
	return NoSubject{Because: CampaignFinished}
}

// Release gives a subject back before its hold expires, which is what a
// volunteer answering does.
//
// It reports whether this volunteer was the one holding it. A release naming
// somebody else's hold changes nothing: a request that arrives late, or twice,
// or for a subject whose hold has already expired and been handed to a second
// volunteer, would otherwise take that second volunteer's subject away while
// they were looking at it.
func (h *Handouts) Release(volunteer string, subject string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	out, taken := h.held[subject]
	if !taken || out.volunteer != volunteer {
		return false
	}
	delete(h.held, subject)
	return true
}

// Out is how many subjects are with volunteers at this moment, expired holds
// excluded. It is what a campaign owner's progress view reads, and it is also
// how a test asks whether a hold was actually given up rather than asking the
// draw again and inferring it.
func (h *Handouts) Out(now time.Time) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.expire(now)
	return len(h.held)
}

// expire drops the holds that have run out. The caller holds the lock.
//
// It runs on every draw rather than on a timer, so nothing here needs a
// goroutine, a ticker or a second process: the sweep is paid for by the request
// that would have been blocked by the stale hold. The cost is that a campaign
// nobody is asking for work in keeps its expired holds until somebody does,
// which changes no answer because a hold nobody is competing for holds nothing
// up.
func (h *Handouts) expire(now time.Time) {
	for subject, out := range h.held {
		if !out.until.After(now) {
			delete(h.held, subject)
		}
	}
}
