package campaign

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// noon is the moment every case here stamps, from a clock that does not move.
// A rule that read a clock would be a test of the machine it ran on, which is
// what Depends exists to stop.
var noon = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

// stored is a write that keeps what it was given, standing in for the store
// #33 brings. It is a slice rather than a map because the order the writes
// arrived in is a thing several cases here assert on.
type stored struct {
	written []Classification
	refuse  error
}

// write is the function Record calls, and the only route a classification takes
// into this fake.
func (s *stored) write(classification Classification) error {
	if s.refuse != nil {
		return s.refuse
	}
	s.written = append(s.written, classification)
	return nil
}

// oneAnswer is a submission of one answer to a one-task campaign.
func oneAnswer(key string, volunteer string, subject string, options ...string) Submission {
	return Submission{
		Key:       key,
		Volunteer: volunteer,
		Subject:   subject,
		Answers:   []Answer{append(Answer(nil), options...)},
	}
}

// TestARepeatedSubmissionProducesOneClassification is the condition the issue
// puts first, and it is the shape of the failure rather than a generic
// duplicate: a volunteer clicks, the connection drops, and the client sends the
// same submission again under the same key.
func TestARepeatedSubmissionProducesOneClassification(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()
	submission := oneAnswer("k1", "v1", "s1", "clear")

	first, written, err := classifications.Record(submission, noon, store.write)
	if err != nil {
		t.Fatalf("the first arrival was refused: %v", err)
	}
	if !written {
		t.Error("the first arrival reported that it wrote nothing")
	}

	again, written, err := classifications.Record(submission, noon.Add(time.Second), store.write)
	if err != nil {
		t.Fatalf("the retry was refused, and a volunteer whose connection dropped would be told their answer did not count: %v", err)
	}
	if written {
		t.Error("the retry wrote a second time")
	}

	if len(store.written) != 1 {
		t.Errorf("the store took %d write(s) for one submission sent twice", len(store.written))
	}
	if classifications.Count("s1") != 1 {
		t.Errorf("the subject carries %d classification(s) after one submission sent twice, which is what the retirement rule would read", classifications.Count("s1"))
	}

	// The retry returns the first result rather than a fresh one, which is
	// what makes the second acknowledgement true. The moment is the first
	// moment, not the retry's.
	if again.Key() != first.Key() || !again.At().Equal(first.At()) {
		t.Errorf("the retry returned %s, and the first arrival recorded %s", again, first)
	}
}

// TestAKeyReusedForADifferentAnswerIsRefused holds the half a key alone does
// not buy. A retry and a client reusing its key for the next subject arrive
// identically, and taking the second one as a retry would silently drop an
// answer somebody gave.
func TestAKeyReusedForADifferentAnswerIsRefused(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	if _, _, err := classifications.Record(oneAnswer("k1", "v1", "s1", "clear"), noon, store.write); err != nil {
		t.Fatalf("the first arrival was refused: %v", err)
	}

	_, written, err := classifications.Record(oneAnswer("k1", "v1", "s2", "marked"), noon, store.write)
	if err == nil {
		t.Fatal("a key already recorded was admitted for a different subject, and the answer under it was discarded as a retry")
	}
	if written {
		t.Error("a refused submission reported that it wrote")
	}
	if !strings.Contains(err.Error(), "reused") {
		t.Errorf("the refusal does not say what was wrong: %v", err)
	}

	// An answer that differs only in the order of its option identifiers is a
	// retry rather than a different answer, because docs/vocabulary.md makes
	// an answer a set.
	several := Submission{Key: "k2", Volunteer: "v1", Subject: "s3", Answers: []Answer{{"a", "b"}}}
	if _, _, err := classifications.Record(several, noon, store.write); err != nil {
		t.Fatalf("the first arrival of a two-option answer was refused: %v", err)
	}
	reordered := Submission{Key: "k2", Volunteer: "v1", Subject: "s3", Answers: []Answer{{"b", "a"}}}
	if _, written, err := classifications.Record(reordered, noon, store.write); err != nil {
		t.Errorf("a retry that reordered the option identifiers was refused: %v", err)
	} else if written {
		t.Error("a retry that reordered the option identifiers wrote a second time")
	}
}

// TestAVolunteerCannotAnswerOneSubjectTwice is the issue's last condition. A
// second answer under a key of its own is not a retry, and admitting it would
// put one person into the consensus rule as two.
func TestAVolunteerCannotAnswerOneSubjectTwice(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	if _, _, err := classifications.Record(oneAnswer("k1", "v1", "s1", "clear"), noon, store.write); err != nil {
		t.Fatalf("the first answer was refused: %v", err)
	}

	_, _, err := classifications.Record(oneAnswer("k2", "v1", "s1", "marked"), noon, store.write)
	if err == nil {
		t.Fatal("one volunteer answered one subject twice, and the second answer counts as a second volunteer agreeing")
	}
	if !strings.Contains(err.Error(), "k1") {
		t.Errorf("the refusal does not name the key that already answered: %v", err)
	}
	if len(store.written) != 1 {
		t.Errorf("the store took %d write(s) where one answer was admitted", len(store.written))
	}

	// A second volunteer on the same subject is the ordinary case and is not
	// what the rule refuses.
	if _, _, err := classifications.Record(oneAnswer("k3", "v2", "s1", "clear"), noon, store.write); err != nil {
		t.Errorf("a second volunteer answering the same subject was refused: %v", err)
	}
	if classifications.Count("s1") != 2 {
		t.Errorf("the subject carries %d classification(s) where two volunteers answered it", classifications.Count("s1"))
	}
}

// TestACampaignMayLetAVolunteerAnswerAgain is the other half of that condition.
// The campaign says they may, and the key still makes a retry one
// classification under either rule.
func TestACampaignMayLetAVolunteerAnswerAgain(t *testing.T) {
	store := &stored{}
	classifications := LettingAVolunteerAnswerAgain()

	if _, _, err := classifications.Record(oneAnswer("k1", "v1", "s1", "clear"), noon, store.write); err != nil {
		t.Fatalf("the first answer was refused: %v", err)
	}
	if _, _, err := classifications.Record(oneAnswer("k2", "v1", "s1", "marked"), noon, store.write); err != nil {
		t.Fatalf("a second answer was refused by a campaign that admits one: %v", err)
	}
	if classifications.Count("s1") != 2 {
		t.Errorf("the subject carries %d classification(s) where two answers were admitted", classifications.Count("s1"))
	}

	if _, written, err := classifications.Record(oneAnswer("k1", "v1", "s1", "clear"), noon, store.write); err != nil {
		t.Errorf("a retry was refused under the rule that admits repeats: %v", err)
	} else if written {
		t.Error("a retry wrote a second time under the rule that admits repeats")
	}
}

// TestAWriteThatFailsIsNotAcknowledgedAndTheKeyStaysFree is the second
// condition of the issue, and it is the position stated in the type comment
// rather than whichever branch happened to fall through.
func TestAWriteThatFailsIsNotAcknowledgedAndTheKeyStaysFree(t *testing.T) {
	broken := errors.New("the store did not commit")
	store := &stored{refuse: broken}
	classifications := OneAnswerPerVolunteerPerSubject()
	submission := oneAnswer("k1", "v1", "s1", "clear")

	_, written, err := classifications.Record(submission, noon, store.write)
	if err == nil {
		t.Fatal("a submission whose write failed was acknowledged, and the volunteer believes their work counted")
	}
	if written {
		t.Error("a submission whose write failed reported that it wrote")
	}
	if !errors.Is(err, broken) {
		t.Errorf("the refusal does not carry what the store said: %v", err)
	}

	// Nothing was registered, so the volunteer's retry is a first arrival.
	if classifications.Count("s1") != 0 {
		t.Errorf("a failed write left the subject carrying %d classification(s)", classifications.Count("s1"))
	}

	store.refuse = nil
	if _, written, err := classifications.Record(submission, noon.Add(time.Second), store.write); err != nil {
		t.Fatalf("the retry after a failed write was refused, and the answer is lost with nothing said about it: %v", err)
	} else if !written {
		t.Error("the retry after a failed write did not write, so the answer is acknowledged and stored nowhere")
	}
	if len(store.written) != 1 {
		t.Errorf("the store holds %d write(s) after one failure and one retry", len(store.written))
	}
}

// TestTheAnswerIsRecordedBeforeAnythingReactsToIt holds the transaction
// boundary the issue asks to be written down. The write of the answer and the
// recomputation that follows it are two steps and not one, so a failure between
// them leaves the answer recorded and the label behind, which is a state
// Label.Stale already names and which is repaired by recomputing from the
// answers rather than by asking the volunteer again.
func TestTheAnswerIsRecordedBeforeAnythingReactsToIt(t *testing.T) {
	task := plateCondition(t)
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	for i, who := range []string{"v1", "v2"} {
		key := string(rune('a' + i))
		if _, _, err := classifications.Record(oneAnswer(key, who, "s1", "clear"), noon, store.write); err != nil {
			t.Fatalf("%s was refused: %v", who, err)
		}
	}

	// The label a caller computed after the second answer and before the third
	// is the state a crash at the boundary leaves behind.
	label := Consensus(task, classifications.AnswersTo("s1", 0), DefaultAgreement())
	if _, _, err := classifications.Record(oneAnswer("c", "v3", "s1", "marked"), noon, store.write); err != nil {
		t.Fatalf("the third answer was refused: %v", err)
	}

	why := label.Stale(classifications.Count("s1"), DefaultAgreement())
	if len(why) == 0 {
		t.Fatal("a label computed before the third answer was not reported stale after it, so a crash between the write and the recomputation leaves nothing that says so")
	}
	if !strings.Contains(why[0], "now carries 3") {
		t.Errorf("the staleness does not name the count the subject now carries: %v", why)
	}

	// Recomputing from the answers alone repairs it. Nothing was asked of the
	// volunteers a second time, which is the property that makes recording
	// first the right order.
	repaired := Consensus(task, classifications.AnswersTo("s1", 0), DefaultAgreement())
	if len(repaired.Stale(classifications.Count("s1"), DefaultAgreement())) != 0 {
		t.Errorf("the recomputed label is still stale: %s", repaired)
	}
}

// TestASubmissionThatIsNotOneIsRefusedWithEveryReason holds the shape refusals.
// Each is a client defect a campaign owner writing their own client will meet,
// and reporting all of them at once is what stops them learning the shape one
// round trip at a time.
func TestASubmissionThatIsNotOneIsRefusedWithEveryReason(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	_, _, err := classifications.Record(Submission{}, noon, store.write)
	if err == nil {
		t.Fatal("an empty submission was recorded")
	}

	var refused NotRecorded
	if !errors.As(err, &refused) {
		t.Fatalf("the refusal is not a NotRecorded, so a caller cannot read the reasons apart: %v", err)
	}
	if len(refused.Says) != 4 {
		t.Errorf("an empty submission produced %d reason(s), and it is wrong in four ways", len(refused.Says))
	}
	if strings.Contains(err.Error(), "definition") {
		t.Errorf("a refused submission reads as a campaign definition problem: %v", err)
	}
	if len(store.written) != 0 {
		t.Errorf("a refused submission reached the store %d time(s)", len(store.written))
	}

	for _, c := range []struct {
		name       string
		submission Submission
	}{
		{"no key", Submission{Volunteer: "v1", Subject: "s1", Answers: []Answer{{"clear"}}}},
		{"no volunteer", Submission{Key: "k", Subject: "s1", Answers: []Answer{{"clear"}}}},
		{"no subject", Submission{Key: "k", Volunteer: "v1", Answers: []Answer{{"clear"}}}},
		{"no answer", Submission{Key: "k", Volunteer: "v1", Subject: "s1"}},
	} {
		if _, _, err := classifications.Record(c.submission, noon, store.write); err == nil {
			t.Errorf("a submission with %s was recorded", c.name)
		}
	}
}

// TestTheRegistersARecordingUpdatesAreTheOnesTheOtherRulesRead is what keeps
// this type and the two rules that read it from parting. The selection rule
// takes Pool.Answered and the retirement rule takes a count, and both come from
// here rather than from a second reading of the same rows.
func TestTheRegistersARecordingUpdatesAreTheOnesTheOtherRulesRead(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	for _, c := range []struct{ key, volunteer, subject string }{
		{"k1", "v1", "s1"},
		{"k2", "v1", "s2"},
		{"k3", "v2", "s1"},
	} {
		if _, _, err := classifications.Record(oneAnswer(c.key, c.volunteer, c.subject, "clear"), noon, store.write); err != nil {
			t.Fatalf("%s was refused: %v", c.key, err)
		}
	}

	answered := classifications.Answered("v1")
	if len(answered) != 2 || !answered["s1"] || !answered["s2"] {
		t.Errorf("the volunteer has answered %v, and the draw would exclude that set", answered)
	}
	if answered := classifications.Answered("v3"); len(answered) != 0 {
		t.Errorf("a volunteer who has answered nothing carries %v", answered)
	}

	if classifications.Count("s1") != 2 || classifications.Count("s2") != 1 {
		t.Errorf("the counts are s1=%d s2=%d", classifications.Count("s1"), classifications.Count("s2"))
	}
	if classifications.Count("s3") != 0 {
		t.Errorf("a subject nobody answered carries %d", classifications.Count("s3"))
	}
}

// TestTheReadingOfOneSubjectIsInTheOrderItWasRecorded holds what makes a
// campaign replayable. Map order is not an order, and a reading that took one
// would produce a different answer set on every run over the same rows.
func TestTheReadingOfOneSubjectIsInTheOrderItWasRecorded(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	for i, who := range []string{"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8"} {
		key := string(rune('a' + i))
		at := noon.Add(time.Duration(i) * time.Minute)
		if _, _, err := classifications.Record(oneAnswer(key, who, "s1", who), at, store.write); err != nil {
			t.Fatalf("%s was refused: %v", who, err)
		}
	}

	want := "v1 v2 v3 v4 v5 v6 v7 v8"
	for run := 0; run < 8; run++ {
		var got []string
		for _, classification := range classifications.Of("s1") {
			got = append(got, classification.Volunteer())
		}
		if strings.Join(got, " ") != want {
			t.Fatalf("run %d read %s", run, strings.Join(got, " "))
		}
	}

	// Two arrivals at one moment, which an injected clock that does not move
	// produces on every campaign these tests run, are ordered by key.
	same := OneAnswerPerVolunteerPerSubject()
	for _, key := range []string{"kb", "ka"} {
		if _, _, err := same.Record(oneAnswer(key, key, "s1", "clear"), noon, store.write); err != nil {
			t.Fatalf("%s was refused: %v", key, err)
		}
	}
	read := same.Of("s1")
	if len(read) != 2 || read[0].Key() != "ka" || read[1].Key() != "kb" {
		t.Errorf("two arrivals at one moment read back as %v", read)
	}
}

// TestTheAnswersOfATaskAreWhatTheConsensusRuleTakes is the join between this
// type and the rule that turns its rows into a label. The position is the
// task's position in the campaign's task order, which is what a classification
// carries one answer per.
func TestTheAnswersOfATaskAreWhatTheConsensusRuleTakes(t *testing.T) {
	condition := plateCondition(t)
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	for i, said := range [][]string{{"clear", "a"}, {"clear", "b"}, {"clear", "a"}} {
		key := string(rune('a' + i))
		submission := Submission{
			Key:       key,
			Volunteer: key,
			Subject:   "s1",
			Answers:   []Answer{{said[0]}, {said[1]}},
		}
		if _, _, err := classifications.Record(submission, noon, store.write); err != nil {
			t.Fatalf("%s was refused: %v", key, err)
		}
	}

	label := Consensus(condition, classifications.AnswersTo("s1", 0), DefaultAgreement())
	if !label.Labelled() || strings.Join(label.Options(), " ") != "clear" {
		t.Errorf("the first task's answers produced %s", label)
	}
	if label.Answers() != 3 {
		t.Errorf("the label was computed from %d answer(s)", label.Answers())
	}

	// The second task's answers are a different set, read at a different
	// position out of the same classifications.
	second := classifications.AnswersTo("s1", 1)
	if len(second) != 3 || strings.Join(second[0], "") != "a" || strings.Join(second[1], "") != "b" {
		t.Errorf("the second task's answers read back as %v", second)
	}

	// A position no task occupies reads back empty rather than panicking, so a
	// caller that lost track of the task order gets a label with no answers
	// behind it instead of a dead process.
	if beyond := classifications.AnswersTo("s1", 2); len(beyond) != 0 {
		t.Errorf("a position past the last task read back %v", beyond)
	}
}

// TestARecordedClassificationCannotBeRewrittenThroughWhatItReturns holds the
// copy every reader in this package makes. A caller holding the answers could
// otherwise rewrite what a volunteer said after it was recorded, and the export
// would carry the rewrite with nothing marking it.
func TestARecordedClassificationCannotBeRewrittenThroughWhatItReturns(t *testing.T) {
	store := &stored{}
	classifications := OneAnswerPerVolunteerPerSubject()

	submission := oneAnswer("k1", "v1", "s1", "clear")
	recorded, _, err := classifications.Record(submission, noon, store.write)
	if err != nil {
		t.Fatalf("the submission was refused: %v", err)
	}

	// The caller's own slice, reused for the next submission.
	submission.Answers[0][0] = "unusable"
	if got := classifications.Of("s1"); len(got) != 1 || strings.Join(got[0].Answers()[0], "") != "clear" {
		t.Errorf("rewriting the submission's slice changed what was recorded: %v", got)
	}

	// The slice the reader was handed.
	recorded.Answers()[0][0] = "marked"
	if got := classifications.Of("s1"); strings.Join(got[0].Answers()[0], "") != "clear" {
		t.Errorf("rewriting the returned answers changed what was recorded: %v", got)
	}
}
