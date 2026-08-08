// Package campaigntest supplies the fakes a test injects in place of the
// runtime: a clock a test moves by hand, a draw that repeats exactly, an
// identifier source that counts, and a dialler that refuses.
//
// It is a package rather than a file inside each test so that one clock is
// shared by the suites of every part, and it sits under the core because the
// core is what every part may import. Only a test file may import it, which
// harness_test.go refuses a departure from, so nothing here reaches the
// program an operator runs.
//
// Nothing here reads the runtime. The clock starts where the test puts it, the
// draw is a fixture generator rather than a source of randomness, and the
// dialler opens nothing. That is what makes a suite built on it the same suite
// on every machine and at every hour.
package campaigntest

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/iderex/lesesaal/internal/campaign"
)

// Epoch is where a clock starts unless a test says otherwise. It is a real
// instant rather than the zero time, because the zero time is year 1 and a
// subtraction against it produces durations no test author expected.
var Epoch = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

// Clock is a clock a test moves. It is not safe for use from several goroutines
// at once, which is deliberate: a test that needs one to be would be a test
// about concurrency, and it should say so and build its own.
type Clock struct {
	now time.Time
}

// NewClock returns a clock reading at the instant given. Pass the zero value to
// start at Epoch, which is what a test that does not care about the absolute
// time wants.
func NewClock(start time.Time) *Clock {
	if start.IsZero() {
		start = Epoch
	}
	return &Clock{now: start}
}

// Now reports the instant the clock is at. It never moves on its own.
func (c *Clock) Now() time.Time { return c.now }

// Advance moves the clock forward by d and reports where it landed. A year
// costs the same as a nanosecond, which is the property the whole harness
// exists for: retirement, expiry and retraining are all reachable in a test
// without waiting for any of them.
//
// It panics on a negative duration. Time going backwards is not a scenario this
// project has, and a test that wrote a minus sign by accident would otherwise
// pass for a reason nobody could see.
func (c *Clock) Advance(d time.Duration) time.Time {
	if d < 0 {
		panic(fmt.Sprintf("campaigntest: Advance called with %v, and this clock does not go backwards", d))
	}
	c.now = c.now.Add(d)
	return c.now
}

// Draw is a deterministic sequence standing in for a random one. The same seed
// gives the same numbers on every machine and in every run, so a test can fix
// which subject is chosen and assert the choice rather than the distribution.
//
// It is a fixture generator and not a source of randomness. It is written out
// here rather than taken from the standard library because nothing outside
// internal/system may read a random source at all, and because a fixture whose
// sequence depends on a library version is a fixture that changes under a test
// that did not.
type Draw struct {
	state uint64
}

// NewDraw returns a sequence seeded with the value given. Two sequences with
// one seed produce identical numbers.
func NewDraw(seed uint64) *Draw {
	// The seed is offset so that a seed of zero is a sequence like any other
	// rather than a stuck one: this generator's step leaves zero at zero.
	return &Draw{state: seed + 0x9e3779b97f4a7c15}
}

// Intn returns the next number in [0, n). It panics on an n that names no
// interval, the same way the real one does, so a test cannot pass against the
// fake and fail against the program.
func (d *Draw) Intn(n int) int {
	if n <= 0 {
		panic(fmt.Sprintf("campaigntest: Intn called with n = %d, which names no interval", n))
	}
	// A 64-bit mix step. What matters here is that it is fixed, not that it is
	// good: this is not used to make anything unpredictable.
	d.state += 0x9e3779b97f4a7c15
	z := d.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	z ^= z >> 31
	return int(z % uint64(n))
}

// Names is an identifier source that counts. It hands out prefix-1, prefix-2
// and so on, so an identifier in a failure message says which call produced it.
type Names struct {
	prefix string
	issued int
}

// NewNames returns a source using the prefix given, or "id" where the prefix is
// empty.
func NewNames(prefix string) *Names {
	if prefix == "" {
		prefix = "id"
	}
	return &Names{prefix: prefix}
}

// NewID returns the next identifier. It has not been returned before by this
// source, which is the only promise the real one makes.
func (n *Names) NewID() string {
	n.issued++
	return fmt.Sprintf("%s-%d", n.prefix, n.issued)
}

// Issued reports how many identifiers this source has handed out, so a test can
// assert that something asked for one rather than reusing another.
func (n *Names) Issued() int { return n.issued }

// RefuseDial is the dialler a test injects. It opens nothing and returns an
// error naming the address, so a dependency that starts calling somewhere is
// caught by the suite with the address in the failure rather than by a
// firewall, or by nothing at all.
func RefuseDial(_ context.Context, network, address string) (net.Conn, error) {
	return nil, fmt.Errorf("campaigntest: this suite makes no outbound connection, and something tried to reach %s over %s", address, network)
}

// Depends returns a complete set of fakes together with the clock, the draw and
// the identifier source behind it, so a test can move time and fix a choice
// without assembling four pieces first.
//
// The dialler refuses. A test that wants an outbound connection to succeed is
// not a unit test, and test/ is where the harnesses that may have one live.
func Depends() (campaign.Depends, *Clock, *Draw, *Names) {
	clock := NewClock(time.Time{})
	draw := NewDraw(1)
	names := NewNames("")
	return campaign.Depends{
		Now:   clock.Now,
		Intn:  draw.Intn,
		NewID: names.NewID,
		Dial:  RefuseDial,
	}, clock, draw, names
}
