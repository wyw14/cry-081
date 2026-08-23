package review_test

import (
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/domain/review"
)

func TestReportRequiresAcceptedAssignmentAndValidRubric(t *testing.T) {
	now := time.Now().UTC()
	assignment, err := review.NewAssignment("a1", "m1", 2, "reviewer1", "editor1", now.Add(48*time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := review.NewReport("r1", *assignment, review.RecommendAccept, []review.ScoreItem{{Criterion: "originality", Score: 5}}, []review.StructuredComment{{ID: "c1", Section: "methods", Comment: "clear", Required: false}}, "editor only", now); err == nil {
		t.Fatal("offered assignment created a report")
	}
	if err := assignment.Accept("reviewer1", now); err != nil {
		t.Fatal(err)
	}
	report, err := review.NewReport("r1", *assignment, review.RecommendMinorChanges, []review.ScoreItem{{Criterion: "originality", Score: 4}}, []review.StructuredComment{{ID: "c1", Section: "methods", Comment: "clarify sampling", Required: true}}, "identity-sensitive note", now)
	if err != nil {
		t.Fatal(err)
	}
	public := report.PublicView()
	if public.ReviewerID != "" || public.AnonymousNote != "" {
		t.Fatalf("public review leaked protected fields: %#v", public)
	}
}

func TestAuthorResponseMustReferenceStructuredComment(t *testing.T) {
	now := time.Now().UTC()
	assignment, _ := review.NewAssignment("a1", "m1", 1, "reviewer1", "editor1", now.Add(time.Hour), now)
	_ = assignment.Accept("reviewer1", now)
	report, _ := review.NewReport("r1", *assignment, review.RecommendMajorChanges, []review.ScoreItem{{Criterion: "methods", Score: 2}}, []review.StructuredComment{{ID: "c1", Section: "methods", Comment: "missing controls", Required: true}}, "", now)
	if err := report.AddAuthorResponse(review.AuthorResponse{CommentID: "missing", Response: "fixed", Changed: true, CreatedAt: now}); err == nil {
		t.Fatal("response accepted an unknown comment")
	}
	if err := report.AddAuthorResponse(review.AuthorResponse{CommentID: "c1", Response: "added a control group", Changed: true, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
}
