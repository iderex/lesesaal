# lesesaal

Zooniverse is open source and in practice not self-hostable: Panoptes is only the API, and a campaign also needs Cellect, Nero, ZooEventStats and Warehouse. A group with 2,000 images and ten volunteers ends up in Google Forms. This is docker compose up for a classification campaign, with active learning built in so the model proposes and humans correct only the uncertain cases. It feeds plattenschrank, whose binding constraint is label scarcity.

Planning happens on the issue tracker first. Every decision that shapes
the architecture is written down there with its reasons before the code
that depends on it exists.

See [NOTICE.md](NOTICE.md) for the intended-use notice.
