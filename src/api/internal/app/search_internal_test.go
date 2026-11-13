package app

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/search"
)

func TestSearchStaysUnimplementedWithoutAHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	NewServer().Search(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), apigen.SearchParams{Q: "ab"})
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestSearchDelegatesWhenSet(t *testing.T) {
	s := NewServer(WithSearch(search.NewHandler(&fakePool{}, allowLimiter{}, nil)))
	rec := httptest.NewRecorder()
	// q is one rune: the real handler answers 422, never the 501 of the unwired route.
	s.Search(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), apigen.SearchParams{Q: "a"})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want the search handler's 422", rec.Code)
	}
}

// Through the real router, the contract validation answers before the handler runs.
func TestSearchRouteStatuses(t *testing.T) {
	routed := httpx.NewRouter(slog.New(slog.DiscardHandler), newDeps(&fakePool{}, WithSearch(search.NewHandler(&fakePool{}, allowLimiter{}, nil))))
	unwired := httpx.NewRouter(slog.New(slog.DiscardHandler), newDeps(&fakePool{}))
	tests := []struct {
		name   string
		router http.Handler
		target string
		status int
	}{
		{"unwired route stays 501", unwired, "/api/v1/search?q=ab", http.StatusNotImplemented},
		{"missing q", routed, "/api/v1/search", http.StatusBadRequest},
		{"per_group not an integer", routed, "/api/v1/search?q=ab&per_group=abc", http.StatusBadRequest},
		{"one-character q", routed, "/api/v1/search?q=a", http.StatusUnprocessableEntity},
		{"per_group 21", routed, "/api/v1/search?q=ab&per_group=21", http.StatusUnprocessableEntity},
		{"unknown type", routed, "/api/v1/search?q=ab&types=answers", http.StatusUnprocessableEntity},
		{"duplicate type", routed, "/api/v1/search?q=ab&types=learn,learn", http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil))
			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.status, rec.Body.String())
			}
		})
	}
}

// allowLimiter lets every request through, so route tests reach the handler.
type allowLimiter struct{}

func (allowLimiter) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}
