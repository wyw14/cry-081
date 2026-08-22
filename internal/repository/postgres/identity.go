package postgres

import (
	"context"
	"encoding/json"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type UserStore struct{ db *Database }

func NewUserStore(db *Database) *UserStore { return &UserStore{db: db} }

func (s *UserStore) Create(ctx context.Context, user identity.User) error {
	roles, err := json.Marshal(user.Roles)
	if err != nil {
		return err
	}
	_, err = s.db.queries(ctx).Exec(ctx, `
		INSERT INTO users (id, email, display_name, password_hash, roles, active, created_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Email, user.DisplayName, user.PasswordHash, roles, user.Active, user.CreatedAt, user.Version)
	return translate(err)
}

func (s *UserStore) Get(ctx context.Context, id string) (identity.User, error) {
	return s.scan(s.db.queries(ctx).QueryRow(ctx, `SELECT id, email, display_name, password_hash, roles, active, created_at, version FROM users WHERE id=$1`, id))
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (identity.User, error) {
	return s.scan(s.db.queries(ctx).QueryRow(ctx, `SELECT id, email, display_name, password_hash, roles, active, created_at, version FROM users WHERE email=$1`, email))
}

type rowScanner interface{ Scan(...any) error }

func (s *UserStore) scan(row rowScanner) (identity.User, error) {
	var user identity.User
	var roles []byte
	if err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &roles, &user.Active, &user.CreatedAt, &user.Version); err != nil {
		return identity.User{}, translate(err)
	}
	if err := json.Unmarshal(roles, &user.Roles); err != nil {
		return identity.User{}, err
	}
	return user, nil
}

func (s *UserStore) Update(ctx context.Context, user identity.User, expectedVersion int64) error {
	roles, err := json.Marshal(user.Roles)
	if err != nil {
		return err
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE users SET display_name=$1, roles=$2, active=$3, version=$4 WHERE id=$5 AND version=$6`, user.DisplayName, roles, user.Active, user.Version, user.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

type TokenStore struct{ db *Database }

func NewTokenStore(db *Database) *TokenStore { return &TokenStore{db: db} }

func (s *TokenStore) Create(ctx context.Context, token identity.RefreshToken) error {
	_, err := s.db.queries(ctx).Exec(ctx, `INSERT INTO refresh_tokens (id, user_id, digest, issued_at, expires_at, revoked_at, version) VALUES ($1,$2,$3,$4,$5,$6,$7)`, token.ID, token.UserID, token.Digest, token.IssuedAt, token.ExpiresAt, token.RevokedAt, token.Version)
	return translate(err)
}

func (s *TokenStore) GetByDigest(ctx context.Context, digest string) (identity.RefreshToken, error) {
	var token identity.RefreshToken
	err := s.db.queries(ctx).QueryRow(ctx, `SELECT id, user_id, digest, issued_at, expires_at, revoked_at, version FROM refresh_tokens WHERE digest=$1`, digest).Scan(&token.ID, &token.UserID, &token.Digest, &token.IssuedAt, &token.ExpiresAt, &token.RevokedAt, &token.Version)
	return token, translate(err)
}

func (s *TokenStore) Update(ctx context.Context, token identity.RefreshToken, expectedVersion int64) error {
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE refresh_tokens SET revoked_at=$1, version=$2 WHERE id=$3 AND version=$4`, token.RevokedAt, token.Version, token.ID, expectedVersion)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

func (s *TokenStore) Revoke(ctx context.Context, request application.RefreshTokenRevocation) error {
	if request.Scope != application.RefreshTokenCurrentSession || request.Token.RevokedAt == nil {
		return shared.NewError("TOKEN_REVOCATION_SCOPE_INVALID", "refresh token revocation scope is invalid", shared.ErrValidation)
	}
	current, err := s.GetByDigest(ctx, request.Token.Digest)
	if err != nil {
		return err
	}
	if current.ID != request.Token.ID || current.Version != request.ExpectedVersion {
		return shared.ErrVersionConflict
	}
	tag, err := s.db.queries(ctx).Exec(ctx, `UPDATE refresh_tokens SET revoked_at=$1, version=version+1 WHERE user_id=$2 AND revoked_at IS NULL`, request.Token.RevokedAt, request.Token.UserID)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() < 1 {
		return shared.ErrVersionConflict
	}
	return nil
}
