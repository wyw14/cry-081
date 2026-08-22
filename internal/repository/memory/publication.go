package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/publication"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type PublicationStore struct {
	mu       sync.RWMutex
	issues   map[string]publication.Issue
	articles map[string]publication.Article
}

func NewPublicationStore() *PublicationStore {
	return &PublicationStore{issues: map[string]publication.Issue{}, articles: map[string]publication.Article{}}
}

type IssueStore struct{ publication *PublicationStore }

func NewIssueStore(store *PublicationStore) *IssueStore { return &IssueStore{publication: store} }

func (s *IssueStore) Create(ctx context.Context, issue publication.Issue) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.publication.mu.Lock()
	defer s.publication.mu.Unlock()
	if _, exists := s.publication.issues[issue.ID]; exists {
		return shared.ErrConflict
	}
	for _, current := range s.publication.issues {
		if current.Volume == issue.Volume && current.Number == issue.Number {
			return shared.ErrConflict
		}
	}
	s.publication.issues[issue.ID] = issue.Clone()
	return nil
}

func (s *IssueStore) Get(ctx context.Context, id string) (publication.Issue, error) {
	if err := ctx.Err(); err != nil {
		return publication.Issue{}, err
	}
	s.publication.mu.RLock()
	defer s.publication.mu.RUnlock()
	issue, ok := s.publication.issues[id]
	if !ok {
		return publication.Issue{}, shared.ErrNotFound
	}
	return issue.Clone(), nil
}

func (s *IssueStore) Update(ctx context.Context, issue publication.Issue, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.publication.mu.Lock()
	defer s.publication.mu.Unlock()
	current, ok := s.publication.issues[issue.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || issue.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.publication.issues[issue.ID] = issue.Clone()
	return nil
}

type ArticleStore struct{ publication *PublicationStore }

func NewArticleStore(store *PublicationStore) *ArticleStore { return &ArticleStore{publication: store} }

func (s *ArticleStore) Create(ctx context.Context, article publication.Article) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.publication.mu.Lock()
	defer s.publication.mu.Unlock()
	if _, exists := s.publication.articles[article.ID]; exists {
		return shared.ErrConflict
	}
	s.publication.articles[article.ID] = article.Clone()
	return nil
}

func (s *ArticleStore) Get(ctx context.Context, id string) (publication.Article, error) {
	if err := ctx.Err(); err != nil {
		return publication.Article{}, err
	}
	s.publication.mu.RLock()
	defer s.publication.mu.RUnlock()
	article, ok := s.publication.articles[id]
	if !ok {
		return publication.Article{}, shared.ErrNotFound
	}
	return article.Clone(), nil
}

func (s *ArticleStore) Update(ctx context.Context, article publication.Article, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.publication.mu.Lock()
	defer s.publication.mu.Unlock()
	current, ok := s.publication.articles[article.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || article.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.publication.articles[article.ID] = article.Clone()
	return nil
}

func (s *ArticleStore) ListForIssue(ctx context.Context, issueID string) ([]publication.Article, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.publication.mu.RLock()
	defer s.publication.mu.RUnlock()
	items := make([]publication.Article, 0)
	for _, article := range s.publication.articles {
		if article.IssueID == issueID {
			items = append(items, article.Clone())
		}
	}
	return items, nil
}

func (s *ArticleStore) Search(ctx context.Context, filter application.PublicationFilter) (application.PublicationPage, error) {
	if err := ctx.Err(); err != nil {
		return application.PublicationPage{}, err
	}
	s.publication.mu.RLock()
	items := make([]publication.Article, 0)
	for _, article := range s.publication.articles {
		if !article.Searchable() {
			continue
		}
		if filter.SectionID != "" && article.SectionID != filter.SectionID {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(article.Title+" "+article.Abstract+" "+article.Body), strings.ToLower(filter.Query)) {
			continue
		}
		if filter.Tag != "" && !contains(article.Tags, strings.ToLower(filter.Tag)) {
			continue
		}
		items = append(items, article.Clone())
	}
	s.publication.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		if filter.Sort == "title" {
			return items[i].Title < items[j].Title
		}
		if items[i].PublishedAt == nil || items[j].PublishedAt == nil {
			return items[i].ID < items[j].ID
		}
		return items[i].PublishedAt.After(*items[j].PublishedAt)
	})
	total := len(items)
	start := (filter.Page - 1) * filter.PageSize
	if start > total {
		start = total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return application.PublicationPage{Items: items[start:end], Total: total}, nil
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
