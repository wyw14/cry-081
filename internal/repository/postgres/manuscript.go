package postgres

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type ManuscriptStore struct{ db *Database }

func NewManuscriptStore(db *Database) *ManuscriptStore { return &ManuscriptStore{db: db} }

func (s *ManuscriptStore) Create(ctx context.Context, item manuscript.Manuscript) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	current, err := item.CurrentVersion()
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `INSERT INTO manuscripts (id, author_id, section_id, status, active_version, fingerprint, aggregate, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, item.ID, item.AuthorID, item.SectionID, item.Status, item.ActiveVersion, current.Fingerprint, payload, item.Version, item.CreatedAt, item.UpdatedAt)
	return translate(err)
}

func (s *ManuscriptStore) Get(ctx context.Context, id string) (manuscript.Manuscript, error) {
	var payload []byte
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT aggregate FROM manuscripts WHERE id=$1`, id).Scan(&payload)
	if err != nil {
		return manuscript.Manuscript{}, translate(err)
	}
	var item manuscript.Manuscript
	if err := json.Unmarshal(payload, &item); err != nil {
		return manuscript.Manuscript{}, err
	}
	return item, nil
}

func (s *ManuscriptStore) Update(ctx context.Context, item manuscript.Manuscript, expectedVersion int64) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	current, err := item.CurrentVersion()
	if err != nil {
		return err
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE manuscripts SET status=$1, active_version=$2, fingerprint=$3, aggregate=$4, version=$5, updated_at=$6 WHERE id=$7 AND version=$8`, item.Status, item.ActiveVersion, current.Fingerprint, payload, item.Version, item.UpdatedAt, item.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

func (s *ManuscriptStore) FindByFingerprint(ctx context.Context, fingerprint string) ([]manuscript.Manuscript, error) {
	rows, err := s.db.queries(ctx).Query(ctx, `SELECT aggregate FROM manuscripts WHERE fingerprint=$1`, fingerprint)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]manuscript.Manuscript, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item manuscript.Manuscript
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *ManuscriptStore) List(ctx context.Context, filter application.ManuscriptFilter) (application.ManuscriptPage, error) {
	query := `SELECT aggregate FROM manuscripts WHERE ($1='' OR author_id=$1) AND ($2='' OR status=$2) AND ($3='' OR section_id=$3) ORDER BY updated_at DESC, id LIMIT $4 OFFSET $5`
	offset := 0
	if filter.Cursor != "" {
		parsed, err := strconv.Atoi(filter.Cursor)
		if err != nil || parsed < 0 {
			return application.ManuscriptPage{}, shared.NewError("CURSOR_INVALID", "cursor is invalid", shared.ErrValidation)
		}
		offset = parsed
	}
	rows, err := s.db.queries(ctx).Query(ctx, query, filter.AuthorID, filter.Status, filter.Section, filter.Limit+1, offset)
	if err != nil {
		return application.ManuscriptPage{}, translate(err)
	}
	defer rows.Close()
	items := make([]manuscript.Manuscript, 0, filter.Limit+1)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return application.ManuscriptPage{}, err
		}
		var item manuscript.Manuscript
		if err := json.Unmarshal(payload, &item); err != nil {
			return application.ManuscriptPage{}, err
		}
		items = append(items, item)
	}
	page := application.ManuscriptPage{Items: items}
	if len(items) > filter.Limit {
		page.Items = items[:filter.Limit]
		page.NextCursor = strconv.Itoa(offset + filter.Limit)
	}
	return page, rows.Err()
}
