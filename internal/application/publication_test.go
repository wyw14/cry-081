package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/publication"
	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/repository/memory"
)

func TestIssueReleasePublishesArticleAndManuscriptTogether(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	for _, status := range []manuscript.Status{manuscript.StatusSubmitted, manuscript.StatusUnderInitial, manuscript.StatusUnderFinal, manuscript.StatusAccepted} {
		if err := target.Transition(status, "editor1", "workflow progression", target.ActiveVersion, fixture.now); err != nil {
			t.Fatal(err)
		}
	}
	stored, _ := fixture.manuscripts.Get(context.Background(), target.ID)
	if err := fixture.manuscripts.Update(context.Background(), target, stored.Version); err != nil {
		t.Fatal(err)
	}
	publicationStore := memory.NewPublicationStore()
	issues := memory.NewIssueStore(publicationStore)
	articles := memory.NewArticleStore(publicationStore)
	service := application.NewPublicationService(fixture.manuscripts, issues, articles, fixture.users, fixture.audits, application.DirectTransactionManager{}, application.NewSequentialIDs("publication"), clock.Fixed(fixture.now))
	issue, err := service.CreateIssue(context.Background(), "editor1", 12, 3, "可复现研究专刊", fixture.now.Add(30*24*time.Hour), "Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	article, err := service.Schedule(context.Background(), "editor1", target.ID, issue.ID, 1, 1, "accepted for the research section")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.PublishIssue(context.Background(), "editor1", issue.ID); err != nil {
		t.Fatal(err)
	}
	published, _ := articles.Get(context.Background(), article.ID)
	updated, _ := fixture.manuscripts.Get(context.Background(), target.ID)
	if published.Status != publication.ArticlePublished || updated.Status != manuscript.StatusPublished {
		t.Fatalf("release was partial: article=%s manuscript=%s", published.Status, updated.Status)
	}
	page, err := service.Search(context.Background(), application.PublicationFilter{Query: "可复现", Page: 1, PageSize: 10, Sort: "published_at"})
	if err != nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("published article was not searchable: %#v %v", page, err)
	}
}

func TestWithdrawalRequiresReasonAndRemovesArticleFromSearch(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	for _, status := range []manuscript.Status{manuscript.StatusSubmitted, manuscript.StatusUnderInitial, manuscript.StatusUnderFinal, manuscript.StatusAccepted} {
		_ = target.Transition(status, "editor1", "workflow progression", target.ActiveVersion, fixture.now)
	}
	stored, _ := fixture.manuscripts.Get(context.Background(), target.ID)
	_ = fixture.manuscripts.Update(context.Background(), target, stored.Version)
	publicationStore := memory.NewPublicationStore()
	issues := memory.NewIssueStore(publicationStore)
	articles := memory.NewArticleStore(publicationStore)
	service := application.NewPublicationService(fixture.manuscripts, issues, articles, fixture.users, fixture.audits, application.DirectTransactionManager{}, application.NewSequentialIDs("publication"), clock.Fixed(fixture.now))
	issue, _ := service.CreateIssue(context.Background(), "editor1", 12, 4, "方法研究", fixture.now.Add(time.Hour), "Asia/Shanghai")
	article, _ := service.Schedule(context.Background(), "editor1", target.ID, issue.ID, 1, 1, "scheduled after acceptance")
	_ = service.PublishIssue(context.Background(), "editor1", issue.ID)
	if err := service.Withdraw(context.Background(), "editor1", article.ID, ""); err == nil {
		t.Fatal("withdrawal without reason succeeded")
	}
	if err := service.Withdraw(context.Background(), "editor1", article.ID, "author disclosed a material correction"); err != nil {
		t.Fatal(err)
	}
	page, _ := service.Search(context.Background(), application.PublicationFilter{Page: 1, PageSize: 10})
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("withdrawn article remained public: %#v", page)
	}
}
