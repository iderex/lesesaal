# What one command is allowed to start

Decided in issue #4.

## The number is one

A default deployment starts one process. That process serves the volunteer page,
serves the campaign owner's page, holds the campaign state in a store embedded in
itself, and serves subject media from the directory this project owns.

There is no second process in the default deployment. No separate database, no
queue, no cache, no worker, no reverse proxy started by this project.

Each part below is inside that one process, with the reason it is not somewhere
else.

The store is embedded because backup then becomes a file copy and restore becomes
putting the file back, which is what makes the backup story short enough that an
operator actually has one. It also means the whole test suite runs with nothing
installed and no container, which is a birth requirement of this project rather
than a convenience. Which embedded store is a separate question, decided with the
language in #17 and the storage layer in #33, and this document fixes only that
it is embedded.

The web surface is in the same process because a separate one would exist only to
forward requests to this one, and the operator would gain a second thing that can
be down in exchange for nothing at the size this project is built for.

Media is served by the same process from the owned directory, which is decided in
`subject-media.md` and costed there.

Background work, meaning model training and any periodic check, runs inside the
same process on its own schedule rather than in a worker. A worker exists to
survive the death of the thing that queued the work. Here the thing that queues
the work and the thing that does it are the same process, so the worker would
add a failure mode rather than remove one.

## What a proposed new service has to pass

Apply this to a concrete proposal without asking the author what they meant. All
five, in order, and a proposal that cannot answer one has failed at that one.

1. Name the property the campaign loses without the new service, in terms a
   campaign owner would notice. Not a property of the code. If the sentence
   cannot be finished without naming an internal component, the proposal is about
   the code and not about the deployment.

2. Show that the property cannot be held inside the existing process, with a
   measurement taken at the design point rather than an argument from general
   practice. The measurement carries the command that produced it. A proposal
   whose evidence is that this is how such systems are usually built has produced
   no evidence.

3. Count what it costs the operator, as a list and not as a sentence: one more
   thing to start, one more thing that can be down, one more thing in the backup,
   one more version to track, one more port or socket, one more way the upgrade
   can half finish. Every item is paid whether or not the service is ever under
   load.

4. Say what the deployment does when the new service is absent or refuses to
   start. Either the campaign still comes up in a reduced form, and that form is
   described, or it does not, and the proposal says plainly that the one command
   now has a second way to fail.

5. Name who decided. Adding a service to the default deployment is a maintainer
   decision. A reviewer may not grant it, and neither may the person proposing
   it.

An optional service does not skip 1 to 3. It answers 4 by construction, which is
the only part being optional pays for.

## What is optional, and what asking costs

Nothing is optional in the default deployment, because there is nothing beside
the one process to make optional.

Two things sit outside it and are the operator's rather than this project's. A
reverse proxy or certificate terminator in front, which #77 covers, and which
costs the operator a second thing to configure and a second place a name can be
wrong. An external store, if the sovereignty position is ever relaxed enough to
allow one, which `subject-media.md` rejects for the default and which is not
offered as a switch today.

Where either arrives later it arrives through the five questions above, and the
answer is written where the change is argued.

## The ceiling this puts on the size

An embedded store serialises writes. That is the ceiling the shape buys, and it
is the number a later argument about a separate database has to beat.

The figures this project is built for are the ones in the readme, around two
thousand subjects and around ten volunteers on one machine. Against those, a
classification is one short write, and ten volunteers answering as fast as a
person can look at a plate produce something under two writes a second.

That comparison is a claim and not a measurement. Nothing has been run, because
there is no code to run: the store is not chosen, the schema does not exist, and
the write is not written. What the claim rests on is arithmetic over the readme's
figures and an assumption about how fast a person classifies, and both belong to
the design point, which is issue #3 and is not yet written down.

So the honest statement is this. The ceiling is unmeasured. The first measurement
that would settle it is the classification write rate a single embedded store
sustains on the assumed host, and it becomes possible once #33 exists. Until
then, a proposal to add a database is answering question 2 above against nothing,
and so is a claim that one is not needed.
