package identity

import (
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type RefreshToken struct {
	ID        string
	UserID    string
	Digest    string
	IssuedAt  time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
	Version   int64
}

func NewRefreshToken(id, userID, digest string, issuedAt, expiresAt time.Time) (*RefreshToken, error) {
	if id == "" || userID == "" || digest == "" || !expiresAt.After(issuedAt) {
		return nil, shared.NewError("TOKEN_INVALID", "refresh token is invalid", shared.ErrValidation)
	}
	return &RefreshToken{
		ID: id, UserID: userID, Digest: digest, IssuedAt: issuedAt.UTC(), ExpiresAt: expiresAt.UTC(), Version: 1,
	}, nil
}

func (t RefreshToken) Usable(now time.Time) bool {
	return t.RevokedAt == nil && now.UTC().Before(t.ExpiresAt)
}

func (t *RefreshToken) Revoke(now time.Time) error {
	if t.RevokedAt != nil {
		return shared.NewError("TOKEN_ALREADY_REVOKED", "refresh token has already been revoked", shared.ErrConflict)
	}
	when := now.UTC()
	t.RevokedAt = &when
	t.Version++
	return nil
}
