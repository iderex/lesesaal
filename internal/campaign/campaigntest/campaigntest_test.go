package campaigntest

import (
	"strings"
	"testing"
	"time"
)

// TestAClockMovesAYearAtOnce is the condition #21 states last: a test can
// advance time by an arbitrary amount without waiting. A year is chosen because
// it is longer than any real campaign and because it is the sort of interval a
// retention rule will be written against.
//
// That it costs no wall time is not asserted here. Asserting it would mean
// reading the real clock, which is the thing this package exists to stop, so
// the evidence is the suite's own reported duration instead.
func TestAClockMovesAYearAtOnce(t *testing.T) {
	clock := NewClock(time.Time{})
	start := clock.Now()

	const year = 365 * 24 * time.Hour
	landed := clock.Advance(year)

	if want := start.Add(year); !landed.Equal(want) {
		t.Errorf("advancing by a year landed at %s and should have landed at %s", landed, want)
	}
	if !clock.Now().Equal(landed) {
		t.Errorf("the clock reads %s after landing at %s", clock.Now(), landed)
	}
	if !start.Equal(Epoch) {
		t.Errorf("a clock made with the zero time started at %s rather than at the epoch %s", start, Epoch)
	}
}

// TestAClockStandsStill refuses a fake that has picked up a life of its own. A
// clock that moves between two reads is the real one wearing this one's name,
// and every assertion about an interval built on it would be a race.
func TestAClockStandsStill(t *testing.T) {
	clock := NewClock(time.Date(2027, time.March, 4, 12, 0, 0, 0, time.UTC))

	first := clock.Now()
	for range 1000 {
		if got := clock.Now(); !got.Equal(first) {
			t.Fatalf("the clock moved on its own, from %s to %s", first, got)
		}
	}
}

// TestAClockRefusesToGoBackwards covers the panic rather than leaving it as a
// comment. A test that wrote a minus sign by accident would otherwise pass for
// a reason nobody could see.
func TestAClockRefusesToGoBackwards(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("advancing by a negative duration was accepted")
		}
	}()

	NewClock(time.Time{}).Advance(-time.Second)
}

// TestTwoDrawsWithOneSeedAgree is what makes a selection rule testable at all.
// Without it a test could assert only that a choice was in range.
func TestTwoDrawsWithOneSeedAgree(t *testing.T) {
	first := NewDraw(7)
	second := NewDraw(7)

	for call := range 100 {
		want := first.Intn(1000)
		if got := second.Intn(1000); got != want {
			t.Fatalf("call %d of two sequences seeded with 7 gave %d and %d", call, want, got)
		}
	}
}

// TestADrawStaysInsideItsInterval checks the one promise the real source makes.
// A draw of n itself, or of a negative number, would be an out-of-range index
// somewhere later and a panic a long way from here.
func TestADrawStaysInsideItsInterval(t *testing.T) {
	draw := NewDraw(0)

	for _, n := range []int{1, 2, 7, 64, 1000} {
		for range 500 {
			got := draw.Intn(n)
			if got < 0 || got >= n {
				t.Fatalf("a draw over %d returned %d, which is outside [0, %d)", n, got, n)
			}
		}
	}
}

// TestADrawSeededWithZeroMoves is the near miss behind the offset in NewDraw.
// This generator's step leaves a state of zero at zero, so a sequence seeded
// with the zero value would return the same number for ever and every test
// using the default seed would agree with itself for the wrong reason.
func TestADrawSeededWithZeroMoves(t *testing.T) {
	draw := NewDraw(0)

	first := draw.Intn(1 << 20)
	for range 20 {
		if draw.Intn(1<<20) != first {
			return
		}
	}
	t.Fatalf("twenty-one draws from a sequence seeded with zero all returned %d", first)
}

// TestNamesDoNotRepeat covers the only promise the real identifier source
// makes. A source handing out one identifier twice would silently join two
// classifications into one.
func TestNamesDoNotRepeat(t *testing.T) {
	names := NewNames("subject")

	seen := make(map[string]bool)
	for range 1000 {
		id := names.NewID()
		if seen[id] {
			t.Fatalf("the identifier %s was handed out twice", id)
		}
		seen[id] = true
	}
	if names.Issued() != len(seen) {
		t.Errorf("the source reports %d issued and handed out %d distinct", names.Issued(), len(seen))
	}
}

// TestTheDiallerRefuses is the no-network half of the harness. A dependency
// that starts calling somewhere has to be caught by the suite rather than by a
// firewall, and the failure has to name the address or nobody can tell what
// called out.
func TestTheDiallerRefuses(t *testing.T) {
	conn, err := RefuseDial(t.Context(), "tcp", "example.invalid:443")

	if err == nil {
		t.Fatal("the dialler opened a connection")
	}
	if conn != nil {
		t.Error("the dialler refused and returned a connection anyway")
	}
	for _, want := range []string{"example.invalid:443", "tcp"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %s: %s", want, err)
		}
	}
}

// TestDependsSuppliesEveryField refuses a set of fakes that has fallen behind
// the fields the program takes. A nil field would panic at the first call in
// whichever test happened to reach it first.
func TestDependsSuppliesEveryField(t *testing.T) {
	depends, clock, draw, names := Depends()

	if missing := depends.Missing(); len(missing) != 0 {
		t.Errorf("the fakes leave %v unsupplied", missing)
	}
	if clock == nil || draw == nil || names == nil {
		t.Fatal("the fakes were returned without one of the things behind them")
	}

	clock.Advance(time.Hour)
	if got := depends.Now(); !got.Equal(clock.Now()) {
		t.Errorf("the Now field reads %s and the clock behind it reads %s", got, clock.Now())
	}
	firstID := depends.NewID()
	if secondID := depends.NewID(); firstID == secondID {
		t.Errorf("the NewID field returned %s twice", firstID)
	}
}
