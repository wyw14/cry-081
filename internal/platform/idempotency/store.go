package idempotency

import (
	"context"
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
	entries map[Identity]Replay
}

func NewMemory() *Memory {
	return NewMemoryWithClock(time.Now)
}

func NewMemoryWithClock(now func() time.Time) *Memory {
	if now == nil {
		now = time.Now
	}
	return &Memory{now: now, entries: make(map[Identity]Replay)}
}

func (m *Memory) Find(ctx context.Context, identity Identity) (Replay, bool, error) {
	if err := ctx.Err(); err != nil {
		return Replay{}, false, err
	}
	m.lock.RLock()
	entry, found := m.entries[identity]
	m.lock.RUnlock()
	if !found || !entry.ExpiresAt.After(m.now().UTC()) {
		return Replay{}, false, nil
	}
	return entry.clone(), true, nil
}

func (m *Memory) Remember(ctx context.Context, replay Replay) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.lock.Lock()
	m.entries[replay.Identity] = replay.clone()
	m.lock.Unlock()
	return nil
}
