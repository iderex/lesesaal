# Designing a campaign that produces usable labels

Owed by issue #113. The last section says what it does not yet cover.

The most common way a citizen science campaign fails is not technical. The
question was ambiguous, the categories overlapped, or the number of answers per
subject was picked because it looked reasonable. None of that produces an error
message. It produces a finished campaign whose labels nobody can use, and it is
discovered at the end, by the person who has to write the paper.

This project cannot prevent any of it. What it can do is tell you, before you
open a campaign, which decisions are the expensive ones and what each one costs.
Everything below is a consequence of a rule this project has already fixed, and
every rule is named so you can read it yourself rather than take this on trust.

## The question is read a thousand times

A campaign of 2,000 subjects at five answers each is ten thousand readings of
one sentence. Every ambiguity in it is paid ten thousand times, and the cost is
not confusion; it is two volunteers answering the same subject differently for a
reason that has nothing to do with the subject.

Write one decision per task. A question that asks whether the plate is clear and
whether anything is written on it is two tasks, and a campaign is an ordered
list of them, so splitting costs nothing. Asked together, an answer of no tells
you neither.

Write what the volunteer should look at, not why you want to know. A volunteer
who has been told what the campaign is for will helpfully answer the question
they think you meant.

Test the wording on the first ten answers rather than on your colleagues. You
may rewrite the text of a question or an option while a campaign is running:
`task-grammar.md` separates an option's identifier from the text a volunteer
reads, precisely so that the edit every campaign owner makes in the first hour
is free. What you may not do is change an identifier or remove an option that
already has answers behind it. That is a different campaign and is refused
rather than accommodated, so decide the list of options with more care than the
words for them.

## Categories, and why a category called other collects half of everything

A task declares a fixed list of options and an answer is a subset of that list.
There is no order between options: a fixed list has no order this project knows
about, so nothing here treats the second option as nearer to the first than the
fifth is. If your categories are really a scale and being one step out matters,
you are asking for a number, and `task-grammar.md` says why that is refused and
what would bring it in.

An option called other, or miscellaneous, or unusual, ends up holding three
different things that cannot afterwards be told apart: the subject your list
forgot, the subject the volunteer could not decide about, and the broken scan.
Those need three different actions from you and the column gives you one.

What to do instead, in order. Name the cases you expect, including the boring
ones, because an option nobody ever chooses costs nothing and tells a reader it
existed. Add one option that means the volunteer cannot tell from this image and
means nothing else, and say so in its text. Then read the stuck list while the
campaign runs, because a subject that reaches no agreement is the campaign
telling you which case your list is missing, and it tells you early enough to
open a second campaign that has it.

A yes or no question is a single choice over two options named yes and no. It is
not a type of its own and nothing in the software knows those words.

## How many answers a subject actually buys

Retirement is a campaign setting chosen before the campaign opens.
`retirement.md` states the four rules a campaign owner may choose in the words
of a campaign rather than of statistics, and this guide deliberately does not
write a fifth version of them. Read them there and come back for the numbers.

The number that surprises people is not the floor. It is what the agreement
threshold does at the counts a real campaign sees. `consensus.md` sets out the
smallest agreeing count that clears the default threshold of 0.7, and the third
row is the one to look at: at three answers, 0.7 is unanimity, because two of
three is 0.67. A campaign that retires at a floor of three and reports fewer
labels than expected has usually met that row rather than a disagreeable set of
volunteers.

Two consequences follow for the numbers you choose.

A floor is a claim about how much evidence a label needs, and at small floors
the threshold turns it into a claim about unanimity. The suggested numbers are a
floor of 3 and a ceiling of 7, and at the default threshold that means the first
three volunteers agreed, or five of seven did. If you want a label from a
majority rather than from near-unanimity, lower the threshold deliberately and
write down that you did, rather than raising the floor and hoping.

A ceiling is a claim about when to stop paying. A subject that has not settled
by the ceiling retires unresolved, and that is a result rather than a failure:
`retirement.md` keeps the two kinds of unresolved apart, and the one where the
answers were level between two readings is usually the scientifically
interesting subject in the collection. Budget for reading that list.

Both numbers are starting positions rather than results. Nothing in this project
has measured them, and what would move them is the pilot in #115 and #116.

## Known subjects are not decided yet

Subjects whose answer you already know are the standard instrument for asking
whether a campaign is working, and how many are needed and what they are allowed
to decide is #38. It is open, and this guide takes no position on it. Until it
is decided, treat any subject you happen to know the answer to as something you
check by hand at the end rather than as something the software will act on.

## What uncertainty ordering does to the sample you collect

Once a model is running, one draw in four is taken from the quarter of the
eligible set the model is least certain about, and three draws in four ignore
the model entirely. `subject-selection.md` is where that mixture and its reasons
live.

What that does to your collected sample depends entirely on whether the campaign
finished, and the two cases are easy to conflate.

A campaign that ran to completion has every subject retired, so the ordering
changed the order in which subjects were seen and not which subjects were seen.
The membership of the collected set is the whole collection either way.

A campaign stopped early has not. The subjects answered are enriched for the
ones the model found hard, in a proportion nobody has measured, and treating
that set as a random sample of the collection is the mistake this section
exists for. If you stop early, say so and say what fraction was answered.

The uneven part survives completion, and it is the evidence per subject rather
than the membership. Where the model shortening rule is switched on, a subject
whose proposal agreed with the volunteers may retire at a reduced floor, so the
finished campaign carries some subjects with five human answers and some with
two. `retirement.md` says why that rule is off unless you switch it on, and that
every subject retired that way is marked in the export.

How to report it honestly, in a paper, in three sentences you can write from the
export alone. State that subject order was mixed, giving the share drawn by
uncertainty and the model version in force. State whether the campaign ran to
completion, and the fraction answered if it did not. State whether the model
shortening rule was on, and that the subjects it shortened are marked, so a
reader who wants the uniform corpus can drop them and re-derive your labels.

That last property is the point of the whole arrangement: the model changed how
much evidence was collected and never what the evidence says, so every claim
above can be checked by somebody holding only the export.

## What a number about a volunteer would not tell you

You will want to know whether a particular volunteer is reliable. Before
computing anything, read `reliability.md`. It sets out what agreement with the
consensus does and does not measure, and the short version is that it measures
conformity with the rest of the group, so the specialist who is right about the
hard class is the volunteer it marks lowest.

## What this guide does not cover yet

There is no worked example end to end. #113 asks for one built from the pilot's
own campaign, and the pilot is #115 and #116 and has not run. An invented
example would be the thing this guide warns against, a set of numbers chosen
because they looked reasonable, presented as experience.

Known subjects are named above and not explained, because #38 has not decided
what they are allowed to decide.

Nobody outside this project has designed a campaign from this document. #113's
last condition is that a researcher who has not used this project can, shown by
asking one, and that has not been done. Until it has, this guide is a claim
about what a researcher needs rather than a measurement of it.
