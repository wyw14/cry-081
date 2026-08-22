package review

import (
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type AssignmentStatus string

const (
	AssignmentOffered   AssignmentStatus = "offered"
	AssignmentAccepted  AssignmentStatus = "accepted"
	AssignmentDeclined  AssignmentStatus = "declined"
	AssignmentCompleted AssignmentStatus = "completed"
	AssignmentCancelled AssignmentStatus = "cancelled"
)

type Assignment struct {
	ID                string
	ManuscriptID      string
	ManuscriptVersion int64
	ReviewerID        string
	AssignedBy        string
	DueAt             time.Time
	Status            AssignmentStatus
	ReminderCount     int
	LastRemindedAt    *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           int64
}

func NewAssignment(id, manuscriptID string, manuscriptVersion int64, reviewerID, assignedBy string, dueAt, now time.Time) (*Assignment, error) {
	if id == "" || manuscriptID == "" || reviewerID == "" || assignedBy == "" || manuscriptVersion < 1 {
		return nil, shared.NewError("ASSIGNMENT_INVALID", "assignment fields are incomplete", shared.ErrValidation)
	}
	if !dueAt.After(now) {
		return nil, shared.NewError("ASSIGNMENT_DEADLINE_INVALID", "review deadline must be in the future", shared.ErrValidation)
	}
	now = now.UTC()
	return &Assignment{
		ID: id, ManuscriptID: manuscriptID, ManuscriptVersion: manuscriptVersion,
		ReviewerID: reviewerID, AssignedBy: assignedBy, DueAt: dueAt.UTC(),
		Status: AssignmentOffered, CreatedAt: now, UpdatedAt: now, Version: 1,
	}, nil
}

func (a *Assignment) Accept(actorID string, now time.Time) error {
	if actorID != a.ReviewerID {
		return shared.NewError("ASSIGNMENT_REVIEWER_REQUIRED", "only the assigned reviewer can accept", shared.ErrForbidden)
	}
	if a.Status != AssignmentOffered {
		return shared.NewError("ASSIGNMENT_STATE_INVALID", "assignment is not available for acceptance", shared.ErrInvalidState)
	}
	if !now.Before(a.DueAt) {
		return shared.NewError("ASSIGNMENT_EXPIRED", "assignment deadline has passed", shared.ErrConflict)
	}
	a.Status = AssignmentAccepted
	a.UpdatedAt = now.UTC()
	a.Version++
	return nil
}

func (a *Assignment) Decline(actorID string, now time.Time) error {
	if actorID != a.ReviewerID {
		return shared.NewError("ASSIGNMENT_REVIEWER_REQUIRED", "only the assigned reviewer can decline", shared.ErrForbidden)
	}
	if a.Status != AssignmentOffered {
		return shared.NewError("ASSIGNMENT_STATE_INVALID", "assignment is not available for decline", shared.ErrInvalidState)
	}
	a.Status = AssignmentDeclined
	a.UpdatedAt = now.UTC()
	a.Version++
	return nil
}

func (a *Assignment) Complete(now time.Time) error {
	if a.Status != AssignmentAccepted {
		return shared.NewError("ASSIGNMENT_NOT_ACCEPTED", "accepted assignment is required", shared.ErrInvalidState)
	}
	a.Status = AssignmentCompleted
	a.UpdatedAt = now.UTC()
	a.Version++
	return nil
}

func (a *Assignment) RecordReminder(now time.Time) error {
	if a.Status != AssignmentAccepted && a.Status != AssignmentOffered {
		return shared.NewError("REMINDER_NOT_APPLICABLE", "assignment no longer accepts reminders", shared.ErrInvalidState)
	}
	when := now.UTC()
	a.LastRemindedAt = &when
	a.ReminderCount++
	a.UpdatedAt = when
	a.Version++
	return nil
}

type Conflict struct {
	EditorID string
	AuthorID string
	Reason   string
	StartsAt time.Time
	EndsAt   *time.Time
}

type ConflictIndex struct {
	byEditor map[string][]Conflict
}

func NewConflictIndex(items []Conflict) ConflictIndex {
	index := ConflictIndex{byEditor: make(map[string][]Conflict)}
	for _, item := range items {
		index.Add(item)
	}
	return index
}

func (i *ConflictIndex) Add(conflict Conflict) {
	if i.byEditor == nil {
		i.byEditor = make(map[string][]Conflict)
	}
	i.byEditor[conflict.EditorID] = append(i.byEditor[conflict.EditorID], conflict)
}

func (i ConflictIndex) Active(editorID, authorID string, now time.Time) bool {
	for _, conflict := range i.byEditor[editorID] {
		if conflict.ActiveAt(now) {
			return true
		}
	}
	return false
}

func (i ConflictIndex) ForEditor(editorID string) []Conflict {
	items := i.byEditor[editorID]
	result := make([]Conflict, len(items))
	copy(result, items)
	return result
}

func (c Conflict) ActiveAt(now time.Time) bool {
	instant := now.UTC()
	startsAt := c.StartsAt.UTC()
	if instant.Before(startsAt) {
		return false
	}
	if c.EndsAt == nil {
		return true
	}
	return instant.Before(c.EndsAt.UTC())
}
