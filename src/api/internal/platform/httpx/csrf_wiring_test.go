package httpx_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

// The CSRF slot must install the provided middleware, not the pass-through.
func TestRouterInstallsCSRFMiddleware(t *testing.T) {
	var seen bool
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = true
			next.ServeHTTP(w, r)
		})
	}
	h := httpx.NewRouter(slog.New(slog.NewJSONHandler(io.Discard, nil)),
		httpx.Deps{Server: apigen.Unimplemented{}, Ready: okPinger{}, CSRF: mw})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/sign-in",
		bytes.NewReader([]byte(`{"email":"a@b.co","password":"correct-horse"}`)))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if !seen {
		t.Fatal("CSRF middleware not installed")
	}
}
