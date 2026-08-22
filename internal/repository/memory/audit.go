package memory

import (
	"context"
	"sync"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/audit"
)

type AuditStore struct {
	mu     sync.RWMutex
	events []audit.Event
}

func NewAuditStore() *AuditStore { return &AuditStore{} }

func (s *AuditStore) Append(ctx context.Context, event audit.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	s.events = append(s.events, event)
	s.mu.Unlock()
	return nil
}

func (s *AuditStore) List(ctx context.Context, filter application.AuditFilter) ([]audit.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	limit := filter.Limit
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	items := make([]audit.Event, 0, limit)
	for i := len(s.events) - 1; i >= 0 && len(items) < limit; i-- {
		event := s.events[i]
		if filter.Resource != "" && event.Resource != filter.Resource {
			continue
		}
		if filter.ResourceID != "" && event.ResourceID != filter.ResourceID {
			continue
		}
		if filter.ActorID != "" && event.ActorID != filter.ActorID {
			continue
		}
		items = append(items, event)
	}
	return items, nil
}
