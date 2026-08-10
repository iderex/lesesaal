# What each check has been observed to refuse

Written for issue #101. A gate nobody has watched refuse anything is a
decoration, and the proofs that it does refuse were scattered across whichever
run happened to demonstrate them. This is where they are collected, one entry
per check: what the smallest change was that tripped it, the run where it was
observed to red, and the failure it prevents.

Every trip case here is a plausible mistake rather than an obviously broken
input. A fixture that could not possibly have passed proves less than one that
nearly did, so where a proof could be made to red for the wrong reason, the
entry says which reason it was.

## The list this is measured against

The checks are what the workflows declare, and the list is derived rather than
remembered:

    git grep -n '^    name:' -- .github/workflows/
    .github/workflows/branch-sweep.yml:74:    name: Sweep the default branch for a non-success run
    .github/workflows/build.yml:60:    name: Build
    .github/workflows/codeql.yml:69:    name: Code scanning (CodeQL)
    .github/workflows/dco.yml:24:    name: DCO sign-off
    .github/workflows/dependency-lock.yml:73:    name: Dependency lock
    .github/workflows/doc-hygiene.yml:81:    name: Documentation formatting, paths and internal links
    .github/workflows/doc-hygiene.yml:118:    name: Documentation spelling
    .github/workflows/doc-hygiene.yml:317:    name: External links
    .github/workflows/lint.yml:67:    name: Formatting, vet and lint
    .github/workflows/sbom.yml:56:    name: Bill of materials
    .github/workflows/scorecard.yml:49:    name: Scorecard analysis
    .github/workflows/test.yml:80:    name: Unit tests
    .github/workflows/text-hygiene.yml:35:    name: Line endings and encoding
    .github/workflows/unicode-guard.yml:32:    name: Reject Trojan Source Unicode
    .github/workflows/zizmor.yml:40:    name: Audit workflows (zizmor)

Fifteen names and one further check run. One job carries no name so that its
job id becomes the check-run name, which is `dependency-review`, and
`docs/required-checks.md` is where that is explained. Sixteen entries follow,
in that order.

Every proof lives on a branch that is never merged. Where the check reports only
on a pull request, the branch was opened as one, observed, and closed unmerged.

## Sweep the default branch for a non-success run

Two runs, because this check has two verdicts and they are opposite states.
Both were taken on the default branch rather than on a branch, which is the
only place this check runs at all, so neither could be taken before it merged.

The trip case for the fail-closed leg is a branch name that does not exist,
which is what a dispatch with a typed branch produces. The run list came back
empty, and an empty run list is a branch this sweep did not judge rather than a
branch with nothing wrong:

    https://github.com/iderex/lesesaal/actions/runs/31303480771/job/93219943491

    No workflow run on no-such-branch-here was examined, so nothing was judged. Failing closed rather than reporting a clean branch.

The second is the sweep meeting that failure. The run above is a real
non-success run on the default branch, produced by this check and not excluded
from what it examines, and the next sweep picked it up and filed it:

    https://github.com/iderex/lesesaal/actions/runs/31303515962/job/93220032635

That sweep concluded success, which is the design rather than a leniency. The
report is the issue it filed and not the run's own verdict, because a red
default branch reported by reddening a run on the default branch would be as
invisible as the thing it reports on. What the issue carried was the workflow
name, the conclusion, the trigger and the link to the failed run, under a label
of its own and assigned.

It prevents a workflow concluding non-success on the default branch and nobody
meeting it, which is the position the weekly external link leg and the
scorecard analysis are both in, and it prevents a sweep that examined nothing
being read as a clean branch.

What has not been shown is the schedule. Both runs above were dispatched by
hand, so the daily trigger has fired nothing here yet, and a cron that never
fires would leave every verdict above true and the check useless.

## Build

The trip case is two commits, because there are two mistakes worth separating.
A package that declares a variable and never uses it, which is a compile error
in this language rather than a lint finding. And a file importing a package the
module file does not carry, which with resolution switched off is refused rather
than quietly added.

    https://github.com/iderex/lesesaal/actions/runs/31259299494/job/93107492218
    https://github.com/iderex/lesesaal/actions/runs/31259368254/job/93107670478

It prevents a tree that does not compile reaching the default branch, and it
prevents an import silently adding a requirement nobody reviewed.

## Code scanning (CodeQL)

The trip case is a request's query parameter handed to the file system with no
check, in the shape the volunteer surface and the subject ingest are both going
to be made of. It compiles, it is formatted, and the toolchain's own analyser
and the linter both pass it, so the red verdict is the analyser and nothing
underneath it.

    https://github.com/iderex/lesesaal/actions/runs/31277683195/job/93153841578

It prevents a path leaving the directory this project owns, which is the
boundary `docs/subject-media.md` draws.

## DCO sign-off

The trip case is a commit with no sign-off trailer. Nothing else about it is
unusual, which is the point: it is what `git commit` without the flag produces.

    https://github.com/iderex/lesesaal/actions/runs/31220312809/job/93003212016

It prevents a contribution whose author never asserted the certificate in
`DCO` reaching the default branch.

## Dependency lock

The trip case is a requirement in the module file that nothing in the tree
imports. It is the residue somebody leaves behind after removing the last use
of a dependency.

    https://github.com/iderex/lesesaal/actions/runs/31278149194/job/93154982494

The other direction was taken on the same branch one commit later, a dependency
imported and used with no lock entry covering it, and there it is `Build` that
refuses with a missing lock entry rather than this check.

It prevents the module file and the tree's imports drifting apart in either
direction.

## Documentation formatting, paths and internal links

The trip case is one document carrying every mistake the three legs judge: a
line over the width this tree wraps at, trailing whitespace, a tab, a backticked
path that does not exist, and a link to a file that is not there.

    https://github.com/iderex/lesesaal/actions/runs/31246610116/job/93076164634

It prevents a document that reflows an unrelated paragraph on every edit, and a
pointer to something renamed, which is how a plan quietly stops being a plan.

## Documentation spelling

The trip case is one letter dropped from one word, in prose, in a document
nobody would look at twice: `deliberately` written `deliberatly` in
`docs/harness.md`. It is the mistake a person actually makes rather than a
string no speller could accept, and it is in a sentence rather than in a code
span, because the leg strips code before it judges a word and a proof placed
there would have proved the stripping instead.

    https://github.com/iderex/lesesaal/actions/runs/31334351163/job/93297633867

    FAIL  docs/harness.md:115 carries "deliberatly", which is in no dictionary this project carries
    Examined 36282 word(s) of prose in 35 tracked document(s), against a project dictionary of 62 entry(s).

Every other check on that commit was green, so the red verdict is this leg and
nothing underneath it:

    gh api repos/iderex/lesesaal/commits/94a1bd6/check-runs --jq '[.check_runs[] | select(.conclusion=="failure")] | length'
    1

It prevents a typo reaching an operator's guide, and it holds the terms
`docs/vocabulary.md` fixes to one spelling wherever they appear, which is what
stops a word this project decided on from arriving in three forms.

The bound on it is in the workflow and belongs here too. It judges words rather
than sentences, so a document can pass this leg and still be wrong: grammar,
style and whether a sentence is true are none of its business. And what the
project dictionary carries is curated rather than derived from
`docs/vocabulary.md`, so a term that document fixes is spelled consistently only
while somebody keeps the dictionary level with it.

## External links

The trip case is a link to a host in the reserved `.invalid` domain, so the
proof contacts nothing outside this repository to make its point.

    https://github.com/iderex/lesesaal/actions/runs/31246822100/job/93076705953

It prevents a document promising something at the other end of a link that is
no longer there. It runs weekly rather than on a pull request, so it reports
rather than blocks.

## Formatting, vet and lint

The trip case is one file that compiles on purpose and trips each of the three
legs once: a space inside the parameter list and spaces for indentation, a
format verb given the wrong type, and a comparison against a boolean constant
in a function nothing calls.

    https://github.com/iderex/lesesaal/actions/runs/31264989341/job/93121582520

It prevents a diff whose real change is hidden inside a reformatting, and the
class of defect the analyser and the linter each name.

## Bill of materials

The trip case is one character in the generator's cataloger selection, a plus
turned into a minus, so the document stops covering an ecosystem the tree
carries while still being produced and still looking complete.

    https://github.com/iderex/lesesaal/actions/runs/31265689351/job/93123312885

The same check refused twice more while the proofs below were being taken, both
times for a mistake nobody was aiming at it, and both refusals were correct. A
branch adding a workflow was told the document covers ten actions and the
workflows pin eleven. A branch adding a requirement was told the document covers
two modules and the module graph holds three.

    https://github.com/iderex/lesesaal/actions/runs/31290254987/job/93186080324
    https://github.com/iderex/lesesaal/actions/runs/31290319529/job/93186240872

It prevents an inventory that answers an advisory question with a confident
wrong answer.

## Scorecard analysis

No demonstrated refusal, and this entry is here rather than omitted because of
it.

It runs on a schedule, on a push to the default branch and on a change to the
branch protection rule, and never on a pull request, so no branch can be made
to red it the way every entry above was. Its verdict is also a score computed
by a service outside this repository against that repository's state, so the
smallest change that would move it is not a change to a file. Every run it has
had here has concluded success:

    gh run list --repo iderex/lesesaal --workflow scorecard.yml --limit 20 --json conclusion --jq '[.[] | .conclusion] | unique | join(" ")'
    success

What that leaves is a check whose behaviour on a bad tree is not known here. It
is advisory in `docs/required-checks.md` for the same reason it cannot be
proven, so nothing is gated on the gap, and it is the one entry in this record
with no evidence behind it.

## Unit tests

Three trip cases, because this check has three legs and each refuses a different
thing. All three are on one branch, one commit each, and every other check that
ran on those three commits was green, so each red verdict is the leg under test
and nothing underneath it.

The first is a suite moved behind a build tag, which is what somebody writes
when they mean to take a slow suite out of the ordinary run and then forgets to
run it anywhere else. Every package still compiles and the toolchain reports a
package with no test file as a success, so the check would have been green over
nothing:

    https://github.com/iderex/lesesaal/actions/runs/31296187043/job/93201545461

The leg reported `Selected 0 test(s) across 7 package(s)` and failed there.

The second is the core given an import of the surface, which the layout guard
refuses. This is the leg that runs the suite, and the run summary carries the
number rather than only the verdict: nineteen tests selected, seventeen
executed, sixteen passed and one failed, the gap being the package that stopped
at its first failing test.

    https://github.com/iderex/lesesaal/actions/runs/31296264233/job/93201734635

The third is a unit test that fetches a subject manifest over HTTP, which is
what somebody writes when they want the ingest tested against a real manifest
rather than a fixture. It compiles, it is formatted, and it skips itself when
the host does not answer, so nothing at run time objects to it. What objects is
`harness_test.go`, which reads the source and refuses a call into the standard
library's HTTP client written outside `internal/system`, and it names the file,
the line, the call and the field to take instead.

    https://github.com/iderex/lesesaal/actions/runs/31296348917/job/93201936767

It prevents a green tick over a suite that was never selected, a failing test
reaching the default branch, and a unit test reaching the network.

The third of those is bounded and the bound belongs here rather than only in the
workflow. The refusal is over the call that is WRITTEN, so a connection opened
through a helper this tree does not own, or by a dependency, is invisible to it,
and the runner has a working network throughout. A suite that reaches the
network without writing the call reaches it.

## Line endings and encoding

The trip case is two files: one stored with CRLF, and one whose bytes are
latin-1 rather than UTF-8. Both are what a file written on another machine or
saved by another editor looks like, and neither is visible in a diff.

    https://github.com/iderex/lesesaal/actions/runs/31222001755/job/93008378502

It prevents the bytes in the repository depending on the machine they were
written on.

## Reject Trojan Source Unicode

The trip case is a document carrying a zero-width space in the middle of a word,
pasted in from a rendered page the way a real one arrives, and a right-to-left
override in ordinary prose. Neither is visible to a reader and neither shows in
a diff. The file is wrapped the way this tree wraps and decodes as UTF-8 with
LF endings, so every other check stayed green on that branch and the single red
verdict is the guard under test.

    https://github.com/iderex/lesesaal/actions/runs/31290210484/job/93185968140

It prevents source that renders differently from how it executes, which is the
attack the check is named for.

## Audit workflows (zizmor)

The trip case is a new workflow whose action is referenced by its release tag
instead of by a commit sha. Everything else the audit asks for is satisfied
deliberately, so the finding is the pin alone: the trigger is manual, the
workflow-level permission set is empty, the job takes only read access to the
contents, and the checkout does not persist the token.

    https://github.com/iderex/lesesaal/actions/runs/31290254992/job/93186080373

It prevents a workflow running whatever the action's owner repoints a tag at,
with this repository's permissions.

This proves one finding class of that check, an unpinned use. It says nothing
about the template injection, excessive permission and dangerous trigger classes
the same check refuses, and those have not been shown to bite here.

## dependency-review

The trip case is a widely used module added at the version somebody already had
in another project, one patch release below the one carrying the fix. It is
locked with its hash, imported, used, and the identifier is exported so the
linter has nothing to say, so the build, the lock check and the linter all
stayed green and the advisory database is the only thing that objected.

    https://github.com/iderex/lesesaal/actions/runs/31290319539/job/93186240914

It prevents a dependency with a known advisory entering the graph.

Until that branch, this check had never read a manifest at all: the module on
the default branch carries no requirement, so every green verdict it had
produced was a verdict about nothing.

## What was proven with no check behind it, and is not any more

The deterministic harness rules have five near misses, one per rule, in the
proof taken for #21. Every check on that branch was green:

    gh api repos/iderex/lesesaal/commits/a4dab05/check-runs --jq '[.check_runs[] | select(.conclusion=="failure")] | length'
    0

That was correct at the time and is not any more. The rules are still guarded by
a test rather than by a workflow leg, but a workflow now runs the tests:

    git grep -n 'go run \. ci' -- .github/workflows/test.yml
    .github/workflows/test.yml:12:# `go run . ci <leg>`, which is the same procedure a contributor runs before
    .github/workflows/test.yml:134:        run: go run . ci test-selection
    .github/workflows/test.yml:138:        run: go run . ci test

That command is not the one this paragraph was first written with. It quoted
two `go test` lines out of the same workflow, and #150 replaced them with the
one verb above, so the old command now matches nothing and a reader running it
would have been left to guess whether the claim or the evidence had died.

So the source guards that landed with the layout and the harness are executed by
`Unit tests`, and the third trip case in that entry is one of them reddening a
check rather than a terminal. What has not been re-taken is the five near misses
themselves: they were observed by somebody running the suite, on a branch whose
checks were all green because no check ran a test, and this record does not
claim they have since been watched by one.

## Nothing refuses a check that arrives without an entry here

The last thing #101 asks for is a mechanism that reds when a check is added and
no row is added with it. There is none, and this paragraph is the record of its
absence rather than a plan to add one.

Nothing in this tree compares a document's contents against the workflows. The
documentation legs judge a document's shape, the paths it names and the links it
carries, and none of them reads what a table says about a check-run name. The
class of check that could do it is the greppable invariant, which is #94, and a
rule that this file names every name the workflows declare is a candidate for
it. Until then this record is kept level by the person adding the check, which
is exactly the kind of rule this project marks as carried by nobody.

The one thing that is derived rather than remembered is the list at the top,
which is pasted under the command that produces it, so a reader can run it and
see the two disagree.
