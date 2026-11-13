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

func TestMemoComputesOncePerKeyAndCachesErrors(t *testing.T) {
	memo := subject.MemoFrom(subject.WithMemo(context.Background()))
	calls := 0
	boom := errors.New("boom")
	fn := func() (any, error) { calls++; return "v", nil }
	for i := 0; i < 3; i++ {
		if v, err := memo.Do("k", fn); v != "v" || err != nil {
			t.Fatalf("Do = %v, %v", v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("fn ran %d times, want 1", calls)
	}
	failing := 0
	for i := 0; i < 2; i++ {
		if _, err := memo.Do("e", func() (any, error) { failing++; return nil, boom }); !errors.Is(err, boom) {
			t.Fatalf("error not returned: %v", err)
		}
	}
	if failing != 1 {
		t.Fatalf("failing fn ran %d times, want 1", failing)
	}
	if v, _ := memo.Do("other", func() (any, error) { return 2, nil }); v != 2 {
		t.Fatalf("keys are not independent: %v", v)
	}
}

func TestMemoWithoutMemoRunsEveryTime(t *testing.T) {
	memo := subject.MemoFrom(context.Background())
	calls := 0
	for i := 0; i < 2; i++ {
		_, _ = memo.Do("k", func() (any, error) { calls++; return nil, nil })
	}
	if calls != 2 {
		t.Fatalf("a nil memo must not cache: %d calls", calls)
	}
}

func TestMiddlewareSharesOneMemoWithTheResolverAndHandlers(t *testing.T) {
	var fromResolver *subject.Memo
	res := resolverFunc(func(r *http.Request) (subject.Subject, error) {
		fromResolver = subject.MemoFrom(r.Context())
		return subject.Subject{}, nil
	})
	var fromHandler *subject.Memo
	h := subject.Middleware(res, nil)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		fromHandler = subject.MemoFrom(r.Context())
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	if fromResolver == nil || fromResolver != fromHandler {
		t.Fatal("the resolver and the handler must share one per-request memo")
	}
}

type resolverFunc func(*http.Request) (subject.Subject, error)

func (f resolverFunc) Resolve(r *http.Request) (subject.Subject, error) { return f(r) }
