package app

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

func TestAskAssistantStaysUnimplementedWithoutAHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	NewServer().AskAssistant(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status %d", rec.Code)
	}
}

// Without AI config the route is wired and answers 503 (never 501, never a call out).
func TestAskAssistantIsUnavailableWhenAIIsOff(t *testing.T) {
	h := newAssistantHandler(nil, &fakePool{}, allowLimiter{}, slog.New(slog.DiscardHandler))
	s := NewServer(WithAssistant(h))
	req := httptest.NewRequestWithContext(subject.With(t.Context(), subject.Subject{UserID: "u1", EmailVerified: true}),
		http.MethodPost, "/", strings.NewReader(`{"question":"how do webhooks work"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.AskAssistant(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAskAssistantRouteRequiresASession(t *testing.T) {
	h := newAssistantHandler(nil, &fakePool{}, allowLimiter{}, nil)
	router := httpx.NewRouter(slog.New(slog.DiscardHandler), newDeps(&fakePool{}, WithAssistant(h)))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/assistant/answers", strings.NewReader(`{"question":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("anonymous status %d, want 401 or 403 (CSRF)", rec.Code)
	}
}

func TestAssistantModelFollowsConfig(t *testing.T) {
	if m, on, _ := assistantModel(nil, nil); on || m != (llm.Disabled{}) {
		t.Fatalf("nil config must give the Disabled client, got %T on=%v", m, on)
	}
	ai := &config.AI{Enabled: true, APIKey: "k", Model: "gemma-test", Timeout: 30 * time.Second}
	m, on, timeout := assistantModel(ai, nil)
	if !on || m == (llm.Disabled{}) {
		t.Fatalf("enabled config must give a real client, got %T on=%v", m, on)
	}
	if timeout != 20*time.Second {
		t.Fatalf("timeout %s, want the 20 s cap inside the 30 s route budget", timeout)
	}
	ai.Timeout = 5 * time.Second
	if _, _, timeout = assistantModel(ai, nil); timeout != 5*time.Second {
		t.Fatalf("a shorter AI_TIMEOUT must be kept, got %s", timeout)
	}
}
