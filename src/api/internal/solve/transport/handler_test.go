package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
	"github.com/Y4NN777/226os/src/api/internal/solve/service"
	"github.com/Y4NN777/226os/src/api/internal/solve/transport"
)

const (
	qID    = "00000000-0000-7000-8000-0000000000d1"
	aID    = "00000000-0000-7000-8000-0000000000a1"
	userID = "00000000-0000-7000-8000-0000000000b1"
	otherU = "00000000-0000-7000-8000-0000000000b2"
)

var (
	when   = time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	alice  = domain.Author{ID: userID, Username: "alice", DisplayName: "Alice", Country: "BF"}
	bob    = domain.Author{ID: otherU, Username: "bob", DisplayName: "Bob"}
	member = subject.Subject{UserID: userID, EmailVerified: true}
)

type fakeStore struct {
	questions []domain.Question
	answers   []domain.Answer
	question  domain.Question
	answer    domain.Answer
	err       error
	guardAs   domain.Question

	created domain.NewQuestion
	patch   domain.QuestionPatch
	newAns  domain.NewAnswer
	writes  int

	vote      domain.VoteState
	gotVote   domain.Vote
	gotChange domain.AcceptChange
}

func (f *fakeStore) ListQuestions(context.Context, domain.QuestionQuery) ([]domain.Question, error) {
	return f.questions, f.err
}
func (f *fakeStore) GetQuestion(context.Context, string, string) (domain.Question, error) {
	return f.question, f.err
}
func (f *fakeStore) CreateQuestion(_ context.Context, in domain.NewQuestion) (domain.Question, error) {
	f.created, f.writes = in, f.writes+1
	return f.question, f.err
}
func (f *fakeStore) UpdateQuestion(_ context.Context, _ string, p domain.QuestionPatch, guard func(domain.Question) error) (domain.Question, error) {
	f.patch = p
	if errors.Is(f.err, domain.ErrNotFound) { // the real store finds the row before it calls the guard
		return domain.Question{}, f.err
	}
	if err := guard(f.guardAs); err != nil {
		return domain.Question{}, err
	}
	f.writes++
	return f.question, f.err
}
func (f *fakeStore) ListAnswers(context.Context, domain.AnswerQuery) ([]domain.Answer, error) {
	return f.answers, f.err
}
func (f *fakeStore) CreateAnswer(_ context.Context, in domain.NewAnswer) (domain.Answer, error) {
	f.newAns, f.writes = in, f.writes+1
	return f.answer, f.err
}

type allowLimiter struct{}

func (allowLimiter) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

func handler(t *testing.T, st *fakeStore) *transport.Handler {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return transport.New(service.New(st, c, allowLimiter{}), nil)
}

func req(t *testing.T, sub subject.Subject, method, path, body string, headers ...string) *http.Request {
	t.Helper()
	r := httptest.NewRequestWithContext(subject.With(t.Context(), sub), method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		r.Header.Set(headers[i], headers[i+1])
	}
	return r
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %d %s: %v", rec.Code, rec.Body.String(), err)
	}
	return v
}

func problemCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type %q: %s", ct, rec.Body.String())
	}
	return decode[map[string]any](t, rec)["code"].(string)
}

func fullQuestion() domain.Question {
	return domain.Question{
		ID: qID, Title: "Titre", BodyMD: "Corps", Lang: "fr", Country: "BF", Tags: []string{"go"},
		Snippets: []domain.Snippet{{Language: "go", Code: "x"}}, Author: alice, AnswerCount: 2, VoteScore: 3,
		AcceptedAnswerID: aID, Related: []domain.Link{{Type: "resource", ID: aID, Title: "Alpha", Lang: "fr"}},
		CreatedAt: when, UpdatedAt: when,
	}
}

func plainQuestion() domain.Question {
	return domain.Question{ID: qID, Title: "Titre", BodyMD: "Corps", Lang: "en", Tags: []string{}, Snippets: []domain.Snippet{}, Author: bob, Related: []domain.Link{}, CreatedAt: when, UpdatedAt: when}
}

func TestListQuestionsEncodesThePage(t *testing.T) {
	st := &fakeStore{questions: []domain.Question{fullQuestion(), plainQuestion()}}
	limit, tag, q, country := 1, "go", "paiement", "BF"
	state, lang := apigen.QuestionState("open"), apigen.LangFilter("fr")
	rec := httptest.NewRecorder()
	handler(t, st).ListQuestions(rec, req(t, subject.Subject{}, http.MethodGet, "/api/v1/questions", ""),
		apigen.ListQuestionsParams{Limit: &limit, Tag: &tag, Q: &q, State: &state, Country: &country, Lang: &lang})
	page := decode[apigen.QuestionPage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || page.NextCursor == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	got := page.Items[0]
	if got.Id.String() != qID || got.State != "resolved" || *got.Country != "BF" || got.AcceptedAnswerId.String() != aID ||
		got.Author.Username != "alice" || *got.Author.Country != "BF" || len(got.Related) != 1 || got.Related[0].Type != "resource" ||
		got.Snippets[0].Code != "x" || got.AnswerCount != 2 || got.VoteScore != 3 {
		t.Fatalf("question %+v", got)
	}
	if rec.Header().Get("Vary") != "Cookie" {
		t.Fatalf("list Vary = %q, want Cookie", rec.Header().Get("Vary"))
	}

	rec = httptest.NewRecorder()
	st.questions = []domain.Question{plainQuestion()}
	handler(t, st).ListQuestions(rec, req(t, subject.Subject{}, http.MethodGet, "/api/v1/questions", ""), apigen.ListQuestionsParams{})
	plain := decode[apigen.QuestionPage](t, rec)
	if plain.NextCursor != nil || plain.Items[0].Country != nil || plain.Items[0].AcceptedAnswerId != nil || plain.Items[0].State != "open" || plain.Items[0].Author.Country != nil {
		t.Fatalf("plain %+v", plain)
	}

	rec = httptest.NewRecorder()
	bad := "garbage"
	handler(t, st).ListQuestions(rec, req(t, subject.Subject{}, http.MethodGet, "/api/v1/questions", ""), apigen.ListQuestionsParams{Cursor: &bad})
	if rec.Code != http.StatusBadRequest || problemCode(t, rec) != "INVALID_CURSOR" {
		t.Fatalf("cursor: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	zero := 0
	handler(t, st).ListQuestions(rec, req(t, subject.Subject{}, http.MethodGet, "/api/v1/questions", ""), apigen.ListQuestionsParams{Limit: &zero})
	if rec.Code != http.StatusUnprocessableEntity || problemCode(t, rec) != "VALIDATION_ERROR" {
		t.Fatalf("limit 0: %d", rec.Code)
	}
}

func TestGetQuestionSendsTheETagAndHidesWhatIsNotPublished(t *testing.T) {
	st := &fakeStore{question: fullQuestion()}
	rec := httptest.NewRecorder()
	handler(t, st).GetQuestion(rec, req(t, subject.Subject{}, http.MethodGet, "/", ""), uuid.MustParse(qID))
	if rec.Code != http.StatusOK || rec.Header().Get("ETag") != service.ETag(fullQuestion(), "") {
		t.Fatalf("%d etag %q", rec.Code, rec.Header().Get("ETag"))
	}
	if rec.Header().Get("Vary") != "Cookie" || rec.Header().Get("Cache-Control") != "private, no-cache" {
		t.Fatalf("a viewer-scoped answer: Vary %q, Cache-Control %q", rec.Header().Get("Vary"), rec.Header().Get("Cache-Control"))
	}
	// The tag and the vote are the caller's: a member gets her own tag and the vote the store read for her.
	st.question.ViewerVote = -1
	rec = httptest.NewRecorder()
	handler(t, st).GetQuestion(rec, req(t, member, http.MethodGet, "/", ""), uuid.MustParse(qID))
	if rec.Header().Get("ETag") != service.ETag(fullQuestion(), member.UserID) || rec.Header().Get("ETag") == service.ETag(fullQuestion(), "") {
		t.Fatalf("etag %q does not cover the viewer", rec.Header().Get("ETag"))
	}
	if got := decode[apigen.Question](t, rec); got.ViewerVote != apigen.ViewerVoteMinus1 {
		t.Fatalf("viewer_vote = %d, want -1", got.ViewerVote)
	}
	st.question.ViewerVote = 0
	st.err = domain.ErrNotFound
	rec = httptest.NewRecorder()
	handler(t, st).GetQuestion(rec, req(t, subject.Subject{}, http.MethodGet, "/", ""), uuid.MustParse(qID))
	if rec.Code != http.StatusNotFound || problemCode(t, rec) != "NOT_FOUND" || rec.Header().Get("ETag") != "" {
		t.Fatalf("%d", rec.Code)
	}
}

const createBody = `{"title":"Comment encaisser ?","body_md":"Corps","lang":"fr","tags":["go"],"country":"BF","snippets":[{"language":"go","code":"x"}],"links":[{"type":"resource","id":"` + aID + `"}]}`

func TestCreateQuestionMapsTheBodyAndLocatesTheQuestion(t *testing.T) {
	st := &fakeStore{question: fullQuestion()}
	rec := httptest.NewRecorder()
	handler(t, st).CreateQuestion(rec, req(t, member, http.MethodPost, "/api/v1/questions", createBody), apigen.CreateQuestionParams{})
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/questions/"+qID {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	in := st.created
	if in.AuthorID != userID || in.Title != "Comment encaisser ?" || in.Country != "BF" || len(in.Snippets) != 1 || len(in.Links) != 1 || in.Links[0].Type != "resource" {
		t.Fatalf("created %+v", in)
	}

	// The author comes from the session, never from the body; unknown fields are refused.
	rec = httptest.NewRecorder()
	handler(t, st).CreateQuestion(rec, req(t, member, http.MethodPost, "/api/v1/questions", `{"title":"Titre","body_md":"C","lang":"fr","tags":["go"],"author_id":"`+otherU+`"}`), apigen.CreateQuestionParams{})
	if rec.Code != http.StatusUnprocessableEntity || st.created.AuthorID != userID {
		t.Fatalf("author in body: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	handler(t, st).CreateQuestion(rec, req(t, member, http.MethodPost, "/api/v1/questions", `{`), apigen.CreateQuestionParams{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	// A question without links, snippets or country needs none of them.
	rec = httptest.NewRecorder()
	handler(t, st).CreateQuestion(rec, req(t, member, http.MethodPost, "/api/v1/questions", `{"title":"Titre","body_md":"C","lang":"fr","tags":["go"]}`), apigen.CreateQuestionParams{})
	if rec.Code != http.StatusCreated || len(st.created.Links) != 0 || len(st.created.Snippets) != 0 || st.created.Country != "" {
		t.Fatalf("minimal: %d %+v", rec.Code, st.created)
	}
	st.err = errors.New("db down")
	rec = httptest.NewRecorder()
	handler(t, st).CreateQuestion(rec, req(t, member, http.MethodPost, "/api/v1/questions", createBody), apigen.CreateQuestionParams{})
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "db down") {
		t.Fatalf("store error leaked: %d %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateQuestionHandsTheIfMatchCheckToTheService(t *testing.T) {
	locked := domain.Question{ID: qID, Author: alice, UpdatedAt: when}
	tag := service.ETag(locked, member.UserID)
	patch := func(st *fakeStore, sub subject.Subject, body string, headers ...string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler(t, st).UpdateQuestion(rec, req(t, sub, http.MethodPatch, "/api/v1/questions/"+qID, body, headers...), uuid.MustParse(qID), apigen.UpdateQuestionParams{})
		return rec
	}

	st := &fakeStore{question: fullQuestion(), guardAs: locked}
	rec := patch(st, member, `{"title":"Nouveau titre","body_md":"Nouveau","tags":["go"],"snippets":[{"language":"go","code":"x"}],"links":[{"type":"project","id":"`+aID+`"}]}`, "If-Match", tag)
	if rec.Code != http.StatusOK || rec.Header().Get("ETag") != service.ETag(fullQuestion(), member.UserID) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	p := st.patch
	if *p.Title != "Nouveau titre" || *p.BodyMD != "Nouveau" || len(*p.Tags) != 1 || len(*p.Snippets) != 1 || len(*p.Links) != 1 || p.Country != nil {
		t.Fatalf("patch %+v", p)
	}

	// country: null clears it, a value sets it, absent leaves it.
	patch(st, member, `{"country":null}`, "If-Match", tag)
	if st.patch.Country == nil || *st.patch.Country != "" {
		t.Fatalf("null country: %+v", st.patch.Country)
	}
	patch(st, member, `{"country":"GH"}`, "If-Match", tag)
	if st.patch.Country == nil || *st.patch.Country != "GH" {
		t.Fatalf("country: %+v", st.patch.Country)
	}
	patch(st, member, `{"title":"Autre titre"}`, "If-Match", tag)
	if st.patch.Country != nil {
		t.Fatalf("absent country: %+v", st.patch.Country)
	}

	// 428 without If-Match, 412 with a stale one; nothing is written.
	st.writes = 0
	if rec := patch(st, member, `{"title":"Autre titre"}`); rec.Code != http.StatusPreconditionRequired || problemCode(t, rec) != "PRECONDITION_REQUIRED" {
		t.Fatalf("missing: %d", rec.Code)
	}
	if rec := patch(st, member, `{"title":"Autre titre"}`, "If-Match", `"stale"`); rec.Code != http.StatusPreconditionFailed || problemCode(t, rec) != "PRECONDITION_FAILED" {
		t.Fatalf("stale: %d", rec.Code)
	}
	if st.writes != 0 {
		t.Fatal("a refused precondition wrote")
	}
	// Unknown question 404, unreadable body, anonymous.
	st.err = domain.ErrNotFound
	if rec := patch(st, member, `{"title":"Autre titre"}`, "If-Match", tag); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown: %d", rec.Code)
	}
	if rec := patch(st, member, `{"title":`, "If-Match", tag); rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	if rec := patch(st, member, `{"country":5}`, "If-Match", tag); rec.Code != http.StatusBadRequest {
		t.Fatalf("country number: %d", rec.Code)
	}
}

// Negative authz for PATCH /api/v1/questions/{questionId} (AUTHORIZATION_MATRIX, allow: owner): anonymous
// callers get 401, a member who is not the author gets 403 on a visible question whatever the tag says,
// an unknown or removed question is 404, and nothing is written.
func TestAuthzUpdateQuestionNonOwner(t *testing.T) {
	locked := domain.Question{ID: qID, Author: alice, UpdatedAt: when}
	cases := []struct {
		name    string
		sub     subject.Subject
		store   *fakeStore
		ifMatch string
		status  int
	}{
		{"anonymous", subject.Subject{}, &fakeStore{guardAs: locked}, service.ETag(locked, otherU), http.StatusUnauthorized},
		{"other member, right tag", subject.Subject{UserID: otherU, EmailVerified: true}, &fakeStore{guardAs: locked}, service.ETag(locked, otherU), http.StatusForbidden},
		{"other member, stale tag", subject.Subject{UserID: otherU, EmailVerified: true}, &fakeStore{guardAs: locked}, `"stale"`, http.StatusForbidden},
		{"unknown or removed question, stale tag", member, &fakeStore{guardAs: locked, err: domain.ErrNotFound}, `"stale"`, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r := req(t, tc.sub, http.MethodPatch, "/api/v1/questions/"+qID, `{"title":"Autre titre"}`, "If-Match", tc.ifMatch)
			handler(t, tc.store).UpdateQuestion(rec, r, uuid.MustParse(qID), apigen.UpdateQuestionParams{})
			if rec.Code != tc.status || tc.store.writes != 0 {
				t.Fatalf("status %d, writes %d; want %d and nothing written", rec.Code, tc.store.writes, tc.status)
			}
		})
	}
}

// Negative authz for POST /api/v1/questions (AUTHORIZATION_MATRIX, allow: member with a verified email):
// anonymous callers get 401, unverified members 403, and nothing is created.
func TestAuthzCreateQuestionAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		status int
	}{
		{"anonymous", subject.Subject{}, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: userID}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{}
			rec := httptest.NewRecorder()
			handler(t, st).CreateQuestion(rec, req(t, tc.sub, http.MethodPost, "/api/v1/questions", createBody), apigen.CreateQuestionParams{})
			if rec.Code != tc.status || st.writes != 0 {
				t.Fatalf("status %d, writes %d; want %d and nothing created", rec.Code, st.writes, tc.status)
			}
		})
	}
}

// Negative authz for POST /api/v1/questions/{questionId}/answers (allow: member with a verified email;
// the question must be visible): anonymous callers get 401, unverified members 403, an unknown or
// removed question 404, and nothing is created.
func TestAuthzCreateAnswerAnonymous(t *testing.T) {
	cases := []struct {
		name   string
		sub    subject.Subject
		err    error
		status int
	}{
		{"anonymous", subject.Subject{}, nil, http.StatusUnauthorized},
		{"member without a verified email", subject.Subject{UserID: userID}, nil, http.StatusForbidden},
		{"unknown or removed question", member, domain.ErrNotFound, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{err: tc.err}
			rec := httptest.NewRecorder()
			handler(t, st).CreateAnswer(rec, req(t, tc.sub, http.MethodPost, "/api/v1/questions/"+qID+"/answers", `{"body_md":"x","lang":"fr"}`), uuid.MustParse(qID), apigen.CreateAnswerParams{})
			if rec.Code != tc.status || (tc.err == nil && st.writes != 0) {
				t.Fatalf("status %d, writes %d; want %d and nothing created", rec.Code, st.writes, tc.status)
			}
		})
	}
}

func TestCreateAnswerMapsTheBodyAndLocatesTheAnswers(t *testing.T) {
	st := &fakeStore{answer: domain.Answer{ID: aID, QuestionID: qID, BodyMD: "Réponse", Lang: "fr", Snippets: []domain.Snippet{}, Sources: []string{"https://a.example"}, Author: bob, VoteScore: 1, CreatedAt: when, UpdatedAt: when}}
	rec := httptest.NewRecorder()
	handler(t, st).CreateAnswer(rec, req(t, member, http.MethodPost, "/", `{"body_md":"Réponse","lang":"fr","sources":["https://a.example"]}`), uuid.MustParse(qID), apigen.CreateAnswerParams{})
	a := decode[apigen.Answer](t, rec)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/questions/"+qID+"/answers" || a.Id.String() != aID || a.QuestionId.String() != qID ||
		len(a.Sources) != 1 || a.IsAccepted || a.VoteScore != 1 || a.Snippets == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if st.newAns.AuthorID != userID || st.newAns.QuestionID != qID || len(st.newAns.Sources) != 1 {
		t.Fatalf("stored %+v", st.newAns)
	}
	rec = httptest.NewRecorder()
	handler(t, st).CreateAnswer(rec, req(t, member, http.MethodPost, "/", `{"body_md":"Réponse"`), uuid.MustParse(qID), apigen.CreateAnswerParams{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	handler(t, st).CreateAnswer(rec, req(t, member, http.MethodPost, "/", `{"body_md":"x","lang":"fr"}`), uuid.MustParse(qID), apigen.CreateAnswerParams{})
	if rec.Code != http.StatusCreated || len(st.newAns.Sources) != 0 {
		t.Fatalf("no sources: %d %+v", rec.Code, st.newAns)
	}
}

func TestListAnswersEncodesThePage(t *testing.T) {
	st := &fakeStore{answers: []domain.Answer{
		{ID: aID, QuestionID: qID, BodyMD: "A", Lang: "fr", Snippets: []domain.Snippet{}, Sources: []string{}, Author: alice, Accepted: true, VoteScore: 4, ViewerVote: 1, CreatedAt: when, UpdatedAt: when},
		{ID: userID, QuestionID: qID, BodyMD: "B", Lang: "fr", Snippets: []domain.Snippet{}, Sources: []string{}, Author: bob, CreatedAt: when, UpdatedAt: when},
	}}
	limit := 1
	rec := httptest.NewRecorder()
	handler(t, st).ListAnswers(rec, req(t, subject.Subject{}, http.MethodGet, "/", ""), uuid.MustParse(qID), apigen.ListAnswersParams{Limit: &limit})
	page := decode[apigen.AnswerPage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || !page.Items[0].IsAccepted || page.Items[0].VoteScore != 4 || page.Items[0].ViewerVote != apigen.ViewerVoteN1 || page.NextCursor == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Vary") != "Cookie" {
		t.Fatalf("answers Vary = %q, want Cookie", rec.Header().Get("Vary"))
	}
	st.err = domain.ErrNotFound
	rec = httptest.NewRecorder()
	handler(t, st).ListAnswers(rec, req(t, subject.Subject{}, http.MethodGet, "/", ""), uuid.MustParse(qID), apigen.ListAnswersParams{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown question: %d", rec.Code)
	}
}
