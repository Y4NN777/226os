package transport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

type fakeReports struct {
	in  []domain.NewReport
	out domain.ErrorReport
	err error
}

func (f *fakeReports) CreateReport(_ context.Context, in domain.NewReport) (domain.ErrorReport, error) {
	f.in = append(f.in, in)
	return f.out, f.err
}

type allowAll struct{}

func (allowAll) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

const (
	repResource = "00000000-0000-7000-8000-0000000000a1"
	repSource   = "00000000-0000-7000-8000-0000000000b1"
	repUser     = "00000000-0000-7000-8000-0000000000d1"
	repID       = "00000000-0000-7000-8000-0000000000e1"
)

func reportHandler(store *fakeReports) *Handler {
	return New(nil, service.NewReports(store, allowAll{}), nil)
}

func post(t *testing.T, h *Handler, sub subject.Subject, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(subject.With(t.Context(), sub), http.MethodPost, "/api/v1/resources/"+repResource+"/error-reports", strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.CreateResourceErrorReport(rec, req, uuid.MustParse(repResource))
	return rec
}

func verifiedMember() subject.Subject { return subject.Subject{UserID: repUser, EmailVerified: true} }

func TestCreateResourceErrorReportAnswers201WithoutAReporter(t *testing.T) {
	created := time.Date(2026, 11, 13, 9, 30, 0, 0, time.FixedZone("x", 3600))
	store := &fakeReports{out: domain.ErrorReport{ID: repID, ResourceID: repResource, Kind: "outdated", Status: "received", CreatedAt: created}}
	rec := post(t, reportHandler(store), verifiedMember(), "application/json",
		`{"kind":"outdated","body":"moved","source_id":"`+repSource+`"}`)
	if rec.Code != http.StatusCreated || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status %d type %s body %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
	want := `{"created_at":"2026-11-13T08:30:00Z","id":"` + repID + `","kind":"outdated","resource_id":"` + repResource + `","status":"received"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("body %s\nwant %s", rec.Body.String(), want)
	}
	if strings.Contains(rec.Body.String(), "reporter") || strings.Contains(rec.Body.String(), repUser) {
		t.Fatalf("the response exposes the reporter: %s", rec.Body.String())
	}
	got := store.in[0]
	if got.ReporterID != repUser || got.ResourceID != repResource || got.SourceID != repSource || got.Body != "moved" || got.Kind != "outdated" {
		t.Fatalf("stored %+v", got)
	}
}

func TestCreateResourceErrorReportWithoutSource(t *testing.T) {
	store := &fakeReports{out: domain.ErrorReport{ID: repID, ResourceID: repResource, Kind: "other", Status: "received"}}
	rec := post(t, reportHandler(store), verifiedMember(), "application/json", `{"kind":"other","body":"x"}`)
	if rec.Code != http.StatusCreated || store.in[0].SourceID != "" {
		t.Fatalf("status %d stored %+v", rec.Code, store.in)
	}
}

func TestCreateResourceErrorReportErrors(t *testing.T) {
	cases := []struct {
		name        string
		sub         subject.Subject
		contentType string
		body        string
		store       *fakeReports
		status      int
		code        string
	}{
		{"anonymous", subject.Subject{}, "application/json", `{"kind":"other","body":"x"}`, &fakeReports{}, 401, "UNAUTHENTICATED"},
		{"unverified member", subject.Subject{UserID: repUser}, "application/json", `{"kind":"other","body":"x"}`, &fakeReports{}, 403, "EMAIL_NOT_VERIFIED"},
		{"unknown field", verifiedMember(), "application/json", `{"kind":"other","body":"x","reporter_id":"` + repUser + `"}`, &fakeReports{}, 422, "VALIDATION_ERROR"},
		{"not JSON", verifiedMember(), "application/json", `{`, &fakeReports{}, 400, "MALFORMED_REQUEST"},
		{"wrong content type", verifiedMember(), "text/plain", `{}`, &fakeReports{}, 422, "VALIDATION_ERROR"},
		{"invalid kind", verifiedMember(), "application/json", `{"kind":"spam","body":"x"}`, &fakeReports{}, 422, "VALIDATION_ERROR"},
		{"resource not visible", verifiedMember(), "application/json", `{"kind":"other","body":"x"}`, &fakeReports{err: domain.ErrNotFound}, 404, "NOT_FOUND"},
		{"source of another resource", verifiedMember(), "application/json", `{"kind":"other","body":"x"}`, &fakeReports{err: domain.ErrSourceMismatch}, 422, "VALIDATION_ERROR"},
		{"database failure", verifiedMember(), "application/json", `{"kind":"other","body":"x"}`, &fakeReports{err: errors.New("select * from secrets")}, 500, "INTERNAL_ERROR"},
		{"store returns a malformed id", verifiedMember(), "application/json", `{"kind":"other","body":"x"}`, &fakeReports{out: domain.ErrorReport{ID: "bad"}}, 500, "INTERNAL_ERROR"},
		{"store returns a malformed resource id", verifiedMember(), "application/json", `{"kind":"other","body":"x"}`, &fakeReports{out: domain.ErrorReport{ID: repID, ResourceID: "bad"}}, 500, "INTERNAL_ERROR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := post(t, reportHandler(tc.store), tc.sub, tc.contentType, tc.body)
			if rec.Code != tc.status || rec.Header().Get("Content-Type") != "application/problem+json" ||
				!strings.Contains(rec.Body.String(), `"`+tc.code+`"`) || strings.Contains(rec.Body.String(), "secrets") {
				t.Fatalf("status %d type %s body %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
			}
		})
	}
}

// Negative authz for POST /api/v1/resources/{resourceId}/error-reports (AUTHORIZATION_MATRIX,
// allow: member with a verified email): anonymous callers get 401 and unverified members 403,
// and nothing is stored.
func TestAuthzCreateResourceErrorReportAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: repUser}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeReports{}
			path := fmt.Sprintf("/api/v1/resources/%s/error-reports", repResource)
			req := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPost, path, strings.NewReader(`{"kind":"other","body":"x"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			reportHandler(store).CreateResourceErrorReport(rec, req, uuid.MustParse(repResource))
			if rec.Code != tc.status || len(store.in) != 0 {
				t.Fatalf("status %d, stored %d; want %d and nothing stored", rec.Code, len(store.in), tc.status)
			}
		})
	}
}
