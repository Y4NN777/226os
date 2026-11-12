package subject_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

type fixedResolver struct {
	s   subject.Subject
	err error
}

func (f fixedResolver) Resolve(*http.Request) (subject.Subject, error) { return f.s, f.err }

func TestAnonymous(t *testing.T) {
	if !(subject.Subject{}).Anonymous() {
		t.Fatal("zero Subject must be anonymous")
	}
	if (subject.Subject{UserID: "u1"}).Anonymous() {
		t.Fatal("Subject with a UserID must not be anonymous")
	}
}

func TestFromWithoutSubjectIsAnonymous(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	if !subject.From(r.Context()).Anonymous() {
		t.Fatal("From on a bare context must be anonymous")
	}
}

func TestWithFromRoundTrip(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	want := subject.Subject{UserID: "u1", EmailVerified: true, Roles: []string{"member"}}
	got := subject.From(subject.With(r.Context(), want))
	if got.UserID != "u1" || !got.EmailVerified || len(got.Roles) != 1 || got.Roles[0] != "member" {
		t.Fatalf("got %+v", got)
	}
}

func TestMiddlewareStoresResolvedSubject(t *testing.T) {
	want := subject.Subject{UserID: "u1"}
	var got subject.Subject
	h := subject.Middleware(fixedResolver{s: want}, nil)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = subject.From(r.Context())
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if got.UserID != "u1" {
		t.Fatalf("got %+v", got)
	}
}

func TestMiddlewareResolverErrorStopsChain(t *testing.T) {
	boom := errors.New("boom")
	var gotErr error
	nextCalled := false
	onErr := func(w http.ResponseWriter, _ *http.Request, err error) {
		gotErr = err
		w.WriteHeader(http.StatusInternalServerError)
	}
	h := subject.Middleware(fixedResolver{err: boom}, onErr)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if nextCalled || !errors.Is(gotErr, boom) || rec.Code != http.StatusInternalServerError {
		t.Fatalf("next=%v err=%v code=%d", nextCalled, gotErr, rec.Code)
	}
}

func TestNilResolverMeansAnonymous(t *testing.T) {
	var anon bool
	h := subject.Middleware(nil, nil)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		anon = subject.From(r.Context()).Anonymous()
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if !anon {
		t.Fatal("nil resolver must yield anonymous")
	}
}

// The subject comes from the Resolver only: client-controlled inputs never
// influence the default resolver (docs/spec-kit/AUTHORIZATION_MATRIX.md).
func TestDefaultResolverIgnoresClientSuppliedIdentity(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?user_id=attacker&UserID=attacker", nil)
	r.Header.Set("X-User-Id", "attacker")
	r.Header.Set("Authorization", "Bearer attacker")
	r.AddCookie(&http.Cookie{Name: "Subject", Value: `{"UserID":"attacker"}`})
	r.AddCookie(&http.Cookie{Name: "subject", Value: "attacker"})

	s, err := subject.AnonymousResolver{}.Resolve(r)
	if err != nil || !s.Anonymous() {
		t.Fatalf("Resolve = %+v, %v", s, err)
	}

	var got subject.Subject
	h := subject.Middleware(nil, nil)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = subject.From(r.Context())
	}))
	h.ServeHTTP(httptest.NewRecorder(), r)
	if !got.Anonymous() {
		t.Fatalf("middleware subject = %+v, want anonymous", got)
	}
}

func TestMiddlewareNilOnErrorFallsBackTo500(t *testing.T) {
	nextCalled := false
	h := subject.Middleware(fixedResolver{err: errors.New("boom")}, nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if nextCalled || rec.Code != http.StatusInternalServerError {
		t.Fatalf("next=%v code=%d", nextCalled, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q, want application/problem+json", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	if body["code"] != "INTERNAL_ERROR" || body["status"] != float64(http.StatusInternalServerError) {
		t.Fatalf("body = %v", body)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("fallback body echoes the resolver error: %q", rec.Body.String())
	}
}
