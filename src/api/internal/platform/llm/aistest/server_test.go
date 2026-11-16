package aistest_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm/aistest"
)

type reply struct {
	StatusCode int
	Header     http.Header
}

func do(t *testing.T, method, url, body string, hdr map[string]string) (reply, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return reply{StatusCode: resp.StatusCode, Header: resp.Header}, string(b)
}

func TestDefaultsAndRecording(t *testing.T) {
	s := aistest.NewServer(t)
	resp, body := do(t, http.MethodPost, s.URL()+"/v1beta/models/m1:generateContent?x=1", `{"a":1}`, map[string]string{"X-Goog-Api-Key": "k1"})
	if resp.StatusCode != 200 {
		t.Fatalf("generate status = %d", resp.StatusCode)
	}
	var gen struct {
		Candidates []struct {
			Content struct{ Parts []struct{ Text string } }
		}
		UsageMetadata struct{ PromptTokenCount, CandidatesTokenCount int }
	}
	if err := json.Unmarshal([]byte(body), &gen); err != nil || gen.Candidates[0].Content.Parts[0].Text == "" {
		t.Fatalf("generate body = %q err = %v", body, err)
	}
	resp, body = do(t, http.MethodGet, s.URL()+"/v1beta/models?pageSize=200", "", nil)
	if resp.StatusCode != 200 || !strings.Contains(body, `"models/`) {
		t.Fatalf("models = %d %q", resp.StatusCode, body)
	}
	resp, _ = do(t, http.MethodGet, s.URL()+"/nope", "", nil)
	if resp.StatusCode != 404 {
		t.Fatalf("unknown path status = %d", resp.StatusCode)
	}

	got := s.Requests()
	if len(got) != 3 {
		t.Fatalf("recorded %d requests", len(got))
	}
	r := got[0]
	if r.Method != "POST" || r.Path != "/v1beta/models/m1:generateContent" || r.Query != "x=1" || r.Body != `{"a":1}` || r.Header.Get("x-goog-api-key") != "k1" {
		t.Fatalf("recorded = %+v", r)
	}
	if got[1].Query != "pageSize=200" || got[2].Path != "/nope" {
		t.Fatalf("recorded = %+v", got)
	}
	got[0].Body = "mutated"
	if s.Requests()[0].Body != `{"a":1}` {
		t.Fatal("Requests returned internal state")
	}
}

func TestScriptedBehavioursAreConsumedInOrderAndTheLastRepeats(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(
		aistest.Behaviour{Status: 429, Body: "first"},
		aistest.Behaviour{Status: 503, Body: "second"},
	), aistest.WithModels(aistest.Behaviour{Body: aistest.ModelsBody("tok", "gemma-x", "other")}))
	for i, want := range []struct {
		code int
		body string
	}{{429, "first"}, {503, "second"}, {503, "second"}} {
		resp, body := do(t, http.MethodPost, s.URL()+"/v1beta/models/m:generateContent", "{}", nil)
		if resp.StatusCode != want.code || body != want.body {
			t.Fatalf("call %d = %d %q", i, resp.StatusCode, body)
		}
	}
	_, body := do(t, http.MethodGet, s.URL()+"/v1beta/models", "", nil)
	var list struct {
		Models []struct {
			Name                       string
			DisplayName                string
			SupportedGenerationMethods []string
			InputTokenLimit            int
			OutputTokenLimit           int
		}
		NextPageToken string
	}
	if err := json.Unmarshal([]byte(body), &list); err != nil {
		t.Fatal(err)
	}
	m := list.Models[0]
	if len(list.Models) != 2 || list.NextPageToken != "tok" || m.Name != "models/gemma-x" || m.DisplayName == "" ||
		len(m.SupportedGenerationMethods) == 0 || m.InputTokenLimit != aistest.ModelInputLimit || m.OutputTokenLimit != aistest.ModelOutputLimit {
		t.Fatalf("list = %+v", list)
	}
}

func TestRedirectAndDelay(t *testing.T) {
	target := aistest.NewServer(t)
	s := aistest.NewServer(t, aistest.WithGenerate(
		aistest.Behaviour{Redirect: target.URL() + "/landed"},
		aistest.Behaviour{Delay: 5 * time.Millisecond, Body: "late"},
	))
	resp, _ := do(t, http.MethodPost, s.URL()+"/v1beta/models/m:generateContent", "{}", nil)
	if resp.StatusCode != http.StatusTemporaryRedirect || resp.Header.Get("Location") != target.URL()+"/landed" {
		t.Fatalf("redirect = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	if len(target.Requests()) != 0 {
		t.Fatal("the test client followed the redirect")
	}
	start := time.Now()
	_, body := do(t, http.MethodPost, s.URL()+"/v1beta/models/m:generateContent", "{}", nil)
	if body != "late" || time.Since(start) < 5*time.Millisecond {
		t.Fatalf("delayed body = %q after %v", body, time.Since(start))
	}
}

func TestStallEndsWhenTheClientGivesUp(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Delay: time.Hour}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL()+"/v1beta/models/m:generateContent", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
		done <- err
	}()
	// Give up only once the server holds the request, so its stall really ends on the client's
	// cancellation (a fixed client timeout could fire before the request arrived on a loaded machine).
	deadline := time.Now().Add(5 * time.Second)
	for len(s.Requests()) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the server never received the request")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; err == nil {
		t.Fatal("expected the cancelled request to fail")
	}
}

func TestEndlessBodyIsBoundedByTheReader(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Endless: true}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.URL()+"/v1beta/models/m:generateContent", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<20))
	if err != nil || n != 4<<20 {
		t.Fatalf("read %d bytes, err = %v", n, err)
	}
}

func TestTruncatedBodyFailsTheClientRead(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Truncate: true, Body: "short"}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.URL()+"/v1beta/models/m:generateContent", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if _, err := io.ReadAll(resp.Body); err == nil {
		t.Fatal("expected a read error on a truncated body")
	}
}
