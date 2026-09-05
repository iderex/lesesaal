# Governance

Written for issue #111. It says who decides, how a decision is recorded, how
somebody proposes a change, and what happens to a proposal nobody has time for.
It is short because the answers are small, and a document implying a committee
that does not exist would be worse than none.

## Who decides

Me, Nils Lehnen, the only person who has committed to this repository and the
only person who has opened anything on it:

    $ gh api repos/iderex/lesesaal/contributors --jq '[.[].login] | join(" ")'
    iderex
    $ gh api --paginate "repos/iderex/lesesaal/issues?state=all&per_page=100" --jq '.[].user.login' | sort -u
    iderex
    $ gh api --paginate "repos/iderex/lesesaal/issues?state=all&per_page=100" --jq '.[].user.login' | wc -l
    159

One person is a legitimate answer and it is this project's answer today. What
it costs is worth stating rather than leaving to be discovered. There is no
second person who can merge, so work stops when I stop. There is no appeal from
a decision to anybody but the person who made it. And a disagreement about
direction is settled by one reading of `docs/design-point.md` and
`docs/scope-of-use.md` rather than by a vote.

That changes when somebody else is doing enough of the work that the sentence
above stops being true, and the change is itself a decision recorded the way
the next section describes.

## How a decision is recorded

On the tracker, before the code that depends on it exists. A decision is an
issue that says what is being decided, what the options cost, and what was
chosen. The milestone that holds nothing else carries sixteen of them:

    $ gh issue list --repo iderex/lesesaal --state all --limit 200 --json milestone --jq '[.[] | select(.milestone.title != null) | select(.milestone.title | startswith("M0"))] | length'
    16

A decision that shapes how the software is built then becomes a document under
`docs/`, present tense, one subject per file, and the issue points at it. That
is why `docs/consensus.md` and `docs/federation.md` read as statements rather
than as arguments: the argument is on the issue that produced them.

Decisions I keep for myself are parked in one place rather than assumed in the
issues they block, and issue #1 is that place. An entry there is
answered by a comment naming the option chosen. An issue blocked on one names
the entry it waits for and stops.

## How somebody proposes a change

Open an issue first. `CONTRIBUTING.md` says what it has to contain, and the
templates under `.github/ISSUE_TEMPLATE/` ask for that shape directly. A change
with no issue has nowhere to record why it exists.

Then a pull request against `main`, with the evidence the body asks for. Direct
pushes are refused by the branch ruleset, for me as well:

    $ gh api repos/iderex/lesesaal/rulesets/20520725 --template '{{.enforcement}} bypass={{len .bypass_actors}} types={{range .rules}}{{.type}} {{end}}'
    active bypass=0 types=deletion non_fast_forward pull_request

Disagreement with a rule is usually disagreement with the position the rule
comes from, and it is cheaper to argue with the position. `docs/scope-of-use.md`
and `docs/design-point.md` are where those live.

## What happens to a proposal nobody has time for

It stays open, and it says what it is waiting for.

Nothing here closes an issue for being old. There is no bot and no schedule
that touches the tracker, which is a fact about the workflows rather than an
intention:

    $ git ls-files .github/workflows
    .github/workflows/build.yml
    .github/workflows/codeql.yml
    .github/workflows/dco.yml
    .github/workflows/dependency-lock.yml
    .github/workflows/dependency-review.yml
    .github/workflows/doc-hygiene.yml
    .github/workflows/lint.yml
    .github/workflows/sbom.yml
    .github/workflows/scorecard.yml
    .github/workflows/text-hygiene.yml
    .github/workflows/unicode-guard.yml
    .github/workflows/zizmor.yml

An open issue with no activity means nobody has done it. It does not mean it
was refused, and a refusal is written into the issue and the issue is closed,
so the two are distinguishable. An issue that cannot move because a decision
has not been taken says which decision.

There is no promised response time for an issue or a pull request. The one
place this project does promise times is a security report, in `SECURITY.md`,
and that section says plainly that nothing enforces them either.

## What this document does not cover

Conduct. `CODE_OF_CONDUCT.md` is that, including who receives a report.

Anything about a deployment somebody else runs. This governs the repository.

Enforcement. No check reads this file, nothing refuses a change that departs
from it, and it is carried by my following it. The three rules this
project measures itself by are in `CONTRIBUTING.md`, and where one of them has
no mechanism behind it, the document saying so is the whole of the guarantee.
