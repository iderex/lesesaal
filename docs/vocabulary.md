# The words this project uses, and the ones it refuses

Decided in issue #5.

## Why one meaning per word

The vocabulary of this field is borrowed from several platforms that mean
slightly different things by the same word. A plan that uses them loosely
produces a schema that has to be renamed later, and a rename after an export
exists is a migration for everybody downstream.

The names below go into the store, the surface and the export. They are close to
permanent, so each one carries the alternatives it is deliberately not using.
Where a word appears in an issue, a document or a column heading, it means what
it means here and nothing else.

## The words

A campaign is one collection of subjects, one ordered list of tasks and the
volunteers invited to it, run from opening to end. Not a project, which is what
one platform calls this and another calls the deployment. Not a workflow, which
is used below in a narrower sense by the platform that coined it and would
collide with the task list.

A subject is one thing being classified, together with the metadata it arrived
with. In the first release that is one image, which is what `task-grammar.md`
assumes throughout. Not an item, an asset, a target or a record.

A task is one question a campaign asks about every subject, with the fixed list
of options it declares. Not a question, which reads as the sentence rather than
the whole declaration, and not a step or a workflow.

An option is one of the choices a task declares. It carries a stable identifier
and the text a volunteer reads, and those are separate fields for the reason
`task-grammar.md` gives. Not a label, which is taken below, and not a class or a
category.

An answer is what one volunteer says about one task for one subject. It is a set
of option identifiers, of size one where the task admits one choice. Not a vote,
which suggests the aggregation is counting and prejudges #7. Not an annotation,
which in the neighbouring platforms means a mark placed on an image, a shape this
grammar refuses. Not a response.

A classification is one volunteer's submission about one subject: exactly one
answer per task in the campaign, recorded together. It is the unit that arrives,
the unit a rate is counted in, and the unit `design-point.md` sizes the system
by. Not an annotation and not a judgement.

The classifications of a subject are every classification recorded about it. For
one task, the answers inside them are that task's answer set for that subject,
which is what the consensus rule reads. Neither has a shorter name on purpose:
a coined word here would be a word only this project knows.

A label is the single derived answer for one subject and one task, computed from
that answer set by the consensus rule. It has the same shape as an answer, a set
of option identifiers, because it answers the same task. Not a consensus, which
is the rule rather than its output. Not an aggregation, a result, a gold standard
or a truth, and the last two for a stronger reason: they claim a correctness this
project cannot establish.

A subject is retired when the campaign's rule says it needs no more
classifications. Retired rather than done, complete or finished, because those
words are also wanted for the campaign as a whole, and a subject that stopped
collecting answers is not necessarily a subject anybody is finished with.

A proposal is what the model says about a subject before or beside the
volunteers. It is never an answer and never a label, and `model-visibility.md`
holds what a volunteer may see of one. Not a prediction, which reads as a claim
about the future rather than about a plate.

A volunteer is a person who classifies. Not a user, which is everybody, and not
an annotator or a worker, which describe paid work this project is not about.

A campaign owner is the researcher who defines a campaign and reads its results.
An operator is whoever runs the deployment. They are frequently the same person
and are never the same role: one answers for the science, the other for the
machine.

Ingest is copying a collection into the deployment and making subjects out of
it. Export is writing a campaign's classifications and labels out as files
somebody else can read. Neither word is used for the other direction.

## Whether a campaign carries one task or several

Several, ordered. This is not decided here: `task-grammar.md` already fixes that
a campaign is an ordered list of tasks, and this document takes it from there
rather than restating it as a fresh choice.

What is decided here is the consequence that made the question worth asking. A
classification carries one answer per task in the campaign, so a volunteer
looking at a subject answers all of its tasks or submits nothing. That removes
partial completion as a state: there is no subject with two of its three tasks
answered by somebody who stopped, and no retirement rule that has to be defined
per task and then reconciled across tasks.

The surface may still present the tasks one at a time, which is what #43
describes, and the campaigns this project is built for will usually declare one
task. Both are true at once, because the presentation and the recorded unit are
different questions.

The cost is stated rather than hidden. A volunteer who can answer the first task
and not the second gives nothing rather than something, and a campaign owner who
wants the partial answer has to declare the tasks as two campaigns. Retirement
per subject then means a subject retires when the rule is satisfied for every
task, which is #9's to state and is written there.

## Whether a label is stored or recomputed

Stored, recomputed on every arrival, and carrying what it was computed from.

The label is written whenever a classification for that subject arrives, and it
records the identity of the consensus rule that produced it and the number of
answers behind it. That makes it a value with its inputs attached rather than a
value alone.

Recomputing on read was the alternative and it fails on the thing that reads it
most. The retirement rule consults the label and its confidence, and it consults
them on the write path, at the moment an answer arrives; so a design that
recomputes on read still computes on write. It also puts the rule in front of
every reader, including the campaign owner's progress view, which is #49.

Storing it costs staleness, and staleness is handled by making it detectable
rather than by assuming it away. A label whose recorded input count differs from
the number of classifications now behind it, or whose recorded rule differs from
the campaign's, is stale by inspection rather than by suspicion. What a
deployment does with a stale label is #37's, and the export carries both the
label and its inputs so that a reader outside this software can recompute it and
compare, which is #7's requirement and #68's file.

## The entities and their cardinalities

Read `1..n` as one or more, `0..n` as none or more.

    campaign      1 --- 1..n  subject
    campaign      1 --- 1..n  task            ordered, at least one
    task          1 --- 2..n  option          at least two to choose between
    campaign      1 --- 1..n  volunteer       invited to it
    subject       1 --- 0..n  classification
    volunteer     1 --- 0..n  classification
    classification 1 --- 1..n answer          exactly one per task
    answer        1 --- 0..n  option          the identifiers chosen
    subject+task  1 --- 0..1  label           none until the first answer
    subject       1 --- 1     retirement state
    subject+task  1 --- 0..n  proposal        by model and time

Four of these carry the whole shape of the system, so they are written out.

A classification names exactly one volunteer and exactly one subject, and the
pair occurs at most once: a volunteer classifies a subject once and never again,
which is what makes the answers about a subject independent and is enforced at
the draw in `subject-selection.md`.

A label belongs to a subject and a task together rather than to a subject alone.
A campaign with three tasks produces three labels per subject and one retirement
state for the subject.

An answer belongs to a classification and a task, so the set of answers for one
subject and one task is reached through the classifications of that subject.
That indirection is the price of the classification being the unit that arrives.

A proposal belongs to a subject and a task and there may be several over time,
because a retrained model proposes again. Which one was in force when an answer
was recorded is #62's, and nothing overwrites an earlier proposal.

## What this document does not fix

The concrete syntax a campaign owner writes, which follows the toolchain in #17.
The store's tables and columns, which are #33's and take their names from here.
The export's file names and column headings, which are #68's and do the same. A
name introduced by any of those that is not in this document is either a synonym
to be removed or a new word this document is missing, and either way the repair
is here rather than there.
