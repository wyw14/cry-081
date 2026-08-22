package application

import (
	"context"
	"time"

	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/platform/outbox"
)

type ReminderSender interface {
	SendReviewReminder(context.Context, string, time.Time) error
}

type ReminderWorker struct {
	assignments AssignmentRepository
	queue       outbox.Store
	sender      ReminderSender
	clock       clock.Source
	pollEvery   time.Duration
	batchSize   int
	maxBackoff  time.Duration
}

func NewReminderWorker(assignments AssignmentRepository, queue outbox.Store, sender ReminderSender, source clock.Source) *ReminderWorker {
	return &ReminderWorker{
		assignments: assignments,
		queue:       queue,
		sender:      sender,
		clock:       source,
		pollEvery:   time.Second,
		batchSize:   20,
		maxBackoff:  time.Minute,
	}
}

func (w *ReminderWorker) Run(ctx context.Context) error {
	nextPoll := time.NewTimer(0)
	defer nextPoll.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-nextPoll.C:
			_ = w.ProcessOnce(ctx)
			nextPoll.Reset(w.pollEvery)
		}
	}
}

func (w *ReminderWorker) ProcessOnce(ctx context.Context) error {
	pending, err := w.queue.Claim(ctx, w.clock.UTCNow(), w.batchSize)
	if err != nil {
		return err
	}
	for index := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		if pending[index].Topic != "review.reminder" {
			continue
		}
		if err := w.deliver(ctx, pending[index]); err != nil {
			return err
		}
	}
	return nil
}

func (w *ReminderWorker) deliver(ctx context.Context, message outbox.Message) error {
	attempt := reminderAttempt{worker: w, message: message, startedAt: w.clock.UTCNow()}
	if err := attempt.reserve(ctx); err != nil {
		return attempt.retry(ctx, err)
	}
	if err := attempt.load(ctx); err != nil {
		return attempt.retry(ctx, err)
	}
	if err := attempt.send(ctx); err != nil {
		return attempt.retry(ctx, err)
	}
	if err := attempt.record(ctx); err != nil {
		return attempt.retry(ctx, err)
	}
	return nil
}

type reminderAttempt struct {
	worker          *ReminderWorker
	message         outbox.Message
	assignment      reviewAssignment
	assignmentReady bool
	startedAt       time.Time
}

type reviewAssignment struct {
	ID         string
	ReviewerID string
	DueAt      time.Time
	Version    int64
	record     func(time.Time) error
	persist    func(context.Context, int64) error
}

func (a *reminderAttempt) reserve(ctx context.Context) error {
	return a.worker.queue.MarkProcessed(ctx, a.message.ID, a.startedAt)
}

func (a *reminderAttempt) load(ctx context.Context) error {
	assignment, err := a.worker.assignments.Get(ctx, a.message.Key)
	if err != nil {
		return err
	}
	a.assignment = reviewAssignment{
		ID: assignment.ID, ReviewerID: assignment.ReviewerID, DueAt: assignment.DueAt,
		Version: assignment.Version,
		record: func(now time.Time) error {
			return assignment.RecordReminder(now)
		},
		persist: func(updateContext context.Context, expectedVersion int64) error {
			return a.worker.assignments.Update(updateContext, assignment, expectedVersion)
		},
	}
	a.assignmentReady = true
	return nil
}

func (a *reminderAttempt) send(ctx context.Context) error {
	if !a.assignmentReady {
		return context.Canceled
	}
	return a.worker.sender.SendReviewReminder(ctx, a.assignment.ReviewerID, a.assignment.DueAt)
}

func (a *reminderAttempt) record(ctx context.Context) error {
	if err := a.assignment.record(a.startedAt); err != nil {
		return err
	}
	return a.assignment.persist(ctx, a.assignment.Version)
}

func (a *reminderAttempt) retry(ctx context.Context, cause error) error {
	nextAttempt := a.message.Attempts + 1
	delay := time.Second
	for count := 1; count < nextAttempt && delay < a.worker.maxBackoff; count++ {
		delay *= 2
	}
	if delay > a.worker.maxBackoff {
		delay = a.worker.maxBackoff
	}
	dead := nextAttempt >= a.message.MaximumAttempts
	return a.worker.queue.Reschedule(ctx, a.message.ID, cause.Error(), a.startedAt.Add(delay), dead)
}
