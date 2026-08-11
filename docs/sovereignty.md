# What stays on the operator's host, and what stands behind that

Written for issue #103. This is the sovereignty statement, addressed to the
person deciding whether to run this on their own machine. It says what data
exists, where it sits, what a default deployment sends anywhere, and what each
of those sentences rests on.

Every clause below is followed by the thing that holds it up. Where that thing
is a test or a check, the command that runs it is written out. Where there is
no such thing yet, the clause says this is a claim in those words and names the
issue that would turn it into something a machine refuses. A reader is meant to
be able to tell the two apart without knowing this project.

Most of this document is claims today. That is what an honest statement looks
like on a project whose deployment does not exist yet, and the alternative is
the sentence people write instead, which is that the software is private by
design.

## The position

A campaign runs on the operator's own machine. The subjects, the answers, the
derived labels, whatever a model is trained on and whatever the deployment logs
are all on that machine, and a default deployment makes no outbound connection
at all. Personal data never leaves the host unless the operator deliberately
switches on one of the four federation modes, one at a time, naming the
destination each one sends to.

`docs/federation.md` is where that was decided and it is the authority for what
a default deployment may send. This document takes the list from there rather
than restating it, so that widening what may leave the host stays one change in
one file.

## What data exists, and where it is

Subject media, which is the images a campaign owner brought. Copied into a
directory this project owns at ingest, stored under its own content digest, and
never written again. `docs/subject-media.md` decided that, and it rejected
referencing the operator's own directory and object storage for reasons written
there.

This is a claim. There is no ingest and no directory yet. #64 copies the bytes,
#33 records the digest beside the subject, and until both exist nothing in this
tree puts an image anywhere.

Subject metadata, which is whatever the campaign owner knew about each image.
Held as they supplied it. This is a claim, for the same reason: #33 is the issue
that stores it.

Classifications, which are the answers volunteers gave. This is a claim. #34 is
the write path and it does not exist.

Derived labels, which are what the consensus rule computed from those answers.
The rule is in `internal/campaign/consensus.go` and it reads a task, the
answers and the campaign's threshold. It reads no volunteer identifier, no
address and no clock, so a label carries nothing about a person by
construction:

    go test ./internal/campaign -run TestTheComputationTakesNothingButItsThreeArguments -count=1 -v

Whatever identifies a volunteer. Not decided. #13 is the inventory of what this
project may hold about a person and it is open, and two entries of #1 decide
its shape. Nothing here assumes an answer in either direction, and this
document will not read as though the strongest option had already been chosen.

Whatever the deployment writes to a log. Not decided. #83 is where a log line
is kept from carrying an answer, a volunteer identity or an address, and #82 is
the configuration that would switch a proxy's own logging off.

## What a default deployment sends outbound

Nothing. Not on start-up, not on first run, not on a schedule, and not while
serving a page. `docs/federation.md` lists the usual exceptions one at a time
and refuses each of them: certificate issuance, update checks, fonts and
scripts from a content network, error reporting, analytics, a model downloaded
at run time, dependency resolution at run time, and time synchronisation.

This is a claim about a deployment, because there is no deployment. #104 is the
test that watches an isolated host and compares every connection against that
list, and it is open.

Two things stand behind it today, and neither is the test #104 owes.

Nothing in this tree opens a socket except through one injected field. A direct
read of the standard library's dialling and HTTP calls is refused everywhere
but the wiring package, one identifier at a time, and the refusal names the
file, the line and what to use instead:

    go test . -run TestOnlyTheWiringReadsTheRuntime -count=1 -v

The field itself is `Dial` on `campaign.Depends`, declared in
`internal/campaign/depends.go`, and `docs/harness.md` is where the shape is
argued. Its whole purpose is that the absence of an outbound connection is
visible rather than assumed: the fake a test injects refuses every address, so
a dependency that starts calling somewhere becomes a failing test rather than a
firewall's problem later.

    go test ./internal/campaign/campaigntest -run TestTheDiallerRefuses -count=1 -v

And there is no third-party code in the build to phone home from. The module
requires nothing outside itself, which the gate reports on every run and which
one command shows:

    go list -mod=readonly -m all
    github.com/iderex/lesesaal

`docs/dependencies.md` is the inventory behind that, and it is honest about the
part it does not cover: two tools are pinned by a version string with nothing
refusing a floating one. The day a requirement does arrive, the constraint in
`docs/federation.md` applies to it, which is that a dependency breaching the
outbound list is a reason to reject that dependency.

## What changes when the operator federates

Four modes, each switched on by itself, each naming its own destination.
Pooling subjects sends subject media and subject metadata. Publishing results
sends the export. Sharing a model sends the weights. Announcing a campaign
sends the campaign's description and the address it is reachable at.

Two of those four are marked in `docs/federation.md` as carrying personal data
today, and they are marked that way because the questions underneath them are
open rather than because an answer was reached. Whether an export says anything
about a volunteer is entry 3 of #1. Whether a model trained on volunteer labels
may be published at all is entry 4. Until those are answered, the safe reading
is the one that does not quietly widen what may leave the host.

This is a claim. No mode is implemented, no configuration reads one, and the
start-up line that would print which modes are on is #85's.

## What this statement does not cover

The operator's own reverse proxy. It sees every request and, in almost every
stack, writes an address into a log by default. That log is the operator's, on
the operator's machine, and turning it off is a change to their proxy rather
than to this software. #83 and #105 are where it is written down; nothing here
reaches it.

The operator's own backups. Where a backup goes is the operator's decision, and
a backup copied to a service somewhere is data leaving the host by a route this
project never sees.

The operator's hosting provider. Running this on a machine somebody else owns
means that party can reach the disk. The position above is about what the
software sends, not about who can read the machine it runs on.

The operator's own decision to publish. An export written to a file and put
somewhere is the point of the export. Nothing here can distinguish that from a
mistake, and nothing tries to.

What the subject media itself contains. A plate envelope photographed with a
name written on it carries personal data into this system as an image, and no
rule in this project can see it. That is the campaign owner's to check before
ingest, and `docs/scope-of-use.md` is where the wider version of that
responsibility sits.

The volunteer's own browser and network. Between a volunteer and the operator's
host there is a connection this project does not own.

Anything about correctness. This statement is about where data goes. It says
nothing about whether a label is right, which is `docs/consensus.md`'s subject
and is not a claim this project makes at all.

## What would move a claim into the other column

Named here so that the balance of this document is a measurement rather than an
impression. #104 is the test that a deployment makes no outbound connection.
#44 is the surface loading nothing but from the operator's host, and the policy
that refuses it. #77 is certificates without an outbound call the operator did
not ask for. #82 is the configuration that fails closed on a mode with no
destination. #85 is the start-up line that says which modes are on. #83 is the
log that cannot carry a volunteer's content. #13 and entry 3 of #1 are what
turn the sentence about personal data from a position into an inventory.

Nothing in that list closes by being written about. Each one is a mechanism or
it stays a claim, and this document is where the difference is visible.
