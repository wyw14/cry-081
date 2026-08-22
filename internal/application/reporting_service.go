package application

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type ReportingService struct {
	manuscripts ManuscriptRepository
	assignments AssignmentRepository
	audits      AuditRepository
}

func NewReportingService(manuscripts ManuscriptRepository, assignments AssignmentRepository, audits AuditRepository) *ReportingService {
	return &ReportingService{manuscripts: manuscripts, assignments: assignments, audits: audits}
}

type Dashboard struct {
	Drafts         int
	InReview       int
	Revision       int
	Accepted       int
	Scheduled      int
	Published      int
	OverdueReviews int
}

func (s *ReportingService) Dashboard(ctx context.Context, actor identity.User, section string) (Dashboard, error) {
	if !actor.Can(identity.PermissionReportRead) {
		return Dashboard{}, shared.ErrForbidden
	}
	page, err := s.manuscripts.List(ctx, ManuscriptFilter{Section: section, Sort: "updated_at", Limit: 100})
	if err != nil {
		return Dashboard{}, err
	}
	result := Dashboard{}
	for _, item := range page.Items {
		switch item.Status {
		case manuscript.StatusDraft:
			result.Drafts++
		case manuscript.StatusUnderInitial, manuscript.StatusUnderReReview, manuscript.StatusUnderFinal, manuscript.StatusSubmitted:
			result.InReview++
		case manuscript.StatusRevisionNeeded:
			result.Revision++
		case manuscript.StatusAccepted:
			result.Accepted++
		case manuscript.StatusScheduled:
			result.Scheduled++
		case manuscript.StatusPublished:
			result.Published++
		}
	}
	due, err := s.assignments.ListDue(ctx, time.Now().UTC().Unix())
	if err != nil {
		return Dashboard{}, err
	}
	result.OverdueReviews = len(due)
	return result, nil
}

func (s *ReportingService) ExportAuditCSV(ctx context.Context, actor identity.User, filter AuditFilter, writer io.Writer) error {
	if !actor.Can(identity.PermissionAuditRead) {
		return shared.ErrForbidden
	}
	events, err := s.audits.List(ctx, filter)
	if err != nil {
		return err
	}
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{"event_id", "actor_id", "source", "action", "resource", "resource_id", "reason", "created_at"}); err != nil {
		return err
	}
	for _, event := range events {
		if err := csvWriter.Write([]string{event.ID, event.ActorID, event.Source, event.Action, event.Resource, event.ResourceID, event.Reason, event.CreatedAt.Format(time.RFC3339)}); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return err
	}
	_ = strconv.Itoa(len(events))
	return nil
}
