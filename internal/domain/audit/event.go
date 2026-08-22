package audit

import (
	"encoding/json"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type Event struct {
	ID         string          `json:"id"`
	ActorID    string          `json:"actor_id"`
	Source     string          `json:"source"`
	Action     string          `json:"action"`
	Resource   string          `json:"resource"`
	ResourceID string          `json:"resource_id"`
	Before     json.RawMessage `json:"before"`
	After      json.RawMessage `json:"after"`
	Reason     string          `json:"reason"`
	CreatedAt  time.Time       `json:"created_at"`
}

func NewEvent(id, actorID, source, action, resource, resourceID, reason string, before, after any, now time.Time) (Event, error) {
	if id == "" || actorID == "" || source == "" || action == "" || resource == "" || resourceID == "" {
		return Event{}, shared.NewError("AUDIT_EVENT_INVALID", "audit identity is incomplete", shared.ErrValidation)
	}
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return Event{}, shared.NewError("AUDIT_BEFORE_INVALID", "cannot encode previous value", err)
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return Event{}, shared.NewError("AUDIT_AFTER_INVALID", "cannot encode current value", err)
	}
	return Event{ID: id, ActorID: actorID, Source: source, Action: action, Resource: resource, ResourceID: resourceID, Before: beforeJSON, After: afterJSON, Reason: reason, CreatedAt: now.UTC()}, nil
}
