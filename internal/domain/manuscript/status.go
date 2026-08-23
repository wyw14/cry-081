package manuscript

import "github.com/wyw14/cry-081/internal/domain/shared"

type Status string

const (
	StatusDraft          Status = "draft"
	StatusSubmitted      Status = "submitted"
	StatusUnderInitial   Status = "under_initial_review"
	StatusRevisionNeeded Status = "revision_needed"
	StatusUnderReReview  Status = "under_re_review"
	StatusUnderFinal     Status = "under_final_review"
	StatusRejected       Status = "rejected"
	StatusAccepted       Status = "accepted"
	StatusScheduled      Status = "scheduled"
	StatusPublished      Status = "published"
	StatusWithdrawn      Status = "withdrawn"
)

type TransitionRule struct {
	From         Status
	To           Status
	ReviewStage  bool
	Publication  bool
	Compensation bool
}

var transitionRules = []TransitionRule{
	{From: StatusDraft, To: StatusSubmitted},
	{From: StatusSubmitted, To: StatusUnderInitial, ReviewStage: true},
	{From: StatusSubmitted, To: StatusRejected, ReviewStage: true},
	{From: StatusUnderInitial, To: StatusRevisionNeeded, ReviewStage: true},
	{From: StatusUnderInitial, To: StatusUnderFinal, ReviewStage: true},
	{From: StatusUnderInitial, To: StatusRejected, ReviewStage: true},
	{From: StatusRevisionNeeded, To: StatusUnderReReview, ReviewStage: true},
	{From: StatusRevisionNeeded, To: StatusRejected, ReviewStage: true},
	{From: StatusUnderReReview, To: StatusRevisionNeeded, ReviewStage: true},
	{From: StatusUnderReReview, To: StatusUnderFinal, ReviewStage: true},
	{From: StatusUnderReReview, To: StatusRejected, ReviewStage: true},
	{From: StatusUnderFinal, To: StatusAccepted, ReviewStage: true},
	{From: StatusUnderFinal, To: StatusRevisionNeeded, ReviewStage: true},
	{From: StatusUnderFinal, To: StatusRejected, ReviewStage: true},
	{From: StatusAccepted, To: StatusScheduled, Publication: true},
	{From: StatusScheduled, To: StatusPublished, Publication: true},
	{From: StatusScheduled, To: StatusAccepted, Publication: true, Compensation: true},
	{From: StatusPublished, To: StatusWithdrawn, Publication: true, Compensation: true},
}

func CanTransition(from, to Status) bool {
	_, exists := FindTransition(from, to)
	return exists
}

func FindTransition(from, to Status) (TransitionRule, bool) {
	for _, rule := range transitionRules {
		if rule.From == from && rule.To == to {
			return rule, true
		}
	}
	return TransitionRule{}, false
}

func AllowedTransitions(from Status) []Status {
	allowed := make([]Status, 0, 3)
	for _, rule := range transitionRules {
		if rule.From == from {
			allowed = append(allowed, rule.To)
		}
	}
	return allowed
}

func IsReviewTransition(from, to Status) bool {
	rule, exists := FindTransition(from, to)
	return exists && rule.ReviewStage
}

func IsPublicationTransition(from, to Status) bool {
	rule, exists := FindTransition(from, to)
	return exists && rule.Publication
}

func IsCompensation(from, to Status) bool {
	rule, exists := FindTransition(from, to)
	return exists && rule.Compensation
}

func RequireTransition(from, to Status) error {
	if CanTransition(from, to) {
		return nil
	}
	return shared.NewError("MANUSCRIPT_STATE_INVALID", "manuscript state transition is not allowed", shared.ErrInvalidState)
}

func (s Status) Terminal() bool {
	return s == StatusRejected || s == StatusWithdrawn
}
