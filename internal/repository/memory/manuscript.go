package memory

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type ManuscriptStore struct {
	mu          sync.RWMutex
	manuscripts map[string]manuscript.Manuscript
}

func NewManuscriptStore() *ManuscriptStore {
	return &ManuscriptStore{manuscripts: map[string]manuscript.Manuscript{}}
}

func (s *ManuscriptStore) Create(ctx context.Context, item manuscript.Manuscript) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.manuscripts[item.ID]; exists {
		return shared.ErrConflict
	}
	s.manuscripts[item.ID] = item.Clone()
	return nil
}

func (s *ManuscriptStore) Get(ctx context.Context, id string) (manuscript.Manuscript, error) {
	if err := ctx.Err(); err != nil {
		return manuscript.Manuscript{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.manuscripts[id]
	if !ok {
		return manuscript.Manuscript{}, shared.ErrNotFound
	}
	return item.Clone(), nil
}

func (s *ManuscriptStore) Update(ctx context.Context, item manuscript.Manuscript, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.manuscripts[item.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || item.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.manuscripts[item.ID] = item.Clone()
	return nil
}

func (s *ManuscriptStore) FindByFingerprint(ctx context.Context, fingerprint string) ([]manuscript.Manuscript, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]manuscript.Manuscript, 0)
	for _, item := range s.manuscripts {
		version, err := item.CurrentVersion()
		if err == nil && version.Fingerprint == fingerprint {
			items = append(items, item.Clone())
		}
	}
	return items, nil
}

func (s *ManuscriptStore) List(ctx context.Context, filter application.ManuscriptFilter) (application.ManuscriptPage, error) {
	if err := ctx.Err(); err != nil {
		return application.ManuscriptPage{}, err
	}
	s.mu.RLock()
	items := make([]manuscript.Manuscript, 0, len(s.manuscripts))
	for _, item := range s.manuscripts {
		if filter.AuthorID != "" && item.AuthorID != filter.AuthorID {
			continue
		}
		if filter.Section != "" && item.SectionID != filter.Section {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		items = append(items, item.Clone())
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		switch filter.Sort {
		case "created_at":
			return items[i].CreatedAt.After(items[j].CreatedAt)
		case "status":
			return strings.Compare(string(items[i].Status), string(items[j].Status)) < 0
		default:
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
	})
	start := 0
	if filter.Cursor != "" {
		parsed, err := strconv.Atoi(filter.Cursor)
		if err != nil || parsed < 0 || parsed > len(items) {
			return application.ManuscriptPage{}, shared.NewError("CURSOR_INVALID", "cursor is invalid", shared.ErrValidation)
		}
		start = parsed
	}
	end := start + filter.Limit
	if end > len(items) {
		end = len(items)
	}
	page := application.ManuscriptPage{Items: items[start:end]}
	if end < len(items) {
		page.NextCursor = strconv.Itoa(end)
	}
	return page, nil
}
