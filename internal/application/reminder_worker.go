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
	assignment, deliveryErr := w.assignments.Get(ctx, message.Key)
	if deliveryErr == nil {
		deliveryErr = w.sender.SendReviewReminder(ctx, assignment.ReviewerID, assignment.DueAt)
	}
	if deliveryErr == nil {
		previousVersion := assignment.Version
		deliveryErr = assignment.RecordReminder(w.clock.UTCNow())
		if deliveryErr == nil {
			deliveryErr = w.assignments.Update(ctx, assignment, previousVersion)
		}
	}
	if deliveryErr == nil {
		return w.queue.MarkProcessed(ctx, message.ID, w.clock.UTCNow())
	}

	nextAttempt := message.Attempts + 1
	retryDelay := time.Second
	for count := 1; count < nextAttempt && retryDelay < w.maxBackoff; count++ {
		retryDelay *= 2
	}
	if retryDelay > w.maxBackoff {
		retryDelay = w.maxBackoff
	}
	dead := nextAttempt >= message.MaximumAttempts
	return w.queue.Reschedule(ctx, message.ID, deliveryErr.Error(), w.clock.UTCNow().Add(retryDelay), dead)
}
