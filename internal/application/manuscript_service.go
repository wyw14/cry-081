package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wyw14/cry-081/internal/domain/audit"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/platform/idempotency"
	"github.com/wyw14/cry-081/internal/platform/outbox"
)

type ManuscriptService struct {
	manuscripts  ManuscriptRepository
	users        UserRepository
	audits       AuditRepository
	transactions TransactionManager
	outbox       outbox.Store
	replays      idempotency.Store
	ids          IDGenerator
	clock        clock.Source
	policy       manuscript.FormatPolicy
}

func NewManuscriptService(manuscripts ManuscriptRepository, users UserRepository, audits AuditRepository, transactions TransactionManager, outboxStore outbox.Store, replayStore idempotency.Store, ids IDGenerator, clock clock.Source, policy manuscript.FormatPolicy) *ManuscriptService {
	return &ManuscriptService{manuscripts: manuscripts, users: users, audits: audits, transactions: transactions, outbox: outboxStore, replays: replayStore, ids: ids, clock: clock, policy: policy}
}

type CreateManuscriptInput struct {
	AuthorID    string
	SectionID   string
	Title       string
	Abstract    string
	Markdown    string
	Tags        []string
	Attachments []manuscript.Attachment
}

func (s *ManuscriptService) Create(ctx context.Context, input CreateManuscriptInput) (manuscript.Manuscript, error) {
	author, err := s.users.Get(ctx, input.AuthorID)
	if err != nil {
		return manuscript.Manuscript{}, err
	}
	if !author.HasRole(identity.RoleAuthor) || !author.Active {
		return manuscript.Manuscript{}, shared.NewError("AUTHOR_ROLE_REQUIRED", "active author role is required", shared.ErrForbidden)
	}
	now := s.clock.UTCNow()
	version, err := manuscript.NewDraftVersion(1, input.Title, input.Abstract, input.Markdown, input.Tags, input.Attachments, now)
	if err != nil {
		return manuscript.Manuscript{}, err
	}
	if result := s.policy.Check(version); !result.Passed {
		return manuscript.Manuscript{}, shared.ValidationError(shared.FieldViolation{Field: "markdown", Message: result.Violations[0]})
	}
	created, err := manuscript.New(s.ids.NewID(), author.ID, input.SectionID, version, now)
	if err != nil {
		return manuscript.Manuscript{}, err
	}
	if err := s.manuscripts.Create(ctx, *created); err != nil {
		return manuscript.Manuscript{}, err
	}
	return created.Clone(), nil
}

type SubmitManuscriptInput struct {
	ManuscriptID    string
	AuthorID        string
	Declaration     manuscript.Declaration
	ExpectedVersion int64
	IdempotencyKey  string
}

func (s *ManuscriptService) Submit(ctx context.Context, input SubmitManuscriptInput) (manuscript.SubmissionSnapshot, error) {
	if input.IdempotencyKey == "" {
		return manuscript.SubmissionSnapshot{}, shared.NewError("IDEMPOTENCY_KEY_REQUIRED", "submission requires an idempotency key", shared.ErrValidation)
	}
	digest, err := submissionRequestDigest(input)
	if err != nil {
		return manuscript.SubmissionSnapshot{}, err
	}
	replayIdentity := submissionReplayIdentity(input)
	previous, found, err := s.replays.Find(ctx, replayIdentity)
	if err != nil {
		return manuscript.SubmissionSnapshot{}, err
	}
	if found {
		if previous.Digest != digest {
			return manuscript.SubmissionSnapshot{}, shared.NewError("IDEMPOTENCY_KEY_CONFLICT", "idempotency key was used for a different submission", shared.ErrConflict)
		}
		var replay manuscript.SubmissionSnapshot
		if err := json.Unmarshal(previous.Response, &replay); err != nil {
			return manuscript.SubmissionSnapshot{}, shared.NewError("IDEMPOTENCY_REPLAY_INVALID", "stored submission response is invalid", err)
		}
		return replay.Clone(), nil
	}
	var result manuscript.SubmissionSnapshot
	err = s.transactions.Within(ctx, func(tx context.Context) error {
		current, err := s.manuscripts.Get(tx, input.ManuscriptID)
		if err != nil {
			return err
		}
		if current.AuthorID != input.AuthorID {
			return shared.NewError("MANUSCRIPT_OWNER_REQUIRED", "only the author can submit this manuscript", shared.ErrForbidden)
		}
		if current.Version != input.ExpectedVersion {
			return shared.NewError("MANUSCRIPT_VERSION_CONFLICT", "manuscript was changed by another request", shared.ErrVersionConflict)
		}
		version, err := current.CurrentVersion()
		if err != nil {
			return err
		}
		if preflight := s.policy.Check(version); !preflight.Passed {
			return shared.NewError("PREFLIGHT_FAILED", "manuscript format preflight failed", shared.ErrValidation)
		}
		matches, err := s.manuscripts.FindByFingerprint(tx, version.Fingerprint)
		if err != nil {
			return err
		}
		for _, match := range matches {
			if match.ID != current.ID {
				return shared.NewError("DUPLICATE_FINGERPRINT", "the manuscript matches another submission", shared.ErrConflict)
			}
		}
		now := s.clock.UTCNow()
		snapshot, err := manuscript.NewSubmissionSnapshot(s.ids.NewID(), current.ID, input.AuthorID, version, input.Declaration, now)
		if err != nil {
			return err
		}
		before := current.Clone()
		expected := current.Version
		if err := current.Submit(snapshot, now); err != nil {
			return err
		}
		if err := s.manuscripts.Update(tx, current, expected); err != nil {
			return err
		}
		event, err := audit.NewEvent(s.ids.NewID(), input.AuthorID, "api", "submit", "manuscript", current.ID, "author submission", before, current, now)
		if err != nil {
			return err
		}
		if err := s.audits.Append(tx, event); err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{"manuscript_id": current.ID, "snapshot_id": snapshot.ID})
		if err != nil {
			return err
		}
		if err := s.outbox.Enqueue(tx, outbox.Message{ID: s.ids.NewID(), Topic: "manuscript.submitted", Key: current.ID, Payload: payload, MaximumAttempts: 5, AvailableAt: now}); err != nil {
			return err
		}
		result = snapshot.Clone()
		response, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return s.replays.Remember(tx, idempotency.Replay{Identity: replayIdentity, Digest: digest, Response: response, StatusCode: 200, ExpiresAt: now.Add(24 * time.Hour)})
	})
	return result, err
}

func submissionReplayIdentity(input SubmitManuscriptInput) idempotency.Identity {
	return idempotency.Identity{
		Operation: "manuscript-submission:" + input.ManuscriptID,
		Key:       input.IdempotencyKey,
	}
}

func submissionRequestDigest(input SubmitManuscriptInput) (string, error) {
	payload, err := json.Marshal(struct {
		ManuscriptID    string
		AuthorID        string
		Declaration     manuscript.Declaration
		ExpectedVersion int64
	}{input.ManuscriptID, input.AuthorID, input.Declaration, input.ExpectedVersion})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func (s *ManuscriptService) AddRevision(ctx context.Context, manuscriptID, authorID string, expectedVersion int64, draft manuscript.DraftVersion) (manuscript.Manuscript, error) {
	current, err := s.manuscripts.Get(ctx, manuscriptID)
	if err != nil {
		return manuscript.Manuscript{}, err
	}
	if current.Version != expectedVersion {
		return manuscript.Manuscript{}, shared.NewError("MANUSCRIPT_VERSION_CONFLICT", "manuscript was changed by another request", shared.ErrVersionConflict)
	}
	before := current.Clone()
	if err := current.AddDraft(draft, authorID, s.clock.UTCNow()); err != nil {
		return manuscript.Manuscript{}, err
	}
	if result := s.policy.Check(draft); !result.Passed {
		return manuscript.Manuscript{}, shared.NewError("PREFLIGHT_FAILED", fmt.Sprintf("revision failed %d checks", len(result.Violations)), shared.ErrValidation)
	}
	if err := s.manuscripts.Update(ctx, current, expectedVersion); err != nil {
		return manuscript.Manuscript{}, err
	}
	event, err := audit.NewEvent(s.ids.NewID(), authorID, "api", "revise", "manuscript", manuscriptID, "author revision", before, current, s.clock.UTCNow())
	if err == nil {
		err = s.audits.Append(ctx, event)
	}
	if err != nil {
		return manuscript.Manuscript{}, err
	}
	return current.Clone(), nil
}

func (s *ManuscriptService) List(ctx context.Context, actor identity.User, filter ManuscriptFilter) (ManuscriptPage, error) {
	if !actor.Can(identity.PermissionSubmissionRead) {
		return ManuscriptPage{}, shared.NewError("SUBMISSION_READ_FORBIDDEN", "actor cannot read submissions", shared.ErrForbidden)
	}
	if actor.HasRole(identity.RoleAuthor) && !actor.HasRole(identity.RoleEditor) && !actor.HasRole(identity.RoleChiefEditor) {
		filter.AuthorID = actor.ID
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 25
	}
	allowedSort := map[string]bool{"created_at": true, "updated_at": true, "status": true}
	if !allowedSort[filter.Sort] {
		filter.Sort = "updated_at"
	}
	return s.manuscripts.List(ctx, filter)
}

func encodeReminderPayload(assignmentID string, dueAt time.Time) ([]byte, error) {
	return json.Marshal(map[string]any{"assignment_id": assignmentID, "due_at": dueAt.UTC().Format(time.RFC3339)})
}
