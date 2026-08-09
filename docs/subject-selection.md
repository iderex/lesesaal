# Which subject a volunteer sees next

Decided in issue #10.

## The rule

A volunteer asking for work gets one subject, drawn as follows.

The eligible set is every subject in the campaign that is not retired, that this
volunteer has not already answered, and that is not currently held by another
volunteer.

From that set, one draw in four is taken uniformly at random from the quarter of
the eligible set the model is least certain about, and three draws in four are
taken uniformly at random from the whole eligible set. Where the model has no
opinion, the uncertain quarter is undefined and every draw is the uniform one.

A subject handed out is held for the volunteer who was given it. The hold expires
after a bounded time and the subject returns to the eligible set. A volunteer who
answers releases the hold at the same moment.

That is the whole rule. The mixture is one number, the hold is one number, and
both are named in the configuration so an operator can see them without reading
code.

## Why a mixture rather than an ordering

Uncertainty ordering is where active learning actually lives, and taken alone it
does the wrong thing with ten people. Ordering by uncertainty hands the single
hardest subject to everybody who asks, so ten volunteers spend ten looks on one
image while the collection stands still. That is sometimes exactly what is
wanted. It is not what is wanted by default.

Two parts of the rule above defuse it. The draw inside the uncertain quarter is
random rather than ordered, so ten volunteers drawing from it land on ten
different subjects. And three of four draws ignore the model entirely, so the
collection keeps moving whatever the model believes.

The proportion is one in four rather than something else because it is the
largest share that leaves the majority of effort on the uniform draw. It is a
choice and not a derivation. What would change it is a measurement from the
pilot, #115 and #116, comparing how many classifications a campaign needs to
finish at different shares. That measurement has not been made and this number
should be read as a starting position rather than a result.

## The properties, and how each one is held

A volunteer rarely sees a subject twice. Held exactly rather than rarely: the
eligible set excludes subjects this volunteer has already answered, so a repeat
is impossible rather than unlikely. The cost is that the selection needs to know
what this volunteer has answered, which is the reliability question in #39 and
the identity question in #13 arriving together at the same place.

Two volunteers working at the same time are mostly not given the same subject.
Held by the hold. The first draw takes the subject out of the eligible set for
everybody else until it is answered or the hold expires. Mostly rather than never
is the honest word: two draws arriving in the same instant can both see the
subject as eligible, and what closes that window is the write that records the
hold, which belongs to #35. This document names the property and #35 owes the
proof that the write is atomic.

Retired subjects never appear. Held by the eligible set, which reads the
retirement state that #36 maintains. Never rather than rarely, because a retired
subject is excluded at the point of the draw and not filtered afterwards.

The order is not predictable. Held by both draws being random within their pool.
A volunteer cannot know what is coming because nothing decided it until they
asked. This is not a security property and should not be relied on as one: it
makes the surface less boring, and that is all it is for.

## What a volunteer sees when there is nothing for them

Three cases, and they are different from each other.

The campaign is finished, meaning every subject is retired. Say so, show what
this volunteer contributed, and stop. Do not hold the page open waiting for work
that will not arrive.

Subjects remain but this volunteer has answered all of them. Say that too, in
different words, because it is a different fact and a volunteer who is told the
campaign is finished when it is not will believe the campaign is finished. Offer
nothing further. Re-serving a subject to somebody who has already answered it
would buy one more classification and cost the independence that makes the
consensus rule mean anything.

Subjects remain and are all held by other volunteers. This is the transient case
and it resolves when a hold expires. Say that work is out with other people and
ask them to try again, with the shortest hold expiry as the wait.

The model having no opinion is not a fourth case. It is the cold start in #60 and
the rule above already covers it: the uncertain quarter is undefined, every draw
is the uniform one, and nothing else changes. A campaign works from the first
minute with no model at all.

## What it costs per request, and how that grows

At the design point, 2,000 subjects, the eligible set is small enough to hold in
memory in its entirety, and a draw is an index lookup and a random choice.
`design-point.md` says so in the same words when it names the 4 GB it assumes,
and it names selection among the things deliberately left slow: a campaign that
runs for weeks does not need a subject chosen in under a millisecond.

Growth is the part worth stating. The eligible set is defined by three
conditions. Not retired and not held are properties of the subject, so an index
over them makes the eligible set a range rather than a scan. Not answered by this
volunteer is a property of the pair, so it is the one that grows: a volunteer who
has answered n subjects carries n exclusions. At the design point n is at most a
few hundred and the exclusion is a set membership test. At a hundred times the
design point it is the first thing to give, and the repair, which is not built
today, is to keep the exclusion as a per volunteer structure rather than as a
join.

Stated against the design point, the numbers are these. Ten volunteers ask for a
subject at up to 30 draws a minute at the busiest moment, each volunteer carries
at most a few hundred exclusions by the end of a campaign of 2,000 subjects, and
the whole request a volunteer waits on has a budget of 300 milliseconds on the
assumed host, which `design-point.md` sets and which this draw is one part of.
At a hundred times the design point `design-point.md` names subject selection as
the second thing to give, after the serialised write path, and for the reason in
the paragraph above.

All of this is a claim. Nothing has been measured, because the selection is not
written and neither is the store, and the 300 milliseconds is a target that
document sets rather than anything this project has observed. The measurement
that would settle the paragraph is the time to serve one subject at the design
point and at ten times it, taken on the assumed host, and it becomes possible
with #35.

What is deliberately not optimised: a campaign that runs for weeks does not need
a subject selected in under a millisecond, and nothing here should be made faster
until a pilot says a volunteer waited.

## What takes its ordering from here

The model milestone does not add a second ordering. #56 implements the mixture
above and #55 supplies the uncertainty number the uncertain quarter is cut from.
Where either wants a different proportion, it changes the number in this document
rather than adding a rule beside it.
