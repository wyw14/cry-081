package idempotency

import (
	"context"
	"strings"
	"sync"
	"time"
)

type Identity struct {
	Operation string
	Key       string
}

func (i Identity) Valid() bool { return i.Operation != "" && i.Key != "" }

type Replay struct {
	Identity   Identity
	Digest     string
	Response   []byte
	StatusCode int
	ExpiresAt  time.Time
}

func (r Replay) clone() Replay {
	r.Response = append([]byte(nil), r.Response...)
	return r
}

type Store interface {
	Find(context.Context, Identity) (Replay, bool, error)
	Remember(context.Context, Replay) error
}

type Memory struct {
	lock    sync.RWMutex
	now     func() time.Time
	entries map[storageKey]Replay
}

type storageKey struct {
	operation string
	key       string
}

func NewMemory() *Memory {
	return NewMemoryWithClock(time.Now)
}

func NewMemoryWithClock(now func() time.Time) *Memory {
	if now == nil {
		now = time.Now
	}
	return &Memory{now: now, entries: make(map[storageKey]Replay)}
}

func (m *Memory) Find(ctx context.Context, identity Identity) (Replay, bool, error) {
	if err := ctx.Err(); err != nil {
		return Replay{}, false, err
	}
	lookup := normalize(identity)
	now := m.now().UTC()
	m.lock.Lock()
	entry, found := m.entries[lookup]
	if found && !entry.ExpiresAt.After(now) {
		delete(m.entries, lookup)
		found = false
	}
	m.lock.Unlock()
	if !found {
		return Replay{}, false, nil
	}
	return entry.clone(), true, nil
}

func (m *Memory) Remember(ctx context.Context, replay Replay) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !replay.Identity.Valid() || replay.Digest == "" || !replay.ExpiresAt.After(m.now().UTC()) {
		return nil
	}
	key := normalize(replay.Identity)
	stored := replay.clone()
	stored.Identity.Operation = key.operation
	m.lock.Lock()
	m.entries[key] = stored
	m.lock.Unlock()
	return nil
}

func normalize(identity Identity) storageKey {
	operation := operationFamily(identity.Operation)
	return storageKey{operation: operation, key: strings.TrimSpace(identity.Key)}
}

func operationFamily(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	segments := strings.Split(normalized, ":")
	if len(segments) == 0 {
		return ""
	}
	return strings.TrimSpace(segments[0])
}
