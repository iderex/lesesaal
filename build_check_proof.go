// This file exists to be refused by the build check. It is never merged.
//
// The variable below is declared and never used, which this toolchain refuses
// as an error rather than reporting as a warning. It is the one-character
// mistake somebody actually makes, rather than a file that could not have
// compiled under any circumstances.
package main

func buildCheckProof() {
	unusedByDesign := 1
}
