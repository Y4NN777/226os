package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
)

func uid(s string) pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true}
}

const (
	resID  = "00000000-0000-7000-8000-0000000000a1"
	srcID  = "00000000-0000-7000-8000-0000000000b1"
	other  = "00000000-0000-7000-8000-0000000000b2"
	projID = "00000000-0000-7000-8000-0000000000c1"
	qID    = "00000000-0000-7000-8000-0000000000c2"
)

func when() pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC), Valid: true}
}

func listVals() []any {
	return []any{
		uid(resID), "alpha", "Alpha Pay", "api", "mobile-money", "Paiement", []string{"fr"},
		pgtype.Text{String: "https://docs.example", Valid: true}, "verified", when(), when(), uid(srcID),
		"acme", "Acme", pgtype.Text{String: "https://acme.example", Valid: true}, int32(2),
	}
}

type script struct {
	t     *testing.T
	steps [][][]any
	errs  []error
	n     int
}

func (s *script) next(kind string) ([][]any, error) {
	s.t.Helper()
	if s.n >= len(s.steps) {
		s.t.Fatalf("unexpected %s", kind)
	}
	rows, err := s.steps[s.n], s.errs[s.n]
	s.n++
	return rows, err
}

func (s *script) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("not used")
}

func (s *script) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	rows, err := s.next("query")
	if err != nil {
		return nil, err
	}
	return &fakeRows{rows: rows}, nil
}

func (s *script) QueryRow(context.Context, string, ...any) pgx.Row {
	rows, err := s.next("query row")
	if err != nil {
		return fakeRow{err: err}
	}
	if len(rows) == 0 {
		return fakeRow{err: pgx.ErrNoRows}
	}
	return fakeRow{vals: rows[0]}
}

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
		return errors.New("scan width")
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(vals[i]))
	}
	return nil
}

func okScript(t *testing.T, extra ...[][]any) *script {
	t.Helper()
	steps := [][][]any{{listVals()}}
	errs := []error{nil}
	steps = append(steps, extra...)
	for range extra {
		errs = append(errs, nil)
	}
	return &script{t: t, steps: steps, errs: errs}
}

func row(vals ...any) []any { return vals }

func children() [][][]any {
	return [][][]any{
		{row(uid(resID), "BF"), row(uid("00000000-0000-7000-8000-0000000000ff"), "CI")},
		{row(uid(resID), "go", "Go SDK", pgtype.Text{String: "https://pkg.example", Valid: true}), row(uid(resID), "js", "JS", pgtype.Text{})},
		{row(uid(resID), "nope", uid(projID), "Ignored", "fr"), row(uid(resID), "project", pgtype.UUID{}, "No id", "fr"), row(uid(resID), "project", uid(projID), "", "fr"), row(uid(resID), "project", uid(projID), "A project", "fr")},
		{row(uid(resID), uid(qID), "A question", "fr")},
	}
}

func TestListAssemblesPublishedRows(t *testing.T) {
	repo := New(okScript(t, children()...))
	got, err := repo.List(t.Context(), domain.Query{Q: "paiement", Country: "BF", Limit: 21})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Alpha Pay" || got[0].VerificationStatus != "verified" {
		t.Fatalf("%+v", got)
	}
	if len(got[0].Countries) != 1 || got[0].Countries[0] != "BF" {
		t.Fatalf("countries %+v", got[0].Countries)
	}
	if len(got[0].SDKs) != 1 || got[0].SDKs[0].URL != "https://pkg.example" {
		t.Fatalf("sdks %+v", got[0].SDKs)
	}
	if len(got[0].Related) != 2 || got[0].Related[0].Type != "project" || got[0].Related[1].Type != "question" {
		t.Fatalf("related %+v", got[0].Related)
	}
	if got[0].DocsURL == nil || got[0].LastVerifiedAt == nil {
		t.Fatal("expected docs url and verified time")
	}
}

func TestListEmptySkipsChildQueries(t *testing.T) {
	repo := New(&script{t: t, steps: [][][]any{nil}, errs: []error{nil}})
	got, err := repo.List(t.Context(), domain.Query{Limit: 21})
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestListRejectsABadCursorID(t *testing.T) {
	repo := New(&script{t: t})
	_, err := repo.List(t.Context(), domain.Query{AfterID: "nope", Limit: 21})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestGetAndSources(t *testing.T) {
	checked := when()
	sources := [][]any{
		row(uid(srcID), "official_docs", "Docs", "https://docs.example", checked, "verified", uid(srcID)),
		row(uid(other), "claim", "A claim", "https://claim.example", pgtype.Timestamptz{}, "verified", uid(srcID)),
	}
	steps := [][][]any{{listVals()}}
	steps = append(steps, children()...)
	steps = append(steps, [][][]any{{listVals()}, sources}...)
	errs := make([]error, len(steps))
	repo := New(&script{t: t, steps: steps, errs: errs})
	got, err := repo.Get(t.Context(), resID)
	if err != nil || got.ID != resID || got.Provider.Slug != "acme" {
		t.Fatalf("%+v %v", got, err)
	}
	list, err := repo.Sources(t.Context(), resID)
	if err != nil || len(list) != 2 {
		t.Fatalf("%+v %v", list, err)
	}
	if list[0].Status != "verified" || list[0].VerifiedAt == nil || list[1].Status != "claimed" || list[1].VerifiedAt != nil {
		t.Fatalf("%+v", list)
	}
}

func TestMissingAndInvalidIDs(t *testing.T) {
	repo := New(&script{t: t, steps: [][][]any{nil, nil}, errs: []error{pgx.ErrNoRows, pgx.ErrNoRows}})
	if _, err := repo.Get(t.Context(), resID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("get %v", err)
	}
	if _, err := repo.Sources(t.Context(), resID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("sources %v", err)
	}
	repo = New(&script{t: t})
	if _, err := repo.Get(t.Context(), "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad get %v", err)
	}
	if _, err := repo.Sources(t.Context(), "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad sources %v", err)
	}
}

func TestListWithNoChildrenAndALongRelatedList(t *testing.T) {
	bare := listVals()
	bare[6] = []string(nil)
	bare[7] = pgtype.Text{}
	bare[9] = pgtype.Timestamptz{}
	links := make([][]any, 21)
	for i := range links {
		links[i] = row(uid(resID), "article", uid(projID), "Article", "en")
	}
	repo := New(&script{t: t, steps: [][][]any{
		{bare}, nil, nil, links, nil,
	}, errs: []error{nil, nil, nil, nil, nil}})
	got, err := repo.List(t.Context(), domain.Query{AfterName: "Alpha Pay", AfterID: resID, Limit: 21})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].DocsURL != nil || got[0].LastVerifiedAt != nil || len(got[0].Related) != 20 {
		t.Fatalf("docs %v verified %v related %d", got[0].DocsURL, got[0].LastVerifiedAt, len(got[0].Related))
	}
	if len(got[0].Countries) != 0 || len(got[0].SDKs) != 0 || len(got[0].DocLanguages) != 0 {
		t.Fatal("expected empty collections")
	}
	empty, err := New(&script{t: t, steps: [][][]any{{bare}, nil, nil, nil, nil}, errs: []error{nil, nil, nil, nil, nil}}).List(t.Context(), domain.Query{Limit: 21})
	if err != nil || len(empty[0].Related) != 0 {
		t.Fatalf("%+v %v", empty, err)
	}
}

func TestQueryErrors(t *testing.T) {
	boom := errors.New("db down")
	one := func(at int) ([][][]any, []error) {
		steps := [][][]any{{listVals()}, nil, nil, nil, nil}
		errs := []error{nil, nil, nil, nil, nil}
		steps[at] = nil
		errs[at] = boom
		return steps, errs
	}
	tests := []struct {
		name string
		at   int
		call string
	}{
		{"list", 0, "list"},
		{"countries", 1, "list"},
		{"sdks", 2, "list"},
		{"related", 3, "list"},
		{"discussions", 4, "list"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			steps, errs := one(tc.at)
			if tc.at == 0 {
				steps, errs = [][][]any{nil}, []error{boom}
			}
			_, err := New(&script{t: t, steps: steps, errs: errs}).List(t.Context(), domain.Query{Limit: 1})
			if err == nil || !strings.Contains(err.Error(), "db down") {
				t.Fatalf("err = %v", err)
			}
		})
	}
	t.Run("get", func(t *testing.T) {
		_, err := New(&script{t: t, steps: [][][]any{nil}, errs: []error{boom}}).Get(t.Context(), resID)
		if err == nil || !strings.Contains(err.Error(), "db down") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("get children", func(t *testing.T) {
		_, err := New(&script{t: t, steps: [][][]any{{listVals()}, nil}, errs: []error{nil, boom}}).Get(t.Context(), resID)
		if err == nil || !strings.Contains(err.Error(), "db down") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("sources", func(t *testing.T) {
		_, err := New(&script{t: t, steps: [][][]any{{listVals()}, nil}, errs: []error{nil, boom}}).Sources(t.Context(), resID)
		if err == nil || !strings.Contains(err.Error(), "db down") {
			t.Fatalf("err = %v", err)
		}
	})
}

// With 20 or more outgoing links, discussions must still appear: both lists are interleaved before the cap.
func TestRelatedInterleavesDiscussionsBeforeTheCap(t *testing.T) {
	links := make([][]any, 25)
	for i := range links {
		links[i] = row(uid(resID), "project", uid(projID), "Project", "en")
	}
	discussions := [][]any{row(uid(resID), uid(qID), "First question", "fr"), row(uid(resID), uid(qID), "Second question", "fr")}
	repo := New(&script{t: t, steps: [][][]any{{listVals()}, nil, nil, links, discussions}, errs: make([]error, 5)})
	got, err := repo.List(t.Context(), domain.Query{Limit: 21})
	if err != nil {
		t.Fatal(err)
	}
	related := got[0].Related
	if len(related) != 20 {
		t.Fatalf("related = %d, want the cap of 20", len(related))
	}
	var questions []string
	for i, l := range related {
		if l.Type == "question" {
			questions = append(questions, l.Title)
			if i > 3 {
				t.Fatalf("discussion at position %d; interleaving must place them first", i)
			}
		}
	}
	if len(questions) != 2 || questions[0] != "First question" || questions[1] != "Second question" {
		t.Fatalf("discussions in related = %v", questions)
	}
}

func TestInterleaveKeepsEveryLinkWhenOneSideIsLonger(t *testing.T) {
	a := []domain.Link{{ID: "a1"}, {ID: "a2"}, {ID: "a3"}}
	b := []domain.Link{{ID: "b1"}}
	got := interleave(a, b)
	want := []string{"a1", "b1", "a2", "a3"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i, l := range got {
		if l.ID != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if len(interleave(nil, b)) != 1 || len(interleave(a, nil)) != 3 || interleave(nil, nil) != nil {
		t.Fatal("empty sides")
	}
}
