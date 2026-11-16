package llm_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm/aistest"
)

// The free tier answers 429, 500 and 503 and times out (V2, ROADMAP). The tracker counts each class per
// model with its latency, so the numbers can be read before the demo. It never records a prompt, an
// answer or the key.

func steppingClock(step time.Duration) func() time.Time {
	var mu sync.Mutex
	now := time.Unix(0, 0)
	return func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		now = now.Add(step)
		return now
	}
}

func TestStatusOfReadsTheHTTPStatus(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: 503}))
	_, err := newClient(t, s).Generate(context.Background(), llm.Request{User: "q"})
	if got := llm.StatusOf(err); got != 503 {
		t.Fatalf("StatusOf = %d, want 503 (err = %v)", got, err)
	}
	if !errors.Is(err, llm.ErrUnavailable) || !strings.Contains(err.Error(), "http 503") {
		t.Fatalf("the sentinel and the message must stay as before: %v", err)
	}
	if got := llm.StatusOf(errors.New("plain")); got != 0 {
		t.Fatalf("StatusOf(plain) = %d, want 0", got)
	}
	if got := llm.StatusOf(nil); got != 0 {
		t.Fatalf("StatusOf(nil) = %d, want 0", got)
	}
}

func TestTrackerCountsEveryClassPerModel(t *testing.T) {
	tr := llm.NewTracker(nil, llm.WithClock(steppingClock(10*time.Millisecond)))
	ok := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Body: aistest.AnswerBody("fine", 1, 1)}))
	for _, c := range []struct {
		status int
		n      int
	}{{429, 2}, {500, 1}, {503, 3}, {502, 1}, {401, 1}} {
		s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: c.status}))
		w := tr.Wrap("m-fail", newClient(t, s))
		for range c.n {
			_, _ = w.Generate(context.Background(), llm.Request{User: "q"})
		}
	}
	good := tr.Wrap("m-ok", newClient(t, ok))
	for range 2 {
		if _, err := good.Generate(context.Background(), llm.Request{User: "q"}); err != nil {
			t.Fatal(err)
		}
	}
	slow := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Delay: 2 * time.Second}))
	short := newClient(t, slow, func(c *llm.AIStudioConfig) { c.Timeout = 50 * time.Millisecond })
	_, _ = tr.Wrap("m-fail", short).Generate(context.Background(), llm.Request{User: "q"})
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = tr.Wrap("m-fail", short).Generate(canceled, llm.Request{User: "q"})
	_, _ = tr.Wrap("m-fail", llm.Disabled{}).Generate(context.Background(), llm.Request{User: "q"})
	_, _ = tr.Wrap("m-fail", llm.NewFake()).Generate(context.Background(), llm.Request{})

	snap := tr.Snapshot()
	f, o := snap["m-fail"], snap["m-ok"]
	if f.Quota429 != 2 || f.Unavailable500 != 1 || f.Unavailable503 != 3 || f.UnavailableOther != 1 || f.Rejected != 1 {
		t.Errorf("failing model counters = %+v", f)
	}
	if f.Timeouts != 1 || f.Canceled != 1 || f.NotConfigured != 1 || f.Invalid != 1 || f.OK != 0 {
		t.Errorf("failing model counters (timeouts, canceled, not configured, invalid) = %+v", f)
	}
	if f.Calls != 12 {
		t.Errorf("failing model calls = %d, want 12", f.Calls)
	}
	if o.Calls != 2 || o.OK != 2 {
		t.Errorf("healthy model counters = %+v", o)
	}
	if o.TotalLatency != 20*time.Millisecond || o.MaxLatency != 10*time.Millisecond {
		t.Errorf("latency total %v max %v, want 20ms and 10ms", o.TotalLatency, o.MaxLatency)
	}
	if got := o.MeanLatency(); got != 10*time.Millisecond {
		t.Errorf("MeanLatency = %v, want 10ms", got)
	}
	if got := (llm.ModelStats{}).MeanLatency(); got != 0 {
		t.Errorf("MeanLatency of nothing = %v, want 0", got)
	}
}

func TestTrackerSnapshotIsACopy(t *testing.T) {
	tr := llm.NewTracker(nil)
	_, _ = tr.Wrap("m", llm.NewFake()).Generate(context.Background(), llm.Request{User: "q"})
	snap := tr.Snapshot()
	snap["m"] = llm.ModelStats{Calls: 99}
	delete(snap, "other")
	if got := tr.Snapshot()["m"].Calls; got != 1 {
		t.Fatalf("a caller changed the tracker through the snapshot: %d", got)
	}
}

func TestTrackerLogsNeitherPromptNorKeyNorAnswer(t *testing.T) {
	var buf bytes.Buffer
	tr := llm.NewTracker(slog.New(slog.NewJSONHandler(&buf, nil)))
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: 429}))
	secret := "PROMPT-WITH-PERSONAL-DATA"
	_, _ = tr.Wrap("m", newClient(t, s)).Generate(context.Background(), llm.Request{User: secret})
	out := buf.String()
	if out == "" {
		t.Fatal("a failure must be logged")
	}
	for _, banned := range []string{secret, fixtureKey()} {
		if strings.Contains(out, banned) {
			t.Fatalf("the log contains %q: %s", banned, out)
		}
	}
	for _, want := range []string{`"model":"m"`, `"class":"quota"`, `"status":429`, `"latency_ms"`} {
		if !strings.Contains(out, want) {
			t.Errorf("log %s lacks %s", out, want)
		}
	}
	buf.Reset()
	ok := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Body: aistest.AnswerBody("answer text", 1, 1)}))
	_, _ = tr.Wrap("m", newClient(t, ok)).Generate(context.Background(), llm.Request{User: secret})
	if buf.Len() != 0 {
		t.Fatalf("a success must stay silent: %s", buf.String())
	}
}

func TestTrackerIsSafeForConcurrentUse(t *testing.T) {
	tr := llm.NewTracker(nil)
	w := tr.Wrap("m", llm.NewFake())
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			for range 10 {
				_, _ = w.Generate(context.Background(), llm.Request{User: "q"})
				_ = tr.Snapshot()
			}
		})
	}
	wg.Wait()
	if got := tr.Snapshot()["m"].Calls; got != 200 {
		t.Fatalf("calls = %d, want 200", got)
	}
}

func TestTrackerPassesTheResponseThrough(t *testing.T) {
	f := llm.NewFake()
	f.Enqueue(llm.Response{Text: "hello", Model: "m"})
	got, err := llm.NewTracker(nil).Wrap("m", f).Generate(context.Background(), llm.Request{User: "q"})
	if err != nil || got.Text != "hello" {
		t.Fatalf("got %+v, %v", got, err)
	}
	f.EnqueueError(fmt.Errorf("%w: boom", llm.ErrUnavailable))
	if _, err := llm.NewTracker(nil).Wrap("m", f).Generate(context.Background(), llm.Request{User: "q"}); !errors.Is(err, llm.ErrUnavailable) {
		t.Fatalf("the error must pass through unchanged: %v", err)
	}
}
