package search

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type stubDB struct{}

func (stubDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (stubDB) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (stubDB) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }

func TestNewHandler(t *testing.T) {
	if NewHandler(stubDB{}, nil) == nil {
		t.Fatal("nil handler")
	}
}
