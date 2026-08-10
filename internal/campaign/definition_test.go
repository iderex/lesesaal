package campaign

import (
	"reflect"
	"strings"
	"testing"
)

// bound is the pointer form a campaign owner's written bound arrives as.
func bound(n int) *int { return &n }

// aSingleChoice is a task that is right in every respect, so that each test
// below can spoil exactly one thing about it.
func aSingleChoice() DeclaredTask {
	return DeclaredTask{
		ID:       "plate-condition",
		Question: "What condition is this plate in?",
		Type:     "single-choice",
		Options: []DeclaredOption{
			{ID: "intact", Text: "Intact"},
			{ID: "cracked", Text: "Cracked"},
		},
	}
}

// aChoiceOfSeveral is the other type, equally right.
func aChoiceOfSeveral() DeclaredTask {
	return DeclaredTask{
		ID:       "defects",
		Question: "Which of these can you see?",
		Type:     "choice-of-several",
		Options: []DeclaredOption{
			{ID: "scratch", Text: "A scratch"},
			{ID: "stain", Text: "A stain"},
			{ID: "fog", Text: "Fogging"},
		},
		AtLeast: bound(0),
		AtMost:  bound(3),
	}
}

// aCampaign carries both types, because a campaign is an ordered list of tasks
// and a fixture with one task would not notice a boundary that judged only the
// first.
func aCampaign() DeclaredCampaign {
	return DeclaredCampaign{
		ID:           "hamburg-plates-1904",
		Title:        "The Hamburg plate envelopes",
		Instructions: "Look at the plate and answer both questions.",
		Tasks:        []DeclaredTask{aSingleChoice(), aChoiceOfSeveral()},
	}
}

// refusedBy runs a definition that is expected not to start and returns what it
// said. A definition that starts is a failure here rather than a skipped
// assertion, because a test that silently passed on an accepted campaign would
// prove nothing about the refusal it is named for.
func refusedBy(t *testing.T, declared DeclaredCampaign) string {
	t.Helper()
	_, err := Define(declared)
	if err == nil {
		t.Fatal("the definition started, and this test is about it being refused")
	}
	return err.Error()
}

// TestADefinitionInsideTheGrammarStarts is the other half of every refusal
// below. Without it they would all pass against a boundary that refuses every
// campaign, which is a different rule from the one written down and one nobody
// could satisfy.
func TestADefinitionInsideTheGrammarStarts(t *testing.T) {
	campaign, err := Define(aCampaign())
	if err != nil {
		t.Fatalf("a definition inside the grammar was refused: %v", err)
	}
	if campaign.ID() != "hamburg-plates-1904" || campaign.Title() != "The Hamburg plate envelopes" {
		t.Errorf("the campaign does not carry what was declared: %q %q", campaign.ID(), campaign.Title())
	}
	tasks := campaign.Tasks()
	if len(tasks) != 2 {
		t.Fatalf("the campaign carries %d task(s) where 2 were declared", len(tasks))
	}
	if tasks[0].ID() != "plate-condition" || tasks[1].ID() != "defects" {
		t.Errorf("the tasks are not in the order they were declared: %q then %q", tasks[0].ID(), tasks[1].ID())
	}
	if tasks[0].Type() != SingleChoice || tasks[1].Type() != ChoiceOfSeveral {
		t.Errorf("the tasks did not keep their types: %q and %q", tasks[0].Type(), tasks[1].Type())
	}
	if text := tasks[0].Options()[1].Text(); text != "Cracked" {
		t.Errorf("an option did not keep the text a volunteer reads: %q", text)
	}
}

// TestATypeOutsideTheGrammarIsRefusedWithTheSupportedSet is the condition #6
// was left open on. A campaign owner learns the boundary from the refusal
// rather than from reading source, so the message has to name the offending
// task and list what is available.
func TestATypeOutsideTheGrammarIsRefusedWithTheSupportedSet(t *testing.T) {
	declared := aCampaign()
	declared.Tasks[1].Type = "bounding-box"

	says := refusedBy(t, declared)

	for _, want := range []string{`"defects"`, `"bounding-box"`, `"single-choice"`, `"choice-of-several"`, "docs/task-grammar.md"} {
		if !strings.Contains(says, want) {
			t.Errorf("the refusal does not carry %s:\n%s", want, says)
		}
	}
}

// TestTheOtherRefusedShapesAreRefusedTheSameWay walks the four shapes
// docs/task-grammar.md names as out. One of them is enough to prove the branch;
// all four are here because each is a thing a campaign owner will actually
// write, and a boundary that happened to admit one of them would be found by a
// campaign rather than by this suite.
func TestTheOtherRefusedShapesAreRefusedTheSameWay(t *testing.T) {
	for _, shape := range []string{"point", "bounding-box", "number", "text"} {
		declared := aCampaign()
		declared.Tasks[0].Type = shape

		says := refusedBy(t, declared)
		if !strings.Contains(says, `"`+shape+`"`) {
			t.Errorf("the refusal of %q does not name what was declared:\n%s", shape, says)
		}
	}
}

// TestAnEmptyTypeIsRefusedRatherThanDefaulted is the near miss for a boundary
// that treated the unset field as the commonest type. A campaign owner who
// forgot the line would get a single choice they never asked for, and every
// answer to it would be recorded as though they had.
func TestAnEmptyTypeIsRefusedRatherThanDefaulted(t *testing.T) {
	declared := aCampaign()
	declared.Tasks[0].Type = ""

	says := refusedBy(t, declared)
	if !strings.Contains(says, "single-choice") {
		t.Errorf("the refusal does not list the supported set:\n%s", says)
	}
}

func TestATaskWithFewerThanTwoOptionsIsRefused(t *testing.T) {
	declared := aCampaign()
	declared.Tasks[0].Options = declared.Tasks[0].Options[:1]

	says := refusedBy(t, declared)
	if !strings.Contains(says, "1 option(s)") {
		t.Errorf("the refusal does not say how many options were declared:\n%s", says)
	}
}

// TestAnOptionDeclaredTwiceIsRefused is the mistake that costs the most later.
// docs/task-grammar.md carries the identifier through the answer, the label and
// the export, so two options under one identifier fold into one column and the
// classifications either side of the fold stop being comparable.
func TestAnOptionDeclaredTwiceIsRefused(t *testing.T) {
	declared := aCampaign()
	declared.Tasks[1].Options[2].ID = "scratch"

	says := refusedBy(t, declared)
	if !strings.Contains(says, `"scratch"`) {
		t.Errorf("the refusal does not name the identifier declared twice:\n%s", says)
	}
}

func TestATaskDeclaredTwiceIsRefused(t *testing.T) {
	declared := aCampaign()
	declared.Tasks[1].ID = "plate-condition"

	says := refusedBy(t, declared)
	if !strings.Contains(says, "declared twice") {
		t.Errorf("the refusal does not say the task is declared twice:\n%s", says)
	}
}

func TestAnOptionWithNoIdentifierOrNoTextIsRefused(t *testing.T) {
	missingID := aCampaign()
	missingID.Tasks[0].Options[0].ID = "   "
	if says := refusedBy(t, missingID); !strings.Contains(says, "no identifier") {
		t.Errorf("the refusal does not say the option names no identifier:\n%s", says)
	}

	missingText := aCampaign()
	missingText.Tasks[0].Options[0].Text = ""
	if says := refusedBy(t, missingText); !strings.Contains(says, "no text") {
		t.Errorf("the refusal does not say the option carries no text:\n%s", says)
	}
}

// TestASingleChoiceThatDeclaresBoundsIsRefused is the case a boundary that
// ignored the field would pass. The campaign owner wrote a constraint, the
// software would keep none of it, and nothing would ever say so.
func TestASingleChoiceThatDeclaresBoundsIsRefused(t *testing.T) {
	declared := aCampaign()
	declared.Tasks[0].AtMost = bound(2)

	says := refusedBy(t, declared)
	if !strings.Contains(says, "choice of several") {
		t.Errorf("the refusal does not say what to declare instead:\n%s", says)
	}
}

// TestAChoiceOfSeveralMissingABoundIsRefused holds the reason the bounds are
// pointers. Zero is a bound a campaign owner writes on purpose, so a boundary
// reading an int could not tell "any number including none" from a field
// nobody filled in.
func TestAChoiceOfSeveralMissingABoundIsRefused(t *testing.T) {
	for _, spoil := range []func(*DeclaredTask){
		func(task *DeclaredTask) { task.AtLeast = nil },
		func(task *DeclaredTask) { task.AtMost = nil },
		func(task *DeclaredTask) { task.AtLeast, task.AtMost = nil, nil },
	} {
		declared := aCampaign()
		spoil(&declared.Tasks[1])

		says := refusedBy(t, declared)
		if !strings.Contains(says, "will not guess") {
			t.Errorf("the refusal does not say why neither bound has a default:\n%s", says)
		}
	}
}

func TestBoundsThatNoAnswerCouldSatisfyAreRefused(t *testing.T) {
	tooLow := aCampaign()
	tooLow.Tasks[1].AtLeast = bound(-1)
	if says := refusedBy(t, tooLow); !strings.Contains(says, "fewer than no options") {
		t.Errorf("a negative lower bound was not refused for what it is:\n%s", says)
	}

	inverted := aCampaign()
	inverted.Tasks[1].AtLeast, inverted.Tasks[1].AtMost = bound(3), bound(2)
	if says := refusedBy(t, inverted); !strings.Contains(says, "no answer satisfies") {
		t.Errorf("a lower bound above the upper one was not refused:\n%s", says)
	}

	beyond := aCampaign()
	beyond.Tasks[1].AtMost = bound(4)
	if says := refusedBy(t, beyond); !strings.Contains(says, "3 option(s)") {
		t.Errorf("an upper bound above the options declared was not refused:\n%s", says)
	}
}

func TestACampaignWithNoTaskOrNoIdentityIsRefused(t *testing.T) {
	empty := aCampaign()
	empty.Tasks = nil
	if says := refusedBy(t, empty); !strings.Contains(says, "declares no task") {
		t.Errorf("a campaign with nothing to ask was not refused:\n%s", says)
	}

	nameless := aCampaign()
	nameless.ID, nameless.Title = "", ""
	says := refusedBy(t, nameless)
	if !strings.Contains(says, "no identifier") || !strings.Contains(says, "no title") {
		t.Errorf("a campaign with neither identifier nor title was not refused for both:\n%s", says)
	}
}

// TestEveryRefusalIsReportedRatherThanTheFirst is what makes the boundary
// usable by hand. A campaign owner learning the grammar one refusal per run
// gets there eventually, and gets there having read the same message six times.
func TestEveryRefusalIsReportedRatherThanTheFirst(t *testing.T) {
	declared := aCampaign()
	declared.ID = ""
	declared.Tasks[0].Type = "number"
	declared.Tasks[1].AtMost = nil

	_, err := Define(declared)
	refused, isRefused := err.(Refused)
	if !isRefused {
		t.Fatalf("the definition was not refused with a Refused: %v", err)
	}
	if len(refused.Refusals) != 3 {
		t.Fatalf("3 things were wrong and %d were reported:\n%s", len(refused.Refusals), err)
	}
	for _, refusal := range refused.Refusals[1:] {
		if refusal.Task == "" {
			t.Errorf("a refusal about one task does not name it: %+v", refusal)
		}
	}
}

// TestASingleChoiceReportsBoundsOfOneAndOne keeps the caller writing one rule
// rather than two. A single choice has no bounds a campaign owner wrote, and
// reporting that it has none would put the special case in every comparison
// between an answer and its task.
func TestASingleChoiceReportsBoundsOfOneAndOne(t *testing.T) {
	campaign, err := Define(aCampaign())
	if err != nil {
		t.Fatalf("a definition inside the grammar was refused: %v", err)
	}
	if atLeast, atMost := campaign.Tasks()[0].Bounds(); atLeast != 1 || atMost != 1 {
		t.Errorf("a single choice reports bounds of %d and %d", atLeast, atMost)
	}
}

// TestAJudgedCampaignCannotBeWrittenThrough is the other half of the fields
// being unexported. A reader handed the slice the campaign holds could change
// a task in a campaign that has already been judged, which is the invariant
// leaving by the back door rather than through the constructor.
func TestAJudgedCampaignCannotBeWrittenThrough(t *testing.T) {
	campaign, err := Define(aCampaign())
	if err != nil {
		t.Fatalf("a definition inside the grammar was refused: %v", err)
	}

	tasks := campaign.Tasks()
	tasks[0] = Task{id: "something-else"}
	tasks[1].Options()[0] = Option{id: "also-something-else"}

	if got := campaign.Tasks()[0].ID(); got != "plate-condition" {
		t.Errorf("writing to the returned slice changed the campaign: %q", got)
	}
	if got := campaign.Tasks()[1].Options()[0].ID(); got != "scratch" {
		t.Errorf("writing to a returned option changed the campaign: %q", got)
	}
}

// TestAJudgedTypeHasNoExportedField is the guard behind the sentence that
// Define is the only way to hold one of these. The moment a field is exported
// for the convenience of the store or the surface, a caller can build a
// campaign that was never judged and every refusal above becomes optional. The
// repair is a reader, not an exported field.
func TestAJudgedTypeHasNoExportedField(t *testing.T) {
	for _, judged := range []any{Campaign{}, Task{}, Option{}} {
		of := reflect.TypeOf(judged)
		for index := range of.NumField() {
			if field := of.Field(index); field.IsExported() {
				t.Errorf("%s.%s is exported, so a %s can be built without passing Define", of.Name(), field.Name, of.Name())
			}
		}
	}
}
