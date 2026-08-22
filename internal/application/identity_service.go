package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/platform/clock"
)

type IdentityService struct {
	users      UserRepository
	tokens     TokenRepository
	hasher     PasswordHasher
	issuer     TokenIssuer
	ids        IDGenerator
	clock      clock.Source
	refreshTTL time.Duration
}

func NewIdentityService(users UserRepository, tokens TokenRepository, hasher PasswordHasher, issuer TokenIssuer, ids IDGenerator, clock clock.Source, refreshTTL time.Duration) *IdentityService {
	return &IdentityService{users: users, tokens: tokens, hasher: hasher, issuer: issuer, ids: ids, clock: clock, refreshTTL: refreshTTL}
}

type RegisterUserInput struct {
	Email       string
	DisplayName string
	Password    string
	Roles       []identity.Role
}

func (s *IdentityService) Register(ctx context.Context, input RegisterUserInput) (identity.User, error) {
	if err := ctx.Err(); err != nil {
		return identity.User{}, err
	}
	if len(input.Password) < 10 {
		return identity.User{}, shared.ValidationError(shared.FieldViolation{Field: "password", Message: "must contain at least ten characters"})
	}
	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return identity.User{}, shared.NewError("PASSWORD_HASH_FAILED", "password could not be secured", err)
	}
	user, err := identity.NewUser(s.ids.NewID(), input.Email, input.DisplayName, input.Roles, hash, s.clock.UTCNow())
	if err != nil {
		return identity.User{}, err
	}
	if err := s.users.Create(ctx, *user); err != nil {
		return identity.User{}, err
	}
	return user.Clone(), nil
}

type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         identity.User
}

func (s *IdentityService) Login(ctx context.Context, email, password string) (Session, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || !user.Active || s.hasher.Compare(user.PasswordHash, password) != nil {
		return Session{}, shared.NewError("CREDENTIALS_INVALID", "email or password is invalid", shared.ErrUnauthenticated)
	}
	access, err := s.issuer.AccessToken(user)
	if err != nil {
		return Session{}, shared.NewError("ACCESS_TOKEN_FAILED", "access token could not be issued", err)
	}
	secret, digest, err := s.issuer.RefreshSecret()
	if err != nil {
		return Session{}, shared.NewError("REFRESH_TOKEN_FAILED", "refresh token could not be issued", err)
	}
	now := s.clock.UTCNow()
	token, err := identity.NewRefreshToken(s.ids.NewID(), user.ID, digest, now, now.Add(s.refreshTTL))
	if err != nil {
		return Session{}, err
	}
	if err := s.tokens.Create(ctx, *token); err != nil {
		return Session{}, err
	}
	return Session{AccessToken: access, RefreshToken: secret, ExpiresAt: token.ExpiresAt, User: user.Clone()}, nil
}

func (s *IdentityService) RevokeRefreshToken(ctx context.Context, secret string) error {
	digestBytes := sha256.Sum256([]byte(secret))
	digest := hex.EncodeToString(digestBytes[:])
	token, err := s.tokens.GetByDigest(ctx, digest)
	if err != nil {
		return err
	}
	version := token.Version
	if err := token.Revoke(s.clock.UTCNow()); err != nil {
		return err
	}
	return s.tokens.Revoke(ctx, RefreshTokenRevocation{Token: token, ExpectedVersion: version, Scope: RefreshTokenCurrentSession})
}
