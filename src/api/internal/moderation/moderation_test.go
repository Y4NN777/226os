package moderation

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeDB struct{}

func (fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (fakeDB) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (fakeDB) Begin(context.Context) (pgx.Tx, error)                   { return nil, nil }

type allowLimiter struct{}

func (allowLimiter) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

func TestNewHandlerBuildsTheModule(t *testing.T) {
	if NewHandler(fakeDB{}, allowLimiter{}, nil) == nil {
		t.Fatal("nil handler")
	}
}
