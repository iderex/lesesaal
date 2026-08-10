package campaign

import (
	"fmt"
	"strings"
)

// TaskType is the shape of the question a task asks. docs/task-grammar.md is
// where the set was decided and it is a closed set: every task in the first
// release declares a fixed list of options and an answer is a subset of that
// list, so the type says only how large a subset may be.
//
// A yes or no question is deliberately not a member. It is a single choice over
// two options, and keeping it out is what holds the number of things needing a
// rendering, a comparison, a consensus rule and an export column at two.
type TaskType string

const (
	// SingleChoice admits exactly one option.
	SingleChoice TaskType = "single-choice"
	// ChoiceOfSeveral admits between the bounds the campaign owner declared.
	ChoiceOfSeveral TaskType = "choice-of-several"
)

// TaskTypes is the supported set, in the order a refusal lists them. It is a
// function rather than a variable because a package-level slice can be written
// to by anything that imports this package, and the set a refusal quotes is the
// last thing that should be mutable.
func TaskTypes() []TaskType {
	return []TaskType{SingleChoice, ChoiceOfSeveral}
}

// The three declared types below are what a campaign owner wrote, before
// anything has judged it. Their fields are exported and every one of them is
// the plain form a file carries, because this is the shape a parser fills in
// and a parser has nothing to validate with.
//
// The concrete syntax of that file is not decided anywhere yet, and this
// boundary does not decide it either. docs/task-grammar.md fixes that a
// definition is a file a person writes by hand and leaves the syntax to the
// issue that chooses it, so what is written here is the shape such a file
// produces rather than the file.

// DeclaredCampaign is one campaign as it was written down.
type DeclaredCampaign struct {
	ID           string
	Title        string
	Instructions string
	Tasks        []DeclaredTask
}

// DeclaredTask is one question as it was written down. Type is a string rather
// than a TaskType because the whole point of this boundary is that a campaign
// owner can write something that is not one, and a field that could only hold a
// member of the set would move that refusal to whoever filled the field in.
type DeclaredTask struct {
	ID       string
	Question string
	Type     string
	Options  []DeclaredOption

	// AtLeast and AtMost are pointers so that "not written" and "zero" are
	// different states. docs/task-grammar.md asks for both bounds to be written
	// by the campaign owner rather than assumed, precisely so that "at least
	// one" and "any number including none" are different declarations; an int
	// here would make the second of those the value somebody gets by leaving
	// the field out.
	AtLeast *int
	AtMost  *int
}

// DeclaredOption is one choice as it was written down.
type DeclaredOption struct {
	ID   string
	Text string
}

// The three judged types below are what a written definition becomes. Their
// fields are unexported and nothing outside this package can fill them in, so
// the only way to hold one is to have passed Define. That is an invariant
// carried by the type system rather than by everything downstream remembering
// to check.

// Campaign is one campaign, judged.
type Campaign struct {
	id           string
	title        string
	instructions string
	tasks        []Task
}

// Task is one question of a campaign, judged.
type Task struct {
	id       string
	question string
	kind     TaskType
	options  []Option
	atLeast  int
	atMost   int
}

// Option is one choice of a task, judged. The identifier and the text are
// separate for the reason docs/task-grammar.md gives: an owner rewrites the
// wording of an option while a campaign runs, and an answer that carried the
// text would split one option into two in the export with nothing saying so.
type Option struct {
	id   string
	text string
}

// The readers below are how a judged value is read, because it cannot be
// reached into. That is what keeps the fields unexported and therefore keeps
// the constructor the only way in. Each returns a copy where what it holds is
// a slice, so a caller cannot write through the result into a campaign that
// has already been judged.

func (c Campaign) ID() string           { return c.id }
func (c Campaign) Title() string        { return c.title }
func (c Campaign) Instructions() string { return c.instructions }
func (c Campaign) Tasks() []Task        { return append([]Task(nil), c.tasks...) }

func (t Task) ID() string       { return t.id }
func (t Task) Question() string { return t.question }
func (t Task) Type() TaskType   { return t.kind }
func (t Task) Options() []Option {
	return append([]Option(nil), t.options...)
}

// Bounds is how many options an answer to this task may name, lower first. A
// single choice answers one and only one, so it reports 1 and 1 rather than
// reporting that it has no bounds: a caller comparing an answer against a task
// then has one rule to write instead of two.
func (t Task) Bounds() (atLeast int, atMost int) { return t.atLeast, t.atMost }

func (o Option) ID() string   { return o.id }
func (o Option) Text() string { return o.text }

// Refusal is one thing wrong with a definition. Task names the task it is
// about, or is empty where the refusal is about the campaign as a whole, so
// that a campaign owner is told which of forty tasks to look at rather than
// being told that something somewhere is wrong.
type Refusal struct {
	Task string
	Says string
}

// Refused is what Define returns when a definition does not start. It carries
// every refusal rather than the first, because a campaign owner fixing one
// problem per run learns the shape of the grammar one round trip at a time.
type Refused struct {
	Refusals []Refusal
}

// Error writes one refusal per line, each naming its task where it has one.
func (r Refused) Error() string {
	lines := make([]string, 0, len(r.Refusals))
	for _, refusal := range r.Refusals {
		if refusal.Task == "" {
			lines = append(lines, refusal.Says)
			continue
		}
		lines = append(lines, fmt.Sprintf("task %q %s", refusal.Task, refusal.Says))
	}
	return fmt.Sprintf("this campaign definition does not start:\n%s", strings.Join(lines, "\n"))
}

// Define judges a written definition and returns the campaign it declares.
//
// This is the boundary docs/task-grammar.md puts the refusal at, and it is the
// only one: a task outside the grammar is refused before a campaign starts
// rather than half supported, because a type that renders and has no comparison
// rule produces answers that can never become a label, which is worse than a
// campaign that would not open.
func Define(declared DeclaredCampaign) (Campaign, error) {
	var refusals []Refusal
	refuse := func(task string, says string) {
		refusals = append(refusals, Refusal{Task: task, Says: says})
	}

	if strings.TrimSpace(declared.ID) == "" {
		refuse("", "the campaign names no identifier")
	}
	if strings.TrimSpace(declared.Title) == "" {
		refuse("", "the campaign names no title")
	}
	if len(declared.Tasks) == 0 {
		refuse("", "the campaign declares no task, and a campaign with nothing to ask a volunteer about a subject is not a campaign")
	}

	tasks := make([]Task, 0, len(declared.Tasks))
	seen := map[string]bool{}
	for _, declaredTask := range declared.Tasks {
		if seen[declaredTask.ID] && declaredTask.ID != "" {
			refuse(declaredTask.ID, "is declared twice, so an answer naming it would name two questions")
		}
		seen[declaredTask.ID] = true
		tasks = append(tasks, judge(declaredTask, refuse))
	}

	if len(refusals) > 0 {
		return Campaign{}, Refused{Refusals: refusals}
	}
	return Campaign{
		id:           declared.ID,
		title:        declared.Title,
		instructions: declared.Instructions,
		tasks:        tasks,
	}, nil
}

// judge reads one task and reports what is wrong with it. It returns the task
// it would have built either way, because the caller throws all of them away as
// soon as one refusal exists and a second pass to build them would be a second
// place for the two to disagree.
func judge(declared DeclaredTask, refuse func(task string, says string)) Task {
	if strings.TrimSpace(declared.ID) == "" {
		refuse("", "a task names no identifier, so nothing an answer carries could say which question it answered")
	}
	if strings.TrimSpace(declared.Question) == "" {
		refuse(declared.ID, "asks no question")
	}

	kind := TaskType(declared.Type)
	known := false
	for _, supported := range TaskTypes() {
		if kind == supported {
			known = true
		}
	}
	if !known {
		refuse(declared.ID, fmt.Sprintf("declares the type %q, which this project does not support. The supported types are %s, and docs/task-grammar.md says what each of the refused shapes would need before it could be one of them", declared.Type, list(TaskTypes())))
	}

	options := make([]Option, 0, len(declared.Options))
	seen := map[string]bool{}
	for _, declaredOption := range declared.Options {
		switch {
		case strings.TrimSpace(declaredOption.ID) == "":
			refuse(declared.ID, "declares an option with no identifier, and the identifier is the only part of an option an answer carries")
		case strings.TrimSpace(declaredOption.Text) == "":
			refuse(declared.ID, fmt.Sprintf("declares the option %q with no text, so a volunteer would be asked to choose between blanks", declaredOption.ID))
		case seen[declaredOption.ID]:
			refuse(declared.ID, fmt.Sprintf("declares the option %q twice, so an answer naming it would be ambiguous and two options would fold into one in the export", declaredOption.ID))
		}
		seen[declaredOption.ID] = true
		options = append(options, Option{id: declaredOption.ID, text: declaredOption.Text})
	}
	if len(options) < 2 {
		refuse(declared.ID, fmt.Sprintf("declares %d option(s), and a question with fewer than two is one every volunteer answers the same way", len(options)))
	}

	atLeast, atMost := bounds(declared, kind, len(options), refuse)
	return Task{
		id:       declared.ID,
		question: declared.Question,
		kind:     kind,
		options:  options,
		atLeast:  atLeast,
		atMost:   atMost,
	}
}

// bounds decides how many options an answer to this task may name, and refuses
// a declaration that wrote them where they mean nothing or left them out where
// they mean everything.
//
// A single choice carrying bounds is refused rather than ignored. The grammar
// says what a task of each type names and that nothing may be added, and a
// bound written on a single choice is a campaign owner believing they have
// constrained something; ignoring it would leave that belief intact and
// unwritten.
func bounds(declared DeclaredTask, kind TaskType, options int, refuse func(task string, says string)) (int, int) {
	if kind == SingleChoice {
		if declared.AtLeast != nil || declared.AtMost != nil {
			refuse(declared.ID, "is a single choice and declares bounds, which say nothing here: a single choice admits exactly one option. A task whose bounds are meant to hold is a choice of several")
		}
		return 1, 1
	}
	if kind != ChoiceOfSeveral {
		return 0, 0
	}

	if declared.AtLeast == nil || declared.AtMost == nil {
		refuse(declared.ID, "is a choice of several and does not write both bounds. Neither has a default: \"at least one\" and \"any number including none\" are different questions and this project will not guess which was meant")
		return 0, 0
	}
	atLeast, atMost := *declared.AtLeast, *declared.AtMost
	switch {
	case atLeast < 0:
		refuse(declared.ID, fmt.Sprintf("declares a lower bound of %d, and an answer cannot name fewer than no options", atLeast))
	case atLeast > atMost:
		refuse(declared.ID, fmt.Sprintf("declares a lower bound of %d above its upper bound of %d, which no answer satisfies", atLeast, atMost))
	case atMost > options:
		refuse(declared.ID, fmt.Sprintf("declares an upper bound of %d above the %d option(s) it declares, so the bound is not the one that holds", atMost, options))
	}
	return atLeast, atMost
}

// list writes a set of types the way a refusal quotes it.
func list(types []TaskType) string {
	quoted := make([]string, 0, len(types))
	for _, one := range types {
		quoted = append(quoted, fmt.Sprintf("%q", string(one)))
	}
	return strings.Join(quoted, " and ")
}
