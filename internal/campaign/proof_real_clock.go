package campaign

// This file exists on this branch only. It is the near miss the harness guard
// has to refuse: the core asking the runtime what time it is, instead of taking
// the Now field it already declares.

import "time"

// Expired reports whether a session that started at the instant given has run
// out. It reads the wall clock, so a test of it can only be written by waiting.
func Expired(startedAt time.Time, lifetime time.Duration) bool {
	return time.Now().After(startedAt.Add(lifetime))
}
