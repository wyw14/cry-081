package manuscript

import (
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type Declaration struct {
	OriginalWork   bool
	AuthorApproved bool
	EthicsCleared  bool
	Text           string
}

func (d Declaration) Complete() bool {
	return d.OriginalWork && d.AuthorApproved && d.EthicsCleared && d.Text != ""
}

type SubmissionSnapshot struct {
	ID            string
	ManuscriptID  string
	VersionNumber int64
	Title         string
	Abstract      string
	Markdown      string
	Tags          []string
	Attachments   []Attachment
	Declaration   Declaration
	ContentDigest string
	SubmittedBy   string
	SubmittedAt   time.Time
}

func NewSubmissionSnapshot(id, manuscriptID, authorID string, version DraftVersion, declaration Declaration, now time.Time) (SubmissionSnapshot, error) {
	if id == "" || manuscriptID == "" || authorID == "" || !declaration.Complete() {
		return SubmissionSnapshot{}, shared.NewError("SUBMISSION_INVALID", "submission declaration is incomplete", shared.ErrValidation)
	}
	return SubmissionSnapshot{
		ID: id, ManuscriptID: manuscriptID, VersionNumber: version.Number,
		Title: version.Title, Abstract: version.Abstract, Markdown: version.Markdown,
		Tags: append([]string(nil), version.Tags...), Attachments: append([]Attachment(nil), version.Attachments...),
		Declaration: declaration, ContentDigest: version.ContentDigest(), SubmittedBy: authorID, SubmittedAt: now.UTC(),
	}, nil
}

func (s SubmissionSnapshot) Clone() SubmissionSnapshot {
	s.Tags = append([]string(nil), s.Tags...)
	s.Attachments = append([]Attachment(nil), s.Attachments...)
	return s
}
