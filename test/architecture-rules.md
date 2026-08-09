# The architecture rules, and what refuses each one

Written for issue #97. Several rules in this plan are about which part of the
code may know about which other part, and about what a part may reach outside
its own source. Those are the rules a well meant change breaks, because
breaking one is convenient and nothing complains until much later, so each one
that can be is written as a test rather than as a sentence somebody remembers.

This is the one place they are listed. Each entry says what the rule is, which
test refuses it, the failure that test names, and the smallest plausible change
made to watch it go red. A rule with no test is here as well, with the reason
and the issue that would bring one, because a list of only the guarded rules
would read as a claim that the rest do not exist.

## What is on this list

A rule about which part may know about which other part, or about what a part
may reach outside this repository's own source. `docs/layout.md` and
`docs/harness.md` are where those are decided, and this file adds nothing to
them.

Not the checks the gate runs and what each has been observed to refuse, which
is `test/gate-refusals.md` beside this file. Not a rule about what a document
says, what a commit message carries or how large a change may be, because
nothing in this tree reads any of those. Not the rules about what the software
decides, which are the campaign rules and belong to the issues that build them.

This file is kept level with those two documents by whoever changes either.
Nothing compares them, and that is a gap rather than a design: the class of
check that could do it is the greppable invariant in #94, and a rule that this
file names every test the guards declare is a candidate for it.

## The rules a test refuses

Every entry below was watched to red on this machine by making the change it
names and running the suite. Where a run on the server has also watched it, the
run is linked; where none has, the entry says so rather than leaving it to be
assumed.

### Every directory the layout names exists

`docs/layout.md` names the parts of the tree, and every rule below about a part
is decided by which directory a file sits in. A renamed directory would leave
those rules matching nothing, and a run that judged nothing prints the same word
as a run that judged everything.

Refused by `TestEveryDirectoryTheLayoutNamesExists` in `layout_test.go`.

The near miss is `deploy` renamed to `deployment`, which is the shape of a
tidying nobody would think to check anything about.

    layout_test.go:68: the layout names deploy and the tree does not carry it: ...

What follows the second colon is the operating system's own sentence for a path
that is not there, in the language that machine is set to, so it is cut here
rather than pasted into a tree that writes in English.

Not watched by a run.

### A dependency points the way the layout says

Everything points inward, at the core. The core imports nothing, the surface
may not import the store, and the model may not import either the store or the
surface. `docs/layout.md` carries the whole table and the reason for each
refused direction.

Refused by `TestDependenciesPointTheWayTheLayoutSays` in `layout_test.go`.

The near miss is the core given an import of the surface, which is the one the
whole layout exists against.

    layout_test.go:143: internal/campaign/depends.go imports github.com/iderex/lesesaal/internal/web, and docs/layout.md does not let internal/campaign depend on internal/web

Watched by a run, on the branch taken for the unit test check:

    https://github.com/iderex/lesesaal/actions/runs/31296264233/job/93201734635

### Only the wiring reaches the runtime

Time, randomness, identifiers and outbound connections are taken as
dependencies rather than read out of the runtime, and `internal/system` is the
one directory that may read them. `docs/harness.md` argues it and carries the
table the guard judges against.

Refused by `TestOnlyTheWiringReadsTheRuntime` in `harness_test.go`.

The near miss is the core asking the wall clock for the moment, which is what
somebody writes who has not yet met the injected field.

    harness_test.go:153: internal/campaign/depends.go:59 calls time.Now, which reaches the runtime and is refused outside internal/system. Take the Now field of campaign.Depends instead.

Watched by a run, for the neighbouring case of a unit test reaching the network
rather than the clock:

    https://github.com/iderex/lesesaal/actions/runs/31296348917/job/93201936767

### Nothing sleeps, including a test

The same test carries a second rule with a different scope. A sleep is refused
in every file in the tree, the wiring and the tests included. In a test it is a
statement that the author could not express what they were waiting for, and it
is the commonest single cause of a suite that fails once a week for no reason
anybody can reproduce.

The near miss is ten milliseconds in front of an assertion, which is the size
nobody argues with.

    harness_test.go:153: internal/campaign/depends_test.go:14 calls time.Sleep, which reaches the runtime and is refused in a test. Take a clock the caller can move, or the thing you are actually waiting for instead.

Not watched by a run.

### Nothing starts a subprocess outside the wiring

`docs/layout.md` names three things the unit suite may not have, no subprocess,
no network and no dependency. The middle one was refused and the other two were
not. This rule is the subprocess half, and it landed with #97 rather than with
the guard it sits in.

The near miss is a test asking git for the commit it is running at, which is
what somebody writes who wants a test to know where it is.

    harness_test.go:153: internal/campaign/depends_test.go:15 calls exec.Command, which reaches the runtime and is refused outside internal/system. Take the thing the command would have done, written in this language instead.

Not watched by a run.

### Only a test file imports the fakes

`internal/campaign/campaigntest` holds a clock a test moves by hand. A fake
clock reaching the binary an operator runs would be a deployment whose time
does not move, and nothing else in this tree would notice.

Refused by `TestOnlyATestImportsTheFakes` in `harness_test.go`.

The near miss is the entry point reading the fixture epoch, which is what is
left behind after somebody debugs a start-up problem against a fixed clock.

    harness_test.go:195: main.go imports github.com/iderex/lesesaal/internal/campaign/campaigntest and is not a test file. The fakes are for tests; a program that carries them carries a clock that does not move.

Not watched by a run.

### No import is made with a dot

Every rule in `harness_test.go` is decided by the name a package is imported
under. A dot import puts identifiers into a file with no qualifier at all, so a
file carrying one walks through that whole table unseen.

Refused by `TestNoDotImport` in `harness_test.go`.

The near miss is a test dropping the qualifier on the fakes, which is the one
place in this tree where a dot import would genuinely read better.

    harness_test.go:217: wiring_test.go imports "github.com/iderex/lesesaal/internal/campaign/campaigntest" with a dot, and every rule in this file is decided by the name a package is imported under.

Not watched by a run.

### The wiring supplies every dependency the program takes

A nil field is a dependency nobody supplied, and a program that starts on one
panics at the first call rather than at start-up, which is the wrong end of a
run to find out. `main.go` refuses to start on an incomplete set, and this is
the test that says the set is complete before anybody starts it.

Refused by `TestTheWiringSuppliesEveryField` in `wiring_test.go`.

The near miss is the dialler dropped from `internal/system/system.go` while the
value that built it stays.

    wiring_test.go:15: the wiring supplies no [Dial], and main refuses to start on that

Not watched by a run.

### The fakes supply what the wiring supplies

The two sets are built in different packages at different times, and a fake set
that has fallen behind the real one is a suite passing against a program shape
that no longer exists.

Refused by `TestTheFakesSupplyWhatTheWiringSupplies` in `wiring_test.go`.

The near miss is the identifier source dropped from the fakes, which is what a
field added to one side and not the other looks like from here.

    wiring_test.go:32: the fakes leave [NewID] unsupplied while the wiring supplies every field

Not watched by a run.

## The rules that are prose today, and why each one is

Each of these is a rule about which part may know about which other part, and
each is unguarded for the same underlying reason: the code the rule is about
has not been written. A guard over an empty package is worth having, and this
tree has landed several, but a guard whose subject is a function signature that
does not exist yet cannot be written without inventing the signature, which is
the other issue's decision to take.

The consensus rule reads no proposal. `docs/model-visibility.md` and
`docs/consensus.md` both state it and #59 owns the test, which is over a
signature. #37 writes the rule the signature belongs to, and nothing here
computes a label today.

The consensus rule reads no volunteer identifier. `docs/consensus.md` fixes it
and the export takes its fields from there. Same subject, same absence.

No proposal reaches a volunteer before their answer is recorded.
`docs/model-visibility.md` is the position and says the test is over the bytes
sent to a browser rather than over what a template renders. There is no
surface, so there are no bytes. #59 owns it.

Scoring never happens while a volunteer waits. `docs/model-boundary.md` fixes
it. The subject is a request path, and there is no request path. #53 owns it.

A default deployment makes no outbound connection. `docs/federation.md` is the
position and #104 owns the test. The written half is guarded already, because
the dialling calls are in the table above, and the bound on that is the same
bound `docs/harness.md` states: it judges the call that is written rather than
the call that is reached, so a connection opened by a dependency or through a
helper this tree does not own is invisible to it. What #104 owes is the
observed half, and that needs a deployment to observe.

An image path stays inside the directory this project owns.
`docs/subject-media.md` draws the boundary. Code scanning has been observed to
refuse a path traversal in exactly the shape the ingest will be made of, which
`test/gate-refusals.md` records, but that is a security analyser reading for a
defect class rather than a test of this rule, and there is no ingest yet. #64
and #66 own it.

One command starts one process. `docs/deployment-ceiling.md` fixes the number
and the five questions a proposed second service has to pass. The subject is
what a deployment starts rather than what the source says, so nothing in this
tree could read it, and it stays prose after the deployment exists as well.
That last one is the only entry here whose reason is not going to expire.

The core reads no environment variable is not on either list, and the reason is
worth writing down. It could be guarded today, and `docs/harness.md` says
plainly that configuration, environment reading and the file system are #82's
and #85's, and that adding a shape for one before there is a caller would be
inventing it rather than deriving it. A guard here would fix where
configuration is read before the issue that decides it has been opened, so the
rule waits on #82 rather than on the code.

## What none of this covers

A green run says the calls that are written obey these rules. It says nothing
about what a dependency does, which is a promise the empty dependency set keeps
today rather than one any test here makes.

It says nothing about a goroutine, a channel or a race, which are the other
half of a suite that fails once a week.

The guards match on the name a package is imported under and do not resolve
scopes, so a local variable shadowing an import would be judged as the package,
and a call reached through a function this tree does not own is not judged at
all.

Nothing here refuses a rule that arrives with no entry in this file, and
nothing refuses an entry here that names a test the tree no longer carries. Both
directions are the same missing check, and #94 is where it would live.
