package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

func TestErrorReportStaysUnimplementedWithoutAHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	NewServer().CreateResourceErrorReport(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil),
		uuid.MustParse("00000000-0000-7000-8000-0000000000a1"))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestErrorReportDelegatesWhenSet(t *testing.T) {
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(WithErrorReports(discover.NewHandler(&fakePool{}, c, allowLimiter{}, nil)))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"kind":"other","body":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// No session subject: the real handler answers 401, never the 501 of the unwired route.
	s.CreateResourceErrorReport(rec, req, uuid.MustParse("00000000-0000-7000-8000-0000000000a1"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want the handler's 401", rec.Code)
	}
}
