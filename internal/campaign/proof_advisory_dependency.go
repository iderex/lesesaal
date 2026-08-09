package campaign

// This file exists on this branch only. It is the near miss the dependency
// review in the gate has to refuse: a dependency added at the version somebody
// already had in another project, one patch release below the one that carries
// the fix.
//
// The requirement is real and the lock file covers it, so the build, the lock
// check and the linter all stay green and the only red verdict comes from the
// advisory database.

import "golang.org/x/text/language"

// ProbeTag exists to make the import above real. An import nothing uses is
// refused by the compiler before any check sees it.
var ProbeTag = language.English
