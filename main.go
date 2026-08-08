// Command lesesaal is the entry point of the deployment.
//
// Today it does one thing: it prints its version and exits. It exists in that
// state deliberately, as the smallest artefact that proves the toolchain chosen
// in issue #17 builds and runs from a clean checkout, and as the thing the
// build check in #19 has to compile.
//
// What a version number promises here is not decided. Issue #118 decides it,
// and the value below is a placeholder that says so rather than a number
// claiming a release exists.
package main

import (
	"fmt"
	"os"
)

// version is a placeholder until the release route in #118 sets it at build
// time. It is deliberately not a plausible release number.
var version = "0.0.0-dev"

func main() {
	fmt.Fprintf(os.Stdout, "lesesaal %s\n", version)
}
