# What a campaign may ask a volunteer

Decided in issue #6.

## The grammar

A campaign asks a volunteer about a subject. The unit of asking is a task, and a
campaign is an ordered list of them.

Every task in the first release has one shape. A task declares a fixed list of
options, and an answer to that task is a subset of that list. Nothing else is a
task.

Two types fall out of that shape, and there is no third.

A single choice declares that exactly one option may be chosen. A choice of
several declares a lower and an upper bound on how many may be chosen, both
written by the campaign owner rather than assumed, so that "at least one" and
"any number including none" are different declarations and neither is a default
somebody has to guess.

A yes or no question is not a type of its own. It is a single choice over two
options, and saying so keeps the count of things that need a rendering, a
comparison, a consensus rule and an export column at two rather than three. A
campaign owner who wants yes and no writes two options named yes and no, and the
software neither knows nor cares that the words are those.

## What an answer is, and what it is compared against

An option carries a stable identifier and the text a volunteer reads, and the two
are separate fields. The identifier is what an answer, a derived label and an
export all carry. The text is what the surface renders and nothing else depends
on.

This separation is the reason it is written here rather than left to the schema.
A campaign owner will rewrite the wording of an option while a campaign is
running, because the first ten volunteers will misread it. If the answer carried
the text, that edit would silently split one option into two in the export, or
merge two into one, and the classifications made before the edit would no longer
be comparable with the ones made after. With identifiers, the edit changes what a
volunteer reads and changes nothing about what was answered. Changing an option's
identifier, or removing an option that has answers behind it, is a different
campaign and is refused rather than accommodated.

An answer to a single choice is one option identifier. Two answers agree when
they name the same identifier and disagree otherwise. There is no partial
agreement and no distance between two options: a fixed list has no order this
project knows about, so nothing here may treat the second option as nearer to the
first than the fifth is. A campaign owner whose options really are ordered, so
that being one step out matters, is asking for a number, which is refused below.

An answer to a choice of several is a set of option identifiers, held sorted by
identifier so that the same answer has one representation. Two answers agree when
the sets are equal. They also decompose: for each declared option, an answer
either contains it or does not, so a set answer over n options is n independent
yes or no answers that happen to have been collected together.

That decomposition is stated here because it is what makes the type cheap
everywhere else. It gives a consensus rule the option of working per option
rather than over whole sets, which is the difference between a rule that has
something to say when nine volunteers agree about three options out of four and a
rule that sees nine disagreements. Which of the two is used is not decided here.
Issue #7 decides it, and this document owes it the two comparisons rather than
the choice between them.

## What the export carries

Per classification, the export carries the campaign, the subject, the task, the
option identifiers the volunteer chose, and whatever the personal data decision
in #1 entry 3 says may identify the volunteer. A single choice carries a list of
one. Both types are written the same way, so a reader does not need to know which
type a task was to read the column, and the campaign definition travels with the
export so a reader can find out.

The derived label has the same shape as an answer, because it is an answer to the
same task: a set of option identifiers, of size one where the task is a single
choice. What it carries beside itself, the confidence and the number of answers
behind it, is #7's and is not restated here.

The export also carries the campaign definition as it stood when the campaign
ran, including every option identifier with its text. Without it an export is a
column of identifiers nobody outside the deployment can read. Issues #68 and #69
own the file and its provenance and take the shape above from here.

## What is refused, and what would bring each one in

Refusal is at the campaign definition boundary. A campaign declaring a task this
document does not describe does not start, and nothing is half supported: a type
that renders but has no comparison rule produces answers that cannot become a
label, which is worse than a campaign that refused to start.

A point marked on the image is out. Two points are never equal, so a comparison
needs a distance and a radius within which two volunteers are held to have marked
the same thing, and that radius is a scientific choice belonging to the campaign
rather than a constant this project can pick. Consensus over points is clustering
rather than counting, and retirement then has to be defined over the number of
clusters rather than the number of answers. What would bring it in is a decided
distance measure and cluster rule in #7's document, a way for a campaign owner to
declare the radius in the units of their own collection, and a surface that can
take a click and give it back, which #43 is not asked to build today.

A box drawn on the image is out, for the reason above and one more. It needs an
overlap measure and a threshold above which two boxes are the same box, and the
common measure carries a known failure on small objects that a campaign owner
would have to be told about. What would bring it in is everything a point needs
plus that measure and its threshold.

A number is out. Comparison over a continuum needs a tolerance, and the tolerance
is per campaign for the same reason the radius is. Consensus is then a median or
a mean with a spread rather than a count, retirement has to be defined over that
spread, and the export column is a distribution rather than a value. What would
bring it in is a way to declare the tolerance and an aggregation decided in #7.
The nearest thing available today is a single choice over declared ranges, which
is a worse instrument and an honest one.

A short piece of text is out, and it is the one most likely to be asked for,
because a plate archive wants what is written on the plate envelope. There is no
comparison rule for free text that anybody has agreed on. Case folding and
whitespace trimming look like one and are not: they make "Hyades" and "hyades"
equal and leave "Hyades" and "the Hyades" apart, which is a rule about typing
rather than about what the volunteer meant. The honest options are to aggregate
it badly or to admit it is never aggregated, and admitting it means a campaign
whose export carries a column that no consensus rule ever touched and no
retirement rule can read. What would bring it in is a decision that this project
supports unaggregated columns at all, which is a decision about what a campaign
is for rather than about text.

None of the four is refused because it is hard to render. Each is refused because
the chain behind it, a comparison, a consensus rule, a retirement rule and an
export column, is longer than the rendering and is where the work actually is.

## What a campaign owner writes

A campaign definition is a plain text file the owner writes by hand and keeps
under their own version control. One campaign to a file.

For the campaign it names an identifier, a title, and the instructions a
volunteer reads before starting. For each task it names an identifier, the
question text, the type, and the options, each with its own identifier and its
text. A choice of several also names the lower and upper bounds. Nothing else is
required and nothing may be omitted.

The concrete syntax is not decided in this document. It follows the toolchain
that #17 chooses, it gets the means check that choice carries, and #32 turns this
shape into types. What is decided here is that the definition is a file a person
writes without this project's help, rather than a form on a page or a sequence of
API calls, so that a campaign can be read, reviewed and kept alongside the
collection it is about.

A definition that names a type outside the grammar is refused before the campaign
starts, with a message that names the offending task and lists the supported set,
so that a campaign owner learns what is available from the refusal rather than
from reading source. That refusal does not exist yet and this document does not
claim it does. It is owed by #32, which builds the definition boundary, and the
test that proves it belongs there.

## Anything that is not an image

The grammar above is written for image subjects, which is what every part of this
plan assumes today.

Whether audio, video, one dimensional spectra or free text subjects are a first
release promise is not settled here and is not assumed either way. It is entry 7
of #1, it belongs to the maintainer, and it is the only thing that reopens this
document. A widening decided there is a change to this file, not a type added
beside it.

## What this document does not decide

How several answers become one label is #7. When a subject is finished is #9.
What the surface looks like and what it may load is #43 and #44. This document
gives all four the same list of types and expects them to name types from it and
no others.
