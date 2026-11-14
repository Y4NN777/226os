package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/assistant/domain"
	"github.com/Y4NN777/226os/src/api/internal/assistant/service"
	"github.com/Y4NN777/226os/src/api/internal/assistant/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

type retriever struct{ passages []domain.Passage }

func (r retriever) Retrieve(context.Context, string, int) ([]domain.Passage, error) {
	return r.passages, nil
}

type allow struct{}

func (allow) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

var passage = domain.Passage{Type: "article", ID: "11111111-1111-1111-1111-111111111111", Title: "Webhooks", Text: "Sign it.", URL: "/article/1"}

func serve(t *testing.T, enabled bool, passages []domain.Passage, fake *llm.Fake, sub subject.Subject, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := transport.New(service.New(service.Config{Enabled: enabled}, retriever{passages}, fake, allow{}, nil), nil)
	req := httptest.NewRequestWithContext(subject.With(t.Context(), sub), http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.AskAssistant(rec, req)
	return rec
}

var verified = subject.Subject{UserID: "u", EmailVerified: true}

func TestAskAssistantAnswered(t *testing.T) {
	fake := llm.NewFake()
	fake.Enqueue(llm.Response{Text: "Sign it [S1].\nHow to verify\nRead the docs."})
	rec := serve(t, true, []domain.Passage{passage}, fake, verified, `{"question":"how do webhooks work","lang":"en"}`)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	var got apigen.AssistantAnswer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Outcome != apigen.AssistantAnswerOutcomeAnswered || got.GeneratedBy != apigen.Ai || got.Lang != apigen.En ||
		got.Answer != "Sign it [1]." || len(got.Sources) != 1 || got.Sources[0].Url != "/article/1" || got.SearchUrl != nil ||
		got.Sources[0].Type != apigen.AssistantSourceTypeArticle || !strings.HasPrefix(got.HowToVerify, "Read the docs.") {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestAskAssistantSourceDates(t *testing.T) {
	dated := passage
	dated.PublishedAt = time.Date(2026, 11, 2, 23, 30, 0, 0, time.FixedZone("x", -2*3600))
	undated := passage
	undated.ID = "22222222-2222-2222-2222-222222222222"
	fake := llm.NewFake()
	fake.Enqueue(llm.Response{Text: "Sign it [S1] and sign it again [S2]."})
	rec := serve(t, true, []domain.Passage{dated, undated}, fake, verified, `{"question":"how do webhooks work","lang":"en"}`)
	var raw struct {
		Sources []map[string]any `json:"sources"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil || len(raw.Sources) != 2 {
		t.Fatalf("%v %s", err, rec.Body.String())
	}
	if raw.Sources[0]["published_at"] != "2026-11-03T01:30:00Z" {
		t.Fatalf("dated source: %v", raw.Sources[0])
	}
	if v, ok := raw.Sources[1]["published_at"]; !ok || v != nil || strings.Contains(rec.Body.String(), "0001-01-01") {
		t.Fatalf("an undated source must carry published_at null, never the zero date: %s", rec.Body.String())
	}
}

func TestAskAssistantNoSource(t *testing.T) {
	rec := serve(t, true, nil, llm.NewFake(), verified, `{"question":"comment faire un paiement"}`)
	var got apigen.AssistantAnswer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || got.Outcome != apigen.AssistantAnswerOutcomeNoSource || got.GeneratedBy != apigen.Ai ||
		got.SearchUrl == nil || len(got.Sources) != 0 || got.Lang != apigen.Fr || got.HowToVerify == "" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"sources":[]`) {
		t.Fatalf("sources must be an empty array: %s", rec.Body.String())
	}
}

func TestAskAssistantProblems(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		sub     subject.Subject
		body    string
		status  int
		code    string
	}{
		{"anonymous", true, subject.Subject{}, `{"question":"how do webhooks work"}`, 401, "UNAUTHENTICATED"},
		{"unverified", true, subject.Subject{UserID: "u"}, `{"question":"how do webhooks work"}`, 403, "EMAIL_NOT_VERIFIED"},
		{"unknown field", true, verified, `{"question":"how do webhooks work","x":1}`, 422, "VALIDATION_ERROR"},
		{"short question", true, verified, `{"question":"ab"}`, 422, "VALIDATION_ERROR"},
		{"disabled", false, verified, `{"question":"how do webhooks work"}`, 503, "SERVICE_UNAVAILABLE"},
	}
	retryAfter := map[string]string{"disabled": "30"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := llm.NewFake()
			rec := serve(t, tc.enabled, []domain.Passage{passage}, fake, tc.sub, tc.body)
			var p struct {
				Code string `json:"code"`
			}
			_ = json.Unmarshal(rec.Body.Bytes(), &p)
			if rec.Code != tc.status || p.Code != tc.code || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/problem+json") {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
			if len(fake.Requests()) != 0 {
				t.Fatal("the provider must not be called")
			}
			if got := rec.Header().Get("Retry-After"); got != retryAfter[tc.name] {
				t.Fatalf("Retry-After %q, want %q", got, retryAfter[tc.name])
			}
		})
	}
}

type brokenLimiter struct{}

func (brokenLimiter) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return false, 0, context.DeadlineExceeded
}

type spentGlobal struct{}

func (spentGlobal) Hit(_ context.Context, scope string, _ string, _ int, _ time.Duration) (bool, time.Duration, error) {
	if scope == "assistant:global" {
		return false, 90 * time.Second, nil
	}
	return true, 0, nil
}

func TestAskAssistant503sCarryRetryAfter(t *testing.T) {
	down := llm.NewFake()
	down.FailWith(llm.ErrUnavailable)
	quota := llm.NewFake()
	quota.FailWith(&llm.StatusError{Kind: llm.ErrQuotaExhausted, Status: 429})
	tests := []struct {
		name  string
		lim   service.Limiter
		model *llm.Fake
		want  string
	}{
		{"limiter failure", brokenLimiter{}, llm.NewFake(), "30"},
		{"provider down", allow{}, down, "30"},
		{"provider 429", allow{}, quota, "30"},
		{"global guard spent keeps the limiter retry", spentGlobal{}, llm.NewFake(), "90"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := transport.New(service.New(service.Config{Enabled: true}, retriever{[]domain.Passage{passage}}, tc.model, tc.lim, nil), nil)
			req := httptest.NewRequestWithContext(subject.With(t.Context(), verified), http.MethodPost, "/", strings.NewReader(`{"question":"how do webhooks work"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.AskAssistant(rec, req)
			if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") != tc.want {
				t.Fatalf("%d Retry-After %q, want 503 and %q", rec.Code, rec.Header().Get("Retry-After"), tc.want)
			}
		})
	}
}

// Negative authz for POST /api/v1/assistant/answers (AUTHORIZATION_MATRIX, allow: member with a
// verified email): anonymous callers get 401, an unverified member 403, and neither reaches the
// retrieval or the provider.
func TestAuthzAskAssistantAnonymous(t *testing.T) {
	const route = "/api/v1/assistant/answers"
	cases := []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: "u"}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := llm.NewFake()
			rec := serve(t, true, []domain.Passage{passage}, fake, tc.sub, `{"question":"how do webhooks work"}`)
			if rec.Code != tc.status || len(fake.Requests()) != 0 {
				t.Fatalf("POST %s: status %d, provider calls %d; want %d and none", route, rec.Code, len(fake.Requests()), tc.status)
			}
		})
	}
}
