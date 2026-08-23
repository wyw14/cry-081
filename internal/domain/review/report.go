package review

import (
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type Recommendation string

const (
	RecommendAccept       Recommendation = "accept"
	RecommendMinorChanges Recommendation = "minor_changes"
	RecommendMajorChanges Recommendation = "major_changes"
	RecommendReject       Recommendation = "reject"
)

type ScoreItem struct {
	Criterion string
	Score     int
	Comment   string
}

type StructuredComment struct {
	ID       string
	Section  string
	Comment  string
	Required bool
}

type AuthorResponse struct {
	CommentID string
	Response  string
	Changed   bool
	CreatedAt time.Time
}

type Report struct {
	ID                string
	AssignmentID      string
	ManuscriptID      string
	ManuscriptVersion int64
	ReviewerID        string
	Recommendation    Recommendation
	Scores            []ScoreItem
	Comments          []StructuredComment
	AnonymousNote     string
	Responses         []AuthorResponse
	SubmittedAt       time.Time
	Version           int64
}

func NewReport(id string, assignment Assignment, recommendation Recommendation, scores []ScoreItem, comments []StructuredComment, anonymousNote string, now time.Time) (*Report, error) {
	if id == "" || assignment.Status != AssignmentAccepted {
		return nil, shared.NewError("REPORT_ASSIGNMENT_INVALID", "an accepted assignment is required", shared.ErrInvalidState)
	}
	if !recommendation.Valid() {
		return nil, shared.NewError("RECOMMENDATION_INVALID", "review recommendation is not supported", shared.ErrValidation)
	}
	if len(scores) == 0 || len(comments) == 0 {
		return nil, shared.NewError("REPORT_INCOMPLETE", "scores and structured comments are required", shared.ErrValidation)
	}
	for _, score := range scores {
		if strings.TrimSpace(score.Criterion) == "" || score.Score < 1 || score.Score > 5 {
			return nil, shared.NewError("SCORE_INVALID", "score must use the one-to-five rubric", shared.ErrValidation)
		}
	}
	return &Report{
		ID: id, AssignmentID: assignment.ID, ManuscriptID: assignment.ManuscriptID,
		ManuscriptVersion: assignment.ManuscriptVersion, ReviewerID: assignment.ReviewerID,
		Recommendation: recommendation, Scores: append([]ScoreItem(nil), scores...),
		Comments: append([]StructuredComment(nil), comments...), AnonymousNote: strings.TrimSpace(anonymousNote),
		SubmittedAt: now.UTC(), Version: 1,
	}, nil
}

func (r Recommendation) Valid() bool {
	switch r {
	case RecommendAccept, RecommendMinorChanges, RecommendMajorChanges, RecommendReject:
		return true
	default:
		return false
	}
}

func (r *Report) AddAuthorResponse(response AuthorResponse) error {
	if response.CommentID == "" || strings.TrimSpace(response.Response) == "" {
		return shared.NewError("AUTHOR_RESPONSE_INVALID", "comment and response are required", shared.ErrValidation)
	}
	found := false
	for _, comment := range r.Comments {
		if comment.ID == response.CommentID {
			found = true
			break
		}
	}
	if !found {
		return shared.NewError("REVIEW_COMMENT_NOT_FOUND", "review comment does not exist", shared.ErrNotFound)
	}
	r.Responses = append(r.Responses, response)
	r.Version++
	return nil
}

func (r Report) PublicView() Report {
	copy := r.Clone()
	copy.ReviewerID = ""
	copy.AnonymousNote = ""
	return copy
}

func (r Report) Clone() Report {
	r.Scores = append([]ScoreItem(nil), r.Scores...)
	r.Comments = append([]StructuredComment(nil), r.Comments...)
	r.Responses = append([]AuthorResponse(nil), r.Responses...)
	return r
}
