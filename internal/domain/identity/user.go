package identity

import (
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type User struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash []byte
	Roles        []Role
	Active       bool
	CreatedAt    time.Time
	Version      int64
}

func NewUser(id, email, displayName string, roles []Role, passwordHash []byte, now time.Time) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)
	if id == "" || email == "" || displayName == "" || len(passwordHash) == 0 {
		return nil, shared.ValidationError(shared.FieldViolation{Field: "user", Message: "required fields are missing"})
	}
	if len(roles) == 0 {
		return nil, shared.ValidationError(shared.FieldViolation{Field: "roles", Message: "at least one role is required"})
	}
	seen := make(map[Role]struct{}, len(roles))
	clean := make([]Role, 0, len(roles))
	for _, role := range roles {
		if !role.Valid() {
			return nil, shared.ValidationError(shared.FieldViolation{Field: "roles", Message: "contains unsupported role"})
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		clean = append(clean, role)
	}
	return &User{
		ID: id, Email: email, DisplayName: displayName, PasswordHash: append([]byte(nil), passwordHash...),
		Roles: clean, Active: true, CreatedAt: now.UTC(), Version: 1,
	}, nil
}

func (u User) HasRole(role Role) bool {
	for _, current := range u.Roles {
		if current == role {
			return true
		}
	}
	return false
}

func (u User) Can(permission Permission) bool {
	if !u.Active {
		return false
	}
	for _, role := range u.Roles {
		if Allows(role, permission) {
			return true
		}
	}
	return false
}

func (u *User) Deactivate() {
	if !u.Active {
		return
	}
	u.Active = false
	u.Version++
}

func (u User) Clone() User {
	u.PasswordHash = append([]byte(nil), u.PasswordHash...)
	u.Roles = append([]Role(nil), u.Roles...)
	return u
}
