package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/search/domain"
)

const hitID = "00000000-0000-7000-8000-0000000000a1"

type fakeDB struct {
	rows     [][]any
	err      error
	sql      string
	args     []any
	deadline bool
}

func (f *fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("not used")
}

func (f *fakeDB) QueryRow(context.Context, string, ...any) pgx.Row { return nil }

func (f *fakeDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	f.sql, f.args = sql, args
	_, f.deadline = ctx.Deadline()
	if f.err != nil {
		return nil, f.err
	}
	return &fakeRows{rows: f.rows}, nil
}

type fakeRows struct {
	rows [][]any
	i    int
}

func (r *fakeRows) Close()                                       {}
func (r *fakeRows) Err() error                                   { return nil }
func (r *fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeRows) Values() ([]any, error)                       { return nil, nil }
func (r *fakeRows) RawValues() [][]byte                          { return nil }
func (r *fakeRows) Conn() *pgx.Conn                              { return nil }
func (r *fakeRows) TypeMap() *pgtype.Map                         { return nil }
func (r *fakeRows) Next() bool {
	if r.i < len(r.rows) {
		r.i++
		return true
	}
	return false
}

func (r *fakeRows) Scan(dest ...any) error {
	vals := r.rows[r.i-1]
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(vals[i]))
	}
	return nil
}

func uid(s string) pgtype.UUID { return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true} }

func TestSearchMapsEveryGroupToPlainTextHits(t *testing.T) {
	long := strings.Repeat("mot ", 100)
	for _, g := range domain.Groups() {
		t.Run(string(g), func(t *testing.T) {
			db := &fakeDB{rows: [][]any{
				{uid(hitID), "Un titre", "**Gras** <b>html</b> [lien](https://x.example)", "fr"},
				{uid(hitID), "Long", long, "en"},
			}}
			got, err := New(db).Search(t.Context(), g, domain.Query{Text: "mot", Country: "BF", Lang: "fr", Limit: 6})
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 2 || got[0].Type != g || got[0].ID != hitID || got[0].Title != "Un titre" || got[0].Lang != "fr" {
				t.Fatalf("hits = %+v", got)
			}
			if got[0].Snippet != "Gras bhtml/b lien" {
				t.Fatalf("snippet = %q", got[0].Snippet)
			}
			if n := len([]rune(got[1].Snippet)); n > domain.MaxSnippetRunes || got[1].Lang != "en" {
				t.Fatalf("long snippet is %d runes", n)
			}
			if !db.deadline {
				t.Fatal("the query must run under a context deadline")
			}
		})
	}
}

func TestSearchBindsTheLimitAndOptionalFilters(t *testing.T) {
	db := &fakeDB{}
	if _, err := New(db).Search(t.Context(), domain.Questions, domain.Query{Text: "mot", Limit: 21}); err != nil {
		t.Fatal(err)
	}
	want := []any{int32(21), "mot", pgtype.Text{}, pgtype.Text{}} // sqlc binds the outer LIMIT first when the query has a CTE
	if !reflect.DeepEqual(db.args, want) {
		t.Fatalf("args = %#v, want %#v", db.args, want)
	}
	if _, err := New(db).Search(t.Context(), domain.Learn, domain.Query{Text: "mot", Country: "BF", Lang: "en", Limit: 3}); err != nil {
		t.Fatal(err)
	}
	// Learn has no country column: the country filter is not bound for that group.
	want = []any{int32(3), "mot", pgtype.Text{String: "en", Valid: true}}
	if !reflect.DeepEqual(db.args, want) {
		t.Fatalf("learn args = %#v, want %#v", db.args, want)
	}
}

func TestSearchReturnsNoHitsAsAnEmptyList(t *testing.T) {
	got, err := New(&fakeDB{}).Search(t.Context(), domain.Projects, domain.Query{Text: "mot", Limit: 6})
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("got %#v err %v", got, err)
	}
}

func TestSearchErrors(t *testing.T) {
	boom := errors.New("db down")
	for _, g := range domain.Groups() {
		_, err := New(&fakeDB{err: boom}).Search(t.Context(), g, domain.Query{Text: "mot", Limit: 6})
		if !errors.Is(err, boom) || !strings.Contains(err.Error(), "search "+string(g)) {
			t.Fatalf("%s: err = %v", g, err)
		}
	}
	// A query that outlives its 2 s deadline keeps the cause for errors.Is; the transport redacts it (#118).
	if _, err := New(&fakeDB{err: context.DeadlineExceeded}).Search(t.Context(), domain.Questions, domain.Query{Text: "mot", Limit: 6}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout err = %v", err)
	}
	if _, err := New(&fakeDB{}).Search(t.Context(), domain.Group("answers"), domain.Query{Limit: 6}); err == nil {
		t.Fatal("an unknown group must be an error")
	}
}
