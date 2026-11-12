package llm_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
)

func TestDisabledReturnsNotConfigured(t *testing.T) {
	var c llm.Client = llm.Disabled{}
	_, err := c.Generate(context.Background(), llm.Request{User: "hi"})
	if !errors.Is(err, llm.ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
}

func TestSentinelsAreDistinctAndWrappable(t *testing.T) {
	all := []error{llm.ErrNotConfigured, llm.ErrInvalidRequest, llm.ErrQuotaExhausted, llm.ErrUnavailable, llm.ErrRejected}
	for i, a := range all {
		wrapped := errors.Join(errors.New("ctx"), a)
		for j, b := range all {
			if errors.Is(wrapped, b) != (i == j) {
				t.Fatalf("errors.Is(%v, %v) wrong", a, b)
			}
		}
	}
}

func TestValidModelName(t *testing.T) {
	for s, want := range map[string]bool{
		"gemma-4_x.1": true, "a": true, strings.Repeat("a", 64): true,
		"": false, strings.Repeat("a", 65): false, "a/b": false, "a?b": false, "a b": false, "a:b": false, "../x": false,
	} {
		if got := llm.ValidModelName(s); got != want {
			t.Errorf("ValidModelName(%q) = %v, want %v", s, got, want)
		}
	}
}

// requestCases are refused by every implementation before any I/O.
var requestCases = []struct {
	name string
	req  llm.Request
}{
	{"empty user", llm.Request{}},
	{"user too long", llm.Request{User: strings.Repeat("a", llm.MaxUserBytes+1)}},
	{"system too long", llm.Request{User: "q", System: strings.Repeat("a", llm.MaxSystemBytes+1)}},
	{"NUL in user", llm.Request{User: "a\x00b"}},
	{"NUL in system", llm.Request{User: "q", System: "a\x00b"}},
	{"negative tokens", llm.Request{User: "q", MaxOutputTokens: -1}},
	{"too many tokens", llm.Request{User: "q", MaxOutputTokens: llm.MaxOutputTokensLimit + 1}},
}

func TestFakeRefusesInvalidRequests(t *testing.T) {
	for _, tc := range requestCases {
		t.Run(tc.name, func(t *testing.T) {
			f := llm.NewFake()
			_, err := f.Generate(context.Background(), tc.req)
			if !errors.Is(err, llm.ErrInvalidRequest) {
				t.Fatalf("err = %v", err)
			}
			if strings.Contains(err.Error(), "aaaa") {
				t.Fatalf("error carries request content: %q", err)
			}
			if len(f.Requests()) != 0 {
				t.Fatal("an invalid request was recorded")
			}
		})
	}
}

func TestFakeAcceptsBoundaryRequests(t *testing.T) {
	f := llm.NewFake()
	for _, r := range []llm.Request{
		{User: "q"},
		{User: strings.Repeat("a", llm.MaxUserBytes), System: strings.Repeat("b", llm.MaxSystemBytes), MaxOutputTokens: llm.MaxOutputTokensLimit},
	} {
		if _, err := f.Generate(context.Background(), r); err != nil {
			t.Fatalf("Generate(%d bytes) = %v", len(r.User), err)
		}
	}
}

func TestFakeDefaultsAndRecordsCopies(t *testing.T) {
	f := llm.NewFake()
	resp, err := f.Generate(context.Background(), llm.Request{User: "q", System: "s"})
	if err != nil || resp.Text == "" || resp.Model == "" {
		t.Fatalf("resp = %+v err = %v", resp, err)
	}
	got := f.Requests()
	if len(got) != 1 || got[0].User != "q" || got[0].System != "s" || got[0].MaxOutputTokens != llm.DefaultMaxOutputTokens {
		t.Fatalf("requests = %+v", got)
	}
	got[0].User = "mutated"
	if f.Requests()[0].User != "q" {
		t.Fatal("Requests returned the internal slice")
	}
}

func TestFakeScripting(t *testing.T) {
	boom := errors.New("boom")
	f := llm.NewFake()
	f.Enqueue(llm.Response{Text: "one"})
	f.EnqueueError(llm.ErrQuotaExhausted)
	f.Respond = func(r llm.Request) (llm.Response, error) { return llm.Response{Text: "echo " + r.User}, nil }
	ctx := context.Background()

	if r, err := f.Generate(ctx, llm.Request{User: "a"}); err != nil || r.Text != "one" {
		t.Fatalf("first = %+v %v", r, err)
	}
	if _, err := f.Generate(ctx, llm.Request{User: "b"}); !errors.Is(err, llm.ErrQuotaExhausted) {
		t.Fatalf("second err = %v", err)
	}
	if r, err := f.Generate(ctx, llm.Request{User: "c"}); err != nil || r.Text != "echo c" {
		t.Fatalf("third = %+v %v", r, err)
	}
	f.FailWith(boom)
	f.Enqueue(llm.Response{Text: "never"})
	if _, err := f.Generate(ctx, llm.Request{User: "d"}); !errors.Is(err, boom) {
		t.Fatalf("FailWith err = %v", err)
	}
	f.FailWith(nil)
	if r, err := f.Generate(ctx, llm.Request{User: "e"}); err != nil || r.Text != "never" {
		t.Fatalf("after FailWith(nil) = %+v %v", r, err)
	}
	if n := len(f.Requests()); n != 5 {
		t.Fatalf("recorded %d requests, want 5", n)
	}
}

func TestFakeHonoursCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := llm.NewFake()
	if _, err := f.Generate(ctx, llm.Request{User: "q"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if len(f.Requests()) != 0 {
		t.Fatal("a cancelled call was recorded")
	}
}

func TestFakeIsSafeForConcurrentUse(t *testing.T) {
	f := llm.NewFake()
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			f.Enqueue(llm.Response{Text: "x"})
			_, _ = f.Generate(context.Background(), llm.Request{User: "q"})
			_ = f.Requests()
		})
	}
	wg.Wait()
	if n := len(f.Requests()); n != 20 {
		t.Fatalf("recorded %d, want 20", n)
	}
}
