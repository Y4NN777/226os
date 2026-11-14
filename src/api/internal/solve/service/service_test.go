package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
	"github.com/Y4NN777/226os/src/api/internal/solve/service"
)

const (
	qID    = "00000000-0000-7000-8000-0000000000d1"
	aID    = "00000000-0000-7000-8000-0000000000a1"
	userID = "00000000-0000-7000-8000-0000000000b1"
	otherU = "00000000-0000-7000-8000-0000000000b2"
)

var (
	errDB = errors.New("db down")
	when  = time.Date(2026, 11, 14, 9, 0, 0, 123456000, time.UTC)

	member     = subject.Subject{UserID: userID, EmailVerified: true}
	unverified = subject.Subject{UserID: userID}
)

// fakeStore is a hand-written Store. guardAs is the locked row UpdateQuestion hands to the guard.
type fakeStore struct {
	questions []domain.Question
	answers   []domain.Answer
	question  domain.Question
	answer    domain.Answer
	err       error
	guardAs   domain.Question

	listQ    domain.QuestionQuery
	listA    domain.AnswerQuery
	created  domain.NewQuestion
	patched  domain.QuestionPatch
	newAns   domain.NewAnswer
	guardErr error
	writes   int

	vote    domain.VoteState
	gotVote domain.Vote
	// viewerSeen is the viewer id GetQuestion was called with.
	viewerSeen string
	gotChange  domain.AcceptChange
}

func (f *fakeStore) ListQuestions(_ context.Context, q domain.QuestionQuery) ([]domain.Question, error) {
	f.listQ = q
	return f.questions, f.err
}

func (f *fakeStore) GetQuestion(_ context.Context, _, viewer string) (domain.Question, error) {
	f.viewerSeen = viewer
	return f.question, f.err
}

func (f *fakeStore) CreateQuestion(_ context.Context, in domain.NewQuestion) (domain.Question, error) {
	f.created, f.writes = in, f.writes+1
	return f.question, f.err
}

func (f *fakeStore) UpdateQuestion(_ context.Context, _ string, p domain.QuestionPatch, guard func(domain.Question) error) (domain.Question, error) {
	f.patched = p
	if f.guardErr = guard(f.guardAs); f.guardErr != nil {
		return domain.Question{}, f.guardErr
	}
	f.writes++
	return f.question, f.err
}

func (f *fakeStore) ListAnswers(_ context.Context, q domain.AnswerQuery) ([]domain.Answer, error) {
	f.listA = q
	return f.answers, f.err
}

func (f *fakeStore) CreateAnswer(_ context.Context, in domain.NewAnswer) (domain.Answer, error) {
	f.newAns, f.writes = in, f.writes+1
	return f.answer, f.err
}

type fakeLimiter struct {
	allowed bool
	retry   time.Duration
	err     error
	hits    int
	scope   string
	limit   int
	window  time.Duration
	value   string
}

func (l *fakeLimiter) Hit(_ context.Context, scope, value string, limit int, window time.Duration) (bool, time.Duration, error) {
	l.hits++
	l.scope, l.value, l.limit, l.window = scope, value, limit, window
	return l.allowed, l.retry, l.err
}

func codec(t *testing.T) *cursor.Codec {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newSvc(t *testing.T, st *fakeStore) (*service.Service, *fakeLimiter) {
	t.Helper()
	lim := &fakeLimiter{allowed: true}
	return service.New(st, codec(t), lim), lim
}

func wantCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	if ae := codeOf(t, err, code); ae == nil {
		t.Fatal("no problem")
	}
}

func codeOf(t *testing.T, err error, code apperr.Code) *apperr.Error {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != code {
		t.Fatalf("err = %v, want code %s", err, code)
	}
	return ae
}

func wantViolation(t *testing.T, err error, field, code string) {
	t.Helper()
	ae := codeOf(t, err, apperr.CodeValidation)
	if len(ae.Violations) != 1 || ae.Violations[0].Field != field || ae.Violations[0].Code != code {
		t.Fatalf("violations = %+v, want %s %s", ae.Violations, field, code)
	}
}

func ptr[T any](v T) *T { return &v }

func question(id string, created time.Time) domain.Question {
	return domain.Question{ID: id, Title: "Titre", BodyMD: "Corps", Lang: "fr", CreatedAt: created, UpdatedAt: created, Author: domain.Author{ID: userID}}
}

func TestListQuestionsPagesAndBindsTheCursorToTheFilters(t *testing.T) {
	rows := []domain.Question{question(qID, when), question(aID, when.Add(-time.Hour)), question(otherU, when.Add(-2*time.Hour))}
	st := &fakeStore{questions: rows}
	svc, _ := newSvc(t, st)
	limit := 2

	page, err := svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Limit: &limit, Tag: ptr("go"), State: ptr("open"), Country: ptr("BF"), Lang: ptr("fr"), Q: ptr("paiement")})
	if err != nil || len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("page = %+v, %v", page, err)
	}
	if st.listQ.Limit != 3 || st.listQ.Tag != "go" || st.listQ.State != "open" || st.listQ.Country != "BF" || st.listQ.Lang != "fr" || st.listQ.Q != "paiement" || st.listQ.AfterCreated != nil {
		t.Fatalf("query = %+v", st.listQ)
	}

	// The next page resumes after the last row, with the same filters.
	_, err = svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Limit: &limit, Tag: ptr("go"), State: ptr("open"), Country: ptr("BF"), Lang: ptr("fr"), Q: ptr("paiement"), Cursor: page.NextCursor})
	if err != nil || st.listQ.AfterID != aID || st.listQ.AfterCreated == nil || !st.listQ.AfterCreated.Equal(when.Add(-time.Hour)) {
		t.Fatalf("after = %v %v, %v", st.listQ.AfterCreated, st.listQ.AfterID, err)
	}
	// The same cursor with another filter is refused.
	_, err = svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Limit: &limit, Cursor: page.NextCursor})
	wantCode(t, err, apperr.CodeInvalidCursor)

	// A short page has no cursor; an empty result is an empty list, never nil.
	st.questions = nil
	page, err = svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{})
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.NextCursor != "" || st.listQ.Limit != 21 {
		t.Fatalf("empty page = %+v, %v (limit %d)", page, err, st.listQ.Limit)
	}
}

func TestListQuestionsRejectsBadInput(t *testing.T) {
	c := codec(t)
	tests := []struct {
		name  string
		req   service.QuestionListRequest
		field string
		code  string
	}{
		{"empty q", service.QuestionListRequest{Q: ptr("")}, "q", "out_of_range"},
		{"long q", service.QuestionListRequest{Q: ptr(strings.Repeat("a", 201))}, "q", "out_of_range"},
		{"NUL in q", service.QuestionListRequest{Q: ptr("a\x00b")}, "q", "invalid_value"},
		{"limit 0", service.QuestionListRequest{Limit: ptr(0)}, "limit", "out_of_range"},
		{"limit 51", service.QuestionListRequest{Limit: ptr(51)}, "limit", "out_of_range"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newSvc(t, &fakeStore{})
			_, err := svc.ListQuestions(t.Context(), subject.Subject{}, tc.req)
			wantViolation(t, err, tc.field, tc.code)
		})
	}
	svc, _ := newSvc(t, &fakeStore{})
	_, err := svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Cursor: "garbage"})
	wantCode(t, err, apperr.CodeInvalidCursor)

	// A validly signed cursor with a bad sort key or id is as invalid as a forged one.
	for _, p := range []cursor.Payload{{Sort: "yesterday", ID: qID}, {Sort: when.Format(time.RFC3339Nano), ID: "nope"}} {
		hash := "8:question0:0:0:0:0:"
		p.FilterHash = hash
		_, err = svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Cursor: c.Encode(p)})
		wantCode(t, err, apperr.CodeInvalidCursor)
	}

	st := &fakeStore{err: errDB}
	svc, _ = newSvc(t, st)
	_, err = svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{})
	wantCode(t, err, apperr.CodeInternal)
}

func TestGetQuestion(t *testing.T) {
	st := &fakeStore{question: question(qID, when)}
	svc, _ := newSvc(t, st)
	got, err := svc.GetQuestion(t.Context(), subject.Subject{}, qID)
	if err != nil || got.ID != qID {
		t.Fatalf("got %+v, %v", got, err)
	}
	st.err = domain.ErrNotFound
	_, err = svc.GetQuestion(t.Context(), subject.Subject{}, qID)
	wantCode(t, err, apperr.CodeNotFound)
	st.err = errDB
	_, err = svc.GetQuestion(t.Context(), subject.Subject{}, qID)
	wantCode(t, err, apperr.CodeInternal)
}

func TestETagCoversTheEditedFieldsOnly(t *testing.T) {
	base := question(qID, when)
	tag := service.ETag(base, "")
	if tag == "" || tag != service.ETag(base, "") {
		t.Fatalf("tag %q is not stable", tag)
	}
	for i, q := range []domain.Question{{ID: aID, UpdatedAt: when}, {ID: qID, UpdatedAt: when.Add(time.Microsecond)}} {
		if service.ETag(q, "") == tag {
			t.Errorf("variant %d keeps the tag", i)
		}
	}
	// The tag is the version as one viewer sees it: another viewer, or none, gets another tag.
	if service.ETag(base, userID) == tag || service.ETag(base, userID) == service.ETag(base, otherU) {
		t.Error("the tag does not cover the viewer")
	}
	// Answers, votes and acceptance do not touch updated_at and do not change the tag.
	same := base
	same.AnswerCount, same.VoteScore, same.AcceptedAnswerID = 3, 2, aID
	if service.ETag(same, "") != tag {
		t.Error("counters changed the tag")
	}
	if service.ETag(domain.Question{ID: qID, UpdatedAt: when.In(time.FixedZone("x", 3600))}, "") != service.ETag(domain.Question{ID: qID, UpdatedAt: when}, "") {
		t.Error("the tag depends on the time zone of the instant")
	}
}

func TestListQuestionsCursorCarriesTheRank(t *testing.T) {
	rows := []domain.Question{question(qID, when), question(aID, when.Add(-time.Hour))}
	rows[0].Rank = 0.0607927
	st := &fakeStore{questions: rows}
	svc, _ := newSvc(t, st)
	limit := 1
	page, err := svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Q: ptr("paiement"), Limit: &limit})
	if err != nil || page.NextCursor == "" {
		t.Fatalf("page %+v, %v", page, err)
	}
	if _, err := svc.ListQuestions(t.Context(), subject.Subject{}, service.QuestionListRequest{Q: ptr("paiement"), Limit: &limit, Cursor: page.NextCursor}); err != nil || st.listQ.AfterRank != 0.0607927 || st.listQ.AfterID != qID {
		t.Fatalf("after rank %v id %s, %v", st.listQ.AfterRank, st.listQ.AfterID, err)
	}
}
func createReq() service.CreateQuestionRequest {
	return service.CreateQuestionRequest{
		Title: "  Comment encaisser ?  ", BodyMD: "Corps\r\n", Lang: "fr", Country: ptr("BF"), Tags: []string{"go", "go", "api"},
		Snippets: []domain.Snippet{{Language: "go", Code: "fmt.Println()"}},
		Links:    []domain.LinkRef{{Type: "resource", ID: qID}, {Type: "resource", ID: qID}},
	}
}

func TestCreateQuestionStoresTheTextAsGivenForTheSubject(t *testing.T) {
	st := &fakeStore{question: question(qID, when)}
	svc, lim := newSvc(t, st)
	got, err := svc.CreateQuestion(t.Context(), member, createReq())
	if err != nil || got.ID != qID {
		t.Fatalf("got %+v, %v", got, err)
	}
	in := st.created
	if in.AuthorID != userID || in.Title != "  Comment encaisser ?  " || in.BodyMD != "Corps\r\n" || in.Country != "BF" ||
		strings.Join(in.Tags, ",") != "api,go" || len(in.Links) != 1 || len(in.Snippets) != 1 {
		t.Fatalf("stored %+v", in)
	}
	if lim.hits != 1 || lim.scope != "write:user" || lim.value != userID || lim.limit != 60 || lim.window != time.Minute {
		t.Fatalf("limiter %+v", lim)
	}
	// No country, no snippet, no link is fine.
	req := createReq()
	req.Country, req.Snippets, req.Links = nil, nil, nil
	if _, err := svc.CreateQuestion(t.Context(), member, req); err != nil || st.created.Country != "" {
		t.Fatalf("minimal question: %v %+v", err, st.created)
	}
}

func TestCreateQuestionAuthorizationAndValidation(t *testing.T) {
	tests := []struct {
		name   string
		sub    subject.Subject
		mutate func(*service.CreateQuestionRequest)
		code   apperr.Code
		field  string
		vcode  string
	}{
		{"anonymous", subject.Subject{}, nil, apperr.CodeUnauthenticated, "", ""},
		{"unverified email", unverified, nil, apperr.CodeEmailNotVerified, "", ""},
		{"blank title", member, func(r *service.CreateQuestionRequest) { r.Title = "   " }, apperr.CodeValidation, "/title", "invalid_value"},
		{"NUL title", member, func(r *service.CreateQuestionRequest) { r.Title = "a\x00b" }, apperr.CodeValidation, "/title", "invalid_value"},
		{"blank body", member, func(r *service.CreateQuestionRequest) { r.BodyMD = "\n\t" }, apperr.CodeValidation, "/body_md", "invalid_value"},
		{"no tag", member, func(r *service.CreateQuestionRequest) { r.Tags = nil }, apperr.CodeValidation, "/tags", "out_of_range"},
		{"six tags", member, func(r *service.CreateQuestionRequest) { r.Tags = []string{"a", "b", "c", "d", "e", "f"} }, apperr.CodeValidation, "/tags", "out_of_range"},
		{"two snippets", member, func(r *service.CreateQuestionRequest) {
			r.Snippets = []domain.Snippet{{Language: "go", Code: "a"}, {Language: "go", Code: "b"}}
		}, apperr.CodeValidation, "/snippets", "too_many_items"},
		{"blank snippet", member, func(r *service.CreateQuestionRequest) { r.Snippets = []domain.Snippet{{Language: "go", Code: " "}} }, apperr.CodeValidation, "/snippets", "invalid_value"},
		{"NUL snippet language", member, func(r *service.CreateQuestionRequest) { r.Snippets = []domain.Snippet{{Language: "g\x00", Code: "a"}} }, apperr.CodeValidation, "/snippets", "invalid_value"},
		{"long snippet language", member, func(r *service.CreateQuestionRequest) {
			r.Snippets = []domain.Snippet{{Language: strings.Repeat("a", 31), Code: "a"}}
		}, apperr.CodeValidation, "/snippets", "invalid_value"},
		{"question link", member, func(r *service.CreateQuestionRequest) { r.Links = []domain.LinkRef{{Type: "question", ID: qID}} }, apperr.CodeValidation, "/links", "invalid_value"},
		{"series link", member, func(r *service.CreateQuestionRequest) { r.Links = []domain.LinkRef{{Type: "series", ID: qID}} }, apperr.CodeValidation, "/links", "invalid_value"},
		{"link id", member, func(r *service.CreateQuestionRequest) { r.Links = []domain.LinkRef{{Type: "article", ID: "x"}} }, apperr.CodeValidation, "/links", "invalid_value"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{}
			svc, lim := newSvc(t, st)
			req := createReq()
			if tc.mutate != nil {
				tc.mutate(&req)
			}
			_, err := svc.CreateQuestion(t.Context(), tc.sub, req)
			if tc.field != "" {
				wantViolation(t, err, tc.field, tc.vcode)
			} else {
				wantCode(t, err, tc.code)
			}
			if st.writes != 0 || lim.hits != 0 {
				t.Fatalf("a refused request wrote %d and counted %d", st.writes, lim.hits)
			}
		})
	}
}

func TestCreateQuestionLimiterAndStoreErrors(t *testing.T) {
	st := &fakeStore{}
	svc, lim := newSvc(t, st)

	lim.err = errDB
	_, err := svc.CreateQuestion(t.Context(), member, createReq())
	wantCode(t, err, apperr.CodeInternal)
	if st.writes != 0 {
		t.Fatal("the limiter failed open")
	}

	lim.err, lim.allowed, lim.retry = nil, false, 12*time.Second
	_, err = svc.CreateQuestion(t.Context(), member, createReq())
	if ae := codeOf(t, err, apperr.CodeRateLimited); ae.RetryAfter != 12*time.Second || st.writes != 0 {
		t.Fatalf("retry %v writes %d", ae.RetryAfter, st.writes)
	}

	lim.allowed = true
	for _, tc := range []struct {
		err   error
		field string
		code  apperr.Code
	}{
		{domain.ErrLinkTarget, "/links", apperr.CodeValidation},
		{domain.ErrTagUnavailable, "/tags", apperr.CodeValidation},
		{domain.ErrUnknownCountry, "/country", apperr.CodeValidation},
		{domain.ErrNotFound, "", apperr.CodeNotFound},
		{errDB, "", apperr.CodeInternal},
	} {
		st.err = tc.err
		_, err = svc.CreateQuestion(t.Context(), member, createReq())
		if tc.field != "" {
			wantViolation(t, err, tc.field, "invalid_value")
		} else {
			wantCode(t, err, tc.code)
		}
	}
}

func TestUpdateQuestionDecidesOwnershipBeforeThePrecondition(t *testing.T) {
	tagSeen := ""
	check := func(tag string) error { tagSeen = tag; return nil }
	locked := domain.Question{ID: qID, Author: domain.Author{ID: userID}, UpdatedAt: when, AnswerCount: 2}

	t.Run("owner with the right tag", func(t *testing.T) {
		st := &fakeStore{question: question(qID, when), guardAs: locked}
		svc, _ := newSvc(t, st)
		got, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Title: ptr("Nouveau titre")}, check)
		if err != nil || got.ID != qID || *st.patched.Title != "Nouveau titre" || tagSeen != service.ETag(locked, userID) {
			t.Fatalf("got %+v, %v, tag %q", got, err, tagSeen)
		}
	})
	t.Run("non-owner is 403 and the tag is never looked at", func(t *testing.T) {
		tagSeen = ""
		st := &fakeStore{guardAs: domain.Question{ID: qID, Author: domain.Author{ID: otherU}}}
		svc, _ := newSvc(t, st)
		_, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Title: ptr("Nouveau titre")}, check)
		wantCode(t, err, apperr.CodeForbidden)
		if tagSeen != "" || st.writes != 0 {
			t.Fatalf("tag %q looked at, %d writes", tagSeen, st.writes)
		}
	})
	t.Run("a failed precondition is returned as it is", func(t *testing.T) {
		st := &fakeStore{guardAs: locked}
		svc, _ := newSvc(t, st)
		for _, code := range []apperr.Code{apperr.CodePreconditionRequired, apperr.CodePreconditionFailed} {
			_, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Title: ptr("Nouveau titre")},
				func(string) error { return apperr.New(code, "precondition") })
			wantCode(t, err, code)
		}
		if st.writes != 0 {
			t.Fatal("a refused precondition wrote")
		}
	})
	t.Run("anonymous is 401", func(t *testing.T) {
		st := &fakeStore{}
		svc, _ := newSvc(t, st)
		_, err := svc.UpdateQuestion(t.Context(), subject.Subject{}, qID, service.UpdateQuestionRequest{Title: ptr("Nouveau titre")}, check)
		wantCode(t, err, apperr.CodeUnauthenticated)
	})
	t.Run("store errors", func(t *testing.T) {
		for _, tc := range []struct {
			err   error
			field string
			code  apperr.Code
		}{
			{domain.ErrNotFound, "", apperr.CodeNotFound},
			{domain.ErrLinkTarget, "/links", apperr.CodeValidation},
			{domain.ErrTagUnavailable, "/tags", apperr.CodeValidation},
			{domain.ErrUnknownCountry, "/country", apperr.CodeValidation},
			{errDB, "", apperr.CodeInternal},
		} {
			st := &fakeStore{guardAs: locked, err: tc.err}
			svc, _ := newSvc(t, st)
			_, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Title: ptr("Nouveau titre")}, check)
			if tc.field != "" {
				wantViolation(t, err, tc.field, "invalid_value")
			} else {
				wantCode(t, err, tc.code)
			}
		}
	})
}

func TestUpdateQuestionBuildsAValidatedMergePatch(t *testing.T) {
	st := &fakeStore{guardAs: domain.Question{Author: domain.Author{ID: userID}}}
	svc, _ := newSvc(t, st)
	ok := func(string) error { return nil }

	empty := []domain.Snippet{}
	noLinks := []domain.LinkRef{}
	_, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{
		Title: ptr("Titre"), BodyMD: ptr("Corps"), Country: ptr(""), Tags: &[]string{"go", "go"}, Snippets: &empty, Links: &noLinks,
	}, ok)
	p := st.patched
	if err != nil || *p.Title != "Titre" || *p.BodyMD != "Corps" || *p.Country != "" || len(*p.Tags) != 1 || p.Snippets == nil || len(*p.Snippets) != 0 || p.Links == nil || len(*p.Links) != 0 {
		t.Fatalf("patch %+v, %v", p, err)
	}
	// Nothing named stays nil.
	if _, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{}, ok); err != nil || st.patched.Title != nil || st.patched.Tags != nil || st.patched.Links != nil || st.patched.Snippets != nil || st.patched.Country != nil {
		t.Fatalf("patch %+v, %v", st.patched, err)
	}
	// A snippet and a link are validated like on create.
	one := []domain.Snippet{{Language: "go", Code: "a"}}
	links := []domain.LinkRef{{Type: "project", ID: qID}}
	if _, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Snippets: &one, Links: &links}, ok); err != nil || len(*st.patched.Snippets) != 1 || len(*st.patched.Links) != 1 {
		t.Fatalf("patch %+v, %v", st.patched, err)
	}

	bad := []struct {
		name  string
		req   service.UpdateQuestionRequest
		field string
		code  string
	}{
		{"title", service.UpdateQuestionRequest{Title: ptr(" ")}, "/title", "invalid_value"},
		{"body", service.UpdateQuestionRequest{BodyMD: ptr(" ")}, "/body_md", "invalid_value"},
		{"tags", service.UpdateQuestionRequest{Tags: &[]string{}}, "/tags", "out_of_range"},
		{"snippets", service.UpdateQuestionRequest{Snippets: &[]domain.Snippet{{Language: "go", Code: ""}}}, "/snippets", "invalid_value"},
		{"links", service.UpdateQuestionRequest{Links: &[]domain.LinkRef{{Type: "series", ID: qID}}}, "/links", "invalid_value"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			writes := st.writes
			_, err := svc.UpdateQuestion(t.Context(), member, qID, tc.req, ok)
			wantViolation(t, err, tc.field, tc.code)
			if st.writes != writes {
				t.Fatal("a refused patch wrote")
			}
		})
	}
}

func answerAt(id string, accepted bool, score int, at time.Time) domain.Answer {
	return domain.Answer{ID: id, QuestionID: qID, Accepted: accepted, VoteScore: score, CreatedAt: at}
}

func TestListAnswersKeysTheCursorOnRankScoreTimeAndID(t *testing.T) {
	rows := []domain.Answer{answerAt(aID, true, -2, when), answerAt(otherU, false, 7, when.Add(time.Second)), answerAt(userID, false, 0, when)}
	st := &fakeStore{answers: rows}
	svc, _ := newSvc(t, st)
	limit := 1

	page, err := svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Limit: &limit})
	if err != nil || len(page.Items) != 1 || page.NextCursor == "" || st.listA.Limit != 2 || st.listA.After != nil || st.listA.QuestionID != qID {
		t.Fatalf("page %+v, %v, query %+v", page, err, st.listA)
	}
	if _, err := svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Limit: &limit, Cursor: page.NextCursor}); err != nil {
		t.Fatal(err)
	}
	k := st.listA.After
	if k == nil || k.Rank != 0 || k.NegScore != 2 || !k.CreatedAt.Equal(when) || k.ID != aID {
		t.Fatalf("after %+v", k)
	}

	// A non-accepted answer is rank 1.
	st.answers = []domain.Answer{answerAt(otherU, false, 7, when), answerAt(userID, false, 0, when)}
	page, _ = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Limit: &limit})
	if _, err := svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Limit: &limit, Cursor: page.NextCursor}); err != nil || st.listA.After.Rank != 1 || st.listA.After.NegScore != -7 {
		t.Fatalf("after %+v, %v", st.listA.After, err)
	}
	// The cursor belongs to its question.
	_, err = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: aID, Limit: &limit, Cursor: page.NextCursor})
	wantCode(t, err, apperr.CodeInvalidCursor)

	st.answers = nil
	page, err = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID})
	if err != nil || page.Items == nil || page.NextCursor != "" {
		t.Fatalf("empty page %+v, %v", page, err)
	}
}

func TestListAnswersRejectsBadInputAndMapsStoreErrors(t *testing.T) {
	c := codec(t)
	svc, _ := newSvc(t, &fakeStore{})
	_, err := svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Limit: ptr(0)})
	wantViolation(t, err, "limit", "out_of_range")
	_, err = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Cursor: "garbage"})
	wantCode(t, err, apperr.CodeInvalidCursor)

	hash := "7:answers36:" + qID
	for _, sort := range []string{"", "0:0", "x:0:" + when.Format(time.RFC3339Nano), "0:x:" + when.Format(time.RFC3339Nano), "0:0:never", "2:0:" + when.Format(time.RFC3339Nano)} {
		_, err = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Cursor: c.Encode(cursor.Payload{Sort: sort, ID: aID, FilterHash: hash})})
		wantCode(t, err, apperr.CodeInvalidCursor)
	}
	_, err = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID, Cursor: c.Encode(cursor.Payload{Sort: "0:0:" + when.Format(time.RFC3339Nano), ID: "nope", FilterHash: hash})})
	wantCode(t, err, apperr.CodeInvalidCursor)

	for _, tc := range []struct {
		err  error
		code apperr.Code
	}{{domain.ErrNotFound, apperr.CodeNotFound}, {errDB, apperr.CodeInternal}} {
		svc, _ := newSvc(t, &fakeStore{err: tc.err})
		_, err = svc.ListAnswers(t.Context(), subject.Subject{}, service.AnswerListRequest{QuestionID: qID})
		wantCode(t, err, tc.code)
	}
}

func TestCreateAnswer(t *testing.T) {
	req := service.CreateAnswerRequest{QuestionID: qID, BodyMD: "Réponse\n", Lang: "fr", Sources: []string{"https://a.example/x", "https://a.example/x", "http://b.example"}}

	st := &fakeStore{answer: answerAt(aID, false, 0, when)}
	svc, lim := newSvc(t, st)
	got, err := svc.CreateAnswer(t.Context(), member, req)
	if err != nil || got.ID != aID || st.newAns.AuthorID != userID || st.newAns.QuestionID != qID || st.newAns.BodyMD != "Réponse\n" ||
		len(st.newAns.Sources) != 2 || lim.hits != 1 || lim.scope != "write:user" {
		t.Fatalf("got %+v, %v, stored %+v", got, err, st.newAns)
	}

	tests := []struct {
		name   string
		sub    subject.Subject
		mutate func(*service.CreateAnswerRequest)
		code   apperr.Code
		field  string
		vcode  string
	}{
		{"anonymous", subject.Subject{}, nil, apperr.CodeUnauthenticated, "", ""},
		{"unverified", unverified, nil, apperr.CodeEmailNotVerified, "", ""},
		{"snippets", member, func(r *service.CreateAnswerRequest) { r.Snippets = []domain.Snippet{{Language: "go", Code: "a"}} }, apperr.CodeValidation, "/snippets", "too_many_items"},
		{"blank body", member, func(r *service.CreateAnswerRequest) { r.BodyMD = "  " }, apperr.CodeValidation, "/body_md", "invalid_value"},
		{"ftp source", member, func(r *service.CreateAnswerRequest) { r.Sources = []string{"ftp://a.example"} }, apperr.CodeValidation, "/sources", "invalid_format"},
		{"NUL source", member, func(r *service.CreateAnswerRequest) { r.Sources = []string{"https://a\x00"} }, apperr.CodeValidation, "/sources", "invalid_format"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{}
			svc, lim := newSvc(t, st)
			r := req
			if tc.mutate != nil {
				tc.mutate(&r)
			}
			_, err := svc.CreateAnswer(t.Context(), tc.sub, r)
			if tc.field != "" {
				wantViolation(t, err, tc.field, tc.vcode)
			} else {
				wantCode(t, err, tc.code)
			}
			if st.writes != 0 || lim.hits != 0 {
				t.Fatalf("a refused request wrote %d and counted %d", st.writes, lim.hits)
			}
		})
	}

	t.Run("limiter and store", func(t *testing.T) {
		st := &fakeStore{}
		svc, lim := newSvc(t, st)
		lim.err = errDB
		_, err := svc.CreateAnswer(t.Context(), member, req)
		wantCode(t, err, apperr.CodeInternal)
		lim.err, lim.allowed, lim.retry = nil, false, time.Second
		_, err = svc.CreateAnswer(t.Context(), member, req)
		wantCode(t, err, apperr.CodeRateLimited)
		lim.allowed = true
		for _, tc := range []struct {
			err  error
			code apperr.Code
		}{{domain.ErrNotFound, apperr.CodeNotFound}, {errDB, apperr.CodeInternal}} {
			st.err = tc.err
			_, err = svc.CreateAnswer(t.Context(), member, req)
			wantCode(t, err, tc.code)
		}
	})
}

func TestUpdateQuestionDrawsOnTheWriteBudgetAndFailsClosed(t *testing.T) {
	ok := func(string) error { return nil }
	req := service.UpdateQuestionRequest{Title: ptr("Nouveau titre")}
	st := &fakeStore{guardAs: domain.Question{Author: domain.Author{ID: userID}}}
	svc, lim := newSvc(t, st)
	if _, err := svc.UpdateQuestion(t.Context(), member, qID, req, ok); err != nil || lim.hits != 1 || lim.scope != "write:user" || lim.limit != 60 || lim.value != userID {
		t.Fatalf("err %v, limiter %+v", err, lim)
	}
	writes := st.writes
	lim.allowed, lim.retry = false, 7*time.Second
	_, err := svc.UpdateQuestion(t.Context(), member, qID, req, ok)
	if ae := codeOf(t, err, apperr.CodeRateLimited); ae.RetryAfter != 7*time.Second || st.writes != writes {
		t.Fatalf("retry %v, writes %d", ae.RetryAfter, st.writes)
	}
	lim.allowed, lim.err = true, errDB
	_, err = svc.UpdateQuestion(t.Context(), member, qID, req, ok)
	wantCode(t, err, apperr.CodeInternal)
	if st.writes != writes {
		t.Fatal("the limiter failed open")
	}
	// A refused body or anonymous caller does not spend the budget.
	lim.err, lim.hits = nil, 0
	_, _ = svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Title: ptr(" ")}, ok)
	_, _ = svc.UpdateQuestion(t.Context(), subject.Subject{}, qID, req, ok)
	if lim.hits != 0 {
		t.Fatalf("%d hits for refused requests", lim.hits)
	}
}

func TestTagsAreSortedAndDistinctBeforeTheyReachTheStore(t *testing.T) {
	st := &fakeStore{guardAs: domain.Question{Author: domain.Author{ID: userID}}}
	svc, _ := newSvc(t, st)
	req := createReq()
	req.Tags = []string{"zeta", "alpha", "zeta", "mid"}
	if _, err := svc.CreateQuestion(t.Context(), member, req); err != nil || strings.Join(st.created.Tags, ",") != "alpha,mid,zeta" {
		t.Fatalf("created tags %v, %v", st.created.Tags, err)
	}
	if _, err := svc.UpdateQuestion(t.Context(), member, qID, service.UpdateQuestionRequest{Tags: &[]string{"b", "a", "b"}}, func(string) error { return nil }); err != nil || strings.Join(*st.patched.Tags, ",") != "a,b" {
		t.Fatalf("patched tags %v, %v", st.patched.Tags, err)
	}
}

// The viewer of a read is the session subject, never anything the request carries: it reaches the store
// as the subject's id, and empty for an anonymous caller (issue #151).
func TestReadsPassTheSessionSubjectAsTheViewer(t *testing.T) {
	for name, sub := range map[string]subject.Subject{"member": member, "anonymous": {}} {
		t.Run(name, func(t *testing.T) {
			st := &fakeStore{question: question(qID, when)}
			svc, _ := newSvc(t, st)
			if _, err := svc.GetQuestion(t.Context(), sub, qID); err != nil || st.viewerSeen != sub.UserID {
				t.Fatalf("detail viewer %q, %v, want %q", st.viewerSeen, err, sub.UserID)
			}
			if _, err := svc.ListQuestions(t.Context(), sub, service.QuestionListRequest{}); err != nil || st.listQ.ViewerID != sub.UserID {
				t.Fatalf("list viewer %q, %v, want %q", st.listQ.ViewerID, err, sub.UserID)
			}
			if _, err := svc.ListAnswers(t.Context(), sub, service.AnswerListRequest{QuestionID: qID}); err != nil || st.listA.ViewerID != sub.UserID {
				t.Fatalf("answers viewer %q, %v, want %q", st.listA.ViewerID, err, sub.UserID)
			}
		})
	}
}
