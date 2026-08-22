package publication_test

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/domain/publication"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

func TestIssueReleaseRequiresArticleForEverySlot(t *testing.T) {
	now := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
	issue, err := publication.NewIssue("issue-1", 1, 1, "Research", now.Add(24*time.Hour), "Asia/Shanghai", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := issue.AddSlot(publication.Slot{ManuscriptID: "manuscript-1", SectionID: "research", Edition: 1, Position: 1}, now); err != nil {
		t.Fatal(err)
	}
	if _, err := issue.ReleaseArticles(nil, now); !errors.Is(err, shared.ErrInvalidState) {
		t.Fatalf("incomplete issue release error=%v", err)
	}
	if issue.Status != publication.IssuePlanning {
		t.Fatalf("failed release changed status to %s", issue.Status)
	}
}
