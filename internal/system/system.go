// Package system supplies the real values behind campaign.Depends: the wall
// clock, a random source, an identifier generator and an outbound dialler.
//
// It is a package of its own so that reading the runtime is a thing one
// directory does rather than a habit spread across the tree. harness_test.go
// refuses a runtime read anywhere else, and docs/layout.md lets only the entry
// point import this package, so the wiring in main.go is the single place the
// real values enter the program.
//
// Nothing here is injectable and nothing here is tested against. A test that
// wants a clock takes campaigntest's, which is the whole point of the split.
package system

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/iderex/lesesaal/internal/campaign"
)

// dialTimeout bounds an outbound connection attempt. A dial with no timeout
// inherits the operating system's, which is minutes on some hosts, and a
// deployment that hangs for minutes on a name that no longer resolves looks
// like a deployment that has stopped rather than one that is waiting.
const dialTimeout = 10 * time.Second

// Depends returns the real dependencies. It is the only constructor here, so
// there is no way to take the real clock without also taking the rest and
// making the wiring visible in one expression.
func Depends() campaign.Depends {
	dialler := &net.Dialer{Timeout: dialTimeout}
	return campaign.Depends{
		Now:   time.Now,
		Intn:  intn,
		NewID: newID,
		Dial:  dialler.DialContext,
	}
}

// intn returns a number in [0, n) drawn from the operating system's source.
// The cryptographic source is used for the ordinary one as well, rather than
// keeping two: this project draws to choose which subject a volunteer sees
// next, that choice is observable by whoever is classifying, and a predictable
// sequence there is a way to work out what somebody else was shown.
//
// It panics on a source failure rather than returning a zero. A zero would be a
// valid draw, so a caller could not tell a failure from a result, and a random
// source that has stopped working is not a condition this project can carry on
// through.
func intn(n int) int {
	if n <= 0 {
		panic(fmt.Sprintf("system: Intn called with n = %d, which names no interval", n))
	}
	drawn, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(fmt.Sprintf("system: the random source failed: %v", err))
	}
	return int(drawn.Int64())
}

// idBytes is the length of an identifier before it is hexadecimal. Sixteen
// bytes is what a version 4 identifier carries, and it is the size at which a
// collision is not something this project has to think about again.
const idBytes = 16

// newID returns an identifier that has not been returned before. It panics for
// the same reason intn does: an identifier that is silently the empty string
// would be accepted everywhere and would collide with the next one.
func newID() string {
	raw := make([]byte, idBytes)
	if _, err := rand.Read(raw); err != nil {
		panic(fmt.Sprintf("system: the random source failed while making an identifier: %v", err))
	}
	return hex.EncodeToString(raw)
}
