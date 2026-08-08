package campaign

import (
	"context"
	"net"
	"time"
)

// Depends is everything this program takes from the runtime instead of reading
// out of it. Most of this system is about time and order: a subject is retired
// once enough answers arrive, a session expires, a campaign opens and closes,
// and a model retrains between sittings. A test written against the real clock
// for any of that is slow, is flaky, and is quietly a test of the machine it
// ran on.
//
// It is declared here, in the core, because the core is the one part every
// other part is allowed to import. Nothing supplies the real values from here:
// internal/system builds them and main.go is the only caller, which is the
// direction docs/layout.md fixes and harness_test.go refuses a departure from.
//
// The fields are functions rather than an interface per concern. A test that
// needs only a clock writes one field and leaves the rest, and adding a concern
// later does not force every existing fake to grow a method it ignores.
type Depends struct {
	// Now is the current time. Nothing outside internal/system reads the
	// runtime clock, so a test moves a year forward in a microsecond by
	// returning a different value here.
	Now func() time.Time

	// Intn returns a number in the half-open interval [0, n). It is what the
	// selection rule in #10 and the sampling in #56 draw from, so a test can
	// fix the draw and assert the choice rather than the distribution.
	Intn func(n int) int

	// NewID returns an identifier that has not been returned before. Injected
	// for the same reason as the draw: an assertion against a generated
	// identifier is only possible where the test decides what it is.
	NewID func() string

	// Dial opens an outbound connection. It exists so that the absence of one
	// is visible: a fake that refuses every address turns a dependency that
	// starts calling somewhere into a failing test rather than into a
	// firewall's problem months later.
	//
	// Nothing in this project makes an outbound connection today, and
	// docs/federation.md is where that is a position rather than an accident.
	Dial func(ctx context.Context, network, address string) (net.Conn, error)
}

// Missing names the fields that were left nil, in the order they are declared.
// A nil field is a dependency nobody supplied, and a program that starts on one
// panics at the first call rather than at startup, which is the wrong end of
// the run to find out.
//
// It returns nil rather than an empty slice when everything is present, so a
// caller can test the result for emptiness either way.
func (d Depends) Missing() []string {
	var missing []string
	if d.Now == nil {
		missing = append(missing, "Now")
	}
	if d.Intn == nil {
		missing = append(missing, "Intn")
	}
	if d.NewID == nil {
		missing = append(missing, "NewID")
	}
	if d.Dial == nil {
		missing = append(missing, "Dial")
	}
	return missing
}
