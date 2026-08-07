# What this is for, and what it is not built for

Decided in issue #110. This is a scope statement. It is not a licence
restriction, and the section at the end says what that difference means.

## What it is for

Research groups and community projects that classify their own collections.

The shape it is built around is a working group with a collection it holds and
volunteers it can name: a few thousand subjects, ten or so people, one machine
the group controls. Photographic plate archives, specimen sheets, survey imagery,
field photographs, scanned instrument output. Collections where the group holding
them wants labels and has nowhere sensible to get them.

Everything in the design follows from that. The volunteer surface asks one
question about one subject. The consensus rule combines several people's answers
because that is how a group of ten gets a label it can defend. The model exists
to spend volunteer attention where it is worth most, and it never answers on a
volunteer's behalf.

## What it is not built for

Surveillance material. Camera footage of people, images collected to watch a
place or a group, and anything gathered for the purpose of knowing where somebody
was.

Grading, ranking or scoring people. Not volunteers and not the people in the
images. This project estimates how reliable a volunteer's answers are, because a
consensus rule needs it, and that estimate exists to weight an answer rather than
to rank a person. It is not a performance measure, it is not shown as a
leaderboard, and it is not an output.

Anything where the subjects are identifiable people who did not agree to be
looked at. A queue of faces, however the collection was assembled and whatever
the labelling task is.

Producing training data whose purpose is any of the above. The output of a
campaign is a labelled collection and a model, and both are more portable than
the campaign that made them. A collection that would be refused as a campaign is
refused as a source of training data too.

These are not edge cases the design overlooked. They are things the design does
well, which is why they are written down: a queue of images, a group of people
answering a question about each one, and a model that learns from the answers is
a useful shape for work this project does not want to be part of.

## Nothing enforces this

The software cannot tell one image from another. It does not look at what a
subject is, it has no list of forbidden collections, it asks no one for a purpose
and it reports nothing anywhere. There is no check, no term in a configuration
file, no key, and nothing that phones home. A deployment used for any of the
things above works exactly as well as one used as intended, and this project
would never know.

Saying otherwise would be the kind of claim this project does not make. A
statement is not a mechanism, and this one is a statement.

What it does do is set what the maintainers will help with. A feature request
that only makes sense for one of the uses above is refused with a pointer at this
document, and that is the whole of its force.

## Why this is not a licence restriction

A licence restriction changes what somebody is legally permitted to do with the
source. That is a different decision with different consequences, it is not made
here, and this document does not make it by implication.

So this document adds no condition to whatever licence this project is released
under and removes none. Somebody who does one of the things above is not
breaching a licence term, because there is no such term. They are outside what
this was built for, which is a statement about the project and not about their
rights.

Which licence this project carries is not settled. Issue #1 holds it.
