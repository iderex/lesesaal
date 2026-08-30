package campaign

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
)

// FileColumn is the one column a manifest has to carry: the name of the file
// the row is about. Everything else is optional or the campaign owner's own.
const FileColumn = "file"

// IDColumn is the column naming the campaign owner's identifier for the
// subject. Where it is absent the file name is the identifier, which is what
// makes a manifest optional at all.
const IDColumn = "id"

// byteOrderMark is what a spreadsheet writes at the very start of a file it
// saved as comma-separated values, and it is not part of the first column's
// name. It is written as an escape rather than as the character, because a
// tracked text file in this tree may not carry one: the unicode leg of the gate
// refuses it by name, for the reason docs are refused it, and a guard that made
// an exception for its own source would be no guard.
const byteOrderMark = "\ufeff"

// DerivedColumns are the column names a manifest may not use, because this
// project computes a field of that name at ingest and two fields under one name
// is an export column a reader cannot resolve.
//
// It is a function rather than a variable for the reason TaskTypes is: a
// package-level slice can be written to by anything that imports this package,
// and the set a refusal quotes is the last thing that should be mutable.
func DerivedColumns() []string {
	return []string{"digest", "bytes", "width", "height", "entered"}
}

// ManifestRefusal is one thing wrong with a manifest. Row is the row it is on,
// counting the header as row 1 the way a spreadsheet does, and is zero where
// the refusal is about the manifest as a whole.
type ManifestRefusal struct {
	Row  int
	Says string
}

// ManifestRefused is what reading a manifest returns when it cannot be used.
//
// It carries every refusal rather than the first. A researcher fixing a two
// thousand row file one error per run will stop, which is #65's own sentence
// and the reason this type exists beside Refused rather than reusing it: a
// refusal about a spreadsheet has to say which row.
type ManifestRefused struct {
	Refusals []ManifestRefusal
}

// Error writes one refusal per line, each naming its row where it has one.
func (m ManifestRefused) Error() string {
	lines := make([]string, 0, len(m.Refusals))
	for _, refusal := range m.Refusals {
		if refusal.Row == 0 {
			lines = append(lines, refusal.Says)
			continue
		}
		lines = append(lines, fmt.Sprintf("row %d %s", refusal.Row, refusal.Says))
	}
	return fmt.Sprintf("this manifest cannot be used:\n%s", strings.Join(lines, "\n"))
}

// Manifest is what a campaign owner wrote about their collection: one row per
// subject, a column naming the file, and every other column carried through
// untouched.
type Manifest struct {
	subjects []DeclaredSubject
	files    []string
}

// Subjects is one declared subject per row, in the order the rows were written.
func (m Manifest) Subjects() []DeclaredSubject {
	return append([]DeclaredSubject(nil), m.subjects...)
}

// Files is the file each row names, in the same order as Subjects.
func (m Manifest) Files() []string { return append([]string(nil), m.files...) }

// Len is how many rows the manifest carries.
func (m Manifest) Len() int { return len(m.subjects) }

// ReadManifest reads a manifest and reports everything wrong with it in one
// pass.
//
// The format is comma-separated values with a header row, because that is what
// a spreadsheet writes when a researcher chooses "save as" and it needs no tool
// from this project to produce. A byte order mark at the start is dropped
// rather than becoming part of the first column's name, which is what the
// spreadsheet most researchers have writes by default.
//
// Column names are compared with their case and their surrounding space
// removed, so "File" and " file " are the column this project requires; the
// name a campaign owner wrote is what an export carries.
func ReadManifest(r io.Reader) (Manifest, error) {
	rows, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return Manifest{}, ManifestRefused{Refusals: []ManifestRefusal{{
			Says: fmt.Sprintf("it does not read as comma-separated values: %v. Every row has to carry the same number of fields as the header", err),
		}}}
	}
	if len(rows) == 0 {
		return Manifest{}, ManifestRefused{Refusals: []ManifestRefusal{{
			Says: "it is empty, and a manifest with no header row names no columns",
		}}}
	}

	var refusals []ManifestRefusal
	refuse := func(row int, says string) {
		refusals = append(refusals, ManifestRefusal{Row: row, Says: says})
	}

	header := rows[0]
	header[0] = strings.TrimPrefix(header[0], byteOrderMark)
	names := make([]string, len(header))
	columnAt := map[string]int{}
	derived := map[string]bool{}
	for _, name := range DerivedColumns() {
		derived[name] = true
	}
	for i, written := range header {
		name := strings.ToLower(strings.TrimSpace(written))
		names[i] = name
		switch {
		case name == "":
			refuse(1, fmt.Sprintf("names column %d with nothing, and a column nobody named cannot be read back out of an export", i+1))
		case derived[name]:
			refuse(1, fmt.Sprintf("names a column %q, which is a field this project derives from the bytes at ingest. Two fields under one name is an export column a reader cannot resolve", written))
		default:
			if _, twice := columnAt[name]; twice {
				refuse(1, fmt.Sprintf("names the column %q twice", written))
			}
		}
		columnAt[name] = i
	}
	fileAt, named := columnAt[FileColumn]
	if !named {
		refuse(1, fmt.Sprintf("names no %q column, and nothing else says which file a row is about", FileColumn))
	}
	idAt, identified := columnAt[IDColumn]

	manifest := Manifest{}
	seenFile := map[string]int{}
	seenID := map[string]int{}
	for i, row := range rows[1:] {
		number := i + 2
		if !named {
			continue
		}

		file := strings.TrimSpace(row[fileAt])
		if file == "" {
			refuse(number, "names no file")
		} else if before, twice := seenFile[file]; twice {
			refuse(number, fmt.Sprintf("names the file %q, which row %d already named. One file is one subject, and two rows about it would be two subjects behind one plate", file, before))
		} else {
			seenFile[file] = number
		}

		id := file
		if identified {
			id = strings.TrimSpace(row[idAt])
			if id == "" {
				refuse(number, fmt.Sprintf("names no identifier, and the manifest carries an %q column. Leave the column out to use the file name, or fill it in", IDColumn))
			}
		}
		if id != "" {
			if before, twice := seenID[id]; twice {
				refuse(number, fmt.Sprintf("names the identifier %q, which row %d already named. An identifier is unique within a campaign", id, before))
			}
			seenID[id] = number
		}

		metadata := make([]Field, 0, len(names))
		for column, name := range names {
			if name == FileColumn || name == IDColumn || name == "" {
				continue
			}
			metadata = append(metadata, Field{Name: header[column], Value: row[column]})
		}
		manifest.subjects = append(manifest.subjects, DeclaredSubject{ID: id, Metadata: metadata})
		manifest.files = append(manifest.files, file)
	}

	if len(refusals) > 0 {
		return Manifest{}, ManifestRefused{Refusals: refusals}
	}
	return manifest, nil
}

// Reconcile compares the manifest against the files that were actually found
// and reports every row and every file that does not have a partner.
//
// Both directions are refusals rather than one being one, and the second is the
// one worth arguing with. A row naming a file that is not there is obviously
// wrong. A file with no row is refused as well, because the manifest is the
// campaign owner's statement of what their campaign is about: ingesting a file
// they did not describe would put a subject in the campaign nobody wrote a line
// about, and its row in the export would carry an identifier and nothing else.
// Reporting it lets them add the row or move the file, which are the two things
// they might have meant. A collection with no manifest at all is a different
// case and is Unmanifested.
func (m Manifest) Reconcile(files []string) error {
	found := map[string]bool{}
	for _, file := range files {
		found[strings.TrimSpace(file)] = true
	}
	listed := map[string]bool{}

	var refusals []ManifestRefusal
	for i, file := range m.files {
		listed[file] = true
		if !found[file] {
			refusals = append(refusals, ManifestRefusal{
				Row:  i + 2,
				Says: fmt.Sprintf("names the file %q, which is not in the collection", file),
			})
		}
	}

	var missing []string
	for file := range found {
		if !listed[file] {
			missing = append(missing, file)
		}
	}
	sort.Strings(missing)
	for _, file := range missing {
		refusals = append(refusals, ManifestRefusal{
			Says: fmt.Sprintf("the collection holds %q and no row names it, so nothing says what it is", file),
		})
	}

	if len(refusals) > 0 {
		return ManifestRefused{Refusals: refusals}
	}
	return nil
}

// Unmanifested is the collection with no manifest at all: every file is a
// subject, the file name is the identifier, and nothing else is known.
//
// This is what makes a manifest optional rather than nearly optional. A working
// group that has a folder and no spreadsheet runs a campaign, and the export
// carries the file names they already use.
func Unmanifested(files []string) (Manifest, error) {
	var refusals []ManifestRefusal
	manifest := Manifest{}
	seen := map[string]bool{}
	for _, written := range files {
		file := strings.TrimSpace(written)
		switch {
		case file == "":
			refusals = append(refusals, ManifestRefusal{Says: "the collection holds a file with no name"})
		case seen[file]:
			refusals = append(refusals, ManifestRefusal{Says: fmt.Sprintf("the collection holds %q twice", file)})
		default:
			seen[file] = true
			manifest.subjects = append(manifest.subjects, DeclaredSubject{ID: file})
			manifest.files = append(manifest.files, file)
		}
	}
	if len(refusals) > 0 {
		return Manifest{}, ManifestRefused{Refusals: refusals}
	}
	return manifest, nil
}
