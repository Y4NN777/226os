// Package ratelimit is the shared fixed-window rate limiter over the PostgreSQL table
// rate_limit_buckets (ADR-0003, migration 00004). Buckets are keyed by an HMAC of
// scope and value, so no IP address or user id is stored.
//
// The limiter reports a database failure as an error and does not decide what to do about it:
// the caller chooses to fail closed (authentication and write limits, ADR-0003) or open
// (anonymous reads such as search).
package ratelimit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

const (
	// callTimeout bounds each database call.
	callTimeout = 2 * time.Second
	// maxWindow is the longest window a caller may ask for (ADR-0003: the longest is one hour).
	maxWindow = time.Hour
	// retention is how long an expired bucket is kept before the purge may delete it. It exceeds
	// maxWindow, so a purge can never delete a bucket that still counts.
	retention = 2 * time.Hour
	// hkdfInfo is the HKDF-Expand info string that separates this subkey from other uses of the secret.
	hkdfInfo = "226os/ratelimit/bucket"
)

// queries is the part of *dbgen.Queries the limiter uses.
type queries interface {
	HitRateBucket(ctx context.Context, arg dbgen.HitRateBucketParams) (int32, error)
	PurgeRateBuckets(ctx context.Context, windowStart pgtype.Timestamptz) error
}

// Limiter counts hits per (scope, value) in fixed windows.
type Limiter struct {
	q      queries
	key    [32]byte
	now    func() time.Time
	logger *slog.Logger
}

// New returns a limiter over db. secret is the application secret (the cursor key, validated by
// config.Load to be at least 32 bytes); the bucket key is an HKDF-Expand subkey of it, never the
// secret itself. A nil logger discards logs.
func New(db dbgen.DBTX, secret []byte, logger *slog.Logger) *Limiter {
	return newLimiter(dbgen.New(db), secret, logger)
}

func newLimiter(q queries, secret []byte, logger *slog.Logger) *Limiter {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Limiter{q: q, key: subkey(secret), now: time.Now, logger: logger}
}

// subkey is HKDF-Expand(secret, hkdfInfo, 32) with SHA-256 (RFC 5869): one block, T(1) = HMAC(secret, info || 0x01).
func subkey(secret []byte) [32]byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(hkdfInfo))
	m.Write([]byte{1})
	var out [32]byte
	copy(out[:], m.Sum(nil))
	return out
}

// bucketKey is HMAC-SHA256(subkey, scope|value). A scope never contains "|" (the table CHECK allows
// only [a-z_] and one colon), so the pair is unambiguous.
func (l *Limiter) bucketKey(scope, value string) []byte {
	m := hmac.New(sha256.New, l.key[:])
	m.Write([]byte(scope + "|" + value))
	return m.Sum(nil)
}

// Hit records one attempt for (scope, value) in the fixed window that contains now and reports
// whether it is within limit. When it is not, retryAfter is the time until the window ends.
// An error means the outcome is unknown: the caller decides between fail closed and fail open.
func (l *Limiter) Hit(ctx context.Context, scope, value string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error) {
	if limit < 1 || window <= 0 || window > maxWindow {
		return false, 0, errors.New("rate limit: invalid limit or window")
	}
	now := l.now()
	start := now.Truncate(window)
	hitCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	hits, err := l.q.HitRateBucket(hitCtx, dbgen.HitRateBucketParams{
		Scope: scope, KeyHash: l.bucketKey(scope, value), WindowStart: pgtype.Timestamptz{Time: start, Valid: true},
	})
	if err != nil {
		return false, 0, fmt.Errorf("rate limit hit: %w", err)
	}
	if hits == 1 {
		l.purge(ctx, now)
	}
	if int(hits) > limit {
		return false, start.Add(window).Sub(now), nil
	}
	return true, 0, nil
}

// purge deletes a bounded batch of long-expired buckets. Housekeeping must never fail a hit, so a
// failure is only logged (without the driver text, which can echo connection details).
func (l *Limiter) purge(ctx context.Context, now time.Time) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	if err := l.q.PurgeRateBuckets(ctx, pgtype.Timestamptz{Time: now.Add(-retention), Valid: true}); err != nil {
		l.logger.Warn("rate limit purge failed", "error_type", fmt.Sprintf("%T", err))
	}
}
