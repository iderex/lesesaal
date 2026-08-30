package campaigntest

import (
	"fmt"
	"time"

	"github.com/iderex/lesesaal/internal/campaign"
)

// PlateCondition is a campaign whose one task is a single choice. It is the
// shape almost every campaign this project is built for has: one question, one
// answer, three options a volunteer can hold in their head at once.
func PlateCondition() campaign.DeclaredCampaign {
	return campaign.DeclaredCampaign{
		ID:           "plate-condition",
		Title:        "The condition of the plates",
		Instructions: "Look at the plate and say what state it is in.",
		Tasks: []campaign.DeclaredTask{{
			ID:       "condition",
			Question: "What condition is this plate in?",
			Type:     string(campaign.SingleChoice),
			Options: []campaign.DeclaredOption{
				{ID: "clear", Text: "Clear"},
				{ID: "marked", Text: "Marked"},
				{ID: "unusable", Text: "Unusable"},
			},
		}},
	}
}

// PlateDefects is a campaign whose one task is a choice of several, with both
// bounds written the way docs/task-grammar.md requires. A plate can carry
// several defects at once and can carry none, which is why the lower bound is
// zero here rather than one.
func PlateDefects() campaign.DeclaredCampaign {
	none, four := 0, 4
	return campaign.DeclaredCampaign{
		ID:           "plate-defects",
		Title:        "What is wrong with the plates",
		Instructions: "Name every defect you can see, or none.",
		Tasks: []campaign.DeclaredTask{{
			ID:       "defects",
			Question: "Which of these can you see on this plate?",
			Type:     string(campaign.ChoiceOfSeveral),
			Options: []campaign.DeclaredOption{
				{ID: "scratch", Text: "A scratch"},
				{ID: "emulsion-loss", Text: "Emulsion loss"},
				{ID: "fogging", Text: "Fogging"},
				{ID: "annotation", Text: "Handwriting on the plate"},
			},
			AtLeast: &none,
			AtMost:  &four,
		}},
	}
}

// Fixtures is one campaign per task type the grammar declares, in the order
// campaign.TaskTypes returns them.
//
// It is a function over that set rather than a list somebody keeps in step with
// it, so a task type added to the grammar with no fixture campaign beside it is
// a failing suite rather than a gap nobody sees. A driver that ran every
// fixture and covered two of three types would report the same green as one
// that covered all of them.
func Fixtures() []campaign.DeclaredCampaign {
	byType := map[campaign.TaskType]campaign.DeclaredCampaign{
		campaign.SingleChoice:    PlateCondition(),
		campaign.ChoiceOfSeveral: PlateDefects(),
	}
	fixtures := make([]campaign.DeclaredCampaign, 0, len(byType))
	for _, kind := range campaign.TaskTypes() {
		declared, held := byType[kind]
		if !held {
			panic(fmt.Sprintf("campaigntest: the grammar declares the task type %q and no fixture campaign asks a question of that type", kind))
		}
		fixtures = append(fixtures, declared)
	}
	return fixtures
}

// Plates builds n subjects with no bytes behind them, named plate-1 upwards in
// the order ingest would have taken them.
//
// The derived fields are the smallest values Enter admits rather than plausible
// ones. A fixture carrying a 50 MB size and a 14000 pixel width would be
// inviting a reader to believe a test loaded something, and nothing here has
// bytes at all: the digest is a number written out to the length a digest has,
// which is what makes two fixture subjects distinguishable without either of
// them standing for a file.
func Plates(n int, entered time.Time) ([]campaign.Subject, error) {
	if entered.IsZero() {
		entered = Epoch
	}
	subjects := make([]campaign.Subject, 0, n)
	for i := 1; i <= n; i++ {
		subject, err := campaign.Enter(
			campaign.DeclaredSubject{
				ID: fmt.Sprintf("plate-%d", i),
				Metadata: []campaign.Field{
					{Name: "plate", Value: fmt.Sprintf("%04d", i)},
				},
			},
			campaign.Derived{
				Digest:  fmt.Sprintf("sha256:%064d", i),
				Bytes:   1,
				Width:   1,
				Height:  1,
				Entered: entered,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("building fixture subject %d: %w", i, err)
		}
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

// VolunteerNames builds n volunteer identifiers, v1 upwards.
//
// A volunteer is an identifier here and nothing else. What this project may
// hold about a person is undecided, which #13 is open for, and a fixture that
// carried a name or an address would be this suite assuming an answer to it.
func VolunteerNames(n int) []string {
	names := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		names = append(names, fmt.Sprintf("v%d", i))
	}
	return names
}
