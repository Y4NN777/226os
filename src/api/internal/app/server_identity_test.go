package app_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// stubStore fails every call: enough to prove the Server delegates to the
// identity handler (responses differ from the 501 fallback).
type stubStore struct{}

var errUnused = errors.New("unused store")

func (stubStore) CreateAccount(context.Context, service.NewAccount) (domain.Account, error) {
	return domain.Account{}, errUnused
}
func (stubStore) AccountByEmail(context.Context, string) (domain.Account, error) {
	return domain.Account{}, errUnused
}
func (stubStore) AccountByID(context.Context, string) (domain.Account, error) {
	return domain.Account{}, errUnused
}
func (stubStore) CreateSession(context.Context, domain.Session) (string, error) {
	return "", errUnused
}
func (stubStore) SessionByTokenHash(context.Context, [32]byte) (domain.Session, error) {
	return domain.Session{}, errUnused
}
func (stubStore) RevokeSession(context.Context, string, string, time.Time) error {
	return errUnused
}
func (stubStore) RevokeUserSessions(context.Context, string, string, time.Time, string) error {
	return errUnused
}
func (stubStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	return errUnused
}
func (stubStore) CreateEmailVerification(context.Context, string, string, [32]byte, time.Time) error {
	return errUnused
}
func (stubStore) ConsumeEmailVerification(context.Context, [32]byte, time.Time) (string, error) {
	return "", errUnused
}
func (stubStore) MarkEmailVerified(context.Context, string, time.Time) error { return errUnused }
func (stubStore) CreatePasswordReset(context.Context, string, [32]byte, time.Time) error {
	return errUnused
}
func (stubStore) ResetPassword(context.Context, [32]byte, string, time.Time) error {
	return errUnused
}
func (stubStore) Hit(context.Context, string, [32]byte, int, time.Duration, time.Time) (bool, time.Duration, error) {
	return false, 0, errUnused
}

func identityServer(t *testing.T) *app.Server {
	t.Helper()
	svc := service.New(stubStore{}, nil, []byte("0123456789abcdef0123456789abcdef"))
	return app.NewServer(app.WithIdentity(identitytransport.New(svc, false, nil)))
}

func postJSON(t *testing.T, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/x", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Without WithIdentity every auth operation stays 501.
func TestServerIdentityUnimplementedWithoutOption(t *testing.T) {
	s := app.NewServer()
	calls := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"GetCsrfToken", s.GetCsrfToken},
		{"SignUp", s.SignUp},
		{"SignIn", s.SignIn},
		{"SignOut", s.SignOut},
		{"GetSession", s.GetSession},
		{"VerifyEmail", s.VerifyEmail},
		{"RequestPasswordReset", s.RequestPasswordReset},
		{"ConfirmPasswordReset", s.ConfirmPasswordReset},
	}
	for _, c := range calls {
		rec := httptest.NewRecorder()
		c.fn(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
		if rec.Code != http.StatusNotImplemented {
			t.Fatalf("%s: status = %d", c.name, rec.Code)
		}
	}
}

// With WithIdentity the same operations reach the real handler.
func TestServerIdentityDelegates(t *testing.T) {
	s := identityServer(t)

	// GET /auth/csrf anonymous -> 200 with a token (no store call needed).
	rec := httptest.NewRecorder()
	s.GetCsrfToken(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "csrf_token") {
		t.Fatalf("csrf: %d %s", rec.Code, rec.Body.String())
	}

	// GET /auth/session anonymous -> 200 authenticated:false.
	rec = httptest.NewRecorder()
	s.GetSession(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"authenticated":false`) {
		t.Fatalf("session: %d %s", rec.Code, rec.Body.String())
	}

	// Sign-up reaches the service: the stub store fails the rate-limit hit
	// and the handler answers a problem, never 501.
	rec = httptest.NewRecorder()
	s.SignUp(rec, postJSON(t, `{"email":"a@b.co","password":"correct-horse","username":"ab"}`))
	if rec.Code == http.StatusNotImplemented {
		t.Fatal("sign-up stayed unimplemented")
	}

	// Sign-in with bad credentials -> 401 INVALID_CREDENTIALS shape.
	rec = httptest.NewRecorder()
	s.SignIn(rec, postJSON(t, `{"email":"a@b.co","password":"correct-horse"}`))
	if rec.Code == http.StatusNotImplemented {
		t.Fatal("sign-in stayed unimplemented")
	}

	// Sign-out anonymous -> 401 (subject check runs before the service).
	rec = httptest.NewRecorder()
	s.SignOut(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("sign-out: %d", rec.Code)
	}

	// Verify-email with a malformed token -> 422 INVALID_TOKEN.
	rec = httptest.NewRecorder()
	s.VerifyEmail(rec, postJSON(t, `{"token":"bad"}`))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("verify: %d %s", rec.Code, rec.Body.String())
	}

	// Reset request reaches the service (stub store -> problem, not 501).
	rec = httptest.NewRecorder()
	s.RequestPasswordReset(rec, postJSON(t, `{"email":"a@b.co"}`))
	if rec.Code == http.StatusNotImplemented {
		t.Fatal("reset request stayed unimplemented")
	}

	// Reset confirm with a malformed token -> 422.
	rec = httptest.NewRecorder()
	s.ConfirmPasswordReset(rec, postJSON(t, `{"token":"bad","new_password":"new-correct-horse"}`))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("confirm: %d %s", rec.Code, rec.Body.String())
	}
}

var _ apigen.ServerInterface = (*app.Server)(nil)
