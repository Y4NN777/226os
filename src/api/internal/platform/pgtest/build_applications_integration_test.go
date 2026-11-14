//go:build integration

package pgtest

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// US-BUILD-001 applications on the production router over a migrated database: a member applies to a
// need, the project owner reads and decides, the applicant withdraws and re-applies; second clicks,
// third parties, cross-project ids and races never create, leak or change anything.

const applyBody = `{"message":"Je connais Go et PostgreSQL."}`

func (s *s2Stack) need(t *testing.T, owner *member, project apigen.Project) apigen.ContributionNeed {
	t.Helper()
	rec := s.write(t, owner, http.MethodPost, "/api/v1/projects/"+project.Id.String()+"/contribution-needs",
		`{"title":"Backend Go","description":"Écrire l'API.","lang":"fr","skills":["go"]}`, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("need %d %s", rec.Code, rec.Body.String())
	}
	return decodeAs[apigen.ContributionNeed](t, rec)
}

func (s *s2Stack) apply(t *testing.T, m *member, need apigen.ContributionNeed, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return s.write(t, m, http.MethodPost, "/api/v1/contribution-needs/"+need.Id.String()+"/applications", applyBody, headers)
}

func (s *s2Stack) decide(t *testing.T, m *member, id, status string) *httptest.ResponseRecorder {
	t.Helper()
	return s.write(t, m, http.MethodPatch, "/api/v1/applications/"+id, `{"status":"`+status+`"}`, nil)
}

func (s *s2Stack) applicationRows(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM applications`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (s *s2Stack) applicationStatus(t *testing.T, id string) string {
	t.Helper()
	var st string
	if err := s.conn.QueryRow(t.Context(), `SELECT status FROM applications WHERE id = $1`, id).Scan(&st); err != nil {
		t.Fatal(err)
	}
	return st
}

// US-BUILD-001/1: Bob applies, Alice reads the application, a second click creates nothing.
func TestBuildApplicationLifecycle(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "app-alice", true, 72)
	bob := s.signIn(t, "app-bob", true, 73)
	unverified := s.signIn(t, "app-unverified", false, 74)
	p := s.createProject(t, &alice, "https://github.com/alice/collab", "public")
	n := s.need(t, &alice, p)

	assertProblem(t, s.apply(t, nil, n, nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	assertProblem(t, s.apply(t, &unverified, n, nil), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
	assertProblem(t, s.apply(t, &alice, n, nil), http.StatusConflict, "STATE_CONFLICT") // own project
	assertProblem(t, s.write(t, &bob, http.MethodPost, "/api/v1/contribution-needs/"+n.Id.String()+"/applications", `{"message":""}`, nil),
		http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if got := s.applicationRows(t); got != 0 {
		t.Fatalf("%d applications after refused writes", got)
	}

	key := map[string]string{"Idempotency-Key": "build-apply-key-000001"}
	rec := s.apply(t, &bob, n, key)
	a := decodeAs[apigen.Application](t, rec)
	if rec.Code != http.StatusCreated || a.Status != "pending" || a.Applicant.Id.String() != bob.id || a.NeedId != n.Id ||
		a.ProjectId != p.Id || a.DecidedAt != nil || rec.Header().Get("Location") != "/api/v1/applications/"+a.Id.String() {
		t.Fatalf("apply %d %s", rec.Code, rec.Body.String())
	}
	// The same key replays the first 201; a second distinct click is a 409. One row either way.
	if rec := s.apply(t, &bob, n, key); rec.Code != http.StatusCreated || rec.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("replay %d %v", rec.Code, rec.Header())
	}
	assertProblem(t, s.apply(t, &bob, n, nil), http.StatusConflict, "DUPLICATE_APPLICATION")
	if got := s.applicationRows(t); got != 1 {
		t.Fatalf("%d applications", got)
	}

	// Alice reads it; the status filter works.
	page := decodeAs[apigen.ApplicationPage](t, s.read(t, &alice, "/api/v1/projects/"+p.Id.String()+"/applications"))
	if len(page.Items) != 1 || page.Items[0].Id != a.Id || page.Items[0].Message != "Je connais Go et PostgreSQL." {
		t.Fatalf("owner list %+v", page)
	}
	if page := decodeAs[apigen.ApplicationPage](t, s.read(t, &alice, "/api/v1/projects/"+p.Id.String()+"/applications?status=accepted")); len(page.Items) != 0 {
		t.Fatalf("status filter %+v", page)
	}

	// Bob withdraws, then re-applies: the same row is reopened.
	if rec := s.decide(t, &bob, a.Id.String(), "withdrawn"); rec.Code != http.StatusOK || decodeAs[apigen.Application](t, rec).Status != "withdrawn" {
		t.Fatalf("withdraw %d %s", rec.Code, rec.Body.String())
	}
	rec = s.apply(t, &bob, n, nil)
	if rec.Code != http.StatusCreated || decodeAs[apigen.Application](t, rec).Id != a.Id || s.applicationStatus(t, a.Id.String()) != "pending" {
		t.Fatalf("re-apply %d %s", rec.Code, rec.Body.String())
	}
	// Alice declines: terminal. No re-apply after a decline, no second decision.
	rec = s.decide(t, &alice, a.Id.String(), "declined")
	if got := decodeAs[apigen.Application](t, rec); rec.Code != http.StatusOK || got.Status != "declined" || got.DecidedAt == nil {
		t.Fatalf("decline %d %s", rec.Code, rec.Body.String())
	}
	assertProblem(t, s.apply(t, &bob, n, nil), http.StatusConflict, "DUPLICATE_APPLICATION")
	assertProblem(t, s.decide(t, &alice, a.Id.String(), "accepted"), http.StatusConflict, "STATE_CONFLICT")
	assertProblem(t, s.decide(t, &bob, a.Id.String(), "withdrawn"), http.StatusConflict, "STATE_CONFLICT")
	if got := s.applicationStatus(t, a.Id.String()); got != "declined" || s.applicationRows(t) != 1 {
		t.Fatalf("status %s, %d rows", got, s.applicationRows(t))
	}

	// A closed need takes no application.
	if _, err := s.conn.Exec(t.Context(), `UPDATE contribution_needs SET status = 'closed', closed_at = now() WHERE id = $1`, n.Id.String()); err != nil {
		t.Fatal(err)
	}
	carol := s.signIn(t, "app-carol", true, 75)
	assertProblem(t, s.apply(t, &carol, n, nil), http.StatusConflict, "STATE_CONFLICT")
}

// A double click without an Idempotency-Key: two concurrent requests, one 201, one 409, one row.
func TestBuildApplicationDoubleClick(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "dbl-alice", true, 76)
	bob := s.signIn(t, "dbl-bob", true, 77)
	n := s.need(t, &alice, s.createProject(t, &alice, "https://github.com/alice/double", "public"))
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Go(func() { codes[i] = s.apply(t, &bob, n, nil).Code })
	}
	wg.Wait()
	slices.Sort(codes)
	if codes[0] != http.StatusCreated || codes[1] != http.StatusConflict || s.applicationRows(t) != 1 {
		t.Fatalf("codes %v, %d rows", codes, s.applicationRows(t))
	}
}

// US-BUILD-001/2: Charlie is not the owner: 403 if the project is visible to him, 404 if private, with
// no item and no count; a maintainer row grants nothing; a third party or another project's owner
// deciding gets 404; nothing changes.
func TestBuildApplicationsPrivacyAndIDOR(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "idor-alice", true, 78)
	bob := s.signIn(t, "idor-bob", true, 79)
	charlie := s.signIn(t, "idor-charlie", true, 80)
	dave := s.signIn(t, "idor-dave", true, 81)
	pub := s.createProject(t, &alice, "https://github.com/alice/open", "public")
	priv := s.createProject(t, &alice, "https://github.com/alice/closed", "private")
	pubNeed, privNeed := s.need(t, &alice, pub), s.need(t, &alice, priv)
	app := decodeAs[apigen.Application](t, s.apply(t, &bob, pubNeed, nil))
	if _, err := s.conn.Exec(t.Context(), `INSERT INTO project_maintainers (project_id, user_id, added_by) VALUES ($1, $2, $3)`,
		pub.Id.String(), charlie.id, alice.id); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		m      *member
		path   string
		status int
	}{
		{nil, pub.Id.String(), http.StatusUnauthorized},
		{&charlie, pub.Id.String(), http.StatusForbidden},
		{&bob, pub.Id.String(), http.StatusForbidden},
		{&charlie, priv.Id.String(), http.StatusNotFound},
	} {
		rec := s.read(t, tc.m, "/api/v1/projects/"+tc.path+"/applications")
		if rec.Code != tc.status || strings.Contains(rec.Body.String(), "items") || strings.Contains(rec.Body.String(), "Je connais") {
			t.Fatalf("list %s: %d %s", tc.path, rec.Code, rec.Body.String())
		}
	}
	// A need of someone else's private project is 404 to an applicant.
	assertProblem(t, s.apply(t, &bob, privNeed, nil), http.StatusNotFound, "NOT_FOUND")

	// Deciding: third party 404, the applicant cannot accept, the owner cannot withdraw.
	assertProblem(t, s.decide(t, &charlie, app.Id.String(), "accepted"), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.decide(t, &charlie, app.Id.String(), "withdrawn"), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.decide(t, &bob, app.Id.String(), "accepted"), http.StatusForbidden, "FORBIDDEN")
	assertProblem(t, s.decide(t, &alice, app.Id.String(), "withdrawn"), http.StatusForbidden, "FORBIDDEN")
	// Cross-project: Dave owns another project; Alice's application id means nothing to him.
	davesNeed := s.need(t, &dave, s.createProject(t, &dave, "https://github.com/dave/tool", "public"))
	davesApp := decodeAs[apigen.Application](t, s.apply(t, &bob, davesNeed, nil))
	assertProblem(t, s.decide(t, &dave, app.Id.String(), "accepted"), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.decide(t, &alice, davesApp.Id.String(), "accepted"), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.decide(t, nil, app.Id.String(), "accepted"), http.StatusUnauthorized, "UNAUTHENTICATED")
	if s.applicationStatus(t, app.Id.String()) != "pending" || s.applicationStatus(t, davesApp.Id.String()) != "pending" {
		t.Fatal("a refused decision changed an application")
	}
	// Dave's list only shows Dave's project.
	page := decodeAs[apigen.ApplicationPage](t, s.read(t, &dave, "/api/v1/projects/"+pub.Id.String()+"/applications"))
	if len(page.Items) != 0 {
		t.Fatalf("dave sees alice's applications: %+v", page)
	}
}

// The owner accepting and the applicant withdrawing at the same time: exactly one wins, the other 409.
func TestBuildConcurrentDecisions(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "race2-alice", true, 82)
	bob := s.signIn(t, "race2-bob", true, 83)
	n := s.need(t, &alice, s.createProject(t, &alice, "https://github.com/alice/race", "public"))
	app := decodeAs[apigen.Application](t, s.apply(t, &bob, n, nil))
	codes := make([]int, 2)
	var wg sync.WaitGroup
	wg.Go(func() { codes[0] = s.decide(t, &alice, app.Id.String(), "accepted").Code })
	wg.Go(func() { codes[1] = s.decide(t, &bob, app.Id.String(), "withdrawn").Code })
	wg.Wait()
	sorted := slices.Clone(codes)
	slices.Sort(sorted)
	final := s.applicationStatus(t, app.Id.String())
	if sorted[0] != http.StatusOK || sorted[1] != http.StatusConflict ||
		(codes[0] == http.StatusOK) != (final == "accepted") || (codes[1] == http.StatusOK) != (final == "withdrawn") {
		t.Fatalf("codes %v, final %s", codes, final)
	}
}

// Issue #181 (note-apply-state-lost): the needs of a project carry the caller's own application (id and
// status), so a reload still knows who applied. It comes from the session only: another applicant's id or
// status never appears, anonymous and non-applicants get null, and the answer varies by cookie.
func TestBuildNeedsCarryTheCallersOwnApplicationOnly(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "viewer-app-alice", true, 75)
	bob := s.signIn(t, "viewer-app-bob", true, 76)
	carol := s.signIn(t, "viewer-app-carol", true, 77)
	dave := s.signIn(t, "viewer-app-dave", true, 78)
	p := s.createProject(t, &alice, "https://github.com/alice/viewer-app", "public")
	n := s.need(t, &alice, p)
	if n.ViewerApplication != nil {
		t.Fatalf("a created need carries %+v, want null", n.ViewerApplication)
	}
	path := "/api/v1/projects/" + p.Id.String() + "/contribution-needs"

	bobApp := decodeAs[apigen.Application](t, s.apply(t, &bob, n, nil))
	carolApp := decodeAs[apigen.Application](t, s.apply(t, &carol, n, nil))
	if rec := s.decide(t, &carol, carolApp.Id.String(), "withdrawn"); rec.Code != http.StatusOK {
		t.Fatalf("withdraw %d %s", rec.Code, rec.Body.String())
	}

	check := func(who string, m *member, want *apigen.ViewerApplication) {
		t.Helper()
		rec := s.read(t, m, path)
		page := decodeAs[apigen.NeedPage](t, rec)
		if len(page.Items) != 1 {
			t.Fatalf("%s: %d needs", who, len(page.Items))
		}
		got := page.Items[0].ViewerApplication
		if (got == nil) != (want == nil) || (got != nil && *got != *want) {
			t.Errorf("%s: viewer_application = %+v, want %+v", who, got, want)
		}
		if !strings.Contains(rec.Header().Get("Vary"), "Cookie") {
			t.Errorf("%s: Vary = %q", who, rec.Header().Get("Vary"))
		}
		if strings.Contains(rec.Body.String(), bobApp.Id.String()) && who != "bob" || strings.Contains(rec.Body.String(), carolApp.Id.String()) && who != "carol" {
			t.Errorf("%s: another applicant's application id is in the response", who)
		}
	}
	check("bob", &bob, &apigen.ViewerApplication{Id: bobApp.Id, Status: "pending"})
	check("carol", &carol, &apigen.ViewerApplication{Id: carolApp.Id, Status: "withdrawn"})
	check("dave", &dave, nil)
	check("owner", &alice, nil)
	check("anonymous", nil, nil)

	// The owner's decision shows in the applicant's own state, and only there.
	if rec := s.decide(t, &alice, bobApp.Id.String(), "accepted"); rec.Code != http.StatusOK {
		t.Fatalf("accept %d %s", rec.Code, rec.Body.String())
	}
	check("bob", &bob, &apigen.ViewerApplication{Id: bobApp.Id, Status: "accepted"})
	check("dave", &dave, nil)
}
