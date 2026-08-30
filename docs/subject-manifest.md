# What a campaign owner writes about their collection

Decided in issue #65.

## The shape

One file of comma-separated values, with a header row naming the columns and
one row per subject. That is what a spreadsheet writes when a researcher
chooses "save as", so producing one needs no tool from this project and no
instruction beyond the sentence above.

One column is required and it is called `file`. It names the image the row is
about. Everything else is the campaign owner's, carried through under the name
and in the order they wrote it, including every column this project
understands nothing about, because the column nobody anticipated is usually the
one the analysis needs.

One column is optional and is called `id`. It is the owner's own identifier for
the subject, which is what an export joins back to their catalogue. Where the
column is absent the file name is the identifier, so a working group that has
never named its plates anything but their file names does not have to invent a
second name for each.

Column names are read with their case and their surrounding space removed, so
`File`, `file` and `" file "` are the same column. What the export carries is
the name as it was written.

## A worked example

Four columns, two of which mean something here and two of which mean something
only to the archive that wrote them:

    id,file,plate,observed,note
    POSS-I-0421,0421.tif,0421,1954-11-02,"Edge lettering, partly rubbed away"
    POSS-I-0422,0422.tif,0422,1954-11-03,

The first row becomes a subject identified as `POSS-I-0421`, made from the
bytes of `0421.tif`, carrying three metadata fields named `plate`, `observed`
and `note` in that order. The second row carries an empty note, which is a
value rather than an absence and travels as one.

The whole of the procedure for producing that file is: open the spreadsheet,
make sure one column is called `file`, and save as comma-separated values.

## What may not be a column name

`digest`, `bytes`, `width`, `height` and `entered`. Each is a field this
project computes from the image at ingest, and two fields under one name is an
export column a reader cannot resolve back to either. A manifest using one is
refused with the name quoted, rather than one of the two silently winning.

The set is derived rather than written twice. `DerivedColumns` in
`internal/campaign/manifest.go` is the list, and this paragraph is prose about
it rather than a second copy: read the function.

## The four awkward cases

**A row naming a file that is not there.** Refused, with the row number. There
is nothing to make a subject from.

**A file with no row.** Refused, with the file named. This is the one worth
arguing with, because ingesting it would also be defensible. The manifest is
the campaign owner's statement of what their campaign is about, so a file they
did not describe would arrive as a subject nobody wrote a line about, and its
row in the export would carry an identifier and nothing else. Refusing it lets
them add the row or move the file, which are the two things they might have
meant, and neither is what silence would have chosen for them.

**Two rows naming the same file.** Refused, naming the earlier row. One file is
one subject; two rows about it would put two subjects behind one plate.

**A column name that collides with a field this project derives.** Refused, as
above.

Two more are refused for the same reasons and are not on that list because
nobody asked about them: two rows naming the same identifier, which the subject
record refuses anyway because an identifier is unique within a campaign, and a
row whose identifier column is empty while the manifest carries one.

## Everything at once, not the first thing

A manifest is read in one pass and every problem in it is reported together,
each with the row it is on and the header counted as row 1 the way a
spreadsheet counts it. A researcher fixing a two thousand row file one error
per run will stop, and this is the difference between a manifest that gets
fixed and one that gets abandoned.

The two comparisons against the collection are reported the same way, and
separately, because neither is a fact about the manifest: both are a comparison
against what was actually found on disk.

## No manifest at all

A campaign runs without one. Every file becomes a subject, the file name is the
identifier, and no metadata travels. That is the smallest thing a working group
can arrive with, and it has to be enough, or the manifest is required while
claiming not to be.

## What is not decided here

Where the files come from and what happens to the ones that are refused. That
is ingest, #64, and this document fixes only what a manifest says and what is
wrong with one.

What the metadata becomes in the export. `#68` writes the export, `#69` carries
the provenance into it, and the field list on the other side is the plate
archive project's schema, which #15 is open for.

Whether a manifest may name something that is not an image. The subject model
is images in the first release, and entry 7 of #1 is where that is a decision
rather than an assumption.
