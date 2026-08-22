package postgres

import (
	"context"
	"time"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/audit"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/platform/idempotency"
	"github.com/wyw14/cry-081/internal/platform/outbox"
)

type AuditStore struct{ db *Database }

func NewAuditStore(db *Database) *AuditStore { return &AuditStore{db: db} }

func (s *AuditStore) Append(ctx context.Context, event audit.Event) error {
	_, err := s.db.queries(ctx).Exec(ctx, `INSERT INTO audit_events (id,actor_id,source,action,resource,resource_id,before_value,after_value,reason,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, event.ID, event.ActorID, event.Source, event.Action, event.Resource, event.ResourceID, event.Before, event.After, event.Reason, event.CreatedAt)
	return translate(err)
}

type IdempotencyStore struct{ db *Database }

func NewIdempotencyStore(db *Database) *IdempotencyStore { return &IdempotencyStore{db: db} }

func (s *IdempotencyStore) Find(ctx context.Context, identity idempotency.Identity) (idempotency.Replay, bool, error) {
	var replay idempotency.Replay
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT key,scope,request_digest,response_body,status_code,expires_at FROM idempotency_records WHERE scope=$1 AND key=$2 AND expires_at > now()`, identity.Operation, identity.Key).Scan(&replay.Identity.Key, &replay.Identity.Operation, &replay.Digest, &replay.Response, &replay.StatusCode, &replay.ExpiresAt)
	if err == nil {
		return replay, true, nil
	}
	if translate(err) == shared.ErrNotFound {
		return idempotency.Replay{}, false, nil
	}
	return idempotency.Replay{}, false, translate(err)
}

func (s *IdempotencyStore) Remember(ctx context.Context, replay idempotency.Replay) error {
	_, err := s.db.queries(ctx).Exec(ctx, `INSERT INTO idempotency_records (scope,key,request_digest,response_body,status_code,expires_at) VALUES ($1,$2,$3,$4,$5,$6)`, replay.Identity.Operation, replay.Identity.Key, replay.Digest, replay.Response, replay.StatusCode, replay.ExpiresAt)
	return translate(err)
}

func (s *AuditStore) List(ctx context.Context, filter application.AuditFilter) ([]audit.Event, error) {
	if filter.Limit < 1 || filter.Limit > 1000 {
		filter.Limit = 100
	}
	rows, err := s.db.queries(ctx).Query(ctx, `SELECT id,actor_id,source,action,resource,resource_id,before_value,after_value,reason,created_at FROM audit_events WHERE ($1='' OR resource=$1) AND ($2='' OR resource_id=$2) AND ($3='' OR actor_id=$3) ORDER BY created_at DESC LIMIT $4`, filter.Resource, filter.ResourceID, filter.ActorID, filter.Limit)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]audit.Event, 0)
	for rows.Next() {
		var event audit.Event
		if err := rows.Scan(&event.ID, &event.ActorID, &event.Source, &event.Action, &event.Resource, &event.ResourceID, &event.Before, &event.After, &event.Reason, &event.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, event)
	}
	return items, rows.Err()
}

type OutboxStore struct{ db *Database }

func NewOutboxStore(db *Database) *OutboxStore { return &OutboxStore{db: db} }

func (s *OutboxStore) Enqueue(ctx context.Context, message outbox.Message) error {
	_, err := s.db.queries(ctx).Exec(ctx, `INSERT INTO outbox_messages (id,topic,message_key,payload,attempts,maximum_attempts,available_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, message.ID, message.Topic, message.Key, message.Payload, message.Attempts, message.MaximumAttempts, message.AvailableAt)
	return translate(err)
}

func (s *OutboxStore) Claim(ctx context.Context, now time.Time, limit int) ([]outbox.Message, error) {
	rows, err := s.db.queries(ctx).Query(ctx, `SELECT id,topic,message_key,payload,attempts,maximum_attempts,available_at,processed_at,dead_at,last_error FROM outbox_messages WHERE processed_at IS NULL AND dead_at IS NULL AND available_at <= $1 ORDER BY available_at FOR UPDATE SKIP LOCKED LIMIT $2`, now, limit)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	items := make([]outbox.Message, 0)
	for rows.Next() {
		var item outbox.Message
		if err := rows.Scan(&item.ID, &item.Topic, &item.Key, &item.Payload, &item.Attempts, &item.MaximumAttempts, &item.AvailableAt, &item.ProcessedAt, &item.DeadAt, &item.LastError); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OutboxStore) MarkProcessed(ctx context.Context, id string, now time.Time) error {
	_, err := s.db.queries(ctx).Exec(ctx, `UPDATE outbox_messages SET processed_at=$1 WHERE id=$2`, now, id)
	return translate(err)
}

func (s *OutboxStore) Reschedule(ctx context.Context, id, reason string, availableAt time.Time, dead bool) error {
	var deadAt *time.Time
	if dead {
		when := time.Now().UTC()
		deadAt = &when
	}
	_, err := s.db.queries(ctx).Exec(ctx, `UPDATE outbox_messages SET attempts=attempts+1,last_error=$1,available_at=$2,dead_at=$3 WHERE id=$4`, reason, availableAt, deadAt, id)
	return translate(err)
}
