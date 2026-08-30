package campaign

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

// DeclaredSubject is one subject as its campaign owner supplied it: the
// identifier they already use for it and whatever else they know about it.
//
// Both fields come out of a folder of images and a spreadsheet, which is the
// bound #33 puts on this shape. Nothing derived is here, because a campaign
// owner cannot supply a digest and would have to be told to compute one.
type DeclaredSubject struct {
	// ID is the campaign owner's own identifier for this subject, carried
	// unchanged so an export joins back to their catalogue without a lookup
	// table. docs/vocabulary.md fixes that a subject is one thing being
	// classified together with the metadata it arrived with, and this is the
	// half of that the owner names it by.
	ID string

	// Metadata is what came with the subject, in the order it came. It is an
	// ordered list rather than a map because a spreadsheet has columns in an
	// order its author chose, and a map would decide that order again on every
	// read.
	Metadata []Field
}

// Field is one thing the campaign owner knows about a subject, as a name and a
// value. Both are strings because a spreadsheet cell is one: giving a column a
// type here would mean this project deciding which of the owner's columns are
// numbers, and being wrong about one of them silently.
type Field struct {
	Name  string
	Value string
}

// Derived is what ingest computed about a subject's bytes, none of which the
// campaign owner supplies.
//
// It is a separate argument from the declaration rather than fields on it, so
// that the two cannot be confused at the call site: everything in the
// declaration is the owner's and survives unmodified, and everything here is
// this project's and is recomputable from the bytes, apart from the moment.
type Derived struct {
	// Digest is the content digest of the bytes, which docs/subject-media.md
	// makes the name the owned copy is stored under and the value a later
	// check compares against. DigestOf is where it comes from.
	Digest string

	// Bytes is the size of the copy, which is the figure docs/design-point.md
	// says the disk cost is arithmetic over and the one it calls most likely
	// to be wrong.
	Bytes int64

	// Width and Height are the dimensions of the image, in pixels. They are
	// derived at ingest rather than read per request because
	// docs/subject-media.md fixes that the sizes a browser is sent are
	// computed there, and a surface that has to open the file to lay out a
	// page is the version of this that makes somebody on a phone wait.
	Width  int
	Height int

	// Entered is when the subject entered the campaign. It is passed in rather
	// than read here, from Depends.Now, because this package reads no clock:
	// the same reason a Decision carries no moment.
	Entered time.Time
}

// Subject is one subject, judged. Its fields are unexported for the reason the
// judged definition types hold theirs that way: the only way to hold one is to
// have passed Enter, so a subject with no digest behind it cannot be built by
// anything downstream forgetting to check.
type Subject struct {
	id       string
	metadata []Field
	derived  Derived
}

// Enter judges a subject as ingest hands it over and returns the subject it
// becomes, or refuses it.
//
// It refuses in both directions, and the second is the one worth stating. A
// declaration missing the owner's identifier is refused because nothing could
// join the export back to their catalogue. A derived set missing its digest,
// its size, its dimensions or its moment is refused because every one of those
// is a thing this project computes rather than asks for, so an absent one is
// this project having failed rather than the owner having.
func Enter(declared DeclaredSubject, derived Derived) (Subject, error) {
	var refusals []Refusal
	refuse := func(says string) {
		refusals = append(refusals, Refusal{Says: says})
	}

	if strings.TrimSpace(declared.ID) == "" {
		refuse("the subject names no identifier, and the identifier is what an export joins back to the campaign owner's own catalogue")
	}

	named := map[string]bool{}
	for _, field := range declared.Metadata {
		switch {
		case strings.TrimSpace(field.Name) == "":
			refuse("the subject carries a metadata field with no name, which nothing could read back out of an export")
		case named[field.Name]:
			refuse(fmt.Sprintf("the subject carries the metadata field %q twice, so a reader would have to choose between two values this project cannot tell apart", field.Name))
		}
		named[field.Name] = true
	}

	if strings.TrimSpace(derived.Digest) == "" {
		refuse("the subject carries no digest, so nothing could tell a re-ingest from a new subject or notice the copy rotting on disk")
	}
	if derived.Bytes <= 0 {
		refuse(fmt.Sprintf("the subject declares %d byte(s), and a subject with no bytes behind it is not one this project copied", derived.Bytes))
	}
	if derived.Width <= 0 || derived.Height <= 0 {
		refuse(fmt.Sprintf("the subject declares dimensions of %dx%d, and no size a browser could be sent is derivable from that", derived.Width, derived.Height))
	}
	if derived.Entered.IsZero() {
		refuse("the subject carries no moment it entered the campaign, which is the one derived field a reader cannot recompute from the bytes afterwards")
	}

	if len(refusals) > 0 {
		return Subject{}, Refused{Refusals: refusals}
	}
	return Subject{
		id:       declared.ID,
		metadata: append([]Field(nil), declared.Metadata...),
		derived:  derived,
	}, nil
}

// ID is the campaign owner's identifier for this subject.
func (s Subject) ID() string { return s.id }

// Metadata is what the campaign owner supplied, in the order they supplied it.
// It returns a copy, so a caller cannot write through the result into a subject
// that has already been judged.
//
// Every field they wrote is here, including the ones this project understands
// nothing about, because the column nobody anticipated is usually the one the
// analysis needs.
func (s Subject) Metadata() []Field { return append([]Field(nil), s.metadata...) }

// Derived is what this project computed about the bytes.
func (s Subject) Derived() Derived { return s.derived }

// Digest is the content digest recorded for this subject.
func (s Subject) Digest() string { return s.derived.Digest }

// Changed says whether bytes carrying the digest given are the bytes this
// subject was made from.
//
// It is the comparison docs/subject-media.md asks the periodic check to make,
// and it is here rather than beside the disk because the recorded digest is the
// subject's and the reading is the caller's. A caller holding a file hashes it
// with DigestOf and asks this, so one definition of the comparison serves the
// rot check, a re-ingest and an export.
func (s Subject) Changed(digest string) bool { return digest != s.derived.Digest }

// DigestOf reads bytes to the end and returns the content digest of them, which
// is what a copy is stored under and what a later check compares against.
//
// It is in the core so that ingest, the periodic rot check and anything that
// later compares two copies cannot each pick their own hash. A second
// definition of this would not fail loudly: it would produce a different digest
// for the same plate, and every comparison across the two would report a change
// that never happened.
//
// The name of the algorithm is carried in the value rather than assumed by
// whoever reads it, so a digest computed under a second one is distinguishable
// from this one instead of merely unequal to it.
func DigestOf(r io.Reader) (string, error) {
	sum := sha256.New()
	if _, err := io.Copy(sum, r); err != nil {
		return "", fmt.Errorf("reading the bytes to digest them: %w", err)
	}
	return "sha256:" + hex.EncodeToString(sum.Sum(nil)), nil
}

// Subjects is the subjects of one campaign. It exists to hold the uniqueness
// #33 asks for: the campaign owner's identifier is unique within a campaign,
// and the second subject to claim one is refused rather than replacing the
// first.
//
// The refusal is here because there is nowhere else for it to be. The store is
// not chosen and its schema does not exist, which docs/deployment-ceiling.md
// says plainly, so a constraint written into a table today would be written
// nowhere. A store arriving later declares it again in its own terms; what this
// type fixes is that a duplicate is refused rather than resolved, which is the
// half a schema cannot decide on its own.
type Subjects struct {
	inOrder []Subject
	byID    map[string]int
}

// Add takes one subject into the campaign, or refuses it because its identifier
// is already taken.
func (s *Subjects) Add(subject Subject) error {
	if _, taken := s.byID[subject.id]; taken {
		return Refused{Refusals: []Refusal{{
			Says: fmt.Sprintf("the subject %q is already in this campaign, and two subjects under one identifier would put two plates behind one row of the export", subject.id),
		}}}
	}
	if s.byID == nil {
		s.byID = map[string]int{}
	}
	s.byID[subject.id] = len(s.inOrder)
	s.inOrder = append(s.inOrder, subject)
	return nil
}

// Len is how many subjects the campaign holds.
func (s Subjects) Len() int { return len(s.inOrder) }

// All is every subject, in the order they were taken in. It returns a copy for
// the reason every other reader here does.
func (s Subjects) All() []Subject { return append([]Subject(nil), s.inOrder...) }

// ByID is the subject the campaign owner named, and whether the campaign holds
// one under that identifier at all.
func (s Subjects) ByID(id string) (Subject, bool) {
	at, held := s.byID[id]
	if !held {
		return Subject{}, false
	}
	return s.inOrder[at], true
}
