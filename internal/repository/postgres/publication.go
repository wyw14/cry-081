package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/publication"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type IssueStore struct{ db *Database }

func NewIssueStore(db *Database) *IssueStore { return &IssueStore{db: db} }

func (s *IssueStore) Create(ctx context.Context, item publication.Issue) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `INSERT INTO issues (id,volume,number,status,release_at,aggregate,version) VALUES ($1,$2,$3,$4,$5,$6,$7)`, item.ID, item.Volume, item.Number, item.Status, item.ReleaseAt, payload, item.Version)
	return translate(err)
}

func (s *IssueStore) Get(ctx context.Context, id string) (publication.Issue, error) {
	var payload []byte
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT aggregate FROM issues WHERE id=$1`, id).Scan(&payload)
	if err != nil {
		return publication.Issue{}, translate(err)
	}
	var item publication.Issue
	return item, json.Unmarshal(payload, &item)
}

func (s *IssueStore) Update(ctx context.Context, item publication.Issue, expectedVersion int64) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE issues SET status=$1,aggregate=$2,version=$3 WHERE id=$4 AND version=$5`, item.Status, payload, item.Version, item.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

type ArticleStore struct{ db *Database }

func NewArticleStore(db *Database) *ArticleStore { return &ArticleStore{db: db} }

func (s *ArticleStore) Create(ctx context.Context, item publication.Article) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `INSERT INTO articles (id,manuscript_id,issue_id,section_id,title,status,published_at,aggregate,version) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, item.ID, item.ManuscriptID, item.IssueID, item.SectionID, item.Title, item.Status, item.PublishedAt, payload, item.Version)
	return translate(err)
}

func (s *ArticleStore) Get(ctx context.Context, id string) (publication.Article, error) {
	var payload []byte
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT aggregate FROM articles WHERE id=$1`, id).Scan(&payload)
	if err != nil {
		return publication.Article{}, translate(err)
	}
	var item publication.Article
	return item, json.Unmarshal(payload, &item)
}

func (s *ArticleStore) Update(ctx context.Context, item publication.Article, expectedVersion int64) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE articles SET status=$1,published_at=$2,aggregate=$3,version=$4 WHERE id=$5 AND version=$6`, item.Status, item.PublishedAt, payload, item.Version, item.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

func (s *ArticleStore) ListForIssue(ctx context.Context, issueID string) ([]publication.Article, error) {
	rows, err := s.db.queries(ctx).Query(ctx, `SELECT aggregate FROM articles WHERE issue_id=$1 ORDER BY id`, issueID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]publication.Article, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item publication.Article
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *ArticleStore) Search(ctx context.Context, filter application.PublicationFilter) (application.PublicationPage, error) {
	order := "published_at DESC"
	if filter.Sort == "title" {
		order = "title ASC"
	}
	query := fmt.Sprintf(`SELECT aggregate, count(*) OVER() FROM articles WHERE status='published' AND ($1='' OR section_id=$1) AND ($2='' OR aggregate::text ILIKE '%%' || $2 || '%%') ORDER BY %s LIMIT $3 OFFSET $4`, order)
	rows, err := s.db.queries(ctx).Query(ctx, query, filter.SectionID, strings.TrimSpace(filter.Query), filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return application.PublicationPage{}, translate(err)
	}
	defer rows.Close()
	page := application.PublicationPage{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload, &page.Total); err != nil {
			return application.PublicationPage{}, err
		}
		var item publication.Article
		if err := json.Unmarshal(payload, &item); err != nil {
			return application.PublicationPage{}, err
		}
		if filter.Tag != "" && !containsTag(item.Tags, strings.ToLower(filter.Tag)) {
			continue
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}
