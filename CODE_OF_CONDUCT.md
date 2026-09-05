# Code of conduct

Written for issue #111. It covers this repository: its issues, its pull
requests, its commit messages and its reviews. Who receives a report and what
they do with it is the section that makes this a process rather than a
statement, and it is below.

This is written here rather than adopted from one of the standard texts. Those
carry their own licence terms, this repository has no licence of its own yet,
and issue #1 is where that is decided. A short document that says what this
project means is worth more here than a longer one whose terms are a second
open question.

## What is expected

Argue with the work. This project settles decisions by writing down what the
options cost, and a disagreement about a rule is usually a disagreement about
the position behind it, which `docs/scope-of-use.md` and `docs/design-point.md`
hold. Take it there.

Say what you measured and what you assumed, and use different words for them. A
contributor who marks a claim as a claim is doing the thing this project asks
for, not admitting a weakness.

Assume the person on the other side has less context than you do and no reason
to be reading carefully. Most sharpness in a review is a sentence that was
compressed rather than a judgement about somebody.

Accept a correction without ceremony, and give one without an audience. If a
change is wrong, say what is wrong with it in its own pull request body.

## What is not acceptable

Personal attacks, demeaning comments, and anything about a person's identity
rather than their work. Sexual attention or imagery. Threats, in any register,
including the joking one.

Publishing somebody's private information, including an address, an employer or
anything about the deployment they run, without their explicit permission.

Sustained disruption: reopening a settled decision without new evidence,
repeating an argument the issue already records, or filing to make a point.

Using this repository to reach the volunteers or the researchers in somebody
else's deployment.

## Reporting

Reports come to me, Nils Lehnen, at the address published on that account,
which is also the account that merges everything here:

    $ gh api users/iderex --jq '{name, email}'
    {"email":"nils.lehnen@proton.me","name":"Nils Lehnen"}

Write what happened, where, and what you want to happen. A link to the comment
or the pull request is enough for the where. You do not have to be the person
it happened to.

What I do, in order. I acknowledge that the report arrived. I read what is
linked and ask the reporter anything that is missing. I decide whether this
document was broken, and say so to the reporter either way. Then I act, which
is one of: nothing, with the reason; asking for an edit; hiding or deleting a
comment; blocking the account from this repository. Which one was chosen is
told to the reporter and, where a comment visibly changed, is recorded in that
thread rather than silently.

A report is not made public and the reporter is not named without their
agreement. There is no promised time here, unlike `SECURITY.md`, which promises
some and says nothing enforces them.

## The gap in that route, stated rather than hidden

One person receives reports and that person is also the only one who can act.
A report about me therefore comes to me, and no route inside this repository
fixes that.

What exists instead is outside this project: GitHub's own abuse reporting,
which reaches somebody other than me and can act on an account regardless of
what any repository says. That is the escalation, it is a weaker answer than a
second person would be, and it is the honest one until there is a second
person. `GOVERNANCE.md` is where the one-person position and what it costs are
recorded.

## This repository is not a deployment's community

This project is self-hosted. Somebody running a campaign has volunteers, and
those volunteers are in that campaign owner's community rather than in this
one. This document does not reach them, I have no view of their deployment and
no way to act inside it, and pretending otherwise would offer a protection that
does not exist.

A campaign owner asking volunteers for their time takes on that responsibility
themselves, including what a volunteer is told before they start, which is
issue #50, and what happens to a volunteer who is treated badly by another one.
Nothing in the software decides any of it.

## Enforcement

Nothing in this repository reads this file. There is no check that judges a
comment and none is possible, so what stands behind this document is me acting
on it and the account-level route named above.
