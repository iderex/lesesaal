package campaign

// This file exists on this branch only. It is the near miss for the sleep rule:
// a test that waits for a wall-clock interval instead of moving a clock, which
// is the commonest single cause of a suite that fails once a week.

import (
	"testing"
	"time"
)

func TestSomethingHappensEventually(t *testing.T) {
	started := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	time.Sleep(10 * time.Millisecond)

	if started.IsZero() {
		t.Fatal("the fixture instant is the zero time")
	}
}
