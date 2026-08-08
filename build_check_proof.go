// This file exists to be refused by the build check. It is never merged.
//
// It imports a package the module file does not carry. With dependency
// resolution disabled the build refuses it rather than quietly adding a
// requirement, which is the shape a stale lock file takes in this toolchain.
package main

import _ "golang.org/x/text/language"
