// Command lesesaal is the entry point of the deployment.
//
// Today it does two things: it builds the dependencies the rest of the program
// will take rather than read, and it prints its version. It exists in that
// state deliberately, as the smallest artefact that proves the toolchain chosen
// in issue #17 builds and runs from a clean checkout, and as the thing the
// build check in #19 has to compile.
//
// The wiring is the point of the first of those. docs/harness.md says that time,
// randomness, identifiers and outbound connections are supplied here and read
// nowhere else, and harness_test.go refuses a departure from it. Nothing
// consumes the set yet, because there is nothing to consume it, so the whole of
// the wiring's job today is to build it and to refuse to start on one that is
// incomplete.
//
// What a version number promises here is not decided. Issue #118 decides it,
// and the value below is a placeholder that says so rather than a number
// claiming a release exists.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/iderex/lesesaal/internal/gate"
	"github.com/iderex/lesesaal/internal/system"
)

// version is a placeholder until the release route in #118 sets it at build
// time. It is deliberately not a plausible release number.
var version = "0.0.0-dev"

func main() {
	// The gate before anything else, because it takes none of the wiring
	// below and a contributor running it on a half-built tree should get the
	// gate's verdict rather than a complaint about a dependency.
	//
	// `ci` with no leg named runs the whole set in order and stops at the
	// first failure. `ci <leg>` runs one and requires it, which is the form a
	// workflow step calls so that one red leg cannot hide the next.
	if len(os.Args) > 1 && os.Args[1] == "ci" {
		leg := ""
		if len(os.Args) > 2 {
			leg = os.Args[2]
		}
		os.Exit(gate.Run(os.Stdout, gate.Env{Run: system.Run, Look: system.Look, Read: system.Read}, leg))
	}

	// A nil field is a dependency nobody supplied. A program that starts on one
	// panics at the first call rather than at startup, which is the wrong end
	// of a run to find out, so this refuses to start instead.
	depends := system.Depends()
	if missing := depends.Missing(); len(missing) != 0 {
		fmt.Fprintf(os.Stderr, "lesesaal: the wiring supplied no %s\n", strings.Join(missing, ", "))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "lesesaal %s\n", version)
}
