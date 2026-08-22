package memory

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-081/internal/domain/review"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type ReviewStore struct {
	mu          sync.RWMutex
	assignments map[string]review.Assignment
	reports     map[string]review.Report
	decisions   map[string]review.EditorialDecision
	conflicts   review.ConflictIndex
}

func NewReviewStore() *ReviewStore {
	return &ReviewStore{
		assignments: map[string]review.Assignment{},
		reports:     map[string]review.Report{},
		decisions:   map[string]review.EditorialDecision{},
		conflicts:   review.NewConflictIndex(nil),
	}
}

func (s *ReviewStore) AddConflict(conflict review.Conflict) {
	s.mu.Lock()
	s.conflicts.Add(conflict)
	s.mu.Unlock()
}

func (s *ReviewStore) Create(ctx context.Context, assignment review.Assignment) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.assignments[assignment.ID]; exists {
		return shared.ErrConflict
	}
	for _, current := range s.assignments {
		if current.ManuscriptID == assignment.ManuscriptID && current.ManuscriptVersion == assignment.ManuscriptVersion && current.ReviewerID == assignment.ReviewerID && current.Status != review.AssignmentCancelled && current.Status != review.AssignmentDeclined {
			return shared.ErrConflict
		}
	}
	s.assignments[assignment.ID] = assignment
	return nil
}

func (s *ReviewStore) Get(ctx context.Context, id string) (review.Assignment, error) {
	if err := ctx.Err(); err != nil {
		return review.Assignment{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	assignment, ok := s.assignments[id]
	if !ok {
		return review.Assignment{}, shared.ErrNotFound
	}
	return assignment, nil
}

func (s *ReviewStore) Update(ctx context.Context, assignment review.Assignment, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.assignments[assignment.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || assignment.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.assignments[assignment.ID] = assignment
	return nil
}

func (s *ReviewStore) ListForManuscript(ctx context.Context, manuscriptID string, version int64) ([]review.Assignment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]review.Assignment, 0)
	for _, assignment := range s.assignments {
		if assignment.ManuscriptID == manuscriptID && assignment.ManuscriptVersion == version {
			items = append(items, assignment)
		}
	}
	return items, nil
}

func (s *ReviewStore) ListDue(ctx context.Context, unix int64) ([]review.Assignment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := time.Unix(unix, 0).UTC()
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]review.Assignment, 0)
	for _, assignment := range s.assignments {
		if assignment.DueAt.Before(now) && (assignment.Status == review.AssignmentAccepted || assignment.Status == review.AssignmentOffered) {
			items = append(items, assignment)
		}
	}
	return items, nil
}

func (s *ReviewStore) HasActiveConflict(ctx context.Context, editorID, authorID string, now time.Time) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conflicts.Active(editorID, authorID, now.UTC()), nil
}

type ReportStore struct{ review *ReviewStore }

func NewReportStore(store *ReviewStore) *ReportStore { return &ReportStore{review: store} }

func (s *ReportStore) Create(ctx context.Context, report review.Report) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.review.mu.Lock()
	defer s.review.mu.Unlock()
	if _, exists := s.review.reports[report.ID]; exists {
		return shared.ErrConflict
	}
	for _, current := range s.review.reports {
		if current.AssignmentID == report.AssignmentID {
			return shared.ErrConflict
		}
	}
	s.review.reports[report.ID] = report.Clone()
	return nil
}

func (s *ReportStore) Get(ctx context.Context, id string) (review.Report, error) {
	if err := ctx.Err(); err != nil {
		return review.Report{}, err
	}
	s.review.mu.RLock()
	defer s.review.mu.RUnlock()
	report, ok := s.review.reports[id]
	if !ok {
		return review.Report{}, shared.ErrNotFound
	}
	return report.Clone(), nil
}

func (s *ReportStore) Update(ctx context.Context, report review.Report, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.review.mu.Lock()
	defer s.review.mu.Unlock()
	current, ok := s.review.reports[report.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || report.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.review.reports[report.ID] = report.Clone()
	return nil
}

func (s *ReportStore) ListForManuscript(ctx context.Context, manuscriptID string, version int64) ([]review.Report, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.review.mu.RLock()
	defer s.review.mu.RUnlock()
	items := make([]review.Report, 0)
	for _, report := range s.review.reports {
		if report.ManuscriptID == manuscriptID && report.ManuscriptVersion == version {
			items = append(items, report.Clone())
		}
	}
	return items, nil
}

type DecisionStore struct{ review *ReviewStore }

func NewDecisionStore(store *ReviewStore) *DecisionStore { return &DecisionStore{review: store} }

func (s *DecisionStore) Create(ctx context.Context, decision review.EditorialDecision) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.review.mu.Lock()
	defer s.review.mu.Unlock()
	if _, exists := s.review.decisions[decision.ID]; exists {
		return shared.ErrConflict
	}
	s.review.decisions[decision.ID] = decision
	return nil
}

func (s *DecisionStore) ListForManuscript(ctx context.Context, manuscriptID string) ([]review.EditorialDecision, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.review.mu.RLock()
	defer s.review.mu.RUnlock()
	items := make([]review.EditorialDecision, 0)
	for _, decision := range s.review.decisions {
		if decision.ManuscriptID == manuscriptID {
			decision.ReviewReportIDs = append([]string(nil), decision.ReviewReportIDs...)
			items = append(items, decision)
		}
	}
	return items, nil
}
