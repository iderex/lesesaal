# What agreement with the consensus does not measure

Owed by issue #39.

A campaign owner will want a number per volunteer eventually, and the cheapest
one to compute is how often somebody agreed with the consensus. The same figure
is a signal a campaign owner can act on and a score about a person held by
software their own group runs, and nothing in the number separates the two
readings. This document is what the number would mean, written before anything
computes one, so that the first person to want it meets the limits before the
figure rather than after it.

## What decides a label today reads nothing about a person

The two functions that decide what a subject's answers mean and when a subject
is finished take no volunteer identifier and nothing derived from one:

    git grep -n 'func Consensus\|func (r Retirement) Decide' -- internal/campaign
    internal/campaign/consensus.go:177:func Consensus(task Task, answers []Answer, rule Rule) Label {
    internal/campaign/retirement.go:246:func (r Retirement) Decide(labels []Label, classifications int, proposals []Answer) Decision {

`consensus.md` fixes that as a decision rather than as a fact about today's
code, and #37's suite holds the first of the two signatures against a change
that would quietly add an argument to it.

So this project computes no reliability estimate at all, and everything below is
about a number somebody may ask for later rather than one that exists. What
would have to be decided before one could exist is the last section.

## Agreement with the consensus is not accuracy

The measure compares a volunteer against the other volunteers, so what it
reports is conformity. Where the group is right, conformity and accuracy move
together and the number is useful. Where the group is wrong, they move apart,
and the number is at its most confident exactly there.

Six specific ways it comes apart, all of which a campaign owner can meet in a
first campaign.

**A volunteer who is consistently right about a hard class looks unreliable.**
The specialist who recognises the one feature nobody else knows disagrees with
the consensus every time they see it. Under any agreement-based measure that is
indistinguishable from somebody answering carelessly, and it is the volunteer a
campaign owner would least want to discount.

**Subject difficulty is not separated from volunteer skill.** A volunteer who
worked through a run of ambiguous subjects disagrees more than one who saw easy
ones, and neither the count nor the share knows which happened. The standard
treatment that does separate them is the joint estimate `consensus.md` rejects
for the first release, so nothing in this project makes that separation.

**Uncertainty ordering means two volunteers are not measured on the same
collection.** Once the model orders the queue, what a volunteer is shown depends
on when they arrived and what had already been answered, so two numbers computed
the same way are computed over different subjects. That is the ordering doing
its job, and it makes the numbers less comparable, not more.

**Subjects that carry no label contribute nothing.** A tie and a spread both
leave a subject without a consensus to agree with, so those subjects drop out of
every volunteer's number. They are the subjects where a disagreement is most
informative, and they are the ones the measure cannot see.

**At small counts the measure is a step function rather than a proportion.**
`consensus.md` sets out what the threshold does at the counts a real campaign
sees. The same steps reach a per-person number: on a subject with three answers,
one dissent is the difference between agreeing with a label and there being no
label to agree with.

**A number from twenty answers reads like a number from two thousand.** A share
carries no count beside it unless one is put there deliberately, and a volunteer
who classified a handful of subjects will otherwise be compared directly against
somebody who classified for a month.

## What such a number would be worth

Enough to look at, and not enough to conclude with. A volunteer far from the
rest of the group is a reason for a campaign owner to read some of that
volunteer's answers, which may show a misread instruction, a task that needed a
clearer question, or a genuine expert. Each of those is a different action, and
the number distinguishes none of them.

Two things follow for anybody writing a paper from a campaign that ran here.
Whatever agreement figures are reported, they are conformity between volunteers
rather than a measurement against truth, and the number of subjects behind each
one belongs beside it. `consensus.md` states the export requirement that keeps
both possible: the per classification rows are what a reader can compute a
different measure from, so somebody who wants one is not left with this one.

## What is not decided here, and what it blocks

A per-person number needs something that tells one volunteer from another, and
whether this project holds such a thing at all is not settled. Entry 2 of #1
decides whether a volunteer arrives by a link or through an account, entry 3
decides whether anything about a person reaches an export, and #13 is where the
inventory of what may be stored about a volunteer is written once those are
answered. Until then there is nothing to compute a number from, and this
document takes no position on any of the three.

Who may see such a number is a second decision and is not this document's
either. The campaign owner is the plausible reader, the volunteer themselves is
a different question, and everybody is not one of the options: a leaderboard
turns a campaign into a competition and this measure into a ranking of people it
cannot support. #49 builds the owner's surface and is where an enforced answer
would live.

What this document does settle is that a number computed under any of those
answers arrives with the six limits above attached, and that `consensus.md`
stays closed to it: an estimate of reliability does not enter a label until that
document is reopened and says otherwise.
