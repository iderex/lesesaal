# When a subject stops being handed out

Decided in issue #9.

## What this rule owns and what it borrows

Retirement decides when a subject has had enough classifications. It owns counts
and nothing else.

Every share, every threshold and every statement about whether volunteers agree
comes from `consensus.md`. A rule below that wants to know whether the
volunteers agree reads the label's confidence and its input count, which that
document defines and writes with the label. No rule here declares a second
threshold, and a campaign owner who changes the agreement threshold changes it
in one place.

A campaign with several tasks retires a subject when its rule is satisfied for
every task, which follows from the classification carrying one answer per task
in `vocabulary.md`. There is no per task retirement and no subject that is half
retired.

## The rules a campaign owner may choose

Each is written the way a campaign owner would say it, with the numbers it takes
from elsewhere named explicitly.

Ask everybody the same number of times. Every subject is handed out until it has
a fixed number of classifications, and agreement is not consulted at all. Its own
number is that count. It borrows nothing from `consensus.md`. It is the rule to
choose when the campaign is measuring the volunteers as much as the plates,
because every subject then carries the same weight of evidence.

Stop as soon as the volunteers agree, and give up after a while. A subject
retires when it has at least a floor of classifications and every task has a
label, and it retires anyway once it reaches a ceiling. Its own numbers are the
floor and the ceiling. Whether a label exists is `consensus.md`'s answer, taken
whole. This is the default, because it spends volunteer time on the plates that
need it.

Stop early when enough volunteers say there is nothing here. The campaign owner
names one option per task as the empty answer, and a subject whose empty answer
reaches a label with a smaller floor retires there. Its own number is that
smaller floor. The agreement it requires is `consensus.md`'s, unchanged. It is
for a collection where most plates carry nothing of interest, which is the
common case in an archive sweep and is where a campaign's time is otherwise
spent confirming absences.

Let the model shorten the easy ones. On top of the default rule, a subject whose
model proposal agrees with the label already derived from the volunteers may
retire at a reduced floor. Its own number is that reduced floor, which may never
go below the absolute floor in the next section. It is off unless a campaign
owner switches it on, and the section after next says why.

The suggested starting numbers are a floor of 3, a ceiling of 7 and an empty
answer floor of 3. At the default agreement threshold of 0.7, a floor of 3 means
the first three volunteers were unanimous, and a ceiling of 7 means a subject
that has not settled by the fifth agreeing answer out of seven is not going to.
The design point assumes 5 classifications per subject on average, which sits
inside that range, and all of these are starting positions rather than results:
the pilot in #116 is what moves them.

## Where a subject goes when it cannot reach agreement

It retires unresolved. Retired, so it stops costing volunteer time, and marked,
so nobody reads the absence of a label as an oversight.

Unresolved is not one thing, and the two cases are kept apart because they mean
different things to a campaign owner. A subject whose answers were level at the
top is a subject the collection is genuinely ambivalent about, and it is usually
the interesting one. A subject whose answers were spread with no option near the
threshold is more often a broken image, a bad scan or a question the volunteers
read differently from each other. `consensus.md` records which of the two
happened, and this rule carries the flag through to retirement rather than
recomputing it.

Both appear in the campaign owner's progress view as the stuck list, which is
#49, and both appear in the export as a retired subject with no label, its
answer counts and its flag, which is #68. A campaign owner who reads the stuck
list and rewrites the question has learned the most useful thing a campaign can
tell them, and the redefinition is a new campaign rather than an edit to this
one, for the reason `task-grammar.md` gives about changing options under
existing answers.

Nothing is deleted and no subject is quietly dropped. A subject that no
volunteer could classify still carries its classifications into the export,
because the answers people gave are evidence about the subject even when they do
not add up to a label.

## Whether retirement is reversible

Only by the campaign owner, only as an explicit action, and never on its own.

A retired subject stores the rule and the numbers it retired under. Changing the
campaign's rule mid-campaign changes what happens to subjects from that moment
forward and leaves the already retired ones alone. The campaign owner may then
ask, in one action, for the new rule to be applied to the subjects already
retired, and the ones it no longer satisfies return to the eligible set. That
action is recorded with its time and appears in the export.

The alternative, recomputing retirement continuously so that tightening the rule
silently un-retires subjects, was rejected for what it does to work already
done. A campaign owner who has exported data would find the export no longer
describes the campaign, with nothing marking the moment it changed. Retirement
that moves only when somebody asks for it keeps the export and the campaign
telling the same story.

Returning a subject costs nothing to the volunteers who already answered it.
They are excluded from its eligible set for good, which `subject-selection.md`
holds, so a returned subject goes to people who have not seen it and the
independence of its answers survives the round trip.

## Whether the rule may consult the model

It may, under one rule, off by default, and never in a way that puts the model
into a label.

The permitted shape is the fourth rule above. The model's proposal may reduce
how many volunteers are asked about a subject whose volunteers already agree.
The label itself is still computed by `consensus.md` from human answers alone,
its confidence is still the share of those answers, and its input count is still
the number of people who answered. So a subject retired this way carries a true
account of the evidence behind it, and what the model changed is how much
evidence was collected rather than what it says.

Three constraints hold it there. There is an absolute floor of classifications
below which no model may retire a subject, and it is the campaign's own floor
rather than a number this rule invents. The model may only shorten a subject
whose humans already agree, so a proposal never breaks a tie and never rescues a
disagreement. And every subject retired this way is marked as such in the
export, so a downstream reader can drop them and re-derive the campaign as if
the model had not been there.

It is off by default because a campaign whose first release produces a mixed
corpus, some subjects with five human answers and some with two, is harder to
analyse than one that produces a uniform one, and a campaign owner switching it
on should be doing so knowingly. `model-visibility.md` already anticipated this
route, and this document is where the conditions on it are written.

What is refused outright is the model contributing an answer, standing in for a
volunteer, or its proposal entering the count that produces a label. That is
#59's property and this rule does not weaken it. It is also why the model can
never retire a subject the volunteers disagree about: the case where a model
would help most is exactly the case where using it would be inventing evidence.

The model milestone assumes this answer. #36 applies the rule, #38 supplies the
reduced floor's condition, #58 is where a model has to earn the right to be
consulted at all before an operator switches this on, and #61 is what notices
when a model that had earned it stops deserving it.

## What a campaign owner is told before choosing

The choice is a campaign setting made before the campaign opens, and every rule
above is stated in the definition file the owner writes, which is
`task-grammar.md`'s file. The guide in #113 explains the rules in the words of a
campaign rather than of statistics, and it takes the four sentences above rather
than writing a fifth version of them.
