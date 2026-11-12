package llm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm/aistest"
)

// fixtureKey is built at run time so no key-shaped literal sits in the source.
func fixtureKey() string { return fmt.Sprintf("%s%s", "AIza", strings.Repeat("Q", 35)) }

const testModel = "gemma-test"

func newClient(t *testing.T, s *aistest.Server, mutate ...func(*llm.AIStudioConfig)) *llm.AIStudio {
	t.Helper()
	cfg := llm.AIStudioConfig{APIKey: fixtureKey(), Model: testModel, Timeout: 5 * time.Second}
	for _, m := range mutate {
		m(&cfg)
	}
	return llm.NewAIStudio(cfg, llm.WithBaseURL(s.URL()))
}

func answer(text string) aistest.Behaviour {
	return aistest.Behaviour{Body: aistest.AnswerBody(text, 1, 1)}
}

func TestGenerateSuccessWithSystemInstruction(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Body: `{
		"candidates":[{"content":{"parts":[{"text":"Hello "},{"text":"world"}]},"finishReason":"STOP"}],
		"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":4}}`}))
	var c llm.Client = newClient(t, s)

	resp, err := c.Generate(context.Background(), llm.Request{System: "be brief", User: "what is 404?", MaxOutputTokens: 77})
	if err != nil {
		t.Fatal(err)
	}
	if resp != (llm.Response{Text: "Hello world", Model: testModel, PromptTokens: 11, OutputTokens: 4}) {
		t.Fatalf("resp = %+v", resp)
	}
	reqs := s.Requests()
	if len(reqs) != 1 {
		t.Fatalf("%d requests", len(reqs))
	}
	r := reqs[0]
	if r.Method != http.MethodPost || r.Path != "/v1beta/models/"+testModel+":generateContent" || r.Query != "" {
		t.Fatalf("request line = %s %s?%s", r.Method, r.Path, r.Query)
	}
	if r.Header.Get("x-goog-api-key") != fixtureKey() || r.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("headers = %v", r.Header)
	}
	var body struct {
		SystemInstruction struct{ Parts []struct{ Text string } }
		Contents          []struct {
			Role  string
			Parts []struct{ Text string }
		}
		GenerationConfig struct{ MaxOutputTokens int }
	}
	if err := json.Unmarshal([]byte(r.Body), &body); err != nil {
		t.Fatal(err)
	}
	if body.SystemInstruction.Parts[0].Text != "be brief" || len(body.Contents) != 1 || body.Contents[0].Role != "user" ||
		body.Contents[0].Parts[0].Text != "what is 404?" || body.GenerationConfig.MaxOutputTokens != 77 {
		t.Fatalf("body = %s", r.Body)
	}
}

func TestGenerateOmitsEmptySystemInstructionAndDefaultsTokens(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(answer("ok")))
	if _, err := newClient(t, s).Generate(context.Background(), llm.Request{User: "q"}); err != nil {
		t.Fatal(err)
	}
	body := s.Requests()[0].Body
	if strings.Contains(body, "systemInstruction") || !strings.Contains(body, `"maxOutputTokens":1024`) {
		t.Fatalf("body = %s", body)
	}
}

func TestTheKeyTravelsInTheHeaderOnly(t *testing.T) {
	s := aistest.NewServer(t)
	c := newClient(t, s)
	if _, err := c.Generate(context.Background(), llm.Request{System: "s", User: "u"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListModels(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, r := range s.Requests() {
		if strings.Contains(r.Path+"?"+r.Query+r.Body, fixtureKey()) {
			t.Fatalf("key leaked into the URL or body of %s %s", r.Method, r.Path)
		}
		if r.Header.Get("x-goog-api-key") != fixtureKey() {
			t.Fatalf("missing key header on %s", r.Path)
		}
	}
}

func TestStatusClassification(t *testing.T) {
	const marker = "BODYMARKER-do-not-leak"
	for _, tc := range []struct {
		status int
		want   error
	}{
		{429, llm.ErrQuotaExhausted},
		{400, llm.ErrRejected}, {401, llm.ErrRejected}, {403, llm.ErrRejected}, {404, llm.ErrRejected},
		{500, llm.ErrUnavailable}, {503, llm.ErrUnavailable}, {418, llm.ErrUnavailable}, {204, llm.ErrUnavailable},
	} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: tc.status, Body: marker + " " + fixtureKey()}))
			c := newClient(t, s)
			_, err := c.Generate(context.Background(), llm.Request{User: "PROMPTMARKER"})
			assertClass(t, err, tc.want, tc.status)
			s2 := aistest.NewServer(t, aistest.WithModels(aistest.Behaviour{Status: tc.status, Body: marker}))
			_, err = newClient(t, s2).ListModels(context.Background())
			assertClass(t, err, tc.want, tc.status)
		})
	}
}

func assertClass(t *testing.T, err, want error, status int) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	msg := err.Error()
	for _, bad := range []string{"BODYMARKER", "PROMPTMARKER", fixtureKey()} {
		if strings.Contains(msg, bad) {
			t.Fatalf("error leaks %q: %s", bad, msg)
		}
	}
	if !strings.Contains(msg, fmt.Sprintf("http %d", status)) {
		t.Fatalf("error %q does not name the status %d", msg, status)
	}
}

func TestUnusableAnswersAreUnavailable(t *testing.T) {
	for name, body := range map[string]string{
		"no candidate":    `{"candidates":[]}`,
		"blocked prompt":  `{"promptFeedback":{"blockReason":"SAFETY"},"candidates":[{"content":{"parts":[{"text":"x"}]}}]}`,
		"safety finish":   `{"candidates":[{"finishReason":"SAFETY","content":{"parts":[{"text":"x"}]}}]}`,
		"no text":         `{"candidates":[{"finishReason":"STOP","content":{"parts":[]}}]}`,
		"empty text":      `{"candidates":[{"content":{"parts":[{"text":""}]}}]}`,
		"not json":        `<html>oops</html>`,
		"json wrong type": `{"candidates":"x"}`,
	} {
		t.Run(name, func(t *testing.T) {
			s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Body: body}))
			_, err := newClient(t, s).Generate(context.Background(), llm.Request{User: "q"})
			if !errors.Is(err, llm.ErrUnavailable) {
				t.Fatalf("err = %v", err)
			}
			if strings.Contains(err.Error(), "<html>") {
				t.Fatalf("error carries the body: %v", err)
			}
		})
	}
}

func TestOversizeAndTruncatedBodiesAreUnavailable(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Endless: true}), aistest.WithModels(aistest.Behaviour{Endless: true}))
	c := newClient(t, s)
	if _, err := c.Generate(context.Background(), llm.Request{User: "q"}); !errors.Is(err, llm.ErrUnavailable) || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("generate err = %v", err)
	}
	if _, err := c.ListModels(context.Background()); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("list err = %v", err)
	}
	short := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Truncate: true, Body: `{"cand`}))
	if _, err := newClient(t, short).Generate(context.Background(), llm.Request{User: "q"}); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("truncated err = %v", err)
	}
}

func TestRedirectsAreRefused(t *testing.T) {
	target := aistest.NewServer(t)
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Redirect: target.URL() + "/v1beta/models/x:generateContent"}),
		aistest.WithModels(aistest.Behaviour{Redirect: target.URL() + "/v1beta/models"}))
	c := newClient(t, s)
	if _, err := c.Generate(context.Background(), llm.Request{User: "q"}); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("generate err = %v", err)
	}
	if _, err := c.ListModels(context.Background()); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("list err = %v", err)
	}
	if n := len(target.Requests()); n != 0 {
		t.Fatalf("the redirect target received %d requests (the key would have followed)", n)
	}
}

func TestLocalRefusalsSendNothing(t *testing.T) {
	s := aistest.NewServer(t)
	ctx := context.Background()
	for _, model := range []string{"", "a/b", "a?b=c", "../v1", "a b", "a:generateContent", strings.Repeat("m", 65), "m\n"} {
		_, err := newClient(t, s, func(c *llm.AIStudioConfig) { c.Model = model }).Generate(ctx, llm.Request{User: "q"})
		if !errors.Is(err, llm.ErrRejected) {
			t.Errorf("model %q: err = %v", model, err)
		}
	}
	_, err := newClient(t, s, func(c *llm.AIStudioConfig) { c.APIKey = "" }).Generate(ctx, llm.Request{User: "q"})
	if !errors.Is(err, llm.ErrNotConfigured) {
		t.Errorf("empty key: err = %v", err)
	}
	if _, err := newClient(t, s, func(c *llm.AIStudioConfig) { c.APIKey = "" }).ListModels(ctx); !errors.Is(err, llm.ErrNotConfigured) {
		t.Errorf("empty key list: err = %v", err)
	}
	if _, err := newClient(t, s).Generate(ctx, llm.Request{}); !errors.Is(err, llm.ErrInvalidRequest) {
		t.Errorf("empty request: err = %v", err)
	}
	if n := len(s.Requests()); n != 0 {
		t.Fatalf("%d requests were sent", n)
	}
}

func TestTransportFailuresAreUnavailable(t *testing.T) {
	cfg := llm.AIStudioConfig{APIKey: fixtureKey(), Model: testModel, Timeout: time.Second}
	bad := llm.NewAIStudio(cfg, llm.WithBaseURL("http://bad host/"))
	if _, err := bad.Generate(context.Background(), llm.Request{User: "q"}); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("bad URL err = %v", err)
	}
	if _, err := bad.ListModels(context.Background()); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("bad URL list err = %v", err)
	}
	dead := llm.NewAIStudio(cfg, llm.WithBaseURL(closedListenerURL(t)))
	_, err := dead.Generate(context.Background(), llm.Request{User: "q"})
	if !errors.Is(err, llm.ErrUnavailable) || strings.Contains(err.Error(), fixtureKey()) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("refused err = %v", err)
	}
	if _, err := dead.ListModels(context.Background()); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("refused list err = %v", err)
	}
}

// closedListenerURL returns the URL of a loopback port nobody listens on.
func closedListenerURL(t *testing.T) string {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return "http://" + addr
}

func TestTimeoutsAndCancellationWrapTheContextError(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Delay: time.Hour}), aistest.WithModels(aistest.Behaviour{Delay: time.Hour}))
	c := newClient(t, s, func(c *llm.AIStudioConfig) { c.Timeout = 50 * time.Millisecond })
	start := time.Now()
	_, err := c.Generate(context.Background(), llm.Request{User: "q"})
	if !errors.Is(err, llm.ErrUnavailable) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("generate err = %v", err)
	}
	if _, err := c.ListModels(context.Background()); !errors.Is(err, llm.ErrUnavailable) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("list err = %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("the timeout was not honoured")
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(30*time.Millisecond, cancel)
	slow := newClient(t, s)
	if _, err := slow.Generate(ctx, llm.Request{User: "q"}); !errors.Is(err, llm.ErrUnavailable) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel err = %v", err)
	}
}

func TestListModelsFollowsPagesAndTrimsThePrefix(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithModels(
		aistest.Behaviour{Body: aistest.ModelsBody("p2", "gemma-a")},
		aistest.Behaviour{Body: aistest.ModelsBody("", "other-b", "gemma-c")},
	))
	got, err := newClient(t, s).ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Name != "gemma-a" || got[2].Name != "gemma-c" || got[0].DisplayName != aistest.DisplayPrefix+"gemma-a" ||
		got[0].InputTokenLimit != aistest.ModelInputLimit || got[0].OutputTokenLimit != aistest.ModelOutputLimit ||
		len(got[0].SupportedMethods) != 2 || got[0].SupportedMethods[0] != "generateContent" {
		t.Fatalf("models = %+v", got)
	}
	reqs := s.Requests()
	if len(reqs) != 2 || reqs[0].Query != "pageSize=200" || reqs[1].Query != "pageSize=200&pageToken=p2" || reqs[0].Method != http.MethodGet {
		t.Fatalf("requests = %+v", reqs)
	}
}

func TestListModelsStopsAfterFivePages(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithModels(aistest.Behaviour{Body: aistest.ModelsBody("more", "gemma-loop")}))
	got, err := newClient(t, s).ListModels(context.Background())
	if err != nil || len(got) != 5 || len(s.Requests()) != 5 {
		t.Fatalf("got %d models, %d requests, err = %v", len(got), len(s.Requests()), err)
	}
}

func TestListModelsRejectsAMalformedPage(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithModels(aistest.Behaviour{Body: "not json"}))
	if _, err := newClient(t, s).ListModels(context.Background()); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestSecretsNeverAppearInFormattedOutput(t *testing.T) {
	cfg := llm.AIStudioConfig{APIKey: fixtureKey(), Model: testModel, Timeout: time.Second}
	c := llm.NewAIStudio(cfg)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	logger.Info("cfg", "config", cfg, "client", c, "clientv", *c)
	outs := map[string]string{"slog": buf.String()}
	for name, v := range map[string]any{"config": cfg, "client": c, "clientval": *c} {
		outs[name+" %v"] = fmt.Sprintf("%v", v)
		outs[name+" %+v"] = fmt.Sprintf("%+v", v)
		outs[name+" %#v"] = fmt.Sprintf("%#v", v)
		outs[name+" %s"] = fmt.Sprintf("%s", v)
	}
	for name, out := range outs {
		if strings.Contains(out, fixtureKey()) || strings.Contains(out, "QQQQ") {
			t.Errorf("%s leaks the key: %s", name, out)
		}
		if !strings.Contains(out, "redacted") || !strings.Contains(out, testModel) {
			t.Errorf("%s lost its content: %s", name, out)
		}
	}
}
