package outbox

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Message struct {
	ID              string
	Topic           string
	Key             string
	Payload         []byte
	Attempts        int
	MaximumAttempts int
	AvailableAt     time.Time
	ProcessedAt     *time.Time
	DeadAt          *time.Time
	LastError       string
}

type Store interface {
	Enqueue(context.Context, Message) error
	Claim(context.Context, time.Time, int) ([]Message, error)
	MarkProcessed(context.Context, string, time.Time) error
	Reschedule(context.Context, string, string, time.Time, bool) error
}

type Handler interface {
	Handle(context.Context, Message) error
}

type MemoryStore struct {
	mu       sync.Mutex
	messages map[string]Message
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{messages: map[string]Message{}} }

func (s *MemoryStore) Enqueue(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if message.ID == "" || message.Topic == "" || message.MaximumAttempts < 1 {
		return errors.New("invalid outbox message")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.messages[message.ID]; exists {
		return errors.New("outbox message already exists")
	}
	message.Payload = append([]byte(nil), message.Payload...)
	s.messages[message.ID] = message
	return nil
}

func (s *MemoryStore) Claim(ctx context.Context, now time.Time, limit int) ([]Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	claimed := make([]Message, 0, limit)
	for _, message := range s.messages {
		if len(claimed) >= limit {
			break
		}
		if message.ProcessedAt != nil || message.DeadAt != nil || message.AvailableAt.After(now) {
			continue
		}
		message.Payload = append([]byte(nil), message.Payload...)
		claimed = append(claimed, message)
	}
	return claimed, nil
}

func (s *MemoryStore) MarkProcessed(ctx context.Context, id string, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	message, ok := s.messages[id]
	if !ok {
		return errors.New("outbox message not found")
	}
	when := now.UTC()
	message.ProcessedAt = &when
	s.messages[id] = message
	return nil
}

func (s *MemoryStore) Reschedule(ctx context.Context, id, reason string, availableAt time.Time, dead bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	message, ok := s.messages[id]
	if !ok {
		return errors.New("outbox message not found")
	}
	message.Attempts++
	message.LastError = reason
	message.AvailableAt = availableAt.UTC()
	if dead || message.Attempts >= message.MaximumAttempts {
		when := time.Now().UTC()
		message.DeadAt = &when
	}
	s.messages[id] = message
	return nil
}
