package campaign

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// collection builds a campaign of n subjects, named plate-1 upwards in the
// order ingest would have taken them. They carry no metadata, because nothing
// the handout does reads any.
func collection(t *testing.T, n int) []Subject {
	t.Helper()
	subjects := make([]Subject, 0, n)
	for i := 1; i <= n; i++ {
		subject, err := Enter(
			DeclaredSubject{ID: fmt.Sprintf("plate-%d", i)},
			Derived{Digest: fmt.Sprintf("sha256:%064d", i), Bytes: 1, Width: 1, Height: 1, Entered: entered},
		)
		if err != nil {
			t.Fatalf("the fixture subject %d is refused: %v", i, err)
		}
		subjects = append(subjects, subject)
	}
	return subjects
}

// first is the draw that always takes the head of the eligible set. It is the
// sharpest draw for a collision test: without the write and the read being one
// step, every volunteer asking at the same moment takes the same head.
func first(int) int { return 0 }

// jostled is first with the processor given up in it, and it is the collision
// test's draw. The draw is called between the read of the eligible set and the
// write of the hold, so yielding there widens the window this mechanism exists
// to close and gives the other volunteers room to reach it. Where the read and
// the write are one step the yielding changes no answer, because the others are
// then waiting on the lock rather than on the scheduler.
//
// It yields rather than sleeps. A test in this tree may not call time.Sleep at
// all, which harness_test.go refuses by name, and waiting on the wall clock is
// what that refusal is for; this asks the scheduler to run somebody else and
// waits for nothing.
func jostled(int) int {
	for range 200 {
		runtime.Gosched()
	}
	return 0
}

// TestTenVolunteersAskingAtOnceAreGivenTenDifferentSubjects is #35's first
// sentence: ten people ask at the same moment and the mechanism must not give
// everybody the same subject because it read a set nobody had updated yet.
//
// The draw always returns the head of the eligible set, so the only thing that
// can make the ten answers differ is the hold being written before the next
// read of the set. A mechanism that read the eligible set outside the lock
// hands plate-1 to all ten.
func TestTenVolunteersAskingAtOnceAreGivenTenDifferentSubjects(t *testing.T) {
	handouts, err := KeepHoldsFor(15 * time.Minute)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: collection(t, 40)}
	now := entered

	const volunteers = 10
	given := make([]string, volunteers)
	var start sync.WaitGroup
	var done sync.WaitGroup
	start.Add(1)
	for i := range volunteers {
		done.Add(1)
		go func() {
			defer done.Done()
			start.Wait()
			handout, err := handouts.Next(fmt.Sprintf("v%d", i), pool, now, jostled)
			if err != nil {
				t.Errorf("volunteer v%d was given nothing: %v", i, err)
				return
			}
			given[i] = handout.Subject().ID()
		}()
	}
	start.Done()
	done.Wait()

	distinct := map[string]int{}
	for i, id := range given {
		if id == "" {
			t.Fatalf("volunteer v%d recorded no subject", i)
		}
		distinct[id]++
	}
	if len(distinct) != volunteers {
		t.Errorf("%d volunteers were given %d distinct subject(s): %v", volunteers, len(distinct), distinct)
	}
	if out := handouts.Out(now); out != volunteers {
		t.Errorf("%d subject(s) are out, want %d", out, volunteers)
	}
}

// TestTheDrawIsOfferedTheWholeEligibleSet is #35's second condition read at the
// place the promise is actually kept. docs/subject-selection.md promises a
// uniform draw over the eligible set; the randomness is injected, so what this
// mechanism owes is that the range handed to the draw is the eligible set and
// nothing narrower. A mechanism that offered the draw a scanned prefix, or the
// whole collection and then rejected the ineligible ones, would still look
// random and would not be uniform over what was promised.
func TestTheDrawIsOfferedTheWholeEligibleSet(t *testing.T) {
	handouts, err := KeepHoldsFor(15 * time.Minute)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	subjects := collection(t, 12)
	pool := Pool{
		Subjects: subjects,
		Retired:  map[string]bool{"plate-3": true, "plate-4": true},
		Answered: map[string]bool{"plate-1": true},
	}
	now := entered

	var offered []int
	watch := func(n int) int {
		offered = append(offered, n)
		return 0
	}

	// Twelve subjects, two retired and one answered, leaves nine. A different
	// volunteer asks each time and each draw holds one more, so the set
	// shrinks by one at every step.
	for want := 9; want > 0; want-- {
		if _, err := handouts.Next(fmt.Sprintf("v%d", want), pool, now, watch); err != nil {
			t.Fatalf("the draw at an eligible set of %d was refused: %v", want, err)
		}
	}
	wanted := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	if len(offered) != len(wanted) {
		t.Fatalf("the draw was called %d time(s), want %d", len(offered), len(wanted))
	}
	for i, n := range wanted {
		if offered[i] != n {
			t.Errorf("draw %d was offered %d subject(s), want %d", i+1, offered[i], n)
		}
	}
}

// TestEveryEligibleSubjectIsReachable is the other half of the distribution:
// no subject sits in the collection unreachable because of where it is in the
// order. The draw walks the whole range here, which is what a uniform source
// does over enough draws, and every subject that is not retired and not
// answered comes out exactly once.
func TestEveryEligibleSubjectIsReachable(t *testing.T) {
	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	subjects := collection(t, 20)
	pool := Pool{Subjects: subjects, Retired: map[string]bool{"plate-20": true}}
	now := entered

	// A draw that walks the range: the last index, then the last but one, and
	// so on, so the choice is not the head every time and the tail of the set
	// is reached as well.
	step := 0
	walk := func(n int) int {
		at := (n - 1 - step) % n
		if at < 0 {
			at += n
		}
		step++
		return at
	}

	seen := map[string]bool{}
	for i := range 19 {
		handout, err := handouts.Next(fmt.Sprintf("v%d", i), pool, now, walk)
		if err != nil {
			t.Fatalf("draw %d was given nothing: %v", i, err)
		}
		id := handout.Subject().ID()
		if seen[id] {
			t.Fatalf("draw %d was given %s, which is already out with somebody", i, id)
		}
		seen[id] = true
	}
	if len(seen) != 19 {
		t.Errorf("%d subject(s) were reached, want 19", len(seen))
	}
	if seen["plate-20"] {
		t.Error("a retired subject was handed out")
	}
}

// TestAVolunteerAskingAgainGetsTheSubjectTheyAlreadyHold is the reading taken
// where docs/subject-selection.md stops. Drawing again would leave the first
// hold standing, so a volunteer reloading the page four times would take four
// subjects out of the campaign and answer one. The hold is not extended by the
// asking either, or a volunteer reloading every minute would hold a subject for
// as long as they cared to and the timeout would bound nothing.
func TestAVolunteerAskingAgainGetsTheSubjectTheyAlreadyHold(t *testing.T) {
	const hold = 10 * time.Minute
	handouts, err := KeepHoldsFor(hold)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: collection(t, 6)}
	now := entered

	given, err := handouts.Next("v1", pool, now, first)
	if err != nil {
		t.Fatalf("the volunteer was given nothing: %v", err)
	}

	later := now.Add(4 * time.Minute)
	again, err := handouts.Next("v1", pool, later, func(int) int {
		t.Error("the draw was consulted for a volunteer who already holds a subject")
		return 0
	})
	if err != nil {
		t.Fatalf("the volunteer asking again was given nothing: %v", err)
	}
	if again.Subject().ID() != given.Subject().ID() {
		t.Errorf("asking again gave %s, want %s", again.Subject().ID(), given.Subject().ID())
	}
	if !again.Until().Equal(given.Until()) {
		t.Errorf("asking again moved the hold to %s, want %s", again.Until(), given.Until())
	}
	if out := handouts.Out(later); out != 1 {
		t.Errorf("%d subject(s) are out after two asks by one volunteer, want 1", out)
	}

	// A hold the pool has overtaken is dropped rather than handed back, and
	// the volunteer is drawn a fresh subject.
	retired := Pool{Subjects: pool.Subjects, Retired: map[string]bool{given.Subject().ID(): true}}
	fresh, err := handouts.Next("v1", retired, later, first)
	if err != nil {
		t.Fatalf("the volunteer was given nothing after their subject retired: %v", err)
	}
	if fresh.Subject().ID() == given.Subject().ID() {
		t.Error("a retired subject was handed back to the volunteer holding it")
	}
	if out := handouts.Out(later); out != 1 {
		t.Errorf("%d subject(s) are out, want 1: the overtaken hold was not dropped", out)
	}
}

// TestAVolunteerWhoAbandonsASubjectReleasesIt is #35's third condition. The
// clock is moved rather than waited on, which is the whole reason Depends.Now
// exists: a fifteen minute hold is reached in a microsecond.
func TestAVolunteerWhoAbandonsASubjectReleasesIt(t *testing.T) {
	const hold = 15 * time.Minute
	handouts, err := KeepHoldsFor(hold)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: collection(t, 1)}
	now := entered

	abandoned, err := handouts.Next("v1", pool, now, first)
	if err != nil {
		t.Fatalf("the first volunteer was given nothing: %v", err)
	}
	if got, want := abandoned.Until(), now.Add(hold); !got.Equal(want) {
		t.Errorf("the hold runs until %s, want %s", got, want)
	}

	// v1 closes the tab. Nothing is released, and while the hold stands there
	// is nothing for anybody else.
	stillHeld := now.Add(hold - time.Second)
	var nothing NoSubject
	_, err = handouts.Next("v2", pool, stillHeld, first)
	if !errors.As(err, &nothing) {
		t.Fatalf("the second volunteer was given a subject that is still held: %v", err)
	}
	if nothing.Because != OutWithOthers {
		t.Errorf("the reason is %q, want %q", nothing.Because, OutWithOthers)
	}
	if nothing.RetryAfter != time.Second {
		t.Errorf("the wait is %s, want %s", nothing.RetryAfter, time.Second)
	}

	// The hold runs out and the subject is eligible again, for somebody who
	// has not answered it.
	expired := now.Add(hold)
	recovered, err := handouts.Next("v2", pool, expired, first)
	if err != nil {
		t.Fatalf("the subject did not come back after its hold expired: %v", err)
	}
	if recovered.Subject().ID() != abandoned.Subject().ID() {
		t.Errorf("the recovered subject is %s, want %s", recovered.Subject().ID(), abandoned.Subject().ID())
	}
}

// TestAnExpiredHoldIsDroppedByTheRequestRatherThanByATimer is #35's fourth
// condition where it is decidable in a test. The clock moves past the hold and
// nothing else happens: no goroutine runs, no ticker fires and no second
// process exists, and the hold is gone by the time anybody asks about it.
func TestAnExpiredHoldIsDroppedByTheRequestRatherThanByATimer(t *testing.T) {
	handouts, err := KeepHoldsFor(time.Minute)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: collection(t, 3)}
	now := entered

	if _, err := handouts.Next("v1", pool, now, first); err != nil {
		t.Fatalf("the volunteer was given nothing: %v", err)
	}
	if out := handouts.Out(now); out != 1 {
		t.Fatalf("%d subject(s) are out, want 1", out)
	}
	if out := handouts.Out(now.Add(time.Minute)); out != 0 {
		t.Errorf("%d subject(s) are still out a minute later, want 0", out)
	}
}

// TestAnsweringReleasesTheHoldAtThatMoment is the other release
// docs/subject-selection.md names, and the pair of refusals under it are what
// keeps a late or repeated release from taking a subject away from the
// volunteer who holds it now.
func TestAnsweringReleasesTheHoldAtThatMoment(t *testing.T) {
	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: collection(t, 2)}
	now := entered

	given, err := handouts.Next("v1", pool, now, first)
	if err != nil {
		t.Fatalf("the volunteer was given nothing: %v", err)
	}
	if !handouts.Release("v1", given.Subject().ID()) {
		t.Fatal("the volunteer who holds the subject could not release it")
	}
	if out := handouts.Out(now); out != 0 {
		t.Errorf("%d subject(s) are out after the answer, want 0", out)
	}
	if handouts.Release("v1", given.Subject().ID()) {
		t.Error("the same release was accepted twice")
	}

	// v2 now holds it. A release arriving late from v1 must not take it away.
	again, err := handouts.Next("v2", pool, now, first)
	if err != nil {
		t.Fatalf("the second volunteer was given nothing: %v", err)
	}
	if handouts.Release("v1", again.Subject().ID()) {
		t.Error("a release named somebody else's hold and was accepted")
	}
	if out := handouts.Out(now); out != 1 {
		t.Errorf("%d subject(s) are out, want 1", out)
	}
}

// TestTheThreeEmptyCasesAreDifferentAnswers holds what
// docs/subject-selection.md is most explicit about: a volunteer told the
// campaign is finished when it is not will believe the campaign is finished.
func TestTheThreeEmptyCasesAreDifferentAnswers(t *testing.T) {
	subjects := collection(t, 2)
	now := entered

	finished, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	_, err = finished.Next("v1", Pool{
		Subjects: subjects,
		Retired:  map[string]bool{"plate-1": true, "plate-2": true},
	}, now, first)
	assertNothing(t, err, CampaignFinished, 0)

	done, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	_, err = done.Next("v1", Pool{
		Subjects: subjects,
		Answered: map[string]bool{"plate-1": true, "plate-2": true},
	}, now, first)
	assertNothing(t, err, AnsweredEverything, 0)

	busy, err := KeepHoldsFor(30 * time.Minute)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: subjects}
	if _, err := busy.Next("v2", pool, now, first); err != nil {
		t.Fatalf("v2 was given nothing: %v", err)
	}
	if _, err := busy.Next("v3", pool, now.Add(time.Minute), first); err != nil {
		t.Fatalf("v3 was given nothing: %v", err)
	}
	_, err = busy.Next("v1", pool, now.Add(2*time.Minute), first)
	assertNothing(t, err, OutWithOthers, 28*time.Minute)

	// A campaign whose subjects are all answered by this volunteer and all
	// retired is finished rather than answered out: the fact about the
	// campaign outranks the fact about the volunteer.
	both, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	_, err = both.Next("v1", Pool{
		Subjects: subjects,
		Retired:  map[string]bool{"plate-1": true, "plate-2": true},
		Answered: map[string]bool{"plate-1": true, "plate-2": true},
	}, now, first)
	assertNothing(t, err, CampaignFinished, 0)
}

// assertNothing reads a refusal to hand out work and checks it is the one
// expected, with the wait it should carry.
func assertNothing(t *testing.T, err error, because Reason, retryAfter time.Duration) {
	t.Helper()
	var nothing NoSubject
	if !errors.As(err, &nothing) {
		t.Fatalf("the draw returned %v, want a campaign.NoSubject", err)
	}
	if nothing.Because != because {
		t.Errorf("the reason is %q, want %q", nothing.Because, because)
	}
	if nothing.RetryAfter != retryAfter {
		t.Errorf("the wait is %s, want %s", nothing.RetryAfter, retryAfter)
	}
	if nothing.Error() == "" {
		t.Error("the refusal says nothing")
	}
}

// TestAHoldThatKeepsNothingIsRefused holds the constructor. A hold of zero
// hands one subject to every volunteer asking at the same moment, which is the
// collision this mechanism exists against, and it would do it silently.
func TestAHoldThatKeepsNothingIsRefused(t *testing.T) {
	for _, hold := range []time.Duration{0, -time.Second, -time.Hour} {
		if _, err := KeepHoldsFor(hold); err == nil {
			t.Errorf("a hold of %s was accepted", hold)
		}
	}
	if _, err := KeepHoldsFor(time.Nanosecond); err != nil {
		t.Errorf("a hold of a nanosecond was refused: %v", err)
	}
}

// TestADrawOutsideItsIntervalHandsOutNothing is the one thing the mechanism
// cannot check about the source it was given. Depends.Intn declares [0, n), and
// a fake or a future source that breaks that would otherwise index whatever the
// number happened to reach, or hand out a subject nobody chose.
func TestADrawOutsideItsIntervalHandsOutNothing(t *testing.T) {
	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	pool := Pool{Subjects: collection(t, 3)}
	now := entered

	for _, at := range []int{-1, 3, 4000} {
		if _, err := handouts.Next("v1", pool, now, func(int) int { return at }); err == nil {
			t.Errorf("a draw returning %d handed out a subject", at)
		}
	}
	if out := handouts.Out(now); out != 0 {
		t.Errorf("%d subject(s) are out after three refused draws, want 0", out)
	}
}

// TestASubjectThisVolunteerAnsweredIsNeverHandedBack is the property
// docs/subject-selection.md holds exactly rather than rarely, because re-serving
// one would buy a classification and cost the independence the consensus rule
// rests on.
func TestASubjectThisVolunteerAnsweredIsNeverHandedBack(t *testing.T) {
	handouts, err := KeepHoldsFor(time.Hour)
	if err != nil {
		t.Fatalf("the holds were refused: %v", err)
	}
	subjects := collection(t, 5)
	answered := map[string]bool{}
	now := entered

	for i := range 5 {
		handout, err := handouts.Next("v1", Pool{Subjects: subjects, Answered: answered}, now, first)
		if err != nil {
			t.Fatalf("draw %d was given nothing: %v", i, err)
		}
		id := handout.Subject().ID()
		if answered[id] {
			t.Fatalf("draw %d handed back %s, which this volunteer had answered", i, id)
		}
		answered[id] = true
		if !handouts.Release("v1", id) {
			t.Fatalf("draw %d could not be released", i)
		}
	}
	_, err = handouts.Next("v1", Pool{Subjects: subjects, Answered: answered}, now, first)
	assertNothing(t, err, AnsweredEverything, 0)
}
