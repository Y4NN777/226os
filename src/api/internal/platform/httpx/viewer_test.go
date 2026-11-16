package httpx_test

import (
	"net/http/httptest"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

func TestViewerScopedVariesByCookieAndStaysPrivate(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Add("Vary", "Accept-Encoding")
	httpx.ViewerScoped(rec)
	if got := rec.Header().Values("Vary"); len(got) != 2 || got[0] != "Accept-Encoding" || got[1] != "Cookie" {
		t.Fatalf("Vary = %v, want the existing value kept and Cookie added", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "private, no-cache" {
		t.Fatalf("Cache-Control = %q", got)
	}
}
