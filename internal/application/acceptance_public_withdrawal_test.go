package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/repository/memory"
)

func TestWithdrawnArticleIsAbsentFromPublicKeywordSearch(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	for _, status := range []manuscript.Status{manuscript.StatusSubmitted, manuscript.StatusUnderInitial, manuscript.StatusUnderFinal, manuscript.StatusAccepted} {
		if err := target.Transition(status, "editor1", "editorial progression", target.ActiveVersion, fixture.now); err != nil {
			t.Fatal(err)
		}
	}
	stored, err := fixture.manuscripts.Get(context.Background(), target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.manuscripts.Update(context.Background(), target, stored.Version); err != nil {
		t.Fatal(err)
	}
	publicationStore := memory.NewPublicationStore()
	issues := memory.NewIssueStore(publicationStore)
	articles := memory.NewArticleStore(publicationStore)
	service := application.NewPublicationService(fixture.manuscripts, issues, articles, fixture.users, fixture.audits, application.DirectTransactionManager{}, application.NewSequentialIDs("public-withdrawal"), clock.Fixed(fixture.now))
	issue, err := service.CreateIssue(context.Background(), "editor1", 18, 2, "研究方法专刊", fixture.now.Add(24*time.Hour), "Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	article, err := service.Schedule(context.Background(), "editor1", target.ID, issue.ID, 1, 1, "scheduled after acceptance")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.PublishIssue(context.Background(), "editor1", issue.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.Withdraw(context.Background(), "editor1", article.ID, "source data cannot be verified"); err != nil {
		t.Fatal(err)
	}
	page, err := service.Search(context.Background(), application.PublicationFilter{Query: "可复现研究", Page: 1, PageSize: 20, Sort: "published_at"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("withdrawn article remains publicly searchable: %#v", page)
	}
}
