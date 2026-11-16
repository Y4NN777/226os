//go:build integration

package pgtest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/build"
	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// US-BUILD-001 (projects and contribution needs) on the production router over a migrated database:
// registration from a public GitHub repository (GitHub faked at the port), owner-only writes (C1),
// visibility, idempotency, the GitHub import limit, and the unique-URL backstop under concurrency.

// fakeGitHub answers every repository as public, except those listed as missing, or all when down.
type fakeGitHub struct {
	mu      sync.Mutex
	calls   int
	missing map[string]bool
	down    bool
}

func (g *fakeGitHub) Repo(_ context.Context, owner, name string) (domain.Repo, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls++
	switch {
	case g.down:
		return domain.Repo{}, domain.ErrUpstream
	case g.missing[owner+"/"+name]:
		return domain.Repo{}, domain.ErrRepoNotFound
	}
	return domain.Repo{Name: strings.ToUpper(name[:1]) + name[1:], Description: "Dépôt " + name}, nil
}

func (g *fakeGitHub) count() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}

func (g *fakeGitHub) setDown(down bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.down = down
}

func projectBody(url, visibility string, extra ...string) string {
	b := `{"github_url":"` + url + `","lang":"fr","visibility":"` + visibility + `","technologies":["go","mobile-money"]`
	for _, e := range extra {
		b += "," + e
	}
	return b + "}"
}

// write sends a JSON write as m, or anonymously (with a pre-session CSRF token) when m is nil.
func (s *s2Stack) write(t *testing.T, m *member, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	h := map[string]string{}
	for k, v := range headers {
		h[k] = v
	}
	if m == nil {
		cookie, token := s.preSession(t)
		h["X-CSRF-Token"] = token
		return s.do(t, method, path, body, h, cookie)
	}
	if _, ok := h["X-CSRF-Token"]; !ok {
		h["X-CSRF-Token"] = m.csrf
	}
	return s.do(t, method, path, body, h, m.session)
}

func (s *s2Stack) read(t *testing.T, m *member, path string) *httptest.ResponseRecorder {
	t.Helper()
	if m == nil {
		return s.do(t, http.MethodGet, path, "", nil)
	}
	return s.do(t, http.MethodGet, path, "", nil, m.session)
}

// liveProjects counts projects registered from GitHub (the catalogue seed has one without a URL).
func (s *s2Stack) liveProjects(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM projects WHERE github_url IS NOT NULL`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (s *s2Stack) needCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM contribution_needs`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (s *s2Stack) createProject(t *testing.T, m *member, url, visibility string) apigen.Project {
	t.Helper()
	rec := s.write(t, m, http.MethodPost, "/api/v1/projects", projectBody(url, visibility), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %s: %d %s", url, rec.Code, rec.Body.String())
	}
	return decodeAs[apigen.Project](t, rec)
}

func projectIDs(page apigen.ProjectPage) []string {
	out := make([]string, len(page.Items))
	for i, p := range page.Items {
		out[i] = p.Id.String()
	}
	return out
}

func TestBuildProjectRegistration(t *testing.T) {
	gh := &fakeGitHub{missing: map[string]bool{"org/ghost": true}}
	s := newBuildStack(t, gh)
	alice := s.signIn(t, "build-alice", true, 61)
	unverified := s.signIn(t, "build-unverified", false, 62)

	// Anonymous 401 and unverified 403: nothing written, GitHub never called.
	assertProblem(t, s.write(t, nil, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/org/repo", "public"), nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	assertProblem(t, s.write(t, &unverified, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/org/repo", "public"), nil), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
	// CSRF: a wrong token is refused before anything happens.
	rec := s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/org/repo", "public"), map[string]string{"X-CSRF-Token": "wrong"})
	if rec.Code != http.StatusForbidden || s.liveProjects(t) != 0 || gh.count() != 0 {
		t.Fatalf("denied writes: %d, projects %d, github %d", rec.Code, s.liveProjects(t), gh.count())
	}
	// Not on the allowlist: refused by the contract validator or the normaliser, GitHub never called.
	for _, bad := range []string{"https://gitlab.com/org/repo", "http://github.com/org/repo", "https://github.com/org/repo/tree/main", "https://github.com/-org/repo"} {
		rec := s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody(bad, "public"), nil)
		if rec.Code != http.StatusUnprocessableEntity || gh.count() != 0 {
			t.Fatalf("%s: %d %s (github %d)", bad, rec.Code, rec.Body.String(), gh.count())
		}
	}

	// Registered with the canonical URL; name comes from GitHub; the owner is the session subject.
	key := map[string]string{"Idempotency-Key": "build-create-0001"}
	link := `"links":[{"type":"question","id":"` + catQuestion + `"}]`
	body := projectBody("https://github.com/Org/Repo.git", "public", `"description":"Paiement mobile"`, link)
	rec = s.write(t, &alice, http.MethodPost, "/api/v1/projects", body, key)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	p := decodeAs[apigen.Project](t, rec)
	if p.GithubUrl != "https://github.com/org/repo" || p.Name != "Repo" || p.Description != "Paiement mobile" ||
		p.Owner.Username != "build-alice" || p.Owner.Id.String() != alice.id || p.Visibility != "public" ||
		strings.Join(p.Technologies, ",") != "go,mobile-money" || len(p.Related) != 1 || p.Related[0].Id.String() != catQuestion ||
		p.OpenNeedsCount != 0 || rec.Header().Get("Location") != "/api/v1/projects/"+p.Id.String() {
		t.Fatalf("project %+v", p)
	}
	// A retry with the same key replays the 201: no second row, no second GitHub call.
	rec = s.write(t, &alice, http.MethodPost, "/api/v1/projects", body, key)
	if rec.Code != http.StatusCreated || rec.Header().Get("Idempotency-Replayed") != "true" || decodeAs[apigen.Project](t, rec).Id != p.Id || gh.count() != 1 {
		t.Fatalf("replay %d %v github %d", rec.Code, rec.Header(), gh.count())
	}
	// The same repository under another spelling is a 409, and costs no GitHub call.
	assertProblem(t, s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/ORG/Repo", "public"), nil), http.StatusConflict, "STATE_CONFLICT")
	if gh.count() != 1 {
		t.Fatalf("github called for a duplicate (%d)", gh.count())
	}
	// GitHub says the repository is missing or private: 422; GitHub down: 503. Nothing is written.
	assertProblem(t, s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/org/ghost", "public"), nil), http.StatusUnprocessableEntity, "INVALID_GITHUB_URL")
	gh.setDown(true)
	assertProblem(t, s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/org/later", "public"), nil), http.StatusServiceUnavailable, "UPSTREAM_UNAVAILABLE")
	gh.setDown(false)
	// A link to removed content is refused and the whole registration is rolled back.
	rec = s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/org/linked", "public", `"links":[{"type":"question","id":"`+catGoneQ+`"}]`), nil)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if n := s.liveProjects(t); n != 1 {
		t.Fatalf("%d projects after refused writes", n)
	}
}

// Two concurrent registrations of the same repository: exactly one 201, the other 409; one row.
func TestBuildConcurrentRegistrationKeepsOneRow(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "race-alice", true, 69)
	bob := s.signIn(t, "race-bob", true, 70)
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i, m := range []*member{&alice, &bob} {
		wg.Go(func() {
			codes[i] = s.write(t, m, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/shared/tool", "public"), nil).Code
		})
	}
	wg.Wait()
	slices.Sort(codes)
	if codes[0] != http.StatusCreated || codes[1] != http.StatusConflict || s.liveProjects(t) != 1 {
		t.Fatalf("codes %v, %d rows", codes, s.liveProjects(t))
	}
}

func TestBuildVisibilityAndOwnerOnlyWrites(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "vis-alice", true, 63)
	bob := s.signIn(t, "vis-bob", true, 64)
	pub := s.createProject(t, &alice, "https://github.com/alice/public-tool", "public")
	priv := s.createProject(t, &alice, "https://github.com/alice/secret-tool", "private")
	pubPath, privPath := "/api/v1/projects/"+pub.Id.String(), "/api/v1/projects/"+priv.Id.String()

	// Private project: 404 for anonymous and for bob, with the same body as an unknown project.
	unknown := problem(t, s.read(t, nil, "/api/v1/projects/019a0000-0000-7000-8000-00000000ffff"))
	for _, m := range []*member{nil, &bob} {
		rec := s.read(t, m, privPath)
		if rec.Code != http.StatusNotFound || fmt.Sprint(problem(t, rec)) != fmt.Sprint(unknown) {
			t.Fatalf("private read %d %s", rec.Code, rec.Body.String())
		}
		if rec := s.read(t, m, privPath+"/contribution-needs"); rec.Code != http.StatusNotFound {
			t.Fatalf("private needs %d", rec.Code)
		}
	}
	if rec := s.read(t, &alice, privPath); rec.Code != http.StatusOK {
		t.Fatalf("owner read %d", rec.Code)
	}
	// Lists: public projects for everyone, plus the caller's own private ones (newest first).
	anonList := decodeAs[apigen.ProjectPage](t, s.read(t, nil, "/api/v1/projects"))
	aliceList := decodeAs[apigen.ProjectPage](t, s.read(t, &alice, "/api/v1/projects"))
	bobList := decodeAs[apigen.ProjectPage](t, s.read(t, &bob, "/api/v1/projects"))
	if strings.Join(projectIDs(anonList), ",") != pub.Id.String() || strings.Join(projectIDs(bobList), ",") != pub.Id.String() ||
		strings.Join(projectIDs(aliceList), ",") != priv.Id.String()+","+pub.Id.String() {
		t.Fatalf("lists anon %v bob %v alice %v", projectIDs(anonList), projectIDs(bobList), projectIDs(aliceList))
	}

	// updateProject: bob 403 on public, 404 on private; a project_maintainers row grants nothing (C1).
	if _, err := s.conn.Exec(t.Context(), `INSERT INTO project_maintainers (project_id, user_id, added_by) VALUES ($1, $2, $3)`,
		pub.Id.String(), bob.id, alice.id); err != nil {
		t.Fatal(err)
	}
	patch := `{"description":"Repris par bob"}`
	assertProblem(t, s.write(t, &bob, http.MethodPatch, pubPath, patch, nil), http.StatusForbidden, "FORBIDDEN")
	assertProblem(t, s.write(t, &bob, http.MethodPatch, privPath, patch, nil), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.write(t, nil, http.MethodPatch, pubPath, patch, nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	needBody := `{"title":"Backend Go","description":"D","lang":"fr","skills":["go"]}`
	assertProblem(t, s.write(t, &bob, http.MethodPost, pubPath+"/contribution-needs", needBody, nil), http.StatusForbidden, "FORBIDDEN")
	assertProblem(t, s.write(t, &bob, http.MethodPost, privPath+"/contribution-needs", needBody, nil), http.StatusNotFound, "NOT_FOUND")
	if got := decodeAs[apigen.Project](t, s.read(t, nil, pubPath)); got.Description != "Dépôt public-tool" || s.needCount(t) != 0 {
		t.Fatalf("non-owner changed the project: %q, %d needs", got.Description, s.needCount(t))
	}
	// github_url is immutable: it is not a ProjectUpdate field.
	if rec := s.write(t, &alice, http.MethodPatch, pubPath, `{"github_url":"https://github.com/bob/x"}`, nil); rec.Code < 400 || rec.Code >= 500 {
		t.Fatalf("github_url patch %d", rec.Code)
	}
	rec := s.write(t, &alice, http.MethodPatch, pubPath, `{"description":"Outil de paiement","technologies":["python"]}`, nil)
	got := decodeAs[apigen.Project](t, rec)
	if rec.Code != http.StatusOK || got.Description != "Outil de paiement" || strings.Join(got.Technologies, ",") != "python" || got.GithubUrl != pub.GithubUrl {
		t.Fatalf("owner patch %d %+v", rec.Code, got)
	}
	// Going private hides it from anonymous callers at once.
	if rec := s.write(t, &alice, http.MethodPatch, pubPath, `{"visibility":"private"}`, nil); rec.Code != http.StatusOK {
		t.Fatalf("visibility %d", rec.Code)
	}
	if rec := s.read(t, nil, pubPath); rec.Code != http.StatusNotFound {
		t.Fatalf("now private %d", rec.Code)
	}
}

func TestBuildContributionNeeds(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "need-alice", true, 65)
	p := s.createProject(t, &alice, "https://github.com/alice/needs", "public")
	other := s.createProject(t, &alice, "https://github.com/alice/other", "public")
	path := "/api/v1/projects/" + p.Id.String() + "/contribution-needs"
	body := `{"title":"Backend Go","description":"Écrire l'API des paiements.","lang":"fr","skills":["go","postgresql"]}`

	assertProblem(t, s.write(t, nil, http.MethodPost, path, body, nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	key := map[string]string{"Idempotency-Key": "build-need-key-000001"}
	rec := s.write(t, &alice, http.MethodPost, path, body, key)
	need := decodeAs[apigen.ContributionNeed](t, rec)
	if rec.Code != http.StatusCreated || need.Status != "open" || strings.Join(need.Skills, ",") != "go,postgresql" || need.ProjectId != p.Id {
		t.Fatalf("need %d %+v", rec.Code, need)
	}
	if rec := s.write(t, &alice, http.MethodPost, path, body, key); rec.Code != http.StatusCreated || s.needCount(t) != 1 {
		t.Fatalf("replay %d, %d needs", rec.Code, s.needCount(t))
	}
	page := decodeAs[apigen.NeedPage](t, s.read(t, nil, path))
	if len(page.Items) != 1 || page.Items[0].Description != "Écrire l'API des paiements." || page.NextCursor != nil {
		t.Fatalf("needs %+v", page)
	}
	if got := decodeAs[apigen.Project](t, s.read(t, nil, "/api/v1/projects/"+p.Id.String())); got.OpenNeedsCount != 1 {
		t.Fatalf("open needs %d", got.OpenNeedsCount)
	}

	// Filters: technology, skill of an open need, open_needs, full text (newest first).
	both := other.Id.String() + "," + p.Id.String()
	for query, want := range map[string]string{
		"technology=go":            both,
		"technology=absent":        "",
		"skill=postgresql":         p.Id.String(),
		"open_needs=true":          p.Id.String(),
		"skill=go&open_needs=true": p.Id.String(),
		"q=needs":                  p.Id.String(),
	} {
		if got := strings.Join(projectIDs(decodeAs[apigen.ProjectPage](t, s.read(t, nil, "/api/v1/projects?"+query))), ","); got != want {
			t.Fatalf("%s: %s, want %s", query, got, want)
		}
	}
}

// The GitHub import budget: 10 registrations per user per hour; the 11th is 429 with Retry-After and
// never reaches GitHub.
func TestBuildImportLimit(t *testing.T) {
	gh := &fakeGitHub{}
	s := newBuildStack(t, gh)
	alice := s.signIn(t, "limit-alice", true, 67)
	for i := range 10 {
		s.createProject(t, &alice, fmt.Sprintf("https://github.com/alice/repo-%d", i), "public")
	}
	rec := s.write(t, &alice, http.MethodPost, "/api/v1/projects", projectBody("https://github.com/alice/repo-10", "public"), nil)
	assertProblem(t, rec, http.StatusTooManyRequests, "RATE_LIMITED")
	if rec.Header().Get("Retry-After") == "" || gh.count() != 10 {
		t.Fatalf("retry-after %q, github %d", rec.Header().Get("Retry-After"), gh.count())
	}
}

// Pages are newest first with no duplicate and no gap across cursors.
func TestBuildProjectPagination(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "page-alice", true, 68)
	var want []string
	for i := range 5 {
		want = append([]string{s.createProject(t, &alice, fmt.Sprintf("https://github.com/alice/page-%d", i), "public").Id.String()}, want...)
	}
	var got []string
	path := "/api/v1/projects?limit=2"
	for range 4 {
		page := decodeAs[apigen.ProjectPage](t, s.read(t, nil, path))
		got = append(got, projectIDs(page)...)
		if page.NextCursor == nil {
			break
		}
		path = "/api/v1/projects?limit=2&cursor=" + *page.NextCursor
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("pages %v, want %v", got, want)
	}
}

// The list queries are bounded: each plan, on the SQL the handler really sends, ends in a Limit
// (EXPLAIN logged as evidence).
func TestBuildListPlansAreBounded(t *testing.T) {
	ctx := t.Context()
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "plan-alice", true, 71)
	p := s.createProject(t, &alice, "https://github.com/alice/plans", "public")
	codec, err := cursor.New(s2Secret)
	if err != nil {
		t.Fatal(err)
	}
	rec := &recordingDB{db: s.pool}
	h := build.NewHandler(beginRecorder{rec, s.pool}, &fakeGitHub{}, codec, nil, nil)
	q, tech, skill, open := "plans", "go", "go", true
	cases := []struct {
		name, query string
		call        func(http.ResponseWriter, *http.Request)
	}{
		{"projects, no filter", "ListProjects", func(w http.ResponseWriter, r *http.Request) {
			h.ListProjects(w, r, apigen.ListProjectsParams{})
		}},
		{"related links of a page", "ProjectRelated", func(w http.ResponseWriter, r *http.Request) {
			h.ListProjects(w, r, apigen.ListProjectsParams{})
		}},
		{"projects, q filter", "ListProjects", func(w http.ResponseWriter, r *http.Request) {
			h.ListProjects(w, r, apigen.ListProjectsParams{Q: &q})
		}},
		{"projects, technology, skill and open needs", "ListProjects", func(w http.ResponseWriter, r *http.Request) {
			h.ListProjects(w, r, apigen.ListProjectsParams{Technology: &tech, Skill: &skill, OpenNeeds: &open})
		}},
		{"needs of a project", "ListNeeds", func(w http.ResponseWriter, r *http.Request) {
			h.ListContributionNeeds(w, r, p.Id, apigen.ListContributionNeedsParams{})
		}},
	}
	for _, tc := range cases {
		rec.reset()
		w := httptest.NewRecorder()
		tc.call(w, httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", tc.name, w.Code, w.Body.String())
		}
		stmt := rec.named(t, tc.query)
		plan := explainPlan(ctx, t, s.conn, stmt.sql, stmt.args...)
		t.Logf("build list plan, %s:\n%s", tc.name, plan)
		if !strings.Contains(plan, "Limit") {
			t.Fatalf("%s plan has no Limit:\n%s", tc.name, plan)
		}
	}
}

// named returns the recorded statement generated from the sqlc query called name.
func (r *recordingDB) named(t *testing.T, name string) recordedQuery {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.calls {
		if strings.Contains(c.sql, "-- name: "+name+" ") {
			return c
		}
	}
	t.Fatalf("no %s statement was recorded", name)
	return recordedQuery{}
}
