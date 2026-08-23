package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wyw14/cry-081/internal/domain/review"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type AssignmentStore struct{ db *Database }

func NewAssignmentStore(db *Database) *AssignmentStore { return &AssignmentStore{db: db} }

func (s *AssignmentStore) Create(ctx context.Context, item review.Assignment) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `INSERT INTO review_assignments (id, manuscript_id, manuscript_version, reviewer_id, due_at, status, aggregate, version) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, item.ID, item.ManuscriptID, item.ManuscriptVersion, item.ReviewerID, item.DueAt, item.Status, payload, item.Version)
	return translate(err)
}

func (s *AssignmentStore) Get(ctx context.Context, id string) (review.Assignment, error) {
	var payload []byte
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT aggregate FROM review_assignments WHERE id=$1`, id).Scan(&payload)
	if err != nil {
		return review.Assignment{}, translate(err)
	}
	var item review.Assignment
	return item, json.Unmarshal(payload, &item)
}

func (s *AssignmentStore) Update(ctx context.Context, item review.Assignment, expectedVersion int64) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE review_assignments SET due_at=$1,status=$2,aggregate=$3,version=$4 WHERE id=$5 AND version=$6`, item.DueAt, item.Status, payload, item.Version, item.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

func (s *AssignmentStore) ListForManuscript(ctx context.Context, manuscriptID string, version int64) ([]review.Assignment, error) {
	return s.list(ctx, `SELECT aggregate FROM review_assignments WHERE manuscript_id=$1 AND manuscript_version=$2 ORDER BY due_at`, manuscriptID, version)
}

func (s *AssignmentStore) ListDue(ctx context.Context, unix int64) ([]review.Assignment, error) {
	return s.list(ctx, `SELECT aggregate FROM review_assignments WHERE due_at < $1 AND status IN ('offered','accepted') ORDER BY due_at`, time.Unix(unix, 0).UTC())
}

func (s *AssignmentStore) list(ctx context.Context, query string, args ...any) ([]review.Assignment, error) {
	rows, err := s.db.queries(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]review.Assignment, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item review.Assignment
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *AssignmentStore) HasActiveConflict(ctx context.Context, editorID, authorID string, now time.Time) (bool, error) {
	var exists bool
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM editorial_conflicts WHERE editor_id=$1 AND author_id=$2 AND starts_at <= $3 AND (ends_at IS NULL OR ends_at > $3))`, editorID, authorID, now.UTC()).Scan(&exists)
	return exists, translate(err)
}

type ReportStore struct{ db *Database }

func NewReportStore(db *Database) *ReportStore { return &ReportStore{db: db} }

func (s *ReportStore) Create(ctx context.Context, item review.Report) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `INSERT INTO review_reports (id, assignment_id, manuscript_id, manuscript_version, reviewer_id, aggregate, version) VALUES ($1,$2,$3,$4,$5,$6,$7)`, item.ID, item.AssignmentID, item.ManuscriptID, item.ManuscriptVersion, item.ReviewerID, payload, item.Version)
	return translate(err)
}

func (s *ReportStore) Get(ctx context.Context, id string) (review.Report, error) {
	var payload []byte
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT aggregate FROM review_reports WHERE id=$1`, id).Scan(&payload)
	if err != nil {
		return review.Report{}, translate(err)
	}
	var item review.Report
	return item, json.Unmarshal(payload, &item)
}

func (s *ReportStore) Update(ctx context.Context, item review.Report, expectedVersion int64) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE review_reports SET aggregate=$1,version=$2 WHERE id=$3 AND version=$4`, payload, item.Version, item.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

func (s *ReportStore) ListForManuscript(ctx context.Context, manuscriptID string, version int64) ([]review.Report, error) {
	rows, err := s.db.queries(ctx).Query(ctx, `SELECT aggregate FROM review_reports WHERE manuscript_id=$1 AND manuscript_version=$2 ORDER BY id`, manuscriptID, version)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]review.Report, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item review.Report
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type DecisionStore struct{ db *Database }

func NewDecisionStore(db *Database) *DecisionStore { return &DecisionStore{db: db} }

func (s *DecisionStore) Create(ctx context.Context, item review.EditorialDecision) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `INSERT INTO editorial_decisions (id,manuscript_id,manuscript_version,aggregate,created_at) VALUES ($1,$2,$3,$4,$5)`, item.ID, item.ManuscriptID, item.ManuscriptVersion, payload, item.CreatedAt)
	return translate(err)
}

func (s *DecisionStore) ListForManuscript(ctx context.Context, manuscriptID string) ([]review.EditorialDecision, error) {
	rows, err := s.db.queries(ctx).Query(ctx, `SELECT aggregate FROM editorial_decisions WHERE manuscript_id=$1 ORDER BY created_at`, manuscriptID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]review.EditorialDecision, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item review.EditorialDecision
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
