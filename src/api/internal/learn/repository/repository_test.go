package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/learn/domain"
)

const (
	artID    = "00000000-0000-7000-8000-0000000000a1"
	userID   = "00000000-0000-7000-8000-0000000000b1"
	seriesID = "00000000-0000-7000-8000-0000000000c1"
	lessonID = "00000000-0000-7000-8000-0000000000c2"
	qID      = "00000000-0000-7000-8000-0000000000d1"
)

var errDB = errors.New("db down")

func uid(s string) pgtype.UUID { return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true} }

func ts() pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC), Valid: true}
}

// step is one scripted database call: Query/QueryRow return rows, Exec reports affected rows.
type step struct {
	rows     [][]any
	affected int64
	err      error
}

type fakeDB struct {
	t         *testing.T
	steps     []step
	n         int
	beginErr  error
	commitErr error
	committed bool
}

func (f *fakeDB) next(kind string) step {
	f.t.Helper()
	if f.n >= len(f.steps) {
		f.t.Fatalf("unexpected %s (call %d)", kind, f.n+1)
	}
	s := f.steps[f.n]
	f.n++
	return s
}

func (f *fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	s := f.next("exec")
	return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", s.affected)), s.err
}

func (f *fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	s := f.next("query")
	if s.err != nil {
		return nil, s.err
	}
	return &fakeRows{rows: s.rows}, nil
}

func (f *fakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	s := f.next("query row")
	if s.err != nil {
		return fakeRow{err: s.err}
	}
	if len(s.rows) == 0 {
		return fakeRow{err: pgx.ErrNoRows}
	}
	return fakeRow{vals: s.rows[0]}
}

func (f *fakeDB) Begin(context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return fakeTx{db: f}, nil
}

type fakeTx struct {
	pgx.Tx
	db *fakeDB
}

func (t fakeTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.db.Exec(ctx, sql, args...)
}

func (t fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.db.Query(ctx, sql, args...)
}

func (t fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.db.QueryRow(ctx, sql, args...)
}

func (t fakeTx) Commit(context.Context) error {
	if t.db.commitErr == nil {
		t.db.committed = true
	}
	return t.db.commitErr
}

func (t fakeTx) Rollback(context.Context) error { return nil }

type fakeRow struct {
	vals []any
	err  error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return scan(dest, r.vals)
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
func (r *fakeRows) Scan(dest ...any) error { return scan(dest, r.rows[r.i-1]) }

func scan(dest, vals []any) error {
	if len(dest) != len(vals) {
		return fmt.Errorf("scan width %d, want %d", len(vals), len(dest))
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(vals[i]))
	}
	return nil
}

func db(t *testing.T, steps ...step) *fakeDB { return &fakeDB{t: t, steps: steps} }

func rows(r ...[]any) step { return step{rows: r} }

func fail(err error) step { return step{err: err} }

func listRow() []any {
	return []any{uid(artID), "tutorial", "Intégrer Mobile Money", "", "fr", "beginner", ts(), uid(userID), "alice", "Alice", "BF"}
}

func getRow(published pgtype.Timestamptz) []any {
	return []any{uid(artID), "article", "Titre d'origine", "Résumé", "# Corps\n", "fr", "beginner", "published",
		published, ts(), uid(userID), "alice", "Alice", ""}
}

func seriesRow(withCount bool) []any {
	r := []any{uid(seriesID), "Formation IA", "Résumé", "fr", "beginner", ts(), uid(userID), "alice", "Alice", ""}
	if withCount {
		r = append(r, int32(3))
	}
	return r
}

func TestListArticlesAttachesTags(t *testing.T) {
	at := ts().Time
	got, err := New(db(t, rows(listRow()), rows([]any{uid(artID), "api"}, []any{uid(artID), "mobile-money"}))).
		ListArticles(t.Context(), domain.ArticleQuery{AfterPublished: &at, AfterID: artID, Limit: 3})
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	a := got[0]
	if a.ID != artID || a.Author.Username != "alice" || a.Author.Country != "BF" || len(a.Tags) != 2 || a.Tags[0] != "api" {
		t.Fatalf("summary %+v", a)
	}
}

func TestListArticlesEmptyAndErrors(t *testing.T) {
	got, err := New(db(t, rows())).ListArticles(t.Context(), domain.ArticleQuery{
		Q: "ussd", Kind: "article", Topic: "mobile", Level: "advanced", Lang: "en", Limit: 21,
	})
	if err != nil || len(got) != 0 {
		t.Fatalf("empty %+v, %v", got, err)
	}
	if _, err := New(db(t)).ListArticles(t.Context(), domain.ArticleQuery{AfterID: "bad"}); err == nil {
		t.Fatal("bad cursor id accepted")
	}
	if _, err := New(db(t, fail(errDB))).ListArticles(t.Context(), domain.ArticleQuery{}); !errors.Is(err, errDB) {
		t.Fatalf("query err %v", err)
	}
	if _, err := New(db(t, rows(listRow()), fail(errDB))).ListArticles(t.Context(), domain.ArticleQuery{}); !errors.Is(err, errDB) {
		t.Fatalf("tags err %v", err)
	}
}

func TestGetArticleReturnsTheOriginalTextAndLinks(t *testing.T) {
	got, err := New(db(t, rows(getRow(ts())), rows([]any{uid(artID), "ia"}),
		rows([]any{"question", uid(qID), "Comment payer ?", "fr"}))).GetArticle(t.Context(), artID, userID)
	if err != nil {
		t.Fatal(err)
	}
	if got.BodyMD != "# Corps\n" || got.Title != "Titre d'origine" || got.PublishedAt == nil || got.Tags[0] != "ia" {
		t.Fatalf("article %+v", got)
	}
	if len(got.Related) != 1 || got.Related[0] != (domain.Link{Type: "question", ID: qID, Title: "Comment payer ?", Lang: "fr"}) {
		t.Fatalf("related %+v", got.Related)
	}
}

func TestGetArticleDraftHasNoPublishedAt(t *testing.T) {
	got, err := New(db(t, rows(getRow(pgtype.Timestamptz{})), rows(), rows())).GetArticle(t.Context(), artID, "")
	if err != nil || got.PublishedAt != nil || got.Tags == nil || len(got.Related) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestGetArticleErrors(t *testing.T) {
	if _, err := New(db(t)).GetArticle(t.Context(), "nope", ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id %v", err)
	}
	if _, err := New(db(t)).GetArticle(t.Context(), artID, "nope"); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad viewer %v", err)
	}
	if _, err := New(db(t, rows())).GetArticle(t.Context(), artID, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows %v", err)
	}
	for name, steps := range map[string][]step{
		"get":     {fail(errDB)},
		"tags":    {rows(getRow(ts())), fail(errDB)},
		"related": {rows(getRow(ts())), rows(), fail(errDB)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(db(t, steps...)).GetArticle(t.Context(), artID, ""); !errors.Is(err, errDB) {
				t.Fatalf("err %v", err)
			}
		})
	}
}

func newArticle(links ...domain.LinkRef) domain.NewArticle {
	return domain.NewArticle{
		AuthorID: userID, Kind: "article", Title: "Titre", BodyMD: "Corps", Lang: "fr", Level: "beginner",
		Status: "published", Tags: []string{"api"}, Links: links,
	}
}

func TestCreateArticleCommitsArticleTagsAndLinks(t *testing.T) {
	f := db(t, rows([]any{uid(artID)}), step{}, step{},
		step{affected: 1}, step{affected: 1}, step{affected: 1})
	id, err := New(f).CreateArticle(t.Context(), newArticle(
		domain.LinkRef{Type: "question", ID: qID}, domain.LinkRef{Type: "resource", ID: qID}, domain.LinkRef{Type: "project", ID: qID}))
	if err != nil || id != artID || !f.committed || f.n != 6 {
		t.Fatalf("id %q err %v committed %v calls %d", id, err, f.committed, f.n)
	}
}

func TestCreateArticleErrors(t *testing.T) {
	topic := &pgconn.PgError{Code: "23503", ConstraintName: "learn_articles_topic_fk"}
	otherFK := &pgconn.PgError{Code: "23503", ConstraintName: "learn_articles_author_fk"}
	inserted := rows([]any{uid(artID)})
	tests := []struct {
		name  string
		in    domain.NewArticle
		steps []step
		begin error
		commt error
		want  error
	}{
		{"unknown topic", newArticle(), []step{fail(topic)}, nil, nil, domain.ErrUnknownTopic},
		{"other fk", newArticle(), []step{fail(otherFK)}, nil, nil, otherFK},
		{"begin", newArticle(), nil, errDB, nil, errDB},
		{"ensure tags", newArticle(), []step{inserted, fail(errDB)}, nil, nil, errDB},
		{"article tags", newArticle(), []step{inserted, {}, fail(errDB)}, nil, nil, errDB},
		{"link exec", newArticle(domain.LinkRef{Type: "question", ID: qID}), []step{inserted, {}, {}, fail(errDB)}, nil, nil, errDB},
		{"link not visible", newArticle(domain.LinkRef{Type: "project", ID: qID}), []step{inserted, {}, {}, {}}, nil, nil, domain.ErrLinkTarget},
		{"link bad id", newArticle(domain.LinkRef{Type: "project", ID: "x"}), []step{inserted, {}, {}}, nil, nil, domain.ErrLinkTarget},
		{"link series", newArticle(domain.LinkRef{Type: "series", ID: qID}), []step{inserted, {}, {}}, nil, nil, domain.ErrLinkTarget},
		{"commit", newArticle(), []step{inserted, {}, {}}, nil, errDB, errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := db(t, tc.steps...)
			f.beginErr, f.commitErr = tc.begin, tc.commt
			if _, err := New(f).CreateArticle(t.Context(), tc.in); !errors.Is(err, tc.want) {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
			if f.committed {
				t.Fatal("committed on failure")
			}
		})
	}
}

func TestCreateArticleRejectsABadAuthor(t *testing.T) {
	in := newArticle()
	in.AuthorID = "nope"
	if _, err := New(db(t)).CreateArticle(t.Context(), in); err == nil {
		t.Fatal("bad author accepted")
	}
}

func TestListSeries(t *testing.T) {
	at := ts().Time
	got, err := New(db(t, rows(seriesRow(true)))).ListSeries(t.Context(), domain.SeriesQuery{AfterPublished: &at, AfterID: seriesID, Limit: 2})
	if err != nil || len(got) != 1 || got[0].LessonCount != 3 || got[0].Author.DisplayName != "Alice" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, err := New(db(t)).ListSeries(t.Context(), domain.SeriesQuery{AfterID: "bad"}); err == nil {
		t.Fatal("bad cursor id accepted")
	}
	if _, err := New(db(t, fail(errDB))).ListSeries(t.Context(), domain.SeriesQuery{}); !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}
}

func TestGetSeriesWithLessons(t *testing.T) {
	got, err := New(db(t, rows(seriesRow(false)), rows([]any{uid(lessonID), int32(1), "Leçon 1", uid(artID)}))).GetSeries(t.Context(), seriesID)
	if err != nil || len(got.Lessons) != 1 || got.Lessons[0] != (domain.Lesson{ID: lessonID, Title: "Leçon 1", ArticleID: artID, Position: 1}) {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestGetSeriesErrors(t *testing.T) {
	if _, err := New(db(t)).GetSeries(t.Context(), "nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id %v", err)
	}
	if _, err := New(db(t, rows())).GetSeries(t.Context(), seriesID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows %v", err)
	}
	if _, err := New(db(t, fail(errDB))).GetSeries(t.Context(), seriesID); !errors.Is(err, errDB) {
		t.Fatalf("get %v", err)
	}
	if _, err := New(db(t, rows(seriesRow(false)), fail(errDB))).GetSeries(t.Context(), seriesID); !errors.Is(err, errDB) {
		t.Fatalf("lessons %v", err)
	}
}
