package idempotency_test

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

type fakeRow struct {
	method, route string
	hash          [32]byte
	done          bool
	attempt       idempotency.Attempt
	stored        idempotency.Response
}

type fakeStore struct {
	mu        sync.Mutex
	rows      map[string]*fakeRow
	begins    int
	completes int
	abandons  int
	lastResp  idempotency.Response
	lastKey   string
	nextTok   idempotency.Attempt

	completeCtxErr error

	beginErr, completeErr, abandonErr error
}

func newFakeStore() *fakeStore { return &fakeStore{rows: map[string]*fakeRow{}} }

func (f *fakeStore) Begin(_ context.Context, userID, key, method, route string, hash [32]byte, _ time.Time) (idempotency.Outcome, idempotency.Attempt, idempotency.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.begins++
	f.lastKey = key
	if f.beginErr != nil {
		return 0, 0, idempotency.Response{}, f.beginErr
	}
	id := userID + "|" + key
	row, ok := f.rows[id]
	if !ok {
		f.nextTok++
		f.rows[id] = &fakeRow{method: method, route: route, hash: hash, attempt: f.nextTok}
		return idempotency.Started, f.nextTok, idempotency.Response{}, nil
	}
	switch {
	case row.method != method || row.route != route || row.hash != hash:
		return idempotency.Reused, 0, idempotency.Response{}, nil
	case row.done:
		return idempotency.Replay, 0, row.stored, nil
	default:
		return idempotency.InProgress, 0, idempotency.Response{}, nil
	}
}

// takeover simulates the key being taken over by a new owner: the old attempt is stale.
func (f *fakeStore) takeover(userID, key string) idempotency.Attempt {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextTok++
	row := f.rows[userID+"|"+key]
	row.attempt = f.nextTok
	return f.nextTok
}

func (f *fakeStore) Complete(ctx context.Context, userID, key string, attempt idempotency.Attempt, resp idempotency.Response, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.completes++
	f.completeCtxErr = ctx.Err()
	if f.completeErr != nil {
		return f.completeErr
	}
	row := f.rows[userID+"|"+key]
	if row == nil || row.done || row.attempt != attempt {
		return idempotency.ErrLostKey
	}
	resp.Body = bytes.Clone(resp.Body)
	f.lastResp = resp
	row.done = true
	row.stored = resp
	return nil
}

func (f *fakeStore) Abandon(_ context.Context, userID, key string, attempt idempotency.Attempt) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.abandons++
	if f.abandonErr != nil {
		return f.abandonErr
	}
	row := f.rows[userID+"|"+key]
	if row == nil || row.done || row.attempt != attempt {
		return idempotency.ErrLostKey
	}
	delete(f.rows, userID+"|"+key)
	return nil
}

func (f *fakeStore) Purge(context.Context, time.Time) (int64, error) { return 0, nil }
