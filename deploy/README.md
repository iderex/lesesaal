# Deployment assets

What belongs here: everything an operator needs to run this that is not the
program itself. The container file, the composition that starts it, and the
example configuration a first run is walked through with.

What does not belong here: anything the program imports. Nothing in this
directory is compiled and nothing in `internal/` may read it, which is the rule
`docs/layout.md` states and the guard in `layout_test.go` holds.

There is nothing here yet. The composition and the image are #72 and #74, the
configuration is #82, and the number of processes any of them may start is
fixed at one in `docs/deployment-ceiling.md` rather than argued again here.

This file is the placeholder `docs/layout.md` promises in every directory where
code does not exist yet. It is replaced by the first asset that lands, not kept
beside it.
