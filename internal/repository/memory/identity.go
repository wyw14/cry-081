package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type IdentityStore struct {
	mu             sync.RWMutex
	users          map[string]identity.User
	usersByEmail   map[string]string
	tokens         map[string]identity.RefreshToken
	tokensByDigest map[string]string
}

func NewIdentityStore() *IdentityStore {
	return &IdentityStore{users: map[string]identity.User{}, usersByEmail: map[string]string{}, tokens: map[string]identity.RefreshToken{}, tokensByDigest: map[string]string{}}
}

func (s *IdentityStore) Create(ctx context.Context, user identity.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	email := strings.ToLower(user.Email)
	if _, exists := s.users[user.ID]; exists {
		return shared.ErrConflict
	}
	if _, exists := s.usersByEmail[email]; exists {
		return shared.NewError("EMAIL_ALREADY_EXISTS", "email is already registered", shared.ErrConflict)
	}
	s.users[user.ID] = user.Clone()
	s.usersByEmail[email] = user.ID
	return nil
}

func (s *IdentityStore) Get(ctx context.Context, id string) (identity.User, error) {
	if err := ctx.Err(); err != nil {
		return identity.User{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return identity.User{}, shared.ErrNotFound
	}
	return user.Clone(), nil
}

func (s *IdentityStore) GetByEmail(ctx context.Context, email string) (identity.User, error) {
	if err := ctx.Err(); err != nil {
		return identity.User{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.usersByEmail[strings.ToLower(email)]
	if !ok {
		return identity.User{}, shared.ErrNotFound
	}
	return s.users[id].Clone(), nil
}

func (s *IdentityStore) Update(ctx context.Context, user identity.User, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.users[user.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || user.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.users[user.ID] = user.Clone()
	return nil
}

type TokenStore struct{ identity *IdentityStore }

func NewTokenStore(identityStore *IdentityStore) *TokenStore {
	return &TokenStore{identity: identityStore}
}

func (s *TokenStore) Create(ctx context.Context, token identity.RefreshToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.identity.mu.Lock()
	defer s.identity.mu.Unlock()
	if _, exists := s.identity.tokens[token.ID]; exists {
		return shared.ErrConflict
	}
	if _, exists := s.identity.tokensByDigest[token.Digest]; exists {
		return shared.ErrConflict
	}
	s.identity.tokens[token.ID] = token
	s.identity.tokensByDigest[token.Digest] = token.ID
	return nil
}

func (s *TokenStore) GetByDigest(ctx context.Context, digest string) (identity.RefreshToken, error) {
	if err := ctx.Err(); err != nil {
		return identity.RefreshToken{}, err
	}
	s.identity.mu.RLock()
	defer s.identity.mu.RUnlock()
	id, ok := s.identity.tokensByDigest[digest]
	if !ok {
		return identity.RefreshToken{}, shared.ErrNotFound
	}
	return s.identity.tokens[id], nil
}

func (s *TokenStore) Update(ctx context.Context, token identity.RefreshToken, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.identity.mu.Lock()
	defer s.identity.mu.Unlock()
	current, ok := s.identity.tokens[token.ID]
	if !ok {
		return shared.ErrNotFound
	}
	if current.Version != expectedVersion || token.Version <= expectedVersion {
		return shared.ErrVersionConflict
	}
	s.identity.tokens[token.ID] = token
	return nil
}
