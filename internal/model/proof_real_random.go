package model

// This file exists on this branch only. It is the near miss for the second
// injected dependency: a random source read where it sits rather than taken as
// a field, which leaves a selection rule testable only as a distribution.

import (
	"crypto/rand"
	"math/big"
)

// Pick returns a position in [0, n). It draws from the operating system's
// source, so no test can fix what it returns.
func Pick(n int) int {
	drawn, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(drawn.Int64())
}
