package httpx_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

type bytesBuf = bytes.Buffer

func apiHandler(buf *bytes.Buffer) func(http.ResponseWriter, *http.Request, error) {
	return httpx.APIErrorHandler(slog.New(slog.NewJSONHandler(buf, nil)))
}

func run(t *testing.T, fn http.HandlerFunc) (*httptest.ResponseRecorder, problemBody) {
	t.Helper()
	rec := httptest.NewRecorder()
	middleware.RequestID(fn).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/x", nil))
	var body problemBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	return rec, body
}

func containsAny(s string, subs ...string) bool {
	for _, x := range subs {
		if strings.Contains(s, x) {
			return true
		}
	}
	return false
}
