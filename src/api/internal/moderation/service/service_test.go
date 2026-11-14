package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/moderation/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	actorID  = "019a0000-0000-7000-8000-000000000001"
	targetID = "019a0000-0000-7000-8000-000000000002"
)

type fakeStore struct {
	hidden, restored, resolved                           int
	actor, target, kind, reason, requestID, status, note string
	err                                                  error
	report                                               domain.ErrorReport
}

func (f *fakeStore) Hide(_ context.Context, kind, target, actor, reason, requestID string) error {
	f.hidden++
	f.kind, f.target, f.actor, f.reason, f.requestID = kind, target, actor, reason, requestID
	return f.err
}
func (f *fakeStore) Restore(_ context.Context, kind, target, actor, reason, requestID string) error {
	f.restored++
	f.kind, f.target, f.actor, f.reason, f.requestID = kind, target, actor, reason, requestID
	return f.err
}
func (f *fakeStore) ResolveReport(_ context.Context, report, actor, status, note, requestID string) (domain.ErrorReport, error) {
	f.resolved++
	f.target, f.actor, f.status, f.note, f.requestID = report, actor, status, note, requestID
	return f.report, f.err
}

type fakeLimiter struct {
	allowed      bool
	retry        time.Duration
	err          error
	hits         int
	scope, value string
	limit        int
	window       time.Duration
}

func (f *fakeLimiter) Hit(_ context.Context, scope, value string, limit int, window time.Duration) (bool, time.Duration, error) {
	f.hits++
	f.scope, f.value, f.limit, f.window = scope, value, limit, window
	return f.allowed, f.retry, f.err
}

func code(t *testing.T, err error, want apperr.Code) *apperr.Error {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != want {
		t.Fatalf("error %v, want %s", err, want)
	}
	return ae
}

func TestModerationRequiresModeratorBeforeRateLimitOrWrite(t *testing.T) {
	for _, tc := range []struct {
		name string
		sub  subject.Subject
		want apperr.Code
	}{
		{"anonymous", subject.Subject{}, apperr.CodeUnauthenticated},
		{"member", subject.Subject{UserID: actorID, Roles: []string{"member"}}, apperr.CodeForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st, lim := &fakeStore{}, &fakeLimiter{allowed: true}
			err := New(st, lim).Hide(t.Context(), tc.sub, "question", targetID, "spam", "req-1")
			_ = code(t, err, tc.want)
			if lim.hits != 0 || st.hidden != 0 {
				t.Fatalf("hits %d writes %d", lim.hits, st.hidden)
			}
		})
	}
}

func TestHideRestoreAndResolveUseSharedBudgetAndMapState(t *testing.T) {
	mod := subject.Subject{UserID: actorID, Roles: []string{"moderator"}}
	st := &fakeStore{report: domain.ErrorReport{ID: targetID, ResourceID: actorID, Kind: "outdated", Status: "resolved"}}
	lim := &fakeLimiter{allowed: true}
	svc := New(st, lim)
	if err := svc.Hide(t.Context(), mod, "question", targetID, "spam", "req-1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Restore(t.Context(), mod, "question", targetID, "appel accepté", "req-2"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveReport(t.Context(), mod, targetID, "resolved", "corrigé", "req-3"); err != nil {
		t.Fatal(err)
	}
	if lim.hits != 3 || lim.scope != "write:user" || lim.value != actorID || lim.limit != 60 || lim.window != time.Minute {
		t.Fatalf("limiter %+v", lim)
	}
	if st.hidden != 1 || st.restored != 1 || st.resolved != 1 || st.actor != actorID {
		t.Fatalf("store %+v", st)
	}

	st.err = domain.ErrStateConflict
	_ = code(t, svc.Hide(t.Context(), mod, "question", targetID, "spam", "req-4"), apperr.CodeStateConflict)
	st.err = domain.ErrNotFound
	_ = code(t, svc.Restore(t.Context(), mod, "question", targetID, "spam", "req-5"), apperr.CodeNotFound)
}

func TestModerationRateLimitStopsTheWrite(t *testing.T) {
	st := &fakeStore{}
	lim := &fakeLimiter{allowed: false, retry: 7 * time.Second}
	err := New(st, lim).Hide(t.Context(), subject.Subject{UserID: actorID, Roles: []string{"moderator"}}, "question", targetID, "spam", "req")
	if ae := code(t, err, apperr.CodeRateLimited); ae.RetryAfter != 7*time.Second {
		t.Fatalf("retry %v", ae.RetryAfter)
	}
	if st.hidden != 0 {
		t.Fatal("rate-limited request wrote")
	}
}

func TestModerationFailurePaths(t *testing.T) {
	mod := subject.Subject{UserID: actorID, Roles: []string{"moderator"}}
	storeErr := errors.New("database unavailable")

	lim := &fakeLimiter{allowed: true, err: storeErr}
	_ = code(t, New(&fakeStore{}, lim).Hide(t.Context(), mod, "question", targetID, "spam", "req"), apperr.CodeInternal)

	st := &fakeStore{err: storeErr}
	_ = code(t, New(st, &fakeLimiter{allowed: true}).Restore(t.Context(), mod, "question", targetID, "reason", "req"), apperr.CodeInternal)
	_ = code(t, New(st, &fakeLimiter{allowed: true}).Restore(t.Context(), subject.Subject{UserID: actorID, Roles: []string{"member"}}, "question", targetID, "reason", "req"), apperr.CodeForbidden)

	st.err = domain.ErrStateConflict
	_, err := New(st, &fakeLimiter{allowed: true}).ResolveReport(t.Context(), mod, targetID, "resolved", "note", "req")
	_ = code(t, err, apperr.CodeStateConflict)
	_, err = New(st, &fakeLimiter{allowed: true}).ResolveReport(t.Context(), subject.Subject{}, targetID, "resolved", "note", "req")
	_ = code(t, err, apperr.CodeUnauthenticated)
}
