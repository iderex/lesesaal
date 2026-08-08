package campaign

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestMissingNamesEveryUnsuppliedField is what the entry point refuses to start
// on. A set with a nil field panics at the first call to it, which can be a long
// way from the wiring that forgot it and long after the run began.
func TestMissingNamesEveryUnsuppliedField(t *testing.T) {
	missing := Depends{}.Missing()

	want := []string{"Now", "Intn", "NewID", "Dial"}
	if len(missing) != len(want) {
		t.Fatalf("an empty set reports %v missing and every field is", missing)
	}
	for i, field := range want {
		if missing[i] != field {
			t.Errorf("field %d of the report is %s and the declaration order says %s", i, missing[i], field)
		}
	}
}

// TestMissingReportsNothingOnACompleteSet is the other direction. A report that
// named a field which was supplied would stop the program from starting at all.
func TestMissingReportsNothingOnACompleteSet(t *testing.T) {
	complete := Depends{
		Now:   func() time.Time { return time.Time{} },
		Intn:  func(int) int { return 0 },
		NewID: func() string { return "" },
		Dial: func(context.Context, string, string) (net.Conn, error) {
			return nil, nil
		},
	}

	if missing := complete.Missing(); len(missing) != 0 {
		t.Errorf("a complete set reports %v missing", missing)
	}
}

// TestMissingNamesOnlyTheAbsentOne is the near miss between the two above: a
// report that named every field whenever any field was absent would send
// whoever read it looking at the wrong three.
func TestMissingNamesOnlyTheAbsentOne(t *testing.T) {
	withoutDial := Depends{
		Now:   func() time.Time { return time.Time{} },
		Intn:  func(int) int { return 0 },
		NewID: func() string { return "" },
	}

	missing := withoutDial.Missing()
	if len(missing) != 1 || missing[0] != "Dial" {
		t.Errorf("a set missing only Dial reports %v", missing)
	}
}
