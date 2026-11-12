package httpx_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// subjectProbe records the subject its handler sees.
type subjectProbe struct {
	apigen.Unimplemented
	called bool
	got    subject.Subject
}

func (p *subjectProbe) GetSession(w http.ResponseWriter, r *http.Request) {
	p.called = true
	p.got = subject.From(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

type fixedResolver struct {
	sub subject.Subject
	err error
}

func (f fixedResolver) Resolve(*http.Request) (subject.Subject, error) { return f.sub, f.err }

func serveSession(t *testing.T, deps httpx.Deps, mutate func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/auth/session", nil)
	if mutate != nil {
		mutate(req)
	}
	rec := httptest.NewRecorder()
	httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps).ServeHTTP(rec, req)
	return rec
}

func TestSubjectDefaultsToAnonymousEvenWithSpoofedIdentity(t *testing.T) {
	probe := &subjectProbe{}
	rec := serveSession(t, httpx.Deps{Server: probe, Ready: okPinger{}}, func(r *http.Request) {
		r.Header.Set("X-User-Id", "attacker")
		r.AddCookie(&http.Cookie{Name: "user_id", Value: "attacker"})
	})
	if rec.Code != http.StatusNoContent || !probe.called {
		t.Fatalf("status = %d, handler called = %v", rec.Code, probe.called)
	}
	if !probe.got.Anonymous() {
		t.Fatalf("subject = %+v, want anonymous", probe.got)
	}
}

func TestSubjectFromTheConfiguredResolverReachesTheHandler(t *testing.T) {
	probe := &subjectProbe{}
	want := subject.Subject{UserID: "0190c3f2-7a3b-7c1e-8a5d-3b1f2c4d5e6f", EmailVerified: true}
	rec := serveSession(t, httpx.Deps{Server: probe, Ready: okPinger{}, Subjects: fixedResolver{sub: want}}, nil)
	if rec.Code != http.StatusNoContent || probe.got.UserID != want.UserID || !probe.got.EmailVerified {
		t.Fatalf("status = %d, subject = %+v", rec.Code, probe.got)
	}
}

func TestSubjectResolverErrorStopsTheRequestWithAProblem(t *testing.T) {
	probe := &subjectProbe{}
	rec := serveSession(t, httpx.Deps{Server: probe, Ready: okPinger{}, Subjects: fixedResolver{err: errors.New("session store down: dial tcp 10.0.0.7")}}, nil)
	if probe.called {
		t.Fatal("the handler ran although the subject could not be resolved")
	}
	if rec.Code != http.StatusInternalServerError || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("status = %d, content-type = %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if body := rec.Body.String(); strings.Contains(body, "10.0.0.7") || strings.Contains(body, "dial tcp") {
		t.Fatalf("the response leaks the resolver error: %s", body)
	}
}
