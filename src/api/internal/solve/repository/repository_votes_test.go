package repository

import (
	"errors"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
)

func byAuthor(id string) step { return rows([]any{uid(id)}) }
func score(n int32) step      { return rows([]any{n}) }

func TestSetVoteLocksTheTargetThenMovesTheScoreInOneTransaction(t *testing.T) {
	cases := []struct {
		name  string
		vote  domain.Vote
		steps []step
		want  domain.VoteState
	}{
		{"cast on an answer", domain.Vote{UserID: userID, TargetType: "answer", TargetID: aID, Value: 1},
			[]step{byAuthor(otherU), score(4)}, domain.VoteState{TargetType: "answer", TargetID: aID, Value: 1, Score: 4}},
		{"change on a question", domain.Vote{UserID: userID, TargetType: "question", TargetID: qID, Value: -1},
			[]step{byAuthor(otherU), score(-2)}, domain.VoteState{TargetType: "question", TargetID: qID, Value: -1, Score: -2}},
		{"retract on an answer", domain.Vote{UserID: userID, TargetType: "answer", TargetID: aID},
			[]step{byAuthor(otherU), score(3)}, domain.VoteState{TargetType: "answer", TargetID: aID, Score: 3}},
		{"retract on a question", domain.Vote{UserID: userID, TargetType: "question", TargetID: qID},
			[]step{byAuthor(otherU), score(3)}, domain.VoteState{TargetType: "question", TargetID: qID, Score: 3}},
		// Having no vote on your own content is a success: only casting one is refused.
		{"retract on own content", domain.Vote{UserID: userID, TargetType: "question", TargetID: qID},
			[]step{byAuthor(userID), score(3)}, domain.VoteState{TargetType: "question", TargetID: qID, Score: 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := db(t, tc.steps...)
			got, err := New(d).SetVote(t.Context(), tc.vote)
			if err != nil || got != tc.want || !d.committed {
				t.Fatalf("got %+v, %v, committed %v; want %+v", got, err, d.committed, tc.want)
			}
		})
	}
}

func TestSetVoteRefusesOwnContentBeforeAnyWrite(t *testing.T) {
	for _, typ := range []string{"question", "answer"} {
		// Only the lock step is scripted: a write after it would fail the fake.
		d := db(t, byAuthor(userID))
		_, err := New(d).SetVote(t.Context(), domain.Vote{UserID: userID, TargetType: typ, TargetID: qID, Value: 1})
		if !errors.Is(err, domain.ErrSelfVote) || d.committed {
			t.Fatalf("%s: err %v, committed %v", typ, err, d.committed)
		}
	}
}

func TestSetVoteErrors(t *testing.T) {
	v := domain.Vote{UserID: userID, TargetType: "answer", TargetID: aID, Value: 1}
	cases := map[string]struct {
		vote  domain.Vote
		steps []step
		want  error
	}{
		"target missing or removed": {v, []step{rows()}, domain.ErrNotFound},
		"question missing":          {domain.Vote{UserID: userID, TargetType: "question", TargetID: qID, Value: 1}, []step{rows()}, domain.ErrNotFound},
		"lock":                      {v, []step{fail(errDB)}, errDB},
		"question lock":             {domain.Vote{UserID: userID, TargetType: "question", TargetID: qID, Value: 1}, []step{fail(errDB)}, errDB},
		"set answer vote":           {v, []step{byAuthor(otherU), fail(errDB)}, errDB},
		"set question vote":         {domain.Vote{UserID: userID, TargetType: "question", TargetID: qID, Value: 1}, []step{byAuthor(otherU), fail(errDB)}, errDB},
		"bad target id":             {domain.Vote{UserID: userID, TargetType: "answer", TargetID: "bad", Value: 1}, nil, domain.ErrNotFound},
		"unknown target type":       {domain.Vote{UserID: userID, TargetType: "resource", TargetID: aID, Value: 1}, nil, domain.ErrNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := db(t, tc.steps...)
			if _, err := New(d).SetVote(t.Context(), tc.vote); !errors.Is(err, tc.want) || d.committed {
				t.Fatalf("err %v, committed %v", err, d.committed)
			}
		})
	}
	d := db(t, byAuthor(otherU), score(1))
	d.commitErr = errDB
	if _, err := New(d).SetVote(t.Context(), v); !errors.Is(err, errDB) {
		t.Fatalf("commit err %v", err)
	}
	d = db(t)
	d.beginErr = errDB
	if _, err := New(d).SetVote(t.Context(), v); !errors.Is(err, errDB) {
		t.Fatalf("begin err %v", err)
	}
	if _, err := New(db(t)).SetVote(t.Context(), domain.Vote{UserID: "bad", TargetType: "answer", TargetID: aID, Value: 1}); err == nil {
		t.Fatal("bad voter id accepted")
	}
}

// acceptTail is the reload of the question after the change: the question, its tags, its links.
func acceptTail() []step {
	return []step{rows(questionRow("", "", uid(aID))), tagRows(), relatedRows()}
}

func TestSetAcceptedAnswerLocksTheQuestionChecksTheAnswerAndReloads(t *testing.T) {
	change := domain.AcceptChange{QuestionID: qID, ActorID: userID, AnswerID: aID}
	// Accept: lock the question, lock the answer, set, reload.
	d := db(t, append([]step{rows(lockRow()), byAuthor(aID), {}}, acceptTail()...)...)
	got, err := New(d).SetAcceptedAnswer(t.Context(), change)
	if err != nil || got.AcceptedAnswerID != aID || !d.committed {
		t.Fatalf("accept: %+v, %v, committed %v", got, err, d.committed)
	}
	// Clear: no answer lock.
	d = db(t, append([]step{rows(lockRow()), {}}, acceptTail()...)...)
	if _, err := New(d).SetAcceptedAnswer(t.Context(), domain.AcceptChange{QuestionID: qID, ActorID: userID}); err != nil || !d.committed {
		t.Fatalf("clear: %v, committed %v", err, d.committed)
	}
}

func TestSetAcceptedAnswerErrors(t *testing.T) {
	change := domain.AcceptChange{QuestionID: qID, ActorID: userID, AnswerID: aID}
	ok := append([]step{rows(lockRow()), byAuthor(aID), {}}, acceptTail()...)
	with := func(i int, s step) []step { return append(append([]step(nil), ok[:i]...), s) }
	cases := map[string]struct {
		change domain.AcceptChange
		steps  []step
		want   error
	}{
		"question missing or removed": {change, []step{rows()}, domain.ErrNotFound},
		"lock question":               {change, []step{fail(errDB)}, errDB},
		"not the author":              {domain.AcceptChange{QuestionID: qID, ActorID: otherU, AnswerID: aID}, []step{rows(lockRow())}, domain.ErrForbidden},
		"answer elsewhere or removed": {change, []step{rows(lockRow()), rows()}, domain.ErrAnswerNotInQuestion},
		"lock answer":                 {change, with(1, fail(errDB)), errDB},
		"set":                         {change, with(2, fail(errDB)), errDB},
		"reload":                      {change, with(3, fail(errDB)), errDB},
		"bad question id":             {domain.AcceptChange{QuestionID: "bad", ActorID: userID, AnswerID: aID}, nil, domain.ErrNotFound},
		"bad answer id":               {domain.AcceptChange{QuestionID: qID, ActorID: userID, AnswerID: "bad"}, nil, domain.ErrAnswerNotInQuestion},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := db(t, tc.steps...)
			if _, err := New(d).SetAcceptedAnswer(t.Context(), tc.change); !errors.Is(err, tc.want) || d.committed {
				t.Fatalf("err %v, committed %v", err, d.committed)
			}
		})
	}
	d := db(t, ok...)
	d.commitErr = errDB
	if _, err := New(d).SetAcceptedAnswer(t.Context(), change); !errors.Is(err, errDB) {
		t.Fatalf("commit err %v", err)
	}
	d = db(t)
	d.beginErr = errDB
	if _, err := New(d).SetAcceptedAnswer(t.Context(), change); !errors.Is(err, errDB) {
		t.Fatalf("begin err %v", err)
	}
	if _, err := New(db(t)).SetAcceptedAnswer(t.Context(), domain.AcceptChange{QuestionID: qID, ActorID: "bad"}); err == nil {
		t.Fatal("bad actor id accepted")
	}
}
