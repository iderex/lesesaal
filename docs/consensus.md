# How several answers become one label

Decided in issue #7.

## The rule

Count the answers. Per task, and per option inside a task, with a threshold the
campaign owner declares.

For a single choice, each answer names one option. The label is the option named
by the largest share of the answers, if that share reaches the campaign's
agreement threshold. If no option reaches it, or if the largest share is held by
more than one option, there is no label.

That second condition is what suppresses a label. It is not the level-at-the-top
flag, which is narrower and is defined once, below, under what the label carries
beside itself. A subject can be left without a label by this condition and still
carry the flag as off.

For a choice of several, each answer is a set. It decomposes: for every option
the task declares, an answer either contains it or does not, which
`task-grammar.md` states and this rule uses. Apply the count per option. The
label is the set of options whose containing share reaches the threshold. If
that set breaks the lower or upper bound the task declared, there is no label.

The threshold is one number per campaign, declared by the campaign owner, and
its default is 0.7. That default is a starting position and not a result, in the
same state as the one in four in `subject-selection.md`. What would move it is
the pilot, #115 and #116, comparing the labels a campaign produces at different
thresholds against the subjects a campaign owner checks by hand.

Nothing else is consulted. Not who answered, not how often they have agreed with
anybody before, not what the model proposed, not the order the answers arrived
in and not the time they arrived. A rule that reads none of those is one an
operator can check by hand from the export, which is the property this project
is trading for.

## What the threshold means at the counts a campaign actually sees

A share threshold is a step function over small counts, and a campaign owner who
reads 0.7 as roughly seventy percent will be surprised by the third row. The
smallest agreeing count that reaches 0.7:

    answers   agreeing answers needed   the share that clears
    1         1                         1.00
    2         2                         1.00
    3         3                         1.00
    4         3                         0.75
    5         4                         0.80
    6         5                         0.83
    7         5                         0.71

At three answers, 0.7 is unanimity, because two of three is 0.67. That is a
consequence worth stating in the campaign owner's guide, #113, rather than
leaving it to be discovered from a campaign that produced fewer labels than
expected.

## Why counting, and what the alternatives would have bought

A weighted vote, where a volunteer who agrees with the consensus more often
counts for more, recovers real information: volunteers do differ, and the
difference is large enough to change labels. It costs a label an operator cannot
check by hand, and it makes the weight of one volunteer depend on the answers of
everybody else, so a single late classification can move labels on subjects that
volunteer never saw. Rejected for the first release.

A joint estimate of volunteer reliability and subject difficulty, which is the
standard probabilistic treatment, recovers more still and is the right answer for
a large campaign. It costs an iterative fit whose output nobody can reproduce
with a spreadsheet, in a project whose argument is that an operator understands
what they are running. Rejected for the first release.

A plain majority with no threshold always produces a label, including from two
answers that disagree with a third. Rejected: it hides disagreement instead of
reporting it, and the subjects it hides are the ones a campaign owner most needs
to see.

Unanimity produces labels nobody argues with and leaves most of a real
collection unlabelled. Rejected as a default, and available to a campaign owner
by declaring a threshold of 1.

First answer wins is what a campaign becomes if it retires on one answer. It is
not rejected here because it is not a consensus rule at all: it is a retirement
setting, and #9 is where it is priced.

Breaking a tie by the option's position in the task, or by which answer arrived
first, is rejected for the reason it looks harmless. Both are deterministic, so
both would satisfy the reproducibility requirement, and both encode something
that is not evidence about the plate. A tie is a fact about the subject and is
reported as one.

None of the rejections is permanent. Every one of them can be run downstream by
somebody who has the export, which is the section below, and that is what makes
this choice a starting position rather than a decision about the science.

## The rule worked by hand

A single choice, agreeing:

    task t1, single choice over {clear, marked, unusable}
    answers   clear, clear, marked, clear, clear
    counts    clear 4, marked 1, unusable 0
    shares    clear 0.80, marked 0.20
    0.80 >= 0.70, so label = {clear}, confidence 0.80, answers 5

A single choice, level at the top, which is the tie:

    task t1, single choice over {clear, marked, unusable}
    answers   clear, marked, clear, marked
    counts    clear 2, marked 2, unusable 0
    shares    clear 0.50, marked 0.50, unusable 0.00
    two options hold the top and unusable was named by nobody, so no label and
    level at the top is on. Recorded as a disagreement, answers 4

A single choice, spread, with two options level at the top and a third named:

    task t1, single choice over {clear, marked, unusable}
    answers   clear, clear, marked, unusable, marked
    counts    clear 2, marked 2, unusable 1
    shares    clear 0.40, marked 0.40, unusable 0.20
    0.40 < 0.70 and the top is shared, so no label. Level at the top is off,
    because unusable was named. Recorded as a disagreement, answers 5

The heading and the counts under it once disagreed about that last line, which
is what #180 was opened on. They are made to agree here rather than in the code
reading whichever of them a reader reached first.

A single choice, three answers naming three different options:

    task t1, single choice over {clear, marked, unusable}
    answers   clear, marked, unusable
    counts    clear 1, marked 1, unusable 1
    shares    clear 0.33, marked 0.33, unusable 0.33
    0.33 < 0.70 and the top is shared, so no label. Level at the top is off,
    because three options hold the top. Recorded as a disagreement, answers 3

That is the most spread result this task can produce, and it is the case the
flag's condition was chosen against.

A single choice with one answer only:

    task t1, single choice over {clear, marked, unusable}
    answers   marked
    shares    marked 1.00
    1.00 >= 0.70, so label = {marked}, confidence 1.00, answers 1

The label exists at one answer and says so in its input count. Whether one
answer is enough to stop asking is not this rule's question; it is the
retirement rule's, and `retirement.md` is where a count is required beside a
threshold.

A choice of several, inside its bounds:

    task t2, choose 1 to 3 of {star, galaxy, plate_defect, annotation_mark}
    answers   {star, galaxy}
              {star}
              {star, galaxy}
              {star, plate_defect}
              {star, galaxy}
    per option, containing share
              star             5/5 = 1.00
              galaxy           3/5 = 0.60
              plate_defect     1/5 = 0.20
              annotation_mark  0/5 = 0.00
    at or above 0.70: {star}
    size 1 is within the declared 1 to 3
    so label = {star}, confidence 1.00, answers 5

Galaxy at 0.60 does not enter the label and is not lost either: the per option
shares are recomputable from the export, so a reader who wants a looser rule can
apply one.

A choice of several whose derived set breaks the declared bound:

    task t3, choose exactly 1 of {a, b}      lower 1, upper 1
    answers   {a}, {b}, {a, b}
    per option, containing share
              a  2/3 = 0.67
              b  2/3 = 0.67
    at or above 0.70: {} , which is below the declared lower bound of 1
    so no label. Recorded as a disagreement, answers 3

The third answer in that example should not have been accepted at all, because
it breaks the bound the task declared. Refusing it is the campaign definition
boundary's job in #32 and the surface's in #43, and this rule still has to be
defined for the case where a bad answer reached the store, because a rule that
assumes clean input is a rule with an undefined branch.

## What the label carries beside itself

Four things, all written with the label and none of them derivable later.

The confidence, which is the share of answers behind the label. For a single
choice it is that option's share. For a choice of several it is the smallest
share among the options in the label, because a set label is only as well
supported as its weakest member and reporting the average would hide exactly the
option a reader should doubt.

The number of classifications the label was computed from.

The identity of the rule that computed it, including the threshold in force at
the time. A campaign owner may change the threshold mid-campaign, and a label
computed under the old one has to be readable as such rather than silently
compared with labels computed under the new one.

Whether the answers were level at the top. This is the one place that condition
is stated, and the condition is exact: the flag is on where two options hold the
largest share and no other option the task declares was named by anybody. It is
off everywhere else, including where three or more options hold the largest
share, and including where two hold it with a third option named once.

The narrowness is the point and it was decided in #180. A subject with no label
because the answers divided between two readings and nothing else was seen is a
subject the collection is genuinely ambivalent about, and it is usually the
interesting one. A subject with no label because the answers were spread is more
often a broken image, a bad scan or a question two volunteers read differently.
`retirement.md` sends both to the same place for different reasons, so a flag
that fired on both would tell a campaign owner nothing, which is what it exists
to tell them.

Two wider readings were rejected on the same case. Any largest share held by
more than one option fires on three answers naming three different options,
because every option then holds a share of 0.33; so does the reading where the
options at the top account for every answer, because three thirds sum to one.
Both would mark the most spread result a campaign can produce as ambivalence.
The condition above refuses that case and is the only one of the three that
agrees with every case worked by hand here.

An answer naming nothing this task declares does not move the flag. It counts in
the denominator and supports no option, so it lowers every share without putting
an option at the top or beside it, and two options level with such an answer
present are still level. What the condition asks is which options the volunteers
named, not how many of the answers were usable.

There is deliberately no fifth field carrying a probability that the label is
correct. This rule does not produce one, and a share presented as a probability
would be the strongest claim in the project resting on the weakest ground.

## What the export has to carry for somebody to recompute this

The requirement is stronger than reproducing the labels above. A downstream
reader has to be able to apply a different rule, including the weighted and
probabilistic ones rejected above, or this project has quietly decided the
science for them.

For every classification: the campaign, the subject, the task, and the option
identifiers the answer named. That is what the count reads, and it is enough to
recompute every label in this document.

For the campaign definition as it stood: each task with its type, each option
with its identifier, and the declared lower and upper bounds. Without it a
reader has a column of identifiers and no way to know that an option existed and
was never chosen, which is the difference between a zero and a missing value.

For every label: the label, the confidence, the input count, the rule identity
with its threshold, and the level-at-the-top flag. These are what make the
export self-checking: a reader recomputes and compares rather than trusting.

For a rule that weights volunteers: a per volunteer identifier on each
classification, stable within a campaign. This project's own rule reads no such
identifier, so a plain count is recomputable from an export that carries nothing
at all about people. Whether the export carries one anyway is not decided here
and is not assumed in either direction. It is entry 3 of #1, it belongs to the
maintainer, and #15 records what the export ends up carrying.

The file that holds all of this is #68's and its provenance is #69's. Both take
the field list from here rather than assembling a second one.

## What takes its numbers from here

`retirement.md` reads the confidence and the input count defined above and adds
no threshold of its own. #37 implements this rule, with the worked cases above
as its table, and owes the proof that the same answers in any order produce the
same label. #39 may estimate volunteer reliability for a campaign owner to look
at, and until this document is reopened that estimate does not enter a label.
