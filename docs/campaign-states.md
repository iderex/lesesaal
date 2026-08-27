# The states a campaign passes through, and what is refused

Decided in issue #42.

## The four states

A campaign is always in exactly one of them.

`draft` is a campaign no volunteer has seen. It is being written and filled
with subjects, and its definition may still change in any way, because nothing
has been answered against it.

`open` is a campaign volunteers are classifying.

`paused` is a campaign that was open and is not taking answers. The campaign
owner is still there and the campaign is expected back.

`closed` is a campaign that is a record. It takes no further answer, and what
it collected is what it will always have collected unless the campaign owner
deliberately reopens it.

Closed rather than finished, complete or done. `vocabulary.md` gives retired to
a subject that needs no more classifications, and the campaign as a whole needs
a word that does not claim anybody is finished with what it collected.

## What decides a transition, beside the state

Three counts, which the store supplies and this rule never goes to look for:
how many subjects the campaign holds, how many classifications have been
recorded against it, and, for a removal, how many of the subjects that removal
concerns already carry a classification.

The third one is there so that a campaign owner correcting a collection
ingested wrong is not refused for the sake of subjects nobody has seen. A
campaign with four thousand answers may still give up a subject that has none.

## The table

Every pair has an entry. A pair nobody decided would be settled by whichever
branch fell through last, which is how a transition that destroys data arrives
without anybody choosing it, so `internal/campaign/state.go` answers all of
them and the suite asks for all of them rather than trusting this table to have
been kept level with the code.

    transition          draft      open       paused     closed
    open to volunteers  allowed*   -          -          -
    pause               -          allowed    -          -
    resume              -          -          allowed    -
    close               conseq.    conseq.    conseq.    -
    reopen              -          -          -          conseq.
    take subjects       allowed    conseq.    allowed    refused
    give up subjects    allowed*   allowed*   allowed*   refused
    reword              allowed    conseq.    conseq.    refused
    redefine tasks      allowed*   allowed*   allowed*   refused

`allowed` is allowed and does nothing beside moving the state. `conseq.` is
allowed with a consequence, which is a thing that will happen and is stated
before it does, because the campaign owner is the only person who can decide
whether it is worth it. `refused` is refused because it would destroy
something. `-` is refused because it is not a move from that state and would
have destroyed nothing.

An asterisk marks a transition whose entry depends on one of the counts above
as well as on the state:

- Opening a campaign that holds no subject is refused. It would open without
  error and then show the first volunteer to arrive what a finished campaign
  shows, which is the failure that is invisible from the owner's side.
- Giving up subjects is refused where any of the subjects concerned already
  carries a classification.
- Redefining the tasks is refused where any classification has been recorded
  against the campaign.

## Editing a definition is two transitions and not one

`task-grammar.md` separates an option's identifier from the text a volunteer
reads, and says why: a campaign owner will rewrite the wording of an option
while a campaign is running, because the first ten volunteers will misread it.
With identifiers, that edit changes what a volunteer reads and changes nothing
about what was answered.

So an edit is judged by what actually moved rather than by what the person
editing believed they were doing, and `WhatChanges` in
`internal/campaign/state.go` is the comparison. Everything an answer carries is
structural: which tasks exist and in what order, each task's type and its
bounds, and each task's option identifiers and their order. `consensus.md`
derives a label in the order the task declares its options, so a change to that
order moves the label's own representation and is not a rewording. Everything
else is text: the title, the instructions, a question, an option's words.

A rewording while a campaign is open carries a consequence rather than being
free. The volunteers before the edit and the volunteers after it read different
words for the same option identifier, and nothing in the export says which of
the two a given answer was made against.

## Redefining after answers exist is refused rather than accommodated

The issue that decided this said both were defensible. Refusal is chosen, and
it is not chosen here first: `retirement.md` already takes that position in its
own words, where a campaign owner who reads the stuck list and rewrites the
question is told that the redefinition is a new campaign rather than an edit to
this one. `task-grammar.md` refuses changing an option identifier, or removing
an option that has answers behind it, for the same reason.

What the other answer would have cost is the reason it lost rather than a
complaint about it. It puts a version on every answer, into the export, into
the consensus rule and into every count that today reads the answers to one
task, and it does so to spare a campaign owner from starting a campaign whose
question they have just discovered was the wrong one. The corpus this project
is built for is a few thousand subjects and ten volunteers; starting again is a
real cost and it is smaller than a version column nobody can drop afterwards.

The refusal says how many classifications it is protecting. A campaign owner
who reads the number is being told what starting again would cost them, which
is the thing they are actually deciding.

## Removing subjects is retirement's question rather than this one

`retirement.md` says nothing is deleted and no subject is quietly dropped, and
that a subject no volunteer could classify still carries its classifications
into the export, because the answers people gave are evidence about the subject
even when they do not add up to a label.

So a subject that has been classified stays, and a campaign owner who wants it
out of circulation retires it, which stops it costing volunteer time and keeps
the answers. The refusal names that route rather than only refusing.

A subject nobody has answered carries no such evidence. Removing it is the
ordinary correction of a collection that was ingested wrong, and it is allowed
in every state but closed.

## Reopening is the one transition that takes something back

`retirement.md` allows a retired subject to come back under three conditions:
only by the campaign owner, only as an explicit action, and never on its own. A
closed campaign reopens under the same three, and for the same reason. What
that document rejects is a rule that recomputes silently, so that a campaign
owner holding an export finds it no longer describes the campaign with nothing
marking the moment it changed.

Reopening is that moment, and the consequence says so: an export already
produced from this campaign stops describing it from here, and has to be
produced again for the two to tell the same story. Producing that export a
second time is #71.

## Closing is allowed from everywhere a campaign can be worked

A campaign owner is entitled to stop, and a machine that refused would be
protecting nothing. What closing carries instead is what it does. A subject
that had not reached its retirement rule stays unresolved rather than being
decided by the closing, and the classifications recorded up to that moment are
what the campaign will have collected.

Closing a draft is allowed as well, and says the smaller thing: the campaign
never opened, so it closes as a definition and a set of subjects with no
classification behind them.

## What a refusal says

A refusal that protects something names what would have been lost. That is the
sentence a campaign owner needs, because the alternative is a message saying a
transition is not allowed, which tells them the machine's opinion and not their
own position.

A refusal that protects nothing says so and says where the transition does
live. Pausing a draft destroys nothing, and a message claiming a loss would be
inventing one. Two kinds of refusal that read the same are two kinds a reader
learns to discount together.

## What this rule does not do

It moves no state anywhere. There is nothing to store a state in yet: #33 is
the store, and until it exists this rule is the decision about what a
transition does and the store is where a campaign is actually in one state
rather than another.

It does not decide who may ask. Whether a campaign owner is authenticated, and
what a volunteer may see of a campaign that is paused, are the surface's
questions, in #43 and #49.

It says nothing about how subjects arrive. Adding subjects here is the
transition; the bytes, the digests and what is refused at ingest are #64.
