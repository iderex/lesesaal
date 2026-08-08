package main

import (
	"testing"

	"github.com/iderex/lesesaal/internal/campaign/campaigntest"
	"github.com/iderex/lesesaal/internal/system"
)

// TestTheWiringSuppliesEveryField is what main refuses to start without. It is
// here rather than in internal/system because the wiring package may only be
// imported from the entry point, and this is the entry point's own suite.
func TestTheWiringSuppliesEveryField(t *testing.T) {
	if missing := system.Depends().Missing(); len(missing) != 0 {
		t.Errorf("the wiring supplies no %v, and main refuses to start on that", missing)
	}
}

// TestTheFakesSupplyWhatTheWiringSupplies is the one that earns this file. The
// two sets are built in different packages by different people at different
// times, and a fake set that has fallen behind the real one is a suite passing
// against a program shape that no longer exists. Comparing them here is cheap
// and it fails the day a field is added to one and not the other.
func TestTheFakesSupplyWhatTheWiringSupplies(t *testing.T) {
	wired := system.Depends().Missing()
	faked, _, _, _ := campaigntest.Depends()

	if len(wired) != 0 {
		t.Fatalf("the wiring itself is incomplete, missing %v, so there is nothing to compare against", wired)
	}
	if missing := faked.Missing(); len(missing) != 0 {
		t.Errorf("the fakes leave %v unsupplied while the wiring supplies every field", missing)
	}
}
