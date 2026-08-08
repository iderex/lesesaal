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
    .github/workflows/build.yml:47:    name: Build
    .github/workflows/dco.yml:24:    name: DCO sign-off
    .github/workflows/doc-hygiene.yml:60:    name: Documentation formatting, paths and internal links
    .github/workflows/doc-hygiene.yml:249:    name: External links
    .github/workflows/scorecard.yml:49:    name: Scorecard analysis
    .github/workflows/scorecard.yml:85:          name: SARIF file
    .github/workflows/text-hygiene.yml:31:    name: Line endings and encoding
    .github/workflows/unicode-guard.yml:23:    name: Reject Trojan Source Unicode
    .github/workflows/zizmor.yml:40:    name: Audit workflows (zizmor)

One job deliberately carries no name, and its own comment says why:
`.github/workflows/dependency-review.yml` leaves the job id to become the
check-run name, which is `dependency-review`.

The last line of that grep is not a check run at all. `SARIF file` is the name of
an upload step inside the scorecard workflow, and it is listed here so that
somebody comparing the grep against the list below does not go looking for it.

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
them freezes nothing. `Build` is the exception and only because it did not exist
when that output was taken: its verdict is recorded on the pull request that
landed the build check, and it is quoted there rather than restated here.

## The blocking list, in the order to add it

Each name is exactly as it appears above.

`Build`. It compiles every package in the tree with the toolchain the module
declares and with dependency resolution switched off, so a change that does not
compile, or that imports something the module file does not carry, cannot reach
the default branch. It is first in the list because every later check is
measured against a tree that builds.

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
added to it by the change that landed the build check, which is what this
section asks of every later one. The unit test check, the lint and format check,
the bill of materials and the code scanning analysis are still to be written, in
#20 through #25, and each adds its check-run name here in the change that lands
it. The quality milestone is where the list is compared against the reference
gate, which is #89's work rather than this document's.

Until the ruleset carries the seven names above, #31 stays open. What closes it
is the ruleset showing them, not this document existing.

## One thing found while writing this

`.github/workflows/dependency-review.yml` says in a comment that the required
status check belongs to a ruleset called "Protect main". This repository's
ruleset is called `gate`, as the first command above shows. The name of a
ruleset changes nothing about how a check is matched, so this is a stale comment
rather than a defect in the gate, and it is recorded here because a reader
following that name would look for a ruleset that does not exist.
