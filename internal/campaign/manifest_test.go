package campaign

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// read reads a manifest and fails the test where it was refused, so the cases
// about a good manifest are not three lines of error handling each.
func read(t *testing.T, written string) Manifest {
	t.Helper()
	manifest, err := ReadManifest(strings.NewReader(written))
	if err != nil {
		t.Fatalf("the manifest was refused: %v", err)
	}
	return manifest
}

// refusals reads a manifest that should be refused and hands back what it said.
func refusals(t *testing.T, written string) []ManifestRefusal {
	t.Helper()
	_, err := ReadManifest(strings.NewReader(written))
	var refused ManifestRefused
	if !errors.As(err, &refused) {
		t.Fatalf("the manifest was accepted or refused as %T, want campaign.ManifestRefused", err)
	}
	return refused.Refusals
}

// TestAManifestIsWhatASpreadsheetSaves is #65's first condition read at the
// format. What a researcher does is choose "save as comma-separated values",
// and this is what comes out of one: a header row of their own column names, a
// column naming the file, and everything else theirs.
func TestAManifestIsWhatASpreadsheetSaves(t *testing.T) {
	manifest := read(t, `id,file,plate,observed,note
POSS-I-0421,0421.tif,0421,1954-11-02,"Edge lettering, partly rubbed away"
POSS-I-0422,0422.tif,0422,1954-11-03,
`)

	if manifest.Len() != 2 {
		t.Fatalf("the manifest holds %d row(s), want 2", manifest.Len())
	}
	subjects := manifest.Subjects()
	if subjects[0].ID != "POSS-I-0421" {
		t.Errorf("the first identifier is %q, want %q", subjects[0].ID, "POSS-I-0421")
	}
	if got := manifest.Files(); !reflect.DeepEqual(got, []string{"0421.tif", "0422.tif"}) {
		t.Errorf("the files are %v, want [0421.tif 0422.tif]", got)
	}

	// Everything that is not the file or the identifier is the owner's, in the
	// order and under the name they wrote, including a value carrying the
	// separator and one that is empty.
	want := []Field{
		{Name: "plate", Value: "0421"},
		{Name: "observed", Value: "1954-11-02"},
		{Name: "note", Value: "Edge lettering, partly rubbed away"},
	}
	if got := subjects[0].Metadata; !reflect.DeepEqual(got, want) {
		t.Errorf("the metadata is %v, want %v", got, want)
	}
	if got := subjects[1].Metadata[2]; got.Value != "" || got.Name != "note" {
		t.Errorf("the empty cell came back as %v", got)
	}

	// And what comes out of it is what the subject boundary takes.
	for _, declared := range subjects {
		if _, err := Enter(declared, Derived{Digest: "sha256:x", Bytes: 1, Width: 1, Height: 1, Entered: entered}); err != nil {
			t.Errorf("a subject read out of the manifest was refused by the boundary: %v", err)
		}
	}
}

// TestASpreadsheetsByteOrderMarkIsNotPartOfTheFirstColumnsName is the case that
// makes the difference between a manifest that works and one that reports a
// missing file column to a researcher looking at a file that plainly has one.
// It is what the spreadsheet most researchers have writes by default.
func TestASpreadsheetsByteOrderMarkIsNotPartOfTheFirstColumnsName(t *testing.T) {
	manifest := read(t, byteOrderMark+"file,plate\n0421.tif,0421\n")
	if manifest.Len() != 1 {
		t.Fatalf("the manifest holds %d row(s), want 1", manifest.Len())
	}
	if got := manifest.Subjects()[0].ID; got != "0421.tif" {
		t.Errorf("the identifier is %q, want %q", got, "0421.tif")
	}
	if got := len(manifest.Subjects()[0].Metadata); got != 1 {
		t.Errorf("the row carries %d metadata field(s), want 1", got)
	}
}

// TestTheFileNameIsTheIdentifierWhereNoColumnNamesOne is what makes the
// identifier column optional. A working group that has a folder and a
// spreadsheet of notes should not have to invent a second name for every plate.
func TestTheFileNameIsTheIdentifierWhereNoColumnNamesOne(t *testing.T) {
	manifest := read(t, "file,plate\n0421.tif,0421\n0422.tif,0422\n")
	for i, declared := range manifest.Subjects() {
		if declared.ID != manifest.Files()[i] {
			t.Errorf("row %d is identified as %q and names the file %q", i+2, declared.ID, manifest.Files()[i])
		}
	}
}

// TestAColumnNamedAfterADerivedFieldIsRefused is one of the four awkward cases
// #65 names. A column called digest would arrive at the export beside the
// digest this project computed, under one name, and a reader could not tell
// which of the two they were looking at.
func TestAColumnNamedAfterADerivedFieldIsRefused(t *testing.T) {
	for _, name := range DerivedColumns() {
		said := refusals(t, "file,"+name+"\n0421.tif,anything\n")
		if len(said) != 1 {
			t.Fatalf("a manifest with a %q column reported %d refusal(s): %v", name, len(said), said)
		}
		if said[0].Row != 1 {
			t.Errorf("the refusal about the %q column is on row %d, want the header", name, said[0].Row)
		}
	}

	// The comparison ignores case and surrounding space, so a spreadsheet's
	// "Digest " is the same column.
	if said := refusals(t, "file, Digest \n0421.tif,anything\n"); len(said) != 1 {
		t.Errorf("a manifest with a \" Digest \" column reported %d refusal(s): %v", len(said), said)
	}
}

// TestEveryProblemIsReportedInOnePass is #65's second condition and the reason
// the refusal type carries rows. A researcher fixing a two thousand row file
// one error per run stops before the fourth run.
func TestEveryProblemIsReportedInOnePass(t *testing.T) {
	said := refusals(t, `id,file,plate,,plate
a,0421.tif,0421,x,0421
b,,0422,x,0422
,0423.tif,0423,x,0423
d,0421.tif,0424,x,0424
a,0425.tif,0425,x,0425
`)
	rows := map[int]int{}
	for _, refusal := range said {
		rows[refusal.Row]++
	}
	// The header carries two problems, an unnamed column and a repeated one,
	// and rows 3 to 6 carry one each.
	for row, want := range map[int]int{1: 2, 3: 1, 4: 1, 5: 1, 6: 1} {
		if rows[row] != want {
			t.Errorf("row %d reported %d refusal(s), want %d. All of them: %v", row, rows[row], want, said)
		}
	}
	if len(said) != 6 {
		t.Errorf("the manifest reported %d refusal(s) in total, want 6: %v", len(said), said)
	}

	// The message names the rows a reader has to look at, and says which
	// earlier row a repeat collides with.
	whole := ManifestRefused{Refusals: said}.Error()
	for _, wanted := range []string{"row 1 ", "row 3 ", "row 5 ", "row 6 ", "which row 2 already named"} {
		if !strings.Contains(whole, wanted) {
			t.Errorf("the refusal does not say %q:\n%s", wanted, whole)
		}
	}
}

// TestARowNamingAFileThatIsNotThereIsRefused and the test below are the other
// two awkward cases, and they are the pair Reconcile exists for. Neither is
// decidable while reading the manifest, because neither is a fact about the
// manifest: both are a comparison against what was actually found.
func TestARowNamingAFileThatIsNotThereIsRefused(t *testing.T) {
	manifest := read(t, "file\n0421.tif\n0422.tif\n")
	err := manifest.Reconcile([]string{"0421.tif"})
	var refused ManifestRefused
	if !errors.As(err, &refused) {
		t.Fatalf("a row naming a missing file was accepted: %v", err)
	}
	if len(refused.Refusals) != 1 {
		t.Fatalf("it reported %d refusal(s), want 1: %v", len(refused.Refusals), refused.Refusals)
	}
	if refused.Refusals[0].Row != 3 {
		t.Errorf("the refusal is on row %d, want 3", refused.Refusals[0].Row)
	}
	if !strings.Contains(refused.Refusals[0].Says, "0422.tif") {
		t.Errorf("the refusal does not name the file: %q", refused.Refusals[0].Says)
	}
}

// TestAFileWithNoRowIsRefusedRatherThanIngestedSilently is the decision this
// change takes where #65 asks for a defined behaviour. Ingesting it would put a
// subject in the campaign its owner never described, and its export row would
// carry an identifier and nothing else.
func TestAFileWithNoRowIsRefusedRatherThanIngestedSilently(t *testing.T) {
	manifest := read(t, "file\n0421.tif\n")
	err := manifest.Reconcile([]string{"0421.tif", "0422.tif", "0423.tif"})
	var refused ManifestRefused
	if !errors.As(err, &refused) {
		t.Fatalf("two files nobody described were accepted: %v", err)
	}
	if len(refused.Refusals) != 2 {
		t.Fatalf("it reported %d refusal(s), want 2: %v", len(refused.Refusals), refused.Refusals)
	}
	// The order is the collection's rather than a map's, so two runs report
	// the same list.
	var named []string
	for _, refusal := range refused.Refusals {
		if refusal.Row != 0 {
			t.Errorf("a refusal about a file with no row carries row %d", refusal.Row)
		}
		named = append(named, refusal.Says)
	}
	if !sort.StringsAreSorted(named) {
		t.Errorf("the files with no row are reported in no fixed order: %v", named)
	}

	if err := manifest.Reconcile([]string{"0421.tif"}); err != nil {
		t.Errorf("a manifest that matches the collection exactly was refused: %v", err)
	}
}

// TestACollectionWithNoManifestStillRuns is #65's fourth condition. A working
// group with a folder and no spreadsheet runs a campaign, and the export
// carries the file names they already use.
func TestACollectionWithNoManifestStillRuns(t *testing.T) {
	manifest, err := Unmanifested([]string{"0421.tif", "0422.tif"})
	if err != nil {
		t.Fatalf("a collection with no manifest was refused: %v", err)
	}
	if manifest.Len() != 2 {
		t.Fatalf("it holds %d subject(s), want 2", manifest.Len())
	}
	for _, declared := range manifest.Subjects() {
		if declared.ID == "" {
			t.Error("a subject with no manifest carries no identifier")
		}
		if len(declared.Metadata) != 0 {
			t.Errorf("%s carries metadata nobody wrote: %v", declared.ID, declared.Metadata)
		}
	}
	if err := manifest.Reconcile([]string{"0421.tif", "0422.tif"}); err != nil {
		t.Errorf("the collection it was built from was refused: %v", err)
	}

	if _, err := Unmanifested([]string{"0421.tif", "0421.tif"}); err == nil {
		t.Error("a collection holding one file twice was accepted")
	}
	if _, err := Unmanifested([]string{"  "}); err == nil {
		t.Error("a collection holding a file with no name was accepted")
	}
}

// TestAManifestWithNothingUsableIsRefusedBeforeItsRows holds the two cases
// where reading the rows would be reporting the same missing column two
// thousand times.
func TestAManifestWithNothingUsableIsRefusedBeforeItsRows(t *testing.T) {
	if said := refusals(t, ""); len(said) != 1 {
		t.Errorf("an empty manifest reported %d refusal(s): %v", len(said), said)
	}
	said := refusals(t, "id,plate\na,0421\nb,0422\nc,0423\n")
	if len(said) != 1 {
		t.Fatalf("a manifest with no file column reported %d refusal(s): %v", len(said), said)
	}
	if !strings.Contains(said[0].Says, FileColumn) {
		t.Errorf("the refusal does not name the column that is missing: %q", said[0].Says)
	}

	// A row with a different number of fields from the header is the other
	// one, and the reader itself refuses it rather than filling in a blank.
	if said := refusals(t, "file,plate\n0421.tif\n"); len(said) != 1 {
		t.Errorf("a short row reported %d refusal(s): %v", len(said), said)
	}
}

// TestTheManifestCannotBeWrittenThroughAfterwards holds the copies the readers
// return, the same property the judged types in this package hold.
func TestTheManifestCannotBeWrittenThroughAfterwards(t *testing.T) {
	manifest := read(t, "file,plate\n0421.tif,0421\n")
	manifest.Subjects()[0].ID = "changed"
	manifest.Files()[0] = "changed.tif"
	if got := manifest.Subjects()[0].ID; got != "0421.tif" {
		t.Errorf("the identifier is now %q", got)
	}
	if got := manifest.Files()[0]; got != "0421.tif" {
		t.Errorf("the file is now %q", got)
	}
}
