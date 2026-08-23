package publication

import (
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type ArticleStatus string

const (
	ArticleScheduled ArticleStatus = "scheduled"
	ArticlePublished ArticleStatus = "published"
	ArticleWithdrawn ArticleStatus = "withdrawn"
)

type Article struct {
	ID                string
	ManuscriptID      string
	ManuscriptVersion int64
	IssueID           string
	SectionID         string
	Edition           int
	Title             string
	Abstract          string
	Body              string
	Tags              []string
	Status            ArticleStatus
	PublishedAt       *time.Time
	WithdrawnAt       *time.Time
	WithdrawReason    string
	Version           int64
}

func ScheduleArticle(id, manuscriptID string, manuscriptVersion int64, issueID, sectionID string, edition int, title, abstract, body string, tags []string) (*Article, error) {
	if id == "" || manuscriptID == "" || manuscriptVersion < 1 || issueID == "" || sectionID == "" || edition < 1 {
		return nil, shared.NewError("ARTICLE_INVALID", "article schedule is incomplete", shared.ErrValidation)
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" {
		return nil, shared.NewError("ARTICLE_CONTENT_INVALID", "article content is incomplete", shared.ErrValidation)
	}
	return &Article{ID: id, ManuscriptID: manuscriptID, ManuscriptVersion: manuscriptVersion, IssueID: issueID, SectionID: sectionID, Edition: edition, Title: title, Abstract: abstract, Body: body, Tags: append([]string(nil), tags...), Status: ArticleScheduled, Version: 1}, nil
}

func (a *Article) Publish(now time.Time) error {
	if a.Status != ArticleScheduled {
		return shared.NewError("ARTICLE_NOT_SCHEDULED", "scheduled article is required", shared.ErrInvalidState)
	}
	when := now.UTC()
	a.Status = ArticlePublished
	a.PublishedAt = &when
	a.Version++
	return nil
}

func (a *Article) Withdraw(reason string, now time.Time) error {
	if a.Status != ArticlePublished {
		return shared.NewError("ARTICLE_NOT_PUBLISHED", "only published articles can be withdrawn", shared.ErrInvalidState)
	}
	if strings.TrimSpace(reason) == "" {
		return shared.NewError("WITHDRAW_REASON_REQUIRED", "withdrawal reason is required", shared.ErrValidation)
	}
	when := now.UTC()
	a.Status = ArticleWithdrawn
	a.WithdrawnAt = &when
	a.WithdrawReason = strings.TrimSpace(reason)
	a.Version++
	return nil
}

type catalogVisibility string

const (
	catalogHidden  catalogVisibility = "hidden"
	catalogVisible catalogVisibility = "visible"
)

type catalogRule struct {
	status     ArticleStatus
	visibility catalogVisibility
	reason     string
}

var catalogRules = []catalogRule{
	{status: ArticleScheduled, visibility: catalogHidden, reason: "not released"},
	{status: ArticlePublished, visibility: catalogVisible, reason: "released"},
	{status: ArticleWithdrawn, visibility: catalogHidden, reason: "withdrawn from public catalog"},
}

func (a Article) Searchable() bool {
	decision := a.catalogDecision()
	return decision.visibility == catalogVisible
}

func (a Article) catalogDecision() catalogRule {
	for _, rule := range catalogRules {
		if rule.status == a.Status {
			return rule
		}
	}
	return catalogRule{status: a.Status, visibility: catalogHidden, reason: "unknown lifecycle state"}
}

func (a Article) PublicCatalogCopy() (Article, bool) {
	decision := a.catalogDecision()
	if decision.visibility != catalogVisible {
		return Article{}, false
	}
	copy := a.Clone()
	return copy, true
}

func (a Article) Clone() Article {
	a.Tags = append([]string(nil), a.Tags...)
	return a
}
