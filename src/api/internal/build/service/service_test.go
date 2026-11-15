package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	projID  = "00000000-0000-7000-8000-0000000000a1"
	aliceID = "00000000-0000-7000-8000-0000000000b1"
	bobID   = "00000000-0000-7000-8000-0000000000b2"
)

var (
	errDB   = errors.New("db down")
	alice   = subject.Subject{UserID: aliceID, EmailVerified: true}
	bob     = subject.Subject{UserID: bobID, EmailVerified: true}
	unverif = subject.Subject{UserID: bobID}
	anon    = subject.Subject{}
	at      = time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
)

type fakeStore struct {
	taken      bool
	takenErr   error
	createErr  error
	created    domain.NewProject
	updateErr  error
	patch      domain.ProjectPatch
	access     domain.Access
	accessErr  error
	authorized error
	getErr     error
	getCalls   int
	getFailAt  int
	list       []domain.Project
	listErr    error
	query      domain.ProjectQuery
	needErr    error
	newNeed    domain.NewNeed
	needs      []domain.Need
	needsErr   error
	needQuery  domain.NeedQuery
	apps       appFake
	total      int
	countErr   error
	countQuery domain.ProjectQuery
}

func (f *fakeStore) URLTaken(context.Context, string) (bool, error) { return f.taken, f.takenErr }

func (f *fakeStore) CreateProject(_ context.Context, in domain.NewProject) (string, error) {
	f.created = in
	return projID, f.createErr
}

func (f *fakeStore) UpdateProject(_ context.Context, _, _ string, p domain.ProjectPatch, authorize func(domain.Access) error) error {
	f.patch = p
	if f.updateErr != nil {
		return f.updateErr
	}
	f.authorized = authorize(f.access)
	return f.authorized
}

func (f *fakeStore) GetProject(context.Context, string, string) (domain.Project, error) {
	f.getCalls++
	if f.getErr != nil && (f.getFailAt == 0 || f.getFailAt == f.getCalls) {
		return domain.Project{}, f.getErr
	}
	return domain.Project{ID: projID, Name: "Repo"}, nil
}

func (f *fakeStore) ListProjects(_ context.Context, q domain.ProjectQuery) ([]domain.Project, error) {
	f.query = q
	return f.list, f.listErr
}

func (f *fakeStore) CountProjects(_ context.Context, q domain.ProjectQuery) (int, error) {
	f.countQuery = q
	return f.total, f.countErr
}

func (f *fakeStore) ProjectAccess(context.Context, string) (domain.Access, error) {
	return f.access, f.accessErr
}

func (f *fakeStore) CreateNeed(_ context.Context, in domain.NewNeed, authorize func(domain.Access) error) (domain.Need, error) {
	f.newNeed = in
	if f.needErr != nil {
		return domain.Need{}, f.needErr
	}
	if err := authorize(f.access); err != nil {
		return domain.Need{}, err
	}
	return domain.Need{ID: "n1", Title: in.Title}, nil
}

func (f *fakeStore) ListNeeds(_ context.Context, q domain.NeedQuery) ([]domain.Need, error) {
	f.needQuery = q
	return f.needs, f.needsErr
}

type fakeGitHub struct {
	repo  domain.Repo
	err   error
	calls int
}

func (g *fakeGitHub) Repo(context.Context, string, string) (domain.Repo, error) {
	g.calls++
	return g.repo, g.err
}

// fakeLimiter denies or fails one scope; every other scope is allowed.
type fakeLimiter struct {
	scope  string
	deny   bool
	err    error
	hits   []string
	value  string
	limit  int
	window time.Duration
}

func (l *fakeLimiter) Hit(_ context.Context, scope, value string, limit int, window time.Duration) (bool, time.Duration, error) {
	l.hits = append(l.hits, scope)
	l.value, l.limit, l.window = value, limit, window
	if scope != l.scope {
		return true, 0, nil
	}
	if l.err != nil {
		return false, 0, l.err
	}
	return !l.deny, 7 * time.Second, nil
}

func codec(t *testing.T) *cursor.Codec {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newSvc(t *testing.T, st *fakeStore, gh *fakeGitHub, lim *fakeLimiter) *Service {
	t.Helper()
	if gh == nil {
		gh = &fakeGitHub{repo: domain.Repo{Name: "Repo", Description: "  Un outil  "}}
	}
	if lim == nil {
		lim = &fakeLimiter{}
	}
	return New(st, gh, codec(t), lim)
}

// problem is what the reject tests read back from an apperr.Error.
type problem struct {
	Violations []apperr.Violation
	RetryAfter time.Duration
}

// problemOf asserts err is a problem with code and returns it.
func problemOf(t *testing.T, err error, code apperr.Code) *problem {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != code {
		t.Fatalf("err %v, want %s", err, code)
	}
	return &problem{Violations: ae.Violations, RetryAfter: ae.RetryAfter}
}

func wantCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	problemOf(t, err, code)
}

func ptr[T any](v T) *T { return &v }

func projects(n int) []domain.Project {
	out := make([]domain.Project, n)
	for i := range out {
		out[i] = domain.Project{ID: projID, PublishedAt: at, UpdatedAt: at}
	}
	return out
}

func TestListProjectsPagesAndBindsTheCursor(t *testing.T) {
	st := &fakeStore{list: projects(3)}
	svc := newSvc(t, st, nil, nil)
	page, err := svc.ListProjects(t.Context(), alice, ProjectListRequest{
		Q: ptr("paiement"), Technology: ptr("go"), Skill: ptr("sql"), OpenNeeds: ptr(true), Limit: ptr(2),
	})
	if err != nil || len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("page %+v %v", page, err)
	}
	if st.query != (domain.ProjectQuery{ViewerID: aliceID, Q: "paiement", Technology: "go", Skill: "sql", OpenNeeds: true, Sort: domain.SortUpdated, Limit: 3}) {
		t.Fatalf("query %+v", st.query)
	}
	st.list = projects(1)
	page2, err := svc.ListProjects(t.Context(), alice, ProjectListRequest{
		Q: ptr("paiement"), Technology: ptr("go"), Skill: ptr("sql"), OpenNeeds: ptr(true), Limit: ptr(2), Cursor: page.NextCursor,
	})
	if err != nil || page2.NextCursor != "" || st.query.AfterID != projID || !st.query.AfterAt.Equal(at) || st.query.AfterNeeds != 0 {
		t.Fatalf("page2 %+v %v %+v", page2, err, st.query)
	}
	// The cursor is bound to the viewer and the filters.
	_, err = svc.ListProjects(t.Context(), bob, ProjectListRequest{Q: ptr("paiement"), Technology: ptr("go"), Skill: ptr("sql"), OpenNeeds: ptr(true), Cursor: page.NextCursor})
	wantCode(t, err, apperr.CodeInvalidCursor)

	st.list = nil
	empty, err := svc.ListProjects(t.Context(), anon, ProjectListRequest{})
	if err != nil || empty.Items == nil || st.query.Limit != defaultLimit+1 || st.query.OpenNeeds {
		t.Fatalf("empty %+v %v", empty, err)
	}
}

// sort: updated (default) keys on updated_at, created on created_at, needs on the open-needs count then
// updated_at; each cursor carries its own key and is refused under another sort.
func TestListProjectsSort(t *testing.T) {
	created, updated := at.Add(-time.Hour), at
	rows := []domain.Project{
		{ID: projID, OpenNeeds: 4, CreatedAt: created, UpdatedAt: updated},
		{ID: projID, OpenNeeds: 1, CreatedAt: created, UpdatedAt: updated},
	}
	st := &fakeStore{list: rows}
	svc := newSvc(t, st, nil, nil)
	cursors := map[string]string{}
	for _, tc := range []struct {
		sort      *string
		want      string
		wantNeeds int
		wantAt    time.Time
	}{
		{nil, domain.SortUpdated, 0, updated},
		{ptr("updated"), domain.SortUpdated, 0, updated},
		{ptr("created"), domain.SortCreated, 0, created},
		{ptr("needs"), domain.SortNeeds, 4, updated},
	} {
		page, err := svc.ListProjects(t.Context(), anon, ProjectListRequest{Sort: tc.sort, Limit: ptr(1)})
		if err != nil || st.query.Sort != tc.want || page.NextCursor == "" {
			t.Fatalf("%s: %+v %v %+v", tc.want, page, err, st.query)
		}
		if _, err := svc.ListProjects(t.Context(), anon, ProjectListRequest{Sort: tc.sort, Limit: ptr(1), Cursor: page.NextCursor}); err != nil ||
			st.query.AfterNeeds != tc.wantNeeds || !st.query.AfterAt.Equal(tc.wantAt) || st.query.AfterID != projID {
			t.Fatalf("%s page 2: %v %+v", tc.want, err, st.query)
		}
		cursors[tc.want] = page.NextCursor
	}
	for from, token := range cursors {
		for _, to := range []string{"created", "needs", "updated"} {
			if to == from {
				continue
			}
			_, err := svc.ListProjects(t.Context(), anon, ProjectListRequest{Sort: ptr(to), Cursor: token})
			wantCode(t, err, apperr.CodeInvalidCursor)
		}
	}
}

// owner reaches the page and the count (the store applies the profile rule) and binds the cursor.
func TestListProjectsOwnerFilter(t *testing.T) {
	st := &fakeStore{list: projects(2)}
	svc := newSvc(t, st, nil, nil)
	page, err := svc.ListProjects(t.Context(), bob, ProjectListRequest{Owner: ptr("alice"), Limit: ptr(1)})
	if err != nil || st.query.Owner != "alice" || st.countQuery.Owner != "alice" || st.query.ViewerID != bobID || page.NextCursor == "" {
		t.Fatalf("page %+v %v %+v", page, err, st.query)
	}
	_, err = svc.ListProjects(t.Context(), bob, ProjectListRequest{Owner: ptr("carol"), Cursor: page.NextCursor})
	wantCode(t, err, apperr.CodeInvalidCursor)
	_, err = svc.ListProjects(t.Context(), bob, ProjectListRequest{Cursor: page.NextCursor})
	wantCode(t, err, apperr.CodeInvalidCursor)
}

// total counts with the same filters and viewer as the page, and does not depend on the cursor.
func TestListProjectsTotal(t *testing.T) {
	st := &fakeStore{list: projects(3), total: 9}
	svc := newSvc(t, st, nil, nil)
	req := ProjectListRequest{Q: ptr("pay"), Stage: ptr("active"), Limit: ptr(2)}
	page, err := svc.ListProjects(t.Context(), alice, req)
	if err != nil || page.Total != 9 {
		t.Fatalf("page %+v %v", page, err)
	}
	want := st.query
	want.Limit, want.AfterAt, want.AfterID = 0, nil, ""
	st.countQuery.Limit = 0
	if st.countQuery != want {
		t.Fatalf("count query %+v, want %+v", st.countQuery, want)
	}
	req.Cursor = page.NextCursor
	if page2, err := svc.ListProjects(t.Context(), alice, req); err != nil || page2.Total != 9 || st.countQuery.AfterAt != nil || st.countQuery.AfterID != "" {
		t.Fatalf("page2 %+v %v %+v", page2, err, st.countQuery)
	}
	_, err = newSvc(t, &fakeStore{countErr: errDB}, nil, nil).ListProjects(t.Context(), alice, ProjectListRequest{})
	wantCode(t, err, apperr.CodeInternal)
}

func TestListProjectsAnonymousQueryUsesSearchBudget(t *testing.T) {
	lim := &fakeLimiter{scope: "search:ip", deny: true}
	_, err := newSvc(t, &fakeStore{}, nil, lim).ListProjects(t.Context(), anon, ProjectListRequest{
		Q: ptr("paiement"), ClientIP: "198.51.100.0/24",
	})
	p := problemOf(t, err, apperr.CodeRateLimited)
	if p.RetryAfter != 7*time.Second {
		t.Fatalf("retry_after = %s", p.RetryAfter)
	}
	if len(lim.hits) != 1 || lim.hits[0] != "search:ip" || lim.value != "198.51.100.0/24" || lim.limit != 60 || lim.window != time.Minute {
		t.Fatalf("limiter = %+v", lim)
	}
}

func TestListProjectsRejects(t *testing.T) {
	c := codec(t)
	hash := lengthPrefixed("project", "", "", "", "", "false", "", "", "", domain.SortUpdated, "")
	tests := []struct {
		name string
		req  ProjectListRequest
		code apperr.Code
		err  error
	}{
		{name: "blank q", req: ProjectListRequest{Q: ptr("   ")}, code: apperr.CodeValidation},
		{name: "nul q", req: ProjectListRequest{Q: ptr("a\x00")}, code: apperr.CodeValidation},
		{name: "limit 0", req: ProjectListRequest{Limit: ptr(0)}, code: apperr.CodeValidation},
		{name: "limit 51", req: ProjectListRequest{Limit: ptr(51)}, code: apperr.CodeValidation},
		{name: "forged cursor", req: ProjectListRequest{Cursor: "x"}, code: apperr.CodeInvalidCursor},
		{name: "bad sort", req: ProjectListRequest{Cursor: c.Encode(cursor.Payload{Sort: "x", ID: projID, FilterHash: hash})}, code: apperr.CodeInvalidCursor},
		{name: "bad id", req: ProjectListRequest{Cursor: c.Encode(cursor.Payload{Sort: "0:" + at.Format(time.RFC3339Nano), ID: "x", FilterHash: hash})}, code: apperr.CodeInvalidCursor},
		{name: "no needs count", req: ProjectListRequest{Cursor: c.Encode(cursor.Payload{Sort: at.Format(time.RFC3339Nano), ID: projID, FilterHash: hash})}, code: apperr.CodeInvalidCursor},
		{name: "negative needs", req: ProjectListRequest{Cursor: c.Encode(cursor.Payload{Sort: "-1:" + at.Format(time.RFC3339Nano), ID: projID, FilterHash: hash})}, code: apperr.CodeInvalidCursor},
		{name: "needs overflow", req: ProjectListRequest{Cursor: c.Encode(cursor.Payload{Sort: "2147483648:" + at.Format(time.RFC3339Nano), ID: projID, FilterHash: hash})}, code: apperr.CodeInvalidCursor},
		{name: "bad time", req: ProjectListRequest{Cursor: c.Encode(cursor.Payload{Sort: "0:x", ID: projID, FilterHash: hash})}, code: apperr.CodeInvalidCursor},
		{name: "store", req: ProjectListRequest{}, code: apperr.CodeInternal, err: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newSvc(t, &fakeStore{listErr: tc.err}, nil, nil).ListProjects(t.Context(), anon, tc.req)
			wantCode(t, err, tc.code)
		})
	}
}

// Project facts (owner decision 2026-11-15): stage, license_spdx and country_code are written on create
// and update, filter the list, and bind the cursor.
func TestProjectFacts(t *testing.T) {
	st := &fakeStore{}
	svc := newSvc(t, st, nil, nil)
	if _, err := svc.CreateProject(t.Context(), alice, createReq()); err != nil {
		t.Fatal(err)
	}
	if st.created.Stage != domain.StageIdea || st.created.License != "" || st.created.Country != "" {
		t.Fatalf("defaults %+v", st.created)
	}
	req := createReq()
	req.Stage, req.License, req.Country = ptr("prototype"), ptr("MIT"), ptr("BF")
	if _, err := svc.CreateProject(t.Context(), alice, req); err != nil {
		t.Fatal(err)
	}
	if st.created.Stage != "prototype" || st.created.License != "MIT" || st.created.Country != "BF" {
		t.Fatalf("facts %+v", st.created)
	}

	st = &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
	clear := &domain.Clearable{}
	if _, err := newSvc(t, st, nil, nil).UpdateProject(t.Context(), alice, projID, UpdateProjectRequest{
		Stage: ptr("active"), License: clear, Country: &domain.Clearable{Value: ptr("SN")},
	}); err != nil {
		t.Fatal(err)
	}
	if *st.patch.Stage != "active" || st.patch.License != clear || *st.patch.Country.Value != "SN" {
		t.Fatalf("patch %+v", st.patch)
	}

	for name, tc := range map[string]struct {
		err   error
		field string
	}{"unknown country": {domain.ErrUnknownCountry, "/country_code"}, "stage without repository": {domain.ErrStageNeedsRepo, "/stage"}} {
		t.Run(name, func(t *testing.T) {
			_, err := newSvc(t, &fakeStore{createErr: tc.err}, nil, nil).CreateProject(t.Context(), alice, createReq())
			if p := problemOf(t, err, apperr.CodeValidation); len(p.Violations) != 1 || p.Violations[0].Field != tc.field {
				t.Fatalf("create violations %+v", p.Violations)
			}
			_, err = newSvc(t, &fakeStore{updateErr: tc.err}, nil, nil).UpdateProject(t.Context(), alice, projID, UpdateProjectRequest{Stage: ptr("active")})
			if p := problemOf(t, err, apperr.CodeValidation); len(p.Violations) != 1 || p.Violations[0].Field != tc.field {
				t.Fatalf("update violations %+v", p.Violations)
			}
		})
	}

	st = &fakeStore{list: projects(2)}
	svc = newSvc(t, st, nil, nil)
	filters := ProjectListRequest{Stage: ptr("active"), Country: ptr("BF"), License: ptr("MIT"), Limit: ptr(1)}
	page, err := svc.ListProjects(t.Context(), anon, filters)
	if err != nil || st.query.Stage != "active" || st.query.Country != "BF" || st.query.License != "MIT" || page.NextCursor == "" {
		t.Fatalf("list %+v %v %+v", page, err, st.query)
	}
	// The cursor is bound to each fact filter.
	for _, other := range []ProjectListRequest{
		{Stage: ptr("idea"), Country: ptr("BF"), License: ptr("MIT")},
		{Stage: ptr("active"), Country: ptr("SN"), License: ptr("MIT")},
		{Stage: ptr("active"), Country: ptr("BF"), License: ptr("GPL-3.0")},
	} {
		other.Cursor = page.NextCursor
		_, err := svc.ListProjects(t.Context(), anon, other)
		wantCode(t, err, apperr.CodeInvalidCursor)
	}
}

func TestGetProject(t *testing.T) {
	if p, err := newSvc(t, &fakeStore{}, nil, nil).GetProject(t.Context(), anon, projID); err != nil || p.ID != projID {
		t.Fatalf("get %+v %v", p, err)
	}
	_, err := newSvc(t, &fakeStore{getErr: domain.ErrNotFound}, nil, nil).GetProject(t.Context(), anon, projID)
	wantCode(t, err, apperr.CodeNotFound)
	_, err = newSvc(t, &fakeStore{getErr: errDB}, nil, nil).GetProject(t.Context(), anon, projID)
	wantCode(t, err, apperr.CodeInternal)
}

func createReq() CreateProjectRequest {
	return CreateProjectRequest{
		GitHubURL: "https://github.com/Org/Repo.git", Lang: "fr", Description: ptr("Mon projet"),
		Technologies: []string{"go", "go", "sql"},
		Links:        []domain.LinkRef{{Type: "question", ID: "q"}, {Type: "question", ID: "q"}, {Type: "article", ID: "a"}},
	}
}

func TestCreateProjectRegistersTheCanonicalRepository(t *testing.T) {
	st, gh, lim := &fakeStore{}, &fakeGitHub{repo: domain.Repo{Name: "Repo", Description: "  Un outil  "}}, &fakeLimiter{}
	p, err := newSvc(t, st, gh, lim).CreateProject(t.Context(), alice, createReq())
	if err != nil || p.ID != projID {
		t.Fatalf("create %+v %v", p, err)
	}
	want := domain.NewProject{
		OwnerID: aliceID, Name: "Repo", Summary: "Un outil", Description: "Mon projet", Lang: "fr",
		GitHubURL: "https://github.com/org/repo", Visibility: "public", Technologies: []string{"go", "sql"},
		Links: []domain.LinkRef{{Type: "question", ID: "q"}, {Type: "article", ID: "a"}},
	}
	if st.created.Name != want.Name || st.created.Summary != want.Summary || st.created.GitHubURL != want.GitHubURL ||
		st.created.Visibility != want.Visibility || strings.Join(st.created.Technologies, ",") != "go,sql" ||
		len(st.created.Links) != 2 || st.created.OwnerID != aliceID || st.created.Description != "Mon projet" {
		t.Fatalf("created %+v", st.created)
	}
	if strings.Join(lim.hits, ",") != "write:user,import:github" || gh.calls != 1 {
		t.Fatalf("hits %v github %d", lim.hits, gh.calls)
	}

	// Private, no description, GitHub has none either: the summary is the repository name.
	st = &fakeStore{}
	req := CreateProjectRequest{GitHubURL: "https://github.com/org/repo", Lang: "en", Visibility: ptr("private")}
	if _, err := newSvc(t, st, &fakeGitHub{repo: domain.Repo{Name: "repo"}}, nil).CreateProject(t.Context(), alice, req); err != nil {
		t.Fatal(err)
	}
	if st.created.Summary != "repo" || st.created.Visibility != "private" || st.created.Description != "" || st.created.Technologies == nil {
		t.Fatalf("created %+v", st.created)
	}
	// A long GitHub description is cut to 500 characters, never mid-rune.
	st = &fakeStore{}
	long := strings.Repeat("é", 600)
	if _, err := newSvc(t, st, &fakeGitHub{repo: domain.Repo{Name: "r", Description: long}}, nil).CreateProject(t.Context(), alice, req); err != nil {
		t.Fatal(err)
	}
	if st.created.Summary != strings.Repeat("é", 500) {
		t.Fatalf("summary length %d", len([]rune(st.created.Summary)))
	}
}

func TestCreateProjectRejects(t *testing.T) {
	tests := []struct {
		name       string
		sub        subject.Subject
		req        func(*CreateProjectRequest)
		st         *fakeStore
		gh         *fakeGitHub
		lim        *fakeLimiter
		code       apperr.Code
		field      string
		noGitHub   bool
		retryAfter bool
	}{
		{name: "anonymous", sub: anon, code: apperr.CodeUnauthenticated, noGitHub: true},
		{name: "unverified", sub: unverif, code: apperr.CodeEmailNotVerified, noGitHub: true},
		{name: "not github", sub: alice, req: func(r *CreateProjectRequest) { r.GitHubURL = "https://gitlab.com/o/r" }, code: apperr.CodeInvalidGitHubURL, field: "/github_url", noGitHub: true},
		{name: "nul description", sub: alice, req: func(r *CreateProjectRequest) { r.Description = ptr("a\x00") }, code: apperr.CodeValidation, field: "/description", noGitHub: true},
		{name: "series link", sub: alice, req: func(r *CreateProjectRequest) { r.Links = []domain.LinkRef{{Type: "series", ID: "s"}} }, code: apperr.CodeValidation, field: "/links", noGitHub: true},
		{name: "write budget", sub: alice, lim: &fakeLimiter{scope: writeScope, deny: true}, code: apperr.CodeRateLimited, noGitHub: true, retryAfter: true},
		{name: "limiter down", sub: alice, lim: &fakeLimiter{scope: writeScope, err: errDB}, code: apperr.CodeInternal, noGitHub: true},
		{name: "url check", sub: alice, st: &fakeStore{takenErr: errDB}, code: apperr.CodeInternal, noGitHub: true},
		{name: "duplicate before github", sub: alice, st: &fakeStore{taken: true}, code: apperr.CodeStateConflict, field: "/github_url", noGitHub: true},
		{name: "import budget", sub: alice, lim: &fakeLimiter{scope: importScope, deny: true}, code: apperr.CodeRateLimited, noGitHub: true, retryAfter: true},
		{name: "repository missing", sub: alice, gh: &fakeGitHub{err: domain.ErrRepoNotFound}, code: apperr.CodeInvalidGitHubURL, field: "/github_url"},
		{name: "github down", sub: alice, gh: &fakeGitHub{err: domain.ErrUpstream}, code: apperr.CodeUpstreamUnavailable},
		{name: "duplicate race", sub: alice, st: &fakeStore{createErr: domain.ErrDuplicateURL}, code: apperr.CodeStateConflict},
		{name: "link target", sub: alice, st: &fakeStore{createErr: domain.ErrLinkTarget}, code: apperr.CodeValidation, field: "/links"},
		{name: "store", sub: alice, st: &fakeStore{createErr: errDB}, code: apperr.CodeInternal},
		{name: "reread", sub: alice, st: &fakeStore{getErr: errDB}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createReq()
			if tc.req != nil {
				tc.req(&req)
			}
			st, gh := tc.st, tc.gh
			if st == nil {
				st = &fakeStore{}
			}
			if gh == nil {
				gh = &fakeGitHub{repo: domain.Repo{Name: "Repo"}}
			}
			_, err := newSvc(t, st, gh, tc.lim).CreateProject(t.Context(), tc.sub, req)
			ae := problemOf(t, err, tc.code)
			if tc.field != "" && (len(ae.Violations) != 1 || ae.Violations[0].Field != tc.field) {
				t.Fatalf("violations %+v", ae.Violations)
			}
			if tc.noGitHub && gh.calls != 0 {
				t.Fatal("GitHub was called")
			}
			if tc.retryAfter && ae.RetryAfter != 7*time.Second {
				t.Fatalf("retry after %v", ae.RetryAfter)
			}
		})
	}
}

func TestUpdateProjectOwnerOnly(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
	req := UpdateProjectRequest{Description: ptr(""), Visibility: ptr("private"), Technologies: &[]string{"go", "go"},
		Links: []domain.LinkRef{{Type: "resource", ID: "r"}}}
	if p, err := newSvc(t, st, nil, nil).UpdateProject(t.Context(), alice, projID, req); err != nil || p.ID != projID {
		t.Fatalf("update %+v %v", p, err)
	}
	if *st.patch.Description != "" || *st.patch.Visibility != "private" || strings.Join(*st.patch.Technologies, ",") != "go" || len(st.patch.Links) != 1 {
		t.Fatalf("patch %+v", st.patch)
	}
	st = &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
	if _, err := newSvc(t, st, nil, nil).UpdateProject(t.Context(), alice, projID, UpdateProjectRequest{Visibility: ptr("public")}); err != nil || st.patch.Technologies != nil {
		t.Fatalf("no technologies %v %+v", err, st.patch)
	}

	tests := []struct {
		name string
		sub  subject.Subject
		req  UpdateProjectRequest
		st   *fakeStore
		lim  *fakeLimiter
		code apperr.Code
	}{
		{name: "anonymous", sub: anon, code: apperr.CodeUnauthenticated},
		{name: "nul description", sub: alice, req: UpdateProjectRequest{Description: ptr("\x00")}, code: apperr.CodeValidation},
		{name: "project link", sub: alice, req: UpdateProjectRequest{Links: []domain.LinkRef{{Type: "project", ID: "p"}}}, code: apperr.CodeValidation},
		{name: "rate limited", sub: alice, lim: &fakeLimiter{scope: writeScope, deny: true}, code: apperr.CodeRateLimited},
		{name: "non-owner public", sub: bob, st: &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}, code: apperr.CodeForbidden},
		{name: "non-owner private", sub: bob, st: &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "private"}}, code: apperr.CodeNotFound},
		{name: "missing", sub: alice, st: &fakeStore{updateErr: domain.ErrNotFound}, code: apperr.CodeNotFound},
		{name: "store", sub: alice, st: &fakeStore{updateErr: errDB}, code: apperr.CodeInternal},
		{name: "reread", sub: alice, st: &fakeStore{access: domain.Access{OwnerID: aliceID}, getErr: errDB}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := tc.st
			if st == nil {
				st = &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
			}
			_, err := newSvc(t, st, nil, tc.lim).UpdateProject(t.Context(), tc.sub, projID, tc.req)
			wantCode(t, err, tc.code)
		})
	}
}

func needs(n int) []domain.Need {
	out := make([]domain.Need, n)
	for i := range out {
		out[i] = domain.Need{ID: projID, CreatedAt: at}
	}
	return out
}

// A signed need cursor whose position cannot be read is still INVALID_CURSOR (the project list has its
// own key since the sort, so the time-and-id decoder is exercised here).
func TestListNeedsRejectsAnUnreadablePosition(t *testing.T) {
	c := codec(t)
	hash := lengthPrefixed("need", projID)
	for name, p := range map[string]cursor.Payload{
		"bad time": {Sort: "x", ID: projID, FilterHash: hash},
		"bad id":   {Sort: at.Format(time.RFC3339Nano), ID: "x", FilterHash: hash},
	} {
		t.Run(name, func(t *testing.T) {
			st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
			_, err := newSvc(t, st, nil, nil).ListNeeds(t.Context(), anon, projID, NeedListRequest{Cursor: c.Encode(p)})
			wantCode(t, err, apperr.CodeInvalidCursor)
		})
	}
}

func TestListNeedsFollowsProjectVisibility(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}, needs: needs(3)}
	svc := newSvc(t, st, nil, nil)
	page, err := svc.ListNeeds(t.Context(), anon, projID, NeedListRequest{Limit: ptr(2)})
	if err != nil || len(page.Items) != 2 || page.NextCursor == "" || st.needQuery.Limit != 3 || st.needQuery.ViewerID != "" {
		t.Fatalf("page %+v %v", page, err)
	}
	st.needs = nil
	page, err = svc.ListNeeds(t.Context(), anon, projID, NeedListRequest{Limit: ptr(2), Cursor: page.NextCursor})
	if err != nil || page.Items == nil || st.needQuery.AfterID != projID {
		t.Fatalf("page2 %+v %v", page, err)
	}
	st.access.Visibility = "private"
	if _, err := svc.ListNeeds(t.Context(), alice, projID, NeedListRequest{}); err != nil || st.needQuery.ViewerID != aliceID {
		t.Fatalf("owner on private %v, viewer %q", err, st.needQuery.ViewerID)
	}
	for _, sub := range []subject.Subject{anon, bob} {
		_, err := svc.ListNeeds(t.Context(), sub, projID, NeedListRequest{})
		wantCode(t, err, apperr.CodeNotFound)
	}
	tests := []struct {
		name string
		st   *fakeStore
		req  NeedListRequest
		code apperr.Code
	}{
		{name: "limit", st: &fakeStore{}, req: NeedListRequest{Limit: ptr(0)}, code: apperr.CodeValidation},
		{name: "missing project", st: &fakeStore{accessErr: domain.ErrNotFound}, code: apperr.CodeNotFound},
		{name: "access error", st: &fakeStore{accessErr: errDB}, code: apperr.CodeInternal},
		{name: "cursor", st: &fakeStore{access: domain.Access{Visibility: "public"}}, req: NeedListRequest{Cursor: "x"}, code: apperr.CodeInvalidCursor},
		{name: "store", st: &fakeStore{access: domain.Access{Visibility: "public"}, needsErr: errDB}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newSvc(t, tc.st, nil, nil).ListNeeds(t.Context(), anon, projID, tc.req)
			wantCode(t, err, tc.code)
		})
	}
}

func TestCreateNeedOwnerOnly(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "private"}}
	req := CreateNeedRequest{Title: "Backend Go", Description: "Écrire l'API", Lang: "fr", Skills: []string{"go", "go"}}
	n, err := newSvc(t, st, nil, nil).CreateNeed(t.Context(), alice, projID, req)
	if err != nil || n.ID != "n1" || st.newNeed.CreatedBy != aliceID || st.newNeed.ProjectID != projID || strings.Join(st.newNeed.Skills, ",") != "go" {
		t.Fatalf("need %+v %v %+v", n, err, st.newNeed)
	}
	tests := []struct {
		name string
		sub  subject.Subject
		req  CreateNeedRequest
		st   *fakeStore
		lim  *fakeLimiter
		code apperr.Code
	}{
		{name: "anonymous", sub: anon, req: req, code: apperr.CodeUnauthenticated},
		{name: "unverified", sub: unverif, req: req, code: apperr.CodeEmailNotVerified},
		{name: "blank title", sub: alice, req: CreateNeedRequest{Title: "   ", Description: "d"}, code: apperr.CodeValidation},
		{name: "nul description", sub: alice, req: CreateNeedRequest{Title: "Titre", Description: "\x00"}, code: apperr.CodeValidation},
		{name: "rate limited", sub: alice, req: req, lim: &fakeLimiter{scope: writeScope, deny: true}, code: apperr.CodeRateLimited},
		{name: "non-owner public", sub: bob, req: req, st: &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}, code: apperr.CodeForbidden},
		{name: "non-owner private", sub: bob, req: req, st: &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "private"}}, code: apperr.CodeNotFound},
		{name: "missing project", sub: alice, req: req, st: &fakeStore{needErr: domain.ErrNotFound}, code: apperr.CodeNotFound},
		{name: "store", sub: alice, req: req, st: &fakeStore{needErr: errDB}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := tc.st
			if st == nil {
				st = &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
			}
			_, err := newSvc(t, st, nil, tc.lim).CreateNeed(t.Context(), tc.sub, projID, tc.req)
			wantCode(t, err, tc.code)
		})
	}
}
