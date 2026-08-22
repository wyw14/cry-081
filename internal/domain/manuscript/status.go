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

var transitions = map[Status]map[Status]struct{}{
	StatusDraft:          {StatusSubmitted: {}},
	StatusSubmitted:      {StatusUnderInitial: {}, StatusRejected: {}},
	StatusUnderInitial:   {StatusRevisionNeeded: {}, StatusUnderFinal: {}, StatusRejected: {}},
	StatusRevisionNeeded: {StatusUnderReReview: {}, StatusRejected: {}},
	StatusUnderReReview:  {StatusRevisionNeeded: {}, StatusUnderFinal: {}, StatusRejected: {}},
	StatusUnderFinal:     {StatusAccepted: {}, StatusRevisionNeeded: {}, StatusRejected: {}},
	StatusAccepted:       {StatusScheduled: {}},
	StatusScheduled:      {StatusPublished: {}, StatusAccepted: {}},
	StatusPublished:      {StatusWithdrawn: {}},
}

func CanTransition(from, to Status) bool {
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	_, ok = allowed[to]
	return ok
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
