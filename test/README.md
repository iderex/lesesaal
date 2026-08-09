# The harnesses the unit suite cannot host

What belongs here: a test that needs something the unit suite is forbidden to
have. A browser driving the volunteer surface, a container runtime starting the
deployment end to end, a real collection of images on disk. Each one is a
harness with its own entry point, its own runner and its own name in a
workflow, so that a run which skipped it says so.

What does not belong here: anything a unit test could do instead. A test placed
here to escape the no-display, no-network and no-elevation rules is the erosion
those rules exist against, and moving it here does not make it a slow test, it
makes it an unwatched one.

A unit test lives beside the code it judges, in the same package directory, as
`layout_test.go` does at the module root. That is this toolchain's own
convention and this directory is not an alternative to it.

There is no harness here yet. The browser harness is #52, the end to end
deployment harness is #80, and the fixtures each of them needs arrive with the
harness rather than in a directory of their own.

One thing that is not a harness lives here anyway. `test/gate-refusals.md`
records what each check in this repository has been observed to refuse, with the
run where it reddened, and it sits beside the harnesses because it is about
whether the gate works rather than about whether the software does.

This file is the placeholder `docs/layout.md` promises in every directory where
code does not exist yet.
