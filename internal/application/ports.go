package application

import (
	"context"
	"time"

	"github.com/wyw14/cry-081/internal/domain/audit"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/publication"
	"github.com/wyw14/cry-081/internal/domain/review"
)

type UserRepository interface {
	Create(context.Context, identity.User) error
	Get(context.Context, string) (identity.User, error)
	GetByEmail(context.Context, string) (identity.User, error)
	Update(context.Context, identity.User, int64) error
}

type TokenRepository interface {
	Create(context.Context, identity.RefreshToken) error
	GetByDigest(context.Context, string) (identity.RefreshToken, error)
	Update(context.Context, identity.RefreshToken, int64) error
	Revoke(context.Context, RefreshTokenRevocation) error
}

type RefreshTokenRevocationScope string

const RefreshTokenCurrentSession RefreshTokenRevocationScope = "current_session"

type RefreshTokenRevocation struct {
	Token           identity.RefreshToken
	ExpectedVersion int64
	Scope           RefreshTokenRevocationScope
}

type ManuscriptRepository interface {
	Create(context.Context, manuscript.Manuscript) error
	Get(context.Context, string) (manuscript.Manuscript, error)
	Update(context.Context, manuscript.Manuscript, int64) error
	FindByFingerprint(context.Context, string) ([]manuscript.Manuscript, error)
	List(context.Context, ManuscriptFilter) (ManuscriptPage, error)
}

type AssignmentRepository interface {
	Create(context.Context, review.Assignment) error
	Get(context.Context, string) (review.Assignment, error)
	Update(context.Context, review.Assignment, int64) error
	ListForManuscript(context.Context, string, int64) ([]review.Assignment, error)
	ListDue(context.Context, int64) ([]review.Assignment, error)
	HasActiveConflict(context.Context, string, string, time.Time) (bool, error)
}

type ReportRepository interface {
	Create(context.Context, review.Report) error
	Get(context.Context, string) (review.Report, error)
	Update(context.Context, review.Report, int64) error
	ListForManuscript(context.Context, string, int64) ([]review.Report, error)
}

type DecisionRepository interface {
	Create(context.Context, review.EditorialDecision) error
	ListForManuscript(context.Context, string) ([]review.EditorialDecision, error)
}

type IssueRepository interface {
	Create(context.Context, publication.Issue) error
	Get(context.Context, string) (publication.Issue, error)
	Update(context.Context, publication.Issue, int64) error
}

type ArticleRepository interface {
	Create(context.Context, publication.Article) error
	Get(context.Context, string) (publication.Article, error)
	Update(context.Context, publication.Article, int64) error
	ListForIssue(context.Context, string) ([]publication.Article, error)
	Search(context.Context, PublicationFilter) (PublicationPage, error)
}

type AuditRepository interface {
	Append(context.Context, audit.Event) error
	List(context.Context, AuditFilter) ([]audit.Event, error)
}

type TransactionManager interface {
	Within(context.Context, func(context.Context) error) error
}

type PasswordHasher interface {
	Hash(string) ([]byte, error)
	Compare([]byte, string) error
}

type TokenIssuer interface {
	AccessToken(identity.User) (string, error)
	RefreshSecret() (string, string, error)
}

type IDGenerator interface {
	NewID() string
}

type ManuscriptFilter struct {
	AuthorID string
	Status   manuscript.Status
	Section  string
	Sort     string
	Cursor   string
	Limit    int
}

type ManuscriptPage struct {
	Items      []manuscript.Manuscript
	NextCursor string
}

type PublicationFilter struct {
	Query     string
	SectionID string
	Tag       string
	Sort      string
	Page      int
	PageSize  int
}

type PublicationPage struct {
	Items []publication.Article
	Total int
}

type AuditFilter struct {
	Resource   string
	ResourceID string
	ActorID    string
	Limit      int
}
