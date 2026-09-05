// Package store is the storage layer: the embedded store the one process in
// docs/deployment-ceiling.md holds, and the reads and writes the core asks for.
//
// It may read the core's words because it stores them. It may not read the
// surface or the model, so a change to either cannot reach into what is
// persisted without passing through main.go first.
//
// There is no code in this package yet. #33 chooses the store and its schema.
//
// The rule that makes a classification arrive exactly once is #34's and is in
// internal/campaign, because it decides what a second arrival means rather than
// how a row is written. What is owed here is the half that rule cannot hold: a
// key registered in memory is lost with the process, so the store carries the
// client's key as its own unique column and refuses a second row under one key
// whatever this process remembers.
package store
