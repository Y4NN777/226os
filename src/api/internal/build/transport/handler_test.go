package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/build/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	projID  = "00000000-0000-7000-8000-0000000000a1"
	aliceID = "00000000-0000-7000-8000-0000000000b1"
	bobID   = "00000000-0000-7000-8000-0000000000b2"
	needID  = "00000000-0000-7000-8000-0000000000e1"
	qID     = "00000000-0000-7000-8000-0000000000d1"
)

var (
	when  = time.Date(2026, 11, 14, 10, 0, 0, 0, time.FixedZone("x", 3600))
	alice = subject.Subject{UserID: aliceID, EmailVerified: true}
	bob   = subject.Subject{UserID: bobID, EmailVerified: true}
)

type fakeStore struct {
	projects []domain.Project
	access   domain.Access
	err      error
	created  []domain.NewProject
	patched  []domain.ProjectPatch
	needs    []domain.Need
	newNeeds []domain.NewNeed
	apps     appFake
}

func (f *fakeStore) URLTaken(context.Context, string) (bool, error) { return false, nil }

func (f *fakeStore) CreateProject(_ context.Context, in domain.NewProject) (string, error) {
	f.created = append(f.created, in)
	return projID, f.err
}

func (f *fakeStore) UpdateProject(_ context.Context, _, _ string, p domain.ProjectPatch, authorize func(domain.Access) error) error {
	if err := authorize(f.access); err != nil {
		return err
	}
	f.patched = append(f.patched, p)
	return f.err
}

func (f *fakeStore) GetProject(context.Context, string, string) (domain.Project, error) {
	if len(f.projects) == 0 {
		return domain.Project{}, domain.ErrNotFound
	}
	return f.projects[0], f.err
}

func (f *fakeStore) ListProjects(context.Context, domain.ProjectQuery) ([]domain.Project, error) {
	return f.projects, f.err
}

func (f *fakeStore) ProjectAccess(context.Context, string) (domain.Access, error) {
	return f.access, f.err
}

func (f *fakeStore) CreateNeed(_ context.Context, in domain.NewNeed, authorize func(domain.Access) error) (domain.Need, error) {
	if err := authorize(f.access); err != nil {
		return domain.Need{}, err
	}
	f.newNeeds = append(f.newNeeds, in)
	return need(), f.err
}

func (f *fakeStore) ListNeeds(context.Context, domain.NeedQuery) ([]domain.Need, error) {
	return f.needs, f.err
}

type repoOK struct{}

func (repoOK) Repo(context.Context, string, string) (domain.Repo, error) {
	return domain.Repo{Name: "repo"}, nil
}

type allowAll struct{}

func (allowAll) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

func handler(t *testing.T, st *fakeStore) *Handler {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return New(service.New(st, repoOK{}, c, allowAll{}), nil)
}

func project() domain.Project {
	return domain.Project{
		ID: projID, Name: "Repo", GitHubURL: "https://github.com/o/repo", Description: "Un outil", Lang: "fr",
		Visibility: "public", Technologies: []string{"go"}, OpenNeeds: 2,
		Owner:       domain.Owner{ID: aliceID, Username: "alice", DisplayName: "Alice"},
		Related:     []domain.Related{{Type: "question", ID: qID, Title: "Q ?", Lang: "fr"}},
		PublishedAt: when, CreatedAt: when, UpdatedAt: when,
	}
}

func need() domain.Need {
	return domain.Need{ID: needID, ProjectID: projID, Title: "Backend Go", Description: "D", Lang: "fr", Status: "open", Skills: []string{"go"}, CreatedAt: when}
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("%v: %s", err, rec.Body.String())
	}
	return v
}

func req(t *testing.T, sub subject.Subject, method, body string) *http.Request {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequestWithContext(subject.With(t.Context(), sub), method, "/", nil)
	} else {
		r = httptest.NewRequestWithContext(subject.With(t.Context(), sub), method, "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	return r
}

const createBody = `{"github_url":"https://github.com/O/Repo","lang":"fr","visibility":"public","technologies":["go"],"links":[{"type":"question","id":"` + qID + `"}]}`

func TestListProjectsEncodesThePage(t *testing.T) {
	rec := httptest.NewRecorder()
	two := []domain.Project{project(), project()}
	handler(t, &fakeStore{projects: two}).ListProjects(rec, req(t, subject.Subject{}, http.MethodGet, ""), apigen.ListProjectsParams{Limit: ptr(1)})
	page := decode[apigen.ProjectPage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || page.NextCursor == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	p := page.Items[0]
	if p.Id.String() != projID || p.Owner.Username != "alice" || p.OpenNeedsCount != 2 || *p.Related[0].Lang != "fr" ||
		p.CreatedAt.Location() != time.UTC || p.Technologies[0] != "go" {
		t.Fatalf("project %+v", p)
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{}).ListProjects(rec, req(t, subject.Subject{}, http.MethodGet, ""), apigen.ListProjectsParams{})
	if !strings.Contains(rec.Body.String(), `"next_cursor":null`) || !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("empty %s", rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{}).ListProjects(rec, req(t, subject.Subject{}, http.MethodGet, ""), apigen.ListProjectsParams{Limit: ptr(0)})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("limit 0: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{}).ListProjects(rec, req(t, subject.Subject{}, http.MethodGet, ""), apigen.ListProjectsParams{Cursor: ptr("forged")})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("forged cursor: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectAnswers201WithLocation(t *testing.T) {
	st := &fakeStore{projects: []domain.Project{project()}}
	rec := httptest.NewRecorder()
	handler(t, st).CreateProject(rec, req(t, alice, http.MethodPost, createBody), apigen.CreateProjectParams{})
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/projects/"+projID {
		t.Fatalf("%d %q %s", rec.Code, rec.Header().Get("Location"), rec.Body.String())
	}
	in := st.created[0]
	if in.OwnerID != aliceID || in.GitHubURL != "https://github.com/o/repo" || len(in.Links) != 1 || in.Links[0].ID != qID || in.Technologies[0] != "go" {
		t.Fatalf("created %+v", in)
	}
}

func TestCreateProjectRejectsAnOwnerInTheBody(t *testing.T) {
	st := &fakeStore{}
	rec := httptest.NewRecorder()
	handler(t, st).CreateProject(rec, req(t, alice, http.MethodPost, `{"github_url":"https://github.com/o/r","lang":"fr","owner_id":"`+bobID+`"}`), apigen.CreateProjectParams{})
	if rec.Code < 400 || rec.Code >= 500 || len(st.created) != 0 {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

// Negative authz for POST /api/v1/projects (createProject, allow: member with a verified email):
// anonymous 401, unverified member 403, nothing created and GitHub never called.
func TestAuthzCreateProjectAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: bobID}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{}
			r := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPost, "/api/v1/projects", strings.NewReader(createBody))
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler(t, st).CreateProject(rec, r, apigen.CreateProjectParams{})
			if rec.Code != tc.status || len(st.created) != 0 {
				t.Fatalf("status %d, created %d; want %d", rec.Code, len(st.created), tc.status)
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	rec := httptest.NewRecorder()
	handler(t, &fakeStore{projects: []domain.Project{project()}}).GetProject(rec, req(t, subject.Subject{}, http.MethodGet, ""), uuid.MustParse(projID))
	if rec.Code != http.StatusOK || decode[apigen.Project](t, rec).Name != "Repo" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{}).GetProject(rec, req(t, subject.Subject{}, http.MethodGet, ""), uuid.MustParse(projID))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("%d", rec.Code)
	}
}

// Negative authz for PATCH /api/v1/projects/{projectId} (updateProject, allow: owner): anonymous 401;
// a non-owner gets 403 on a public project and 404 on a private one (moderators included, C1); a
// project_maintainers row grants nothing (the store only knows the owner); nothing is written.
func TestAuthzUpdateProjectNonOwner(t *testing.T) {
	moderator := subject.Subject{UserID: bobID, EmailVerified: true, Roles: []string{"moderator"}}
	cases := []struct {
		name       string
		sub        subject.Subject
		visibility string
		status     int
	}{
		{"anonymous", subject.Subject{}, "public", http.StatusUnauthorized},
		{"non-owner on a public project", bob, "public", http.StatusForbidden},
		{"non-owner on a private project", bob, "private", http.StatusNotFound},
		{"moderator on a public project", moderator, "public", http.StatusForbidden},
		{"moderator on a private project", moderator, "private", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{projects: []domain.Project{project()}, access: domain.Access{OwnerID: aliceID, Visibility: tc.visibility}}
			r := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPatch, "/api/v1/projects/"+projID, strings.NewReader(`{"description":"pris"}`))
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler(t, st).UpdateProject(rec, r, uuid.MustParse(projID))
			if rec.Code != tc.status || len(st.patched) != 0 {
				t.Fatalf("status %d, patched %d; want %d", rec.Code, len(st.patched), tc.status)
			}
		})
	}
}

func TestUpdateProjectByTheOwner(t *testing.T) {
	st := &fakeStore{projects: []domain.Project{project()}, access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
	rec := httptest.NewRecorder()
	handler(t, st).UpdateProject(rec, req(t, alice, http.MethodPatch, `{"visibility":"private","technologies":["go"],"links":[{"type":"article","id":"`+qID+`"}]}`), uuid.MustParse(projID))
	if rec.Code != http.StatusOK || len(st.patched) != 1 || *st.patched[0].Visibility != "private" || st.patched[0].Links[0].Type != "article" {
		t.Fatalf("%d %s %+v", rec.Code, rec.Body.String(), st.patched)
	}
	rec = httptest.NewRecorder()
	handler(t, st).UpdateProject(rec, req(t, alice, http.MethodPatch, `{"github_url":"https://github.com/x/y"}`), uuid.MustParse(projID))
	if rec.Code < 400 || rec.Code >= 500 || len(st.patched) != 1 {
		t.Fatalf("github_url patch %d", rec.Code)
	}
}

func TestListContributionNeeds(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}, needs: []domain.Need{need()}}
	rec := httptest.NewRecorder()
	handler(t, st).ListContributionNeeds(rec, req(t, subject.Subject{}, http.MethodGet, ""), uuid.MustParse(projID), apigen.ListContributionNeedsParams{})
	page := decode[apigen.NeedPage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || page.Items[0].Status != "open" || page.NextCursor != nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	st.access.Visibility = "private"
	rec = httptest.NewRecorder()
	handler(t, st).ListContributionNeeds(rec, req(t, subject.Subject{}, http.MethodGet, ""), uuid.MustParse(projID), apigen.ListContributionNeedsParams{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("private %d", rec.Code)
	}
}

func TestCreateContributionNeedAnswers201(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}
	rec := httptest.NewRecorder()
	body := `{"title":"Backend Go","description":"Écrire l'API","lang":"fr","skills":["go"]}`
	handler(t, st).CreateContributionNeed(rec, req(t, alice, http.MethodPost, body), uuid.MustParse(projID), apigen.CreateContributionNeedParams{})
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/projects/"+projID+"/contribution-needs" ||
		st.newNeeds[0].CreatedBy != aliceID || st.newNeeds[0].Title != "Backend Go" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, st).CreateContributionNeed(rec, req(t, alice, http.MethodPost, `[`), uuid.MustParse(projID), apigen.CreateContributionNeedParams{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed %d", rec.Code)
	}
}

// Negative authz for POST /api/v1/projects/{projectId}/contribution-needs (createContributionNeed,
// allow: owner, C1): anonymous 401, unverified 403, non-owner 403 on a public project and 404 on a
// private one; nothing is created.
func TestAuthzCreateContributionNeedNonOwner(t *testing.T) {
	cases := []struct {
		name       string
		sub        subject.Subject
		visibility string
		status     int
	}{
		{"anonymous", subject.Subject{}, "public", http.StatusUnauthorized},
		{"unverified member", subject.Subject{UserID: bobID}, "public", http.StatusForbidden},
		{"non-owner on a public project", bob, "public", http.StatusForbidden},
		{"non-owner on a private project", bob, "private", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: tc.visibility}}
			body := `{"title":"Backend Go","description":"D","lang":"fr","skills":["go"]}`
			r := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPost, "/api/v1/projects/"+projID+"/contribution-needs", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler(t, st).CreateContributionNeed(rec, r, uuid.MustParse(projID), apigen.CreateContributionNeedParams{})
			if rec.Code != tc.status || len(st.newNeeds) != 0 {
				t.Fatalf("status %d, created %d; want %d", rec.Code, len(st.newNeeds), tc.status)
			}
		})
	}
}

func TestErrorsNeverLeakTheCause(t *testing.T) {
	rec := httptest.NewRecorder()
	h := handler(t, &fakeStore{err: errors.New("pg: secret detail")})
	h.ListProjects(rec, req(t, subject.Subject{}, http.MethodGet, ""), apigen.ListProjectsParams{})
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.fail(rec, req(t, subject.Subject{}, http.MethodGet, ""), errors.New("plain secret"))
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("plain %d %s", rec.Code, rec.Body.String())
	}
	for name, call := range map[string]func(*httptest.ResponseRecorder){
		"create": func(rec *httptest.ResponseRecorder) {
			h.CreateProject(rec, req(t, alice, http.MethodPost, `[`), apigen.CreateProjectParams{})
		},
		"create service": func(rec *httptest.ResponseRecorder) {
			h.CreateProject(rec, req(t, alice, http.MethodPost, `{"github_url":"https://github.com/o/r","lang":"en"}`), apigen.CreateProjectParams{})
		},
		"update": func(rec *httptest.ResponseRecorder) {
			h.UpdateProject(rec, req(t, alice, http.MethodPatch, `[`), uuid.MustParse(projID))
		},
		"needs": func(rec *httptest.ResponseRecorder) {
			h.ListContributionNeeds(rec, req(t, alice, http.MethodGet, ""), uuid.MustParse(projID), apigen.ListContributionNeedsParams{})
		},
		"need service": func(rec *httptest.ResponseRecorder) {
			h.CreateContributionNeed(rec, req(t, alice, http.MethodPost, `{"title":"T","description":"D","lang":"fr","skills":["go"]}`), uuid.MustParse(projID), apigen.CreateContributionNeedParams{})
		},
	} {
		rec := httptest.NewRecorder()
		call(rec)
		if rec.Code < 400 || strings.Contains(rec.Body.String(), "secret") {
			t.Fatalf("%s: %d %s", name, rec.Code, rec.Body.String())
		}
	}
}

func TestNewHandlerWithNilLogger(t *testing.T) {
	if New(nil, nil).logger == nil {
		t.Fatal("nil logger kept")
	}
}

func ptr[T any](v T) *T { return &v }
