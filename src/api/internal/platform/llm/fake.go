package llm

import (
	"context"
	"slices"
	"sync"
)

// Fake is an in-memory Client for tests of other slices. It validates like the real adapter, so a
// test catches a bad request. It is safe for concurrent use.
//
// Precedence on each call: FailWith, then the queue, then Respond, then a fixed default answer.
// Set Respond before sharing the Fake between goroutines.
type Fake struct {
	// Respond scripts the answer when the queue is empty. It runs outside the Fake lock.
	Respond func(Request) (Response, error)

	mu    sync.Mutex
	queue []queued
	err   error
	reqs  []Request
}

type queued struct {
	resp Response
	err  error
}

// NewFake returns a Fake whose default answer is a short fixed text.
func NewFake() *Fake { return &Fake{} }

// Enqueue appends a response returned by a later call, in order.
func (f *Fake) Enqueue(r Response) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queue = append(f.queue, queued{resp: r})
}

// EnqueueError appends an error returned by a later call, in order.
func (f *Fake) EnqueueError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queue = append(f.queue, queued{err: err})
}

// FailWith makes every following call return err; FailWith(nil) restores normal behaviour.
func (f *Fake) FailWith(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

// Requests returns copies of the requests received, in order. Rejected requests are not recorded.
func (f *Fake) Requests() []Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.reqs)
}

// Generate validates r, records it, then answers as described on Fake.
func (f *Fake) Generate(ctx context.Context, r Request) (Response, error) {
	r, err := normalize(r)
	if err != nil {
		return Response{}, err
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	o := f.next(r)
	if o.respond != nil {
		return o.respond(r)
	}
	return o.resp, o.err
}

// outcome is what one call returns, or the Respond callback it must run outside the lock.
type outcome struct {
	resp    Response
	err     error
	respond func(Request) (Response, error)
}

// next records r and picks the scripted outcome under the lock.
func (f *Fake) next(r Request) outcome {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reqs = append(f.reqs, r)
	if f.err != nil {
		return outcome{err: f.err}
	}
	if len(f.queue) > 0 {
		q := f.queue[0]
		f.queue = f.queue[1:]
		return outcome{resp: q.resp, err: q.err}
	}
	if f.Respond != nil {
		return outcome{respond: f.Respond}
	}
	return outcome{resp: Response{Text: "fake answer", Model: "fake"}}
}
