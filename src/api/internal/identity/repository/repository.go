// Package repository adapts the sqlc identity queries to the service Store
// port (ADR-0003). It maps driver rows to domain types and SQLSTATE
// constraint violations to domain sentinels; it never sees plaintext tokens.
package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

const queryTimeout = 2 * time.Second

// Repository is the sqlc adapter for the identity Store port.
type Repository struct {
	q *dbgen.Queries
}

// New returns a repository over db (a *pgxpool.Pool or *pgx.Conn in tests).
func New(db dbgen.DBTX) *Repository { return &Repository{q: dbgen.New(db)} }

var _ service.Store = (*Repository)(nil)

// ---------- accounts ----------

// CreateAccount writes user + credential + profile in one statement: the CTE
// is atomic, so no partial account can exist.
func (r *Repository) CreateAccount(ctx context.Context, in service.NewAccount) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	user, err := r.q.CreateAccount(ctx, dbgen.CreateAccountParams{
		Email: in.Email, PasswordHash: in.PasswordHash,
		Username: in.Username, DisplayName: in.DisplayName,
	})
	if err != nil {
		return domain.Account{}, mapWriteErr(err, "identity create account")
	}
	return accountFromRow(user.ID, user.Email, user.EmailVerifiedAt, user.Role, user.Status,
		user.CreatedAt, in.PasswordHash, in.Username, in.DisplayName), nil
}

func (r *Repository) AccountByEmail(ctx context.Context, email string) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	row, err := r.q.GetAccountByEmail(ctx, email)
	if err != nil {
		return domain.Account{}, accountErr(err, "identity account by email")
	}
	return accountFromRow(row.ID, row.Email, row.EmailVerifiedAt, row.Role, row.Status,
		row.CreatedAt, row.PasswordHash, row.Username, row.DisplayName), nil
}

func (r *Repository) AccountByID(ctx context.Context, id string) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	uid, err := parseUUID(id)
	if err != nil {
		return domain.Account{}, domain.ErrNotFound
	}
	row, err := r.q.GetAccountByID(ctx, uid)
	if err != nil {
		return domain.Account{}, accountErr(err, "identity account by id")
	}
	return accountFromRow(row.ID, row.Email, row.EmailVerifiedAt, row.Role, row.Status,
		row.CreatedAt, row.PasswordHash, row.Username, row.DisplayName), nil
}

func accountFromRow(id pgtype.UUID, email string, verified pgtype.Timestamptz, role, status string,
	created pgtype.Timestamptz, phc, username, displayName string) domain.Account {
	return domain.Account{
		ID: uuidString(id), Email: email, EmailVerified: verified.Valid,
		Role: role, Status: status, Username: username, DisplayName: displayName,
		PasswordHash: phc, CreatedAt: created.Time,
	}
}

func accountErr(err error, what string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", what, err)
}

// ---------- sessions ----------

func (r *Repository) CreateSession(ctx context.Context, s domain.Session) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	id, err := parseUUID(s.ID)
	if err != nil {
		return "", fmt.Errorf("identity create session id: %w", err)
	}
	uid, err := parseUUID(s.UserID)
	if err != nil {
		return "", fmt.Errorf("identity create session user: %w", err)
	}
	err = r.q.CreateSession(ctx, dbgen.CreateSessionParams{
		ID: id, UserID: uid,
		TokenHash: s.TokenHash[:], CsrfTokenHash: s.CSRFHash[:],
		IdleExpiresAt: ts(s.IdleExpiresAt), AbsoluteExpiresAt: ts(s.AbsoluteExpiresAt),
	})
	if err != nil {
		return "", fmt.Errorf("identity create session: %w", err)
	}
	return s.ID, nil
}

func (r *Repository) SessionByTokenHash(ctx context.Context, hash [32]byte) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	row, err := r.q.GetSessionByTokenHash(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, domain.ErrNotFound
		}
		return domain.Session{}, fmt.Errorf("identity session lookup: %w", err)
	}
	return sessionFromRow(row), nil
}

func sessionFromRow(row dbgen.GetSessionByTokenHashRow) domain.Session {
	s := domain.Session{
		ID: uuidString(row.ID), UserID: uuidString(row.UserID),
		CreatedAt: row.CreatedAt.Time, LastSeenAt: row.LastSeenAt.Time,
		AuthenticatedAt: row.AuthenticatedAt.Time,
		IdleExpiresAt:   row.IdleExpiresAt.Time, AbsoluteExpiresAt: row.AbsoluteExpiresAt.Time,
	}
	copy(s.TokenHash[:], row.TokenHash)
	copy(s.CSRFHash[:], row.CsrfTokenHash)
	if row.RevokedAt.Valid {
		revoked := row.RevokedAt.Time
		s.RevokedAt = &revoked
		s.RevokeReason = row.RevokedReason.String
	}
	return s
}

func (r *Repository) RevokeSession(ctx context.Context, id, reason string, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	sid, err := parseUUID(id)
	if err != nil {
		return domain.ErrNotFound
	}
	if err := r.q.RevokeSession(ctx, dbgen.RevokeSessionParams{
		ID: sid, RevokedReason: pgtype.Text{String: reason, Valid: true}, RevokedAt: ts(now),
	}); err != nil {
		return fmt.Errorf("identity revoke session: %w", err)
	}
	return nil
}

func (r *Repository) RevokeUserSessions(ctx context.Context, userID, reason string, now time.Time, keepID string) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.ErrNotFound
	}
	var keep pgtype.UUID
	if keepID != "" {
		keep, err = parseUUID(keepID)
		if err != nil {
			return domain.ErrNotFound
		}
	}
	if err := r.q.RevokeUserSessions(ctx, dbgen.RevokeUserSessionsParams{
		UserID: uid, RevokedReason: pgtype.Text{String: reason, Valid: true},
		RevokedAt: ts(now), Column4: keep,
	}); err != nil {
		return fmt.Errorf("identity revoke user sessions: %w", err)
	}
	return nil
}

func (r *Repository) TouchSession(ctx context.Context, id string, lastSeenAt, idleExpiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	sid, err := parseUUID(id)
	if err != nil {
		return domain.ErrNotFound
	}
	if err := r.q.TouchSession(ctx, dbgen.TouchSessionParams{
		ID: sid, LastSeenAt: ts(lastSeenAt), IdleExpiresAt: ts(idleExpiresAt),
	}); err != nil {
		return fmt.Errorf("identity touch session: %w", err)
	}
	return nil
}

// ---------- tokens ----------

func (r *Repository) CreateEmailVerification(ctx context.Context, userID, email string, hash [32]byte, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.ErrNotFound
	}
	if err := r.q.CreateEmailVerification(ctx, dbgen.CreateEmailVerificationParams{
		UserID: uid, Email: email, TokenHash: hash[:], ExpiresAt: ts(expiresAt),
	}); err != nil {
		return fmt.Errorf("identity create email verification: %w", err)
	}
	return nil
}

// VerifyEmail consumes the token and marks the account verified in one
// atomic statement.
func (r *Repository) VerifyEmail(ctx context.Context, hash [32]byte, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if _, err := r.q.VerifyEmail(ctx, dbgen.VerifyEmailParams{
		TokenHash: hash[:], ConsumedAt: ts(now),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTokenInvalid
		}
		return fmt.Errorf("identity verify email: %w", err)
	}
	return nil
}

func (r *Repository) CreatePasswordReset(ctx context.Context, userID string, hash [32]byte, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.ErrNotFound
	}
	if err := r.q.CreatePasswordReset(ctx, dbgen.CreatePasswordResetParams{
		UserID: uid, TokenHash: hash[:], ExpiresAt: ts(expiresAt),
	}); err != nil {
		return fmt.Errorf("identity create password reset: %w", err)
	}
	return nil
}

// ResetPassword consumes the token, stores the new PHC hash and revokes
// every session of the account in one atomic statement.
func (r *Repository) ResetPassword(ctx context.Context, hash [32]byte, phc string, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if _, err := r.q.ResetPassword(ctx, dbgen.ResetPasswordParams{
		TokenHash: hash[:], ConsumedAt: ts(now), PasswordHash: phc,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTokenInvalid
		}
		return fmt.Errorf("identity reset password: %w", err)
	}
	return nil
}

// ---------- rate limits ----------

// Count reports the hits already recorded in the window that contains now,
// without recording one. A window with no row has zero hits.
func (r *Repository) Count(ctx context.Context, scope string, key [32]byte, window time.Duration, now time.Time) (int, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	windowStart := now.Truncate(window)
	hits, err := r.q.CountRateLimit(ctx, dbgen.CountRateLimitParams{
		Scope: scope, KeyHash: key[:], WindowStart: ts(windowStart),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, windowStart.Add(window).Sub(now), nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("identity rate limit count: %w", err)
	}
	return int(hits), windowStart.Add(window).Sub(now), nil
}

// purgeAfter is how long a finished window is kept before it is purged.
const purgeAfter = 2 * time.Hour

// Hit records one attempt in the fixed window that contains now. windowStart
// is aligned to the window so concurrent processes agree on bucket identity.
func (r *Repository) Hit(ctx context.Context, scope string, key [32]byte, limit int, window time.Duration, now time.Time) (bool, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	windowStart := now.Truncate(window)
	hits, err := r.q.HitRateLimit(ctx, dbgen.HitRateLimitParams{
		Scope: scope, KeyHash: key[:], WindowStart: ts(windowStart),
	})
	if err != nil {
		return false, 0, fmt.Errorf("identity rate limit hit: %w", err)
	}
	// The table grows only through new buckets, so old windows are purged when
	// one opens (bounded DELETE; every window is at most an hour).
	if hits == 1 {
		// Housekeeping must not deny a request whose hit is already counted:
		// a failed purge is logged (no key or address in it) and retried by the
		// next new bucket.
		if err := r.q.PurgeRateLimitBuckets(ctx, ts(now.Add(-purgeAfter))); err != nil {
			slog.WarnContext(ctx, "identity rate limit purge failed", slog.String("error", err.Error()))
		}
	}
	if int(hits) > limit {
		return false, windowStart.Add(window).Sub(now), nil
	}
	return true, 0, nil
}

// ---------- helpers ----------

// mapWriteErr translates SQLSTATE uniqueness violations into domain sentinels
// so SignUp can answer the generic 202 instead of leaking which field clashed.
func mapWriteErr(err error, what string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "users_email_uq":
			return domain.ErrEmailTaken
		case "profiles_username_uq":
			return domain.ErrUsernameTaken
		}
	}
	return fmt.Errorf("%s: %w", what, err)
}

func parseUUID(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

func uuidString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
