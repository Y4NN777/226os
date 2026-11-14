package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
	"github.com/Y4NN777/226os/src/api/internal/solve/service"
)

func (f *fakeStore) SetVote(_ context.Context, v domain.Vote) (domain.VoteState, error) {
	f.gotVote = v
	if f.err != nil {
		return domain.VoteState{}, f.err
	}
	f.writes++
	return f.vote, nil
}

func (f *fakeStore) SetAcceptedAnswer(_ context.Context, c domain.AcceptChange) (domain.Question, error) {
	f.gotChange = c
	if f.err != nil {
		return domain.Question{}, f.err
	}
	f.writes++
	return f.question, nil
}

// castVote: the voter is the session subject, a value is +1 or -1, the target is a question or an answer,
// and every store refusal maps to one problem.
func TestCastVote(t *testing.T) {
	good := service.VoteRequest{TargetType: domain.TargetAnswer, TargetID: aID, Value: 1}
	cases := []struct {
		name  string
		sub   subject.Subject
		req   service.VoteRequest
		store *fakeStore
		deny  bool
		code  apperr.Code
		field string
	}{
		{"anonymous", subject.Subject{}, good, &fakeStore{}, false, apperr.CodeUnauthenticated, ""},
		{"unverified email", unverified, good, &fakeStore{}, false, apperr.CodeEmailNotVerified, ""},
		{"value zero", member, service.VoteRequest{TargetType: domain.TargetAnswer, TargetID: aID, Value: 0}, &fakeStore{}, false, apperr.CodeValidation, "/value"},
		{"value two", member, service.VoteRequest{TargetType: domain.TargetAnswer, TargetID: aID, Value: 2}, &fakeStore{}, false, apperr.CodeValidation, "/value"},
		{"unknown target type", member, service.VoteRequest{TargetType: "resource", TargetID: aID, Value: 1}, &fakeStore{}, false, apperr.CodeValidation, "/target_type"},
		{"malformed target id", member, service.VoteRequest{TargetType: domain.TargetAnswer, TargetID: "nope", Value: 1}, &fakeStore{}, false, apperr.CodeValidation, "/target_id"},
		{"rate limited", member, good, &fakeStore{}, true, apperr.CodeRateLimited, ""},
		{"target missing or removed", member, good, &fakeStore{err: domain.ErrNotFound}, false, apperr.CodeNotFound, ""},
		{"own content", member, good, &fakeStore{err: domain.ErrSelfVote}, false, apperr.CodeSelfVoteForbidden, ""},
		{"database down", member, good, &fakeStore{err: errDB}, false, apperr.CodeInternal, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, lim := newSvc(t, tc.store)
			lim.allowed = !tc.deny
			_, err := svc.CastVote(t.Context(), tc.sub, tc.req)
			if tc.field != "" {
				wantViolation(t, err, tc.field, "invalid_value")
			} else {
				wantCode(t, err, tc.code)
			}
			if tc.store.writes != 0 {
				t.Fatalf("%d writes on a refused vote", tc.store.writes)
			}
		})
	}
}

func TestCastVoteStoresTheSubjectsVoteAndReturnsTheState(t *testing.T) {
	st := &fakeStore{vote: domain.VoteState{TargetType: domain.TargetQuestion, TargetID: qID, Value: -1, Score: 4}}
	svc, lim := newSvc(t, st)
	got, err := svc.CastVote(t.Context(), member, service.VoteRequest{TargetType: domain.TargetQuestion, TargetID: qID, Value: -1})
	if err != nil || got != st.vote {
		t.Fatalf("got %+v, %v", got, err)
	}
	if want := (domain.Vote{UserID: userID, TargetType: domain.TargetQuestion, TargetID: qID, Value: -1}); st.gotVote != want {
		t.Fatalf("stored %+v, want %+v", st.gotVote, want)
	}
	if lim.hits != 1 || lim.scope != "write:user" || lim.value != userID || lim.limit != 60 || lim.window != time.Minute {
		t.Fatalf("limiter %+v", lim)
	}
}

func TestCastVoteRateLimiterFailsClosedAndTellsWhenToRetry(t *testing.T) {
	st := &fakeStore{}
	svc, lim := newSvc(t, st)
	lim.allowed, lim.retry = false, 7*time.Second
	_, err := svc.CastVote(t.Context(), member, service.VoteRequest{TargetType: domain.TargetAnswer, TargetID: aID, Value: 1})
	if ae := codeOf(t, err, apperr.CodeRateLimited); ae.RetryAfter != 7*time.Second {
		t.Fatalf("retry after %v", ae.RetryAfter)
	}
	lim.err = errDB
	_, err = svc.CastVote(t.Context(), member, service.VoteRequest{TargetType: domain.TargetAnswer, TargetID: aID, Value: 1})
	wantCode(t, err, apperr.CodeInternal)
	if st.writes != 0 {
		t.Fatal("a vote was written past the limiter")
	}
}

// retractVote is a vote of value 0, with the same session, target and budget rules.
func TestRetractVote(t *testing.T) {
	cases := []struct {
		name  string
		sub   subject.Subject
		typ   string
		id    string
		store *fakeStore
		deny  bool
		code  apperr.Code
		field string
	}{
		{"anonymous", subject.Subject{}, domain.TargetAnswer, aID, &fakeStore{}, false, apperr.CodeUnauthenticated, ""},
		{"unverified email", unverified, domain.TargetAnswer, aID, &fakeStore{}, false, apperr.CodeEmailNotVerified, ""},
		{"unknown target type", member, "resource", aID, &fakeStore{}, false, apperr.CodeValidation, "target_type"},
		{"malformed target id", member, domain.TargetAnswer, "nope", &fakeStore{}, false, apperr.CodeValidation, "target_id"},
		{"rate limited", member, domain.TargetAnswer, aID, &fakeStore{}, true, apperr.CodeRateLimited, ""},
		{"target missing or removed", member, domain.TargetAnswer, aID, &fakeStore{err: domain.ErrNotFound}, false, apperr.CodeNotFound, ""},
		{"database down", member, domain.TargetAnswer, aID, &fakeStore{err: errDB}, false, apperr.CodeInternal, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, lim := newSvc(t, tc.store)
			lim.allowed = !tc.deny
			_, err := svc.RetractVote(t.Context(), tc.sub, tc.typ, tc.id)
			if tc.field != "" {
				wantViolation(t, err, tc.field, "invalid_value")
			} else {
				wantCode(t, err, tc.code)
			}
		})
	}
	st := &fakeStore{vote: domain.VoteState{TargetType: domain.TargetAnswer, TargetID: aID, Score: 2}}
	svc, _ := newSvc(t, st)
	got, err := svc.RetractVote(t.Context(), member, domain.TargetAnswer, aID)
	if err != nil || got != st.vote {
		t.Fatalf("got %+v, %v", got, err)
	}
	if want := (domain.Vote{UserID: userID, TargetType: domain.TargetAnswer, TargetID: aID}); st.gotVote != want {
		t.Fatalf("stored %+v, want a value-0 vote %+v", st.gotVote, want)
	}
}

// acceptAnswer: only the question author, on a visible question, for one of its own answers.
func TestAcceptAnswer(t *testing.T) {
	cases := []struct {
		name  string
		sub   subject.Subject
		qid   string
		aid   string
		store *fakeStore
		deny  bool
		code  apperr.Code
		field string
	}{
		{"anonymous", subject.Subject{}, qID, aID, &fakeStore{}, false, apperr.CodeUnauthenticated, ""},
		{"unverified email", unverified, qID, aID, &fakeStore{}, false, apperr.CodeEmailNotVerified, ""},
		{"malformed answer id", member, qID, "nope", &fakeStore{}, false, apperr.CodeValidation, "/answer_id"},
		{"rate limited", member, qID, aID, &fakeStore{}, true, apperr.CodeRateLimited, ""},
		{"question missing or removed", member, qID, aID, &fakeStore{err: domain.ErrNotFound}, false, apperr.CodeNotFound, ""},
		{"not the question author", member, qID, aID, &fakeStore{err: domain.ErrForbidden}, false, apperr.CodeForbidden, ""},
		{"answer of another question or removed", member, qID, aID, &fakeStore{err: domain.ErrAnswerNotInQuestion}, false, apperr.CodeValidation, "/answer_id"},
		{"database down", member, qID, aID, &fakeStore{err: errDB}, false, apperr.CodeInternal, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, lim := newSvc(t, tc.store)
			lim.allowed = !tc.deny
			_, err := svc.AcceptAnswer(t.Context(), tc.sub, tc.qid, tc.aid)
			if tc.field != "" {
				wantViolation(t, err, tc.field, "invalid_value")
			} else {
				wantCode(t, err, tc.code)
			}
			if tc.store.writes != 0 {
				t.Fatalf("%d writes on a refused accept", tc.store.writes)
			}
		})
	}
	st := &fakeStore{question: domain.Question{ID: qID, AcceptedAnswerID: aID}}
	svc, lim := newSvc(t, st)
	got, err := svc.AcceptAnswer(t.Context(), member, qID, aID)
	if err != nil || got.AcceptedAnswerID != aID {
		t.Fatalf("got %+v, %v", got, err)
	}
	if want := (domain.AcceptChange{QuestionID: qID, ActorID: userID, AnswerID: aID}); st.gotChange != want {
		t.Fatalf("change %+v, want %+v", st.gotChange, want)
	}
	if lim.scope != "write:user" || lim.value != userID {
		t.Fatalf("limiter %+v", lim)
	}
}

// unacceptAnswer clears the accepted answer: the same change with no answer id.
func TestUnacceptAnswer(t *testing.T) {
	cases := []struct {
		name  string
		sub   subject.Subject
		store *fakeStore
		deny  bool
		code  apperr.Code
	}{
		{"anonymous", subject.Subject{}, &fakeStore{}, false, apperr.CodeUnauthenticated},
		{"unverified email", unverified, &fakeStore{}, false, apperr.CodeEmailNotVerified},
		{"rate limited", member, &fakeStore{}, true, apperr.CodeRateLimited},
		{"question missing or removed", member, &fakeStore{err: domain.ErrNotFound}, false, apperr.CodeNotFound},
		{"not the question author", member, &fakeStore{err: domain.ErrForbidden}, false, apperr.CodeForbidden},
		{"database down", member, &fakeStore{err: errDB}, false, apperr.CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, lim := newSvc(t, tc.store)
			lim.allowed = !tc.deny
			_, err := svc.UnacceptAnswer(t.Context(), tc.sub, qID)
			wantCode(t, err, tc.code)
			if tc.store.writes != 0 {
				t.Fatalf("%d writes on a refused unaccept", tc.store.writes)
			}
		})
	}
	st := &fakeStore{question: domain.Question{ID: qID}}
	svc, _ := newSvc(t, st)
	if _, err := svc.UnacceptAnswer(t.Context(), member, qID); err != nil {
		t.Fatal(err)
	}
	if want := (domain.AcceptChange{QuestionID: qID, ActorID: userID}); st.gotChange != want {
		t.Fatalf("change %+v, want %+v", st.gotChange, want)
	}
}
