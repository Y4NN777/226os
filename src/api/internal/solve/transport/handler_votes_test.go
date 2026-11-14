package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
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

const voteBody = `{"target_type":"answer","target_id":"` + aID + `","value":-1}`

func TestCastVoteEncodesTheVoteState(t *testing.T) {
	st := &fakeStore{vote: domain.VoteState{TargetType: "answer", TargetID: aID, Value: -1, Score: 2}}
	rec := httptest.NewRecorder()
	handler(t, st).CastVote(rec, req(t, member, http.MethodPost, "/api/v1/votes", voteBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	got := decode[apigen.VoteState](t, rec)
	if got.TargetType != "answer" || got.TargetId.String() != aID || got.Value != -1 || got.Score != 2 {
		t.Fatalf("state %+v", got)
	}
	if want := (domain.Vote{UserID: userID, TargetType: "answer", TargetID: aID, Value: -1}); st.gotVote != want {
		t.Fatalf("stored %+v, want %+v (the voter comes from the session)", st.gotVote, want)
	}
}

func TestCastVoteRefusesABadBodyAndMapsTheSelfVote(t *testing.T) {
	rec := httptest.NewRecorder()
	handler(t, &fakeStore{}).CastVote(rec, req(t, member, http.MethodPost, "/api/v1/votes", `{"target_type":"answer","bogus":1}`))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown field: status %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{err: domain.ErrSelfVote}).CastVote(rec, req(t, member, http.MethodPost, "/api/v1/votes", voteBody))
	if rec.Code != http.StatusForbidden || problemCode(t, rec) != "SELF_VOTE_FORBIDDEN" {
		t.Fatalf("self vote: %d %s", rec.Code, rec.Body.String())
	}
}

func TestRetractVoteEncodesTheVoteState(t *testing.T) {
	st := &fakeStore{vote: domain.VoteState{TargetType: "question", TargetID: qID, Score: 5}}
	rec := httptest.NewRecorder()
	handler(t, st).RetractVote(rec, req(t, member, http.MethodDelete, "/api/v1/votes", ""),
		apigen.RetractVoteParams{TargetType: "question", TargetId: uuid.MustParse(qID)})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	if got := decode[apigen.VoteState](t, rec); got.Value != 0 || got.Score != 5 || got.TargetType != "question" {
		t.Fatalf("state %+v", got)
	}
	if want := (domain.Vote{UserID: userID, TargetType: "question", TargetID: qID}); st.gotVote != want {
		t.Fatalf("stored %+v, want %+v", st.gotVote, want)
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{err: domain.ErrNotFound}).RetractVote(rec, req(t, member, http.MethodDelete, "/api/v1/votes", ""),
		apigen.RetractVoteParams{TargetType: "question", TargetId: uuid.MustParse(qID)})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing target: %d", rec.Code)
	}
}

func TestAcceptAnswerEncodesTheQuestion(t *testing.T) {
	q := fullQuestion()
	st := &fakeStore{question: q}
	rec := httptest.NewRecorder()
	handler(t, st).AcceptAnswer(rec, req(t, member, http.MethodPost, "/", `{"answer_id":"`+aID+`"}`), uuid.MustParse(qID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	if got := decode[apigen.Question](t, rec); got.AcceptedAnswerId == nil || got.AcceptedAnswerId.String() != aID || got.State != "resolved" {
		t.Fatalf("question %+v", got)
	}
	if want := (domain.AcceptChange{QuestionID: qID, ActorID: userID, AnswerID: aID}); st.gotChange != want {
		t.Fatalf("change %+v, want %+v", st.gotChange, want)
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{}).AcceptAnswer(rec, req(t, member, http.MethodPost, "/", `{"answer_id":"`+aID+`","extra":1}`), uuid.MustParse(qID))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown field: %d", rec.Code)
	}
}

func TestUnacceptAnswerEncodesTheQuestion(t *testing.T) {
	st := &fakeStore{question: plainQuestion()}
	rec := httptest.NewRecorder()
	handler(t, st).UnacceptAnswer(rec, req(t, member, http.MethodPost, "/", ""), uuid.MustParse(qID))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	if got := decode[apigen.Question](t, rec); got.AcceptedAnswerId != nil || got.State != "open" {
		t.Fatalf("question %+v", got)
	}
	if want := (domain.AcceptChange{QuestionID: qID, ActorID: userID}); st.gotChange != want {
		t.Fatalf("change %+v, want %+v", st.gotChange, want)
	}
}

// Negative authz for POST /api/v1/votes (AUTHORIZATION_MATRIX, allow: member with a verified email, not
// own content): anonymous callers get 401, unverified members 403 EMAIL_NOT_VERIFIED, an unknown or removed
// target 404, the author's own content 403 SELF_VOTE_FORBIDDEN, and no vote is written.
func TestAuthzCastVoteAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		err    error
		status int
		code   string
	}{
		{"anonymous", subject.Subject{}, nil, http.StatusUnauthorized, "UNAUTHENTICATED"},
		{"member without a verified email", subject.Subject{UserID: userID}, nil, http.StatusForbidden, "EMAIL_NOT_VERIFIED"},
		{"unknown or removed target", member, domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"own content", member, domain.ErrSelfVote, http.StatusForbidden, "SELF_VOTE_FORBIDDEN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{err: tc.err}
			rec := httptest.NewRecorder()
			handler(t, st).CastVote(rec, req(t, tc.sub, http.MethodPost, "/api/v1/votes", voteBody))
			if rec.Code != tc.status || st.writes != 0 || problemCode(t, rec) != tc.code {
				t.Fatalf("status %d, writes %d; want %d %s", rec.Code, st.writes, tc.status, tc.code)
			}
		})
	}
}

// Negative authz for DELETE /api/v1/votes (allow: member with a verified email): anonymous 401, unverified
// 403, an unknown or removed target 404, and nothing is written.
func TestAuthzRetractVoteAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		err    error
		status int
	}{
		{"anonymous", subject.Subject{}, nil, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: userID}, nil, http.StatusForbidden},
		{"unknown or removed target", member, domain.ErrNotFound, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{err: tc.err}
			rec := httptest.NewRecorder()
			handler(t, st).RetractVote(rec, req(t, tc.sub, http.MethodDelete, "/api/v1/votes", ""),
				apigen.RetractVoteParams{TargetType: "answer", TargetId: uuid.MustParse(aID)})
			if rec.Code != tc.status || st.writes != 0 {
				t.Fatalf("status %d, writes %d; want %d and nothing written", rec.Code, st.writes, tc.status)
			}
		})
	}
}

// Negative authz for POST /api/v1/questions/{questionId}/accept (AUTHORIZATION_MATRIX, allow: owner, the
// question author): anonymous 401, unverified 403, a member or moderator who is not the author 403 and nothing changes,
// an unknown or removed question 404, an answer of another question 422.
func TestAuthzAcceptAnswerNonOwner(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		err    error
		status int
	}{
		{"anonymous", subject.Subject{}, nil, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: userID}, nil, http.StatusForbidden},
		{"other member", subject.Subject{UserID: otherU, EmailVerified: true}, domain.ErrForbidden, http.StatusForbidden},
		{"moderator is not the owner", subject.Subject{UserID: otherU, EmailVerified: true, Roles: []string{"moderator"}}, domain.ErrForbidden, http.StatusForbidden},
		{"unknown or removed question", member, domain.ErrNotFound, http.StatusNotFound},
		{"answer of another question", member, domain.ErrAnswerNotInQuestion, http.StatusUnprocessableEntity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{err: tc.err}
			rec := httptest.NewRecorder()
			handler(t, st).AcceptAnswer(rec, req(t, tc.sub, http.MethodPost, "/api/v1/questions/"+qID+"/accept", `{"answer_id":"`+aID+`"}`), uuid.MustParse(qID))
			if rec.Code != tc.status || st.writes != 0 {
				t.Fatalf("status %d, writes %d; want %d and nothing changed", rec.Code, st.writes, tc.status)
			}
		})
	}
}

// Negative authz for POST /api/v1/questions/{questionId}/unaccept (allow: owner): anonymous 401, unverified
// 403, a member or moderator who is not the author 403 and nothing changes, an unknown or removed question 404.
func TestAuthzUnacceptAnswerNonOwner(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		err    error
		status int
	}{
		{"anonymous", subject.Subject{}, nil, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: userID}, nil, http.StatusForbidden},
		{"other member", subject.Subject{UserID: otherU, EmailVerified: true}, domain.ErrForbidden, http.StatusForbidden},
		{"moderator is not the owner", subject.Subject{UserID: otherU, EmailVerified: true, Roles: []string{"moderator"}}, domain.ErrForbidden, http.StatusForbidden},
		{"unknown or removed question", member, domain.ErrNotFound, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{err: tc.err}
			rec := httptest.NewRecorder()
			handler(t, st).UnacceptAnswer(rec, req(t, tc.sub, http.MethodPost, "/api/v1/questions/"+qID+"/unaccept", ""), uuid.MustParse(qID))
			if rec.Code != tc.status || st.writes != 0 {
				t.Fatalf("status %d, writes %d; want %d and nothing changed", rec.Code, st.writes, tc.status)
			}
		})
	}
}
