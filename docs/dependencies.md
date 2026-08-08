# Dependencies, and what pins each of them

Written for issue #23. Everything the quality milestone asks for assumes the
dependency set is known. A project that resolves versions at build time has no
such set, so a scan of it proves nothing about what will be installed tomorrow.

This document says what the set is, what pins each part of it, what the pin is
worth, and what happens when one of them is updated.

## The Go module

The set is empty:

    go list -mod=readonly -m all
    github.com/iderex/lesesaal

    git ls-files go.sum | wc -l
    0

One line is this module itself. There is no lock file because the toolchain
creates one the moment a requirement needs one, so its absence and an empty
requirement set are the same fact rather than two.

Three things hold it, and they refuse in different directions.

`-mod=readonly` refuses to add or update a requirement in order to satisfy an
import. It is on every command the build check and the lint check run, so a file
importing something the module file does not carry fails rather than being
resolved quietly.

The build check refuses a module the lock file does not cover. Measured at
d5ed98f, with a requirement added to the module file and a file importing it and
no lock file written:

    go build -mod=readonly ./... ; echo $?
    internal\campaign\tmpprobe.go:3:8: missing go.sum entry for module providing package golang.org/x/text/language (imported by github.com/iderex/lesesaal/internal/campaign); to add:
    1

The second line of that message names the command that would add the entry. It
is cut here because it arrives indented with a tab, and a tab is a byte this
tree does not store in a document.

`Dependency lock` refuses the other direction, which nothing refused before it.
A requirement in the module file that nothing in the tree imports passes the
build check, because every import still resolves. Measured on the same tree:

    go list -mod=readonly -deps ./... > /dev/null ; echo $?
    0
    go mod tidy -diff ; echo $?
    diff current/go.mod tidy/go.mod
    --- current/go.mod
    +++ tidy/go.mod
    @@ -1,5 +1,3 @@
    -require golang.org/x/text v0.3.8
    1

The unchanged context lines of that diff are cut and nothing else is.

That leg reports what a tidy run would change and changes nothing, which is the
difference between a verdict and a repair. A check that tidied the file would
let a drifted one land and then correct it, and nobody would see the drift.

## The GitHub Actions the workflows use

Ten distinct actions, each pinned to a commit sha with the version in a comment
beside it:

    grep -rhoE '^\s*uses:\s*\S+' .github/workflows/ | sed -E 's/^\s*uses:\s*//' \
      | sort -u | wc -l
    10

There is no lock file for this ecosystem and no format that would hold one. A
commit sha is the integrity pin: it names a tree rather than a moving label, and
it is the strongest pin available here. `Audit workflows (zizmor)` refuses an
action that is not pinned that way, so this ecosystem has a guard even though it
has no lock file.

What the sha does not give is a transitive set. An action can itself use other
actions, and nothing here reads those. The bill of materials counts what the
workflows declare, not what the actions they call declare.

## The tools a workflow installs while it runs

Two, both pinned by a version string rather than by a hash:

    git grep -nE 'go install |ZIZMOR_VERSION' -- .github/workflows/
    .github/workflows/lint.yml:141:          GOFLAGS= go install honnef.co/go/tools/cmd/staticcheck@v0.7.0
    .github/workflows/zizmor.yml:47:      ZIZMOR_VERSION: "1.26.1"

The linter is fetched through the Go module proxy, whose checksum database is
what makes that version mean a fixed set of bytes. The workflow auditor is
fetched from a Python package index, where a version is a label the index
resolves and no checksum in this tree constrains it; `--no-build` narrows what
that installs to a prebuilt wheel rather than a source build that runs a script,
which is a different guarantee and a smaller one.

**Nothing refuses a floating version here.** A change replacing either pin with
a moving label would pass every check in this repository. The guard that would
refuse it belongs to this ecosystem rather than to the module, and it is not
written. This is the second ecosystem #23 warns about, and the warning is
accurate.

## What happens when a dependency is updated

A dependency bump is a change like any other. It starts as an issue, it lands as
a pull request, and it carries the same evidence.

What the person reviewing it looks at, in this order. Whether the diff to the
module file and the lock file is only the dependency being named, so a bump does
not carry an unrelated requirement with it. Whether the new version's own
requirements entered the graph, which the module listing shows and the bill of
materials confirms. What the upstream change actually was, read at the tag being
moved to rather than in a summary of it. Whether the reason for the bump is
written down: a fix that matters here, an advisory, or a version this project
needs for something it is doing, and not that a newer one exists.

The cadence is on demand rather than on a timer. Nothing in this repository
opens a dependency update on a schedule today, so an update happens when
somebody has a reason. The cost of that is stated rather than hidden: an
advisory against a pinned version is noticed when somebody looks, and the thing
that would notice it sooner is a scheduled update route this repository does not
have.

Two checks already stand behind a bump whenever one is made. `dependency-review`
refuses a newly introduced or upgraded dependency carrying a known advisory at
any severity. `Bill of materials` refuses a document that stopped covering an
ecosystem the tree carries, so a bump that quietly removes one is caught.

## What is not covered

A lock file with integrity hashes, because there is nothing to lock. The first
condition of #23 is met by the day the first requirement lands, not by this
document, and the checks above are in place so that day is uneventful.

A second ecosystem in the sense #23 means it, a runtime with its own package
manager beside the Go module. There is none. The web surface is served by the
standard library and the model runs in the same process, both decided in
`docs/toolchain.md`, so the shape where a project has a locked backend and an
unlocked frontend does not exist here yet.

The container image. `deploy/` holds a placeholder and #74 builds the image; the
base image and whatever it carries are outside every count on this page.
