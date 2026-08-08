# Schema changes

What belongs here: the ordered set of changes that takes an existing store from
one schema to the next, one file per change, never edited once it has run
anywhere.

What does not belong here: the schema itself. What the store holds today is
declared in `internal/store/`, and a reader asking what a column means should
not have to replay a directory of changes to find out.

It sits under the storage layer rather than beside it because a migration is
the storage layer changing its own shape. Nothing outside `internal/store/`
reads this directory, and the direction rule in `docs/layout.md` applies to
anything compiled here exactly as it applies to the parent.

There is nothing here yet. The store and its first schema are #33, and until
that exists there is no shape to change.

This file is the placeholder `docs/layout.md` promises in every directory where
code does not exist yet.
