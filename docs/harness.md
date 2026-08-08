# The test harness: an injected clock, no network, no sleeps

Built for issue #21, before the tests that need it exist. That order is
deliberate. Almost everything this system does is about time and order, so a
suite written against the real clock is slow, is flaky, and is quietly a test of
the machine it ran on. Recovering that on the hundredth test costs a rewrite of
all hundred; having it on the first costs one file.

## What is injected

`campaign.Depends` holds four things this program takes rather than reads:

    Now     the current time
    Intn    a number in [0, n)
    NewID   an identifier not returned before
    Dial    an outbound connection

They are function fields rather than an interface each. A test that needs only a
clock writes one field and leaves the rest, and adding a fifth concern later
does not force every existing fake to grow a method it ignores.

`Missing` names the fields left nil. A nil field is a dependency nobody
supplied, and a program that starts on one panics at the first call rather than
at startup, which is the wrong end of a run to find out. `main.go` refuses to
start on an incomplete set.

## Where the real ones come from, and where they may not

`internal/system` builds them: the wall clock, the operating system's random
source, a hexadecimal identifier from that same source, and a dialler with a
timeout. It is a package of its own so that reading the runtime is something one
directory does rather than a habit spread across the tree, and `docs/layout.md`
lets only the entry point import it.

`internal/campaign/campaigntest` holds the fakes: a clock a test moves by hand,
a draw that repeats exactly, an identifier source that counts, and a dialler
that refuses every address and names it in the error. Nothing in it reads the
runtime, so a suite built on it is the same suite on every machine and at every
hour. Only a test file may import it.

## The mechanism

`harness_test.go` at the module root, which is the same shape as
`layout_test.go` and sits beside it on purpose. It reads the source with the
toolchain's own parser, so it needs no subprocess, no network and no dependency.

It refuses three things.

A direct read of the runtime anywhere but the wiring package. The table it
judges against is in that file, and it names the injected field to take instead,
so a failure says which rule was broken rather than only that one was. What is
in it, by the package it reaches:

    time             the reads that ask the clock for the moment
    math/rand        every identifier
    math/rand/v2     every identifier
    crypto/rand      every identifier
    net              the dialling and the listening calls
    net/http         the calls that go out through a package-level client

Type names are not in it, so a signature naming an instant or a connection is
untouched.

A sleep, in every file in the tree, including the wiring and including a test.
In a test a sleep is a statement that the author could not express what they
were waiting for, and it is the commonest single cause of a suite that fails
once a week for no reason anybody can reproduce. In the program it is a wait
nobody can move a clock past. That is the search #21 asks for, and the check
that keeps it empty is `TestOnlyTheWiringReadsTheRuntime`.

An import of the fakes from anything that is not a test file. A fake clock
reaching the binary an operator runs would be a deployment whose time does not
move, and nothing else here would notice.

A fourth test refuses a dot import, because every rule above is decided by the
name a package is imported under and a dot import has none.

## What the guard cannot see

It matches on the name a package is imported under and does not resolve scopes,
so a local variable shadowing an import would be judged as the package.

It reads this repository's own source and nothing a dependency does. That is a
promise the empty dependency set keeps today rather than one this test makes,
and `docs/dependencies.md` is where the set is.

It judges the call that is written rather than the call that is reached, so a
runtime read behind a function this tree does not own is invisible to it.

It says nothing about a goroutine, a channel or a race. Those are the other half
of a flaky suite and no check here covers them.

## Advancing time

    go test ./internal/campaign/campaigntest -run TestAClockMovesAYearAtOnce -v

The clock moves a year for the same cost as a nanosecond, which is what makes
retirement, expiry and retraining reachable in a test at all. That it costs no
wall time is not asserted inside the test: asserting it would mean reading the
real clock, which is the thing the package exists to stop, so the evidence is
the suite's own reported duration.

## What this does not do

It does not make the suite concurrent-safe. The fake clock is not safe for use
from several goroutines at once, deliberately: a test that needed it to be would
be a test about concurrency and should say so and build its own.

It does not stop a dependency from opening a socket, and it does not stop the
process the suite runs in from opening one either. What this guard refuses is
the call written in this repository's own source. Whether the run itself is
confined is a property of how the suite is started rather than of what it is
written with, and #20 is where the run is.

It supplies no configuration, no environment reading and no file system. Those
are #82's and #85's, and adding a fifth field here before there is a caller for
it would be inventing a shape rather than deriving one.
