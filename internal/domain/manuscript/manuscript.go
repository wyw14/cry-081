package manuscript

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type Transition struct {
	From      Status
	To        Status
	ActorID   string
	Reason    string
	Version   int64
	CreatedAt time.Time
}

type Manuscript struct {
	ID              string
	AuthorID        string
	SectionID       string
	Status          Status
	Versions        []DraftVersion
	Submissions     []SubmissionSnapshot
	ActiveVersion   int64
	DecisionVersion int64
	Version         int64
	Transitions     []Transition
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func New(id, authorID, sectionID string, first DraftVersion, now time.Time) (*Manuscript, error) {
	if id == "" || authorID == "" || sectionID == "" || first.Number != 1 {
		return nil, shared.NewError("MANUSCRIPT_INVALID", "manuscript identity is incomplete", shared.ErrValidation)
	}
	now = now.UTC()
	return &Manuscript{
		ID: id, AuthorID: authorID, SectionID: sectionID, Status: StatusDraft,
		Versions: []DraftVersion{first.Clone()}, ActiveVersion: 1, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (m *Manuscript) CurrentVersion() (DraftVersion, error) {
	for i := len(m.Versions) - 1; i >= 0; i-- {
		if m.Versions[i].Number == m.ActiveVersion {
			return m.Versions[i].Clone(), nil
		}
	}
	return DraftVersion{}, shared.NewError("VERSION_NOT_FOUND", "active manuscript version does not exist", shared.ErrNotFound)
}

func (m *Manuscript) AddDraft(version DraftVersion, actorID string, now time.Time) error {
	request := revisionDraftRequest{actorID: actorID, version: version.Clone(), now: now.UTC()}
	if err := m.validateRevisionDraft(request); err != nil {
		return err
	}
	locks := m.revisionLocks()
	if locks.blocks(m.ID, request.version.Number) {
		return shared.NewError("DRAFT_LOCKED_BY_SUBMISSION", "submitted versions cannot be edited", shared.ErrInvalidState)
	}
	m.applyRevisionDraft(request)
	return nil
}

type revisionDraftRequest struct {
	actorID string
	version DraftVersion
	now     time.Time
}

type revisionLockSet struct {
	items []SubmissionVersionLock
}

func (m Manuscript) validateRevisionDraft(request revisionDraftRequest) error {
	if request.actorID != m.AuthorID {
		return shared.NewError("MANUSCRIPT_OWNER_REQUIRED", "only the author can revise the draft", shared.ErrForbidden)
	}
	if m.Status != StatusDraft && m.Status != StatusRevisionNeeded {
		return shared.NewError("DRAFT_LOCKED", "manuscript cannot be edited in its current state", shared.ErrInvalidState)
	}
	if request.version.Number != m.ActiveVersion+1 {
		return shared.NewError("VERSION_SEQUENCE_INVALID", "draft version must be sequential", shared.ErrConflict)
	}
	if request.version.Locked {
		return shared.NewError("NEW_DRAFT_ALREADY_LOCKED", "a new revision must remain editable until submission", shared.ErrConflict)
	}
	if request.now.IsZero() {
		return shared.NewError("REVISION_TIME_REQUIRED", "revision time is required", shared.ErrValidation)
	}
	return nil
}

func (m Manuscript) revisionLocks() revisionLockSet {
	set := revisionLockSet{items: make([]SubmissionVersionLock, 0, len(m.Submissions))}
	for _, snapshot := range m.Submissions {
		lock := snapshot.VersionLock()
		if lock.Valid() {
			set.items = append(set.items, lock)
		}
	}
	return set
}

func (s revisionLockSet) blocks(manuscriptID string, versionNumber int64) bool {
	for _, lock := range s.items {
		if lock.AppliesTo(manuscriptID, versionNumber) {
			return true
		}
	}
	return false
}

func (m *Manuscript) applyRevisionDraft(request revisionDraftRequest) {
	m.Versions = append(m.Versions, request.version.Clone())
	m.ActiveVersion = request.version.Number
	m.Version++
	m.UpdatedAt = request.now
}

func (m *Manuscript) AddCorrection(version DraftVersion, actorID string, now time.Time) error {
	if actorID != m.AuthorID {
		return shared.NewError("MANUSCRIPT_OWNER_REQUIRED", "only the author can submit a correction", shared.ErrForbidden)
	}
	if m.Status != StatusAccepted && m.Status != StatusScheduled && m.Status != StatusPublished {
		return shared.NewError("CORRECTION_NOT_ALLOWED", "corrections are only allowed after acceptance", shared.ErrInvalidState)
	}
	if !version.Correction || version.Number != m.ActiveVersion+1 {
		return shared.NewError("CORRECTION_INVALID", "a correction must be the next marked version", shared.ErrValidation)
	}
	m.Versions = append(m.Versions, version.Clone())
	m.ActiveVersion = version.Number
	m.Version++
	m.UpdatedAt = now.UTC()
	return nil
}

func (m *Manuscript) Submit(snapshot SubmissionSnapshot, now time.Time) error {
	if snapshot.ManuscriptID != m.ID || snapshot.VersionNumber != m.ActiveVersion {
		return shared.NewError("SUBMISSION_VERSION_MISMATCH", "submission must snapshot the active version", shared.ErrConflict)
	}
	if m.Status != StatusDraft && m.Status != StatusRevisionNeeded {
		return shared.NewError("SUBMISSION_STATE_INVALID", "manuscript cannot be submitted in its current state", shared.ErrInvalidState)
	}
	current, err := m.CurrentVersion()
	if err != nil {
		return err
	}
	if current.ContentDigest() != snapshot.ContentDigest {
		return shared.NewError("SUBMISSION_DIGEST_MISMATCH", "submission snapshot does not match the active draft", shared.ErrConflict)
	}
	next := StatusSubmitted
	if m.Status == StatusRevisionNeeded {
		next = StatusUnderReReview
	}
	for i := range m.Versions {
		if m.Versions[i].Number == snapshot.VersionNumber {
			m.Versions[i].Locked = true
		}
	}
	m.Submissions = append(m.Submissions, snapshot.Clone())
	return m.transition(next, snapshot.SubmittedBy, fmt.Sprintf("submission %s", snapshot.ID), snapshot.VersionNumber, now)
}

func (m *Manuscript) Transition(to Status, actorID, reason string, decisionVersion int64, now time.Time) error {
	if reason == "" && (to == StatusRevisionNeeded || to == StatusRejected || to == StatusAccepted || to == StatusWithdrawn) {
		return shared.NewError("TRANSITION_REASON_REQUIRED", "a decision reason is required", shared.ErrValidation)
	}
	return m.transition(to, actorID, reason, decisionVersion, now)
}

func (m *Manuscript) transition(to Status, actorID, reason string, version int64, now time.Time) error {
	if actorID == "" {
		return shared.NewError("ACTOR_REQUIRED", "transition actor is required", shared.ErrValidation)
	}
	if err := RequireTransition(m.Status, to); err != nil {
		return err
	}
	if version != m.ActiveVersion {
		return shared.NewError("DECISION_VERSION_MISMATCH", "decision must target the active manuscript version", shared.ErrVersionConflict)
	}
	previous := m.Status
	m.Status = to
	m.DecisionVersion = version
	m.Version++
	m.UpdatedAt = now.UTC()
	m.Transitions = append(m.Transitions, Transition{From: previous, To: to, ActorID: actorID, Reason: reason, Version: version, CreatedAt: now.UTC()})
	return nil
}

func (m Manuscript) Clone() Manuscript {
	versions := m.Versions
	submissions := m.Submissions
	transitions := m.Transitions
	m.Versions = make([]DraftVersion, len(versions))
	for i, version := range versions {
		m.Versions[i] = version.Clone()
	}
	m.Submissions = make([]SubmissionSnapshot, len(submissions))
	for i, submission := range submissions {
		m.Submissions[i] = submission.Clone()
	}
	m.Transitions = append([]Transition(nil), transitions...)
	return m
}
