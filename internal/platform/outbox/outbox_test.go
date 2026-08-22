package outbox_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/platform/outbox"
)

func TestMemoryStoreHonorsCancellationAndDeadLetterState(t *testing.T) {
	store := outbox.NewMemoryStore()
	now := time.Now().UTC()
	message := outbox.Message{ID: "message-1", Topic: "review.reminder", Key: "assignment-1", Payload: []byte("payload"), MaximumAttempts: 2, AvailableAt: now}
	if err := store.Enqueue(context.Background(), message); err != nil {
		t.Fatal(err)
	}
	items, err := store.Claim(context.Background(), now, 10)
	if err != nil || len(items) != 1 {
		t.Fatalf("claim failed: %#v %v", items, err)
	}
	if err := store.Reschedule(context.Background(), message.ID, "temporary", now, false); err != nil {
		t.Fatal(err)
	}
	if err := store.Reschedule(context.Background(), message.ID, "permanent", now, true); err != nil {
		t.Fatal(err)
	}
	items, _ = store.Claim(context.Background(), now.Add(time.Hour), 10)
	if len(items) != 0 {
		t.Fatal("dead-lettered message was claimed again")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.Enqueue(cancelled, outbox.Message{ID: "message-2", Topic: "x", MaximumAttempts: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context was ignored: %v", err)
	}
}
