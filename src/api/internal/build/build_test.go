package build

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Y4NN777/226os/src/api/internal/build/github"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

type stubDB struct{}

func (stubDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (stubDB) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (stubDB) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (stubDB) Begin(context.Context) (pgx.Tx, error)                   { return nil, errors.New("unused") }

func TestNewHandler(t *testing.T) {
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	if NewHandler(stubDB{}, github.New("", nil), c, nil, nil) == nil {
		t.Fatal("nil handler")
	}
}
