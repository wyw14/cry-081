package application_test

import (
	"context"
	"testing"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
)

func TestSubmittedManuscriptCannotSkipInitialReview(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	declaration := manuscript.Declaration{OriginalWork: true, AuthorApproved: true, EthicsCleared: true, Text: "all authors approve"}
	if _, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{
		ManuscriptID: target.ID, AuthorID: target.AuthorID, ExpectedVersion: target.Version,
		IdempotencyKey: "initial-review-gate", Declaration: declaration,
	}); err != nil {
		t.Fatal(err)
	}
	stored, err := fixture.manuscripts.Get(context.Background(), target.ID)
	if err != nil {
		t.Fatal(err)
	}
	moveErr := fixture.reviewService.MoveToFinalReview(context.Background(), stored.ID, "editor1", stored.Version)
	if moveErr == nil {
		t.Fatal("submitted manuscript skipped directly to final review")
	}
	after, err := fixture.manuscripts.Get(context.Background(), target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Status != manuscript.StatusSubmitted || after.Version != stored.Version {
		t.Fatalf("rejected transition changed manuscript: status=%s version=%d", after.Status, after.Version)
	}
}
