package idempotency

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DB is the part of *pgxpool.Pool the PostgreSQL store needs.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PgStore is the Store over the idempotency_keys table (migration 00004).
type PgStore struct{ db DB }

// NewPgStore returns a Store over db.
func NewPgStore(db DB) *PgStore { return &PgStore{db: db} }

// maxBeginAttempts bounds the insert/select dance when rows vanish between the two statements.
const maxBeginAttempts = 3

// claimSQL inserts the key, or takes over a row that is expired, or abandoned
// by a crashed owner and identical to this request. One statement, so two
// simultaneous first requests yield exactly one insert.
const claimSQL = `
INSERT INTO idempotency_keys (user_id, idempotency_key, method, route, request_hash, state, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, 'in_progress', $6, $7)
ON CONFLICT (user_id, idempotency_key) DO UPDATE SET
  method = EXCLUDED.method, route = EXCLUDED.route, request_hash = EXCLUDED.request_hash,
  state = 'in_progress', response_status = NULL, response_body = NULL, completed_at = NULL,
  created_at = EXCLUDED.created_at, expires_at = EXCLUDED.expires_at
WHERE idempotency_keys.expires_at <= $6
   OR (idempotency_keys.state = 'in_progress' AND idempotency_keys.created_at <= $8
       AND idempotency_keys.method = EXCLUDED.method AND idempotency_keys.route = EXCLUDED.route
       AND idempotency_keys.request_hash = EXCLUDED.request_hash)
RETURNING created_at`

const selectSQL = `
SELECT method, route, request_hash, state, response_status, response_body
FROM idempotency_keys
WHERE user_id = $1 AND idempotency_key = $2 AND expires_at > $3`

// Begin implements Store.
func (s *PgStore) Begin(ctx context.Context, userID, key, method, route string, hash [32]byte, now time.Time) (Outcome, Attempt, Response, error) {
	for range maxBeginAttempts {
		var createdAt time.Time
		err := s.db.QueryRow(ctx, claimSQL, userID, key, method, route, hash[:], now, now.Add(Retention), now.Add(-AbandonAfter)).Scan(&createdAt)
		if err == nil {
			return Started, Attempt(createdAt.UnixMicro()), Response{}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, Response{}, fmt.Errorf("claim idempotency key: %w", err)
		}
		outcome, attempt, stored, err := s.existing(ctx, userID, key, method, route, hash, now)
		if errors.Is(err, pgx.ErrNoRows) {
			continue // the row vanished (abandoned, purged): claim again
		}
		return outcome, attempt, stored, err
	}
	return 0, 0, Response{}, errors.New("claim idempotency key: row keeps changing")
}

func (s *PgStore) existing(ctx context.Context, userID, key, method, route string, hash [32]byte, now time.Time) (Outcome, Attempt, Response, error) {
	var (
		gotMethod, gotRoute, state string
		gotHash                    []byte
		status                     *int16
		body                       []byte
	)
	err := s.db.QueryRow(ctx, selectSQL, userID, key, now).Scan(&gotMethod, &gotRoute, &gotHash, &state, &status, &body)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, Response{}, err
		}
		return 0, 0, Response{}, fmt.Errorf("read idempotency key: %w", err)
	}
	switch {
	case gotMethod != method || gotRoute != route || !bytes.Equal(gotHash, hash[:]):
		return Reused, 0, Response{}, nil
	case state != "completed" || status == nil:
		return InProgress, 0, Response{}, nil
	}
	return Replay, 0, decodeEnvelope(int(*status), body), nil
}

// completeSQL stores the envelope $4 only when its jsonb text form fits the column check
// (jsonb adds spaces, so an envelope within the raw cap can still exceed it), else the
// fallback envelope $6; a mismatching attempt ($7, the created_at of the claim) touches nothing.
const completeSQL = `
UPDATE idempotency_keys SET
  state = 'completed', response_status = $3, completed_at = $5,
  response_body = CASE WHEN octet_length($4::jsonb::text) <= 16384 THEN $4::jsonb ELSE $6::jsonb END
WHERE user_id = $1 AND idempotency_key = $2 AND state = 'in_progress' AND created_at = $7`

// Complete implements Store. A value PostgreSQL refuses (a \u0000 escape, invalid UTF-8) must not
// leave the key in progress: the envelope is retried body-less with the headers, then minimal
// without them; only when the minimal one fails too is the error returned.
func (s *PgStore) Complete(ctx context.Context, userID, key string, attempt Attempt, resp Response, now time.Time) error {
	tiers, err := encodeEnvelopes(resp)
	if err != nil {
		return err
	}
	at := time.UnixMicro(int64(attempt)).UTC()
	var tag pgconn.CommandTag
	for i, env := range tiers {
		// The oversize fallback ($6) is the next tier, or the envelope itself on the last one.
		fallback := env
		if i+1 < len(tiers) {
			fallback = tiers[i+1]
		}
		tag, err = s.db.Exec(ctx, completeSQL, userID, key, resp.Status, env, now, fallback, at)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("complete idempotency key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLostKey
	}
	return nil
}

// Abandon implements Store.
func (s *PgStore) Abandon(ctx context.Context, userID, key string, attempt Attempt) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM idempotency_keys WHERE user_id = $1 AND idempotency_key = $2 AND state = 'in_progress' AND created_at = $3`,
		userID, key, time.UnixMicro(int64(attempt)).UTC())
	if err != nil {
		return fmt.Errorf("abandon idempotency key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLostKey
	}
	return nil
}

// Purge implements Store.
func (s *PgStore) Purge(ctx context.Context, now time.Time) (int64, error) {
	tag, err := s.db.Exec(ctx, `DELETE FROM idempotency_keys WHERE expires_at <= $1`, now)
	if err != nil {
		return 0, fmt.Errorf("purge idempotency keys: %w", err)
	}
	return tag.RowsAffected(), nil
}
