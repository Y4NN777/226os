package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	appID     = "00000000-0000-7000-8000-0000000000f1"
	charlieID = "00000000-0000-7000-8000-0000000000b3"
)

var charlie = subject.Subject{UserID: charlieID, EmailVerified: true}
var moderator = subject.Subject{UserID: charlieID, EmailVerified: true, Roles: []string{"moderator"}}

type appFake struct {
	need    domain.NeedAccess
	app     domain.ApplicationAccess
	created []domain.NewApplication
	decided []string
	list    []domain.Application
}

func (f *fakeStore) CreateApplication(_ context.Context, in domain.NewApplication, authorize func(domain.NeedAccess) error) (string, error) {
	if err := authorize(f.apps.need); err != nil {
		return "", err
	}
	f.apps.created = append(f.apps.created, in)
	return appID, f.err
}

func (f *fakeStore) DecideApplication(_ context.Context, _, _, status string, authorize func(domain.ApplicationAccess) error) error {
	if err := authorize(f.apps.app); err != nil {
		return err
	}
	f.apps.decided = append(f.apps.decided, status)
	return f.err
}

func (f *fakeStore) GetApplication(context.Context, string) (domain.Application, error) {
	return application(), nil
}

func (f *fakeStore) ListApplications(context.Context, domain.ApplicationQuery) ([]domain.Application, error) {
	return f.apps.list, f.err
}

func application() domain.Application {
	at := when
	return domain.Application{
		ID: appID, NeedID: needID, ProjectID: projID, Message: "Je peux aider", Status: "accepted",
		Applicant: domain.Owner{ID: bobID, Username: "bob", DisplayName: "Bob"}, CreatedAt: when, DecidedAt: &at,
	}
}

var (
	openNeed = domain.NeedAccess{NeedStatus: "open", OwnerID: aliceID, Visibility: "public"}
	pending  = domain.ApplicationAccess{ApplicantID: bobID, OwnerID: aliceID, Status: "pending"}
)

func TestCreateApplicationAnswers201WithLocation(t *testing.T) {
	st := &fakeStore{apps: appFake{need: openNeed}}
	rec := httptest.NewRecorder()
	handler(t, st).CreateApplication(rec, req(t, bob, http.MethodPost, `{"message":"Je peux aider"}`), uuid.MustParse(needID), apigen.CreateApplicationParams{})
	a := decode[apigen.Application](t, rec)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/applications/"+appID ||
		a.Applicant.Username != "bob" || a.DecidedAt == nil || a.DecidedAt.Location().String() != "UTC" || st.apps.created[0].UserID != bobID {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, st).CreateApplication(rec, req(t, bob, http.MethodPost, `{"message":"m","user_id":"`+charlieID+`"}`), uuid.MustParse(needID), apigen.CreateApplicationParams{})
	if rec.Code < 400 || rec.Code >= 500 || len(st.apps.created) != 1 {
		t.Fatalf("unknown field %d", rec.Code)
	}
}

// Negative authz for POST /api/v1/contribution-needs/{needId}/applications (createApplication, allow:
// member with a verified email, not the project owner): anonymous 401, unverified 403, a private
// project of someone else 404, the owner on their own project 409; nothing is created.
func TestAuthzCreateApplicationAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		need   domain.NeedAccess
		status int
	}{
		{"anonymous", subject.Subject{}, openNeed, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: bobID}, openNeed, http.StatusForbidden},
		{"need of someone else's private project", bob, domain.NeedAccess{NeedStatus: "open", OwnerID: aliceID, Visibility: "private"}, http.StatusNotFound},
		{"owner on their own project", alice, openNeed, http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{apps: appFake{need: tc.need}}
			r := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPost,
				"/api/v1/contribution-needs/"+needID+"/applications", strings.NewReader(`{"message":"Je peux aider"}`))
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler(t, st).CreateApplication(rec, r, uuid.MustParse(needID), apigen.CreateApplicationParams{})
			if rec.Code != tc.status || len(st.apps.created) != 0 {
				t.Fatalf("status %d, created %d; want %d", rec.Code, len(st.apps.created), tc.status)
			}
		})
	}
}

func TestListProjectApplicationsForTheOwner(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}, apps: appFake{list: []domain.Application{application()}}}
	rec := httptest.NewRecorder()
	pending := apigen.ApplicationStatus("accepted")
	handler(t, st).ListProjectApplications(rec, req(t, alice, http.MethodGet, ""), uuid.MustParse(projID), apigen.ListProjectApplicationsParams{Status: &pending})
	page := decode[apigen.ApplicationPage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || page.Items[0].Message != "Je peux aider" || page.NextCursor != nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

// Negative authz for GET /api/v1/projects/{projectId}/applications (listProjectApplications, allow:
// owner, C1): anonymous 401; a third party or moderator gets 403 on a public project and 404 on a private one, with a
// problem body that carries no item and no count.
func TestAuthzListProjectApplicationsNonOwner(t *testing.T) {
	cases := []struct {
		name       string
		sub        subject.Subject
		visibility string
		status     int
	}{
		{"anonymous", subject.Subject{}, "public", http.StatusUnauthorized},
		{"third party on a public project", charlie, "public", http.StatusForbidden},
		{"third party on a private project", charlie, "private", http.StatusNotFound},
		{"applicant on a public project", bob, "public", http.StatusForbidden},
		{"moderator on a public project", moderator, "public", http.StatusForbidden},
		{"moderator on a private project", moderator, "private", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: tc.visibility}, apps: appFake{list: []domain.Application{application()}}}
			r := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodGet, "/api/v1/projects/"+projID+"/applications", nil)
			rec := httptest.NewRecorder()
			handler(t, st).ListProjectApplications(rec, r, uuid.MustParse(projID), apigen.ListProjectApplicationsParams{})
			body := rec.Body.String()
			if rec.Code != tc.status || strings.Contains(body, "items") || strings.Contains(body, "Je peux aider") {
				t.Fatalf("status %d %s; want %d and no data", rec.Code, body, tc.status)
			}
		})
	}
}

func TestDecideApplicationByTheOwner(t *testing.T) {
	st := &fakeStore{apps: appFake{app: pending}}
	rec := httptest.NewRecorder()
	handler(t, st).DecideApplication(rec, req(t, alice, http.MethodPatch, `{"status":"accepted","note":"Bienvenue"}`), uuid.MustParse(appID))
	if rec.Code != http.StatusOK || decode[apigen.Application](t, rec).Status != "accepted" || st.apps.decided[0] != "accepted" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, st).DecideApplication(rec, req(t, alice, http.MethodPatch, `[`), uuid.MustParse(appID))
	if rec.Code != http.StatusBadRequest || len(st.apps.decided) != 1 {
		t.Fatalf("malformed %d", rec.Code)
	}
}

// Negative authz for PATCH /api/v1/applications/{applicationId} (decideApplication, allow: project
// owner accepts or declines, applicant withdraws): anonymous 401; a third party 404 (cross-project
// IDOR); a moderator who is not a party gets the same 404; the applicant accepting 403; the owner withdrawing 403; nothing changes.
func TestAuthzDecideApplicationNonParty(t *testing.T) {
	cases := []struct {
		name, decision string
		sub            subject.Subject
		status         int
	}{
		{"anonymous", "accepted", subject.Subject{}, http.StatusUnauthorized},
		{"third party accepts", "accepted", charlie, http.StatusNotFound},
		{"third party withdraws", "withdrawn", charlie, http.StatusNotFound},
		{"moderator accepts", "accepted", moderator, http.StatusNotFound},
		{"applicant accepts their own", "accepted", bob, http.StatusForbidden},
		{"owner withdraws", "withdrawn", alice, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{apps: appFake{app: pending}}
			r := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPatch, "/api/v1/applications/"+appID,
				strings.NewReader(`{"status":"`+tc.decision+`"}`))
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler(t, st).DecideApplication(rec, r, uuid.MustParse(appID))
			if rec.Code != tc.status || len(st.apps.decided) != 0 {
				t.Fatalf("status %d, decided %v; want %d", rec.Code, st.apps.decided, tc.status)
			}
		})
	}
}

func TestApplicationErrorsAndPendingHasNoDecisionTime(t *testing.T) {
	rec := httptest.NewRecorder()
	handler(t, &fakeStore{}).CreateApplication(rec, req(t, bob, http.MethodPost, `[`), uuid.MustParse(needID), apigen.CreateApplicationParams{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed %d", rec.Code)
	}
	a := application()
	a.DecidedAt, a.Status = nil, "pending"
	if j := applicationJSON(a); j.DecidedAt != nil {
		t.Fatalf("pending decided_at %v", j.DecidedAt)
	}
}
