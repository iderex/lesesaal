# Where the model runs, and what it is allowed to see

Decided in issue #11.

## The shape

Inference runs inside the one process. Training runs inside the same process, on
a trigger the operator pulls. Neither exists in a default deployment, because a
default deployment has no model at all.

That is three statements and the third is the one that matters most. Absence is
not a degraded mode here. It is what an operator gets on the first day, what the
test suite runs in, and what every campaign has to complete in.

The service count implied is one, which is what `deployment-ceiling.md` fixes.
Background work inside the single process is exactly the arrangement that
document describes, and model training is the background work it had in mind.
There is no difference between the two documents to argue.

## The two shapes that were not taken, priced rather than dismissed

Inference in a second process, with a narrow interface between them, buys any
model format including the ones that need a runtime this tree would otherwise
not carry. It costs the service count this project has decided to defend: one
more thing to start, one more thing that can be down, one more thing in the
backup, one more version to track. Refused as a default. If it ever arrives, it
arrives through the five questions in `deployment-ceiling.md` and not as a
setting.

Inference nowhere, with the operator training and scoring by hand between
sessions, is the honest minimum, and it is worth stating that this document does
not refuse it. It is what the absence default plus one import path amounts to: an
operator may produce a model artefact on whatever machine they like, by whatever
means, and hand the file to the deployment. What that costs this project is a
file format and a validation step rather than a runtime, and it is the route by
which a working group with a GPU somewhere else gets to use it without this
project growing a second service. #53 owns the format.

A model reached over the network is refused for a different reason. A default
deployment makes no outbound connection at all, which `federation.md` states and
lists exhaustively. A hosted model API is an outbound connection on every score,
so it is not an inference option here; it would be a federation mode, with the
switch and the disclosure that go with one.

## What the model is given when it scores

Per subject, one at a time or in a batch:

The subject's own media, as the derived size this project already holds, and its
content digest so a proposal can be tied to the exact bytes that produced it.

The campaign's task declarations: each task identifier, its type, its bounds and
the identifier of every option it declares. Identifiers rather than the text a
volunteer reads, because the model is being asked which option applies and not
asked to read German.

Nothing else by default. Subject metadata that arrived with the collection is
not passed unless the campaign owner names the fields, one by one, in the
campaign definition. That metadata is whatever the operator's archive happened
to carry, it can hold anything including a person's name, and a default that
forwards it would make the boundary below depend on somebody else's spreadsheet.

## What the model returns

Per subject and per task:

A number per declared option. For a single choice they are read as a
distribution over the options; for a choice of several, each option's number is
read on its own.

One uncertainty number. What that number means, and what it may not be read as,
is #55's and this document does not define it beyond requiring that exactly one
accompanies each proposal.

The identity of the model and its version, and the time the score was produced.
A proposal that cannot say which model made it cannot be excluded later when
that model turns out to have been wrong.

No free text, no explanation, no confidence about a volunteer, and no
recommendation about retirement. A proposal is a set of numbers about a plate.

## What the model is given when it trains

Retired subjects that reached a label, each as its media, its content digest, and
its label per task with that label's confidence and input count from
`consensus.md`.

Nothing about the volunteers who produced those labels. Not an identifier, not
how many of them there were beyond the input count that travels with the label,
not who answered what, not when they answered, not how long they took and not
which of them agreed with which.

That is the whole of the training input and it is a short list on purpose. A
label is already the aggregate of what people said, so training on labels rather
than on individual answers gives the model the campaign's conclusion without
giving it anybody's contribution.

## What the model is never given

Anything about a volunteer. Their identifier, their answers, their timing, their
session, their device, their address, and the fact that a particular person
answered a particular subject at all.

Individual answers, aggregated or not, beyond the label and the two numbers that
travel with it.

Operator or campaign owner free text, other than the option identifiers above
and the metadata fields a campaign owner has named explicitly.

Anything from another campaign on the same host, unless the operator has asked
for that model to be trained across campaigns, which is a decision with its own
consequences and is not assumed here.

This list is what makes the sentence in the data protection documentation, #105,
a short one: the model sits behind the same boundary as everything else and is
not a route by which volunteer data reaches a place the rest of the system does
not put it.

## Scoring never happens while a volunteer waits

Scores are produced in the background and stored beside the subject. The
volunteer's request reads a stored score and never calls a model.

That is not only a latency decision. `subject-selection.md` draws one time in
four from the quarter of the eligible set the model is least certain about, and
cutting a quarter out of a set requires a score for every member of it. A design
that scored on demand could not answer that question at all, so batch scoring is
what the selection rule already assumes.

A scoring run happens after ingest, for the subjects that arrived, and after
training, for every subject in the campaign. Both are the operator's to trigger
or to schedule, and #57 owns the trigger.

## Absent, slow, failed, wrong

Absent is the default and costs nothing. With no model there are no proposals,
`subject-selection.md`'s uncertain quarter is undefined and every draw is the
uniform one, the retirement rule that may consult a model is unavailable rather
than silently satisfied, and `model-visibility.md`'s per classification field
records that no proposal existed. A campaign runs from the first minute to the
export with no model, which is #41's property and #60's cold start.

Slow costs a stale score and nothing else. A scoring run that does not finish
leaves the previous scores in place, and a subject with no score yet is simply
not in the uncertain quarter. Nothing waits on it, because nothing in a
volunteer's request path calls it.

Failed is recorded and left in the last good state. A run that errors does not
half update the scores, the failure is one of the decisions the software records
under #87, and repeated failure is a thing the operator is told about through
#85 rather than a thing that quietly stops happening.

Wrong is the case the model milestone is built around, and it is not this
document's to solve. A model earns the right to propose by being evaluated
first, #58. Drift between the model and the volunteers is noticed, #61. And a
proposal never becomes an answer under any circumstances, #59, which is the
property that makes the other three recoverable.

## What each half assumes about the host

Inference assumes the design point host in `design-point.md`, two cores and 4 GB,
with no accelerator. That constrains the model formats this project can carry to
ones that run on a CPU in-process, and the constraint is accepted rather than
regretted: a default deployment that needs a graphics card is not a default
deployment for the working group this is for.

Training assumes the same host and takes as long as it takes. At the design
point the training set is at most 2,000 images, it runs when the operator asks,
and it competes with page serving for two cores, which is a real cost and the
reason the trigger is the operator's rather than a timer's. A training run that
would not finish on the assumed host is a model this project cannot carry
in-process, and the import route above is where such a model belongs.

## What takes its shape from here

#53 builds the interface above and makes absence the default. #54 supplies the
first model that trains on the operator's own machine within those assumptions.
#55 defines the uncertainty number. #62 records what was proposed, by which
model and when, from the fields listed above. #63 builds the harness that tests
all of it without a model present.
