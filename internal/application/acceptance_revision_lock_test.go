package application_test

import (
	"context"
	"testing"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
)

func TestRevisionCreatesEditableVersionWithoutUnlockingSubmission(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	declaration := manuscript.Declaration{OriginalWork: true, AuthorApproved: true, EthicsCleared: true, Text: "all authors approve"}
	if _, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{ManuscriptID: target.ID, AuthorID: target.AuthorID, ExpectedVersion: target.Version, IdempotencyKey: "revision-lock", Declaration: declaration}); err != nil {
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
	if err := stored.Transition(manuscript.StatusRevisionNeeded, "editor1", "methods need clarification", stored.ActiveVersion, fixture.now); err != nil {
		t.Fatal(err)
	}
	if err := fixture.manuscripts.Update(context.Background(), stored, beforeReview); err != nil {
		t.Fatal(err)
	}
	revision, err := manuscript.NewDraftVersion(2, "可复现研究修订稿", "补充抽样和质量控制说明", testBody()+" The revision explains the requested methodological changes and the author response in detail.", []string{"research", "methods"}, nil, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := fixture.manuscriptService.AddRevision(context.Background(), stored.ID, stored.AuthorID, stored.Version, revision)
	if err != nil {
		t.Fatalf("add revision: %v", err)
	}
	if updated.ActiveVersion != 2 || len(updated.Versions) != 2 {
		t.Fatalf("revision was not added: %#v", updated)
	}
	if !updated.Versions[0].Locked || updated.Versions[1].Locked {
		t.Fatalf("version locks are wrong: old=%v new=%v", updated.Versions[0].Locked, updated.Versions[1].Locked)
	}
}
