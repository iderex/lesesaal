package campaign

import (
	"fmt"
	"testing"
	"time"
)

// scores gives every subject an uncertainty, highest for plate-1 and falling
// from there, so the least certain share of a collection built by collection is
// always its first subjects and the most certain one is always its last.
func scores(subjects []Subject) map[string]float64 {
	out := map[string]float64{}
	for i, subject := range subjects {
		out[subject.ID()] = 1 - float64(i)/float64(len(subjects))
	}
	return out
}

// counting is a draw that records what it was asked for and answers from a
// fixed sequence, repeating the last entry once the sequence runs out. It is
// how a case here says which lane a draw took: the first call of a draw that
// consults the model is the lane, and the second is the index inside it.
type counting struct {
	asked   []int
	answers []int
	at      int
}

func (c *counting) Intn(n int) int {
	c.asked = append(c.asked, n)
	if len(c.answers) == 0 {
		return 0
	}
	answer := c.answers[min(c.at, len(c.answers)-1)]
	c.at++
	if answer >= n {
		return n - 1
	}
	return answer
}

// TestWhereTheModelHasNoOpinionEveryDrawIsTheUniformOne is the sentence
// docs/subject-selection.md ends its rule with, and it is what makes a campaign
// before its first model need no special handling, which is #60.
func TestWhereTheModelHasNoOpinionEveryDrawIsTheUniformOne(t *testing.T) {
	subjects := collection(t, 40)

	for _, c := range []struct {
		name string
		pool Pool
	}{
		{"no model at all", Pool{Subjects: subjects}},
		{"a model that scored nothing", Pool{Subjects: subjects, Uncertainty: map[string]float64{}}},
	} {
		handouts, err := KeepHoldsFor(time.Hour)
		if err != nil {
			t.Fatalf("%s: the holds were refused: %v", c.name, err)
		}
		draw := &counting{}
		handout, err := handouts.Next("v-"+c.name, c.pool, entered, draw.Intn)
		if err != nil {
			t.Fatalf("%s: the draw failed: %v", c.name, err)
		}

		// One call, for the whole eligible set. A second call would be the
		// lane being chosen, and there is no lane to choose.
		if len(draw.asked) != 1 || draw.asked[0] != len(subjects) {
			t.Errorf("%s: the draw was asked for %v, and the eligible set is %d", c.name, draw.asked, len(subjects))
		}
		if handout.Subject().ID() != "plate-1" {
			t.Errorf("%s: a draw of 0 over the whole set gave %s", c.name, handout.Subject().ID())
		}
	}
}

// TestOneDrawInFourComesFromTheShareTheModelIsLeastCertainAbout is the mixture
// itself. The lane is chosen by a draw, so this asserts on both answers to that
// draw rather than only on the one that consults the model.
func TestOneDrawInFourComesFromTheShareTheModelIsLeastCertainAbout(t *testing.T) {
	subjects := collection(t, 40)
	pool := Pool{Subjects: subjects, Uncertainty: scores(subjects)}

	for _, c := range []struct {
		name string
		lane int
		want string
		set  int
	}{
		// The mixture's own draw answers 0 one time in four, and that is the
		// draw that consults the model. The share is a quarter of 40, so the
		// lane holds 10 and its head is the least certain subject.
		{"the model is consulted", 0, "plate-1", 10},
		// Any other answer is one of the three draws in four that ignore the
		// model, over the whole eligible set of 40.
		{"the model is not consulted", 1, "plate-1", 40},
	} {
		handouts, err := KeepHoldsFor(time.Hour)
		if err != nil {
			t.Fatalf("%s: the holds were refused: %v", c.name, err)
		}
		draw := &counting{answers: []int{c.lane, 0}}

		handout, err := handouts.Next("v1", pool, entered, draw.Intn)
		if err != nil {
			t.Fatalf("%s: the draw failed: %v", c.name, err)
		}
		if len(draw.asked) != 2 {
			t.Fatalf("%s: the draw was asked %d time(s), and a pool carrying scores asks twice", c.name, len(draw.asked))
		}
		if draw.asked[0] != DefaultOneDrawIn {
			t.Errorf("%s: the lane was drawn over %d rather than the mixture's %d", c.name, draw.asked[0], DefaultOneDrawIn)
		}
		if draw.asked[1] != c.set {
			t.Errorf("%s: the subject was drawn over a set of %d, want %d", c.name, draw.asked[1], c.set)
		}
		if handout.Subject().ID() != c.want {
			t.Errorf("%s: the draw gave %s, want %s", c.name, handout.Subject().ID(), c.want)
		}
	}
}

// TestTheDrawInsideTheUncertainShareIsUniformRatherThanOrdered is the half of
// the rule that keeps ten volunteers off one image. Ordering by uncertainty and
// taking the top would give all of them the same subject; drawing uniformly
// inside the share gives them different ones.
func TestTheDrawInsideTheUncertainShareIsUniformRatherThanOrdered(t *testing.T) {
	subjects := collection(t, 40)
	pool := Pool{Subjects: subjects, Uncertainty: scores(subjects)}

	given := map[string]bool{}
	for i := range 10 {
		handouts, err := KeepHoldsFor(time.Hour)
		if err != nil {
			t.Fatalf("the holds were refused: %v", err)
		}
		// Every volunteer takes the model's lane, and each lands on a
		// different index inside it.
		draw := &counting{answers: []int{0, i}}
		handout, err := handouts.Next(fmt.Sprintf("v%d", i), pool, entered, draw.Intn)
		if err != nil {
			t.Fatalf("the draw for v%d failed: %v", i, err)
		}
		given[handout.Subject().ID()] = true
	}

	if len(given) != 10 {
		t.Errorf("ten volunteers drawing the model's lane were given %d different subject(s): %v", len(given), given)
	}
}

// TestTheShareIsCutFromTheSubjectsTheModelScored holds the reading the document
// does not state. A subject the model has never scored is not a subject the
// model is uncertain about, so it is never in the share, and treating it as
// maximally uncertain would fill the lane with whatever was ingested last.
func TestTheShareIsCutFromTheSubjectsTheModelScored(t *testing.T) {
	subjects := collection(t, 8)

	// Only the last four are scored, and plate-8 is the least certain of them.
	uncertainty := map[string]float64{
		"plate-5": 0.1, "plate-6": 0.2, "plate-7": 0.3, "plate-8": 0.4,
	}
	pool := Pool{Subjects: subjects, Uncertainty: uncertainty}

	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	draw := &counting{answers: []int{0, 0}}
	handout, err := handouts.Next("v1", pool, entered, draw.Intn)
	if err != nil {
		t.Fatalf("the draw failed: %v", err)
	}

	// A quarter of the four scored subjects is one, and it is the least
	// certain of the scored ones rather than an unscored newcomer.
	if draw.asked[1] != 1 {
		t.Errorf("the share held %d subject(s) where four were scored", draw.asked[1])
	}
	if handout.Subject().ID() != "plate-8" {
		t.Errorf("the model's lane gave %s, and plate-8 is the least certain scored subject", handout.Subject().ID())
	}

	// The unscored subjects are not excluded from the campaign: the uniform
	// draw still reaches them, which is what keeps a newly ingested subject
	// from waiting for a training run before anybody sees it.
	uniform := &counting{answers: []int{1, 0}}
	next, err := handouts.Next("v2", pool, entered, uniform.Intn)
	if err != nil {
		t.Fatalf("the uniform draw failed: %v", err)
	}
	if next.Subject().ID() != "plate-1" {
		t.Errorf("the uniform draw gave %s, and plate-1 is an unscored subject at the head of the eligible set", next.Subject().ID())
	}
}

// TestWhereNothingEligibleIsScoredTheDrawFallsBackToUniform is the second cold
// start in #60: a campaign whose model has scored subjects, all of which have
// since retired or been answered, is in the same position as day one.
func TestWhereNothingEligibleIsScoredTheDrawFallsBackToUniform(t *testing.T) {
	subjects := collection(t, 6)
	pool := Pool{
		Subjects:    subjects,
		Retired:     map[string]bool{"plate-1": true, "plate-2": true, "plate-3": true},
		Uncertainty: map[string]float64{"plate-1": 0.9, "plate-2": 0.8, "plate-3": 0.7},
	}

	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	draw := &counting{answers: []int{0, 0}}
	handout, err := handouts.Next("v1", pool, entered, draw.Intn)
	if err != nil {
		t.Fatalf("the draw failed: %v", err)
	}

	if draw.asked[1] != 3 {
		t.Errorf("the draw was taken over a set of %d, and three unscored subjects are eligible", draw.asked[1])
	}
	if handout.Subject().ID() != "plate-4" {
		t.Errorf("the draw gave %s, and plate-4 is the head of what is left", handout.Subject().ID())
	}
}

// sequence is a fixed pseudo-random draw, written here because nothing outside
// internal/system may read a random source and because a fixture whose numbers
// depend on a library version changes under a test that did not. It is the same
// mix step the fixture draw uses.
type sequence struct {
	state uint64
}

func newSequence(seed uint64) *sequence {
	return &sequence{state: seed + 0x9e3779b97f4a7c15}
}

func (s *sequence) Intn(n int) int {
	s.state += 0x9e3779b97f4a7c15
	z := s.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	z ^= z >> 31
	return int(z % uint64(n))
}

// TestNoSubjectIsStarvedAtTheDesignPoint is the condition the issue states in
// those words, at docs/design-point.md's campaign of 2,000 subjects.
//
// It is proved twice, because the two proofs answer different halves. The first
// is the property itself and is not a matter of luck: three draws in four are
// taken over the whole eligible set, so every eligible subject is in the set
// those draws are taken from, including the one the model is most certain about
// and which is therefore never in the uncertain share. The second shows that
// this is not only true of the set but observable of the campaign, by running
// the draw until that subject is handed out.
func TestNoSubjectIsStarvedAtTheDesignPoint(t *testing.T) {
	const subjectCount = 2000

	subjects := collection(t, subjectCount)
	uncertainty := scores(subjects)
	mostCertain := subjects[len(subjects)-1].ID()

	// The subject is outside the uncertain share, so this case is about the
	// uniform draw reaching it rather than about the share being generous.
	for _, subject := range leastCertain(subjects, uncertainty, DefaultOneDrawIn) {
		if subject.ID() == mostCertain {
			t.Fatalf("%s is in the uncertain share, so this case proves nothing", mostCertain)
		}
	}

	handouts, err := KeepHoldsFor(time.Nanosecond)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: subjects, Uncertainty: uncertainty}

	// The property. The lane the three-in-four draws take is the whole
	// eligible set, whatever the model believes.
	uniform, err := handouts.lane(subjects, pool, func(int) int { return 1 })
	if err != nil {
		t.Fatalf("the uniform lane failed: %v", err)
	}
	if len(uniform) != subjectCount {
		t.Fatalf("the uniform lane holds %d of %d subject(s)", len(uniform), subjectCount)
	}
	found := false
	for _, subject := range uniform {
		if subject.ID() == mostCertain {
			found = true
		}
	}
	if !found {
		t.Errorf("%s is not in the set the uniform draw is taken over", mostCertain)
	}

	// And the campaign reaching it. A uniform draw is three rounds in four
	// over 2,000 subjects, so the expected wait is under three thousand
	// rounds; the bound below is generous and the seed is fixed, so this case
	// either passes or does not rather than passing most of the time.
	draw := newSequence(20260821)
	at := entered
	for round := range 30000 {
		at = at.Add(time.Second)
		handout, err := handouts.Next("v1", pool, at, draw.Intn)
		if err != nil {
			t.Fatalf("round %d failed: %v", round, err)
		}
		handouts.Release("v1", handout.Subject().ID())
		if handout.Subject().ID() == mostCertain {
			return
		}
	}

	t.Errorf("the subject the model is most certain about was not handed out in 30000 rounds over %d subjects", subjectCount)
}

// TestTheMixtureRefusesTheNumbersThatAreNotOne holds what an operator can type.
// One draw in one is uncertainty ordering, which the document rejects by name,
// and it would arrive here looking like a setting rather than like a decision.
func TestTheMixtureRefusesTheNumbersThatAreNotOne(t *testing.T) {
	for _, n := range []int{-1, 0, 1} {
		if _, err := OneDrawIn(n); err == nil {
			t.Errorf("a mixture of one draw in %d was admitted", n)
		}
	}
	for _, n := range []int{2, 4, 100} {
		mixture, err := OneDrawIn(n)
		if err != nil {
			t.Errorf("a mixture of one draw in %d was refused: %v", n, err)
			continue
		}
		if mixture.OneDrawIn() != n {
			t.Errorf("a mixture of one draw in %d reads back as %d", n, mixture.OneDrawIn())
		}
	}

	// The zero value is refused where it would otherwise reach a draw, because
	// nothing else stops a caller building a Handouts around it.
	if _, err := KeepHoldsForDrawing(time.Hour, Mixture{}); err == nil {
		t.Error("holds were built around a mixture of one draw in zero")
	}

	// The default is the document's proportion, and the constructor that names
	// no mixture takes it.
	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	if handouts.Mixture() != DefaultMixture() || handouts.Mixture().OneDrawIn() != DefaultOneDrawIn {
		t.Errorf("the default mixture is %s", handouts.Mixture())
	}
}

// TestADeclaredMixtureIsTheOneTheDrawUses is what makes the proportion a
// setting rather than a constant, which is the issue's first condition.
func TestADeclaredMixtureIsTheOneTheDrawUses(t *testing.T) {
	mixture, err := OneDrawIn(2)
	if err != nil {
		t.Fatalf("one draw in two was refused: %v", err)
	}
	handouts, err := KeepHoldsForDrawing(time.Hour, mixture)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}

	subjects := collection(t, 40)
	pool := Pool{Subjects: subjects, Uncertainty: scores(subjects)}
	draw := &counting{answers: []int{0, 0}}

	if _, err := handouts.Next("v1", pool, entered, draw.Intn); err != nil {
		t.Fatalf("the draw failed: %v", err)
	}
	if draw.asked[0] != 2 {
		t.Errorf("the lane was drawn over %d, and the declared mixture is one draw in 2", draw.asked[0])
	}
	// Half of 40 rather than a quarter: the one number decides how often the
	// model is consulted and how much of the set it is consulted about.
	if draw.asked[1] != 20 {
		t.Errorf("the share held %d subject(s), and half of 40 is 20", draw.asked[1])
	}
}

// TestTheShareIsCutTheSameWayTwice holds replayability. Two subjects at one
// uncertainty are ordered by identifier, so a pool a caller built from a map
// cuts the same share on every run rather than one that depends on the order
// the eligible set happened to arrive in.
func TestTheShareIsCutTheSameWayTwice(t *testing.T) {
	subjects := collection(t, 8)
	level := map[string]float64{}
	for _, subject := range subjects {
		level[subject.ID()] = 0.5
	}

	want := ""
	for run := range 8 {
		share := leastCertain(subjects, level, 4)
		got := ""
		for _, subject := range share {
			got += subject.ID() + " "
		}
		if run == 0 {
			want = got
			continue
		}
		if got != want {
			t.Fatalf("run %d cut %q where run 0 cut %q", run, got, want)
		}
	}
	if want != "plate-1 plate-2 " {
		t.Errorf("eight subjects at one uncertainty cut to %q", want)
	}
}
