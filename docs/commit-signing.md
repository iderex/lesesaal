# Commit signing on the protected branch

Requested in issue #100. This document is the request rather than the change:
only the maintainer can edit a ruleset, and nothing here does it.

## The setting

Add the `required_signatures` rule to the `gate` ruleset on `main`. What that
ruleset carries today:

    $ gh api repos/iderex/lesesaal/rulesets/20520725 --jq '{enforcement, bypass: .bypass_actors, types: [.rules[].type]}'
    {"bypass":[],"enforcement":"active","types":["deletion","non_fast_forward","pull_request"]}

The request is one entry added to that list. Enforcement stays active and the
bypass list stays empty, which is the half that makes the rule mean anything: a
control with a bypass actor is a control the person most able to skip it does not
have.

## What it would refuse today

Nothing. Every commit on the default branch is already verified:

    $ gh api "repos/iderex/lesesaal/commits?per_page=100&sha=main" --jq '[.[] | .commit.verification.verified] | group_by(.) | map({v: .[0], n: length})'
    [{"n":21,"v":true}]
    $ git rev-list --count main
    21

So the setting costs this repository no repair on the way in. It is a floor under
what happens next rather than a cleanup of what has happened.

That number is worth reading carefully, because verified covers two different
facts here:

    $ gh api "repos/iderex/lesesaal/commits?per_page=100&sha=main" --jq '[.[] | {r: .commit.verification.reason, c: (.committer.login // "-")}] | group_by(.c + " " + .r) | map({k: (.[0].c + " " + .[0].r), n: length})'
    [{"k":"iderex valid","n":12},{"k":"web-flow valid","n":9}]

Twelve carry a signature made on somebody's machine with their own key. Nine are
merge commits made through the web interface and signed by GitHub's key, which
says the merge was made by an authenticated session and says nothing about a key
anybody holds.

## What it costs a contributor

A signing key, generated once. Either an SSH key or an OpenPGP key works, and the
SSH route reuses a key most contributors already have.

Three lines of git configuration: the format, the key, and signing on by default.
Without the last one, signing is something you remember, and the commit you
forget is the one that stops the merge.

The public half registered with GitHub as a signing key, which is a different
list from the authentication keys. A key that signs correctly but is not
registered produces a commit GitHub reports as unverified, and the failure looks
like a broken key rather than a missing registration.

Every commit in a branch, not only the tip. A branch is refused for its worst
commit.

## What it costs an outside contribution

Somebody sending a first change learns the requirement at the moment the merge is
refused, which is the end of the line and the most expensive place to learn
anything. That is why the contribution guide has to carry it before the rule
lands rather than after, and #29 is where that sentence belongs.

Nobody here can repair a fork's history for them. A maintainer with write access
to this repository has no write access to the branch the change came from, so the
rewrite below is work the contributor has to do, on a change they have already
finished once.

Nobody can wave it through either, and that is deliberate. The bypass list is
empty, so the rule applies to the maintainer's own branches on the same terms.

## A branch carrying one unsigned commit

It cannot be merged, and it cannot be fixed in place. Signing a commit changes
it, and changing a commit changes every commit after it, so the repair is a
rewrite of the branch from the first unsigned commit onwards.

Find them first:

    git log --format='%h %G? %s' origin/main..HEAD

`G` is a good signature. `N` is no signature and is what this rule refuses.

Then rewrite, and prove the rewrite changed nothing but the signatures:

    git checkout -b <name>-signed origin/main
    git cherry-pick origin/main..<name>
    git log --format=%H origin/main..HEAD | xargs -n1 git show | git patch-id --stable

That last command prints one line per commit, the patch identifier first and the
commit it came from second. Run over three commits of this repository it looks
like this:

    $ git log --format=%H origin/main~3..origin/main | xargs -n1 git show | git patch-id --stable
    3f9be1820139838c1fed25612b877fa2beeaad04 54209cb9df525e2cd6e3683c39a447c2f00ac145
    3f62b70e8fe11e4dc460208f71688f86b703cec3 9826e86ac709c3f94641e0899a757c5a1a681562
    d5e161bffaa58bf7959043f8d41545234c1acecb bc8455d866102452aded749734239db7115638bb
    e52c95c8ddabfe9150d70bb5f93123554f217808 081f9af54402d58ddb5e6b4ab63c5955f8a62f4f

Run it against the old branch and the new one and compare the first column only.
The second differs by design, since rewriting is what produced the new commits.
Equal patch identifiers mean the content is the change that was already reviewed,
which is what makes the replacement a replacement rather than a second attempt.

The superseded pull request is closed with the reason written into its body, and
the replacement carries the comparison above.

Two spellings that make the problem go away and must not be used:

    git commit --no-gpg-sign
    git -c commit.gpgsign=false commit

Neither is refused before the merge. A bypassed commit builds, passes every check
in this repository and reads exactly like a signed one, and the only thing that
says otherwise is the merge at the end of the line. A signing failure is a thing
to fix, not a thing to route around.

## What signing does not prove

It proves a key was used. That is all it proves.

It does not prove anybody reviewed anything. A signed commit and an unreviewed
commit are the same commit.

It does not prove the content is correct, tested, or what its message says it is.

It does not prove who was at the keyboard. A key that leaked, a key on a machine
somebody else has, and a key used by an agent acting for its holder all sign
identically, and the signature is evidence about the key rather than about a
person.

It says nothing about anything that reached this project by another route, which
includes every dependency the tree will eventually carry. The supply chain
position, #99, is where that lives, and signing is one link in it rather than the
whole chain.

## What is not measured here

Whether GitHub refuses a merge whose branch carries an unsigned commit, and
whether that answer is the same for a merge commit as for a squash, has not been
measured on this repository. Measuring it means switching the rule on, which is
the change this document is asking for and not one to make in order to find out
what it does.

So the repair procedure above is written against the refusal it expects. If the
rule lands and the refusal turns out to be narrower than that, this document is
where the correction goes, and the correction is a measurement rather than a
sentence.

## What this document is for

#100 is the request and stays open until the ruleset shows the rule. #29 is the
contribution guide and owes the sentence that sends a contributor here before
they have written anything to sign.
