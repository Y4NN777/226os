package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/build"
	"github.com/Y4NN777/226os/src/api/internal/build/github"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// The six S5 routes answer 501 without a handler and reach the handler when it is set. The calls are
// anonymous, so no wired route reaches GitHub.
func TestBuildRoutesDelegateOnlyWhenWired(t *testing.T) {
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	calls := map[string]func(*Server, http.ResponseWriter, *http.Request){
		"listProjects": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ListProjects(w, r, apigen.ListProjectsParams{})
		},
		"createProject": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.CreateProject(w, r, apigen.CreateProjectParams{})
		},
		"getProject":    func(s *Server, w http.ResponseWriter, r *http.Request) { s.GetProject(w, r, id) },
		"updateProject": func(s *Server, w http.ResponseWriter, r *http.Request) { s.UpdateProject(w, r, id) },
		"listContributionNeeds": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.ListContributionNeeds(w, r, id, apigen.ListContributionNeedsParams{})
		},
		"createContributionNeed": func(s *Server, w http.ResponseWriter, r *http.Request) {
			s.CreateContributionNeed(w, r, id, apigen.CreateContributionNeedParams{})
		},
	}
	wired := NewServer(WithBuild(build.NewHandler(&fakePool{}, github.New(), c, allowLimiter{}, nil)))
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
