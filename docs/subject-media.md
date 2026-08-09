# Where the images live

Decided in issue #14.

## This project owns a directory

Subject media is copied into a directory this project owns, at ingest, and is
never written again afterwards. The copy is the subject. The file the operator
pointed at is an input to ingest and nothing after that.

Each copy is stored under its own content digest, recorded with the subject, and
the digest is what later checks compare against.

The directory is one path under the deployment's data directory, beside the
store, so that the whole of a deployment's state is one thing to back up and one
thing to move.

## The options that were not taken

Referencing the operator's own directory, storing only paths and copying no
bytes, makes ingest instant and costs the thing this project is for. A label is a
statement about an image, and if the image can change or vanish after the label
was made then the label is a statement about nothing, with no way to tell which
labels are affected.

Object storage over a network protocol adds a service and an endpoint that by
default is not on this host, which is refused below.

Bytes in the store beside the metadata makes the backup one artefact, which is
genuinely attractive, and makes the store large enough that restoring it is the
slow step. It also puts every image byte through the store's own write path, on a
store chosen for short campaign writes.

## The sovereignty check on the option that leaves the host

Object storage is the only candidate that puts subject media anywhere but the
operator's machine, so it is the one that has to be checked against the position
that nothing leaves the host unless federation is deliberately switched on.

It fails that check as a default. A default deployment that stores subject media
in object storage makes an outbound connection on every ingest and on every image
a browser fetches, whether or not the operator ever asked to federate anything.
That is not an exception to the position, it is the position not holding.

Recorded as the answer to that check: object storage is not in the default
deployment and is not offered as a switch today. Where it arrives later it
arrives as a federation mode with the rest of them, defined in `federation.md`,
and not as a storage setting.

The other three options keep every byte on the host, so the check does not
constrain the choice between them.

## What happens when the input changes or disappears

For the owned copy the question does not arise, and that is the point of owning
it. The copy is written once at ingest and never rewritten, so nothing this
project serves can change under a label that was already made.

For the operator's original, which this project does not control, three things
are true and each is worth stating on its own.

Its disappearance costs nothing. Ingest already copied it, and re-running ingest
against a folder that has lost a file refuses that file and says so, which is
#64's job.

Its modification costs nothing either, for the same reason. A modified original
presented at a later ingest is a different digest and therefore a different
subject, not a changed one. It gets its own identity and its own classifications.

The owned copy itself can still rot on disk. A periodic check reads each copy and
compares its digest against the recorded one. A copy that fails is marked, and
its classifications are kept with the mark rather than deleted, because a label
whose image is gone is still evidence about what volunteers saw, and silently
dropping it would change a campaign's results after the fact. What the export
carries about a marked subject belongs to #69.

## Who serves an image, and what that costs

The one process serves it, from the owned directory, as decided in
`deployment-ceiling.md`.

It does not serve the original bytes to a browser. Derived sizes are computed at
ingest and stored beside the copy, and the browser is sent the derived size that
fits what it asked for. That is #67's work and this document only fixes that the
derivation happens at ingest rather than per request. Deriving per request would
put image processing in the path of every page a volunteer opens, which is the
version of this that makes somebody on a phone wait.

The disk cost of the choice is the doubling: the operator's original stays where
it was and this project holds its own copy, plus the derived sizes.

`design-point.md` is where the per subject size now comes from, so the cost has
a figure instead of a shape. It assumes 50 MB for a plate as the operator hands
it over and 1.5 MB for the derived sizes kept beside it, and at 2,000 subjects
that is about 103 GB: 100 GB of owned copies and 3 GB derived. The operator's
original collection is another 100 GB and sits outside that figure, on the
operator's own disk, which is what the doubling above costs.

That is also the number `design-point.md` calls the one most likely to be wrong,
because the size of a digitised plate varies widely between archives, and it is
what makes disk the resource this project consumes: at the design point the
campaign's own data is under 100 MB against 103 GB of media.

Nothing here is measured. Nothing has been ingested, no derived size has been
chosen, which is #67, and this project has no plate images to weigh. So the
statement is that the storage cost is about 103 GB at the design point plus the
operator's own copy, as arithmetic over an assumed per subject size rather than
as a measurement, and the number that would replace it is a real ingest of a real
collection with the disk usage recorded, which is possible with #64 and belongs
to the pilot in #116.

## What takes its scope from here

Backup, #78. What has to be backed up follows directly from this document: the
store, the owned media directory, and the configuration. The operator's original
directory is not backed up by this project, because this project does not own it
and holds its own copy of everything it needs from it.
