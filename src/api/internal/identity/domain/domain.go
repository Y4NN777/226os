// Package domain holds the identity model: accounts, credentials, sessions
// and single-use tokens. It has no database or HTTP types (ADR-0003).
package domain

import (
	"errors"
	"time"
)

// Sentinel errors mapped to problem codes by the service/transport edge.
var (
	// ErrInvalidCredentials covers unknown email, wrong password and
	// suspended/deleted accounts alike: sign-in never reveals which.
	ErrInvalidCredentials = errors.New("identity: invalid credentials")
	// ErrEmailTaken means the normalized email or username is already used.
	// Sign-up never surfaces it (generic 202); it exists for internal callers.
	ErrEmailTaken = errors.New("identity: email or username taken")
	// ErrUsernameTaken is ErrEmailTaken for the username alone.
	ErrUsernameTaken = errors.New("identity: username taken")
	// ErrTokenInvalid covers unknown, expired, consumed or mismatched tokens.
	ErrTokenInvalid = errors.New("identity: token invalid")
	// ErrUnauthenticated means no valid session.
	ErrUnauthenticated = errors.New("identity: unauthenticated")
	// ErrRateLimited means a fixed-window bucket is exhausted.
	ErrRateLimited = errors.New("identity: rate limited")
	// ErrNotFound means the account, session or token row does not exist.
	ErrNotFound = errors.New("identity: not found")
)

// Account is the user row plus its credential state.
type Account struct {
	ID            string
	Email         string
	EmailVerified bool
	Role          string
	Status        string
	Username      string
	DisplayName   string
	PasswordHash  string // PHC string; never leaves the repository boundary
	CreatedAt     time.Time
}

// Session is one opaque server-side session. Only token hashes are stored.
type Session struct {
	ID                string
	UserID            string
	TokenHash         [32]byte
	CSRFHash          [32]byte
	CreatedAt         time.Time
	LastSeenAt        time.Time
	AuthenticatedAt   time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         *time.Time
	RevokeReason      string
}

// Active reports whether the session may authenticate right now.
func (s Session) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.IdleExpiresAt) && now.Before(s.AbsoluteExpiresAt)
}

// TokenKind distinguishes verification from reset tokens; both are 32 random
// bytes, single-use, stored as SHA-256 hashes (ADR-0003).
type TokenKind int

const (
	TokenVerifyEmail TokenKind = iota
	TokenPasswordReset
)

// IssuedToken is a plaintext token returned once to the caller (emailed or
// set as a cookie). Only its hash is persisted.
type IssuedToken struct {
	Plaintext string
	Hash      [32]byte
}
