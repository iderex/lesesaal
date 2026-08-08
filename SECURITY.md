# Security policy

Written for issue #108.

## Reporting something

Report privately, through this repository's own private reporting route: the
Security tab, then "Report a vulnerability". It is switched on, which is a fact
about the repository rather than a claim about intent:

    $ gh api repos/iderex/lesesaal/private-vulnerability-reporting
    {"enabled":true}

That route opens a draft advisory only the reporter and the maintainer can read,
it carries attachments, and it becomes the published advisory later without
anything being copied between places.

Do not open a public issue for a suspected vulnerability, and do not put the
detail in a pull request. A public issue is readable by everybody the moment it
exists, including by anybody running a deployment that has no fix yet. If the
private route is unavailable to you, open a public issue that says only that you
have something to report and gives no detail, and a private route will be opened
for you.

A report is most useful when it says what an attacker gets, which version or
commit it was seen on, and what has to be true of the deployment for it to work.
A working demonstration is worth more than a description, and a description is
worth more than nothing.

## What happens next

An acknowledgement that the report arrived, within five days.

An assessment saying whether it is in scope, whether it is accepted as a
vulnerability, and what severity it is being treated as, within fourteen days.

A fix on the default branch, or a written reason why there will not be one,
within ninety days of the acknowledgement. A published advisory when the fix
lands, or at ninety days, whichever comes first.

Credit in the advisory under whatever name the reporter asks for, or no credit
if that is what they want.

Those are the intended times and nothing enforces them. If an acknowledgement
does not arrive within the five days, the report did not reach anybody, and the
right response is the detail-free public issue described above rather than a
second private report into the same silence.

## What is in scope

The software in this repository, and the artefacts this repository publishes.
That means the campaign core, the boundary that accepts classifications, the
volunteer surface, the campaign owner surface, the export, the model interface,
the container image and the deployment composition, and the workflow files
under `.github/`.

Some examples in this project's own terms, none of which are known to exist,
because there is no code yet:

A volunteer able to read or change another volunteer's answers, or to submit as
another volunteer.

A classification accepted for a subject that has retired, in a way that moves a
label already exported.

A subject media path that escapes the directory this project owns, which
`docs/subject-media.md` says is the only place subject bytes live.

Anything in a default deployment making an outbound connection.
`docs/federation.md` states the permitted list for a default deployment and the
list is empty, so any outbound connection at all is a finding, whether it comes
from this project's own code or arrives through a dependency.

A campaign owner's credential, a volunteer's session or an operator's secret
reaching a log line, an error page or a bug report, against what #83 and #84
require of logs and of secrets.

A model proposal reaching a volunteer's browser before their answer is recorded.
`docs/model-visibility.md` makes that impossible by construction rather than by
setting, so a way to obtain one is a defect in this project's central scientific
claim and is treated as a vulnerability rather than as a bug.

## What is out of scope

The operator's own deployment and the choices they made in it. A host reachable
from the internet when the operator meant it not to be, a reverse proxy
terminating TLS in a way this project never saw, a certificate that expired, a
backup left readable. This project is self-hosted, so a report about a
particular running instance belongs to whoever runs it.

Anything a campaign owner is entitled to do with their own credential. Reading
their own campaign's classifications is the product working.

A finding in a third-party dependency. Report it to that project. Telling this
project as well is welcome and useful, because the response here is to pin,
update or drop the dependency, but the fix is not here.

A missing response header, a weak cipher suite in an operator's proxy, or a
scanner's output with no demonstration of what an attacker gets. A report of this
kind that does show the impact is in scope, and it is the demonstration that
moves it.

Denial of service by volume against somebody's deployment. A deployment is one
process on one machine by decision, in `docs/deployment-ceiling.md`, and it is
not built to absorb that. An amplification where a small input costs the
deployment a large amount of work is a different thing and is in scope.

Social engineering of the maintainer, physical access, and anything requiring the
operator to run something they were told not to run.

## Which versions get a fix

There is no release and there is no tag:

    $ gh api repos/iderex/lesesaal/releases --jq 'length'
    0
    $ git ls-remote --tags https://github.com/iderex/lesesaal.git | wc -l
    0

So there is exactly one supported version and it is the current tip of `main`.
Nothing is backported, because there is nothing to backport to, and no
deployment is known to exist.

That changes with the first release, #123. From then the policy is that the
latest release gets the fix and older ones do not, until a stated support window
says otherwise, and #121 owns the upgrade path a deployment follows to reach it.

## What this policy does not cover

Advisories arriving in the other direction, about the dependencies this project
carries rather than about this project, are a separate procedure. It belongs to
the supply chain issue, #99, and it does not exist yet. Both documents have to
agree once it does, and today there is nothing here to agree with.

There is no threat model. #107 owes it and it is blocked on a decision that has
not been made, which is whether a deployment is expected to face the open
internet at all, entry 2 of #1. The scope boundary above is therefore drawn from
this plan's documents rather than derived from a threat model, and it will move
when one exists.

Nothing in this repository enforces any of this. There is no check that reads
this file and no route that measures a response against the times above.
