// Package model is the model interface: what a model is given when it scores a
// subject, what a proposal is, and how a proposal is recorded, under the rules
// docs/model-boundary.md and docs/model-visibility.md fix.
//
// It may read the core's words because it answers the core's tasks. It may not
// read the surface or the store: a proposal reaches a page only through
// main.go, which is what keeps docs/model-visibility.md's position provable
// rather than merely intended.
//
// There is no code in this package yet. #54 trains the first model and #62
// records what was proposed.
package model
