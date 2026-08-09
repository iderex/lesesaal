# The checks to require on the protected branch

Requested in issue #31. This document is the request. It changes no setting,
because the ruleset is the maintainer's to edit and nothing here can edit it.

## What the branch refuses today

    gh api repos/iderex/lesesaal/rulesets --template '{{range .}}{{.name}} {{.target}} {{.enforcement}}
    {{end}}'
    gate branch active

    gh api repos/iderex/lesesaal/rulesets/20520725 --template '{{.enforcement}} bypass={{len .bypass_actors}} types={{range .rules}}{{.type}} {{end}}'
    active bypass=0 types=deletion non_fast_forward pull_request

So a change reaches `main` only through a pull request, the branch cannot be
deleted and cannot be force pushed, and no actor may bypass any of that. What is
missing is the fourth rule: no status check is required, so a pull request with
every workflow red merges exactly like one with every workflow green.

## The names, quoted from the workflows that produce them

A required check is matched by its literal check-run name, so the list has to be
quoted rather than described.

    git grep -n '^  *name:' -- .github/workflows/
    .github/workflows/branch-sweep.yml:74:    name: Sweep the default branch for a non-success run
    .github/workflows/build.yml:47:    name: Build
    .github/workflows/codeql.yml:69:    name: Code scanning (CodeQL)
    .github/workflows/dco.yml:24:    name: DCO sign-off
    .github/workflows/dependency-lock.yml:73:    name: Dependency lock
    .github/workflows/doc-hygiene.yml:60:    name: Documentation formatting, paths and internal links
    .github/workflows/doc-hygiene.yml:249:    name: External links
    .github/workflows/lint.yml:60:    name: Formatting, vet and lint
    .github/workflows/sbom.yml:56:    name: Bill of materials
    .github/workflows/sbom.yml:281:          name: sbom-spdx-json
    .github/workflows/scorecard.yml:49:    name: Scorecard analysis
    .github/workflows/scorecard.yml:85:          name: SARIF file
    .github/workflows/test.yml:69:    name: Unit tests
    .github/workflows/text-hygiene.yml:31:    name: Line endings and encoding
    .github/workflows/unicode-guard.yml:23:    name: Reject Trojan Source Unicode
    .github/workflows/zizmor.yml:40:    name: Audit workflows (zizmor)

That output is longer than it was when this document was written, and it is
pasted again rather than added to, because six checks landed between the two
runs and a block that carried only the newest name would disagree with the
command above it.

One job deliberately carries no name, and its own comment says why:
`.github/workflows/dependency-review.yml` leaves the job id to become the
check-run name, which is `dependency-review`.

Two lines of that grep are not check runs at all. `SARIF file` and
`sbom-spdx-json` name upload steps, inside the scorecard workflow and the bill
of materials workflow, and they are listed here so that somebody comparing the
grep against the list below does not go looking for them.

Four names in the grep are check runs and are not in the blocking list below:
`Code scanning (CodeQL)`, `Dependency lock`, `Formatting, vet and lint` and
`Bill of materials`. All four exist and all four report on a pull request. Each
landed without being added here, which is what the last section of this document
asks each change to do, so the omission is recorded rather than repaired in
passing. Adding a name to the blocking list is an argument about what the branch
should refuse, and that argument belongs to whoever makes it.

## The names as they actually appear on a pull request

The grep above says what the workflows declare. This says what the platform
reported on a real pull request, which is what the ruleset will match:

    gh pr checks 143 --json name,state --template '{{range .}}{{.name}} | {{.state}}
    {{end}}'
    zizmor | SUCCESS
    DCO sign-off | SUCCESS
    Audit workflows (zizmor) | SUCCESS
    dependency-review | SUCCESS
    Documentation formatting, paths and internal links | SUCCESS
    Line endings and encoding | SUCCESS
    Reject Trojan Source Unicode | SUCCESS
    External links | SKIPPED

That output continues past what is pasted: each of the three workflows that run
on a push as well as on a pull request reports twice, and the repeated lines are
cut rather than shown. Nothing else follows them.

Every name in the blocking list below appears there and was green, so requiring
them freezes nothing. Two are exceptions and only because neither existed when
that output was taken. `Build`'s verdict is recorded on the pull request that
landed the build check, and `Unit tests`'s on the one that landed the unit test
check, and both are quoted there rather than restated here.

## The blocking list, in the order to add it

Each name is exactly as it appears above.

`Build`. It compiles every package in the tree with the toolchain the module
declares and with dependency resolution switched off, so a change that does not
compile, or that imports something the module file does not carry, cannot reach
the default branch. It is first in the list because every later check is
measured against a tree that builds.

`Unit tests`. It runs the whole unit suite unprivileged and with no display, and
it refuses a run that selected no test at all. That last leg is why it is second:
without it the toolchain reports a package with no test file as a success, so a
change that moves the suite where the job does not look leaves every later green
meaning nothing. What it does not refuse is a socket opened at run time, and the
workflow that produces this name says so at the top of the file.

`DCO sign-off`. It is the only thing standing between an unsigned commit and the
default branch, and this project cannot accept a contribution whose author never
asserted the certificate. It fails closed and passes on a clean tree.

`Line endings and encoding`. It refuses a tracked text file stored with CRLF or
mixed endings and one that does not decode as UTF-8. Both are the kind of defect
that is invisible in a diff and expensive afterwards.

`Reject Trojan Source Unicode`. It refuses bidirectional and invisible control
characters, which exist to make a file read differently from how it behaves. A
guard against deception that a merge can ignore is not a guard.

`Documentation formatting, paths and internal links`. Everything this repository
holds today is documents, so this is the check that judges the whole of it. Its
path and link legs are the ones that catch a plan quietly ceasing to be a plan.

`Audit workflows (zizmor)`. The workflows are the only thing here that runs with
this repository's permissions, and this is what refuses an unpinned action, an
excessive permission or a template injection in them.

`dependency-review`. It refuses a newly introduced dependency carrying a known
vulnerability at any severity. There is no manifest in the tree yet, so it
passes on nothing today, and requiring it now means the guard is in place before
the first dependency rather than after it.

## What is advisory, and why each one cannot be required

`External links`. It runs on a schedule and on request and never on a pull
request, which its own condition states and the SKIPPED verdict above confirms.
A required check that never reports leaves every pull request waiting forever,
so requiring it would freeze the branch rather than guard it. It is advisory by
construction, not by preference.

`Scorecard analysis`. Its triggers are the branch protection rule, a schedule
and a push to the default branch, so it does not report on a pull request
either. It also scores against an external service, which makes somebody else's
availability a condition of merging here.

`zizmor`, the bare name in the pull request list. That check run comes from the
code scanning upload rather than from the job, and its name follows the tool and
the category in the uploaded results rather than anything written in the
workflow file. Requiring `Audit workflows (zizmor)` pins the same finding set to
a name this repository controls.

`Sweep the default branch for a non-success run`. It runs on a schedule and on
request and never on a pull request, so it is advisory for the same reason
`External links` is. It is also the one check here whose subject is the other
checks: it exists because the two above never report on a pull request and can
therefore stay red on the default branch unmet, which is what
`.github/workflows/branch-sweep.yml` sets out. Its own report is an issue on
this tracker rather than its verdict, so a red sweep and a sweep that filed
something are different states and neither is a thing to gate a merge on.

## No bypass actor is wanted

The ruleset has none today, and the request is that it keeps none.

A bypass actor makes every rule above advisory for whoever holds it, and the
holder is normally the person most likely to be merging at speed. The rules here
are cheap to satisfy and there is no case where a merge has to happen with a red
gate, so the exception would exist only to be used on the day it should not be.

Where a check turns out to be wrong, the repair is to fix the check or to move
it to the advisory list in this document, both of which leave a trace. A bypass
leaves none.

## This request is repeated rather than made once

The list is bounded by what exists at the moment it is asked for. `Build` was
added to it by the change that landed the build check, and `Unit tests` by the
change that landed the unit test check, which is what this section asks of every
later one. Four checks that already exist did not, and the section above says
which. The quality milestone is where the list is compared against the reference
gate, which is #89's work rather than this document's.

Until the ruleset carries the eight names above, #31 stays open. What closes it
is the ruleset showing them, not this document existing.

## One thing found while writing this

`.github/workflows/dependency-review.yml` says in a comment that the required
status check belongs to a ruleset called "Protect main". This repository's
ruleset is called `gate`, as the first command above shows. The name of a
ruleset changes nothing about how a check is matched, so this is a stale comment
rather than a defect in the gate, and it is recorded here because a reader
following that name would look for a ruleset that does not exist.
