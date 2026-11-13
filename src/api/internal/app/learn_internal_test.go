package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/learn"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// The five Learn routes answer 501 without a handler and reach the handler when it is set.
func TestLearnRoutesDelegateOnlyWhenWired(t *testing.T) {
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	calls := map[string]func(*Server, http.ResponseWriter, *http.Request){
		"listLearnArticles": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ListLearnArticles(w, r, apigen.ListLearnArticlesParams{})
		},
		"createLearnArticle": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.CreateLearnArticle(w, r, apigen.CreateLearnArticleParams{})
		},
		"getLearnArticle": func(s *Server, w http.ResponseWriter, r *http.Request) { s.GetLearnArticle(w, r, id) },
		"listTrainingSeries": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ListTrainingSeries(w, r, apigen.ListTrainingSeriesParams{})
		},
		"getTrainingSeries": func(s *Server, w http.ResponseWriter, r *http.Request) { s.GetTrainingSeries(w, r, id) },
	}
	wired := NewServer(WithLearn(learn.NewHandler(&fakePool{}, c, allowLimiter{}, nil)))
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
