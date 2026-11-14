package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/moderation"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// The three S6 routes keep the generated 501 fallback until the moderation handler is wired.
func TestModerationRoutesDelegateOnlyWhenWired(t *testing.T) {
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	calls := map[string]func(*Server, http.ResponseWriter, *http.Request){
		"hideContent": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.HideContent(w, r, apigen.ModerationTargetTypeQuestion, id)
		},
		"restoreContent": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.RestoreContent(w, r, apigen.ModerationTargetTypeQuestion, id)
		},
		"resolveErrorReport": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ResolveErrorReport(w, r, id)
		},
	}
	wired := NewServer(WithModeration(moderation.NewHandler(&fakePool{}, allowLimiter{}, nil)))
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			body := `{"reason":"spam"}`
			if name == "resolveErrorReport" {
				body = `{"status":"resolved"}`
			}
			request := func() *http.Request {
				r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				return r
			}
			rec := httptest.NewRecorder()
			call(NewServer(), rec, request())
			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("unwired status %d", rec.Code)
			}
			rec = httptest.NewRecorder()
			call(wired, rec, request())
			if rec.Code == http.StatusNotImplemented {
				t.Fatal("wired moderation route still 501")
			}
		})
	}
}
