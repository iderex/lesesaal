package campaign

// This file exists on this branch only. With the requirement added by the
// commit before it, the module file now names a module and one file imports it,
// and no lock file covers either. This is what a dependency added without
// updating the lock looks like, and the build check is what refuses it.

import "golang.org/x/text/language"

// probeTag exists to make the import above real. An import nothing uses is
// removed by the compiler's own rules before any check sees it.
var probeTag = language.English
