// Package store is the storage layer: the embedded store the one process in
// docs/deployment-ceiling.md holds, and the reads and writes the core asks for.
//
// It may read the core's words because it stores them. It may not read the
// surface or the model, so a change to either cannot reach into what is
// persisted without passing through main.go first.
//
// There is no code in this package yet. #33 chooses the store and its schema,
// #34 makes a classification arrive exactly once.
package store
