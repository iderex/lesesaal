# The shape of the tree, and which way a dependency may point

Decided in issue #18.

## The property this exists for

The campaign core has to be runnable in a test with no browser, no images, no
model and no store behind it. That is a property of the layout as much as of
the code, and it is much cheaper to have on the first day than to recover on
the hundredth.

A boundary either exists in the directory structure or it does not exist at
all. So the parts below are directories rather than a convention, and the
direction a dependency runs between them is a rule with a guard rather than a
paragraph somebody remembers.

## The tree

    .
    +-- main.go                   the one entry point and the only wiring
    +-- layout_test.go            the guard behind this document
    +-- harness_test.go           the guard behind harness.md
    +-- go.mod                    the module and the toolchain it declares
    +-- internal/
    |   +-- campaign/             the campaign core
    |   |   +-- campaigntest/     the fakes a test injects
    |   +-- gate/                 the procedure that decides a change passes
    |   +-- model/                the model interface
    |   +-- store/                the storage layer
    |   |   +-- migration/        changes to the store's own shape
    |   +-- system/               the real values behind the injected ones
    |   +-- web/                  the surface
    +-- deploy/                   deployment assets
    +-- test/                     harnesses the unit suite cannot host
    +-- docs/                     the documents

`.github/` is not in that list and is not part of the layout. It belongs to
the platform this repository is hosted on rather than to this program, and
nothing in it is compiled.

## What belongs in each directory

`internal/campaign/` holds the campaign core, which is campaigns, subjects,
tasks, options, answers, classifications, labels and retirement in the words
`vocabulary.md` fixes them, together with the rules that decide what a
classification means and when a subject is finished; it does not hold a
handler, a query, a file path, a model or anything else that would make the
core need one of them present to run.

`internal/campaign/campaigntest/` holds the fakes a test injects in place of
the runtime, which are a clock a test moves, a draw that repeats, an identifier
source that counts and a dialler that refuses; it does not hold a test, and
nothing but a test file may import it, because a fake clock reaching the binary
an operator runs would be a deployment whose time does not move.

`internal/system/` holds the real values behind `campaign.Depends`, which are
the wall clock, the random source, the identifier generator and the outbound
dialler, together with the two functions that start a program and look for one
on the path; it does not hold a rule, a handler or anything a test would want
to stand in for, because it is the one directory permitted to read the runtime
and everything in it is therefore unreachable from a deterministic suite.
Starting a process belongs here for the same reason reading the clock does:
what it finds is whatever the machine happens to hold.

`internal/gate/` holds the procedure that decides whether a change passes,
which is the legs `ci` runs, the order they run in and what each one reports;
it does not hold a subprocess, because it reaches the machine only through the
two functions the entry point hands it, and its own suite decides what a
command returned instead of running one.

`internal/model/` holds the model interface, which is what a model is given
when it scores a subject, what a proposal is and how a proposal is recorded,
under `model-boundary.md` and `model-visibility.md`; it does not hold the code
that decides what a volunteer sees, because that decision is the surface's to
carry out and the core's to define.

`internal/store/` holds the storage layer, which is the embedded store the one
process in `deployment-ceiling.md` keeps and the reads and writes the core asks
for; it does not hold a rule about campaigns, because a rule that lives in a
query can only be tested against a store.

`internal/store/migration/` holds the ordered set of changes that takes an
existing store from one schema to the next, one file per change and never
edited once it has run anywhere; it does not hold the schema itself, which is
declared in the storage layer beside the code that reads it.

`internal/web/` holds the surface, which is the pages a volunteer classifies on
and a campaign owner reads results from, and the handlers behind them; it does
not hold a decision about which subject comes next or when one is retired,
because those belong to the core and a surface is one of several ways to reach
them.

`deploy/` holds the deployment assets, which are the container file, the
composition that starts it and the example configuration a first run is walked
through with; it does not hold anything the program imports, and nothing in it
is compiled.

`test/` holds the harnesses that need something the unit suite is forbidden to
have, a browser or a container runtime or a real collection on disk, each with
its own entry point and its own name in a workflow; it does not hold a test
that could have been a unit test, because moving one here does not make it slow
but does make it unwatched.

`docs/` holds the documents, which are part of the product rather than a
commentary on it; it holds no generated file, because a generated document
drifts from its source and nothing says which of the two is wrong.

A unit test lives beside the code it judges, in the same directory, which is
this toolchain's own convention. `test/` is not an alternative to that.

## Which way a dependency may point

Everything points inward, at the core. Read the table as: a file in the
directory on the left may import the packages on the right and no other package
inside this module.

    from                            may import inside this module
    .                               every part below
    internal/campaign               nothing
    internal/campaign/campaigntest  internal/campaign
    internal/gate                   nothing
    internal/model                  internal/campaign
    internal/store                  internal/campaign
    internal/store/migration        internal/campaign
    internal/system                 internal/campaign
    internal/web                    internal/campaign, internal/model
    deploy, test, docs              no compiled code lives there

Imports of the standard library are not this rule's subject and are not
restricted by it. What the module may depend on from outside itself is the
build check's, which refuses an import the module file does not carry.

Four directions are refused, and each is refused for its own reason rather than
for symmetry.

The core importing anything. This is the one the whole layout exists for. A
core that reaches the surface, the store or the model can only be tested with
those present, and the property is gone the day it happens rather than the day
somebody notices.

The surface importing the store. A handler that queries the store directly is a
rule about campaigns written where nothing can test it without a browser, and
it is how the core quietly stops being the place the rules live.

The model importing the store or the surface. `model-visibility.md` says a
proposal never reaches a volunteer before their answer is recorded. That
position is provable only while the model has no route to the page except
through `main.go`, and an import is such a route.

Anything importing `deploy/` or `test/`. They are the ends of the tree rather
than parts of it: the program is what gets deployed and what gets tested, and a
dependency in that direction inverts both.

Two of the parts above are refused in a direction this table cannot express, and
`harness_test.go` holds both. `internal/system` may be imported only from the
entry point, so the real clock enters the program in one place. And
`internal/campaign/campaigntest` may be imported only from a test file, which is
a rule about the kind of file rather than about the directory it sits in.
`harness.md` is where both are argued.

`main.go` may import everything, and it is the only file that may. Wiring the
parts together is a job somewhere, and putting it in the one place that is not
a part is what keeps the parts from doing it to each other.

## What holds it

`layout_test.go` at the module root. It reads the import declarations out of
every tracked Go file with the toolchain's own parser, decides which part each
file sits in from its directory, and refuses an import of this module that the
table above does not permit. It uses no subprocess, no network and no
dependency, which are the three things the unit suite is not allowed to have.

It carries a second test, on the table rather than on the imports: every
directory the table names has to exist. Without it, renaming a directory would
leave every rule about it unenforced and the run still green, because nothing
would be found inside it to judge. For the same reason the import test fails
when it examined no file at all: a run that judged nothing and a run that
judged everything and found nothing print the same word otherwise.

    go test ./... -run TestDependenciesPointTheWayTheLayoutSays -v

What that command reports today is that six Go files carry no import of this
module at all, because five of the six are a package clause and a comment. So
the guard is in place before the first import rather than after it, and the
proof that it bites is the failing run recorded on the change that landed it
rather than a claim here.

## Where the things that are not code live

Documentation is `docs/`, and this document is in it.

The deployment composition is `deploy/`, along with the container file and the
example configuration.

Schema changes are `internal/store/migration/`, under the storage layer rather
than beside it, because a migration is the storage layer changing its own
shape.

Fixtures live with whatever they are fixtures for. A unit fixture goes in a
`testdata` directory beside the test that reads it, which is the directory name
this toolchain already ignores when it builds. A collection of real images for
a harness is not a fixture in that sense and belongs with the harness.

The harnesses that need a browser, a container runtime or real media are
`test/`, one directory per harness, each with the workflow name that runs it.

## What is not here yet

Most of the tree above still holds a placeholder rather than code:
`deploy/README.md`, `test/README.md`, `internal/store/migration/README.md`, and
a package comment in `internal/model/doc.go`, `internal/store/doc.go` and
`internal/web/doc.go`. Each says what it is waiting for and which issue brings
it. A placeholder is replaced by the first real thing that lands, not kept
beside it.

Four directories now hold code. `internal/campaign` holds the dependencies the
program takes rather than reads, `internal/system` holds the real values behind
them, and `internal/campaign/campaigntest` holds the fakes. Those three arrived
with #21 and `harness.md` is where they are argued. `internal/gate` arrived with
#150 and holds the gate the entry point runs as `ci`.

`internal/campaign` also holds the campaign definition boundary, which is the
first of the core's own rules to land and arrived with #32. It is the shape a
written definition becomes and the refusal of one outside the grammar
`task-grammar.md` fixes, and it reaches nothing: the boundary judges a value it
was handed, so the property at the top of this document is what makes its suite
a suite of the rule rather than of a store.

Ingest, export and subject media have no directory here, and that is deliberate
rather than an omission. Where each one goes is decided by the direction rule
above at the moment the issue that needs it starts, and inventing a home for it
now would be laying out a tree around code nobody has written. What this
document fixes is the rule that decides the next case, which is the same shape
`deployment-ceiling.md` uses for a proposed new service.
