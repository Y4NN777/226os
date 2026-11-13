package llm

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// ModelStats counts the calls of one model by outcome. Latency covers every call, failed or not.
type ModelStats struct {
	Calls            int
	OK               int
	Quota429         int // ErrQuotaExhausted: the provider answered HTTP 429
	Unavailable500   int
	Unavailable503   int
	UnavailableOther int // other 5xx, empty or blocked output, network errors
	Timeouts         int
	Canceled         int
	Rejected         int // bad key or unknown model: a configuration problem
	NotConfigured    int
	Invalid          int // refused locally before any I/O
	TotalLatency     time.Duration
	MaxLatency       time.Duration
}

// MeanLatency is the average call duration, zero when there was no call.
func (s ModelStats) MeanLatency() time.Duration {
	if s.Calls == 0 {
		return 0
	}
	return s.TotalLatency / time.Duration(s.Calls)
}

// Tracker counts calls, failures by class and latency per model, for the free tier whose limits and
// failures must be read before the demo (V2). It records no prompt, no answer and no key: a failure
// log line carries the model, the class, the HTTP status and the latency only.
// A Tracker is safe for concurrent use.
type Tracker struct {
	logger *slog.Logger
	now    func() time.Time

	mu    sync.Mutex
	stats map[string]ModelStats
}

// TrackerOption customizes NewTracker. Options exist for tests only.
type TrackerOption func(*Tracker)

// WithClock replaces the time source.
func WithClock(now func() time.Time) TrackerOption { return func(t *Tracker) { t.now = now } }

// NewTracker returns a Tracker that logs failures to logger. A nil logger discards them.
func NewTracker(logger *slog.Logger, opts ...TrackerOption) *Tracker {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	t := &Tracker{logger: logger, now: time.Now, stats: map[string]ModelStats{}}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Wrap returns a Client that records every call of c under model. Responses and errors pass through
// unchanged.
func (t *Tracker) Wrap(model string, c Client) Client {
	return trackedClient{t: t, model: model, next: c}
}

// Snapshot returns a copy of the counters per model.
func (t *Tracker) Snapshot() map[string]ModelStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[string]ModelStats, len(t.stats))
	for k, v := range t.stats {
		out[k] = v
	}
	return out
}

type trackedClient struct {
	t     *Tracker
	model string
	next  Client
}

func (c trackedClient) Generate(ctx context.Context, r Request) (Response, error) {
	start := c.t.now()
	resp, err := c.next.Generate(ctx, r)
	c.t.record(c.model, err, c.t.now().Sub(start))
	return resp, err
}

func (t *Tracker) record(model string, err error, d time.Duration) {
	class := classify(err)
	status := StatusOf(err)

	t.mu.Lock()
	s := t.stats[model]
	s.Calls++
	s.TotalLatency += d
	s.MaxLatency = max(s.MaxLatency, d)
	switch class {
	case classOK:
		s.OK++
	case classTimeout:
		s.Timeouts++
	case classCanceled:
		s.Canceled++
	case classQuota:
		s.Quota429++
	case classRejected:
		s.Rejected++
	case classNotConfigured:
		s.NotConfigured++
	case classInvalid:
		s.Invalid++
	default:
		switch status {
		case 500:
			s.Unavailable500++
		case 503:
			s.Unavailable503++
		default:
			s.UnavailableOther++
		}
	}
	t.stats[model] = s
	t.mu.Unlock()

	if class != classOK {
		t.logger.Warn("llm call failed", "model", model, "class", class, "status", status, "latency_ms", d.Milliseconds())
	}
}

const (
	classOK            = "ok"
	classTimeout       = "timeout"
	classCanceled      = "canceled"
	classQuota         = "quota"
	classRejected      = "rejected"
	classNotConfigured = "not_configured"
	classInvalid       = "invalid"
	classUnavailable   = "unavailable"
)

// classify maps an error to its class. Anything unexpected counts as unavailable.
func classify(err error) string {
	switch {
	case err == nil:
		return classOK
	case errors.Is(err, context.DeadlineExceeded):
		return classTimeout
	case errors.Is(err, context.Canceled):
		return classCanceled
	case errors.Is(err, ErrQuotaExhausted):
		return classQuota
	case errors.Is(err, ErrRejected):
		return classRejected
	case errors.Is(err, ErrNotConfigured):
		return classNotConfigured
	case errors.Is(err, ErrInvalidRequest):
		return classInvalid
	default:
		return classUnavailable
	}
}
