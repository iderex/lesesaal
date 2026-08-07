# Whether a volunteer ever sees what the model thinks

Decided in issue #12.

## The position

A volunteer is never shown what the model thinks before their answer is recorded.
Not as a pre-filled answer, not as a highlighted option, not as a hint, not as an
ordering of the choices, and not as anything a volunteer could read as the
software's opinion. This is not a campaign setting and a campaign owner cannot
switch it on.

After the answer is recorded, the campaign may show the volunteer what the model
proposed, as feedback. That is a campaign setting and it defaults to off.

The model still does everything else. It orders part of the queue, which is
`subject-selection.md`, and it may let a subject retire on fewer human answers,
which is #36 and #38. Neither of those puts a proposal in front of a person.

## Why

A volunteer shown a suggestion agrees with it more often than a volunteer asked
cold. That sentence is a claim in this document, not a measurement: it is the
reason the position was taken, it comes from the annotation literature rather
than from anything this project has run, and no command here backs it. It is
written as a claim deliberately, because the whole point of the position is that
labels made under a suggestion are not the same evidence as labels made without
one, and stating that as a fact this project verified would be the same mistake
in the opposite direction.

What follows from it does not depend on the size of the effect. If the effect is
real the labels are contaminated. If the effect is small, showing the proposal
first buys a speed-up that costs the one property that makes a campaign's output
worth having, which is that a label is what a person thought before being told.

Feedback after the answer defaults to off for a smaller reason, and it is a
different reason. The answer just given is safe: it was recorded before anything
was shown. What feedback moves is the volunteer's next answer, and over a session
that is a real drift with nobody watching it. Off by default means a campaign
owner who wants it opts in and gets it recorded, which is the next section.

## What the rejected positions would have cost

Never show the proposal at all, in any position. This is the simplest to
implement and the easiest to defend. It costs the feedback that keeps somebody
classifying for an hour, and it throws away a cheap way for a volunteer to learn
a task they were not trained for. It is the position this one departs from by
exactly one step, and the step is taken after the answer rather than before it.

Show the proposal first, as a starting point the volunteer corrects. This is the
fastest route through a collection and the reason people build systems this way.
It produces labels that look like human labels and are not, with no way to
separate them afterwards. Refused.

A per campaign setting for showing it first. This moves the decision to a
campaign owner who has no reason to know what it costs them, and it produces a
corpus where some labels were anchored and some were not. A dataset that is mixed
with no way to tell them apart is worse than either alone. Refused.

## What the export records

Per classification, permanently, three fields rather than one:

Whether a proposal existed for this subject at the moment the answer was
recorded. This is independent of whether anybody saw it, and it is what makes the
next field readable.

Whether a proposal was visible before the answer was recorded. Under this
position it is always false. It is written out on every classification anyway
rather than omitted, because a field that is absent cannot be distinguished from
a field that was true, and a reader in a year has no way to know which position
was in force when the campaign ran.

Whether feedback had been shown to this volunteer earlier in the same session.
This is the field that carries the setting above, and it is per classification
rather than per campaign because a volunteer's first answer in a session was
given before any feedback and their twentieth was not.

The export format that carries these is #68 and #69, and what the model half of
it looks like is #62.

## The measurement that would detect anchoring here

The direct measurement, comparing labels made with a visible proposal against
labels made without one, is not available in this project's own data, because
this position never shows a proposal first. That is a consequence of the decision
and not an oversight, and it should not be quietly replaced with a weaker
measurement presented as the same thing.

What is measurable is the feedback effect, which is the only place this project
lets the model reach a volunteer at all. Within a campaign with feedback switched
on, compare a volunteer's agreement with the model on the answers they gave
before their first feedback in a session against the answers they gave after it,
and compare that difference against volunteers in the same campaign whose
feedback is off. The comparison is within campaign and within volunteer, so it
does not depend on the two groups classifying the same subjects.

That measurement needs a campaign with real volunteers and both settings in use.
It belongs to the pilot, and #116 carries it.

If the measurement finds a drift, the setting's default does not change, because
it is already off. What changes is whether the setting is offered at all, and
that is a decision this document would be reopened for.

## What the surface owes

The volunteer page implements exactly this and no variation of it. A test proves
that under the position above a proposal cannot reach the page before an answer
is recorded, and the test is a refusal rather than an assertion about markup: the
proposal is not sent to the browser at all, so it cannot be read out of the page
source, out of a request the page made, or out of anything cached. #43 builds the
page, #44 constrains what it may load, and #59 owes the proof.
