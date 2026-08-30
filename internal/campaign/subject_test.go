package campaign

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

// entered is the moment the fixtures below stamp a subject with. A test that
// needs a second moment writes its own; what matters everywhere else is only
// that it is not the zero time, which Enter refuses.
var entered = time.Date(2026, 8, 30, 21, 0, 0, 0, time.UTC)

// plate is a subject as ingest would hand it over, with the metadata a plate
// archive actually carries beside the identifier.
func plate(t *testing.T) Subject {
	t.Helper()
	subject, err := Enter(
		DeclaredSubject{
			ID: "POSS-I-0421",
			Metadata: []Field{
				{Name: "plate", Value: "0421"},
				{Name: "observed", Value: "1954-11-02"},
				{Name: "emulsion", Value: "103a-O"},
			},
		},
		Derived{
			Digest:  "sha256:" + strings.Repeat("ab", 32),
			Bytes:   52_428_800,
			Width:   14000,
			Height:  14000,
			Entered: entered,
		},
	)
	if err != nil {
		t.Fatalf("the fixture subject is refused: %v", err)
	}
	return subject
}

// TestMetadataThisProjectUnderstandsNothingAboutSurvivesTheBoundary is #33's
// first condition. The fields below are deliberately ones no part of this
// project reads, in an order nothing here would choose, and the round trip is
// through the boundary rather than past it: what comes back out is fed in
// again, so a field dropped or reordered on either leg shows up as a
// difference rather than as a subject that still looks plausible.
func TestMetadataThisProjectUnderstandsNothingAboutSurvivesTheBoundary(t *testing.T) {
	declared := DeclaredSubject{
		ID: "arch/1998/plate-7",
		Metadata: []Field{
			{Name: "zeta_index", Value: "  4.1e-3  "},
			{Name: "Bemerkung", Value: "Randbeschriftung teilweise abgerieben"},
			{Name: "cabinet", Value: "3"},
			{Name: "plate", Value: "0421"},
		},
	}

	first, err := Enter(declared, plate(t).Derived())
	if err != nil {
		t.Fatalf("a subject carrying metadata this project does not read was refused: %v", err)
	}
	if got := first.Metadata(); !reflect.DeepEqual(got, declared.Metadata) {
		t.Fatalf("the metadata came back as %v, want %v", got, declared.Metadata)
	}

	second, err := Enter(DeclaredSubject{ID: first.ID(), Metadata: first.Metadata()}, first.Derived())
	if err != nil {
		t.Fatalf("the subject read back out of the first one was refused: %v", err)
	}
	if !reflect.DeepEqual(second.Metadata(), first.Metadata()) {
		t.Errorf("the second trip through the boundary produced %v, want %v", second.Metadata(), first.Metadata())
	}
	if second.ID() != declared.ID {
		t.Errorf("the identifier came back as %q, want %q", second.ID(), declared.ID)
	}
	if second.Derived() != first.Derived() {
		t.Errorf("the derived fields came back as %+v, want %+v", second.Derived(), first.Derived())
	}
}

// TestTheOwnersMetadataCannotBeWrittenThroughAfterwards holds the copy the
// reader returns. Without it a caller could reach into a judged subject and
// change what the campaign owner supplied, which is the one thing this record
// promises not to do.
func TestTheOwnersMetadataCannotBeWrittenThroughAfterwards(t *testing.T) {
	subject := plate(t)
	read := subject.Metadata()
	read[0] = Field{Name: "plate", Value: "9999"}
	if got := subject.Metadata()[0].Value; got != "0421" {
		t.Errorf("the subject now reports %q for the first field, want %q", got, "0421")
	}
}

// TestAMetadataFieldNamedTwiceOrNotAtAllIsRefused is the pair of cases a
// spreadsheet produces on its own: a blank column heading, and the same heading
// written twice. Carrying either would mean an export column a reader cannot
// resolve back to what the owner wrote.
func TestAMetadataFieldNamedTwiceOrNotAtAllIsRefused(t *testing.T) {
	derived := plate(t).Derived()
	for _, metadata := range [][]Field{
		{{Name: "", Value: "3"}},
		{{Name: "   ", Value: "3"}},
		{{Name: "plate", Value: "0421"}, {Name: "plate", Value: "0422"}},
	} {
		if _, err := Enter(DeclaredSubject{ID: "p1", Metadata: metadata}, derived); err == nil {
			t.Errorf("the metadata %v was accepted", metadata)
		}
	}
}

// TestASubjectWithNoDerivedFieldsIsRefused holds the direction that is this
// project's own failure rather than the campaign owner's. Every field named
// here is computed at ingest, so an absent one means ingest did not do its
// work and the subject would carry a hole nothing later could fill.
func TestASubjectWithNoDerivedFieldsIsRefused(t *testing.T) {
	whole := plate(t).Derived()
	for name, derived := range map[string]Derived{
		"no digest":     {Bytes: whole.Bytes, Width: whole.Width, Height: whole.Height, Entered: whole.Entered},
		"blank digest":  {Digest: "  ", Bytes: whole.Bytes, Width: whole.Width, Height: whole.Height, Entered: whole.Entered},
		"no bytes":      {Digest: whole.Digest, Width: whole.Width, Height: whole.Height, Entered: whole.Entered},
		"no width":      {Digest: whole.Digest, Bytes: whole.Bytes, Height: whole.Height, Entered: whole.Entered},
		"no height":     {Digest: whole.Digest, Bytes: whole.Bytes, Width: whole.Width, Entered: whole.Entered},
		"no moment":     {Digest: whole.Digest, Bytes: whole.Bytes, Width: whole.Width, Height: whole.Height},
		"negative size": {Digest: whole.Digest, Bytes: -1, Width: whole.Width, Height: whole.Height, Entered: whole.Entered},
	} {
		if _, err := Enter(DeclaredSubject{ID: "p1"}, derived); err == nil {
			t.Errorf("a subject with %s was accepted", name)
		}
	}
	if _, err := Enter(DeclaredSubject{ID: "  "}, whole); err == nil {
		t.Error("a subject naming no identifier was accepted")
	}
}

// TestARefusedSubjectSaysEverythingWrongWithIt is the same property Refused
// carries for a campaign definition, and it is here because the caller is
// ingest rather than a person: a folder refused one file at a time is a folder
// somebody re-runs ingest over five times.
func TestARefusedSubjectSaysEverythingWrongWithIt(t *testing.T) {
	_, err := Enter(DeclaredSubject{}, Derived{})
	var refused Refused
	if !errors.As(err, &refused) {
		t.Fatalf("a subject with nothing in it was refused as %T, want campaign.Refused", err)
	}
	if len(refused.Refusals) != 5 {
		t.Errorf("it reported %d refusal(s), want 5: %v", len(refused.Refusals), refused.Refusals)
	}
}

// TestTheSecondSubjectUnderOneIdentifierIsRefused is #33's second condition.
// The identifier is the campaign owner's own, so this project cannot invent a
// second one to resolve the collision, and replacing the first would silently
// discard whatever had already been classified against it.
func TestTheSecondSubjectUnderOneIdentifierIsRefused(t *testing.T) {
	var subjects Subjects
	if err := subjects.Add(plate(t)); err != nil {
		t.Fatalf("the first subject was refused: %v", err)
	}

	same, err := Enter(
		DeclaredSubject{ID: "POSS-I-0421", Metadata: []Field{{Name: "plate", Value: "9999"}}},
		Derived{Digest: "sha256:" + strings.Repeat("cd", 32), Bytes: 1, Width: 1, Height: 1, Entered: entered},
	)
	if err != nil {
		t.Fatalf("the second subject is refused by the boundary rather than by the set: %v", err)
	}
	if err := subjects.Add(same); err == nil {
		t.Fatal("a second subject under one identifier was accepted")
	}

	if subjects.Len() != 1 {
		t.Errorf("the campaign holds %d subject(s), want 1", subjects.Len())
	}
	held, there := subjects.ByID("POSS-I-0421")
	if !there {
		t.Fatal("the campaign no longer holds the subject it accepted")
	}
	if got := held.Metadata()[0].Value; got != "0421" {
		t.Errorf("the refused subject overwrote the first: the field reads %q, want %q", got, "0421")
	}
}

// TestSubjectsKeepsTheOrderTheyArrivedIn is what the set owes beyond
// uniqueness. Ingest walks a folder, and a set that reordered it would make
// every later report about "the first hundred subjects" a report about a
// different hundred on each run.
func TestSubjectsKeepsTheOrderTheyArrivedIn(t *testing.T) {
	var subjects Subjects
	derived := plate(t).Derived()
	for _, id := range []string{"c", "a", "b"} {
		subject, err := Enter(DeclaredSubject{ID: id}, derived)
		if err != nil {
			t.Fatalf("the subject %q is refused: %v", id, err)
		}
		if err := subjects.Add(subject); err != nil {
			t.Fatalf("the subject %q was not taken: %v", id, err)
		}
	}
	var got []string
	for _, subject := range subjects.All() {
		got = append(got, subject.ID())
	}
	if want := []string{"c", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("the subjects came back as %v, want %v", got, want)
	}
	if _, there := subjects.ByID("nothing here"); there {
		t.Error("the set reported a subject nobody added")
	}
}

// TestAChangedFileIsDetectableFromTheRecordedDigest is #33's third condition.
// The case is docs/subject-media.md's owned copy rotting on disk: the bytes are
// read again later and compared against what was recorded, and one byte is
// enough to make them a different plate.
func TestAChangedFileIsDetectableFromTheRecordedDigest(t *testing.T) {
	original := []byte("\x89PNG\r\n\x1a\n a plate, as far as this test is concerned")
	digest, err := DigestOf(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("digesting the bytes failed: %v", err)
	}

	subject, err := Enter(
		DeclaredSubject{ID: "POSS-I-0421"},
		Derived{Digest: digest, Bytes: int64(len(original)), Width: 10, Height: 10, Entered: entered},
	)
	if err != nil {
		t.Fatalf("the subject is refused: %v", err)
	}

	again, err := DigestOf(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("digesting the same bytes a second time failed: %v", err)
	}
	if again != digest {
		t.Fatalf("the same bytes digested to %q and then to %q", digest, again)
	}
	if subject.Changed(again) {
		t.Error("the subject reports its own bytes as changed")
	}

	rotted := append([]byte(nil), original...)
	rotted[len(rotted)-1] ^= 0x01
	after, err := DigestOf(bytes.NewReader(rotted))
	if err != nil {
		t.Fatalf("digesting the changed bytes failed: %v", err)
	}
	if after == digest {
		t.Fatal("one changed byte produced the same digest")
	}
	if !subject.Changed(after) {
		t.Error("the subject reports changed bytes as its own")
	}
}

// TestTheDigestNamesTheAlgorithmItUsed holds the prefix. A bare hex string
// would be indistinguishable from one computed under a different hash the day
// this project ever has two, and every comparison across the two would then
// report a change that never happened.
func TestTheDigestNamesTheAlgorithmItUsed(t *testing.T) {
	digest, err := DigestOf(strings.NewReader(""))
	if err != nil {
		t.Fatalf("digesting an empty reader failed: %v", err)
	}
	const empty = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if digest != empty {
		t.Errorf("the empty digest is %q, want %q", digest, empty)
	}
}

// TestTheDigestReportsAReadThatFailedRatherThanDigestingWhatItGot is the case a
// disk gives ingest halfway through a file. Returning the digest of the bytes
// that did arrive would record a subject under a name that matches nothing, and
// the rot check would then report it as changed forever.
func TestTheDigestReportsAReadThatFailedRatherThanDigestingWhatItGot(t *testing.T) {
	broken := io.MultiReader(strings.NewReader("half a plate"), failingReader{})
	if _, err := DigestOf(broken); err == nil {
		t.Fatal("a read that failed produced a digest")
	}
}

// failingReader is a disk that stops halfway.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("the disk stopped") }

// TestASubjectAsksForNothingBeyondAFolderAndASpreadsheet is #33's fourth
// condition, read off the declared shape rather than described. A field added
// to it later that a campaign owner cannot fill in from what they already have
// is invisible in a table of cases: it would simply be left empty in every test
// somebody wrote afterwards.
func TestASubjectAsksForNothingBeyondAFolderAndASpreadsheet(t *testing.T) {
	shape := reflect.TypeOf(DeclaredSubject{})
	want := []struct {
		name string
		kind string
	}{
		{"ID", "string"},
		{"Metadata", "[]campaign.Field"},
	}
	if got := shape.NumField(); got != len(want) {
		t.Fatalf("a declared subject carries %d field(s), want %d", got, len(want))
	}
	for i, field := range want {
		got := shape.Field(i)
		if got.Name != field.name || got.Type.String() != field.kind {
			t.Errorf("field %d is %s %s, want %s %s", i, got.Name, got.Type, field.name, field.kind)
		}
	}

	metadata := reflect.TypeOf(Field{})
	if got := metadata.NumField(); got != 2 {
		t.Fatalf("a metadata field carries %d field(s), want 2", got)
	}
	for i, name := range []string{"Name", "Value"} {
		got := metadata.Field(i)
		if got.Name != name || got.Type.Kind() != reflect.String {
			t.Errorf("metadata field %d is %s %s, want %s string", i, got.Name, got.Type, name)
		}
	}

	// A folder of images and no spreadsheet at all is the smallest thing an
	// owner can arrive with, and it has to be enough.
	if _, err := Enter(DeclaredSubject{ID: "POSS-I-0421"}, plate(t).Derived()); err != nil {
		t.Errorf("a subject with an identifier and no metadata was refused: %v", err)
	}
}
