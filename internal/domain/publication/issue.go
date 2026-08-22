package publication

import (
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type IssueStatus string

const (
	IssuePlanning IssueStatus = "planning"
	IssueLocked   IssueStatus = "locked"
	IssueReleased IssueStatus = "released"
)

type Slot struct {
	ManuscriptID string
	SectionID    string
	Edition      int
	Position     int
}

type Issue struct {
	ID        string
	Volume    int
	Number    int
	Title     string
	ReleaseAt time.Time
	Timezone  string
	Status    IssueStatus
	Slots     []Slot
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewIssue(id string, volume, number int, title string, releaseAt time.Time, timezone string, now time.Time) (*Issue, error) {
	if id == "" || volume < 1 || number < 1 || strings.TrimSpace(title) == "" || strings.TrimSpace(timezone) == "" {
		return nil, shared.NewError("ISSUE_INVALID", "issue fields are incomplete", shared.ErrValidation)
	}
	if !releaseAt.After(now) {
		return nil, shared.NewError("ISSUE_DATE_INVALID", "release date must be in the future", shared.ErrValidation)
	}
	now = now.UTC()
	return &Issue{ID: id, Volume: volume, Number: number, Title: strings.TrimSpace(title), ReleaseAt: releaseAt.UTC(), Timezone: timezone, Status: IssuePlanning, Version: 1, CreatedAt: now, UpdatedAt: now}, nil
}

func (i *Issue) AddSlot(slot Slot, now time.Time) error {
	if i.Status != IssuePlanning {
		return shared.NewError("ISSUE_LOCKED", "issue schedule is locked", shared.ErrInvalidState)
	}
	if slot.ManuscriptID == "" || slot.SectionID == "" || slot.Edition < 1 || slot.Position < 1 {
		return shared.NewError("ISSUE_SLOT_INVALID", "issue slot fields are incomplete", shared.ErrValidation)
	}
	for _, current := range i.Slots {
		if current.ManuscriptID == slot.ManuscriptID {
			return shared.NewError("MANUSCRIPT_ALREADY_SCHEDULED", "manuscript already has a slot", shared.ErrConflict)
		}
		if current.Edition == slot.Edition && current.Position == slot.Position {
			return shared.NewError("ISSUE_POSITION_OCCUPIED", "issue position is already occupied", shared.ErrConflict)
		}
	}
	i.Slots = append(i.Slots, slot)
	i.Version++
	i.UpdatedAt = now.UTC()
	return nil
}

func (i *Issue) Lock(now time.Time) error {
	if i.Status != IssuePlanning || len(i.Slots) == 0 {
		return shared.NewError("ISSUE_NOT_LOCKABLE", "a planned issue with content is required", shared.ErrInvalidState)
	}
	i.Status = IssueLocked
	i.Version++
	i.UpdatedAt = now.UTC()
	return nil
}

func (i *Issue) Release(now time.Time) error {
	if i.Status != IssueLocked {
		return shared.NewError("ISSUE_NOT_LOCKED", "issue must be locked before release", shared.ErrInvalidState)
	}
	i.Status = IssueReleased
	i.Version++
	i.UpdatedAt = now.UTC()
	return nil
}

// ReleaseArticles validates the complete issue layout before making any article public.
func (i *Issue) ReleaseArticles(articles []Article, now time.Time) ([]Article, error) {
	if len(i.Slots) == 0 || len(articles) != len(i.Slots) {
		return nil, shared.NewError("ISSUE_CONTENT_INCOMPLETE", "every issue slot must have one scheduled article", shared.ErrInvalidState)
	}
	byManuscript := make(map[string]Article, len(articles))
	for _, article := range articles {
		if article.IssueID != i.ID {
			return nil, shared.NewError("ARTICLE_ISSUE_MISMATCH", "article belongs to another issue", shared.ErrConflict)
		}
		byManuscript[article.ManuscriptID] = article.Clone()
	}
	published := make([]Article, 0, len(i.Slots))
	for _, slot := range i.Slots {
		article, found := byManuscript[slot.ManuscriptID]
		if !found || article.SectionID != slot.SectionID || article.Edition != slot.Edition {
			return nil, shared.NewError("ISSUE_LAYOUT_MISMATCH", "scheduled article does not match its issue slot", shared.ErrConflict)
		}
		if err := article.Publish(now); err != nil {
			return nil, err
		}
		published = append(published, article)
	}
	if i.Status == IssuePlanning {
		if err := i.Lock(now); err != nil {
			return nil, err
		}
	}
	if err := i.Release(now); err != nil {
		return nil, err
	}
	return published, nil
}

func (i Issue) Clone() Issue {
	i.Slots = append([]Slot(nil), i.Slots...)
	return i
}
