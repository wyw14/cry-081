package application_test

import (
	"context"
	"testing"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
)

func TestSubmissionIdempotencyIsScopedPerManuscript(t *testing.T) {
	fixture := newWorkflowFixture(t)
	first := fixture.createDraft(t)
	second, err := fixture.manuscriptService.Create(context.Background(), application.CreateManuscriptInput{
		AuthorID: "author1", SectionID: "research", Title: "Independent replication",
		Abstract: "A second submission with independent evidence",
		Markdown: testBody() + " This second manuscript has distinct observations and conclusions.",
		Tags:     []string{"replication"},
	})
	if err != nil {
		t.Fatal(err)
	}
	declaration := manuscript.Declaration{OriginalWork: true, AuthorApproved: true, EthicsCleared: true, Text: "all authors approve"}
	firstSnapshot, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{
		ManuscriptID: first.ID, AuthorID: first.AuthorID, ExpectedVersion: first.Version,
		IdempotencyKey: "author-browser-retry", Declaration: declaration,
	})
	if err != nil {
		t.Fatal(err)
	}
	secondSnapshot, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{
		ManuscriptID: second.ID, AuthorID: second.AuthorID, ExpectedVersion: second.Version,
		IdempotencyKey: "author-browser-retry", Declaration: declaration,
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstSnapshot.ManuscriptID == secondSnapshot.ManuscriptID || secondSnapshot.ManuscriptID != second.ID {
		t.Fatalf("submission replay crossed manuscripts: first=%s second=%s", firstSnapshot.ManuscriptID, secondSnapshot.ManuscriptID)
	}
	stored, err := fixture.manuscripts.Get(context.Background(), second.ID)
	if err != nil || stored.Status != manuscript.StatusSubmitted {
		t.Fatalf("second manuscript was not submitted: status=%s err=%v", stored.Status, err)
	}
}
