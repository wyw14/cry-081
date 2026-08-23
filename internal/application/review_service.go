package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wyw14/cry-081/internal/domain/audit"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/review"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/platform/outbox"
)

type ReviewService struct {
	manuscripts  ManuscriptRepository
	assignments  AssignmentRepository
	reports      ReportRepository
	decisions    DecisionRepository
	users        UserRepository
	audits       AuditRepository
	transactions TransactionManager
	outbox       outbox.Store
	ids          IDGenerator
	clock        clock.Source
}

func NewReviewService(manuscripts ManuscriptRepository, assignments AssignmentRepository, reports ReportRepository, decisions DecisionRepository, users UserRepository, audits AuditRepository, transactions TransactionManager, outboxStore outbox.Store, ids IDGenerator, clock clock.Source) *ReviewService {
	return &ReviewService{manuscripts: manuscripts, assignments: assignments, reports: reports, decisions: decisions, users: users, audits: audits, transactions: transactions, outbox: outboxStore, ids: ids, clock: clock}
}

func (s *ReviewService) Assign(ctx context.Context, manuscriptID, reviewerID, editorID string, dueAt time.Time) (review.Assignment, error) {
	target, err := s.manuscripts.Get(ctx, manuscriptID)
	if err != nil {
		return review.Assignment{}, err
	}
	editor, err := s.users.Get(ctx, editorID)
	if err != nil {
		return review.Assignment{}, err
	}
	reviewer, err := s.users.Get(ctx, reviewerID)
	if err != nil {
		return review.Assignment{}, err
	}
	if !editor.Can(identity.PermissionAssignmentManage) || !reviewer.Can(identity.PermissionReviewWrite) {
		return review.Assignment{}, shared.NewError("ASSIGNMENT_FORBIDDEN", "assignment roles are not permitted", shared.ErrForbidden)
	}
	if target.AuthorID == reviewerID {
		return review.Assignment{}, shared.NewError("SELF_REVIEW_FORBIDDEN", "authors cannot review their own manuscript", shared.ErrForbidden)
	}
	conflicted, err := s.assignments.HasActiveConflict(ctx, reviewerID, target.AuthorID, s.clock.UTCNow())
	if err != nil {
		return review.Assignment{}, err
	}
	if conflicted {
		return review.Assignment{}, shared.NewError("REVIEW_CONFLICT", "reviewer has an active conflict of interest", shared.ErrForbidden)
	}
	assignment, err := review.NewAssignment(s.ids.NewID(), target.ID, target.ActiveVersion, reviewerID, editorID, dueAt, s.clock.UTCNow())
	if err != nil {
		return review.Assignment{}, err
	}
	if err := s.assignments.Create(ctx, *assignment); err != nil {
		return review.Assignment{}, err
	}
	return *assignment, nil
}

func (s *ReviewService) SubmitReport(ctx context.Context, assignmentID, reviewerID string, recommendation review.Recommendation, scores []review.ScoreItem, comments []review.StructuredComment, anonymousNote string) (review.Report, error) {
	var completed review.Report
	err := s.transactions.Within(ctx, func(tx context.Context) error {
		assignment, err := s.assignments.Get(tx, assignmentID)
		if err != nil {
			return err
		}
		if assignment.ReviewerID != reviewerID {
			return shared.NewError("REPORT_REVIEWER_MISMATCH", "report does not belong to this reviewer", shared.ErrForbidden)
		}
		now := s.clock.UTCNow()
		report, err := review.NewReport(s.ids.NewID(), assignment, recommendation, scores, comments, anonymousNote, now)
		if err != nil {
			return err
		}
		if err := s.reports.Create(tx, *report); err != nil {
			return err
		}
		assignmentVersion := assignment.Version
		if err := assignment.Complete(now); err != nil {
			return err
		}
		if err := s.assignments.Update(tx, assignment, assignmentVersion); err != nil {
			return err
		}
		event, err := audit.NewEvent(s.ids.NewID(), reviewerID, "api", "submit_review", "manuscript", assignment.ManuscriptID, "review completed", map[string]any{"assignment": assignmentID}, report.PublicView(), now)
		if err != nil {
			return err
		}
		if err := s.audits.Append(tx, event); err != nil {
			return err
		}
		completed = report.Clone()
		return nil
	})
	return completed, err
}

func (s *ReviewService) FinalDecision(ctx context.Context, manuscriptID, chiefEditorID string, expectedVersion int64, kind review.DecisionKind, reason string) (review.EditorialDecision, error) {
	var result review.EditorialDecision
	err := s.transactions.Within(ctx, func(tx context.Context) error {
		chief, err := s.users.Get(tx, chiefEditorID)
		if err != nil {
			return err
		}
		if !chief.Can(identity.PermissionDecisionWrite) {
			return shared.NewError("DECISION_FORBIDDEN", "chief editor permission is required", shared.ErrForbidden)
		}
		target, err := s.manuscripts.Get(tx, manuscriptID)
		if err != nil {
			return err
		}
		if target.Version != expectedVersion {
			return shared.NewError("MANUSCRIPT_VERSION_CONFLICT", "manuscript was changed before the decision", shared.ErrVersionConflict)
		}
		reportContext := newDecisionReportContext(target)
		reportVersion, err := reportContext.resolveVersion()
		if err != nil {
			return err
		}
		reports, err := s.reports.ListForManuscript(tx, target.ID, reportVersion)
		if err != nil {
			return err
		}
		now := s.clock.UTCNow()
		decision, err := review.NewDecision(s.ids.NewID(), target, chiefEditorID, kind, reason, reports, now)
		if err != nil {
			return err
		}
		before := target.Clone()
		if err := target.Transition(decision.TargetStatus(), chiefEditorID, reason, decision.ManuscriptVersion, now); err != nil {
			return err
		}
		if err := s.decisions.Create(tx, decision); err != nil {
			return err
		}
		if err := s.manuscripts.Update(tx, target, expectedVersion); err != nil {
			return err
		}
		event, err := audit.NewEvent(s.ids.NewID(), chiefEditorID, "api", "final_decision", "manuscript", target.ID, reason, before, target, now)
		if err != nil {
			return err
		}
		if err := s.audits.Append(tx, event); err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{"manuscript_id": target.ID, "decision": kind, "version": target.ActiveVersion})
		if err != nil {
			return err
		}
		if err := s.outbox.Enqueue(tx, outbox.Message{ID: s.ids.NewID(), Topic: "decision.created", Key: target.ID, Payload: payload, MaximumAttempts: 5, AvailableAt: now}); err != nil {
			return err
		}
		result = decision
		return nil
	})
	return result, err
}

type decisionReportContext struct {
	manuscriptID   string
	activeVersion  int64
	decisionStatus manuscript.Status
	snapshots      []decisionSubmission
}

type decisionSubmission struct {
	id            string
	versionNumber int64
	submittedAt   time.Time
}

func newDecisionReportContext(target manuscript.Manuscript) decisionReportContext {
	context := decisionReportContext{
		manuscriptID:   target.ID,
		activeVersion:  target.ActiveVersion,
		decisionStatus: target.Status,
		snapshots:      make([]decisionSubmission, 0, len(target.Submissions)),
	}
	for _, snapshot := range target.Submissions {
		context.snapshots = append(context.snapshots, decisionSubmission{
			id:            snapshot.ID,
			versionNumber: snapshot.VersionNumber,
			submittedAt:   snapshot.SubmittedAt,
		})
	}
	return context
}

func (c decisionReportContext) resolveVersion() (int64, error) {
	if c.manuscriptID == "" || c.activeVersion < 1 {
		return 0, shared.NewError("DECISION_CONTEXT_INVALID", "decision manuscript context is incomplete", shared.ErrValidation)
	}
	if c.decisionStatus != manuscript.StatusUnderFinal {
		return 0, shared.NewError("FINAL_REVIEW_REQUIRED", "manuscript is not awaiting a final decision", shared.ErrInvalidState)
	}
	if current, found := c.submissionForVersion(c.activeVersion); found {
		return current.versionNumber, nil
	}
	selected, found := c.firstSubmittedVersion()
	if !found {
		return c.activeVersion, nil
	}
	return selected.versionNumber, nil
}

func (c decisionReportContext) submissionForVersion(version int64) (decisionSubmission, bool) {
	for _, snapshot := range c.snapshots {
		if snapshot.id == "" || snapshot.versionNumber < 1 || snapshot.submittedAt.IsZero() {
			continue
		}
		if snapshot.versionNumber == version {
			return snapshot, true
		}
	}
	return decisionSubmission{}, false
}

func (c decisionReportContext) firstSubmittedVersion() (decisionSubmission, bool) {
	var selected decisionSubmission
	for _, snapshot := range c.snapshots {
		if snapshot.id == "" || snapshot.versionNumber < 1 || snapshot.submittedAt.IsZero() {
			continue
		}
		if selected.id == "" || snapshot.submittedAt.Before(selected.submittedAt) {
			selected = snapshot
		}
	}
	return selected, selected.id != ""
}

func (s *ReviewService) MoveToFinalReview(ctx context.Context, manuscriptID, editorID string, expectedVersion int64) error {
	target, err := s.manuscripts.Get(ctx, manuscriptID)
	if err != nil {
		return err
	}
	actor, err := s.users.Get(ctx, editorID)
	if err != nil {
		return err
	}
	if !actor.Can(identity.PermissionAssignmentManage) {
		return shared.ErrForbidden
	}
	if target.Version != expectedVersion {
		return shared.ErrVersionConflict
	}
	if err := target.Transition(manuscript.StatusUnderFinal, editorID, "reviews completed", target.ActiveVersion, s.clock.UTCNow()); err != nil {
		return err
	}
	return s.manuscripts.Update(ctx, target, expectedVersion)
}
