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
// What is here so far is the dependencies the program takes rather than reads,
// the campaign definition boundary, which is the types a campaign owner's
// written definition becomes and the refusal of one that asks something this
// project does not support, the consensus rule, which is the answer, the label
// and the count that turns a subject's answers into one, the campaign state
// machine, which is what a transition does and which ones are refused, and the
// retirement rule, which is when a subject has had enough classifications. The
// rest of #32 is the subject and the classification, and #34 records one.
package campaign
