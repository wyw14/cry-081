package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/wyw14/cry-081/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{ Cost int }

func (h BcryptHasher) Hash(password string) ([]byte, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

func (h BcryptHasher) Compare(hash []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}

type LocalTokenIssuer struct {
	AccessTTL time.Duration
}

func (i LocalTokenIssuer) AccessToken(user identity.User) (string, error) {
	payload := fmt.Sprintf("%s:%d", user.ID, time.Now().UTC().Add(i.AccessTTL).Unix())
	return base64.RawURLEncoding.EncodeToString([]byte(payload)), nil
}

func (i LocalTokenIssuer) RefreshSecret() (string, string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(data)
	digest := sha256.Sum256([]byte(secret))
	return secret, hex.EncodeToString(digest[:]), nil
}

type SequentialIDs struct {
	prefix string
	next   atomic.Uint64
}

func NewSequentialIDs(prefix string) *SequentialIDs { return &SequentialIDs{prefix: prefix} }

func (g *SequentialIDs) NewID() string {
	return fmt.Sprintf("%s-%012d", g.prefix, g.next.Add(1))
}

type DirectTransactionManager struct{}

func (DirectTransactionManager) Within(ctx context.Context, callback func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return callback(ctx)
}
