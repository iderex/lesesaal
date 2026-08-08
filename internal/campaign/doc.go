// Package campaign is the campaign core: campaigns, subjects, tasks, options,
// answers, classifications, labels and retirement, in the words
// docs/vocabulary.md fixes them.
//
// It holds the rules that decide what a classification means and when a
// subject is finished. It imports nothing else in this module, and that is the
// property docs/layout.md exists to keep: the core runs in a test with no
// browser, no images, no model and no store behind it.
//
// Where the core needs something from outside itself, it declares the
// interface here and main.go supplies the implementation.
//
// There is no code in this package yet. #32 models the domain and #34 records
// a classification.
package campaign
