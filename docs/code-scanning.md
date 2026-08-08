# Code scanning, and the query set it runs

Added in issue #25. This is the first check that reads the code of this project
for a security defect. It exists before the code it will judge, which is the
cheap moment to add it: a scanner switched on over a finished codebase starts
with a backlog nobody triages, and a scanner switched on over an empty one
starts at zero and stays there or fails.

The check is `.github/workflows/codeql.yml` and its check-run name is
`Code scanning (CodeQL)`.

## The query set

`security-extended`.

The default set is what the analyser runs when nothing is asked of it. It is
tuned for precision, which means it holds back queries that find real defects at
the cost of also reporting things that are not. `security-extended` is the same
set plus those, and the trade is more findings to read against fewer missed.

The reason to take that trade here is the shape of what this project will hold.
A volunteer surface takes bytes from people the operator has never met, a
subject ingest takes image files from wherever an owner got them, and an export
writes files somebody else parses. Those are the surfaces the extended queries
are about: path handling, request-derived data reaching a file or a command, and
data crossing a trust boundary. A precision-tuned set is the right default for a
codebase where a false positive costs a developer an afternoon; it is the wrong
one where a missed finding costs a volunteer's data.

The cost is stated rather than assumed away. The extended set reports more, so
this check will need triage that the default set would not have asked for, and
the section on dismissals below is what stops that triage from becoming a habit
of clicking things away.

## What was not chosen

`security-and-quality` is `security-extended` plus maintainability and
reliability queries. It was not chosen because those overlap what this tree
already judges twice: the toolchain's own correctness analyser and the linter
both run in `Formatting, vet and lint`, they are configured in
`staticcheck.conf`, and a third opinion on an unused variable is noise rather
than a lens. Where a quality query turns out to catch something the linter
cannot, adding it by name is cheaper than moving the whole set.

A severity threshold was not chosen either. The leg that fails the build fails
on any result the query set reports, rather than on anything above a line. This
tree has no history of findings to weigh a threshold against, and a number
picked before the first finding would be a guess presented as a policy. The
first time a finding here is genuinely not worth failing on, that is the
evidence, and this document is where the threshold gets written down with it.

The threat model is the analyser's default, which treats a remote request as
untrusted and local input as trusted. Widening it is a decision this project
cannot make yet, because whether a deployment faces the open internet is entry 2
of #1 and is unanswered.

## What is not analysed here

The workflow files. `Audit workflows (zizmor)` reads them against what a
workflow is permitted to do, and a second analyser over the same files would
produce two verdicts on one question. The languages input in the workflow names
Go and nothing else.

Anything that is not Go. There is no browser code and no template in this tree
yet, so the volunteer surface in #43 arrives after this check rather than
before it.

An unchanged tree, after the fact. The check runs on a pull request and on the
default branch and on no schedule, so a query that is added upstream next month
does not re-read code that stopped changing. That is a real gap. Closing it
costs a scheduled run, and the argument belongs with #90, which adds the second
analyser and can schedule both together rather than separately.

This is the first of two. #90 is the second, with a different lens, and two
analysers that agree by construction are one analyser.

## Dismissals

A finding that is not a real problem is dismissed with a reason, in the
dismissal comment on the alert itself. That is where a reason travels with the
thing it is about; a list of dismissals kept in a document drifts against the
dashboard that decides them, and the document is always the one that is wrong.

What is dismissed today, and the command that answers it at any later commit:

    gh api "repos/iderex/lesesaal/code-scanning/alerts?state=dismissed&per_page=100" \
      --jq '.[] | "\(.number) \(.rule.id) \(.dismissed_reason) \(.dismissed_comment)"'

    gh api "repos/iderex/lesesaal/code-scanning/alerts?state=dismissed&per_page=100" --jq 'length'
    0

Nothing refuses a dismissal that carries no reason. The dismissal comment is
optional at the interface that takes it, and a check that could refuse an empty
one would have to read the dashboard, which nothing in this repository does.
That is the residual on this rule and it is left visible rather than written as
though the rule enforces itself.

## What this does not settle

Whether the branch should refuse a merge on this check. That argument belongs to
`docs/required-checks.md`, which is a request to the maintainer rather than a
setting, and this change does not add the name to it.
