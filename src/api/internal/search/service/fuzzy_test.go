package service

import (
	"errors"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/search/domain"
)

// With Fuzzy on, a group whose full text finds nothing falls back to the trigram query; a group with
// full-text hits keeps them and never runs the fallback (SPEC 6, typo tolerance).
func TestFuzzySearchFallsBackOnlyForGroupsWithoutFullTextHits(t *testing.T) {
	r := &fakeReader{
		rows:    map[domain.Group]int{domain.Questions: 2},
		similar: map[domain.Group]int{domain.Questions: 9, domain.Projects: 7},
	}
	got, err := New(r, allow, nil).Search(t.Context(), Request{Q: "paiemnt", Fuzzy: true, PerGroup: ints(5)})
	if err != nil {
		t.Fatal(err)
	}
	if q := got.Groups[domain.Questions]; len(q.Items) != 2 || q.HasMore || q.Items[0].ID != "id-questions-0" {
		t.Fatalf("questions kept their full-text hits? %+v", q)
	}
	p := got.Groups[domain.Projects]
	if len(p.Items) != 5 || !p.HasMore || p.Items[0].ID != "near-projects-0" || p.Items[0].Title != "near" {
		t.Fatalf("projects fallback: %+v", p)
	}
	if res := got.Groups[domain.Resources]; len(res.Items) != 0 || res.Items == nil {
		t.Fatalf("resources must stay an empty list: %#v", res.Items)
	}
	groups := map[domain.Group]bool{}
	for _, c := range r.similarCalls {
		groups[c.group] = true
		if c.query.Limit != 6 || c.query.Text != "paiemnt" {
			t.Fatalf("fallback query %+v, want the same text and per_group+1", c.query)
		}
	}
	if groups[domain.Questions] || !groups[domain.Resources] || !groups[domain.Projects] || !groups[domain.Learn] {
		t.Fatalf("fallback ran for %v, want resources, projects and learn only", groups)
	}
}

// Without Fuzzy (the assistant's retrieval), nothing found stays nothing: no fallback, so the assistant
// never grounds an answer on a near match (AI-1).
func TestSearchWithoutFuzzyNeverFallsBack(t *testing.T) {
	r := &fakeReader{similar: map[domain.Group]int{domain.Questions: 3}}
	got, err := New(r, allow, nil).Search(t.Context(), Request{Q: "paiemnt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.similarCalls) != 0 || len(got.Groups[domain.Questions].Items) != 0 {
		t.Fatalf("fallback ran without Fuzzy: %d calls", len(r.similarCalls))
	}
}

// A trigram needs 3 characters: a shorter trimmed query never falls back (DATABASE_DESIGN section 7).
func TestFuzzySearchNeedsThreeCharacters(t *testing.T) {
	r := &fakeReader{similar: map[domain.Group]int{domain.Questions: 3}}
	if _, err := New(r, allow, nil).Search(t.Context(), Request{Q: " ab ", Fuzzy: true}); err != nil {
		t.Fatal(err)
	}
	if len(r.similarCalls) != 0 {
		t.Fatalf("fallback ran for a 2-character query: %d calls", len(r.similarCalls))
	}
}

// pg_trgm ignores "-", so the fallback would return the very titles a negated word excludes. A query
// that negates a word in websearch_to_tsquery syntax never falls back; a hyphen inside a word or
// standing alone is not a negation.
func TestFuzzySearchSkipsTheFallbackWhenTheQueryNegatesAWord(t *testing.T) {
	tests := []struct {
		name         string
		q            string
		wantFallback bool
	}{
		{"trailing negation", "paiemnt -mobile", false},
		{"leading negation", "-mobile paiemnt", false},
		{"negated quoted word", `paiemnt "-mobile"`, false},
		{"negated word in parentheses", "paiemnt (-mobile", false},
		{"negated digit", "paiemnt -3d", false},
		{"negated phrase", `paiemnt -"mobile app"`, false},
		{"negated group", "paiemnt -(mobile)", false},
		{"negated open group", "paiemnt -(mobile", false},
		{"negated quoted word after the hyphen", `paiemnt -"mobile"`, false},
		{"negation after an operator", "paiemnt !-mobile", false},
		{"hyphen inside a word", "e-commerce paiemnt", true},
		{"lone hyphen", "paiemnt - mobile", true},
		{"hyphen before a symbol", "paiemnt -+ mobile", true},
		{"no hyphen", "paiemnt", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeReader{similar: map[domain.Group]int{domain.Questions: 3}}
			if _, err := New(r, allow, nil).Search(t.Context(), Request{Q: tc.q, Fuzzy: true}); err != nil {
				t.Fatal(err)
			}
			if got := len(r.similarCalls) > 0; got != tc.wantFallback {
				t.Fatalf("q %q: fallback ran = %v (%d calls), want %v", tc.q, got, len(r.similarCalls), tc.wantFallback)
			}
		})
	}
}

func TestFuzzySearchFallbackErrorIsReturned(t *testing.T) {
	boom := errors.New("db down")
	r := &fakeReader{similarErr: map[domain.Group]error{domain.Learn: boom}}
	if _, err := New(r, allow, nil).Search(t.Context(), Request{Q: "paiemnt", Fuzzy: true}); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the fallback error", err)
	}
}
