# The size this is built for, and what gives first above it

Decided in issue #3.

## Nothing below is measured

Every number in this document is either assumed or derived by arithmetic from
something assumed. Nothing has been measured, because there is nothing to
measure: no code exists, no store is chosen, no collection has been ingested,
and this project has never served a page to a volunteer.

Each number says which of the two it is. Each section that ends in a number also
says what measurement would replace it and which issue produces that
measurement, so a later reader can tell an assumption that was never revisited
from one that survived being checked.

The point of writing them down before they are measurable is that every later
argument about a store, a queue, an index or a model is an argument about cost
relative to a size. Without the size fixed, those arguments are settled by
taste.

## Where the two starting numbers come from

Two numbers are inherited rather than chosen here:

    git grep -n '2,000 images and ten volunteers' -- README.md
    README.md:12:A group with 2,000 images and ten volunteers ends up in Google Forms. That is a

Everything else below is derived from those two together with assumptions that
are named where they are used.

## The design point

Subjects in one campaign: 2,000. Inherited from the readme. A campaign is one
collection somebody wants classified, and the collection that motivated this
project is a plate archive drawer rather than a survey.

Volunteers in one campaign: 10. Inherited from the readme. These are people the
campaign owner knows, which is what makes the number small and stable.

Volunteers classifying at the same moment: 10. Assumed, and deliberately the
worst case rather than an average. Ten people in one room at a classification
session is exactly how a working group uses an afternoon, so the busiest moment
is every volunteer at once rather than some fraction of them.

Time one volunteer spends on one subject: 20 seconds. Assumed. It covers looking
at a plate, choosing from a fixed list of options and submitting. A task with
more options or a harder image is slower, and slower is the easy direction: it
lowers every rate below.

Classifications per minute at the busiest moment: 30, which is one every two
seconds. Derived from the two numbers above, 10 volunteers divided by 20 seconds
each. This is the peak arrival rate the one process has to absorb.

Answers before a subject is finished: 5. Assumed here only to make the totals
below computable. The rule that actually decides it is #9's, and the numbers
that rule consults are #7's. If it settles on a different figure, the totals in
this section move with it and the shape of the argument does not.

Classifications in a finished campaign: 10,000. Derived, 2,000 subjects times 5
answers.

Volunteer effort to finish a campaign: about 56 hours in total, which is under
six hours each for ten people. Derived, 10,000 classifications at 20 seconds.
Spread over the weeks a working group actually has, that is a campaign that ends
rather than one that stalls, and it is the reason this size is the design point
instead of a smaller one.

Size of one subject as the operator hands it over: 50 MB. Assumed, and this is
the number in this document most likely to be wrong. A digitised photographic
plate is a large image and the range across archives is wide. It is called out
because it is also the number that dominates everything on disk.

Size of the derived sizes kept per subject: 1.5 MB in total across them.
Assumed. What the derived set actually is belongs to #67, and this is a
placeholder for its order of magnitude rather than a decision taken here.

Disk a finished campaign occupies: about 103 GB. Derived, 2,000 times 51.5 MB.
Of that, 100 GB is the copy this project owns under the rule in
`subject-media.md` and 3 GB is the derived set. The operator's original
collection sits outside this figure, on the operator's own disk, which is the
doubling that document costs out.

Disk everything that is not media occupies: under 100 MB. Derived from an
assumed few hundred bytes per classification row, 10,000 of them, plus subject
rows, indexes and the record of what the software decided.

## Media is three orders of magnitude larger than everything else

That is the single most useful consequence of the numbers above, so it is
stated on its own rather than left to be noticed.

At the design point the campaign's own data is about 100 MB and its media is
about 103 GB. A change that makes the store twice as compact saves 50 MB. A
plate archive whose scans are 100 MB rather than 50 MB costs another 100 GB.

So a disk argument in this project is an argument about images, and an argument
about rows is not a disk argument at all. The same conclusion pushes back on
schema cleverness sold as a space saving.

## The host this assumes

Two cores, 4 GB of memory, and 250 GB of disk.

Two cores because the one process serves pages, holds the store and runs
background work, and nothing above asks for parallelism beyond that: 30
classifications a minute is one short unit of work every two seconds.

4 GB because the eligible set for 2,000 subjects, the campaign definition and
the model's per subject scores are all small enough to hold in memory in their
entirety, and because 4 GB is what the smallest machine a working group is
likely to have or rent already carries.

250 GB because 103 GB of media plus room to ingest a second campaign beside the
first is the shape of the disk requirement, and because disk is the resource
this project actually consumes.

No accelerator is assumed, anywhere. Training on the operator's own machine
without one is #54's constraint and this document does not relax it.

This host is an assumption about what an operator has, not a minimum this
project has verified it runs on. The measurement that replaces it is #117,
which weighs a running deployment on a stated machine.

## What gives first at ten times the design point

Ten times is 20,000 subjects, 100 volunteers, and 300 classifications a minute,
which is 5 a second.

Disk gives first. At the assumed subject size the media directory is about 1 TB,
which has already left the assumed host, and it is the only figure here that
grows linearly with the collection and has no algorithmic repair. Nothing about
the software changes the number; either the operator brings a bigger disk or the
collection is smaller than assumed.

Ingest gives second. Copying 20,000 plates and deriving sizes for each is a
single-process job an operator waits on once, and at ten times it is the step
that turns from minutes into a working day. It is not in the path of any
volunteer, which is why it is second rather than first.

The write rate is not what gives. Five short writes a second is an ordinary load
for an embedded store, and that sentence is arithmetic against an assumption
rather than a measurement, in the same state as the equivalent paragraph in
`deployment-ceiling.md`.

## What gives first at a hundred times the design point

A hundred times is 200,000 subjects, 1,000 volunteers, and 3,000 classifications
a minute, which is 50 a second, alongside the reads that serve them.

The serialised write path gives first. An embedded store takes writes in one
line, which is the ceiling `deployment-ceiling.md` names as the price of the
shape, and at 50 writes a second competing with the reads that pick the next
subject that line is the thing under pressure.

Subject selection gives second, and for a different reason. The eligible set
excludes the subjects this volunteer has already answered, which is a property
of the pair rather than of the subject, so it is the one part of the draw that
grows with a volunteer's own history. `subject-selection.md` states this and
names the repair it does not build.

Neither is repaired by a bigger machine, and that is the honest summary of the
ceiling. At a hundred times this is a different product with a separate database
and a queue in front of it, and the way such a proposal is judged is the five
questions in `deployment-ceiling.md` rather than a decision taken here.

## What is deliberately left slow

Naming these prevents work being spent making them fast, which is work that also
buys complexity nobody asked for.

Subject selection. A campaign that runs for weeks does not need a subject chosen
in under a millisecond. A scan of an in-memory eligible set of 2,000 is fine, and
nothing here should be indexed further until a pilot reports a volunteer waiting.

Ingest. It runs once per campaign, an operator started it, and it may take
minutes or hours. What it owes is a clear account of what it refused, which is
#64, not speed.

Export. It runs rarely and somebody is willing to wait for it. What it owes is
being repeatable and complete, which is #71, and using memory that does not grow
with the campaign, which is #68. Neither of those is a speed requirement.

Model training. It may take an hour and it is never in the path of a volunteer
waiting for a subject. The operator triggers it, which is #57.

Backup and restore. Copying a file that is mostly plate images takes as long as
the disk takes, and the whole argument for the embedded store is that the
procedure is short rather than fast.

## The one thing that is not allowed to be slow

The request a volunteer waits on, which is the one that hands out the next
subject and the page that renders it.

The budget is 300 milliseconds on the assumed host for the server's part,
excluding the image bytes and the network. That is a target this project has not
measured and is not a claim about anything that exists. It is set here so that
#45, which is about a phone on a bad connection, is arguing against a number
rather than against a feeling.

## What would replace the numbers here

Each of these is a measurement that does not exist yet, with the issue that
produces it.

The per subject size and the ingest cost: a real collection ingested and its
disk usage recorded, which becomes possible with #64 and is done for real in
the pilot, #116.

The classification write rate a single embedded store sustains: #35, once the
store and the schema exist.

The footprint of a running deployment on a stated machine: #117.

The time a volunteer actually spends on one subject and the number of answers a
subject actually needs: #115 and #116, which is also where the assumed 20 seconds
and the assumed 5 answers are first exposed to people.

## What takes its numbers from here

Five documents point at this one rather than carrying a number of their own,
which is the arrangement this document exists to complete:

    git grep -c 'design point' -- docs/
    docs/deployment-ceiling.md:3
    docs/design-point.md:6
    docs/model-boundary.md:1
    docs/retirement.md:1
    docs/subject-media.md:2
    docs/subject-selection.md:5

Six lines, because the command counts this document's own uses of the phrase as
well. That output was pasted here when three documents pointed at this one, and
it is taken again rather than added to, so it is what the command produces at
the commit that carries it.

Where a later issue needs a size, it cites this document rather than restating a
figure, and where it disagrees with one it changes the figure here instead of
carrying a second one beside it.
