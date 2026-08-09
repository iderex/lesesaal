# Parity with the reference gate

Written for issue #89. The gate this repository is measured against runs on the
single sign-on plugin at github.com/Flowfin/jellyfin-plugin-sso. It is public,
its ruleset is readable, and it is the target because it is a standard already
in force somewhere rather than one invented here.

Parity is not copying. That gate guards a plugin loaded into somebody else's
process, written in another language, shipped through a package manifest a
media server reads. This is a self-hosted web application that holds data about
volunteers, decodes images somebody else supplied, and produces a scientific
artefact. Some of the reference applies unchanged, some applies in a different
form, some does not apply at all, and some of what this project needs has no
counterpart there. Each of those is a deviation and each owes its line.

Three verdicts are used and they are used strictly. Adopt is the same check
under the same name. Adapt is the same purpose under a different mechanism, a
different subject or a different name. Drop is that this repository will not
have it, with the reason it does not apply.

## What the two gates require today

    gh api repos/Flowfin/jellyfin-plugin-sso/rulesets --jq '.[] | select(.name=="Protect main and 5.0") | .id'
    18802863

    gh api repos/Flowfin/jellyfin-plugin-sso/rulesets/18802863 --jq '{enforcement, bypass: .bypass_actors, required: [.rules[] | select(.type=="required_status_checks") | .parameters.required_status_checks[].context]}'
    {"bypass":[],"enforcement":"active","required":["build","ABI floor build","Package (JPRM) / Build package","Package (JPRM) / Generate SBOM","CodeQL","Analyze (csharp)","DCO sign-off","Deterministic PR-hygiene checks","Enforce greppable invariants","Reject Trojan Source Unicode","Audit workflows (zizmor)","prettier","dependency-review"]}

    gh api repos/iderex/lesesaal/rulesets/20520725 --jq '{enforcement, bypass: .bypass_actors, types: [.rules[].type]}'
    {"bypass":[],"enforcement":"active","types":["deletion","non_fast_forward","pull_request"]}

Thirteen against none. The distance is the whole list, because this branch
requires no status check at all, and that gap is #31's rather than this
document's: the checks exist here and the ruleset does not name them.

What the reference runs beyond what it requires:

    gh api repos/Flowfin/jellyfin-plugin-sso/contents/.github/workflows --jq '[.[].name] | join(" ")'
    build.yml codeql.yml dco.yml dependency-review.yml dotnet.yml e2e-login.yml fuzz.yml manifest-freshness.yml nightly-betas.yml opengrep.yml pr-hygiene.yml prettier.yml publish-beta.yml publish-failure-alert.yml publish-jf12-beta.yml publish-jf12-stable.yml publish.yml regenerate-manifest.yml scorecard.yml stryker-mutation.yml unicode-guard.yml wiki-lint.yml zizmor.yml

And the coverage bar it pins to the code that decides a security outcome, which
is not a required check but a step inside its build:

    gh api repos/Flowfin/jellyfin-plugin-sso/contents/scripts/check-coverage.py --jq .content | base64 -d | grep -n 'SECURITY_LINE_BAR ='
    68:SECURITY_LINE_BAR = 92.0

What this repository runs, quoted from the workflows that produce the names:

    git grep -n '^    name:' -- .github/workflows/
    .github/workflows/build.yml:47:    name: Build
    .github/workflows/codeql.yml:69:    name: Code scanning (CodeQL)
    .github/workflows/dco.yml:24:    name: DCO sign-off
    .github/workflows/dependency-lock.yml:73:    name: Dependency lock
    .github/workflows/doc-hygiene.yml:60:    name: Documentation formatting, paths and internal links
    .github/workflows/doc-hygiene.yml:249:    name: External links
    .github/workflows/lint.yml:60:    name: Formatting, vet and lint
    .github/workflows/sbom.yml:56:    name: Bill of materials
    .github/workflows/scorecard.yml:49:    name: Scorecard analysis
    .github/workflows/text-hygiene.yml:31:    name: Line endings and encoding
    .github/workflows/unicode-guard.yml:23:    name: Reject Trojan Source Unicode
    .github/workflows/zizmor.yml:40:    name: Audit workflows (zizmor)

One job here carries no name so that its check run takes the job id, which is
`dependency-review`, and `docs/required-checks.md` is where that is explained.

## The thirteen the reference requires

`build`. Adopt, as `Build` in `.github/workflows/build.yml`, delivered by #19.
The reference compiles and tests in one job; here the test half is a separate
check because a compile failure and a failing test are different verdicts and
#20 is where the second one is added.

`ABI floor build`. Drop. It asserts that the plugin still loads into the oldest
host version it claims to support, and nothing loads this into a host process:
the deployment is the whole program, so there is no binary interface to hold
still.

`Package (JPRM) / Build package`. Adapt, to the container image, #74. The
reference packages for a plugin installer; the artefact an operator here
receives is an image and a composition, so the packaging check is the image
build rather than a manifest package.

`Package (JPRM) / Generate SBOM`. Adopt in substance, as `Bill of materials` in
`.github/workflows/sbom.yml`, delivered by #24, with the release half in #120.
The reference generates it during packaging; here it runs on every build and
refuses one that has stopped covering an ecosystem the tree carries.

`CodeQL` and `Analyze (csharp)`. Adapt, both onto one name.
`.github/workflows/codeql.yml` produces `Code scanning (CodeQL)` and #25
delivered it. The reference requires the job's name and the code scanning
upload's own name together. Only the job's name is a repository's to keep
stable, which `docs/required-checks.md` argues for the same duality in the
workflow audit, so this repository requires one name and not two.

`DCO sign-off`. Adopt, same name, in `.github/workflows/dco.yml`, with the
certificate delivered by #26. Nothing about the reason changes between a plugin
and a web application.

`Deterministic PR-hygiene checks`. Adapt, #95. The class is the same, facts
about the pull request that a machine can decide without opinion, and the rules
differ because the artefacts differ: a change to the deployment composition or
to the export format has no counterpart at the reference.

`Enforce greppable invariants`. Adapt, #94, with the analyser that carries them
in #90. The reference's invariants are about its own tree. The ones here are
this project's rules that happen to be text facts, and a personal field
reaching a log call is the one with no counterpart there at all.

`Reject Trojan Source Unicode`. Adopt, same name, already in
`.github/workflows/unicode-guard.yml`. It came with the inherited automation
rather than from an issue in this plan, and what is still owed is #31, which
asks for it on the protected branch.

`Audit workflows (zizmor)`. Adopt, same name, already in
`.github/workflows/zizmor.yml`, and owed to the branch by #31. The workflows
are the only thing in either repository that runs with the repository's own
permissions, so the argument transfers unchanged.

`prettier`. Adapt, and it splits in two. The formatter for the language this
project chose is inside `Formatting, vet and lint` in
`.github/workflows/lint.yml`, delivered by #22. The reference formats its
documents with the same tool; here the documents are judged by
`Documentation formatting, paths and internal links`, delivered by #96, which
also checks that a path a document names exists.

`dependency-review`. Adopt, same name, already in
`.github/workflows/dependency-review.yml`, owed to the branch by #31. It passes
on nothing today because the module carries no requirement, which is a reason
to require it now rather than later.

## What the reference runs without requiring it

`Fuzz (SharpFuzz)`, on a schedule. Adapt, #93. The reference fuzzes an
authentication surface; the surfaces here are the classification boundary, the
image ingest and the manifest parser, and the image decoder is the one where
the honest target is containment rather than correctness.

`Stryker mutation testing`, on a schedule. Adapt, #92, reporting rather than
gating in both places. The subject changes to the code that decides a label.

The coverage bar. Adapt, #91. The number and the shape carry over, a high bar
pinned to the code that decides an outcome and the rest left ungated, and the
subject changes from authentication decision code to the consensus rule, the
retirement rule, the selection rule, the classification boundary and the
export. An error in any of those produces wrong data that looks right, which is
this project's equivalent of an authentication bypass.

`E2E Login Harness`. Adapt, and it splits in two. The browser half is #51, the
deployment half is #80, and both are named for what they require rather than
called integration tests, because a harness whose requirement is not in its
name is how a suite quietly stops running on a stock machine.

`Scorecard analysis`. Adopt, same name, already in
`.github/workflows/scorecard.yml`. It is advisory in both places for the same
two reasons, that it never reports on a pull request and that it scores against
an external service, and `docs/required-checks.md` records that.

`Repo Invariant Lint (Opengrep)`. Counted above as the greppable invariants.

`Wiki Lint`. Adapt, #96. The reference's documentation lives in a wiki outside
its tree, so linting it needs a separate route; here every document is tracked
and the same leg that judges the source tree judges them.

`Manifest freshness`. Drop. It asserts that a published plugin manifest lists
the newest release per host generation, and this project publishes no manifest
for a host to read: an operator pulls an image or a release, so the freshness
question does not arise.

`Nightly betas`, `Publish Beta`, `Publish JF12 Beta`, `Publish JF12 Stable`,
`Publish Release` and `Regenerate manifest`. Adapt as one element, the release
route, #118 for what a version number promises, #119 for reproducibility from a
tag, #120 for signatures and #123 for the first release. The per-generation
channels drop with the manifest: there are no host generations to publish
against.

`Publish failure alert`. Adapt, #159. Nothing in this plan swept the default
branch for a workflow that concluded non-success, and writing this document is
where that was found, so the issue was opened for it rather than the element
recorded as unplaced.

## What this repository has that the reference does not

`Line endings and encoding`, in `.github/workflows/text-hygiene.yml`, delivered
by #28 and owed to the branch by #31. The reference's formatter normalises what
it formats and nothing there refuses a tracked file stored with CRLF or one
that does not decode as UTF-8.

`Dependency lock`, in `.github/workflows/dependency-lock.yml`, delivered by
#23. The reference restores from a lock file its ecosystem maintains for it.
Here the module file and the lock have to be shown to account for each other,
and the check exists before the first requirement rather than after.

`External links`, in `.github/workflows/doc-hygiene.yml`, weekly, delivered by
#96. The reference checks its wiki's links on its own schedule; this is the
same idea kept off the pull request path so that somebody else's uptime is not
a condition of merging here.

The unit test check that needs no display and no elevation, #20, and the audit
that every test is in exactly one set and each named harness reds rather than
skips when its requirement is absent, #98. The reference runs its tests inside
the build and has no equivalent audit. Headless testability was declared a
birth requirement here, so it is checked rather than assumed.

The deterministic harness in `harness_test.go`, delivered by #21. Time,
randomness, identifiers and outbound connections are taken as dependencies and
supplied in one place, and a departure is a failing test rather than a review
comment.

The layout test in `layout_test.go`, delivered by #18 and widened by #97.
Which part of the tree may know about which other part is a verdict here. The
reference states its equivalent as a conformance test over one directory; this
one is over the dependency direction of the whole tree.

Proof that nothing leaves the host unless federation is switched on, #104. This
is the deviation upward and it has no counterpart at the reference at all. It
is also the most checkable sentence in this project's legal position, which is
why it is a test rather than a paragraph.

One record per check naming the smallest change that trips it and the run where
it reddened, #101. The reference proves its guards bite one pull request at a
time and keeps no collected record. This project's position is that a guard
nobody has watched refuse anything is a decoration, so the record is the
artefact.

## What this document does not settle

It places each element from the reference's ruleset, its workflow file names,
its job names and its triggers, and from the one script quoted above. It is not
a full reading of all twenty-three workflow files there, so an element that
exists only inside a step of a workflow whose name does not suggest it would
not have been seen.

It records what this repository intends per element, not what it has. Most of
the issues named above are open, and an element with a verdict and an issue is
still an element with nothing running behind it. `docs/required-checks.md` is
the list of what exists and can be required today; this is the list of what the
distance is made of.

It settles nothing about the ruleset. Requiring a check is a repository setting,
no pull request changes one, and #31 is where that request lives.
