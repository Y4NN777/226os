package assistant_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/assistant"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	searchdomain "github.com/Y4NN777/226os/src/api/internal/search/domain"
	searchservice "github.com/Y4NN777/226os/src/api/internal/search/service"
)

type searcher struct {
	res  searchservice.Results
	err  error
	last searchservice.Request
}

func (s *searcher) Search(_ context.Context, req searchservice.Request) (searchservice.Results, error) {
	s.last = req
	return s.res, s.err
}

func hit(g searchdomain.Group, id, snippet string) searchdomain.Hit {
	h := searchdomain.Hit{Type: g, ID: id, Title: "T " + id, Snippet: snippet}
	if id == "q1" {
		h.PublishedAt = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	}
	return h
}

func TestRetrieveSearchesAsAMemberWithAnOrQuery(t *testing.T) {
	s := &searcher{}
	got, err := assistant.NewRetriever(s).Retrieve(t.Context(), "How do I verify a webhook signature?", 5)
	if err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
	if s.last.Q != "verify or webhook or signature" || s.last.Anonymous || s.last.PerGroup == nil || *s.last.PerGroup != 5 {
		t.Fatalf("search request: %+v", s.last)
	}
}

func TestRetrieveWithoutKeywordsDoesNotSearch(t *testing.T) {
	s := &searcher{err: errors.New("must not be called")}
	got, err := assistant.NewRetriever(s).Retrieve(t.Context(), "how do you do", 5)
	if err != nil || got != nil {
		t.Fatalf("%v %v", got, err)
	}
}

func TestRetrieveReportsSearchFailure(t *testing.T) {
	s := &searcher{err: errors.New("pg down")}
	if _, err := assistant.NewRetriever(s).Retrieve(t.Context(), "webhook retry", 5); err == nil {
		t.Fatal("want the error")
	}
}

func TestRetrieveTakesOneHitPerGroupInTurnAndCapsText(t *testing.T) {
	long := strings.Repeat("é", 2000)
	s := &searcher{res: searchservice.Results{Groups: map[searchdomain.Group]searchservice.GroupResult{
		searchdomain.Questions: {Items: []searchdomain.Hit{hit(searchdomain.Questions, "q1", long), hit(searchdomain.Questions, "q2", "b"), hit(searchdomain.Questions, "q3", "c")}},
		searchdomain.Learn:     {Items: []searchdomain.Hit{hit(searchdomain.Learn, "a1", "d"), hit(searchdomain.Learn, "a2", "e")}},
		searchdomain.Resources: {Items: []searchdomain.Hit{hit(searchdomain.Resources, "r1", "f")}},
		searchdomain.Projects:  {Items: []searchdomain.Hit{hit(searchdomain.Projects, "p1", "g")}},
	}}}
	got, err := assistant.NewRetriever(s).Retrieve(t.Context(), "webhook retry", 5)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, p := range got {
		ids = append(ids, p.ID)
	}
	if !slices.Equal(ids, []string{"q1", "r1", "p1", "a1", "q2"}) {
		t.Fatalf("order %v", ids)
	}
	if n := len([]rune(got[0].Text)); n > 1500 {
		t.Fatalf("passage has %d runes", n)
	}
	want := map[string][2]string{"q1": {"question", "/questions/q1"}, "r1": {"resource", "/discover/r1"}, "p1": {"project", "/build/p1"}, "a1": {"article", "/apprendre/articles/a1"}}
	if !got[0].PublishedAt.Equal(time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)) || !got[1].PublishedAt.IsZero() {
		t.Fatalf("dates must follow the hits (zero = undated): %v %v", got[0].PublishedAt, got[1].PublishedAt)
	}
	for _, p := range got[:4] {
		if w := want[p.ID]; p.Type != w[0] || p.URL != w[1] {
			t.Errorf("%s: %+v", p.ID, p)
		}
	}
}

type deny struct{}

func (deny) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return false, time.Minute, nil
}

func TestNewHandlerWires(t *testing.T) {
	if assistant.NewHandler(false, 0, &searcher{}, llm.Disabled{}, deny{}, nil) == nil {
		t.Fatal("nil handler")
	}
}
