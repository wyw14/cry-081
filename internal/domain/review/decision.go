package review

import (
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type DecisionKind string

const (
	DecisionRevision DecisionKind = "revision"
	DecisionReject   DecisionKind = "reject"
	DecisionAccept   DecisionKind = "accept"
)

type EditorialDecision struct {
	ID                string
	ManuscriptID      string
	ManuscriptVersion int64
	ChiefEditorID     string
	Kind              DecisionKind
	Reason            string
	ReviewReportIDs   []string
	CreatedAt         time.Time
}

func NewDecision(id string, target manuscript.Manuscript, chiefEditorID string, kind DecisionKind, reason string, reports []Report, now time.Time) (EditorialDecision, error) {
	if id == "" || chiefEditorID == "" || strings.TrimSpace(reason) == "" {
		return EditorialDecision{}, shared.NewError("DECISION_INVALID", "decision identity and reason are required", shared.ErrValidation)
	}
	if target.Status != manuscript.StatusUnderFinal {
		return EditorialDecision{}, shared.NewError("FINAL_REVIEW_REQUIRED", "manuscript is not awaiting a final decision", shared.ErrInvalidState)
	}
	if kind != DecisionRevision && kind != DecisionReject && kind != DecisionAccept {
		return EditorialDecision{}, shared.NewError("DECISION_KIND_INVALID", "decision kind is not supported", shared.ErrValidation)
	}
	if len(reports) == 0 {
		return EditorialDecision{}, shared.NewError("DECISION_REPORT_REQUIRED", "at least one completed review report is required", shared.ErrValidation)
	}
	ids := make([]string, 0, len(reports))
	for _, report := range reports {
		if report.ManuscriptID != target.ID || report.ManuscriptVersion != target.ActiveVersion {
			return EditorialDecision{}, shared.NewError("DECISION_REPORT_VERSION_MISMATCH", "all reports must target the active manuscript version", shared.ErrVersionConflict)
		}
		ids = append(ids, report.ID)
	}
	return EditorialDecision{
		ID: id, ManuscriptID: target.ID, ManuscriptVersion: target.ActiveVersion,
		ChiefEditorID: chiefEditorID, Kind: kind, Reason: strings.TrimSpace(reason),
		ReviewReportIDs: ids, CreatedAt: now.UTC(),
	}, nil
}

func (d EditorialDecision) TargetStatus() manuscript.Status {
	switch d.Kind {
	case DecisionRevision:
		return manuscript.StatusRevisionNeeded
	case DecisionReject:
		return manuscript.StatusRejected
	default:
		return manuscript.StatusAccepted
	}
}
