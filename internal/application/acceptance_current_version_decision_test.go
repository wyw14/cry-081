package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/review"
)

func TestFinalDecisionUsesReportsFromCurrentRevision(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	declaration := manuscript.Declaration{OriginalWork: true, AuthorApproved: true, EthicsCleared: true, Text: "all authors approve"}
	if _, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{ManuscriptID: target.ID, AuthorID: target.AuthorID, ExpectedVersion: target.Version, IdempotencyKey: "decision-v1", Declaration: declaration}); err != nil {
		t.Fatal(err)
	}
	stored, err := fixture.manuscripts.Get(context.Background(), target.ID)
	if err != nil {
		t.Fatal(err)
	}
	beforeReview := stored.Version
	if err := stored.Transition(manuscript.StatusUnderInitial, "editor1", "initial review started", stored.ActiveVersion, fixture.now); err != nil {
		t.Fatal(err)
	}
	if err := stored.Transition(manuscript.StatusRevisionNeeded, "editor1", "revise methods", stored.ActiveVersion, fixture.now); err != nil {
		t.Fatal(err)
	}
	if err := fixture.manuscripts.Update(context.Background(), stored, beforeReview); err != nil {
		t.Fatal(err)
	}
	revision, err := manuscript.NewDraftVersion(2, "可复现研究修订稿", "补充方法说明", testBody()+" The revised version addresses every requested methodological clarification with supporting detail.", []string{"research", "revision"}, nil, fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	updated, err := fixture.manuscriptService.AddRevision(context.Background(), stored.ID, stored.AuthorID, stored.Version, revision)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{ManuscriptID: updated.ID, AuthorID: updated.AuthorID, ExpectedVersion: updated.Version, IdempotencyKey: "decision-v2", Declaration: declaration}); err != nil {
		t.Fatal(err)
	}
	current, err := fixture.manuscripts.Get(context.Background(), target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.reviewService.MoveToFinalReview(context.Background(), current.ID, "editor1", current.Version); err != nil {
		t.Fatal(err)
	}
	current, err = fixture.manuscripts.Get(context.Background(), target.ID)
	if err != nil {
		t.Fatal(err)
	}
	assignment, err := fixture.reviewService.Assign(context.Background(), current.ID, "reviewer1", "editor1", fixture.now.Add(48*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := fixture.reviews.Get(context.Background(), assignment.ID)
	if err != nil {
		t.Fatal(err)
	}
	assignmentVersion := accepted.Version
	if err := accepted.Accept("reviewer1", fixture.now); err != nil {
		t.Fatal(err)
	}
	if err := fixture.reviews.Update(context.Background(), accepted, assignmentVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.reviewService.SubmitReport(context.Background(), accepted.ID, accepted.ReviewerID, review.RecommendAccept, []review.ScoreItem{{Criterion: "methodology", Score: 5}}, []review.StructuredComment{{ID: "current-version-comment", Section: "methods", Comment: "revision resolves the concern"}}, "ready for final decision"); err != nil {
		t.Fatal(err)
	}
	decision, err := fixture.reviewService.FinalDecision(context.Background(), current.ID, "editor1", current.Version, review.DecisionAccept, "current revision satisfies the journal standard")
	if err != nil {
		t.Fatalf("final decision: %v", err)
	}
	if decision.ManuscriptVersion != 2 {
		t.Fatalf("decision targeted version %d instead of 2", decision.ManuscriptVersion)
	}
	decided, err := fixture.manuscripts.Get(context.Background(), current.ID)
	if err != nil || decided.Status != manuscript.StatusAccepted {
		t.Fatalf("current revision was not accepted: status=%s err=%v", decided.Status, err)
	}
}
