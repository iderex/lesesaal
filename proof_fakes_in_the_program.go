package main

// This file exists on this branch only. It is the near miss for the rule that
// keeps the fakes out of the program: a non-test file importing them, which
// would put a clock that does not move into the binary an operator runs.

import "github.com/iderex/lesesaal/internal/campaign/campaigntest"

// The blank identifier rather than a name, so that the import is the whole of
// what this file does and no linter has an opinion about the rest of it.
var _ = campaigntest.Epoch
