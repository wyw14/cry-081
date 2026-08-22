package application

import (
	"context"
	"time"

	"github.com/wyw14/cry-081/internal/domain/audit"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/publication"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/platform/clock"
)

type PublicationService struct {
	manuscripts  ManuscriptRepository
	issues       IssueRepository
	articles     ArticleRepository
	users        UserRepository
	audits       AuditRepository
	transactions TransactionManager
	ids          IDGenerator
	clock        clock.Source
}

func NewPublicationService(manuscripts ManuscriptRepository, issues IssueRepository, articles ArticleRepository, users UserRepository, audits AuditRepository, transactions TransactionManager, ids IDGenerator, clock clock.Source) *PublicationService {
	return &PublicationService{manuscripts: manuscripts, issues: issues, articles: articles, users: users, audits: audits, transactions: transactions, ids: ids, clock: clock}
}

func (s *PublicationService) CreateIssue(ctx context.Context, actorID string, volume, number int, title string, releaseAt time.Time, timezone string) (publication.Issue, error) {
	actor, err := s.users.Get(ctx, actorID)
	if err != nil {
		return publication.Issue{}, err
	}
	if !actor.Can(identity.PermissionIssueManage) {
		return publication.Issue{}, shared.NewError("ISSUE_MANAGE_FORBIDDEN", "issue manager permission is required", shared.ErrForbidden)
	}
	issue, err := publication.NewIssue(s.ids.NewID(), volume, number, title, releaseAt, timezone, s.clock.UTCNow())
	if err != nil {
		return publication.Issue{}, err
	}
	if err := s.issues.Create(ctx, *issue); err != nil {
		return publication.Issue{}, err
	}
	return issue.Clone(), nil
}

func (s *PublicationService) Schedule(ctx context.Context, actorID, manuscriptID, issueID string, edition, position int, reason string) (publication.Article, error) {
	var scheduled publication.Article
	err := s.transactions.Within(ctx, func(tx context.Context) error {
		actor, err := s.users.Get(tx, actorID)
		if err != nil {
			return err
		}
		if !actor.Can(identity.PermissionPublicationWrite) {
			return shared.NewError("PUBLICATION_FORBIDDEN", "publication permission is required", shared.ErrForbidden)
		}
		target, err := s.manuscripts.Get(tx, manuscriptID)
		if err != nil {
			return err
		}
		if target.Status != manuscript.StatusAccepted {
			return shared.NewError("MANUSCRIPT_NOT_ACCEPTED", "only accepted manuscripts can be scheduled", shared.ErrInvalidState)
		}
		issue, err := s.issues.Get(tx, issueID)
		if err != nil {
			return err
		}
		issueVersion := issue.Version
		if err := issue.AddSlot(publication.Slot{ManuscriptID: target.ID, SectionID: target.SectionID, Edition: edition, Position: position}, s.clock.UTCNow()); err != nil {
			return err
		}
		version, err := target.CurrentVersion()
		if err != nil {
			return err
		}
		article, err := publication.ScheduleArticle(s.ids.NewID(), target.ID, target.ActiveVersion, issue.ID, target.SectionID, edition, version.Title, version.Abstract, version.Markdown, version.Tags)
		if err != nil {
			return err
		}
		targetVersion := target.Version
		if err := target.Transition(manuscript.StatusScheduled, actorID, reason, target.ActiveVersion, s.clock.UTCNow()); err != nil {
			return err
		}
		if err := s.issues.Update(tx, issue, issueVersion); err != nil {
			return err
		}
		if err := s.articles.Create(tx, *article); err != nil {
			return err
		}
		if err := s.manuscripts.Update(tx, target, targetVersion); err != nil {
			return err
		}
		event, err := audit.NewEvent(s.ids.NewID(), actorID, "api", "schedule", "manuscript", target.ID, reason, map[string]any{"status": manuscript.StatusAccepted}, map[string]any{"issue_id": issue.ID, "edition": edition, "position": position}, s.clock.UTCNow())
		if err != nil {
			return err
		}
		if err := s.audits.Append(tx, event); err != nil {
			return err
		}
		scheduled = article.Clone()
		return nil
	})
	return scheduled, err
}

func (s *PublicationService) PublishIssue(ctx context.Context, actorID, issueID string) error {
	return s.transactions.Within(ctx, func(tx context.Context) error {
		actor, err := s.users.Get(tx, actorID)
		if err != nil {
			return err
		}
		if !actor.Can(identity.PermissionPublicationWrite) {
			return shared.ErrForbidden
		}
		issue, err := s.issues.Get(tx, issueID)
		if err != nil {
			return err
		}
		now := s.clock.UTCNow()
		issueVersion := issue.Version
		articles, err := s.articles.ListForIssue(tx, issue.ID)
		if err != nil {
			return err
		}
		published, err := issue.ReleaseArticles(articles, now)
		if err != nil {
			return err
		}
		if err := s.issues.Update(tx, issue, issueVersion); err != nil {
			return err
		}
		for _, article := range published {
			articleVersion := article.Version
			if err := s.articles.Update(tx, article, articleVersion-1); err != nil {
				return err
			}
			target, err := s.manuscripts.Get(tx, article.ManuscriptID)
			if err != nil {
				return err
			}
			targetVersion := target.Version
			if err := target.Transition(manuscript.StatusPublished, actorID, "issue released", article.ManuscriptVersion, now); err != nil {
				return err
			}
			if err := s.manuscripts.Update(tx, target, targetVersion); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PublicationService) Withdraw(ctx context.Context, actorID, articleID, reason string) error {
	return s.transactions.Within(ctx, func(tx context.Context) error {
		actor, err := s.users.Get(tx, actorID)
		if err != nil {
			return err
		}
		if !actor.Can(identity.PermissionPublicationWrite) {
			return shared.ErrForbidden
		}
		article, err := s.articles.Get(tx, articleID)
		if err != nil {
			return err
		}
		version := article.Version
		before := article.Clone()
		if err := article.Withdraw(reason, s.clock.UTCNow()); err != nil {
			return err
		}
		if err := s.articles.Update(tx, article, version); err != nil {
			return err
		}
		target, err := s.manuscripts.Get(tx, article.ManuscriptID)
		if err != nil {
			return err
		}
		targetVersion := target.Version
		if err := target.Transition(manuscript.StatusWithdrawn, actorID, reason, article.ManuscriptVersion, s.clock.UTCNow()); err != nil {
			return err
		}
		if err := s.manuscripts.Update(tx, target, targetVersion); err != nil {
			return err
		}
		event, err := audit.NewEvent(s.ids.NewID(), actorID, "api", "withdraw", "article", article.ID, reason, before, article, s.clock.UTCNow())
		if err != nil {
			return err
		}
		return s.audits.Append(tx, event)
	})
}

func (s *PublicationService) Search(ctx context.Context, filter PublicationFilter) (PublicationPage, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	allowedSort := map[string]bool{"published_at": true, "title": true, "issue": true}
	if !allowedSort[filter.Sort] {
		filter.Sort = "published_at"
	}
	return s.articles.Search(ctx, filter)
}
