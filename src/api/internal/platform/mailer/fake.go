package mailer

import (
	"context"
	"slices"
	"sync"
)

// Fake is an in-memory Mailer for tests of other slices. It validates like the real adapter,
// so a test catches a bad message. It is safe for concurrent use.
type Fake struct {
	mu   sync.Mutex
	sent []Message
	err  error
}

// NewFake returns an empty Fake.
func NewFake() *Fake { return &Fake{} }

// Send validates m, then either returns the error set by FailWith or records a copy of m.
func (f *Fake) Send(ctx context.Context, m Message) error {
	if _, err := validate(m); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, clone(m))
	return nil
}

// FailWith makes every following Send return err; FailWith(nil) restores success.
func (f *Fake) FailWith(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

// Sent returns copies of the recorded messages, in order.
func (f *Fake) Sent() []Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Message, len(f.sent))
	for i, m := range f.sent {
		out[i] = clone(m)
	}
	return out
}

func clone(m Message) Message {
	m.To = slices.Clone(m.To)
	return m
}
