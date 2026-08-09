# Contributing

This project plans on the issue tracker first and lands every change as a pull
request. What follows is what somebody arriving with a fix needs before the
first one. Read it once; the commands are the part worth coming back to.

## Getting a working environment

Git, and the Go toolchain at the version `go.mod` declares. Nothing else is
needed to build the program or to run its tests.

    git clone https://github.com/iderex/lesesaal.git
    cd lesesaal
    go build ./...
    go test ./...

`go.mod` is the only place the toolchain version is written down, and every
workflow reads it from there rather than repeating it, so a contributor and the
server compile with the same toolchain by construction rather than by
agreement. `docs/toolchain.md` is where the language was chosen, and it says
what the choice costs as well as what it buys.

The module carries no dependency at all today, so a first build needs no
network and there is nothing to restore. One gate leg needs a program the
toolchain does not ship, installed by pinned version rather than from a package
manager:

    go install honnef.co/go/tools/cmd/staticcheck@v0.7.0

## Running the gate before you push

One command, and it is a verb of the program rather than a script beside it:

    go run . ci

The legs run in the order a reader wants them and the run stops at the first
failure, because the second finding is usually the first one again. Every leg
says what it examined, so a run that covered less than the whole set cannot be
read as one that covered it and found nothing.

The workflow steps call the same command by leg name, `go run . ci gofmt` and
so on, so the tree holds one procedure and not two. What the legs are is
printed by the command and is not listed here: a list in this document drifts
against the thing that decides it.

The run ends by saying what it did not do and why, which is the part worth
reading. Some of the gate is still shell inside a workflow file, some of it
needs a tool this tree does not carry, and some of it is server-side by
nature. All three are named in the report rather than left out of it.

The command does not install the linter the section above pins, and on a clone
without it the linter leg reports that it did not run and says that install
line back to you. A leg that did not run is not a leg that passed, and the
summary counts the three states apart rather than two.

Naming a leg requires it. `go run . ci staticcheck` on a machine without the
linter fails rather than reporting a disclosure, because asking for a leg by
name is what makes its absence the answer to a question somebody asked.

## No work without an issue

Every change starts as an issue. A change with no issue has nowhere to record
why it exists, and this project settles a decision on the tracker before
writing the code that depends on it.

An issue says what is wrong, what the evidence is, and what done means. Where
the evidence is a number, it carries the command that produced it. A `Scope:`
line at column zero names the paths the work may touch. Nothing in this
repository reads that line today; it is a convention every issue here already
follows, and the section below on foreign paths is what it is for.

The templates under `.github/ISSUE_TEMPLATE/` ask for that shape directly. A
bug report asks for the smallest reproduction rather than a description of one,
cut down until removing one more step makes the problem stop happening.

## Signing off a commit

Every commit carries a Developer Certificate of Origin sign-off, and the gate
refuses one that does not. The certificate is [DCO](DCO) at the root of this
repository, and signing off is how you assert it.

    git commit -s

The trailer has to match the commit's author exactly, name and address, which
is why it is produced by the flag rather than typed. Where a branch already
carries commits without it:

    git rebase --signoff <base>

Those are the two commands the check itself prints when it fails, in
`.github/workflows/dco.yml`. Commits authored by GitHub's own bots are exempt,
and the check lists those identities one at a time rather than matching
anything shaped like a bot address, so nobody can exempt themselves by choosing
a convincing author address.

Signing off is not signing. A cryptographic signature on a commit is a separate
control, the branch does not require one today, and `docs/commit-signing.md` is
the request that it should.

## What a commit message says

What changed, and what failure it prevents. Where the change is a correction it
also says what was wrong and how it was found, because the next person to make
the same mistake will search for the symptom rather than for the fix.

One topic per commit and per pull request. A commit carrying two unrelated
changes has a message describing one of them, and the other one lands unread.

Nothing here judges what a message says. The only route that reads a message at
all is the sign-off check, and it reads one trailer out of it. So this is
carried by review and by nothing else.

## How large a change may be

Small enough that one person can hold the whole of it while reading it. There
is no number in this repository and no check that measures one.

A change that will not fit is usually an issue whose scope was planned wrong
rather than a change that needs an exception, and the first response is to
divide the issue rather than the finished diff. Two pull requests carved out of
one change only make sense together and neither is reviewable alone, which
satisfies the size and defeats the reason for it. Dividing the issue gives each
half its own reason to exist and its own definition of done.

Two practical triggers. Where the change touches more than one part of the tree
as `docs/layout.md` names them, and those parts did not have to move together,
it was two changes. Where the first section of the pull request body needs more
than a sentence to say what the change does, it was two changes.

## What a pull request body carries

`.github/pull_request_template.md` is the shape, and it is not a formality.

Every asserted fact carries the command that produced it, run at the commit
being pushed and against the reference a reader will have rather than against
your working tree. Paste the command and its output. Where a claim cannot be
backed by a command, write it as a claim and say so in the same sentence.
Measured, not measured, and not evaluated on this route are three different
states, and they take three different words.

The last section of the template is the one most often left empty and the one
worth most. A condition on the issue that the change does not meet belongs
there by name rather than left to be noticed. A sentence admitting that
something was not done stays negative through every later edit; turning such a
line into a tick is worse than deleting it.

Where a change introduces a language, a format, a tool, a runtime or a
dependency the tree does not already carry, the body names the means and says
why it fits, including what it costs to add. `docs/toolchain.md` is what that
answer looks like at full length.

## A change that touches something you do not own

Before pushing, look at what the branch actually changes:

    git diff --name-only origin/main...HEAD

A path outside what the issue's `Scope:` line names is a path somebody else may
be editing at the same time, and two changes in one file is the collision that
costs a merge nobody can land. Push the branch, write into the issue which path
it reached and why it had to, and leave the merge to whoever holds that path.

## The checks a change can trip

The workflow files are the authority rather than this list, and a check that
lands after this paragraph was last touched will be in them and not in it. Run
the command rather than trusting the paste:

    git grep -n '^    name:' -- .github/workflows/
    .github/workflows/branch-sweep.yml:74:    name: Sweep the default branch for a non-success run
    .github/workflows/build.yml:47:    name: Build
    .github/workflows/codeql.yml:69:    name: Code scanning (CodeQL)
    .github/workflows/dco.yml:24:    name: DCO sign-off
    .github/workflows/dependency-lock.yml:73:    name: Dependency lock
    .github/workflows/doc-hygiene.yml:60:    name: Documentation formatting, paths and internal links
    .github/workflows/doc-hygiene.yml:249:    name: External links
    .github/workflows/lint.yml:60:    name: Formatting, vet and lint
    .github/workflows/sbom.yml:56:    name: Bill of materials
    .github/workflows/scorecard.yml:49:    name: Scorecard analysis
    .github/workflows/test.yml:69:    name: Unit tests
    .github/workflows/text-hygiene.yml:31:    name: Line endings and encoding
    .github/workflows/unicode-guard.yml:23:    name: Reject Trojan Source Unicode
    .github/workflows/zizmor.yml:40:    name: Audit workflows (zizmor)

One job deliberately carries no name so that its check run takes the job id
instead, which is `dependency-review`, and its own comment says why.

What each one refuses, in the order above. `Sweep the default branch for a
non-success run` is the one that judges no change of yours: it runs daily, asks
which runs on the default branch concluded something other than success, and
files what it finds as an issue on this tracker, because two of the checks
below never report on a pull request and can stay red where nobody looks.
`Build` compiles every package with
dependency resolution switched off. `Code scanning (CodeQL)` reads the Go source
for a security defect and reds on any finding, which `docs/code-scanning.md`
argues. `DCO sign-off` is the section above. `Dependency lock` refuses a module
file that is not what the tree's imports require, and `docs/dependencies.md` is
the inventory behind it.
`Documentation formatting, paths and internal links` refuses a document that is
not wrapped the way this tree wraps, that names a path which does not exist, or
that links to something which is not there. `External links` follows the links
that leave this repository, weekly rather than on a pull request. `Formatting,
vet and lint` is the formatter, the toolchain's correctness analyser and the
linter, in that order. `Bill of materials` produces the inventory of what is in
here and refuses one that has stopped covering an ecosystem the tree carries.
`Scorecard analysis` scores supply chain hygiene and reports to the security
dashboard. `Unit tests` runs the suite on a stock runner with no display and no
elevation, and fails a run that selected no test at all.
`Line endings and encoding` refuses a tracked text file that is not
stored with LF or does not decode as UTF-8. `Reject Trojan Source Unicode`
refuses bidirectional and invisible control characters, which exist to make a
file read differently from how it behaves. `Audit workflows (zizmor)` audits
the workflow files themselves, which are the only thing here that runs with
this repository's permissions. `dependency-review` refuses a newly introduced
dependency carrying a known advisory.

Which of them the branch requires is a different question from which of them
run, and today the answer is none of them. `docs/required-checks.md` is the
request for that and is not a description of the current setting.

## Style

English in artefacts. No attribution to a tool, no generated-by marker, and
nothing naming what produced a change rather than what the change does.

Documents are hard wrapped at 81 columns, carry no trailing whitespace and no
tab, and end with exactly one newline. That is enforced rather than preferred,
and the width is the width this tree already wrapped at on the day the check
landed.

Every tracked text file is stored with LF and decodes as UTF-8. `.gitattributes`
declares it, and a check refuses a file that breaks the declaration rather than
leaving the declaration to be believed.

A path written in backticks in a document has to exist, because a pointer to
something renamed stays invisible until somebody follows it. Name no path you
do not intend to resolve.

## What this guide does not cover

How to run a deployment. There is nothing here to deploy yet, and `deploy/`
holds a placeholder saying which issue brings it.

The architecture. `docs/layout.md` fixes the shape of the tree and which way a
dependency may point, and `layout_test.go` is the guard behind it rather than a
paragraph somebody remembers.

What this project is for, and the specific things it is not built for.
`README.md`, `NOTICE.md` and `docs/scope-of-use.md` hold that, and nothing in
the software enforces it.

How to report a vulnerability. `SECURITY.md` is that route, and it is not the
issue tracker.
