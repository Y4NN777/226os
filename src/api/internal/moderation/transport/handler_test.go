package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/moderation/domain"
	"github.com/Y4NN777/226os/src/api/internal/moderation/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	modID = "019a0000-0000-7000-8000-000000000001"
	objID = "019a0000-0000-7000-8000-000000000002"
)

type storeFake struct {
	hides, restores, resolves int
	report                    domain.ErrorReport
	requestID                 string
}

func (f *storeFake) Hide(_ context.Context, _, _, _, _, requestID string) error {
	f.hides++
	f.requestID = requestID
	return nil
}
func (f *storeFake) Restore(_ context.Context, _, _, _, _, requestID string) error {
	f.restores++
	f.requestID = requestID
	return nil
}
func (f *storeFake) ResolveReport(_ context.Context, _, _, _, _, requestID string) (domain.ErrorReport, error) {
	f.resolves++
	f.requestID = requestID
	return f.report, nil
}

type allow struct{}

func (allow) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

func request(t *testing.T, sub subject.Subject, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequestWithContext(subject.With(t.Context(), sub), http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestAuthzHideContentNonModerator(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member", subject.Subject{UserID: modID, Roles: []string{"member"}}, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &storeFake{}
			rec := httptest.NewRecorder()
			New(service.New(st, allow{}), nil).HideContent(rec, request(t, tc.sub, `{"reason":"spam"}`), "question", uuid.MustParse(objID))
			if rec.Code != tc.status || st.hides != 0 {
				t.Fatalf("status %d hides %d", rec.Code, st.hides)
			}
		})
	}
}

func TestAuthzRestoreContentNonModerator(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member", subject.Subject{UserID: modID, Roles: []string{"member"}}, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &storeFake{}
			rec := httptest.NewRecorder()
			New(service.New(st, allow{}), nil).RestoreContent(rec, request(t, tc.sub, `{"reason":"appel"}`), "question", uuid.MustParse(objID))
			if rec.Code != tc.status || st.restores != 0 {
				t.Fatalf("status %d restores %d", rec.Code, st.restores)
			}
		})
	}
}

func TestAuthzResolveErrorReportNonModerator(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member", subject.Subject{UserID: modID, Roles: []string{"member"}}, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &storeFake{}
			rec := httptest.NewRecorder()
			New(service.New(st, allow{}), nil).ResolveErrorReport(rec, request(t, tc.sub, `{"status":"resolved"}`), uuid.MustParse(objID))
			if rec.Code != tc.status || st.resolves != 0 {
				t.Fatalf("status %d resolves %d", rec.Code, st.resolves)
			}
		})
	}
}

func TestModeratorActionsAndReporterSafeResolution(t *testing.T) {
	mod := subject.Subject{UserID: modID, Roles: []string{"moderator"}}
	created := time.Date(2026, 11, 14, 17, 0, 0, 0, time.UTC)
	st := &storeFake{report: domain.ErrorReport{ID: objID, ResourceID: modID, Kind: "outdated", Status: "resolved", CreatedAt: created}}
	h := New(service.New(st, allow{}), nil)
	for _, call := range []func(*httptest.ResponseRecorder){
		func(w *httptest.ResponseRecorder) {
			h.HideContent(w, request(t, mod, `{"reason":"spam"}`), apigen.ModerationTargetTypeQuestion, uuid.MustParse(objID))
		},
		func(w *httptest.ResponseRecorder) {
			h.RestoreContent(w, request(t, mod, `{"reason":"appel accepté"}`), apigen.ModerationTargetTypeQuestion, uuid.MustParse(objID))
		},
	} {
		rec := httptest.NewRecorder()
		call(rec)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status %d %s", rec.Code, rec.Body.String())
		}
	}
	rec := httptest.NewRecorder()
	h.ResolveErrorReport(rec, request(t, mod, `{"status":"resolved","note":"corrigé"}`), uuid.MustParse(objID))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "reporter") || strings.Contains(rec.Body.String(), "message") || !strings.Contains(rec.Body.String(), `"status":"resolved"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if st.hides != 1 || st.restores != 1 || st.resolves != 1 {
		t.Fatalf("store %+v", st)
	}
}

func TestModerationAuditRequestIDIsCapped(t *testing.T) {
	mod := subject.Subject{UserID: modID, Roles: []string{"moderator"}}
	st := &storeFake{}
	h := New(service.New(st, allow{}), nil)
	r := request(t, mod, `{"reason":"spam"}`)
	r.Header.Set(middleware.RequestIDHeader, strings.Repeat("x", 65))
	rec := httptest.NewRecorder()

	middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HideContent(w, r, apigen.ModerationTargetTypeQuestion, uuid.MustParse(objID))
	})).ServeHTTP(rec, r)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if st.requestID != strings.Repeat("x", 64) {
		t.Fatalf("audit request id length = %d, want 64", len(st.requestID))
	}
}

func TestMalformedModerationBodyDoesNotWrite(t *testing.T) {
	st := &storeFake{}
	h := New(service.New(st, allow{}), nil)
	mod := subject.Subject{UserID: modID, Roles: []string{"moderator"}}
	for name, call := range map[string]func(*httptest.ResponseRecorder){
		"hide": func(rec *httptest.ResponseRecorder) {
			h.HideContent(rec, request(t, mod, `[`), "question", uuid.MustParse(objID))
		},
		"restore": func(rec *httptest.ResponseRecorder) {
			h.RestoreContent(rec, request(t, mod, `[`), "question", uuid.MustParse(objID))
		},
		"resolve": func(rec *httptest.ResponseRecorder) {
			h.ResolveErrorReport(rec, request(t, mod, `[`), uuid.MustParse(objID))
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			call(rec)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status %d", rec.Code)
			}
		})
	}
	if st.hides != 0 || st.restores != 0 || st.resolves != 0 {
		t.Fatalf("store %+v", st)
	}
}

func TestFailWrapsUnexpectedErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	New(service.New(&storeFake{}, allow{}), nil).fail(rec, request(t, subject.Subject{}, `{}`), errors.New("database detail"))
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "database detail") {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}
