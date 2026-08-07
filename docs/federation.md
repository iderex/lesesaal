# Federation, and what a default deployment is allowed to send

Decided in issue #16.

The legal position of this project is that personal data never leaves the host
unless the operator deliberately federates. That sentence is worth something only
if federation is a defined thing with a switch, rather than a word covering
whatever outbound traffic happens to exist. This document defines it and lists
what a default deployment may send.

## Federation is four modes, and they are separate

Each is switched on by itself. Switching on one never implies another, and there
is no setting that means federation generally.

Pooling subjects across hosts, so several groups classify one collection. What
travels is subject media and subject metadata, outbound and inbound. Whether the
classifications themselves travel is a property of this mode and the answer is
that they do not: pooling shares the work, not the results. No personal data,
provided the subject media itself carries none, which is the operator's
responsibility and belongs in the operator's data protection documentation, #105.

Publishing a campaign's results somewhere central. What travels is the export.
Whether that carries personal data is not decided here and this document does not
assume it: it depends on what an export says about a volunteer, which is entry 3
of #1 and belongs to the maintainer. Until that entry is answered, this mode is
marked as carrying personal data, because the safe reading of an undecided
question is the one that does not quietly widen what may leave the host.

Sharing a trained model. What travels is the weights and whatever the model
integration records alongside them. Whether a model derived from volunteer labels
carries personal data has no settled answer anywhere, and entry 4 of #1 holds
whether such a model may be published at all. Marked as carrying personal data
for the same reason as above.

Announcing that a campaign exists, so volunteers can find it. What travels is the
campaign's own description and the address it is reachable at. No personal data
about volunteers. It does say publicly that this operator is running this
campaign, which is information about the operator rather than about a volunteer,
and an operator should be told that in the sentence that offers the switch.

## What a default deployment may send outbound

Nothing.

This is the list, and it is intended to be exhaustive. A default deployment makes
no outbound connection at all: not on start-up, not on first run, not on a
schedule, and not while serving a page.

Each of the usual exceptions is named and refused, because they arrive through a
dependency rather than through a decision, which is why the list has to be
written before the dependencies are chosen rather than after.

No certificate issuance. Certificates reach the deployment by the route #77
decides, and the default deployment does not talk to a certificate authority
itself.

No software update check. Not at start-up, not in the background, not once a
week. An operator learns about a new release from the release, #118 through #123.

No font, script, stylesheet, icon or map tile from a content network. The
volunteer surface loads only from the operator's host, which is #44, and this is
the same rule stated from the other end.

No error reporting, crash reporting or telemetry service. Errors go to the
operator's own log, #83.

No analytics of any kind, including a self-hosted one that is not on this host.

No model download at run time. A model that needs weights it does not have does
not fetch them; it says it has none, which is the absence default in #53.

No package or dependency resolution at run time. Everything the deployment needs
is in the image, #74.

No time synchronisation, no telemetry endpoint, and no health reporting to
anywhere but the operator's own health endpoint, #85.

Any dependency that would breach this list is a reason to reject that dependency,
and the reason is written where the dependency is argued. A dependency that
phones home for any of the reasons above is not made acceptable by being popular,
by being configurable, or by being off in the version that was reviewed. The
deployment issues carry this as a constraint on what may be added: #74 for the
image, #82 for what configuration may contain, and #109 for what the bill of
materials has to make visible.

## How federation is switched on, and why not by accident

Three things together, and all three are required.

The configuration names the mode explicitly, one setting per mode, and the value
is the mode's name rather than true. A configuration file with a single boolean
called federation is exactly the accident this is written to prevent.

Each setting names the destination it federates with. A mode switched on with no
destination is a configuration error and the deployment refuses to start, which
is the fail closed rule in #82. There is no default destination and this project
ships no address for one.

The start-up self check prints, every time, which modes are on and where each one
sends. An operator who did not intend to federate finds out on the next restart
rather than from somebody else, and an operator who did intended it and can read
the confirmation. #85 owes that line.

There is no user interface control that switches federation on. It is a
configuration change on the host, made by the person who runs the host.

## What this document is for

The sovereignty statement, #103, states the position in the words an operator
reads. The test that proves nothing leaves the host, #104, refuses a deployment
that makes a connection this list does not permit. Both take the permitted list
from here rather than restating it, so that widening what may leave the host is
one change in one file and is visible as one.
