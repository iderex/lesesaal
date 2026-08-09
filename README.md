# lesesaal

Zooniverse is open source and in practice not self-hostable. Its API component's
development composition starts four services before anything is added to it:

    $ gh api repos/zooniverse/panoptes/contents/docker-compose.yml --jq .content | base64 -d | sed -n '/^services:/,/^volumes:/p' | grep -E '^  [a-z_]+:'
      postgres:
      redis:
      panoptes:
      sidekiq:

A group with 2,000 images and ten volunteers ends up in Google Forms. That is a
claim rather than a measurement, and so is the rest of this paragraph. This is
docker compose up for a classification campaign, with active learning built in so
the model proposes and humans correct only the uncertain cases. It feeds
plattenschrank, whose binding constraint is label scarcity.

Issue #2 holds the evidence for the gap, including what has been measured and
what has not.

Planning happens on the issue tracker first. Every decision that shapes
the architecture is written down there with its reasons before the code
that depends on it exists.

See [NOTICE.md](NOTICE.md) for the intended-use notice, and
[docs/scope-of-use.md](docs/scope-of-use.md) for what this is built for and the
specific things it is not built for. Nothing in the software enforces that
statement.

[GOVERNANCE.md](GOVERNANCE.md) says who decides and what happens to a proposal
nobody has time for. [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) says what is
expected here and who receives a report. Both cover this repository and not a
deployment somebody else runs.
