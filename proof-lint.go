// This file exists to make the three legs of the lint check refuse, each for
// the reason it names. It lives on a proof branch that is never merged.
//
// It compiles on purpose, so the build check stays green and the refusals come
// from the legs under test rather than from a compile error underneath them.
package main

import "fmt"

func proofLint( ) {
    // gofmt: the parameter list carries a space and the body is indented with
    // spaces rather than a tab.
    //
    // go vet: %d is given a string.
    fmt.Printf("%d\n", "not a number")

    // staticcheck: a comparison with a boolean constant (S1002), in a function
    // nothing calls (U1000).
    flag := true
    if flag == true {
        fmt.Println("reached")
    }
}
