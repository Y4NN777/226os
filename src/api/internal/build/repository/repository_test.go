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

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
)

const (
	projID = "00000000-0000-7000-8000-0000000000a1"
	userID = "00000000-0000-7000-8000-0000000000b1"
	qID    = "00000000-0000-7000-8000-0000000000d1"
	needID = "00000000-0000-7000-8000-0000000000e1"
)

var (
	errDB   = errors.New("db down")
	errDeny = errors.New("denied")
)

func uid(s string) pgtype.UUID { return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true} }

func ts() pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC), Valid: true}
}

// step is one scripted database call: Query/QueryRow return rows, Exec reports affected rows.
type step struct {
	rows [][]any
	err  error
	zero bool // Exec reports no affected row
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
	if s.zero {
		return pgconn.NewCommandTag("UPDATE 0"), s.err
	}
	return pgconn.NewCommandTag("UPDATE 1"), s.err
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

var (
	ok     = step{}
	none   = rows()
	yes    = rows([]any{true})
	no     = rows([]any{false})
	access = rows([]any{uid(userID), "public"})
)

func projectRow(techs []string) []any {
	return []any{uid(projID), "Repo", "https://github.com/o/r", "D", "fr", "public", uid(userID), "alice", "Alice", techs, int32(1), ts(), ts(), ts()}
}

func allow(domain.Access) error { return nil }

func newProject(links ...domain.LinkRef) domain.NewProject {
	return domain.NewProject{
		OwnerID: userID, Name: "Repo", Summary: "S", Lang: "fr", GitHubURL: "https://github.com/o/r",
		Visibility: "public", Technologies: []string{"go"}, Links: links,
	}
}

func TestURLTaken(t *testing.T) {
	if taken, err := New(db(t, yes)).URLTaken(t.Context(), "u"); err != nil || !taken {
		t.Fatalf("taken %v %v", taken, err)
	}
	if _, err := New(db(t, fail(errDB))).URLTaken(t.Context(), "u"); !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}
}

func TestCreateProject(t *testing.T) {
	links := []domain.LinkRef{{Type: "question", ID: qID}, {Type: "resource", ID: qID}, {Type: "article", ID: qID}}
	f := db(t, rows([]any{uid(projID)}), ok, ok, yes, yes, yes)
	id, err := New(f).CreateProject(t.Context(), newProject(links...))
	if err != nil || id != projID || !f.committed {
		t.Fatalf("create %q %v committed %v", id, err, f.committed)
	}
	// No technologies: no tag statements at all.
	in := newProject()
	in.Technologies = nil
	if _, err := New(db(t, rows([]any{uid(projID)}))).CreateProject(t.Context(), in); err != nil {
		t.Fatal(err)
	}

	dup := &pgconn.PgError{Code: "23505", ConstraintName: githubURLUnique}
	other := &pgconn.PgError{Code: "23505", ConstraintName: "something_else"}
	tests := []struct {
		name   string
		f      *fakeDB
		in     domain.NewProject
		want   error
		begin  error
		commit error
	}{
		{name: "bad owner", f: db(t), in: domain.NewProject{OwnerID: "x"}, want: nil},
		{name: "begin", f: db(t), in: newProject(), begin: errDB, want: errDB},
		{name: "duplicate url", f: db(t, fail(dup)), in: newProject(), want: domain.ErrDuplicateURL},
		{name: "other unique", f: db(t, fail(other)), in: newProject(), want: other},
		{name: "ensure tags", f: db(t, rows([]any{uid(projID)}), fail(errDB)), in: newProject(), want: errDB},
		{name: "insert tags", f: db(t, rows([]any{uid(projID)}), ok, fail(errDB)), in: newProject(), want: errDB},
		{name: "bad link id", f: db(t, rows([]any{uid(projID)}), ok, ok), in: newProject(domain.LinkRef{Type: "question", ID: "x"}), want: domain.ErrLinkTarget},
		{name: "unknown link type", f: db(t, rows([]any{uid(projID)}), ok, ok), in: newProject(domain.LinkRef{Type: "series", ID: qID}), want: domain.ErrLinkTarget},
		{name: "link error", f: db(t, rows([]any{uid(projID)}), ok, ok, fail(errDB)), in: newProject(domain.LinkRef{Type: "question", ID: qID}), want: errDB},
		{name: "unusable target", f: db(t, rows([]any{uid(projID)}), ok, ok, no), in: newProject(domain.LinkRef{Type: "article", ID: qID}), want: domain.ErrLinkTarget},
		{name: "commit", f: db(t, rows([]any{uid(projID)}), ok, ok), in: newProject(), commit: errDB, want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t, tc.f.beginErr, tc.f.commitErr = t, tc.begin, tc.commit
			_, err := New(tc.f).CreateProject(t.Context(), tc.in)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || tc.f.committed {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	desc, vis, techs := "Nouveau", "private", []string{"go"}
	f := db(t, access, ok, ok, ok, ok, yes)
	err := New(f).UpdateProject(t.Context(), projID, userID, domain.ProjectPatch{
		Description: &desc, Visibility: &vis, Technologies: &techs, Links: []domain.LinkRef{{Type: "question", ID: qID}},
	}, allow)
	if err != nil || !f.committed {
		t.Fatalf("update %v", err)
	}
	empty := []string{}
	if err := New(db(t, access, ok, ok)).UpdateProject(t.Context(), projID, userID, domain.ProjectPatch{Technologies: &empty}, allow); err != nil {
		t.Fatalf("clear technologies %v", err)
	}
	var seen domain.Access
	capture := func(a domain.Access) error { seen = a; return errDeny }
	if err := New(db(t, access)).UpdateProject(t.Context(), projID, userID, domain.ProjectPatch{}, capture); !errors.Is(err, errDeny) ||
		seen != (domain.Access{OwnerID: userID, Visibility: "public"}) {
		t.Fatalf("authorize %v %+v", err, seen)
	}
	tests := []struct {
		name          string
		f             *fakeDB
		id, actor     string
		patch         domain.ProjectPatch
		want          error
		begin, commit error
	}{
		{name: "bad id", f: db(t), id: "x", actor: userID},
		{name: "bad actor", f: db(t), id: projID, actor: "x"},
		{name: "begin", f: db(t), id: projID, actor: userID, begin: errDB, want: errDB},
		{name: "missing", f: db(t, none), id: projID, actor: userID, want: domain.ErrNotFound},
		{name: "lock", f: db(t, fail(errDB)), id: projID, actor: userID, want: errDB},
		{name: "update", f: db(t, access, fail(errDB)), id: projID, actor: userID, want: errDB},
		{name: "delete technologies", f: db(t, access, ok, fail(errDB)), id: projID, actor: userID, patch: domain.ProjectPatch{Technologies: &techs}, want: errDB},
		{name: "link", f: db(t, access, ok, no), id: projID, actor: userID, patch: domain.ProjectPatch{Links: []domain.LinkRef{{Type: "resource", ID: qID}}}, want: domain.ErrLinkTarget},
		{name: "commit", f: db(t, access, ok), id: projID, actor: userID, commit: errDB, want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t, tc.f.beginErr, tc.f.commitErr = t, tc.begin, tc.commit
			err := New(tc.f).UpdateProject(t.Context(), tc.id, tc.actor, tc.patch, allow)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || tc.f.committed {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	related := rows([]any{uid(projID), "question", uid(qID), "Q", "fr"})
	p, err := New(db(t, rows(projectRow(nil)), related)).GetProject(t.Context(), projID, userID)
	want := domain.Project{
		ID: projID, Name: "Repo", GitHubURL: "https://github.com/o/r", Description: "D", Lang: "fr", Visibility: "public",
		Technologies: []string{}, Owner: domain.Owner{ID: userID, Username: "alice", DisplayName: "Alice"}, OpenNeeds: 1,
		Related:     []domain.Related{{Type: "question", ID: qID, Title: "Q", Lang: "fr"}},
		PublishedAt: ts().Time, CreatedAt: ts().Time, UpdatedAt: ts().Time,
	}
	if err != nil || !reflect.DeepEqual(p, want) {
		t.Fatalf("got %+v %v", p, err)
	}
	p, err = New(db(t, rows(projectRow([]string{"go"})), none)).GetProject(t.Context(), projID, "")
	if err != nil || p.Related == nil || len(p.Related) != 0 || p.Technologies[0] != "go" {
		t.Fatalf("anonymous %+v %v", p, err)
	}
	tests := []struct {
		name       string
		f          *fakeDB
		id, viewer string
		want       error
	}{
		{name: "bad id", f: db(t), id: "x"},
		{name: "bad viewer", f: db(t), id: projID, viewer: "x"},
		{name: "missing", f: db(t, none), id: projID, want: domain.ErrNotFound},
		{name: "query", f: db(t, fail(errDB)), id: projID, want: errDB},
		{name: "related", f: db(t, rows(projectRow(nil)), fail(errDB)), id: projID, want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t = t
			_, err := New(tc.f).GetProject(t.Context(), tc.id, tc.viewer)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestListProjects(t *testing.T) {
	at := ts().Time
	list, err := New(db(t, rows(projectRow(nil)), none)).ListProjects(t.Context(), domain.ProjectQuery{
		ViewerID: userID, Q: "pay", Technology: "go", Skill: "go", OpenNeeds: true, AfterPublished: &at, AfterID: projID, Limit: 21,
	})
	if err != nil || len(list) != 1 || list[0].Related == nil {
		t.Fatalf("list %+v %v", list, err)
	}
	if list, err := New(db(t, none)).ListProjects(t.Context(), domain.ProjectQuery{Limit: 21}); err != nil || len(list) != 0 {
		t.Fatalf("empty %+v %v", list, err)
	}
	tests := []struct {
		name string
		f    *fakeDB
		q    domain.ProjectQuery
		want error
	}{
		{name: "bad viewer", f: db(t), q: domain.ProjectQuery{ViewerID: "x"}},
		{name: "bad cursor id", f: db(t), q: domain.ProjectQuery{AfterID: "x"}},
		{name: "query", f: db(t, fail(errDB)), want: errDB},
		{name: "related", f: db(t, rows(projectRow(nil)), fail(errDB)), want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t = t
			if _, err := New(tc.f).ListProjects(t.Context(), tc.q); err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestProjectAccess(t *testing.T) {
	a, err := New(db(t, rows([]any{uid(userID), "private"}))).ProjectAccess(t.Context(), projID)
	if err != nil || a != (domain.Access{OwnerID: userID, Visibility: "private"}) {
		t.Fatalf("access %+v %v", a, err)
	}
	if _, err := New(db(t)).ProjectAccess(t.Context(), "x"); err == nil {
		t.Fatal("bad id accepted")
	}
	if _, err := New(db(t, none)).ProjectAccess(t.Context(), projID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing %v", err)
	}
	if _, err := New(db(t, fail(errDB))).ProjectAccess(t.Context(), projID); !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}
}

func TestCreateNeed(t *testing.T) {
	in := domain.NewNeed{ProjectID: projID, CreatedBy: userID, Title: "Backend Go", Description: "D", Lang: "fr", Skills: []string{"go"}}
	f := db(t, access, rows([]any{uid(needID), "open", ts()}), ok, ok)
	n, err := New(f).CreateNeed(t.Context(), in, allow)
	want := domain.Need{ID: needID, ProjectID: projID, Title: "Backend Go", Description: "D", Lang: "fr", Status: "open", Skills: []string{"go"}, CreatedAt: ts().Time}
	if err != nil || !reflect.DeepEqual(n, want) || !f.committed {
		t.Fatalf("need %+v %v", n, err)
	}
	badProject, badBy := in, in
	badProject.ProjectID, badBy.CreatedBy = "x", "x"
	tests := []struct {
		name          string
		f             *fakeDB
		in            domain.NewNeed
		auth          func(domain.Access) error
		want          error
		begin, commit error
	}{
		{name: "bad project", f: db(t), in: badProject, auth: allow},
		{name: "bad creator", f: db(t), in: badBy, auth: allow},
		{name: "begin", f: db(t), in: in, auth: allow, begin: errDB, want: errDB},
		{name: "missing", f: db(t, none), in: in, auth: allow, want: domain.ErrNotFound},
		{name: "denied", f: db(t, access), in: in, auth: func(domain.Access) error { return errDeny }, want: errDeny},
		{name: "insert", f: db(t, access, fail(errDB)), in: in, auth: allow, want: errDB},
		{name: "ensure tags", f: db(t, access, rows([]any{uid(needID), "open", ts()}), fail(errDB)), in: in, auth: allow, want: errDB},
		{name: "skills", f: db(t, access, rows([]any{uid(needID), "open", ts()}), ok, fail(errDB)), in: in, auth: allow, want: errDB},
		{name: "commit", f: db(t, access, rows([]any{uid(needID), "open", ts()}), ok, ok), in: in, auth: allow, commit: errDB, want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t, tc.f.beginErr, tc.f.commitErr = t, tc.begin, tc.commit
			_, err := New(tc.f).CreateNeed(t.Context(), tc.in, tc.auth)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || tc.f.committed {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestListNeeds(t *testing.T) {
	at := ts().Time
	row := []any{uid(needID), uid(projID), "T", "D", "fr", "open", ts(), []string(nil)}
	needs, err := New(db(t, rows(row))).ListNeeds(t.Context(), domain.NeedQuery{ProjectID: projID, AfterCreated: &at, AfterID: needID, Limit: 21})
	if err != nil || len(needs) != 1 || needs[0].Skills == nil || needs[0].ProjectID != projID {
		t.Fatalf("needs %+v %v", needs, err)
	}
	if _, err := New(db(t)).ListNeeds(t.Context(), domain.NeedQuery{ProjectID: "x"}); err == nil {
		t.Fatal("bad project accepted")
	}
	if _, err := New(db(t)).ListNeeds(t.Context(), domain.NeedQuery{ProjectID: projID, AfterID: "x"}); err == nil {
		t.Fatal("bad cursor accepted")
	}
	if _, err := New(db(t, fail(errDB))).ListNeeds(t.Context(), domain.NeedQuery{ProjectID: projID}); !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}
}
