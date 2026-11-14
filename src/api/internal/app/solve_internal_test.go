package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/solve"
)

// The six Solve routes answer 501 without a handler and reach the handler when it is set.
func TestSolveRoutesDelegateOnlyWhenWired(t *testing.T) {
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	calls := map[string]func(*Server, http.ResponseWriter, *http.Request){
		"listQuestions": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ListQuestions(w, r, apigen.ListQuestionsParams{})
		},
		"createQuestion": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.CreateQuestion(w, r, apigen.CreateQuestionParams{})
		},
		"getQuestion": func(s *Server, w http.ResponseWriter, r *http.Request) { s.GetQuestion(w, r, id) },
		"updateQuestion": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.UpdateQuestion(w, r, id, apigen.UpdateQuestionParams{})
		},
		"listAnswers": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ListAnswers(w, r, id, apigen.ListAnswersParams{})
		},
		"createAnswer": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.CreateAnswer(w, r, id, apigen.CreateAnswerParams{})
		},
	}
	wired := NewServer(WithSolve(solve.NewHandler(&fakePool{}, c, allowLimiter{}, nil)))
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			req := func() *http.Request {
				r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{}`))
				r.Header.Set("Content-Type", "application/json")
				return r
			}
			rec := httptest.NewRecorder()
			call(NewServer(), rec, req())
			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("unwired status %d", rec.Code)
			}
			rec = httptest.NewRecorder()
			call(wired, rec, req())
			if rec.Code == http.StatusNotImplemented {
				t.Fatal("wired route still 501")
			}
		})
	}
}
