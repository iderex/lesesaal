# The language this is written in

Decided in issue #17.

## The choice

Go, minimum toolchain 1.26, declared in `go.mod` and nowhere else.

The rest of this document is the check that the means fits, answered against
what this project has to do rather than against a preference. It answers in both
directions, and the requirement it fits worst is written out at the same length
as the ones it fits best.

## What the means had to carry

An operator starts the deployment with one command and maintains no runtime for
it. Go builds a single executable with no interpreter, no virtual environment
and no shared library to install, so the thing an operator copies is the thing
that runs. This is the promise the project is for rather than a convenience, and
it is the requirement that carries the most weight here.

A web surface and an API served from one process. The standard library serves
both, so this costs no framework, no dependency and no second thing to keep
current. The service count that `deployment-ceiling.md` fixes at one is a
property of how the program is written rather than something the language has to
be argued out of.

A model in the loop. This is the requirement the choice fits worst and the next
section is about it.

A test story strong enough for the campaign core to run with an injected clock,
no browser, no images and no network. The test runner is part of the toolchain,
needs no display and no elevation, and injection through interfaces is the
ordinary way to write this rather than a pattern imported for the occasion. What
#21 has to build is the injection itself, not an apparatus to make injection
possible.

A dependency graph that can be locked, scanned and reproduced. `go.mod` and its
checksum file pin every direct and transitive dependency with a hash, the build
can be told to refuse anything the lock does not already carry, and the
verification is part of the toolchain rather than a third-party tool. #23 is
where that is turned into a check.

A container image small enough that the footprint claim is not embarrassing. A
statically linked executable copied into an image with nothing else in it is
tens of megabytes, and it carries no package manager and no shell for anything
to be found in later. #74 builds it and #117 weighs it, and neither number is
claimed here.

## What it costs the model work

The honest cost, stated in the words of `model-boundary.md` rather than in
general ones.

That document puts inference and training inside the one process, on a CPU, with
no accelerator assumed. Go has no ecosystem for training anything substantial,
so the first model in #54 has to be one that can be trained on a CPU in this
language: a linear or tree model over features derived in-process, rather than
anything that needs a deep learning framework to fit.

`model-boundary.md` already priced the way out and it is not being invented
here. An operator may produce a model artefact on any machine, by any means, and
hand the file to the deployment, which costs this project a file format and a
validation step rather than a runtime. So the ceiling on what can be trained
in-process is real, and it is not a ceiling on what can be run.

What the choice refuses is the shape where the model is a second service in a
second language. That was refused in `model-boundary.md` before this decision,
which is why the language question does not reopen it. This document only says
what that refusal costs once the language is fixed: no in-process training of a
model that needs a framework this toolchain does not have.

## What it adds to the tree

One toolchain, where the tree had none. Before this change the repository held
documents and workflow files and nothing that compiles.

That cost is paid knowingly and it is paid once. The alternative shapes cost
more than one: a service in one language with a model in another is two
toolchains, two dependency graphs to lock, two scanners to configure and two
sets of build reproducibility to argue.

No dependency is added at all. There is no checksum file in this change because
the dependency set is empty, which is also why the entry point builds with no
network access. #23 owns the moment that stops being true.

## What was priced and not chosen

Python is the answer that wins the model requirement outright and loses the
first one. An operator would maintain an interpreter and an environment, the
image is larger by an order of magnitude, and the dependency graph is the
hardest of the candidates to lock and reproduce exactly. For a project whose
product is that a working group can run this without having to look after the
software as well, that is the wrong trade.

A runtime-based platform with a self-contained publish sits between the two. It
would satisfy the deployment requirement and has a strong test story, its model
story through a portable inference format is about as good as Go's, and it costs
a larger image and a heavier build. It is a defensible answer and not the one
taken.

A systems language without a garbage collector is stronger than Go on image size
and on nothing else that matters at the size in `design-point.md`, and it costs
build time and the ease with which somebody can send a fix. That last part is a
claim rather than a measurement.

None of the four is the right answer for every artefact this project will ever
build. The rule is that the means is checked per artefact, so a later thing that
is genuinely forced elsewhere gets its own check rather than inheriting this
one.

## What exists today, and what it proves

`main.go` prints its version and exits. That is the whole program.

It exists in that state because it is the smallest artefact that proves the
toolchain builds and runs from a clean checkout, and because #19 needs something
to compile before it can be a build check. The version string is a placeholder
and says so: what a version number promises here is #118's decision.
