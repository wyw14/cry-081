package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/review"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/platform/idempotency"
	"github.com/wyw14/cry-081/internal/platform/outbox"
	"github.com/wyw14/cry-081/internal/repository/memory"
)

type workflowFixture struct {
	now               time.Time
	users             *memory.IdentityStore
	manuscripts       *memory.ManuscriptStore
	reviews           *memory.ReviewStore
	reports           *memory.ReportStore
	audits            *memory.AuditStore
	manuscriptService *application.ManuscriptService
	reviewService     *application.ReviewService
}

func newWorkflowFixture(t *testing.T) workflowFixture {
	t.Helper()
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	users := memory.NewIdentityStore()
	for _, item := range []struct {
		id    string
		roles []identity.Role
	}{
		{"author1", []identity.Role{identity.RoleAuthor}},
		{"reviewer1", []identity.Role{identity.RoleEditor}},
		{"reviewer2", []identity.Role{identity.RoleEditor}},
		{"editor1", []identity.Role{identity.RoleChiefEditor}},
	} {
		user, err := identity.NewUser(item.id, item.id+"@example.test", item.id, item.roles, []byte("hash"), now)
		if err != nil || users.Create(context.Background(), *user) != nil {
			t.Fatalf("seed user %s: %v", item.id, err)
		}
	}
	manuscripts := memory.NewManuscriptStore()
	reviews := memory.NewReviewStore()
	reports := memory.NewReportStore(reviews)
	audits := memory.NewAuditStore()
	ids := application.NewSequentialIDs("test")
	fixed := clock.Fixed(now)
	return workflowFixture{
		now: now, users: users, manuscripts: manuscripts, reviews: reviews, reports: reports, audits: audits,
		manuscriptService: application.NewManuscriptService(manuscripts, users, audits, application.DirectTransactionManager{}, outbox.NewMemoryStore(), idempotency.NewMemory(), ids, fixed, manuscript.DefaultFormatPolicy()),
		reviewService:     application.NewReviewService(manuscripts, reviews, reports, memory.NewDecisionStore(reviews), users, audits, application.DirectTransactionManager{}, outbox.NewMemoryStore(), ids, fixed),
	}
}

func (f workflowFixture) createDraft(t *testing.T) manuscript.Manuscript {
	t.Helper()
	created, err := f.manuscriptService.Create(context.Background(), application.CreateManuscriptInput{AuthorID: "author1", SectionID: "research", Title: "可复现研究", Abstract: "关于可复现方法的摘要", Markdown: testBody(), Tags: []string{"research", "methods"}})
	if err != nil {
		t.Fatal(err)
	}
	return created
}

func TestSubmitCreatesLockedSnapshotAndRejectsDuplicateFingerprint(t *testing.T) {
	fixture := newWorkflowFixture(t)
	first := fixture.createDraft(t)
	declaration := manuscript.Declaration{OriginalWork: true, AuthorApproved: true, EthicsCleared: true, Text: "all authors approve"}
	snapshot, err := fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{ManuscriptID: first.ID, AuthorID: first.AuthorID, ExpectedVersion: first.Version, IdempotencyKey: "submission-1", Declaration: declaration})
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := fixture.manuscripts.Get(context.Background(), first.ID)
	if stored.Status != manuscript.StatusSubmitted || !stored.Versions[0].Locked || snapshot.ContentDigest == "" {
		t.Fatalf("submission did not close the snapshot transition: %#v", stored)
	}
	second := fixture.createDraft(t)
	_, err = fixture.manuscriptService.Submit(context.Background(), application.SubmitManuscriptInput{ManuscriptID: second.ID, AuthorID: second.AuthorID, ExpectedVersion: second.Version, IdempotencyKey: "submission-2", Declaration: declaration})
	if !errors.Is(err, shared.ErrConflict) {
		t.Fatalf("duplicate fingerprint was not rejected: %v", err)
	}
}

func TestAssignmentRejectsSelfReviewAndActiveConflict(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	if _, err := fixture.reviewService.Assign(context.Background(), target.ID, target.AuthorID, "editor1", fixture.now.Add(48*time.Hour)); !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("self review was not rejected: %v", err)
	}
	fixture.reviews.AddConflict(review.Conflict{EditorID: "reviewer1", AuthorID: "author1", Reason: "same institution", StartsAt: fixture.now.Add(-time.Hour)})
	if _, err := fixture.reviewService.Assign(context.Background(), target.ID, "reviewer1", "editor1", fixture.now.Add(48*time.Hour)); !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("active conflict was not rejected: %v", err)
	}
}

func TestConcurrentReviewReportsDoNotOverwriteEachOther(t *testing.T) {
	fixture := newWorkflowFixture(t)
	target := fixture.createDraft(t)
	assignments := make([]review.Assignment, 0, 2)
	for _, reviewerID := range []string{"reviewer1", "reviewer2"} {
		assignment, err := fixture.reviewService.Assign(context.Background(), target.ID, reviewerID, "editor1", fixture.now.Add(48*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		stored, _ := fixture.reviews.Get(context.Background(), assignment.ID)
		expected := stored.Version
		if err := stored.Accept(reviewerID, fixture.now); err != nil {
			t.Fatal(err)
		}
		if err := fixture.reviews.Update(context.Background(), stored, expected); err != nil {
			t.Fatal(err)
		}
		assignments = append(assignments, stored)
	}
	start := make(chan struct{})
	var wait sync.WaitGroup
	errorsChannel := make(chan error, 2)
	for i, assignment := range assignments {
		wait.Add(1)
		go func(index int, item review.Assignment) {
			defer wait.Done()
			<-start
			_, err := fixture.reviewService.SubmitReport(context.Background(), item.ID, item.ReviewerID, review.Recommendation([]review.Recommendation{review.RecommendAccept, review.RecommendMinorChanges}[index]), []review.ScoreItem{{Criterion: "originality", Score: 4}}, []review.StructuredComment{{ID: item.ID + "-comment", Section: "methods", Comment: "clear evidence", Required: false}}, "editor note")
			errorsChannel <- err
		}(i, assignment)
	}
	close(start)
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatal(err)
		}
	}
	reports, err := fixture.reports.ListForManuscript(context.Background(), target.ID, target.ActiveVersion)
	if err != nil || len(reports) != 2 || reports[0].ID == reports[1].ID {
		t.Fatalf("concurrent reports were not preserved independently: %#v %v", reports, err)
	}
}

func testBody() string {
	return "研究正文详细说明背景和目的，并依次描述资料来源、样本范围、实验方法、质量控制、统计结果、讨论、局限与结论。This body includes enough distinct words for deterministic preflight validation and supports a realistic editorial submission workflow without placeholder content."
}
